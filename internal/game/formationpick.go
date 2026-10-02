package game

import "github.com/wicanr2/logh3/internal/rules"

// 「編隊」的選格子畫面（原版 `2BE1:0328`，`docs/internals/46`）。
//
// 命令選單選了「編隊」之後，畫面上把艦隊 41 個隊形槽裡**還沒有人**的
// 那些各畫一個 16×16 的標記，玩家點一個就換過去；右鍵或點空白取消。

// formationSlotBox 是標記的邊長，也是命中框（原版寫死 16×16）。
const formationSlotBox = 16

// FormationSlot 是一個可以選的格子。X／Y 是世界座標的左上角。
type FormationSlot struct {
	Index int
	X, Y  int
}

// FormationPickPending 回「正在等玩家點隊形格」。
func (g *Game) FormationPickPending() bool { return g.battleForm }

// beginFormationPick 進入選格子的狀態。艦隊資料不齊就不進去。
func (g *Game) beginFormationPick(unit int) {
	if g.res.Formation == nil || unit < 0 || unit >= len(g.battleUnits) {
		return
	}
	if len(g.formationSlots(unit)) == 0 {
		return
	}
	g.battleForm, g.battleFormUnit = true, unit
}

// cancelFormationPick 取消選格子。
func (g *Game) cancelFormationPick() { g.battleForm = false }

// FormationSlots 回現在該畫出來的格子（世界座標）。
//
// 原版是先把 41 格全標成可選，再照艦隊分組表把有人的標掉，最後**把
// 發令這個分隊自己那一格放回可選**——所以「留在原地」也選得到。
func (g *Game) FormationSlots() []FormationSlot {
	if !g.battleForm {
		return nil
	}
	return g.formationSlots(g.battleFormUnit)
}

func (g *Game) formationSlots(unit int) []FormationSlot {
	s := g.res.Formation
	if s == nil || unit < 0 || unit >= len(g.battleUnits) {
		return nil
	}
	u := g.battleUnits[unit]
	fx, fy, a, ok := g.battleFleetPose(u.Fleet)
	if !ok {
		return nil
	}
	taken := make(map[int]bool, s.Count)
	for i := range g.battleUnits {
		o := &g.battleUnits[i]
		if o.Fleet == u.Fleet && i != unit {
			taken[o.Slot] = true
		}
	}
	out := make([]FormationSlot, 0, s.Count)
	for i := 0; i < s.Count; i++ {
		if taken[i] {
			continue
		}
		x, y := s.Station(fx, fy, a, i)
		out = append(out, FormationSlot{Index: i, X: x, Y: y})
	}
	return out
}

// battleFleetPose讀目前typed艦隊的+5C/+5E/+4A；缺艦隊時保留旗艦診斷fallback。
func (g *Game) battleFleetPose(fleet int) (int, int, int, bool) {
	if f := g.battleFleet(fleet); f != nil {
		return f.X, f.Y, f.Facing, true
	}
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		if u.Fleet == fleet && u.Squadron == 0 {
			return u.X, u.Y, rules.BattleFacing(u.X, u.Y), true
		}
	}
	return 0, 0, 0, false
}

// formationPickClick 處理選格子時的左鍵；回有沒有吃掉這一下。
//
// 命中判定照原版：畫面座標換成世界座標，**從槽 0 掃到槽 40、先中先贏**，
// 命中框就是那個 16×16 的標記。沒中當取消。
func (g *Game) formationPickClick(x, y int) bool {
	if !g.battleForm {
		return false
	}
	unit := g.battleFormUnit
	slots := g.FormationSlots()
	g.cancelFormationPick()
	if unit < 0 || unit >= len(g.battleUnits) {
		return true
	}
	wx, wy, ok := g.battleFieldPoint(x, y)
	if !ok {
		return true
	}
	for _, s := range slots {
		if wx < s.X || wx >= s.X+formationSlotBox {
			continue
		}
		if wy < s.Y || wy >= s.Y+formationSlotBox {
			continue
		}
		g.applyFormationSlot(unit, s.Index)
		return true
	}
	return true
}

// applyFormationSlot 把分隊換到第 slot 格：寫槽、把行動清成「沒有」，
// 再依新槽算出目的地（原版選完就叫 `2BE1:0059`，不必再下「移動」）。
func (g *Game) applyFormationSlot(unit, slot int) {
	s := g.res.Formation
	u := &g.battleUnits[unit]
	u.Slot = slot
	u.Action = squadronNone
	if fx, fy, a, ok := g.battleFleetPose(u.Fleet); ok && s != nil {
		u.TargetX, u.TargetY = s.Station(fx, fy, a, slot)
	}
}

// battleFieldPoint 把畫面座標換成戰場世界座標，只做邊界不做地形判斷
// ——選格子是在既有的格子上點，通不通行由格子本身決定。
func (g *Game) battleFieldPoint(x, y int) (int, int, bool) {
	s := g.res.Battle
	if s == nil {
		return 0, 0, false
	}
	fw := s.Field.Cols * s.Field.TileW
	fh := s.Field.Rows * s.Field.TileH
	if x < s.Field.X || y < s.Field.Y || x >= s.Field.X+fw || y >= s.Field.Y+fh {
		return 0, 0, false
	}
	return x - s.Field.X + g.battleView.Col*s.Field.TileW,
		y - s.Field.Y + g.battleView.Row*s.Field.TileH, true
}

// formationMarkSlots 回現在要畫標記的格子（世界座標左上角）。
func (g *Game) formationMarkSlots() [][2]int {
	slots := g.FormationSlots()
	if len(slots) == 0 {
		return nil
	}
	out := make([][2]int, 0, len(slots))
	for _, s := range slots {
		out = append(out, [2]int{s.X, s.Y})
	}
	return out
}
