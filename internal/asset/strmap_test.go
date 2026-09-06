package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// 星圖底圖的高兩個位元要跟原版主畫面一致。
//
// 低兩個位元是遊戲依勢力與治安重畫的（`docs/re/17`），所以這裡只驗線稿；
// 剩下的差異是畫在上面的東西（游標、覆蓋物）。
func TestSTRMAPHighPlanesMatchOriginal(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(origDir(t), "STRMAP.BIT"))
	if err != nil {
		t.Skip("沒有原版 STRMAP.BIT，跳過")
	}
	img, err := DecodeSTRMAP(raw)
	if err != nil {
		t.Fatal(err)
	}
	if img.W != 320 || img.H != 300 {
		t.Fatalf("尺寸 %d×%d，預期 320×300", img.W, img.H)
	}
	screen, err := os.ReadFile("../../workplace/oracle/main1/screen.bin")
	if err != nil {
		t.Skip("沒有 main1 收據，跳過")
	}
	const screenW, top = 640, 40
	same := 0
	for y := 0; y < img.H; y++ {
		for x := 0; x < img.W; x++ {
			if img.Pix[y*img.W+x]>>2 == screen[(top+y)*screenW+x]>>2 {
				same++
			}
		}
	}
	total := img.W * img.H
	if same*100 < total*90 {
		t.Errorf("高兩個位元只有 %d / %d 相同（%.1f%%），預期至少 90%%",
			same, total, float64(same)*100/float64(total))
	}
	t.Logf("高兩個位元 %d / %d 相同（%.1f%%）", same, total,
		float64(same)*100/float64(total))
}
