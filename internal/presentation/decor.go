package presentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

// 保持核准PNG的原始預乘alpha轉換；不在等尺寸載入時再取樣。
func loadMapSprite(dir, name string) (*image.RGBA, error) {
	b, err := os.ReadFile(filepath.Join(dir, "MANIFEST.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m struct {
		Schema int
		Files  []struct{ File, SHA256 string }
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Schema != 1 {
		return nil, fmt.Errorf("HD 星圖素材格式無效")
	}
	for _, f := range m.Files {
		if f.File != name {
			continue
		}
		data, err := checkedFile(dir, f.File, f.SHA256)
		if err != nil {
			return nil, err
		}
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if config.Width < 256 || config.Height < 256 || config.Width > 4096 || config.Height > 4096 {
			return nil, fmt.Errorf("HD 星圖素材尺寸無效")
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		dst := image.NewRGBA(im.Bounds())
		draw.Draw(dst, dst.Rect, im, im.Bounds().Min, draw.Src)
		return dst, nil
	}
	return nil, nil
}

func alphaBox(im image.Image, rect image.Rectangle) image.Rectangle {
	box := image.Rectangle{}
	found := false
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			_, _, _, a := im.At(x, y).RGBA()
			if a < 32768 {
				continue
			}
			p := image.Rect(x, y, x+1, y+1)
			if !found {
				box = p
				found = true
			} else {
				box = box.Union(p)
			}
		}
	}
	if !found {
		return rect
	}
	return box
}
func (r *Renderer) loadMapDecor(dir string) error {
	atlas, e := loadMapSprite(dir, "planets.png")
	if e != nil {
		return e
	}
	if atlas != nil && (atlas.Rect.Dx()%2 != 0 || atlas.Rect.Dy()%2 != 0) {
		return fmt.Errorf("HD 星球圖集必須為等分2×2")
	}
	for i := range r.mapPlanets {
		if atlas == nil {
			break
		}
		w, h := atlas.Rect.Dx()/2, atlas.Rect.Dy()/2
		cell := image.Rect(i%2*w, i/2*h, (i%2+1)*w, (i/2+1)*h)
		b := alphaBox(atlas, cell)
		r.mapPlanets[i] = resize(atlas.SubImage(b), 14*Scale, 14*Scale)
	}
	for i, name := range []string{"empire", "alliance"} {
		im, e := loadMapSprite(dir, name+".png")
		if e != nil {
			return e
		}
		if im == nil {
			continue
		}
		b := alphaBox(im, im.Rect)
		size := image.Pt(38*Scale, 38*Scale)
		if i == 1 {
			size = image.Pt(28*Scale, 29*Scale)
		}
		r.mapEmblems[i] = resize(im.SubImage(b), size.X, size.Y)
	}
	return nil
}
func (r *Renderer) drawStarDecor(im *image.RGBA, c gfx.DrawCommand, p gfx.Palette) {
	old := r.frame
	r.frame = im
	defer func() { r.frame = old }()
	background := func(x, y int) {
		at := y*320 + x
		id := uint8(0)
		if (x+y)%2 == 1 {
			id = 1
		}
		if id == 0 && c.Regions[at] > 0 {
			id = c.Colours[int(c.Regions[at])-1]
		}
		r.fill(image.Rect(x*Scale, y*Scale, (x+1)*Scale, (y+1)*Scale), paletteColour(p, id))
	}
	place := func(rect image.Rectangle, src *image.RGBA) {
		if src == nil {
			return
		}
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				if x >= 0 && x < 320 && y >= 0 && y < 300 {
					background(x, y)
				}
			}
		}
		draw.Draw(im, scaledRect(rect), src, src.Rect.Min, draw.Over)
	}
	type label struct {
		name               string
		x, y, px, py, kind int
	}
	labels := []label{
		{"VALHALLA", 10, 48, 0, 0, 0}, {"FLAYA", 95, 18, 88, 22, 1}, {"KASTROP", 171, 15, 161, 21, 2}, {"GANDHARVA", 248, 16, 240, 21, 2},
		{"MAL-ADETA", 265, 34, 309, 39, 2}, {"ALTEMER", 63, 66, 57, 72, 1}, {"SHANTAU", 121, 66, 113, 72, 3}, {"POLEUIT", 181, 66, 213, 73, 2},
		{"VERMILION", 226, 73, 258, 64, 2}, {"ALDEBALAN", 20, 92, 15, 88, 4}, {"PHEZZAN", 144, 92, 180, 98, 3}, {"ALUISE", 107, 111, 97, 118, 3},
		{"TASSILI", 235, 106, 265, 112, 4}, {"VINAS", 282, 125, 307, 124, 4}, {"RANTEMARIO", 171, 127, 219, 132, 2}, {"EL-FACIL", 151, 152, 186, 158, 3},
		{"AMLIZER", 107, 164, 98, 170, 3}, {"LYGALL", 243, 158, 273, 164, 4}, {"KIFFISER", 25, 148, 17, 153, 1}, {"ISERLOHN", 155, 195, 147, 199, 1},
		{"VYAMSEED", 194, 173, 233, 177, 2}, {"ELUCERA", 280, 188, 309, 184, 2}, {"VAUANHAI", 67, 206, 59, 211, 2}, {"ASTATE", 171, 213, 200, 213, 2},
		{"TRIPLA", 245, 208, 274, 215, 4}, {"JOTUNHEIM", 109, 225, 99, 231, 3}, {"KERIM", 204, 233, 230, 239, 3}, {"DOUELUG", 26, 243, 17, 249, 3},
		{"TIAMAT", 143, 262, 174, 267, 3}, {"HOLSIN", 80, 273, 70, 279, 4}, {"BALLAT", 287, 257, 0, 0, 0}, {"DORIA", 204, 280, 230, 286, 3},
	}
	for _, l := range labels {
		if l.kind < 2 {
			continue
		}
		at := 0
		if l.kind == 2 {
			at = 1
		}
		if l.kind == 4 {
			at = 3
		}
		place(image.Rect(l.px-7, l.py-7, l.px+7, l.py+7), r.mapPlanets[at])
	}
	place(image.Rect(5, 3, 43, 41), r.mapEmblems[0])
	place(image.Rect(281, 270, 309, 299), r.mapEmblems[1])
	// Preserve every original name pixel and its exact baseline.
	for _, l := range labels {
		for y := l.y; y < l.y+6; y++ {
			for x := l.x; x < l.x+len(l.name)*4; x++ {
				if x < 0 || x >= 320 || y < 0 || y >= 300 {
					continue
				}
				at := y*320 + x
				id := c.Base[at] & 15
				if id == 0 && c.Regions[at] > 0 {
					id = c.Colours[int(c.Regions[at])-1]
				}
				r.fill(image.Rect(x*Scale, y*Scale, (x+1)*Scale, (y+1)*Scale), paletteColour(p, id))
			}
		}
	}
}
