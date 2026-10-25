package game

import (
	"image"

	"github.com/wicanr2/logh3/internal/gfx"
)

// 人事視窗沒有世界動畫。點擊、鍵盤及捲動使背景失效；移動游標只複製
// 背景並在最上層畫游標，避免重建整張視窗阻塞短點擊。
func (g *Game) personnelFrame() *gfx.Framebuffer {
	if g.personnelBase == nil {
		cursor := g.res.Cursor
		g.res.Cursor = nil
		g.personnelBuilding = true
		func() {
			defer func() { g.res.Cursor = cursor; g.personnelBuilding = false }()
			g.personnelBase = g.Frame()
		}()
		g.personnelBase.HoverControl = image.Rectangle{}
	}
	base := g.personnelBase
	fb := *base
	fb.Pix = append([]uint8(nil), base.Pix...)
	fb.Commands = append([]gfx.DrawCommand(nil), base.Commands...)
	fb.Artwork = append([]gfx.Artwork(nil), base.Artwork...)
	for i := range fb.Artwork {
		fb.Artwork[i].Visible = append([]uint8(nil), base.Artwork[i].Visible...)
	}
	fb.HoverControl = g.hoverControlBounds()
	if fb.HoverControl.Empty() {
		fb.HoverControl = gfx.VisibleButtonAt(fb.Commands, image.Pt(g.mouseX, g.mouseY))
	}
	fb.CursorBackground = base
	if g.res.Cursor != nil && g.mouseY >= 0 {
		g.res.Cursor.Draw(&fb, g.mouseX, g.mouseY-mouseDY)
	}
	return &fb
}
