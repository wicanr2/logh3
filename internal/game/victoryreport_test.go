package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

func TestVictoryReportLifecycle(t *testing.T) {
	for _, c := range []struct {
		name            string
		viewer, winner  int
		versus, wantWon bool
	}{
		{name: "human_wins", viewer: 0, winner: 0, wantWon: true},
		{name: "human_loses", viewer: 0, winner: 1},
		{name: "versus_winner_views", viewer: 1, winner: 1, versus: true, wantWon: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newRegionTurnGame(t)
			g.res.VictoryReport = &screen.VictoryReportPanel{}
			g.res.Endings[c.winner] = &asset.Image{W: 1, H: 1, Pix: []uint8{7}}
			g.faction, g.versus = 0, c.versus
			g.regionRecord(rules.CapitalRegion(0))[0x12] = 0
			g.regionRecord(rules.CapitalRegion(1))[0x12] = 1
			if c.winner == 0 {
				g.regionRecord(rules.CapitalRegion(1))[0x12] = 0
			} else {
				g.regionRecord(rules.CapitalRegion(0))[0x12] = 1
			}
			if !g.startVictoryReport() || g.state != StateVictoryReport {
				t.Fatalf("沒有進勝負報告：%v", g.state)
			}
			r, ok := g.CurrentVictoryReport()
			if !ok || r.Viewer != c.viewer || r.Winner != c.winner || r.Won != c.wantWon {
				t.Fatalf("報告=%+v,%v", r, ok)
			}
			g.Click(300, 300)
			if g.state != StateEnding || g.ending != g.res.Endings[c.winner] {
				t.Fatalf("關閉後狀態=%v、結局=%p", g.state, g.ending)
			}
		})
	}
}

func TestNoVictoryReportWhileBothCapitalsRemain(t *testing.T) {
	g := newRegionTurnGame(t)
	g.res.VictoryReport = &screen.VictoryReportPanel{}
	g.regionRecord(rules.CapitalRegion(0))[0x12] = 0
	g.regionRecord(rules.CapitalRegion(1))[0x12] = 1
	if g.startVictoryReport() {
		t.Fatal("兩座首都仍在時不應開勝負報告")
	}
}
