package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 「指定待命行星」（原版 `0DEA:1320`，`docs/re/314`、規格 `docs/spec/051` 待命那一節）。
//
// 作戰選單第 4 項「待命」開出來的視窗：18×11 格疊在艦隊視窗上，左邊是
// 這個星域的戰場地圖縮成 128×128 的單色小圖，右邊一顆行星一顆鈕，
// 每顆鈕拉一條引線到小圖上那顆行星；「決定」把選的那一顆寫進艦隊記錄
// `+0x07`（`+0x08` 留 5）。這一層只會排版與畫圖；值由 `internal/game` 湊齊。

// FleetStandbySpec 是版面、格式字串、熱區與小圖的常數
// （`docs/internals/data/fleet-standby-panel.json`，`tools/scripts/standby_spec.py`）。
type FleetStandbySpec struct {
	CellX    int               `json:"cell_x"`
	CellY    int               `json:"cell_y"`
	Cols     int               `json:"cols"`
	Rows     int               `json:"rows"`
	Style    int               `json:"style"`
	Colour   int               `json:"colour"`
	Formats  map[string]string `json:"formats"`
	Hotspots []struct {
		CellX int `json:"cell_x"`
		CellY int `json:"cell_y"`
		Cols  int `json:"cols"`
		Rows  int `json:"rows"`
		Code  int `json:"code"`
	} `json:"hotspots"`
	ButtonColour  int               `json:"button_colour"`
	PressedColour int               `json:"pressed_colour"`
	Decide        string            `json:"decide"`
	Cancel        string            `json:"cancel"`
	Minimap       RegionMinimapSpec `json:"minimap"`
	Leader        LeaderSpec        `json:"leader"`
}

// RegionMinimapSpec 是 128×128 的星域小圖：`1727:05C6` 縮圖、`2376:0B44` 貼在 (X, Y)、
// `3870:0723` 的外框。指定待命行星那張與星系情報視窗共用（`docs/re/314` §四、
// `docs/re/320`）。
type RegionMinimapSpec struct {
	X, Y, W, H int
	Frame      struct {
		X0  int `json:"x0"`
		Y0  int `json:"y0"`
		X1  int `json:"x1"`
		Y1  int `json:"y1"`
		Ink int `json:"ink"`
	} `json:"frame"`
	Ink            int   `json:"ink"`
	Bg             int   `json:"bg"`
	GridRowEvery   int   `json:"grid_row_every"`
	GridColEvery   int   `json:"grid_col_every"`
	EmptyTiles     []int `json:"empty_tiles"`
	RowRepeatEvery int   `json:"row_repeat_every"`
}

// LeaderSpec 是每顆行星從文字拉到小圖上位置的引線（`3870:07C9`，先畫往左上
// 偏一個像素的影子）。
type LeaderSpec struct {
	X0        int `json:"x0"`
	Y0Base    int `json:"y0_base"`
	Y0Pitch   int `json:"y0_pitch"`
	ShadowInk int `json:"shadow_ink"`
}

// MinimapPlanet 是小圖上一顆行星：位置（載入戰場地圖時算的，`rules.PlanetSites`）
// 與擁有者（引線的顏色）。
type MinimapPlanet struct {
	Owner int
	X, Y  int
}

// 熱區表 `ds:2248` 的動作碼：0–4 是第幾顆行星，5 決定、6 中止。
const (
	StandbyHotspotDecide = 5
	StandbyHotspotCancel = 6
	StandbyHotspotNone   = -1
)

// LoadFleetStandbySpec 讀 `docs/internals/data/fleet-standby-panel.json`。
func LoadFleetStandbySpec(path string) (*FleetStandbySpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FleetStandbySpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Cols == 0 || s.Rows == 0 || len(s.Formats) == 0 || len(s.Hotspots) < 7 {
		return nil, fmt.Errorf("fleet-standby-panel：規格不完整（%d×%d、%d 條格式字串、%d 筆熱區）",
			s.Cols, s.Rows, len(s.Formats), len(s.Hotspots))
	}
	return &s, nil
}

func (s *FleetStandbySpec) fmtBytes(name string) []byte {
	h, ok := s.Formats[name]
	if !ok {
		panic("fleet-standby-panel：規格沒有格式字串 " + name)
	}
	return decodeHex(h)
}

// FleetStandbyPlanet 是清單上的一顆行星：名字、擁有者，與它在戰場地圖上
// 的位置（`rules.PlanetSite` 的 X／Y，像素）。
type FleetStandbyPlanet struct {
	Name  []byte
	Owner int
	X, Y  int
}

// FleetStandbyView 是這張視窗要印的每一個值。
type FleetStandbyView struct {
	FactionName []byte // 帶顏色跳脫碼的陣營名（`ds:60F4[陣營]`，與艦隊視窗同一張表）
	Number      int    // 艦隊編號
	AdmiralName []byte // 提督姓名（Big5）
	RegionName  []byte // 艦隊所在星域的名字（Big5，印成視窗底色不換色）
	Planets     []FleetStandbyPlanet
	Selected    int // 反白的那一顆（−1 ＝ 沒有）
	// Repainted 是視窗顯示之後**換過顏色**的行星鈕：原版點一顆行星時把
	// 舊反白改回 0x87、新的改成 0x27，`341B:02E5` 只重畫變過的格子——
	// 所以那幾顆鈕蓋在引線上面，沒動過的鈕引線還壓在字上
	// （收據 `fcmd-standby-pick` 的 (288,250)）。
	Repainted []bool

	// Cells 是這個星域的戰場地圖（`MAP<nn>.CMP` 解開的圖塊，MapCols × MapRows），
	// 小圖從它縮出來；nil 就整塊底色。
	Cells            []uint8
	MapCols, MapRows int
}

// Draw 把視窗畫在畫面上（疊在艦隊視窗上面）。回沒有字模的字碼。
//
// 順序照原版：`3603:0006` 開窗、`0DEA:14D2` 排字與鈕、`341B:0145` 顯示、
// `2376:0B44` 貼小圖、`0DEA:1594` 外框與引線。
func (s *FleetStandbySpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	v *FleetStandbyView) []uint16 {
	if s == nil || fb == nil || v == nil {
		return nil
	}
	if tiles != nil {
		(ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}).
			Draw(fb, s.CellX, s.CellY, s.Cols+2, s.Rows+2)
	}
	b := s.Buffer(v)
	missing := ui.DrawTextBuffer(fb, fnt, tiles, b, s.CellX+1, s.CellY+1)
	s.drawMinimap(fb, v)
	s.drawLeaders(fb, v)
	// 換過顏色的鈕只重畫本體那幾格（`37A9:0332` 改的就是本體格的顏色）。
	for _, h := range s.Hotspots {
		if h.Code >= 0 && h.Code < len(v.Repainted) && v.Repainted[h.Code] {
			ui.DrawTextBufferRect(fb, fnt, tiles, b, s.CellX+1, s.CellY+1, h.CellX, h.CellY, h.Cols, h.Rows)
		}
	}
	return missing
}

// Buffer 把整張視窗排進內容緩衝區——`0DEA:14D2` 那一串列印。
func (s *FleetStandbySpec) Buffer(v *FleetStandbyView) *ui.TextBuffer {
	f := s.fmtBytes
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	b.Write(f("title"))
	// `3982:00B6`：「%s第」帶色的陣營名、`3982:0038` 全形編號、「艦隊  %s」提督。
	b.Printf(f("fleet_prefix"), v.FactionName)
	b.Attr |= ui.AttrWiden
	b.Printf([]byte("%d"), v.Number)
	b.Attr &^= ui.AttrWiden
	b.Printf(f("fleet_suffix"), v.AdmiralName)
	b.Printf(f("region"), v.RegionName)
	// `37A9:000F` 畫熱區表上啟用的鈕：行星鈕的標籤是 `0DEA:1433` 排好的
	// `%-16s`（只有 i < 行星數的那幾筆啟用），決定／中止是固定的字。
	for _, h := range s.Hotspots {
		var label []byte
		colour := uint8(s.ButtonColour)
		switch {
		case h.Code == StandbyHotspotDecide:
			label = decodeHex(s.Decide)
		case h.Code == StandbyHotspotCancel:
			label = decodeHex(s.Cancel)
		case h.Code >= 0 && h.Code < len(v.Planets):
			label = ui.Sprintf(f("planet_label"), v.Planets[h.Code].Name)
			if h.Code == v.Selected {
				colour = uint8(s.PressedColour)
			}
		default:
			continue
		}
		putRawButton(b, h.CellX, h.CellY, h.Cols, colour, label)
	}
	return b
}

// drawMinimap 是 `1727:05C6` ＋ `2376:0B44`：把 128 欄 × 102 列的圖塊縮成
// 128×128 的單色圖貼在 (X, Y)，一格一個像素——欄不縮，列每 RowRepeatEvery
// 列重用前一列來源（102 列拉成 128 列）。
//
// 一格算「有東西」（畫 Ink）的規則是載入地圖時塞進去的格線（`1588:073C`／
// `078C`：列 %12 ＝ 0 或欄 %15 ＝ 0 的空格換成格線圖塊）再查 `ds:3C8C`：
// 圖塊 0／1／5 不算，其餘都算。
//
// 最後一列（127）的來源列是 102，超出地圖——原版讀到緩衝區後面的東西，
// 那 128 個像素是垃圾；remake 畫底色，規格列為已知差異。
func (s *FleetStandbySpec) drawMinimap(fb *gfx.Framebuffer, v *FleetStandbyView) {
	s.Minimap.Draw(fb, v.Cells, v.MapCols, v.MapRows)
}

// Draw 把地圖（MapCols × MapRows 的圖塊）縮成小圖貼上去。
func (m *RegionMinimapSpec) Draw(fb *gfx.Framebuffer, cells []uint8, mapCols, mapRows int) {
	ink, bg := uint8(m.Ink), uint8(m.Bg)
	empty := func(t uint8) bool {
		for _, e := range m.EmptyTiles {
			if int(t) == e {
				return true
			}
		}
		return false
	}
	for py := 0; py < m.H; py++ {
		src := py
		if m.RowRepeatEvery > 0 {
			src = py - py/m.RowRepeatEvery
		}
		for px := 0; px < m.W; px++ {
			c := bg
			if cells != nil && src < mapRows && px < mapCols {
				t := cells[src*mapCols+px]
				grid := t == 0 && ((m.GridRowEvery > 0 && src%m.GridRowEvery == 0) ||
					(m.GridColEvery > 0 && px%m.GridColEvery == 0))
				if grid || !empty(t) {
					c = ink
				}
			}
			fb.Set(m.X+px, m.Y+py, c)
		}
	}
}

// drawLeaders 是 `0DEA:1594`：小圖外框（`3870:0723`，四邊含端點），再為每顆
// 行星從鈕的左緣拉一條線到小圖上的位置——先畫往左上偏一個像素的影子
// （色 ShadowInk），再畫擁有者陣營色的本體。線是 `3870:07C9` 的直接上色版。
func (s *FleetStandbySpec) drawLeaders(fb *gfx.Framebuffer, v *FleetStandbyView) {
	planets := make([]MinimapPlanet, len(v.Planets))
	for i, p := range v.Planets {
		planets[i] = MinimapPlanet{Owner: p.Owner, X: p.X, Y: p.Y}
	}
	s.Leader.Draw(fb, &s.Minimap, planets)
}

// Draw 畫小圖外框與每顆行星的引線。
func (l *LeaderSpec) Draw(fb *gfx.Framebuffer, m *RegionMinimapSpec, planets []MinimapPlanet) {
	clip := ClipRect{X0: 0, Y0: 0, X1: gfx.Width - 1, Y1: gfx.Height - 1}
	fr := m.Frame
	ink := uint8(fr.Ink)
	clip.Line(fb, fr.X0, fr.Y0, fr.X1, fr.Y0, ink)
	clip.Line(fb, fr.X0, fr.Y1, fr.X1, fr.Y1, ink)
	clip.Line(fb, fr.X0, fr.Y0, fr.X0, fr.Y1, ink)
	clip.Line(fb, fr.X1, fr.Y0, fr.X1, fr.Y1, ink)
	for i, p := range planets {
		x0, y0 := l.X0, l.Y0Base+i*l.Y0Pitch
		x1, y1 := m.X+p.X>>4, m.Y+p.Y>>4
		clip.Line(fb, x0-1, y0-1, x1-1, y1-1, uint8(l.ShadowInk))
		clip.Line(fb, x0, y0, x1, y1, rules.FactionColour(p.Owner))
	}
}

// HotspotAt 回畫面座標 (x,y) 落在哪一筆熱區上（`37A9:01F9` 查 `ds:2248`）：
// 0–4 行星（只有 i < nPlanets 的啟用）、`StandbyHotspotDecide`、
// `StandbyHotspotCancel`；都不是回 `StandbyHotspotNone`。
func (s *FleetStandbySpec) HotspotAt(x, y, nPlanets int) int {
	for _, h := range s.Hotspots {
		if h.Code >= 0 && h.Code < StandbyHotspotDecide && h.Code >= nPlanets {
			continue
		}
		bx := (s.CellX + 1 + h.CellX) * ui.FrameTileW
		by := (s.CellY+1+h.CellY)*ui.FrameTileH + ui.FrameCellOriginY
		if x >= bx && x < bx+h.Cols*ui.FrameTileW && y >= by && y < by+h.Rows*ui.FrameTileH {
			return h.Code
		}
	}
	return StandbyHotspotNone
}

// putRawButton 與 `putTextButton` 同一件事，標籤直接給位元組（兩個一格）。
func putRawButton(b *ui.TextBuffer, x, y, cols int, colour uint8, raw []byte) {
	if c := b.At(x-1, y); c != nil {
		c.Flags |= ui.CellButtonLeft
	}
	for i := 0; i < cols; i++ {
		c := b.At(x+i, y)
		if c == nil {
			continue
		}
		c.Colour = colour
		c.Flags |= ui.CellButtonBody
		if 2*i+1 < len(raw) {
			c.Code = uint16(raw[2*i])<<8 | uint16(raw[2*i+1])
		}
	}
	if c := b.At(x+cols, y); c != nil {
		c.Flags |= ui.CellButtonRight
	}
}
