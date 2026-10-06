package asset

import "github.com/wicanr2/logh3/internal/gfx"

// Draw 畫出原圖，同時保留可選 HD 圖片的原索引與座標。
func (img *Image) Draw(fb *gfx.Framebuffer, x, y, w, h int) {
	if img == nil {
		return
	}
	w, h = min(w, img.W), min(h, img.H)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			fb.Set(x+i, y+j, img.Pix[j*img.W+i])
		}
	}
	fb.MarkArtwork(img.ArtworkKey, x, y, w, h, img.W, img.H)
}
