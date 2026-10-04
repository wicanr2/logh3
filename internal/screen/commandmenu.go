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

	// 行星那兩張（`docs/re/133`）。有要塞（行星記錄 `+0x12` 非 0）的
	// 多一項「要塞砲」，內寬也從 2 變 3。
	CommandMenuPlanet       CommandMenuKind = "planet"        // ds:42EA 情報／補給／補充
	CommandMenuFortress     CommandMenuKind = "fortress"      // ds:4318 ＋要塞砲
	CommandMenuLoadFaction  CommandMenuKind = "load-faction"  // DS:022E前三項，088
	CommandMenuCombatWeapon CommandMenuKind = "combat-weapon" // 1ED2:03E9，內寬8
	CommandMenuCombatChoice CommandMenuKind = "combat-choice" // 1ED2:0616，內寬2
)

// CommandMenuSpec 是戰場命令選單的版面與五張表
// （`docs/internals/data/battle-command-menus.json`）。
//
// 開啟常式與戰術選單同一支，配色與框線也同一套，所以這裡只放每一張的
// 項目與內寬；框的位置與大小走 layout 的公式（`docs/re/117`）。
type CommandMenuSpec struct {
	Width int `json:"width"` // 預設內寬，單位是 16 像素
	Menus map[CommandMenuKind]struct {
		Address string  `json:"address"`
		When    string  `json:"when"`
		Width   int     `json:"width,omitempty"` // 0 ＝ 用上面那個預設值
		Items   []label `json:"items"`
	} `json:"menus"`
	Layout struct {
		CellW          int   `json:"cell_w"`
		CellH          int   `json:"cell_h"`
		OriginBias     int   `json:"origin_bias"`
		CursorCellBias int   `json:"cursor_cell_bias"`
		ClampRows      int   `json:"clamp_rows"`
		FramePadW      int   `json:"frame_pad_w"`
		FramePadH      int   `json:"frame_pad_h"`
		HighlightDX    int   `json:"highlight_dx"`
		HighlightDY    int   `json:"highlight_dy"`
		TextDY         int   `json:"text_dy"`
		ColPitch       int   `json:"col_pitch"`
		RowPitch       int   `json:"row_pitch"`
		Ink            uint8 `json:"ink"`
		Shadow         uint8 `json:"shadow"`
		HighlightInk   uint8 `json:"highlight_ink"`
		DisabledInk    uint8 `json:"disabled_ink"`
		Fill           uint8 `json:"fill"`
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

// WidthOf 回某一張選單的內寬。戰場那五張都是 2，行星那兩張是 2 與 3
// （`1CC5:0FB9`／`1CC5:0FCB`，`docs/re/133`）。
func (s *CommandMenuSpec) WidthOf(kind CommandMenuKind) int {
	if w := s.Menus[kind].Width; w > 0 {
		return w
	}
	return s.Width
}

// Size 回一張有 rows 項的選單的框大小。
//
// 框大小 ＝ (內寬 × 16 ＋ 30) × (項目數 × 20 ＋ 38)，三張收據驗過
// （`docs/re/117`）。
func (s *CommandMenuSpec) Size(kind CommandMenuKind, rows int) (int, int) {
	l := &s.Layout
	return s.WidthOf(kind)*l.CellW + l.FramePadW, rows*l.CellH + l.FramePadH
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
func (s *CommandMenuSpec) CommandMenuRowAt(kind CommandMenuKind, left, top, x, y, rows int) int {
	l := &s.Layout
	hx := left + l.HighlightDX
	if x < hx || x >= hx+s.WidthOf(kind)*l.CellW {
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
//
// 框與戰術選單同一組（`FRAME.BLK` 起始碼 9）；tiles 缺了就不畫框。
func DrawCommandMenu(s *CommandMenuSpec, fb *gfx.Framebuffer, fnt *font.Font,
	tiles *ui.FrameTiles, kind CommandMenuKind, enabled []bool, x, y, sel int) []uint16 {
	items := s.Menus[kind].Items
	return drawCommandMenuItems(s, fb, fnt, tiles, kind, items, enabled, x, y, sel)
}

// DrawCommandMenuLabels沿同一份幾何畫動態武器數量；不改共享規格。
func DrawCommandMenuLabels(s *CommandMenuSpec, fb *gfx.Framebuffer, fnt *font.Font,
	tiles *ui.FrameTiles, kind CommandMenuKind, codes [][]uint16, enabled []bool, x, y, sel int) []uint16 {
	items := make([]label, len(codes))
	for i := range codes {
		items[i].Codes = codes[i]
	}
	return drawCommandMenuItems(s, fb, fnt, tiles, kind, items, enabled, x, y, sel)
}

func drawCommandMenuItems(s *CommandMenuSpec, fb *gfx.Framebuffer, fnt *font.Font,
	tiles *ui.FrameTiles, kind CommandMenuKind, items []label, enabled []bool, x, y, sel int) []uint16 {
	if len(items) == 0 {
		return nil
	}
	l := &s.Layout
	w, h := s.Size(kind, len(items))
	ui.MenuFrame(tiles, l.Fill).DrawRect(fb, x, y, w, h)

	hx, hy := x+l.HighlightDX, y+l.HighlightDY
	on := func(r int) bool { return r >= len(enabled) || enabled[r] }
	if sel >= 0 && sel < len(items) && on(sel) {
		fb.FillRect(hx, hy+sel*l.RowPitch, s.WidthOf(kind)*l.CellW, l.RowPitch, l.HighlightInk)
	}
	var missing []uint16
	for r, item := range items {
		ink := l.Ink
		if !on(r) {
			ink = l.DisabledInk
		}
		for c, code := range item.Codes {
			if code < 0x8000 && kind == CommandMenuCombatWeapon {
				ui.DrawHalfGlyph(fb, fnt, hx+c*l.ColPitch, hy+l.TextDY+r*l.RowPitch, uint8(code>>8), ink, l.Shadow, -1, false)
				ui.DrawHalfGlyph(fb, fnt, hx+c*l.ColPitch+8, hy+l.TextDY+r*l.RowPitch, uint8(code), ink, l.Shadow, -1, false)
				continue
			}
			if !ui.DrawGlyph(fb, fnt, hx+c*l.ColPitch, hy+l.TextDY+r*l.RowPitch,
				code, ink, l.Shadow, -1, false) {
				missing = append(missing, code)
			}
		}
	}
	return missing
}
