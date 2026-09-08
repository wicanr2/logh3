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
		Cursor struct {
			DX, DY int
		} `json:"cursor_offset"`
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
func (s *BattleMenuSpec) BattleMenuAt(mx, my int) (int, int) {
	return mx + s.Layout.Cursor.DX, my + s.Layout.Cursor.DY
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
// 框與主選單同一組（`ui.MenuFrame`）；文字用遊戲既有的立體邊規則——
// 本體色 7、立體邊色 3，字模每列左移一位再上移一列（`docs/internals/25`）。
func DrawBattleMenu(s *BattleMenuSpec, fb *gfx.Framebuffer, fnt *font.Font, x, y, sel int) []uint16 {
	l := &s.Layout
	f := ui.MenuFrame()
	f.Fill = l.Fill
	f.Draw(fb, x, y, l.Window.W, l.Window.H)

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
