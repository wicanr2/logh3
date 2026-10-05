package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type planetEligibilityState struct {
	Planets []BattlePlanet `json:"planets"`
	Fleets  []BattleFleet  `json:"fleets"`
	Units   []BattleUnit   `json:"units"`
	Targets []int          `json:"targets"`
	Bits    []int          `json:"bits"`
	Seed    uint32         `json:"seed"`
}

func planetEligibilityReceipt(t *testing.T) struct {
	Cases []struct {
		Name            string
		Input, Expected planetEligibilityState
	}
	Pipeline struct {
		Input, Expected planetEligibilityState
		TotalLoss       int `json:"total_loss"`
	}
} {
	t.Helper()
	var receipt struct {
		Cases []struct {
			Name            string
			Input, Expected planetEligibilityState
		}
		Pipeline struct {
			Input, Expected planetEligibilityState
			TotalLoss       int `json:"total_loss"`
		}
	}
	b, err := os.ReadFile(filepath.Join("../..", "docs/internals/data/planet-eligibility.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	return receipt
}

// 預期值由 dosgolem 重新執行000C／0091後的RAM生成，非重寫判定算式。
func TestPlanetAssaultEligibilityOriginalMatrix(t *testing.T) {
	receipt := planetEligibilityReceipt(t)
	if len(receipt.Cases) != 40 {
		t.Fatal("原版矩陣必須完整40案")
	}
	for _, c := range receipt.Cases {
		t.Run(c.Name, func(t *testing.T) {
			target, bits := PlanetAssaultTargets(c.Input.Planets, c.Input.Fleets, c.Input.Units)
			if !reflect.DeepEqual(target, c.Expected.Targets) || !reflect.DeepEqual(bits, c.Expected.Bits) {
				t.Fatalf("got target=%v bits=%v; original target=%v bits=%v", target, bits, c.Expected.Targets, c.Expected.Bits)
			}
		})
	}
}

// 完整父state重算Targets後續行至181B:03F0。比較七艦隊所有分隊、行星與RND。
func TestPlanetAssaultCounterOnlyMatchesOriginal(t *testing.T) {
	c := planetEligibilityReceipt(t).Pipeline
	kinds, err := LoadShipKindTable(filepath.Join("../..", "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	rng := NewRand(c.Input.Seed)
	hits := PlanetAssault(c.Input.Planets, c.Input.Fleets, c.Input.Units, kinds, rng)
	if len(hits) != 1 || hits[0].Fleet != 31 || hits[0].Planet != 0 || hits[0].Damage != 0 || hits[0].Captured || hits[0].Score != 0 || hits[0].Counter != c.TotalLoss {
		t.Fatalf("counter-only結果不符原版：%+v; loss=%d", hits, c.TotalLoss)
	}
	if !reflect.DeepEqual(c.Input.Planets, c.Expected.Planets) {
		t.Fatal("counter-only改變行星或庫存")
	}
	if !reflect.DeepEqual(c.Input.Units, c.Expected.Units) {
		t.Errorf("分隊艘數、修復上限或庫存不符原版")
	}
	if rng.Seed() != c.Expected.Seed {
		t.Errorf("seed=%08X original=%08X", rng.Seed(), c.Expected.Seed)
	}
	for i := range c.Input.Fleets {
		f := &c.Input.Fleets[i]
		if got := FleetIntegrity(f, c.Input.Units); got != c.Expected.Fleets[i].Integrity {
			t.Errorf("fleet%d integrity=%d original=%d", f.Index, got, c.Expected.Fleets[i].Integrity)
		}
	}
}
