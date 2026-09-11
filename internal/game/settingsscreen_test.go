package game

import (
	"bytes"
	"testing"
)

// snapshot 複製目前這一幀。`Frame()` 回的是同一塊緩衝區，下一次呼叫會
// 就地重畫，所以要比較前後就得先抄一份。
func snapshot(g *Game) []uint8 {
	return append([]uint8(nil), g.Frame().Pix...)
}

// diffBounds 回兩幀所有不同像素的外接矩形與筆數。
func diffBounds(a, b []uint8) (x0, y0, x1, y1, n int) {
	x0, y0 = 640, 480
	x1, y1 = -1, -1
	for i := range a {
		if a[i] == b[i] {
			continue
		}
		n++
		x, y := i%640, i/640
		if x < x0 {
			x0 = x
		}
		if x > x1 {
			x1 = x
		}
		if y < y0 {
			y0 = y
		}
		if y > y1 {
			y1 = y
		}
	}
	return
}

// TestSettingsMusicClicks 設定畫面上那三行點得動，而且點完立刻反映在畫面上
// （熱區來自 `docs/re/285`，三格的語意是 `docs/re/274` §4／§5）。
//
// **不比收據**：原版沒有「點完設定」那一張收據，所以這裡驗的是兩件
// 收據以外也成立的事——值換對了，以及**畫面只在該換的那一組框裡變**。
// 進設定畫面那一張本身由 `TestCommandPhaseNavigation` 對著 `settings1e`
// 逐位元組釘住，這支不重複。
//
// 每次都先把游標移到要點的位置再拍前一張：`Click` 自己會先搬游標
// （`moveBeforePress`），不先搬的話兩張差的是游標，不是那幾個框。
func TestSettingsMusicClicks(t *testing.T) {
	g := reachCommandPhase(t)
	g.Click(544, 125) // 命令列「系統指令」
	g.Click(560, 185) // 選單第 3 項「設定」
	if g.State() != StateSettings {
		t.Fatalf("點完「設定」之後在 %v", g.State())
	}
	// 開局是出貨的那一組：ＦＭ、波琍璐（`BGM_0`）、全艦攻擊（`BGM_6`）。
	if got, want := g.MusicSettings(), (MusicSettings{Device: MusicFM}); got != want {
		t.Fatalf("設定畫面開起來是 %+v，原版出貨是 %+v", got, want)
	}

	for _, c := range []struct {
		name           string
		x, y           int
		want           MusicSettings
		rx0, ry0       int // 允許變動的範圍（該組小框的外框）
		rx1, ry1       int
		expectNoChange bool
	}{
		{name: "英雄交響曲", x: 420, y: 150,
			want: MusicSettings{Device: MusicFM, BGM: 1},
			rx0:  270, ry0: 120, rx1: 433, ry1: 159},
		{name: "決戰前夕", x: 420, y: 270,
			want: MusicSettings{Device: MusicFM, BGM: 1, Battle: 1},
			rx0:  270, ry0: 240, rx1: 433, ry1: 279},
		{name: "關閉（音源）", x: 410, y: 105,
			want: MusicSettings{Device: MusicOff, BGM: 1, Battle: 1},
			rx0:  350, ry0: 100, rx1: 433, ry1: 119},
		// 「ＭＩＤＩ」那一格的可選旗標是 0（出貨的 `MUSIC.INI` 沒有 MIDI
		// 介面），原版的命中測試直接跳過它，畫面上連框都沒有。
		{name: "ＭＩＤＩ（不可選）", x: 310, y: 105,
			want:           MusicSettings{Device: MusicOff, BGM: 1, Battle: 1},
			expectNoChange: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g.MoveMouse(c.x, c.y)
			before := snapshot(g)
			g.Click(c.x, c.y)
			if got := g.MusicSettings(); got != c.want {
				t.Fatalf("點了「%s」之後是 %+v，預期 %+v", c.name, got, c.want)
			}
			after := snapshot(g)
			x0, y0, x1, y1, n := diffBounds(before, after)
			if c.expectNoChange {
				if !bytes.Equal(before, after) {
					t.Fatalf("點不下去的那一格卻改了 %d 個像素，"+
						"範圍 (%d,%d)-(%d,%d)", n, x0, y0, x1, y1)
				}
				return
			}
			if n == 0 {
				t.Fatalf("點了「%s」畫面一個像素都沒變", c.name)
			}
			if x0 < c.rx0 || x1 > c.rx1 || y0 < c.ry0 || y1 > c.ry1 {
				t.Fatalf("點了「%s」變動範圍是 (%d,%d)-(%d,%d)，"+
					"超出那一組框 (%d,%d)-(%d,%d)",
					c.name, x0, y0, x1, y1, c.rx0, c.ry0, c.rx1, c.ry1)
			}
		})
	}

	// 音源關掉之後右鍵回地圖，再進來一次：選的還是剛剛那三格
	// （設定不跟著畫面走）。
	g.RightClick(300, 240)
	g.Click(544, 125)
	g.Click(560, 185)
	if got, want := g.MusicSettings(),
		(MusicSettings{Device: MusicOff, BGM: 1, Battle: 1}); got != want {
		t.Fatalf("重新進設定畫面之後是 %+v，離開時是 %+v", got, want)
	}
}
