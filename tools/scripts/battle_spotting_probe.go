//go:build ignore

// 可丟棄第三階段狀態模型。只用於087的證據審查，不接正式玩家路徑。
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/wicanr2/logh3/internal/rules"
)

type record struct {
	Index int
	Hex   string
}
type frame struct {
	Name                   string
	Units, Fleets, Planets []record
	Seats, Controllers     []int
}
type scenario struct {
	Scenario string
	Frames   []frame
}
type report struct {
	Source string `json:"source"`
	Index  int    `json:"index"`
	Pools  []int  `json:"pools"`
}

func fail(err error) {
	if err != nil {
		panic(err)
	}
}
func raw(r record) []byte   { b, err := hex.DecodeString(r.Hex); fail(err); return b }
func w(b []byte, n int) int { return int(binary.LittleEndian.Uint16(b[n:])) }
func typed(b []byte) rules.BattleUnit {
	return rules.BattleUnit{Faction: int(b[0]), Fleet: int(b[1]), Squadron: int(b[3]), Kind: int(b[4]),
		Out: b[0x1E]&0x80 != 0, X: w(b, 0x34), Y: w(b, 0x36), TargetX: w(b, 0x28), TargetY: w(b, 0x2A),
		Seen: b[0x16]&1 != 0, Detail: int(b[0x17]), DetailHold: int(b[0x18])}
}

func main() {
	b, err := os.ReadFile("docs/internals/data/battle-spotting-phase.json")
	fail(err)
	var data struct{ Cases []scenario }
	fail(json.Unmarshal(b, &data))
	results := []any{}
	for _, c := range data.Cases {
		first := c.Frames[0]
		current := map[int][]byte{}
		units := []rules.BattleUnit{}
		visiblePools := []int{}
		fleets := map[int][]byte{}
		for _, r := range first.Fleets {
			fleets[r.Index] = raw(r)
		}
		for _, r := range first.Units {
			current[r.Index] = raw(r)
			p := current[r.Index]
			if p[4] == 7 && fleets[int(p[1])][58] == 255 {
				continue
			}
			units = append(units, typed(p))
			visiblePools = append(visiblePools, r.Index)
		}
		var planets []rules.BattlePlanet
		for _, r := range first.Planets {
			p := raw(r)
			planets = append(planets, rules.BattlePlanet{Owner: int(p[17]), X: w(p, 70), Y: w(p, 72), Defence: w(p, 20)})
		}
		// 真實入口一次轉成現行產品規則，不反推preDetail，不以輸出填輸入。
		rules.UpdateSpotting(units, planets)
		var existingDifferences []any
		wants := map[int][]byte{}
		for _, r := range c.Frames[len(c.Frames)-1].Units {
			wants[r.Index] = raw(r)
		}
		for i, pool := range visiblePools {
			want, got := typed(wants[pool]), units[i]
			if want.Detail != got.Detail || want.Seen != got.Seen || want.DetailHold != got.DetailHold {
				existingDifferences = append(existingDifferences, map[string]any{"pool": pool,
					"want": []any{want.Seen, want.Detail, want.DetailHold}, "got": []any{got.Seen, got.Detail, got.DetailHold}})
			}
		}
		var reports []report
		observe := func(source string, index, faction, x, y, radius int) {
			var changed []int
			for _, r := range first.Units {
				p := current[r.Index]
				if int(p[0]) != 1-faction || p[0x1E]&0x80 != 0 || rules.Distance16(x, y, w(p, 52), w(p, 54)) > int64(radius)*rules.FixOne || p[22]&1 != 0 {
					continue
				}
				p[22] |= 1
				if p[23] == 3 {
					continue
				}
				p[23]++
				if p[23] == 3 {
					p[24] = 2
				}
				changed = append(changed, r.Index)
			}
			if len(changed) == 0 {
				return
			}
			human := false
			for seat, f := range first.Seats {
				if f == faction && first.Controllers[seat] == 0 {
					human = true
				}
			}
			if !human {
				return
			}
			sort.Slice(changed, func(i, j int) bool {
				a, b := current[changed[i]], current[changed[j]]
				if a[1] != b[1] {
					return a[1] < b[1]
				}
				return a[3] < b[3]
			})
			reports = append(reports, report{source, index, changed})
		}
		// 窄模型直接使用122-byte艦隊的觀測者索引與3A狀態，不從Squadron猜旗艦。
		for _, faction := range first.Seats {
			for _, r := range first.Fleets {
				f := raw(r)
				if int(f[1]) != faction {
					continue
				}
				indices := []int{w(f, 16)}
				if f[58] != 255 {
					indices = append(indices, w(f, 56))
				}
				for _, pool := range indices {
					p := current[pool]
					if p == nil {
						panic("缺觀測者池")
					}
					radius := 200
					if p[4] == 7 {
						radius = 250
					}
					observe("fleet", r.Index, faction, w(p, 52), w(p, 54), radius)
				}
			}
			for _, r := range first.Planets {
				p := raw(r)
				if int(p[17]) == faction {
					observe("planet", r.Index, faction, w(p, 70), w(p, 72), rules.PlanetQueryRadius(w(p, 20)))
				}
			}
		}
		// S1沒有任何變更；S3只有一份8筆行星報告。若更多報告，拒絕套用此有限模型。
		if c.Scenario == "s1" && len(reports) != 0 || c.Scenario == "s3" && (len(reports) != 1 || len(reports[0].Pools) != 8) {
			panic("報告分支超出已量測模型")
		}
		checked := 0
		for _, f := range c.Frames {
			got := current
			if f.Name == "entry" || f.Name == "label" {
				got = map[int][]byte{}
				for _, r := range first.Units {
					got[r.Index] = raw(r)
				}
			} else if len(reports) > 0 {
				n := -1
				if _, e := fmt.Sscanf(f.Name, "unit-%d", &n); e == nil {
					p := current[reports[0].Pools[n]]
					p[47] = byte(rules.BattleUnitSprite(typed(p), -1))
				}
			}
			for _, r := range f.Units {
				if !bytes.Equal(got[r.Index], raw(r)) {
					panic(fmt.Sprintf("%s/%s池%d完整78-byte不同", c.Scenario, f.Name, r.Index))
				}
				checked++
			}
		}
		results = append(results, map[string]any{"scenario": c.Scenario, "reports": reports,
			"whole_records_checked": checked, "existing_rule_differences": existingDifferences,
			"scope": "disposable current-position model only; production/UI/save not verified"})
	}
	out, err := json.MarshalIndent(results, "", "  ")
	fail(err)
	fmt.Println(string(out))
}
