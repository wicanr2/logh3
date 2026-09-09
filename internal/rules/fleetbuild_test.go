package rules

import (
	"path/filepath"
	"testing"
)

func compTables(t *testing.T) *CompositionTables {
	t.Helper()
	data := filepath.Join("../..", "docs/internals/data")
	tt, err := LoadCompositionTables(
		filepath.Join(data, "fleet-composition-templates.json"),
		filepath.Join(data, "ship-kind-fallback.json"))
	if err != nil {
		t.Fatalf("讀編制樣板／替代表：%v", err)
	}
	return tt
}

// TestLoadCompositionTables 兩張表都讀得進來、形狀對。
func TestLoadCompositionTables(t *testing.T) {
	tt := compTables(t)
	// 每個類別鍵的第 0 組都是 (0, 20)（`docs/re/212`）。
	for k := 0; k < 16; k++ {
		if got := tt.Template[k][0]; got != [2]int{0, 20} {
			t.Errorf("類別鍵 %d 的第 0 組 ＝ %v，預期 [0 20]", k, got)
		}
	}
	// 替代表：艦種 5 是 6 4 2 0 3（`docs/re/216`）。
	if got := tt.Fallback[5]; len(got) != 5 ||
		got[0] != 6 || got[1] != 4 || got[2] != 2 || got[3] != 0 || got[4] != 3 {
		t.Errorf("艦種 5 的替代順序 ＝ %v，預期 [6 4 2 0 3]", got)
	}
	// 0xFF 不該留在清單裡。
	for k, order := range tt.Fallback {
		for _, c := range order {
			if c == 0xFF {
				t.Errorf("艦種 %d 的清單裡留著 0xFF", k)
			}
		}
	}
}

// TestBuildCompositionDeductsStock 編制會從星域存量扣掉，而且組數
// 等於 `FleetSlots(等級)`（存量夠的時候）。
func TestBuildCompositionDeductsStock(t *testing.T) {
	kinds, tt := kindTable(t), compTables(t)
	rec := make([]byte, RegionRecord)
	rec[0x12] = 0 // 擁有者：帝國
	// 七個艦種各給一萬艘，夠編滿。
	for k := 0; k < regionShipKinds; k++ {
		putStock(rec, 0, k, 10000)
	}
	before := 0
	for _, v := range RegionShips(rec, 0) {
		before += v
	}

	got := BuildComposition(BuildInput{
		Region: rec, Faction: 0, Kinds: kinds, Tables: tt,
		AdmiralLevel: 20, AdmiralKind: 0, ClassKey: 0,
	})
	if len(got) != FleetSlots(20) {
		t.Fatalf("編出 %d 組，預期 %d", len(got), FleetSlots(20))
	}
	after := 0
	for _, v := range RegionShips(rec, 0) {
		after += v
	}
	taken := 0
	for _, g := range got {
		if g.Kind == shipKindStandalone {
			continue // 那一種不扣存量
		}
		taken += g.Ships
	}
	if before-after != taken {
		t.Errorf("存量少了 %d，編制拿走 %d——對不上", before-after, taken)
	}
	if taken == 0 {
		t.Error("一艘都沒搬")
	}
	// 每一組的艘數不該超過滿編。
	for i, g := range got {
		if g.Full > 0 && g.Ships > g.Full {
			t.Errorf("第 %d 組 %d 艘超過滿編 %d", i, g.Ships, g.Full)
		}
	}
}

// TestBuildCompositionStopsWhenStockRunsOut 存量空了就編不下去，
// 但旗艦那一組一定在（原版 `297B:0066` 先填它）。
func TestBuildCompositionStopsWhenStockRunsOut(t *testing.T) {
	kinds, tt := kindTable(t), compTables(t)
	rec := make([]byte, RegionRecord)
	got := BuildComposition(BuildInput{
		Region: rec, Faction: 0, Kinds: kinds, Tables: tt,
		AdmiralLevel: 39, AdmiralKind: 0, ClassKey: 0,
	})
	if len(got) != 1 {
		t.Fatalf("存量全 0 卻編出 %d 組，預期只有旗艦那一組", len(got))
	}
	if got[0].Ships != 0 {
		t.Errorf("旗艦那一組拿到 %d 艘，預期 0", got[0].Ships)
	}
}

// TestBuildCompositionSubstitutesKind 樣板要的艦種沒存量時會換一個。
func TestBuildCompositionSubstitutesKind(t *testing.T) {
	kinds, tt := kindTable(t), compTables(t)
	rec := make([]byte, RegionRecord)
	// 只給艦種 2 存量——樣板要別的時候只能換到它（艦種 0 的替代清單
	// 第一個就是 2，`docs/re/216`）。
	putStock(rec, 0, 2, 10000)
	got := BuildComposition(BuildInput{
		Region: rec, Faction: 0, Kinds: kinds, Tables: tt,
		AdmiralLevel: 20, AdmiralKind: 2, ClassKey: 0,
	})
	if len(got) < 2 {
		t.Fatalf("只編出 %d 組", len(got))
	}
	for i, g := range got {
		if g.Kind != 2 && g.Kind != shipKindStandalone {
			t.Errorf("第 %d 組是艦種 %d，只有艦種 2 有存量", i, g.Kind)
		}
	}
}

func putStock(rec []byte, faction, kind, n int) {
	o := regionShips + faction*regionShipKinds*4 + kind*4
	rec[o] = byte(n)
	rec[o+1] = byte(n >> 8)
	rec[o+2] = byte(n >> 16)
	rec[o+3] = byte(n >> 24)
}
