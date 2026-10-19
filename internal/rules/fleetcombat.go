package rules

import "fmt"

// 艦隊對艦隊的純規則核心（規格093、研究402至404）。完整第④步的
// 命令、演出與保存仍由規格092驗收；本檔不替玩家或CPU選擇命令。
type fleetWeaponProfile struct {
	power, major, minor int
	y                   [6]int
}

// 原版DS:4FDA及DS:477C指向的六點曲線。第六點不是1.0，而是
// 0x7fffffff；刪除它會改變超出第五點後的原版定點插值。
var fleetWeaponProfiles = [...]fleetWeaponProfile{
	{180, 100, 40, [6]int{10, 30, 60, 100, 90, 0}},
	{140, 100, 40, [6]int{20, 40, 60, 100, 90, 0}},
	{70, 80, 36, [6]int{25, 50, 75, 100, 100, 0}},
	{45, 90, 28, [6]int{60, 80, 95, 100, 105, 0}},
	{20, 60, 20, [6]int{40, 80, 100, 80, 40, 0}},
	{300, 80, 60, [6]int{30, 60, 80, 100, 90, 0}},
	{280, 120, 55, [6]int{30, 60, 90, 110, 100, 0}},
	{50, 70, 64, [6]int{40, 80, 100, 100, 95, 0}},
	{220, 66, 55, [6]int{30, 50, 90, 100, 100, 0}},
	{160, 80, 74, [6]int{40, 60, 100, 100, 95, 0}},
	{45, 60, 60, [6]int{100, 100, 100, 90, 80, 0}},
	{60, 60, 60, [6]int{100, 100, 100, 90, 80, 0}},
}

var fleetCurveX = [...]Fixed{0, 16384, 32768, 49152, 65536, 0x7fffffff}
var fleetOrderBases = [...]int{50, 36, 23, 10, 12}
var fleetMobilityCoefficients = [...]int{0, 0, 7, 14, 21, 28, 35, 42, 49, 56, 63}

// FleetCombatCommand帶入已提交命令與FLEET／ADML欄位。三能力分別是
// FLEET+3D／3E／3F，TargetAttribute是目標提督ADML+12。
type FleetCombatCommand struct {
	Weapon             int
	AttackAbilities    [3]int
	DefendAbilities    [3]int
	TargetFacing       int
	TargetWeapon       int
	TargetAttribute    int
	TargetComputer     bool
	AttackBonus        int // 規格149的remake加成，零值保持原版核心算式。
	FighterAttackBonus int // 僅武器2；提督／參謀所在艦隊的獨立倍率。
}

// FleetCombatHit保留演出所需係數與逐擊增量。呼叫端依Scorer累加兩張
// 分數表，寫提督／副官功績，再刷新艦隊完整度；本函式只修改兩支分隊。
type FleetCombatHit struct {
	Skipped            bool
	Attack, Defence    int
	NormalizedDistance Fixed
	CurveY, Shots      int
	Percent, Requested int
	Lost, Scorer       int
	Score              int
	AdmiralMerit       int
	AdjunctMerit       int
	Fallback           bool
}

func fleetCombatKind(kinds *ShipKindTable, u *BattleUnit) (ShipKind, error) {
	if u == nil {
		return ShipKind{}, fmt.Errorf("交戰分隊缺失")
	}
	k, ok := kinds.Kind(u.Faction, u.Kind)
	if !ok || len(k.Words) != 13 {
		return ShipKind{}, fmt.Errorf("交戰艦種表缺失：陣營%d艦種%d", u.Faction, u.Kind)
	}
	return k, nil
}

func fleetCombatAttack(p fleetWeaponProfile, own, target *BattleUnit, ability int) (int, Fixed, int, error) {
	normalizer := FixedMul(FixedFromInt(p.major), FixedOne+EllipseEccentricity(p.major, p.minor))
	q := FixedDiv(FixedDist(own.X, own.Y, target.X, target.Y), normalizer)
	if normalizer <= 0 || q < 0 {
		return 0, 0, 0, fmt.Errorf("交戰距離超出定點數範圍")
	}
	// 使用者2026-10-04裁決：修正原版q=0讀未初始化stack指標。
	y := p.y[0]
	if q != 0 {
		for i := 1; i < len(fleetCurveX); i++ {
			if q > fleetCurveX[i] {
				continue
			}
			f := FixedDiv(q-fleetCurveX[i-1], fleetCurveX[i]-fleetCurveX[i-1])
			// 原版203A:00EC取signed32乘積的高word；負斜率向下截斷。
			y = p.y[i-1] + int(int32(int64(f)*int64(p.y[i]-p.y[i-1]))>>16)
			break
		}
	}
	return y * p.power / 100 * ability / 100, q, y, nil
}

// ApplyFleetCombatDamage依1DD6:0C54提交一擊。錯誤在庫存與RND變動前
// 回傳。FF／3及零艘數沿原版呼叫端略過，不消費傷害亂數。
func ApplyFleetCombatDamage(own, target *BattleUnit, kinds *ShipKindTable, c FleetCombatCommand, r *Rand) (FleetCombatHit, error) {
	var h FleetCombatHit
	if own == nil || target == nil {
		return h, fmt.Errorf("交戰分隊缺失")
	}
	if c.Weapon == 0xff || c.Weapon == 3 || own.Ships == 0 || target.Ships == 0 {
		h.Skipped = true
		return h, nil
	}
	if c.Weapon < 0 || c.Weapon > 2 || r == nil || own == target {
		return h, fmt.Errorf("交戰命令、目標或亂數無效")
	}
	for _, u := range []*BattleUnit{own, target} {
		if u.Ships < 0 || u.Ships > 0xffff || u.FullShips < 0 || u.FullShips > 0xffff ||
			u.RepairCap < 0 || u.RepairCap > 0xffff || u.Missiles < 0 || u.Fighters < 0 ||
			uint64(u.Missiles) > 0xffffffff || uint64(u.Fighters) > 0xffffffff ||
			u.X < 0 || u.X > 0x7fff || u.Y < 0 || u.Y > 0x7fff {
			return h, fmt.Errorf("交戰分隊欄位超出原版數值範圍")
		}
	}
	if target.FullShips == 0 || c.TargetFacing < 0 || c.TargetFacing > 255 ||
		c.TargetAttribute < 0 || c.TargetAttribute > 255 ||
		(c.TargetWeapon != 0xff && (c.TargetWeapon < 0 || c.TargetWeapon > 3)) {
		return h, fmt.Errorf("交戰防禦欄位無效")
	}
	for _, a := range [][3]int{c.AttackAbilities, c.DefendAbilities} {
		for _, v := range a {
			if v < 0 || v > 255 {
				return h, fmt.Errorf("交戰能力超出byte範圍")
			}
		}
	}
	ok, err := fleetCombatKind(kinds, own)
	if err != nil {
		return h, err
	}
	tk, err := fleetCombatKind(kinds, target)
	if err != nil {
		return h, err
	}
	typeID := ok.Byte(10 + c.Weapon*4)
	if typeID < 0 || typeID >= len(fleetWeaponProfiles) {
		return h, fmt.Errorf("未搭載或未知交戰武器type%d", typeID)
	}
	if tk.SemiMajor <= 0 || tk.SemiMinor <= 0 || tk.SemiMinor > tk.SemiMajor {
		return h, fmt.Errorf("目標裝甲橢圓無效")
	}
	h.Attack, h.NormalizedDistance, h.CurveY, err = fleetCombatAttack(fleetWeaponProfiles[typeID], own, target, c.AttackAbilities[c.Weapon])
	if err != nil {
		return h, err
	}
	h.Defence = AttackCoefficient(tk.SemiMajor, tk.SemiMinor,
		AngleDiff256(Atan256(own.X-target.X, own.Y-target.Y), c.TargetFacing), c.DefendAbilities[c.Weapon])
	if h.Defence <= 0 {
		return h, fmt.Errorf("交戰防禦係數為零")
	}
	if c.AttackBonus < 0 || c.AttackBonus > 200 {
		return h, fmt.Errorf("攻擊加成超出範圍")
	}
	h.Attack = BonusAmount(h.Attack, c.AttackBonus)
	if c.FighterAttackBonus < 0 || c.FighterAttackBonus > 200 {
		return h, fmt.Errorf("戰鬥機攻擊加成超出範圍")
	}
	if c.Weapon == 2 {
		h.Attack = BonusAmount(h.Attack, c.FighterAttackBonus)
	}
	h.Shots = ok.Byte(11+c.Weapon*4) * own.Ships
	switch c.Weapon {
	case 1:
		h.Shots = min(h.Shots, own.Missiles)
		own.Missiles -= h.Shots
	case 2:
		h.Shots = min(h.Shots, own.Fighters)
		own.Fighters -= h.Shots * 20 / 100
		if own.Kind == 3 {
			h.Shots /= 2
		}
	}
	h.Percent = int(uint16(int32(int64(h.Shots)*int64(h.Attack)) / int32(h.Defence) / 750))
	h.Percent = min(h.Percent, 100)
	if c.TargetWeapon == 3 {
		h.Percent /= 2
	}
	if target.Squadron == 0 {
		if c.TargetAttribute > 100 {
			h.Percent = h.Percent * 100 / c.TargetAttribute
		} else if h.Percent >= 80 && c.TargetComputer {
			h.Percent -= 10
		}
	}
	noise := r.Next()%200 - c.TargetAttribute
	if target.Ships < 100 {
		noise /= 2
	}
	h.Requested = max(0, int(int16(target.Ships*h.Percent/100+noise)))
	if h.Requested == 0 {
		h.Fallback = true
		h.Requested = r.Next() % 10
	}
	h.Lost = ApplyShipLoss(target, tk, h.Requested)
	h.Scorer = 1 - target.Faction
	h.Score, h.AdmiralMerit, h.AdjunctMerit = h.Lost/20, h.Lost/2, h.Lost/4
	return h, nil
}

// FleetCombatOrderEntry按原版FLEET編制順序提供，包含FF／3命令的分隊。
// Attribute是該艦隊提督ADML+0B，Mobility沿FleetCombatMobility計算。
type FleetCombatOrderEntry struct {
	Unit      BattleUnit
	Attribute int
	Mobility  int
}

// FleetCombatMobility保留297B:083F的lowword乘法、signed除法與AL回值。
// operations取提督及非FF副官ADML+17較大值，groups是一般分隊數。
func FleetCombatMobility(operations, groups int) (int, error) {
	if operations < 0 || operations > 255 || groups < 0 || groups >= len(fleetMobilityCoefficients) {
		return 0, fmt.Errorf("交戰機動力欄位無效")
	}
	return int(byte(100 - int(int16((270-2*operations)*fleetMobilityCoefficients[groups]))/100)), nil
}

// FleetCombatOrder回傳輸入索引的降序排列及按輸入索引保存的unsigned鍵。
// 先驗全部條目，成功才逐筆取兩個RND；不先篩無攻擊命令。
func FleetCombatOrder(entries []FleetCombatOrderEntry, kinds *ShipKindTable, r *Rand) ([]int, []uint16, error) {
	if r == nil {
		return nil, nil, fmt.Errorf("交戰排序缺少亂數")
	}
	bases := make([]int, len(entries))
	for i := range entries {
		e := &entries[i]
		k, err := fleetCombatKind(kinds, &e.Unit)
		if err != nil {
			return nil, nil, err
		}
		class := k.Byte(9)
		if class >= len(fleetOrderBases) || e.Attribute <= 0 || e.Attribute > 255 ||
			e.Unit.Morale < 0 || e.Unit.Morale > 255 || e.Mobility < 0 || e.Mobility > 255 {
			return nil, nil, fmt.Errorf("交戰排序第%d筆欄位無效", i)
		}
		bases[i] = fleetOrderBases[class]
	}
	order, keys := make([]int, len(entries)), make([]uint16, len(entries))
	for i, e := range entries {
		order[i] = i
		keys[i] = uint16(r.Next()%bases[i] + bases[i] + r.Next()%e.Attribute + e.Attribute + e.Unit.Morale + e.Mobility)
	}
	gap := len(order)
	for {
		gap = max(1, gap*10/13)
		if gap == 9 || gap == 10 {
			gap = 11
		}
		swapped := false
		for i := 0; i+gap < len(order); i++ {
			if keys[order[i]] < keys[order[i+gap]] {
				order[i], order[i+gap] = order[i+gap], order[i]
				swapped = true
			}
		}
		if gap == 1 && !swapped {
			return order, keys, nil
		}
	}
}
