package rules

import "testing"

func TestCommandMovementBudgetFreshAndIndependent(t *testing.T) {
	f := BattleFleet{PerTurn: 32 * FixOne, ScoutMode: 0xFF}
	u := BattleUnit{Squadron: 1, CruiseSpeed: 4, Budget: 100}
	scout := BattleUnit{Squadron: -1, CruiseSpeed: 4, Budget: 456}
	groups := []MovementFleet{{Fleet: &f, Units: []MovementOccupant{{Unit: &u}, {Unit: &scout}}, CommandBoost: true}}
	NewBattleMovement(groups, nil, nil, nil)
	if f.Budget != 48*FixOne || u.Budget != 48*FixOne+100 || scout.Budget != 456 || f.PerTurn != 32*FixOne || u.CruiseSpeed != 4 {
		t.Fatalf("加成改基本速度、餘額或未啟用索敵隊：%+v %+v %+v", f, u, scout)
	}
	NewBattleMovement(groups, nil, nil, nil)
	if f.Budget != 48*FixOne || u.Budget != 48*FixOne {
		t.Fatal("重建累乘")
	}
	groups[0].CommandBoost = false
	NewBattleMovement(groups, nil, nil, nil)
	if f.Budget != 32*FixOne || u.Budget != 32*FixOne {
		t.Fatal("取消未恢復")
	}
	groups[0].Frozen, groups[0].CommandBoost = true, true
	f.Budget, u.Budget = 111, 222
	NewBattleMovement(groups, nil, nil, nil)
	if f.Budget != 111 || u.Budget != 222 {
		t.Fatal("凍結也補充額度")
	}
}

func TestCommandMovementActuallyMovesOneAndHalf(t *testing.T) {
	for _, boost := range []bool{false, true} {
		f := BattleFleet{X: 200, Y: 200, DestX: 1800, DestY: 200, PerTurn: 32 * FixOne, ShortMove: true}
		m := NewBattleMovement([]MovementFleet{{Fleet: &f, CommandBoost: boost}}, nil, nil, nil)
		steps := 0
		for m.Next() {
			steps++
			if steps > 20 {
				t.Fatal("移動未終止")
			}
		}
		want := 232
		if boost {
			want = 248
		}
		if f.X != want || f.Y != 200 {
			t.Fatalf("boost%v 落點%v", boost, f)
		}
	}
}

func TestCommandMovementStoppedAndActiveScout(t *testing.T) {
	f := BattleFleet{X: 200, Y: 200, DestX: 1800, DestY: 200, PerTurn: 32 * FixOne, Action: 2, ScoutMode: 0}
	u := BattleUnit{Squadron: -1, Kind: 7, CruiseSpeed: 4, Action: 0, X: 300, Y: 300, TargetX: 500, TargetY: 300}
	m := NewBattleMovement([]MovementFleet{{Fleet: &f, CommandBoost: true, Units: []MovementOccupant{{Unit: &u}}}}, nil, nil, nil)
	if u.Budget != 48*FixOne {
		t.Fatal("活動索敵隊未加成")
	}
	for n := 0; m.Next(); n++ {
		if n > 20 {
			t.Fatal("移動未終止")
		}
	}
	if f.X != 200 || f.Y != 200 || u.X != 348 {
		t.Fatal("加成強迫停止艦隊移動或索敵行程錯誤", f, u)
	}
}
