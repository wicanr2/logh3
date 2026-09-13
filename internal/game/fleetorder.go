package game

import "github.com/wicanr2/logh3/internal/rules"

// 艦隊選單的「移動」與「戰鬥」（原版 `1CC5:0813`／`1CC5:08BD` →
// `2583:0003`，`docs/re/120`–`122`）。
//
// 兩段互動：先點一個目的地，再**繞著目的地轉游標**選抵達朝向，左鍵
// 確定、右鍵取消。兩項共用同一支常式，差別只有畫範圍時用哪個速度，
// 以及「戰鬥」成功之後把艦隊行動寫成 1。

// FleetOrderStage 是艦隊命令現在停在哪一段。
type FleetOrderStage int

// 艦隊命令的三種狀態。
const (
	FleetOrderNone   FleetOrderStage = iota
	FleetOrderPoint                  // 在等玩家點目的地
	FleetOrderFacing                 // 在等玩家選抵達朝向
)

// FleetOrderStage 回現在停在哪一段。
func (g *Game) FleetOrderStage() FleetOrderStage { return g.fleetOrder }

// FleetOrderTarget 回已經點好的目的地（世界座標）與現在的朝向。
// 只有在 FleetOrderFacing 這一段才有意義。
func (g *Game) FleetOrderTarget() (x, y, facing int) {
	return g.fleetOrderX, g.fleetOrderY, g.fleetOrderFacing
}

// beginFleetOrder 進入選目的地的狀態。battle ＝ 這一次是「戰鬥」。
func (g *Game) beginFleetOrder(unit int, battle bool) {
	if unit < 0 || unit >= len(g.battleUnits) {
		return
	}
	if g.battleFleet(g.battleUnits[unit].Fleet) == nil {
		return
	}
	g.fleetOrder, g.fleetOrderUnit, g.fleetOrderBattle = FleetOrderPoint, unit, battle
}

// cancelFleetOrder 取消（右鍵）。原版兩段都回 `0xFFFF` 就整個作廢，
// 不會留下半個命令。
func (g *Game) cancelFleetOrder() { g.fleetOrder = FleetOrderNone }

// fleetOrderClick 處理艦隊命令兩段的左鍵；回有沒有吃掉這一下。
func (g *Game) fleetOrderClick(x, y int) bool {
	switch g.fleetOrder {
	case FleetOrderPoint:
		wx, wy, ok := g.battleWorldAt(x, y)
		if !ok {
			// 與分隊的「移動」同一條檢查鏈：原版跳「這是無法行動的
			// 地方。」再整個作廢（`docs/re/121`、`138`）。
			g.showMessage("move_blocked")
			g.cancelFleetOrder()
			return true
		}
		g.fleetOrderX, g.fleetOrderY = wx, wy
		g.fleetOrderFacing = g.fleetDefaultFacing()
		g.fleetOrder = FleetOrderFacing
		return true
	case FleetOrderFacing:
		g.fleetOrderFacing = g.fleetFacingAt(x, y)
		g.commitFleetOrder()
		return true
	}
	return false
}

// fleetOrderMouse 在選朝向那一段跟著游標轉。回有沒有吃掉這一下。
func (g *Game) fleetOrderMouse(x, y int) bool {
	if g.fleetOrder != FleetOrderFacing {
		return false
	}
	g.fleetOrderFacing = g.fleetFacingAt(x, y)
	return true
}

// fleetFacingAt 回游標在 (x,y) 時的抵達朝向：**目的地指向游標**的方位
// （`2583:03A9`）。游標正好停在目的地上就維持預設值。
func (g *Game) fleetFacingAt(x, y int) int {
	wx, wy, ok := g.battleFieldPoint(x, y)
	if !ok || (wx == g.fleetOrderX && wy == g.fleetOrderY) {
		return g.fleetOrderFacing
	}
	return rules.Atan256(wx-g.fleetOrderX, wy-g.fleetOrderY)
}

// fleetDefaultFacing 回下命令那一刻艦隊現在的朝向（`+0x4A`）。
func (g *Game) fleetDefaultFacing() int {
	if f := g.orderedFleet(); f != nil {
		return f.Facing
	}
	return 0
}

func (g *Game) orderedFleet() *rules.BattleFleet {
	if g.fleetOrderUnit < 0 || g.fleetOrderUnit >= len(g.battleUnits) {
		return nil
	}
	return g.battleFleet(g.battleUnits[g.fleetOrderUnit].Fleet)
}

// commitFleetOrder 把命令寫進艦隊記錄：先備份現在的位置，再寫目的地與
// 抵達朝向；「戰鬥」另外把行動寫成 1。
func (g *Game) commitFleetOrder() {
	f := g.orderedFleet()
	battle := g.fleetOrderBattle
	x, y, facing := g.fleetOrderX, g.fleetOrderY, g.fleetOrderFacing
	unit := g.fleetOrderUnit
	g.cancelFleetOrder()
	if f == nil {
		return
	}
	flag := g.battleUnits[unit]
	f.OriginX, f.OriginY = flag.X, flag.Y
	f.X, f.Y = flag.X, flag.Y
	f.DestX, f.DestY = x, y
	f.TargetFacing = facing
	f.Ordered = true
	// 一回合的行程 ＝ 這一次用的速度 × 8（`307D:1468`，`docs/re/102`），
	// 回轉速度是艦隊記錄的 `+0x4D`。兩者都取各分隊的最小值
	// （`docs/re/112`）。
	f.TurnSpeed, f.PerTurn = g.fleetSpeeds(f.Index, battle)
	f.Action = 0
	if battle {
		f.Action = fleetActionBattle
	}
	f.ShortMove = rules.Distance16(f.X, f.Y, f.DestX, f.DestY) <
		int64(f.PerTurn/3+4*rules.FixOne)
	g.battleUnits[unit].Action = squadronMove
}

// fleetSpeeds 回一支艦隊的回轉速度與這一次的每回合行程（16.16）。
//
// 回轉速度 ＝ 各分隊艦艇記錄 `+0x1F` 的最小值；速度看命令：
// 「移動」用巡航 `+0x21`、「戰鬥」用戰鬥 `+0x22`，都取最小值再 × 8
// （`docs/re/112`、`docs/re/102`）。
func (g *Game) fleetSpeeds(fleet int, battle bool) (turn, perTurn int) {
	turn, speed := 0, 0
	for _, u := range g.battleUnits {
		if u.Fleet != fleet {
			continue
		}
		s := u.CruiseSpeed
		if battle {
			s = u.FightSpeed
		}
		if turn == 0 || (u.TurnSpeed > 0 && u.TurnSpeed < turn) {
			turn = u.TurnSpeed
		}
		if speed == 0 || (s > 0 && s < speed) {
			speed = s
		}
	}
	return turn, speed * rules.BattleStep * rules.FixOne
}

// fleetActionBattle 是艦隊記錄 `+0x48` 低四位的「戰鬥」（`docs/internals/33`）。
const fleetActionBattle = 1

// rangeShape 是畫範圍圖形要用的幾個數字。
type rangeShape struct {
	cx, cy, turn, dist, facing int
}

// orderedFleetForRange 回正在選點的那一支要畫的範圍；沒有就回 nil。
//
// 圓心是旗艦圖塊的中心（艦隊位置各 ＋8，`165F:0003`／`165F:004D`），
// 距離是這一次用的速度 × 8：「移動」用巡航、「戰鬥」用戰鬥速度。
func (g *Game) orderedFleetForRange() *rangeShape {
	unit := g.fleetOrderUnit
	if unit < 0 || unit >= len(g.battleUnits) {
		return nil
	}
	f := g.battleFleet(g.battleUnits[unit].Fleet)
	if f == nil {
		return nil
	}
	turn, perTurn := g.fleetSpeeds(g.battleUnits[unit].Fleet, g.fleetOrderBattle)
	dist := perTurn / rules.FixOne
	if turn <= 0 || dist <= 0 {
		return nil
	}
	u := g.battleUnits[unit]
	return &rangeShape{
		cx: u.X + battleUnitTile/2, cy: u.Y + battleUnitTile/2,
		turn: turn, dist: dist, facing: f.Facing,
	}
}
