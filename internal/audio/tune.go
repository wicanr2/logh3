package audio

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/game"
)

// Settings 是原版設定畫面上那三行，也就是 `ds:8F33`／`8F34`／`8F35`
// （[`docs/internals/17`](../../docs/internals/17-mds.md) §哪個畫面播哪一首、
// [`docs/re/274`](../../docs/re/274-which-screen-plays-which-tune.md)）。
//
// 這三格跟著存檔走：`WORK.DT?` 的 `+0x1F`／`+0x20`／`+0x21`。
// **原版沒有「畫面 → 曲子」的對照表**——八首 `BGM_*` 裡六首是玩家挑的
// 平時配樂、兩首是玩家挑的戰鬥配樂，所以 [TuneFor] 吃的是設定不是畫面表。
type Settings struct {
	// Device 是音源：0 ＝ MIDI（`_M`）、1 ＝ ＦＭ（`_F`）、2 ＝ 關閉。
	Device int

	// BGM 是平時配樂 0–5 → `BGM_0`–`BGM_5`。
	BGM int

	// Battle 是戰鬥配樂：0 → `BGM_6`、1 → `BGM_7`、2 ＝ 沿用 BGM 那一首。
	Battle int
}

// DefaultSettings 是原版拿出貨的 `MUSIC.INI` 開新遊戲時的三格，
// 也就是 [game.New] 給的那一組。
//
// `Device` 是 1（ＦＭ）而不是 0：出貨的 `MUSIC.INI` 把「ＭＩＤＩ Ｉ／Ｆ」
// 設成「無」（`+4` ＝ `FFFF`），偵測不到 MPU-401 就退到 OPL，於是原版
// 開箱讀的一律是 `_F` 那一組。457 份收據裡出現過的 `.MDS` 沒有一個是 `_M`。
//
// `BGM`／`Battle` 都是 0：開新遊戲時原版把兩格清零
// （簽章 `31 C0 A2 34 8F A2 35 8F CB`）。
var DefaultSettings = Settings{Device: game.MusicFM}

// suffix 是檔名最後那個字元；關閉的時候回空字串。
func (s Settings) suffix() string {
	switch s.Device {
	case 0:
		return "M"
	case 1:
		return "F"
	}
	return ""
}

// TuneFor 回這個畫面該播哪一首，沒有配樂就回空字串。
//
// faction 是玩家陣營（0 帝國、1 同盟），只有結局用得到。
//
// 對照的來源全部是反組譯或收據，沒有猜的：
//
//   - 開頭動畫一律 `OPEN_M`。`OPEN.EXE` 的 FM 與 MIDI 兩個外皮都寫死 `'m'`，
//     所以 `OPEN_F` 在原版裡永遠不會被打開。
//   - **主選單與防拷問答沒有配樂**：`GIN3PS.EXE` 從啟動到答完防拷題之間
//     一個 `.MDS` 都沒開過（收據 `key1`／`main1e`／`chain1` 的檔案存取序列）。
//   - 結局照命令列的第二個數字：`9 1` ＝ 帝國勝利 → `END_1`、
//     `9 2` ＝ 同盟勝利 → `END_2`（`docs/internals/16`）。
//   - 其餘畫面都是平時配樂，**戰術畫面（戰場）也是**：收據 `s3assault` 裡
//     `BGM_3F` 一路撐過進戰場那一段，換成戰鬥曲的是交戰演出
//     （`ANI_PNL1.A98` ＋ `COMANI.BMG`）而不是進戰場本身。
//
// remake 還沒有交戰演出那個狀態，所以 [Settings.Battle] 目前沒有出口；
// 欄位先擺著，接上演出之後那一支畫面直接呼叫 [Settings.BattleTune]。
func TuneFor(st game.State, faction int, s Settings) string {
	suf := s.suffix()
	if suf == "" {
		return ""
	}
	switch st {
	case game.StateIntro:
		return "OPEN_M"
	case game.StateMainMenu, game.StateCopyProtection:
		return ""
	case game.StateEnding:
		return fmt.Sprintf("END_%d%s", faction+1, suf)
	}
	return fmt.Sprintf("BGM_%d%s", s.BGM, suf)
}

// BattleTune 是交戰演出那一段的曲子：`Battle` ＝ 2 時沿用平時那一首。
func (s Settings) BattleTune() string {
	suf := s.suffix()
	if suf == "" {
		return ""
	}
	n := s.BGM
	if s.Battle != 2 {
		n = s.Battle + 6
	}
	return fmt.Sprintf("BGM_%d%s", n, suf)
}
