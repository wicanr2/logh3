package screen

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	// Button 是子元件的配色（`ds:8BE4`）。
	Button struct {
		Colour int `json:"colour"`
	} `json:"button"`

	// Children 是每一種訊息框的按鈕（`37A9:000F` 的 12 bytes 一筆表）。
	Children []struct {
		Kind  int `json:"kind"`
		Items []struct {
			CellX int      `json:"cell_x"`
			CellY int      `json:"cell_y"`
			Cols  int      `json:"cols"`
			Rows  int      `json:"rows"`
			Codes []string `json:"codes"`
			Text  string   `json:"text"`
		} `json:"items"`
	} `json:"children"`

	// Messages 是三句訊息（`docs/re/121`、`131`）。`bytes` 是原版資料段裡
	// 那一串 Big5，照抄。
	Messages []struct {
		Name   string `json:"name"`
		Offset string `json:"offset"`
		Kind   int    `json:"kind"`
		Colour int    `json:"colour"`
		Text   string `json:"text"`
		Bytes  string `json:"bytes"`
	} `json:"messages"`
}

// Message 回名為 name 的那一句：訊息框種類與字碼。
//
// 半形（高位元組 < 0x80，例如原版字串尾巴那兩個空白）跳過——底色已經
// 由框那一趟填好，畫空白是沒有作用的。
func (s *MessageBoxSpec) Message(name string) (kind int, codes []uint16, ok bool) {
	for _, m := range s.Messages {
		if m.Name != name {
			continue
		}
		raw, err := hex.DecodeString(m.Bytes)
		if err != nil {
			return 0, nil, false
		}
		for i := 0; i+1 < len(raw); i += 2 {
			c := uint16(raw[i])<<8 | uint16(raw[i+1])
			if c < 0x8000 {
				continue
			}
			codes = append(codes, c)
		}
		return m.Kind, codes, true
	}
	return 0, nil, false
}

// MessageBoxButtonAt 回 (x, y) 有沒有落在第 kind 種訊息框的某顆按鈕上，
// 以及那顆按鈕的編號（`紀錄[+5]`，0 ＝ 確認／決定、1 ＝ 中止）。
//
// 命中範圍就是按鈕那幾格本身；左右那兩格只是端點的裝飾，原版的命中測試
// 走的是子元件自己的格範圍（`37A9:01F9` → `341B:44A`）。
func (s *MessageBoxSpec) MessageBoxButtonAt(kind, x, y int) (int, bool) {
	t, ok := s.template(kind)
	if !ok {
		return 0, false
	}
	ox := (t.cellX + 1) * ui.FrameTileW
	oy := (t.cellY+1)*ui.FrameTileH + ui.FrameCellOriginY
	for _, c := range s.Children {
		if c.Kind != kind {
			continue
		}
		for i, it := range c.Items {
			bx := ox + it.CellX*ui.FrameTileW
			by := oy + it.CellY*ui.FrameTileH
			if x >= bx && x < bx+it.Cols*ui.FrameTileW &&
				y >= by && y < by+it.Rows*ui.FrameTileH {
				return i, true
			}
		}
	}
	return 0, false
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

// DrawMessageBoxButtons 把訊息框的按鈕畫出來。
//
// `37A9:004C` 是寫進視窗的內容緩衝區，不是直接畫：按鈕那幾格拿到顏色
// `ds:8BE4[…]`（0x87 ＝ 底色 8、字色 7）與旗標 `0x08`，左邊多一格拿
// `0x10`、右邊多一格拿 `0x20`。左右那兩格**沒有換顏色**，所以底仍是視窗
// 的 9，只有格內邊緣那兩行像素被畫上（`docs/re/138`）。
func DrawMessageBoxButtons(fb *gfx.Framebuffer, s *MessageBoxSpec, kind int,
	fnt *font.Font) []uint16 {
	t, ok := s.template(kind)
	if !ok {
		return nil
	}
	bg := uint8(s.Button.Colour >> 4)
	ink := uint8(s.Button.Colour & 0x0F)
	// 內容格 (0,0) 的左上角像素。
	ox := (t.cellX + 1) * ui.FrameTileW
	oy := (t.cellY+1)*ui.FrameTileH + ui.FrameCellOriginY

	var missing []uint16
	for _, c := range s.Children {
		if c.Kind != kind {
			continue
		}
		for _, it := range c.Items {
			// 順序照 `2376:10B7`：底色 → 立體邊 → 字 → 裝飾。裝飾**最後**畫，
			// 所以按鈕上緣那兩列會蓋掉字的立體邊——差就差在那兩個像素。
			for r := 0; r < it.Rows; r++ {
				y := oy + (it.CellY+r)*ui.FrameTileH
				for i := 0; i < it.Cols; i++ {
					fb.FillRect(ox+(it.CellX+i)*ui.FrameTileW, y,
						ui.FrameTileW, ui.FrameTileH, bg)
				}
			}
			for i, h := range it.Codes {
				var code uint16
				if _, err := fmt.Sscanf(h, "%04X", &code); err != nil {
					continue
				}
				if !ui.DrawGlyph(fb, fnt,
					ox+(it.CellX+i)*ui.FrameTileW+s.Text.DX,
					oy+it.CellY*ui.FrameTileH+s.Text.DY,
					code, ink, s.Text.Shadow, s.Text.ShadowDY, false) {
					missing = append(missing, code)
				}
			}
			for r := 0; r < it.Rows; r++ {
				y := oy + (it.CellY+r)*ui.FrameTileH
				for i := 0; i < it.Cols; i++ {
					ui.DrawCellDecor(fb, ox+(it.CellX+i)*ui.FrameTileW, y, ui.CellButtonBody)
				}
				ui.DrawCellDecor(fb, ox+(it.CellX-1)*ui.FrameTileW, y, ui.CellButtonLeft)
				ui.DrawCellDecor(fb, ox+(it.CellX+it.Cols)*ui.FrameTileW, y, ui.CellButtonRight)
			}
		}
	}
	return missing
}