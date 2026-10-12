package game

import (
	"bytes"
	"github.com/wicanr2/logh3/internal/rules"
	"reflect"
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

// TestSettingsDisplayClicks 設定畫面另外三行（訊息速度、動畫效果、動畫速度）
// 點得動，而且畫面只在該變的那一塊變。
//
// 值的去向與邊界都是反組譯釘的（`docs/re/286` §二）：
// `12`／`13` 把 `ds:8F38` ±10（到 300 不再加、不到 10 不再減）、
// `14`／`15` 把 `ds:8F36` 設成 1／0、`16`–`18` 把 `ds:8F37` 設成 0／1／2。
func TestSettingsDisplayClicks(t *testing.T) {
	g := reachCommandPhase(t)
	g.Click(544, 125)
	g.Click(560, 185)
	if g.State() != StateSettings {
		t.Fatalf("點完「設定」之後在 %v", g.State())
	}
	// 開機預設：動畫開啟、普通、90（原版三份寫入點與出貨存檔一致）。
	if got, want := g.DisplaySettings(),
		(DisplaySettings{Anim: 1, MsgSpeed: 90}); got != want {
		t.Fatalf("設定畫面開起來是 %+v，原版出貨是 %+v", got, want)
	}

	for _, c := range []struct {
		name     string
		x, y     int
		want     DisplaySettings
		rx0, ry0 int // 允許變動的範圍
		rx1, ry1 int
	}{
		{name: "訊息速度 ＋", x: 405, y: 310,
			want: DisplaySettings{Anim: 1, MsgSpeed: 100},
			rx0:  351, ry0: 300, rx1: 384, ry1: 319},
		{name: "動畫效果 關閉", x: 410, y: 330,
			want: DisplaySettings{Anim: 0, MsgSpeed: 100},
			rx0:  350, ry0: 320, rx1: 433, ry1: 339},
		{name: "動畫速度 最快", x: 410, y: 350,
			want: DisplaySettings{Anim: 0, AnimSpeed: 2, MsgSpeed: 100},
			rx0:  302, ry0: 340, rx1: 433, ry1: 359},
	} {
		t.Run(c.name, func(t *testing.T) {
			g.MoveMouse(c.x, c.y)
			before := snapshot(g)
			g.Click(c.x, c.y)
			if got := g.DisplaySettings(); got != c.want {
				t.Fatalf("點了「%s」之後是 %+v，預期 %+v", c.name, got, c.want)
			}
			x0, y0, x1, y1, n := diffBounds(before, snapshot(g))
			if n == 0 {
				t.Fatalf("點了「%s」畫面一個像素都沒變", c.name)
			}
			if x0 < c.rx0 || x1 > c.rx1 || y0 < c.ry0 || y1 > c.ry1 {
				t.Fatalf("點了「%s」變動範圍是 (%d,%d)-(%d,%d)，"+
					"超出 (%d,%d)-(%d,%d)",
					c.name, x0, y0, x1, y1, c.rx0, c.ry0, c.rx1, c.ry1)
			}
		})
	}

	// 上限 300：從 100 按 20 下只到 300，再按不動。
	for i := 0; i < 21; i++ {
		g.Click(405, 310)
	}
	if got := g.DisplaySettings().MsgSpeed; got != 300 {
		t.Fatalf("一直按「＋」之後訊息速度是 %d，原版停在 300", got)
	}
	// 下限：原版的判斷是「不到 10 就不減」，所以按到底是 0。
	for i := 0; i < 40; i++ {
		g.Click(421, 310)
	}
	if got := g.DisplaySettings().MsgSpeed; got != 0 {
		t.Fatalf("一直按「−」之後訊息速度是 %d，原版停在 0", got)
	}
}

// TestSettingsCloseIsRightClick 「結束」與右鍵走同一條出口。
//
// 原版的「結束」只是讓熱區分派回傳 1，設定畫面那個訊息迴圈跟著結束、
// `retf` 回呼叫端；右鍵（輸入結果 2）收在**同一個** `retf` 上
// （`docs/re/286` §三）。所以兩條路離開之後應該落在同一張畫面。
func TestSettingsCloseIsRightClick(t *testing.T) {
	open := func(t *testing.T) *Game {
		t.Helper()
		g := reachCommandPhase(t)
		g.Click(544, 125)
		g.Click(560, 185)
		if g.State() != StateSettings {
			t.Fatalf("點完「設定」之後在 %v", g.State())
		}
		return g
	}
	byRight := open(t)
	byRight.RightClick(410, 370)

	byClose := open(t)
	byClose.Click(410, 370) // 「結束」那一格
	if byClose.State() != byRight.State() {
		t.Fatalf("「結束」離開之後在 %v，右鍵離開之後在 %v",
			byClose.State(), byRight.State())
	}
	if !bytes.Equal(snapshot(byClose), snapshot(byRight)) {
		t.Fatal("「結束」與右鍵離開之後的畫面不一樣")
	}
}

// 診斷戰場起點；實際主選單至S1的返回另由規格139的正常X11收據驗證。
func TestBattleSettingsReturnsToCallerWithoutAdvancing(t *testing.T) {
	for _, right := range []bool{false, true} {
		g := reachDashiliBattle(t)
		units := append([]rules.BattleUnit(nil), g.battleUnits...)
		ship, fleet := bytes.Clone(g.slot.Ship), bytes.Clone(g.slot.Fleet)
		turn, seed := g.battleTurn, g.battleRand.Seed()
		g.Click(600, 240)
		g.Click(600, 342)
		if g.State() != StateSettings {
			t.Fatal("戰術設定入口未到達")
		}
		g.Click(232, 370)
		if !g.ShowEnemies() {
			t.Fatal("設定未開啟敵軍顯示")
		}
		if right {
			g.RightClick(410, 370)
		} else {
			g.Click(410, 370)
		}
		if !g.BattleInputReady() || !g.ShowEnemies() {
			t.Fatal("設定沒有返回戰場或遺失偏好")
		}
		if turn != g.battleTurn || seed != g.battleRand.Seed() || !reflect.DeepEqual(units, g.battleUnits) || !bytes.Equal(ship, g.slot.Ship) || !bytes.Equal(fleet, g.slot.Fleet) {
			t.Fatal("設定返回推進回合、亂數或戰場資料")
		}
		g.Frame()
	}
}
