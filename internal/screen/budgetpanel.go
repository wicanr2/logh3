package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// BudgetPanel 是「次月份預算計劃」與「徵稅計劃」那個視窗裡的數字
// （規格 `docs/spec/003`）。兩張畫面共用同一份版面，只有底層不同。
type BudgetPanel struct {
	Ink   uint8 `json:"ink"`
	Right int   `json:"right"`
	Rows  []struct {
		Y        int    `json:"y"`
		Field    string `json:"field"`
		Text     string `json:"text"`
		PctRight int    `json:"pct_right"`
	} `json:"rows"`
}

// LoadBudgetPanel 讀版面。
func LoadBudgetPanel(path string) (*BudgetPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p BudgetPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Draw 把數字畫進視窗。t 是這個月的財政、b 是次月份的預算。
func (p *BudgetPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, t rules.Treasury, b rules.Budget) {
	for _, r := range p.Rows {
		var v int
		switch r.Field {
		case "budget":
			v = t.Budget
		case "ship_paid":
			v = t.ShipPaid
		case "fleet_upkeep":
			v = t.FleetUpkeep
		case "movement":
			v = t.Movement
		case "next_fleet_upkeep":
			v = t.NextUpkeep
		case "next_movement":
			v = t.NextMovement
		case "tax":
			v = b.Tax
		case "military":
			// 預估軍事費要扣掉上面那四格（`docs/re/239`）。
			v = b.Military() - t.Paid()
		case "ship_cost":
			// **船艦建造費用的是扣之前的數**：收據 `fiscalrows5` 把
			// 預估軍事費壓成 −5925，這一格仍然是 15185（`docs/re/239`）。
			v = b.ShipCost()
		case "rest":
			v = b.Military() - t.Paid() - b.ShipCost()
		}
		drawRight(fb, fnt, p.Right, r.Y, strconv.Itoa(v), p.Ink)
		if r.PctRight != 0 {
			drawRight(fb, fnt, r.PctRight, r.Y, strconv.Itoa(b.ShipPct)+"%", p.Ink)
		}
	}
}
