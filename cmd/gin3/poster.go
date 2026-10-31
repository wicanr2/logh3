package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/logh3/internal/localize"
)

const (
	posterWidth  = 320
	posterGameW  = 1920
	posterHeight = 1440
	posterTotalW = posterGameW + 2*posterWidth
)

// Spec151: art stays outside the engine. CPU images load once; GPU panels and
// the central viewport are reused. Game coordinates and the six save tables
// never include the decorative sides.
type posterArt struct {
	approved   *approvedPosterArt
	images     [2]image.Image
	sources    [2]*ebiten.Image
	panels     [2]*ebiten.Image
	viewport   *ebiten.Image
	alternates [2]*posterArt
	selected   int
	catalog    *localize.Catalog
	quotes     map[string][2]*ebiten.Image
	narrow     bool
}

func loadPosterArt(art string) (*posterArt, error) {
	approved, approvedErr := loadApprovedPosterArt(art)
	if approvedErr == nil {
		return &posterArt{approved: approved}, approved.quoteError
	}
	p, err := loadPosterPair(art, "")
	if err != nil {
		p = &posterArt{}
	}
	for i, prefix := range []string{"v2-", "v3-"} {
		p.alternates[i], _ = loadPosterPair(art, prefix)
	}
	if !p.available() {
		return nil, err
	}
	if !p.pairAvailable() {
		for i, pair := range p.alternates {
			if pair.pairAvailable() {
				p.selected = i + 1
				break
			}
		}
	}
	if !os.IsNotExist(approvedErr) {
		return p, approvedErr
	}
	return p, nil
}

func loadPosterPair(art, prefix string) (*posterArt, error) {
	p := &posterArt{}
	if prefix != "" {
		_, left := os.Stat(filepath.Join(art, "poster", prefix+"empire-narrow.png"))
		_, right := os.Stat(filepath.Join(art, "poster", prefix+"alliance-narrow.png"))
		p.narrow = left == nil && right == nil
	}
	for i, name := range []string{"empire.png", "alliance.png"} {
		if p.narrow {
			name = strings.TrimSuffix(name, ".png") + "-narrow.png"
		}
		path := filepath.Join(art, "poster", prefix+name)
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		config, err := png.DecodeConfig(f)
		if err == nil && (config.Width <= 0 || config.Height <= 0 || config.Width > 8192 || config.Height > 8192) {
			err = fmt.Errorf("海報尺寸無效：%s", path)
		}
		if err == nil {
			_, err = f.Seek(0, 0)
		}
		if err == nil {
			p.images[i], err = png.Decode(f)
		}
		_ = f.Close()
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *posterArt) available() bool {
	if p == nil {
		return false
	}
	if p.pairAvailable() {
		return true
	}
	for _, pair := range p.alternates {
		if pair.pairAvailable() {
			return true
		}
	}
	return false
}

func (p *posterArt) pairAvailable() bool {
	return p != nil && (p.approved != nil || p.images[0] != nil && p.images[1] != nil)
}

func (p *posterArt) setAvailable(index int) bool {
	if p == nil || index < 0 || index >= 3 {
		return false
	}
	if p.approved != nil {
		return p.approved.bases[index] != nil || p.approved.files[fmt.Sprintf("base-v%d.png", index+1)] != nil
	}
	if index == 0 {
		return p.pairAvailable()
	}
	return p.alternates[index-1].pairAvailable()
}

func (p *posterArt) draw(screen, game *ebiten.Image) {
	pair := p
	if p.selected > 0 {
		pair = p.alternates[p.selected-1]
	}
	for i := range pair.panels {
		if pair.panels[i] == nil {
			pair.sources[i] = ebiten.NewImageFromImage(pair.images[i])
			pair.panels[i] = ebiten.NewImage(posterWidth, posterHeight)
			pair.panels[i].Fill(color.RGBA{5, 16, 32, 255})
			b := pair.sources[i].Bounds()
			artHeight := float64(posterHeight)
			if p.selected == 0 {
				artHeight = 1120
			}
			scale := math.Max(float64(posterWidth)/float64(b.Dx()), artHeight/float64(b.Dy()))
			if p.selected > 0 && !pair.narrow {
				scale = math.Min(float64(posterWidth)/float64(b.Dx()), artHeight/float64(b.Dy()))
			}
			focus := .5
			if i == 0 && p.selected == 0 {
				focus = .56 // Keep Reinhard's boots and body; the outer cape can reach the bezel edge.
			}
			op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(-(float64(b.Dx())*scale-posterWidth)*focus, (artHeight-float64(b.Dy())*scale)/2)
			pair.panels[i].DrawImage(pair.sources[i], op)
		}
		op := &ebiten.DrawImageOptions{}
		if i == 1 {
			op.GeoM.Translate(posterWidth+posterGameW, 0)
		}
		screen.DrawImage(pair.panels[i], op)
	}
	if p.selected == 0 && p.catalog != nil {
		p.drawQuotes(screen)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(posterGameW)/float64(game.Bounds().Dx()), float64(posterHeight)/float64(game.Bounds().Dy()))
	op.GeoM.Translate(posterWidth, 0)
	screen.DrawImage(game, op)
}

func (p *posterArt) setName() string {
	if p == nil {
		return "v1"
	}
	return []string{"v1", "v2", "v3"}[p.selected]
}

func (p *posterArt) selectSet(name string) bool {
	if p == nil {
		return false
	}
	for i, value := range []string{"v1", "v2", "v3"} {
		if name == value && p.setAvailable(i) {
			p.selected = i
			return true
		}
	}
	return false
}

func readPosterSetPreference(path string) (string, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "v1", nil
	}
	if err != nil {
		return "v1", err
	}
	var prefs map[string]json.RawMessage
	if json.Unmarshal(b, &prefs) != nil || prefs == nil {
		return "v1", fmt.Errorf("invalid poster preferences")
	}
	if raw, ok := prefs["poster_set"]; ok {
		var set string
		if json.Unmarshal(raw, &set) != nil || set != "v1" && set != "v2" && set != "v3" {
			return "v1", fmt.Errorf("invalid saved poster set")
		}
		return set, nil
	}
	return "v1", nil
}

func (a *app) cyclePoster() {
	if !a.posters.available() {
		return
	}
	if !a.posterMode {
		a.setPosterMode(true)
		return
	}
	old := a.posters.selected
	for next := old + 1; next < 3; next++ {
		if !a.posters.setAvailable(next) {
			continue
		}
		name := []string{"v1", "v2", "v3"}[next]
		if a.preferencesPath != "" {
			raw, _ := json.Marshal(name)
			if err := writePreferences(a.preferencesPath, map[string]json.RawMessage{"poster_set": raw}); err != nil {
				a.posterError = err.Error()
				return
			}
		}
		a.posters.selected, a.posterError = next, ""
		return
	}
	if a.preferencesPath != "" {
		if err := writePreferences(a.preferencesPath, map[string]json.RawMessage{"poster_mode": json.RawMessage("false"), "poster_set": json.RawMessage(`"v1"`)}); err != nil {
			a.posterError = err.Error()
			return
		}
	}
	a.posters.selected = 0
	// Preference transaction above succeeded; resize without another write.
	path := a.preferencesPath
	a.preferencesPath = ""
	a.setPosterMode(false)
	a.preferencesPath = path
}

func (p *posterArt) drawQuotes(screen *ebiten.Image) {
	locale := p.catalog.Locale
	if p.quotes == nil {
		p.quotes = map[string][2]*ebiten.Image{}
	}
	textures, ok := p.quotes[locale]
	if !ok {
		for i, key := range []string{"empire", "alliance"} {
			pixels := image.NewRGBA(image.Rect(0, 0, posterWidth, 320))
			value := p.catalog.Value("poster.quote." + key)
			var lines []string
			line := ""
			// Spaces are preferred break points for English; CJK may break by rune.
			parts := strings.Fields(value)
			if locale != "en" && locale != "ko" {
				parts = nil
				for _, ch := range value {
					parts = append(parts, string(ch))
				}
			}
			for _, part := range parts {
				candidate := line + part
				if (locale == "en" || locale == "ko") && line != "" {
					candidate = line + " " + part
				}
				w, _ := localize.Measure(locale, candidate, 36)
				if w > 260 && line != "" {
					lines = append(lines, line)
					line = part
				} else {
					line = candidate
				}
			}
			if line != "" {
				lines = append(lines, line)
			}
			for j, text := range lines {
				localize.DrawRGBA(pixels, locale, text, image.Rect(28, 28+j*48, 288, 76+j*48), color.RGBA{238, 221, 177, 255}, 36)
			}
			localize.DrawRGBA(pixels, locale, p.catalog.Value("poster.name."+key), image.Rect(32, 260, 288, 300), color.RGBA{170, 192, 211, 255}, 25)
			source := ebiten.NewImageFromImage(pixels)
			textures[i] = ebiten.NewImage(posterWidth, 320)
			op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
			op.GeoM.Skew(-.12, 0)
			op.GeoM.Translate(22, 0)
			textures[i].DrawImage(source, op)
			source.Dispose()
		}
		p.quotes[locale] = textures
	}
	for i, texture := range textures {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Translate(float64(i)*(posterWidth+posterGameW), 1120)
		screen.DrawImage(texture, op)
	}
}

func readPosterPreference(path string) (bool, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var prefs map[string]json.RawMessage
	if err := json.Unmarshal(b, &prefs); err != nil || prefs == nil {
		return false, fmt.Errorf("海報偏好檔不是有效物件")
	}
	raw, exists := prefs["poster_mode"]
	if !exists {
		return false, nil
	}
	var enabled bool
	if string(raw) == "null" {
		return false, fmt.Errorf("海報偏好不是開關")
	}
	err = json.Unmarshal(raw, &enabled)
	return enabled, err
}

func posterWindowWidth(height int, enabled bool) int {
	ratio := 4.0 / 3
	if enabled {
		ratio = 16.0 / 9
	}
	return max(1, int(math.Round(float64(height)*ratio)))
}

func (a *app) setPosterMode(enabled bool) bool {
	if enabled && !a.posters.available() {
		return false
	}
	if a.preferencesPath != "" {
		raw, _ := json.Marshal(enabled)
		if err := writePreferences(a.preferencesPath, map[string]json.RawMessage{"poster_mode": raw}); err != nil {
			a.posterError = err.Error()
			return false
		}
	}
	a.posterMode, a.posterError = enabled, ""
	if a.windowReady {
		_, height := ebiten.WindowSize()
		ebiten.SetWindowSize(posterWindowWidth(height, enabled), height)
	}
	return true
}
