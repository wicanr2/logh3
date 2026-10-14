// 144 DRAFT完整單幀診斷；原版像素僅在合成後作比較。
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
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func read(p string) []byte { b, e := os.ReadFile(p); must(e); return b }
func raw(s string) []byte  { b, e := hex.DecodeString(s); must(e); return b }

type fortressFrame struct {
	Name       string
	Planet     screen.PlanetTextView
	Target     screen.FleetUnitView
	Gauges     screen.EngagementGaugePair
	ClearSides []uint8
	Pictures   []struct {
		Side       uint8
		SegmentHex string
		LastFrame  int
	}
	Scene                    *screen.PlanetSceneView
	Destroyed                uint16
	ShowResult               bool
	NativePath, NativeSHA256 string
}

func render(c fortressFrame, formats map[string]string, s *screen.EngagementSpec, fs *screen.FleetPresentationSpec,
	panel *asset.Image, fnt *font.Font, bmg *asset.BMG, sprites *asset.FleetSceneSprites) *gfx.Framebuffer {
	fb := gfx.NewFramebuffer(0)
	s.Draw(fb, fnt, nil, panel, nil)
	pw, fw := s.Windows.Planet, s.Windows.Fleet
	if c.Target.Side == 0 {
		pw, fw = fw, pw
	}
	p := ui.NewTextBuffer(pw.W, pw.H, uint8(pw.Colour))
	p.Colour = uint8(pw.Colour&240) | s.PlanetFrame.Colours[c.Planet.Owner]
	p.Printf(raw(formats["title"]), raw(s.PlanetFrame.Factions[c.Planet.Owner]))
	p.Colour = uint8(pw.Colour)
	p.Printf(raw(formats["name"]), raw(s.PlanetFrame.Kinds[c.Planet.Kind]), c.Planet.Name)
	p.Printf(raw(formats["defence"]), c.Planet.Defence, c.Planet.DefenceMax)
	ui.DrawTextBuffer(fb, fnt, nil, p, pw.X, pw.Y)
	local := *fs
	local.Formats = formats
	u, e := local.UnitBuffer(c.Target)
	must(e)
	ui.DrawTextBuffer(fb, fnt, nil, u, fw.X, fw.Y)
	w := s.Windows.Middle
	m := ui.NewTextBuffer(w.W, w.H, 4)
	m.Colour = s.PlanetFrame.Colours[c.Planet.Owner]
	m.Write(raw(formats["label"]))
	m.Colour = 4
	if c.ShowResult {
		m.Printf(raw(formats["goto"]), 0, 1)
		m.Colour = s.PlanetFrame.Colours[c.Target.Side]
		m.Printf(raw(formats["destroyed"]), c.Destroyed)
		m.Colour = 4
	}
	ui.DrawTextBuffer(fb, fnt, nil, m, w.X, w.Y)
	if !c.Gauges.Draw(fb) {
		panic("要塞量表無效")
	}
	for _, side := range c.ClearSides {
		x := 8
		if side == 1 {
			x = 376
		}
		fb.FillRect(x, 56, 256, 102, 0)
	}
	planes := asset.PlanesFrom(fb.Pix)
	for _, p := range c.Pictures {
		frames, e := bmg.Frames(string(raw(p.SegmentHex)))
		must(e)
		x := 1
		if p.Side == 1 {
			x = 47
		}
		for _, f := range frames[:p.LastFrame+1] {
			must(f.Draw(planes, asset.BMGWindowBase(x, 4)))
		}
	}
	copy(fb.Pix, planes.Pixels())
	if c.Scene != nil {
		ops, e := screen.PlanPlanetBlocks(*c.Scene, sprites)
		must(e)
		must(screen.DrawFleetBlocks(fb, ops))
	}
	return fb
}

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		panic("probe typed.json new-output-directory")
	}
	must(os.Mkdir(os.Args[2], 0755))
	var data struct {
		Cases   []fortressFrame
		Formats map[string]string
	}
	must(json.Unmarshal(read(os.Args[1]), &data))
	s, e := screen.LoadEngagementSpec("docs/internals/data/engagement-panel.json")
	must(e)
	fs, e := screen.LoadFleetPresentationSpec("docs/internals/data/fleet-presentation-spec.json")
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
	classification := "same-state complete typed diagnostic;144 DRAFT;not Game player path"
	if len(os.Args) == 4 {
		classification = "same-state complete typed production renderer;144 READY;not Game player path"
	}
	for _, c := range data.Cases {
		var fb *gfx.Framebuffer
		if len(os.Args) == 4 {
			v := screen.FortressFrameView{Planet: c.Planet, Target: c.Target, Gauges: c.Gauges, ClearSides: c.ClearSides, Scene: c.Scene, Destroyed: c.Destroyed, ShowResult: c.ShowResult}
			for _, p := range c.Pictures {
				v.Pictures = append(v.Pictures, screen.FleetScenePicture{Side: p.Side, Segment: string(raw(p.SegmentHex)), LastFrame: p.LastFrame})
			}
			var err error
			fb, err = s.RenderFortressFrame(v, fs, panel, sprites, bmg, fnt, nil)
			must(err)
		} else {
			fb = render(c, data.Formats, s, fs, panel, fnt, bmg, sprites)
		}
		want := read(c.NativePath)
		if len(want) != len(fb.Pix) || fmt.Sprintf("%x", sha256.Sum256(want)) != c.NativeSHA256 {
			panic("原版來源不符")
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
			r, g, b := fs.Palette.RGB8(int(p))
			wr, wg, wb := fs.Palette.RGB8(int(want[i]))
			for _, d := range []int{int(r) - int(wr), int(g) - int(wg), int(b) - int(wb)} {
				maxChannel = max(maxChannel, d, -d)
			}
		}
		var regions []map[string]any
		for _, r := range []struct {
			Name       string
			X, Y, W, H int
		}{
			{"left-picture", 8, 56, 256, 102}, {"right-picture", 376, 56, 256, 102},
			{"left-text", 16, 160, 240, 40}, {"right-text", 384, 160, 240, 40}, {"middle-text", 272, 160, 96, 40},
			{"left-gauge", 298, 63, 11, 66}, {"right-gauge", 333, 63, 11, 66}, {"lower", 8, 208, 624, 224}} {
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
		stats = append(stats, map[string]any{"case": c.Name, "classification": classification,
			"pixels": len(want), "same": len(want) - n, "different": n, "max_channel_difference": maxChannel, "difference_bbox": bbox, "regions": regions})
		for name, im := range map[string]*gfx.Framebuffer{"original": native, "remake": fb, "diff": diff} {
			must(im.WritePNG(filepath.Join(os.Args[2], c.Name+"-"+name+".png"), fs.Palette))
		}
		must(os.WriteFile(filepath.Join(os.Args[2], c.Name+"-remake.bin"), fb.Pix, 0644))
		if n != 0 {
			fmt.Println(c.Name, "different", n, "bbox", bbox)
		}
	}
	b, e := json.MarshalIndent(stats, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(os.Args[2], "comparison.json"), append(b, '\n'), 0644))
	fmt.Println("complete frames", len(stats))
}
