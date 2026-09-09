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

// 捲動：`AttrScroll` 沒設就停在最後一列，設了就把內容往上搬一列。
func TestTextBufferScroll(t *testing.T) {
	mk := func(attr uint8) *TextBuffer {
		b := NewTextBuffer(1, 2, 0x97)
		b.Attr = attr
		b.Write([]byte{0xB3, 0x6F, 0x0A, 0xAC, 0x4F, 0x0A, 0xB5, 0x4C})
		return b
	}
	if b := mk(0); b.At(0, 0).Code != 0xB36F || b.At(0, 1).Code != 0xB54C {
		t.Errorf("不捲動時是 %04X／%04X，該是 B36F／B54C",
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
