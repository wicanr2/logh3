package game

import (
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 回合處理中的過場視窗（規格 059、`docs/re/332`）。
//
// 原版的回合是一支同步程式，每支人類座位的艦隊走一格時在 `3765:006C(n)` 停下來：
// 出發前往先等 `ds:8F38 >> 2`、塗完地圖再等 `ds:8F38`；被擋與經費不足只等一段。
// 計時到或滑鼠左／右鍵按一下結束一段。**等完才套結果**（`113A:047D` 先叫
// `0B46` 才從 `+0585` 起扣錢、改位置、到達收尾），所以駐留的解散在
// 「已經駐留在星系上。」那一張之後。

// turnWindowState 是停在哪一支、印哪一張、等到哪裡。
type turnWindowState struct {
	fleet  int
	view   screen.TurnWindowView
	face   int
	from   int // 出發前往：現在格（第一段塗色）
	next   int // 下一格（第二段塗色）
	stage  int // 0 ＝ 第一段；出發前往才有 1
	ticks  int
	rem    time.Duration
	resume State
	// apply 在最後一段結束時套這一支的結果，回 true ＝ 這一支從艦隊表拿掉了。
	apply func() bool
}

// turnProcessing 回現在是不是「實行」之後、回合處理中停下來的畫面（過場視窗或問話）：
// 右側標題只剩「戰略模式處理中」、清單清空、地圖不看座位（規格 059）。
func (g *Game) turnProcessing() bool {
	return g.state == StateTurnWindow || g.state == StateOccupyAsk || g.state == StateTurnReport
}

// humanSeat 回這個陣營的艦隊走的時候要不要印（`113A:060F`：`3982:022D(陣營) == 0`，
// 單人模式只有玩家那一方、對戰模式兩方都是）。`ds:8F1C` bit0 的除錯「顯示」沒有模型化。
func (g *Game) humanSeat(faction int) bool {
	if faction < 0 || faction > 1 {
		return false
	}
	return g.versus || faction == g.faction
}

func (g *Game) regionOwner(r int) int {
	if rec := g.regionRecord(r); rec != nil {
		return int(rec[regionOwnerOff])
	}
	return rules.RegionOwnerNone
}

// openTurnWindow 停在第 i 支：印第一段、存好結束之後要套的結果。from 是走之前的
// 所在星域（出發前往那一張印的現在格；呼叫時艦隊可能已經改了位置）。
func (g *Game) openTurnWindow(i int, kind screen.TurnWindowKind, from int, plan rules.StepPlan, apply func() bool) {
	f := &g.fleets[i]
	v := screen.TurnWindowView{
		Kind:   kind,
		Number: f.Number,
	}
	if g.res.FleetWindow != nil && f.Faction >= 0 && f.Faction < len(g.res.FleetWindow.FactionNamesColoured) {
		v.FactionName = decodeHexBytes(g.res.FleetWindow.FactionNamesColoured[f.Faction])
	}
	adm := g.admiralRec(f.Admiral)
	v.AdmiralName = codesBig5(g.admiralNameCodes(f.Admiral))
	v.Rank = codesBig5(g.rankCodes(rules.RankOf(byteAt(adm, admiralMeritOff))))
	w := turnWindowState{fleet: i, face: f.Admiral, from: -1, next: -1, resume: g.state, apply: apply}
	if kind == screen.TurnWindowDepart {
		w.from, w.next = from, plan.Next
		// 地名的顏色讀星域記錄的擁有者 `+0x12`（`3982:0263`）。
		v.FromName, v.FromOwner = codesBig5(g.regionName(from)), g.regionOwner(from)
		v.NextName, v.NextOwner = codesBig5(g.regionName(plan.Next)), g.regionOwner(plan.Next)
		if plan.Next == f.Target {
			v.Arrival = screen.TurnWindowArrived
			if f.Move == rules.MoveGarrison {
				v.Arrival = screen.TurnWindowGarrisoned
			}
		}
	}
	w.view = v
	g.turnWin = w
	g.state = StateTurnWindow
}

// turnWindowWait 回目前這一段要等幾個 tick。
func (g *Game) turnWindowWait() int {
	speed := g.display.MsgSpeed
	if g.turnWin.view.Kind == screen.TurnWindowDepart && g.turnWin.stage == 0 && g.res.TurnWindow != nil {
		return g.res.TurnWindow.FirstWait(speed)
	}
	return speed
}

// turnWindowTick 讓等待往前走 n 個 tick。
func (g *Game) turnWindowTick(n int) {
	for ; n > 0 && g.state == StateTurnWindow; n-- {
		g.turnWin.ticks++
		if g.turnWin.ticks >= g.turnWindowWait() {
			g.endTurnWindowStage()
		}
	}
}

// endTurnWindowStage 結束一段（計時到或一次左／右鍵按下）。出發前往的第一段結束
// 換第二段；最後一段結束就套結果、接下一支。
func (g *Game) endTurnWindowStage() {
	w := &g.turnWin
	w.ticks, w.rem = 0, 0
	if w.view.Kind == screen.TurnWindowDepart && w.stage == 0 {
		w.stage = 1
		w.view.Second = true
		return
	}
	g.keepMapMark()
	g.state = w.resume
	i := w.fleet
	if w.apply != nil && w.apply() {
		g.moveFleetsFrom(i) // 這一支解散了，下一支補進同一個索引
		return
	}
	// `047D` 回到 `0350` 之後的士氣檢查（`+03C5`）。
	if g.startTurnReports(g.fleetStepMorale(i), func() { g.moveFleetsFrom(i + 1) }) {
		return
	}
	g.moveFleetsFrom(i + 1)
}

// turnMapMark 回回合處理中地圖上塗色 5 的那一格（`3705:03FD(星域, 5)`），沒有回 −1。
//
// 塗了就留著，要等下一次 `3705:0349` 整片還原（`docs/re/332` §二、`docs/re/335` §五）：
//   - 過場視窗：第一段現在格、第二段下一格；視窗關掉之後下一格留著，到 32 支走完（`056A:044B`）才還原。
//   - 占據（`056A:059E`）、叛亂（`06F3`）、易主（`089D`）報告：先整片還原，再塗報告那一格，印完不還原。
//   - 回合結束回到命令階段時整片還原（`04F8` 的 `3705:0349(1)`）。
func (g *Game) turnMapMark() int {
	switch g.state {
	case StateTurnWindow:
		if g.turnWin.view.Kind == screen.TurnWindowDepart {
			if g.turnWin.stage == 0 {
				return g.turnWin.from
			}
			return g.turnWin.next
		}
	case StateOccupationReport:
		if r, ok := g.CurrentOccupationReport(); ok {
			return r.Region
		}
	case StateRebellionReport:
		if r, ok := g.CurrentRebellionReport(); ok {
			return r.Region
		}
	case StateRegionFlipReport:
		if f, ok := g.CurrentFlip(); ok {
			return f.Region
		}
	}
	return g.mapMark
}

// mapMarkColour 是 `3705:03FD` 塗的色號。
func (g *Game) mapMarkColour() uint8 {
	if g.res.TurnWindow == nil {
		return 5
	}
	return uint8(g.res.TurnWindow.DepartColour)
}

// keepMapMark 在報告或過場關掉的那一刻把當下塗的那一格留下來。
func (g *Game) keepMapMark() { g.mapMark = g.turnMapMark() }

// mapProcessing 回地圖是不是回合處理中的樣子：沒有反白格（`ds:F182 = FF`）、不看座位（`ds:8F1D = FF`）。
func (g *Game) mapProcessing() bool {
	switch g.state {
	case StateOccupationReport, StateRebellionReport, StateRegionFlipReport:
		return true
	}
	return g.turnProcessing()
}

// drawTurnWindow 把過場視窗疊在戰略地圖介面上。
func (g *Game) drawTurnWindow(fb *gfx.Framebuffer) {
	if g.res.TurnWindow == nil {
		return
	}
	var face *asset.Image
	if g.res.Faces != nil {
		if img, err := g.res.Faces.Decode(g.turnWin.face); err == nil {
			face = img
		}
	}
	v := g.turnWin.view
	g.res.TurnWindow.Draw(fb, g.res.Font, g.res.FrameTiles, &v, face)
}
