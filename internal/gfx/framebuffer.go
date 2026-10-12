// Package gfx 是 remake 的畫面層。
//
// 畫布刻意做成**與原版同構**：640×480、每像素一個 4 位元色號、外加一份
// 16 色的 6 位元 DAC。這不是為了懷舊，是為了讓對拍成立——oracle 倒出來的
// 就是色號陣列與 DAC，兩邊同型才能逐位元組比較，而不是比壓縮過的圖片。
package gfx

import "image"

const (
	// Width／Height 是 VGA mode 12h 的邏輯畫布。
	Width  = 640
	Height = 480

	// Colors 是 16 色模式的色號數。
	Colors = 16
)

// Framebuffer 是一張色號陣列。**不存 RGB**——顏色由 Palette 決定，
// 兩者分開才對得上原版「改調色盤就整片變色」的行為。
type Framebuffer struct {
	Pix            []uint8         // 長度 Width*Height，值 0–15
	Artwork        []Artwork       // 可選圖片顯示層；不改變色號或遊戲資料。
	Commands       []DrawCommand   // 正式高解析度顯示依實際畫序重播。
	InterfaceTheme string          // 可選顯示識別；不參與色號、規則或存檔。
	HoverControl   image.Rectangle // 正式命中區的HD回饋，不改索引像素。
	mute           int
}

// NewFramebuffer 造一張全部填 c 的畫布。
func NewFramebuffer(c uint8) *Framebuffer {
	f := &Framebuffer{Pix: make([]uint8, Width*Height)}
	f.Clear(c)
	return f
}

// Clear 整片填色。
func (f *Framebuffer) Clear(c uint8) {
	f.Artwork = nil
	f.Commands = []DrawCommand{{Kind: DrawFill, W: Width, H: Height, Colour: c}}
	for i := range f.Pix {
		f.Pix[i] = c
	}
}

// Set 畫一個像素。超出畫布就丟掉——裁切是繪圖層的責任，
// 讓呼叫端自己判邊界的話，每個呼叫端都會漏掉一種情況。
func (f *Framebuffer) Set(x, y int, c uint8) {
	if x < 0 || y < 0 || x >= Width || y >= Height {
		return
	}
	f.Pix[y*Width+x] = c
	f.recordPixel(x, y, c)
	for i := range f.Artwork {
		a := &f.Artwork[i]
		if x >= a.X && y >= a.Y && x < a.X+a.W && y < a.Y+a.H {
			a.Visible[(y-a.Y)*a.W+x-a.X] = 0
		}
	}
}

// At 讀一個像素；界外回 0。
func (f *Framebuffer) At(x, y int) uint8 {
	if x < 0 || y < 0 || x >= Width || y >= Height {
		return 0
	}
	return f.Pix[y*Width+x]
}

// FillRect 填一個矩形（含左上、不含右下）。
func (f *Framebuffer) FillRect(x, y, w, h int, c uint8) {
	f.mute++
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			f.Set(i, j, c)
		}
	}
	f.mute--
	if f.mute == 0 && w > 0 && h > 0 {
		f.Commands = append(f.Commands, DrawCommand{Kind: DrawFill, X: x, Y: y, W: w, H: h, Colour: c})
	}
}

// HLine／VLine 是一維的 FillRect，讓框線的程式碼讀起來像它在描述的東西。
func (f *Framebuffer) HLine(x, y, w int, c uint8) { f.FillRect(x, y, w, 1, c) }
func (f *Framebuffer) VLine(x, y, h int, c uint8) { f.FillRect(x, y, 1, h, c) }

// Palette 是 16 色的 6 位元 DAC（每個分量 0–63），與原版同格式。
type Palette [Colors][3]uint8

// RGB8 把 6 位元分量轉成 8 位元，用「高位補低位」與 dosgolem 一致。
func (p Palette) RGB8(i int) (r, g, b uint8) {
	v := p[i&(Colors-1)]
	f := func(x uint8) uint8 { return x<<2 | x>>4 }
	return f(v[0]), f(v[1]), f(v[2])
}
