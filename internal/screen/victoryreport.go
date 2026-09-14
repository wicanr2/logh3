package screen

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// VictoryReportPanel 是原版勝負確定後的模態報告（04C8:00F3）。
type VictoryReportPanel struct {
	CellX  int                      `json:"cell_x"`
	CellY  int                      `json:"cell_y"`
	Cols   int                      `json:"cols"`
	Rows   int                      `json:"rows"`
	Style  int                      `json:"style"`
	Colour int                      `json:"colour"`
	Face   struct{ X, Y, W, H int } `json:"face"`
	Button struct {
		CellX  int      `json:"cell_x"`
		CellY  int      `json:"cell_y"`
		Cols   int      `json:"cols"`
		Colour int      `json:"colour"`
		Codes  []string `json:"codes"`
	} `json:"button"`
	Formats      map[string]string `json:"formats"`
	FactionNames [2]string         `json:"faction_names"`
	PlainNames   [2]string         `json:"plain_names"`
}

func LoadVictoryReportPanel(path string) (*VictoryReportPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p VictoryReportPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func decodeHex(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

// Draw 畫指定人類座位看到的勝／敗頁。score 是整局累積戰略分數。
func (p *VictoryReportPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font,
	tiles *ui.FrameTiles, face *asset.Image, viewer int, won bool,
	year, month, day, regions int, score [2]int) []uint16 {
	if p == nil || fb == nil || tiles == nil || viewer < 0 || viewer > 1 {
		return nil
	}
	(ui.TileFrame{Tiles: tiles, Style: p.Style, Fill: uint8(p.Colour >> 4)}).
		Draw(fb, p.CellX, p.CellY, p.Cols+2, p.Rows+2)
	b := ui.NewTextBuffer(p.Cols, p.Rows, uint8(p.Colour))
	b.Printf(decodeHex(p.Formats["date"]), year, month, day)
	if won {
		b.Write(decodeHex(p.Formats["win_intro"]))
		b.Printf(decodeHex(p.Formats["win_name"]), decodeHex(p.FactionNames[viewer]))
		b.Printf(decodeHex(p.Formats["win_regions"]), regions)
	} else {
		b.Write(decodeHex(p.Formats["loss_intro"]))
		b.Printf(decodeHex(p.Formats["loss_name"]), decodeHex(p.FactionNames[viewer]))
	}
	opponent := 1 - viewer
	b.Printf(decodeHex(p.Formats["victory"]), score[viewer]-score[opponent])
	b.Printf(decodeHex(p.Formats["score"]), decodeHex(p.PlainNames[viewer]), score[viewer])
	b.Printf(decodeHex(p.Formats["score"]), decodeHex(p.PlainNames[opponent]), score[opponent])
	p.putButton(b)
	missing := ui.DrawTextBuffer(fb, fnt, tiles, b, p.CellX+1, p.CellY+1)
	if face != nil {
		for y := 0; y < p.Face.H && y < face.H; y++ {
			for x := 0; x < p.Face.W && x < face.W; x++ {
				fb.Set(p.Face.X+x, p.Face.Y+y, face.Pix[y*face.W+x])
			}
		}
	}
	return missing
}

func (p *VictoryReportPanel) putButton(b *ui.TextBuffer) {
	x, y := p.Button.CellX, p.Button.CellY
	if c := b.At(x-1, y); c != nil {
		c.Flags |= ui.CellButtonLeft
	}
	for i := 0; i < p.Button.Cols; i++ {
		if c := b.At(x+i, y); c != nil {
			c.Colour = uint8(p.Button.Colour)
			c.Flags |= ui.CellButtonBody
			if i < len(p.Button.Codes) {
				var code uint16
				if raw := decodeHex(p.Button.Codes[i]); len(raw) == 2 {
					code = uint16(raw[0])<<8 | uint16(raw[1])
				}
				c.Code = code
			}
		}
	}
	if c := b.At(x+p.Button.Cols, y); c != nil {
		c.Flags |= ui.CellButtonRight
	}
}
