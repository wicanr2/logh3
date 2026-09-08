package game

import (
	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// battleSprite 回一個分隊要畫哪一塊 `SPRITE.SPR`（`rules.BattleSprite`）。
func battleSprite(u rules.BattleUnit) int {
	kind := u.Kind
	if kind < 0 {
		kind = -1
	}
	return rules.BattleSprite(u.Faction, u.Detail, kind)
}

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
		ux, uy := rules.FormationBase(rules.BattleFacing(x, y))
		for _, u := range g.slot.Deploy(g.res.Weights, f, x, y, ux, uy) {
			g.battleUnits = append(g.battleUnits, u)
		}
		joined = append(joined, f.Index)
	}
	g.battleFleets = g.slot.BattleFleetsFor(joined)
	g.battlePlanets = g.battleRegionPlanets(region)
	g.battleRally = g.battleRegionRally(region)
	g.battleMenu = false
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
	for i := range g.battleUnits {
		rules.AdvanceBattle(&g.battleUnits[i])
	}
	g.battleLosses = rules.ResolveBattleTurn(g.battlePlanets, g.battleFleets,
		g.battleUnits, g.battleRand)
	for _, l := range g.battleLosses {
		if l.Scorer >= 0 && l.Scorer < 2 {
			g.battleScore[l.Scorer] += l.Score
		}
	}
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
// 主迴圈繼續下一個回合；其餘六項回非 0，跳出主迴圈去處理那個命令。
// 這裡實作了「實行」與「撤退」；取進度／存進度／設定／結束會離開戰場
// 去別的畫面，**那條路還沒逆向**，所以先只把選單關掉。
func (g *Game) battleClick(x, y int) {
	if !g.battleMenu {
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
	case 2: // 撤退
		g.Withdraw()
	}
}

// Withdraw 讓一方退出戰鬥——戰術選單的第三項。
func (g *Game) Withdraw() {
	if g.state != StateBattle {
		return
	}
	g.battleOver = true
	g.EndBattle()
}

// EndBattle 收掉戰場：把剩下的艘數寫回艦艇池，重算戰力，回戰略地圖。
//
// 原版戰後還會做什麼（撤退、星域易主、分數怎麼結算）還沒解，所以這裡
// 只做「把損失寫回去」這一件確定的事（`docs/spec/018`）。
func (g *Game) EndBattle() {
	if g.slot != nil {
		g.slot.ApplyBattleResult(g.battleUnits)
		g.fleets = g.slot.Fleets(g.res.Weights)
	}
	g.state = StateCommandPhase
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
