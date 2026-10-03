package game

import (
	"time"

	"github.com/wicanr2/logh3/internal/screen"
)

// 規格089：清單速度是使用者選定的刻意差異；經濟箭頭沿八倒數近似。
const (
	leftListDelay  = 350 * time.Millisecond
	leftListPeriod = 100 * time.Millisecond
	leftPlanPeriod = time.Duration(8 * 10 * 2048 * int64(time.Second) / 1193182)
)

type leftRepeatKind uint8

const (
	leftRepeatNone leftRepeatKind = iota
	leftRepeatList
	leftRepeatTabs
	leftRepeatStandby
	leftRepeatAdjutant
	leftRepeatStaff
	leftRepeatBattleData
	leftRepeatTax
	leftRepeatPercent
)

type leftRepeatAction struct {
	kind     leftRepeatKind
	state    State
	dir, tab int
	region   int
	x, y     int
}

type leftRepeatHold struct {
	action leftRepeatAction
	next   time.Duration
}

func (a leftRepeatAction) economic() bool {
	return a.kind == leftRepeatTax || a.kind == leftRepeatPercent
}

func (a leftRepeatAction) sameAction(b leftRepeatAction) bool {
	a.x, a.y, b.x, b.y = 0, 0, 0, 0
	return a == b
}

// leftRepeatAt只辨識可以重複的箭頭；不把確認、選擇與載入當保持動作。
func (g *Game) leftRepeatAt(x, y int) leftRepeatAction {
	a := leftRepeatAction{state: g.state, x: x, y: y}
	if g.msgOn || g.staff.detailOn || g.adj.detailOn || g.adj.menuOn {
		return a
	}
	if g.staff.listOn {
		a.kind, a.dir = leftRepeatStaff, g.adjListScrollAt(x, y)
	} else if g.adj.listOn {
		a.kind, a.dir = leftRepeatAdjutant, g.adjListScrollAt(x, y)
	} else {
		switch g.state {
		case StateStarList, StateFleetList, StateAdmiralListView, StateEnemyList:
			if g.res.ListScreen != nil && g.lst.detail < 0 && !g.lst.bio {
				a.kind, a.dir, a.tab = leftRepeatList, g.res.ListScreen.ArrowAt(x, y), g.listScreenTab()
			}
		case StateCommandPhase:
			if g.res.FleetList != nil && !g.listOverlayOn() {
				a.tab, a.region = g.listTab(), g.listRegion()
				if a.tab != listTabPlanet {
					a.kind, a.dir = leftRepeatTabs, g.res.FleetList.ScrollAt(x, y)
				}
			}
		case StateStandbyAdmirals, StateAdmiralList:
			a.kind, a.dir = leftRepeatStandby, g.standbyScrollAt(x, y)
		case StateBattle:
			if g.battleData && !g.battleDetail && g.res.BattleData != nil {
				if len(g.battleDataRows(g.battleDataTab)) > g.res.BattleData.Row.Count {
					a.kind, a.dir, a.tab = leftRepeatBattleData, g.res.BattleData.BattleDataScrollAt(x, y), g.battleDataTab
				}
			}
		case StateTaxPlan, StateBudgetPlan:
			switch {
			case planTaxUp.hit(x, y):
				a.kind, a.dir = leftRepeatTax, 1
			case planTaxDown.hit(x, y):
				a.kind, a.dir = leftRepeatTax, -1
			case planPctUp.hit(x, y):
				a.kind, a.dir = leftRepeatPercent, 1
			case planPctDown.hit(x, y):
				a.kind, a.dir = leftRepeatPercent, -1
			}
		}
	}
	if a.dir == 0 {
		a.kind = leftRepeatNone
	}
	return a
}

// LeftRelease停止左鍵保持。保持狀態只在記憶體，不進原版存檔。
func (g *Game) LeftRelease() { g.leftHeld = leftRepeatHold{} }

// LeftHold由顯示層每幀送入游標與經過時間；先前Click捕獲的箭頭才會更新。
// 清單移出仍捲動；經濟箭頭只在原熱區內更新，等待期間不累積補發。
func (g *Game) LeftHold(x, y int, elapsed time.Duration) {
	g.MoveMouse(x, y)
	a := g.leftHeld.action
	if a.kind == leftRepeatNone || elapsed <= 0 {
		return
	}
	if g.leftRepeatAt(a.x, a.y) != a {
		g.LeftRelease()
		return
	}
	period := leftListPeriod
	if a.economic() {
		period = leftPlanPeriod
	}
	g.leftHeld.next -= elapsed
	for g.leftHeld.next <= 0 {
		g.leftHeld.next += period
		if !a.economic() || a.sameAction(g.leftRepeatAt(x, y)) {
			g.applyLeftRepeat(a)
		}
		if g.leftRepeatAt(a.x, a.y) != a {
			g.LeftRelease()
			return
		}
	}
}

func (g *Game) applyLeftRepeat(a leftRepeatAction) {
	switch a.kind {
	case leftRepeatList:
		g.listScreenStep(a.dir)
	case leftRepeatTabs:
		g.listScroll(a.dir)
	case leftRepeatStandby:
		g.standbyStep(a.dir)
	case leftRepeatAdjutant:
		g.adjListScroll(a.dir)
	case leftRepeatStaff:
		g.staff.first = clampListOffset(g.staff.first+a.dir, len(g.staff.list), g.adjListPage())
	case leftRepeatBattleData:
		g.battleDataOffset = clampListOffset(g.battleDataOffset+a.dir, len(g.battleDataRows(a.tab)), g.res.BattleData.Row.Count)
	case leftRepeatTax:
		g.stepTaxLevel(a.dir)
	case leftRepeatPercent:
		g.stepShipPct(a.dir)
	}
}

func clampListOffset(offset, count, rows int) int {
	last := count - rows
	if last < 0 {
		last = 0
	}
	if offset < 0 {
		return 0
	}
	if offset > last {
		return last
	}
	return offset
}

func (g *Game) standbySpec() *screen.MainGameSpec {
	if g.state == StateAdmiralList {
		return g.res.AdmiralList
	}
	return g.commandSpec(StateStandbyAdmirals)
}

func (g *Game) standbyScrollAt(x, y int) int {
	return admiralScrollAt(g.standbySpec(), x, y)
}

func (g *Game) standbyStep(dir int) {
	g.setListRows(g.standbySpec())
	g.standbyScroll = clampListOffset(g.standbyScroll+dir, g.listScreenAdmiralCount, g.adjListPage())
}
