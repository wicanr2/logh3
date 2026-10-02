package game

import "github.com/wicanr2/logh3/internal/rules"

// 艦隊選單第三項「索敵」（原版 `1CC5:0966`，`docs/re/132`）。
//
// 它做的事很短：把艦隊的 `+0x3A`／`+0x3B` 清成 0（＝ 已經放出去了）、
// 把 `+0x38` 指的那一筆艦艇擺出來、行動設成 3（歸隊）、`+0x19` 設成
// 0x80（要畫航跡），然後**開分隊「移動」那支選點常式**讓玩家指定它
// 要去哪裡。所以「索敵」＝ 放偵察隊 ＋ 替它下一個移動命令。

// beginScout 放出艦隊的索敵分遣隊，並進入選目的地的狀態。
func (g *Game) beginScout(unit int) {
	if unit < 0 || unit >= len(g.battleUnits) {
		return
	}
	fleet := g.battleUnits[unit].Fleet
	bf := g.battleFleet(fleet)
	if bf == nil || bf.ScoutOut {
		return
	}
	f, ok := g.saveFleet(fleet)
	if !ok {
		return
	}
	// 偵察隊從旗艦的位置出發。
	from := flagshipOf(g.battleUnits, fleet)
	if from == nil {
		return
	}
	u, ok := g.slot.ScoutUnit(g.res.Weights, f, from.X, from.Y)
	if !ok {
		return
	}
	u.Action = squadronRejoin // 原版寫 3，等玩家指定目的地才會改
	u.Out = false
	u.Marked = true // 艦艇 `+0x19` bit7：上小地圖
	u.MarkerFlags = 0xC0 | 5
	if u.Faction == 1 {
		u.MarkerFlags = 0xC0 | 9
	}
	u.MarkerX, u.MarkerY = u.X, u.Y
	// 已回收的索敵隊仍保留池身份，再次放出不能增加同池副本。
	for i := range g.battleUnits {
		if g.battleUnits[i].Fleet == fleet && g.battleUnits[i].Squadron < 0 {
			g.battleUnits[i] = u
			bf.ScoutOut, bf.ScoutMode, bf.ScoutState = true, 0, 0
			g.beginMoveTarget(i)
			return
		}
	}
	g.battleUnits = append(g.battleUnits, u)
	bf.ScoutOut, bf.ScoutMode, bf.ScoutState = true, 0, 0
	g.beginMoveTarget(len(g.battleUnits) - 1)
}

// saveFleet 回存檔裡的第 index 支艦隊。
func (g *Game) saveFleet(index int) (rules.Fleet, bool) {
	for _, f := range g.slot.Fleets(g.res.Weights) {
		if f.Index == index {
			return f, true
		}
	}
	return rules.Fleet{}, false
}
