package screen

import (
	"bytes"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
	"testing"
)

func TestAnimationPixelsPreserveOuterJournalAndFallback(t *testing.T) {
	fb := gfx.NewFramebuffer(0)
	fb.SurfaceRect(300, 200, 16, 15, 9)
	pixels := append([]byte(nil), fb.Pix...)
	pixels[57*gfx.Width+9] = 12
	DrawAnimationPixels(fb, pixels, []image.Rectangle{image.Rect(8, 56, 264, 158), image.Rect(376, 56, 632, 158)})
	if !bytes.Equal(pixels, fb.Pix) || !bytes.Equal(fb.Pix, gfx.ReplayIndexed(fb.Commands)) {
		t.Fatal("雙視窗原色或重播改變")
	}
	found := false
	for _, c := range fb.Commands {
		if c.Kind == gfx.DrawSurface {
			found = true
		}
	}
	if !found {
		t.Fatal("動畫抹掉外側HD命令")
	}
	pixels = append([]byte(nil), fb.Pix...)
	pixels[400*gfx.Width+600] = 7
	DrawAnimationPixels(fb, pixels, []image.Rectangle{image.Rect(8, 56, 264, 158)})
	if !bytes.Equal(pixels, fb.Pix) || !bytes.Equal(pixels, gfx.ReplayIndexed(fb.Commands)) {
		t.Fatal("界外變更未保留完整快照")
	}
	if len(fb.Commands) != 1 || fb.Commands[0].Kind != gfx.DrawImage {
		t.Fatal("界外變更不應猜裁切")
	}
}
