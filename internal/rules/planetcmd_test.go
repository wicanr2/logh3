package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// 要塞砲的射程：外圈 ＝ 防衛力 ÷ 200 ＋ 80、內圈 ＝ 外圈 ÷ 4
// （`2264:0EC3`）。伊謝爾倫防衛力 12800 → 144 與 36。
func TestFortressRange(t *testing.T) {
	for _, c := range []struct{ def, outer, inner int }{
		{12800, 144, 36}, // 伊謝爾倫
		{9800, 129, 32},  // 連典貝爾克
		{0, 80, 20},
	} {
		o, i := FortressRange(c.def)
		if o != c.outer || i != c.inner {
			t.Errorf("防衛力 %d → (%d,%d)，該是 (%d,%d)", c.def, o, i, c.outer, c.inner)
		}
	}
	for _, c := range []struct {
		state int
		ok    bool
	}{{0, true}, {3, true}, {1, false}, {2, false}, {4, false}} {
		if got := FortressGunReady(c.state); got != c.ok {
			t.Errorf("狀態 %d 能不能重新指定 ＝ %v，該是 %v", c.state, got, c.ok)
		}
	}
}

// 行星存量的欄位位置：拿執行期那份 81 筆行星表驗。
//
//	+0x3C 與 +0x40 開局相等（都約等於經濟力的 1.3 倍）
//	+0x20 起七格是艦種 0–6 各一份的補充存量
func TestPlanetStockFields(t *testing.T) {
	p := filepath.Join("..", "..", "workplace/s3datamem.bin")
	mem, err := os.ReadFile(p)
	if err != nil || len(mem) < 0x53132+81*PlanetRecord {
		t.Skip("缺 s3datamem.bin，跳過（不用代用品）")
	}
	same, all := 0, 0
	for i := 0; i < 81; i++ {
		rec := mem[0x53132+i*PlanetRecord : 0x53132+(i+1)*PlanetRecord]
		s, ok := ReadPlanetStock(rec)
		if !ok {
			t.Fatal("讀不出存量")
		}
		all++
		if s.Missiles == s.Fighters && s.Missiles > 0 {
			same++
		}
	}
	if same != all {
		t.Errorf("%d／%d 顆的飛彈與艦載機存量相等，開局該是全部相等", same, all)
	}
}

// 補給用行星的存量，滿載量的算法與補給艦那支相同；補充補到**建制滿編**
// `+0x0A`（不是修理的 `+0x0C`），扣的是該艦種那一格。
func TestPlanetSupplyAndReplenish(t *testing.T) {
	tab, err := LoadShipKindTable(filepath.Join("..", "..",
		"docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	k, ok := tab.Kind(1, 0)
	if !ok {
		t.Fatal("讀不到陣營 1 艦種 0")
	}
	per, full := k.Byte(kindMissilePer), k.Byte(kindFighterFull)

	s := &PlanetStock{Missiles: 1000, Fighters: 1000}
	u := &BattleUnit{Faction: 1, Kind: 0, Ships: 100, FullShips: 200, RepairCap: 150}
	m, f := PlanetSupply(tab, s, u)
	if m != per*100 || f != full*100 {
		t.Errorf("補了 %d／%d，該是 %d／%d", m, f, per*100, full*100)
	}
	if s.Missiles != 1000-m || s.Fighters != 1000-f {
		t.Errorf("行星剩 %d／%d，該是 %d／%d", s.Missiles, s.Fighters, 1000-m, 1000-f)
	}
	// 再補一次就沒得補了。
	if m2, f2 := PlanetSupply(tab, s, u); m2 != 0 || f2 != 0 {
		t.Errorf("已經滿了還補 %d／%d", m2, f2)
	}

	// 補充：補到 200（`+0x0A`），不是 150（`+0x0C`）。
	s.Ships[0] = 60
	if n := PlanetReplenish(s, u); n != 60 || u.Ships != 160 || s.Ships[0] != 0 {
		t.Fatalf("存量 60 補了 %d 艘（→ %d），行星剩 %d", n, u.Ships, s.Ships[0])
	}
	s.Ships[0] = 1000
	if n := PlanetReplenish(s, u); n != 40 || u.Ships != u.FullShips {
		t.Fatalf("存量充足時補了 %d 艘（→ %d），該補到滿編 %d",
			n, u.Ships, u.FullShips)
	}
	// 不同艦種扣不同格。
	u2 := &BattleUnit{Faction: 1, Kind: 2, Ships: 10, FullShips: 20}
	s.Ships[2] = 5
	if n := PlanetReplenish(s, u2); n != 5 || s.Ships[0] != 960 {
		t.Errorf("補了 %d 艘，艦種 0 那一格變成 %d（不該動）", n, s.Ships[0])
	}
}
