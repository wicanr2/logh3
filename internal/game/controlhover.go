package game

import (
	"image"

	"github.com/wicanr2/logh3/internal/screen"
)

// 使用正式熱區與計劃控制常數，不能從資訊格的外觀猜測可操作性。
func (g *Game) hoverControlBounds() image.Rectangle {
	x, y := g.mouseX, g.mouseY
	point := image.Pt(x, y)
	if g.state == StateCommandPhase {
		for _, bar := range commandBar {
			r := image.Rect(bar.X, barRow, bar.X+bar.W, barRow+rowHeight)
			if point.In(r) && g.commandSpec(bar.Menu) != nil {
				return r
			}
		}
		if spec := g.commandSpec(StateCommandPhase); spec != nil {
			for _, w := range spec.Windows {
				if w.Tabs == nil {
					continue
				}
				for _, row := range w.Tabs.Rows {
					for _, tab := range row.Items {
						r := image.Rect(tab.X, row.Y, tab.X+tab.W, row.Y+w.Tabs.H)
						if !tab.NoBox && point.In(r) {
							return r
						}
					}
				}
			}
		}
	}
	if i := menuOf(g.state); i >= 0 {
		bar := commandBar[i]
		for row, item := range bar.Items {
			r := image.Rect(bar.X, barRow+row*rowHeight, bar.X+bar.ItemW, barRow+(row+1)*rowHeight)
			if point.In(r) && (item == StateExecute || g.commandSpec(item) != nil) {
				return r
			}
		}
		return image.Rectangle{}
	}
	if g.state == StateCopyProtection && g.res.CopyProt != nil && g.res.CopyProt.InputBox != nil {
		b := g.res.CopyProt.InputBox
		r := image.Rect(b.X, b.Y, b.X+b.W, b.Y+b.H)
		if point.In(r) {
			return r
		}
		return image.Rectangle{}
	}
	if g.state == StateTaxPlan || g.state == StateBudgetPlan || g.state == StateProductionPlan {
		boxes := []planBox{planProdEnd}
		if g.state != StateProductionPlan {
			boxes = []planBox{planTaxUp, planTaxDown, planPctUp, planPctDown, planStaff, planEnd}
		}
		for _, b := range boxes {
			if b.hit(x, y) {
				return image.Rect(b.X0, b.Y0, b.X1+1, b.Y1+1)
			}
		}
		return image.Rectangle{}
	}
	if g.state == StateSettings && point.In(enemyVisibilityButton) {
		return enemyVisibilityButton
	}
	var spec *screen.MainGameSpec
	dy := 0
	switch g.state {
	case StateMainGame:
		spec, dy = g.res.MainGame, mouseDY
	case StateShipOrder:
		spec, dy = g.res.ShipOrder, mouseDY
	case StateBuildPlan:
		spec, dy = g.res.BuildPlan, mouseDY
	default:
		spec = g.commandSpec(g.state)
	}
	if spec == nil {
		return image.Rectangle{}
	}
	for _, h := range spec.Hotspots {
		r := image.Rect(h.X, h.Y+dy, h.X+h.W, h.Y+h.H+dy)
		if !h.Disabled && point.In(r) {
			return r
		}
	}
	// 雙人報告及交棒的正式輸入使用overlay.BoxAt。
	if g.state == StateVersusAccount || g.state == StateVersusShipOrder || g.state == StateVersusInput {
		if w := spec.Windows["overlay"]; w != nil {
			if i := w.BoxAt(x, y); i >= 0 {
				b := w.Boxes[i]
				return image.Rect(b.X, b.Y, b.X+b.W, b.Y+b.H)
			}
		}
	}
	return image.Rectangle{}
}
