package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// 自家星域補充的數字（#71，規格 063、`docs/re/346` §六）：拿原版補充前後的
// 艦艇池與星域記錄（收據 `fcmd-supply-mem`）驗算式。
//
// 局面是 poke 造的：帝國第 1 艦隊旗艦那一組（艦種 0）艘數壓到 100、滿編 3000，
// 首都星域那個艦種的存量 75000。所以 `給 ＝ min(3000−100, 75000) ＝ 2900`。
//
// **兩格存量不在這裡比**：原版那兩格各多一塊 108600，那是星域層的飛彈／艦載機
// 被整批補給進來的，不屬於這一支（`docs/re/345` 末節）。
func TestResupplyNumbersMatchOracle(t *testing.T) {
	root := "../.."
	dir := filepath.Join(root, "workplace/oracle/fcmd-supply-mem")
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skip("沒有收據 fcmd-supply-mem，跳過")
		}
		return b
	}
	poolBefore, poolAfter := read("pool-before.bin"), read("pool-after.bin")
	regBefore, regAfter := read("region-before.bin"), read("region-after.bin")
	if len(poolAfter) < ShipRecord || len(regAfter) < RegionRecord {
		t.Fatal("收據太短")
	}
	word := func(b []byte, o int) int { return int(b[o]) | int(b[o+1])<<8 }

	rec := poolBefore[:ShipRecord]
	faction, kind := int(rec[shipFaction]), int(rec[shipKind])
	g := FleetGroup{
		Kind:      kind,
		Ships:     word(rec, shipAlive),
		Full:      word(rec, shipFull),
		RepairCap: word(rec, shipRepairCap),
		Missiles:  dwordAt(rec, shipMissiles),
		Fighters:  dwordAt(rec, shipFighters),
	}
	stock := regionShipStock(regBefore[:RegionRecord], faction, kind)
	need := g.Full - g.Ships
	give := need
	if stock < give {
		give = stock
	}
	if give != 2900 {
		t.Fatalf("算出來要補 %d 艘，收據的局面是 2900（滿編 %d、現有 %d、存量 %d）",
			give, g.Full, g.Ships, stock)
	}
	kinds, err := LoadShipKindTable(filepath.Join(root, "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	addGroupShips(kinds, faction, &g, give)

	after := poolAfter[:ShipRecord]
	if got, want := g.Ships, word(after, shipAlive); got != want {
		t.Errorf("補完現有艘數 %d，原版 %d", got, want)
	}
	if got, want := g.RepairCap, word(after, shipRepairCap); got != want {
		t.Errorf("補完修理上限 %d，原版 %d", got, want)
	}
	if got, want := stock-give, regionShipStock(regAfter[:RegionRecord], faction, kind); got != want {
		t.Errorf("補完星域存量 %d，原版 %d", got, want)
	}
	// 那兩格的差就是「不屬於這一支」的那一塊，兩格應該一樣多。
	extra1 := dwordAt(after, shipMissiles) - g.Missiles
	extra2 := dwordAt(after, shipFighters) - g.Fighters
	if extra1 != extra2 {
		t.Errorf("兩格多出來的量不同（%d／%d）——`docs/re/345` 末節的假設是同一塊星域層存量", extra1, extra2)
	}
	if extra1 != regionSupplyAt(regAfter[:RegionRecord], regionMissiles, faction) {
		t.Errorf("多出來的 %d 與星域的飛彈存量對不上", extra1)
	}
}

// regionSupplyAt 回星域記錄某一格 32 位元存量（飛彈 `+0x68`／艦載機 `+0x70`，各陣營一份）。
func regionSupplyAt(record []byte, off, faction int) int {
	o := off + faction*4
	if o+4 > len(record) {
		return 0
	}
	return int(le32(record[o:]))
}
