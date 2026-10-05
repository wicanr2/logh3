// 第110輪118 READY候選renderer驗證；不接Game、不借原版畫面像素。
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
		Count      int
	}
	Landing                   bool
	LabelColour, ResultColour uint8
	LowerView                 *landingFrame
	PlanetKind                int
	FleetDetail               uint8
	NativePath, NativeSHA256  string
}

func render(c landingFrame, s *screen.EngagementSpec, panel *asset.Image, fnt *font.Font, bmg *asset.BMG, sprites *asset.FleetSceneSprites) *gfx.Framebuffer {
	v := screen.PlanetFrameView{
		Planet: screen.PlanetTextView{Owner: c.View.Faction, Kind: c.PlanetKind, Name: c.View.PlanetName, Defence: c.View.Defence, DefenceMax: c.View.DefenceMax},
		Fleet:  screen.PlanetFleetTextView{Side: uint8(c.FleetSide), Number: uint16(c.View.FleetNumber), AdmiralName: c.View.AdmiralName, Detail: c.FleetDetail, Visibility: c.Visibility, Landing: c.Landing, Attack: uint16(c.View.Ships), Ships: uint16(c.View.Ships), ShipsMax: uint16(c.View.ShipsMax)},
		Middle: screen.PlanetMiddleView{Label: c.View.Label, LabelColour: c.LabelColour, ResultColour: c.ResultColour, Destroyed: uint16(c.View.Destroyed), ShowResult: c.View.ShowResult},
		Gauges: *c.View.Gauges,
	}
	for _, pc := range c.Pictures {
		v.Pictures = append(v.Pictures, screen.FleetScenePicture{Side: pc.Side, Segment: string(raw(pc.SegmentHex)), LastFrame: pc.Count - 1})
	}
	if c.Lower {
		x := c.LowerView
		v.Scene = &screen.PlanetSceneView{Units: x.Units, Visibility: x.Visibility, Facing: x.Facing, X: x.PlanetPoint.X, Y: x.PlanetPoint.Y, Owner: x.View.Faction, Defence: uint16(x.View.Defence)}
		projection, e := screen.ProjectPlanetScene(*v.Scene)
		must(e)
		for i, w := range x.ExpectedUnits {
			g := projection.Units[i]
			if g.ID != w.ID || g.X != w.X || g.Y != w.Y || g.Depth != w.Depth || g.Visible != w.Visible {
				panic("正式投影與原池不符 " + c.Name)
			}
		}
		if [2]int{projection.Planet.X, projection.Planet.Y} != x.ExpectedPlanet {
			panic("正式行星投影不符")
		}
		for i, id := range projection.Order {
			if id != x.ExpectedOrder[i] {
				panic("正式NULL排序不符")
			}
		}
	}
	fb, e := s.RenderPlanetFrame(v, panel, sprites, bmg, fnt, nil)
	must(e)
	return fb
}

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		panic("candidate <typed.json> <existing-output-directory> [candidate-panel.json]")
	}
	var receipt struct {
		Schema string
		Cases  []landingFrame
	}
	must(json.Unmarshal(read(os.Args[1]), &receipt))
	if receipt.Schema != "logh3.landing-caller-probe.v1" {
		panic("typed版本不符")
	}
	panelPath := "workplace/landing48-r110-panel-candidate.json"
	if len(os.Args) == 4 {
		panelPath = os.Args[3]
	}
	s, e := screen.LoadEngagementSpec(panelPath)
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
		stats = append(stats, map[string]any{"case": c.Name, "classification": "same-state complete candidate renderer diagnostic;118 READY;not Game player path", "pixels": len(want), "same": len(want) - n, "different": n, "max_channel_difference": maxChannel, "difference_bbox": bbox, "regions": regions})
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
