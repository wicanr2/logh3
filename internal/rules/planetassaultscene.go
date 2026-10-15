package rules

// PlanetAssaultSceneKind follows the three visible stages in 20B4:0503 (spec118).
type PlanetAssaultSceneKind uint8

const (
	PlanetLandingScene PlanetAssaultSceneKind = iota
	PlanetCounterScene
	PlanetCaptureScene
)

// PlanetAssaultSnapshot is detached at the rule boundary, before ownership or
// another attacker changes the data. It is transient and never enters a save.
type PlanetAssaultSnapshot struct {
	Planet    BattlePlanet
	Fleet     BattleFleet
	Units     []BattleUnit
	Attack    int
	Integrity int
	Seed      uint32
}

type PlanetAssaultScene struct {
	Kind          PlanetAssaultSceneKind
	PlanetIndex   int
	Before, After PlanetAssaultSnapshot
	Destroyed     int
	Captured      bool
}

// PlanetAssaultWithScenes runs the same rules and RNG sequence as PlanetAssault.
// Observation only copies data; it cannot modify the calculation or input.
func PlanetAssaultWithScenes(planets []BattlePlanet, fleets []BattleFleet,
	units []BattleUnit, kinds *ShipKindTable, r *Rand) ([]PlanetAssaultHit, []PlanetAssaultScene) {
	return PlanetAssaultWithAfterCounter(planets, fleets, units, kinds, r, nil)
}

// A remake transition may run after a counterattack and before the next stage.
func PlanetAssaultWithAfterCounter(planets []BattlePlanet, fleets []BattleFleet,
	units []BattleUnit, kinds *ShipKindTable, r *Rand, afterCounter func(PlanetAssaultScene)) ([]PlanetAssaultHit, []PlanetAssaultScene) {
	var scenes []PlanetAssaultScene
	hits := planetAssaultObserved(planets, fleets, units, kinds, r, func(s PlanetAssaultScene) {
		scenes = append(scenes, s)
		if afterCounter != nil && s.Kind == PlanetCounterScene {
			afterCounter(s)
		}
	})
	return hits, scenes
}

func assaultSnapshot(p *BattlePlanet, f *BattleFleet, units []BattleUnit, r *Rand) PlanetAssaultSnapshot {
	planet := *p
	planet.NameBytes = append([]byte(nil), p.NameBytes...)
	return PlanetAssaultSnapshot{
		Planet: planet, Fleet: *f, Units: append([]BattleUnit(nil), units...),
		Attack: FleetAttack(f, units), Integrity: f.Integrity, Seed: r.Seed(),
	}
}
