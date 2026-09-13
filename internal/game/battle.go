package game

import (
	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// battleSprite 回一個分隊要畫哪一塊 `SPRITE.SPR`（`rules.BattleSprite`）。
func battleSprite(u rules.BattleUnit, viewer int) int { return rules.BattleUnitSprite(u, viewer) }

// StartBattle 在某個星域開一場戰鬥。
//
// 參戰的是所在星域等於 region 的艦隊（`docs/internals/38`）。每一支照
// 入口表放到它的入口格上，再把分隊散開（`docs/internals/39`）；同一個
// 入口進來的第二支以後會相撞，原版會錯開 32 像素，錯開的順序還沒解，
// 所以這裡照順序往下堆。
//
// 視野放在第一支參戰艦隊那裡——原版怎麼選還沒解（`docs/internals/34`）。
func (g *Game) StartBattle(region int) {
	g.battleRegion = region
	g.battleTurn = 1
	g.battleUnits = nil
	g.battleFleets = nil
	g.battleScore = [2]int{}
	g.battleOver = false
	var joined []int
	if g.res.Entries == nil {
		return
	}
	seen := map[int]int{}
	for _, f := range g.fleets {
		if f.Region != region {
			continue
		}
		e, ok := g.res.Entries.EntryFor(region, f.Region)
		if !ok {
			// 從別處打進來的：原版查的是「來源星域」，remake 還沒有
			// 行軍來源，先用該星域的第一個入口。
			if len(g.res.Entries.Regions[region].Entries) == 0 {
				continue
			}
			e = g.res.Entries.Regions[region].Entries[0]
		}
		n := seen[e.From]
		seen[e.From]++
		x, y := e.X, e.Y-n*battleFleetGap
		facing := rules.BattleFacing(x, y)
		for _, u := range g.slot.Deploy(
			g.res.Weights, g.res.Formation, f, x, y, facing) {
			g.battleUnits = append(g.battleUnits, u)
		}
		joined = append(joined, f.Index)
	}
	g.battleFleets = g.slot.BattleFleetsFor(joined)
	g.battlePlanets = g.battleRegionPlanets(region)
	g.battleRally = g.battleRegionRally(region)
	g.battleMenu = false
	g.closeCommandMenu()
	g.cancelMoveTarget()
	g.cancelFormationPick()
	g.cancelSupplyPick()
	g.cancelFleetOrder()
	g.battleGunOn = false
	for i := range g.battleFleets {
		if u := flagshipOf(g.battleUnits, g.battleFleets[i].Index); u != nil {
			g.battleFleets[i].Facing = rules.BattleFacing(u.X, u.Y)
		}
	}
	if len(g.battleUnits) > 0 {
		// 缺規格時用量到的預設值（`docs/internals/36`）。
		tw, th, cols, rows := asset.BattleTileW, asset.BattleTileH, 28, 18
		if g.res.Battle != nil {
			tw, th = g.res.Battle.Field.TileW, g.res.Battle.Field.TileH
			cols, rows = g.res.Battle.Field.Cols, g.res.Battle.Field.Rows
		}
		u := g.battleUnits[0]
		g.battleView = screen.BattleView{
			Col: clampCol(u.X/tw-cols/2, cols),
			Row: clampRow(u.Y/th-rows/2, rows),
		}
	}
	g.state = StateBattle
}

// battleFleetGap 是同一個入口進來的艦隊之間錯開多少（`docs/internals/38`）。
const battleFleetGap = 32

func clampCol(c, cols int) int {
	if c < 0 {
		return 0
	}
	if c > asset.BattleMapCols-cols {
		return asset.BattleMapCols - cols
	}
	return c
}

func clampRow(r, rows int) int {
	if r < 0 {
		return 0
	}
	if r > asset.BattleMapRows-rows {
		return asset.BattleMapRows - rows
	}
	return r
}

// AdvanceBattleTurn 推進一個戰鬥回合：兩邊各下一次令，然後所有分隊走一步。
//
// 損失怎麼算還沒解（`docs/re/86`），所以這裡只有移動。
func (g *Game) AdvanceBattleTurn() {
	if g.battleRand == nil {
		g.battleRand = rules.NewRand(uint32(g.battleTurn))
	}
	rules.DecaySpotting(g.battleUnits)
	for _, faction := range []int{0, 1} {
		if faction == g.faction && !g.autoBattle {
			continue
		}
		rules.RevealEnemiesIfBlind(g.battleUnits, faction, g.battleTurn)
	}
	if g.tactic != nil {
		b := ai.Battlefield{
			Units: g.battleUnits, Planets: g.battlePlanets,
			Rally: g.battleRally, Terrain: g.res.Terrain,
			Cells: g.battleCells(g.battleRegion), Rand: g.battleRand,
		}
		for _, faction := range []int{0, 1} {
			if faction == g.faction && !g.autoBattle {
				continue // 玩家自己下令
			}
			for _, m := range g.tactic.Moves(b, faction, g.battleTurn) {
				if m.Unit >= 0 && m.Unit < len(g.battleUnits) {
					g.battleUnits[m.Unit].TargetX = m.X
					g.battleUnits[m.Unit].TargetY = m.Y
				}
			}
		}
	}
	// 收到過命令的艦隊走艦隊層那一套：艦隊自己動，分隊的目的地跟著
	// 換成隊形位置（`docs/re/129`、`docs/re/126`）。沒下過命令的
	// 仍然逐分隊各自走——原版兩層本來就各走各的，這裡只是還沒把
	// 全部艦隊搬過去（`docs/spec/027`）。
	g.advanceOrderedFleets()
	for i := range g.battleUnits {
		rules.AdvanceBattle(&g.battleUnits[i], g.res.Terrain, g.battleCells(g.battleRegion))
	}
	g.battleLosses = rules.ResolveBattleTurn(g.battlePlanets, g.battleFleets,
		g.battleUnits, g.battleRand)
	for _, l := range g.battleLosses {
		if l.Scorer >= 0 && l.Scorer < 2 {
			g.battleScore[l.Scorer] += l.Score
		}
	}
	// 原版一回合八步，第五步是要塞砲開火（`181B:0385`，規格
	// `docs/spec/028`）、第六步是打行星（`181B:03D0`，規格
	// `docs/spec/019`）。**順序有意義**：瞄好的要塞砲會讓敵方那一回合
	// 打不到這顆行星，所以倒數要排在打行星之後（`docs/re/149`）。
	g.battleGunHits = rules.FortressGunFire(g.battlePlanets, g.battleFleets,
		g.battleUnits, g.res.ShipKinds, g.battleRand)
	for _, h := range g.battleGunHits {
		if h.Scorer >= 0 && h.Scorer < 2 {
			g.battleScore[h.Scorer] += h.Score
		}
	}
	// 打行星讀艦隊的編制完整度（`+0x3C`），先重算一次（`docs/re/153`）。
	rules.RefreshFleetIntegrity(g.battleFleets, g.battleUnits)
	planetOwners := make([]int, len(g.battlePlanets))
	for i := range g.battlePlanets {
		planetOwners[i] = g.battlePlanets[i].Owner
	}
	g.battleAssaults = rules.PlanetAssault(g.battlePlanets, g.battleFleets,
		g.battleUnits, g.res.ShipKinds, g.battleRand)
	for _, h := range g.battleAssaults {
		if h.Planet >= 0 && h.Planet < len(planetOwners) {
			owner := planetOwners[h.Planet]
			if owner == 0 || owner == 1 {
				rules.RevealFleet(g.battleUnits, h.Fleet, 2)
			}
		}
		if h.Scorer >= 0 && h.Scorer < 2 {
			g.battleScore[h.Scorer] += h.Score
		}
	}
	rules.FortressGunTickAll(g.battlePlanets)
	rules.UpdateSpotting(g.battleUnits, g.battlePlanets)
	g.battleTurn++
	// 九個回合打完就結束（`docs/re/92`）；中途一方歸零也結束。
	g.battleOver = g.battleTurn > rules.MaxBattleTurns || battleDecided(g.battleUnits)
	if g.battleOver {
		g.EndBattle()
	}
}

// battleClick 處理戰術畫面的左鍵。
//
// 原版每一回合結束時在**游標所在位置**開一張戰術選單（`12AA:02D4`），
// 第一項「實行」正好落在游標底下（`docs/internals/34`）——所以同一點
// 連點兩下就等於推進一個回合：第一下開選單，第二下落在第一項上。
//
// `12AA:02D4` 的回傳值就是選了哪一項（`docs/re/98`）：選「實行」回 0，
// 主迴圈繼續下一個回合；其餘六項回非 0，**跳出主迴圈**去處理那個命令。
//
// 後四項的字串與戰略地圖上「系統指令」的後四項一字不差，所以接到同一
// 組畫面。**離開之後會不會回到這一場戰鬥還沒逆向**——這裡不動戰場狀態
// （不寫回損失、不清單位），純粹換畫面，回得來就接得上。
//
// 第二項「資料」開的是什麼還沒逆向，所以只把選單關掉。
func (g *Game) battleClick(x, y int) {
	// 訊息框是模態的，開著的時候什麼都不用往下傳（`docs/re/138`）。
	if g.messageClick(x, y) {
		return
	}
	// 行星情報視窗按任一鍵關掉（`docs/re/144`）。
	if g.planetInfoOn {
		g.closePlanetInfo()
		return
	}
	if g.battleDataClick(x, y) {
		return
	}
	if g.moveTargetClick(x, y) {
		return
	}
	if g.formationPickClick(x, y) {
		return
	}
	if g.supplyPickClick(x, y) {
		return
	}
	if g.fleetOrderClick(x, y) {
		return
	}
	if g.fortressTargetClick(x, y) {
		return
	}
	if g.commandMenuClick(x, y) {
		return
	}
	if !g.battleMenu {
		// 先看有沒有點到自己陣營的分隊：有的話開命令選單，沒有才開
		// 戰術選單（`docs/spec/023`）。
		if u := g.battleUnitAt(x, y); u >= 0 && g.battleUnits[u].Faction == g.faction {
			g.openCommandMenu(u, x, y)
			if g.CommandMenuOpen() {
				return
			}
		}
		// 分隊底下才輪到行星。**誰的行星都開得起來**（原版索引最高位元
		// 設起來就代表點到行星，一律可以），不是自己的只剩「情報」
		// （`docs/re/133`）。
		if p := g.battlePlanetAt(x, y); p >= 0 {
			g.openPlanetMenu(p, x, y)
			if g.CommandMenuOpen() {
				return
			}
		}
		if g.res.BattleMenu == nil {
			return
		}
		g.battleMenuX, g.battleMenuY = g.res.BattleMenu.BattleMenuAt(x, y)
		g.battleMenu = true
		return
	}
	row := g.res.BattleMenu.BattleMenuRowAt(g.battleMenuX, g.battleMenuY, x, y)
	g.battleMenu = false
	switch row {
	case 0: // 實行
		g.AdvanceBattleTurn()
	case 1: // 資料
		g.OpenBattleData()
	case 2: // 撤退
		g.Withdraw()
	case 3, 4, 5, 6: // 取進度／存進度／設定／結束
		if to := battleMenuScreen[row]; g.commandSpec(to) != nil {
			g.state = to
		}
	}
}

// battleMenuScreen 是戰術選單後四項對到的畫面。
//
// 字串與系統指令的後四項相同（`ds:334F` 與戰略地圖那張是各自一份，
// 但內容一字不差），所以走同一組畫面。
var battleMenuScreen = map[int]State{
	3: StateLoadGame,
	4: StateSaveGame,
	5: StateSettings,
	6: StateQuitConfirm,
}

// Withdraw 讓一方退出戰鬥——戰術選單的第三項。
func (g *Game) Withdraw() {
	if g.state != StateBattle {
		return
	}
	g.battleOver = true
	g.EndBattle()
}

// EndBattle 收掉戰場：把剩下的艘數寫回艦艇池、把行星狀態寫回行星表，
// 重算戰力，回戰略地圖。
//
// **星域不會易主**：兩份收據量過了（`docs/re/105`），連把一方整個抹掉
// 讓戰鬥提早結束，星域記錄的 `+0x12` 都沒有被寫。所以這裡不動星域歸屬
// 不是偷懶，是照原版——原版是**回合結束**依行星重算的（`2703:159B`、
// `docs/re/244`）。
//
// **行星那一半要寫回來**：原版整場都直接改行星記錄，佔領、防衛力損失、
// 要塞砲狀態與被打掉的存量打完就留在那裡（`docs/re/150`）。remake 為了
// 讓戰場層是純函式而先複製一份，所以這裡補寫回去。
//
// 戰後會不會撤退、分數怎麼結算還沒解。
func (g *Game) EndBattle() {
	if g.slot != nil {
		g.slot.ApplyBattleResult(g.battleUnits)
		g.fleets = g.slot.Fleets(g.res.Weights)
	}
	if g.res.Planets != nil {
		oldOwner := rules.RegionOwnerNone
		if rec := g.regionRecord(g.battleRegion); rec != nil {
			oldOwner = rules.RegionOwnerOf(rec)
		}
		g.res.Planets.ApplyBattlePlanets(g.battlePlanets)
		newOwner := g.res.Planets.RegionOwner(g.battleRegion)
		if g.battleRegion == rules.PhezzanRegion && oldOwner != newOwner {
			// 戰術戰鬥的原版呼叫端只有在人類座位自己拿下時登記；電腦
			// 拿下後，回合尾巴因 +0x12 已更新，也不會補登記。
			g.phezzanBattleResolved = true
			if (g.versus || newOwner == g.faction) &&
				!rules.PhezzanOccurred(g.workRecord()) {
				g.phezzanPending = true
			}
		}
	}
	// 原版一個回合會把**每一個**交戰星域都打完：`0374:03E8` 逐格掃，
	// 打完一格接著掃下一格，全部打完才走回合結束那條鏈
	// （`docs/re/251`）。所以這裡接著掃，掃完才收尾。
	if r := g.contestedRegionFrom(g.battleRegion + 1); r >= 0 {
		g.StartBattle(r)
		return
	}
	// 回合結束那條鏈確實會跑（收據 `s3win`，`docs/re/105`：它記的
	// `2703:0944`／`2703:0990` 就是 `2703:0933` 的內部寫入點，而
	// `2703:0933` 只有新遊戲初始化與 `2703:08DE` 兩個呼叫者）。
	// 那一份 `+0x12` 沒變是因為達希利的行星仍然混雜、重算出同一個值。
	g.endOfTurn()
}

// battleRegionPlanets 回這個星域的行星在戰場上的樣子（`docs/re/99`）。
//
// 位置要掃該星域的地圖（圖塊 `PlanetTile` 的格子依列序對到行星，
// `docs/re/89`），所以缺地圖或缺行星表就回空的——沒有行星就不開火。
func (g *Game) battleRegionPlanets(region int) []rules.BattlePlanet {
	if g.res.Planets == nil || g.res.BattleMaps == nil || g.res.Battle == nil {
		return nil
	}
	if region < 0 || region >= len(g.res.BattleMaps) {
		return nil
	}
	cells := g.res.BattleMaps[region]
	if len(cells) == 0 {
		return nil
	}
	sites := rules.PlanetSites(cells, g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
	return g.res.Planets.BattlePlanetsFor(region, sites)
}

// battleCells 回一個星域的戰場地形，沒有就回 nil。
func (g *Game) battleCells(region int) []uint8 {
	if region < 0 || region >= len(g.res.BattleMaps) {
		return nil
	}
	return g.res.BattleMaps[region]
}

// battleRegionRally 回這個星域埋了哪些集結點（`docs/re/100`）。
func (g *Game) battleRegionRally(region int) []rules.RallySite {
	if g.res.Battle == nil {
		return nil
	}
	cells := g.battleCells(region)
	if len(cells) == 0 {
		return nil
	}
	return rules.RallySites(cells, g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
}

// flagshipOf 回一支艦隊的旗艦分隊（第 0 個）。
func flagshipOf(units []rules.BattleUnit, fleet int) *rules.BattleUnit {
	for i := range units {
		if units[i].Fleet == fleet && units[i].Squadron == 0 {
			return &units[i]
		}
	}
	return nil
}

// battleDecided 回「有一方的艦艇全沒了」。
func battleDecided(units []rules.BattleUnit) bool {
	var left [2]int
	for _, u := range units {
		if u.Faction >= 0 && u.Faction < 2 {
			left[u.Faction] += u.Ships
		}
	}
	return left[0] == 0 || left[1] == 0
}

// SetTactic 換戰術層的 AI（`-ai` 旗標會設它）。
func (g *Game) SetTactic(t ai.Tactic, auto bool) {
	g.tactic, g.autoBattle = t, auto
}

// advanceOrderedFleets 把收到過移動／戰鬥命令的艦隊各推進一個回合。
func (g *Game) advanceOrderedFleets() {
	if g.res.Formation == nil {
		return
	}
	cells := g.battleCells(g.battleRegion)
	passable := func(x, y int) bool {
		return g.res.Terrain == nil || cells == nil || g.res.Terrain.Passable(cells, x, y)
	}
	for i := range g.battleFleets {
		f := &g.battleFleets[i]
		if !f.Ordered {
			continue
		}
		var idx []int
		for k := range g.battleUnits {
			if g.battleUnits[k].Fleet == f.Index {
				idx = append(idx, k)
			}
		}
		if len(idx) == 0 {
			continue
		}
		m := rules.FleetMover{
			X: f.X, Y: f.Y, DestX: f.DestX, DestY: f.DestY,
			Facing: f.Facing, TargetFacing: f.TargetFacing,
			TurnSpeed: f.TurnSpeed, PerTurn: f.PerTurn, ShortMove: f.ShortMove,
		}
		rules.AdvanceFleetTurn(&m, g.res.Formation, g.battleUnits, idx, passable)
		f.X, f.Y, f.Facing = m.X, m.Y, m.Facing
		f.Budget = m.Budget
		if m.Arrived {
			f.Ordered = false
		}
	}
}
