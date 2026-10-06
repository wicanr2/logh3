package game

import "github.com/wicanr2/logh3/internal/gfx"

// 這兩張沒有游標裁切或收據補丁；動態數字後補游標（規格121）。
func (g *Game) drawMonthlyCursor(fb *gfx.Framebuffer) {
	if g.res.Cursor != nil && g.mouseY >= 0 {
		g.res.Cursor.Draw(fb, g.mouseX, g.mouseY-mouseDY)
	}
}
