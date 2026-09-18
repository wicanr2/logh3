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

// 星域層的飛彈／艦載機補給（`1930:1572`，`docs/re/349` §三）：拿同一份收據驗
// **兩支加起來**的結果。原版的順序是 `1930:1572` → `1AAB:0FAF`，所以：
//
//	1930:15F4   把星域存量補進兩格（缺口用補充「前」的艘數算）
//	1AAB:0FAF   補艘數，再照 2703:14CF 加「係數 × 這次補了幾艘」
//
// remake 在缺口為負時夾 0、不照抄原版的無號溢位（使用者裁決 2026-09-19），
// 所以**收據那一筆走的正是溢位分支，這裡只驗不溢位的那一半**：
// 星域存量夠、缺口為正時兩邊該一致。
func TestSupplyRegionFleetsFillsToFull(t *testing.T) {
	root := "../.."
	kinds, err := LoadShipKindTable(filepath.Join(root, "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	k, ok := kinds.Kind(0, 0)
	if !ok {
		t.Fatal("艦種表讀不到帝國艦種 0")
	}
	perMissile, perFighter := k.Byte(kindMissilePer), k.Byte(kindFighterFull)

	rec := make([]byte, RegionRecord)
	rec[regionOwner] = 0
	putLE32(rec[regionMissiles:], 100000)
	putLE32(rec[regionFighters:], 100000)

	// 一組 1000 艘、兩格都空的編制：缺口是正的，補到滿就停。
	f := &Fleet{Faction: 0, Region: 3, Composition: []FleetGroup{{Kind: 0, Ships: 1000}}}
	SupplyRegionFleets(rec, 3, []*Fleet{f}, kinds)

	g := f.Composition[0]
	if want := perMissile * 1000; g.Missiles != want {
		t.Errorf("飛彈補成 %d，該補到滿 %d", g.Missiles, want)
	}
	if want := perFighter * 1000; g.Fighters != want {
		t.Errorf("艦載機補成 %d，該補到滿 %d", g.Fighters, want)
	}
	if want := 100000 - perMissile*1000; int(le32(rec[regionMissiles:])) != want {
		t.Errorf("星域飛彈剩 %d，該剩 %d", le32(rec[regionMissiles:]), want)
	}

	// 再跑一次：已經滿了，不該再動（缺口 0）。
	before := int(le32(rec[regionMissiles:]))
	SupplyRegionFleets(rec, 3, []*Fleet{f}, kinds)
	if got := int(le32(rec[regionMissiles:])); got != before {
		t.Errorf("滿編之後又補了 %d", before-got)
	}

	// 缺口為負（超編）時夾 0——**不照抄原版的無號溢位**（`docs/re/349` §四）。
	f.Composition[0].Missiles = perMissile*1000 + 5000
	before = int(le32(rec[regionMissiles:]))
	SupplyRegionFleets(rec, 3, []*Fleet{f}, kinds)
	if got := int(le32(rec[regionMissiles:])); got != before {
		t.Errorf("超編時動了星域存量 %d → %d；原版會被抽乾，remake 該夾 0 不補", before, got)
	}
}
