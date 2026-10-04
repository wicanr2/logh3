package rules

import (
	"reflect"
	"testing"
)

const fleetAutoReceiptSHA = "23627c23a9d2ef93991c842e00fbf43df75f00d2043c78788454e868d3a19401"

func autoNativeInputs(t *testing.T, m fcManifest, name string) ([]byte, []byte, *ShipKindTable, []FleetAutoCandidate, map[string][]fcObservation) {
	t.Helper()
	dir := "workplace/oracle/command79-r75-" + name + "-a"
	ram, end := m.ram(t, dir+"/input-mem.bin"), m.ram(t, dir+"/end-mem.bin")
	obs := fcObservations(t, m.read(t, dir+"/report.txt"))
	m.read(t, dir+"/flags.txt")
	m.read(t, dir+"/meta.txt")
	kinds := fcKinds(t)
	for side := range kinds.Factions {
		for i := range kinds.Factions[side].Kinds {
			k := &kinds.Factions[side].Kinds[i]
			for j := range k.Words {
				k.Words[j] = fcWord(ram, fcDS+0x4e3a+side*208+k.Kind*26+j*2)
			}
		}
	}
	var candidates []FleetAutoCandidate
	// 候選原始次序直接取00BB的callee目標far pointer，而不是已排序末端列表。
	for _, h := range obs["1F6F:00BB"] {
		pool := (h.bp[8] + 16*h.bp[10] - fcPool) / 78
		u := candidateUnit(ram, pool)
		of, tf := fcDS+0xc074+12*122, fcDS+0xc074+u.Fleet*122
		c := FleetCombatCommand{TargetWeapon: int(ram[fcPool+pool*78+0x2e]), TargetFacing: int(ram[tf+0x4a])}
		for j := range c.AttackAbilities {
			c.AttackAbilities[j], c.DefendAbilities[j] = int(ram[of+0x3d+j]), int(ram[tf+0x3d+j])
		}
		candidates = append(candidates, FleetAutoCandidate{Target: pool, Unit: u,
			Mask: end[fcPool+pool*78+0x3c], Anticipated: uint16(fcWord(ram, fcPool+pool*78+0x3e)), Command: c})
	}
	return ram, end, kinds, candidates, obs
}

func TestFleetAutoCommandMatchesNative(t *testing.T) {
	fcRequireOriginal(t)
	m := fcLoadManifest(t, "fleet-auto-command.json", fleetAutoReceiptSHA)
	fcEqual(t, "自動命令案例數", len(m.Cases), 31)
	forecasts := 0
	for _, row := range m.Cases {
		t.Run(row.Name, func(t *testing.T) {
			ram, end, kinds, candidates, obs := autoNativeInputs(t, m, row.Name)
			own := fcUnit(ram, 80)
			before := append([]FleetAutoCandidate(nil), candidates...)
			r := NewRand(uint32(fcLong(ram, fcDS+0x94dc)))
			computer := ram[fcDS+0x8f21] == 1 // 本批攻方固定座位0，不能外推陣營1。
			got, err := SelectFleetAutoCommand(own, uint16(fcWord(ram, fcPool+80*78+0x3e)), candidates,
				kinds, computer, int(ram[fcDS+0xc074+own.Fleet*122+0x45]), r)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, candidates) && len(candidates) > 0 || own != fcUnit(ram, 80) {
				t.Fatal("自動預估改寫輸入分隊")
			}
			fcEqual(t, "weapon", int(got.Weapon), int(end[fcPool+80*78+0x2e]))
			fcEqual(t, "own anticipated", int(got.Anticipated), fcWord(end, fcPool+80*78+0x3e))
			fcEqual(t, "seed", int(r.Seed()), fcLong(end, fcDS+0x94dc))
			if len(candidates) == 0 {
				fcEqual(t, "no candidate", got.Target, -1)
				fcEqual(t, "target retained", fcWord(end, fcPool+80*78+0x2c), fcWord(ram, fcPool+80*78+0x2c))
				return
			}
			fcEqual(t, "target", got.Target, fcWord(end, fcPool+80*78+0x2c))
			fcEqual(t, "choice", got.Choice, fcValue(t, obs, "1F6F:00F0", 0, 0, false))
			n := 0
			for i, c := range candidates {
				e := got.Evaluations[i]
				fcEqual(t, "sorted pool", candidates[got.Order[i]].Target, fcWord(end, fcDS+0xa49a+i*2))
				fcEqual(t, "score", int(e.Score), fcWord(end, fcPool+c.Target*78+0x46))
				for w := 0; w < 3; w++ {
					fcEqual(t, "weapon score", int(e.WeaponScores[w]), fcWord(end, fcPool+c.Target*78+0x48+w*2))
					if c.Mask&(1<<w) == 0 {
						continue
					}
					f, err := fleetAutoForecast(own, c, kinds, w)
					if err != nil {
						t.Fatal(err)
					}
					fcEqual(t, "attack", f.Attack, fcValue(t, obs, "1F6F:072F", n, 0, false))
					fcEqual(t, "defence", f.Defence, fcValue(t, obs, "1F6F:0745", n, 0, false))
					fcEqual(t, "shots", f.Shots, fcValue(t, obs, "1F6F:086C", n, 10, true)+65536*fcValue(t, obs, "1F6F:086C", n, 12, true))
					fcEqual(t, "percent", f.Percent, fcValue(t, obs, "1F6F:08CC", n, 4, true))
					fcEqual(t, "forecast", int(f.Score), fcValue(t, obs, "1F6F:08CC", n, 6, true))
					n++
				}
			}
			fcEqual(t, "forecast count", n, len(obs["1F6F:08CC"]))
			forecasts += n
		})
	}
	t.Logf("31案，%d筆預估係數／百分比／結果直接符合原版", forecasts)
}

func TestFleetAutoCommandRejectsBeforeRandom(t *testing.T) {
	fcRequireOriginal(t)
	m := fcLoadManifest(t, "fleet-auto-command.json", fleetAutoReceiptSHA)
	for _, name := range []string{"rand", "ability", "target", "mask", "side", "kind", "weapon", "defence", "attack", "facing", "coordinate", "stock"} {
		t.Run(name, func(t *testing.T) {
			ram, _, kinds, c, _ := autoNativeInputs(t, m, "base")
			own := fcUnit(ram, 80)
			r := NewRand(0x12345678)
			arg, ability := r, 100
			switch name {
			case "rand":
				arg = nil
			case "ability":
				ability = 256
			case "target":
				c = append(c, c[0])
			case "mask":
				c[0].Mask = 8
			case "side":
				c[0].Unit.Faction = own.Faction
			case "kind":
				c[0].Unit.Kind = 8
			case "weapon":
				c[0].Command.TargetWeapon = 4
			case "defence":
				c[0].Command.DefendAbilities[1] = 0 // 在第一個有效預估之後失敗，仍不得取RND。
			case "attack":
				c[0].Command.AttackAbilities[0] = 256
			case "facing":
				c[0].Command.TargetFacing = -1
			case "coordinate":
				own.X = 2048
			case "stock":
				own.Fighters = -1
			}
			before := append([]FleetAutoCandidate(nil), c...)
			if _, err := SelectFleetAutoCommand(own, 3000, c, kinds, false, ability, arg); err == nil {
				t.Fatal("錯誤輸入被接受")
			}
			if r.Seed() != 0x12345678 || !reflect.DeepEqual(before, c) {
				t.Fatal("失敗消費RND或改寫候選")
			}
		})
	}
	// 無候選無需亂數與資料表，也不改anticipated。
	got, err := SelectFleetAutoCommand(BattleUnit{}, 1234, nil, nil, false, 0, nil)
	if err != nil || got.Target != -1 || got.Weapon != 3 || got.Anticipated != 1234 {
		t.Fatal("無候選不符合防禦分支")
	}
}
