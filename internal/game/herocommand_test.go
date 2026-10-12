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
	if g.battleFleets[0].DestX != 800 || g.battleFleets[0].PerTurn != 3*8*rules.FixOne || g.battleUnits[1].Action != 3 {
		t.Fatal("未透過正常整隊引擎下令")
	}
	if g.battleUnits[2].Action != 0 || g.battleUnits[2].TargetX != 1200 {
		t.Fatal("改寫獨立索敵命令")
	}
}
