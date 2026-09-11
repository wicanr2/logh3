package game

// 設定畫面（系統指令 →「設定」）上那三行音樂設定。
//
// 原版把選項存成一份熱區清單（資料段 `ds:03D8`，22 筆 12 bytes），
// 命中之後看動作碼分三段（`docs/re/274` §5、`docs/re/285`）：
//
//	0–2   [8F33] ＝ 碼        音源：ＭＩＤＩ／ＦＭ／關閉，重畫那一列後試聽平時曲
//	3–8   [8F34] ＝ 碼 − 3    平時配樂六選一，重畫清單後試聽
//	9–11  [8F35] ＝ 碼 − 9    戰鬥配樂三選一，重畫清單，**不試聽**
//
// 試聽不必另外做：`cmd/gin3` 每一幀都問「現在該播哪一首」，曲名一變
// 就換曲，而戰鬥配樂不參與那個算式（`internal/audio.TuneFor`）——
// 原版那三段的差別在 remake 這邊自然成立。
//
// 同一份清單上另外三行（訊息速度、動畫效果、動畫速度）與「結束」、
// 「ＭＩＤＩ Ｉ／Ｆ」也在規格裡，但它們存在哪還沒解，所以只讀出來沒有接。

import (
	"strconv"
	"strings"

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
// 其餘填 8。三組各自獨立，同一組裡只有一個是選中的。
const (
	settingsPickedFill uint8 = 2
	settingsPlainFill  uint8 = 8
)

// musicGroups 是「動作碼前綴 → 存哪一格、幾個選項」。
// 三組的上限就是原版那三段動作碼的長度（3／6／3）。
func (m *MusicSettings) groups() []struct {
	prefix string
	field  *int
	count  int
} {
	return []struct {
		prefix string
		field  *int
		count  int
	}{
		{"device-", &m.Device, 3},
		{"bgm-", &m.BGM, 6},
		{"battle-", &m.Battle, 3},
	}
}

// MusicSettings 回那三格目前的值（`cmd/gin3` 拿它決定播哪一首）。
func (g *Game) MusicSettings() MusicSettings { return g.music }

// settingsClick 處理設定畫面上的點擊。
//
// 「ＭＩＤＩ」那一格點不下去：原版的可選旗標由開機偵測寫（`ds:03DC`），
// 出貨設定下是 0，`HitTest` 直接跳過它，所以這裡連 `device-0` 都收不到。
func (g *Game) settingsClick(x, y int) {
	spec := g.commandSpec(StateSettings)
	if spec == nil {
		return
	}
	act := spec.HitTest(x, y)
	for _, c := range g.music.groups() {
		rest, ok := strings.CutPrefix(act, c.prefix)
		if !ok {
			continue
		}
		n, err := strconv.Atoi(rest)
		if err != nil || n < 0 || n >= c.count {
			return
		}
		*c.field = n
		return
	}
}

// applyMusicSettings 把那三格反映到畫面上：每一組裡選中的那一格換成
// 反白底色。原版選完會重畫那一列或整張清單（`docs/re/274` §5），
// remake 是每一幀重畫，所以只要把底色算對就行。
//
// 找不到框的熱區直接跳過——「ＭＩＤＩ」不可選，原版連框都沒畫。
func (g *Game) applyMusicSettings(spec *screen.MainGameSpec) {
	if spec == nil {
		return
	}
	w := spec.Windows["overlay"]
	if w == nil {
		return
	}
	for _, c := range g.music.groups() {
		for _, h := range spec.Hotspots {
			rest, ok := strings.CutPrefix(h.Action, c.prefix)
			if !ok {
				continue
			}
			n, err := strconv.Atoi(rest)
			if err != nil {
				continue
			}
			fill := settingsPlainFill
			if n == *c.field {
				fill = settingsPickedFill
			}
			w.SetHotspotFill(h.X, h.Y, h.W, h.H, fill)
		}
	}
}
