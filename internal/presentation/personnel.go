package presentation

import (
	"image"

	"github.com/wicanr2/logh3/internal/gfx"
)

// 遊戲提供不含游標的完整背景。背景仍經正式renderer與完整key驗證，
// 色盤、文字、疊窗或資料變動都會重建，游標不進背景快取。
func (r *Renderer) renderPersonnelCursor(fb *gfx.Framebuffer, pal gfx.Palette) *image.RGBA {
	key := r.renderKey(fb, pal)
	if r.frameValid && key == r.frameKey {
		return r.frame
	}
	base := *fb.CursorBackground
	base.InterfaceTheme, base.HoverControl, base.HoverMenu = fb.InterfaceTheme, image.Rectangle{}, false
	baseKey := r.renderKey(&base, pal)
	if r.cursorBasePix == nil || baseKey != r.cursorBaseKey {
		r.frameValid = false
		r.Render(&base, pal)
		r.cursorBasePix = append(r.cursorBasePix[:0], r.frame.Pix...)
		r.cursorBaseMask = append(r.cursorBaseMask[:0], r.uiMask...)
		r.cursorBaseKey = baseKey
	}
	copy(r.frame.Pix, r.cursorBasePix)
	r.uiMask = append(r.uiMask[:0], r.cursorBaseMask...)
	cursor := fb.Commands[len(fb.Commands)-1]
	r.replay([]gfx.DrawCommand{cursor}, pal)
	r.hoverControl(fb, pal)
	r.frameKey, r.frameValid = key, true
	r.generation++
	return r.frame
}
