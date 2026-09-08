package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// CommandMenuKind 是戰場上五張命令選單之一（`docs/spec/023`）。
type CommandMenuKind string

const (
	CommandMenuFleet    CommandMenuKind = "fleet"    // ds:4120 移動／戰鬥／索敵
	CommandMenuSquadron CommandMenuKind = "squadron" // ds:413C 移動／編隊
	CommandMenuSupply   CommandMenuKind = "supply"   // ds:4158 移動／編隊／補給
	CommandMenuRepair   CommandMenuKind = "repair"   // ds:417A 移動／編隊／修理
	CommandMenuScout    CommandMenuKind = "scout"    // ds:4196 移動／歸隊
)

// CommandMenuSpec 是戰場命令選單的版面與五張表
// （`docs/internals/data/battle-command-menus.json`）。
//
// 開啟常式與戰術選單同一支，配色與框線也同一套，所以這裡只放每一張的
// 項目與內寬；框的位置與大小走 layout 的公式（`docs/re/117`）。
type CommandMenuSpec struct {
	Width int `json:"width"` // 內寬，單位是 16 像素
	Menus map[CommandMenuKind]struct {
		Address string  `json:"address"`
		When    string  `json:"when"`
		Items   []label `json:"items"`
	} `json:"menus"`
	Layout struct {
		CellW           int   `json:"cell_w"`
		CellH           int   `json:"cell_h"`
		OriginBias      int   `json:"origin_bias"`
		CursorCellBias  int   `json:"cursor_cell_bias"`
		ClampRows       int   `json:"clamp_rows"`
		FramePadW       int   `json:"frame_pad_w"`
		FramePadH       int   `json:"frame_pad_h"`
		HighlightDX     int   `json:"highlight_dx"`
		HighlightDY     int   `json:"highlight_dy"`
		TextDY          int   `json:"text_dy"`
		ColPitch        int   `json:"col_pitch"`
		RowPitch        int   `json:"row_pitch"`
		Ink             uint8 `json:"ink"`
		Shadow          uint8 `json:"shadow"`
		HighlightInk    uint8 `json:"highlight_ink"`
		DisabledInk     uint8 `json:"disabled_ink"`
		Fill            uint8 `json:"fill"`
	} `json:"layout"`
}

// LoadCommandMenuSpec 讀戰場命令選單的版面。
func LoadCommandMenuSpec(path string) (*CommandMenuSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s CommandMenuSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ItemCount 回某一張選單有幾項；沒有那一張回 0。
func (s *CommandMenuSpec) ItemCount(kind CommandMenuKind) int {
	return len(s.Menus[kind].Items)
}

// Size 回一張有 rows 項的選單的框大小。
//
// 框大小 ＝ (內寬 × 16 ＋ 30) × (項目數 × 20 ＋ 38)，三張收據驗過
// （`docs/re/117`）。
func (s *CommandMenuSpec) Size(rows int) (int, int) {
	l := &s.Layout
	return s.Width*l.CellW + l.FramePadW, rows*l.CellH + l.FramePadH
}

// CommandMenuAt 回把有 rows 項的選單開在游標 (mx,my) 時，框的左上角。
//
// 位置是**格**不是像素：格 ＝ 游標所在格 − (1,1)，再夾成
// 「格y ≤ 19 − 項目數」讓整個框留在戰場那一塊裡面。
func (s *CommandMenuSpec) CommandMenuAt(mx, my, rows int) (int, int) {
	l := &s.Layout
	cx := mx/l.CellW + l.CursorCellBias
	cy := my/l.CellH + l.CursorCellBias
	if cx < 0 {
		cx = 0
	}
	if cy < 0 {
		cy = 0
	}
	if max := l.ClampRows - rows; cy > max {
		cy = max
	}
	return cx*l.CellW + l.OriginBias, cy*l.CellH + l.OriginBias
}

// CommandMenuRowAt 回畫面座標 (x,y) 落在第幾項，-1 ＝ 不在任何一項上。
func (s *CommandMenuSpec) CommandMenuRowAt(left, top, x, y, rows int) int {
	l := &s.Layout
	hx := left + l.HighlightDX
	if x < hx || x >= hx+s.Width*l.CellW {
		return -1
	}
	first := top + l.HighlightDY
	if y < first {
		return -1
	}
	r := (y - first) / l.RowPitch
	if r < 0 || r >= rows {
		return -1
	}
	return r
}

// DrawCommandMenu 把命令選單畫在 (x,y)（框的左上角）。
//
// sel 是游標所在的那一項（-1 ＝ 不在任何一項上）；enabled 逐項對應
// 「可不可選」。不可選的項目字色 8，**而且游標停在上面也不畫反白帶**
// （收據 `s3cmdoff`）。回沒有字模的字碼。
func DrawCommandMenu(s *CommandMenuSpec, fb *gfx.Framebuffer, fnt *font.Font,
	kind CommandMenuKind, enabled []bool, x, y, sel int) []uint16 {
	items := s.Menus[kind].Items
	if len(items) == 0 {
		return nil
	}
	l := &s.Layout
	w, h := s.Size(len(items))
	f := ui.MenuFrame()
	f.Fill = l.Fill
	f.Draw(fb, x, y, w, h)

	hx, hy := x+l.HighlightDX, y+l.HighlightDY
	on := func(r int) bool { return r >= len(enabled) || enabled[r] }
	if sel >= 0 && sel < len(items) && on(sel) {
		fb.FillRect(hx, hy+sel*l.RowPitch, s.Width*l.CellW, l.RowPitch, l.HighlightInk)
	}
	var missing []uint16
	for r, item := range items {
		ink := l.Ink
		if !on(r) {
			ink = l.DisabledInk
		}
		for c, code := range item.Codes {
			if !ui.DrawGlyph(fb, fnt, hx+c*l.ColPitch, hy+l.TextDY+r*l.RowPitch,
				code, ink, l.Shadow, -1, false) {
				missing = append(missing, code)
			}
		}
	}
	return missing
}
