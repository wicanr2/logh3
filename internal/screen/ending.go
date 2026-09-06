package screen

import (
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
)

// EndingImageY 是結局 CG 畫在畫面上的縱向位置。
//
// 640×400 的圖擺在 640×480 的畫布中間，上下各留 40 列——量自原版
// 畫面（`docs/internals/18-vo5.md`：`ENDT30.VO5` 的 plane 2／3 與
// 原版畫面在 (0,80) 逐位元組相同）。
const EndingImageY = 80

// Ending 把一張結局 CG 畫成整張畫布。
//
// 原版是四個 plane 一個一個寫上去的，中途擷取會看到「新圖的一半 ＋
// 舊圖的一半」；這裡畫的是寫完之後的狀態。
func Ending(img *asset.Image) *gfx.Framebuffer {
	fb := gfx.NewFramebuffer(0)
	if img == nil {
		return fb
	}
	for y := 0; y < img.H; y++ {
		for x := 0; x < img.W; x++ {
			fb.Set(x, EndingImageY+y, img.Pix[y*img.W+x])
		}
	}
	return fb
}
