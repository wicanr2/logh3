package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// CompositionPanel 是「艦隊組成」那個視窗裡算出來的部分
// （規格 `docs/spec/015`）：七列的待命部隊與數量、部隊數、機動力、
// 編隊那八格。
//
// 「待命部隊」＝ 該星域的艦艇艘數 ÷ 一單位幾艘，「數量」印的就是
// 一單位幾艘（`docs/re/69`）。
type CompositionPanel struct {
	Ink  uint8 `json:"ink"`
	Rows struct {
		Y             int
		Pitch         int
		Count         int
		CountRight    int `json:"count_right"`
		QuantityRight int `json:"quantity_right"`
	} `json:"rows"`
	Units struct {
		Y                int
		NumeratorRight   int `json:"numerator_right"`
		SlashX           int `json:"slash_x"`
		DenominatorRight int `json:"denominator_right"`
		Slots            int
	} `json:"units"`
	Mobility struct {
		Y     int
		Right int
	} `json:"mobility"`
	Formation struct {
		X, Y  int
		Pitch int
		Slots int
		Empty uint16
	} `json:"formation"`
	UnitLabels []struct {
		Slot int
		Code uint16
		Tile int
	} `json:"unit_labels"`
}

// LoadCompositionPanel 讀版面。
func LoadCompositionPanel(path string) (*CompositionPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p CompositionPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// CompositionView 是畫這張畫面要的一整組值。
type CompositionView struct {
	Ships    []int // 七種艦艇的艘數
	PerUnit  []int // 七種艦艇一單位幾艘
	Units    int   // 已編成的部隊數
	Mobility int   // 機動力（%）
	Slots    []int // 每一格是哪一種艦艇；-1 ＝ 空
}

// Draw 把算出來的部分畫上去。格盤的圖塊另外畫（要 `SPRITE.SPR`）。
func (p *CompositionPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, v CompositionView) {
	for i := 0; i < p.Rows.Count; i++ {
		y := p.Rows.Y + i*p.Rows.Pitch
		if i < len(v.PerUnit) {
			per := v.PerUnit[i]
			if i < len(v.Ships) && per > 0 {
				drawRight(fb, fnt, p.Rows.CountRight, y, strconv.Itoa(v.Ships[i]/per), p.Ink)
			}
			drawRight(fb, fnt, p.Rows.QuantityRight, y, strconv.Itoa(per), p.Ink)
		}
	}
	drawRight(fb, fnt, p.Units.NumeratorRight, p.Units.Y, strconv.Itoa(v.Units), p.Ink)
	drawHalf(fb, fnt, p.Units.SlashX, p.Units.Y, '/', p.Ink)
	drawRight(fb, fnt, p.Units.DenominatorRight, p.Units.Y, strconv.Itoa(p.Units.Slots), p.Ink)
	drawRight(fb, fnt, p.Mobility.Right, p.Mobility.Y, strconv.Itoa(v.Mobility)+"%", p.Ink)

	for i := 0; i < p.Formation.Slots; i++ {
		code := p.Formation.Empty
		if i < len(v.Slots) {
			if c, ok := p.labelCode(i, v.Slots[i]); ok {
				code = c
			}
		}
		drawFull(fb, fnt, p.Formation.X+i*p.Formation.Pitch, p.Formation.Y,
			[]uint16{code}, p.Ink)
	}
}

// labelCode 回編隊第 slot 格該印哪一個字；空的或還沒解的回 false。
//
// **字母怎麼決定還沒解**：收據只有一個資料點（第 0 格是「Ｂ」），
// 所以對照表也只有那一格（`docs/internals/data/fleet-composition-panel.json`）。
func (p *CompositionPanel) labelCode(slot, ship int) (uint16, bool) {
	if ship < 0 {
		return 0, false
	}
	for _, l := range p.UnitLabels {
		if l.Slot == slot {
			return l.Code, true
		}
	}
	return 0, false
}

// TileAt 回格盤第 slot 格該畫哪一張 `SPRITE.SPR` 圖塊；沒有回 -1。
func (p *CompositionPanel) TileAt(slot int) int {
	for _, l := range p.UnitLabels {
		if l.Slot == slot {
			return l.Tile
		}
	}
	return -1
}
