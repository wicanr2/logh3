package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// FleetList 是戰略地圖介面右下角那張清單的「艦隊」分頁
// （`docs/internals/data/fleet-list-panel.json`，`docs/re/70`）。
//
// 一列是「第Ｎ艦隊」＋ 提督姓名 ＋ 一個數字。編號那四格印成色 6，
// 姓名與數字是色 7。
type FleetList struct {
	Y      int
	Pitch  int
	Rows   int
	Number struct {
		PrefixX int   `json:"prefix_x"`
		DigitsX int   `json:"digits_x"`
		SuffixX int   `json:"suffix_x"`
		Ink     uint8 `json:"ink"`
	} `json:"number"`
	Name struct {
		X   int
		Ink uint8
	} `json:"name"`
	Value struct {
		Right int
		Ink   uint8
		// 士氣值低於 AlertBelow 換成 AlertInk（`cmp [bx+3Ch], 32h`）。
		AlertBelow int   `json:"alert_below"`
		AlertInk   uint8 `json:"alert_ink"`
	} `json:"value"`
	// Marker 是「第」前面那一格的記號（`ds:1A66` 那張表，依 `+0x08` 查；
	// 停駐的再看星域旗標）。
	Marker struct {
		X      int
		ByMove []struct {
			Move int    `json:"move"`
			Code string `json:"code"`
			Ink  uint8  `json:"ink"`
		} `json:"by_move"`
		ParkedBattle struct {
			Code string `json:"code"`
			Ink  uint8  `json:"ink"`
		} `json:"parked_battle"`
		ParkedBit4 struct {
			Code string `json:"code"`
			Ink  uint8  `json:"ink"`
		} `json:"parked_bit4"`
	} `json:"marker"`
	// Row 是一列按鈕樣式的那 17 格：點開艦隊視窗時整列底色換成
	// `PressedBg`（`0AD0:0272` 的 `3603:00DB(…, 0x20, …)`，`docs/re/312`）。
	// 上下各 `Inset` 列與左邊兩個像素是按鈕的立體邊，不動（X 已扣掉）。
	Row struct {
		X, W, Top, Inset int
		PressedBg        uint8 `json:"pressed_bg"`
	} `json:"row"`
	// Tabs 是其餘三個分頁與 ▲▼ 的版面（`listtabs.go`）。
	Tabs *ListTabsLayout `json:"tabs"`
}

// LoadFleetList 讀版面。
func LoadFleetList(path string) (*FleetList, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l FleetList
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// FleetRow 是清單上的一列。
type FleetRow struct {
	Number int      // 第幾艦隊
	Name   []uint16 // 帶隊提督的姓名
	Value  int      // 右邊那個數字 ＝ 士氣值（`+0x3C`）
	// Move 是艦隊的 `+0x08`；停駐（5）時 RegionFlags 是所在星域的 `+0x13`，
	// 兩者決定「第」前面那一格印哪個記號。
	Move        int
	RegionFlags int
	// ValueText／ValueCodes 換掉右邊那個數字（一覽表艦隊資料的「部隊」「作戰」
	// 分頁）：半形字串或全形字碼，一樣靠右對齊 `Value.Right`。
	ValueText  string
	ValueCodes []uint16
}

// markerFor 回這一列要印的記號（Big5 字碼）與字色；不印回 0。
func (l *FleetList) markerFor(r FleetRow) (uint16, uint8) {
	for _, m := range l.Marker.ByMove {
		if m.Move == r.Move {
			return big5Code(m.Code), m.Ink
		}
	}
	switch {
	case r.RegionFlags&0x04 != 0:
		return big5Code(l.Marker.ParkedBattle.Code), l.Marker.ParkedBattle.Ink
	case r.RegionFlags&0x10 != 0:
		return big5Code(l.Marker.ParkedBit4.Code), l.Marker.ParkedBit4.Ink
	}
	return 0, 0
}

// big5Code 把規格檔裡的 "a1b8" 換成字碼 0xA1B8。
func big5Code(hex string) uint16 {
	var v uint16
	for i := 0; i < len(hex); i++ {
		c := hex[i]
		var n uint16
		switch {
		case c >= '0' && c <= '9':
			n = uint16(c - '0')
		case c >= 'a' && c <= 'f':
			n = uint16(c-'a') + 10
		case c >= 'A' && c <= 'F':
			n = uint16(c-'A') + 10
		}
		v = v<<4 | n
	}
	return v
}

// Prefix／Suffix 是「第」與「艦隊」的字碼——固定的字，不隨局面變。
var fleetRowPrefix = []uint16{0xB2C4}
var fleetRowSuffix = []uint16{0xC4A5, 0xB6A4}

// Draw 畫這幾列；`pressed` 是被點開視窗的那一列（−1 ＝ 沒有），整列底色
// 換成 `Row.PressedBg`，字照畫在上面（收據 `fcmd-win`）。
//
// **每一列的字裁在自己那一列裡**：立體邊那一趟會往上移一列，不裁的話
// 會咬掉上面那條分隔線（收據 `dec1`：那 16 個像素原版是線的色 7）。
// 與按鈕上的字同一件事（`ui.DrawGlyphClipped`）。
func (l *FleetList) Draw(fb *gfx.Framebuffer, fnt *font.Font, rows []FleetRow, pressed int) {
	for i, r := range rows {
		if i >= l.Rows {
			break
		}
		y := l.Y + i*l.Pitch
		if i == pressed && l.Row.W > 0 {
			top := l.Row.Top + i*l.Pitch
			fb.FillRect(l.Row.X, top+l.Row.Inset, l.Row.W, l.Pitch-2*l.Row.Inset, l.Row.PressedBg)
		}
		full := func(x int, codes []uint16, ink uint8) {
			for _, c := range codes {
				ui.DrawGlyphClipped(fb, fnt, x, y, c, ink, panelShadow, -1, false,
					0, y, gfx.Width, l.Pitch)
				x += font.GlyphW
			}
		}
		if code, ink := l.markerFor(r); code != 0 {
			full(l.Marker.X, []uint16{code}, ink)
		}
		full(l.Number.PrefixX, fleetRowPrefix, l.Number.Ink)
		drawFleetNumberRow(fb, fnt, l.Number.DigitsX, y, r.Number, l.Number.Ink, l.Pitch)
		full(l.Number.SuffixX, fleetRowSuffix, l.Number.Ink)
		full(l.Name.X, r.Name, l.Name.Ink)
		if len(r.ValueCodes) > 0 {
			full(l.Value.Right-len(r.ValueCodes)*font.GlyphW, r.ValueCodes, l.Value.Ink)
			continue
		}
		s := strconv.Itoa(r.Value)
		ink := l.Value.Ink
		if r.ValueText != "" {
			s = r.ValueText
		} else if l.Value.AlertBelow > 0 && r.Value < l.Value.AlertBelow {
			ink = l.Value.AlertInk
		}
		x := l.Value.Right - len(s)*font.HalfW
		for k := 0; k < len(s); k++ {
			if s[k] != ' ' {
				ui.DrawHalfGlyphClipped(fb, fnt, x, y, s[k], ink, panelShadow,
					-1, false, 0, y, gfx.Width, l.Pitch)
			}
			x += font.HalfW
		}
	}
}
