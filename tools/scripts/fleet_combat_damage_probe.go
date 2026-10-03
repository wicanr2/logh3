// 第64輪可丟棄戰損probe。092仍DRAFT；本程式不由產品流程引用。
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

type hit struct {
	Regs map[string]int
	BP   map[int]int
}

var regPattern = regexp.MustCompile(`([A-Z]+)=([0-9A-F]{4})`)
var bpPattern = regexp.MustCompile(`([0-9A-F]{2}):([0-9A-F]{4})`)

func mustRead(p string) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}
	return b
}
func word(b []byte, off int) int { return int(binary.LittleEndian.Uint16(b[off:])) }
func long(b []byte, off int) int { return int(binary.LittleEndian.Uint32(b[off:])) }
func hex(s string) int {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		panic(err)
	}
	return int(n)
}

func observations(b []byte) map[string][]hit {
	out := make(map[string][]hit)
	var owner string
	var current *hit
	for _, line := range strings.Split(string(b), "\n") {
		if strings.Contains(line, "執行時的暫存器（前") {
			owner = strings.Fields(line)[0]
			current = nil
		} else if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(line, "AX=") && owner != "" {
			h := hit{Regs: map[string]int{}, BP: map[int]int{}}
			for _, m := range regPattern.FindAllStringSubmatch(line, -1) {
				h.Regs[m[1]] = hex(m[2])
			}
			out[owner] = append(out[owner], h)
			current = &out[owner][len(out[owner])-1]
		} else if current != nil && strings.Contains(line, "[bp+]") {
			for _, m := range bpPattern.FindAllStringSubmatch(line, -1) {
				current.BP[hex(m[1])] = hex(m[2])
			}
		}
	}
	return out
}

func unit(m []byte, pool int) rules.BattleUnit {
	a := 0x668b2 + pool*78
	return rules.BattleUnit{Faction: int(m[a]), Fleet: int(m[a+1]), Squadron: int(m[a+3]), Kind: int(m[a+4]),
		Ships: word(m, a+8), FullShips: word(m, a+10), RepairCap: word(m, a+12),
		Missiles: long(m, a+14), Fighters: long(m, a+18), X: word(m, a+0x34), Y: word(m, a+0x36)}
}
func kind(m []byte, u rules.BattleUnit) rules.ShipKind {
	a := ds + 0x4e3a + u.Faction*208 + u.Kind*26
	k := rules.ShipKind{Kind: u.Kind, SemiMajor: word(m, a+22), SemiMinor: word(m, a+24)}
	for i := 0; i < 13; i++ {
		k.Words = append(k.Words, word(m, a+2*i))
	}
	return k
}

// 原版00EC的比例插值取32位元乘積高word；負斜率不是Fixed.Int()的往零截斷。
func attack(m []byte, own, target rules.BattleUnit, weapon int) (int, int, int) {
	k := kind(m, own)
	t := k.Byte(10 + weapon*4)
	if t == 255 {
		panic("不合法武器，不能猜fallback")
	}
	a := ds + 0x4fda + t*6
	power, major, minor := word(m, a), word(m, a+2), word(m, a+4)
	normalizer := rules.FixedMul(rules.FixedFromInt(major), rules.FixedOne+rules.EllipseEccentricity(major, minor))
	q := rules.FixedDiv(rules.FixedDist(own.X, own.Y, target.X, target.Y), normalizer)
	if q <= 0 {
		panic("重疊座標的curve前項尚未知")
	}
	ptr := ds + 0x477c + t*4
	curve := word(m, ptr+2)*16 + word(m, ptr)
	for i := 1; i < 16; i++ {
		left, right := curve+(i-1)*6, curve+i*6
		x0, x1 := rules.Fixed(long(m, left)), rules.Fixed(long(m, right))
		if q > x1 {
			continue
		}
		fraction := rules.FixedDiv(q-x0, x1-x0)
		y0, y1 := word(m, left+4), word(m, right+4)
		curveY := y0 + int(int32(int64(fraction)*int64(y1-y0))>>16)
		rawPower := curveY * power / 100
		fleet := ds + 0xc074 + own.Fleet*122
		return rawPower * int(m[fleet+0x3d+weapon]) / 100, int(q), curveY
	}
	panic("curve沒有合法上界")
}

func check(name string, got, want int) {
	if got != want {
		panic(fmt.Sprintf("%s: probe=%d original=%d", name, got, want))
	}
}

func run(dir string) map[string]any {
	m, end := mustRead(filepath.Join(dir, "input-mem.bin")), mustRead(filepath.Join(dir, "end-mem.bin"))
	if len(m) != 1048576 || len(end) != 1048576 {
		panic("RAM大小不符")
	}
	obs := observations(mustRead(filepath.Join(dir, "report.txt")))
	own, target := unit(m, 80), unit(m, 102)
	name := filepath.Base(dir)
	if len(obs["1DD6:0C54"]) == 0 {
		if !strings.Contains(name, "zero-own") && !strings.Contains(name, "zero-target") {
			panic("應有傷害call")
		}
		check("zero own", unit(end, 80).Ships, own.Ships)
		check("zero target", unit(end, 102).Ships, target.Ships)
		for pool, u := range map[int]rules.BattleUnit{80: own, 102: target} {
			v := unit(end, pool)
			check("zero missiles", v.Missiles, u.Missiles)
			check("zero fighters", v.Fighters, u.Fighters)
			check("zero repair cap", v.RepairCap, u.RepairCap)
		}
		for _, off := range []int{0x8fe4, 0x8fec} {
			check("zero score", long(end, ds+off), long(m, ds+off))
		}
		r := rules.NewRand(uint32(long(m, 0x5e00c)))
		r.Advance(16)
		check("zero damage seed", int(r.Seed()), long(end, 0x5e00c))
		return map[string]any{"case": name, "guard_no_damage_call": true}
	}
	if len(obs["1DD6:0C54"]) != 1 {
		panic("限定單擊probe")
	}
	weapon := int(m[0x668b2+80*78+0x2e])
	atk, q, curveY := attack(m, own, target, weapon)
	tf := ds + 0xc074 + target.Fleet*122
	tk := kind(m, target)
	bearing := rules.Atan256(own.X-target.X, own.Y-target.Y)
	def := rules.AttackCoefficient(tk.SemiMajor, tk.SemiMinor, rules.AngleDiff256(bearing, int(m[tf+0x4a])), int(m[tf+0x3d+weapon]))
	if def <= 0 {
		panic("零防禦不在已證實domain")
	}
	check(name+" attack", atk, obs["1DD6:0C6D"][0].Regs["AX"])
	check(name+" defence", def, obs["1DD6:0C83"][0].Regs["AX"])
	shots := kind(m, own).Byte(11+4*weapon) * own.Ships
	if weapon == 1 {
		if shots > own.Missiles {
			shots = own.Missiles
		}
		own.Missiles -= shots
	}
	if weapon == 2 {
		if shots > own.Fighters {
			shots = own.Fighters
		}
		own.Fighters -= shots * 20 / 100
		if own.Kind == 3 {
			shots /= 2
		}
	}
	check(name+" shots", shots, obs["1DD6:0E17"][0].BP[14]+65536*obs["1DD6:0E17"][0].BP[16])
	pct := int(uint16(int32(shots*atk) / int32(def) / 750))
	if pct > 100 {
		pct = 100
	}
	if m[0x668b2+102*78+0x2e] == 3 {
		pct /= 2
	}
	ad := ds + 0xb77c + int(m[tf+4])*28
	attr := int(m[ad+0x12])
	if target.Squadron == 0 {
		if attr > 100 {
			pct = pct * 100 / attr
		} else if pct >= 80 {
			faction := int(m[ad+1])
			seat := 1
			if faction == int(m[ds+0x8f25]) {
				seat = 0
			}
			if faction != 2 && m[ds+0x8f21+seat] == 1 {
				pct -= 10
			}
		}
	}
	check(name+" percent", pct, obs["1DD6:0F59"][0].BP[4])
	r := rules.NewRand(uint32(long(m, 0x5e00c)))
	var plan struct {
		InputStep int `json:"input_dump_step"`
	}
	if err := json.Unmarshal(mustRead(filepath.Join("workplace", name+"-plan.json")), &plan); err != nil {
		panic(err)
	}
	// 觀測及早注入快照位於030F之前；本固定局面8筆已由第63輪證實各兩次排序亂數。
	if plan.InputStep < 496477068 {
		r.Advance(16)
	}
	noise := r.Next()%200 - attr
	if target.Ships < 100 {
		noise /= 2
	}
	request := int(int16(target.Ships*pct/100 + noise))
	if request < 0 {
		request = 0
	}
	fallback := request == 0
	if fallback {
		request = r.Next() % 10
	}
	check(name+" request", request, obs["1DD6:0F59"][0].BP[6])
	lost := rules.ApplyShipLoss(&target, tk, request)
	check(name+" actual", lost, obs["1DD6:0F60"][0].Regs["AX"])
	result := unit(end, 102)
	for field, pair := range map[string][2]int{"ships": {target.Ships, result.Ships}, "repair_cap": {target.RepairCap, result.RepairCap}, "missiles": {target.Missiles, result.Missiles}, "fighters": {target.Fighters, result.Fighters}} {
		check(name+" target "+field, pair[0], pair[1])
	}
	resultOwn := unit(end, 80)
	check(name+" own missiles", own.Missiles, resultOwn.Missiles)
	check(name+" own fighters", own.Fighters, resultOwn.Fighters)
	check(name+" seed", int(r.Seed()), long(end, 0x5e00c))
	for _, offset := range []int{0x8fe4, 0x8fec} {
		check(name+" score", long(end, ds+offset)-long(m, ds+offset), lost/20)
	}
	// 0743既有契約：攻方提督+量/2、副官再除二，存於ADML+02 u32。
	of := ds + 0xc074 + own.Fleet*122
	for i, divisor := range []int{2, 4} {
		id := int(m[of+4+i])
		if id == 255 {
			continue
		}
		a := ds + 0xb77c + id*28 + 2
		check(name+" merit", long(end, a)-long(m, a), lost/divisor)
	}
	return map[string]any{"case": name, "attack": atk, "defence": def, "normalized_distance": q, "curve_y": curveY, "shots": shots, "target_attribute": attr, "percent": pct, "requested": request, "actual": lost, "fallback": fallback, "seed_after": fmt.Sprintf("%08x", r.Seed()), "projected_target_after": target, "projected_own_after": own, "score_each": lost / 20}
}

func main() {
	dirs, err := filepath.Glob("workplace/oracle/damage79-r64-*-a")
	if err != nil {
		panic(err)
	}
	var results []map[string]any
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(dir, "input-mem.bin")); err != nil {
			continue
		}
		if strings.Contains(dir, "flag-high101") {
			continue
		} // 舊案名稱誤讀，保留但不列目前集合。
		results = append(results, run(dir))
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		panic(err)
	}
}
