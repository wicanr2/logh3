package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 戰術選單「資料」點一列開出來的艦隊詳細資料（`docs/re/112`、
// 規格 `docs/spec/022`）。
//
// 它與「資料」共用同一個外框，換掉的是裡面那一層：一塊內凹面板、
// 一張 80×100 的頭像，加上十四列文字。文字仍然是原版那套字元格引擎
// （一格 16 像素、全形一格半形半格），**每一行的靜態字與欄位位置都由
// `tools/scripts/fleetdetail_spec.py` 走一遍原版的格式字串算出來**，
// 這一層只負責把值填進去。

// FleetDetailStatic 是一行裡的一段固定字。
type FleetDetailStatic struct {
	X     int      `json:"x"`
	Codes []uint16 `json:"codes"`
}

// FleetDetailField 是一行裡的一個欄位。`W` 是半形格數（`%W` 是 2）。
type FleetDetailField struct {
	X   int    `json:"x"`
	W   int    `json:"w"`
	Fmt string `json:"fmt"`
}

// FleetDetailLine 是一整行。
type FleetDetailLine struct {
	Y      int                 `json:"y"`
	Static []FleetDetailStatic `json:"static"`
	Fields []FleetDetailField  `json:"fields"`
}

// FleetDetailValue 是能力那十二格的其中一格。
type FleetDetailValue struct {
	Name   string `json:"name"`
	X, Y   int
	W      int  `json:"w"`
	Source *int `json:"source"` // 顏色比的提督位移；nil ＝ 這一格沒有顏色參數
}

// FleetDetailName 是一個名字（艦種、武器）。
type FleetDetailName struct {
	Text  string   `json:"text"`
	Codes []uint16 `json:"codes"`
}

// FleetDetailSpec 是這一層的版面。
type FleetDetailSpec struct {
	Panel struct {
		X, Y, W, H int
		Frame      BattleDataFrame `json:"frame"`
	} `json:"panel"`
	Face struct{ X, Y int } `json:"face"`
	Cell struct {
		X, Y, Col, Row, Half, Rows int
	} `json:"cell"`
	// Rules 是純色的線（上下兩段之間那條）。
	Rules []struct {
		X, Y, W, H int
		Colour     uint8 `json:"colour"`
	} `json:"rules"`
	Lines       map[string]FleetDetailLine `json:"lines"`
	Values      []FleetDetailValue         `json:"values"`
	ShipKinds   []FleetDetailName          `json:"ship_kinds"`
	Weapons     []FleetDetailName          `json:"weapons"`
	NotEquipped []struct {
		Row  int             `json:"row"`
		Text string          `json:"text"`
		Line FleetDetailLine `json:"line"`
	} `json:"not_equipped"`
}

// LoadFleetDetailSpec 讀 `fleet-detail.json`。
func LoadFleetDetailSpec(path string) (*FleetDetailSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FleetDetailSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(s.Lines) == 0 || len(s.Values) != 12 {
		return nil, fmt.Errorf("%s: 版面不完整", path)
	}
	return &s, nil
}

// FleetDetailWeapon 是一組裝備要印的東西。
type FleetDetailWeapon struct {
	Name     []uint16 // 空 ＝ 沒搭載
	Cur, Max int
}

// FleetDetailView 是這一層要顯示的內容。
type FleetDetailView struct {
	Faction     []uint16 // 陣營名（用陣營色）
	FactionInk  uint8
	FleetNo     int
	Commander   []uint16 // 名字 ＋ 軍階
	CmdLevel    int
	CmdExp      int
	Adjutant    []uint16 // 空 ＝ 沒有副官
	AdjLevel    int
	AdjExp      int
	Values      [12]int  // 順序同規格的 values
	ValueInks   [12]uint8 // 0 ＝ 用預設色
	UnitNo      int
	KindName    []uint16
	Ships, Full int
	UnitMorale  int
	TopSpeed    int
	BattleSpeed int
	Durability  int
	Beam        FleetDetailWeapon
	Missile     FleetDetailWeapon
	Launcher    FleetDetailWeapon // Name 不用（原版那一格是固定的「發射裝置」）
	Fighter     FleetDetailWeapon
	Face        *asset.Image
}

// DrawFleetDetail 把這一層畫上去。回畫不出來的字碼。
func DrawFleetDetail(s *FleetDetailSpec, fb *gfx.Framebuffer, fnt *font.Font,
	v FleetDetailView) []uint16 {
	if s == nil || fb == nil {
		return nil
	}
	var missing []uint16
	s.Panel.Frame.frame().Draw(fb, s.Panel.X, s.Panel.Y, s.Panel.W, s.Panel.H)
	if v.Face != nil {
		for y := 0; y < v.Face.H; y++ {
			for x := 0; x < v.Face.W; x++ {
				fb.Set(s.Face.X+x, s.Face.Y+y, v.Face.Pix[y*v.Face.W+x])
			}
		}
	}
	for _, r := range s.Rules {
		fb.FillRect(r.X, r.Y, r.W, r.H, r.Colour)
	}
	ink, shadow := uint8(7), uint8(3)
	line := func(name string, ins ...[]uint16) {
		l, ok := s.Lines[name]
		if !ok {
			return
		}
		for _, st := range l.Static {
			missing = append(missing, drawCells(fb, fnt,
				cellsAt(st.Codes, st.X), l.Y, ink, shadow, false, 0, 0)...)
		}
		for i, f := range l.Fields {
			if i >= len(ins) || ins[i] == nil {
				continue
			}
			missing = append(missing, drawCells(fb, fnt,
				cellsAt(ins[i], f.X), l.Y, ink, shadow, false, 0, 0)...)
		}
	}
	// 陣營名用陣營色，其餘色 7。
	if l, ok := s.Lines["faction"]; ok && len(v.Faction) > 0 {
		fi := v.FactionInk
		if fi == 0 {
			fi = ink
		}
		missing = append(missing, drawCells(fb, fnt,
			cellsAt(v.Faction, l.Fields[0].X), l.Y, fi, shadow, false, 0, 0)...)
	}
	line("fleet_no", padLeft(v.FleetNo, 2))
	line("commander", v.Commander)
	line("cmd_level", padLeft(v.CmdLevel, 2), padLeft(v.CmdExp, 6))
	line("adjutant", v.Adjutant)
	if len(v.Adjutant) > 0 {
		line("adj_level", padLeft(v.AdjLevel, 2), padLeft(v.AdjExp, 6))
	}
	for i := 0; i < 4; i++ {
		line(fmt.Sprintf("grid%d", i))
	}
	line("unit_no", []uint16{fullDigit(v.UnitNo)})
	line("unit_ship", padRight(v.KindName, 8), padLeft(v.Ships, 4), padLeft(v.Full, 4))
	line("u_morale", padLeft(v.UnitMorale, 3))
	line("u_top", padLeft(v.TopSpeed, 3))
	line("u_battle", padLeft(v.BattleSpeed, 3))
	line("u_dur", padLeft(v.Durability, 3))
	weapon := func(name string, w FleetDetailWeapon, row int, withName bool) {
		if len(w.Name) == 0 && withName {
			for _, e := range s.NotEquipped {
				if e.Row == row {
					for _, st := range e.Line.Static {
						missing = append(missing, drawCells(fb, fnt,
							cellsAt(st.Codes, st.X), e.Line.Y, ink, shadow, false, 0, 0)...)
					}
				}
			}
			return
		}
		if withName {
			line(name, padRight(w.Name, 14), padLeft(w.Cur, 6), padLeft(w.Max, 6))
			return
		}
		line(name, padLeft(w.Cur, 6), padLeft(w.Max, 6))
	}
	weapon("w_beam", v.Beam, 10, true)
	weapon("w_missile", v.Missile, 11, true)
	weapon("w_launch", v.Launcher, 12, false)
	weapon("w_fighter", v.Fighter, 13, true)

	for i, val := range s.Values {
		vi := ink
		if v.ValueInks[i] != 0 {
			vi = v.ValueInks[i]
		}
		missing = append(missing, drawCells(fb, fnt,
			cellsAt(padLeft(v.Values[i], 3), val.X+8), val.Y, vi, shadow, false, 0, 0)...)
	}
	return missing
}

// cellsAt 把一串字碼排在 x 起：全形進 16、半形兩個併一格
// （與「資料」那一層同一套，見 `layoutCells`）。
func cellsAt(codes []uint16, x int) []BattleDataCell {
	cs, _ := layoutCells(codes, x)
	return cs
}

// padLeft 回靠右對齊的 w 個半形字碼（同原版的 `%Nu`）。
func padLeft(v, w int) []uint16 { return rightHalf(v, w) }

func rightHalf(v, w int) []uint16 {
	out := make([]uint16, w)
	for i := range out {
		out[i] = ' '
	}
	if v < 0 {
		v = 0
	}
	for i := w - 1; i >= 0; i-- {
		out[i] = uint16('0' + v%10)
		v /= 10
		if v == 0 {
			break
		}
	}
	return out
}

// padRight 回靠左對齊、寬度 w 個半形格的字碼（同原版的 `%-Ns`）。
// 全形字佔兩格；不足補半形空白。
func padRight(codes []uint16, w int) []uint16 {
	out := append([]uint16(nil), codes...)
	used := 0
	for _, c := range codes {
		if c >= 0x100 {
			used += 2
		} else {
			used++
		}
	}
	for ; used < w; used++ {
		out = append(out, ' ')
	}
	return out
}

// fullDigit 回全形數字的字碼（０ ＝ A2AF，往後連續）。
func fullDigit(n int) uint16 {
	if n < 0 || n > 9 {
		return ' '
	}
	return uint16(0xA2AF + n)
}
