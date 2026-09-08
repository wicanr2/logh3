package rules

import "testing"

// TestBattleMoveMatchesOracle 對原版量到的一回合行程。
//
// 收據 `s3move`（`-watch 613A0-613A3`）逐步印出同盟第 16 艦隊旗艦的座標：
// 一回合走 **29 步、每步 (−4,−7)**，行程餘額從 240.0 扣到 7.09 就停。
// 起點 (1536,1840)、目標 (1168,1160)（集結點 3，`docs/re/100`），
// 第 1 回合結束在 (1420,1637) 附近、第 2 回合到 (1306,1434)。
//
// 原版沿的是艦隊**現在的朝向**、另有一支轉向邏輯還沒逆向；這一段路上
// 朝向已經指著目標，所以兩個回合的落點都逐像素相同。
func TestBattleMoveMatchesOracle(t *testing.T) {
	u := BattleUnit{X: 1536, Y: 1840, TargetX: 1168, TargetY: 1160, Speed: 240}
	for i, want := range [][2]int{{1420, 1637}, {1306, 1434}} {
		x0, y0 := u.X, u.Y
		if !AdvanceBattle(&u, nil, nil) {
			t.Fatalf("第 %d 回合沒動", i+1)
		}
		d := isqrt((u.X-x0)*(u.X-x0) + (u.Y-y0)*(u.Y-y0))
		t.Logf("第 %d 回合：(%d,%d) → (%d,%d)，走了 %d", i+1, x0, y0, u.X, u.Y, d)
		if u.X != want[0] || u.Y != want[1] {
			t.Errorf("第 %d 回合走到 (%d,%d)，原版是 (%d,%d)",
				i+1, u.X, u.Y, want[0], want[1])
		}
	}
}

// TestBattleMoveStopsAtTerrain 釘住「撞到地形就停」。
//
// 原版每走一步都先問 `2583:040C`（→ `1588:04EA`），不能站就不走
// （`docs/re/102`）。
func TestBattleMoveStopsAtTerrain(t *testing.T) {
	// 一整張都是圖塊 44（遮罩 0F ＝ 不通），只有起點那一格是 0。
	cells := make([]uint8, 128*102)
	for i := range cells {
		cells[i] = 44
	}
	cells[50*128+50] = 0
	terrain := &BattleTerrain{
		BlockMask:    make([]uint8, 64),
		QuadrantBits: []uint8{0, 2, 4, 8},
		HalfCellW:    8, HalfCellH: 10, Footprint: 16, Step: 8,
	}
	for i := 8; i < 64; i++ {
		terrain.BlockMask[i] = 0x0F
	}
	u := BattleUnit{X: 50 * 16, Y: 50 * 20, TargetX: 100 * 16, TargetY: 50 * 20, Speed: 240}
	AdvanceBattle(&u, terrain, cells)
	if u.X > 50*16+8 {
		t.Errorf("撞到障礙還走到 x=%d", u.X)
	}
}
