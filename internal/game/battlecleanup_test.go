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
)

type cleanupRow struct {
	Name               string
	Fleet, Count, Mask int
	Before, After      map[string]string
	PromotionBefore    map[string]string `json:"promotion_before"`
	SeedAfter          string            `json:"seed_after_hex"`
}

func cleanupRows(t *testing.T) []cleanupRow {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "workplace/cleanup79-r88-a.json"))
	if os.IsNotExist(err) {
		t.Skip("缺固定原版收據；沿tools/battle_cleanup_entry_oracle.sh及battle_cleanup_oracle.sh重生")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "80f108f6db41137c245dd52040616234779cf93349fbd7c0fae87d1aa2ef9949" {
		t.Fatal("清理原版收據不符", err)
	}
	var d struct{ Rows []cleanupRow }
	if err = json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Rows) != 35 {
		t.Fatal("原版案例不足")
	}
	return d.Rows
}

func cleanupGame(t *testing.T, base *Resources, row cleanupRow) *Game {
	t.Helper()
	g := promotionGame(t, base, promotionGameRow{Case: []int{row.Fleet, 0, 0, 0}, Before: row.Before})
	list := ownershipHex(t, row.Before["lists"])
	var fleets []int
	for _, n := range list[4 : 4+int(binary.LittleEndian.Uint16(list))] {
		fleets = append(fleets, int(n))
	}
	g.battleFleets = g.slot.BattleFleetsFor(fleets)
	g.battleUnits = nil
	for _, index := range fleets {
		g.battleUnits = append(g.battleUnits, g.slot.BattleUnitsFor(base.Weights, *g.fleetByIndex(index))...)
	}
	return g
}

func TestBattleCleanupGameOriginal(t *testing.T) {
	base := loadWithRules(t)
	removed := 0
	for _, row := range cleanupRows(t) {
		t.Run(row.Name, func(t *testing.T) {
			g := cleanupGame(t, base, row)
			before := append([]rules.BattleUnit(nil), g.battleUnits...)
			if g.battleFinalPhase(0, func() { t.Fatal("沒有報告卻額外續接") }) || g.saveErr != nil {
				t.Fatal("清理停住", g.saveErr)
			}
			promotionGameTables(t, g, row.After)
			list := ownershipHex(t, row.After["lists"])
			prep := g.preparationUnits()
			if len(prep) != int(binary.LittleEndian.Uint16(list[2:])) {
				t.Fatal("下一回合補給清單長度不同")
			}
			for j, entry := range prep {
				if entry.Pool != int(binary.LittleEndian.Uint16(list[38+2*j:])) {
					t.Fatal("下一回合補給清單池序與原版不同")
				}
			}
			for _, bf := range g.battleFleets {
				f := g.fleetByIndex(bf.Index)
				pools := g.slot.FleetGroupShips(bf.Index)
				if f.Groups != len(pools) || len(f.Composition) != len(pools) {
					t.Fatal("編制數未同步")
				}
				for j, gp := range f.Composition {
					if gp.Pool != pools[j] {
						t.Fatal("池引用錯位")
					}
				}
			}
			for _, old := range before {
				if old.Squadron >= 0 && old.Ships == 0 {
					removed++
					continue
				}
				found := false
				for _, u := range g.battleUnits {
					if u.Fleet == old.Fleet && u.Slot == old.Slot && u.Kind == old.Kind {
						copy := u
						copy.Squadron = old.Squadron
						if copy != old {
							t.Fatal("清理改了活分隊命令／XY／庫存／額度")
						}
						if u.Squadron >= 0 && g.battleUnitPool(u) != g.slot.FleetGroupShips(u.Fleet)[u.Squadron] {
							t.Fatal("下一回合池索引錯位")
						}
						found = true
						break
					}
				}
				if !found {
					t.Fatal("存活分隊消失")
				}
			}
			if g.battleRand.Seed() != 0x12345678 {
				t.Fatal("清理取了RND")
			}
			if !bytes.Equal(g.res.Planets.Data, ownershipHex(t, row.After["planet"])) {
				t.Fatal("清理改了行星")
			}
		})
	}
	if removed != 77 {
		t.Fatal("移除數不同", removed)
	}
}

func TestBattleCleanupStaleCompositionSaveReload(t *testing.T) {
	base := loadWithRules(t)
	for _, row := range cleanupRows(t) {
		if row.Mask != (1<<row.Count)-2 {
			continue
		}
		t.Run(row.Name, func(t *testing.T) {
			g := cleanupGame(t, base, row)
			// 交戰已改原池與BattleUnit，編制仍持有戰前資源的真實反例。
			for i := range g.fleetByIndex(row.Fleet).Composition {
				gp := &g.fleetByIndex(row.Fleet).Composition[i]
				gp.Ships = gp.Full
				gp.Missiles += 27
				gp.Fighters += 29
				gp.RepairCap += 3
			}
			if g.battleFinalPhase(0, func() {}) || g.saveErr != nil {
				t.Fatal(g.saveErr)
			}
			promotionGameTables(t, g, row.After)
			g.faction = 0
			g.pct, g.foePct = g.slot.BuildPct[0], g.slot.BuildPct[1]
			g.treasury = rules.Treasury{Budget: g.slot.BudgetStart[0], ShipPaid: g.slot.ShipPaid[0]}
			g.state = StateSaveGame // 明示保存診斷，不宣稱經玩家戰場選單。
			dir := t.TempDir()
			if output := os.Getenv("LOGH3_CLEANUP_OUTPUT"); output != "" {
				dir = filepath.Join(output, row.Name+"-save")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
			}
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			read := func(name string) []byte {
				b, e := os.ReadFile(filepath.Join(dir, name+".DT0"))
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			for name, key := range map[string]string{"ADMIRAL": "admiral", "FLEET": "fleet", "SHIP": "ship", "WORK": "work", "STELLAR": "stellar", "PLANET": "planet"} {
				want := ownershipHex(t, row.After[key])
				if name == "PLANET" {
					want = want[64:]
				}
				got := read(name)
				if !bytes.Equal(got, want) {
					for i, v := range got {
						if v != want[i] {
							t.Fatalf("保存%s+%X %02X／原版%02X", name, i, v, want[i])
						}
					}
					t.Fatal("保存長度不符")
				}
			}
			s, err := rules.LoadSlot(rules.SlotFiles{Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Work: read("WORK"), Stellar: read("STELLAR"), Planet: read("PLANET"), PlanetHeader: ownershipHex(t, row.After["planet"])[:64]})
			if err != nil {
				t.Fatal(err)
			}
			g.LoadSlot(s)
			if len(g.fleetByIndex(row.Fleet).Composition) != 1 || g.slot.FleetGroupShips(row.Fleet)[0] != g.fleetByIndex(row.Fleet).Composition[0].Pool || g.battleRand.Seed() != 0x12345678 {
				t.Fatal("重讀編制／seed不符")
			}
		})
	}
}

func TestBattleCleanupRejectsInvalidAtomically(t *testing.T) {
	base := loadWithRules(t)
	row := cleanupRows(t)[1]
	for _, bad := range []string{"pool mismatch", "duplicate pool", "wrong owner", "truncated ship", "missing unit"} {
		t.Run(bad, func(t *testing.T) {
			g := cleanupGame(t, base, row)
			f := g.fleetByIndex(row.Fleet)
			switch bad {
			case "pool mismatch":
				f.Composition[1].Pool++
			case "duplicate pool":
				f.Composition[1].Pool = f.Composition[0].Pool
				binary.LittleEndian.PutUint16(g.slot.Fleet[row.Fleet*122+20:], uint16(f.Composition[0].Pool))
			case "wrong owner":
				g.slot.Ship[f.Composition[1].Pool*78+1] = 255
			case "truncated ship":
				g.slot.Ship = g.slot.Ship[:len(g.slot.Ship)-1]
			case "missing unit":
				for i, u := range g.battleUnits {
					if u.Fleet == row.Fleet && u.Squadron == 1 {
						g.battleUnits = append(g.battleUnits[:i], g.battleUnits[i+1:]...)
						break
					}
				}
			}
			fb, sb := bytes.Clone(g.slot.Fleet), bytes.Clone(g.slot.Ship)
			before, _ := json.Marshal(struct {
				F []rules.Fleet
				U []rules.BattleUnit
			}{g.fleets, g.battleUnits})
			if err := g.cleanupBattleFleet(row.Fleet); err == nil {
				t.Fatal("非法資料未回錯")
			}
			after, _ := json.Marshal(struct {
				F []rules.Fleet
				U []rules.BattleUnit
			}{g.fleets, g.battleUnits})
			if !bytes.Equal(fb, g.slot.Fleet) || !bytes.Equal(sb, g.slot.Ship) || !bytes.Equal(before, after) || g.battleRand.Seed() != 0x12345678 {
				t.Fatal("失敗仍提交部分狀態")
			}
		})
	}
}

func TestBattleCleanupPromotionChainOriginal(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(root, "workplace/cleanup-chain79-r88-a.json"))
	if os.IsNotExist(err) {
		t.Skip("缺原版清理至升級鏈；沿tools/battle_cleanup_chain_oracle.sh重生")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "0e45c28c9cfe015a6db29aa8d6a37a9e51a2dc559a981cf0f3b9a317e822b3db" {
		t.Fatal("原版鏈收據不符", err)
	}
	var data struct{ Rows []cleanupRow }
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 4 {
		t.Fatal("原版鏈案數不同")
	}
	base := loadWithRules(t)
	for _, row := range data.Rows {
		t.Run(row.Name, func(t *testing.T) {
			g := cleanupGame(t, base, row)
			start, n := g.res.Planets.Range(g.battleRegion)
			var sites []rules.PlanetSite
			for i := 0; i < n; i++ {
				r := g.res.Planets.Record(start + i)
				sites = append(sites, rules.PlanetSite{X: int(binary.LittleEndian.Uint16(r[0x46:])), Y: int(binary.LittleEndian.Uint16(r[0x48:]))})
			}
			g.battlePlanets = g.res.Planets.BattlePlanetsFor(g.battleRegion, sites)
			if g.battleFinalPhase(0, func() {}) {
				t.Fatal("清理意外停止", g.saveErr)
			}
			done := 0
			g.battleOwnershipPhase(func() { g.battleScoutFuelPhase(0, func() { done++ }) })
			if done != 1 || g.state != StateBattle {
				t.Fatal("清理至升級入口未續接", g.saveErr)
			}
			promotionGameTables(t, g, row.PromotionBefore)
			done = 0
			g.battlePromotionPhase(0, false, func() { done++ })
			if done != 1 || g.state != StateBattle || g.battleRand.Seed() != binary.LittleEndian.Uint32(ownershipHex(t, row.SeedAfter)) {
				t.Fatal("升級未續接或seed不同", g.saveErr)
			}
			promotionGameTables(t, g, row.After)
			for _, u := range g.battleUnits {
				if u.Squadron < 0 {
					continue
				}
				sh := g.slot.ShipAt(g.battleUnitPool(u))
				if u.Morale != int(sh[6]) || u.TurnSpeed != int(sh[0x1f]) || u.CruiseSpeed != int(sh[0x21]) || u.FightSpeed != int(sh[0x22]) {
					t.Fatal("清理後升級更新了錯誤池")
				}
			}
		})
	}
}
