package screen

import (
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 艦隊編號印進視窗：`3982:0038(視窗, 編號)`（`docs/re/143`）。
//
// 兩條路都只佔一格：
//
//	編號 < 10  → ds:8E86 "\x06%d\x07"  ; 開半形轉全形 → 一個全形數字 → 關
//	編號 >= 10 → ds:8E8B "%2d"          ; 兩個半形塞一格
//
// 面板的分行是靠每一段剛好佔滿視窗寬度擠出來的，所以「一格」這件事是版面的
// 前提，不是美觀問題：兩位數印成兩個全形就會把後面整列推掉一格。
// 過場視窗那一套（`3982:0073`）在 `TurnWindowSpec.FleetNumber`，同一條規則。

// writeFleetNumber 把編號寫進內容緩衝區。
func writeFleetNumber(b *ui.TextBuffer, n int) {
	if n >= 10 {
		b.Printf([]byte("%2d"), n)
		return
	}
	b.Attr |= ui.AttrWiden
	b.Printf([]byte("%d"), n)
	b.Attr &^= ui.AttrWiden
}

// fleetNumberBytes 是同一條規則的「先排成位元組再 Write」版本（交戰演出用）。
// 小於 10 時夾在 `\x06`／`\x07` 之間，由 PutByte 那一層轉全形。
func fleetNumberBytes(n int) []byte {
	if n >= 10 {
		return ui.Sprintf([]byte("%2d"), n)
	}
	return ui.Sprintf([]byte("\x06%d\x07"), n)
}

// drawFleetNumberRow 畫清單列上的編號（`0B20:059B` 那一類自己畫字的列）。原版在
// 那裡也是同一條兩位數規則：小於 10 一個全形數字，否則兩個半形塞進同一格
// ——後面的「艦隊」在固定欄位，不會被推開。
func drawFleetNumberRow(fb *gfx.Framebuffer, fnt *font.Font, x, y, n int, ink uint8, pitch int) {
	if n < 10 {
		ui.DrawGlyphClipped(fb, fnt, x, y, fullDigit0+uint16(n), ink, panelShadow,
			-1, false, 0, y, gfx.Width, pitch)
		return
	}
	s := strconv.Itoa(n)
	for k := 0; k < len(s); k++ {
		ui.DrawHalfGlyphClipped(fb, fnt, x, y, s[k], ink, panelShadow,
			-1, false, 0, y, gfx.Width, pitch)
		x += font.HalfW
	}
}
