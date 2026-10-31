package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/screen"
)

func TestScoutTargetRestoresCameraAndUsesSelectionView(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	flag := findUnit(g, g.faction, true)
	if flag < 0 {
		t.Fatal("沒有旗艦")
	}
	bf := g.battleFleet(g.battleUnits[flag].Fleet)
	bf.Flag0, bf.ScoutOut = false, false
	original := g.battleView
	g.beginScout(flag)
	if !g.battleMoveRestore || !g.battleMove {
		t.Fatal("索敵未切P視窗")
	}
	u := g.battleMoveUnit
	if g.battleView != g.battleViewAt(g.battleUnits[u].X, g.battleUnits[u].Y) {
		t.Fatal("未聚焦P")
	}
	// 小地圖只移選點鏡頭，不提前確認目的地。
	m := g.res.Battle.Minimap
	g.Click(m.X+m.W/2, m.Y+m.H/2)
	if !g.battleMove || !g.battleMoveRestore {
		t.Fatal("點小地圖誤取消索敵")
	}
	var px, py, wx, wy int
	found := false
	s := g.res.Battle.Field
	for y := s.Y + 10; y < s.Y+s.Rows*s.TileH && !found; y += 20 {
		for x := s.X + 10; x < s.X+s.Cols*s.TileW; x += 16 {
			if a, b, ok := g.battleWorldAt(x, y); ok {
				px, py, wx, wy, found = x, y, a, b, true
				break
			}
		}
	}
	if !found {
		t.Fatal("找不到合法目的地")
	}
	g.Click(px, py)
	if g.battleMove || g.battleMoveRestore || g.battleView != original {
		t.Fatal("確認未還原原視窗")
	}
	if got := g.battleUnits[u]; got.TargetX != wx || got.TargetY != wy || got.Action != squadronMove {
		t.Fatal("使用還原後視窗換算了錯誤座標", got)
	}
	for _, cancel := range []bool{true, false} {
		g.battleView = screen.BattleView{Col: 1, Row: 1}
		want := g.battleView
		g.beginMoveTarget(u)
		if cancel {
			g.RightClick(220, 220)
		} else {
			g.Click(639, 479)
		}
		if g.battleMove || g.battleMoveRestore || g.battleView != want {
			t.Fatal("取消／無效點未還原", cancel)
		}
	}
}
