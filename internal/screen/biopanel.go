package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// BioPanel 是「提督經歷」那個視窗裡的內文版面（規格 `docs/spec/012`）。
//
// `KEIREKI.DAT` 那一段字裡沒有換行控制碼，斷行是遊戲自己算的：
// 一列固定 Columns 個全形字，最多 Rows 列。
type BioPanel struct {
	Ink     uint8 `json:"ink"`
	X, Y    int
	Pitch   int `json:"pitch"`
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

// LoadBioPanel 讀版面。
func LoadBioPanel(path string) (*BioPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p BioPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Draw 把一段人物傳排進視窗。
func (p *BioPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, codes []uint16) {
	if p.Columns <= 0 {
		return
	}
	for row := 0; row*p.Columns < len(codes); row++ {
		if p.Rows > 0 && row >= p.Rows {
			break
		}
		end := (row + 1) * p.Columns
		if end > len(codes) {
			end = len(codes)
		}
		drawFull(fb, fnt, p.X, p.Y+row*p.Pitch, codes[row*p.Columns:end], p.Ink)
	}
}
