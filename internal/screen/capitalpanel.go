package screen

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// CapitalPanel 是戰略地圖介面右側那塊：右上的日期與錢，下面自己首都的摘要
// （規格 `docs/spec/002`，座標在 `docs/internals/data/capital-panel.json`）。
//
// 最上面那一塊摘要與左下角**是同一個版面**，只差 (+320, −200)，
// 所以直接借 `RegionPanel.DrawAt` 畫。
type CapitalPanel struct {
	// Summary 是首都摘要相對左下角面板的位移。
	Summary struct{ DX, DY int } `json:"summary"`

	FactionLine struct {
		Y       int `json:"y"`
		PrefixX int `json:"prefix_x"`
		Prefix  []struct {
			Faction int      `json:"faction"`
			Text    string   `json:"text"`
			Codes   []uint16 `json:"codes"`
		} `json:"prefix"`
		EraX      int      `json:"era_x"`
		Era       []uint16 `json:"era"`
		YearRight int      `json:"year_right"`
		YearMarkX int      `json:"year_mark_x"`
		YearMark  uint16   `json:"year_mark"`
		MonthRight int     `json:"month_right"`
		MonthMarkX int     `json:"month_mark_x"`
		MonthMark  uint16  `json:"month_mark"`
		DayRight   int     `json:"day_right"`
		DayMarkX   int     `json:"day_mark_x"`
		DayMark    uint16  `json:"day_mark"`
		Ink        uint8   `json:"ink"`
	} `json:"faction_line"`

	Rows []struct {
		Y     int         `json:"y"`
		Left  *PanelField `json:"left"`
		Right *PanelField `json:"right"`
	} `json:"rows"`

	Ink uint8 `json:"ink"`
}

// PanelField 是「標籤 ＋ 一個靠右對齊的數字」。
type PanelField struct {
	LabelX int      `json:"label_x"`
	Label  []uint16 `json:"label"`
	Text   string   `json:"text"`
	Right  int      `json:"right"`
	Field  string   `json:"field"`

	// Format 是 `per_mille` 時印成 `4.5%`（千分之一 → 一位小數的百分比）。
	Format string `json:"format"`
}

// LoadCapitalPanel 讀右側面板的版面。
func LoadCapitalPanel(path string) (*CapitalPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p CapitalPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Draw 把右側面板畫出來。region 是左下角那塊的版面（首都摘要借它畫）。
func (p *CapitalPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, region *RegionPanel,
	c rules.CapitalSummary, faction int) {
	ink := p.FactionLine.Ink
	for _, pre := range p.FactionLine.Prefix {
		if pre.Faction == faction {
			drawFull(fb, fnt, p.FactionLine.PrefixX, p.FactionLine.Y, pre.Codes,
				rules.FactionColour(faction))
		}
	}
	f := p.FactionLine
	drawFull(fb, fnt, f.EraX, f.Y, f.Era, ink)
	drawRight(fb, fnt, f.YearRight, f.Y, strconv.Itoa(c.Year), ink)
	drawFull(fb, fnt, f.YearMarkX, f.Y, []uint16{f.YearMark}, ink)
	drawRight(fb, fnt, f.MonthRight, f.Y, strconv.Itoa(c.Month), ink)
	drawFull(fb, fnt, f.MonthMarkX, f.Y, []uint16{f.MonthMark}, ink)
	drawRight(fb, fnt, f.DayRight, f.Y, strconv.Itoa(c.Day), ink)
	drawFull(fb, fnt, f.DayMarkX, f.Y, []uint16{f.DayMark}, ink)

	for _, r := range p.Rows {
		for _, fld := range []*PanelField{r.Left, r.Right} {
			if fld == nil {
				continue
			}
			drawFull(fb, fnt, fld.LabelX, r.Y, fld.Label, p.Ink)
			drawRight(fb, fnt, fld.Right, r.Y, fld.value(c), p.Ink)
		}
	}

	region.DrawAt(fb, fnt, c.Region, faction, p.Summary.DX, p.Summary.DY)
}

// value 回這一格要印的字串。
func (f *PanelField) value(c rules.CapitalSummary) string {
	var v int
	switch f.Field {
	case "military":
		v = c.Military
	case "tax_total":
		v = c.TaxTotal
	case "defence":
		v = c.Defence
	case "economy":
		v = c.Economy
	case "tax":
		v = c.Tax
	case "growth_per_mille":
		v = c.GrowthPerMille
	case "tax_rate_per_mille":
		v = c.TaxRatePerMille
	}
	if f.Format == "per_mille" {
		// 千分之一印成一位小數的百分比：45 → `4.5%`。
		return fmt.Sprintf("%d.%d%%", v/10, v%10)
	}
	return strconv.Itoa(v)
}
