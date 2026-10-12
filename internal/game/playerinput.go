package game

import "github.com/wicanr2/logh3/internal/gfx"

func (g *Game) monthlyAccountState() State {
	if g.versus {
		return StateVersusAccount
	}
	return StateMainGame
}

func (g *Game) monthlyShipOrderState() State {
	if g.versus {
		return StateVersusShipOrder
	}
	return StateShipOrder
}

func (g *Game) openMonthlyBuildPlan() {
	if g.res.BuildPlan != nil {
		g.state = StateBuildPlan
		p := g.thisPlan()
		g.plan = &p
	}
}

// 這兩張沒有游標裁切或收據補丁；動態數字後補游標（規格121）。
func (g *Game) drawMonthlyCursor(fb *gfx.Framebuffer) {
	if g.res.Cursor != nil && g.mouseY >= 0 {
		g.res.Cursor.Draw(fb, g.mouseX, g.mouseY-mouseDY)
	}
}
