// 第65輪可丟棄候選／合法戰鬥機探針。092 DRAFT，不由產品引用。
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/wicanr2/logh3/internal/rules"
)

const ds = 0x54b30

func read(p string) []byte {
	b, e := os.ReadFile(p)
	if e != nil {
		panic(e)
	}
	return b
}
func word(m []byte, a int) int { return int(binary.LittleEndian.Uint16(m[a:])) }
func long(m []byte, a int) int { return int(binary.LittleEndian.Uint32(m[a:])) }
func require(got, want int, label string) {
	if got != want {
		panic(fmt.Sprintf("%s: probe=%d original=%d", label, got, want))
	}
}
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
func unit(m []byte, pool int) rules.BattleUnit {
	a := 0x668b2 + pool*78
	return rules.BattleUnit{Faction: int(m[a]), Fleet: int(m[a+1]), Squadron: int(m[a+3]), Kind: int(m[a+4]), Ships: word(m, a+8), FullShips: word(m, a+10), RepairCap: word(m, a+12), Missiles: long(m, a+14), Fighters: long(m, a+18), X: word(m, a+0x34), Y: word(m, a+0x36)}
}
func kind(m []byte, u rules.BattleUnit) rules.ShipKind {
	a := ds + 0x4e3a + u.Faction*208 + u.Kind*26
	k := rules.ShipKind{Kind: u.Kind, SemiMajor: word(m, a+22), SemiMinor: word(m, a+24)}
	for i := 0; i < 13; i++ {
		k.Words = append(k.Words, word(m, a+i*2))
	}
	return k
}
func objects(m []byte) []int {
	var out []int
	for i := 0; i < word(m, ds+0x9896); i++ {
		out = append(out, word(m, ds+0x98ba+i*2))
	}
	return out
}

// 091C以兩方+7中心走整數射線；B0A先收原點在包圍盒內的可見物件。
func clearRay(m []byte, own, target int) bool {
	a, b := unit(m, own), unit(m, target)
	var blocks []rules.BattleUnit
	for _, p := range objects(m) {
		if p == own || p == target || m[0x668b2+p*78+0x1e]&128 != 0 {
			continue
		}
		u := unit(m, p)
		if u.X >= min(a.X, b.X) && u.X <= max(a.X, b.X) && u.Y >= min(a.Y, b.Y) && u.Y <= max(a.Y, b.Y) {
			blocks = append(blocks, u)
		}
	}
	x1, y1, x2, y2 := a.X+7, a.Y+7, b.X+7, b.Y+7
	dx, dy := abs(x2-x1), abs(y2-y1)
	check := func(x, y, count int) bool {
		if count&7 != 0 {
			return true
		}
		cx, cy := x/8, y/10
		ptr := word(m, ds+0x3c1a) + 16*word(m, ds+0x3c1c)
		tile := int(m[ptr+(cy/2)*128+cx/2])
		mask := m[ds+0x9bb4+tile] & m[ds+0x3c3e+2*(cy&1)+(cx&1)]
		if mask != 0 {
			return false
		}
		for _, u := range blocks {
			if x >= u.X && x < u.X+16 && y >= u.Y && y < u.Y+16 {
				return false
			}
		}
		return true
	}
	if dx > dy {
		if x1 > x2 {
			x1, x2 = x2, x1
			y1, y2 = y2, y1
		}
		sign := 1
		if y1 >= y2 {
			sign = -1
		}
		err := dx / 2
		for x, count := x1+1, 1; x <= x2; x, count = x+1, count+1 {
			err -= dy
			if err < 0 {
				err += dx
				y1 += sign
			}
			if !check(x, y1, count) {
				return false
			}
		}
	} else {
		if y1 > y2 {
			y1, y2 = y2, y1
			x1, x2 = x2, x1
		}
		sign := 1
		if x1 >= x2 {
			sign = -1
		}
		err := dy / 2
		for y, count := y1+1, 1; y <= y2; y, count = y+1, count+1 {
			err -= dx
			if err < 0 {
				err += dy
				x1 += sign
			}
			if !check(x1, y, count) {
				return false
			}
		}
	}
	return true
}
func mask(m []byte, own, target int) int {
	a, b := unit(m, own), unit(m, target)
	k := kind(m, a)
	dist := rules.FixedDist(a.X, a.Y, b.X, b.Y)
	facing := word(m, ds+0xc074+a.Fleet*122+0x4a)
	delta := rules.AngleDiff256(rules.Atan256(b.X-a.X, b.Y-a.Y), facing)
	out := 0
	for w := 0; w < 3; w++ {
		t := k.Byte(10 + w*4)
		if t == 255 || w == 1 && a.Missiles == 0 || w == 2 && a.Fighters == 0 {
			continue
		}
		s := ds + 0x4fda + t*6
		if dist <= rules.EllipseRadius(word(m, s+2), word(m, s+4), delta) && (w != 0 || clearRay(m, own, target)) {
			out |= 1 << w
		}
	}
	return out
}

type hit struct {
	Step int
	Reg  map[string]int
}

var regRE = regexp.MustCompile(`([A-Z]+)=([0-9A-F]{4})`)

func observations(s string) map[string][]hit {
	out := map[string][]hit{}
	owner := ""
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, "執行時的暫存器（前") {
			owner = strings.Fields(line)[0]
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(line, "AX=") {
			h := hit{Reg: map[string]int{}}
			h.Step, _ = strconv.Atoi(strings.TrimPrefix(strings.Fields(line)[0], "#"))
			for _, p := range regRE.FindAllStringSubmatch(line, -1) {
				n, e := strconv.ParseUint(p[2], 16, 16)
				if e != nil {
					panic(e)
				}
				h.Reg[p[1]] = int(n)
			}
			out[owner] = append(out[owner], h)
		}
	}
	return out
}
func candidate(dir string) map[string]any {
	m := read(filepath.Join(dir, "input-mem.bin"))
	obs := observations(string(read(filepath.Join(dir, "report.txt"))))
	queries := obs["1F6F:02D2"]
	returns := obs["1ED2:0097"]
	require(len(queries), len(returns), "query/return count")
	var counts []int
	for i, q := range queries {
		own := (0x668b0 + q.Reg["SI"] - 0x668b2) / 78
		u := unit(m, own)
		count := 0
		for _, p := range objects(m) {
			a := 0x668b2 + p*78
			v := unit(m, p)
			if v.Faction != 1-u.Faction || m[a+0x1e]&128 != 0 || rules.FixedDist(u.X, u.Y, v.X, v.Y) > rules.FixedFromInt(300) {
				continue
			}
			if m[a+0x17] == 0 || v.Kind == 7 {
				continue
			}
			got := mask(m, own, p)
			if got != 0 {
				count++
			}
			var writes []hit
			for _, h := range obs["1F6F:0382"] {
				if h.Step > q.Step && (i+1 == len(queries) || h.Step < queries[i+1].Step) && h.Reg["DI"] == p {
					writes = append(writes, h)
				}
			}
			require(len(writes), 1, "mask write count")
			require(writes[0].Reg["AX"]&255, got, dir+" mask")
		}
		require(returns[i].Reg["AX"], count, dir+" accepted count")
		counts = append(counts, count)
	}
	return map[string]any{"case": filepath.Base(dir), "query_counts": counts, "weapon_mask_before_target_filter": mask(m, 80, 102), "beam_clear": clearRay(m, 80, 102)}
}
func fighter(dir string) map[string]any {
	m, e := read(filepath.Join(dir, "input-mem.bin")), read(filepath.Join(dir, "end-mem.bin"))
	own, target := unit(m, 80), unit(m, 102)
	k := kind(m, own)
	s := ds + 0x4fda + k.Byte(18)*6
	q := rules.FixedDiv(rules.FixedDist(own.X, own.Y, target.X, target.Y), rules.FixedMul(rules.FixedFromInt(word(m, s+2)), rules.FixedOne+rules.EllipseEccentricity(word(m, s+2), word(m, s+4))))
	if q <= 0 {
		panic("distance0 curve domain尚未知")
	}
	ptr := ds + 0x477c + k.Byte(18)*4
	curve := word(m, ptr) + 16*word(m, ptr+2)
	power := -1
	for i := 1; i < 16; i++ {
		l, r := curve+(i-1)*6, curve+i*6
		x0, x1 := rules.Fixed(long(m, l)), rules.Fixed(long(m, r))
		if q > x1 {
			continue
		}
		fraction := rules.FixedDiv(q-x0, x1-x0)
		y0, y1 := word(m, l+4), word(m, r+4)
		power = (y0 + int(int32(int64(fraction)*int64(y1-y0))>>16)) * word(m, s) / 100
		break
	}
	if power < 0 {
		panic("曲線上限未知")
	}
	of, tf := ds+0xc074+own.Fleet*122, ds+0xc074+target.Fleet*122
	atk := power * int(m[of+0x3f]) / 100
	tk := kind(m, target)
	def := rules.AttackCoefficient(tk.SemiMajor, tk.SemiMinor, rules.AngleDiff256(rules.Atan256(own.X-target.X, own.Y-target.Y), int(m[tf+0x4a])), int(m[tf+0x3f]))
	if def <= 0 {
		panic("零防禦domain未知")
	}
	shots := min(k.Byte(19)*own.Ships, own.Fighters)
	own.Fighters -= shots * 20 / 100
	if own.Kind == 3 {
		shots /= 2
	}
	pct := min(int(uint16(int32(shots*atk)/int32(def)/750)), 100)
	attr := int(m[ds+0xb77c+int(m[tf+4])*28+0x12])
	if target.Squadron == 0 && attr > 100 {
		pct = pct * 100 / attr
	}
	rng := rules.NewRand(uint32(long(m, 0x5e00c)))
	rng.Advance(16)
	noise := rng.Next()%200 - attr
	if target.Ships < 100 {
		noise /= 2
	}
	request := max(int(int16(target.Ships*pct/100+noise)), 0)
	if request == 0 {
		request = rng.Next() % 10
	}
	lost := rules.ApplyShipLoss(&target, tk, request)
	actual := unit(e, 102)
	for label, pair := range map[string][2]int{"ships": {target.Ships, actual.Ships}, "repair": {target.RepairCap, actual.RepairCap}, "missiles": {target.Missiles, actual.Missiles}, "fighters": {target.Fighters, actual.Fighters}, "own_fighters": {own.Fighters, unit(e, 80).Fighters}, "own_missiles": {own.Missiles, unit(e, 80).Missiles}, "seed": {int(rng.Seed()), long(e, 0x5e00c)}} {
		require(pair[0], pair[1], label)
	}
	for _, off := range []int{0x8fe4, 0x8fec} {
		require(long(e, ds+off)-long(m, ds+off), lost/20, "score")
	}
	return map[string]any{"actual": lost, "requested": request, "attack": atk, "defence": def, "percent": pct, "shots": shots, "target_ships": actual.Ships, "own_fighters": own.Fighters, "seed_after": fmt.Sprintf("%08X", rng.Seed())}
}
func main() {
	dirs, err := filepath.Glob("workplace/oracle/weapon79-r65-*-a")
	if err != nil {
		panic(err)
	}
	var results []map[string]any
	for _, dir := range dirs {
		var plan struct {
			Stage string `json:"stage"`
		}
		if err = json.Unmarshal(read(filepath.Join("workplace", filepath.Base(dir)+"-plan.json")), &plan); err != nil {
			panic(err)
		}
		if plan.Stage == "entry" {
			results = append(results, candidate(dir))
		}
	}
	out := map[string]any{"candidates": results, "fighter": fighter("workplace/oracle/weapon79-r65-fighter-phase-end-a")}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err = enc.Encode(out); err != nil {
		panic(err)
	}
}
