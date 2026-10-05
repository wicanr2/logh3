// 第109輪118 DRAFT單幀探針；不接Game、不借原版畫面像素。
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func read(p string) []byte { b, e := os.ReadFile(p); must(e); return b }
func raw(h string) []byte  { b, e := hex.DecodeString(h); must(e); return b }

type landingFrame struct {
	Name                        string
	View                        screen.EngagementView
	Formats                     map[string]string
	FactionText, PlanetKindText []byte
	Units                       []screen.FleetSceneUnit
	Visibility                  screen.FleetVisibility
	FleetSide                   int
	FleetColour, PlanetColour   uint8
	Facing                      uint8
	Lower                       bool
	PlanetPoint                 struct{ X, Y uint16 }
	ExpectedUnits               []struct {
		ID, X, Y, Depth int
		Visible         bool
	}
	ExpectedPlanet [2]int
	ExpectedOrder  []int
	Pictures       []struct {
		Side       uint8
		SegmentHex string
		AllFrames  bool
	}
	NativePath, NativeSHA256 string
}

// PLANET作另一側的唯一參考點；仍按原版先艦隊、末尾NULL的序列排序。
func lowerOps(c landingFrame, sprites *asset.FleetSceneSprites) []screen.FleetBlockOp {
	side := int(c.Units[0].Side)
	var sides [2][]rules.FleetProjectionUnit
	for _, u := range c.Units {
		sides[side] = append(sides[side], u.FleetProjectionUnit)
	}
	sides[1-side] = []rules.FleetProjectionUnit{{ID: 65535, Side: uint8(1 - side), X: c.PlanetPoint.X, Y: c.PlanetPoint.Y}}
	p, e := rules.ProjectFleetScene(sides)
	must(e)
	positions := map[int]rules.FleetProjectedUnit{}
	for _, u := range p.Units {
		positions[u.ID] = u
	}
	order := make([]int, 0, len(c.Units)+1)
	byID := map[int]screen.FleetSceneUnit{}
	for _, u := range c.Units {
		order = append(order, u.ID)
		byID[u.ID] = u
	}
	order = append(order, 65535)
	gap := len(order)
	for {
		gap = max(1, gap*10/13)
		if gap == 9 || gap == 10 {
			gap = 11
		}
		swapped := false
		for i := 0; i+gap < len(order); i++ {
			a, b := positions[order[i]], positions[order[i+gap]]
			if uint16(a.Y) > uint16(b.Y) || (uint16(a.Y) == uint16(b.Y) && uint16(a.X) > uint16(b.X)) {
				order[i], order[i+gap] = order[i+gap], order[i]
				swapped = true
			}
		}
		if gap == 1 && !swapped {
			break
		}
	}
	for _, want := range c.ExpectedUnits {
		got := positions[want.ID]
		if got.X != want.X || got.Y != want.Y || got.Depth != want.Depth || got.Visible != want.Visible {
			panic(fmt.Sprintf("投影不符 %s ID%d got%+v want%+v", c.Name, want.ID, got, want))
		}
	}
	planet := positions[65535]
	if [2]int{planet.X, planet.Y} != c.ExpectedPlanet {
		panic("行星投影不符 " + c.Name)
	}
	for i, id := range order {
		if id == 65535 {
			id = -1
		}
		if id != c.ExpectedOrder[i] {
			panic("NULL排序不符 " + c.Name)
		}
	}
	ops := []screen.FleetBlockOp{{Kind: screen.FleetBlockBackground}}
	for _, id := range order {
		pos := positions[id]
		n, sprite := 0, 0
		if id == 65535 {
			n = int(uint16(c.View.Defence+1999)) / 2000
			sprite = 1
			if c.View.Faction != 0 {
				sprite = 3
			}
		} else {
			if !pos.Visible {
				continue
			}
			u := byID[id]
			n = int(uint16(u.Ships+499)) / 500
			control := c.Visibility.SideControl[1]
			if u.Side == c.Visibility.PlayerSide {
				control = c.Visibility.SideControl[0]
			}
			known := u.Squadron == 0 && (u.Detail == 3 || (c.Visibility.Mode != 2 && control == 0))
			sprite = int(u.Side) * 2
			if !known {
				sprite++
			}
		}
		if n != 0 {
			ops = append(ops, screen.FleetBlockOp{Kind: screen.FleetBlockSprite, X: pos.X, Y: pos.Y, Count: n, Image: sprites.Blocks[sprite]})
		}
	}
	bx, by := 48, 200
	if side == 1 {
		bx, by = 592, 368
	}
	angle := int(uint8(c.Facing + p.Angle))
	for _, s := range [4][4]int{{16, 0, -16, -16}, {16, 0, -16, 16}, {-16, -16, 0, 0}, {-16, 16, 0, 0}} {
		x0, y0 := rules.Rotate256(s[0], s[1], angle)
		x1, y1 := rules.Rotate256(s[2], s[3], angle)
		ops = append(ops, screen.FleetBlockOp{Kind: screen.FleetBlockLine, X: x0 - y0/2 + bx, Y: y0/2 + by, X1: x1 - y1/2 + bx, Y1: y1/2 + by, Ink: 13})
	}
	return ops
}

func render(c landingFrame, s *screen.EngagementSpec, panel *asset.Image, fnt *font.Font, bmg *asset.BMG, sprites *asset.FleetSceneSprites) *gfx.Framebuffer {
	fb := gfx.NewFramebuffer(0)
	v := c.View
	v.Gauges = nil
	local := *s
	local.Formats.FleetShips = c.Formats["0x483f"]
	local.Formats.FleetKnownPrefix = c.Formats["fleet_prefix"]
	local.Formats.FleetKnownSuffix = c.Formats["fleet_suffix"]
	local.Formats.FleetKnownName = c.Formats["fleet_name"]
	if c.FleetSide == 0 {
		local.Windows.Planet, local.Windows.Fleet = local.Windows.Fleet, local.Windows.Planet
	}
	local.Draw(fb, fnt, nil, panel, nil)
	_, _, fleet := local.Buffers(&v)
	w := local.Windows.Fleet
	ui.DrawTextBuffer(fb, fnt, nil, fleet, w.X, w.Y)
	w = local.Windows.Planet
	planet := ui.NewTextBuffer(w.W, w.H, uint8(w.Colour))
	planet.Colour = uint8(w.Colour&240) | c.PlanetColour
	planet.Printf(raw(c.Formats["0x4882"]), c.FactionText)
	planet.Colour = uint8(w.Colour)
	planet.Printf(raw(c.Formats["0x4886"]), c.PlanetKindText, v.PlanetName)
	planet.Printf(raw(c.Formats["0x488c"]), v.Defence, v.DefenceMax)
	ui.DrawTextBuffer(fb, fnt, nil, planet, w.X, w.Y)
	w = local.Windows.Middle
	middle := ui.NewTextBuffer(w.W, w.H, 4)
	middle.Colour = c.FleetColour
	middle.Write(raw(c.Formats["0x485b"]))
	middle.Colour = 4
	if v.ShowResult {
		middle.Colour = c.PlanetColour
		middle.Printf(raw(c.Formats["0x48b1"]), v.Destroyed)
		middle.Colour = 4
	}
	if v.Label == screen.EngagementCaptured {
		middle.Write(raw(c.Formats["0x48d1"]))
	}
	ui.DrawTextBuffer(fb, fnt, nil, middle, w.X, w.Y)
	if !c.View.Gauges.Draw(fb) {
		panic("量表參數不符")
	}
	planes := asset.PlanesFrom(fb.Pix)
	for _, p := range c.Pictures {
		frames, e := bmg.Frames(string(raw(p.SegmentHex)))
		must(e)
		last := 0
		if p.AllFrames {
			last = len(frames) - 1
		}
		x := 1
		if p.Side == 1 {
			x = 47
		}
		for _, f := range frames[:last+1] {
			must(f.Draw(planes, asset.BMGWindowBase(x, 4)))
		}
	}
	copy(fb.Pix, planes.Pixels())
	if c.Lower {
		must(screen.DrawFleetBlocks(fb, lowerOps(c, sprites)))
	}
	return fb
}

func main() {
	if len(os.Args) != 3 {
		panic("probe <typed.json> <existing-output-directory>")
	}
	var receipt struct {
		Schema string
		Cases  []landingFrame
	}
	must(json.Unmarshal(read(os.Args[1]), &receipt))
	if receipt.Schema != "logh3.landing-frame-probe.v1" {
		panic("typed版本不符")
	}
	s, e := screen.LoadEngagementSpec("docs/internals/data/engagement-panel.json")
	must(e)
	palette, e := screen.LoadFleetPresentationSpec("docs/internals/data/fleet-presentation-spec.json")
	must(e)
	panel, e := asset.DecodeAMGImage(read("/orig/ANI_PNL1.A98"))
	must(e)
	fnt, e := font.Load("/orig/GRAPH.IMG")
	must(e)
	bmg, e := asset.OpenBMG(read("/orig/COMANI.BMG"))
	must(e)
	sprites, e := asset.DecodeFleetSceneSprites(read("/orig/3DBLOCK.SPR"), read("/orig/COM_FLAG.SPR"))
	must(e)
	var stats []map[string]any
	for _, c := range receipt.Cases {
		fb := render(c, s, panel, fnt, bmg, sprites)
		// 原版像素只在合成完成後讀取，不能進renderer或ops。
		want := read(c.NativePath)
		if len(want) != len(fb.Pix) || fmt.Sprintf("%x", sha256.Sum256(want)) != c.NativeSHA256 {
			panic("原版畫面身份不符")
		}
		native, diff := gfx.NewFramebuffer(0), gfx.NewFramebuffer(0)
		copy(native.Pix, want)
		n, maxChannel := 0, 0
		box := [4]int{640, 480, -1, -1}
		for i, p := range fb.Pix {
			if p == want[i] {
				continue
			}
			n++
			diff.Pix[i] = 15
			x, y := i%640, i/640
			box = [4]int{min(box[0], x), min(box[1], y), max(box[2], x), max(box[3], y)}
			r, g, b := palette.Palette.RGB8(int(p))
			wr, wg, wb := palette.Palette.RGB8(int(want[i]))
			for _, d := range []int{int(r) - int(wr), int(g) - int(wg), int(b) - int(wb)} {
				maxChannel = max(maxChannel, d, -d)
			}
		}
		var regions []map[string]any
		for _, r := range []struct {
			Name       string
			X, Y, W, H int
		}{{"left-picture", 8, 56, 256, 102}, {"right-picture", 376, 56, 256, 102}, {"left-text", 16, 160, 240, 40}, {"right-text", 384, 160, 240, 40}, {"middle-text", 272, 160, 96, 40}, {"left-gauge", 298, 63, 11, 66}, {"right-gauge", 333, 63, 11, 66}, {"lower", 8, 208, 624, 224}} {
			different := 0
			for y := r.Y; y < r.Y+r.H; y++ {
				for x := r.X; x < r.X+r.W; x++ {
					i := y*640 + x
					if fb.Pix[i] != want[i] {
						different++
					}
				}
			}
			regions = append(regions, map[string]any{"name": r.Name, "rect": [4]int{r.X, r.Y, r.W, r.H}, "pixels": r.W * r.H, "same": r.W*r.H - different, "different": different})
		}
		var bbox any
		if n != 0 {
			bbox = box
		}
		stats = append(stats, map[string]any{"case": c.Name, "classification": "same-state complete typed diagnostic;118 DRAFT;not Game player path", "pixels": len(want), "same": len(want) - n, "different": n, "max_channel_difference": maxChannel, "difference_bbox": bbox, "regions": regions})
		for name, im := range map[string]*gfx.Framebuffer{"original": native, "remake": fb, "diff": diff} {
			must(im.WritePNG(filepath.Join(os.Args[2], c.Name+"-"+name+".png"), palette.Palette))
		}
		must(os.WriteFile(filepath.Join(os.Args[2], c.Name+"-remake.bin"), fb.Pix, 0644))
		fmt.Println(c.Name, "different", n, "bbox", bbox)
	}
	data, e := json.MarshalIndent(stats, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(os.Args[2], "comparison.json"), append(data, '\n'), 0644))
}
