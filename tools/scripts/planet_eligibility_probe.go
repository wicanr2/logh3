// 第110輪#91唯讀診斷：實際原版第⑥步鎖定向量與目前正式Targets比較。
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/rules"
)

func main() {
	if len(os.Args) != 3 {
		panic("probe <original-phase6-input-mem.bin> <new-report.json>")
	}
	m, err := os.ReadFile(os.Args[1])
	if os.IsNotExist(err) {
		fmt.Println("SKIP 缺固定原版RAM")
		return
	}
	if err != nil {
		panic(err)
	}
	if len(m) != 1048576 {
		panic("原版RAM大小不符")
	}
	word := func(a int) int { return int(binary.LittleEndian.Uint16(m[a:])) }
	ptr := func(a int) int { return word(a) + 16*word(a+2) }
	ds := 0x54b30
	ship := ptr(ds + 0xcfb4)
	planet := ptr(ds + 0xcff8)
	var planets []rules.BattlePlanet
	for i := 0; i < int(m[ds+0xcffc]); i++ {
		a := planet + i*82
		planets = append(planets, rules.BattlePlanet{Owner: int(m[a+17]), X: word(a + 70), Y: word(a + 72), Defence: word(a + 20), Fortress: m[a+18] != 0, GunState: int(m[a+69])})
	}
	var fleets []rules.BattleFleet
	var units []rules.BattleUnit
	var expectedTargets, expectedBits []int
	for i := 0; i < int(m[ds+0x9894]); i++ {
		id := int(m[ds+0x9898+i])
		a := ds + 0xc074 + id*122
		fleets = append(fleets, rules.BattleFleet{Index: id, Faction: int(m[a+1]), Flag0: m[a]&1 != 0, Action: int(m[a+72] & 15), X: word(a + 92), Y: word(a + 94)})
		s := ship + word(a+16)*78
		units = append(units, rules.BattleUnit{Fleet: id, Squadron: 0, Ships: word(s + 8)})
		target := int(m[a+90])
		if target == 255 {
			target = -1
		}
		expectedTargets = append(expectedTargets, target)
		expectedBits = append(expectedBits, word(a+112))
	}
	target, bits := rules.PlanetAssaultTargets(planets, fleets, units)
	var differences []map[string]int
	for i, f := range fleets {
		if target[i] != expectedTargets[i] || bits[i] != expectedBits[i] {
			differences = append(differences, map[string]int{"fleet": f.Index, "action": f.Action, "original_target": expectedTargets[i], "current_target": target[i], "original_bits": expectedBits[i], "current_bits": bits[i]})
		}
	}
	report := map[string]any{"classification": "原版已過eligibility的父RAM診斷；非重新執行0091；不改正式規則", "source": os.Args[1], "sha256": fmt.Sprintf("%x", sha256.Sum256(m)), "original_segment": "20B4:0091", "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "planets": planets, "fleets": fleets, "units": units, "original_targets": expectedTargets, "original_bits": expectedBits, "current_targets": target, "current_bits": bits, "differences": differences}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(append(b, '\n')); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Println("fleets", len(fleets), "differences", differences)
}
