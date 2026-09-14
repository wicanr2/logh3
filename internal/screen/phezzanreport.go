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

type PhezzanReportPanel struct {
	CellX   int                      `json:"cell_x"`
	CellY   int                      `json:"cell_y"`
	Cols    int                      `json:"cols"`
	Rows    int                      `json:"rows"`
	Style   int                      `json:"style"`
	Colour  int                      `json:"colour"`
	Face    struct{ X, Y, W, H int } `json:"face"`
	Reports map[string]struct {
		Pages []struct {
			Raw string `json:"raw"`
		} `json:"pages"`
	} `json:"reports"`
}

func LoadPhezzanReportPanel(path string) (*PhezzanReportPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p PhezzanReportPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *PhezzanReportPanel) Page(key string, page int) ([]byte, bool) {
	r, ok := p.Reports[key]
	if !ok || page < 0 || page >= len(r.Pages) {
		return nil, false
	}
	b, err := hex.DecodeString(r.Pages[page].Raw)
	return b, err == nil
}

func (p *PhezzanReportPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font,
	tiles *ui.FrameTiles, key string, page int, face *asset.Image) []uint16 {
	if p == nil || fb == nil || tiles == nil {
		return nil
	}
	(ui.TileFrame{Tiles: tiles, Style: p.Style, Fill: uint8(p.Colour >> 4)}).
		Draw(fb, p.CellX, p.CellY, p.Cols+2, p.Rows+2)
	b := ui.NewTextBuffer(p.Cols, p.Rows, uint8(p.Colour))
	if raw, ok := p.Page(key, page); ok {
		b.Write(raw)
	}
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
