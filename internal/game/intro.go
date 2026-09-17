package game

import (
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
)

// 開頭動畫（`docs/spec/020`）。
//
// 原版的路徑是 `GIN3.COM` → `OPEN.EXE` → 主選單，動畫那一段在
// `OPEN.EXE` 裡跑：mode 10h（640×350）、來源 374 格（一格四個 plane），
// 原版玩家可見序列是 421 格。
// remake 這一側把它接在主選單前面，素材缺一個就整段跳過。
//
// **播的是原版的分塊順序。** dosgolem 收據證明原版顯示 0–348、略過
// 349，再把 350–373 播三趟；`asset.IntroPlaybackDecoder` 封裝這個
// 玩家可見序列與每趟前的 plane 清除。

// startIntro 解第一格，切到開頭動畫。
func (g *Game) startIntro() error {
	ok, err := g.res.Intro.Next()
	if err != nil {
		return err
	}
	if !ok {
		return nil // 空的動畫：留在主選單
	}
	g.introPage = screen.IntroPage(g.res.Intro)
	g.introFrame = 0
	// 第 0 格已經畫出；下一次等待是切到第 1 格之前的時程。
	g.introLeft = screen.IntroFrameIntervalAt(1)
	g.state = StateIntro
	return nil
}

// SkipIntro 跳過開頭動畫，直接進主選單。左鍵、右鍵、任何按鍵都算。
func (g *Game) SkipIntro() {
	g.introPage = nil
	g.state = StateMainMenu
}

// IntroFrame 回開頭動畫播到第幾格（0 起算）。
func (g *Game) IntroFrame() int { return g.introFrame }

// Tick 讓時間往前走 d。開頭動畫、防拷問答的計時與交戰演出（`engagement.go`）在用；
// 其他畫面都是事件驅動的。
//
// **不讀牆上時鐘**：顯示層把每格的長度傳進來，所以同一串 Tick 一定
// 重現同一串畫面，無頭驗收與視窗跑出來的東西一樣。
func (g *Game) Tick(d time.Duration) {
	if d <= 0 {
		return
	}
	if g.state == StateCopyProtection {
		g.copyProtTickRemainder += d
		period := time.Duration(screen.CopyProtPITDivisor) * time.Second / screen.CopyProtPITHz
		for g.copyProtTickRemainder >= period {
			g.copyProtTickRemainder -= period
			g.copyProtTicks++
		}
	}
	if g.state == StateTurnReport {
		g.turnRep.rem += d
		n := 0
		for tick := g.engagementTickDuration(); g.turnRep.rem >= tick; g.turnRep.rem -= tick {
			n++
		}
		g.turnReportTick(n)
		return
	}
	if g.state == StateTurnWindow {
		g.turnWin.rem += d
		n := 0
		for tick := g.engagementTickDuration(); g.turnWin.rem >= tick; g.turnWin.rem -= tick {
			n++
		}
		g.turnWindowTick(n)
		return
	}
	if g.state == StateEngagement {
		g.engTickRemainder += d
		n := 0
		for tick := g.engagementTickDuration(); g.engTickRemainder >= tick; g.engTickRemainder -= tick {
			n++
		}
		g.engagementTick(n)
		return
	}
	if g.state != StateIntro {
		return
	}
	g.introLeft -= d
	for g.introLeft <= 0 {
		ok, err := g.res.Intro.Next()
		if err != nil || !ok {
			// 播完（或解不下去）就進主選單。
			g.SkipIntro()
			return
		}
		g.introFrame++
		g.introPage = screen.IntroPage(g.res.Intro)
		// IntroTicksToFrame 的量測定義是「切到該格之前」；目前已切到
		// introFrame，所以累加下一格的等待。
		g.introLeft += screen.IntroFrameIntervalAt(g.introFrame + 1)
	}
}

// Palette 回目前畫面用的調色盤。
//
// 開頭動畫是自己的一組灰階（`BU16IBM.RGB`），與遊戲本體那組不同，
// 所以顯示層要每格問一次，不能開機時抓一份用到底。
func (g *Game) Palette() gfx.Palette {
	if g.state == StateIntro {
		return g.res.IntroPal
	}
	return g.res.Menu.Palette()
}
