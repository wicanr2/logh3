package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

func TestResolveAdmiralFatesQueuesOnlyHumanReports(t *testing.T) {
	admirals := make([]byte, 82*rules.AdmiralRecord)
	admirals[0] = 0x80
	admirals[1] = 0
	names := make([][]uint16, 82)
	g := &Game{
		res:          &Resources{Admiral: admirals, AdmiralNames: names},
		faction:      0,
		strategyRand: rules.NewRand(0),
	}
	g.resolveAdmiralFates()
	if len(g.admiralFateReports) != 1 {
		t.Fatalf("排了 %d 則，預期 1", len(g.admiralFateReports))
	}
	r := g.admiralFateReports[0]
	if r.Admiral != 0 || r.Faction != 0 || r.Dead {
		t.Fatalf("報告 %+v，預期帝國提督 0 生還", r)
	}
}

func TestAdmiralFateReportsQueueAndResume(t *testing.T) {
	g := &Game{
		res: &Resources{
			FlipReport:  &screen.FlipReportPanel{},
			AdmiralNames: make([][]uint16, 82),
		},
		admiralFateReports: []AdmiralFateReport{
			{Admiral: 1, Faction: 0, Dead: true},
			{Admiral: 2, Faction: 0},
		},
		state: StateCommandPhase,
	}
	resumed := 0
	if !g.startAdmiralFateReports(func() { resumed++ }) {
		t.Fatal("兩則 fate 沒有開報告")
	}
	if g.state != StateAdmiralFateReport {
		t.Fatalf("狀態 %v，預期 StateAdmiralFateReport", g.state)
	}
	g.Click(2, 470)
	if r, ok := g.CurrentAdmiralFateReport(); !ok || r.Admiral != 2 || resumed != 0 {
		t.Fatalf("第一下後 ＝ %+v,%v resumed=%d", r, ok, resumed)
	}
	g.RightClick(300, 200)
	if _, ok := g.CurrentAdmiralFateReport(); ok || resumed != 1 {
		t.Fatalf("第二下後仍有報告或續跑 %d 次", resumed)
	}
	if g.afterAdmiralFateReports != nil {
		t.Fatal("報告清空後仍保留續跑 callback")
	}
}
