package game

import (
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
)

// 開頭動畫（`docs/spec/020`）。
//
// 原版的路徑是 `GIN3.COM` → `OPEN.EXE` → 主選單，動畫那一段在
// `OPEN.EXE` 裡跑：mode 10h（640×350）、374 格、一格四個 plane。
// remake 這一側把它接在主選單前面，素材缺一個就整段跳過。
//
// **播的是檔案順序。** 原版實際播出來的順序被讀檔緩衝的副作用打亂了
// （第 1–116 格從來沒被解過，`docs/re/108`），成因未解；那是原版的
// 瑕疵不是動畫的內容，所以 remake 不照抄。要不要重現等成因解出來再談。

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
	g.introLeft = screen.IntroFrameInterval
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

// Tick 讓時間往前走 d。目前只有開頭動畫在用；其他畫面都是事件驅動的。
//
// **不讀牆上時鐘**：顯示層把每格的長度傳進來，所以同一串 Tick 一定
// 重現同一串畫面，無頭驗收與視窗跑出來的東西一樣。
func (g *Game) Tick(d time.Duration) {
	if g.state != StateIntro || d <= 0 {
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
		g.introLeft += screen.IntroFrameInterval
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
