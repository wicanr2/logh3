package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// battleUnitPool 保留來源的池身份；比較時另讀該池目前的擁有者。
func (g *Game) battleUnitPool(u rules.BattleUnit) int {
	f := g.fleetByIndex(u.Fleet)
	if f != nil {
		if u.Squadron < 0 {
			if f.Scout >= 0 && f.Scout != 0xFFFF {
				return f.Scout
			}
			return -1
		}
		if u.Squadron < len(f.Composition) && f.Composition[u.Squadron].Pool >= 0 {
			return f.Composition[u.Squadron].Pool
		}
	}
	if g.slot != nil && u.Squadron >= 0 {
		ships := g.slot.FleetGroupShips(u.Fleet)
		if u.Squadron < len(ships) {
			return ships[u.Squadron]
		}
	}
	return -1
}

func (g *Game) rememberPreparationSource(unit int, repair bool) {
	if unit < 0 || unit >= len(g.battleUnits) {
		return
	}
	pool := g.battleUnitPool(g.battleUnits[unit])
	s := rules.PreparationSource{Pool: pool, Known: pool >= 0}
	if repair {
		g.battleSources.Repair = s
	} else {
		g.battleSources.Supply = s
	}
}

// preparationUnits 重建原版清單：每支一般組後接索敵分遣隊。
// 未放出的索敵隊不進renderer，但仍能是資源清單最後的來源。
func (g *Game) preparationUnits() []rules.PreparationUnit {
	var out []rules.PreparationUnit
	for _, bf := range g.battleFleets {
		for squad := 0; squad < 10; squad++ {
			for i := range g.battleUnits {
				u := &g.battleUnits[i]
				if u.Fleet == bf.Index && u.Squadron == squad {
					out = append(out, rules.PreparationUnit{Pool: g.battleUnitPool(*u), Unit: u})
				}
			}
		}
		scout := false
		for i := range g.battleUnits {
			u := &g.battleUnits[i]
			if u.Fleet == bf.Index && u.Squadron < 0 {
				out = append(out, rules.PreparationUnit{Pool: g.battleUnitPool(*u), Unit: u})
				scout = true
			}
		}
		if !scout && g.slot != nil {
			if f := g.fleetByIndex(bf.Index); f != nil {
				if u, ok := g.slot.ScoutUnit(g.res.Weights, *f, bf.X, bf.Y); ok {
					u.Out = true // 未放出時原版SHIP+1E bit7。
					out = append(out, rules.PreparationUnit{Pool: f.Scout, Unit: &u})
				}
			}
		}
	}
	return out
}

func (g *Game) beginBattlePreparation(after func()) {
	g.battlePreparing = true
	g.battlePreparationAfter = after
	g.battlePreparation = rules.NewBattlePreparation(g.res.ShipKinds, g.preparationUnits(),
		g.battlePlanets, &g.battleSources, func(pool int) int {
			if g.slot != nil {
				if r := g.slot.ShipAt(pool); len(r) > 1 {
					return int(r[1])
				}
			}
			return 0xFF
		})
	g.continueBattlePreparation()
}

func (g *Game) continueBattlePreparation() {
	g.battlePreparationSupply, g.battlePreparationRepair = nil, nil
	for {
		r, ok := g.battlePreparation.Next()
		if !ok {
			g.battlePreparation = nil
			g.battlePreparationMorale(0)
			return
		}
		if r.From == nil || !g.humanSeat(r.To.Faction) || g.res.SupplyPanel == nil {
			continue
		}
		from, to := r.From, r.To
		if from.Faction < 0 || from.Faction >= len(g.res.FactionNames) {
			continue
		}
		name := []byte(nil)
		if to.Kind >= 0 && to.Kind < len(g.res.ShipKindNames) {
			name = g.res.ShipKindNames[to.Kind]
		}
		fleet := from.FleetNumber
		if f := g.fleetByIndex(from.Fleet); f != nil {
			fleet = f.Number
		}
		if r.Repair {
			g.battlePreparationRepair = &screen.RepairPanel{
				FactionName: g.res.FactionNames[from.Faction], Fleet: fleet,
				Repairers: from.Ships, Capacity: int64(from.Missiles),
				Target: g.repairTargetView(*to),
			}
		} else if k, has := g.res.ShipKinds.Kind(to.Faction, to.Kind); has {
			g.battlePreparationSupply = &screen.SupplyPanel{
				FactionName: g.res.FactionNames[from.Faction], Fleet: fleet,
				Transport: from.Ships, Missiles: int64(from.Missiles), Fighters: int64(from.Fighters),
				Target: &screen.SupplyTarget{KindName: name, Ships: to.Ships, NoStock: k.Byte(0x0E) == 0xFF,
					Missiles: int64(to.Missiles), MissileFull: int64(k.Byte(0x10)) * int64(to.Ships),
					Fighters: int64(to.Fighters), FighterFull: int64(k.Byte(0x14)) * int64(to.Ships)},
			}
		}
		if g.battlePreparationSupply != nil || g.battlePreparationRepair != nil {
			return // 下一次Tick自行續接，不需要按鍵，也沒有「結束」鈕。
		}
	}
}

// 全部資源搬完後逐支同步編制、重算士氣並當場報告。
func (g *Game) battlePreparationMorale(start int) {
	for i := start; i < len(g.battleFleets); i++ {
		index := g.battleFleets[i].Index
		for j := range g.fleets {
			f := &g.fleets[j]
			if f.Index != index {
				continue
			}
			for _, u := range g.battleUnits {
				if u.Fleet != index || u.Squadron < 0 || u.Squadron >= len(f.Composition) {
					continue
				}
				gp := &f.Composition[u.Squadron]
				gp.Ships, gp.Full, gp.RepairCap = u.Ships, u.FullShips, u.RepairCap
				gp.Missiles, gp.Fighters = u.Missiles, u.Fighters
			}
			reports := g.fleetStepMorale(j)
			g.battleFleets[i].Integrity = f.Morale
			resume := i + 1
			if g.startBattleReports(reports, func() { g.battlePreparationMorale(resume) }) {
				return
			}
			break
		}
	}
	rules.RecoverBattleDefence(g.battlePlanets)
	g.battlePreparing = false
	after := g.battlePreparationAfter
	g.battlePreparationAfter = nil
	if after != nil {
		after()
	}
}
