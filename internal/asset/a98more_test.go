package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// ANI_PNL1／ANI_PNL2 是什麼：先看解出來多大、非空的部分落在哪、
// 以及與收據 s3ai6 的畫面吻合到什麼程度。
func TestANIPanel1And2(t *testing.T) {
	scr, err := os.ReadFile(filepath.Join("../..", "workplace/oracle/s3ai6/screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("沒有收據 s3ai6，跳過")
	}
	for _, name := range []string{"ANI_PNL1.A98", "ANI_PNL2.A98"} {
		data, err := os.ReadFile(filepath.Join(origDir(t), name))
		if err != nil {
			t.Skipf("讀不到 %s", name)
		}
		img, err := DecodeAMGImage(data)
		if err != nil {
			t.Errorf("%s 解不開：%v", name, err)
			continue
		}
		// 非空（色號不是 0）的 bounding box。
		x0, y0, x1, y1 := img.W, img.H, -1, -1
		for y := 0; y < img.H; y++ {
			for x := 0; x < img.W; x++ {
				if img.Pix[y*img.W+x] != 0 {
					if x < x0 {
						x0 = x
					}
					if y < y0 {
						y0 = y
					}
					if x > x1 {
						x1 = x
					}
					if y > y1 {
						y1 = y
					}
				}
			}
		}
		t.Logf("%s：%d×%d，非空 x %d..%d y %d..%d", name, img.W, img.H, x0, x1, y0, y1)
		// 在畫面上找最合的位置（只比非空的那一塊）。
		best, bestBad := -1, 1<<30
		for off := 0; off+img.H <= 480; off++ {
			bad := 0
			for y := y0; y <= y1; y++ {
				for x := x0; x <= x1; x++ {
					if img.Pix[y*img.W+x] != scr[(off+y)*640+x] {
						bad++
					}
				}
			}
			if bad < bestBad {
				best, bestBad = off, bad
			}
		}
		n := (x1 - x0 + 1) * (y1 - y0 + 1)
		t.Logf("%s：最合的起點 y=%d，不同 %d／%d（%.1f%%）",
			name, best, bestBad, n, 100*float64(bestBad)/float64(n))
		// 畫在 y=40 時，狀態欄那一塊（畫面 464..623 × 220..419）像不像。
		bad := 0
		for y := 220; y < 420; y++ {
			for x := 464; x < 624; x++ {
				if img.Pix[(y-40)*img.W+x] != scr[y*640+x] {
					bad++
				}
			}
		}
		t.Logf("%s：畫在 y=40 時狀態欄那塊不同 %d／32000（%.1f%%）",
			name, bad, 100*float64(bad)/32000)
		for y := 221; y < 228; y++ {
			line := ""
			for x := 460; x < 501; x++ {
				line += string("0123456789ABCDEF"[img.Pix[(y-40)*img.W+x]])
			}
			t.Logf("  %s y=%d %s", name, y, line)
		}
	}
}
