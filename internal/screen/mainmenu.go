// Package screen 是各個畫面的組裝。
//
// 每個畫面都要能在無頭情況下畫成色號陣列，才能跟 oracle 逐位元組比。
// 版面數字一律來自 `docs/internals/data/screen-*.json`，不寫死在程式裡
// ——那些數字是量出來的證據，改動要連著證據一起改。
package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// MenuSpec 是主選單的版面規格（從 JSON 讀進來）。
type MenuSpec struct {
	Background uint8 `json:"background"`
	Window     struct {
		X, Y, W, H      int
		CursorClipInset int `json:"cursor_clip_inset"`
	} `json:"window"`
	Text struct {
		AttrNormal      uint8 `json:"attr_normal"`
		AttrHighlight   uint8 `json:"attr_highlight"`
		ShadowNormal    uint8 `json:"shadow_normal"`
		ShadowHighlight uint8 `json:"shadow_highlight"`
		CellPadY        int   `json:"cell_pad_y"`
		ColPitch        int   `json:"col_pitch"`
		RowPitch        int   `json:"row_pitch"`
		ColsX           []int `json:"cols_x"`
		RowsY           []int `json:"rows_y"`
	} `json:"text"`
	Items     [][]string `json:"items"`
	Highlight *struct {
		Row int `json:"row"`
	} `json:"highlight"`
	Palette6bit [][3]uint8 `json:"palette_6bit"`
}

// LoadMenuSpec 讀主選單規格。
func LoadMenuSpec(path string) (*MenuSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s MenuSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Palette 回規格裡的 16 色調色盤。
func (s *MenuSpec) Palette() gfx.Palette {
	var p gfx.Palette
	for i := 0; i < gfx.Colors && i < len(s.Palette6bit); i++ {
		p[i] = s.Palette6bit[i]
	}
	return p
}

// big5 把一個中文字轉成 Big5 字碼。全形空白（`　`）與半形空白都當空格。
func big5(s string) uint16 {
	switch s {
	case "　", " ", "":
		return ui.Blank
	}
	b := toBig5(s)
	if len(b) != 2 {
		return ui.Blank
	}
	return uint16(b[0])<<8 | uint16(b[1])
}

// MainMenu 畫主選單，反白列取規格裡的預設值。
func MainMenu(s *MenuSpec, fnt *font.Font) (*gfx.Framebuffer, []uint16) {
	row := -1
	if s.Highlight != nil {
		row = s.Highlight.Row
	}
	return MainMenuWith(s, fnt, row)
}

// MainMenuWith 畫主選單並指定反白哪一列（-1 ＝ 不反白）。
func MainMenuWith(s *MenuSpec, fnt *font.Font, highlight int) (*gfx.Framebuffer, []uint16) {
	return MainMenuAt(s, fnt, highlight, nil, -1, -1)
}

// CursorHotY 是游標熱點相對滑鼠 y 的位移。
//
// 遊戲把滑鼠座標當成**游標的位置**在用：命中判定與游標繪製是同一個點，
// 不是兩套座標。這個數字是量出來的（`docs/internals/07`）。
const CursorHotY = 41

// RowAt 回滑鼠 y（遊戲收到的座標）落在哪一列，-1 ＝ 不在任何一列上。
//
// 判準是**游標熱點**落在哪一格的矩形裡，不是滑鼠座標本身。
func RowAt(s *MenuSpec, y int) int {
	hot := y + CursorHotY
	first := s.Text.RowsY[0] - s.Text.CellPadY
	if hot < first {
		return -1
	}
	r := (hot - first) / s.Text.RowPitch
	if r >= len(s.Text.RowsY) {
		return -1
	}
	return r
}

// MainMenuAt 畫主選單，並在指定的滑鼠座標畫游標（cur 為 nil 或座標為負就不畫）。
func MainMenuAt(s *MenuSpec, fnt *font.Font, highlight int, cur *ui.Cursor, mx, my int) (*gfx.Framebuffer, []uint16) {
	fb := gfx.NewFramebuffer(s.Background)
	w := ui.NewTextWindow(s.Window.X, s.Window.Y, len(s.Text.ColsX), len(s.Text.RowsY),
		ui.MenuFrame(), s.Text.AttrNormal, s.Text.ShadowNormal)
	w.ColPitch, w.RowPitch = s.Text.ColPitch, s.Text.RowPitch
	w.CellPadY = s.Text.CellPadY
	w.OriginX = s.Text.ColsX[0] - s.Window.X
	w.OriginY = s.Text.RowsY[0] - s.Window.Y
	for r, line := range s.Items {
		codes := make([]uint16, 0, len(line))
		for _, ch := range line {
			codes = append(codes, big5(ch))
		}
		w.SetText(r, s.Text.AttrNormal, s.Text.ShadowNormal, codes)
	}
	// **順序是量出來的**：一般的列先畫完，接著畫游標，最後才畫反白的那一列。
	// 所以游標會被反白列蓋掉一部分——換個順序畫面就會多出一截箭頭，
	// 而那看起來只像「多了一個箭頭」，不像錯。
	missing := w.Draw(fb, fnt, s.Window.W, s.Window.H)
	if cur != nil && mx >= 0 && my >= 0 {
		cur.Draw(fb, mx, my)
	}
	if highlight >= 0 && highlight < w.Rows {
		w.SetRowAttr(highlight, s.Text.AttrHighlight, s.Text.ShadowHighlight)
		missing = append(missing, w.DrawRow(fb, fnt, highlight)...)
	}
	return fb, missing
}
