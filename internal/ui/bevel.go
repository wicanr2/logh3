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

// Patch 是一個明列的角落像素。座標為負表示從右／下數回來
// （-1 ＝ 最後一個像素）。
type Patch struct {
	DX, DY int
	Color  uint8
}

// Frame 是一組由外往內的層，加上內部填色與角落修正。
//
// ⚠ **角落規則還沒逆向出來。** 斜角在四個角的切法不是單純的「誰後畫誰贏」
// ——不同層的切點不一致（`docs/internals/07-window-frame.md`）。目前用
// 量自 oracle 的七個像素修正，做得到逐位元組一致，但**換一個視窗尺寸
// 一定要重驗**：這些座標是相對矩形算的，規則對不對還沒被第二個樣本檢驗過。
type Frame struct {
	Layers  []Layer
	Fill    uint8
	Patches []Patch
}

// MenuFrame 是主選單那種視窗的外框，量自 oracle 收據。
func MenuFrame() Frame {
	return Frame{
		Layers: []Layer{
			{15, 15, 8, 8},   // 外凸斜角：亮 15 / 暗 8
			{7, 7, 3, 3},     // 第二道斜角：亮 7 / 暗 3
			{15, 15, 15, 15}, // 亮帶
			{15, 15, 15, 15},
			{3, 3, 7, 7}, // 內凹斜角（亮暗反過來）
			{8, 8, 8, 8},
			{3, 3, 3, 3},
		},
		Fill: 9,
		// 量自 oracle（`workplace/re06/screen.bin`）的角落像素。
		Patches: []Patch{
			{0, 1, 15}, {1, 1, 3}, {4, 5, 8},
			{-6, -5, 8}, {-5, -5, 8}, {-2, -2, 3}, {-2, -1, 15},
		},
	}
}

// Draw 把外框與內部填色畫到畫布上。
func (f Frame) Draw(fb *gfx.Framebuffer, x, y, w, h int) {
	for r, l := range f.Layers {
		if w-2*r <= 0 || h-2*r <= 0 {
			return
		}
		lx, ly, lw, lh := x+r, y+r, w-2*r, h-2*r
		// **先暗後亮**：角落歸亮邊。順序倒過來的話左上角會變暗，
		// 而那是唯一看得出差別的地方——整片框其他部分完全一樣。
		fb.VLine(lx, ly, lh, l.Left)
		fb.HLine(lx, ly+lh-1, lw, l.Bottom)
		fb.HLine(lx, ly, lw, l.Top)
		fb.VLine(lx+lw-1, ly, lh, l.Right)
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
