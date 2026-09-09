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
			// 與分隊的「移動」同一條檢查鏈；原版會跳「這是無法行動的
			// 地方。」，remake 只取消（`docs/spec/024`）。
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
	f.DestX, f.DestY = x, y
	f.TargetFacing = facing
	f.Ordered = true
	if battle {
		f.Action = fleetActionBattle
	}
	// 艦隊層的移動引擎（`rules.AdvanceFleet`）還沒接進回合流程，
	// remake 目前逐分隊各自走——所以先讓旗艦朝新目的地移動，
	// 畫面上才看得到命令生效（差異記在 `docs/spec/027`）。
	g.battleUnits[unit].TargetX, g.battleUnits[unit].TargetY = x, y
	g.battleUnits[unit].Action = squadronMove
}

// fleetActionBattle 是艦隊記錄 `+0x48` 低四位的「戰鬥」（`docs/internals/33`）。
const fleetActionBattle = 1
