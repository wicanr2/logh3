package game

// PlanetPlaybackObservation describes the currently displayed detached stage.
// Reading it never advances the scene, changes orders or consumes randomness.
type PlanetPlaybackObservation struct {
	Planet, Fleet, Kind, Label      int
	Captured                        bool
	OwnerBefore, OwnerAfter         int
	DefenceBefore, DefenceAfter     int
	ShipsBefore, ShipsAfter         int
	AttackBefore, AttackAfter       int
	IntegrityBefore, IntegrityAfter int
}

func (g *Game) PlanetPlaybackObservation() *PlanetPlaybackObservation {
	if g.state != StateEngagement || len(g.engQueue) == 0 || g.engQueue[0].stage == nil {
		return nil
	}
	e := g.engQueue[0]
	s := e.stage.scene
	b, _ := snapshotShips(s.Before)
	a, _ := snapshotShips(s.After)
	return &PlanetPlaybackObservation{e.planet, e.fleet, int(s.Kind), int(e.label), s.Captured,
		s.Before.Planet.Owner, s.After.Planet.Owner, s.Before.Planet.Defence, s.After.Planet.Defence,
		b, a, s.Before.Attack, s.After.Attack, s.Before.Integrity, s.After.Integrity}
}

type TacticalUnitObservation struct {
	Index, Fleet, Faction, Squadron, Kind, X, Y, Ships int
	Out                                                bool
}

type TacticalCombatChoiceObservation struct {
	Target, Mask int
	Point        [2]int
}
