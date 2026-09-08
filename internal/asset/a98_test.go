package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// TestANIPanelIsAMG `ANI_PNL0.A98` 是戰鬥畫面的面板：整個檔就是一張
// `.AMG` 影像（`docs/internals/37`），解出來要與收據 `s3ai6` 的畫面
// 逐像素相同。
func TestANIPanelIsAMG(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(origDir(t), "ANI_PNL0.A98"))
	if err != nil {
		t.Skipf("讀不到 ANI_PNL0.A98：%v", err)
	}
	img, err := DecodeAMGImage(data)
	if err != nil {
		t.Fatal(err)
	}
	if img.W != 640 || img.H != 400 {
		t.Fatalf("解出 %d×%d，檔頭寫的是 640×400", img.W, img.H)
	}
	scr, err := os.ReadFile(filepath.Join("../..", "workplace/oracle/s3ai6/screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("沒有收據 s3ai6，跳過")
	}
	// 面板畫在畫面的哪一列還沒定，掃一遍找最合的那一列。
	best, bestBad := -1, 1<<30
	for off := 0; off+img.H <= 480; off++ {
		bad := 0
		for y := 0; y < img.H; y++ {
			for x := 0; x < 640; x++ {
				if img.Pix[y*640+x] != scr[(off+y)*640+x] {
					bad++
				}
			}
			if bad >= bestBad {
				break
			}
		}
		if bad < bestBad {
			best, bestBad = off, bad
		}
	}
	t.Logf("最合的起點 y=%d，不同的像素 %d／%d", best, bestBad, img.H*640)
	if best != 40 {
		t.Fatalf("最合的起點是 y=%d，預期 40", best)
	}

	// 面板是**底圖**，上面還會疊三塊：戰場、小地圖、狀態欄
	// （`docs/internals/34`）。把那三塊遮掉，剩下的應該完全一樣。
	masked := [][4]int{
		{16, 60, 463, 419},  // 戰場
		{464, 40, 639, 219}, // 小地圖那一區
		{464, 220, 623, 419},// 狀態欄
	}
	bad, total := 0, 0
	for y := 0; y < img.H; y++ {
		for x := 0; x < 640; x++ {
			sy := best + y
			skip := false
			for _, m := range masked {
				if x >= m[0] && x <= m[2] && sy >= m[1] && sy <= m[3] {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			total++
			if img.Pix[y*640+x] != scr[sy*640+x] {
				bad++
			}
		}
	}
	t.Logf("遮掉疊在上面那三塊之後：不同 %d／%d", bad, total)
	if bad != 0 {
		t.Fatalf("外框有 %d 個像素對不上", bad)
	}
}
