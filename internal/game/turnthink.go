package game

import (
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
)

// 「戰略模式思考中」（`04F8:030E` → `04F8:0355(1)`，規格 059、`docs/re/332` §一）。
//
// 原版按「實行」之後兩個座位各印一次標題窗「<陣營> 宇宙曆…／戰略模式思考中」，
// 電腦座位那一次接著跑戰略 AI（`2D10:0005`），畫面停到 AI 算完才進「戰略模式處理中」。
// 停多久是 AI 在機器上的運算時間，不是計時器；remake 的 AI 瞬間算完，照使用者裁決
// （2026-09-17）固定停 `think_title.ticks`（dosgolem 在 `fcmd-turn-think` 那條量到的長度）。
// 人類座位那一次不到半個 tick 就被蓋掉，不重現；對戰模式兩邊都是人，不停。

// turnThinkState 是停了幾個 tick。
type turnThinkState struct {
	ticks int
	rem   time.Duration
}

// startTurnThinking 進「思考中」；沒有版面或對戰模式回 false，呼叫端直接跑回合。
func (g *Game) startTurnThinking() bool {
	if g.versus || g.res.TurnWindow == nil || g.res.TurnWindow.ThinkTitle.Ticks <= 0 {
		return false
	}
	g.setActiveSeat(1)
	g.turnThink = turnThinkState{}
	g.state = StateTurnThinking
	return true
}

// turnThinkTick 走 n 個計時器 tick；停滿就接回合處理。
func (g *Game) turnThinkTick(n int) {
	for ; n > 0 && g.state == StateTurnThinking; n-- {
		g.turnThink.ticks++
		if g.turnThink.ticks >= g.res.TurnWindow.ThinkTitle.Ticks {
			// 回合處理照舊從系統指令選單那一刻接（`execute` 原本就在那個狀態裡跑）。
			g.state = StateSysMenu
			g.advanceTurn()
		}
	}
}

// drawThinkTitle 畫標題窗：列 0 電腦座位的陣營與日期，列 1「戰略模式思考中」。
func (g *Game) drawThinkTitle(fb *gfx.Framebuffer) {
	g.res.TurnWindow.DrawThinkTitle(fb, g.res.Font, g.res.FrameTiles)
	if c, _, ok := g.capitalSummary(); ok {
		g.res.CapitalPanel.DrawFactionLine(fb, g.res.Font, c, 1-g.faction)
	}
}
