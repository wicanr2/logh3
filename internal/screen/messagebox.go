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

// 訊息框（原版 `3826:0005`，`docs/re/137`、`138`、`141`）。全遊戲的呼叫端
// 與訊息清單在 `docs/internals/48`。
//
// 版面來自 `ds:8BEC` 的樣板表（一種 6 bytes：格 x、格 y、寬、高…），
// 框與字的幾何與命令選單同一套。**只有 y 多 2 格**——樣板寫 7，
// 收據上量到的框頂是 181 ＝ (7＋2)×20＋1（`s3msg`）。

// MessageBoxSpec 是訊息框的版面。
type MessageBoxSpec struct {
	Templates []struct {
		Kind  int `json:"kind"`
		CellX int `json:"cell_x"`
		CellY int `json:"cell_y"`
		Cols  int `json:"cols"`
		Rows  int `json:"rows"`
	} `json:"templates"`
	RowBias int `json:"row_bias"` // 樣板的格 y 要再加幾格
	Style   int `json:"style"`    // 外框樣式的起始碼（`視窗[+3]`）
	Colour  int `json:"colour"`   // 視窗顏色，高四位是底色

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

// Message 回名為 name 的那一句：訊息框種類與原版資料段那一串 Big5
// （控制碼夾在裡面，交給 `ui.TextBuffer` 解）。
func (s *MessageBoxSpec) Message(name string) (kind int, raw []byte, ok bool) {
	for _, m := range s.Messages {
		if m.Name != name {
			continue
		}
		b, err := hex.DecodeString(m.Bytes)
		if err != nil {
			return 0, nil, false
		}
		return m.Kind, b, true
	}
	return 0, nil, false
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

// Box 回第 kind 種訊息框的框左上角與大小。
//
// 這是**看得見的**那個矩形（框的最外圈那一像素與底色同色，肉眼上不算），
// 所以位置比格座標晚一個像素、大小少兩個。`RowBias` 那兩格就是格座標系
// 原點與畫面頂端的 40 像素。
func (s *MessageBoxSpec) Box(kind int, l *CommandMenuSpec) (x, y, w, h int, ok bool) {
	if s == nil || l == nil {
		return 0, 0, 0, 0, false
	}
	t, found := s.template(kind)
	if !found {
		return 0, 0, 0, 0, false
	}
	g := &l.Layout
	return t.cellX*g.CellW + g.OriginBias,
		(t.cellY+s.RowBias)*g.CellH + g.OriginBias,
		t.cols*g.CellW + g.FramePadW,
		t.rows*g.CellH + g.FramePadH, true
}

// DrawMessageBox 把第 kind 種訊息框整個畫出來：框 ＋ 字 ＋ 按鈕。
//
// 內容走通用的文字緩衝區（`ui.TextBuffer`，`docs/re/141`）——原版的字都是
// 先寫進視窗自己的內容緩衝區，再由另一趟畫成像素，這裡照做。按鈕不是
// 另一種元件，是那幾格的顏色與旗標（`37A9:004C`，`docs/re/138`）。
//
// raw 是原版資料段那一串 Big5。回沒有字模的字碼。
func DrawMessageBox(fb *gfx.Framebuffer, s *MessageBoxSpec, kind int,
	tiles *ui.FrameTiles, fnt *font.Font, raw []byte) []uint16 {
	if fb == nil || s == nil || tiles == nil {
		return nil
	}
	t, ok := s.template(kind)
	if !ok {
		return nil
	}
	(ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}).
		Draw(fb, t.cellX, t.cellY, t.cols+2, t.rows+2)

	b := ui.NewTextBuffer(t.cols, t.rows, uint8(s.Colour))
	b.Write(raw)
	s.putButtons(b, kind)
	return ui.DrawTextBuffer(fb, fnt, tiles, b, t.cellX+1, t.cellY+1)
}

// putButtons 把子元件寫進內容緩衝區，照 `37A9:004C`：文字那幾格換顏色
// 並帶旗標 `0x08`，左邊多一格 `0x10`、右邊多一格 `0x20`（那兩格**不換
// 顏色**，底仍是視窗的）。
func (s *MessageBoxSpec) putButtons(b *ui.TextBuffer, kind int) {
	for _, c := range s.Children {
		if c.Kind != kind {
			continue
		}
		for _, it := range c.Items {
			for r := 0; r < it.Rows; r++ {
				if cell := b.At(it.CellX-1, it.CellY+r); cell != nil {
					cell.Flags |= ui.CellButtonLeft
				}
				for i := 0; i < it.Cols; i++ {
					if cell := b.At(it.CellX+i, it.CellY+r); cell != nil {
						cell.Colour = uint8(s.Button.Colour)
						cell.Flags |= ui.CellButtonBody
					}
				}
				if cell := b.At(it.CellX+it.Cols, it.CellY+r); cell != nil {
					cell.Flags |= ui.CellButtonRight
				}
			}
			for i, h := range it.Codes {
				var code uint16
				if _, err := fmt.Sscanf(h, "%04X", &code); err != nil {
					continue
				}
				if cell := b.At(it.CellX+i, it.CellY); cell != nil {
					cell.Code = code
				}
			}
		}
	}
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

// ButtonIsDecide 回第 kind 種訊息框的第 i 顆鈕是不是「決定」／「確認」
// （另一顆是「中止」）。`MessageBoxButtonAt` 回的是子元件表裡的順序，
// 兩顆鈕的框裡「中止」排在前面，所以不能拿索引 0 當「決定」。
func (s *MessageBoxSpec) ButtonIsDecide(kind, i int) bool {
	for _, c := range s.Children {
		if c.Kind != kind || i < 0 || i >= len(c.Items) {
			continue
		}
		t := c.Items[i].Text
		return t == "決定" || t == "確認"
	}
	return false
}
