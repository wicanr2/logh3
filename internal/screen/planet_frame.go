package screen

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// PlanetFrameFormats保存118審查的原版DS字串，與舊056診斷格式分開。
type PlanetFrameFormats struct {
	Factions, Kinds                                    []string
	Colours                                            []uint8
	Attack, Ships, Unknown                             string
	ControlledPrefix, ControlledSuffix, ControlledName string
	KnownPrefix, KnownSuffix, KnownName                string
	Title, Name, Defence                               string
	Labels                                             [4]string
	Destroyed                                          string
}

type PlanetTextView struct {
	Owner, Kind         int
	Name                []byte
	Defence, DefenceMax int
}

type PlanetFleetTextView struct {
	Side                    uint8
	Number                  uint16
	AdmiralName             []byte
	Detail                  uint8 // 只看旗艦。
	Visibility              FleetVisibility
	Landing                 bool
	Attack, Ships, ShipsMax uint16
}

type PlanetMiddleView struct {
	Label                     EngagementLabel
	LabelColour, ResultColour uint8
	Destroyed                 uint16
	ShowResult                bool
}

// PlanetSceneView獨立保存最後一次321F:0897重畫的資料。
// 換艦隊或扣防衛力時，不能用目前上方文字取代此快照。
type PlanetSceneView struct {
	Units      []FleetSceneUnit
	Visibility FleetVisibility
	Facing     uint8
	X, Y       uint16
	Owner      int
	Defence    uint16
}

type PlanetFrameView struct {
	Planet   PlanetTextView
	Fleet    PlanetFleetTextView
	Middle   PlanetMiddleView
	Gauges   EngagementGaugePair
	Pictures []FleetScenePicture
	Scene    *PlanetSceneView // nil表示下方尚未首次重畫。
}

type PlanetSceneProjection struct {
	Units  []rules.FleetProjectedUnit
	Planet rules.FleetProjectedUnit
	Order  []int // -1為原版NULL行星。
	Angle  uint8
}

// ProjectPlanetScene沿321F單艦隊加行星NULL的投影及unsigned comb sort。
func ProjectPlanetScene(v PlanetSceneView) (PlanetSceneProjection, error) {
	var out PlanetSceneProjection
	if len(v.Units) == 0 || len(v.Units) > 22 || v.Owner < 0 || v.Owner > 2 {
		return out, fmt.Errorf("行星場景編制或擁有者無效")
	}
	side := v.Units[0].Side
	if side > 1 {
		return out, fmt.Errorf("行星場景陣營無效")
	}
	var sides [2][]rules.FleetProjectionUnit
	seen := map[int]bool{}
	for _, u := range v.Units {
		if u.Side != side || u.ID < 0 || u.ID >= 65535 || seen[u.ID] {
			return out, fmt.Errorf("行星場景分隊身份無效")
		}
		seen[u.ID] = true
		sides[side] = append(sides[side], u.FleetProjectionUnit)
		out.Order = append(out.Order, u.ID)
	}
	sides[1-side] = []rules.FleetProjectionUnit{{ID: 65535, Side: 1 - side, X: v.X, Y: v.Y}}
	p, err := rules.ProjectFleetScene(sides)
	if err != nil {
		return out, err
	}
	positions := map[int]rules.FleetProjectedUnit{}
	for _, u := range p.Units {
		positions[u.ID] = u
	}
	out.Order = append(out.Order, -1)
	position := func(id int) rules.FleetProjectedUnit {
		if id == -1 {
			id = 65535
		}
		return positions[id]
	}
	gap := len(out.Order)
	for {
		gap = max(1, gap*10/13)
		if gap == 9 || gap == 10 {
			gap = 11
		}
		swapped := false
		for i := 0; i+gap < len(out.Order); i++ {
			a, b := position(out.Order[i]), position(out.Order[i+gap])
			if uint16(a.Y) > uint16(b.Y) || uint16(a.Y) == uint16(b.Y) && uint16(a.X) > uint16(b.X) {
				out.Order[i], out.Order[i+gap] = out.Order[i+gap], out.Order[i]
				swapped = true
			}
		}
		if gap == 1 && !swapped {
			break
		}
	}
	for _, u := range v.Units {
		out.Units = append(out.Units, positions[u.ID])
	}
	out.Planet, out.Angle = positions[65535], p.Angle
	return out, nil
}

func PlanPlanetBlocks(v PlanetSceneView, sprites *asset.FleetSceneSprites) ([]FleetBlockOp, error) {
	if sprites == nil {
		return nil, fmt.Errorf("缺行星場景方塊素材")
	}
	p, err := ProjectPlanetScene(v)
	if err != nil {
		return nil, err
	}
	positions := map[int]rules.FleetProjectedUnit{-1: p.Planet}
	units := map[int]FleetSceneUnit{}
	for i, u := range v.Units {
		positions[u.ID] = p.Units[i]
		units[u.ID] = u
	}
	ops := []FleetBlockOp{{Kind: FleetBlockBackground}}
	for _, id := range p.Order {
		pos := positions[id]
		n, sprite := 0, 1
		if id == -1 {
			n = int(uint16(v.Defence+1999)) / 2000
			if v.Owner != 0 {
				sprite = 3
			}
		} else {
			if !pos.Visible {
				continue
			}
			u := units[id]
			n = int(uint16(u.Ships+499)) / 500
			control := v.Visibility.SideControl[1]
			if u.Side == v.Visibility.PlayerSide {
				control = v.Visibility.SideControl[0]
			}
			known := u.Squadron == 0 && (u.Detail == 3 || v.Visibility.Mode != 2 && control == 0)
			sprite = int(u.Side) * 2
			if !known {
				sprite++
			}
		}
		if n != 0 {
			ops = append(ops, FleetBlockOp{Kind: FleetBlockSprite, X: pos.X, Y: pos.Y, Count: n, Image: sprites.Blocks[sprite]})
		}
	}
	bx, by := 48, 200
	if v.Units[0].Side == 1 {
		bx, by = 592, 368
	}
	angle := int(uint8(v.Facing + p.Angle))
	for _, s := range [4][4]int{{16, 0, -16, -16}, {16, 0, -16, 16}, {-16, -16, 0, 0}, {-16, 16, 0, 0}} {
		x0, y0 := rules.Rotate256(s[0], s[1], angle)
		x1, y1 := rules.Rotate256(s[2], s[3], angle)
		ops = append(ops, FleetBlockOp{Kind: FleetBlockLine, X: x0 - y0/2 + bx, Y: y0/2 + by, X1: x1 - y1/2 + bx, Y1: y1/2 + by, Ink: 13})
	}
	return ops, nil
}

// RenderPlanetFrame從空白畫布合成118完整幀，各層互不借用狀態。
func (s *EngagementSpec) RenderPlanetFrame(v PlanetFrameView, panel *asset.Image,
	sprites *asset.FleetSceneSprites, comani *asset.BMG, fnt *font.Font, tiles *ui.FrameTiles) (*gfx.Framebuffer, error) {
	if s == nil || s.PlanetFrame == nil || panel == nil || fnt == nil || v.Fleet.Side > 1 {
		return nil, fmt.Errorf("缺行星交戰畫面資料")
	}
	f := s.PlanetFrame
	if v.Planet.Owner < 0 || v.Planet.Owner >= len(f.Factions) || v.Planet.Owner >= len(f.Colours) ||
		v.Planet.Kind < 0 || v.Planet.Kind >= len(f.Kinds) || v.Middle.Label < 0 || v.Middle.Label > EngagementFailed {
		return nil, fmt.Errorf("行星交戰文字欄位無效")
	}
	fb := gfx.NewFramebuffer(0)
	s.Draw(fb, fnt, tiles, panel, nil)
	pw, fw := s.Windows.Planet, s.Windows.Fleet
	if v.Fleet.Side == 0 {
		pw, fw = fw, pw
	}
	p := ui.NewTextBuffer(pw.W, pw.H, uint8(pw.Colour))
	p.Colour = uint8(pw.Colour&240) | f.Colours[v.Planet.Owner]
	p.Printf(fmtBytes(f.Title), fmtBytes(f.Factions[v.Planet.Owner]))
	p.Colour = uint8(pw.Colour)
	p.Printf(fmtBytes(f.Name), fmtBytes(f.Kinds[v.Planet.Kind]), v.Planet.Name)
	p.Printf(fmtBytes(f.Defence), v.Planet.Defence, v.Planet.DefenceMax)
	ui.DrawTextBuffer(fb, fnt, tiles, p, pw.X, pw.Y)
	fl := ui.NewTextBuffer(fw.W, fw.H, uint8(fw.Colour))
	control := v.Fleet.Visibility.SideControl[1]
	if v.Fleet.Side == v.Fleet.Visibility.PlayerSide {
		control = v.Fleet.Visibility.SideControl[0]
	}
	controlled := v.Fleet.Visibility.Mode != 2 && control == 0
	if controlled || v.Fleet.Detail == 3 {
		prefix, suffix, name := f.KnownPrefix, f.KnownSuffix, f.KnownName
		if controlled {
			prefix, suffix, name = f.ControlledPrefix, f.ControlledSuffix, f.ControlledName
		}
		fl.Write(fmtBytes(prefix))
		fl.Write(fleetNumberBytes(int(byte(v.Fleet.Number))))
		fl.Write(fmtBytes(suffix))
		fl.Printf(fmtBytes(name), v.Fleet.AdmiralName)
	} else {
		fl.Write(fmtBytes(f.Unknown))
	}
	if v.Fleet.Landing {
		fl.Printf(fmtBytes(f.Attack), v.Fleet.Attack)
	} else {
		fl.Printf(fmtBytes(f.Ships), v.Fleet.Ships, v.Fleet.ShipsMax)
	}
	ui.DrawTextBuffer(fb, fnt, tiles, fl, fw.X, fw.Y)
	w := s.Windows.Middle
	m := ui.NewTextBuffer(w.W, w.H, 4)
	m.Colour = v.Middle.LabelColour
	m.Write(fmtBytes(f.Labels[v.Middle.Label]))
	m.Colour = 4
	if v.Middle.ShowResult {
		m.Colour = v.Middle.ResultColour
		m.Printf(fmtBytes(f.Destroyed), v.Middle.Destroyed)
		m.Colour = 4
	}
	ui.DrawTextBuffer(fb, fnt, tiles, m, w.X, w.Y)
	if !v.Gauges.Draw(fb) {
		return nil, fmt.Errorf("行星交戰量表無效")
	}
	if len(v.Pictures) > 0 {
		if comani == nil {
			return nil, fmt.Errorf("缺行星交戰動畫庫")
		}
		planes := asset.PlanesFrom(fb.Pix)
		for _, pc := range v.Pictures {
			frames, err := comani.Frames(pc.Segment)
			if err != nil {
				return nil, err
			}
			if pc.Side > 1 || pc.LastFrame < 0 || pc.LastFrame >= len(frames) {
				return nil, fmt.Errorf("行星交戰動畫格無效")
			}
			x := 1
			if pc.Side == 1 {
				x = 47
			}
			for _, frame := range frames[:pc.LastFrame+1] {
				if err := frame.Draw(planes, asset.BMGWindowBase(x, 4)); err != nil {
					return nil, err
				}
			}
		}
		fb.CopyPixels(planes.Pixels())
	}
	if v.Scene != nil {
		ops, err := PlanPlanetBlocks(*v.Scene, sprites)
		if err != nil {
			return nil, err
		}
		if err := DrawFleetBlocks(fb, ops); err != nil {
			return nil, err
		}
	}
	return fb, nil
}
