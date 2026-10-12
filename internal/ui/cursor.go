package ui

import (
	"encoding/json"
	"image"
	"os"

	"github.com/wicanr2/logh3/internal/gfx"
)

// MouseDY 是畫面座標與原版滑鼠座標的 y 差。
//
// 原版的字元平面從畫面 y ＝ 40 開始（`docs/internals/31`），滑鼠驅動程式
// 回的是平面座標，所以畫面 y ＝ 滑鼠 y ＋ 40。remake 對外一律用畫面座標，
// 只有要餵給吃滑鼠座標的東西（游標點陣、畫面規格的 `mouse` 欄位、
// dosgolem 的 `-clicks`）時才減掉它。
//
// 這個值與 `cursor.json` 的 `offset_from_mouse[1]` 是同一件事：游標點陣的
// 左上角就畫在熱點上，那個 40 全部來自座標系的差。
const MouseDY = 40

// Cursor 是遊戲自己畫的滑鼠游標。
//
// **游標是畫面內容的一部分**，不是模擬器疊上去的東西——原版自己畫它，
// 所以對拍時兩邊位置差一格就會在兩個地方各差一整塊，而畫面看起來完全正常。
type Cursor struct {
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Offset [2]int   `json:"offset_from_mouse"`
	Rows   []string `json:"rows"`

	// Colors 把點陣的字元對到色號。游標是**雙色**的：白色本體加一圈
	// 色號 3 的描邊。在黑色背景上描邊看不見，所以只量黑底會少掉一半。
	Colors map[string]uint8 `json:"colors"`
}

// LoadCursor 讀游標定義。
func LoadCursor(path string) (*Cursor, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Draw 依滑鼠座標畫游標。透明的地方不畫。
func (c *Cursor) Draw(fb *gfx.Framebuffer, mouseX, mouseY int) {
	c.DrawClipped(fb, mouseX, mouseY, 0, 0, gfx.Width, gfx.Height)
}

// DrawClipped 畫游標並裁切在指定矩形內。
//
// **裁切不是可有可無的**：游標壓到視窗外框上時，原版不會蓋掉框——
// 它把游標裁在視窗的內容區裡。少了這條，框上會多出幾個像素的箭頭尖端，
// 那在畫面上完全看不出來。
func (c *Cursor) DrawClipped(fb *gfx.Framebuffer, mouseX, mouseY, cx, cy, cw, ch int) {
	x0, y0 := mouseX+c.Offset[0], mouseY+c.Offset[1]
	start := fb.BeginGroup()
	for y, row := range c.Rows {
		for x, r := range row {
			px, py := x0+x, y0+y
			if px < cx || py < cy || px >= cx+cw || py >= cy+ch {
				continue
			}
			if col, ok := c.Colors[string(r)]; ok {
				fb.Set(px, py, col)
			}
		}
	}
	fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawCursor, X: x0, Y: y0, W: c.Width, H: c.Height, Clip: image.Rect(cx, cy, cx+cw, cy+ch)})
}

// Overlaps 回游標的矩形是否與指定矩形相交。
func (c *Cursor) Overlaps(mouseX, mouseY, x, y, w, h int) bool {
	cx, cy := mouseX+c.Offset[0], mouseY+c.Offset[1]
	return cx < x+w && cx+c.Width > x && cy < y+h && cy+c.Height > y
}
