package rules

import (
	"path/filepath"
	"testing"
)

// 1261:03A5全滅只設bit3；後續行星階段不能把原+3C清零。
func TestRefreshIntegrityKeepsAnnihilatedFleetWithScout(t *testing.T) {
	fleets := []BattleFleet{{Index: 7, Integrity: 42, Flag0: true, AdmiralLossFactor: 100}, {Index: 8, Integrity: 100, AdmiralLossFactor: 100}}
	units := []BattleUnit{{Fleet: 7, Kind: 0, Squadron: 0, Ships: 0, FullShips: 3000}, {Fleet: 7, Kind: 7, Squadron: -1, Ships: 10, FullShips: 10}, {Fleet: 8, Kind: 0, Squadron: 0, Ships: 600, FullShips: 1000}}
	RefreshFleetIntegrity(fleets, units)
	if fleets[0].Integrity != 42 || !fleets[0].Flag0 || fleets[1].Integrity != 60 {
		t.Fatalf("全滅／獨立斥候／存活刷新不符：%+v", fleets)
	}
}

func TestPlanetDefenceGaugesFollowCounterSnapshots(t *testing.T) {
	kinds, err := LoadShipKindTable(filepath.Join("../..", "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	f := BattleFleet{Index: 7, Faction: 0, Integrity: 71, Ability3: 80, Defence: 50, AdmiralLossFactor: 100}
	p := BattlePlanet{Owner: 1, Defence: 3300, DefenceMax: 3300}
	u := []BattleUnit{{Fleet: 7, Faction: 0, Kind: 0, Squadron: 0, Ships: 1000, FullShips: 2000},
		{Fleet: 7, Faction: 0, Kind: 1, Squadron: 1, Ships: 7000, FullShips: 7000}}
	before := append([]BattleUnit(nil), u...)
	r := NewRand(0x12345678)
	h, ok := assaultOne(&p, 0, &f, 2, u, kinds, r) // 明示只反擊的局部診斷
	if !ok || h.DefenceGauges == nil || h.Counter == 0 {
		t.Fatalf("無反擊快照 %+v", h)
	}
	g := h.DefenceGauges
	if g.BeforeIntegrity != 71 || g.BeforeAttack != FleetAttack(&f, before) || g.AfterAttack != FleetAttack(&f, u) || g.AfterIntegrity != FleetIntegrity(&f, u) {
		t.Fatalf("快照未承接實際反擊 %+v", g)
	}
	if g.AfterAttack >= g.BeforeAttack || g.AfterIntegrity == 100*(u[0].Ships+u[1].Ships)/(u[0].FullShips+u[1].FullShips) {
		t.Fatal("未更新攻擊力或誤用艘數比例")
	}
	// 直接跑既有反擊，確認演出快照沒有新增RND或資料修改。
	p2 := BattlePlanet{Owner: 1, Defence: 3300, DefenceMax: 3300}
	r2 := NewRand(0x12345678)
	if got := assaultCounter(&p2, &f, 2, before, kinds, r2); got != h.Counter || r2.Seed() != r.Seed() {
		t.Fatal("演出快照改動戰損或RNG")
	}
}
