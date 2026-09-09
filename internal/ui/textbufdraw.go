package ui

import (
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 把內容緩衝區畫到畫面上（原版 `34CB:01A4` 把內容格搬進畫面格表，
// `2376:040E` → `2376:0E0D` 再畫；`docs/re/138`、`141`）。
//
// 一格的順序是**底色 → 立體邊 → 字 → 裝飾**，不能顛倒：裝飾最後畫才會
// 蓋掉字的立體邊，順序反了整個框差兩個像素。

// 字色與立體邊：`2376:10B7` 拿 `顏色 & 0x0F` 當本體色，立體邊看
// `旗標 & 0x40` 選 7 或 3。目前看到的視窗都是 3。
const (
	TextShadowInk    = 3
	TextShadowAltInk = 7
	TextShadowDY     = -1
	TextCellDY       = 2 // 字在格內往下 2 像素（收據 s3msg 量到 (80,202)）
)

// DrawTextBuffer 把緩衝區畫在畫面格 (cellX, cellY)。tiles 可以是 nil，
// 那樣就不畫 `0xFF00` 那一段的外框圖塊。回沒有字模的字碼。
func DrawTextBuffer(fb *gfx.Framebuffer, fnt *font.Font, tiles *FrameTiles,
	b *TextBuffer, cellX, cellY int) []uint16 {
	if fb == nil || b == nil {
		return nil
	}
	var missing []uint16
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			cell := b.Cells[r*b.Cols+c]
			x := (cellX + c) * FrameTileW
			y := (cellY+r)*FrameTileH + FrameCellOriginY
			if m := drawContentCell(fb, fnt, tiles, cell, x, y); m != 0 {
				missing = append(missing, m)
			}
		}
	}
	return missing
}

func drawContentCell(fb *gfx.Framebuffer, fnt *font.Font, tiles *FrameTiles,
	cell ContentCell, x, y int) uint16 {
	hi := uint8(cell.Code >> 8)
	if hi == 0xFF {
		// 外框圖塊：色號 0 填視窗底色（`2376:0E4A`）。
		if tiles != nil {
			(TileFrame{Tiles: tiles, Style: 1, Fill: cell.Colour >> 4}).
				drawTile(fb, x, y, int(cell.Code&0xFF)-1)
		}
		return 0
	}
	bg := cell.Colour >> 4
	ink := cell.Colour & 0x0F
	shadow := uint8(TextShadowInk)
	if cell.Flags&0x40 != 0 {
		shadow = TextShadowAltInk
	}
	fb.FillRect(x, y, FrameTileW, FrameTileH, bg)
	var missing uint16
	if fnt != nil {
		if hi >= 0x80 {
			if !DrawGlyph(fb, fnt, x, y+TextCellDY, cell.Code,
				ink, shadow, TextShadowDY, cell.Flags&CellBold != 0) {
				missing = cell.Code
			}
		} else {
			// 兩個半形：高位元組是左邊那個。
			if hi != 0x20 {
				DrawHalfGlyph(fb, fnt, x, y+TextCellDY, hi,
					ink, shadow, TextShadowDY, cell.Flags&CellBold != 0)
			}
			if lo := uint8(cell.Code); lo != 0x20 {
				DrawHalfGlyph(fb, fnt, x+FrameTileW/2, y+TextCellDY, lo,
					ink, shadow, TextShadowDY, cell.Flags&CellBold != 0)
			}
		}
	}
	DrawCellDecor(fb, x, y, cell.Flags)
	return missing
}
