package screen

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// FleetCommandSpec保存原版格式；發射量由Game按目前分隊供給。
type FleetCommandSpec struct {
	HeaderPrefix  string   `json:"header_prefix_hex"`
	HeaderSuffix  string   `json:"header_suffix_hex"`
	WeaponFormats []string `json:"weapon_formats_hex"`
	Choices       []string `json:"auto_defence_hex"`
	KindCodes     string   `json:"kind_codes_hex"`
	WeaponCodes   string   `json:"weapon_codes_hex"`
}

func LoadFleetCommandSpec(path string) (*FleetCommandSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FleetCommandSpec
	if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if len(s.WeaponFormats) != 3 || len(s.Choices) != 2 {
		return nil, fmt.Errorf("交戰命令格式數量不符")
	}
	for _, v := range append(append([]string{s.HeaderPrefix, s.HeaderSuffix, s.KindCodes, s.WeaponCodes}, s.WeaponFormats...), s.Choices...) {
		if b, err := hex.DecodeString(v); err != nil || len(b) == 0 {
			return nil, fmt.Errorf("交戰命令原始格式無效")
		}
	}
	if len(fmtBytes(s.KindCodes)) != 12 || len(fmtBytes(s.WeaponCodes)) != 3 {
		return nil, fmt.Errorf("交戰圖段名字表形狀不符")
	}
	return &s, nil
}

func (s *FleetCommandSpec) AnimationKind(kind int) byte     { return fmtBytes(s.KindCodes)[kind] }
func (s *FleetCommandSpec) AnimationWeapon(weapon int) byte { return fmtBytes(s.WeaponCodes)[weapon] }

func (s *FleetCommandSpec) Header(faction []byte, number []byte) []byte {
	b := ui.Sprintf(fmtBytes(s.HeaderPrefix), faction)
	b = append(b, number...)
	return append(b, fmtBytes(s.HeaderSuffix)...)
}

// 1ED2:0433／064D傳flag1給3982:0198。定位寬度與實際框內寬分開。
func (s *FleetCommandSpec) MenuAt(mx, my int, choice bool) (int, int) {
	width, rows := 8, 3
	if choice {
		width, rows = 4, 2
	}
	cx := min(max(1, mx/16-1), 29-(width+5))
	cy := min(max(1, my/20-1), 19-(rows+2))
	return cx*16 + 1, cy*20 + 1
}

// 3826:0298的3603:0006參數：格4,7、20×3、style1、colour97。
func DrawFleetCommandHeader(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles, raw []byte) {
	(ui.TileFrame{Tiles: tiles, Style: 1, Fill: 9}).Draw(fb, 4, 7, 22, 5)
	b := ui.NewTextBuffer(20, 3, 0x97)
	b.Write(raw)
	ui.DrawTextBuffer(fb, fnt, tiles, b, 5, 8)
}

func (s *FleetCommandSpec) Labels(shots [3]int, choice bool) [][]uint16 {
	var labels [][]uint16
	if choice {
		for _, h := range s.Choices {
			labels = append(labels, fleetCommandCodes(fmtBytes(h)))
		}
	} else {
		for i, h := range s.WeaponFormats {
			labels = append(labels, fleetCommandCodes(ui.Sprintf(fmtBytes(h), shots[i])))
		}
	}
	return labels
}

func fleetCommandCodes(raw []byte) []uint16 {
	var codes []uint16
	for i := 0; i < len(raw); i++ {
		if raw[i] >= 128 && i+1 < len(raw) {
			codes = append(codes, uint16(raw[i])<<8|uint16(raw[i+1]))
			i++
		} else {
			first := raw[i]
			second := byte(' ')
			if i+1 < len(raw) && raw[i+1] < 128 {
				second = raw[i+1]
				i++
			}
			codes = append(codes, uint16(first)<<8|uint16(second))
		}
	}
	return codes
}
