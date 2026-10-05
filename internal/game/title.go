package game

import (
	"time"

	"github.com/wicanr2/logh3/internal/screen"
)

// 112／研究431：兩段末格等待1150、2000與標題等待9000，倒數<=3返回。
// 秒數是hardware-spec approximation，不模擬解碼／重畫成本或PIT相位。
const (
	IntroTitleDelay   = time.Duration(1150-3+2000-3) * screen.IntroPITDivisor * time.Second / screen.IntroPITHz
	TitleWaitDuration = time.Duration(9000-3) * screen.IntroPITDivisor * time.Second / screen.IntroPITHz
)

func (g *Game) introNextInterval() time.Duration {
	if g.res.Title != nil && g.introFrame == g.res.Intro.Frames()-1 {
		return IntroTitleDelay
	}
	return screen.IntroFrameIntervalAt(g.introFrame + 1)
}

func (g *Game) replayIntro() bool {
	if g.res.IntroRestart == nil {
		return false
	}
	dec, err := g.res.IntroRestart()
	if err != nil || dec == nil {
		return false
	}
	g.res.Intro = dec
	return g.startIntro() == nil && g.state == StateIntro
}
