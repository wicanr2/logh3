package rules

// 戰場索敵（`2526:000E`／`2526:00F8`，`docs/re/290`）。

const (
	FlagshipVision = 200
	ScoutVision    = 250
	SpottingMax    = 3
)

// DecaySpotting 是回合開頭的衰減：上一輪沒被看到才降級，然後清 Seen。
func DecaySpotting(units []BattleUnit) {
	for i := range units {
		u := &units[i]
		if u.Seen {
			u.Seen = false
			continue
		}
		switch u.Detail {
		case 3:
			if u.DetailHold > 0 {
				u.DetailHold--
			}
			if u.DetailHold == 0 {
				u.Detail = 2
			}
		case 2:
			u.Detail = 1
		case 1:
			u.Detail = 0
		}
	}
}

// UpdateSpotting 是回合末的兩方索敵。每支艦隊只有旗艦是 200 像素的眼睛；
// 已放出的偵察隊是 250 像素的眼睛；有主行星的半徑由防衛力夾限後除 25。
// 候選位置讀 TargetX／TargetY，偵察隊自己不會被看到。
func UpdateSpotting(units []BattleUnit, planets []BattlePlanet) {
	observe := func(faction, x, y, radius int) {
		for i := range units {
			u := &units[i]
			if u.Faction == faction || u.Out || u.Kind == 7 || u.Seen {
				continue
			}
			if Distance16(x, y, u.TargetX, u.TargetY) > int64(radius*FixOne) {
				continue
			}
			u.Seen = true
			if u.Detail < SpottingMax {
				u.Detail++
				if u.Detail == SpottingMax {
					u.DetailHold = 2
				}
			}
		}
	}
	for i := range units {
		u := &units[i]
		switch {
		case u.Squadron == 0:
			observe(u.Faction, u.TargetX, u.TargetY, FlagshipVision)
		case u.Kind == 7:
			observe(u.Faction, u.TargetX, u.TargetY, ScoutVision)
		}
	}
	for _, p := range planets {
		if p.Owner == 0 || p.Owner == 1 {
			observe(p.Owner, p.X, p.Y, PlanetQueryRadius(p.Defence))
		}
	}
}

// RevealFleet 顯示一支艦隊的所有一般分隊。打有主行星時 hold=2；AI 的
// 第七回合防僵局特例使用 RevealEnemiesIfBlind 的 hold=6。
func RevealFleet(units []BattleUnit, fleet, hold int) {
	for i := range units {
		if units[i].Fleet == fleet && units[i].Kind != 7 {
			units[i].Detail = SpottingMax
			units[i].DetailHold = hold
		}
	}
}

// RevealEnemiesIfBlind 是 `3009:008D`：畫面第七回合起，AI 若完全看不到
// 任何敵方一般分隊，就把敵方全部揭露六輪並標成這輪已看到。
func RevealEnemiesIfBlind(units []BattleUnit, observerFaction, turn int) bool {
	if turn < 7 {
		return false
	}
	for i := range units {
		u := &units[i]
		if u.Faction != observerFaction && u.Kind != 7 && u.Detail != 0 {
			return false
		}
	}
	changed := false
	for i := range units {
		u := &units[i]
		if u.Faction == observerFaction || u.Kind == 7 {
			continue
		}
		u.Seen = true
		u.Detail = SpottingMax
		u.DetailHold = 6
		changed = true
	}
	return changed
}
