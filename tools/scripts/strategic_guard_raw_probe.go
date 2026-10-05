//go:build ignore

package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
)

func main() {
	input := flag.String("input", "", "固定原版")
	output := flag.String("output", "", "新JSON")
	flag.Parse()
	must := func(e error) {
		if e != nil {
			panic(e)
		}
	}
	b, e := os.ReadFile(*input)
	must(e)
	var src struct {
		Rows []struct {
			Name          string
			Region        int
			Before, After map[string]string
			SeedAfterHex  string `json:"seed_after_hex"`
		}
	}
	must(json.Unmarshal(b, &src))
	kinds, e := rules.LoadShipKindTable("docs/internals/data/ship-kind-table.json")
	must(e)
	slots, e := rules.LoadFormationSlots("docs/internals/data/battle-formation-slots.json")
	must(e)
	terrain, e := rules.LoadBattleTerrain("docs/internals/data/battle-terrain.json")
	must(e)
	results := []map[string]any{}
	decode := func(s string) []byte { b, e := hex.DecodeString(s); must(e); return b }
	for _, row := range src.Rows {
		m, e := os.ReadFile(fmt.Sprintf("/orig/MAP%02d.CMP", row.Region))
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
		c, e := rules.NewWinnerOccupation(s, row.Region, attacker, 0, pools, kinds, slots, func(x, y int) bool { return terrain.Passable(cells, x, y) }, rnd)
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
		results = append(results, map[string]any{"name": row.Name, "differences": diffs, "first_offsets": offsets, "seed": rnd.Seed(), "wanted_seed": wantSeed})
	}
	b, e = json.MarshalIndent(map[string]any{"schema": "logh3.strategic-guard-raw-probe.v1", "scope": "容器副本沿原共用1512原池常式，回合0；僅放寬戰場池必屬目前星域的驗證", "rows": results}, "", "  ")
	must(e)
	f, e := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(e)
	_, e = f.Write(append(b, '\n'))
	must(e)
	must(f.Close())
}
