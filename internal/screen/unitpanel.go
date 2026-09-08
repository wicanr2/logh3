package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// UnitPanelSpec 是戰場右側分隊面板的版面（`docs/internals/data/unit-panel.json`）。
//
// 點自己陣營的分隊時，右側的狀態欄換成這一塊：八列、列距 20、與狀態欄
// 同一組字元格（`docs/re/118`）。
type UnitPanelSpec struct {
	Cell struct {
		X, Y, Col, Row, Half, Rows int
	} `json:"cell"`
	Ink             uint8            `json:"ink"`
	MoraleLowInk    uint8            `json:"morale_low_ink"`
	MoraleLowBelow  int              `json:"morale_low_below"`
	FactionInk      []uint8          `json:"faction_ink"`
	Labels          map[string]label `json:"labels"`
	Factions        []label          `json:"factions"`
	FleetActions    []label          `json:"fleet_actions"`
	FleetActionInk  []uint8          `json:"fleet_action_ink"`
	SquadronActions []label          `json:"squadron_actions"`
	WeaponInk       struct {
		Full, Part, Low uint8
	} `json:"weapon_ink"`
	Halves map[string]int `json:"halves"`
}

// LoadUnitPanelSpec 讀分隊面板的版面。
func LoadUnitPanelSpec(path string) (*UnitPanelSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s UnitPanelSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// UnitPanelView 是面板上要印的東西。
type UnitPanelView struct {
	Faction  int      // 0 帝國、1 同盟
	FleetNo  int      // 艦隊記錄 `+0x02`：畫面上的艦隊編號
	Flagship bool     // 分隊序號 0
	Admiral  []uint16 // 旗艦印提督名
	UnitNo   int      // 其餘印「第Ｎ部隊」
	Scout    bool     // 艦種 7 印「索敵部隊」
	KindName []uint16
	Ships    int
	Morale   int
	// FleetMorale ＝ 標籤用「艦隊士氣值」（旗艦），否則「部隊士氣值」。
	FleetMorale bool
	// CannotFight ＝ 艦隊 `+0x00` bit0：整列換成「無法戰鬥」。
	CannotFight bool
	ActionName  []uint16
	ActionInk   uint8
	// UnitAction ＝ 分隊有自己的命令，標籤用「部隊行動」。
	UnitAction             bool
	Beam, Missile, Fighter int
	MissileInk, FighterInk uint8
}

// FactionColour 回陣營色（`ds:6096`，低四位才是色號）。
func (s *UnitPanelSpec) FactionColour(faction int) uint8 {
	if faction < 0 || faction >= len(s.FactionInk) {
		return s.Ink
	}
	return s.FactionInk[faction] & 0x0F
}

// FleetActionInkOf 回艦隊行動的顏色（`ds:6930`）。
func (s *UnitPanelSpec) FleetActionInkOf(action int) uint8 {
	if action < 0 || action >= len(s.FleetActionInk) {
		return s.Ink
	}
	return s.FleetActionInk[action] & 0x0F
}

// DrawUnitPanel 把分隊面板畫上去。
//
// 每一段的位置是「第幾個半形格」，走原版的格式字串算出來的：全形一個字
// 兩格、半形一格，x ＝ 格原點 ＋ 半形格 × 8（`docs/re/118`）。
func DrawUnitPanel(s *UnitPanelSpec, fb *gfx.Framebuffer, fnt *font.Font, v UnitPanelView) {
	x := func(half int) int { return s.Cell.X + half*s.Cell.Half }
	y := func(row int) int { return s.Cell.Y + row*s.Cell.Row }
	h := func(k string) int { return s.Halves[k] }
	lab := func(k string) []uint16 { return s.Labels[k].Codes }

	// 第 0 列：陣營名（陣營色）＋「第」＋ 編號 ＋「艦隊」。
	if v.Faction >= 0 && v.Faction < len(s.Factions) {
		drawFull(fb, fnt, x(h("faction")), y(0),
			s.Factions[v.Faction].Codes, s.FactionColour(v.Faction))
	}
	drawFull(fb, fnt, x(h("di")), y(0), lab("di"), s.Ink)
	drawRightHalves(fb, fnt, x(h("fleet_no")), y(0), itoa(v.FleetNo), 2, s.Ink)
	drawFull(fb, fnt, x(h("fleet_word")), y(0), lab("fleet"), s.Ink)

	// 第 1 列：旗艦印提督名，索敵分遣隊印「索敵部隊」，其餘印「第Ｎ部隊」。
	switch {
	case v.Flagship:
		drawFull(fb, fnt, x(h("admiral")), y(1), v.Admiral, s.Ink)
	case v.Scout:
		drawFull(fb, fnt, x(h("admiral")), y(1), lab("scout_unit"), s.Ink)
	default:
		drawFull(fb, fnt, x(h("unit_di")), y(1), lab("di"), s.Ink)
		// 樣板上「第？部隊」的 ？ 只有一格：個位數用全形數字，
		// 兩位數就塞兩個半形（寬度一樣，`docs/internals/34`）。
		if v.UnitNo >= 1 && v.UnitNo <= 9 {
			drawFullNumber(fb, fnt, x(h("unit_no")), y(1), v.UnitNo, fullDigitCodes, s.Ink)
		} else {
			drawRightHalves(fb, fnt, x(h("unit_no")), y(1), itoa(v.UnitNo), 2, s.Ink)
		}
		drawFull(fb, fnt, x(h("unit_word")), y(1), lab("unit"), s.Ink)
	}

	// 第 2 列：艦種名 ＋ 艘數。
	drawFull(fb, fnt, x(h("kind_name")), y(2), v.KindName, s.Ink)
	drawRightHalves(fb, fnt, x(h("ships")), y(2), itoa(v.Ships), h("ships_width"), s.Ink)

	// 第 3 列：士氣值。標籤用陣營色，值低於門檻換色。
	label := "unit_morale"
	if v.FleetMorale {
		label = "fleet_morale"
	}
	drawFull(fb, fnt, x(h("morale_label")), y(3), lab(label), s.FactionColour(v.Faction))
	ink := s.Ink
	if v.Morale < s.MoraleLowBelow {
		ink = s.MoraleLowInk
	}
	drawRightHalves(fb, fnt, x(h("morale_value")), y(3), itoa(v.Morale),
		h("morale_width"), ink)

	// 第 4 列：行動。艦隊 `+0x00` bit0 有設的話整列換成「無法戰鬥」。
	if v.CannotFight {
		drawFull(fb, fnt, x(h("action_label")), y(4), lab("cannot_fight"), s.MoraleLowInk)
	} else {
		al := "fleet_action"
		if v.UnitAction {
			al = "unit_action"
		}
		drawFull(fb, fnt, x(h("action_label")), y(4), lab(al), s.Ink)
		drawFull(fb, fnt, x(h("action_value")), y(4), v.ActionName, v.ActionInk)
	}

	// 第 5–7 列：三組武器。光線砲永遠是本色，另外兩組的顏色是算出來的。
	for i, w := range []struct {
		key   string
		value int
		ink   uint8
	}{
		{"beam", v.Beam, s.Ink},
		{"missile", v.Missile, v.MissileInk},
		{"fighter", v.Fighter, v.FighterInk},
	} {
		row := 5 + i
		drawFull(fb, fnt, x(h("weapon_label")), y(row), lab(w.key), s.Ink)
		drawRightHalves(fb, fnt, x(h("weapon_value")), y(row), itoa(w.value),
			h("weapon_width"), w.ink)
	}
}

// drawRightHalves 把一串半形字靠右對齊到寬 halves 個半形格的欄位裡，
// 左邊補空白——就是原版 `%4d`／`%6lu` 那種欄位。
func drawRightHalves(fb *gfx.Framebuffer, fnt *font.Font, x, y int, s string,
	halves int, ink uint8) {
	if len(s) > halves {
		s = s[len(s)-halves:]
	}
	x += (halves - len(s)) * font.HalfW
	for i := 0; i < len(s); i++ {
		drawHalf(fb, fnt, x+i*font.HalfW, y, s[i], ink)
	}
}

// fullDigitCodes 是全形數字０–９的倚天 Big5 碼（`docs/internals/04`）。
var fullDigitCodes = []uint16{
	0xA2AF, 0xA2B0, 0xA2B1, 0xA2B2, 0xA2B3,
	0xA2B4, 0xA2B5, 0xA2B6, 0xA2B7, 0xA2B8,
}
