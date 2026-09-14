package game

import (
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// VictoryReport 是 `0374:0750` 選出來、交給 `04C8:00F3` 的單頁報告。
type VictoryReport struct {
	Viewer int
	Winner int
	Won    bool
}

func (g *Game) startVictoryReport() bool {
	if g.res == nil || g.res.VictoryReport == nil || !g.regionsReady() {
		return false
	}
	winner := -1
	if r := g.regionRecord(rules.CapitalRegion(0)); r != nil &&
		rules.RegionOwnerOf(r) == 1 {
		winner = 1
	} else if r := g.regionRecord(rules.CapitalRegion(1)); r != nil &&
		rules.RegionOwnerOf(r) == 0 {
		winner = 0
	}
	if winner < 0 {
		return false
	}
	viewer := g.faction
	if g.versus {
		viewer = winner
	}
	g.victoryReport = VictoryReport{Viewer: viewer, Winner: winner, Won: viewer == winner}
	g.state = StateVictoryReport
	return true
}

func (g *Game) dismissVictoryReport() {
	r := g.victoryReport
	if r.Winner < 0 || r.Winner > 1 {
		return
	}
	g.ShowEnding(g.res.Endings[r.Winner])
}

func (g *Game) CurrentVictoryReport() (VictoryReport, bool) {
	return g.victoryReport, g.state == StateVictoryReport
}

func (g *Game) drawVictoryReport() *gfx.Framebuffer {
	fb := gfx.NewFramebuffer(0)
	if g.res == nil || g.res.VictoryReport == nil {
		return fb
	}
	r := g.victoryReport
	faceIndex := -1
	if work := g.workRecord(); len(work) > 0x37+r.Viewer*4 {
		faceIndex = int(work[0x37+r.Viewer*4])
	}
	var face *asset.Image
	if g.res.Faces != nil {
		face, _ = g.res.Faces.Decode(faceIndex)
	}
	year, month, day := g.Date()
	regions := 0
	for i := 0; i < rules.RegionCount; i++ {
		if rec := g.regionRecord(i); rec != nil && rules.RegionOwnerOf(rec) == r.Viewer {
			regions++
		}
	}
	g.res.VictoryReport.Draw(fb, g.res.Font, g.res.FrameTiles, face,
		r.Viewer, r.Won, year, month, day, regions, g.campaignScore)
	return fb
}
