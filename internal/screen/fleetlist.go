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
	} `json:"value"`
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
	Value  int      // 右邊那個數字
}

// Prefix／Suffix 是「第」與「艦隊」的字碼——固定的字，不隨局面變。
var fleetRowPrefix = []uint16{0xB2C4}
var fleetRowSuffix = []uint16{0xC4A5, 0xB6A4}

// Draw 畫這幾列。
//
// **每一列的字裁在自己那一列裡**：立體邊那一趟會往上移一列，不裁的話
// 會咬掉上面那條分隔線（收據 `dec1`：那 16 個像素原版是線的色 7）。
// 與按鈕上的字同一件事（`ui.DrawGlyphClipped`）。
func (l *FleetList) Draw(fb *gfx.Framebuffer, fnt *font.Font, rows []FleetRow) {
	for i, r := range rows {
		if i >= l.Rows {
			break
		}
		y := l.Y + i*l.Pitch
		full := func(x int, codes []uint16, ink uint8) {
			for _, c := range codes {
				ui.DrawGlyphClipped(fb, fnt, x, y, c, ink, panelShadow, -1, false,
					0, y, gfx.Width, l.Pitch)
				x += font.GlyphW
			}
		}
		full(l.Number.PrefixX, fleetRowPrefix, l.Number.Ink)
		full(l.Number.DigitsX, fullDigits(r.Number), l.Number.Ink)
		full(l.Number.SuffixX, fleetRowSuffix, l.Number.Ink)
		full(l.Name.X, r.Name, l.Name.Ink)
		s := strconv.Itoa(r.Value)
		x := l.Value.Right - len(s)*font.HalfW
		for k := 0; k < len(s); k++ {
			ui.DrawHalfGlyphClipped(fb, fnt, x, y, s[k], l.Value.Ink, panelShadow,
				-1, false, 0, y, gfx.Width, l.Pitch)
			x += font.HalfW
		}
	}
}
