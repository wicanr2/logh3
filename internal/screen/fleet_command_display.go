package screen

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

type FleetCommandDisplaySpec struct {
	Formats map[string]string `json:"formats_hex"`
}

func LoadFleetCommandDisplaySpec(path string) (*FleetCommandDisplaySpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FleetCommandDisplaySpec
	if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	for _, k := range []string{"enemy_unknown_fleet", "enemy_unknown_name", "enemy_known_ships", "enemy_unknown_ships", "enemy_distance"} {
		if b, err := hex.DecodeString(s.Formats[k]); err != nil || len(b) == 0 {
			return nil, fmt.Errorf("交戰摘要格式缺失：%s", k)
		}
	}
	return &s, nil
}

// DrawFleetCombatRange沿165F:034F及0505／060B／071A／07CA。
func DrawFleetCombatRange(s *BattleSpec, fb *gfx.Framebuffer, x, y, facing int, ranges [3]rules.FleetWeaponRange, v BattleView) {
	if s == nil || fb == nil {
		return
	}
	cx, cy := x+8, y+8
	ox, oy := s.Field.X-v.Col*s.Field.TileW, s.Field.Y-v.Row*s.Field.TileH
	clip := FieldClip(s)
	for w, r := range ranges {
		if !r.Enabled {
			continue
		}
		if w == 2 {
			xorFleetCombatCircle(fb, cx+ox, cy+oy, r.Major, 5, clip)
			continue
		}
		ink := uint8(6)
		if w == 1 {
			ink = 1
		}
		for _, p := range FortressRangePoints(cx, cy, r.Major, r.Minor, facing) {
			xx, yy := p[0]+ox, p[1]+oy
			if clip.Inside(xx, yy) {
				fb.Set(xx, yy, fb.At(xx, yy)^ink)
			}
		}
	}
	for _, a := range [][4]int{{-2048, 0, 2048, 0}, {0, -2048, 0, 2048}} {
		x0, y0 := rules.Rotate256(a[0], a[1], facing)
		x1, y1 := rules.Rotate256(a[2], a[3], facing)
		clip.XorLine(fb, cx+ox+x0, cy+oy+y0, cx+ox+x1, cy+oy+y1, 7)
	}
}

// 39D7:008C：誤差初值為半徑，先使用舊y更新，再遞增y。
// 八個對稱端點照原版逐次XOR，不合併重複點。
func xorFleetCombatCircle(fb *gfx.Framebuffer, cx, cy, radius int, ink uint8, clip ClipRect) {
	if radius <= 0 {
		return
	}
	x, y, e := radius, 0, radius
	for x >= y {
		for _, p := range [8][2]int{{x, y}, {x, -y}, {-x, y}, {-x, -y}, {y, x}, {y, -x}, {-y, x}, {-y, -x}} {
			px, py := cx+p[0], cy+p[1]
			if clip.Inside(px, py) {
				fb.Set(px, py, fb.At(px, py)^ink)
			}
		}
		e -= 2*y - 1
		y++
		if e < 0 {
			e += 2 * (x - 1)
			x--
		}
	}
}

// 我方僅四列；敵方後四列不借用完整八列武器與行動面板。
func (s *FleetCommandDisplaySpec) DrawPanel(panel *UnitPanelSpec, fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles, own UnitPanelView, target *UnitPanelView, detail, distance int, faction, kind []byte) {
	if panel == nil || fb == nil {
		return
	}
	x, y := panel.Cell.X, panel.Cell.Y-ui.TextCellDY
	w, h := 8*panel.Cell.Col, 8*panel.Cell.Row
	fb.FillRect(x, y, w, h, 9)
	DrawUnitPanelSummary(panel, fb, fnt, own)
	if target == nil {
		return
	}
	p := *panel
	p.Cell.Y += 4 * p.Cell.Row
	if detail == 3 {
		DrawUnitPanelSummary(&p, fb, fnt, *target)
		b := ui.NewTextBuffer(8, 1, 0x97)
		b.Printf(fmtBytes(s.Formats["enemy_distance"]), distance)
		ui.DrawTextBuffer(fb, fnt, tiles, b, x/ui.FrameTileW, (y+7*panel.Cell.Row-ui.FrameCellOriginY)/ui.FrameTileH)
		return
	}
	b := ui.NewTextBuffer(8, 4, 0x97)
	b.Printf(fmtBytes(s.Formats["enemy_unknown_fleet"]), faction)
	b.Write(fmtBytes(s.Formats["enemy_unknown_name"]))
	if detail == 2 {
		b.Printf(fmtBytes(s.Formats["enemy_known_ships"]), kind, target.Ships)
	} else {
		b.Write(fmtBytes(s.Formats["enemy_unknown_ships"]))
	}
	b.Printf(fmtBytes(s.Formats["enemy_distance"]), distance)
	ui.DrawTextBuffer(fb, fnt, tiles, b, x/ui.FrameTileW, (y+4*panel.Cell.Row-ui.FrameCellOriginY)/ui.FrameTileH)
}
