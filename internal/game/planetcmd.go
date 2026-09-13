package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 戰場上點行星開的命令選單（原版 `1D9E:01F1`，`docs/re/133`）。
//
// 沒有要塞的行星開三項（情報／補給／補充，`ds:42EA`，內寬 2），有要塞的
// 開四項（多「要塞砲」，`ds:4318`，內寬 3）。點誰的行星都開得起來——
// 不是自己的就只剩「情報」可選。

// 行星在戰場上佔一格（地圖一格 16×20），座標記的是**格中央**
// （`rules.PlanetSites`：欄×16＋8、列×20＋12）。原版 `341B:039D`
// 也先把像素換成 16×20 格，再按已註冊視窗的順序判最上層；戰場 viewport
// 的原點與大小都是完整格，所以此處的整格判定等價（`docs/re/136`）。
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
// 補充 `1AAB:0A33`、要塞砲 `2264:0049`。
func (g *Game) planetMenuClick(x, y, n int) bool {
	s := g.res.CommandMenus
	planet := g.battleCmdPlanet
	row := s.CommandMenuRowAt(g.battleCmdKind, g.battleCmdX, g.battleCmdY, x, y, n)
	on := g.planetMenuEnabled(planet)
	g.closeCommandMenu()
	if row < 0 || (row < len(on) && !on[row]) {
		return true
	}
	switch row {
	case 0:
		// 「情報」開一個 22 × 10 格的資訊視窗（`1373:1CBC`，
		// `docs/re/144`）。原版按任一鍵關掉。
		g.planetInfoOn, g.planetInfo = true, planet
	case 1:
		g.beginPlanetSupplyPick(planet, false)
	case 2:
		g.beginPlanetSupplyPick(planet, true)
	case 3:
		g.beginFortressTarget(planet)
	}
	return true
}

// FortressTargetPending 回「正在等玩家替要塞砲指定目標」。
func (g *Game) FortressTargetPending() bool { return g.battleGunOn }

// beginFortressTarget 進入要塞砲指定目標的狀態（`2264:0049`）。
func (g *Game) beginFortressTarget(planet int) {
	if planet < 0 || planet >= len(g.battlePlanets) {
		return
	}
	if !rules.FortressGunReady(g.battlePlanets[planet].GunState) {
		return
	}
	g.battleGunOn, g.battleGun, g.battleGunAim = true, planet, 0
	g.fortressAimMouse(g.mouseX, g.mouseY)
}

// fortressAimMouse 讓射程橢圓跟著游標轉（原版每一圈都重算一次，
// 方位沒變就不重畫，`2264:0133`）。
func (g *Game) fortressAimMouse(x, y int) {
	if !g.battleGunOn {
		return
	}
	if a, ok := g.fortressAimAt(x, y); ok {
		g.battleGunAim = a
	}
}

// cancelFortressTarget 右鍵：取消瞄準（原版寫 `+0x50` ← 0xFFFF、
// `+0x45` ← 0）。
func (g *Game) cancelFortressTarget() {
	if !g.battleGunOn {
		return
	}
	g.battleGunOn = false
	if g.battleGun < 0 || g.battleGun >= len(g.battlePlanets) {
		return
	}
	p := &g.battlePlanets[g.battleGun]
	p.GunAim, p.GunState = rules.FortressGunNoAim, rules.FortressGunIdle
}

// fortressTargetClick 左鍵：把游標的**方位**記下來（`+0x50` ← 方位、
// `+0x45` ← 3）。點的不是某一支分隊——要塞砲掃的是一個橢圓，
// 誰在裡面是開火那一刻才算的（`docs/re/145`）。
func (g *Game) fortressTargetClick(x, y int) bool {
	if !g.battleGunOn {
		return false
	}
	planet := g.battleGun
	g.battleGunOn = false
	if planet < 0 || planet >= len(g.battlePlanets) {
		return true
	}
	aim, ok := g.fortressAimAt(x, y)
	if !ok {
		return true
	}
	p := &g.battlePlanets[planet]
	p.GunAim, p.GunState = aim, rules.FortressGunAimed
	return true
}

// fortressAimAt 回「游標在 (x,y) 時要塞砲的方位」。x／y 是畫面座標。
// 游標正好落在要塞中心那一格時方位不變（原版直接回舊值，
// `2264:0133`）。
func (g *Game) fortressAimAt(x, y int) (int, bool) {
	if !g.battleGunOn || g.battleGun < 0 || g.battleGun >= len(g.battlePlanets) {
		return 0, false
	}
	p := g.battlePlanets[g.battleGun]
	wx, wy, ok := g.battleWorldPixel(x, y)
	if !ok {
		return 0, false
	}
	cx, cy := p.X+fortressCentreBias, p.Y+fortressCentreBias
	if wx == cx && wy == cy {
		return 0, false
	}
	return rules.Atan256(wx-cx, wy-cy), true
}

// FortressAim 回要塞砲正在瞄的方位；沒在瞄就回 false。
func (g *Game) FortressAim() (int, bool) {
	if !g.battleGunOn {
		return 0, false
	}
	return g.battleGunAim, true
}

// FortressTargetRange 回正在瞄準的那座要塞的長短半軸。**吃的是防衛力
// 上限**（記錄 `+0x16`），不是目前防衛力（收據 `s3fort`：3300 → 96／24）。
func (g *Game) FortressTargetRange() (outer, inner int, ok bool) {
	if !g.battleGunOn || g.battleGun >= len(g.battlePlanets) {
		return 0, 0, false
	}
	o, i := rules.FortressRange(g.battlePlanets[g.battleGun].DefenceMax)
	return o, i, true
}

// fortressCentreBias 是「行星記錄的座標 → 佔位框中心」的偏移
// （原版 `行星[+0x46] ＋ 7`、`行星[+0x48] ＋ 7`）。
const fortressCentreBias = 7

// PlanetInfoOpen 回行星情報視窗開著沒有。
func (g *Game) PlanetInfoOpen() bool { return g.planetInfoOn }

// closePlanetInfo 關掉情報視窗。原版按任一鍵就關（`1373:1C71`）。
func (g *Game) closePlanetInfo() { g.planetInfoOn = false }

// planetInfoView 回情報視窗要顯示的東西；沒開就回 false。
func (g *Game) planetInfoView() (screen.PlanetInfo, bool) {
	if !g.planetInfoOn || g.res.PlanetInfo == nil ||
		g.planetInfo < 0 || g.planetInfo >= len(g.battlePlanets) {
		return screen.PlanetInfo{}, false
	}
	p := g.battlePlanets[g.planetInfo]
	v := screen.PlanetInfo{
		RegionOwner: g.regionOwnerOf(g.battleRegion),
		RegionName:  g.regionNameBytesOf(g.battleRegion),
		PlanetOwner: p.Owner,
		PlanetName:  p.NameBytes,
		Fortress:    p.Fortress,
		Own:         p.Owner == g.faction,
		Economy:     int64(p.Economy),
		Defence:     p.Defence,
		DefenceMax:  p.DefenceMax,
		Recover:     p.Recover,
		Missiles:    int64(p.Stock.Missiles),
		Fighters:    int64(p.Stock.Fighters),
	}
	for i := range v.Kinds {
		v.Kinds[i] = p.Stock.Ships[i]
	}
	return v, true
}

// regionRecordOf 回第 i 個星域的記錄；越界回 nil。
func (g *Game) regionRecordOf(i int) []byte {
	if i < 0 || (i+1)*rules.RegionRecord > len(g.res.Stellar) {
		return nil
	}
	return g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
}

// regionOwnerOf 回星域記錄的 `+0x12`（0 帝國、1 同盟、2 無主）。
func (g *Game) regionOwnerOf(i int) int {
	r := g.regionRecordOf(i)
	if len(r) <= 0x12 {
		return 2
	}
	return int(r[0x12])
}

// regionNameBytesOf 回星域名那串 Big5（記錄開頭，`0x00` 結尾）。
func (g *Game) regionNameBytesOf(i int) []byte {
	r := g.regionRecordOf(i)
	for j, b := range r {
		if b == 0 {
			out := make([]byte, j)
			copy(out, r[:j])
			return out
		}
	}
	return nil
}
