package game

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 規格125：戰場「修理」指令的候選。不依賴原版檔案，只用 repo 內的艦種表。
func repairRuleGame(t *testing.T) *Game {
	t.Helper()
	kinds, err := rules.LoadShipKindTable(filepath.Join("../..", "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{res: &Resources{ShipKinds: kinds}}
	g.battleUnits = []rules.BattleUnit{
		{Faction: 0, Kind: shipKindRepair, X: 100, Y: 100}, // 0：發令的修護艦
		{Faction: 0, Kind: shipKindRepair, X: 110, Y: 100}, // 1：另一艘修護艦
		{Faction: 0, Kind: 2, X: 120, Y: 100},              // 2：一般艦
		{Faction: 0, Kind: shipKindScout, X: 105, Y: 100},  // 3：偵察隊
		{Faction: 1, Kind: shipKindRepair, X: 108, Y: 100}, // 4：他方修護艦
		{Faction: 0, Kind: shipKindSupply, X: 115, Y: 100}, // 5：運補艦
	}
	return g
}

func TestSupplyCandidatesRepairShipsDefaultOff(t *testing.T) {
	g := repairRuleGame(t)
	if got, want := g.supplyCandidates(0, true), []int{2, 5}; !slices.Equal(got, want) {
		t.Errorf("預設修理候選 %v，該是 %v（排除艦種 6／7）", got, want)
	}
	if got, want := g.supplyCandidates(0, false), []int{2}; !slices.Equal(got, want) {
		t.Errorf("預設補給候選 %v，該是 %v（排除艦種 5／6／7）", got, want)
	}
}

func TestSupplyCandidatesRepairShipsOn(t *testing.T) {
	g := repairRuleGame(t)
	g.SetRepairShipsRepairable(true)
	if !g.RepairShipsRepairable() {
		t.Fatal("偏好沒有設進去")
	}
	if got, want := g.supplyCandidates(0, true), []int{1, 2, 5}; !slices.Equal(got, want) {
		t.Errorf("開啟後修理候選 %v，該是 %v（含另一艘修護艦、不含自己、偵察隊與他方）", got, want)
	}
	if got, want := g.supplyCandidates(0, false), []int{2}; !slices.Equal(got, want) {
		t.Errorf("補給候選不該受影響：%v，該是 %v", got, want)
	}
}
