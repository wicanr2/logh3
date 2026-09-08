package game

import (
	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// battleSprite 回一個分隊要畫哪一塊 `SPRITE.SPR`。
//
// 用 `-poke` 把同一支艦隊七個分隊的艦種改成 0–6，再看畫面上各是哪一塊
// （收據 `s3kinds2`，`docs/re/85`）：艦種 0、1、2 依序是第 66、67、68 塊，
// 艦種 3 到 6 全部畫成第 74 塊。為什麼從 3 開始就共用一塊還沒解。
func battleSprite(kind int) int {
	if kind >= 0 && kind <= 2 {
		return battleSpriteBase + kind
	}
	return battleSpriteOther
}

// SPRITE.SPR 裡單位圖示的位置（`docs/re/85`）。
const (
	battleSpriteBase  = 66
	battleSpriteOther = 74
)

// shipKind 回一個分隊的艦種（艦艇記錄的 `+0x04`）。
func (g *Game) shipKind(u rules.BattleUnit) int {
	ships := g.slot.FleetShips(u.Fleet)
	if u.Squadron < 0 || u.Squadron >= len(ships) {
		return -1
	}
	n := ships[u.Squadron]
	if n < 0 || (n+1)*rules.ShipRecord > len(g.slot.Ship) {
		return -1
	}
	return int(g.slot.Ship[n*rules.ShipRecord+4])
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
	if g.tactic != nil {
		b := ai.Battlefield{Units: g.battleUnits}
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
	g.battleTurn++
}

// SetTactic 換戰術層的 AI（`-ai` 旗標會設它）。
func (g *Game) SetTactic(t ai.Tactic, auto bool) {
	g.tactic, g.autoBattle = t, auto
}
