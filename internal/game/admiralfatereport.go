package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// AdmiralFateReport 保存一則失蹤提督生死報告的事件時狀態。
type AdmiralFateReport struct {
	Admiral int
	Faction int
	Dead    bool
}

// resolveAdmiralFates 執行原版 `056A:0219`。單人只顯示玩家陣營；
// 對戰兩邊都是人類座位。狀態轉移本身仍會套到所有陣營。
func (g *Game) resolveAdmiralFates() {
	g.admiralFateReports = nil
	if g.res == nil || len(g.res.Admiral) < rules.AdmiralRecord {
		return
	}
	human := [2]bool{}
	if g.versus {
		human = [2]bool{true, true}
	} else if g.faction >= 0 && g.faction < len(human) {
		human[g.faction] = true
	}
	for _, fate := range rules.ResolveMissingAdmirals(g.res.Admiral, human, g.rand94DC()) {
		if fate.Faction < 0 || fate.Faction >= len(human) || !human[fate.Faction] {
			continue
		}
		g.admiralFateReports = append(g.admiralFateReports, AdmiralFateReport(fate))
	}
}

func (g *Game) startAdmiralFateReports(after func()) bool {
	if len(g.admiralFateReports) == 0 || g.res == nil || g.res.FlipReport == nil ||
		len(g.res.AdmiralNames) == 0 {
		return false
	}
	g.pendingAdmiralFateReports = append(g.pendingAdmiralFateReports[:0], g.admiralFateReports...)
	g.afterAdmiralFateReports = after
	g.state = StateAdmiralFateReport
	return true
}

func (g *Game) CurrentAdmiralFateReport() (AdmiralFateReport, bool) {
	if len(g.pendingAdmiralFateReports) == 0 {
		return AdmiralFateReport{}, false
	}
	return g.pendingAdmiralFateReports[0], true
}

func (g *Game) dismissAdmiralFateReport() {
	if len(g.pendingAdmiralFateReports) > 0 {
		g.pendingAdmiralFateReports = g.pendingAdmiralFateReports[1:]
	}
	if len(g.pendingAdmiralFateReports) > 0 {
		return
	}
	after := g.afterAdmiralFateReports
	g.afterAdmiralFateReports = nil
	if after != nil {
		after()
		return
	}
	g.state = StateCommandPhase
}

func (g *Game) drawAdmiralFateReport(fb *gfx.Framebuffer) {
	report, ok := g.CurrentAdmiralFateReport()
	if !ok || g.res == nil || g.res.FlipReport == nil ||
		report.Admiral < 0 || report.Admiral >= len(g.res.AdmiralNames) {
		return
	}
	g.res.FlipReport.DrawAdmiralFate(fb, g.res.Font,
		g.res.AdmiralNames[report.Admiral], report.Dead)
	g.drawAdmiralReportFace(fb, report.Admiral)
}

// drawAdmiralReportFace 與 `3826:0356(-1, admiral)` 相同：直接用事件中的
// 提督索引，而不是 `ds:8F4D` 的情報參謀索引。
func (g *Game) drawAdmiralReportFace(fb *gfx.Framebuffer, admiral int) {
	p := g.res.FlipReport
	if p == nil || g.res.Faces == nil {
		return
	}
	img, err := g.res.Faces.Decode(admiral)
	if err != nil {
		return
	}
	img.Draw(fb, p.Face.X, p.Face.Y, p.Face.W, p.Face.H)
}
