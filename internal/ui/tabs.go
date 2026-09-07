package ui

import (
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// TabStyle 是分頁鈕的共通外觀。
//
// 一個分頁鈕在字元平面上是「一格內建圖形 `FF16`（核取方塊 ＋ 左端立體邊）
// ＋ 文字格（樣式 `0x08`）＋ 右端格（樣式 `0x20`）」（`docs/internals/31`）。
// 核取方塊內部填的是那一格的底色——選中的填 2、沒選中的填 9，
// 這是原版分辨選了哪一頁的方式。
type TabStyle struct {
	H      int        `json:"h"`
	Fill   uint8      `json:"fill"`
	Order  string     `json:"order"`
	Layers [][4]uint8 `json:"layers"`
	Ink    uint8      `json:"ink"`
	Shadow uint8      `json:"shadow"`
	TextDY int        `json:"text_dy"`
	Box    struct {
		DX, DY, W, H int
		Order        string     `json:"order"`
		Layers       [][4]uint8 `json:"layers"`
		FillOn       uint8      `json:"fill_on"`
		FillOff      uint8      `json:"fill_off"`
		Patches      []struct {
			DX, DY int
			Color  uint8 `json:"color"`
		} `json:"patches"`
	} `json:"box"`
}

// Tab 是一個分頁鈕。
//
// NoBox ＝ 左端被上層視窗蓋掉的分頁：核取方塊與左緣的立體邊都看不見，
// 只剩文字條與右端的立體邊。原版沒有「部分遮蔽」這個概念——它是逐格畫的，
// 被蓋掉的那幾格就換成別的內容，剩下的照畫。
type Tab struct {
	X, W, TextX int
	Text        string   `json:"text"`
	Codes       []uint16 `json:"codes"`
	On          bool     `json:"on"`
	NoBox       bool     `json:"no_box"`
}

// TabRow 是同一列上的一組分頁鈕。
type TabRow struct {
	Y     int   `json:"y"`
	Items []Tab `json:"items"`
}

// TabBar 是一份分頁鈕規格。
type TabBar struct {
	TabStyle
	Rows []TabRow `json:"rows"`
}

// Draw 畫所有分頁鈕。字裁在內容區裡，立體邊那一列不會咬到上緣。
func (t *TabBar) Draw(fb *gfx.Framebuffer, fnt *font.Font) {
	if t == nil {
		return
	}
	outer := FrameOf(t.Fill, t.Order, t.Layers)
	for _, row := range t.Rows {
		for _, it := range row.Items {
			if it.NoBox {
				t.drawStrip(fb, fnt, it, row.Y)
				continue
			}
			outer.Draw(fb, it.X, row.Y, it.W, t.H)
			fill := t.Box.FillOff
			if it.On {
				fill = t.Box.FillOn
			}
			bx, by := it.X+t.Box.DX, row.Y+t.Box.DY
			FrameOf(fill, t.Box.Order, t.Box.Layers).Draw(fb, bx, by, t.Box.W, t.Box.H)
			for _, p := range t.Box.Patches {
				fb.Set(bx+p.DX, by+p.DY, p.Color)
			}
			cx, cy := it.TextX, row.Y+t.TextDY
			cw, ch := it.X+it.W-2-cx, t.H-2*t.TextDY
			for i, code := range it.Codes {
				DrawGlyphClipped(fb, fnt, cx+i*font.GlyphW, cy, code,
					t.Ink, t.Shadow, -1, false, cx, cy, cw, ch)
			}
		}
	}
}

// drawStrip 畫左端被蓋掉的分頁：整條填底色，再畫上、右、下三邊的立體邊。
func (t *TabBar) drawStrip(fb *gfx.Framebuffer, fnt *font.Font, it Tab, y int) {
	fb.FillRect(it.X, y, it.W, t.H, t.Fill)
	for r, l := range t.Layers {
		lw, ly, lh := it.W-r, y+r, t.H-2*r
		if lw <= 0 || lh <= 0 {
			break
		}
		for x := it.X; x < it.X+lw; x++ { // 上
			fb.Set(x, ly, l[0])
		}
		for yy := ly; yy < ly+lh; yy++ { // 右
			fb.Set(it.X+lw-1, yy, l[1])
		}
		for x := it.X; x < it.X+lw; x++ { // 下
			fb.Set(x, ly+lh-1, l[2])
		}
	}
	cx, cy := it.TextX, y+t.TextDY
	cw, ch := it.X+it.W-2-cx, t.H-2*t.TextDY
	for i, code := range it.Codes {
		DrawGlyphClipped(fb, fnt, cx+i*font.GlyphW, cy, code,
			t.Ink, t.Shadow, -1, false, cx, cy, cw, ch)
	}
}
