package game

// 設定畫面（系統指令 →「設定」）上那六行。
//
// 原版把選項存成一份熱區清單（資料段 `ds:03D8`，22 筆 12 bytes），
// 命中之後一支分派常式照動作碼一條一條比下去（`0429:0391` ＝ 線性
// `0x04681`，`docs/re/274` §5、`docs/re/285`、`docs/re/286`）：
//
//	0–2    [8F33] ＝ 碼        音源：ＭＩＤＩ／ＦＭ／關閉，重畫那一列後試聽平時曲
//	3–8    [8F34] ＝ 碼 − 3    平時配樂六選一，重畫清單後試聽
//	9–11   [8F35] ＝ 碼 − 9    戰鬥配樂三選一，重畫清單，**不試聽**
//	12／13 [8F38] ±10          訊息速度：上限 300、低於 10 就不再減
//	14／15 [8F36] ＝ 1／0      動畫效果 開啟／關閉
//	16–18  [8F37] ＝ 0／1／2   動畫速度 普通／快速／最快
//	19     回傳 1             「結束」——與右鍵走同一條出口
//
// 試聽不必另外做：`cmd/gin3` 每一幀都問「現在該播哪一首」，曲名一變
// 就換曲，而戰鬥配樂不參與那個算式（`internal/audio.TuneFor`）——
// 原版那三段的差別在 remake 這邊自然成立。
//
// 沒接的兩筆：「ＭＩＤＩ Ｉ／Ｆ」（動作碼 22）要先有 `MUSIC.INI` 的讀寫
// （`docs/re/274` §10），`BUG!`（20）開的是另一張畫面（`ds:902C` 那份
// 七項清單，`docs/re/286` §六）。

import (
	"strconv"

	"github.com/wicanr2/logh3/internal/screen"
)

// 音源的三個值（原版 `ds:8F33`）。
const (
	// MusicMIDI 走 MPU-401，讀的是 `_M` 那一組。
	MusicMIDI = 0

	// MusicFM 走 OPL，讀的是 `_F` 那一組。**出貨的原版就是這個**：
	// `MUSIC.INI` 的「ＭＩＤＩ Ｉ／Ｆ」是「無」，偵測不到 MPU-401 就退到
	// OPL（`docs/re/274` §3、§10）。既有的執行期傾印裡 `ds:8F33` 一律是 1。
	MusicFM = 1

	// MusicOff ＝ 不播。
	MusicOff = 2
)

// MusicSettings 是設定畫面上那三行的值，也就是原版的
// `ds:8F33`／`8F34`／`8F35`。跟著存檔走（`WORK.DT?` 的
// `+0x1F`／`+0x20`／`+0x21`，`docs/internals/30`）。
//
// 播哪一首的算式在 `internal/audio`（那一層是純的、測得到）；
// 這裡只管值本身怎麼變、存到哪裡去。
type MusicSettings struct {
	// Device 是音源：0 ＝ MIDI、1 ＝ ＦＭ、2 ＝ 關閉。
	Device int

	// BGM 是平時配樂 0–5 → `BGM_0`–`BGM_5`。
	BGM int

	// Battle 是戰鬥配樂：0 → `BGM_6`、1 → `BGM_7`、2 ＝ 沿用平時那一首。
	Battle int
}

// 設定畫面上那幾組選項的框底色（量自收據 `settings1e`）：選中的填 2，
// 其餘填 8。原版寫進去的是 `0x27`／`0x87`——高四位是底色、低四位是字色
// （`docs/re/286` §二），字色兩種都是 7。各組互相獨立，同一組裡只有一個
// 是選中的。
const (
	settingsPickedFill uint8 = 2
	settingsPlainFill  uint8 = 8
)

// 訊息速度（原版 `ds:8F38`）的步進與上限。原版兩支加減常式的判斷是
// 「已經 >= 300 就不加」與「不到 10 就不減」，所以**按到底是 0，不是 10**
// （`docs/re/286` §二）。
const (
	msgSpeedStep = 10
	msgSpeedMax  = 300
)

// DisplaySettings 是設定畫面上另外三行，也就是原版的 `ds:8F36`／`8F37`／
// `8F38`。跟著存檔走（`WORK.DT?` 的 `+0x22`／`+0x23`／`+0x24`，
// `docs/internals/30`）。
//
// remake 目前只管值本身怎麼變、存到哪裡去。三個值在原版各自控制什麼
// 已經解出來了（記在 `docs/re/286` §五），但那三樣演出 remake 還沒做，
// 所以沒有地方讀它們。
type DisplaySettings struct {
	// Anim 是動畫效果：1 開啟、0 關閉。原版是 0 的時候，交戰演出那支
	// 常式進門就返回。
	Anim int

	// AnimSpeed 是動畫速度：0 普通、1 快速、2 最快。原版拿它當索引查
	// `ds:8F10` 的 `{8, 4, 0}`，查到的就是每一格等幾個計時器 tick。
	AnimSpeed int

	// MsgSpeed 是訊息停留的計時器 tick 數，也就是畫面上印的那個數字。
	// **數字大 ＝ 停得久**；等待途中按左鍵或右鍵可以提早結束。
	MsgSpeed int
}

// newDisplaySettings 回原版的開機預設：動畫開啟、普通、90
// （`0x03A28`／`0x03A2C`／`0x039E1`）。46 份執行期傾印、三格出貨存檔與
// 新遊戲底稿 `work-796-01.json` 的 `+0x22`–`+0x25` 全是 `01 00 5A 00`。
func newDisplaySettings() DisplaySettings {
	return DisplaySettings{Anim: 1, AnimSpeed: 0, MsgSpeed: 90}
}

// settingsChoice 是一個單選格：點下去把某一格設成某個值。
type settingsChoice struct {
	action string
	field  *int
	value  int
}

// settingsChoices 把五組單選攤平成一張表——原版的分派就是一條一條比
// 動作碼比下去，沒有共用的迴圈（`docs/re/286` §二）。每個動作碼在表裡
// 只出現一次，所以同一張表既能處理點擊也能算底色。
func (g *Game) settingsChoices() []settingsChoice {
	c := make([]settingsChoice, 0, 17)
	add := func(prefix string, field *int, n int) {
		for i := 0; i < n; i++ {
			c = append(c, settingsChoice{prefix + strconv.Itoa(i), field, i})
		}
	}
	add("device-", &g.music.Device, 3)
	add("bgm-", &g.music.BGM, 6)
	add("battle-", &g.music.Battle, 3)
	c = append(c,
		settingsChoice{"anim-on", &g.display.Anim, 1},
		settingsChoice{"anim-off", &g.display.Anim, 0})
	add("anim-speed-", &g.display.AnimSpeed, 3)
	return c
}

// MusicSettings 回那三格目前的值（`cmd/gin3` 拿它決定播哪一首）。
func (g *Game) MusicSettings() MusicSettings { return g.music }

// DisplaySettings 回設定畫面上另外三行目前的值。
func (g *Game) DisplaySettings() DisplaySettings { return g.display }

// settingsClick 處理設定畫面上的點擊。
//
// 「ＭＩＤＩ」那一格點不下去：原版的可選旗標由開機偵測寫（`ds:03DC`），
// 出貨設定下是 0，`HitTest` 直接跳過它，所以這裡連 `device-0` 都收不到。
//
// MIDI I/F所在列由正式app的音色選單接管（114）。BUG!出貨不可選（113）。
func (g *Game) settingsClick(x, y int) {
	if enemyVisibilityButtonContains(x, y) {
		g.ToggleEnemyVisibility()
		return
	}
	spec := g.commandSpec(StateSettings)
	if spec == nil {
		return
	}
	switch act := spec.HitTest(x, y); act {
	case "":
	case "msg-speed-inc":
		if g.display.MsgSpeed < msgSpeedMax {
			g.display.MsgSpeed += msgSpeedStep
		}
	case "msg-speed-dec":
		if g.display.MsgSpeed >= msgSpeedStep {
			g.display.MsgSpeed -= msgSpeedStep
		}
	case "close":
		// 原版的「結束」只是讓熱區分派回傳 1，設定畫面那個訊息迴圈跟著
		// 結束、`retf` 回呼叫端——**與右鍵（輸入結果 2）同一條出口**，
		// 兩條路收在同一個 `retf` 上（`docs/re/286` §三）。所以這裡直接
		// 走右鍵那條路，以後右鍵改去哪，「結束」自動跟著改。
		g.rightClick(x, y)
	default:
		for _, c := range g.settingsChoices() {
			if c.action == act {
				*c.field = c.value
				return
			}
		}
	}
}

// applySettings 把目前的值反映到畫面上：每一組裡選中的那一格換成反白
// 底色，訊息速度那個數字換成目前的值。原版是選完重畫那一列或整張清單
// （`docs/re/274` §5、`docs/re/286` §二），remake 是每一幀重畫，所以只要
// 把底色與數字算對就行。
//
// 找不到框的熱區直接跳過——「ＭＩＤＩ」不可選，原版連框都沒畫。
func (g *Game) applySettings(spec *screen.MainGameSpec) {
	if spec == nil {
		return
	}
	w := spec.Windows["overlay"]
	if w == nil {
		return
	}
	for _, c := range g.settingsChoices() {
		fill := settingsPlainFill
		if *c.field == c.value {
			fill = settingsPickedFill
		}
		for _, h := range spec.Hotspots {
			if h.Action != c.action {
				continue
			}
			w.SetHotspotFill(h.X, h.Y, h.W, h.H, fill)
			break
		}
	}
	if f := spec.MsgSpeed; f != nil {
		w.SetFieldText(f, strconv.Itoa(g.display.MsgSpeed))
	}
}
