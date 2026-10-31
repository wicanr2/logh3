package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

func TestLiveHeroCommandBoundary(t *testing.T) {
	g := &Game{aiMode: 2, fleets: []rules.Fleet{{Index: 10, Faction: 0, Admiral: 0}, {Index: 20, Faction: 1, Admiral: 52}}, battleUnits: []rules.BattleUnit{{Fleet: 10, Faction: 0, Squadron: 0, Ships: 100}, {Fleet: 20, Faction: 1, Squadron: 0, Ships: 100}}}
	if g.battleCommanders() != [2]ai.Commander{ai.Reinhard, ai.Yang} {
		t.Fatal("雙方名將未觸發")
	}
	g.battleUnits[0].Out = true
	g.battleUnits[1].Ships = 0
	if g.battleCommanders() != [2]ai.Commander{} {
		t.Fatal("離場或全損仍觸發")
	}
	g.battleUnits[0].Out = false
	g.battleUnits[1].Ships = 100
	g.battleUnits[0].Squadron = 1
	g.fleets[1].Admiral, g.fleets[1].Adjutant = 53, 52
	if g.battleCommanders() != [2]ai.Commander{} {
		t.Fatal("非旗艦或副官觸發")
	}
	g.battleUnits[0].Squadron = 0
	g.fleets[1].Admiral = 52
	g.aiMode = 0
	if g.battleCommanders() != [2]ai.Commander{} {
		t.Fatal("原版模式觸發")
	}
}

func TestHeroFormationOrderKeepsSquadronsAndScout(t *testing.T) {
	g := &Game{battleFleets: []rules.BattleFleet{{Index: 10}}, battleUnits: []rules.BattleUnit{
		{Fleet: 10, Squadron: 0, Ships: 100, X: 100, Y: 100, CruiseSpeed: 4, FightSpeed: 3, TurnSpeed: 8},
		{Fleet: 10, Squadron: 1, Ships: 100, Action: 0, CruiseSpeed: 4, FightSpeed: 3, TurnSpeed: 8},
		{Fleet: 10, Squadron: -1, Kind: 7, Ships: 1, Action: 0, TargetX: 1200, TargetY: 1200},
	}}
	g.applyFormationMove(0, 800, 500)
	if g.battleFleets[0].DestX != 800 || g.battleFleets[0].PerTurn != 3*8*rules.FixOne || g.battleUnits[1].Action != squadronNone {
		t.Fatal("未透過正常整隊引擎下令")
	}
	if g.battleUnits[2].Action != 0 || g.battleUnits[2].TargetX != 1200 {
		t.Fatal("改寫獨立索敵命令")
	}
}

func TestFormationMarchKeepsSlotsInOpenSpaceForBothFactions(t *testing.T) {
	g := newTestGame(t)
	if g.res.Formation == nil {
		t.Fatal("缺少真實槽位表")
	}
	for _, faction := range []int{0, 1} {
		g.battleFleets = []rules.BattleFleet{{Index: 10, Faction: faction, X: 400, Y: 800, Facing: 0, ScoutMode: 255}}
		g.battleUnits = nil
		for squad, slot := range []int{20, 19, 21} {
			x, y := g.res.Formation.Station(400, 800, 0, slot)
			g.battleUnits = append(g.battleUnits, rules.BattleUnit{Fleet: 10, Faction: faction, Squadron: squad, Slot: slot, X: x, Y: y, TargetX: x, TargetY: y, Action: squadronNone, Ships: 100, CruiseSpeed: 8, FightSpeed: 8, TurnSpeed: 8})
		}
		g.applyFormationMove(0, 464, 800)
		group := rules.MovementFleet{Fleet: &g.battleFleets[0]}
		for i := range g.battleUnits {
			group.Units = append(group.Units, rules.MovementOccupant{Pool: i, Unit: &g.battleUnits[i]})
		}
		m := rules.NewBattleMovement([]rules.MovementFleet{group}, g.res.Formation, nil, nil)
		for n := 0; m.Next(); n++ {
			if n > 100 {
				t.Fatal("空地行軍未結束")
			}
			f := g.battleFleets[0]
			for _, u := range g.battleUnits {
				x, y := g.res.Formation.Station(f.X, f.Y, f.Facing, u.Slot)
				if u.X != x || u.Y != y {
					t.Fatalf("陣營%d行軍時收攏：slot%d (%d,%d)應為(%d,%d)", faction, u.Slot, u.X, u.Y, x, y)
				}
			}
		}
		if g.battleFleets[0].X != 464 {
			t.Fatal("沒有完成64像素行軍")
		}
	}
}
