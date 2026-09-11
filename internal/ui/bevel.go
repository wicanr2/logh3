// Package ui 是原版的視窗與文字元件。
package ui

import "github.com/wicanr2/logh3/internal/gfx"

// Layer 是視窗外框的一層：四個邊各一個色號。
//
// 原版的框是一層層套進去的斜角，光源在**右上**——上緣與右緣拿亮色、
// 左緣與下緣拿暗色。角落由畫的順序決定：先畫暗邊再畫亮邊，所以左上角
// 與右下角是亮色、左下角是暗色。這個順序不是美感選擇，是對拍對出來的
// （`docs/internals/data/frame-profile.json`）。
type Layer struct{ Top, Right, Bottom, Left uint8 }

// Patch 是一個明列的像素。座標為負表示從右／下數回來
// （-1 ＝ 最後一個像素）。
//
// 修正像素是「線條模型與畫面實際長相的差」，不是遊戲裡的東西。對得上
// 16 × 20 格線的框**不需要它**——那種框是 `FRAME.BLK` 的圖塊貼出來的，
// 角落就是圖塊自己的像素（`TileFrame`、`docs/re/289`）。
type Patch struct {
	DX, DY int
	Color  uint8
}

// Frame 是一組由外往內的層，加上內部填色與修正像素。
//
// 這是**線條重建版**，只給還沒對到圖塊的那些框用（主遊戲的大盤面、
// 防拷畫面等）：它們的矩形都不落在 16 × 20 的格線上，畫的常式不是
// `2376:0E4A` 那一條，所以還不知道像素從哪裡來。視窗那種對得上格線的框
// 一律走 `TileFrame`。
type Frame struct {
	Layers  []Layer
	Fill    uint8
	Patches []Patch

	// Order 是四個邊的畫序，**角落歸最後畫的那一邊**。空字串 ＝ "LBTR"
	// （先暗後亮，角落歸上緣與右緣），主選單那種外凸的框用這個。
	// 內凹的框（例如輸入框）要 "TRBL"：角落歸左緣與下緣。
	Order string

	// Alt 讓某一邊變成 1:1 交錯（網點）：非 0 的邊會依 (x+y) 的奇偶
	// 在 Layers 的主色與這裡的色號之間換。長度可以短於 Layers。
	//
	// 防拷問答那個視窗的左緣就是這樣畫的（8 與 15 交錯）——
	// 當成純色會得到「一半的像素對、一半不對」，而框看起來完全正常。
	Alt []Layer
}

// hline／vline 畫一條線；alt 非 0 時依 (x+y) 的奇偶交錯兩個色號。
func hline(fb *gfx.Framebuffer, x, y, n int, c, alt uint8) {
	if alt == 0 {
		fb.HLine(x, y, n, c)
		return
	}
	for i := 0; i < n; i++ {
		fb.Set(x+i, y, pick(x+i, y, c, alt))
	}
}

func vline(fb *gfx.Framebuffer, x, y, n int, c, alt uint8) {
	if alt == 0 {
		fb.VLine(x, y, n, c)
		return
	}
	for i := 0; i < n; i++ {
		fb.Set(x, y+i, pick(x, y+i, c, alt))
	}
}

func pick(x, y int, c, alt uint8) uint8 {
	if (x+y)%2 == 0 {
		return c
	}
	return alt
}

// Draw 把外框與內部填色畫到畫布上。
func (f Frame) Draw(fb *gfx.Framebuffer, x, y, w, h int) {
	for r, l := range f.Layers {
		if w-2*r <= 0 || h-2*r <= 0 {
			return
		}
		lx, ly, lw, lh := x+r, y+r, w-2*r, h-2*r
		var alt Layer
		if r < len(f.Alt) {
			alt = f.Alt[r]
		}
		// 角落歸**最後畫的那一邊**。順序錯了只有四個角不一樣，
		// 而那是唯一看得出差別的地方——整片框其他部分完全相同。
		order := f.Order
		if order == "" {
			order = "LBTR"
		}
		for _, side := range order {
			switch side {
			case 'L':
				vline(fb, lx, ly, lh, l.Left, alt.Left)
			case 'B':
				hline(fb, lx, ly+lh-1, lw, l.Bottom, alt.Bottom)
			case 'T':
				hline(fb, lx, ly, lw, l.Top, alt.Top)
			case 'R':
				vline(fb, lx+lw-1, ly, lh, l.Right, alt.Right)
			}
		}
	}
	r := len(f.Layers)
	if w-2*r > 0 && h-2*r > 0 {
		fb.FillRect(x+r, y+r, w-2*r, h-2*r, f.Fill)
	}
	for _, p := range f.Patches {
		px, py := p.DX, p.DY
		if px < 0 {
			px += w
		}
		if py < 0 {
			py += h
		}
		fb.Set(x+px, y+py, p.Color)
	}
}
