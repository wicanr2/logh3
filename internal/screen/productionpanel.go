package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// ProductionPanel 是生產計劃那個視窗裡的數字（規格 `docs/spec/004`）。
//
// 七列各六個數字：百分比、金額、船艦數、部隊，加上「建造中」的船艦數與部隊
// ——建造中那兩欄是**上一份計劃**買到的船，月底會照這個數交付
// （`docs/re/21`）。
type ProductionPanel struct {
	Ink  uint8 `json:"ink"`
	Cost struct {
		Y     int `json:"y"`
		Right int `json:"right"`
	} `json:"cost"`
	RowsY   []int `json:"rows_y"`
	Columns struct {
		Pct           int `json:"pct"`
		Amount        int `json:"amount"`
		Ships         int `json:"ships"`
		Units         int `json:"units"`
		Building      int `json:"building"`
		BuildingUnits int `json:"building_units"`
	} `json:"columns"`
	Rest struct {
		Y      int `json:"y"`
		Pct    int `json:"pct"`
		Amount int `json:"amount"`
	} `json:"rest"`
}

// LoadProductionPanel 讀版面。
func LoadProductionPanel(path string) (*ProductionPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p ProductionPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Draw 把數字畫上去。plan 是次月份的計劃，building 是上一份計劃
// （買到的船這個月還在造，月底交付）。
func (p *ProductionPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, plan, building rules.BuildPlan) {
	c := p.Columns
	drawRight(fb, fnt, p.Cost.Right, p.Cost.Y, strconv.Itoa(plan.Cost), p.Ink)
	for i, y := range p.RowsY {
		drawRight(fb, fnt, c.Pct, y, strconv.Itoa(plan.Pct[i])+"%", p.Ink)
		drawRight(fb, fnt, c.Amount, y, strconv.Itoa(plan.Amount(i)), p.Ink)
		drawRight(fb, fnt, c.Ships, y, strconv.Itoa(plan.Ships(i)), p.Ink)
		drawRight(fb, fnt, c.Units, y, strconv.Itoa(plan.Units(i)), p.Ink)
		drawRight(fb, fnt, c.Building, y, strconv.Itoa(building.Ships(i)), p.Ink)
		drawRight(fb, fnt, c.BuildingUnits, y, strconv.Itoa(building.Units(i)), p.Ink)
	}
	drawRight(fb, fnt, p.Rest.Pct, p.Rest.Y, strconv.Itoa(plan.RestPct())+"%", p.Ink)
	drawRight(fb, fnt, p.Rest.Amount, p.Rest.Y, strconv.Itoa(plan.RestAmount()), p.Ink)
}
