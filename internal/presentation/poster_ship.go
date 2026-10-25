package presentation

import "image"

// PosterShip只取已驗圖集的顯示複本；不改戰鬥來源或原PNG。
func (r *Renderer) PosterShip(key string) *image.RGBA {
	if r == nil || r.ships[key] == nil {
		return nil
	}
	im := resize(r.ships[key], 192, 76)
	for i := 0; i < len(im.Pix); i += 4 {
		if im.Pix[i] == 0 && im.Pix[i+1] == 0 && im.Pix[i+2] == 0 {
			im.Pix[i+3] = 0
		}
	}
	return im
}
