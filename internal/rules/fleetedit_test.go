package rules

import (
	"encoding/binary"
	"path/filepath"
	"testing"
)

// 艦隊組成編輯器的規則（`docs/re/317`）。畫面層的同狀態對拍在
// `internal/game/fleetedit_test.go`；這裡釘住規則本身的邊界。

func editorRegion(stock [7]int) []byte {
	rec := make([]byte, RegionRecord)
	for i, v := range stock {
		binary.LittleEndian.PutUint32(rec[regionShips+i*4:], uint32(v))
	}
	return rec
}

func editorTables(t *testing.T) *EditorTables {
	t.Helper()
	tb, err := LoadEditorTables(filepath.Join("..", "..", "docs", "internals", "data", "fleet-editor.json"))
	if err != nil {
		t.Fatal(err)
	}
	return tb
}

func TestFleetEditorPlaceRemoveCancel(t *testing.T) {
	rec := editorRegion([7]int{75000, 36000, 108000, 10000, 27000, 15000, 5000})
	flag := []FleetGroup{{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20}}
	e := NewFleetEditor(BuildInput{Region: rec, Faction: 0}, editorTables(t), 8, flag)
	if e.Used() != 1 || e.Free() != 7 {
		t.Fatalf("開編輯器 used %d free %d", e.Used(), e.Free())
	}
	// 中央只收 0／1。
	e.Kind = 2
	if r := e.Place(20); r != EditFlagshipKind {
		t.Fatalf("巡航艦放中央 → %v", r)
	}
	// 放、換、拆。
	e.Kind = 0
	if r := e.Place(19); r != EditOK || RegionShips(rec, 0)[0] != 72000 {
		t.Fatalf("放戰艦 → %v，存量 %v", r, RegionShips(rec, 0))
	}
	e.Kind = 2
	if r := e.Place(19); r != EditOK {
		t.Fatalf("換巡航艦 → %v", r)
	}
	if s := RegionShips(rec, 0); s[0] != 75000 || s[2] != 104000 {
		t.Fatalf("換掉之後存量 %v", s)
	}
	if r := e.Remove(20); r != EditFlagshipRemove {
		t.Fatalf("拆旗艦 → %v", r)
	}
	if r := e.Remove(0); r != EditNoop {
		t.Fatalf("拆空格 → %v", r)
	}
	if r := e.Remove(19); r != EditOK || RegionShips(rec, 0)[2] != 108000 {
		t.Fatalf("拆巡航艦 → %v，存量 %v", r, RegionShips(rec, 0))
	}
	// 沒貨。
	binary.LittleEndian.PutUint32(rec[regionShips+5*4:], 0)
	e.Kind = 5
	if r := e.Place(19); r != EditNoShips {
		t.Fatalf("運補艦沒貨 → %v", r)
	}
	// 滿了：放到 8/8 再放 → 超過；放在有人的格子上仍可以換。
	e.Kind = 0
	for _, slot := range []int{0, 1, 2, 3, 4, 5, 6} {
		if r := e.Place(slot); r != EditOK {
			t.Fatalf("放第 %d 格 → %v", slot, r)
		}
	}
	if r := e.Place(7); r != EditOverCapacity {
		t.Fatalf("8/8 再放 → %v", r)
	}
	e.Kind = 2
	if r := e.Place(6); r != EditOK {
		t.Fatalf("8/8 換一格 → %v", r)
	}
	// 中止：存量倒回。
	e.Cancel()
	if s := RegionShips(rec, 0); s[0] != 75000 || s[2] != 108000 || s[5] != 15000 {
		t.Fatalf("中止之後存量 %v", s)
	}
}

func TestFleetEditorGroupsOrderAndFormation(t *testing.T) {
	rec := editorRegion([7]int{75000, 36000, 108000, 10000, 27000, 0, 5000})
	flag := []FleetGroup{{Kind: 0, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20}}
	e := NewFleetEditor(BuildInput{Region: rec, Faction: 0}, editorTables(t), 8, flag)
	e.Kind = 2
	e.Place(21)
	e.Kind = 0
	e.Place(19)
	got := e.Groups()
	want := [][2]int{{0, 20}, {0, 19}, {2, 21}}
	for i, w := range want {
		if got[i].Slot2 != w[0] || got[i].Slot3 != w[1] {
			t.Fatalf("壓實排序 %v，要 %v", got, want)
		}
	}
	// 圓錐：旗艦以外全拆，(5,29) 換成修護艦（替代表第 5 列第一個有貨的 6），
	// 樣板第 8、9 組放不下。
	e.ApplyFormation(0)
	got = e.Groups()
	want = [][2]int{{0, 20}, {0, 24}, {0, 25}, {2, 32}, {2, 35}, {4, 28}, {4, 30}, {6, 29}}
	if len(got) != len(want) {
		t.Fatalf("圓錐之後 %d 組：%v", len(got), got)
	}
	for i, w := range want {
		if got[i].Slot2 != w[0] || got[i].Slot3 != w[1] {
			t.Fatalf("圓錐之後 %v，要 %v", got, want)
		}
	}
	if s := RegionShips(rec, 0); s[0] != 69000 || s[2] != 100000 || s[4] != 21000 || s[6] != 4000 {
		t.Fatalf("圓錐之後存量 %v", s)
	}
	if e.Kind != 0 {
		t.Fatalf("套完陣形反白列變成 %d", e.Kind)
	}
}

func TestFleetMobility(t *testing.T) {
	table := editorTables(t).Mobility
	for _, c := range []struct{ groups, want int }{{1, 100}, {2, 98}, {3, 95}, {8, 81}} {
		if got := FleetMobility(table, 115, c.groups); got != c.want {
			t.Errorf("%d 組：%d，原版 %d", c.groups, got, c.want)
		}
	}
}

// 副官候選與換副官（`docs/re/318`）。
func TestAdjutantCandidatesAndSetAdjutant(t *testing.T) {
	mk := func(faction, region, level int, post byte) []byte {
		r := make([]byte, AdmiralRecord)
		r[0], r[admiralFaction], r[admiralMerit], r[admiralFleet], r[admiralStaffFlag], r[admiralRegion] = post, byte(faction), byte(level), 0xFF, 0x80, byte(region)
		return r
	}
	recs := [][]byte{
		mk(0, 0, 29, 0xFF), // 0 帶隊提督
		mk(0, 0, 20, 0xFF), // 1 合格
		mk(0, 1, 20, 0xFF), // 2 別的星域
		mk(1, 0, 10, 0xFF), // 3 別的陣營
		mk(0, 0, 29, 0xFF), // 4 等級不比提督低
		mk(0, 0, 8, 0x01),  // 5 有職務
		mk(0, 0, 8, 0xFF),  // 6 合格，但呼叫端說在艦隊裡
	}
	all := make([]Admiral, 0, len(recs))
	for i, r := range recs {
		all = append(all, Admiral{Index: i, Faction: int(r[admiralFaction]), Merit: int(r[admiralMerit]), Record: r})
	}
	busy := func(i int) bool { return i == 0 || i == 6 }
	if got := AdjutantCandidates(all, recs[0], busy); len(got) != 1 || got[0] != 1 {
		t.Fatalf("候選 %v，要 [1]", got)
	}
	recs[0][admiralStaffFlag] = 0x03 // 帶隊提督在槽 3
	SetAdjutant(recs[0], nil, recs[1], 1)
	if recs[0][admiralFleet] != 1 || recs[1][admiralStaffFlag] != 0x03 {
		t.Fatalf("換副官之後 +07=%#x 新副官 +09=%#x", recs[0][admiralFleet], recs[1][admiralStaffFlag])
	}
	SetAdjutant(recs[0], recs[1], nil, -1)
	if recs[0][admiralFleet] != 0xFF || recs[1][admiralStaffFlag] != 0x80 {
		t.Fatalf("卸任之後 +07=%#x 舊副官 +09=%#x", recs[0][admiralFleet], recs[1][admiralStaffFlag])
	}
}

// 右下角「提督」分頁的清單與五列一頁（`docs/re/319`）。
func TestIdleAdmiralsInRegionAndListPage(t *testing.T) {
	mk := func(faction, region int, post byte) []byte {
		r := make([]byte, AdmiralRecord)
		r[0], r[admiralFaction], r[admiralFleet], r[admiralStaffFlag], r[admiralRegion] = post, byte(faction), 0xFF, 0x80, byte(region)
		return r
	}
	recs := [][]byte{
		mk(0, 0, 0xFF), // 0 合格
		mk(0, 1, 0xFF), // 1 別的星域
		mk(1, 0, 0xFF), // 2 別的陣營
		mk(0, 0, 0x02), // 3 有職務
		mk(0, 0, 0xFF), // 4 合格，但呼叫端說在艦隊裡
		mk(0, 0, 0xFF), // 5 合格
	}
	all := make([]Admiral, 0, len(recs))
	for i, r := range recs {
		all = append(all, Admiral{Index: i, Faction: int(r[admiralFaction]), Record: r})
	}
	got := IdleAdmiralsInRegion(all, 0, 0, func(i int) bool { return i == 4 })
	if len(got) != 2 || got[0] != 0 || got[1] != 5 {
		t.Fatalf("閒置提督 %v，要 [0 5]", got)
	}
	p := ListPage{Count: 35}
	if p.Arrows() != true {
		t.Fatal("35 筆要有 ▲▼")
	}
	p.Step(1)
	p.Step(1)
	if p.Scroll != 2 {
		t.Fatalf("▼ 兩下位移 %d，要 2", p.Scroll)
	}
	for i := 0; i < 40; i++ {
		p.Step(1)
	}
	if p.Scroll != 30 {
		t.Fatalf("捲到底位移 %d，要 30（35 − 5）", p.Scroll)
	}
	for i := 0; i < 40; i++ {
		p.Step(-1)
	}
	if p.Scroll != 0 {
		t.Fatalf("▲ 到頂位移 %d，要 0", p.Scroll)
	}
	p = ListPage{Count: 4, Scroll: 3}
	p.Clamp()
	if p.Scroll != 0 || p.Arrows() {
		t.Fatalf("四筆的頁 %+v，要位移 0、沒有 ▲▼", p)
	}
	p = ListPage{Count: 6, Scroll: 5}
	p.Clamp()
	if p.Scroll != 1 {
		t.Fatalf("六筆位移 5 夾成 %d，要 1", p.Scroll)
	}
}
