package game

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
)

// Real decoded COMANI frames, without relying on optional old memory dumps.
func TestFleetVisualCacheNativeFrames(t *testing.T) {
	read := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(root, "org_game/extracted", name))
		if os.IsNotExist(e) {
			t.Skip("缺原版素材")
		}
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	panel, e := asset.DecodeAMGImage(read("ANI_PNL1.A98"))
	if e != nil {
		t.Fatal(e)
	}
	bmg, e := asset.OpenBMG(read("COMANI.BMG"))
	if e != nil {
		t.Fatal(e)
	}
	sprites, e := asset.DecodeFleetSceneSprites(read("3DBLOCK.SPR"), read("COM_FLAG.SPR"))
	if e != nil {
		t.Fatal(e)
	}
	fnt, e := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if e != nil {
		t.Fatal(e)
	}
	spec, e := screen.LoadFleetPresentationSpec(filepath.Join(root, "docs/internals/data/fleet-presentation-spec.json"))
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "docs/internals/data/fleet-frame.json"))
	if e != nil {
		t.Fatal(e)
	}
	var receipt struct {
		Cases []struct {
			View     screen.FleetFrameView
			Pictures []struct {
				Side       uint8
				SegmentHex string
				AllFrames  bool
			}
		}
	}
	if e = json.Unmarshal(raw, &receipt); e != nil {
		t.Fatal(e)
	}
	checked := 0
	for _, c := range receipt.Cases {
		v := c.View
		var cache fleetCombatSceneState
		for _, picture := range c.Pictures {
			name, e := hex.DecodeString(picture.SegmentHex)
			if e != nil {
				t.Fatal(e)
			}
			frames, e := bmg.Frames(string(name))
			if e != nil {
				t.Fatal(e)
			}
			v.Pictures = append(v.Pictures, screen.FleetScenePicture{Side: picture.Side, Segment: string(name)})
			for i := range frames {
				v.Pictures[len(v.Pictures)-1].LastFrame = i
				build := func() (*gfx.Framebuffer, error) { return spec.RenderFleetFrame(v, panel, sprites, bmg, fnt, nil) }
				got, e := cache.renderFrame(v, build)
				if e != nil {
					t.Fatal(e)
				}
				hit, e := cache.renderFrame(v, func() (*gfx.Framebuffer, error) { t.Fatal("same view rebuilt"); return nil, nil })
				if e != nil || hit != got {
					t.Fatal("cache miss")
				}
				fresh, e := build()
				if e != nil {
					t.Fatal(e)
				}
				if !bytes.Equal(got.Pix, fresh.Pix) || !reflect.DeepEqual(got.Commands, fresh.Commands) || !reflect.DeepEqual(got.Artwork, fresh.Artwork) {
					t.Fatal("cached native frame differs")
				}
				checked++
				if !picture.AllFrames {
					break
				}
			}
		}
	}
	if checked < 4 {
		t.Fatal("insufficient native frames")
	}
	t.Logf("%d decoded native frames match uncached pixels, commands and artwork", checked)
}

func TestFleetFrameCacheMatchesReconstruction(t *testing.T) {
	m := fleetSceneReceipt(t)
	g := fleetChainRAM(t, m, m.read(t, "workplace/oracle/scene48-r69-v2-laser-a/start-mem.bin"))
	g.fleetCombat = fleetCombatState{active: true, fleets: []int{12}, current: -1}
	g.beginFleetCombatScenes()
	s := g.fleetCombat.scene
	if s == nil {
		t.Fatal("missing fleet combat scene")
	}
	for tick := 0; tick <= s.resultTick+10; tick++ {
		s.tick = tick
		seed := g.battleRand.Seed()
		cached := g.Frame()
		if g.Frame() != cached {
			t.Fatal("same visual view was rebuilt")
		}
		s.frameCache = nil
		fresh := g.Frame()
		if !bytes.Equal(cached.Pix, fresh.Pix) || !reflect.DeepEqual(cached.Commands, fresh.Commands) || !reflect.DeepEqual(cached.Artwork, fresh.Artwork) {
			t.Fatalf("cache changed visual output at tick %d", tick)
		}
		if g.battleRand.Seed() != seed {
			t.Fatal("drawing consumed random state")
		}
	}
}
