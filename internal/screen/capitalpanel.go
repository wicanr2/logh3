package screen

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
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

	// Buttons 是中間那三顆命令按鈕（戰略命令／一覽表／系統指令）。
	// 立體邊與輸入框同一種：兩層 top/left 色號 7、bottom/right 色號 3。
	Buttons *struct {
		Y, W, H int
		Fill    uint8      `json:"fill"`
		Order   string     `json:"order"`
		Layers  [][4]uint8 `json:"layers"`
		Ink     uint8      `json:"ink"`
		Shadow  uint8      `json:"shadow"`
		TextDX  int        `json:"text_dx"`
		TextDY  int        `json:"text_dy"`
		Items   []struct {
			X     int      `json:"x"`
			Text  string   `json:"text"`
			Codes []uint16 `json:"codes"`
		} `json:"items"`
	} `json:"buttons"`

	// Tabs 是一覽表那四個分頁鈕。核取方塊是三層立體邊，內部填的是那一格的
	// 底色——選中的填 2、沒選中的填 9，這是原版分辨選了哪一頁的方式。
	Tabs *struct {
		Y, H   int
		Fill   uint8      `json:"fill"`
		Order  string     `json:"order"`
		Layers [][4]uint8 `json:"layers"`
		Ink    uint8      `json:"ink"`
		Shadow uint8      `json:"shadow"`
		TextDX int        `json:"text_dx"`
		TextDY int        `json:"text_dy"`
		Box    struct {
			DX, DY, W, H int
			Order        string     `json:"order"`
			Layers       [][4]uint8 `json:"layers"`
			FillOn       uint8      `json:"fill_on"`
			FillOff      uint8      `json:"fill_off"`
			Patches      []struct {
				DX, DY int
				Color  uint8 `json:"color"`
			} `json:"patches"`
		} `json:"box"`
		Selected int `json:"selected"`
		Items    []struct {
			X     int      `json:"x"`
			W     int      `json:"w"`
			Text  string   `json:"text"`
			Codes []uint16 `json:"codes"`
		} `json:"items"`
	} `json:"tabs"`
}

// frameOf 把規格裡的層色轉成 ui.Frame。
func frameOf(fill uint8, order string, layers [][4]uint8) ui.Frame {
	f := ui.Frame{Fill: fill, Order: order}
	for _, l := range layers {
		f.Layers = append(f.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
	}
	return f
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

	p.drawButtons(fb, fnt)
	p.drawTabs(fb, fnt)
	region.DrawAt(fb, fnt, c.Region, faction, p.Summary.DX, p.Summary.DY)
}

// DrawChrome 只畫按鈕與分頁鈕，不畫算出來的文字。
//
// 月底結算的當下右側面板的字整塊是空的（畫面上只有「戰略模式處理中」），
// 但按鈕與分頁鈕還在。
func (p *CapitalPanel) DrawChrome(fb *gfx.Framebuffer, fnt *font.Font) {
	p.drawButtons(fb, fnt)
	p.drawTabs(fb, fnt)
}

// drawButtons 畫中間那三顆命令按鈕。
//
// 字裁在按鈕的內容區裡：立體邊往上移那一列會落到按鈕的白色上緣上，
// 原版的上緣是乾淨的。
func (p *CapitalPanel) drawButtons(fb *gfx.Framebuffer, fnt *font.Font) {
	b := p.Buttons
	if b == nil {
		return
	}
	f := frameOf(b.Fill, b.Order, b.Layers)
	for _, it := range b.Items {
		f.Draw(fb, it.X, b.Y, b.W, b.H)
		cx, cy := it.X+b.TextDX, b.Y+b.TextDY
		cw, ch := b.W-2*b.TextDX, b.H-2*b.TextDY
		for i, code := range it.Codes {
			ui.DrawGlyphClipped(fb, fnt, cx+i*font.GlyphW, cy, code,
				b.Ink, b.Shadow, -1, false, cx, cy, cw, ch)
		}
	}
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

// drawTabs 畫一覽表那四個分頁鈕。
func (p *CapitalPanel) drawTabs(fb *gfx.Framebuffer, fnt *font.Font) {
	t := p.Tabs
	if t == nil {
		return
	}
	outer := frameOf(t.Fill, t.Order, t.Layers)
	for i, it := range t.Items {
		outer.Draw(fb, it.X, t.Y, it.W, t.H)
		fill := t.Box.FillOff
		if i == t.Selected {
			fill = t.Box.FillOn
		}
		bx, by := it.X+t.Box.DX, t.Y+t.Box.DY
		frameOf(fill, t.Box.Order, t.Box.Layers).Draw(fb, bx, by, t.Box.W, t.Box.H)
		for _, pt := range t.Box.Patches {
			fb.Set(bx+pt.DX, by+pt.DY, pt.Color)
		}
		cx, cy := it.X+t.TextDX, t.Y+t.TextDY
		cw, ch := it.W-t.TextDX-2, t.H-2*t.TextDY
		for j, code := range it.Codes {
			ui.DrawGlyphClipped(fb, fnt, cx+j*font.GlyphW, cy, code,
				t.Ink, t.Shadow, -1, false, cx, cy, cw, ch)
		}
	}
}
