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

// 回合處理中占領命令的問話（`113A:0899`，`docs/re/322`）：占領命令的下一格
// 不是自己的、也不是目的地時，停下來問「<目的地>無法直接到達此星系。
// <下一格>要占領此星系嗎？」，兩顆鈕：占領／待命。

// OccupyAskSpec 是版面與格式字串（`docs/internals/data/occupy-ask-panel.json`）。
type OccupyAskSpec struct {
	CellX        int               `json:"cell_x"`
	CellY        int               `json:"cell_y"`
	Cols         int               `json:"cols"`
	Rows         int               `json:"rows"`
	Style        int               `json:"style"`
	Colour       int               `json:"colour"`
	ButtonColour int               `json:"button_colour"`
	Formats      map[string]string `json:"formats"`
	Hotspots     []struct {
		CellX int    `json:"cell_x"`
		CellY int    `json:"cell_y"`
		Cols  int    `json:"cols"`
		Rows  int    `json:"rows"`
		Code  int    `json:"code"`
		Label string `json:"label"`
	} `json:"hotspots"`
	Face struct {
		X, Y, W, H int
	} `json:"face"`
}

// LoadOccupyAskSpec 讀版面。
func LoadOccupyAskSpec(path string) (*OccupyAskSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s OccupyAskSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Cols == 0 || s.Rows == 0 || len(s.Formats) < 4 || len(s.Hotspots) != 2 {
		return nil, fmt.Errorf("occupy-ask-panel：規格不完整（%d×%d、%d 條格式字串、%d 筆熱區）",
			s.Cols, s.Rows, len(s.Formats), len(s.Hotspots))
	}
	return &s, nil
}

// OccupyAskView 是這張問話要印的值。
type OccupyAskView struct {
	FactionName []byte // 帶色的陣營名（`3982:00B6`）
	Number      int
	AdmiralName []byte
	Rank        []byte
	TargetName  []byte // 目的地星系
	NextName    []byte // 下一格星系
}

func (s *OccupyAskSpec) fmtBytes(name string) []byte { return decodeHex(s.Formats[name]) }

// Buffer 把整張問話排進內容緩衝區（`113A:0899` 的列印順序）。
func (s *OccupyAskSpec) Buffer(v *OccupyAskView) *ui.TextBuffer {
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	b.Printf(s.fmtBytes("fleet_prefix"), v.FactionName)
	b.Attr |= ui.AttrWiden
	b.Printf([]byte("%d"), v.Number)
	b.Attr &^= ui.AttrWiden
	b.Printf(s.fmtBytes("fleet_suffix"), v.AdmiralName, v.Rank)
	b.Printf(s.fmtBytes("unreachable"), v.TargetName)
	b.Printf(s.fmtBytes("ask"), v.NextName)
	for _, h := range s.Hotspots {
		putRawButton(b, h.CellX, h.CellY, h.Cols, uint8(s.ButtonColour), decodeHex(h.Label))
	}
	return b
}

// Draw 把問話疊在畫面上：開窗、排字與鈕、頭像。
func (s *OccupyAskSpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	v *OccupyAskView, face *asset.Image) []uint16 {
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

// ButtonAt 回畫面座標 (x, y) 落在哪顆鈕上（熱區碼），沒有回 −1。
func (s *OccupyAskSpec) ButtonAt(x, y int) int {
	for _, h := range s.Hotspots {
		bx := (s.CellX + 1 + h.CellX) * ui.FrameTileW
		by := (s.CellY+1+h.CellY)*ui.FrameTileH + ui.FrameCellOriginY
		if x >= bx && x < bx+h.Cols*ui.FrameTileW && y >= by && y < by+h.Rows*ui.FrameTileH {
			return h.Code
		}
	}
	return -1
}
