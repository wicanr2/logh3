package rules_test

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
)

func TestStrategicGuardRawMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	b, e := os.ReadFile(filepath.Join(root, "workplace/assault56-r99-v3-a.json"))
	if os.IsNotExist(e) {
		t.Skip("缺固定dosgolem守軍收據")
	}
	must(e)
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "d22b0e4b93c7283c23c0c26c0e8dca859c487fabb9ffec5e6b5f6ef3ff373653" {
		t.Fatal("原版身份不符")
	}
	var src struct {
		Rows []struct {
			Name          string
			Region        int
			Before, After map[string]string
			SeedAfterHex  string `json:"seed_after_hex"`
		}
	}
	must(json.Unmarshal(b, &src))
	kinds, e := rules.LoadShipKindTable(filepath.Join(root, "docs/internals/data/ship-kind-table.json"))
	must(e)
	slots, e := rules.LoadFormationSlots(filepath.Join(root, "docs/internals/data/battle-formation-slots.json"))
	must(e)
	terrain, e := rules.LoadBattleTerrain(filepath.Join(root, "docs/internals/data/battle-terrain.json"))
	must(e)
	checked := 0
	decode := func(s string) []byte { b, e := hex.DecodeString(s); must(e); return b }
	for _, row := range src.Rows {
		m, e := os.ReadFile(filepath.Join(root, fmt.Sprintf("org_game/extracted/MAP%02d.CMP", row.Region)))
		must(e)
		cells, e := asset.DecodeCMP(m, asset.BattleMapCols, asset.BattleMapRows)
		must(e)
		p, e := rules.LoadPlanets(decode(row.Before["planet"]))
		must(e)
		s := &rules.Slot{Admiral: decode(row.Before["admiral"]), Fleet: decode(row.Before["fleet"]), Ship: decode(row.Before["ship"]), Stellar: decode(row.Before["stellar"]), Planets: p, Work: decode(row.Before["work"])}
		list := decode(row.Before["lists"])
		n := int(binary.LittleEndian.Uint16(list[2:]))
		pools := make([]int, n)
		for i := range pools {
			pools[i] = int(binary.LittleEndian.Uint16(list[38+2*i:]))
		}
		attacker := 1
		for _, f := range s.Fleets(nil) {
			if f.Region == row.Region && f.Faction == 0 && f.Flags&1 == 0 {
				attacker = 0
			}
		}
		rnd := rules.NewRand(0x12345678)
		c, e := rules.NewStrategicOccupation(s, row.Region, attacker, pools, kinds, slots, func(x, y int) bool { return terrain.Passable(cells, x, y) }, rnd)
		must(e)
		for c.Step() != nil {
		}
		diffs := map[string]int{}
		offsets := map[string][]int{}
		for name, got := range map[string][]byte{"admiral": s.Admiral, "fleet": s.Fleet, "ship": s.Ship, "stellar": s.Stellar, "planet": s.Planets.Data, "work": s.Work, "lists": list} {
			want := decode(row.After[name])
			if len(got) != len(want) {
				panic("長度不同")
			}
			diffs[name] = 0
			for i, v := range got {
				if v != want[i] {
					diffs[name]++
					if len(offsets[name]) < 32 {
						offsets[name] = append(offsets[name], i)
					}
				}
			}
		}
		wantSeed := binary.LittleEndian.Uint32(decode(row.SeedAfterHex))
		for name, n := range diffs {
			if n != 0 {
				t.Fatalf("%s原池%s差%d bytes，位移%v", row.Name, name, n, offsets[name])
			}
		}
		if rnd.Seed() != wantSeed {
			t.Fatal("seed不同", row.Name)
		}
		checked++
		t.Log(row.Name, "完整六表／清單／seed零差")
		if row.Name == "near" {
			// 434證實前一戰場池跨入下一星域；戰略入口不可套勝方同星域驗證。
			s.Fleet[6] = 27
			_, err := rules.NewStrategicOccupation(s, 27, 0, pools, kinds, slots, func(int, int) bool { return true }, rules.NewRand(0x12345678))
			if err != nil {
				t.Fatal("保留池被新星域拒絕", err)
			}
			if _, err := rules.NewWinnerOccupation(s, 27, 0, 0, pools, kinds, slots, func(int, int) bool { return true }, rules.NewRand(0x12345678)); err == nil {
				t.Fatal("勝方入口的同星域檢查被放寬")
			}
		}

	}
	if checked != 5 {
		t.Fatal("不是五案", checked)
	}
}
