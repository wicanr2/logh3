package screen

import "github.com/wicanr2/logh3/internal/gfx"

// EngagementGauge 是規格094的原版量表輸入；Value的意義依Planet區分。
type EngagementGauge struct {
	Planet       bool
	FleetFaction int
	Value        int // 艦隊完整度，或行星防衛力
	Divisor      int // 行星caller的攻擊力
}

// Rect 回原生640×480座標，亦供差分報告使用。
func (g EngagementGauge) Rect() (EngagementRect, bool) {
	if g.FleetFaction < 0 || g.FleetFaction > 1 {
		return EngagementRect{}, false
	}
	side := g.FleetFaction
	if g.Planet {
		side = 1 - side
	}
	x := 298
	if side == 1 {
		x = 333
	}
	return EngagementRect{X: x, Y: 63, W: 11, H: 66}, true
}

// Height 保留原版word倍增及整數捨去；分母0原版直接滿格。
func (g EngagementGauge) Height() (int, bool) {
	if _, ok := g.Rect(); !ok {
		return 0, false
	}
	if !g.Planet {
		if g.Value < 0 || g.Value > 100 {
			return 0, false
		}
		return g.Value * 66 / 100, true
	}
	if g.Value < 0 || g.Value > 65535 || g.Divisor < 0 || g.Divisor > 65535 {
		return 0, false
	}
	if uint16(g.Divisor*2) <= uint16(g.Value) {
		return 66, true
	}
	return g.Value * 33 / g.Divisor, true
}

// Draw 只改量表矩形；非法資料不動畫布。
func (g EngagementGauge) Draw(fb *gfx.Framebuffer) bool {
	h, ok := g.Height()
	if !ok || fb == nil {
		return false
	}
	r, _ := g.Rect()
	fill, edge := uint8(14), uint8(3)
	if r.X == 333 {
		fill, edge = 12, 7
	}
	fb.FillRect(r.X, r.Y, r.W, r.H-h, 0)
	y := r.Y + r.H - h
	fb.FillRect(r.X, y, 1, h, edge)
	fb.FillRect(r.X+1, y, 7, h, fill)
	fb.FillRect(r.X+8, y, 3, h, edge)
	return true
}

// EngagementGaugePair 是同一幀的兩側量表；不接受船艦文字的比例。
type EngagementGaugePair struct {
	FleetFaction int
	Integrity    int
	Defence      int
	Attack       int
}

func (p EngagementGaugePair) Draw(fb *gfx.Framebuffer) bool {
	f := EngagementGauge{FleetFaction: p.FleetFaction, Value: p.Integrity}
	pl := EngagementGauge{Planet: true, FleetFaction: p.FleetFaction, Value: p.Defence, Divisor: p.Attack}
	if _, ok := f.Height(); !ok {
		return false
	}
	if _, ok := pl.Height(); !ok {
		return false
	}
	return f.Draw(fb) && pl.Draw(fb)
}
