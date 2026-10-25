package presentation

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
)

func (r *Renderer) hoverControl(fb *gfx.Framebuffer, pal gfx.Palette) {
	b := fb.HoverControl.Intersect(image.Rect(0, 0, gfx.Width, gfx.Height))
	if b.Empty() {
		return
	}
	// 只在實際繪出的按鈕上反光，熱區清單內的地圖及資訊列不增加亮框。
	var box image.Rectangle
	var cursors []gfx.DrawCommand
	var visit func([]gfx.DrawCommand)
	visit = func(commands []gfx.DrawCommand) {
		for _, c := range commands {
			if c.Kind == gfx.DrawCursor {
				cursors = append(cursors, c)
				continue
			}
			if c.Kind == gfx.DrawPanel {
				v := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)
				button := c.Key == "ui/button" || c.Key == "ui/primary" || c.Key == "" && v.Min.Y == b.Min.Y && v.Max.Y == b.Max.Y
				if button && b.In(v) {
					box = v
					continue
				}
			}
			// 後畫的不透明面板或圖像遮住控制時，不把反光畫回上層。
			switch c.Kind {
			case gfx.DrawFill, gfx.DrawSurface:
				v := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)
				// 分頁鈕會在框畫好後補文字底色；框內的補畫仍屬同一顆按鈕。
				if !box.Empty() && !v.In(box) && !box.Intersect(v).Empty() {
					box = image.Rectangle{}
				}
			case gfx.DrawPanel, gfx.DrawImage, gfx.DrawTerrain, gfx.DrawMinimap:
				if !box.Empty() && !box.Intersect(image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)).Empty() {
					box = image.Rectangle{}
				}
			}
			visit(c.Fallback)
		}
	}
	visit(fb.Commands)
	if fb.HoverMenu {
		box = b
	}
	if box.Empty() {
		return
	}
	x0, y0, x1, y1 := box.Min.X*Scale+2, box.Min.Y*Scale+2, box.Max.X*Scale-2, box.Max.Y*Scale-2
	for y := max(0, y0); y < min(r.frame.Bounds().Dy(), y1); y++ {
		for x := max(0, x0); x < min(r.frame.Bounds().Dx(), x1); x++ {
			edge := x < x0+2 || x >= x1-2 || y < y0+2 || y >= y1-2
			i := (y/Scale)*gfx.Width + x/Scale
			if !edge && (len(r.uiMask) == 0 || r.uiMask[i]&themeGlyph != 0) {
				continue
			}
			alpha := uint32(52)
			if y < y0+(y1-y0)/3 {
				alpha = 84
			}
			if edge {
				alpha = 220
			}
			o := y*r.frame.Stride + x*4
			for k, v := range [3]uint32{255, 235, 181} {
				r.frame.Pix[o+k] = uint8((uint32(r.frame.Pix[o+k])*(255-alpha) + v*alpha + 127) / 255)
			}
		}
	}
	// 游標最後重播，保持透明部分並避免亮框改變指標顏色。
	for _, c := range cursors {
		if !r.cursor(c) {
			r.replay(c.Fallback, pal)
		}
	}
}
