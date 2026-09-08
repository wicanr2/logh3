package rules

import (
	"path/filepath"
	"testing"
)

func loadKinds(t *testing.T) *ShipKindTable {
	t.Helper()
	tab, err := LoadShipKindTable(filepath.Join("../..",
		"docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Skipf("讀不到艦種表，跳過：%v", err)
	}
	return tab
}

// 表裡的值要與 `docs/re/94` 記的相同。
func TestShipKindTableValues(t *testing.T) {
	tab := loadKinds(t)
	cases := []struct{ faction, kind, a, b int }{
		{0, 0, 100, 95},
		{0, 3, 240, 180},
		{0, 4, 80, 75},
		{1, 4, 78, 73},
		{0, 7, 50, 50},
		{1, 7, 60, 60},
	}
	for _, c := range cases {
		k, ok := tab.Kind(c.faction, c.kind)
		if !ok {
			t.Fatalf("陣營 %d 艦種 %d 查不到", c.faction, c.kind)
		}
		if k.SemiMajor != c.a || k.SemiMinor != c.b {
			t.Errorf("陣營 %d 艦種 %d：半軸 %d／%d，要 %d／%d",
				c.faction, c.kind, k.SemiMajor, k.SemiMinor, c.a, c.b)
		}
	}
}

// 攻擊係數會隨方向變，而且 K 在分母——從目標防禦弱的方向打，係數越大。
// （傳進去的陣營／艦種／能力都是**目標**的，`docs/re/94`。）
func TestSkirmishAttackDirection(t *testing.T) {
	tab := loadKinds(t)
	const attack, ability = 20000, 100
	near := tab.SkirmishAttack(0, 0, BattleTurn/2, ability, attack) // K 最小
	far := tab.SkirmishAttack(0, 0, 0, ability, attack)             // K 最大
	if near <= far {
		t.Errorf("K 在分母，近拱點那一側該比較強：%d vs %d", near, far)
	}
	if tab.SkirmishAttack(0, 99, 0, ability, attack) != 0 {
		t.Error("查不到的艦種該回 0")
	}
}
