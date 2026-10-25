package game

import (
	"bytes"
	"fmt"

	"github.com/wicanr2/logh3/internal/rules"
)

// Called exactly once at a positive-to-zero damage boundary. Keeping unit
// indices stable preserves already queued combat commands and scene targets.
func (g *Game) tryFlagshipTransfer(destroyed int) bool {
	if !g.flagshipTransfer {
		return false
	}
	target := rules.FlagshipTransferTargetWithPolicy(g.battleUnits, destroyed, g.nonBattleshipFlagship)
	if target < 0 {
		return false
	}
	flag, survivor := g.battleUnits[destroyed], g.battleUnits[target]
	f, bf := g.fleetByIndex(flag.Fleet), g.battleFleet(flag.Fleet)
	if f == nil || bf == nil || g.slot == nil || survivor.Squadron >= len(f.Composition) {
		g.saveErr = fmt.Errorf("旗艦轉移缺現行編制")
		return false
	}
	pool := g.battleUnitPool(flag)
	if g.flagshipTransferTried[pool] {
		return false
	}
	// Validate a detached transaction before consuming transfer RNG.
	work := *g.slot
	work.Fleet, work.Ship = bytes.Clone(g.slot.Fleet), bytes.Clone(g.slot.Ship)
	var units []rules.BattleUnit
	for _, unit := range g.battleUnits {
		if unit.Fleet == flag.Fleet && unit.Squadron >= 0 {
			units = append(units, unit)
			work.ApplyBattleUnitMovement(g.battleUnitPool(unit), unit)
		}
	}
	work.ApplyBattleResult(units)
	if err := work.TransferFlagshipWithPolicy(flag.Fleet, survivor.Squadron, g.nonBattleshipFlagship); err != nil {
		g.saveErr = err
		return false
	}
	roll := 0
	chance := g.admiralPreferences.Special(f.Admiral).Transfer
	if chance == 0 {
		return false
	}
	if chance < 100 {
		if g.battleRand == nil {
			g.saveErr = fmt.Errorf("旗艦轉移缺本場亂數狀態")
			return false
		}
		roll = g.battleRand.Next()
	}
	if g.flagshipTransferTried == nil {
		g.flagshipTransferTried = map[int]bool{}
	}
	g.flagshipTransferTried[pool] = true
	if chance < 100 && roll%100 >= chance {
		return false
	}
	g.slot.Fleet, g.slot.Ship = work.Fleet, work.Ship
	f.Composition = work.FleetComposition(flag.Fleet)
	f.X, f.Y = survivor.X, survivor.Y
	bf.X, bf.Y = survivor.X, survivor.Y
	g.battleUnits[destroyed].Squadron, g.battleUnits[destroyed].Slot = survivor.Squadron, survivor.Slot
	g.battleUnits[target].Squadron, g.battleUnits[target].Slot = 0, rules.FlagshipFormationSlot
	g.slot.ApplyBattleMovementFleet(*bf)
	return true
}

func (g *Game) transferAfterCounter(scene rules.PlanetAssaultScene) {
	if scene.Kind != rules.PlanetCounterScene {
		return
	}
	for i, before := range scene.Before.Units {
		if before.Fleet == scene.Before.Fleet.Index && before.Squadron == 0 && before.Ships > 0 && i < len(g.battleUnits) && g.battleUnits[i].Ships == 0 {
			g.tryFlagshipTransfer(i)
			return
		}
	}
}

func (g *Game) transferAfterGun(scene rules.FortressGunScene) {
	i := scene.Unit
	if i >= 0 && scene.Hit.Lost > 0 && scene.Before.Units[i].Squadron == 0 && scene.Before.Units[i].Ships > 0 && g.battleUnits[i].Ships == 0 {
		g.tryFlagshipTransfer(i)
	}
}
