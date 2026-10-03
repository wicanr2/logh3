package screen

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// 上方保留原版入口；下方全部從typed來源生成，沒有借用原版投影或操作參數。
func TestFleetSceneTypedNative(t *testing.T) {
	data, err := os.ReadFile("../../docs/internals/data/fleet-scene-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "4988a5b95d2167bc14166a8fb173fb26bc0b6d4b9f6e7f070c86c9bcc3b126f4" {
		t.Fatal("原版場景收據雜湊不符")
	}
	var receipt struct {
		Palette [16][3]uint8
		Sprites map[string]struct {
			Width  int `json:"width_bytes"`
			Height int
			Raw    string `json:"raw_hex"`
		}
		Cases []struct {
			Name       string
			Sides      [2][]FleetSceneUnit
			Facing     [2]uint8
			AttackerID int `json:"attacker_id"`
			TargetID   int `json:"target_id"`
			Visibility struct {
				Mode        uint8
				PlayerSide  uint8    `json:"player_side"`
				SideControl [2]uint8 `json:"side_control"`
			}
			Operations []fleetBlockFixtureOp
			Entry      string `json:"entry_native_zlib_base64"`
			Complete   string `json:"complete_native_zlib_base64"`
		}
	}
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Cases) != 2 {
		t.Fatal("原版場景數不符")
	}
	sprites := &asset.FleetSceneSprites{}
	keys := map[*asset.Image]string{}
	for key, s := range receipt.Sprites {
		raw, err := hex.DecodeString(s.Raw)
		if err != nil {
			t.Fatal(err)
		}
		im, err := asset.DecodeFleetSceneSprite(raw, s.Width*8, s.Height)
		if err != nil {
			t.Fatal(err)
		}
		keys[im] = key
		var offset int
		if _, err := fmt.Sscanf(key, "3D22:%X:", &offset); err != nil {
			t.Fatal(err)
		}
		if s.Width == 6 {
			sprites.Blocks[(offset-0x551e)/384] = im
		} else {
			sprites.Labels[(offset-0x5b1e)/288] = im
		}
	}
	decode := func(s string) []byte {
		raw, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		r, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		p, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		if len(p) != gfx.Width*gfx.Height {
			t.Fatal("畫面形狀不符")
		}
		return p
	}
	var stats []map[string]any
	for _, c := range receipt.Cases {
		t.Run(c.Name, func(t *testing.T) {
			view := FleetSceneView{Sides: c.Sides, AttackerID: c.AttackerID, TargetID: c.TargetID, Facing: c.Facing,
				Visibility: FleetVisibility{Mode: c.Visibility.Mode, PlayerSide: c.Visibility.PlayerSide, SideControl: c.Visibility.SideControl}}
			ops, err := BuildFleetSceneOps(view, sprites)
			if err != nil {
				t.Fatal(err)
			}
			if len(ops) != len(c.Operations) {
				t.Fatalf("操作數%d want%d", len(ops), len(c.Operations))
			}
			for i, o := range ops {
				want := c.Operations[i]
				switch o.Kind {
				case FleetBlockBackground:
					if want.Kind != "background" {
						t.Fatal("背景順序不符")
					}
				case FleetBlockSprite:
					if want.Kind != "sprite" || o.X != want.X || o.Y != want.Y || o.Count != want.Count || keys[o.Image] != want.Sprite {
						t.Fatalf("操作%d got%+v sprite%s want%+v", i, o, keys[o.Image], want)
					}
				case FleetBlockLine:
					if want.Kind != "line" || len(want.Args) != 6 || o.X != want.Args[0] || o.Y != want.Args[1] || o.X1 != want.Args[2] || o.Y1 != want.Args[3] || int(o.Ink) != want.Args[4]&15 {
						t.Fatalf("線段%d got%+v want%+v", i, o, want)
					}
				}
			}
			fb := gfx.NewFramebuffer(0)
			copy(fb.Pix, decode(c.Entry))
			want := decode(c.Complete)
			if err := DrawFleetBlocks(fb, ops); err != nil {
				t.Fatal(err)
			}
			different, lowerDifference := 0, 0
			diff := make([]byte, len(want))
			r := FleetBlockBackgroundRect()
			for i, p := range fb.Pix {
				if p != want[i] {
					different++
					diff[i] = 15
					if r.Inside(i%640, i/640) {
						lowerDifference++
					}
				}
			}
			stats = append(stats, map[string]any{"case": c.Name, "classification": "same-state typed lower-scene; native above retained; not Game player path", "full_pixels": len(want), "same": len(want) - different, "different": different, "lower_rect": r, "lower_pixels": (r.X1 - r.X0 + 1) * (r.Y1 - r.Y0 + 1), "lower_different": lowerDifference, "difference_bbox": nil, "max_channel_difference": 0})
			if out := os.Getenv("LOGH3_FLEET_SCENE_OUTPUT"); out != "" {
				pal := make(color.Palette, 16)
				for i, p := range receipt.Palette {
					pal[i] = color.RGBA{p[0], p[1], p[2], 255}
				}
				for label, p := range map[string][]byte{"original": want, "remake": fb.Pix, "diff": diff} {
					im := image.NewPaletted(image.Rect(0, 0, 640, 480), pal)
					copy(im.Pix, p)
					f, err := os.OpenFile(filepath.Join(out, c.Name+"-"+label+".png"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
					if err != nil {
						t.Fatal(err)
					}
					if err := png.Encode(f, im); err != nil {
						t.Fatal(err)
					}
					if err := f.Close(); err != nil {
						t.Fatal(err)
					}
				}
			}
			if different != 0 {
				t.Fatalf("全圖%d／下方%d像素差", different, lowerDifference)
			}
		})
	}
	if out := os.Getenv("LOGH3_FLEET_SCENE_OUTPUT"); out != "" {
		b, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "comparison.json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFleetSceneRejectsUnresolvedClip(t *testing.T) {
	sprites := &asset.FleetSceneSprites{}
	for i := range sprites.Blocks {
		sprites.Blocks[i] = &asset.Image{W: 48, H: 16, Pix: make([]uint8, 48*16)}
	}
	for i := range sprites.Labels {
		sprites.Labels[i] = &asset.Image{W: 64, H: 9, Pix: make([]uint8, 64*9)}
	}
	view := FleetSceneView{AttackerID: 1, TargetID: 3, Sides: [2][]FleetSceneUnit{
		{{FleetProjectionUnit: rules.FleetProjectionUnit{ID: 1, X: 1000, Y: 1000}, Ships: 3000},
			{FleetProjectionUnit: rules.FleetProjectionUnit{ID: 2, X: 1000, Y: 1104}, Ships: 65001}},
		{{FleetProjectionUnit: rules.FleetProjectionUnit{ID: 3, Side: 1, X: 1020, Y: 1000}, Ships: 3000}},
	}}
	// +104投影為Y376。65001艘產生131塊，碰到095未支援的零高度裁切。
	before, _ := json.Marshal(view)
	ops, err := BuildFleetSceneOps(view, sprites)
	if err == nil || ops != nil {
		t.Fatal("未閉合裁切不得回傳部分場景")
	}
	after, _ := json.Marshal(view)
	if string(before) != string(after) {
		t.Fatal("錯誤修改來源")
	}
}
