package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

func TestOccupationMessageSelection(t *testing.T) {
	for _, c := range []struct {
		occupier, seat int
		continuing     bool
		want           string
	}{
		{0, 0, false, "occupation_new_us"},
		{0, 1, false, "occupation_new_enemy"},
		{1, 1, true, "occupation_continuing_us"},
		{1, 0, true, "occupation_continuing_enemy"},
	} {
		r := OccupationReport{Occupier: c.occupier, Seat: c.seat, Continuing: c.continuing}
		if got := r.MessageKey(); got != c.want {
			t.Errorf("occupier=%d seat=%d continuing=%v：%s，預期 %s",
				c.occupier, c.seat, c.continuing, got, c.want)
		}
	}
}

func TestPrepareOccupationReportsUsesOldBitForContinuing(t *testing.T) {
	g := newRegionTurnGame(t)
	regionPass(g)
	g.faction = 0
	g.fleets = []rules.Fleet{{Faction: 0, Region: 20}}
	g.prepareOccupationReports()
	if len(g.occupationReports) != 1 {
		t.Fatalf("首次占據排了 %d 則，預期 1", len(g.occupationReports))
	}
	r := g.occupationReports[0]
	if r.Region != 20 || r.Owner != 1 || r.Occupier != 0 || r.Continuing {
		t.Fatalf("首次報告 ＝ %+v，預期帝國新占據同盟星域 20", r)
	}
	if g.regionRecord(20)[regionFlagsOff]&0x10 == 0 {
		t.Fatal("準備報告後沒有先寫入 bit4")
	}

	g.prepareOccupationReports()
	if len(g.occupationReports) != 1 || !g.occupationReports[0].Continuing {
		t.Fatalf("第二回合報告 ＝ %+v，預期持續占據", g.occupationReports)
	}
}

func TestOccupationReportsQueueAndResume(t *testing.T) {
	g := &Game{
		res: &Resources{FlipReport: &screen.FlipReportPanel{}},
		occupationReports: []OccupationReport{
			{Region: 16, Owner: 1, Occupier: 0, Seat: 0},
			{Region: 20, Owner: 1, Occupier: 0, Seat: 0, Continuing: true},
		},
		state: StateCommandPhase,
	}
	resumed := 0
	if !g.startOccupationReports(func() { resumed++ }) {
		t.Fatal("兩格占據沒有開報告")
	}
	if g.state != StateOccupationReport {
		t.Fatalf("狀態 %v，預期 StateOccupationReport", g.state)
	}
	g.Click(2, 470)
	if r, ok := g.CurrentOccupationReport(); !ok || r.Region != 20 || resumed != 0 {
		t.Fatalf("第一下後 ＝ %+v,%v resumed=%d，預期星域 20、尚未續跑", r, ok, resumed)
	}
	g.RightClick(300, 200)
	if _, ok := g.CurrentOccupationReport(); ok || resumed != 1 {
		t.Fatalf("第二下後仍有報告或續跑 %d 次，預期清空且一次", resumed)
	}
	if g.afterOccupationReports != nil {
		t.Fatal("報告清空後仍保留續跑 callback")
	}
}
