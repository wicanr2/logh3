package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰場的行星情報視窗（原版 `1373:1CBC`，`docs/re/144`）。
//
// 它是自己的一個視窗：`3603:0006(色 0x97, 格 (3,4), 內容 22 × 10 格,
// 旗標 1)`，實際佔 24 × 12 格。與補給面板不同，這裡的分行是**格式字串
// 自己帶的 `\n`**，不是靠「剛好佔滿一列」。

// PlanetInfoSpec 是情報視窗的版面與格式字串。
type PlanetInfoSpec struct {
	CellX  int `json:"cell_x"`
	CellY  int `json:"cell_y"`
	Cols   int `json:"cols"`
	Rows   int `json:"rows"`
	Colour int `json:"colour"`

	// OwnerInk 是 `ds:6092`（字色），OwnerColour 是 `ds:6096`（整個位元組）。
	// `OwnerInk[n] | 0x90` 正好等於 `OwnerColour[n]`。
	OwnerInk    []int `json:"owner_ink"`
	OwnerColour []int `json:"owner_colour"`

	OwnerNames []hexStr `json:"owner_names"`

	Lines struct {
		Region        hexStr `json:"region"`
		RegionName    hexStr `json:"region_name"`
		Fortress      hexStr `json:"fortress"`
		Planet        hexStr `json:"planet"`
		KindAndOwner  hexStr `json:"kind_and_owner"`
		PlanetName    hexStr `json:"planet_name"`
		Blank         hexStr `json:"blank"`
		Economy       hexStr `json:"economy"`
		Defence       hexStr `json:"defence"`
		KindHeader    hexStr `json:"kind_header"`
		KindValue     hexStr `json:"kind_value"`
		StockHeader   hexStr `json:"stock_header"`
		StockValue    hexStr `json:"stock_value"`
		EconomyOther  hexStr `json:"economy_other"`
		DefenceOther  hexStr `json:"defence_other"`
	} `json:"lines"`
}

// LoadPlanetInfoSpec 讀情報視窗的版面。
func LoadPlanetInfoSpec(path string) (*PlanetInfoSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s PlanetInfoSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// PlanetInfo 是情報視窗要顯示的東西。
type PlanetInfo struct {
	RegionOwner int    // 星域 `+0x12`
	RegionName  []byte // 星域記錄開頭那個名字
	PlanetOwner int    // 行星 `+0x11`
	PlanetName  []byte
	Fortress    bool // 行星 `+0x12` 非 0

	// Own ＝ 這顆行星是玩家陣營的（`行星 +0x11 == ds:8F1D`）。
	// 不是的話經濟力與防衛力那兩行換成 `----` 的版本，後面幾段不印。
	Own bool

	Economy            int64 // `+0x1A` 32 位元
	GrowthInt, GrowthF int   // 成長率的整數與小數（`2703:0D47` ÷ 10）
	Defence, DefenceMax int  // `+0x14`／`+0x16`
	Recover            int   // `+0x18`

	Kinds    [7]int // 七個艦種的補充存量
	Missiles int64
	Fighters int64
}

// DrawPlanetInfo 把情報視窗畫出來（含框）。
func DrawPlanetInfo(fb *gfx.Framebuffer, s *PlanetInfoSpec, fnt *font.Font,
	tiles *ui.FrameTiles, p PlanetInfo) {
	if fb == nil || s == nil || tiles == nil {
		return
	}
	(ui.TileFrame{Tiles: tiles, Style: ui.PanelFrameStyle, Fill: uint8(s.Colour >> 4)}).
		Draw(fb, s.CellX, s.CellY, s.Cols+2, s.Rows+2)

	l := &s.Lines
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	b.Colour = uint8(s.ownerColour(p.RegionOwner))
	b.Printf(l.Region.Raw(), s.ownerName(p.RegionOwner))
	b.Colour = uint8(s.Colour)
	b.Printf(l.RegionName.Raw(), p.RegionName)

	b.Colour = uint8(s.ownerInk(p.PlanetOwner) | 0x90)
	kind := l.Planet.Raw()
	if p.Fortress {
		kind = l.Fortress.Raw()
	}
	// `"%s%s  "` 的兩個參數是**擁有者名在前、要塞／行星在後**——原版的
	// 推入順序是由後往前，先推「詞」再推「名字」（`docs/re/144`）。
	b.Printf(l.KindAndOwner.Raw(), s.ownerName(p.PlanetOwner), kind)
	b.Colour = uint8(s.Colour)
	b.Printf(l.PlanetName.Raw(), p.PlanetName)
	b.Write(l.Blank.Raw())

	if !p.Own {
		b.Printf(l.EconomyOther.Raw(), p.Economy)
		b.Printf(l.DefenceOther.Raw(), p.Defence, p.DefenceMax)
		ui.DrawTextBuffer(fb, fnt, tiles, b, s.CellX+1, s.CellY+1)
		return
	}
	b.Printf(l.Economy.Raw(), p.Economy, p.GrowthInt, p.GrowthF)
	b.Printf(l.Defence.Raw(), p.Defence, p.DefenceMax, p.Recover)
	b.Write(l.KindHeader.Raw())
	for _, v := range p.Kinds {
		b.Printf(l.KindValue.Raw(), v)
	}
	b.Write(l.StockHeader.Raw())
	b.Printf(l.StockValue.Raw(), p.Missiles, p.Fighters)
	ui.DrawTextBuffer(fb, fnt, tiles, b, s.CellX+1, s.CellY+1)
}

func (s *PlanetInfoSpec) ownerInk(n int) int {
	if n < 0 || n >= len(s.OwnerInk) {
		return 7
	}
	return s.OwnerInk[n]
}

func (s *PlanetInfoSpec) ownerColour(n int) int {
	if n < 0 || n >= len(s.OwnerColour) {
		return s.Colour
	}
	return s.OwnerColour[n]
}

func (s *PlanetInfoSpec) ownerName(n int) []byte {
	if n < 0 || n >= len(s.OwnerNames) {
		return nil
	}
	return s.OwnerNames[n].Raw()
}
