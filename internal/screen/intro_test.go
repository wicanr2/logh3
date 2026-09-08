package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
)

// TestIntroPagePlacesImageAtRow61 釘住「畫面畫在頁面第 61 列」。
//
// 那是量出來的（`docs/re/108`）：搬進 VRAM 的目的地是 `A0000 + 0x1310`，
// `0x1310 ＝ 61 × 80`；把執行期的四個 plane 組成色號去比收據 `introvram`
// 的畫面、逐一試 y 位移，61 也是明確的峰值（74.5% 對鄰近的 61–62%）。
func TestIntroPagePlacesImageAtRow61(t *testing.T) {
	orig := os.Getenv("LOGH3_ORIG")
	if orig == "" {
		orig = filepath.Join("..", "..", "org_game", "extracted")
	}
	data, err := os.ReadFile(filepath.Join(orig, "IBM16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	lengths, err := os.ReadFile(filepath.Join(orig, "IBMFLS16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	d, err := asset.NewIntroDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	// 走到有內容的一格再看（前面幾格幾乎全黑）。
	for i := 0; i < 200; i++ {
		if ok, err := d.Next(); err != nil || !ok {
			t.Fatalf("解到第 %d 格停了：%v", i, err)
		}
	}
	page := IntroPage(d)
	if page.W != IntroPageCols || page.H != IntroPageRows {
		t.Fatalf("頁面 %d×%d，應該是 %d×%d", page.W, page.H, IntroPageCols, IntroPageRows)
	}
	img := d.Pixels()
	nonzero := 0
	for y := 0; y < img.H; y++ {
		for x := 0; x < img.W; x++ {
			if page.Pix[(IntroTopRow+y)*IntroPageCols+x] != img.Pix[y*img.W+x] {
				t.Fatalf("第 %d 列第 %d 行對不上", IntroTopRow+y, x)
			}
			if img.Pix[y*img.W+x] != 0 {
				nonzero++
			}
		}
	}
	if nonzero == 0 {
		t.Fatal("第 200 格是全黑的，這條測試就沒有意義")
	}
	for y := 0; y < IntroTopRow; y++ {
		for x := 0; x < IntroPageCols; x++ {
			if page.Pix[y*IntroPageCols+x] != 0 {
				t.Fatalf("上緣第 %d 列不是 0", y)
			}
		}
	}
	t.Logf("第 200 格：%d 個非零像素，畫在第 %d–%d 列",
		nonzero, IntroTopRow, IntroTopRow+img.H-1)
}
