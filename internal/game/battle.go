package game

import (
	"encoding/binary"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
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
	if err := g.syncFleetTables(); err != nil {
		g.saveErr = err
		return
	}
	g.battleRegion = region
	g.battleTurn = 1
	g.battlePhase = screen.BattlePhaseInput
	g.battleUnits = nil
	g.battleSpotting = battleSpottingState{highlight: -1}
	g.battleDisplayKnown = false
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
	g.initBattleMarkers()
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
			g.battleFleets[i].X, g.battleFleets[i].Y = u.X, u.Y
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
		// 開戰置中到第一支分隊；之後只有按住右鍵會動它（`battleScroll`）。
		g.battleView = screen.BattleView{
			Col: clampCol(u.X/tw-cols/2, cols),
			Row: clampRow(u.Y/th-rows/2, rows),
		}
	}
	g.state = StateBattle
	// 088只取得兩個單人入口的活動控制；雙人戰場邊界仍未驗。
	if !g.versus {
		g.setHumanSeat()
	}
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

// AdvanceBattleTurn經準備及逐幀交錯移動後續接結算；完整八步對齊仍依規格062。
func (g *Game) AdvanceBattleTurn() {
	if g.battlePreparing || g.battleMoving || g.battleSpotting.active {
		return
	}
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
	// +16 bit0與typed Seen共用同一格。先保存衰減／揭露結果，
	// 避免移動寫回舊Attack低byte而重新設回上一輪Seen。
	for _, r := range g.preparationUnits() {
		g.saveSpottingUnit(r)
	}
	if g.tactic != nil {
		b := ai.Battlefield{
			Units: g.battleUnits, Planets: g.battlePlanets,
			Rally: g.battleRally, Terrain: g.res.Terrain,
			Cells: g.battleCells(g.battleRegion), Rand: g.battleRand,
		}
		cpu, passive := g.passiveComputer()
		for _, faction := range []int{0, 1} {
			if faction == g.faction && !g.autoBattle {
				continue // 玩家自己下令
			}
			if passive && faction == cpu {
				continue // 電腦不下令、留在原地（`passivebattle.go`）
			}
			for _, m := range g.tactic.Moves(b, faction, g.battleTurn) {
				if m.Unit >= 0 && m.Unit < len(g.battleUnits) {
					g.applyTacticalMove(m.Unit, m.X, m.Y)
				}
			}
		}
	}
	g.beginBattlePreparation(g.advanceBattleTurnAfterPreparation)
}

// 既有remake自動戰術意圖接到與玩家相同的移動引擎；不新增原版AI行為聲明。
func (g *Game) applyTacticalMove(index, x, y int) {
	u := &g.battleUnits[index]
	u.TargetX, u.TargetY, u.Action = x, y, squadronMove
	if u.Squadron != 0 {
		return
	}
	f := g.battleFleet(u.Fleet)
	if f == nil {
		return
	}
	f.OriginX, f.OriginY, f.X, f.Y = u.X, u.Y, u.X, u.Y
	f.DestX, f.DestY, f.TargetFacing, f.Action, f.Ordered = x, y, f.Facing, 0, true
	f.TurnSpeed, f.PerTurn = g.fleetSpeeds(u.Fleet, false)
	f.ShortMove = rules.Distance16(u.X, u.Y, x, y) < int64(f.PerTurn/3+4*rules.FixOne)
}

func (g *Game) advanceBattleTurnAfterPreparation() {
	g.beginBattleMovement()
}

func (g *Game) advanceBattleTurnAfterMovement() {
	g.beginBattleSpotting(g.advanceBattleTurnAfterSpotting)
}

func (g *Game) advanceBattleTurnAfterSpotting() {
	g.battleLosses = g.resolveBattleTurn()
	for _, l := range g.battleLosses {
		if l.Scorer >= 0 && l.Scorer < 2 {
			g.battleScore[l.Scorer] += l.Score
		}
	}
	// 原版一回合八步，第五步是要塞砲開火（`181B:0385`，規格
	// `docs/spec/028`）、第六步是打行星（`181B:03D0`，規格
	// `docs/spec/019`）。**順序有意義**：瞄好的要塞砲會讓敵方那一回合
	// 打不到這顆行星，所以倒數要排在打行星之後（`docs/re/149`）。
	g.battleGunHits = g.fortressGunFire()
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
	g.battleAssaults = g.planetAssault()
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
	// 打行星那一步有人打就切交戰演出（`20B4:0276` → `1DD6:0152`，規格 056），
	// 看完再回來做這一回合剩下的（`finishBattleTurn`）。
	if g.queueEngagements(g.battleAssaults) {
		return
	}
	g.finishBattleTurn()
}

// finishBattleTurn 是一回合最後的幾步：要塞砲倒數、回合數與結束判定。
//
// 原版的第 ⑧ 步（`181B:03F0` → `185C:0001`，規格 062）排在這裡的最前面：全滅判定、
// 報告、清掉空的組、重建戰場表（子步 8-1）。報告會停下來等，所以走 `battleFinalPhase`
// 的續接寫法——停住時這一回合剩下的幾步由報告看完之後接回來。
func (g *Game) finishBattleTurn() {
	if g.battleFinalPhase(0, g.endBattleTurn) {
		return
	}
	g.endBattleTurn()
}

func (g *Game) endBattleTurn() {
	g.battlePhase = screen.BattlePhaseInput
	rules.FortressGunTickAll(g.battlePlanets)
	g.battleSpotting = battleSpottingState{highlight: -1}
	g.refreshBattleMarkers()
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
	if g.battleSpotting.active {
		if g.battleSpotting.stage == 2 || g.battleSpotting.stage == 5 {
			g.battleSpotting.ticks = 0
		}
		return
	}
	if g.battlePreparing || g.battleMoving {
		return
	}
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
// 每場戰術得分在這裡累加進整局分數，供最終勝負報告顯示
// （`ds:8FEC`／`ds:8FF0`，`docs/re/309`）。戰後撤退仍未解。
func (g *Game) EndBattle() {
	for f := 0; f < 2; f++ {
		g.campaignScore[f] += g.battleScore[f]
	}
	if g.slot != nil {
		if err := g.syncFleetTables(); err != nil {
			g.saveErr = err
			return
		}
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
	// 原版的戰後艦隊解散／能力重算會先經 `090A:0E6D` 重算全圖未安定度，
	// 然後 `0374:03E8` 才從下一個星域續掃。後面的叛亂效果必須讀這份戰後值，
	// 不能在第一場戰鬥前就批次套完（`docs/spec/042`、`docs/re/310`）。
	g.recomputeUnrest(g.turn)
	g.scanRegionsFrom(g.battleRegion + 1)
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
	sites := g.loadRegionPlanetSites(region)
	return g.res.Planets.BattlePlanetsFor(region, sites)
}

// loadRegionPlanetSites 是 2AF5:00CB 的載圖副作用（研究 354／規格 066）。
// 只在實際載入事件呼叫，不能在開局預載或 Draw 時提前改整張行星表。
func (g *Game) loadRegionPlanetSites(region int) []rules.PlanetSite {
	if g.res.Planets == nil || g.res.Battle == nil {
		return nil
	}
	sites := rules.PlanetSites(g.battleCells(region), g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
	start, n := g.res.Planets.Range(region)
	for i := 0; i < n && i < len(sites); i++ {
		rec := g.res.Planets.Record(start + i)
		if len(rec) >= 0x4A {
			binary.LittleEndian.PutUint16(rec[0x46:], uint16(sites[i].X))
			binary.LittleEndian.PutUint16(rec[0x48:], uint16(sites[i].Y))
		}
	}
	return sites
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

// battleScroll 是戰場視野的滑鼠捲動（`1588:043F` → `1588:03A9`，`docs/re/326`）：
// 戰場視窗上沒開任何東西時按住右鍵，每一圈重讀游標、往它相對於滑鼠座標
// (224,180) 的方位走一格，放開才停；游標停在邊緣不按什麼都不動（收據
// `cam-edge-*`）。remake 沒有「一圈多久」：按下那一刻走一格（原版至少走一格
// 才看放開），之後顯示層每一幀按著就再走一格（`RightHold`）——一次按住走幾格
// 是已知差異（規格 023）。
func (g *Game) battleScroll(x, y int) {
	s := g.res.Battle
	if s == nil || x < s.Field.X || x >= s.Field.X+s.Field.Cols*s.Field.TileW ||
		y < s.Field.Y || y >= s.Field.Y+s.Field.Rows*s.Field.TileH {
		return
	}
	dir := rules.ViewScrollDir(x, y-ui.FrameCellOriginY)
	col, row, moved := rules.ViewScrollStep(g.battleView.Col, g.battleView.Row, dir)
	if moved {
		g.battleView = screen.BattleView{Col: col, Row: row}
	}
}

// RightHold 是右鍵按著沒放的每一幀：戰場上就再捲一格。
func (g *Game) RightHold(x, y int) {
	if g.state != StateBattle || !g.battleScrollHeld || g.battleMoving || g.battlePreparing || g.battleSpotting.active {
		return
	}
	g.moveMouse(x, y)
	g.battleScroll(x, y)
}

// RightRelease 是右鍵放開。
func (g *Game) RightRelease() { g.battleScrollHeld = false }
