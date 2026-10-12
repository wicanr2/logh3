package ui

import (
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
)

// Cell 是文字視窗的一格：一個全形字碼與一個屬性。
//
// 原版的儲存格就是 4 bytes `{字碼, 屬性, 0}`，清除時填兩個空格 `0x2020`
// （`docs/internals/04-text-windows-and-big5.md`）。
//
// **屬性是「底色 << 4 | 字色」**，不是單純的字色。反白就是換一組屬性重畫
// 那一格——底色與字色都是 2 的話，整格變成實心紅、字看不見。原版確實是
// 這樣做的：probe 的檔案存取記錄顯示反白那兩格的字形在畫完之後被**重讀
// 重畫了一次**，所以反白不是「蓋一條橫條」（`docs/internals/07`）。
type Cell struct {
	Code   uint16 // Big5 全形字碼；0x2020 ＝ 空白
	Attr   uint8  // 高 4 位 ＝ 底色、低 4 位 ＝ 字色
	Shadow uint8  // 立體邊的色號
}

// FG／BG 拆出字色與底色。
func (c Cell) FG() uint8 { return c.Attr & 0x0F }
func (c Cell) BG() uint8 { return c.Attr >> 4 }

// Blank 是清除後的格子內容。
const Blank uint16 = 0x2020

// Shadow 是文字的立體邊：同一個字先往上一列、每列位元左移一位，
// 用 Color 畫一次，再用字色畫在原位。
//
// ⚠ **左移是對整列的位元做的，不是把整個字平移。** 所以字身最左邊那一行
// 的像素在立體邊裡會被移掉，不會溢出到左邊那一格。差別只有幾個像素，
// 但那幾個像素每個字都出現一次。
type Shadow struct {
	DY int
	On bool
}

// TextWindow 是一個有外框的文字視窗。
type TextWindow struct {
	X, Y               int // 外框左上角（畫布座標）
	Cols, Rows         int // 格子數
	OriginX, OriginY   int // 第一格**字身**左上角相對外框的位移
	CellPadY           int // 格子矩形比字身高出來的上緣（原版是 2）
	ColPitch, RowPitch int
	Frame              TileFrame
	Shadow             Shadow
	Cells              []Cell
}

// NewTextWindow 造一個清空的文字視窗。attr 是清除時填進每一格的屬性。
//
// 外框是 `FRAME.BLK` 的圖塊（`TileFrame`）：原版的視窗位置存的是**格**
// （`視窗[+5]`／`[+6]`，`docs/internals/47`），框那一圈是寫進畫面格表的
// 八個字碼，所以視窗只可能落在 16 × 20 的格線上。
func NewTextWindow(x, y, cols, rows int, f TileFrame, attr, shadow uint8) *TextWindow {
	w := &TextWindow{
		X: x, Y: y, Cols: cols, Rows: rows,
		CellPadY: 2,
		ColPitch: font.GlyphW, RowPitch: 20,
		Frame:  f,
		Shadow: Shadow{DY: -1, On: true},
		Cells:  make([]Cell, cols*rows),
	}
	for i := range w.Cells {
		w.Cells[i] = Cell{Code: Blank, Attr: attr, Shadow: shadow}
	}
	return w
}

// SetText 把一串全形字碼填進第 row 列，從第 0 欄開始。
func (w *TextWindow) SetText(row int, attr, shadow uint8, codes []uint16) {
	for c, code := range codes {
		if c >= w.Cols || row >= w.Rows {
			return
		}
		w.Cells[row*w.Cols+c] = Cell{Code: code, Attr: attr, Shadow: shadow}
	}
}

// SetRowAttr 換掉一整列的屬性（反白／取消反白就是這個）。
func (w *TextWindow) SetRowAttr(row int, attr, shadow uint8) {
	for c := 0; c < w.Cols; c++ {
		w.Cells[row*w.Cols+c].Attr = attr
		w.Cells[row*w.Cols+c].Shadow = shadow
	}
}

// cellRect 回第 (col,row) 格的矩形。格子比字身高 CellPadY，
// 高度是列距——反白時填滿的就是這個矩形。
func (w *TextWindow) cellRect(col, row int) (x, y, cw, ch int) {
	return w.X + w.OriginX + col*w.ColPitch,
		w.Y + w.OriginY + row*w.RowPitch - w.CellPadY,
		w.ColPitch, w.RowPitch
}

// DrawRow 只畫一列的格子。**繪製順序會影響畫面**：原版是把一般的列畫完、
// 畫游標，最後才畫反白的那一列——所以游標會被反白列蓋掉一部分。
func (w *TextWindow) DrawRow(fb *gfx.Framebuffer, fnt *font.Font, r int) []uint16 {
	var missing []uint16
	for c := 0; c < w.Cols; c++ {
		if m := w.drawCell(fb, fnt, c, r); m != 0 {
			missing = append(missing, m)
		}
	}
	return missing
}

// drawCell 畫一格，回缺字的字碼（0 ＝ 沒缺）。
func (w *TextWindow) drawCell(fb *gfx.Framebuffer, fnt *font.Font, c, r int) uint16 {
	cell := w.Cells[r*w.Cols+c]
	cx, cy, cw, ch := w.cellRect(c, r)
	fb.SurfaceRect(cx, cy, cw, ch, cell.BG())
	if cell.Code == Blank {
		return 0
	}
	g := fnt.Glyph(cell.Code)
	if g == nil {
		return cell.Code
	}
	px, py := cx, cy+w.CellPadY
	start := fb.BeginGroup()
	defer func() {
		fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: px, Y: py, W: font.GlyphW, H: font.GlyphH,
			Code: cell.Code, Colour: cell.FG(), Shadow: cell.Shadow, ShadowDY: w.Shadow.DY, Clip: image.Rect(0, 0, gfx.Width, gfx.Height)})
	}()
	blit := func(shift, oy int, col uint8) {
		for gy, bits := range g {
			row := bits << uint(shift)
			for gx := 0; gx < font.GlyphW; gx++ {
				if row>>(font.GlyphW-1-gx)&1 != 0 {
					fb.Set(px+gx, py+oy+gy, col)
				}
			}
		}
	}
	if w.Shadow.On {
		blit(1, w.Shadow.DY, cell.Shadow)
	}
	blit(0, 0, cell.FG())
	return 0
}

// Draw 畫外框與所有格子。缺字（字型裡沒有那個碼）會被跳過並回報，
// **不畫成空白**——空白與缺字在畫面上分不出來。
//
// 矩形對不上格線時 `DrawRect` 不畫框（寧可少一個框也不要畫錯位置）。
// 產品裡用到的每一個矩形都由 `TestMenuRectsAreCellAligned` 釘著。
func (w *TextWindow) Draw(fb *gfx.Framebuffer, fnt *font.Font, frameW, frameH int) []uint16 {
	w.Frame.DrawRect(fb, w.X, w.Y, frameW, frameH)
	var missing []uint16
	for r := 0; r < w.Rows; r++ {
		missing = append(missing, w.DrawRow(fb, fnt, r)...)
	}
	return missing
}

// DrawGlyph 在指定像素座標畫一個字（不填背景）。
//
// 給「不是格狀排版」的畫面用：防拷問答那三行字每行的字數不同、
// 輸入框那一格又另起一個 x，硬塞進 TextWindow 的格點要多算一堆空白。
// 陰影規則與 drawCell 相同：先把每一列往左位移一格畫陰影，再畫本體。
//
// bold ＝ 原版的粗體：**先**把字形每一列往右加粗一格，陰影與本體都用
// 加粗後的點陣（`docs/internals/25`）。順序不能顛倒——先畫陰影再加粗
// 會多出一排像素。
func DrawGlyph(fb *gfx.Framebuffer, fnt *font.Font, x, y int, code uint16, fg, shadow uint8, shadowDY int, bold bool) bool {
	return DrawGlyphClipped(fb, fnt, x, y, code, fg, shadow, shadowDY, bold,
		0, 0, gfx.Width, gfx.Height)
}

// DrawGlyphClipped 同 DrawGlyph，但把整個字（含立體邊）裁在指定矩形內。
//
// 按鈕上的字要用這條：立體邊那一趟會往上移一列，落到按鈕外框上。
// 原版把字裁在按鈕的內容區裡，外框不會被立體邊咬掉一格。
func DrawGlyphClipped(fb *gfx.Framebuffer, fnt *font.Font, x, y int, code uint16,
	fg, shadow uint8, shadowDY int, bold bool, cx, cy, cw, ch int) bool {
	g := fnt.Glyph(code)
	if g == nil {
		return false
	}
	start := fb.BeginGroup()
	defer func() {
		fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: x, Y: y, W: font.GlyphW, H: font.GlyphH,
			Code: code, Colour: fg, Shadow: shadow, ShadowDY: shadowDY, Bold: bold, Clip: image.Rect(cx, cy, cx+cw, cy+ch)})
	}()
	if bold {
		for i, v := range g {
			g[i] = v | v>>1
		}
	}
	blit := func(shift, oy int, col uint8) {
		for gy, bits := range g {
			row := bits << uint(shift)
			for gx := 0; gx < font.GlyphW; gx++ {
				if row>>(font.GlyphW-1-gx)&1 == 0 {
					continue
				}
				px, py := x+gx, y+oy+gy
				if px < cx || py < cy || px >= cx+cw || py >= cy+ch {
					continue
				}
				fb.Set(px, py, col)
			}
		}
	}
	if shadow != 0 {
		blit(1, shadowDY, shadow)
	}
	blit(0, 0, fg)
	return true
}

// DrawHalfGlyph 畫一個半形字（8×15）。數字與百分比走這條。
//
// 半形字是**兩個一組塞進同一個 16 像素的格子**（`docs/internals/25`）。
// 陰影那一趟是整格 16 位元一起往左移一格，所以**右半那個字最左邊那一行的
// 陰影會落到左半的地盤上**，不會被丟掉；左半那個字移出格外的那一行才會。
// 判準是這個字在不在 16 的邊界上。
func DrawHalfGlyph(fb *gfx.Framebuffer, fnt *font.Font, x, y int, code uint8, fg, shadow uint8, shadowDY int, bold bool) bool {
	return DrawHalfGlyphClipped(fb, fnt, x, y, code, fg, shadow, shadowDY, bold,
		0, 0, gfx.Width, gfx.Height)
}

// DrawHalfGlyphClipped 同 DrawHalfGlyph，但把整個字（含立體邊）裁在矩形內。
func DrawHalfGlyphClipped(fb *gfx.Framebuffer, fnt *font.Font, x, y int, code uint8,
	fg, shadow uint8, shadowDY int, bold bool, cx, cy, cw, ch int) bool {
	start := fb.BeginGroup()
	defer func() {
		fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: x, Y: y, W: font.HalfW, H: font.GlyphH,
			Code: uint16(code), Half: true, Colour: fg, Shadow: shadow, ShadowDY: shadowDY, Bold: bold, Clip: image.Rect(cx, cy, cx+cw, cy+ch)})
	}()
	if shadow != 0 {
		halfPass(fb, fnt, x, y, code, shadow, 1, shadowDY, bold, cx, cy, cw, ch)
	}
	return halfPass(fb, fnt, x, y, code, fg, 0, 0, bold, cx, cy, cw, ch)
}

// halfPass 畫半形字的一趟：shift ＝ 1 是立體邊那一趟（整個字往左移一行）。
func halfPass(fb *gfx.Framebuffer, fnt *font.Font, x, y int, code uint8,
	col uint8, shift, oy int, bold bool, cx, cy, cw, ch int) bool {
	g := fnt.Half(code)
	if g == nil {
		return false
	}
	rightHalf := x%(2*font.HalfW) == font.HalfW
	// 加粗是**整格 16 位元一起**做的（`docs/internals/25`），所以左半那個
	// 字最右一行加粗出來的那一行會落到右半的地盤上，右半那個字的則移出
	// 格外被丟掉。拆成兩次 8 位元去做的話兩邊都丟——症狀是粗體的半形字
	// 右邊少一豎（`BHCADTE` 的 E，`docs/re/111`）。
	// 這裡把一行擺成 9 位（第 gx 行 ＝ 第 8−gx 位），左半才留得住第 8 行。
	width := font.HalfW
	if bold && !rightHalf {
		width = font.HalfW + 1
	}
	rows := make([]uint16, len(g))
	for i, v := range g {
		w := uint16(v) << 1
		if bold {
			w |= w >> 1
		}
		rows[i] = w
	}
	for gy, bits := range rows {
		for gx := 0; gx < width; gx++ {
			if bits>>(font.HalfW-gx)&1 == 0 {
				continue
			}
			px, py := x+gx-shift, y+oy+gy
			if px < x && !rightHalf {
				continue
			}
			if px < cx || py < cy || px >= cx+cw || py >= cy+ch {
				continue
			}
			fb.Set(px, py, col)
		}
	}
	return true
}

// DrawCellClipped 畫字元平面上的一格。
//
// 原版依字碼的高位元組分路（`docs/internals/31`）：≥ 0x80 是一個全形字，
// < 0x80 是**兩個半形字塞在同一格**（高位元組在左、低位元組在右）。
// 「MIDI I/F」那種按鈕就是後者。
func DrawCellClipped(fb *gfx.Framebuffer, fnt *font.Font, x, y int, code uint16,
	fg, shadow uint8, shadowDY int, bold bool, cx, cy, cw, ch int) bool {
	if code>>8 >= 0x80 {
		return DrawGlyphClipped(fb, fnt, x, y, code, fg, shadow, shadowDY, bold, cx, cy, cw, ch)
	}
	start := fb.BeginGroup()
	// **整格先畫立體邊、再畫兩個字身。** 一個字一個字畫的話，右半那個字的
	// 立體邊會蓋掉左半那個字加粗溢出來的最後一行（`docs/re/111`）。
	if shadow != 0 {
		halfPass(fb, fnt, x, y, uint8(code>>8), shadow, 1, shadowDY, bold, cx, cy, cw, ch)
		halfPass(fb, fnt, x+font.HalfW, y, uint8(code), shadow, 1, shadowDY, bold, cx, cy, cw, ch)
	}
	a := halfPass(fb, fnt, x, y, uint8(code>>8), fg, 0, 0, bold, cx, cy, cw, ch)
	b := halfPass(fb, fnt, x+font.HalfW, y, uint8(code), fg, 0, 0, bold, cx, cy, cw, ch)
	// 保留原本「兩字陰影先畫」的 fallback，HD 則以兩個半形格重播。
	fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: x, Y: y, W: font.GlyphW, H: font.GlyphH,
		Code: code, Half: true, Colour: fg, Shadow: shadow, ShadowDY: shadowDY, Bold: bold, Clip: image.Rect(cx, cy, cx+cw, cy+ch)})
	return a && b
}
