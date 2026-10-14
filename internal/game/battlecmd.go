package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 戰場的命令選單（`docs/spec/023`）。
//
// 點自己陣營的分隊會開一張命令選單：旗艦（分隊序號 0）開艦隊選單
// 「移動／戰鬥／索敵」，其餘按艦種開四張分隊選單之一。點不到分隊
// （或點到別人的）就照舊開戰術選單。

// battleUnitAt 回畫面座標 (x,y) 底下是哪一個分隊，-1 ＝ 沒有。
//
// 單位是 16×16 的圖塊、左上角對齊世界座標，與 `DrawBattleUnits` 同一組
// 幾何。重疊時取**畫在最上面的那一個**（繪製順序的最後一個）——原版的
// 命中判定還沒逆向到這一層，這是 remake 的取捨（`docs/spec/023`）。
func (g *Game) battleUnitAt(x, y int) int {
	s := g.res.Battle
	if s == nil {
		return -1
	}
	fx, fy := s.Field.X, s.Field.Y
	ox := g.battleView.Col * s.Field.TileW
	oy := g.battleView.Row * s.Field.TileH
	for i := len(g.battleUnits) - 1; i >= 0; i-- {
		u := g.battleUnits[i]
		if u.Out {
			continue // 已回收／離場單位與正式renderer一致，不攔截點選（規格080）。
		}
		sx, sy := fx+u.X-ox, fy+u.Y-oy
		if x >= sx && x < sx+battleUnitTile && y >= sy && y < sy+battleUnitTile {
			return i
		}
	}
	return -1
}

// battleUnitTile 是戰場上一個分隊的圖塊邊長（`SPRITE.SPR`，`docs/re/85`）。
const battleUnitTile = 16

// 分隊行動（艦艇記錄 `+0x1E`）。0x7F ＝ 沒有自己的命令。
const (
	squadronMove   = 0
	squadronForm   = 1
	squadronScout  = 2
	squadronRejoin = 3
	squadronNone   = 0x7F
)

// 艦種：5 運補艦、6 修護艦、7 索敵分遣隊（`docs/internals/40`）。
const (
	shipKindSupply = 5
	shipKindRepair = 6
	shipKindScout  = 7
)

// commandMenuKind 回點到這個分隊該開哪一張選單。
func commandMenuKind(u rules.BattleUnit) screen.CommandMenuKind {
	if u.Squadron == 0 {
		return screen.CommandMenuFleet
	}
	switch u.Kind {
	case shipKindSupply:
		return screen.CommandMenuSupply
	case shipKindRepair:
		return screen.CommandMenuRepair
	case shipKindScout:
		return screen.CommandMenuScout
	}
	return screen.CommandMenuSquadron
}

// commandMenuEnabled 回每一項可不可選（`docs/spec/023`）。
func (g *Game) commandMenuEnabled(kind screen.CommandMenuKind, u rules.BattleUnit) []bool {
	if kind == screen.CommandMenuFleet {
		// 三項先全開，再按艦隊記錄關掉。
		on := []bool{true, true, true}
		f := g.battleFleet(u.Fleet)
		if f == nil {
			return on
		}
		if f.Flag0 {
			on[1], on[2] = false, false
		}
		if f.ScoutOut {
			on[2] = false
		}
		return on
	}
	// 分隊選單四張共用同一條：已經在移動或已經在歸隊就不能再下移動。
	move := u.Action != squadronMove && u.Action != squadronRejoin
	n := g.res.CommandMenus.ItemCount(kind)
	on := make([]bool, n)
	for i := range on {
		on[i] = true
	}
	if n > 0 {
		on[0] = move
	}
	return on
}

// battleFleet 回戰場上這一支艦隊的記錄；找不到回 nil。
func (g *Game) battleFleet(index int) *rules.BattleFleet {
	for i := range g.battleFleets {
		if g.battleFleets[i].Index == index {
			return &g.battleFleets[i]
		}
	}
	return nil
}

// openCommandMenu 在游標處開命令選單。
func (g *Game) openCommandMenu(unit, x, y int) {
	s := g.res.CommandMenus
	if s == nil || unit < 0 || unit >= len(g.battleUnits) {
		return
	}
	kind := commandMenuKind(g.battleUnits[unit])
	n := s.ItemCount(kind)
	if n == 0 {
		return
	}
	g.battleCmd = true
	g.battleCmdUnit = unit
	g.battleCmdPlanet = -1
	g.battleCmdKind = kind
	g.battleCmdX, g.battleCmdY = s.CommandMenuAt(x, y, n)
}

// closeCommandMenu 關掉命令選單。
func (g *Game) closeCommandMenu() { g.battleCmd, g.battleCmdPlanet = false, -1 }

// CommandMenuOpen 回命令選單開著沒有。
func (g *Game) CommandMenuOpen() bool { return g.battleCmd }

// commandMenuClick 處理命令選單開著時的左鍵；回有沒有吃掉這一下。
func (g *Game) commandMenuClick(x, y int) bool {
	if !g.CommandMenuOpen() {
		return false
	}
	s := g.res.CommandMenus
	n := s.ItemCount(g.battleCmdKind)
	if g.battleCmdPlanet >= 0 {
		return g.planetMenuClick(x, y, n)
	}
	unit := g.battleCmdUnit
	if unit < 0 || unit >= len(g.battleUnits) {
		g.closeCommandMenu()
		return true
	}
	row := s.CommandMenuRowAt(g.battleCmdKind, g.battleCmdX, g.battleCmdY, x, y, n)
	on := g.commandMenuEnabled(g.battleCmdKind, g.battleUnits[unit])
	g.closeCommandMenu()
	if row < 0 || (row < len(on) && !on[row]) {
		return true // 點在選單外面或點到不可選的：只把選單關掉
	}
	// 「歸隊」沒有子常式，選下去就是寫一個位元組（`docs/re/117`）。
	if g.battleCmdKind == screen.CommandMenuScout && row == 1 {
		g.battleUnits[unit].Action = squadronRejoin
	}
	// 第一項是「移動」：分隊直接等點目的地，旗艦（艦隊選單）走兩段
	// （點目的地 → 選抵達朝向，`docs/spec/027`）。
	if row == 0 {
		if g.battleCmdKind == screen.CommandMenuFleet {
			g.beginFleetOrder(unit, false)
		} else {
			g.beginMoveTarget(unit)
		}
	}
	// 艦隊選單第二項是「戰鬥」：與「移動」同一條路，成功之後艦隊行動
	// 寫成 1（`docs/re/120`）。
	if row == 1 && g.battleCmdKind == screen.CommandMenuFleet {
		g.beginFleetOrder(unit, true)
	}
	// 「編隊」是分隊三張表的第二項：接著等玩家點隊形格
	// （`docs/internals/46`）。索敵那張的第二項是「歸隊」，不算。
	if row == 1 && g.battleCmdKind != screen.CommandMenuFleet &&
		g.battleCmdKind != screen.CommandMenuScout {
		g.beginFormationPick(unit)
	}
	// 運補艦與修護艦那兩張表的第三項是「補給」／「修理」：接著等玩家
	// 點對象（`docs/re/131`）。艦隊選單第三項是「索敵」（`docs/re/132`）。
	if row == 2 {
		switch g.battleCmdKind {
		case screen.CommandMenuSupply:
			g.beginSupplyPick(unit, false)
		case screen.CommandMenuRepair:
			g.beginSupplyPick(unit, true)
		case screen.CommandMenuFleet:
			g.beginScout(unit)
		}
	}
	// 其餘各項的行為還沒逆向（`docs/spec/023`「還沒做的」）。
	return true
}

// 正常Input的純移動摘要依規格088及研究389；命令選單沿023既有範圍。

func (g *Game) battleInputIdle() bool {
	return g.state == StateBattle && g.battlePhase == screen.BattlePhaseInput &&
		!g.fortressEngagement &&
		!g.fleetCombat.active &&
		!g.battlePreparing && !g.battleMoving && !g.battleSpotting.active &&
		!g.battleMenu && !g.battleCmd && !g.battleMove && !g.battleForm &&
		!g.battleSupply && g.fleetOrder == FleetOrderNone && !g.battleGunOn &&
		!g.planetInfoOn && !g.battleData && !g.battleDetail && !g.msgOn && !g.battleScrollHeld
}

func (g *Game) refreshBattleHover() {
	if !g.battleInputIdle() {
		return
	}
	i := g.battleUnitAt(g.mouseX, g.mouseY)
	if i >= 0 && g.battleUnits[i].Faction != g.faction && !g.showEnemies {
		i = -1 // 敵方摘要仍待原版取證，不使用己方完整摘要洩漏資料。
	}
	if i == g.battleHoverUnit {
		return
	}
	g.battleHoverUnit, g.battleHoverOn = i, false
	if i < 0 {
		return
	}
	v, ok := g.unitPanelFor(i)
	if !ok {
		return
	}
	u := g.battleUnits[i]
	g.battleHoverPanel, g.battleHoverOn = v, true
	g.battleHoverCurrent = [2]int{u.X, u.Y}
	g.battleHoverTarget = [2]int{u.TargetX, u.TargetY}
	g.battleHoverMinimapLine = false
	if u.Squadron == 0 {
		if f := g.battleFleet(u.Fleet); f != nil {
			g.battleHoverTarget = [2]int{f.DestX, f.DestY}
			g.battleHoverMinimapFrom = [2]int{f.X, f.Y}
			g.battleHoverMinimapLine = true
		}
	}
}

// 艦種列上這一層讀的四個「每艘幾件」（`docs/internals/40`）。
const (
	kindBeamPer     = 0x0B
	kindLauncherPer = 0x0F
	kindMissilePer  = 0x10
	kindFighterPer  = 0x13
	kindFighterFull = 0x14
)

// unitPanelView 組出分隊面板要顯示的內容。
func (g *Game) unitPanelView() (screen.UnitPanelView, bool) {
	if g.battleInputIdle() && g.battleHoverOn {
		return g.battleHoverPanel, true
	}
	sel, ok := g.selectedUnit()
	if !ok {
		return screen.UnitPanelView{}, false
	}
	return g.unitPanelFor(sel)
}

func (g *Game) unitPanelFor(sel int) (screen.UnitPanelView, bool) {
	var v screen.UnitPanelView
	s := g.res.UnitPanel
	if s == nil || sel < 0 || sel >= len(g.battleUnits) {
		return v, false
	}
	u := g.battleUnits[sel]
	f := g.fleetByIndex(u.Fleet)
	if f == nil {
		return v, false
	}
	v.Faction = u.Faction
	v.FleetNo = f.Number
	v.Flagship = u.Squadron == 0
	v.Scout = u.Kind == shipKindScout
	v.UnitNo = u.Squadron + 1
	if v.Flagship {
		v.Admiral = g.res.AdmiralNames[f.Admiral]
	}
	if fd := g.res.FleetDetail; fd != nil && u.Kind >= 0 && u.Kind < len(fd.ShipKinds) {
		v.KindName = fd.ShipKinds[u.Kind].Codes
	}
	v.Ships = u.Ships
	bf := g.battleFleet(u.Fleet)

	// 士氣：旗艦印艦隊的（建立艦隊時寫死 100），其餘印自己那筆。
	if v.Flagship {
		v.FleetMorale, v.Morale = true, f.Morale
		if bf != nil {
			v.Morale = bf.Integrity // 原版標籤為士氣，實際欄位仍是FLEET+3C。
		}
	} else {
		v.Morale = u.Morale
	}

	// 行動：分隊沒有自己的命令（`+0x1E ＝ 0x7F`）就印艦隊的。
	switch {
	case u.Action != squadronNone && u.Action < len(s.SquadronActions):
		v.UnitAction = true
		v.ActionName = s.SquadronActions[u.Action].Codes
		v.ActionInk = s.Ink
	case bf != nil && bf.Flag0:
		v.CannotFight = true
	default:
		a := f.Action
		if bf != nil {
			a = bf.Action & 0x0F
		}
		if a >= 0 && a < len(s.FleetActions) {
			v.ActionName = s.FleetActions[a].Codes
			v.ActionInk = s.FleetActionInkOf(a)
		}
	}

	// 三組武器。光線砲印的是算出來的量，另外兩組印艦艇記錄裡的現有量，
	// 顏色由「滿不滿」決定。
	if k, ok := g.res.ShipKinds.Kind(u.Faction, u.Kind); ok {
		v.Beam = u.Ships * k.Byte(kindBeamPer)
		v.Missile, v.Fighter = u.Missiles, u.Fighters
		v.MissileInk = weaponInk(s, u.Missiles,
			u.Ships*k.Byte(kindMissilePer), u.Ships*k.Byte(kindLauncherPer))
		v.FighterInk = weaponInk(s, u.Fighters,
			u.Ships*k.Byte(kindFighterFull), u.Ships*k.Byte(kindFighterPer))
	}
	return v, true
}

// selectedUnit 回現在選著的分隊——命令選單開著，或是選了「移動」正在等
// 玩家點目的地。收據 `s3cmdrange` 上選目的地的時候面板還在。
func (g *Game) selectedUnit() (int, bool) {
	i := -1
	switch {
	case g.battleCmd:
		i = g.battleCmdUnit
	case g.battleMove:
		i = g.battleMoveUnit
	}
	if i < 0 || i >= len(g.battleUnits) {
		return -1, false
	}
	return i, true
}

// weaponInk 回武器數的顏色：等於滿量是本色，達到門檻是色 5，否則色 2
// （`0x1CFFE` 起那兩段比較，`docs/re/118`）。
func weaponInk(s *screen.UnitPanelSpec, have, full, part int) uint8 {
	switch {
	case have == full:
		return s.WeaponInk.Full
	case have >= part:
		return s.WeaponInk.Part
	}
	return s.WeaponInk.Low
}

// 「移動」：選一個目的地（`docs/re/119`、`docs/re/121`、規格 `docs/spec/024`）。
//
// 分隊層那支（原版 `2583:00DE`）沒有速度上限也不畫範圍——寫進艦艇
// `+0x28`／`+0x2A` 的就是點到的世界座標，逐位元組驗過（收據 `s3cmdmove`）。
// 艦隊層那支（旗艦）要再選一次抵達朝向，走 `beginFleetOrder` 的兩段
// （`docs/spec/027`）。

// 世界座標的上下限（原版 `2583:022B` 的 `cmp ax,7EFh`／`cmp ax,7E7h`）。
const (
	battleWorldMaxX = 0x7EF // 2031
	battleWorldMaxY = 0x7E7 // 2023
)

// MoveTargetPending 回「正在等玩家點目的地」。
func (g *Game) MoveTargetPending() bool { return g.battleMove }

// beginMoveTarget 進入選目的地的狀態。
func (g *Game) beginMoveTarget(unit int) {
	g.battleMove, g.battleMoveUnit = true, unit
}

// cancelMoveTarget 取消選目的地。
func (g *Game) cancelMoveTarget() { g.battleMove = false }

// moveTargetClick 處理選目的地時的左鍵；回有沒有吃掉這一下。
//
// 點在戰場外、超出世界範圍或那一格過不去，原版跳「這是無法行動的
// 地方。」（`ds:4A24`）**然後回 `0xFFFF`**，也就是整個命令作廢
// （`docs/re/121`）。訊息框照畫（`docs/re/138`）。
func (g *Game) moveTargetClick(x, y int) bool {
	if !g.battleMove {
		return false
	}
	unit := g.battleMoveUnit
	g.cancelMoveTarget()
	if unit < 0 || unit >= len(g.battleUnits) {
		return true
	}
	wx, wy, ok := g.battleWorldAt(x, y)
	if !ok {
		g.showMessage("move_blocked")
		return true
	}
	g.battleUnits[unit].TargetX, g.battleUnits[unit].TargetY = wx, wy
	g.battleUnits[unit].Action = squadronMove
	return true
}

// battleWorldPixel 把畫面座標換成戰場的世界座標，只做邊界判斷。
// 瞄準要塞砲那類「只要方向」的用途走這一支（原版 `1588:0362`）；
// 要下移動命令才用 battleWorldAt，因為那還得過地形。
func (g *Game) battleWorldPixel(x, y int) (int, int, bool) {
	s := g.res.Battle
	if s == nil {
		return 0, 0, false
	}
	fw := s.Field.Cols * s.Field.TileW
	fh := s.Field.Rows * s.Field.TileH
	if x < s.Field.X || y < s.Field.Y || x >= s.Field.X+fw || y >= s.Field.Y+fh {
		return 0, 0, false
	}
	wx := x - s.Field.X + g.battleView.Col*s.Field.TileW
	wy := y - s.Field.Y + g.battleView.Row*s.Field.TileH
	if wx < 0 || wx > battleWorldMaxX || wy < 0 || wy > battleWorldMaxY {
		return 0, 0, false
	}
	return wx, wy, true
}

// battleWorldAt 把畫面座標換成戰場的世界座標，並照原版做邊界與地形
// 判斷（`2583:022B` ＋ `1588:04EA`）。第三個回傳值 false ＝ 不能去。
func (g *Game) battleWorldAt(x, y int) (int, int, bool) {
	s := g.res.Battle
	if s == nil {
		return 0, 0, false
	}
	fw := s.Field.Cols * s.Field.TileW
	fh := s.Field.Rows * s.Field.TileH
	if x < s.Field.X || y < s.Field.Y || x >= s.Field.X+fw || y >= s.Field.Y+fh {
		return 0, 0, false // 點在戰場那一塊外面
	}
	wx := x - s.Field.X + g.battleView.Col*s.Field.TileW
	wy := y - s.Field.Y + g.battleView.Row*s.Field.TileH
	if wx < 0 || wx > battleWorldMaxX || wy < 0 || wy > battleWorldMaxY {
		return 0, 0, false
	}
	if g.res.Terrain != nil &&
		!g.res.Terrain.Passable(g.battleCells(g.battleRegion), wx, wy) {
		return 0, 0, false
	}
	return wx, wy, true
}
