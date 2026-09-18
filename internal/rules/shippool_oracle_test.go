package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// 艦艇池那一筆的欄位（`docs/internals/64`、`docs/re/344`）：拿**原版執行期的池**
// （收據 `fcmd-pool-s3` 的 `pool.bin`，S3 開局從 `ds:CFB4` 倒出來的 352×78）
// 與 remake 從同一格存檔解出來的 `FleetGroup` 逐欄比。
//
// 這不是循環驗證：左邊是原版跑起來之後記憶體裡的樣子，右邊是 remake 自己
// 照欄位位移解出來的值。位移錯一格就對不上。
func TestFleetCompositionMatchesOraclePool(t *testing.T) {
	root := "../.."
	pool, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-pool-s3/pool.bin"))
	if err != nil {
		t.Skip("沒有收據 fcmd-pool-s3，跳過")
	}
	orig := filepath.Join(root, "org_game/extracted")
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(orig, name))
		if err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", name)
		}
		return b
	}
	s, err := LoadSlot(SlotFiles{
		Stellar: read("STELLAR.DTA"), Planet: read("PLANET.DTA"),
		Admiral: read("ADMIRAL.DTA"), Fleet: read("FLEET.DTA"),
		Ship: read("SHIP.DTA"), Work: read("WORK.DTA"),
		PlanetHeader: read("PLADAT.DAT"),
	})
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for i := 0; i < FleetCount; i++ {
		if !s.FleetUsed(i) {
			continue
		}
		for _, g := range s.FleetComposition(i) {
			if (g.Pool+1)*ShipRecord > len(pool) {
				t.Fatalf("槽 %d 的池索引 %d 超出收據的池", i, g.Pool)
			}
			rec := pool[g.Pool*ShipRecord : (g.Pool+1)*ShipRecord]
			for _, c := range []struct {
				name string
				got  int
				want int
			}{
				{"艦種", g.Kind, int(rec[shipKind])},
				{"現有艘數", g.Ships, int(rec[shipAlive]) | int(rec[shipAlive+1])<<8},
				{"滿編", g.Full, int(rec[shipFull]) | int(rec[shipFull+1])<<8},
				{"士氣", g.Morale, int(rec[shipMorale])},
				{"修理上限", g.RepairCap, int(rec[shipRepairCap]) | int(rec[shipRepairCap+1])<<8},
				{"存量1", g.Stock1, dwordAt(rec, shipStock1)},
				{"存量2", g.Stock2, dwordAt(rec, shipStock2)},
				{"索敵等級", g.Detail, int(rec[shipDetail])},
				{"隊形槽", g.Slot3, int(rec[shipSlot])},
				{"擁有者", i, int(rec[shipOwner])},
			} {
				if c.got != c.want {
					t.Errorf("槽 %d 池索引 %d 的%s ＝ %d，原版池是 %d",
						i, g.Pool, c.name, c.got, c.want)
				}
			}
			checked++
		}
	}
	if checked < 20 {
		t.Fatalf("只比到 %d 組，S3 開局應該有幾十組——收據或存檔挑錯了", checked)
	}
	t.Logf("逐欄比過 %d 組", checked)
}
