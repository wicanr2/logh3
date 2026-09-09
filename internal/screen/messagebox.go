package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 訊息框（原版 `3826:0005`，`docs/re/137`）。三個地方用它：選點不合法的
// 「這是無法行動的地方。」、補給與修理沒有對象的兩句。
//
// 版面來自 `ds:8BEC` 的樣板表（一種 6 bytes：格 x、格 y、寬、高…），
// 框與字的幾何與命令選單同一套。**只有 y 多 2 格**——樣板寫 7，
// 收據上量到的框頂是 181 ＝ (7＋2)×20＋1（`s3msg`）。

// MessageBoxSpec 是訊息框的版面。
type MessageBoxSpec struct {
	Templates []struct {
		Kind   int `json:"kind"`
		CellX  int `json:"cell_x"`
		CellY  int `json:"cell_y"`
		Cols   int `json:"cols"`
		Rows   int `json:"rows"`
	} `json:"templates"`
	RowBias int `json:"row_bias"` // 樣板的格 y 要再加幾格
	Style   int `json:"style"`    // 外框樣式的起始碼（`視窗[+3]`）
	Colour  int `json:"colour"`   // 視窗顏色，高四位是底色

	// Text 是內容區裡文字的排版。原點是內容格的左上角。
	Text struct {
		DX       int   `json:"dx"`
		DY       int   `json:"dy"`
		ColPitch int   `json:"col_pitch"`
		RowPitch int   `json:"row_pitch"`
		Ink      uint8 `json:"ink"`
		Shadow   uint8 `json:"shadow"`
		ShadowDY int   `json:"shadow_dy"`
	} `json:"text"`
}

// LoadMessageBoxSpec 讀訊息框的版面。
func LoadMessageBoxSpec(path string) (*MessageBoxSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s MessageBoxSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Box 回第 kind 種訊息框的框左上角與大小。
func (s *MessageBoxSpec) Box(kind int, l *CommandMenuSpec) (x, y, w, h int, ok bool) {
	if s == nil || l == nil {
		return 0, 0, 0, 0, false
	}
	for _, t := range s.Templates {
		if t.Kind != kind {
			continue
		}
		g := &l.Layout
		return t.CellX*g.CellW + g.OriginBias,
			(t.CellY+s.RowBias)*g.CellH + g.OriginBias,
			t.Cols*g.CellW + g.FramePadW,
			t.Rows*g.CellH + g.FramePadH, true
	}
	return 0, 0, 0, 0, false
}

// DrawMessageBox 把第 kind 種訊息框的框與底畫出來（不含文字）。
//
// 框是圖塊不是線：外圈八張取自 `FRAME.BLK`，內部填視窗底色。旗標讓框
// 往外多長兩格，所以實際佔 (寬＋2) × (高＋2) 格（`docs/re/137`、`138`）。
func DrawMessageBox(fb *gfx.Framebuffer, s *MessageBoxSpec, kind int, tiles *ui.FrameTiles) bool {
	if fb == nil || s == nil || tiles == nil {
		return false
	}
	t, ok := s.template(kind)
	if !ok {
		return false
	}
	f := ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}
	// 格 y **不加 RowBias**：那兩格就是格座標系原點與畫面頂端的 40 像素
	// （`ui.FrameCellOriginY`），Draw 自己會加。
	f.Draw(fb, t.cellX, t.cellY, t.cols+2, t.rows+2)
	return true
}

// tmpl 是樣板的一列，欄位換成小寫免得外面誤用匿名結構。
type tmpl struct{ cellX, cellY, cols, rows int }

func (s *MessageBoxSpec) template(kind int) (tmpl, bool) {
	for _, t := range s.Templates {
		if t.Kind == kind {
			return tmpl{t.CellX, t.CellY, t.Cols, t.Rows}, true
		}
	}
	return tmpl{}, false
}

// MessageBoxTextOrigin 回第 kind 種訊息框裡文字第 0 列第 0 欄的左上角。
//
// 內容區從框往內縮一格（旗標多出來的那一圈），文字再往下 `Text.DY`
// 像素——收據 `s3msg` 量到 (80, 202)，內容格是 (5, 8)（`docs/re/138`）。
func (s *MessageBoxSpec) MessageBoxTextOrigin(kind int) (x, y int, ok bool) {
	t, ok := s.template(kind)
	if !ok {
		return 0, 0, false
	}
	return (t.cellX+1)*ui.FrameTileW + s.Text.DX,
		(t.cellY+1)*ui.FrameTileH + ui.FrameCellOriginY + s.Text.DY, true
}

// DrawMessageBoxText 把一行全形字畫進訊息框。row 是內容區的第幾列。
func DrawMessageBoxText(fb *gfx.Framebuffer, s *MessageBoxSpec, kind int,
	fnt *font.Font, row int, codes []uint16) []uint16 {
	x, y, ok := s.MessageBoxTextOrigin(kind)
	if !ok {
		return codes
	}
	y += row * s.Text.RowPitch
	var missing []uint16
	for i, c := range codes {
		if !ui.DrawGlyph(fb, fnt, x+i*s.Text.ColPitch, y, c,
			s.Text.Ink, s.Text.Shadow, s.Text.ShadowDY, false) {
			missing = append(missing, c)
		}
	}
	return missing
}
