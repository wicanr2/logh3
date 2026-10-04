package rules

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type candidateObservation struct {
	step int
	regs map[string]int
}

var candidateRegRE = regexp.MustCompile(`(AX|SI|DI)=([0-9A-F]{4})`)

func candidateObservations(t *testing.T, data []byte) map[string][]candidateObservation {
	t.Helper()
	out := map[string][]candidateObservation{}
	owner := ""
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "執行時的暫存器（前") {
			owner = strings.Fields(line)[0]
		} else if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(line, "AX=") {
			h := candidateObservation{regs: map[string]int{}}
			var err error
			h.step, err = strconv.Atoi(strings.TrimPrefix(strings.Fields(line)[0], "#"))
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range candidateRegRE.FindAllStringSubmatch(line, -1) {
				n, err := strconv.ParseUint(v[2], 16, 16)
				if err != nil {
					t.Fatal(err)
				}
				h.regs[v[1]] = int(n)
			}
			out[owner] = append(out[owner], h)
		}
	}
	return out
}

func candidateUnit(m []byte, pool int) BattleUnit {
	u := fcUnit(m, pool)
	u.Action, u.Detail = int(m[fcPool+pool*78+0x1e]), int(m[fcPool+pool*78+0x17])
	u.Out = u.Action&128 != 0
	return u
}

// 期望直接取原版每次查詢AX與0382的mask寫入，完全不讀model.candidates。
func TestFleetCombatCandidatesNativeQueries(t *testing.T) {
	fcRequireOriginal(t)
	manifest := fcLoadManifest(t, "weapon-candidates.json", "59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f")
	count, queries := 0, 0
	for _, c := range manifest.Pairs {
		dir := "workplace/oracle/weapon79-r65-" + c.Name + "-a"
		// 非entry的五個UI／結果案例另由核心與命令驗收，不混入候選查詢。
		if c.Name == "command" || c.Name == "selected" || strings.HasPrefix(c.Name, "fighter-") &&
			(c.Name == "fighter-end" || c.Name == "fighter-phase-end" || c.Name == "fighter-submit") {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			m := manifest.ram(t, dir+"/input-mem.bin")
			obs := candidateObservations(t, manifest.read(t, dir+"/report.txt"))
			q, returns := obs["1F6F:02D2"], obs["1ED2:0097"]
			if len(q) == 0 || len(q) != len(returns) {
				t.Fatal("原版查詢與返回觀測不完整")
			}
			var units []BattleUnit
			var pools []int
			for i := 0; i < fcWord(m, fcDS+0x9896); i++ {
				pool := fcWord(m, fcDS+0x98ba+i*2)
				pools = append(pools, pool)
				units = append(units, candidateUnit(m, pool))
			}
			kinds := fcKinds(t)
			for side := range kinds.Factions {
				for i := range kinds.Factions[side].Kinds {
					k := &kinds.Factions[side].Kinds[i]
					k.Words = make([]int, 13)
					for w := range k.Words {
						k.Words[w] = fcWord(m, fcDS+0x4e3a+side*208+k.Kind*26+w*2)
					}
				}
			}
			profiles := append([]fleetWeaponProfile(nil), fleetWeaponProfiles[:]...)
			for i := range profiles {
				profiles[i].major = fcWord(m, fcDS+0x4fda+i*6+2)
				profiles[i].minor = fcWord(m, fcDS+0x4fda+i*6+4)
			}
			terrain := &BattleTerrain{BlockMask: bytes.Clone(m[fcDS+0x9bb4 : fcDS+0x9bb4+256]),
				QuadrantBits: bytes.Clone(m[fcDS+0x3c3e : fcDS+0x3c3e+4])}
			ptr := fcWord(m, fcDS+0x3c1a) + 16*fcWord(m, fcDS+0x3c1c)
			cells := m[ptr : ptr+128*102]
			for n, query := range q {
				pool := (query.regs["SI"] - 2) / 78
				own := -1
				for i, p := range pools {
					if p == pool {
						own = i
					}
				}
				if own < 0 {
					t.Fatalf("原版查詢分隊%d不在物件清單", pool)
				}
				before, ramBefore := append([]BattleUnit(nil), units...), bytes.Clone(m)
				got, err := fleetCombatCandidates(units, own, fcWord(m, fcDS+0xc074+units[own].Fleet*122+0x4a), kinds, terrain, cells, profiles)
				if err != nil {
					t.Fatal(err)
				}
				fcEqual(t, "candidate count", len(got), returns[n].regs["AX"])
				masks := map[int]int{}
				for _, v := range got {
					masks[pools[v.Target]] = int(v.Mask)
				}
				writes := 0
				for _, h := range obs["1F6F:0382"] {
					if h.step <= query.step || n+1 < len(q) && h.step >= q[n+1].step {
						continue
					}
					writes++
					fcEqual(t, fmt.Sprintf("query%d target%d mask", n, h.regs["DI"]), masks[h.regs["DI"]], h.regs["AX"]&255)
					delete(masks, h.regs["DI"])
				}
				if len(masks) != 0 {
					t.Fatalf("remake候選沒有原版逐查詢寫入：%v", masks)
				}
				if !reflect.DeepEqual(units, before) || !bytes.Equal(m, ramBefore) {
					t.Fatal("候選查詢改寫分隊、地形或RAM")
				}
				queries++
				t.Logf("query%d pool%d count%d writes%d", n, pool, len(got), writes)
			}
		})
	}
	if count != 48 {
		t.Fatalf("候選案例數%d，應為48", count)
	}
	t.Logf("48案、%d次逐查詢原版AX／mask符合", queries)
}

func TestFleetCombatCandidatesInvalidInput(t *testing.T) {
	u := []BattleUnit{{Faction: 0, Kind: 0, Action: 127, X: 10, Y: 10}, {Faction: 1, Kind: 0, Detail: 3, Action: 127, X: 30, Y: 10}}
	kinds := fcKinds(t)
	terrain := &BattleTerrain{BlockMask: make([]byte, 128), QuadrantBits: []byte{1, 2, 4, 8}}
	cells := make([]byte, 128*102)
	for _, tc := range []string{"own", "facing", "map", "terrain", "kind", "type", "coordinate", "stock", "tile"} {
		t.Run(tc, func(t *testing.T) {
			v := append([]BattleUnit(nil), u...)
			k := fcKinds(t)
			terr := *terrain
			m := bytes.Clone(cells)
			own, facing := 0, 0
			switch tc {
			case "own":
				own = -1
			case "facing":
				facing = 256
			case "map":
				m = m[:100]
			case "terrain":
				terr.QuadrantBits = nil
			case "kind":
				v[0].Kind = 8
			case "type":
				k.Factions[0].Kinds[0].Words[5] = 12
			case "coordinate":
				v[1].Y = 2040
			case "stock":
				v[0].Missiles = -1
			case "tile":
				m[0] = 128
			}
			before, mapBefore := append([]BattleUnit(nil), v...), bytes.Clone(m)
			got, err := FleetCombatCandidates(v, own, facing, k, &terr, m)
			if err == nil || got != nil || !reflect.DeepEqual(v, before) || !bytes.Equal(m, mapBefore) {
				t.Fatal("非法輸入未失敗或改寫輸入")
			}
		})
	}
	if got, err := FleetCombatCandidates(u, 0, 0, kinds, nil, cells); err == nil || got != nil {
		t.Fatal("缺地形被當開放射線")
	}
	if got, err := FleetCombatCandidates(u, 0, 0, nil, terrain, cells); err == nil || got != nil {
		t.Fatal("缺艦種表被當作有效武器")
	}
}
