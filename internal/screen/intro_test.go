package screen

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
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

// TestIntroPaletteMatchesOracle 釘住「檔案值 × 4 才是 6 位元 DAC 值」。
//
// `BU16IBM.RGB` 是 48 bytes ＝ 16 組 RGB，裡面的值只有 0–15，直接當
// 顏色用會暗到看不見。收據 `introvram/pal.bin` 是 dosgolem 從執行期的
// DAC 倒出來的 8 位元值，這裡把「檔案值 ×4 → 6 位元 →`(v<<2)|(v>>4)`」
// 走一遍，48 個位元組要逐一相同。
func TestIntroPaletteMatchesOracle(t *testing.T) {
	orig := os.Getenv("LOGH3_ORIG")
	if orig == "" {
		orig = filepath.Join("..", "..", "org_game", "extracted")
	}
	rgb, err := os.ReadFile(filepath.Join(orig, "BU16IBM.RGB"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	want, err := os.ReadFile(filepath.Join("..", "..", "workplace", "oracle", "introvram", "pal.bin"))
	if err != nil {
		t.Skipf("沒有收據，跳過：%v", err)
	}
	pal, err := IntroPalette(rgb)
	if err != nil {
		t.Fatal(err)
	}
	for i := range pal {
		r, g, b := pal.RGB8(i)
		got := [3]uint8{r, g, b}
		for c := 0; c < 3; c++ {
			if got[c] != want[i*3+c] {
				t.Fatalf("第 %d 色第 %d 通道：remake %d，收據 %d", i, c, got[c], want[i*3+c])
			}
		}
	}
	// 十六階的亮度梯：G ＝ B ＝ 色號，偶數色號的 R 少一階（略偏青）。
	for i := range pal {
		wantG := uint8(i * IntroPaletteScale)
		wantR := wantG
		if i%2 == 0 && i > 0 {
			wantR = uint8((i - 1) * IntroPaletteScale)
		}
		if pal[i] != [3]uint8{wantR, wantG, wantG} {
			t.Fatalf("第 %d 色 %v，預期 %v", i, pal[i], [3]uint8{wantR, wantG, wantG})
		}
	}
}

// TestIntroFrameIntervalMatchesMeasurement 釘住量到的播放速度。
//
// `OPEN.EXE` 把 PIT 除數改成 0x0800，解碼器入口四個一組、前段每格相差
// 75 個計時器中斷（dosgolem 收據 `introfull`，`docs/re/283`）。
func TestIntroFrameIntervalMatchesMeasurement(t *testing.T) {
	want := time.Duration(IntroTicksPerFrame) * IntroPITDivisor * time.Second / IntroPITHz
	if IntroFrameInterval != want {
		t.Fatalf("一格 %v，算出來是 %v", IntroFrameInterval, want)
	}
	if IntroFrameInterval < 128*time.Millisecond || IntroFrameInterval > 130*time.Millisecond {
		t.Fatalf("一格 %v 不在 129 ms 附近", IntroFrameInterval)
	}
	if IntroScreenTop+IntroTopRow != (gfx.Height-asset.IntroPlaneRows)/2 {
		t.Fatalf("畫面沒有垂直置中：%d + %d", IntroScreenTop, IntroTopRow)
	}
}

func TestIntroFrameScheduleMatchesDosgolem(t *testing.T) {
	want := map[int]int{0: 75, 1: 75, 102: 336, 247: 2236,
		348: 75, 349: 2850, 350: 1798, 372: 77, 373: 1240,
		396: 77, 397: 1240, 420: 77}
	for frame, ticks := range want {
		if got := IntroTicksToFrame(frame); got != ticks {
			t.Errorf("播放格 %d 等待 %d ticks，預期 %d", frame, got, ticks)
		}
	}
}
