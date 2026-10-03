package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// newBattleMovement保留一般組及索敵隊順序；未放出的隊只在本階段持有，不進可見名單。
func (g *Game) newBattleMovement() *rules.BattleMovement {
	g.battleMovementGroups = nil
	for i := range g.battleFleets {
		bf := &g.battleFleets[i]
		group := rules.MovementFleet{Fleet: bf}
		if cpu, passive := g.passiveComputer(); passive && bf.Faction == cpu {
			group.Frozen = true
		}
		for squad := 0; squad < 10; squad++ {
			for j := range g.battleUnits {
				u := &g.battleUnits[j]
				if u.Fleet == bf.Index && u.Squadron == squad {
					group.Units = append(group.Units, rules.MovementOccupant{Pool: g.battleUnitPool(*u), Unit: u})
				}
			}
		}
		scout := false
		for j := range g.battleUnits {
			u := &g.battleUnits[j]
			if u.Fleet == bf.Index && u.Squadron < 0 {
				group.Units = append(group.Units, rules.MovementOccupant{Pool: g.battleUnitPool(*u), Unit: u})
				scout = true
			}
		}
		if !scout && g.slot != nil {
			if f := g.fleetByIndex(bf.Index); f != nil {
				if u, ok := g.slot.ScoutUnit(g.res.Weights, *f, bf.X, bf.Y); ok {
					group.Units = append(group.Units, rules.MovementOccupant{Pool: f.Scout, Unit: &u})
				}
			}
		}
		g.battleMovementGroups = append(g.battleMovementGroups, group)
	}
	return rules.NewBattleMovement(g.battleMovementGroups, g.res.Formation, g.res.Terrain, g.battleCells(g.battleRegion))
}

func (g *Game) beginBattleMovement() {
	// 181B:011E先印文字，0126才初始化額度（083）。
	g.battlePhase = screen.BattlePhaseMovement
	g.battleMoving = true
	g.battleMovement = g.newBattleMovement()
	g.continueBattleMovement()
}

func (g *Game) continueBattleMovement() {
	if g.battleMovement.Next() {
		return
	}
	g.saveBattleMovement()
	g.battleMovement = nil
	g.battleMoving = false
	g.advanceBattleTurnAfterMovement()
}

func (g *Game) saveBattleMovement() {
	for _, group := range g.battleMovementGroups {
		bf := group.Fleet
		for j := range g.fleets {
			f := &g.fleets[j]
			if f.Index != bf.Index {
				continue
			}
			f.X, f.Y, f.Action = bf.X, bf.Y, bf.Action&15
			for _, entry := range group.Units {
				u := entry.Unit
				if u.Squadron >= 0 && u.Squadron < len(f.Composition) {
					gp := &f.Composition[u.Squadron]
					gp.X, gp.Y, gp.Slot3 = u.X, u.Y, u.Slot
				}
			}
			break
		}
		if g.slot != nil {
			g.slot.ApplyBattleMovementFleet(*bf)
			for _, entry := range group.Units {
				g.slot.ApplyBattleUnitMovement(entry.Pool, *entry.Unit)
			}
		}
	}
	g.battleMovementGroups = nil
}
