package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 跳脫碼的狀態機（`docs/re/141`）：`\x01[欄,列` 設游標、`\x03[色` 換色、
// `\x0c` 清空、`\x0a` 換行。**結束參數的那個字元被吃掉，不會印出來。**
func TestTextBufferEscapes(t *testing.T) {
	b := NewTextBuffer(8, 3, 0x97)
	b.Write([]byte("\x01[2,1,"))
	if b.Col != 2 || b.Row != 1 {
		t.Fatalf("游標在 (%d,%d)，該是 (2,1)", b.Col, b.Row)
	}
	b.Write([]byte("\x02[1,1,"))
	if b.Col != 3 || b.Row != 2 {
		t.Fatalf("相對移動之後在 (%d,%d)，該是 (3,2)", b.Col, b.Row)
	}
	b.Write([]byte("\x03[135,"))
	if b.Colour != 135 {
		t.Fatalf("顏色是 %d，該是 135", b.Colour)
	}
	b.Write([]byte("\x08[8,"))
	if b.Flags != 8 {
		t.Fatalf("旗標是 %d，該是 8", b.Flags)
	}
	b.Write([]byte("\x0c"))
	if b.Col != 0 || b.Row != 0 {
		t.Error("`\\x0c` 該把游標歸零")
	}
	for i, c := range b.Cells {
		if c.Code != 0x2020 {
			t.Fatalf("第 %d 格清完是 %04X，該是 2020", i, c.Code)
		}
	}

	// 全形字兩個位元組佔一格；換行回到第 0 欄。
	b = NewTextBuffer(4, 2, 0x97)
	b.Write([]byte{0xB3, 0x6F, 0xAC, 0x4F, 0x0A, 0xB5, 0x4C})
	if got := b.At(0, 0).Code; got != 0xB36F {
		t.Errorf("(0,0) 是 %04X，該是 B36F", got)
	}
	if got := b.At(1, 0).Code; got != 0xAC4F {
		t.Errorf("(1,0) 是 %04X，該是 AC4F", got)
	}
	if got := b.At(0, 1).Code; got != 0xB54C {
		t.Errorf("(0,1) 是 %04X，該是 B54C", got)
	}

	// 寫滿一列會自己折到下一列（`3617:02A5` 先比 [+1C] 與 [+0B]）。
	b = NewTextBuffer(2, 2, 0x97)
	b.Write([]byte{0xB3, 0x6F, 0xAC, 0x4F, 0xB5, 0x4C})
	if got := b.At(0, 1).Code; got != 0xB54C {
		t.Errorf("折行之後 (0,1) 是 %04X，該是 B54C", got)
	}
}

// 捲動：`AttrScroll` 沒設就列回 0、整窗清空（`3617:049C` → `35AA:0068`，`docs/re/336` §七），
// 設了就把內容往上搬一列。
func TestTextBufferScroll(t *testing.T) {
	mk := func(attr uint8) *TextBuffer {
		b := NewTextBuffer(1, 2, 0x97)
		b.Attr = attr
		b.Write([]byte{0xB3, 0x6F, 0x0A, 0xAC, 0x4F, 0x0A, 0xB5, 0x4C})
		return b
	}
	if b := mk(0); b.At(0, 0).Code != 0xB54C || b.At(0, 1).Code != 0x2020 {
		t.Errorf("不捲動時是 %04X／%04X，該是 B54C／2020",
			b.At(0, 0).Code, b.At(0, 1).Code)
	}
	if b := mk(AttrScroll); b.At(0, 0).Code != 0xAC4F || b.At(0, 1).Code != 0xB54C {
		t.Errorf("捲動時是 %04X／%04X，該是 AC4F／B54C",
			b.At(0, 0).Code, b.At(0, 1).Code)
	}
}

// 收據 `s3msg`：**整個訊息框走通用的文字緩衝區重畫一次**，逐位元組比。
// 這是這一套引擎的驗收——訊息框是目前唯一有同狀態收據的視窗。
func TestTextBufferRedrawsMessageBox(t *testing.T) {
	root := "../.."
	blk, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skipf("缺原版 FRAME.BLK：%v", err)
	}
	tiles, err := LoadFrameTiles(blk)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺原版字型：%v", err)
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3msg/screen.bin"))
	if err != nil {
		t.Skipf("缺收據 s3msg：%v", err)
	}

	const cellX, cellY, cols, rows = 4, 7, 20, 4
	fb := gfx.NewFramebuffer(0)
	(TileFrame{Tiles: tiles, Style: PanelFrameStyle, Fill: 9}).
		Draw(fb, cellX, cellY, cols+2, rows+2)

	b := NewTextBuffer(cols, rows, 0x97)
	// `ds:4A24`「這是無法行動的地方。」後面原版還有兩個半形空白。
	b.Write([]byte{
		0xb3, 0x6f, 0xac, 0x4f, 0xb5, 0x4c, 0xaa, 0x6b, 0xa6, 0xe6,
		0xb0, 0xca, 0xaa, 0xba, 0xa6, 0x61, 0xa4, 0xe8, 0xa1, 0x43,
		0x20, 0x20,
	})
	// 按鈕：`37A9:004C` 直接寫內容格（不走文字引擎）。
	b.At(16, 3).Flags |= CellButtonLeft
	for i, code := range []uint16{0xBD54, 0xBB7B} {
		c := b.At(17+i, 3)
		c.Code, c.Colour, c.Flags = code, 0x87, CellButtonBody
	}
	b.At(19, 3).Flags |= CellButtonRight

	if m := DrawTextBuffer(fb, fnt, tiles, b, cellX+1, cellY+1); len(m) != 0 {
		t.Fatalf("有缺字：%04X", m)
	}

	var diff int
	for r := 0; r < rows+2; r++ {
		for c := 0; c < cols+2; c++ {
			for y := 0; y < FrameTileH; y++ {
				for x := 0; x < FrameTileW; x++ {
					px := (cellX+c)*FrameTileW + x
					py := (cellY+r)*FrameTileH + FrameCellOriginY + y
					if fb.At(px, py) != want[py*gfx.Width+px] {
						diff++
					}
				}
			}
		}
	}
	if diff != 0 {
		t.Errorf("走文字緩衝區重畫的訊息框有 %d 個像素與收據不同", diff)
	}
}

// 半形轉全形（`3B6B:0025`）：數字／英文用加法，標點查表。
// **表是線性找、第一個中的算數**——`0x60` 在表上出現兩次，
// 所以反引號變的是「，」（A141）不是「‘」（A1A5）。
func TestWidenASCII(t *testing.T) {
	cases := []struct {
		in   byte
		want uint16
	}{
		{'0', 0xA2AF}, {'9', 0xA2B8},
		{'a', 0xA2E9}, {'z', 0xA302},
		{'A', 0xA2CF}, {'Z', 0xA2E8},
		{' ', 0xA140}, {'!', 0xA149}, {'?', 0xA148},
		{'`', 0xA141}, // 表上兩筆，取前面那一筆
		{'"', 0xA149}, // 原版對到「！」，照抄不修正
		{0x01, 0x0001},
	}
	for _, c := range cases {
		if got := WidenASCII(c.in); got != c.want {
			t.Errorf("WidenASCII(%02X) ＝ %04X，該是 %04X", c.in, got, c.want)
		}
	}

	// 開了 AttrWiden 之後 ASCII 一個字佔一格。
	b := NewTextBuffer(4, 1, 0x97)
	b.Write([]byte("\x06AB"))
	if got := b.At(0, 0).Code; got != 0xA2CF {
		t.Errorf("(0,0) 是 %04X，該是 A2CF", got)
	}
	if got := b.At(1, 0).Code; got != 0xA2D0 {
		t.Errorf("(1,0) 是 %04X，該是 A2D0", got)
	}
	// 關掉之後回到「兩個半形塞一格」。
	b.Write([]byte("\x07CD"))
	if got := b.At(2, 0).Code; got != 0x4344 {
		t.Errorf("(2,0) 是 %04X，該是 4344", got)
	}
}

// 格式化只做原版格式字串用得到的那幾種；沒見過的轉換原樣輸出。
//
// **格式字串是原版資料段的 Big5 位元組**，不是 Go 原始碼裡的 UTF-8，
// 所以這裡一律用位元組字面值。
func TestTextBufferPrintf(t *testing.T) {
	bytesOf := func(b *TextBuffer, n int) []byte {
		var out []byte
		for i := 0; i < n; i++ {
			out = append(out, byte(b.Cells[i].Code>>8), byte(b.Cells[i].Code))
		}
		return out
	}
	// `ds:3F36`「運輸艦      %4u」：3 個全形 ＋ 6 個空白 ＋ 4 欄的數字
	// ＝ 剛好 8 格，也就是視窗一整列。
	b := NewTextBuffer(8, 2, 0x97)
	b.Printf([]byte{0xB9, 0x42, 0xBF, 0xE9, 0xC4, 0xA5, ' ', ' ', ' ', ' ', ' ', ' ',
		'%', '4', 'u'}, 12)
	want := []byte{0xB9, 0x42, 0xBF, 0xE9, 0xC4, 0xA5, ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', '1', '2'}
	if got := bytesOf(b, 8); string(got) != string(want) {
		t.Errorf("得到 % X\n該是   % X", got, want)
	}

	// 靠左、補零、`%%`；`l` 只是取值寬度，輸出一樣。
	b = NewTextBuffer(8, 2, 0x97)
	b.Printf([]byte("%-8s|%6lu"), []byte("abc"), 1234)
	// 15 個位元組 ＝ 7 格半，最後一格只寫了高位元組（原版也是這樣）。
	if got := bytesOf(b, 8); string(got) != "abc     |  1234\x00" {
		t.Errorf("%q", got)
	}
	b = NewTextBuffer(4, 1, 0x97)
	b.Printf([]byte("%03d%%"), 7)
	if got := bytesOf(b, 2); string(got) != "007%" {
		t.Errorf("%q", got)
	}
}
