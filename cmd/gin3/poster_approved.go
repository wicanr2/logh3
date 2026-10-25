package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

const approvedPosterSHA = "ac843a911fce876e67292940285d22f19c988e092465fe09ab0746f6bf34c202"

type posterViewport struct{ X, Y, W, H float64 }

func (a *app) usingApprovedPoster() bool {
	return a.posterMode && a.posters != nil && a.posters.approved != nil
}

// Spec157: drawing and input share this transform, in the fixed 2560x1440 canvas.
func (a *app) posterViewport() posterViewport {
	v := posterViewport{posterWidth, 0, posterGameW, posterHeight}
	if a.usingApprovedPoster() {
		v.W *= .88
		v.H *= .88
		v.X = (posterTotalW - v.W) / 2
		v.Y = (posterHeight - v.H) / 2
	}
	return v
}

type approvedPosterArt struct {
	bases         [3]image.Image
	hulls         image.Image
	labels        map[string]image.Image
	baseTextures  [3]*ebiten.Image
	hullTexture   *ebiten.Image
	labelTextures map[string]*ebiten.Image
	overlay       *ebiten.Image
}

func loadApprovedPosterArt(art string) (*approvedPosterArt, error) {
	dir := filepath.Join(art, "poster", "approved")
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var manifest struct {
		Format   string `json:"format"`
		Canvas   [2]int `json:"canvas"`
		Proposal string `json:"proposal_sha256"`
		Scale    int    `json:"game_scale"`
		Files    map[string]struct {
			SHA    string `json:"sha256"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"files"`
	}
	if err = json.Unmarshal(b, &manifest); err != nil {
		return nil, err
	}
	if manifest.Format != "logh3-approved-poster-v1" || manifest.Canvas != [2]int{posterTotalW, posterHeight} || manifest.Proposal != approvedPosterSHA || manifest.Scale != 88 {
		return nil, fmt.Errorf("定案海報manifest不相符")
	}
	read := func(name string) (image.Image, error) {
		item, ok := manifest.Files[name]
		if !ok || item.Width != posterTotalW || item.Height != posterHeight {
			return nil, fmt.Errorf("定案海報缺少或尺寸錯誤：%s", name)
		}
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > 32<<20 {
			return nil, fmt.Errorf("定案海報檔案無效：%s", name)
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if fmt.Sprintf("%x", sha256.Sum256(bytes)) != item.SHA {
			return nil, fmt.Errorf("定案海報雜湊錯誤：%s", name)
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		config, err := png.DecodeConfig(f)
		if err != nil {
			return nil, err
		}
		if config.Width != posterTotalW || config.Height != posterHeight {
			return nil, fmt.Errorf("定案海報PNG尺寸錯誤：%s", name)
		}
		if _, err = f.Seek(0, 0); err != nil {
			return nil, err
		}
		return png.Decode(f)
	}
	p := &approvedPosterArt{labels: map[string]image.Image{}, labelTextures: map[string]*ebiten.Image{}}
	for i, set := range []string{"v1", "v2", "v3"} {
		if p.bases[i], err = read("base-" + set + ".png"); err != nil {
			return nil, err
		}
	}
	if p.hulls, err = read("hulls.png"); err != nil {
		return nil, err
	}
	for _, locale := range []string{"zh-TW", "zh-CN", "ja", "en", "ko"} {
		if p.labels[locale], err = read("labels-" + locale + ".png"); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func approvedPosterBand(v posterViewport, b image.Rectangle, gameW int, bottom bool) image.Rectangle {
	if b.Empty() || gameW <= 0 {
		return image.Rectangle{}
	}
	sy := v.W / float64(gameW)
	y0, y1 := 0, int(math.Floor(v.Y+float64(b.Max.Y)*sy))
	if bottom {
		y0, y1 = int(math.Ceil(v.Y+float64(b.Min.Y)*sy)), posterHeight
	}
	// Include the fractional edge pixels of the scaled empty rows, so their
	// antialiasing cannot leave a one-pixel dark seam against the side portraits.
	return image.Rect(int(math.Floor(v.X)), y0, int(math.Ceil(v.X+v.W)), y1).Intersect(image.Rect(0, 0, posterTotalW, posterHeight))
}

func (a *app) drawApprovedPoster(screen, game *ebiten.Image) {
	p := a.posters
	art := p.approved
	i := p.selected
	if art.baseTextures[i] == nil {
		art.baseTextures[i] = ebiten.NewImageFromImage(art.bases[i])
	}
	if art.hullTexture == nil {
		art.hullTexture = ebiten.NewImageFromImage(art.hulls)
	}
	base := art.baseTextures[i]
	screen.DrawImage(base, nil)
	v := a.posterViewport()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(v.W/float64(game.Bounds().Dx()), v.H/float64(game.Bounds().Dy()))
	op.GeoM.Translate(v.X, v.Y)
	screen.DrawImage(game, op)
	if d := a.posterDecor; d != nil {
		for band, b := range d.bands {
			bounds := approvedPosterBand(v, b, game.Bounds().Dx(), band == 1)
			if bounds.Empty() {
				continue
			}
			target := screen.SubImage(bounds).(*ebiten.Image)
			target.DrawImage(base, nil)
			target.DrawImage(art.hullTexture, nil)
		}
	}
	if i == 0 {
		locale := "zh-TW"
		if p.catalog != nil {
			locale = p.catalog.Locale
		}
		if art.labels[locale] == nil {
			locale = "zh-TW"
		}
		if art.labelTextures[locale] == nil {
			art.labelTextures[locale] = ebiten.NewImageFromImage(art.labels[locale])
		}
		screen.DrawImage(art.labelTextures[locale], nil)
	}
	a.drawApprovedPosterWeapons(screen, game.Bounds().Dx())
	// Pointer, toolbar and help stay above the restored empty bands and artwork.
	if art.overlay == nil || art.overlay.Bounds() != game.Bounds() {
		art.overlay = ebiten.NewImage(game.Bounds().Dx(), game.Bounds().Dy())
	}
	art.overlay.Clear()
	a.drawPointer(art.overlay, a.posterPointer)
	a.drawToolbar(art.overlay, a.posterTheme)
	screen.DrawImage(art.overlay, op)
}

func (a *app) drawApprovedPosterWeapons(screen *ebiten.Image, gameW int) {
	d := a.posterDecor
	if d == nil || a.g.DisplaySettings().Anim == 0 {
		return
	}
	if d.white == nil {
		d.white = ebiten.NewImage(1, 1)
		d.white.Fill(color.White)
	}
	t := float64(d.phase) / 15
	for band, b := range d.bands {
		if band == 0 && (t <= 1 || t >= 2.2) || band == 1 && (t <= 6 || t >= 7.2) {
			continue
		}
		clip := approvedPosterBand(a.posterViewport(), b, gameW, band == 1)
		if clip.Empty() {
			continue
		}
		x, y, length := 365.44*2, 97.76*.51*2, -100.
		colour := color.NRGBA{189, 215, 232, 204}
		if band == 1 {
			x, y, length = (1062.4-78)*2, (594.96+125.04*.49+24)*2, 100
			colour = color.NRGBA{226, 205, 157, 204}
		}
		op := &ebiten.DrawImageOptions{}
		if length < 0 {
			x += length
			length = -length
		}
		op.GeoM.Scale(length, 1.6)
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(colour)
		screen.SubImage(clip).(*ebiten.Image).DrawImage(d.white, op)
	}
}
