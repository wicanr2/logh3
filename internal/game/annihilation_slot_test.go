package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 戰略全滅要寫回存檔表（#78）。
//
// 存檔是從 `g.slot` 寫出去的（`rules.(*Slot).Save`），`annihilateFleet` 本來只改
// 執行期切片 `g.fleets`，所以那一支在存檔裡還活著，讀回來就復活。
//
// 這是**窄幅診斷**：直接叫 `annihilateFleet`，不走完整的逐格掃描——那一條由
// `TestTurnReportAnnihilatedMatchesOracle` 照收據整張比。這裡只釘存檔表那一側，
// 因為完整路徑那支測試走的是新遊戲（沒有 `g.slot`），碰不到這個欄位。
func TestStrategicAnnihilationWritesBackToSlot(t *testing.T) {
	g := newTestGame(t)
	if g.slot == nil {
		t.Fatal("newTestGame 應該從存檔載入，卻沒有存檔槽")
	}
	if len(g.fleets) == 0 {
		t.Fatal("存檔裡一支艦隊都沒有")
	}
	f := g.fleets[0]
	slot := f.Index
	if !g.slot.FleetUsed(slot) {
		t.Fatalf("全滅之前艦隊表槽 %d 就已經是空的", slot)
	}
	ships := make([]int, 0, len(f.Composition))
	for _, gp := range f.Composition {
		ships = append(ships, gp.Pool)
	}

	g.annihilateFleet(0)

	if g.slot.FleetUsed(slot) {
		t.Errorf("全滅之後艦隊表槽 %d 還在用——存檔之後那一支會復活（#78）", slot)
	}
	for _, fl := range g.fleets {
		if fl.Index == slot {
			t.Errorf("執行期切片裡槽 %d 也還在", slot)
		}
	}
	// 編制那幾筆艦艇要一起作廢（`+0x01` ← FF），否則池子會漏著不還。
	for _, pool := range ships {
		if pool < 0 {
			continue
		}
		rec := g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord]
		if rec[1] != 0xFF {
			t.Errorf("艦艇池 #%d 的 +0x01 ＝ %02X，全滅之後原版是 FF", pool, rec[1])
		}
	}
	if a := g.admiralRecord(f.Admiral); len(a) > 0 && a[0] != 0xFF && a[0x09]&0x80 == 0 {
		t.Errorf("提督 %d 沒有記成失蹤：+00 %02X／+09 %02X", f.Admiral, a[0], a[0x09])
	}
}

// 提督表也要與存檔同步（#78 的「同一類問題不只這一處」）。
//
// `mutableAdmirals()` 第一次改提督時會把 `g.res.Admiral` 換成一份可寫的副本。
// 換了卻不換 `g.slot.Admiral` 的話，**從那一刻起兩份就分家**——存檔寫出去的是
// `g.slot.Admiral`（`rules.(*Slot).Save`），所以之後的失蹤、升級、任命全部漏掉。
func TestAdmiralWritesStayInSlot(t *testing.T) {
	g := newTestGame(t)
	if g.slot == nil || len(g.fleets) == 0 {
		t.Fatal("newTestGame 應該從存檔載入且有艦隊")
	}
	// 真實路徑（`LoadGame`）載存檔時會把兩份接成同一個切片；測試的 helper 沒有，
	// 這裡補上，否則測的就不是同步而是「本來就不同源」。
	g.res.Admiral = g.slot.Admiral

	adm := g.fleets[0].Admiral
	off := adm * rules.AdmiralRecord
	g.annihilateFleet(0)

	if got, want := g.slot.Admiral[off], g.res.Admiral[off]; got != want {
		t.Errorf("提督 %d 的 +00：存檔表 %02X、執行期 %02X——兩份分家了，失蹤不會進存檔（#78）",
			adm, got, want)
	}
	if got, want := g.slot.Admiral[off+0x09], g.res.Admiral[off+0x09]; got != want {
		t.Errorf("提督 %d 的 +09：存檔表 %02X、執行期 %02X", adm, got, want)
	}
}
