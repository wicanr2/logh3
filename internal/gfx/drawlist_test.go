package gfx

import (
	"bytes"
	"testing"
)

func TestDrawingJournalPreservesIndexedWrites(t *testing.T) {
	f := NewFramebuffer(3)
	f.Set(0, 0, 7)
	f.FillRect(-2, -2, 5, 5, 9)
	start := f.BeginGroup()
	f.SurfaceRect(8, 7, 30, 20, 5)
	nested := f.BeginGroup()
	f.Set(11, 11, 15)
	f.Set(12, 12, 7)
	f.EndGroup(nested, DrawCommand{Kind: DrawGlyph})
	f.EndGroup(start, DrawCommand{Kind: DrawPanel})
	f.Set(10, 10, 5) // 同色寫入亦不能消失或改畫序。
	f.WriteRow(8, 8, []uint8{0, 1, 2, 3})
	source := []uint8{4, 5, 6, 7, 8, 9}
	f.DrawSource("test", -1, 15, 2, 2, 3, 2, source)
	source[1] = 15 // 紀錄綁本次影像，不能隨來源變動。
	if got := ReplayIndexed(f.Commands); !bytes.Equal(got, f.Pix) {
		t.Fatal("原色重播與畫布不同")
	}
	f.CopyPixels(bytes.Repeat([]uint8{8}, Width*Height))
	f.Set(639, 479, 15)
	if got := ReplayIndexed(f.Commands); !bytes.Equal(got, f.Pix) {
		t.Fatal("整張取代仍保留舊命令")
	}
	f.Clear(2)
	if got := ReplayIndexed(f.Commands); !bytes.Equal(got, f.Pix) {
		t.Fatal("清空仍保留舊命令")
	}
}
