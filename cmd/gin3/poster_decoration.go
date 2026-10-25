package main

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
)

type posterDecoration struct {
	fighters        [2]image.Image
	ships           [2][]*image.RGBA
	fighterTextures [2]*ebiten.Image
	shipTextures    [2][]*ebiten.Image
	white           *ebiten.Image
	bands           [2]image.Rectangle
	stamp           uint64
	width           int
	phase, page     int
	epoch           time.Time
}

type posterDecorationTrace struct {
	Approved  bool       `json:"approved"`
	Viewport  [4]float64 `json:"viewport"`
	Bands     [2][4]int  `json:"bands"`
	Flagships [2]int     `json:"flagships"`
	Fighters  int        `json:"fighters"`
	Phase     int        `json:"phase"`
	Page      int        `json:"page"`
	Animated  bool       `json:"animated"`
}

func (a *app) posterDecorationObservation() *posterDecorationTrace {
	d := a.posterDecor
	if !a.posterMode || d == nil {
		return nil
	}
	o := &posterDecorationTrace{Flagships: [2]int{len(d.ships[0]), len(d.ships[1])}, Phase: d.phase, Page: d.page, Animated: a.g.DisplaySettings().Anim != 0}
	v := a.posterViewport()
	o.Viewport = [4]float64{v.X, v.Y, v.W, v.H}
	if a.usingApprovedPoster() {
		o.Approved, o.Flagships = true, [2]int{1, 1}
	}
	scale := 1
	if a.hd != nil {
		scale = presentation.Scale
	}
	for i, b := range d.bands {
		o.Bands[i] = [4]int{b.Min.X / scale, b.Min.Y / scale, b.Max.X / scale, b.Max.Y / scale}
	}
	for _, im := range d.fighters {
		if im != nil {
			o.Fighters++
		}
	}
	return o
}

func posterFlagshipKeys(codes []byte) [2][]string {
	var result [2][]string
	for side := 0; side < 2; side++ {
		seen := map[byte]bool{}
		order := []int{}
		if side == 1 {
			order = append(order, 12)
		} // 楊威利的現有旗艦碼，先展示休伯利安。
		for i := 0; i < 16; i++ {
			order = append(order, i)
		}
		for _, i := range order {
			index := side*40 + i
			if index >= len(codes) || seen[codes[index]] {
				continue
			}
			seen[codes[index]] = true
			result[side] = append(result[side], fmt.Sprintf("%02X%02X", 'E'+side, codes[index]))
		}
	}
	return result
}

func posterFighterRect(src image.Image, side int) image.Rectangle {
	b := src.Bounds()
	cell := image.Rect(b.Min.X+side*b.Dx()/2, b.Min.Y, b.Min.X+(side+1)*b.Dx()/2, b.Max.Y)
	box := image.Rectangle{}
	for y := cell.Min.Y; y < cell.Max.Y; y++ {
		for x := cell.Min.X; x < cell.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			if a >= 192*257 {
				p := image.Rect(x, y, x+1, y+1)
				if box.Empty() {
					box = p
				} else {
					box = box.Union(p)
				}
			}
		}
	}
	if box.Empty() {
		return cell
	}
	return box.Inset(-8).Intersect(cell)
}

func loadPosterDecoration(art string, renderer *presentation.Renderer, flagshipCodes string) (*posterDecoration, error) {
	d := &posterDecoration{}
	codes, err := hex.DecodeString(flagshipCodes)
	if err != nil {
		return nil, err
	}
	for side, keys := range posterFlagshipKeys(codes) {
		for _, key := range keys {
			if im := renderer.PosterShip(key); im != nil {
				d.ships[side] = append(d.ships[side], im)
			}
		}
	}
	path := filepath.Join(art, "poster", "fighters.png")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return d, err
	}
	if info.Size() > 16<<20 {
		return d, fmt.Errorf("戰機圖過大")
	}
	c, err := png.DecodeConfig(f)
	if err != nil {
		return d, err
	}
	if c.Width < 2 || c.Width%2 != 0 || c.Width > 4096 || c.Height <= 0 || c.Height > 4096 {
		return d, fmt.Errorf("戰機圖尺寸無效")
	}
	if _, err = f.Seek(0, 0); err != nil {
		return d, err
	}
	im, err := png.Decode(f)
	if err != nil {
		return d, err
	}
	sub := im.(interface {
		SubImage(image.Rectangle) image.Image
	})
	for side := range d.fighters {
		d.fighters[side] = sub.SubImage(posterFighterRect(im, side))
	}
	return d, nil
}

// 對實際無游標RGB背景找完整黑列，不從狀態名稱猜測空白。
func posterBlackBands(pixels []byte, stride, width, height int) [2]image.Rectangle {
	var result [2]image.Rectangle
	if width <= 0 || height <= 0 || stride < width*4 || len(pixels) < stride*height {
		return result
	}
	black := func(y int) bool {
		row := pixels[y*stride : y*stride+width*4]
		for x := 0; x < len(row); x += 4 {
			if row[x] != 0 || row[x+1] != 0 || row[x+2] != 0 {
				return false
			}
		}
		return true
	}
	top := 0
	for top < height && black(top) {
		top++
	}
	if top == height {
		return result
	}
	bottom := height
	for bottom > top && black(bottom-1) {
		bottom--
	}
	scale := width / gfx.Width
	if scale < 1 {
		scale = 1
	}
	if top >= 8*scale {
		result[0] = image.Rect(0, 0, width, min(top, 48*scale))
	}
	if height-bottom >= 20*scale {
		result[1] = image.Rect(0, max(bottom, height-80*scale), width, height)
	}
	return result
}

func posterDecorationPhase(elapsed time.Duration, animated bool) (phase, page int) {
	if !animated {
		return 0, 0
	}
	frames := int(elapsed/time.Second)*15 + int((elapsed%time.Second)*15/time.Second)
	return frames % 180, frames / 180
}

func posterDecorationIndex(page, slot, count int) int {
	if count == 0 {
		return -1
	}
	n := (page*8 + slot) % count
	if n < 0 {
		n += count
	}
	return n
}

func (d *posterDecoration) textures() {
	if d.white != nil {
		return
	}
	d.white = ebiten.NewImage(1, 1)
	d.white.Fill(color.White)
	for side, im := range d.fighters {
		if im == nil {
			continue
		}
		source := ebiten.NewImageFromImage(im)
		d.fighterTextures[side] = ebiten.NewImage(128, 96)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Scale(128/float64(source.Bounds().Dx()), 96/float64(source.Bounds().Dy()))
		d.fighterTextures[side].DrawImage(source, op)
		source.Dispose()
	}
	for side, ships := range d.ships {
		for _, im := range ships {
			d.shipTextures[side] = append(d.shipTextures[side], ebiten.NewImageFromImage(im))
		}
	}
}

func (a *app) drawPosterDecoration(screen *ebiten.Image, fb *gfx.Framebuffer, pixels []byte, stride, width, height int, stamp uint64) {
	d := a.posterDecor
	if d == nil {
		return
	}
	if !a.posterMode || fb.InterfaceTheme == "" {
		d.bands = [2]image.Rectangle{}
		d.stamp = 0
		return
	}
	if stamp == 0 || d.stamp != stamp || d.width != width {
		d.bands = posterBlackBands(pixels, stride, width, height)
		factor := screen.Bounds().Dx() / width
		for i, b := range d.bands {
			d.bands[i] = image.Rect(b.Min.X*factor, b.Min.Y*factor, b.Max.X*factor, b.Max.Y*factor)
		}
		d.stamp, d.width = stamp, width
	}
	if d.epoch.IsZero() {
		d.epoch = time.Now()
	}
	d.phase, d.page = posterDecorationPhase(time.Since(d.epoch), a.g.DisplaySettings().Anim != 0)
	if a.usingApprovedPoster() {
		return // New artwork is composited outside the game viewport.
	}
	d.textures()
	for band, rect := range d.bands {
		if rect.Empty() {
			continue
		}
		target := screen.SubImage(rect).(*ebiten.Image)
		d.drawBand(target, rect, band, float64(screen.Bounds().Dx())/gfx.Width)
	}
}

func (d *posterDecoration) drawBand(target *ebiten.Image, b image.Rectangle, band int, scale float64) {
	w, h := float64(b.Dx())/scale, float64(b.Dy())/scale
	quad := func(tex *ebiten.Image, x, y, width, height, alpha float64, flip bool, tint color.NRGBA) {
		if tex == nil || alpha <= 0 {
			return
		}
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		sx := width * scale / float64(tex.Bounds().Dx())
		if flip {
			sx = -sx
			x += width
		}
		op.GeoM.Scale(sx, height*scale/float64(tex.Bounds().Dy()))
		op.GeoM.Translate(float64(b.Min.X)+x*scale, float64(b.Min.Y)+y*scale)
		op.ColorScale.ScaleWithColor(tint)
		op.ColorScale.ScaleAlpha(float32(alpha))
		target.DrawImage(tex, op)
	}
	fill := func(x, y, width, height float64, c color.NRGBA) { quad(d.white, x, y, width, height, 1, false, c) }
	line := func(x, y, x2, y2, thickness float64, c color.NRGBA) {
		op := &ebiten.DrawImageOptions{}
		dx, dy := x2-x, y2-y
		op.GeoM.Scale(math.Hypot(dx, dy)*scale, thickness*scale)
		op.GeoM.Rotate(math.Atan2(dy, dx))
		op.GeoM.Translate(float64(b.Min.X)+x*scale, float64(b.Min.Y)+y*scale)
		op.ColorScale.ScaleWithColor(c)
		target.DrawImage(d.white, op)
	}
	fill(0, 0, w, h, color.NRGBA{2, 6, 13, 255})
	for i := 0; i < 28; i++ {
		x := float64((i*149+37)%640) / 640 * w
		y := float64((i*47+19)%83) / 83 * h
		fill(x, y, .7, .7, color.NRGBA{105, 144, 178, uint8(75 + i%3*25)})
	}
	if band == 0 {
		fill(0, h-.5, w, .5, color.NRGBA{144, 122, 78, 180})
	} else {
		fill(0, 0, w, .5, color.NRGBA{144, 122, 78, 180})
	}
	t := float64(d.phase) / 15
	white := color.NRGBA{255, 255, 255, 255}
	if band == 0 {
		for side := 0; side < 2; side++ {
			for i := 0; i < 4; i++ {
				f := math.Mod(t/12+float64(i)/4+float64(side)*.125, 1)
				x := -44 + f*(w+88)
				if side == 1 {
					x = w - 44 - f*(w+88)
				}
				y := h/2 - 13 + math.Sin(t*.9+float64(i)*1.7+float64(side))*min(5., max(0., h/2-14))
				quad(d.fighterTextures[side], x, y, 38, 27, 1, side == 1, white)
				if d.fighterTextures[side] != nil && (d.phase+i*11+side*17)%60 < 8 {
					start, end := x+36, x+89
					if side == 1 {
						start, end = x+2, x-51
					}
					c := color.NRGBA{105, 184, 255, 170}
					if side == 1 {
						c = color.NRGBA{255, 177, 99, 170}
					}
					line(start, y+13, end, y+13+math.Sin(float64(i))*3, .7, c)
				}
			}
		}
		return
	}
	for side, ships := range d.shipTextures {
		if len(ships) == 0 {
			continue
		}
		for slot := 0; slot < 8; slot++ {
			col, row := slot%4, slot/4
			x := 8 + float64(col)*75
			if side == 1 {
				x = w - 76 - float64(col)*75
			}
			y := h*.03 + float64(row)*h*.49 + math.Sin(t*.45+float64(slot))*min(1., h*.02)
			width, height := 68., min(27., h*.43)
			alpha := 1.
			if d.page > 0 && d.phase < 12 {
				alpha = float64(d.phase) / 12
				quad(ships[posterDecorationIndex(d.page-1, slot, len(ships))], x, y, width, height, 1-alpha, false, white)
			}
			quad(ships[posterDecorationIndex(d.page, slot, len(ships))], x, y, width, height, alpha, false, white)
		}
	}
	if d.phase%90 < 12 {
		for row := 0; row < 2; row++ {
			y := h * (.24 + float64(row)*.48)
			line(w/2-26, y, w/2+22, y-.8, .6, color.NRGBA{125, 193, 255, 155})
			line(w/2+25, y+2, w/2-20, y+1, .6, color.NRGBA{255, 178, 103, 155})
		}
	}
}
