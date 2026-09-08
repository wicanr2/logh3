package screen

import (
	"github.com/wicanr2/logh3/internal/asset"
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
func IntroPage(d *asset.IntroDecoder) *asset.Image {
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
