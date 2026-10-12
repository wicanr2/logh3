package presentation

import "github.com/wicanr2/logh3/internal/gfx"

// refinedCorners 只細化3×3像素的四角；中心和四邊中點保持原色。
// 所有畫面共用同一取樣，動畫不讀前後幀或測試局面。
func refinedCorners(fb *gfx.Framebuffer, x, y int) [4]uint8 {
	width := gfx.Width
	centre := fb.Pix[y*width+x] & 15
	up := fb.Pix[max(y-1, 0)*width+x] & 15
	down := fb.Pix[min(y+1, gfx.Height-1)*width+x] & 15
	left := fb.Pix[y*width+max(x-1, 0)] & 15
	right := fb.Pix[y*width+min(x+1, width-1)] & 15
	result := [4]uint8{centre, centre, centre, centre}
	if up != down && left != right {
		if up == left {
			result[0] = up
		}
		if up == right {
			result[1] = up
		}
		if down == left {
			result[2] = down
		}
		if down == right {
			result[3] = down
		}
	}
	return result
}
