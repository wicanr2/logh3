package screen

import (
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 右下角清單其餘三個分頁（行星／敵艦隊／提督）與捲動鈕的版面
// （`fleet-list-panel.json` 的 "tabs" 一節；列印常式 `0B99`／`0B20`／`0BCE`
// 段，`docs/re/319`）。四個分頁共用同一塊 18 × 5 格的內容區與同一對
// ▲▼，只有一列印什麼不同。

// ListTabsLayout 是那三個分頁一列的欄位位置。
type ListTabsLayout struct {
	// Scroll 是欄 17 的 ▲（列 7）與 ▼（列 11）：字碼 `\xFF\x11`／`\xFF\x15`
	// 是 `FRAME.BLK` 的第 16／20 張圖塊，箭頭裡色號 0 的像素填印那一刻的
	// 視窗底色 9；筆數不超過一頁時印 `\xFF\x21`（第 32 張，整張 8 的空白）。
	Scroll struct {
		X        int
		UpY      int   `json:"up_y"`
		DownY    int   `json:"down_y"`
		UpTile   int   `json:"up_tile"`
		DownTile int   `json:"down_tile"`
		Fill     uint8 `json:"fill"`
	} `json:"scroll"`
	// Planet：`\x03[色,%6s%4s` 擁有者名（陣營色）＋ 行星／要塞 → `%-16s` 名字
	// → ` %5u` 防衛力（滿的 7、掉了不到一半 5、掉了一半以上 2）。
	Planet struct {
		OwnerX           int `json:"owner_x"`
		KindX            int `json:"kind_x"`
		NameX            int `json:"name_x"`
		ValueRight       int `json:"value_right"`
		Ink              uint8
		FullInk          uint8      `json:"full_ink"`
		HalfInk          uint8      `json:"half_ink"`
		LowInk           uint8      `json:"low_ink"`
		Owners           [][]uint16 `json:"owners"`
		Planet, Fortress []uint16
	} `json:"planet"`
	// Admiral：`%-18s` 姓名 → `%-8s` 軍階。
	Admiral struct {
		NameX int `json:"name_x"`
		RankX int `json:"rank_x"`
		Ink   uint8
	} `json:"admiral"`
	// Enemy：看得清楚（索敵等級 3）時與「艦隊」分頁同一列版面、記號那格空白；
	// 看不清楚時印「第？艦隊」＋ 九個全形問號。
	Enemy struct {
		UnknownX int      `json:"unknown_x"`
		Unknown  []uint16 `json:"unknown"`
		Prefix   []uint16 `json:"prefix"`
	} `json:"enemy"`
}

// PlanetRow 是「行星」分頁的一列。
type PlanetRow struct {
	Owner    int
	Fortress bool
	Name     []uint16
	Defence  int
	Max      int
}

// AdmiralRow 是「提督」分頁的一列。
type AdmiralRow struct {
	Name []uint16
	Rank []uint16
}

// EnemyRow 是「敵艦隊」分頁的一列；Known ＝ 索敵等級到 3。
type EnemyRow struct {
	Known   bool
	Faction int
	Number  int
	Name    []uint16
	Morale  int
}

// DrawScroll 印 ▲▼（或空白圖塊）。tiles 是 `FRAME.BLK`。
func (l *FleetList) DrawScroll(fb *gfx.Framebuffer, tiles *ui.FrameTiles, arrows bool) {
	t := l.Tabs
	if t == nil || tiles == nil {
		return
	}
	up, down := t.Scroll.UpTile, t.Scroll.DownTile
	if !arrows {
		// 第 32 張整張是 8（沒有色號 0），貼上去就是一格空白。
		up, down = len(tiles.Tiles)-1, len(tiles.Tiles)-1
	}
	tiles.DrawTile(fb, t.Scroll.X, t.Scroll.UpY, up, t.Scroll.Fill)
	tiles.DrawTile(fb, t.Scroll.X, t.Scroll.DownY, down, t.Scroll.Fill)
}

// ScrollAt 回畫面座標落在 ▲（−1）或 ▼（＋1）上；沒有回 0。
func (l *FleetList) ScrollAt(x, y int) int {
	t := l.Tabs
	if t == nil || x < t.Scroll.X || x >= t.Scroll.X+ui.FrameTileW {
		return 0
	}
	switch {
	case y >= t.Scroll.UpY && y < t.Scroll.UpY+ui.FrameTileH:
		return -1
	case y >= t.Scroll.DownY && y < t.Scroll.DownY+ui.FrameTileH:
		return 1
	}
	return 0
}

func (l *FleetList) rowText(fb *gfx.Framebuffer, fnt *font.Font, x, y int, codes []uint16, ink uint8) {
	for _, c := range codes {
		ui.DrawGlyphClipped(fb, fnt, x, y, c, ink, panelShadow, -1, false,
			0, y, gfx.Width, l.Pitch)
		x += font.GlyphW
	}
}

func (l *FleetList) rowNumber(fb *gfx.Framebuffer, fnt *font.Font, right, y, v int, ink uint8) {
	s := strconv.Itoa(v)
	x := right - len(s)*font.HalfW
	for k := 0; k < len(s); k++ {
		ui.DrawHalfGlyphClipped(fb, fnt, x, y, s[k], ink, panelShadow,
			-1, false, 0, y, gfx.Width, l.Pitch)
		x += font.HalfW
	}
}

// DrawPlanets 畫「行星」分頁的幾列（`0B99:00A0`）。
func (l *FleetList) DrawPlanets(fb *gfx.Framebuffer, fnt *font.Font, rows []PlanetRow, pressed int) {
	t := l.Tabs
	if t == nil {
		return
	}
	for i, r := range rows {
		if i >= l.Rows {
			break
		}
		y := l.Y + i*l.Pitch
		l.pressRow(fb, i, pressed)
		owner := ownerColour(r.Owner)
		if r.Owner >= 0 && r.Owner < len(t.Planet.Owners) {
			l.rowText(fb, fnt, t.Planet.OwnerX, y, t.Planet.Owners[r.Owner], owner)
		}
		kind := t.Planet.Planet
		if r.Fortress {
			kind = t.Planet.Fortress
		}
		l.rowText(fb, fnt, t.Planet.KindX, y, kind, owner)
		l.rowText(fb, fnt, t.Planet.NameX, y, r.Name, t.Planet.Ink)
		// 防衛力的顏色：與上限相同 7；掉了的比剩下的多（剩不到一半）2；否則 5。
		ink := t.Planet.FullInk
		switch {
		case r.Defence == r.Max:
		case r.Max-r.Defence > r.Defence:
			ink = t.Planet.LowInk
		default:
			ink = t.Planet.HalfInk
		}
		l.rowNumber(fb, fnt, t.Planet.ValueRight, y, r.Defence, ink)
	}
}

// DrawAdmirals 畫「提督」分頁的幾列（`0BCE:0132`）。
func (l *FleetList) DrawAdmirals(fb *gfx.Framebuffer, fnt *font.Font, rows []AdmiralRow, pressed int) {
	t := l.Tabs
	if t == nil {
		return
	}
	for i, r := range rows {
		if i >= l.Rows {
			break
		}
		y := l.Y + i*l.Pitch
		l.pressRow(fb, i, pressed)
		l.rowText(fb, fnt, t.Admiral.NameX, y, r.Name, t.Admiral.Ink)
		l.rowText(fb, fnt, t.Admiral.RankX, y, r.Rank, t.Admiral.Ink)
	}
}

// DrawEnemies 畫「敵艦隊」分頁的幾列（`0B20:059B`）。
func (l *FleetList) DrawEnemies(fb *gfx.Framebuffer, fnt *font.Font, rows []EnemyRow, pressed int) {
	t := l.Tabs
	if t == nil {
		return
	}
	for i, r := range rows {
		if i >= l.Rows {
			break
		}
		y := l.Y + i*l.Pitch
		l.pressRow(fb, i, pressed)
		faction := ownerColour(r.Faction)
		if !r.Known {
			l.rowText(fb, fnt, l.Number.PrefixX, y, t.Enemy.Prefix, faction)
			l.rowText(fb, fnt, t.Enemy.UnknownX, y, t.Enemy.Unknown, l.Name.Ink)
			continue
		}
		l.rowText(fb, fnt, l.Number.PrefixX, y, fleetRowPrefix, faction)
		l.rowText(fb, fnt, l.Number.DigitsX, y, fullDigits(r.Number), faction)
		l.rowText(fb, fnt, l.Number.SuffixX, y, fleetRowSuffix, faction)
		l.rowText(fb, fnt, l.Name.X, y, r.Name, l.Name.Ink)
		ink := l.Value.Ink
		if l.Value.AlertBelow > 0 && r.Morale < l.Value.AlertBelow {
			ink = l.Value.AlertInk
		}
		l.rowNumber(fb, fnt, l.Value.Right, y, r.Morale, ink)
	}
}

// pressRow 把被點開視窗的那一列底色換掉（與「艦隊」分頁同一件事）。
func (l *FleetList) pressRow(fb *gfx.Framebuffer, i, pressed int) {
	if i == pressed && l.Row.W > 0 {
		top := l.Row.Top + i*l.Pitch
		fb.FillRect(l.Row.X, top+l.Row.Inset, l.Row.W, l.Pitch-2*l.Row.Inset, l.Row.PressedBg)
	}
}

// ownerColour 是 `ds:6092`：帝國 6、同盟 5、其餘 7。
func ownerColour(owner int) uint8 {
	switch owner {
	case 0:
		return 6
	case 1:
		return 5
	}
	return 7
}
