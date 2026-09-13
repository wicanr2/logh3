package rules

import (
	"encoding/binary"
	"reflect"
	"testing"
)

func phezzanFixture() ([]byte, *Planets) {
	per := make([]int, RegionCount)
	per[0], per[1], per[10], per[14], per[16], per[31] = 1, 32, 1, 1, 45, 1
	p := &Planets{PerRegion: per, Data: make([]byte, PlanetHeader+81*PlanetRecord)}
	for i := 0; i < 81; i++ {
		p.Record(i)[planetOwner] = RegionOwnerNone
	}
	records := make([]byte, RegionCount*RegionRecord)
	for region := 0; region < RegionCount; region++ {
		records[region*RegionRecord+regionOwner] = RegionOwnerNone
	}
	setOwner := func(region, owner int) {
		records[region*RegionRecord+regionOwner] = byte(owner)
		start, n := p.Range(region)
		for i := start; i < start+n; i++ {
			p.Record(i)[planetOwner] = byte(owner)
		}
	}
	for _, region := range []int{0, 1, 10, 14} {
		setOwner(region, 0)
	}
	setOwner(31, 1)
	for i := 1; i <= 32; i++ {
		p.SetPlanetEconomy(i, 10)
	}
	p.SetPlanetEconomy(0, 1000)
	p.SetPlanetEconomy(33, 900)
	p.SetPlanetEconomy(34, 103100)
	p.SetPlanetEconomy(80, 1000)
	return records, p
}

func TestPhezzanEventEmpireMatchesConfirmedRules(t *testing.T) {
	records, planets := phezzanFixture()
	got := RunPhezzanEvent(records, planets, 0, true, 0, nil)
	if got.Transferred != 82480 || !reflect.DeepEqual(got.Rebellions, []int{10}) {
		t.Fatalf("本體結果 ＝ %+v，預期搬 82480、星域 10 叛離", got)
	}
	if got := planets.PlanetEconomy(34); got != 20620 {
		t.Errorf("費沙經濟力 ＝ %d，預期 20620", got)
	}
	if got := planets.PlanetEconomy(80); got != 83480 {
		t.Errorf("海尼森經濟力 ＝ %d，預期 83480", got)
	}
	if got := planets.PlanetEconomy(33); got != 600 {
		t.Errorf("叛離行星經濟力 ＝ %d，預期 600", got)
	}
	if got := planets.Owner(33); got != RegionOwnerNone {
		t.Errorf("叛離行星擁有者 ＝ %d，預期無主", got)
	}
	if got := planets.PlanetEconomy(0); got != 500 {
		t.Errorf("人類方懲罰後奧丁經濟力 ＝ %d，預期 500", got)
	}
	if got := records[RegionMorale]; got != 100 {
		t.Errorf("懲罰後星域 0 民心 ＝ %d，預期 100", got)
	}
	if got := RegionEconomy(records[31*RegionRecord:]); got != 83480 {
		t.Errorf("對手首都星域經濟力 ＝ %d，預期 83480", got)
	}
}

func TestPhezzanEventGarrisonAndComputerPenalty(t *testing.T) {
	records, planets := phezzanFixture()
	RunPhezzanEvent(records, planets, 0, false, 0,
		func(region, faction int) int {
			if region == 10 && faction == 0 {
				return 1
			}
			return 0
		})
	if got := planets.Owner(33); got != 0 {
		t.Errorf("有駐軍的星域仍叛離，行星擁有者 ＝ %d", got)
	}
	if got := planets.PlanetEconomy(0); got != 1000 {
		t.Errorf("電腦方受到了人類懲罰：%d", got)
	}
}

func TestPhezzanPenaltyPercentAndSaveFlag(t *testing.T) {
	for _, c := range []struct{ turn, want int }{
		{0, 50}, {99, 50}, {100, 40}, {499, 10}, {500, 12}, {599, 12}, {600, 0},
	} {
		if got := phezzanPenaltyPercent(c.turn); got != c.want {
			t.Errorf("回合 %d 的懲罰 ＝ %d，預期 %d", c.turn, got, c.want)
		}
	}
	work := make([]byte, WorkRecord)
	if PhezzanOccurred(work) {
		t.Fatal("空白 WORK 不該已有費沙事件")
	}
	SetPhezzanOccurred(work)
	if !PhezzanOccurred(work) || binary.LittleEndian.Uint16(work[0x34:]) != 1 {
		t.Fatal("WORK +0x34 沒寫成 1")
	}
}
