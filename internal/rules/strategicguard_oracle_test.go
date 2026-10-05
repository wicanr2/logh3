package rules_test

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
)

type nativeRow struct {
	Name          string
	Region        int
	Before, After map[string]string
	SeedAfterHex  string `json:"seed_after_hex"`
	Calls         []struct {
		Location string
		Args     []uint16
		ReturnAX uint16 `json:"return_ax"`
	}
}

func TestStrategicGuardMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	data, err := os.ReadFile(filepath.Join(root, "workplace/assault56-r99-v3-a.json"))
	if os.IsNotExist(err) {
		t.Skip("缺固定dosgolem守軍收據")
	}
	must(err)
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "d22b0e4b93c7283c23c0c26c0e8dca859c487fabb9ffec5e6b5f6ef3ff373653" {
		t.Fatal("原版收據身份不符")
	}
	var src struct {
		Seed string
		Rows []nativeRow
	}
	must(json.Unmarshal(data, &src))
	if len(src.Rows) != 5 {
		t.Fatal("不是五案")
	}
	seed, err := strconv.ParseUint(src.Seed, 0, 32)
	must(err)
	terrain, err := rules.LoadBattleTerrain(filepath.Join(root, "docs/internals/data/battle-terrain.json"))
	must(err)
	slots, err := rules.LoadFormationSlots(filepath.Join(root, "docs/internals/data/battle-formation-slots.json"))
	must(err)
	decode := func(m map[string]string) map[string][]byte {
		out := map[string][]byte{}
		for k, v := range m {
			b, err := hex.DecodeString(v)
			must(err)
			out[k] = b
		}
		return out
	}
	cases, negative := 0, 0
	for _, row := range src.Rows {
		mapData, err := os.ReadFile(filepath.Join(root, fmt.Sprintf("org_game/extracted/MAP%02d.CMP", row.Region)))
		must(err)
		cells, err := asset.DecodeCMP(mapData, asset.BattleMapCols, asset.BattleMapRows)
		must(err)
		terrainChecks, blocked := 0, 0
		for _, c := range row.Calls {
			if c.Location != "1588:04EA" {
				continue
			}
			ok := terrain.Passable(cells, int(c.Args[0]), int(c.Args[1]))
			if ok != (c.ReturnAX != 0) {
				panic(fmt.Sprintf("%s原版地形返回不符：%v AX=%d", row.Name, c.Args[:2], c.ReturnAX))
			}
			terrainChecks++
			if !ok {
				blocked++
			}
		}
		for _, withTerrain := range []bool{false, true} {
			b, w := decode(row.Before), decode(row.After)
			p, err := rules.LoadPlanets(b["planet"])
			must(err)
			slot := &rules.Slot{Fleet: b["fleet"], Ship: b["ship"], Admiral: b["admiral"], Stellar: b["stellar"], Work: b["work"], Planets: p}
			fleets := slot.Fleets(nil)
			attacker := 1
			for _, f := range fleets {
				if f.Region == row.Region && f.Faction == 0 && f.Flags&1 == 0 {
					attacker = 0
				}
			}
			place := &rules.StrategicPlacement{Slots: slots}
			count := int(binary.LittleEndian.Uint16(b["lists"][2:]))
			if count > (len(b["lists"])-38)/2 {
				t.Fatal("原版池清單越界")
			}
			pools := make([]int, count)
			for i := range pools {
				pools[i] = int(binary.LittleEndian.Uint16(b["lists"][38+2*i:]))
			}
			place.Guard, err = rules.NewStrategicGuard(slot, pools)
			must(err)

			if withTerrain {
				place.Passable = func(x, y int) bool { return terrain.Passable(cells, x, y) }
			}
			var attackers []rules.StrategicFleet
			for i := range fleets {
				f := &fleets[i]
				if f.Region != row.Region {
					continue
				}
				if f.Faction != attacker {
					place.Others = append(place.Others, f)
					continue
				}
				rec := b["fleet"][f.Index*rules.FleetRecord:]
				attackers = append(attackers, rules.StrategicFleet{Fleet: f, Ability40: int(rec[0x40]), Ability41: int(rec[0x41]), Command: int(b["admiral"][f.Admiral*rules.AdmiralRecord+0xb])})
			}
			start, n := p.Range(row.Region)
			lo := rules.PlanetHeader + start*rules.PlanetRecord
			planets := p.Data[lo : lo+n*rules.PlanetRecord]
			sites := make([]rules.PlanetSite, n)
			for i := range sites {
				p := planets[i*rules.PlanetRecord:]
				sites[i] = rules.PlanetSite{X: int(binary.LittleEndian.Uint16(p[0x46:])), Y: int(binary.LittleEndian.Uint16(p[0x48:]))}
			}
			rnd := rules.NewRand(uint32(seed))
			res := rules.StrategicAssault(planets, sites, attacker, attackers, rnd, place)
			diffs := []string{}
			check := func(label string, got, want int64) {
				if got != want {
					diffs = append(diffs, fmt.Sprintf("%s %d/%d", label, got, want))
				}
			}
			seedAfter, err := hex.DecodeString(row.SeedAfterHex)
			must(err)
			if len(seedAfter) != 4 {
				panic("原版seed不是四bytes")
			}
			check("seed", int64(rnd.Seed()), int64(binary.LittleEndian.Uint32(seedAfter)))
			for i := 0; i < n; i++ {
				got, want := planets[i*rules.PlanetRecord:], w["planet"][lo+i*rules.PlanetRecord:]
				for _, off := range append([]int{0x11, 0x14, 0x15}, byteRange(0x20, 0x44)...) {
					check(fmt.Sprintf("planet%d+%02X", i, off), int64(got[off]), int64(want[off]))
				}
			}
			wantSlot := &rules.Slot{Fleet: w["fleet"], Ship: w["ship"]}
			for _, sf := range attackers {
				f := sf.Fleet
				want := w["fleet"][f.Index*rules.FleetRecord:]
				for _, v := range []struct {
					name      string
					got, want int
				}{{"flags", f.Flags, int(want[0])}, {"morale", f.Morale, int(want[0x3c])}, {"target", f.Target, int(want[7])}, {"action", f.Action, int(want[0x48])}, {"x", f.X, int(binary.LittleEndian.Uint16(want[0x5c:]))}, {"y", f.Y, int(binary.LittleEndian.Uint16(want[0x5e:]))}} {
					check(fmt.Sprintf("fleet%d.%s", f.Index, v.name), int64(v.got), int64(v.want))
				}
				wc := wantSlot.FleetComposition(f.Index)
				for i, g := range f.Composition {
					for _, v := range []struct {
						name      string
						got, want int
					}{{"ships", g.Ships, wc[i].Ships}, {"x", g.X, wc[i].X}, {"y", g.Y, wc[i].Y}} {
						check(fmt.Sprintf("fleet%d.group%d.%s", f.Index, i, v.name), int64(v.got), int64(v.want))
					}
				}
			}
			for adm, merit := range res.Merit {
				off := adm*rules.AdmiralRecord + 2
				check(fmt.Sprintf("admiral%d.merit", adm), int64(binary.LittleEndian.Uint32(b["admiral"][off:]))+int64(merit), int64(binary.LittleEndian.Uint32(w["admiral"][off:])))
			}
			for f := 0; f < 2; f++ {
				for _, base := range []int{0xd0, 0xd8} {
					off := base + f*4
					check(fmt.Sprintf("work+%02X", off), int64(binary.LittleEndian.Uint32(b["work"][off:]))+int64(res.Score[f]), int64(binary.LittleEndian.Uint32(w["work"][off:])))
				}
			}
			cases++
			if withTerrain && len(diffs) > 0 {
				t.Fatalf("%s正式守軍差異：%v", row.Name, diffs)
			}
			if !withTerrain && row.Name == "terrain" {
				if len(diffs) == 0 {
					t.Fatal("缺地形反例沒有失敗")
				}
				negative++
				t.Log("缺地形反例差異", len(diffs))
			}
			t.Log(row.Name, withTerrain, "差異", len(diffs), "地形查詢", terrainChecks, "拒絕", blocked)

		}
	}
	if cases != 10 || negative != 1 {
		t.Fatal("五案／地形反例不足", cases, negative)
	}
}

func byteRange(a, b int) []int {
	out := []int{}
	for i := a; i < b; i++ {
		out = append(out, i)
	}
	return out
}
