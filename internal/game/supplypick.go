package game

import "github.com/wicanr2/logh3/internal/rules"

// 「補給」與「修理」的選目標（原版 `1930:05EC`、`1AAB:0583`，
// `docs/re/131`）。兩者只差候選的條件與動手時搬什麼。

// SupplyPickPending 回「正在等玩家點補給／修理的對象」。
func (g *Game) SupplyPickPending() bool { return g.battleSupply }

// SupplyPickRepair 回這一次是不是「修理」。
func (g *Game) SupplyPickRepair() bool { return g.battleSupplyRepair }

// beginPlanetSupplyPick 進入「行星補給／補充」的選對象狀態。
//
// 與艦艇版同一個骨架，只換中心、陣營與半徑（`docs/re/134`）：中心是
// 行星在戰場上的位置，陣營是行星的擁有者，半徑是
// `rules.PlanetQueryRadius(防衛力)`。
func (g *Game) beginPlanetSupplyPick(planet int, replenish bool) {
	if planet < 0 || planet >= len(g.battlePlanets) {
		return
	}
	if len(g.planetSupplyCandidates(planet)) == 0 {
		return
	}
	g.battleSupply, g.battleSupplyPlanet = true, planet
	g.battleSupplyRepair = replenish
}

// planetSupplyCandidates 回行星半徑內可以補的分隊。篩選條件與艦艇版的
// 「補給」逐條相同（`1930:10B2` 對 `1930:0485`）。
func (g *Game) planetSupplyCandidates(planet int) []int {
	if planet < 0 || planet >= len(g.battlePlanets) {
		return nil
	}
	p := g.battlePlanets[planet]
	r := rules.PlanetQueryRadius(p.Defence)
	var out []int
	for i := range g.battleUnits {
		u := g.battleUnits[i]
		if u.Faction != p.Owner || u.Out {
			continue
		}
		if u.Kind == shipKindSupply || u.Kind == shipKindRepair || u.Kind == shipKindScout {
			continue
		}
		dx, dy := u.X-p.X, u.Y-p.Y
		if dx*dx+dy*dy > r*r {
			continue
		}
		out = append(out, i)
	}
	return out
}

// beginSupplyPick 進入選對象的狀態；一個候選都沒有就不進去。
//
// 原版這時會跳「沒有需要補給的船艦。」（`ds:3F14`）／「沒有需要修理的
// 船艦在附近。」（`ds:403E`）的訊息框——版面還沒量，所以這裡只是不進去
// （比照 `docs/spec/024` 的作法，差異記在規格裡）。
func (g *Game) beginSupplyPick(unit int, repair bool) {
	if len(g.supplyCandidates(unit, repair)) == 0 {
		return
	}
	g.battleSupply, g.battleSupplyUnit, g.battleSupplyRepair = true, unit, repair
	g.battleSupplyPlanet = -1
}

// cancelSupplyPick 取消選對象。
func (g *Game) cancelSupplyPick() { g.battleSupply, g.battleSupplyPlanet = false, -1 }

// SupplyCandidates 回現在可以點的分隊索引。
func (g *Game) SupplyCandidates() []int {
	if !g.battleSupply {
		return nil
	}
	if g.battleSupplyPlanet >= 0 {
		return g.planetSupplyCandidates(g.battleSupplyPlanet)
	}
	return g.supplyCandidates(g.battleSupplyUnit, g.battleSupplyRepair)
}

// supplyCandidates 照原版收候選：`17A0:00AC` 先用半徑與陣營篩一遍，
// 再套 `1930:0485`／`1AAB:043D` 那個篩選函式。
//
//   - 同陣營、`+0x1E` bit7 沒設（沒出局）、距離 ≤ 收集半徑；
//   - 補給另外排除艦種 5／6／7，修理排除 6／7（**修理修得了運補艦**）。
//
// 發令的分隊自己距離 0，補給時會被艦種 5 那條排掉；修理時排不掉
// ——照原版留著。
func (g *Game) supplyCandidates(unit int, repair bool) []int {
	if unit < 0 || unit >= len(g.battleUnits) || g.res.ShipKinds == nil {
		return nil
	}
	me := g.battleUnits[unit]
	kind := rules.SupplyKind
	if repair {
		kind = rules.RepairKind
	}
	r := rules.SupplyListRange(g.res.ShipKinds, me.Faction, kind)
	if r == 0 {
		return nil
	}
	var out []int
	for i := range g.battleUnits {
		u := g.battleUnits[i]
		if u.Faction != me.Faction || u.Out {
			continue
		}
		if u.Kind == shipKindRepair || u.Kind == shipKindScout {
			continue
		}
		if !repair && u.Kind == shipKindSupply {
			continue
		}
		if !rules.InSupplyRange(me.X, me.Y, u.X, u.Y, r) {
			continue
		}
		out = append(out, i)
	}
	return out
}

// supplyPickClick 處理選對象時的左鍵；回有沒有吃掉這一下。
//
// 按下去那一關原版**再驗一次**同陣營與距離，而且用的是另一個半徑
// （`rules.SupplyRange`，比收候選的小 12／65536）。照抄。
func (g *Game) supplyPickClick(x, y int) bool {
	if !g.battleSupply {
		return false
	}
	unit, repair := g.battleSupplyUnit, g.battleSupplyRepair
	planet := g.battleSupplyPlanet
	cands := g.SupplyCandidates()
	g.cancelSupplyPick()
	hit := g.battleUnitAt(x, y)
	if hit < 0 || unit < 0 || unit >= len(g.battleUnits) {
		return true
	}
	found := false
	for _, c := range cands {
		if c == hit {
			found = true
			break
		}
	}
	if !found {
		return true
	}
	if planet >= 0 {
		p := &g.battlePlanets[planet]
		if repair {
			rules.PlanetReplenish(&p.Stock, &g.battleUnits[hit])
		} else {
			rules.PlanetSupply(g.res.ShipKinds, &p.Stock, &g.battleUnits[hit])
		}
		return true
	}
	from, to := &g.battleUnits[unit], &g.battleUnits[hit]
	kind := rules.SupplyKind
	if repair {
		kind = rules.RepairKind
	}
	if to.Faction != from.Faction {
		return true
	}
	if !rules.InSupplyRange(from.X, from.Y, to.X, to.Y,
		rules.SupplyRange(g.res.ShipKinds, from.Faction, kind)) {
		return true
	}
	if repair {
		rules.Repair(from, to)
	} else {
		rules.Supply(g.res.ShipKinds, from, to)
	}
	return true
}
