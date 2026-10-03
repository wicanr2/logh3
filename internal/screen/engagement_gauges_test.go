package screen

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func TestEngagementGaugesNativePixels(t *testing.T) {
	b, err := os.ReadFile("../../docs/internals/data/engagement-gauges.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "e1ed3aad4d7577f82df4449c097cee0461144ef0f5e555b669fac5c9a2914d86" {
		t.Fatal("凍結收據雜湊不符")
	}
	var receipt struct {
		Palette [256][3]uint8 `json:"base_palette"`
		Cases   []struct {
			Name    string `json:"case"`
			Kind    string `json:"kind"`
			Faction int    `json:"fleet_faction"`
			Value   int    `json:"value"`
			Divisor int    `json:"divisor"`
			Height  int    `json:"native_height"`
			Rect    [4]int `json:"rect"`
			Pixels  string `json:"native_pixels_hex"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Cases) != 32 {
		t.Fatal("收據案例數改變")
	}
	var stats []map[string]any
	for _, c := range receipt.Cases {
		t.Run(c.Name, func(t *testing.T) {
			g := EngagementGauge{Planet: c.Kind == "planet", FleetFaction: c.Faction, Value: c.Value, Divisor: c.Divisor}
			h, ok := g.Height()
			if !ok || h != c.Height {
				t.Fatalf("高度 %d，原版 %d", h, c.Height)
			}
			r, ok := g.Rect()
			if !ok || [4]int{r.X, r.Y, r.W, r.H} != c.Rect {
				t.Fatal("座標不符原版")
			}
			fb := gfx.NewFramebuffer(13)
			if !g.Draw(fb) {
				t.Fatal("未繪製")
			}
			want, err := hex.DecodeString(c.Pixels)
			if err != nil {
				t.Fatal(err)
			}
			var got []byte
			for y := r.Y; y < r.Y+r.H; y++ {
				got = append(got, fb.Pix[y*gfx.Width+r.X:y*gfx.Width+r.X+r.W]...)
			}
			if !bytes.Equal(got, want) {
				t.Fatal("量表色號不符原版")
			}
			for y := 0; y < gfx.Height; y++ {
				for x := 0; x < gfx.Width; x++ {
					if (x < r.X || x >= r.X+r.W || y < r.Y || y >= r.Y+r.H) && fb.At(x, y) != 13 {
						t.Fatalf("越界寫入 %d,%d", x, y)
					}
				}
			}
			stats = append(stats, map[string]any{"case": c.Name, "rect": c.Rect, "pixels": len(got), "same": len(got), "different": 0, "max_channel_difference": 0, "difference_bbox": nil})
			if out := os.Getenv("LOGH3_GAUGE_OUTPUT"); out != "" {
				for label, px := range map[string][]byte{"original": want, "remake": got, "diff": make([]byte, len(got))} {
					path := filepath.Join(out, c.Name+"-"+label+".png")
					pal := make(color.Palette, 256)
					for i, p := range receipt.Palette {
						pal[i] = color.RGBA{p[0], p[1], p[2], 255}
					}
					im := image.NewPaletted(image.Rect(0, 0, r.W, r.H), pal)
					copy(im.Pix, px)
					f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
					if e != nil {
						t.Fatal(e)
					}
					e = png.Encode(f, im)
					closeErr := f.Close()
					if e != nil {
						t.Fatal(e)
					}
					if closeErr != nil {
						t.Fatal(closeErr)
					}
				}
			}
		})
	}
	if out := os.Getenv("LOGH3_GAUGE_OUTPUT"); out != "" {
		b, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "comparison.json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEngagementGaugesRejectInvalidWithoutDrawing(t *testing.T) {
	for _, g := range []EngagementGauge{
		{FleetFaction: -1}, {FleetFaction: 2}, {Value: -1}, {Value: 101},
		{Planet: true, Value: 65536}, {Planet: true, Divisor: -1}, {Planet: true, Divisor: 65536},
	} {
		fb := gfx.NewFramebuffer(13)
		if g.Draw(fb) {
			t.Fatalf("接受非法量表 %+v", g)
		}
		if !bytes.Equal(fb.Pix, bytes.Repeat([]byte{13}, gfx.Width*gfx.Height)) {
			t.Fatal("非法資料動畫布")
		}
	}
	if h, ok := (EngagementGauge{Planet: true, Divisor: 0}).Height(); !ok || h != 66 {
		t.Fatal("原版分母0應滿格")
	}
}
