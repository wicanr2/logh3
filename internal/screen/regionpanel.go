package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// RegionPanel 是戰略地圖介面左下角那塊星域摘要的版面
// （規格 `docs/spec/001`，量到的座標在 `docs/internals/data/region-panel.json`）。
//
// 面板的文字是**算出來的**，不像其他畫面那樣量好畫死——換一個星域整塊就換。
type RegionPanel struct {
	Rect struct{ X, Y, W, H int } `json:"rect"`

	FactionLine struct {
		Y       int `json:"y"`
		PrefixX int `json:"prefix_x"`
		Prefix  []struct {
			Owner int      `json:"owner"`
			Text  string   `json:"text"`
			Codes []uint16 `json:"codes"`
		} `json:"prefix"`
		NameX   int   `json:"name_x"`
		NameInk uint8 `json:"name_ink"`
	} `json:"faction_line"`

	ValueLine struct {
		Y        int   `json:"y"`
		LabelInk uint8 `json:"label_ink"`
		Security struct {
			LabelX  int      `json:"label_x"`
			Label   []uint16 `json:"label"`
			Right   int      `json:"right"`
			Unknown string   `json:"unknown"`
		} `json:"security"`
		Planets struct {
			LabelX   int      `json:"label_x"`
			Label    []uint16 `json:"label"`
			OwnRight int      `json:"own_right"`
			SlashX   int      `json:"slash_x"`
			TotalX   int      `json:"total_x"`
		} `json:"planets"`
		Fleets struct {
			LabelX     int      `json:"label_x"`
			Label      []uint16 `json:"label"`
			OwnRight   int      `json:"own_right"`
			ColonX     int      `json:"colon_x"`
			Colon      uint16   `json:"colon"`
			EnemyRight int      `json:"enemy_right"`
		} `json:"fleets"`
	} `json:"value_line"`
}

// LoadRegionPanel 讀面板版面。
func LoadRegionPanel(path string) (*RegionPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p RegionPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// panelShadow 是面板文字的立體邊色號，與其他視窗一樣。
const panelShadow = 3

// DrawRegionPanel 把一個星域的摘要畫進面板。
//
// faction 是玩家陣營（決定「自軍」是哪一邊，以及數字用什麼顏色）。
func (p *RegionPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, s rules.RegionSummary, faction int) {
	own, enemy := rules.FactionColour(faction), rules.FactionColour(1-faction)

	// 第一行：陣營前綴 ＋ 星域名。
	for _, pre := range p.FactionLine.Prefix {
		if pre.Owner != s.Owner {
			continue
		}
		p.full(fb, fnt, p.FactionLine.PrefixX, p.FactionLine.Y, pre.Codes,
			rules.FactionColour(s.Owner))
	}
	p.full(fb, fnt, p.FactionLine.NameX, p.FactionLine.Y, s.Name, p.FactionLine.NameInk)

	y, ink := p.ValueLine.Y, p.ValueLine.LabelInk

	// 治安：自己這一方印百分比，其餘印 `----`。
	sec := p.ValueLine.Security
	p.full(fb, fnt, sec.LabelX, y, sec.Label, ink)
	if s.Own {
		p.right(fb, fnt, sec.Right, y, strconv.Itoa(s.Security())+"%", s.SecurityColour())
	} else {
		p.right(fb, fnt, sec.Right, y, sec.Unknown, ink)
	}

	// 行星 n/m：分子是玩家陣營的行星數，用玩家的陣營色。
	pl := p.ValueLine.Planets
	p.full(fb, fnt, pl.LabelX, y, pl.Label, ink)
	p.right(fb, fnt, pl.OwnRight, y, strconv.Itoa(s.Planets), own)
	p.half(fb, fnt, pl.SlashX, y, '/', ink)
	p.half(fb, fnt, pl.TotalX, y, byte('0'+s.Total%10), ink)

	// 艦隊 a：b，兩個數字各用自己那一邊的陣營色。
	fl := p.ValueLine.Fleets
	p.full(fb, fnt, fl.LabelX, y, fl.Label, ink)
	p.right(fb, fnt, fl.OwnRight, y, strconv.Itoa(s.OwnFleets), own)
	p.full(fb, fnt, fl.ColonX, y, []uint16{fl.Colon}, ink)
	p.right(fb, fnt, fl.EnemyRight, y, strconv.Itoa(s.EnemyFleets), enemy)
}

// full 從 x 起畫一串全形字。
func (p *RegionPanel) full(fb *gfx.Framebuffer, fnt *font.Font, x, y int, codes []uint16, ink uint8) {
	for _, c := range codes {
		ui.DrawGlyph(fb, fnt, x, y, c, ink, panelShadow, -1, false)
		x += font.GlyphW
	}
}

// half 畫一個半形字。
func (p *RegionPanel) half(fb *gfx.Framebuffer, fnt *font.Font, x, y int, c byte, ink uint8) {
	ui.DrawHalfGlyph(fb, fnt, x, y, c, ink, panelShadow, -1, false)
}

// right 把一串半形字**靠右對齊**到 right——面板上的數字全是這樣排的，
// 位數一多就往左長（`docs/spec/001` 用 -poke 造的三位數局面量到的）。
func (p *RegionPanel) right(fb *gfx.Framebuffer, fnt *font.Font, right, y int, s string, ink uint8) {
	x := right - len(s)*font.HalfW
	for i := 0; i < len(s); i++ {
		p.half(fb, fnt, x, y, s[i], ink)
		x += font.HalfW
	}
}
