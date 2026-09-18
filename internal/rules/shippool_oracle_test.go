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
				{"飛彈存量", g.Missiles, dwordAt(rec, shipMissiles)},
				{"艦載機存量", g.Fighters, dwordAt(rec, shipFighters)},
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

// 建立時兩格存量的算式（`2703:14CF`，`docs/re/345`）：飛彈 ＝ 艦種表 `+0x10`
// ×現有艘數、艦載機 ＝ `+0x14`×現有艘數。拿原版執行期的池逐筆驗。
//
// 兩個例外照原版的規則排除，不是為了讓測試過：
//
//   - 艦種 7（索敵分遣隊）走 `2AD2:0029` 的另一條分支，不呼叫 `2703:14CF`；
//   - 艦種 6（修護艦）的兩個係數都是 0，但自家星域補充那一支
//     （`1AAB:1235`，#71）每回合把它的 `+0x0E` 重設成現有艘數——所以存檔
//     裡看到的不是建立值。
func TestInitGroupStockMatchesOraclePool(t *testing.T) {
	root := "../.."
	pool, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-pool-s3/pool.bin"))
	if err != nil {
		t.Skip("沒有收據 fcmd-pool-s3，跳過")
	}
	kinds, err := LoadShipKindTable(filepath.Join(root, "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for i := 0; i < ShipCount && (i+1)*ShipRecord <= len(pool); i++ {
		rec := pool[i*ShipRecord : (i+1)*ShipRecord]
		faction, kind := int(rec[shipFaction]), int(rec[shipKind])
		if rec[shipOwner] == 0xFF || kind == shipKindEmpty {
			continue
		}
		if faction > 1 || kind >= 8 || kind == 6 || kind == shipKindStandalone {
			continue
		}
		g := FleetGroup{
			Kind:  kind,
			Ships: int(rec[shipAlive]) | int(rec[shipAlive+1])<<8,
			Full:  int(rec[shipFull]) | int(rec[shipFull+1])<<8,
		}
		InitGroupStock(kinds, faction, &g)
		if got, want := g.Missiles, dwordAt(rec, shipMissiles); got != want {
			t.Errorf("池 #%d（陣營 %d 艦種 %d）飛彈存量：算出 %d，原版 %d", i, faction, kind, got, want)
		}
		if got, want := g.Fighters, dwordAt(rec, shipFighters); got != want {
			t.Errorf("池 #%d（陣營 %d 艦種 %d）艦載機存量：算出 %d，原版 %d", i, faction, kind, got, want)
		}
		if got, want := g.RepairCap, int(rec[shipRepairCap])|int(rec[shipRepairCap+1])<<8; got != want {
			t.Errorf("池 #%d 修理上限：算出 %d，原版 %d", i, got, want)
		}
		checked++
	}
	if checked < 100 {
		t.Fatalf("只驗到 %d 筆，收據不對？", checked)
	}
	t.Logf("驗了 %d 筆", checked)
}
