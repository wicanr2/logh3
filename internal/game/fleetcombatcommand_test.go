package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 明示第65輪受控入口資料，非正常玩家交戰完成證據。
func commandDataFixture(t *testing.T) (*Game, int, int) {
	t.Helper()
	g := newTestGame(t)
	m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/weapon79-r65-near-a/input-mem.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺第65輪原版候選RAM")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1048576 || fmt.Sprintf("%x", sha256.Sum256(m)) != "b0b72e8a6d92a31b6dec3dae6b0016f3aad976fc216fb9b5ebf5187b94d16be1" {
		t.Fatal("第65輪near原版RAM身份不符")
	}
	g.slot.Fleet = bytes.Clone(m[0x60ba4 : 0x60ba4+rules.FleetCount*rules.FleetRecord])
	g.slot.Ship = bytes.Clone(m[0x668b2 : 0x668b2+rules.ShipCount*rules.ShipRecord])
	g.fleets = g.slot.Fleets(g.res.Weights)
	if g.res.ShipKinds, err = rules.LoadShipKindTable(filepath.Join(root, "docs/internals/data/ship-kind-table.json")); err != nil {
		t.Fatal(err)
	}
	g.battleFleets = g.slot.BattleFleetsFor([]int{12, 16, 17, 28, 29, 30, 31})
	byPool := map[int]rules.BattleUnit{}
	for _, f := range g.fleets {
		for _, u := range g.slot.BattleUnitsFor(g.res.Weights, f) {
			byPool[g.battleUnitPool(u)] = u
		}
		if u, ok := g.slot.ScoutUnit(g.res.Weights, f, 0, 0); ok {
			p := f.Scout * rules.ShipRecord
			u.X = int(binary.LittleEndian.Uint16(g.slot.Ship[p+0x34:]))
			u.Y = int(binary.LittleEndian.Uint16(g.slot.Ship[p+0x36:]))
			byPool[f.Scout] = u
		}
	}
	own, target := -1, -1
	for i := 0; i < int(binary.LittleEndian.Uint16(m[0x54b30+0x9896:])); i++ {
		pool := int(binary.LittleEndian.Uint16(m[0x54b30+0x98ba+i*2:]))
		u, ok := byPool[pool]
		if !ok {
			t.Fatalf("原版物件池%d沒有typed來源", pool)
		}
		if pool == 80 {
			own = i
		}
		if pool == 102 {
			target = i
		}
		g.battleUnits = append(g.battleUnits, u)
	}
	if own < 0 || target < 0 {
		t.Fatal("原版攻守池缺失")
	}
	g.battleRegion, g.battleTurn, g.faction, g.state = 26, 1, 0, StateBattle
	g.battleRand = rules.NewRand(0x12345678)
	return g, own, target
}

func TestFleetCombatCommandDataAndSave(t *testing.T) {
	g, own, target := commandDataFixture(t)
	beforeUnits := append([]rules.BattleUnit(nil), g.battleUnits...)
	before := bytes.Clone(g.slot.Ship)
	candidates, err := g.fleetCombatCandidates(own)
	if err != nil || len(candidates) != 1 || candidates[0].Target != target || candidates[0].Mask != 7 {
		t.Fatalf("正式Game原版near候選不符：%v %v", candidates, err)
	}
	for _, weapon := range []int{0, 1, 2, 255, 3} {
		t.Run(fmt.Sprintf("weapon%d", weapon), func(t *testing.T) {
			g.slot.Ship = bytes.Clone(before)
			if err := g.commitFleetCombatCommand(own, target, weapon); err != nil {
				t.Fatal(err)
			}
			pool, stored, err := g.fleetCombatCommand(own)
			wantPool := int(binary.LittleEndian.Uint16(before[80*rules.ShipRecord+0x2c:]))
			if weapon < 3 {
				wantPool = 102
			}
			if err != nil || pool != wantPool || int(stored) != weapon {
				t.Fatal("命令目標／武器沒有保存在原池或FF／3改寫目標")
			}
			for i, b := range g.slot.Ship {
				o := i - 80*rules.ShipRecord
				if b != before[i] && o != 0x2c && o != 0x2d && o != 0x2e {
					t.Fatalf("提交意外改寫SHIP+%X", i)
				}
			}
			if !reflect.DeepEqual(g.battleUnits, beforeUnits) || g.battleRand.Seed() != 0x12345678 {
				t.Fatal("提交消費庫存、變動分隊或RND")
			}
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			read := func(name string) []byte {
				b, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			loaded, err := rules.LoadSlot(rules.SlotFiles{Stellar: read("STELLAR.DT0"), Planet: read("PLANET.DT0"),
				Admiral: read("ADMIRAL.DT0"), Fleet: read("FLEET.DT0"), Ship: read("SHIP.DT0"), Work: read("WORK.DT0"),
				PlanetHeader: g.slot.Planets.Data[:rules.PlanetHeader]})
			if err != nil {
				t.Fatal(err)
			}
			fresh := &Game{slot: loaded, res: g.res, fleets: loaded.Fleets(g.res.Weights), battleUnits: g.battleUnits}
			pool2, weapon2, err := fresh.fleetCombatCommand(own)
			if err != nil || pool2 != pool || weapon2 != stored {
				t.Fatal("SaveSlot與正式Slot重讀遺失交戰命令")
			}
		})
	}
	// 已有目標的取消／防禦亦不清word，不能只測初值FFFF。
	if err := g.commitFleetCombatCommand(own, target, 0); err != nil {
		t.Fatal(err)
	}
	for _, weapon := range []int{255, 3} {
		if err := g.commitFleetCombatCommand(own, -1, weapon); err != nil {
			t.Fatal(err)
		}
		if pool, _, err := g.fleetCombatCommand(own); err != nil || pool != 102 {
			t.Fatal("取消或防禦清除了既有目標池")
		}
	}
}

func TestFleetCombatCommandRejectedWithoutMutation(t *testing.T) {
	for _, name := range []string{"weapon", "target", "stock", "moved", "terrain", "map", "action", "scout", "pool", "cpu"} {
		t.Run(name, func(t *testing.T) {
			g, own, target := commandDataFixture(t)
			weapon := 1
			switch name {
			case "weapon":
				weapon = 4
			case "target":
				target = -1
			case "stock":
				g.battleUnits[own].Missiles = 0
			case "moved":
				g.battleUnits[target].X += 301
			case "terrain":
				g.res.Terrain = nil
			case "map":
				g.res.BattleMaps[g.battleRegion] = nil
			case "action":
				g.battleUnits[own].Action = 0
			case "scout":
				g.battleUnits[own].Squadron = -1
			case "pool":
				g.slot.Ship = g.slot.Ship[:rules.ShipRecord]
			case "cpu":
				g.faction = 1
			}
			before, units := bytes.Clone(g.slot.Ship), append([]rules.BattleUnit(nil), g.battleUnits...)
			if err := g.commitFleetCombatCommand(own, target, weapon); err == nil {
				t.Fatal("非法命令被提交")
			}
			if !bytes.Equal(before, g.slot.Ship) || !reflect.DeepEqual(units, g.battleUnits) || g.battleRand.Seed() != 0x12345678 {
				t.Fatal("錯誤命令變動資料或RND")
			}
		})
	}
}
