package ui

import (
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// Cell 是文字視窗的一格：一個全形字碼與一個屬性。
//
// 原版的儲存格就是 4 bytes `{字碼, 屬性, 0}`，清除時填兩個空格 `0x2020`
// （`docs/internals/04-text-windows-and-big5.md`）。這裡沿用同一個模型，
// 因為版面規則（欄距、列距、反白）都是照格子走的。
type Cell struct {
	Code uint16 // Big5 全形字碼；0x2020 ＝ 空白
	Attr uint8  // 色號
}

// Shadow 是文字的立體邊：同一個字先往左上偏一格用 Color 畫一次，
// 再用本色畫在原位。少了它，字會薄一圈，而畫面上看起來只是
// 「字比較細」，不像錯。
//
// ⚠ **左移是對整列的位元做的，不是把整個字平移。** 所以字身最左邊那一
// 行的像素在立體邊裡會被移掉，不會溢出到左邊那一格。差別只有幾個像素，
// 但那幾個像素每個字都出現一次。
type Shadow struct {
	DY    int
	Color uint8
	On    bool
}

// Blank 是清除後的格子內容。
const Blank uint16 = 0x2020

// TextWindow 是一個有外框的文字視窗。
type TextWindow struct {
	X, Y               int // 外框左上角（畫布座標）
	Cols, Rows         int // 格子數
	OriginX, OriginY   int // 第一格左上角相對外框的位移
	ColPitch, RowPitch int
	Frame              Frame
	Shadow             Shadow
	Cells              []Cell
}

// NewTextWindow 造一個清空的文字視窗。
func NewTextWindow(x, y, cols, rows int, f Frame) *TextWindow {
	w := &TextWindow{
		X: x, Y: y, Cols: cols, Rows: rows,
		ColPitch: font.GlyphW, RowPitch: 20,
		Frame:  f,
		Shadow: Shadow{DY: -1, Color: 3, On: true},
		Cells:  make([]Cell, cols*rows),
	}
	for i := range w.Cells {
		w.Cells[i] = Cell{Code: Blank}
	}
	return w
}

// SetText 把一串全形字碼填進第 row 列，從第 0 欄開始。
func (w *TextWindow) SetText(row int, attr uint8, codes []uint16) {
	for c, code := range codes {
		if c >= w.Cols || row >= w.Rows {
			return
		}
		w.Cells[row*w.Cols+c] = Cell{Code: code, Attr: attr}
	}
}

// PixelSize 回外框要多大才裝得下格子與邊框。
func (w *TextWindow) PixelSize(frameW, frameH int) (int, int) {
	return frameW, frameH
}

// Draw 畫外框與文字。缺字（字型裡沒有那個碼）會被跳過並回報，
// **不畫成空白**——空白與缺字在畫面上分不出來。
func (w *TextWindow) Draw(fb *gfx.Framebuffer, fnt *font.Font, frameW, frameH int) []uint16 {
	w.Frame.Draw(fb, w.X, w.Y, frameW, frameH)
	var missing []uint16
	for r := 0; r < w.Rows; r++ {
		for c := 0; c < w.Cols; c++ {
			cell := w.Cells[r*w.Cols+c]
			if cell.Code == Blank {
				continue
			}
			g := fnt.Glyph(cell.Code)
			if g == nil {
				missing = append(missing, cell.Code)
				continue
			}
			px := w.X + w.OriginX + c*w.ColPitch
			py := w.Y + w.OriginY + r*w.RowPitch
			blit := func(shift int, oy int, col uint8) {
				for gy, bits := range g {
					row := bits << uint(shift) // 左移：最左邊那一行被移掉
					for gx := 0; gx < font.GlyphW; gx++ {
						if row>>(font.GlyphW-1-gx)&1 != 0 {
							fb.Set(px+gx, py+oy+gy, col)
						}
					}
				}
			}
			if w.Shadow.On {
				blit(1, w.Shadow.DY, w.Shadow.Color)
			}
			blit(0, 0, cell.Attr)
		}
	}
	return missing
}
