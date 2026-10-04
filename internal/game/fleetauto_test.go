package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

type autoGameReceipt struct {
	Files map[string]struct {
		SHA string `json:"sha256"`
	} `json:"files"`
	Cases []struct {
		Name string `json:"case"`
	} `json:"cases"`
}

func autoGameManifest(t *testing.T) autoGameReceipt {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/fleet-auto-command.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "23627c23a9d2ef93991c842e00fbf43df75f00d2043c78788454e868d3a19401" {
		t.Fatal("第75輪自動命令收據身份不符", err)
	}
	var m autoGameReceipt
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func (m autoGameReceipt) read(t *testing.T, path string) []byte {
	t.Helper()
	s, ok := m.Files[path]
	if !ok {
		t.Fatal("自動命令來源未索引", path)
	}
	b, err := os.ReadFile(filepath.Join(root, path))
	if os.IsNotExist(err) {
		t.Skip("缺原版自動命令來源", path)
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != s.SHA {
		t.Fatal("自動命令來源身份不符", path, err)
	}
	return b
}

func autoGameFixture(t *testing.T, m autoGameReceipt, name string) (*Game, int, []byte) {
	t.Helper()
	g, _, _ := commandDataFixture(t)
	dir := "workplace/oracle/command79-r75-" + name + "-a/"
	ram, end := m.read(t, dir+"input-mem.bin"), m.read(t, dir+"end-mem.bin")
	if len(ram) != 1048576 || len(end) != 1048576 {
		t.Fatal("自動命令原版RAM不完整")
	}
	g.slot.Ship = bytes.Clone(ram[0x668b2 : 0x668b2+rules.ShipCount*rules.ShipRecord])
	g.slot.Fleet = bytes.Clone(ram[0x60ba4 : 0x60ba4+rules.FleetCount*rules.FleetRecord])
	g.fleets = g.slot.Fleets(g.res.Weights)
	g.battleFleets = g.slot.BattleFleetsFor([]int{12, 16, 17, 28, 29, 30, 31})
	if k, err := rules.LoadShipKindTable(filepath.Join(root, "docs/internals/data/ship-kind-table.json")); err != nil {
		t.Fatal(err)
	} else {
		g.res.ShipKinds = k
	}
	for side := range g.res.ShipKinds.Factions {
		for i := range g.res.ShipKinds.Factions[side].Kinds {
			k := &g.res.ShipKinds.Factions[side].Kinds[i]
			for j := range k.Words {
				k.Words[j] = int(binary.LittleEndian.Uint16(ram[0x54b30+0x4e3a+side*208+k.Kind*26+j*2:]))
			}
		}
	}
	own := -1
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		pool := g.battleUnitPool(*u)
		r := g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord]
		u.Kind, u.Detail, u.Action, u.Out = int(r[4]), int(r[0x17]), int(r[0x1e]), r[0x1e]&128 != 0
		u.Ships = int(binary.LittleEndian.Uint16(r[8:]))
		u.Missiles, u.Fighters = int(binary.LittleEndian.Uint32(r[14:])), int(binary.LittleEndian.Uint32(r[18:]))
		u.X, u.Y = int(binary.LittleEndian.Uint16(r[0x34:])), int(binary.LittleEndian.Uint16(r[0x36:]))
		if pool == 80 {
			own = i
		}
	}
	if own < 0 {
		t.Fatal("自動命令原池80沒有typed分隊")
	}
	g.battleRand = rules.NewRand(binary.LittleEndian.Uint32(ram[0x5e00c:]))
	if strings.HasPrefix(name, "computer-") {
		g.faction, g.originalBattleAI = 1, true // 明示原版CPU診斷；不改061預設。
	}
	return g, own, end
}

func TestFleetAutoCommandGameAndSave(t *testing.T) {
	m := autoGameManifest(t)
	for _, c := range m.Cases {
		// 早期局部RE明示把pool103偵察隊改一般組，或把旗艦+03改1但不改編制。
		// 純規則照原bytes核對；Game另用general案例保持真正池身分，不猜映射。
		if c.Name == "computer-kind0" || c.Name == "computer-kind5" || c.Name == "computer-kind6" ||
			strings.HasPrefix(c.Name, "three-") || c.Name == "two-low" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			g, own, end := autoGameFixture(t, m, c.Name)
			before, units := bytes.Clone(g.slot.Ship), append([]rules.BattleUnit(nil), g.battleUnits...)
			selected, err := g.selectFleetAutoCombatCommand(own)
			if err != nil {
				t.Fatal(err)
			}
			a := 80 * rules.ShipRecord
			for _, offset := range []int{0x2c, 0x2d, 0x2e, 0x3e, 0x3f} {
				if g.slot.Ship[a+offset] != end[0x668b2+a+offset] {
					t.Fatalf("原池80+%X與原版不同", offset)
				}
			}
			allowed := map[int]bool{a + 0x2c: true, a + 0x2d: true, a + 0x2e: true, a + 0x3e: true, a + 0x3f: true}
			for _, e := range selected.Evaluations {
				pool := g.battleUnitPool(g.battleUnits[e.Target])
				base := pool * rules.ShipRecord
				for _, off := range []int{0x3c, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d} {
					allowed[base+off] = true
					if g.slot.Ship[base+off] != end[0x668b2+base+off] {
						t.Fatalf("原池%d+%X預估不同", pool, off)
					}
				}
			}
			for i, b := range g.slot.Ship {
				if b != before[i] && !allowed[i] {
					t.Fatalf("自動命令意外改SHIP+%X", i)
				}
			}
			if !reflect.DeepEqual(units, g.battleUnits) || g.battleRand.Seed() != binary.LittleEndian.Uint32(end[0x5e00c:]) {
				t.Fatal("自動命令改typed庫存或RND不符")
			}
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			read := func(name string) []byte {
				b, err := os.ReadFile(filepath.Join(dir, name+".DT0"))
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			loaded, err := rules.LoadSlot(rules.SlotFiles{Stellar: read("STELLAR"), Planet: read("PLANET"),
				Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Work: read("WORK"),
				PlanetHeader: g.slot.Planets.Data[:rules.PlanetHeader]})
			if err != nil {
				t.Fatal(err)
			}
			for i := range allowed {
				if loaded.Ship[i] != g.slot.Ship[i] {
					t.Fatalf("保存重讀遺失原池+%X", i)
				}
			}
		})
	}
}

func TestFleetAutoCommandVersusHumanControl(t *testing.T) {
	m := autoGameManifest(t)
	g, own, end := autoGameFixture(t, m, "base")
	g.versus, g.faction = true, 1
	selected, err := g.selectFleetAutoCombatCommand(own)
	if err != nil || len(selected.Evaluations) != 1 || int(selected.Evaluations[0].Score) !=
		int(binary.LittleEndian.Uint16(end[0x668b2+102*rules.ShipRecord+0x46:])) {
		t.Fatal("雙人另一陣營被誤當電腦權重", err)
	}
}

func TestFleetAutoCommandGameRejectsWithoutMutation(t *testing.T) {
	m := autoGameManifest(t)
	for _, name := range []string{"cpu", "rand", "fleet", "ability", "target-weapon", "action"} {
		t.Run(name, func(t *testing.T) {
			g, own, _ := autoGameFixture(t, m, "base")
			seed := g.battleRand.Seed()
			switch name {
			case "cpu":
				g.faction = 1
			case "rand":
				g.battleRand = nil
			case "fleet":
				g.slot.Fleet = g.slot.Fleet[:rules.FleetRecord]
			case "ability":
				g.slot.Fleet[16*rules.FleetRecord+0x3e] = 0
			case "target-weapon":
				g.slot.Ship[102*rules.ShipRecord+0x2e] = 4
			case "action":
				g.battleUnits[own].Action = 0
			}
			before, units := bytes.Clone(g.slot.Ship), append([]rules.BattleUnit(nil), g.battleUnits...)
			if _, err := g.selectFleetAutoCombatCommand(own); err == nil {
				t.Fatal("錯誤自動命令被提交")
			}
			if !bytes.Equal(before, g.slot.Ship) || !reflect.DeepEqual(units, g.battleUnits) ||
				g.battleRand != nil && g.battleRand.Seed() != seed {
				t.Fatal("拒絕分支改資料或消費RND")
			}
		})
	}
}
