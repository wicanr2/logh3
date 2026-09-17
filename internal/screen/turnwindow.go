package screen

import (
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
	BusyTitle      struct {
		CellX int    `json:"cell_x"`
		CellY int    `json:"cell_y"`
		Cols  int    `json:"cols"`
		Rows  int    `json:"rows"`
		Text  string `json:"text"`
	} `json:"busy_title"`
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
	b.Attr |= ui.AttrWiden
	b.Printf([]byte("%d"), v.Number)
	b.Attr &^= ui.AttrWiden
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
