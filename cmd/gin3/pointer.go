package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
	"image"
	"image/color"
)

func (a *app) drawNativeHover(screen *ebiten.Image, bounds image.Rectangle) {
	b := bounds.Intersect(image.Rect(0, 0, 640, 480)).Inset(1)
	if b.Empty() {
		return
	}
	if a.pointerFill == nil {
		a.pointerFill = ebiten.NewImage(1, 1)
		a.pointerFill.Fill(color.White)
	}
	scale := 1
	if a.hd != nil {
		scale = presentation.Scale
	}
	fill := func(r image.Rectangle, alpha uint8) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(r.Dx()*scale), float64(r.Dy()*scale))
		op.GeoM.Translate(float64(r.Min.X*scale), float64(r.Min.Y*scale))
		op.ColorScale.ScaleWithColor(color.NRGBA{255, 235, 181, alpha})
		screen.DrawImage(a.pointerFill, op)
	}
	fill(b, 52)
	fill(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+1), 220)
}

// 背景材質不含游標。相同形狀只上傳一次，位置逐次用GPU平移。
func (a *app) drawPointer(screen *ebiten.Image, command *gfx.DrawCommand) {
	if command == nil || a.cursor == nil {
		return
	}
	c := *command
	clip := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H).Intersect(c.Clip).Intersect(image.Rect(0, 0, gfx.Width, gfx.Height))
	if clip.Empty() {
		return
	}
	scale, spriteScale, which := 1, 1, 0
	if a.hd != nil {
		scale = presentation.Scale
	}
	if a.hd != nil && !a.original && c.W == 13 && c.H == 16 {
		which, spriteScale = 1, presentation.Scale
	}
	if which == 1 && a.cursorTextures[1] == nil {
		a.cursorTextures[1] = ebiten.NewImageFromImage(presentation.CursorSprite())
	}
	if which == 0 && (a.cursorTextures[0] == nil || a.cursorPalette != a.g.Palette()) {
		im := image.NewRGBA(image.Rect(0, 0, a.cursor.Width, a.cursor.Height))
		pal := a.g.Palette()
		for y, row := range a.cursor.Rows {
			for x, ch := range row {
				if index, ok := a.cursor.Colors[string(ch)]; ok {
					r, g, b := pal.RGB8(int(index))
					o := y*im.Stride + x*4
					im.Pix[o], im.Pix[o+1], im.Pix[o+2], im.Pix[o+3] = r, g, b, 255
				}
			}
		}
		a.cursorTextures[0] = ebiten.NewImageFromImage(im)
		a.cursorPalette = pal
	}
	rect := image.Rect((clip.Min.X-c.X)*spriteScale, (clip.Min.Y-c.Y)*spriteScale, (clip.Max.X-c.X)*spriteScale, (clip.Max.Y-c.Y)*spriteScale)
	texture := a.cursorTextures[which].SubImage(rect).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale)/float64(spriteScale), float64(scale)/float64(spriteScale))
	op.GeoM.Translate(float64(clip.Min.X*scale), float64(clip.Min.Y*scale))
	screen.DrawImage(texture, op)
}
