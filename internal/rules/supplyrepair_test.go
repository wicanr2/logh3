package rules

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// 收據 `s3form` 那份 64 KB 傾印裡的艦艇池（線性 `0x668B2` 起，148 筆
// × 78 bytes；傾印從 `0x60000` 開始）。
func loadPool(t *testing.T) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "workplace/oracle/s3form/form_mem.bin")
	b, err := os.ReadFile(p)
	if err != nil || len(b) < 0x668B2-0x60000+148*ShipRecord {
		t.Skip("缺收據 s3form 的 form_mem.bin，跳過（不用代用品）")
	}
	return b[0x668B2-0x60000:]
}

func poolUnit(pool []byte, i int) BattleUnit {
	r := pool[i*ShipRecord:]
	return BattleUnit{
		Faction:   int(r[0]),
		Kind:      int(r[4]),
		Ships:     int(binary.LittleEndian.Uint16(r[8:])),
		FullShips: int(binary.LittleEndian.Uint16(r[10:])),
		RepairCap: int(binary.LittleEndian.Uint16(r[12:])),
		Missiles:  int(binary.LittleEndian.Uint32(r[14:])),
		Fighters:  int(binary.LittleEndian.Uint32(r[18:])),
	}
}

func kindTable(t *testing.T) *ShipKindTable {
	t.Helper()
	tab, err := LoadShipKindTable(filepath.Join("..", "..",
		"docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	return tab
}

// 補給的滿載量 ＝ 艦種列 `+0x10`（飛彈每艘幾發）與 `+0x14`（艦載機滿量）
// × **現有**艘數 `+0x08`。拿收據裡 148 筆真的記錄驗：從一個裝不完的
// 補給艦補一次之後，每一筆都要剛好等於那個乘積。
//
// 收據裡只有六筆是缺的——戰損會把 `+0x0E`／`+0x12` 按「剩下 ÷ 原有」
// 等比例縮小（`docs/internals/38`），所以打過的分隊才會不滿。
func TestSupplyCapacityMatchesPool(t *testing.T) {
	pool, tab := loadPool(t), kindTable(t)
	short, checked := 0, 0
	for i := 0; i < 148; i++ {
		u := poolUnit(pool, i)
		k, ok := tab.Kind(u.Faction, u.Kind)
		if !ok {
			continue
		}
		from := &BattleUnit{Faction: u.Faction, Kind: SupplyKind,
			Missiles: 1 << 24, Fighters: 1 << 24}
		to := u
		m, f := Supply(tab, from, &to)
		// 缺多少 ＝ 公式算出的滿載減掉現有；負的（修理艦的 `+0x0E` 是
		// 修理存量，艦種列 `+0x10` 是 0）就一發都不給。
		wantM := max0(k.Byte(kindMissilePer)*u.Ships - u.Missiles)
		wantF := max0(k.Byte(kindFighterFull)*u.Ships - u.Fighters)
		if m != wantM || f != wantF {
			t.Errorf("第 %d 筆（陣營 %d 艦種 %d，%d 艘）給了 %d／%d，"+
				"公式算出 %d／%d", i, u.Faction, u.Kind, u.Ships, m, f, wantM, wantF)
		}
		if to.Missiles != u.Missiles+m || to.Fighters != u.Fighters+f {
			t.Errorf("第 %d 筆：給了 %d／%d，身上卻變成 %d／%d（原本 %d／%d）",
				i, m, f, to.Missiles, to.Fighters, u.Missiles, u.Fighters)
		}
		// 給出去的量與補給艦身上少掉的量要對得起來。
		if from.Missiles != 1<<24-m || from.Fighters != 1<<24-f {
			t.Errorf("第 %d 筆：給了 %d／%d，補給艦卻少了 %d／%d", i, m, f,
				1<<24-from.Missiles, 1<<24-from.Fighters)
		}
		if m != 0 || f != 0 {
			short++
		}
		checked++
	}
	if checked != 148 {
		t.Fatalf("只驗到 %d 筆，收據裡是 148 筆", checked)
	}
	if short != 6 {
		t.Fatalf("有 %d 筆需要補，收據上是 6 筆", short)
	}
}

// 打過的十六筆：`+0x0A`（建制滿編）與 `+0x0C`（可修復上限）不是同一個
// 東西——三筆的 `+0x0C` 掉到 3999 而 `+0x0A` 還是 4000。修理補到
// `+0x0C`，所以那三筆最多補回 83／68／68 艘，不是 84／69／69。
func TestRepairCapIsNotFullShips(t *testing.T) {
	pool := loadPool(t)
	var loss, cap0 []int
	for i := 0; i < 148; i++ {
		u := poolUnit(pool, i)
		if u.Ships == u.FullShips {
			continue
		}
		loss = append(loss, u.FullShips-u.Ships)
		cap0 = append(cap0, u.RepairCap-u.Ships)
	}
	if len(loss) != 16 {
		t.Fatalf("收據上有 %d 筆掉過艘數，該是 16 筆", len(loss))
	}
	diff := 0
	for i := range loss {
		if loss[i] != cap0[i] {
			diff++
		}
	}
	if diff != 3 {
		t.Fatalf("`+0x0A` 與 `+0x0C` 的差距出現在 %d 筆，收據上是 3 筆", diff)
	}

	// 拿其中一筆真的修一次：修理艦身上有多少就補多少，不超過上限。
	for i := 0; i < 148; i++ {
		u := poolUnit(pool, i)
		if u.RepairCap-u.Ships != 83 {
			continue
		}
		from := &BattleUnit{Kind: RepairKind, Missiles: 20}
		to := u
		if n := Repair(from, &to); n != 20 || to.Ships != u.Ships+20 || from.Missiles != 0 {
			t.Fatalf("存量 20 補了 %d 艘（%d → %d），存量剩 %d",
				n, u.Ships, to.Ships, from.Missiles)
		}
		from = &BattleUnit{Kind: RepairKind, Missiles: 1000}
		to = u
		if n := Repair(from, &to); n != 83 || to.Ships != to.RepairCap {
			t.Fatalf("存量充足時補了 %d 艘，上限是 %d", n, u.RepairCap-u.Ships)
		}
		return
	}
	t.Fatal("收據裡找不到差 83 艘的那一筆")
}

// 作用半徑 ＝ 補給／修理艦種那一列的 `+0x06`（最高速度）× 6.4。
// 兩個陣營的兩個艦種最高速度都是 30，所以半徑都是 192.0。
func TestSupplyRange(t *testing.T) {
	tab := kindTable(t)
	for _, c := range []struct {
		faction, kind int
	}{{0, SupplyKind}, {1, SupplyKind}, {0, RepairKind}, {1, RepairKind}} {
		if r := SupplyListRange(tab, c.faction, c.kind); r != 12582912 {
			t.Errorf("陣營 %d 艦種 %d 收集候選的半徑 %d（%.5f），要 12582912（192.0）",
				c.faction, c.kind, r, float64(r)/FixOne)
		}
		if r := SupplyRange(tab, c.faction, c.kind); r != 12582900 {
			t.Errorf("陣營 %d 艦種 %d 確認時的半徑 %d（%.5f），要 12582900（191.99982）",
				c.faction, c.kind, r, float64(r)/FixOne)
		}
	}
	// 正好 192 像素列得出來，卻按不下去——原版兩個半徑差 12／65536。
	list, ok := SupplyListRange(tab, 1, SupplyKind), SupplyRange(tab, 1, SupplyKind)
	if !InSupplyRange(0, 0, 192, 0, list) {
		t.Error("正好 192 像素該收得進候選")
	}
	if InSupplyRange(0, 0, 192, 0, ok) {
		t.Error("正好 192 像素在確認那一關該被擋掉")
	}
	if !InSupplyRange(0, 0, 191, 0, ok) {
		t.Error("191 像素該過得了確認")
	}
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
