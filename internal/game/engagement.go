package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 交戰演出——打行星那一種（`docs/re/324`、規格 056）。
//
// 原版戰場回合處理的打行星那一步（`20B4:0222` 起）對每一顆有人打的行星：
// `1DD6:0152` 開場（`ANI_PNL1.A98` 底圖、三個樣式 0 的文字視窗、換戰鬥曲）→
// `20B4:0486` 打（同時放方塊動畫、印「擊毀」）→ `1DD6:026B` 收場（換回平時曲）。
// 演出約 29M 步，按鍵只把動畫跳到結果、不縮短；動畫效果（`ds:8F36`）關掉時
// 沒有方塊動畫、演出短九百萬步，面板與換曲照舊。
//
// remake 接的是「兩幀」：開場那一幀（反擊之前的數字）與結果那一幀（「擊毀」與
// 反擊之後的艘數），左鍵或右鍵換幀，看完回戰場繼續那一回合剩下的處理。
// **行星圖、艦隊圖、量表與方塊動畫沒畫**：`COMANI.BMG` 的編碼還沒解
// （`docs/internals/20`），事先列在版面的 `known_diff`。

// engagement 是一場演出要印的數。
type engagement struct {
	planet, fleet int
	// defence／ships 是演出開始那一刻（反擊之前）的值。
	defence, defenceMax int
	ships, shipsMax     int
	destroyed           int
	label               screen.EngagementLabel
}

// queueEngagements 把這一回合打行星的結果排成演出（一顆行星一場：原版
// `20B4:0222` 逐顆行星開場；同一顆被幾支打就在同一場裡逐支印——這裡先
// 一支一場，假設待驗）。回 true ＝ 有得演。
func (g *Game) queueEngagements(hits []rules.PlanetAssaultHit) bool {
	if g.res.Engagement == nil || g.res.EngagementPanel == nil {
		return false
	}
	g.engQueue = g.engQueue[:0]
	for _, h := range hits {
		if h.Planet < 0 || h.Planet >= len(g.battlePlanets) {
			continue
		}
		p := g.battlePlanets[h.Planet]
		after := g.battleFleetShips(h.Fleet)
		e := engagement{
			planet: h.Planet, fleet: h.Fleet,
			defence: p.Defence + h.Damage, defenceMax: p.DefenceMax,
			ships: after + h.Counter, shipsMax: g.battleFleetShipsMax(h.Fleet),
			destroyed: h.Counter,
		}
		// 中間那格的標題：原版四種字串（防衛攻擊／登陸艦攻擊／占領成功／占領失敗）
		// 各在什麼情形印沒讀完；收據那一場是「只挨反擊」→ 防衛攻擊。其餘假設待驗。
		switch {
		case h.Captured:
			e.label = screen.EngagementCaptured
		case h.Damage > 0:
			e.label = screen.EngagementLanding
		default:
			e.label = screen.EngagementDefence
		}
		g.engQueue = append(g.engQueue, e)
	}
	if len(g.engQueue) == 0 {
		return false
	}
	g.engPhase = 0
	g.state = StateEngagement
	return true
}

// battleFleetShips 回這一支現在的艘數合計（船艦總數的分子）。
func (g *Game) battleFleetShips(fleet int) int {
	n := 0
	for _, u := range g.battleUnits {
		if u.Fleet == fleet {
			n += u.Ships
		}
	}
	return n
}

// battleFleetShipsMax 回這一支的滿編艘數合計（船艦總數的分母；假設待驗，
// 收據那一支剛好滿編 22000/22000）。
func (g *Game) battleFleetShipsMax(fleet int) int {
	n := 0
	for _, u := range g.battleUnits {
		if u.Fleet == fleet {
			n += u.FullShips
		}
	}
	return n
}

// engagementView 把目前這一場排成畫面要印的東西。
func (g *Game) engagementView() *screen.EngagementView {
	if len(g.engQueue) == 0 {
		return nil
	}
	e := g.engQueue[0]
	if e.planet < 0 || e.planet >= len(g.battlePlanets) {
		return nil
	}
	p := g.battlePlanets[e.planet]
	v := &screen.EngagementView{
		Faction: p.Owner, PlanetName: g.battlePlanetName(e.planet),
		Defence: e.defence, DefenceMax: e.defenceMax,
		Ships: e.ships, ShipsMax: e.shipsMax,
		Label: e.label, Destroyed: e.destroyed, ShowResult: g.engPhase > 0,
	}
	if g.engPhase > 0 {
		v.Ships = e.ships - e.destroyed
	}
	// 艦隊那格：自軍或索敵等級 3 的才印編號與提督（`ds:4801` 那句「第？艦隊
	// ？？？？？？？？？」是給看不清的），門檻照右下角敵艦隊分頁的規則
	// （`docs/re/319`；假設待驗）。
	for _, f := range g.battleFleets {
		if f.Index != e.fleet {
			continue
		}
		known := f.Faction == g.faction
		if !known {
			for _, u := range g.battleUnits {
				if u.Fleet == e.fleet && u.Detail >= 3 {
					known = true
					break
				}
			}
		}
		if known {
			v.FleetKnown, v.FleetNumber = true, f.Number
			if i := g.fleetAdmiral(e.fleet); i >= 0 && i < len(g.res.AdmiralNames) {
				v.AdmiralName = codesBig5(g.res.AdmiralNames[i])
			}
		}
	}
	return v
}

// fleetAdmiral 回戰略層那一支艦隊的提督索引；找不到回 −1。
func (g *Game) fleetAdmiral(index int) int {
	for _, f := range g.fleets {
		if f.Index == index {
			return f.Admiral
		}
	}
	return -1
}

// battlePlanetName 回戰場上第 i 顆行星的名字（Big5）。
func (g *Game) battlePlanetName(i int) []byte {
	if g.res.Planets == nil || g.battleRegion < 0 {
		return nil
	}
	start, n := g.res.Planets.Range(g.battleRegion)
	if i < 0 || i >= n {
		return nil
	}
	return planetNameBytes(g.res.Planets.Record(start + i))
}

// engagementClick 是演出期間的左鍵／右鍵：開場那一幀 → 結果那一幀 → 下一場
// 或回戰場（原版按鍵只把動畫跳到結果，演出的長度不變；remake 沒有計時）。
func (g *Game) engagementClick() {
	if g.engPhase == 0 {
		g.engPhase = 1
		return
	}
	if len(g.engQueue) > 0 {
		g.engQueue = g.engQueue[1:]
	}
	g.engPhase = 0
	if len(g.engQueue) > 0 {
		return
	}
	g.state = StateBattle
	g.finishBattleTurn()
}

// drawEngagement 畫演出：底圖與三個視窗（沒有游標）。
func (g *Game) drawEngagement() *gfx.Framebuffer {
	fb := gfx.NewFramebuffer(0)
	if g.res.Engagement == nil {
		return fb
	}
	g.res.Engagement.Draw(fb, g.res.Font, g.res.FrameTiles, g.res.EngagementPanel, g.engagementView())
	return fb
}
