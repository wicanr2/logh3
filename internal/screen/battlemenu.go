package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// BattleMenuSpec 是戰術選單的版面（`docs/internals/data/battle-menu.json`）。
//
// 戰鬥每一回合結束時，原版在**游標所在位置**開這張選單（`12AA:02D4`），
// 第一項「實行」正好落在游標底下——所以同一點連點兩下就等於推進一回合
// （`docs/internals/34`）。
type BattleMenuSpec struct {
	Title  label   `json:"title"`
	Items  []label `json:"items"`
	Layout struct {
		Window struct{ X, Y, W, H int } `json:"window"`
		Origin struct {
			CellW          int `json:"cell_w"`
			CellH          int `json:"cell_h"`
			OriginBias     int `json:"origin_bias"`
			CursorCellBias int `json:"cursor_cell_bias"`
			ClampRows      int `json:"clamp_rows"`
		} `json:"origin"`
		Text struct {
			X, Y        int
			ColPitch    int `json:"col_pitch"`
			RowPitch    int `json:"row_pitch"`
			Cols, Rows  int
			Ink, Shadow uint8
		} `json:"text"`
		Highlight struct {
			X, Y, W, H int
			PadY       int `json:"pad_y"`
			Ink        uint8
		} `json:"highlight"`
		Fill uint8 `json:"fill"`
	} `json:"layout"`
}

// LoadBattleMenuSpec 讀戰術選單的版面。
func LoadBattleMenuSpec(path string) (*BattleMenuSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s BattleMenuSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// BattleMenuAt 回把選單開在游標 (mx,my) 時，框的左上角在哪裡。
//
// 位置是**格**不是像素：全遊戲 22 張選單共用同一支開啟常式
// （`3667:0004`），位置參數是兩個 byte（`341B:04DB`，放不下 480），
// 格 ＝ 游標所在格 − (1,1)，再由 `3982:0198` 夾成「格y ≤ 19 − 項目數」，
// 框左上角 ＝ `(格x × 16 + 1, 格y × 20 + 1)`（`docs/re/117`）。
//
// 先前這裡寫的是「游標 ＋ (−15,−19)」。收據 `s3menu` 的游標落在
// (240,240)——剛好是 16 與 20 的整數倍，兩個模型在那一點給出同一個答案，
// 所以那一份收據分不出對錯。像素模型在游標不落在格線上時會給出對不上
// 格線的框，而原版的視窗位置存的就是格（`視窗[+5]`／`[+6]`），畫不出來。
func (s *BattleMenuSpec) BattleMenuAt(mx, my int) (int, int) {
	o := &s.Layout.Origin
	cx := mx/o.CellW + o.CursorCellBias
	cy := my/o.CellH + o.CursorCellBias
	if cx < 0 {
		cx = 0
	}
	if cy < 0 {
		cy = 0
	}
	if max := o.ClampRows - len(s.Items); cy > max {
		cy = max
	}
	return cx*o.CellW + o.OriginBias, cy*o.CellH + o.OriginBias
}

// BattleMenuRowAt 回畫面座標 (x,y) 落在第幾項，-1 ＝ 不在任何一項上。
//
// 判準是反白帶的矩形——第一項的帶頂剛好等於游標的 y（`docs/re/101`），
// 所以在同一點連點兩下時第二下一定落在第一項「實行」上。
func (s *BattleMenuSpec) BattleMenuRowAt(left, top, x, y int) int {
	l := &s.Layout
	hx := left + (l.Highlight.X - l.Window.X)
	if x < hx || x >= hx+l.Highlight.W {
		return -1
	}
	first := top + (l.Highlight.Y - l.Window.Y)
	if y < first {
		return -1
	}
	r := (y - first) / l.Text.RowPitch
	if r >= l.Text.Rows {
		return -1
	}
	return r
}

// DrawBattleMenu 把戰術選單畫在 (x,y)（框的左上角），sel 是反白哪一項。
//
// 框與主選單同一組（`FRAME.BLK` 起始碼 9）；文字用遊戲既有的立體邊規則
// ——本體色 7、立體邊色 3，字模每列左移一位再上移一列（`docs/internals/25`）。
// tiles 缺了就不畫框（不用代用品）。
func DrawBattleMenu(s *BattleMenuSpec, fb *gfx.Framebuffer, fnt *font.Font,
	tiles *ui.FrameTiles, x, y, sel int) []uint16 {
	l := &s.Layout
	ui.MenuFrame(tiles, l.Fill).DrawRect(fb, x, y, l.Window.W, l.Window.H)

	tx := x + (l.Text.X - l.Window.X)
	ty := y + (l.Text.Y - l.Window.Y)
	if sel >= 0 && sel < l.Text.Rows {
		// 反白帶的頂端比字身高 2 列（`docs/re/101`）。
		fb.FillRect(x+(l.Highlight.X-l.Window.X),
			y+(l.Highlight.Y-l.Window.Y)+sel*l.Text.RowPitch,
			l.Highlight.W, l.Highlight.H, l.Highlight.Ink)
	}
	var missing []uint16
	for r, item := range s.Items {
		if r >= l.Text.Rows {
			break
		}
		for c, code := range item.Codes {
			if c >= l.Text.Cols {
				break
			}
			if !ui.DrawGlyph(fb, fnt, tx+c*l.Text.ColPitch, ty+r*l.Text.RowPitch,
				code, l.Text.Ink, l.Text.Shadow, -1, false) {
				missing = append(missing, code)
			}
		}
	}
	return missing
}
