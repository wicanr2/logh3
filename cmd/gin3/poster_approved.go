package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"time"

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
	files         map[string]*posterPNG
	bases         [3]image.Image
	hulls         image.Image
	labels        map[string]image.Image
	baseTextures  [3]*ebiten.Image
	hullTexture   *ebiten.Image
	labelTextures map[string]*ebiten.Image
	overlay       *ebiten.Image
	rotation      *posterQuoteRotation
	quoteError    error
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
	p := &approvedPosterArt{labels: map[string]image.Image{}, labelTextures: map[string]*ebiten.Image{}, files: map[string]*posterPNG{}}
	names := []string{"base-v1.png", "base-v2.png", "base-v3.png", "hulls.png"}
	for _, locale := range []string{"zh-TW", "zh-CN", "ja", "en", "ko"} {
		p.labels[locale] = nil
		names = append(names, "labels-"+locale+".png")
	}
	for _, name := range names {
		item, ok := manifest.Files[name]
		if !ok || item.Width != posterTotalW || item.Height != posterHeight {
			return nil, fmt.Errorf("定案海報缺少或尺寸錯誤：%s", name)
		}
		p.files[name], err = registerPosterPNG(filepath.Join(dir, name), item.SHA, item.Width, item.Height, 32<<20)
		if err != nil {
			return nil, err
		}
	}
	p.rotation, err = loadPosterQuoteRotation(dir)
	if err != nil && !os.IsNotExist(err) {
		p.quoteError = err
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
		im := art.bases[i]
		if im == nil {
			im = art.files[fmt.Sprintf("base-v%d.png", i+1)].image()
		}
		if im != nil {
			for n, texture := range art.baseTextures {
				if n != i && texture != nil {
					texture.Dispose()
					art.baseTextures[n] = nil
				}
			}
			art.baseTextures[i] = ebiten.NewImageFromImage(im)
			art.bases[i] = nil
		}
	}
	if art.hullTexture == nil {
		im := art.hulls
		if im == nil {
			im = art.files["hulls.png"].image()
		}
		if im != nil {
			art.hullTexture = ebiten.NewImageFromImage(im)
			art.hulls = nil
		}
	}
	base := art.baseTextures[i]
	if base != nil {
		screen.DrawImage(base, nil)
	}
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
			if base != nil {
				target.DrawImage(base, nil)
			}
			if art.hullTexture != nil {
				target.DrawImage(art.hullTexture, nil)
			}
		}
	}
	if i == 0 {
		locale := "zh-TW"
		if p.catalog != nil {
			locale = p.catalog.Locale
		}
		art.drawQuote(screen, locale, time.Now())
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
