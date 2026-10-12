package asset

import "github.com/wicanr2/logh3/internal/gfx"

// Draw 畫出原圖，同時保留可選 HD 圖片的原索引與座標。
func (img *Image) Draw(fb *gfx.Framebuffer, x, y, w, h int) {
	if img == nil {
		return
	}
	w, h = min(w, img.W), min(h, img.H)
	start := -1
	if img.StarMap != nil {
		start = fb.BeginGroup()
	}
	fb.DrawSource(img.ArtworkKey, x, y, w, h, img.W, img.H, img.Pix)
	fb.MarkArtwork(img.ArtworkKey, x, y, w, h, img.W, img.H)
	if start >= 0 {
		m := img.StarMap
		fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawStarMap, X: x, Y: y, W: w, H: h, SourceW: img.W, SourceH: img.H, Source: img.Pix, Base: m.Base, Regions: m.Regions, Colours: m.Colours})
	}
}
