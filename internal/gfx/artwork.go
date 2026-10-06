package gfx

// Artwork 記錄原圖座標與後續繪製留下的可見像素。
// SourceW/H 是未裁切原圖尺寸，W/H 是本次畫出的部分。
type Artwork struct {
	Key              string
	X, Y, W, H       int
	SourceW, SourceH int
	Visible          []uint8
}

// MarkArtwork 在原圖畫完後附加 HD 索引；之後的 Set 會遮擋這個圖層。
func (f *Framebuffer) MarkArtwork(key string, x, y, w, h, sourceW, sourceH int) {
	if key == "" || w <= 0 || h <= 0 || sourceW <= 0 || sourceH <= 0 {
		return
	}
	w, h = min(w, sourceW), min(h, sourceH)
	a := Artwork{Key: key, X: x, Y: y, W: w, H: h, SourceW: sourceW, SourceH: sourceH,
		Visible: make([]uint8, w*h)}
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			if x+i >= 0 && x+i < Width && y+j >= 0 && y+j < Height {
				a.Visible[j*w+i] = 1
			}
		}
	}
	f.Artwork = append(f.Artwork, a)
}

// WriteRow 沿 Set 的繪製順序寫一段色號，保留 HD 遮擋語意。
func (f *Framebuffer) WriteRow(x, y int, pix []uint8) {
	if len(f.Artwork) == 0 && y >= 0 && y < Height && x >= 0 && x+len(pix) <= Width {
		copy(f.Pix[y*Width+x:], pix)
		return
	}
	for i, c := range pix {
		f.Set(x+i, y, c)
	}
}

// CopyPixels 以新的整張底圖取代畫面，舊圖片顯示層同時失效。
func (f *Framebuffer) CopyPixels(pix []uint8) {
	f.Artwork = nil
	copy(f.Pix, pix)
}
