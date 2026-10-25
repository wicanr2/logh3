package gfx

import "image"

// VisibleButtonAt只讀既有繪製角色，後畫視窗／圖像遮住的按鈕不算可見。
func VisibleButtonAt(commands []DrawCommand, point image.Point) image.Rectangle {
	var box image.Rectangle
	var visit func([]DrawCommand)
	visit = func(cs []DrawCommand) {
		for _, c := range cs {
			if c.Kind == DrawCursor {
				continue
			}
			v := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)
			if c.Kind == DrawPanel && (c.Key == "ui/button" || c.Key == "ui/primary") && point.In(v) {
				box = v
				continue
			}
			switch c.Kind {
			case DrawFill, DrawSurface, DrawPanel, DrawImage, DrawTerrain, DrawMinimap:
				if !box.Empty() && point.In(v) && !v.In(box) {
					box = image.Rectangle{}
				}
			}
			visit(c.Fallback)
		}
	}
	visit(commands)
	return box
}
