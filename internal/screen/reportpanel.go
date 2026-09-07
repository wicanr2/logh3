package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// ReportPanel 是主畫面那份「會計報告」裡跟局面走的東西：署名的參謀、
// 他的頭像，以及五個金額（規格 `docs/spec/011`）。
//
// 視窗其餘的字（標題、「內政專任參謀」、各列的名目、按鈕）是版面，
// 留在畫面規格裡。
type ReportPanel struct {
	Ink   uint8 `json:"ink"`
	Right int   `json:"right"`
	Staff struct {
		Y int `json:"y"`
		X int `json:"x"`
	} `json:"staff"`
	Face struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"face"`
	Rows []struct {
		Y     int    `json:"y"`
		Field string `json:"field"`
	} `json:"rows"`
}

// LoadReportPanel 讀版面。
func LoadReportPanel(path string) (*ReportPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p ReportPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Draw 把署名與五個金額畫上去。頭像另外畫（要 `FACE.AMG`，見
// `ReportPanel.Face`）。
func (p *ReportPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, b rules.Budget,
	name, rank []uint16) {
	if len(name) > 0 {
		codes := append(append([]uint16(nil), name...), rank...)
		drawFull(fb, fnt, p.Staff.X, p.Staff.Y, codes, p.Ink)
	}
	for _, r := range p.Rows {
		var v int
		switch r.Field {
		case "treasury":
			v = b.Treasury
		case "tax":
			v = b.Tax
		case "military":
			v = b.Military()
		case "ship_cost":
			v = b.ShipCost()
		case "rest":
			v = b.Rest()
		}
		drawRight(fb, fnt, p.Right, r.Y, strconv.Itoa(v), p.Ink)
	}
}
