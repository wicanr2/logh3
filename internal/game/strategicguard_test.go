package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

type strategicGuardRow struct {
	Name          string
	Region        int
	Before, After map[string]string
	SeedAfterHex  string `json:"seed_after_hex"`
}

func strategicGuardRows(t *testing.T) []strategicGuardRow {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "workplace/assault56-r99-v3-a.json"))
	if os.IsNotExist(err) {
		t.Skip("缺固定dosgolem守軍收據")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "d22b0e4b93c7283c23c0c26c0e8dca859c487fabb9ffec5e6b5f6ef3ff373653" {
		t.Fatal("原版守軍收據身份不符", err)
	}
	var d struct{ Rows []strategicGuardRow }
	if err = json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Rows) != 5 {
		t.Fatal("五案不足")
	}
	return d.Rows
}

func strategicGuardGame(t *testing.T, base *Resources, row strategicGuardRow) *Game {
	t.Helper()
	g := finishGame(t, base, finishGameRow{promotionGameRow: promotionGameRow{Before: row.Before}, Human: 1, Round: 2})
	var err error
	g.res.Battle, err = screen.LoadBattleSpec(filepath.Join(root, "docs/internals/data/screen-battle.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.strategyRand = rules.NewRand(0x12345678)
	// 留下身份，刻意污染不應消費的戰場快取；正式守軍必須重讀原池。
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		u.X, u.Y, u.Ships, u.Faction, u.Out = 1800, 1800, 1, 0, true
	}
	g.battleRegion = 25
	return g
}

func TestStrategicGuardGameOriginal(t *testing.T) {
	base := loadWithRules(t)
	for _, row := range strategicGuardRows(t) {
		t.Run(row.Name, func(t *testing.T) {
			g := strategicGuardGame(t, base, row)
			if _, err := g.strategicAssault(row.Region); err != nil {
				t.Fatal(err)
			}
			if g.rand94DC().Seed() != binary.LittleEndian.Uint32(ownershipHex(t, row.SeedAfterHex)) {
				t.Fatalf("亂數終態不同 %X/%X", g.rand94DC().Seed(), binary.LittleEndian.Uint32(ownershipHex(t, row.SeedAfterHex)))
			}
			start, n := g.res.Planets.Range(row.Region)
			wantPlanet := ownershipHex(t, row.After["planet"])
			for i := start; i < start+n; i++ {
				got, want := g.res.Planets.Record(i), wantPlanet[rules.PlanetHeader+i*rules.PlanetRecord:]
				if got[0x11] != want[0x11] || !bytes.Equal(got[0x14:0x16], want[0x14:0x16]) || !bytes.Equal(got[0x20:0x44], want[0x20:0x44]) {
					t.Fatal("行星歸屬／防衛力／庫存不同", i)
				}
			}
			wantSlot := &rules.Slot{Fleet: ownershipHex(t, row.After["fleet"]), Ship: ownershipHex(t, row.After["ship"])}
			for _, f := range g.fleets {
				if f.Region != row.Region || f.Faction != 0 {
					continue
				}
				w := wantSlot.Fleet[f.Index*rules.FleetRecord:]
				if f.Flags != int(w[0]) || f.Morale != int(w[0x3c]) || f.X != int(binary.LittleEndian.Uint16(w[0x5c:])) || f.Y != int(binary.LittleEndian.Uint16(w[0x5e:])) {
					t.Fatal("攻方旗標／士氣／位置不同", f.Index)
				}
				wc := wantSlot.FleetComposition(f.Index)
				for i, c := range f.Composition {
					if c.Ships != wc[i].Ships || c.X != wc[i].X || c.Y != wc[i].Y {
						t.Fatal("分隊艘數／位置不同", f.Index, i)
					}
				}
			}
			wantAdmiral := ownershipHex(t, row.After["admiral"])
			for i := 0; i < len(wantAdmiral)/rules.AdmiralRecord; i++ {
				if !bytes.Equal(g.admiralRecord(i)[2:6], wantAdmiral[i*rules.AdmiralRecord+2:i*rules.AdmiralRecord+6]) {
					t.Fatal("功績不同", i)
				}
			}
			wantWork := ownershipHex(t, row.After["work"])
			for f := 0; f < 2; f++ {
				if g.battleScore[f] != int(binary.LittleEndian.Uint32(wantWork[0xd0+4*f:])) || g.campaignScore[f] != int(binary.LittleEndian.Uint32(wantWork[0xd8+4*f:])) {
					t.Fatal("分數不同", f)
				}
			}
			for name, got := range map[string][]byte{"admiral": g.res.Admiral, "fleet": g.slot.Fleet, "ship": g.slot.Ship, "stellar": g.res.Stellar, "planet": g.res.Planets.Data, "work": g.workRecord()} {
				if !bytes.Equal(got, ownershipHex(t, row.After[name])) {
					t.Fatal("保存前完整原池不同", name)
				}
			}
			list := ownershipHex(t, row.After["lists"])
			if len(g.battleUnits) != int(binary.LittleEndian.Uint16(list[2:])) {
				t.Fatal("守軍清單筆數不同")
			}
			for i, u := range g.battleUnits {
				if g.battleUnitPool(u) != int(binary.LittleEndian.Uint16(list[38+2*i:])) {
					t.Fatal("守軍清單池序不同")
				}
			}
			// 正式保存消費端與六檔讀回；注入入口不宣稱正常玩家流程。
			g.state = StateSaveGame
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			// 對拍聲明限於上述欄位；完整原池差異另保存，不能由讀回自洽推成原版全同。
			for name, got := range map[string][]byte{"admiral": g.res.Admiral, "fleet": g.slot.Fleet, "ship": g.slot.Ship, "planet": g.res.Planets.Data} {
				want := ownershipHex(t, row.After[name])
				diff := 0
				offsets := []int{}
				for i, v := range got {
					if v != want[i] {
						diff++
						if len(offsets) < 16 {
							offsets = append(offsets, i)
						}
					}
				}
				t.Log("完整原池", name, "不同bytes", diff, "前16位移", offsets)
				if diff != 0 {
					t.Fatal("保存改變原版原池", name)
				}
			}
			read := func(name string) []byte {
				b, e := os.ReadFile(filepath.Join(dir, name))
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			ext := rules.SlotExt(0)
			loaded, err := rules.LoadSlot(rules.SlotFiles{Admiral: read("ADMIRAL.DT" + ext), Fleet: read("FLEET.DT" + ext), Ship: read("SHIP.DT" + ext), Stellar: read("STELLAR.DT" + ext), Work: read("WORK.DT" + ext), PlanetHeader: g.res.Planets.Data[:64], Planet: read("PLANET.DT" + ext)})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(loaded.Ship, g.slot.Ship) || !bytes.Equal(loaded.Fleet, g.slot.Fleet) || !bytes.Equal(loaded.Admiral, g.res.Admiral) || !bytes.Equal(loaded.Planets.Data, g.res.Planets.Data) || loaded.StrategicScore != g.campaignScore {
				t.Fatal("保存讀回丟失守軍戰損／功績／行星／分數")
			}
		})
	}
}

func TestStrategicGuardGameInvalidPoolStops(t *testing.T) {
	rows := strategicGuardRows(t)
	g := strategicGuardGame(t, loadWithRules(t), rows[1])
	g.battleUnits = []rules.BattleUnit{{Fleet: 9999, Squadron: 0}}
	g.rebellionOccupation[26].Occupied = true
	before := bytes.Clone(g.res.Planets.Data)
	seed := g.rand94DC().Seed()
	if !g.scanRegionTail(26) || g.SaveErr() == nil {
		t.Fatal("壞清單被當成0守軍繼續")
	}
	if !bytes.Equal(before, g.res.Planets.Data) || seed != g.rand94DC().Seed() {
		t.Fatal("壞清單仍開火")
	}
}

func TestStrategicGuardEndBattleRetainsPool(t *testing.T) {
	rows := strategicGuardRows(t)
	var row strategicGuardRow
	for _, r := range rows {
		if r.Name == "near" {
			row = r
		}
	}
	g := finishGame(t, loadWithRules(t), finishGameRow{promotionGameRow: promotionGameRow{Before: row.Before}, Human: 1, Round: 2})
	g.strategyRand = rules.NewRand(0x12345678)
	if len(g.battleUnits) != 1 {
		t.Fatal("缺近距池清單")
	}
	pool := g.battleUnitPool(g.battleUnits[0])
	g.EndBattle()
	if g.SaveErr() != nil {
		t.Fatal(g.SaveErr())
	}
	if len(g.battleUnits) != 1 || g.battleUnitPool(g.battleUnits[0]) != pool {
		t.Fatal("戰後續掃清掉或重選池清單")
	}
	guard, err := rules.NewStrategicGuard(g.slot, []int{pool})
	start, _ := g.res.Planets.Range(row.Region)
	if err != nil || guard(g.res.Planets.Record(start+1)) != 30 {
		t.Fatal("戰後原池不再提供守軍", err)
	}
}
