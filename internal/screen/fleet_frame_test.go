package screen

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

func TestFleetFrameNative(t *testing.T) {
	raw, err := os.ReadFile("../../docs/internals/data/fleet-frame.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "17d968aa06150de3f67e98f8ead12f8636c2624b78a08802ab1e8bd203a4a027" {
		t.Fatal("完整畫面來源雜湊不符")
	}
	var receipt struct {
		Assets map[string]string
		Cases  []struct {
			Name     string
			View     FleetFrameView
			Pictures []struct {
				Side       uint8
				SegmentHex string
				AllFrames  bool
			}
			Native string `json:"native_zlib_base64"`
		}
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Cases) != 2 {
		t.Fatal("完整原版幀數不符")
	}
	readAsset := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("../../org_game/extracted", name))
		if os.IsNotExist(err) {
			t.Skipf("缺原版%s", name)
		}
		if err != nil {
			t.Fatal(err)
		}
		if want, ok := receipt.Assets[name]; ok && fmt.Sprintf("%x", sha256.Sum256(b)) != want {
			t.Fatalf("原版%s來源不符", name)
		}
		return b
	}
	readAsset("GRAPH.IMG")
	fnt, err := font.Load("../../org_game/extracted/GRAPH.IMG")
	if err != nil {
		t.Fatal(err)
	}
	panel, err := asset.DecodeAMGImage(readAsset("ANI_PNL1.A98"))
	if err != nil {
		t.Fatal(err)
	}
	bmg, err := asset.OpenBMG(readAsset("COMANI.BMG"))
	if err != nil {
		t.Fatal(err)
	}
	sprites, err := asset.DecodeFleetSceneSprites(readAsset("3DBLOCK.SPR"), readAsset("COM_FLAG.SPR"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadFleetPresentationSpec("../../docs/internals/data/fleet-presentation-spec.json")
	if err != nil {
		t.Fatal(err)
	}
	var stats []map[string]any
	for _, c := range receipt.Cases {
		t.Run(c.Name, func(t *testing.T) {
			v := c.View
			for _, p := range c.Pictures {
				name, err := hex.DecodeString(p.SegmentHex)
				if err != nil {
					t.Fatal(err)
				}
				last := 0
				if p.AllFrames {
					frames, err := bmg.Frames(string(name))
					if err != nil {
						t.Fatal(err)
					}
					last = len(frames) - 1
				}
				v.Pictures = append(v.Pictures, FleetScenePicture{Side: p.Side, Segment: string(name), LastFrame: last})
			}
			before, _ := json.Marshal(v)
			fb, err := s.RenderFleetFrame(v, panel, sprites, bmg, fnt, nil)
			if err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(v)
			if !bytes.Equal(before, after) {
				t.Fatal("合成修改來源")
			}
			packed, err := base64.StdEncoding.DecodeString(c.Native)
			if err != nil {
				t.Fatal(err)
			}
			z, err := zlib.NewReader(bytes.NewReader(packed))
			if err != nil {
				t.Fatal(err)
			}
			want, err := io.ReadAll(z)
			if err != nil {
				t.Fatal(err)
			}
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			if len(want) != gfx.Width*gfx.Height {
				t.Fatal("原版完整畫布形狀不符")
			}
			diff := gfx.NewFramebuffer(0)
			different, maxChannel := 0, 0
			bbox := [4]int{640, 480, -1, -1}
			for i, p := range fb.Pix {
				if p == want[i] {
					continue
				}
				different++
				diff.Pix[i] = 15
				x, y := i%640, i/640
				bbox = [4]int{min(bbox[0], x), min(bbox[1], y), max(bbox[2], x), max(bbox[3], y)}
				r, g, b := s.Palette.RGB8(int(p))
				wr, wg, wb := s.Palette.RGB8(int(want[i]))
				for _, delta := range []int{int(r) - int(wr), int(g) - int(wg), int(b) - int(wb)} {
					maxChannel = max(maxChannel, delta, -delta)
				}
			}
			var regions []map[string]any
			for _, r := range []struct {
				Name       string
				X, Y, W, H int
			}{{"left-picture", 8, 56, 256, 102}, {"right-picture", 376, 56, 256, 102},
				{"left-text", 16, 160, 240, 40}, {"right-text", 384, 160, 240, 40}, {"middle-text", 272, 160, 96, 40},
				{"left-gauge", 298, 63, 11, 66}, {"right-gauge", 333, 63, 11, 66}, {"lower", 8, 208, 624, 224}} {
				n := 0
				for y := r.Y; y < r.Y+r.H; y++ {
					for x := r.X; x < r.X+r.W; x++ {
						i := y*640 + x
						if fb.Pix[i] != want[i] {
							n++
						}
					}
				}
				regions = append(regions, map[string]any{"name": r.Name, "rect": [4]int{r.X, r.Y, r.W, r.H}, "pixels": r.W * r.H, "same": r.W*r.H - n, "different": n})
			}
			var bounds any
			if different != 0 {
				bounds = bbox
			}
			stats = append(stats, map[string]any{"case": c.Name, "classification": "same-state complete typed presentation; no native pixels borrowed; diagnostic, not Game player path", "pixels": len(want), "same": len(want) - different, "different": different, "max_channel_difference": maxChannel, "difference_bbox": bounds, "regions": regions})
			if out := os.Getenv("LOGH3_FLEET_FRAME_OUTPUT"); out != "" {
				native := gfx.NewFramebuffer(0)
				copy(native.Pix, want)
				for name, image := range map[string]*gfx.Framebuffer{"original": native, "remake": fb, "diff": diff} {
					if err := image.WritePNG(filepath.Join(out, c.Name+"-"+name+".png"), s.Palette); err != nil {
						t.Fatal(err)
					}
				}
			}
			if different != 0 {
				t.Fatalf("全圖%d像素差，範圍%v", different, bbox)
			}
			// 以已驗完整來源探查錯誤邊界；不能回傳部分畫布或修改輸入。
			for _, mutate := range []func(*FleetFrameView){
				func(v *FleetFrameView) { v.Weapon.Side = 1 - v.Weapon.Side },
				func(v *FleetFrameView) { v.Integrity[0] = 101 },
				func(v *FleetFrameView) { v.Units[0].ShipsMax = 4001 },
				func(v *FleetFrameView) { v.Units[0].ShipsMax = v.Units[0].Ships - 1 },
				func(v *FleetFrameView) { v.Units[0].Visibility.Mode = 2 },
				func(v *FleetFrameView) { v.Scene.Sides[0][0].Ships = 4001 },
				func(v *FleetFrameView) { v.Pictures = []FleetScenePicture{{Side: 2, Segment: "PPD0"}} },
				func(v *FleetFrameView) { v.Pictures = []FleetScenePicture{{Segment: "NONE"}} },
				func(v *FleetFrameView) { v.Pictures = []FleetScenePicture{{Segment: "PPD0", LastFrame: 65535}} },
			} {
				bad := v
				for side := range bad.Scene.Sides {
					bad.Scene.Sides[side] = append([]FleetSceneUnit(nil), v.Scene.Sides[side]...)
				}
				mutate(&bad)
				before, _ := json.Marshal(bad)
				if fb, err := s.RenderFleetFrame(bad, panel, sprites, bmg, fnt, nil); err == nil || fb != nil {
					t.Fatal("非法來源未拒絕")
				}
				after, _ := json.Marshal(bad)
				if !bytes.Equal(before, after) {
					t.Fatal("錯誤修改來源")
				}
			}
		})
	}
	if out := os.Getenv("LOGH3_FLEET_FRAME_OUTPUT"); out != "" {
		b, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "comparison.json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// 099支援域的每種堆疊數、零艘標籤與所有可見Y；既有095原版裁切收據另行保持。
func TestFleetFrameSupportedClipDomain(t *testing.T) {
	blocks := &asset.Image{W: 48, H: 16, Pix: make([]uint8, 48*16)}
	labels := &asset.Image{W: 64, H: 9, Pix: make([]uint8, 64*9)}
	for y := 220; y <= 376; y++ {
		for n := 0; n <= 8; n++ {
			for _, x := range []int{8, 568} {
				if n != 0 {
					if err := validateFleetBlockOp(FleetBlockOp{Kind: FleetBlockSprite, X: x, Y: y, Count: n, Image: blocks}); err != nil {
						t.Fatal(err)
					}
				}
				for _, dy := range []int{0, 9} {
					if err := validateFleetBlockOp(FleetBlockOp{Kind: FleetBlockSprite, X: x - 6, Y: y - 4*(n-1) - 20 + dy, Count: 1, Image: labels}); err != nil {
						t.Fatalf("Y%d count%d: %v", y, n, err)
					}
				}
			}
		}
	}
}
