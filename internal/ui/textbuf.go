package ui

// 視窗的文字緩衝區與跳脫碼引擎（原版 `3617:0099`，`docs/re/141`）。
//
// 原版所有的字都經過同一支「印字進視窗」（`3617:003A`，611 個呼叫端）：
// 它把格式字串交給一支 printf，輸出回呼一個位元組一個位元組地寫進視窗
// 自己的內容緩衝區——**一格 4 bytes**（字碼 word ＋ 顏色 byte ＋ 旗標 byte）。
// 真正畫到畫面上是另一趟（`34CB:01A4` 把內容格搬進畫面格表，
// `2376:040E` 再畫）。
//
// 所以文字排版與繪製是分開的兩件事，這裡做的是前一半。

// ContentCell 是內容緩衝區的一格。
type ContentCell struct {
	Code   uint16 // 高位元組 >= 0x80 ＝ 全形；0xFF ＝ 外框圖塊；否則兩個半形
	Colour uint8  // 低四位字色、高四位底色
	Flags  uint8  // 裝飾位元，見 CellUnderline 等
}

// 視窗 `[+1A]` 的兩個位元。bit0 是「結構是建構函式自己配的」，與文字無關。
const (
	AttrWiden  = 0x80 // 半形（0x20–0x7E）轉成全形，一個字佔一格
	AttrScroll = 0x40 // 寫到最後一列還要換行時把內容往上捲一列
)

// 跳脫碼（`ds:6BDC` 那張表；表上沒有的碼一律忽略）。
const (
	EscGoto     = 0x01 // `[欄,列` 設游標
	EscMove     = 0x02 // `[Δ欄,Δ列` 相對移動
	EscColour   = 0x03 // `[色` 設顏色
	EscWidenOn  = 0x06 // 半形轉全形 開
	EscWidenOff = 0x07 // 關
	EscFlags    = 0x08 // `[旗標` 設裝飾位元
	EscNewline  = 0x0A // 換行
	EscClear    = 0x0C // 清空並把游標歸零
)

// TextBuffer 是一個視窗的內容緩衝區。
type TextBuffer struct {
	Cols, Rows int
	Cells      []ContentCell

	Colour uint8 // 視窗 [+1]
	Flags  uint8 // 視窗 [+2]
	Attr   uint8 // 視窗 [+1A]
	Col    int   // 視窗 [+1C]
	Row    int   // 視窗 [+1D]

	// half 記著「上一個位元組是雙位元組字的前半」（視窗 [+1B]）。
	half bool

	// 跳脫碼的狀態機（`ds:F124`／`ds:F125`）。0 ＝ 沒有在等參數。
	state uint8
	arg   uint8
}

// NewTextBuffer 造一個 cols × rows 的緩衝區，並照 `35AA:0068` 清成空白。
func NewTextBuffer(cols, rows int, colour uint8) *TextBuffer {
	b := &TextBuffer{Cols: cols, Rows: rows, Colour: colour,
		Cells: make([]ContentCell, cols*rows)}
	b.Clear()
	return b
}

// Clear 把每一格清成兩個半形空白 ＋ 視窗顏色，游標歸零（`35AA:0068` ＋
// 跳脫碼 `0x0C`）。
func (b *TextBuffer) Clear() {
	for i := range b.Cells {
		b.Cells[i] = ContentCell{Code: 0x2020, Colour: b.Colour}
	}
	b.Col, b.Row, b.half = 0, 0, false
}

// At 回第 (col, row) 格；超出範圍回 nil。
func (b *TextBuffer) At(col, row int) *ContentCell {
	if col < 0 || row < 0 || col >= b.Cols || row >= b.Rows {
		return nil
	}
	return &b.Cells[row*b.Cols+col]
}

// Write 送一串位元組進去。字串是原版資料段的 Big5，控制碼夾在裡面。
func (b *TextBuffer) Write(raw []byte) {
	for _, c := range raw {
		b.PutByte(c)
	}
}

// PutByte 送一個位元組進去，完全照 `3617:0099` 的分派。（不叫 WriteByte
// 是為了不撞上 io.ByteWriter 那個要回 error 的簽章。）
func (b *TextBuffer) PutByte(c byte) {
	if b.state != 0 {
		b.escapeArg(c)
		return
	}
	if c < 0x20 {
		b.control(c)
		return
	}
	b.put(c)
}

// escapeArg 收跳脫碼的十進位參數（`3617:038A`）：`[` 跳過、數字累加、
// 其他字元結束並提交。**結束的那個字元被吃掉，不會印出來。**
func (b *TextBuffer) escapeArg(c byte) {
	if c >= '0' && c <= '9' {
		b.arg = b.arg*10 + (c - '0')
		return
	}
	if c == '[' {
		return
	}
	v, st := b.arg, b.state
	b.arg = 0
	switch st {
	case 1: // 欄，接著還要收列
		b.Col = int(v)
		b.state = 4
		return
	case 2: // Δ欄，接著還要收 Δ列
		b.Col += int(v)
		b.state = 5
		return
	case 3:
		b.Colour = v
	case 4:
		b.Row = int(v)
	case 5:
		b.Row += int(v)
	case 8:
		b.Flags = v
	}
	b.state = 0
}

func (b *TextBuffer) control(c byte) {
	switch c {
	case EscGoto:
		b.state, b.arg = 1, 0
	case EscMove:
		b.state, b.arg = 2, 0
	case EscColour:
		b.state, b.arg = 3, 0
	case EscFlags:
		b.state, b.arg = 8, 0
	case EscWidenOn:
		b.Attr |= AttrWiden
	case EscWidenOff:
		b.Attr &^= AttrWiden
	case EscNewline:
		b.newline()
	case EscClear:
		b.Clear()
	}
	// 表上沒有的碼原版也是什麼都不做。
}

// newline 換行（`3617:03BE`）。寫到最後一列還要換行時，`AttrScroll`
// 決定是把內容往上捲一列還是停在原地。
func (b *TextBuffer) newline() {
	b.Col = 0
	b.Row++
	if b.Row < b.Rows {
		return
	}
	if b.Attr&AttrScroll == 0 {
		b.Row = b.Rows - 1
		return
	}
	copy(b.Cells, b.Cells[b.Cols:])
	last := (b.Rows - 1) * b.Cols
	for i := last; i < len(b.Cells); i++ {
		b.Cells[i] = ContentCell{Code: 0x2020, Colour: b.Colour}
	}
	b.Row = b.Rows - 1
}

// put 寫一個可見位元組（`3617:02A5`）。
func (b *TextBuffer) put(c byte) {
	if b.Col >= b.Cols {
		b.half = false
		b.newline()
	}
	cell := b.At(b.Col, b.Row)
	if cell == nil {
		return
	}
	if b.half {
		// 雙位元組字的後半：OR 進低位元組，這一格才算寫完。
		cell.Code |= uint16(c)
		b.half = false
		b.Col++
		return
	}
	cell.Colour, cell.Flags = b.Colour, b.Flags
	if b.Attr&AttrWiden != 0 && c >= 0x20 && c <= 0x7E {
		// 半形轉全形（原版查 `3B6B:0002` 的表）。**表還沒抄出來**，
		// 所以這裡先原樣放進高位元組再補一個空白——與原版不同，
		// 用到的地方要先量收據（`docs/re/141`）。
		cell.Code = uint16(c)<<8 | 0x20
		b.Col++
		return
	}
	cell.Code = uint16(c) << 8
	b.half = true
}
