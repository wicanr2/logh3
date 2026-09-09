package rules

import "testing"

// TestRefitMatchesReceipt 拿收據 `refit2` 的局面跑一次整補，逐組比。
//
// 局面（`docs/re/231`）：同盟提督 52 那一支停在星域 30，編制 4／4，
// 兩組艦種 0 被 `-poke` 改成 1000／3000（不到七成五）。
//
//	前  組0 艦種0 槽20 3000/3000   星域待命部隊 [3000 0 4000 0 3000 0 0]
//	    組1 艦種0 槽10 1000/3000
//	    組2 艦種0 槽12 1000/3000
//	    組3 艦種6 槽29 1000/1000
//	後  組0 艦種0 槽20 3000/3000   星域待命部隊 [2000 0    0 0 3000 0 0]
//	    組1 艦種0 槽10 3000/3000
//	    組2 艦種2 槽12 4000/4000
//	    組3 艦種6 槽29 1000/1000
func TestRefitMatchesReceipt(t *testing.T) {
	kinds, tt := kindTable(t), compTables(t)
	rec := make([]byte, RegionRecord)
	rec[0x12] = 1
	for k, v := range []int{3000, 0, 4000, 0, 3000, 0, 0} {
		putStock(rec, 1, k, v)
	}
	groups := []FleetGroup{
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20},
		{Kind: 0, Ships: 1000, Full: 3000, Slot2: 0, Slot3: 10},
		{Kind: 0, Ships: 1000, Full: 3000, Slot2: 0, Slot3: 12},
		{Kind: 6, Ships: 1000, Full: 1000, Slot2: 6, Slot3: 29},
	}
	in := BuildInput{Region: rec, Faction: 1, Kinds: kinds, Tables: tt,
		AdmiralKind: 0, ClassKey: 4}
	got := Refit(in, groups, 4)

	want := []struct{ kind, ships, slot3 int }{
		{0, 3000, 20}, {0, 3000, 10}, {2, 4000, 12}, {6, 1000, 29},
	}
	if len(got) != len(want) {
		t.Fatalf("整補之後有 %d 組，收據是 %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Kind != w.kind || got[i].Ships != w.ships || got[i].Slot3 != w.slot3 {
			t.Errorf("組%d ＝ 艦種 %d／%d 艘／槽 %d，收據是 %d／%d／%d",
				i, got[i].Kind, got[i].Ships, got[i].Slot3, w.kind, w.ships, w.slot3)
		}
	}
	if s, w := RegionShips(rec, 1), []int{2000, 0, 0, 0, 3000, 0, 0}; !equalInts(s, w) {
		t.Errorf("待命部隊 ＝ %v，收據是 %v", s, w)
	}
}

func equalInts(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestRefitLeavesFullFleetAlone 全部滿編就什麼都不做——`turn260mem`
// 第一個 AI 回合走到整補時就是這個局面。
func TestRefitLeavesFullFleetAlone(t *testing.T) {
	kinds, tt := kindTable(t), compTables(t)
	rec := make([]byte, RegionRecord)
	rec[0x12] = 1
	for k := 0; k < regionShipKinds; k++ {
		putStock(rec, 1, k, 10000)
	}
	before := RegionShips(rec, 1)
	groups := []FleetGroup{
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20},
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 10},
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 12},
		{Kind: 6, Ships: 1000, Full: 1000, Slot2: 6, Slot3: 29},
	}
	in := BuildInput{Region: rec, Faction: 1, Kinds: kinds, Tables: tt,
		AdmiralKind: 0, ClassKey: 4}
	got := Refit(in, append([]FleetGroup(nil), groups...), 4)
	if len(got) != 4 {
		t.Fatalf("滿編的艦隊被改成 %d 組", len(got))
	}
	for i := range groups {
		if got[i] != groups[i] {
			t.Errorf("組%d 被動過：%+v → %+v", i, groups[i], got[i])
		}
	}
	if !equalInts(RegionShips(rec, 1), before) {
		t.Errorf("待命部隊被動過：%v → %v", before, RegionShips(rec, 1))
	}
}

// TestRefitSacrificesStaleSlot 釘「編制已滿要犧牲一組」那一條
// （原版 `2D76:0D40` ＋ `2D76:0DB2`，收據 `refit3`／`docs/re/232`）。
//
// 把提督 52 那一支的組1 的隊形槽 `-poke` 成 99（樣板裡沒有這個槽），
// 編制仍然是 4／4。整補時樣板第 1 格要的槽 10 找不到、編制又滿了，
// 所以挑「隊形槽已經不在樣板裡」的那一組犧牲掉，重建成槽 10。
//
//	前  組0 艦種0 槽20  組1 艦種0 槽**99**  組2 艦種0 槽12  組3 艦種6 槽29
//	後  組0 艦種0 槽20  組1 艦種0 槽**10**  組2 艦種0 槽12  組3 艦種6 槽29
//	待命部隊前後相同——拆掉還回去 3000、重建又拿走 3000。
func TestRefitSacrificesStaleSlot(t *testing.T) {
	kinds, tt := kindTable(t), compTables(t)
	rec := make([]byte, RegionRecord)
	rec[0x12] = 1
	for k, v := range []int{3000, 0, 4000, 0, 3000, 0, 0} {
		putStock(rec, 1, k, v)
	}
	before := RegionShips(rec, 1)
	groups := []FleetGroup{
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20},
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 99},
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 12},
		{Kind: 6, Ships: 1000, Full: 1000, Slot2: 6, Slot3: 29},
	}
	in := BuildInput{Region: rec, Faction: 1, Kinds: kinds, Tables: tt,
		AdmiralKind: 0, ClassKey: 4}
	got := Refit(in, groups, 4)

	want := []struct{ kind, ships, slot3 int }{
		{0, 3000, 20}, {0, 3000, 10}, {0, 3000, 12}, {6, 1000, 29},
	}
	if len(got) != len(want) {
		t.Fatalf("整補之後有 %d 組，收據是 %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Kind != w.kind || got[i].Ships != w.ships || got[i].Slot3 != w.slot3 {
			t.Errorf("組%d ＝ 艦種 %d／%d 艘／槽 %d，收據是 %d／%d／%d",
				i, got[i].Kind, got[i].Ships, got[i].Slot3, w.kind, w.ships, w.slot3)
		}
	}
	if s := RegionShips(rec, 1); !equalInts(s, before) {
		t.Errorf("待命部隊 ＝ %v，收據是前後相同的 %v", s, before)
	}
}

// TestRefitBalancesFiveSixFromReceipt 釘替代清單裡的 `0xFE`
// （收據 `refit4`／`docs/re/233`）。
//
// 把星域 30 的艦種 0–4 存量 `-poke` 成 0、5 與 6 各 30000，再把艦隊槽 28
// 的現有組數改成 2。補編制要艦種 0 的那一格拿不到存量，走替代清單
// `2 → 4 → FE → 3 → 1`：2 與 4 都沒有存量，`0xFE` 那一格兩邊都是 0 組
// **先試 6**，拿到了。
//
//	後  組0 艦種0 槽20  組1 艦種0 槽10  組2 艦種6 槽12  組3 艦種6 槽29
//	待命部隊 艦種6 30000 → 28000（兩組各 1000），艦種5 不動
func TestRefitBalancesFiveSixFromReceipt(t *testing.T) {
	kinds, tt := kindTable(t), compTables(t)
	rec := make([]byte, RegionRecord)
	rec[0x12] = 1
	for k, v := range []int{0, 0, 0, 0, 0, 30000, 30000} {
		putStock(rec, 1, k, v)
	}
	groups := []FleetGroup{
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20},
		{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 10},
	}
	in := BuildInput{Region: rec, Faction: 1, Kinds: kinds, Tables: tt,
		AdmiralKind: 0, ClassKey: 4}
	got := Refit(in, groups, 4)

	want := []struct{ kind, ships, slot3 int }{
		{0, 3000, 20}, {0, 3000, 10}, {6, 1000, 12}, {6, 1000, 29},
	}
	if len(got) != len(want) {
		t.Fatalf("整補之後有 %d 組，收據是 %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Kind != w.kind || got[i].Ships != w.ships || got[i].Slot3 != w.slot3 {
			t.Errorf("組%d ＝ 艦種 %d／%d 艘／槽 %d，收據是 %d／%d／%d",
				i, got[i].Kind, got[i].Ships, got[i].Slot3, w.kind, w.ships, w.slot3)
		}
	}
	if s, w := RegionShips(rec, 1), []int{0, 0, 0, 0, 0, 30000, 28000}; !equalInts(s, w) {
		t.Errorf("待命部隊 ＝ %v，收據是 %v", s, w)
	}
}
