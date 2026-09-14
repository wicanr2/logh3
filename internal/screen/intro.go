package screen

import (
	"fmt"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 開頭動畫的版面（`docs/internals/19`、`docs/re/108`）。
//
// `OPEN.EXE` 切到 mode 10h（640×350×16），一頁 350 列 × 80 個位元組。
// 解出來的畫面只有 227 列，搬進 VRAM 的目的地是 `A0000 + 0x1310 + 頁`
// ——`0x1310 ＝ 61 × 80`，所以畫面是**垂直置中**在頁面上。
// 61 這個值是量出來的：把執行期的四個 plane 組成色號去比收據
// `introvram` 的畫面，逐一試 y 位移，61 是明確的峰值。
const (
	IntroPageRows = 350 // mode 10h 一頁幾列
	IntroPageCols = 640
	IntroTopRow   = 61 // 227 列的畫面畫在第幾列
)

// IntroPage 把解出來的一格畫成一整頁 640×350 的色號圖。
//
// 頁面其餘的列留 0——原版那兩塊也是 0（收據 `intro52` 的 `A0000` 兩頁
// 各只有兩、三千個非零位元組，`docs/re/104`）。
func IntroPage(d interface{ Pixels() *asset.Image }) *asset.Image {
	page := &asset.Image{
		W: IntroPageCols, H: IntroPageRows,
		Pix: make([]uint8, IntroPageCols*IntroPageRows),
	}
	if d == nil {
		return page
	}
	img := d.Pixels()
	for y := 0; y < img.H && IntroTopRow+y < IntroPageRows; y++ {
		copy(page.Pix[(IntroTopRow+y)*IntroPageCols:(IntroTopRow+y+1)*IntroPageCols],
			img.Pix[y*img.W:(y+1)*img.W])
	}
	return page
}

// 開頭動畫的調色盤（`BU16IBM.RGB`，48 bytes ＝ 16 組 RGB）。
//
// 檔案裡的值只有 0–15，**乘 4 才是 VGA DAC 的 6 位元值**。收據
// `introvram/pal.bin` 的 48 個位元組逐一相符（dosgolem 把 6 位元展成
// 8 位元的算法是 `(v<<2)|(v>>4)`：檔案值 12 → DAC 48 → 195）。
// 整份是十六階的亮度梯：G ＝ B ＝ 色號，偶數色號的 R 少一階
// （色號 2 是 `01 02 02`），所以是略偏青的近灰階。
const IntroPaletteScale = 4

// IntroPalette 把 `BU16IBM.RGB` 轉成 6 位元調色盤。
func IntroPalette(b []byte) (gfx.Palette, error) {
	var p gfx.Palette
	if len(b) < gfx.Colors*3 {
		return p, fmt.Errorf("BU16IBM.RGB 只有 %d bytes，不足 %d", len(b), gfx.Colors*3)
	}
	for i := 0; i < gfx.Colors; i++ {
		for c := 0; c < 3; c++ {
			v := int(b[i*3+c]) * IntroPaletteScale
			if v > 63 {
				return p, fmt.Errorf("第 %d 色第 %d 通道 %d 超出 6 位元", i, c, v)
			}
			p[i][c] = uint8(v)
		}
	}
	return p, nil
}

// 開頭動畫的播放速度。
//
// `OPEN.EXE` 進場時把 PIT 的除數改成 0x0800（`mov dx,0800h` →
// `out 43h,36h` → 分兩次 `out 40h`，映像位移 0xD90），離場時寫回 0
// ＝ 65536。所以動畫期間計時器是 1193182 / 2048 ≈ 582.6 Hz。
//
// 一格要解四個 plane，量到的解碼器入口（`041F:024B`）四個一組；
// 組與組的間隔在修正後的 dosgolem `introfull` 收據是 75 個計時器中斷；
// 尾段是 77 個（完整時程見 `docs/re/283`）。
// 換算 75 × 2048 / 1193182 ≈ 128.7 ms，約每秒 7.8 格。
const (
	IntroPITHz         = 1193182 // PIT 的輸入頻率
	IntroPITDivisor    = 0x0800  // OPEN.EXE 設的除數
	IntroTicksPerFrame = 75      // 前段一格幾個計時器中斷
	IntroTailTicks     = 77      // 尾段一格幾個計時器中斷
)

// IntroFrameInterval 是前段一格的長度（約 128.7 ms）。
const IntroFrameInterval = time.Duration(IntroTicksPerFrame) *
	IntroPITDivisor * time.Second / IntroPITHz

// IntroTicksToFrame 回「切到指定播放格之前」原版等待的 PIT 中斷數。
// 播放格是玩家可見序列的索引：0–348 是前段，349–420 是尾段三趟。
// 336、2236、2850、1798、1240 是 dosgolem `introfull` 收據量出的停頓；
// 原版儲存這些值的位置尚未解出，因此明確保留為量測常數。
func IntroTicksToFrame(frame int) int {
	switch {
	case frame == 102:
		return 336
	case frame == 247:
		return 2236
	case frame >= 349:
		pass := (frame - 349) / 24
		within := (frame - 349) % 24
		if within == 0 {
			if pass == 0 {
				return 2850
			}
			return 1240
		}
		if pass == 0 && within == 1 {
			return 1798
		}
		return IntroTailTicks
	default:
		return IntroTicksPerFrame
	}
}

// IntroFrameIntervalAt 是切到指定播放格之前的等待時間。
func IntroFrameIntervalAt(frame int) time.Duration {
	return time.Duration(IntroTicksToFrame(frame)) * IntroPITDivisor * time.Second / IntroPITHz
}

// IntroScreenTop 是 350 列的頁面畫在 640×480 畫布上的第幾列。
//
// 原版這一段是 mode 10h（640×350），remake 的畫布固定 640×480
// （mode 12h），兩者對不起來——**這是 remake 自己的決定**：把整頁
// 垂直置中。(480−350)/2 ＝ 65，加上頁面內的 61 列剛好是 126，
// 與「227 列的畫面在 480 列裡置中」((480−227)/2 ＝ 126) 相同。
const IntroScreenTop = (gfx.Height - IntroPageRows) / 2

// IntroScreen 把一格畫成 640×480 的畫面。
func IntroScreen(page *asset.Image) *gfx.Framebuffer {
	fb := gfx.NewFramebuffer(0)
	if page == nil {
		return fb
	}
	for y := 0; y < page.H; y++ {
		for x := 0; x < page.W; x++ {
			fb.Set(x, IntroScreenTop+y, page.Pix[y*page.W+x])
		}
	}
	return fb
}
