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
// 編隊那一行格子。
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
	} `json:"units"`
	Mobility struct {
		Y     int
		Right int
	} `json:"mobility"`
	Formation struct {
		X, Y  int
		Pitch int
		Empty uint16
	} `json:"formation"`
	Buttons []struct {
		Action     string
		Text       string
		X, Y, W, H int
	} `json:"buttons"`

	// UnitLetters ＝ 原版 `ds:6172`：艦種 → 全形代號字母
	// （Ｂ Ｈ Ｃ Ａ Ｄ Ｔ Ｅ Ｐ，`docs/internals/56`）。
	UnitLetters struct {
		Codes []uint16 `json:"codes"`
	} `json:"unit_letters"`

	// StrengthInk 是字母的顏色：滿編、半編以上、半編以下三段
	// （`0FF4:03FE`，`docs/internals/56`）。
	StrengthInk struct {
		Full uint8 `json:"full"`
		Half uint8 `json:"half"`
		Low  uint8 `json:"low"`
	} `json:"strength_ink"`

	// GridSlots ＝ 原版 `ds:28B8`：隊形槽 0–40 各畫在格盤的哪一格。
	GridSlots struct {
		Cells []struct {
			Index int `json:"index"`
			Col   int `json:"col"`
			Row   int `json:"row"`
		} `json:"cells"`
	} `json:"grid_slots"`
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

// CompositionGroup 是編隊那一行的一格：一組編制。
//
// 對到原版 `ds:981E` 的一筆 4 bytes（旗艦在第 0 筆，其餘依艦種、
// 隊形槽排序）——`Kind` 是 `[2]`、`Slot` 是 `[3]`，`Ships`／`Full`
// 是那一組艦艇記錄的 `+0x08`／`+0x0A`（`docs/internals/56`）。
type CompositionGroup struct {
	Kind  int // 艦種
	Slot  int // 隊形槽 0–40
	Ships int // 現有艘數
	Full  int // 滿編艘數
}

// CompositionView 是畫這張畫面要的一整組值。
type CompositionView struct {
	Ships    []int // 七種艦艇的艘數
	PerUnit  []int // 七種艦艇一單位幾艘
	Mobility int   // 機動力（%）

	// Groups 是已經編成的那幾組（＝ 部隊數的分子，原版 `ds:984B`）。
	Groups []CompositionGroup
	// Slots ＝ 編制槽的應有數（艦隊 `+0x0E`），也就是編隊那一行畫幾格、
	// 部隊數的分母。**不是常數 8**（`docs/re/293`）。
	Slots int
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
	drawRight(fb, fnt, p.Units.NumeratorRight, p.Units.Y, strconv.Itoa(len(v.Groups)), p.Ink)
	drawHalf(fb, fnt, p.Units.SlashX, p.Units.Y, '/', p.Ink)
	drawRight(fb, fnt, p.Units.DenominatorRight, p.Units.Y, strconv.Itoa(v.Slots), p.Ink)
	drawRight(fb, fnt, p.Mobility.Right, p.Mobility.Y, strconv.Itoa(v.Mobility)+"%", p.Ink)

	// 編隊那一行：前 len(Groups) 格印艦種的代號字母，其餘印「□」。
	// 格數是 v.Slots ＝ 艦隊 `+0x0E`（原版 `ds:984B ＋ ds:984C`）。
	for i := 0; i < v.Slots; i++ {
		code, ink := p.Formation.Empty, p.Ink
		if i < len(v.Groups) {
			g := v.Groups[i]
			if c, ok := p.UnitLetter(g.Kind); ok {
				code, ink = c, p.StrengthInkFor(g.Ships, g.Full)
			}
		}
		drawFull(fb, fnt, p.Formation.X+i*p.Formation.Pitch, p.Formation.Y,
			[]uint16{code}, ink)
	}
}

// UnitLetter 回艦種的全形代號字母（原版 `ds:6172` 第 kind 筆）。
func (p *CompositionPanel) UnitLetter(kind int) (uint16, bool) {
	if kind < 0 || kind >= len(p.UnitLetters.Codes) {
		return 0, false
	}
	return p.UnitLetters.Codes[kind], true
}

// StrengthInkFor 是字母的顏色：原版 `0FF4:03FE` 拿那一組艦艇記錄的
// 現有艘數與滿編艘數比，分成三段（`docs/internals/56`）。
func (p *CompositionPanel) StrengthInkFor(ships, full int) uint8 {
	switch {
	case ships == full:
		return p.StrengthInk.Full
	case ships*2 < full:
		return p.StrengthInk.Low
	default:
		return p.StrengthInk.Half
	}
}

// GridCell 回隊形槽 slot 畫在格盤的第幾欄第幾列（原版 `ds:28B8`）。
func (p *CompositionPanel) GridCell(slot int) (col, row int, ok bool) {
	for _, c := range p.GridSlots.Cells {
		if c.Index == slot {
			return c.Col, c.Row, true
		}
	}
	return 0, 0, false
}

// ButtonAt 回畫面座標 (x,y) 落在哪一顆按鈕上；沒有回空字串。
// 矩形是熱區，比畫出來的框左右各窄 2 像素（`docs/re/68`）。
func (p *CompositionPanel) ButtonAt(x, y int) string {
	for _, b := range p.Buttons {
		if x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H {
			return b.Action
		}
	}
	return ""
}
