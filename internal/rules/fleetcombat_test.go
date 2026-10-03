package rules

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// 期望值直接取原版RAM／AX／BP，不取可丟棄probe計算結果。
// 原版輸入、flags與dosgolem版本的雜湊綁在受版控凍結收據中；缺原版就skip。
const fcDS = 0x54b30
const fcPool = 0x668b2
const fcGolem = "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"

type fcSource struct {
	SHA string `json:"sha256"`
}
type fcManifest struct {
	Head    string              `json:"dosgolem_head"`
	Files   map[string]fcSource `json:"files"`
	Sources map[string]fcSource `json:"sources"`
	Cases   []struct {
		Name string `json:"case"`
	} `json:"cases"`
	Pairs []struct {
		Name string `json:"case"`
	} `json:"pairs"`
}

func fcLoadManifest(t *testing.T, name, digest string) fcManifest {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../..", "docs/internals/data", name))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != digest {
		t.Fatal("凍結收據雜湊改變，須重新審查原版證據")
	}
	var m fcManifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m.Head != fcGolem {
		t.Fatal("原版收據執行器版本不符")
	}
	return m
}

func fcRequireOriginal(t *testing.T) {
	t.Helper()
	b, err := os.ReadFile("../../org_game/extracted/GIN3.COM")
	if os.IsNotExist(err) {
		t.Skip("缺原版GIN3.COM，不用自製代用品")
	}
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		t.Fatal("原版GIN3.COM雜湊不符")
	}
}

func (m fcManifest) read(t *testing.T, path string) []byte {
	t.Helper()
	s, ok := m.Files[path]
	if !ok {
		s, ok = m.Sources[path]
	}
	if !ok {
		t.Fatalf("收據沒有索引%s", path)
	}
	b, err := os.ReadFile(filepath.Join("../..", path))
	if os.IsNotExist(err) {
		t.Skipf("缺原版收據%s，不用代用品", path)
	}
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != s.SHA {
		t.Fatalf("原版收據%s雜湊不符", path)
	}
	return b
}

func (m fcManifest) ram(t *testing.T, path string) []byte {
	t.Helper()
	b := m.read(t, path)
	if len(b) != 1048576 {
		t.Fatalf("%s不是完整1MiB RAM", path)
	}
	return b
}

func fcWord(m []byte, a int) int { return int(binary.LittleEndian.Uint16(m[a:])) }
func fcLong(m []byte, a int) int { return int(binary.LittleEndian.Uint32(m[a:])) }
func fcUnit(m []byte, pool int) BattleUnit {
	a := fcPool + pool*78
	return BattleUnit{Faction: int(m[a]), Fleet: int(m[a+1]), Squadron: int(m[a+3]), Kind: int(m[a+4]),
		Morale: int(m[a+6]), Ships: fcWord(m, a+8), FullShips: fcWord(m, a+10), RepairCap: fcWord(m, a+12),
		Missiles: fcLong(m, a+14), Fighters: fcLong(m, a+18), X: fcWord(m, a+0x34), Y: fcWord(m, a+0x36)}
}
func fcCommand(m []byte, own, target BattleUnit, pool int) FleetCombatCommand {
	of, tf := fcDS+0xc074+own.Fleet*122, fcDS+0xc074+target.Fleet*122
	a := fcDS + 0xb77c + int(m[tf+4])*28
	c := FleetCombatCommand{Weapon: int(m[fcPool+pool*78+0x2e]),
		TargetWeapon: int(m[fcPool+fcWord(m, fcPool+pool*78+0x2c)*78+0x2e]),
		TargetFacing: int(m[tf+0x4a]), TargetAttribute: int(m[a+0x12])}
	for i := range c.AttackAbilities {
		c.AttackAbilities[i], c.DefendAbilities[i] = int(m[of+0x3d+i]), int(m[tf+0x3d+i])
	}
	seat := 1
	if m[a+1] == m[fcDS+0x8f25] {
		seat = 0
	}
	c.TargetComputer = m[a+1] != 2 && m[fcDS+0x8f21+seat] == 1
	return c
}
func fcKinds(t *testing.T) *ShipKindTable {
	t.Helper()
	k, err := LoadShipKindTable("../../docs/internals/data/ship-kind-table.json")
	if err != nil {
		t.Fatal(err)
	}
	return k
}

type fcObservation struct{ regs, bp map[int]int }

var fcRegRE = regexp.MustCompile(`(AX|BP)=([0-9A-F]{4})`)
var fcStackRE = regexp.MustCompile(`SS:SP=([0-9A-F]{4}):([0-9A-F]{4})`)
var fcBPRE = regexp.MustCompile(`([0-9A-F]{2}):([0-9A-F]{4})`)

func fcObservations(t *testing.T, b []byte) map[string][]fcObservation {
	t.Helper()
	parse := func(s string) int {
		n, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		return int(n)
	}
	out := map[string][]fcObservation{}
	owner := ""
	var current *fcObservation
	for _, line := range strings.Split(string(b), "\n") {
		if strings.Contains(line, "執行時的暫存器（前") {
			owner, current = strings.Fields(line)[0], nil
		} else if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(line, "AX=") && owner != "" {
			h := fcObservation{regs: map[int]int{}, bp: map[int]int{}}
			for _, v := range fcRegRE.FindAllStringSubmatch(line, -1) {
				reg := 0
				if v[1] == "BP" {
					reg = 1
				}
				h.regs[reg] = parse(v[2])
			}
			if v := fcStackRE.FindStringSubmatch(line); v != nil {
				h.regs[2] = parse(v[1])
			}
			out[owner] = append(out[owner], h)
			current = &out[owner][len(out[owner])-1]
		} else if current != nil && strings.Contains(line, "[bp+]") {
			for _, v := range fcBPRE.FindAllStringSubmatch(line, -1) {
				current.bp[parse(v[1])] = parse(v[2])
			}
		}
	}
	return out
}

func fcValue(t *testing.T, obs map[string][]fcObservation, addr string, index, field int, bp bool) int {
	t.Helper()
	if len(obs[addr]) <= index {
		t.Fatalf("缺原版%s第%d筆觀測", addr, index)
	}
	values := obs[addr][index].regs
	if bp {
		values = obs[addr][index].bp
	}
	v, ok := values[field]
	if !ok {
		t.Fatalf("缺原版%s第%d筆欄位%d", addr, index, field)
	}
	return v
}

func fcEqual(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s：remake=%d，原版=%d", field, got, want)
	}
}
func fcCheckUnit(t *testing.T, u, v BattleUnit) {
	t.Helper()
	for field, pair := range map[string][2]int{"ships": {u.Ships, v.Ships}, "full": {u.FullShips, v.FullShips},
		"repair": {u.RepairCap, v.RepairCap}, "missiles": {u.Missiles, v.Missiles}, "fighters": {u.Fighters, v.Fighters}} {
		fcEqual(t, field, pair[0], pair[1])
	}
}
func fcCheckHit(t *testing.T, h FleetCombatHit, obs map[string][]fcObservation, index int) {
	t.Helper()
	fcEqual(t, "attack", h.Attack, fcValue(t, obs, "1DD6:0C6D", index, 0, false))
	fcEqual(t, "defence", h.Defence, fcValue(t, obs, "1DD6:0C83", index, 0, false))
	fcEqual(t, "shots", h.Shots, fcValue(t, obs, "1DD6:0E17", index, 14, true)+65536*fcValue(t, obs, "1DD6:0E17", index, 16, true))
	fcEqual(t, "percent", h.Percent, fcValue(t, obs, "1DD6:0F59", index, 4, true))
	fcEqual(t, "requested", h.Requested, fcValue(t, obs, "1DD6:0F59", index, 6, true))
	fcEqual(t, "actual", h.Lost, fcValue(t, obs, "1DD6:0F60", index, 0, false))
}

func fcCheckTotals(t *testing.T, m, end []byte, fleet, scorer, score, lead, adj int) {
	t.Helper()
	for _, offset := range []int{0x8fe4 + scorer*4, 0x8fec + scorer*4} {
		fcEqual(t, "score", score, fcLong(end, fcDS+offset)-fcLong(m, fcDS+offset))
	}
	for i, increment := range []int{lead, adj} {
		id := m[fcDS+0xc074+fleet*122+4+i]
		if id != 255 {
			a := fcDS + 0xb77c + int(id)*28 + 2
			fcEqual(t, "merit", increment, fcLong(end, a)-fcLong(m, a))
		}
	}
}

func fcSingle(t *testing.T, manifest fcManifest, name string, fullObservation bool) {
	t.Helper()
	dir := "workplace/oracle/" + name
	m, end := manifest.ram(t, dir+"/input-mem.bin"), manifest.ram(t, dir+"/end-mem.bin")
	obs := fcObservations(t, manifest.read(t, dir+"/report.txt"))
	// 也驗flags及meta的來源雜湊，防止拿到未綁版本的同名傾印。
	manifest.read(t, dir+"/flags.txt")
	manifest.read(t, dir+"/meta.txt")
	var plan struct {
		InputStep int `json:"input_dump_step"`
	}
	if err := json.Unmarshal(manifest.read(t, "workplace/"+name+"-plan.json"), &plan); err != nil {
		t.Fatal(err)
	}
	own, target := fcUnit(m, 80), fcUnit(m, 102)
	r := NewRand(uint32(fcLong(m, fcDS+0x94dc)))
	if plan.InputStep < 496477068 {
		r.Advance(16) // 本固定局面有8筆排序；多擊測試另核全部原版鍵與排列。
	}
	h, err := ApplyFleetCombatDamage(&own, &target, fcKinds(t), fcCommand(m, own, target, 80), r)
	if err != nil {
		t.Fatal(err)
	}
	if len(obs["1DD6:0C54"]) == 0 {
		if !h.Skipped || (!strings.Contains(name, "zero-own") && !strings.Contains(name, "zero-target")) {
			t.Fatal("原版與remake的零艘數guard不一致")
		}
	} else {
		fcEqual(t, "call count", len(obs["1DD6:0C54"]), 1)
		if h.Skipped {
			t.Fatal("原版執行傷害，remake卻略過")
		}
		if fullObservation {
			fcCheckHit(t, h, obs, 0)
		}
	}
	fcCheckUnit(t, own, fcUnit(end, 80))
	fcCheckUnit(t, target, fcUnit(end, 102))
	fcEqual(t, "seed", int(r.Seed()), fcLong(end, fcDS+0x94dc))
	fcCheckTotals(t, m, end, own.Fleet, 1-target.Faction, h.Score, h.AdmiralMerit, h.AdjunctMerit)
}

func TestFleetCombatDamageMatchesOracle(t *testing.T) {
	fcRequireOriginal(t)
	m := fcLoadManifest(t, "fleet-combat-damage.json", "360277c8dcbf6214da429baead75d4a7bf20b09b03f1ca2af3c0313df4171b30")
	fcEqual(t, "固定戰損案例數", len(m.Cases), 31)
	for _, c := range m.Cases {
		t.Run(c.Name, func(t *testing.T) { fcSingle(t, m, "damage79-r64-"+c.Name+"-a", true) })
	}
}

func TestFleetCombatCurvesAndApprovedZeroDistance(t *testing.T) {
	fcRequireOriginal(t)
	m := fcLoadManifest(t, "combat-sequence.json", "199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8")
	near, zero := 0, 0
	for _, c := range m.Pairs {
		if !strings.HasPrefix(c.Name, "curve") {
			continue
		}
		if strings.HasSuffix(c.Name, "-near") {
			near++
			t.Run(c.Name, func(t *testing.T) { fcSingle(t, m, "sequence79-r66-"+c.Name+"-a", true) })
		} else if strings.HasSuffix(c.Name, "-zero") || c.Name == "curve0-stack-first" {
			zero++
			t.Run(c.Name+"-approved-difference", func(t *testing.T) {
				ram := m.ram(t, "workplace/oracle/sequence79-r66-"+c.Name+"-a/input-mem.bin")
				own, target := fcUnit(ram, 80), fcUnit(ram, 102)
				command := fcCommand(ram, own, target, 80)
				k, _ := fcKinds(t).Kind(own.Faction, own.Kind)
				typeID := k.Byte(10 + 4*command.Weapon)
				ptr := fcDS + 0x477c + 4*typeID
				curve := fcWord(ram, ptr) + fcWord(ram, ptr+2)*16
				firstY := fcWord(ram, curve+4)
				power := fcWord(ram, fcDS+0x4fda+typeID*6)
				h, err := ApplyFleetCombatDamage(&own, &target, fcKinds(t), command, NewRand(123))
				if err != nil {
					t.Fatal(err)
				}
				fcEqual(t, "q", int(h.NormalizedDistance), 0)
				fcEqual(t, "核准首點Y", h.CurveY, firstY)
				fcEqual(t, "核准首點attack", h.Attack, firstY*power/100*command.AttackAbilities[command.Weapon]/100)
				// 此處刻意不比原版殘留stack所造成的戰損，不能宣稱其parity。
			})
		}
	}
	fcEqual(t, "近距離武器案例數", near, 11)
	fcEqual(t, "零距離及stack反例數", zero, 12)
}

func TestFleetCombatLegalFighterMatchesOracle(t *testing.T) {
	fcRequireOriginal(t)
	m := fcLoadManifest(t, "weapon-candidates.json", "59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f")
	// 此正常原版UI續接收據只保存入口及末端，沒有0C6D等中途係數點。
	// 比較呼叫次數、完整戰損欄位、庫存、分數、功績與seed；係數由其他案直接核。
	fcSingle(t, m, "weapon79-r65-fighter-end-a", false)
}

func TestFleetCombatSequenceMatchesOracle(t *testing.T) {
	fcRequireOriginal(t)
	manifest := fcLoadManifest(t, "combat-sequence.json", "199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8")
	for _, name := range []string{"multi-ranked", "same-target-measured"} {
		t.Run(name, func(t *testing.T) {
			dir := "workplace/oracle/sequence79-r66-" + name + "-a"
			m, end := manifest.ram(t, dir+"/input-mem.bin"), manifest.ram(t, dir+"/end-mem.bin")
			keysRAM, sortedRAM := manifest.ram(t, dir+"/keys-mem.bin"), manifest.ram(t, dir+"/sorted-mem.bin")
			obs := fcObservations(t, manifest.read(t, dir+"/report.txt"))
			manifest.read(t, dir+"/flags.txt")
			manifest.read(t, dir+"/meta.txt")
			kinds := fcKinds(t)
			fleet := fcDS + 0xc074 + 12*122
			count := int(m[fleet+15])
			fcEqual(t, "一般分隊數", count, 8)
			entries, pools := make([]FleetCombatOrderEntry, count), make([]int, count)
			for i := range entries {
				pools[i] = fcWord(m, fleet+0x10+4*i)
				lead := fcDS + 0xb77c + int(m[fleet+4])*28
				ops := int(m[lead+0x17])
				if id := m[fleet+5]; id != 255 {
					ops = max(ops, int(m[fcDS+0xb77c+int(id)*28+0x17]))
				}
				mobility, err := FleetCombatMobility(ops, count)
				if err != nil {
					t.Fatal(err)
				}
				fcEqual(t, "原版機動力AL", mobility, fcValue(t, obs, "1DD6:0522", i, 0, false)&255)
				entries[i] = FleetCombatOrderEntry{fcUnit(m, pools[i]), int(m[lead+11]), mobility}
			}
			r := NewRand(uint32(fcLong(m, fcDS+0x94dc)))
			order, keys, err := FleetCombatOrder(entries, kinds, r)
			if err != nil {
				t.Fatal(err)
			}
			for i, pool := range pools {
				fcEqual(t, "原版排序鍵", int(keys[i]), fcWord(keysRAM, fcPool+pool*78+0x42))
			}
			array := fcValue(t, obs, "1DD6:033E", 0, 2, false)*16 + fcValue(t, obs, "1DD6:033E", 0, 1, false)
			units := map[int]*BattleUnit{}
			for _, pool := range []int{80, 82, 102, 104} {
				u := fcUnit(m, pool)
				units[pool] = &u
			}
			score, leadMerit, adjMerit, hitIndex := 0, 0, 0, 0
			for rank, index := range order {
				pool := pools[index]
				fcEqual(t, "完整原版排列", pool, fcWord(sortedRAM, array+rank*2))
				weapon := int(m[fcPool+pool*78+0x2e])
				if weapon == 0xff || weapon == 3 {
					continue
				}
				checkpoint := manifest.ram(t, fmt.Sprintf("%s/damage%d-mem.bin", dir, hitIndex+1))
				for p, u := range units {
					fcCheckUnit(t, *u, fcUnit(checkpoint, p))
				}
				targetPool := fcWord(m, fcPool+pool*78+0x2c)
				fcEqual(t, "原版攻方指標", fcValue(t, obs, "1DD6:0C54", hitIndex, 8, true)+16*fcValue(t, obs, "1DD6:0C54", hitIndex, 10, true), fcPool+pool*78)
				fcEqual(t, "原版目標指標", fcValue(t, obs, "1DD6:0C54", hitIndex, 12, true)+16*fcValue(t, obs, "1DD6:0C54", hitIndex, 14, true), fcPool+targetPool*78)
				h, err := ApplyFleetCombatDamage(units[pool], units[targetPool], kinds, fcCommand(m, *units[pool], *units[targetPool], pool), r)
				if err != nil {
					t.Fatal(err)
				}
				fcCheckHit(t, h, obs, hitIndex)
				score, leadMerit, adjMerit = score+h.Score, leadMerit+h.AdmiralMerit, adjMerit+h.AdjunctMerit
				hitIndex++
			}
			fcEqual(t, "兩擊", hitIndex, 2)
			fcEqual(t, "原版呼叫數", len(obs["1DD6:0C54"]), hitIndex)
			for pool, u := range units {
				fcCheckUnit(t, *u, fcUnit(end, pool))
			}
			fcEqual(t, "序列seed", int(r.Seed()), fcLong(end, fcDS+0x94dc))
			fcCheckTotals(t, m, end, 12, 0, score, leadMerit, adjMerit)
		})
	}
}

func TestFleetCombatTablesMatchOriginal(t *testing.T) {
	fcRequireOriginal(t)
	manifest := fcLoadManifest(t, "combat-sequence.json", "199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8")
	m := manifest.ram(t, "workplace/oracle/normal79-r61-a/end-mem.bin")
	for i, p := range fleetWeaponProfiles {
		for j, v := range []int{p.power, p.major, p.minor} {
			fcEqual(t, "武器原表", v, fcWord(m, fcDS+0x4fda+i*6+j*2))
		}
		ptr := fcDS + 0x477c + i*4
		curve := fcWord(m, ptr) + fcWord(m, ptr+2)*16
		for j, x := range fleetCurveX {
			fcEqual(t, "六點X", int(x), fcLong(m, curve+j*6))
			fcEqual(t, "六點Y", p.y[j], fcWord(m, curve+j*6+4))
		}
	}
	for i, v := range fleetOrderBases {
		fcEqual(t, "排序基值", v, fcWord(m, fcDS+0x4426+i*2))
	}
	for i, v := range fleetMobilityCoefficients {
		fcEqual(t, "機動係數", v, int(m[fcDS+0x4e2e+i]))
	}
	for _, f := range fcKinds(t).Factions {
		if f.Faction > 1 {
			continue
		}
		for _, k := range f.Kinds {
			for i, v := range k.Words {
				fcEqual(t, "艦種原表", v, fcWord(m, fcDS+0x4e3a+f.Faction*208+k.Kind*26+i*2))
			}
		}
	}
}

// 已提交的兩命令沿原版呼叫次序作測試輸入；此處只驗死亡後停止，
// 排序本身由TestFleetCombatSequenceMatchesOracle獨立核原版全部鍵。
func TestFleetCombatDeathStopsFollowingCommand(t *testing.T) {
	fcRequireOriginal(t)
	manifest := fcLoadManifest(t, "fleet-scene.json", "7d7e5b388eebc8348825e91352ff10e9821ce2b0cf7912d6f9b9cf1cafc95379")
	dir := "workplace/oracle/scene48-r69-v2-death-a"
	start, end := manifest.ram(t, dir+"/start-mem.bin"), manifest.ram(t, dir+"/end-mem.bin")
	var report struct {
		Events []struct {
			Location string   `json:"original_location"`
			Seed     uint32   `json:"seed"`
			Words    []uint16 `json:"stack_words"`
		} `json:"events"`
	}
	if err := json.Unmarshal(manifest.read(t, dir+"/report.json"), &report); err != nil {
		t.Fatal(err)
	}
	target := fcUnit(start, 102)
	fcEqual(t, "明示注入目標艘數", target.Ships, 1)
	var r *Rand
	owners, damages, hits := 0, 0, 0
	for _, e := range report.Events {
		if e.Location == "1DD6:0C54" {
			damages++
		}
		if e.Location != "1DD6:05D7" {
			continue
		}
		owners++
		if r == nil {
			r = NewRand(e.Seed)
		}
		fcEqual(t, "原版逐命令seed", int(r.Seed()), int(e.Seed))
		pool := int(e.Words[2])
		own := fcUnit(start, pool)
		beforeOwn, beforeTarget, beforeSeed := own, target, r.Seed()
		h, err := ApplyFleetCombatDamage(&own, &target, fcKinds(t), fcCommand(start, own, target, pool), r)
		if err != nil {
			t.Fatal(err)
		}
		if h.Skipped {
			if own != beforeOwn || target != beforeTarget || r.Seed() != beforeSeed {
				t.Fatal("死亡後的命令仍扣庫存、修改目標或消費RND")
			}
		} else {
			hits++
			fcEqual(t, "首擊實扣", h.Lost, 1)
		}
		fcCheckUnit(t, own, fcUnit(end, pool))
	}
	fcEqual(t, "原版到達兩命令入口", owners, 2)
	fcEqual(t, "原版傷害呼叫", damages, 1)
	fcEqual(t, "remake傷害次數", hits, damages)
	fcCheckUnit(t, target, fcUnit(end, 102))
	fcEqual(t, "末端seed", int(r.Seed()), fcLong(end, fcDS+0x94dc))
}

func TestFleetCombatInvalidInputDoesNotMutate(t *testing.T) {
	own := BattleUnit{Faction: 0, Kind: 0, Ships: 3000, FullShips: 3000, RepairCap: 3000, Fighters: 36000, X: 20}
	target := BattleUnit{Faction: 1, Kind: 0, Ships: 3000, FullShips: 3000, RepairCap: 3000}
	valid := FleetCombatCommand{Weapon: 2, TargetWeapon: 0xff, AttackAbilities: [3]int{100, 100, 100}, DefendAbilities: [3]int{100, 100, 100}, TargetAttribute: 104}
	for _, name := range []string{"weapon", "zero-defence", "missing-kind", "zero-full"} {
		t.Run(name, func(t *testing.T) {
			a, b, c, kinds := own, target, valid, fcKinds(t)
			switch name {
			case "weapon":
				c.Weapon = 4
			case "zero-defence":
				c.DefendAbilities[2] = 0
			case "missing-kind":
				kinds = nil
			case "zero-full":
				b.FullShips = 0
			}
			beforeA, beforeB := a, b
			r := NewRand(123)
			if _, err := ApplyFleetCombatDamage(&a, &b, kinds, c, r); err == nil {
				t.Fatal("非法輸入沒有回錯誤")
			}
			if a != beforeA || b != beforeB || r.Seed() != 123 {
				t.Fatal("錯誤前已扣庫存或消費RND")
			}
		})
	}
	for _, w := range []int{3, 0xff} {
		a, b, c := own, target, valid
		c.Weapon = w
		h, err := ApplyFleetCombatDamage(&a, &b, nil, c, nil)
		if err != nil || !h.Skipped || a != own || b != target {
			t.Fatal("防禦／無命令guard改變分隊")
		}
	}
	entries := []FleetCombatOrderEntry{{Unit: own, Attribute: 93, Mobility: 72}, {Unit: target, Attribute: 0, Mobility: 72}}
	before := append([]FleetCombatOrderEntry(nil), entries...)
	r := NewRand(123)
	if _, _, err := FleetCombatOrder(entries, fcKinds(t), r); err == nil || r.Seed() != 123 || !reflect.DeepEqual(entries, before) {
		t.Fatal("排序須先驗全部條目再取亂數")
	}
}
