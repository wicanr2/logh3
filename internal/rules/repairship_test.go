package rules

import (
	"path/filepath"
	"testing"
)

// 規格125：修護艦可被其他修護艦修理（可選規則，預設關閉）。
// 場景：A、B 是同陣營修護艦，A 受損且有庫存；E 是一般受損艦；C 是偵察隊；D 是他方修護艦。
func repairShipScenario() []BattleUnit {
	return []BattleUnit{
		{Faction: 0, Kind: RepairKind, X: 100, Y: 100, Ships: 900, RepairCap: 1000, Missiles: 1000, Fleet: 1, Squadron: 1}, // A：來源，自己也受損
		{Faction: 0, Kind: RepairKind, X: 110, Y: 100, Ships: 600, RepairCap: 1000, Missiles: 0, Fleet: 1, Squadron: 2},    // B：受損的另一艘修護艦
		{Faction: 0, Kind: 2, X: 120, Y: 100, Ships: 100, RepairCap: 200, Fleet: 1, Squadron: 3},                           // E：一般受損艦
		{Faction: 0, Kind: 7, X: 105, Y: 100, Ships: 10, RepairCap: 50, Fleet: 1, Squadron: 4},                             // C：偵察隊
		{Faction: 1, Kind: RepairKind, X: 108, Y: 100, Ships: 300, RepairCap: 500, Fleet: 2, Squadron: 1},                  // D：他方修護艦
	}
}

func runRepairPhase(t *testing.T, on bool) []BattleUnit {
	t.Helper()
	kinds, err := LoadShipKindTable(filepath.Join("../..", "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	units := repairShipScenario()
	refs := make([]PreparationUnit, len(units))
	for i := range units {
		refs[i] = PreparationUnit{Pool: i, Unit: &units[i]}
	}
	prep := NewBattlePreparation(kinds, refs, nil, &PreparationSources{}, func(i int) int { return units[i].Fleet })
	prep.SetRepairShipsRepairable(on)
	prep.phase = 2 // 只執行修理階段，與 TestPreparationResourcesMatchesOriginal 相同作法。
	for {
		if _, ok := prep.Next(); !ok {
			break
		}
	}
	return units
}

func TestPreparationRepairShipsOffMatchesOriginal(t *testing.T) {
	u := runRepairPhase(t, false)
	if u[1].Ships != 600 {
		t.Errorf("預設關閉時修護艦不該被修理：B 艘數 %d", u[1].Ships)
	}
	if u[2].Ships != 200 || u[0].Missiles != 900 {
		t.Errorf("一般受損艦該補滿 100 艘且 A 扣 100：E=%d A庫存=%d", u[2].Ships, u[0].Missiles)
	}
}

func TestPreparationRepairShipsOn(t *testing.T) {
	u := runRepairPhase(t, true)
	if u[1].Ships != 1000 {
		t.Errorf("開啟後 B 該被 A 修到上限：B 艘數 %d", u[1].Ships)
	}
	if u[2].Ships != 200 {
		t.Errorf("一般受損艦 E 該補滿：%d", u[2].Ships)
	}
	if u[0].Missiles != 1000-400-100 {
		t.Errorf("A 該扣 B 的 400 與 E 的 100：庫存 %d", u[0].Missiles)
	}
	if u[0].Ships != 900 {
		t.Errorf("來源 A 不能修理自己：艘數 %d", u[0].Ships)
	}
	if u[3].Ships != 10 {
		t.Errorf("偵察隊仍排除：%d", u[3].Ships)
	}
	if u[4].Ships != 300 {
		t.Errorf("他方修護艦不該被修理：%d", u[4].Ships)
	}
}

// 非修護艦的單位在兩種偏好下的結果相同。
func TestPreparationRepairShipsKeepsOtherKinds(t *testing.T) {
	off, on := runRepairPhase(t, false), runRepairPhase(t, true)
	for _, i := range []int{2, 3, 4} {
		if off[i] != on[i] {
			t.Errorf("單位 %d（艦種 %d）在兩種偏好下不同：關 %+v 開 %+v", i, off[i].Kind, off[i], on[i])
		}
	}
}
