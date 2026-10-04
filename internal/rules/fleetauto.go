package rules

import "fmt"

// FleetAutoCandidate是規格101的預估輸入。Target是呼叫端索引；
// Anticipated來自原池+3E，Command提供兩側能力、目標朝向及武器。
type FleetAutoCandidate struct {
	Target      int
	Unit        BattleUnit
	Mask        uint8
	Anticipated uint16
	Command     FleetCombatCommand
}

type FleetAutoEvaluation struct {
	Target       int
	Mask         uint8
	WeaponScores [3]uint16
	Score        uint16
}

type FleetAutoSelection struct {
	Target      int
	Weapon      uint8
	Anticipated uint16
	Choice      int
	// Evaluations按輸入次序；Order保存降序排列的輸入索引。
	Evaluations []FleetAutoEvaluation
	Order       []int
}

func fleetAutoUnit(kinds *ShipKindTable, u *BattleUnit) (ShipKind, error) {
	if u == nil || u.Faction < 0 || u.Faction > 1 || u.Kind < 0 || u.Kind > 7 ||
		u.Ships < 0 || u.Ships > 65535 || u.X < 0 || u.X > 2047 || u.Y < 0 || u.Y > 2039 ||
		u.Missiles < 0 || uint64(u.Missiles) > 0xffffffff || u.Fighters < 0 || uint64(u.Fighters) > 0xffffffff {
		return ShipKind{}, fmt.Errorf("自動交戰分隊欄位無效")
	}
	return fleetCombatKind(kinds, u)
}

// fleetAutoForecast保持0714的unsigned32運算，與093實際傷害的signed32、
// 噪音及旗艦修正分開。預估不消耗彈藥或RND。
type fleetAutoForecastResult struct {
	Attack, Defence, Shots, Percent int
	Score                           uint16
}

func fleetAutoForecast(own BattleUnit, c FleetAutoCandidate, kinds *ShipKindTable, weapon int) (fleetAutoForecastResult, error) {
	var result fleetAutoForecastResult
	ok, err := fleetAutoUnit(kinds, &own)
	if err != nil {
		return result, err
	}
	tk, err := fleetAutoUnit(kinds, &c.Unit)
	if err != nil {
		return result, err
	}
	typeID := ok.Byte(10 + weapon*4)
	if typeID < 0 || typeID >= len(fleetWeaponProfiles) || tk.SemiMajor <= 0 || tk.SemiMinor <= 0 || tk.SemiMinor > tk.SemiMajor {
		return result, fmt.Errorf("自動交戰武器或裝甲表無效")
	}
	attack, _, _, err := fleetCombatAttack(fleetWeaponProfiles[typeID], &own, &c.Unit, c.Command.AttackAbilities[weapon])
	if err != nil {
		return result, err
	}
	defence := AttackCoefficient(tk.SemiMajor, tk.SemiMinor,
		AngleDiff256(Atan256(own.X-c.Unit.X, own.Y-c.Unit.Y), c.Command.TargetFacing), c.Command.DefendAbilities[weapon])
	if defence <= 0 {
		return result, fmt.Errorf("自動交戰防禦係數為零")
	}
	shots := ok.Byte(11+weapon*4) * own.Ships
	switch weapon {
	case 1:
		shots = min(shots, own.Missiles)
	case 2:
		shots = min(shots, own.Fighters)
		if own.Kind == 3 {
			shots /= 2
		}
	}
	percent := min(int(uint16(uint32(uint64(shots)*uint64(attack))/uint32(defence)/750)), 100)
	if c.Command.TargetWeapon == 3 {
		percent /= 2
	}
	result.Attack, result.Defence, result.Shots, result.Percent = attack, defence, shots, percent
	result.Score = uint16(int(c.Anticipated) * percent / 100)
	return result, nil
}

// SelectFleetAutoCommand不改分隊資料；所有錯誤均在RND前返回。
// computer是攻方座位的控制模式，selectionAbility是FLEET+45原byte。
func SelectFleetAutoCommand(own BattleUnit, anticipated uint16, candidates []FleetAutoCandidate,
	kinds *ShipKindTable, computer bool, selectionAbility int, r *Rand) (FleetAutoSelection, error) {
	result := FleetAutoSelection{Target: -1, Weapon: 3, Anticipated: anticipated, Choice: -1}
	if len(candidates) == 0 {
		return result, nil
	}
	if r == nil || selectionAbility < 0 || selectionAbility > 255 {
		return FleetAutoSelection{}, fmt.Errorf("自動交戰缺亂數或選擇倍率無效")
	}
	if _, err := fleetAutoUnit(kinds, &own); err != nil {
		return FleetAutoSelection{}, err
	}
	seen := map[int]bool{}
	lastWeapon := uint8(255)
	for _, c := range candidates {
		if c.Target < 0 || seen[c.Target] || c.Unit.Faction == own.Faction || c.Mask == 0 || c.Mask&^7 != 0 ||
			c.Command.TargetFacing < 0 || c.Command.TargetFacing > 255 ||
			(c.Command.TargetWeapon != 255 && (c.Command.TargetWeapon < 0 || c.Command.TargetWeapon > 3)) {
			return FleetAutoSelection{}, fmt.Errorf("自動交戰候選或防禦欄位無效")
		}
		seen[c.Target] = true
		for _, abilities := range [][3]int{c.Command.AttackAbilities, c.Command.DefendAbilities} {
			for _, a := range abilities {
				if a < 0 || a > 255 {
					return FleetAutoSelection{}, fmt.Errorf("自動交戰能力超出byte")
				}
			}
		}
		e := FleetAutoEvaluation{Target: c.Target, Mask: c.Mask}
		for weapon := 0; weapon < 3; weapon++ {
			if c.Mask&(1<<weapon) == 0 {
				continue
			}
			score, err := fleetAutoForecast(own, c, kinds, weapon)
			if err != nil {
				return FleetAutoSelection{}, err
			}
			e.WeaponScores[weapon] = score.Score
			e.Score = max(e.Score, score.Score)
			lastWeapon = uint8(weapon)
		}
		if computer {
			switch {
			case c.Unit.Squadron == 0:
				e.Score *= 3
			case c.Unit.Kind == 6:
				e.Score *= 2
			case c.Unit.Kind == 5:
				e.Score = (e.Score * 3) / 2
			}
		}
		result.Order = append(result.Order, len(result.Evaluations))
		result.Evaluations = append(result.Evaluations, e)
	}
	gap := len(result.Order)
	for {
		gap = max(1, gap*10/13)
		if gap == 9 || gap == 10 {
			gap = 11
		}
		swapped := false
		for i := 0; i+gap < len(result.Order); i++ {
			if result.Evaluations[result.Order[i]].Score < result.Evaluations[result.Order[i+gap]].Score {
				result.Order[i], result.Order[i+gap] = result.Order[i+gap], result.Order[i]
				swapped = true
			}
		}
		if gap == 1 && !swapped {
			break
		}
	}
	value := uint16(r.Next() * selectionAbility / 100)
	result.Choice = 0
	if value < 8192 && len(candidates) >= 3 {
		result.Choice = 2
	} else if value < 16384 && len(candidates) >= 2 {
		result.Choice = 1
	}
	selected := result.Evaluations[result.Order[result.Choice]]
	result.Target, result.Weapon = selected.Target, lastWeapon
	best := uint16(65535)
	if selected.Mask&1 != 0 {
		result.Weapon, best = 0, selected.WeaponScores[0]
	}
	for weapon := 1; weapon < 3; weapon++ {
		if selected.Mask&(1<<weapon) != 0 && selected.WeaponScores[weapon] > best {
			result.Weapon, best = uint8(weapon), selected.WeaponScores[weapon]
		}
	}
	// 原版018B扣自身暫存；缺beam時FFFF造成+1回繞，仍沿最後預估武器。
	result.Anticipated -= best
	return result, nil
}
