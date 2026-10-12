package screen

import (
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
)

// DrawAnimationPixels只替換已解碼的矩形。界外原像素有變更時保留完整快照。
// 保留外側HD文字命令，不能用整張CopyPixels蓋掉它們。
func DrawAnimationPixels(fb *gfx.Framebuffer, pixels []uint8, rects []image.Rectangle, bases ...*asset.Image) {
	if fb == nil || len(pixels) != gfx.Width*gfx.Height {
		return
	}
	for i, v := range pixels {
		if fb.Pix[i] == v {
			continue
		}
		p := image.Pt(i%gfx.Width, i/gfx.Width)
		inside := false
		for _, rect := range rects {
			if p.In(rect) {
				inside = true
				break
			}
		}
		if !inside {
			fb.CopyPixels(pixels)
			return
		}
	}
	for i, rect := range rects {
		rect = rect.Intersect(image.Rect(0, 0, gfx.Width, gfx.Height))
		if rect.Empty() {
			continue
		}
		source := make([]uint8, rect.Dx()*rect.Dy())
		for y := 0; y < rect.Dy(); y++ {
			copy(source[y*rect.Dx():], pixels[(rect.Min.Y+y)*gfx.Width+rect.Min.X:(rect.Min.Y+y)*gfx.Width+rect.Max.X])
		}
		start := fb.BeginGroup()
		fb.DrawSource("", rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), rect.Dx(), rect.Dy(), source)
		command := gfx.DrawCommand{Kind: gfx.DrawAnimation, X: rect.Min.X, Y: rect.Min.Y, W: rect.Dx(), H: rect.Dy(), SourceW: rect.Dx(), SourceH: rect.Dy(), Source: source}
		if i < len(bases) && bases[i] != nil && bases[i].W == rect.Dx() && bases[i].H == rect.Dy() {
			command.Key = bases[i].ArtworkKey
			command.Base = bases[i].Pix
		}
		fb.EndGroup(start, command)
	}
}
