package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰略地圖介面的星系情報視窗（`073F:060A`，左鍵點右側面板看的那一格；
// `docs/re/320`、規格 053）：整個左半一張 20 × 19 格的視窗，上面是星系名與
// 小圖＋行星名，下面是那一格待命船艦的七列（艦數、部隊、飛彈與戰鬥機），
// 右下角一顆「結束」。

// RegionInfoSpec 是版面與格式字串（`docs/internals/data/region-info-panel.json`）。
type RegionInfoSpec struct {
	CellX    int                `json:"cell_x"`
	CellY    int                `json:"cell_y"`
	Cols     int                `json:"cols"`
	Rows     int                `json:"rows"`
	Colour   int                `json:"colour"`
	Style    int                `json:"style"`
	Formats  map[string]hexStr  `json:"-"`
	RawFmts  map[string]any     `json:"formats"`
	Letters  []string           `json:"kind_letters"`
	Layout   map[string][]int   `json:"layout"`
	Hotspots []struct {
		CellX int `json:"cell_x"`
		CellY int `json:"cell_y"`
		Cols  int `json:"cols"`
		Rows  int `json:"rows"`
		Code  int      `json:"code"`
		Codes []string `json:"codes"`
	} `json:"hotspots"`
	ButtonColour int               `json:"button_colour"`
	Minimap      RegionMinimapSpec `json:"minimap"`
	Leader       LeaderSpec        `json:"leader"`

	otherRows []hexStr
}

// LoadRegionInfoSpec 讀版面。
func LoadRegionInfoSpec(path string) (*RegionInfoSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s RegionInfoSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	s.Formats = map[string]hexStr{}
	for k, v := range s.RawFmts {
		switch t := v.(type) {
		case map[string]any:
			s.Formats[k] = hexStr{Bytes: fmt.Sprint(t["bytes"]), Text: fmt.Sprint(t["text"])}
		case []any:
			for _, e := range t {
				if m, ok := e.(map[string]any); ok {
					s.otherRows = append(s.otherRows, hexStr{Bytes: fmt.Sprint(m["bytes"])})
				}
			}
		}
	}
	if s.Cols == 0 || s.Rows == 0 || len(s.Formats) < 9 || len(s.otherRows) != 7 || len(s.Hotspots) == 0 {
		return nil, fmt.Errorf("region-info-panel：規格不完整（%d×%d、%d 條格式字串、%d 條破折號列、%d 筆熱區）",
			s.Cols, s.Rows, len(s.Formats), len(s.otherRows), len(s.Hotspots))
	}
	return &s, nil
}

// RegionInfoView 是這張視窗要印的每一個值。
type RegionInfoView struct {
	Owner      int    // 星域 `+0x12`（陣營色與陣營名）
	OwnerName  []byte // `ds:60AE[擁有者]`（帝國軍／同盟軍／中立軍，沒有跳脫碼）
	RegionName []byte
	// Own ＝ 星域是玩家陣營的；不是的話七列印 `--------`。
	Own bool
	// Ships 是那一格擁有者陣營七種艦艇的待命艘數（星域 `+0x30`）、Units 是
	// 艘數 ÷ 滿編；Missiles／Fighters 是星域層的飛彈與艦載機（`+0x68`／`+0x70`）。
	Ships, Units       [7]int
	Missiles, Fighters int

	Planets []RegionInfoPlanet

	// Cells 是這個星域的戰場地圖（小圖從它縮出來）；nil 就整塊底色。
	Cells            []uint8
	MapCols, MapRows int
}

// RegionInfoPlanet 是一顆行星：名字（擁有者色）與小圖上的位置。
type RegionInfoPlanet struct {
	Name  []byte
	Owner int
	X, Y  int
}

func (s *RegionInfoSpec) fmt(name string) []byte { return s.Formats[name].Raw() }

func (s *RegionInfoSpec) one(name string) int {
	if v, ok := s.Layout[name]; ok && len(v) >= 1 {
		return v[0]
	}
	return 0
}

func (s *RegionInfoSpec) at(name string) (int, int) {
	if v, ok := s.Layout[name]; ok && len(v) == 2 {
		return v[0], v[1]
	}
	return 0, 0
}

// Buffer 把整張視窗排進內容緩衝區（`073F:06C5`）。
func (s *RegionInfoSpec) Buffer(v *RegionInfoView) *ui.TextBuffer {
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	goto_ := func(name string) {
		c, r := s.at(name)
		b.Printf(s.fmt("goto"), c, r)
	}
	// 標題：陣營名（陣營色）＋「星系」＋ 兩格空白 ＋ 星域名（視窗色）。
	b.Printf(s.fmt("title"), int(ownerColour(v.Owner))|0x90, v.OwnerName, s.Colour, v.RegionName)
	goto_("planets_label_at")
	b.Write(s.fmt("planets_label"))
	goto_("ships_header_at")
	b.Write(s.fmt("ships_header"))
	if v.Own {
		for k := 0; k < 7; k++ {
			letter := []byte(nil)
			if k < len(s.Letters) {
				letter = big5Bytes(s.Letters[k])
			}
			b.Printf(s.fmt("ships_row"), letter, int64(v.Ships[k]), int64(v.Units[k]))
		}
		goto_("missiles_label_at")
		b.Write(s.fmt("missiles_label"))
		goto_("missiles_value_at")
		b.Printf(s.fmt("missiles_value"), int64(v.Missiles))
		goto_("fighters_label_at")
		b.Write(s.fmt("fighters_label"))
		goto_("fighters_value_at")
		b.Printf(s.fmt("fighters_value"), int64(v.Fighters))
	} else {
		for _, row := range s.otherRows {
			b.Write(row.Raw())
		}
	}
	// 行星名：第 i 顆在 (9, 2＋i)，擁有者色。
	col, row0 := s.one("planet_name_col"), s.one("planet_name_row0")
	for i, p := range v.Planets {
		b.Printf(s.fmt("goto"), col, row0+i)
		b.Colour = uint8(int(ownerColour(p.Owner)) | 0x90)
		b.Write(p.Name)
	}
	b.Colour = uint8(s.Colour)
	// 「結束」：字那幾格換顏色帶按鈕旗標，左右各一格端點。
	for _, h := range s.Hotspots {
		if cell := b.At(h.CellX-1, h.CellY); cell != nil {
			cell.Flags |= ui.CellButtonLeft
		}
		for i := 0; i < h.Cols; i++ {
			if cell := b.At(h.CellX+i, h.CellY); cell != nil {
				cell.Colour = uint8(s.ButtonColour)
				cell.Flags |= ui.CellButtonBody
			}
		}
		if cell := b.At(h.CellX+h.Cols, h.CellY); cell != nil {
			cell.Flags |= ui.CellButtonRight
		}
		for i, hx := range h.Codes {
			var code uint16
			if _, err := fmt.Sscanf(hx, "%04X", &code); err != nil {
				continue
			}
			if cell := b.At(h.CellX+i, h.CellY); cell != nil {
				cell.Code = code
			}
		}
	}
	return b
}

// Draw 把視窗畫在畫面上。順序照原版：開窗、排字與鈕、顯示、貼小圖、外框與引線。
func (s *RegionInfoSpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	v *RegionInfoView) []uint16 {
	if s == nil || fb == nil || v == nil {
		return nil
	}
	if tiles != nil {
		(ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}).
			Draw(fb, s.CellX, s.CellY, s.Cols+2, s.Rows+2)
	}
	missing := ui.DrawTextBuffer(fb, fnt, tiles, s.Buffer(v), s.CellX+1, s.CellY+1)
	s.Minimap.Draw(fb, v.Cells, v.MapCols, v.MapRows)
	planets := make([]MinimapPlanet, len(v.Planets))
	for i, p := range v.Planets {
		planets[i] = MinimapPlanet{Owner: p.Owner, X: p.X, Y: p.Y}
	}
	s.Leader.Draw(fb, &s.Minimap, planets)
	return missing
}

// ButtonAt 回畫面座標 (x, y) 有沒有落在「結束」上（`37A9:01F9` 查 `ds:0DA2`）。
func (s *RegionInfoSpec) ButtonAt(x, y int) bool {
	for _, h := range s.Hotspots {
		bx := (s.CellX + 1 + h.CellX) * ui.FrameTileW
		by := (s.CellY+1+h.CellY)*ui.FrameTileH + ui.FrameCellOriginY
		if x >= bx && x < bx+h.Cols*ui.FrameTileW && y >= by && y < by+h.Rows*ui.FrameTileH {
			return true
		}
	}
	return false
}

// big5Bytes 把規格檔裡的字串（UTF-8）換回 Big5 位元組——字母表只有全形英文字母。
func big5Bytes(s string) []byte {
	var out []byte
	for _, r := range s {
		if r >= 0xFF21 && r <= 0xFF3A { // 全形 Ａ–Ｚ ＝ Big5 A2CF 起
			c := 0xA2CF + int(r-0xFF21)
			out = append(out, byte(c>>8), byte(c))
		}
	}
	return out
}
