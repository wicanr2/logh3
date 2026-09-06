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
		X, Y, W, H int
	} `json:"window"`
	Text struct {
		Color    uint8 `json:"color"`
		ColPitch int   `json:"col_pitch"`
		RowPitch int   `json:"row_pitch"`
		ColsX    []int `json:"cols_x"`
		RowsY    []int `json:"rows_y"`
	} `json:"text"`
	Items     [][]string `json:"items"`
	Highlight *struct {
		Row   int   `json:"row"`
		Color uint8 `json:"color"`
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

// MainMenu 畫主選單，回畫布與缺字清單。
func MainMenu(s *MenuSpec, fnt *font.Font) (*gfx.Framebuffer, []uint16) {
	fb := gfx.NewFramebuffer(s.Background)
	w := ui.NewTextWindow(s.Window.X, s.Window.Y, len(s.Text.ColsX), len(s.Text.RowsY),
		ui.MenuFrame())
	w.ColPitch, w.RowPitch = s.Text.ColPitch, s.Text.RowPitch
	w.OriginX = s.Text.ColsX[0] - s.Window.X
	w.OriginY = s.Text.RowsY[0] - s.Window.Y
	if s.Highlight != nil {
		w.Highlight, w.HighlightColor = s.Highlight.Row, s.Highlight.Color
	}
	for r, line := range s.Items {
		codes := make([]uint16, 0, len(line))
		for _, ch := range line {
			codes = append(codes, big5(ch))
		}
		w.SetText(r, s.Text.Color, codes)
	}
	return fb, w.Draw(fb, fnt, s.Window.W, s.Window.H)
}
