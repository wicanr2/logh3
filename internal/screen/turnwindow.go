package screen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 回合處理中的過場視窗（`056A:03E0` 開的那一個，`docs/re/332`、規格 059）：
// 每支人類座位的艦隊走一格印一張「出發前往／已經到達」，走不了印被擋或
// 經費不足。

// TurnWindowSpec 是版面與格式字串（`docs/internals/data/turn-window-panel.json`）。
type TurnWindowSpec struct {
	CellX          int               `json:"cell_x"`
	CellY          int               `json:"cell_y"`
	Cols           int               `json:"cols"`
	Rows           int               `json:"rows"`
	Style          int               `json:"style"`
	Colour         int               `json:"colour"`
	RegionInk      []int             `json:"region_ink"`
	RegionWidth    int               `json:"region_width"`
	FirstWaitShift int               `json:"first_wait_shift"`
	DepartColour   int               `json:"depart_colour"`
	Formats        map[string]string `json:"formats"`
	ReportFormats  map[string]string `json:"report_formats"`
	BusyTitle      struct {
		CellX int    `json:"cell_x"`
		CellY int    `json:"cell_y"`
		Cols  int    `json:"cols"`
		Rows  int    `json:"rows"`
		Text  string `json:"text"`
	} `json:"busy_title"`
	// ThinkTitle 是「戰略模式思考中」：列 Row 的字與停留的 tick 數（規格 059）。
	ThinkTitle struct {
		CellX int    `json:"cell_x"`
		CellY int    `json:"cell_y"`
		Cols  int    `json:"cols"`
		Rows  int    `json:"rows"`
		Row   int    `json:"row"`
		Text  string `json:"text"`
		Ticks int    `json:"ticks"`
	} `json:"think_title"`
	// BattleReport 是戰鬥中的士氣報告視窗（`3826:0405`）：格座標、內容格數、頭像左上角。
	BattleReport struct {
		CellX int `json:"cell_x"`
		CellY int `json:"cell_y"`
		Cols  int `json:"cols"`
		Rows  int `json:"rows"`
		Face  struct {
			X int `json:"x"`
			Y int `json:"y"`
		} `json:"face"`
	} `json:"battle_report"`
	Face struct {
		X, Y, W, H int
	} `json:"face"`
}

// LoadTurnWindowSpec 讀版面。
func LoadTurnWindowSpec(path string) (*TurnWindowSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s TurnWindowSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Cols == 0 || s.Rows == 0 || len(s.Formats) < 8 || len(s.RegionInk) != 3 {
		return nil, fmt.Errorf("turn-window-panel：規格不完整（%d×%d、%d 條格式字串、%d 個地名色）",
			s.Cols, s.Rows, len(s.Formats), len(s.RegionInk))
	}
	return &s, nil
}

// TurnWindowKind 是這張過場印哪一種本文。
type TurnWindowKind int

const (
	TurnWindowDepart  TurnWindowKind = iota // 113A:0B46
	TurnWindowBlocked                       // 113A:0797
	TurnWindowNoFunds                       // 113A:0A4E
)

// TurnWindowArrival 是出發前往第二段加的那一句。
type TurnWindowArrival int

const (
	TurnWindowOnward     TurnWindowArrival = iota // 沒到目的地：不加
	TurnWindowArrived                             // ds:2ED7
	TurnWindowGarrisoned                          // ds:2EEF
)

// TurnWindowView 是這張過場要印的值。
type TurnWindowView struct {
	Kind        TurnWindowKind
	FactionName []byte // 帶色的陣營名
	Number      int
	AdmiralName []byte
	Rank        []byte
	// 出發前往那一種才用：兩格的名字與擁有者、第二段的到達句。
	FromName  []byte
	FromOwner int
	NextName  []byte
	NextOwner int
	Second    bool // 第二段（地圖已換、印到達句）
	Arrival   TurnWindowArrival
}

func (s *TurnWindowSpec) fmtBytes(name string) []byte { return decodeHex(s.Formats[name]) }

// FirstWait 回出發前往第一段等幾個 tick（`ds:8F38 >> 2`）。
func (s *TurnWindowSpec) FirstWait(msgSpeed int) int { return msgSpeed >> s.FirstWaitShift }

// regionName 是 `3982:0263`：擁有者色、名字補到固定寬、還原成視窗色。
func (s *TurnWindowSpec) regionName(b *ui.TextBuffer, name []byte, owner int) {
	if owner < 0 || owner >= len(s.RegionInk) {
		owner = len(s.RegionInk) - 1
	}
	b.Printf(s.fmtBytes("region_name"), s.RegionInk[owner], s.RegionWidth, name, s.Colour)
}

// Buffer 把整張過場排進內容緩衝區（`0B46`／`0797`／`0A4E` 的列印順序）。
func (s *TurnWindowSpec) Buffer(v *TurnWindowView) *ui.TextBuffer {
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	b.Printf(s.fmtBytes("fleet_prefix"), v.FactionName)
	// `3982:0038`：小於 10 全形、否則 `%2d` 兩個半形（`docs/re/143`；收據 `fcmd-s3-w2a` 第１３艦隊）。
	b.Write(s.FleetNumber(v.Number))
	b.Printf(s.fmtBytes("fleet_suffix"), v.AdmiralName, v.Rank)
	switch v.Kind {
	case TurnWindowBlocked:
		b.Write(s.fmtBytes("blocked"))
	case TurnWindowNoFunds:
		b.Write(s.fmtBytes("no_funds"))
	default:
		s.regionName(b, v.FromName, v.FromOwner)
		b.Write(s.fmtBytes("depart"))
		s.regionName(b, v.NextName, v.NextOwner)
		if v.Second {
			switch v.Arrival {
			case TurnWindowArrived:
				b.Write(s.fmtBytes("arrived"))
			case TurnWindowGarrisoned:
				b.Write(s.fmtBytes("garrisoned"))
			}
		}
	}
	return b
}

// Draw 把過場疊在畫面上：外框、排字、頭像。回沒有字模的字碼。
func (s *TurnWindowSpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	v *TurnWindowView, face *asset.Image) []uint16 {
	if s == nil || fb == nil || v == nil {
		return nil
	}
	if tiles != nil {
		(ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}).
			Draw(fb, s.CellX, s.CellY, s.Cols+2, s.Rows+2)
	}
	missing := ui.DrawTextBuffer(fb, fnt, tiles, s.Buffer(v), s.CellX+1, s.CellY+1)
	if face != nil {
		for y := 0; y < s.Face.H && y < face.H; y++ {
			for x := 0; x < s.Face.W && x < face.W; x++ {
				fb.Set(s.Face.X+x, s.Face.Y+y, face.Pix[y*face.W+x])
			}
		}
	}
	return missing
}

// DrawBusyTitle 把右上標題窗換成「戰略模式處理中」一列（`056A:011C`）。
func (s *TurnWindowSpec) DrawBusyTitle(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles) {
	if s == nil || fb == nil || s.BusyTitle.Cols == 0 {
		return
	}
	t := s.BusyTitle
	b := ui.NewTextBuffer(t.Cols, t.Rows, uint8(s.Colour))
	b.Write(decodeHex(t.Text))
	ui.DrawTextBuffer(fb, fnt, tiles, b, t.CellX+1, t.CellY+1)
}

// DrawThinkTitle 是 `04F8:0355(1)` 的標題窗：整窗清空，列 Row 印「戰略模式思考中」。
// 列 0 的陣營與日期由呼叫端照命令階段那一行畫。
func (s *TurnWindowSpec) DrawThinkTitle(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles) {
	if s == nil || fb == nil || s.ThinkTitle.Cols == 0 {
		return
	}
	t := s.ThinkTitle
	b := ui.NewTextBuffer(t.Cols, t.Rows, uint8(s.Colour))
	b.Write(append(bytes.Repeat([]byte{'\n'}, t.Row), decodeHex(t.Text)...))
	ui.DrawTextBuffer(fb, fnt, tiles, b, t.CellX+1, t.CellY+1)
}

// ReportFormat 回報告用的一條格式字串（`report_formats`）。
func (s *TurnWindowSpec) ReportFormat(name string) []byte { return decodeHex(s.ReportFormats[name]) }

// FleetNumber 是 `3982:0073`（過場視窗的 `3982:0038` 同一套）：編號小於 10 印全形（`\x06%d\x07`），否則兩個半形。
func (s *TurnWindowSpec) FleetNumber(n int) []byte {
	if n < 10 {
		return ui.Sprintf(s.ReportFormat("number_small"), n)
	}
	return ui.Sprintf(s.ReportFormat("number_large"), n)
}

// PlainName 是 `3982:0296(…, 擁有者, 0, 記錄)`：擁有者色的名字，不補寬。
func (s *TurnWindowSpec) PlainName(name []byte, owner int) []byte {
	if owner < 0 || owner >= len(s.RegionInk) {
		owner = len(s.RegionInk) - 1
	}
	return ui.Sprintf(s.ReportFormat("name_plain"), s.RegionInk[owner], name, s.Colour)
}

// DrawReport 是 `3826:0307`：同一個矩形開窗、清空、印 text、畫頭像。
func (s *TurnWindowSpec) DrawReport(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	text []byte, face *asset.Image) []uint16 {
	if s == nil {
		return nil
	}
	return s.drawReportIn(fb, fnt, tiles, s.CellX, s.CellY, s.Cols, s.Rows, s.Face.X, s.Face.Y, text, face)
}

// DrawBattleReport 是 `3826:0405`：戰鬥中（`ds:8F16 != 0`）的士氣報告。每次在堆疊上開臨時視窗
// `3603:0006(&local, 0x97, 2, 7, 24, 5, 1)`、清空、印、頭像 `3958:0215(提督, 0x2C, 0x28)`、等、關
// （`docs/re/336` §七、§八）。文字與戰略階段那一則相同。
func (s *TurnWindowSpec) DrawBattleReport(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	text []byte, face *asset.Image) []uint16 {
	if s == nil || s.BattleReport.Cols == 0 {
		return nil
	}
	r := s.BattleReport
	return s.drawReportIn(fb, fnt, tiles, r.CellX, r.CellY, r.Cols, r.Rows, r.Face.X, r.Face.Y, text, face)
}

func (s *TurnWindowSpec) drawReportIn(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	cx, cy, cols, rows, faceX, faceY int, text []byte, face *asset.Image) []uint16 {
	if fb == nil {
		return nil
	}
	if tiles != nil {
		(ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}).
			Draw(fb, cx, cy, cols+2, rows+2)
	}
	b := ui.NewTextBuffer(cols, rows, uint8(s.Colour))
	b.Write(text)
	missing := ui.DrawTextBuffer(fb, fnt, tiles, b, cx+1, cy+1)
	if face != nil {
		for y := 0; y < s.Face.H && y < face.H; y++ {
			for x := 0; x < s.Face.W && x < face.W; x++ {
				fb.Set(faceX+x, faceY+y, face.Pix[y*face.W+x])
			}
		}
	}
	return missing
}
