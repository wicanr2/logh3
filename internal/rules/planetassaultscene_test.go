package rules

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestPlanetScenesKeepCaptureOwnerAndDetachedData(t *testing.T) {
	kinds, err := LoadShipKindTable(filepath.Join("../..", "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	planets := []BattlePlanet{{Owner: 1, Defence: 10, DefenceMax: 10, NameBytes: []byte{1, 2}}}
	fleets := []BattleFleet{{Index: 7, Faction: 0, Action: 1, Integrity: 100, Ability3: 100, Defence: 100, AdmiralLossFactor: 100}}
	units := []BattleUnit{{Fleet: 7, Faction: 0, Kind: 0, Squadron: 0, Ships: 1000, FullShips: 1000}}
	r := NewRand(0x12345678)
	hits, scenes := PlanetAssaultWithScenes(planets, fleets, units, kinds, r)
	if len(hits) != 1 || !hits[0].Captured || len(scenes) != 2 || scenes[0].Kind != PlanetLandingScene || scenes[1].Kind != PlanetCaptureScene {
		t.Fatalf("capture stages: hits=%+v scenes=%+v", hits, scenes)
	}
	if scenes[0].Before.Planet.Owner != 1 || scenes[0].After.Planet.Owner != 1 || scenes[1].Before.Planet.Owner != 1 || scenes[1].After.Planet.Owner != 0 {
		t.Fatal("capture ownership leaked into earlier scenes")
	}
	if scenes[0].Before.Planet.Defence != 10 || scenes[0].After.Planet.Defence != 0 {
		t.Fatal("landing damage was reconstructed from the final planet")
	}
	units[0].Ships = 1
	planets[0].NameBytes[0] = 9
	if scenes[0].Before.Units[0].Ships != 1000 || scenes[0].Before.Planet.NameBytes[0] != 1 {
		t.Fatal("snapshot aliases live data")
	}
}

func TestPlanetSceneObservationKeepsOriginalCounterOnlyResult(t *testing.T) {
	c := planetEligibilityReceipt(t).Pipeline
	kinds, err := LoadShipKindTable(filepath.Join("../..", "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	r := NewRand(c.Input.Seed)
	hits, scenes := PlanetAssaultWithScenes(c.Input.Planets, c.Input.Fleets, c.Input.Units, kinds, r)
	if len(hits) != 1 || len(scenes) != 1 || scenes[0].Kind != PlanetCounterScene || scenes[0].Destroyed != c.TotalLoss {
		t.Fatalf("counter-only scene does not match the original receipt: %+v", scenes)
	}
	if !reflect.DeepEqual(c.Input.Planets, c.Expected.Planets) || !reflect.DeepEqual(c.Input.Units, c.Expected.Units) || r.Seed() != c.Expected.Seed {
		t.Fatal("observing scenes changed original state or RNG")
	}
	if scenes[0].Before.Seed != c.Input.Seed || scenes[0].After.Seed != c.Expected.Seed {
		t.Fatal("scene omitted RNG boundary")
	}
}
