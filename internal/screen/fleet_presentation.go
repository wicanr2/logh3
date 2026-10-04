package screen

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
	"os"
)

// FleetPresentationSpec 是096確認的艦隊交戰文字及logical色盤。
// 不持有交戰規則、投影、等待輸入或原版記憶體。
type FleetPresentationSpec struct {
	Schema      string              `json:"schema"`
	Windows     [3]EngagementWindow `json:"windows"` // 左、右、中
	Formats     map[string]string   `json:"formats"`
	WeaponNames []string            `json:"weapon_names"`
	Palette     gfx.Palette         `json:"palette"`
}

func LoadFleetPresentationSpec(path string) (*FleetPresentationSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FleetPresentationSpec
	if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	var shape struct {
		Windows []json.RawMessage `json:"windows"`
		Palette [][]uint8         `json:"palette"`
	}
	if err = json.Unmarshal(b, &shape); err != nil {
		return nil, err
	}
	if len(shape.Windows) != 3 || len(shape.Palette) != 16 {
		return nil, fmt.Errorf("交戰視窗或色盤長度錯誤")
	}
	for _, rgb := range shape.Palette {
		if len(rgb) != 3 {
			return nil, fmt.Errorf("交戰色盤分量長度錯誤")
		}
	}
	if s.Schema != "logh3.fleet-presentation-spec.v1" || len(s.WeaponNames) != 13 {
		return nil, fmt.Errorf("艦隊交戰格式版本或武器表形狀錯誤")
	}
	for i, w := range s.Windows {
		want := [][5]int{{1, 6, 15, 2, 228}, {24, 6, 15, 2, 196}, {17, 6, 6, 2, 4}}[i]
		if [5]int{w.X, w.Y, w.W, w.H, w.Colour} != want {
			return nil, fmt.Errorf("交戰視窗%d形狀錯誤", i)
		}
	}
	for _, key := range []string{"clear", "unknown", "prefix", "suffix", "squadron", "admiral", "ships", "weapon", "goto", "destroyed"} {
		raw, err := hex.DecodeString(s.Formats[key])
		if err != nil || len(raw) == 0 {
			return nil, fmt.Errorf("交戰格式%s錯誤", key)
		}
	}
	for _, name := range s.WeaponNames {
		raw, err := hex.DecodeString(name)
		if err != nil || len(raw) == 0 {
			return nil, fmt.Errorf("交戰武器名稱錯誤")
		}
	}
	for _, rgb := range s.Palette {
		for _, v := range rgb {
			if v > 63 {
				return nil, fmt.Errorf("交戰DAC值超界")
			}
		}
	}
	return &s, nil
}

// FleetVisibility保留3982:022D的兩格原始控制值；不是按陣營索引的設定。
type FleetVisibility struct {
	Mode, PlayerSide uint8
	SideControl      [2]uint8
}
type FleetUnitView struct {
	Side, Squadron, Detail uint8
	FleetNumber            uint16 // FLEET+2；顯示只消費低byte
	AdmiralName, KindName  []byte // Big5
	Ships, ShipsMax        uint16 // SHIP+8／+0A，分隊而非艦隊總數
	Visibility             FleetVisibility
}
type FleetWeaponView struct {
	Side, Weapon uint8
	WeaponTypes  [3]uint8 // 艦種參數+0A／+0E／+12
	ShowResult   bool
	Destroyed    uint16
}

// FleetResultView對應0F6D的loss及目標陣營，與0B90的攻方分開。
type FleetResultView struct {
	Side      uint8
	Destroyed uint16
}

func (s *FleetPresentationSpec) AppendResult(b *ui.TextBuffer, v FleetResultView) error {
	if s == nil || b == nil || v.Side > 1 {
		return fmt.Errorf("未支援的交戰結果狀態")
	}
	b.Printf(fmtBytes(s.Formats["goto"]), 0, 1)
	b.Colour = 2
	if v.Side == 1 {
		b.Colour = 13
	}
	b.Printf(fmtBytes(s.Formats["destroyed"]), int(v.Destroyed))
	b.Colour = 4
	return nil
}

func (s *FleetPresentationSpec) UnitBuffer(v FleetUnitView) (*ui.TextBuffer, error) {
	if s == nil || v.Side > 1 || v.Visibility.PlayerSide > 1 || v.Visibility.Mode > 2 {
		return nil, fmt.Errorf("未支援的交戰文字狀態")
	}
	w := s.Windows[v.Side]
	b := ui.NewTextBuffer(w.W, w.H, uint8(w.Colour))
	b.Write(fmtBytes(s.Formats["clear"]))
	control := v.Visibility.SideControl[1]
	if v.Side == v.Visibility.PlayerSide {
		control = v.Visibility.SideControl[0]
	}
	known := v.Detail == 3 || (v.Visibility.Mode != 2 && control == 0)
	if len(v.KindName) == 0 || (known && v.Squadron == 0 && len(v.AdmiralName) == 0) {
		return nil, fmt.Errorf("缺交戰艦種或提督名稱")
	}
	if known {
		b.Write(fmtBytes(s.Formats["prefix"]))
		b.Write(fleetNumberBytes(int(uint8(v.FleetNumber))))
		b.Write(fmtBytes(s.Formats["suffix"]))
		if v.Squadron == 0 {
			b.Printf(fmtBytes(s.Formats["admiral"]), v.AdmiralName)
		} else {
			b.Printf(fmtBytes(s.Formats["squadron"]), int(v.Squadron)+1)
		}
	} else {
		b.Write(fmtBytes(s.Formats["unknown"]))
	}
	b.Printf(fmtBytes(s.Formats["ships"]), v.KindName, int(v.Ships), int(v.ShipsMax))
	return b, nil
}

func (s *FleetPresentationSpec) MiddleBuffer(v FleetWeaponView) (*ui.TextBuffer, error) {
	if s == nil || v.Side > 1 || len(s.WeaponNames) != 13 {
		return nil, fmt.Errorf("未支援的交戰武器狀態")
	}
	w := s.Windows[2]
	b := ui.NewTextBuffer(w.W, w.H, uint8(w.Colour))
	kind := uint8(12)
	if v.Weapon < 3 {
		kind = v.WeaponTypes[v.Weapon]
		if kind > 11 {
			kind = 12
		}
	}
	colour := 2
	if v.Side == 1 {
		colour = 13
	}
	b.Printf(fmtBytes(s.Formats["weapon"]), colour, fmtBytes(s.WeaponNames[kind]), 4)
	if v.ShowResult {
		if err := s.AppendResult(b, FleetResultView{Side: v.Side, Destroyed: v.Destroyed}); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func (s *FleetPresentationSpec) DrawUnit(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles, v FleetUnitView) error {
	b, err := s.UnitBuffer(v)
	if err != nil {
		return err
	}
	if fb == nil {
		return fmt.Errorf("缺交戰畫布")
	}
	w := s.Windows[v.Side]
	ui.DrawTextBuffer(fb, fnt, tiles, b, w.X, w.Y)
	return nil
}
func (s *FleetPresentationSpec) DrawMiddle(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles, v FleetWeaponView) error {
	b, err := s.MiddleBuffer(v)
	if err != nil {
		return err
	}
	if fb == nil {
		return fmt.Errorf("缺交戰畫布")
	}
	w := s.Windows[2]
	ui.DrawTextBuffer(fb, fnt, tiles, b, w.X, w.Y)
	return nil
}
