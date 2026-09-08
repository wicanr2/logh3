package rules

import "testing"

func testUnits() []BattleUnit {
	return []BattleUnit{
		// 帝國艦隊 0：旗艦在 (1000,1000)
		{Fleet: 0, Squadron: 0, Faction: 0, X: 1000, Y: 1000, TargetX: 1000, TargetY: 1000,
			Kind: 1, Ships: 3000, FullShips: 3000, Attack: 4000},
		{Fleet: 0, Squadron: 1, Faction: 0, X: 1020, Y: 1000, TargetX: 1020, TargetY: 1000,
			Kind: 2, Ships: 4000, FullShips: 4000, Attack: 4000},
		// 同盟艦隊 1：旗艦在 (1100,1000)，距離 100 < 射程 × 2
		{Fleet: 1, Squadron: 0, Faction: 1, X: 1100, Y: 1000, TargetX: 1100, TargetY: 1000,
			Kind: 1, Ships: 3000, FullShips: 3000, Attack: 4000},
		{Fleet: 1, Squadron: 1, Faction: 1, X: 1120, Y: 1000, TargetX: 1120, TargetY: 1000,
			Kind: 2, Ships: 4000, FullShips: 4000, Attack: 4000},
	}
}

func testFleets() []BattleFleet {
	return []BattleFleet{
		{Index: 0, Faction: 0, Bonus: 100, Defence: 100, Facing: 0},
		{Index: 1, Faction: 1, Bonus: 100, Defence: 100, Facing: 0},
	}
}

func TestResolveBattleTurnDeducts(t *testing.T) {
	units, fleets := testUnits(), testFleets()
	before := units[2].Ships + units[3].Ships
	got := ResolveBattleTurn(fleets, units, NewRand(1))
	if len(got) != 2 {
		t.Fatalf("兩支艦隊都在射程內，該有兩次交火，實得 %d", len(got))
	}
	after := units[2].Ships + units[3].Ships
	if after >= before {
		t.Errorf("同盟該被扣，扣前 %d 扣後 %d", before, after)
	}
	for i, u := range units {
		if u.Ships < 0 {
			t.Errorf("分隊 %d 的艘數變成負的：%d", i, u.Ships)
		}
	}
}

// 超出射程就不開火。
func TestResolveBattleTurnOutOfRange(t *testing.T) {
	units, fleets := testUnits(), testFleets()
	for i := range units {
		if units[i].Faction == 1 {
			units[i].X += 2000
			units[i].TargetX += 2000
		}
	}
	if got := ResolveBattleTurn(fleets, units, NewRand(1)); len(got) != 0 {
		t.Fatalf("離那麼遠不該開火，卻有 %d 次", len(got))
	}
}

// 推很多回合不會卡住，也不會出現負的艘數。
func TestResolveBattleTurnRunsDown(t *testing.T) {
	units, fleets := testUnits(), testFleets()
	rnd := NewRand(12345)
	for turn := 0; turn < 500; turn++ {
		ResolveBattleTurn(fleets, units, rnd)
	}
	for i, u := range units {
		if u.Ships < 0 {
			t.Fatalf("分隊 %d 的艘數變成負的：%d", i, u.Ships)
		}
	}
}

// 防禦係數的兩項就是能力的第 3、4 項。
func TestNewBattleFleetDefence(t *testing.T) {
	var ab [FleetAbilityCount]byte
	ab[defenceAbilityA], ab[defenceAbilityB] = 120, 90
	rec := make([]byte, FleetRecord)
	rec[1], rec[0x12], rec[0x14] = 1, 1, 137
	f := NewBattleFleet(31, rec, ab)
	if f.Defence != 108 {
		t.Errorf("防禦係數 ＝ 120 × 90 ÷ 100 ＝ 108，實得 %d", f.Defence)
	}
	if f.Faction != 1 || f.Bonus != 137 || !f.Halved {
		t.Errorf("欄位讀錯：%+v", f)
	}
}
