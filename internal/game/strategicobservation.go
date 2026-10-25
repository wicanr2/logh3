package game

import "github.com/wicanr2/logh3/internal/rules"

// StrategicObservation is read-only telemetry for normal player recordings.
// It neither changes state nor supplies input, seeds or test fixtures.
type StrategicObservation struct {
	Turn         int                         `json:"turn"`
	Faction      int                         `json:"faction"`
	Region       int                         `json:"region"`
	Stock        []int                       `json:"stock,omitempty"`
	CapitalStock []int                       `json:"capital_stock,omitempty"`
	Idle         []int                       `json:"idle,omitempty"`
	Detail       int                         `json:"detail"`
	Message      bool                        `json:"message"`
	Fleets       []StrategicFleetObservation `json:"fleets,omitempty"`
	Editor       []rules.FleetGroup          `json:"editor,omitempty"`
}

type StrategicFleetObservation struct {
	Index, Number, Admiral, Adjutant, Region, Target, Move int
	Groups                                                 []rules.FleetGroup
}

func (g *Game) StrategicObservation() *StrategicObservation {
	if g.state == StateIntro || g.state == StateTitle || g.state == StateMainMenu || g.state == StateCopyProtection {
		return nil
	}
	r := g.listRegion()
	v := &StrategicObservation{Turn: g.turn, Faction: g.faction, Region: r, Detail: g.detailAdmiral, Message: g.msgOn}
	if rec := g.regionRecord(r); rec != nil {
		v.Stock = rules.RegionShips(rec, g.faction)
	}
	if rec := g.regionRecord(rules.CapitalRegion(g.faction)); rec != nil {
		v.CapitalStock = rules.RegionShips(rec, g.faction)
	}
	v.Idle = append([]int(nil), g.idleAdmiralsInRegion()...)
	if g.tabs.detail >= 0 {
		v.Detail = g.tabs.detail
	}
	for _, f := range g.fleets {
		if f.Faction == g.faction {
			v.Fleets = append(v.Fleets, StrategicFleetObservation{f.Index, f.Number, f.Admiral, f.Adjutant, f.Region, f.Target, f.Move, append([]rules.FleetGroup(nil), f.Composition...)})
		}
	}
	if g.editor != nil {
		v.Editor = g.editor.Groups()
	}
	return v
}
