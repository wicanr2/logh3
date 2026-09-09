package screen

import (
	"encoding/json"
	"os"
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

// DrawMessageBox **還沒做**：框的位置與大小已經對過收據（`Box`），
// 但框線的畫法與命令選單那一支不同（原版走 `3603:0006`）——收據左緣是
// `8 F 8 7 F F 3`、拿 `ui.MenuFrame` 畫出來是 `8 3 F F 7 8 3`，元素一樣
// 但排法不同，開頭幾列還有階梯狀的斜切。把 `3603:0006` 讀出來之前
// 不畫（`docs/re/137`）。
