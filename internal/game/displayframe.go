package game

import "github.com/wicanr2/logh3/internal/gfx"

// FrameWithoutCursor供app疊加設定後只畫一次最上層游標；不改正式Frame。
func (g *Game) FrameWithoutCursor() *gfx.Framebuffer {
	cursor := g.res.Cursor
	g.res.Cursor = nil
	defer func() { g.res.Cursor = cursor }()
	return g.Frame()
}
