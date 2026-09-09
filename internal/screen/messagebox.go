package screen

import (
	"encoding/json"
	"os"

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

// MessageBoxStyle／MessageBoxColour 是訊息框傳給視窗系統的兩個值：
// `3826:0005` 拿樣板的 `[+4]` 當旗標（三種訊息都是 1），顏色固定 0x97。
// 圖塊裡色號 0 的像素填的是顏色的**高四位**（`docs/re/138`）。
const (
	MessageBoxStyle  = 1
	MessageBoxColour = 0x97
)

// DrawMessageBox 把第 kind 種訊息框的框與底畫出來（不含文字）。
//
// 框是圖塊不是線：外圈八張取自 `FRAME.BLK`，內部填視窗底色。旗標讓框
// 往外多長兩格，所以實際佔 (寬＋2) × (高＋2) 格（`docs/re/137`、`138`）。
func DrawMessageBox(fb *gfx.Framebuffer, s *MessageBoxSpec, kind int, tiles *ui.FrameTiles) bool {
	if fb == nil || s == nil || tiles == nil {
		return false
	}
	for _, t := range s.Templates {
		if t.Kind != kind {
			continue
		}
		f := ui.TileFrame{Tiles: tiles, Style: MessageBoxStyle, Fill: MessageBoxColour >> 4}
		// 格 y **不加 RowBias**：那兩格就是格座標系原點與畫面頂端
		// 的 40 像素（`ui.FrameCellOriginY`），Draw 自己會加。
		f.Draw(fb, t.CellX, t.CellY, t.Cols+2, t.Rows+2)
		return true
	}
	return false
}
