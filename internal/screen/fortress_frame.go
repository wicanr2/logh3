package screen

import (
	"fmt"
	"image"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// FortressFrameFormats保留2264的原格式，不沿用艦隊交火的空格。
type FortressFrameFormats struct {
	Formats       map[string]string
	Cues          [6]int
	NoticeColours [3]int
}

// FortressFrameView的上下方快照各自獨立，不讀正式Game的末端狀態。
type FortressFrameView struct {
	Planet     PlanetTextView
	Target     FleetUnitView
	Gauges     EngagementGaugePair
	Pictures   []FleetScenePicture
	ClearSides []uint8
	Scene      *PlanetSceneView
	Destroyed  uint16
	ShowResult bool
}

func (s *EngagementSpec) RenderFortressFrame(v FortressFrameView, fs *FleetPresentationSpec,
	panel *asset.Image, sprites *asset.FleetSceneSprites, bmg *asset.BMG, fnt *font.Font, tiles *ui.FrameTiles) (*gfx.Framebuffer, error) {
	if s == nil || s.FortressFrame == nil || s.PlanetFrame == nil || fs == nil || panel == nil || fnt == nil ||
		v.Target.Side > 1 || v.Planet.Owner < 0 || v.Planet.Owner > 2 || v.Planet.Kind < 0 || v.Planet.Kind > 2 || v.Target.Ships > v.Target.ShipsMax {
		return nil, fmt.Errorf("缺要塞呈現資料或欄位無效")
	}
	f := s.FortressFrame.Formats
	fb := gfx.NewFramebuffer(0)
	s.Draw(fb, fnt, tiles, panel, nil)
	pw, fw := s.Windows.Planet, s.Windows.Fleet
	if v.Target.Side == 0 {
		pw, fw = fw, pw
	}
	p := ui.NewTextBuffer(pw.W, pw.H, uint8(pw.Colour))
	p.Colour = uint8(pw.Colour&240) | s.PlanetFrame.Colours[v.Planet.Owner]
	p.Printf(fmtBytes(f["title"]), fmtBytes(s.PlanetFrame.Factions[v.Planet.Owner]))
	p.Colour = uint8(pw.Colour)
	p.Printf(fmtBytes(f["name"]), fmtBytes(s.PlanetFrame.Kinds[v.Planet.Kind]), v.Planet.Name)
	p.Printf(fmtBytes(f["defence"]), v.Planet.Defence, v.Planet.DefenceMax)
	ui.DrawTextBuffer(fb, fnt, tiles, p, pw.X, pw.Y)
	local := *fs
	local.Formats = f
	u, err := local.UnitBuffer(v.Target)
	if err != nil {
		return nil, err
	}
	ui.DrawTextBuffer(fb, fnt, tiles, u, fw.X, fw.Y)
	w := s.Windows.Middle
	m := ui.NewTextBuffer(w.W, w.H, 4)
	m.Colour = s.PlanetFrame.Colours[v.Planet.Owner]
	m.Write(fmtBytes(f["label"]))
	m.Colour = 4
	if v.ShowResult {
		m.Printf(fmtBytes(f["goto"]), 0, 1)
		m.Colour = s.PlanetFrame.Colours[v.Target.Side]
		m.Printf(fmtBytes(f["destroyed"]), v.Destroyed)
		m.Colour = 4
	}
	ui.DrawTextBuffer(fb, fnt, tiles, m, w.X, w.Y)
	if !v.Gauges.Draw(fb) {
		return nil, fmt.Errorf("要塞量表無效")
	}
	for _, side := range v.ClearSides {
		if side > 1 {
			return nil, fmt.Errorf("要塞清圖側無效")
		}
		x := 8
		if side == 1 {
			x = 376
		}
		fb.FillRect(x, 56, 256, 102, 0)
	}
	if len(v.Pictures) > 0 {
		if bmg == nil {
			return nil, fmt.Errorf("缺要塞動畫庫")
		}
		planes := asset.PlanesFrom(fb.Pix)
		var bases [2]*asset.Image
		for _, p := range v.Pictures {
			frames, err := bmg.Frames(p.Segment)
			if err != nil {
				return nil, err
			}
			if p.Side > 1 || p.LastFrame < 0 || p.LastFrame >= len(frames) {
				return nil, fmt.Errorf("要塞動畫格超界")
			}
			x := 1
			if p.Side == 1 {
				x = 47
			}
			for _, f := range frames[:p.LastFrame+1] {
				if err := f.Draw(planes, asset.BMGWindowBase(x, 4)); err != nil {
					return nil, err
				}
			}
			bases[p.Side], err = bmg.StaticPicture(p.Segment[:2])
			if err != nil {
				return nil, err
			}
		}
		DrawAnimationPixels(fb, planes.Pixels(), []image.Rectangle{image.Rect(8, 56, 264, 158), image.Rect(376, 56, 632, 158)}, bases[:]...)
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

func (s *EngagementSpec) DrawFortressNotice(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles, owner int, name []byte) {
	if s == nil || s.FortressFrame == nil || fb == nil || fnt == nil || tiles == nil || owner < 0 || owner > 2 {
		return
	}
	(ui.TileFrame{Tiles: tiles, Style: 1, Fill: 9}).Draw(fb, 4, 7, 22, 5)
	b := ui.NewTextBuffer(20, 3, 0x97)
	b.Printf(fmtBytes(s.FortressFrame.Formats["notice"]), s.FortressFrame.NoticeColours[owner], name, 0x97)
	ui.DrawTextBuffer(fb, fnt, tiles, b, 5, 8)
}
