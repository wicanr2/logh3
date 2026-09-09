package game

import "github.com/wicanr2/logh3/internal/screen"

// 戰場上點行星開的命令選單（原版 `1D9E:01F1`，`docs/re/133`）。
//
// 沒有要塞的行星開三項（情報／補給／補充，`ds:42EA`，內寬 2），有要塞的
// 開四項（多「要塞砲」，`ds:4318`，內寬 3）。點誰的行星都開得起來——
// 不是自己的就只剩「情報」可選。

// 行星在戰場上佔一格（地圖一格 16×20），座標記的是**格中央**
// （`rules.PlanetSites`：欄×16＋8、列×20＋12）。原版的命中判定要等
// `341B:039D` 的分層邏輯讀出來，這裡先用整格。
const (
	planetCellW = 16
	planetCellH = 20
)

// battlePlanetAt 回畫面座標 (x,y) 底下是第幾顆行星，-1 ＝ 沒有。
func (g *Game) battlePlanetAt(x, y int) int {
	s := g.res.Battle
	if s == nil {
		return -1
	}
	fx, fy := s.Field.X, s.Field.Y
	ox := g.battleView.Col * s.Field.TileW
	oy := g.battleView.Row * s.Field.TileH
	for i, p := range g.battlePlanets {
		sx := fx + p.X - ox - planetCellW/2
		sy := fy + p.Y - oy - planetCellH/2 - 2 // 格中央是 +12 不是 +10
		if x >= sx && x < sx+planetCellW && y >= sy && y < sy+planetCellH {
			return i
		}
	}
	return -1
}

// planetMenuKind 回這顆行星該開哪一張選單。
func (g *Game) planetMenuKind(i int) screen.CommandMenuKind {
	if i >= 0 && i < len(g.battlePlanets) && g.battlePlanets[i].Fortress {
		return screen.CommandMenuFortress
	}
	return screen.CommandMenuPlanet
}

// planetMenuEnabled 回行星選單各項可不可選（`1D9E:0220` 起那一段）。
//
//	三項（或四項）先全開
//	行星擁有者 != 玩家陣營 → 關掉「補給」「補充」（有要塞的連「要塞砲」）
//	是自己的、而且有要塞：`+0x45` 不是 0 也不是 3 → 關掉「要塞砲」
//
// 所以**別人的行星只剩「情報」**。`+0x45` 是要塞砲自己的狀態，開局
// 81 顆全是 0；它什麼時候變還沒解，remake 先當成一直可以打。
func (g *Game) planetMenuEnabled(i int) []bool {
	kind := g.planetMenuKind(i)
	n := g.res.CommandMenus.ItemCount(kind)
	on := make([]bool, n)
	for k := range on {
		on[k] = true
	}
	if i < 0 || i >= len(g.battlePlanets) {
		return on
	}
	if g.battlePlanets[i].Owner != g.faction {
		for k := 1; k < n; k++ {
			on[k] = false
		}
	}
	return on
}

// openPlanetMenu 在游標處開行星的命令選單。
func (g *Game) openPlanetMenu(planet, x, y int) {
	s := g.res.CommandMenus
	if s == nil || planet < 0 || planet >= len(g.battlePlanets) {
		return
	}
	kind := g.planetMenuKind(planet)
	n := s.ItemCount(kind)
	if n == 0 {
		return
	}
	g.battleCmd = true
	g.battleCmdUnit = -1
	g.battleCmdPlanet = planet
	g.battleCmdKind = kind
	g.battleCmdX, g.battleCmdY = s.CommandMenuAt(x, y, n)
}

// PlanetMenuOpen 回現在開著的是不是行星的命令選單。
func (g *Game) PlanetMenuOpen() bool {
	return g.battleCmd && g.battleCmdPlanet >= 0
}

// planetMenuClick 處理行星選單開著時的左鍵。四項各自對應一支常式
// （`ds:4396` 的跳躍表）：情報 `1373:1C71`、補給 `1930:0DA2`、
// 補充 `1AAB:0A33`、要塞砲 `2264:0049`——**四支都還沒逆向**，所以
// 現在只把選單關掉。
func (g *Game) planetMenuClick(x, y, n int) bool {
	s := g.res.CommandMenus
	planet := g.battleCmdPlanet
	row := s.CommandMenuRowAt(g.battleCmdKind, g.battleCmdX, g.battleCmdY, x, y, n)
	on := g.planetMenuEnabled(planet)
	g.closeCommandMenu()
	if row < 0 || (row < len(on) && !on[row]) {
		return true
	}
	return true
}
