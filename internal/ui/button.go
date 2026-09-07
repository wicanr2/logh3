package ui

import (
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// ButtonStyle 是一顆按鈕的共通外觀：兩層立體邊（上／左色號 7、下／右色號 3）、
// 底色 8、字色 7、立體邊色號 3。
//
// 原版沒有「按鈕」這個東西——它是字元平面上三種樣式旗標拼出來的：
// 左端一格 `0x10`、中間文字格 `0x08`、右端一格 `0x20`
// （`docs/internals/31`）。矩形換算成像素是
// 「左端格的最後兩個像素」到「右端格的前兩個像素」。
type ButtonStyle struct {
	Fill   uint8      `json:"fill"`
	Order  string     `json:"order"`
	Layers [][4]uint8 `json:"layers"`
	Ink    uint8      `json:"ink"`
	Shadow uint8      `json:"shadow"`
	TextDX int        `json:"text_dx"`
	TextDY int        `json:"text_dy"`
}

// Button 是一顆按鈕的矩形與字。
type Button struct {
	X, Y, W, H int
	Text       string   `json:"text"`
	Codes      []uint16 `json:"codes"`
}

// ButtonBar 是一組共用同一種外觀的按鈕。
type ButtonBar struct {
	ButtonStyle
	Items []Button `json:"items"`
}

// FrameOf 把規格裡的層色轉成 Frame。
func FrameOf(fill uint8, order string, layers [][4]uint8) Frame {
	f := Frame{Fill: fill, Order: order}
	for _, l := range layers {
		f.Layers = append(f.Layers, Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
	}
	return f
}

// Draw 畫這一組按鈕。
//
// 字裁在按鈕的內容區裡：立體邊往上移那一列會落到按鈕的白色上緣，
// 原版的上緣是乾淨的。
func (b *ButtonBar) Draw(fb *gfx.Framebuffer, fnt *font.Font) {
	if b == nil {
		return
	}
	f := FrameOf(b.Fill, b.Order, b.Layers)
	for _, it := range b.Items {
		f.Draw(fb, it.X, it.Y, it.W, it.H)
		cx, cy := it.X+b.TextDX, it.Y+b.TextDY
		cw, ch := it.W-2*b.TextDX, it.H-2*b.TextDY
		for i, code := range it.Codes {
			DrawGlyphClipped(fb, fnt, cx+i*font.GlyphW, cy, code,
				b.Ink, b.Shadow, -1, false, cx, cy, cw, ch)
		}
	}
}
