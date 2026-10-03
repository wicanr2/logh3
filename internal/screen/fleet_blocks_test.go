package screen

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type fleetBlockFixtureOp struct {
	Kind        string `json:"kind"`
	X, Y, Count int
	Sprite      string
	Args        []int
}

func TestFleetBlocksNativeScene(t *testing.T) {
	b, err := os.ReadFile("../../docs/internals/data/fleet-blocks.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "20874578ab46987b007f746efa0e180ff0906e5cc8c23ca3afdd0f1eadc9afeb" {
		t.Fatal("凍結收據不符")
	}
	var receipt struct {
		Palette [16][3]uint8 `json:"logical_palette_rgb8"`
		Sprites map[string]struct {
			Width  int `json:"width_bytes"`
			Height int
			Raw    string `json:"raw_hex"`
		}
		Frames []struct {
			Metadata struct {
				Name, Stage string
				Scene       int
			}
			Operations []fleetBlockFixtureOp
			Native     string `json:"native_zlib_base64"`
		}
		Edges []struct {
			Metadata  struct{ Name string }
			Operation fleetBlockFixtureOp
			Native    string `json:"native_zlib_base64"`
		}
	}
	if err = json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Frames) != 40 || len(receipt.Edges) != 20 {
		t.Fatal("收據數量改變")
	}
	sprites := map[string]*asset.Image{}
	for key, source := range receipt.Sprites {
		raw, err := hex.DecodeString(source.Raw)
		if err != nil {
			t.Fatal(err)
		}
		im, err := asset.DecodeFleetSceneSprite(raw, source.Width*8, source.Height)
		if err != nil {
			t.Fatal(err)
		}
		sprites[key] = im
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
		if err = r.Close(); err != nil {
			t.Fatal(err)
		}
		if len(p) != gfx.Width*gfx.Height {
			t.Fatal("原版畫面形狀不符")
		}
		return p
	}
	op := func(f fleetBlockFixtureOp) FleetBlockOp {
		switch f.Kind {
		case "background":
			return FleetBlockOp{Kind: FleetBlockBackground}
		case "sprite":
			return FleetBlockOp{Kind: FleetBlockSprite, X: f.X, Y: f.Y, Count: f.Count, Image: sprites[f.Sprite]}
		case "line":
			if len(f.Args) != 6 || f.Args[5] != 65535 {
				t.Fatal("線段遮罩不符")
			}
			return FleetBlockOp{Kind: FleetBlockLine, X: f.Args[0], Y: f.Args[1], X1: f.Args[2], Y1: f.Args[3], Ink: uint8(f.Args[4] & 15)}
		default:
			t.Fatalf("未知操作%q", f.Kind)
			return FleetBlockOp{}
		}
	}
	var stats []map[string]any
	check := func(name string, base, want []byte, ops []FleetBlockOp) {
		t.Helper()
		fb := gfx.NewFramebuffer(0)
		copy(fb.Pix, base)
		if err := DrawFleetBlocks(fb, ops); err != nil {
			t.Fatal(err)
		}
		different := 0
		var first string
		for i, p := range fb.Pix {
			if p != want[i] {
				different++
				if first == "" {
					first = fmt.Sprintf("%d,%d: %d/%d", i%640, i/640, p, want[i])
				}
			}
		}
		r := FleetBlockBackgroundRect()
		region := (r.X1 - r.X0 + 1) * (r.Y1 - r.Y0 + 1)
		stat := map[string]any{"case": name, "full_pixels": len(want), "same": len(want) - different, "different": different,
			"lower_rect": r, "lower_pixels": region, "difference_bbox": nil, "max_channel_difference": 0,
			"classification": "same-state compositor fixture; outside composition borrowed from native entry"}
		stats = append(stats, stat)
		if out := os.Getenv("LOGH3_BLOCKS_OUTPUT"); out != "" {
			diff := make([]byte, len(want))
			for i, p := range fb.Pix {
				if p != want[i] {
					diff[i] = 15
				}
			}
			pal := make(color.Palette, 16)
			for i, p := range receipt.Palette {
				pal[i] = color.RGBA{p[0], p[1], p[2], 255}
			}
			for label, p := range map[string][]byte{"original": want, "remake": fb.Pix, "diff": diff} {
				im := image.NewPaletted(image.Rect(0, 0, 640, 480), pal)
				copy(im.Pix, p)
				path := filepath.Join(out, name+"-"+label+".png")
				f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
				if e != nil {
					t.Fatal(e)
				}
				e = png.Encode(f, im)
				ce := f.Close()
				if e != nil {
					t.Fatal(e)
				}
				if ce != nil {
					t.Fatal(ce)
				}
				bp := filepath.Join(out, name+"-"+label+".bin")
				f, e = os.OpenFile(bp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
				if e != nil {
					t.Fatal(e)
				}
				_, e = f.Write(p)
				ce = f.Close()
				if e != nil {
					t.Fatal(e)
				}
				if ce != nil {
					t.Fatal(ce)
				}
			}
		}
		if different != 0 {
			t.Fatalf("%s %d像素差，首差%s", name, different, first)
		}
	}
	var base []byte
	var ops []FleetBlockOp
	for _, f := range receipt.Frames {
		want := decode(f.Native)
		if f.Metadata.Stage == "entry" {
			base = append([]byte(nil), want...)
			ops = nil
		}
		for _, v := range f.Operations {
			ops = append(ops, op(v))
		}
		t.Run(f.Metadata.Name, func(t *testing.T) { check(f.Metadata.Name, base, want, ops) })
	}
	base = decode(receipt.Frames[0].Native)
	for _, c := range receipt.Edges {
		t.Run(c.Metadata.Name, func(t *testing.T) {
			check(c.Metadata.Name, base, decode(c.Native), []FleetBlockOp{{Kind: FleetBlockBackground}, op(c.Operation)})
		})
	}
	if out := os.Getenv("LOGH3_BLOCKS_OUTPUT"); out != "" {
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.OpenFile(filepath.Join(out, "comparison.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Write(append(data, '\n'))
		ce := f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if ce != nil {
			t.Fatal(ce)
		}
	}
}

func TestFleetBlocksRejectUnresolvedStateBeforeDrawing(t *testing.T) {
	im := &asset.Image{W: 48, H: 16, Pix: make([]uint8, 48*16)}
	cases := []FleetBlockOp{
		{Kind: FleetBlockSprite, X: 245, Y: 376, Count: 131, Image: im},
		{Kind: FleetBlockSprite, X: 624, Y: 220, Count: 1, Image: im},
		{Kind: FleetBlockSprite, X: 8, Y: 220, Count: 1, Image: &asset.Image{W: 48, H: 16, Pix: []uint8{1}}},
		{Kind: FleetBlockLine, Ink: 16}, {Kind: 99},
	}
	for _, op := range cases {
		fb := gfx.NewFramebuffer(7)
		if err := DrawFleetBlocks(fb, []FleetBlockOp{{Kind: FleetBlockBackground}, op}); err == nil {
			t.Fatal("未拒絕未支援狀態")
		}
		if !bytes.Equal(fb.Pix, bytes.Repeat([]byte{7}, 640*480)) {
			t.Fatal("錯誤仍修改畫布")
		}
	}
	for _, raw := range [][]byte{nil, make([]byte, 383), make([]byte, 385)} {
		if _, err := asset.DecodeFleetSceneSprite(raw, 48, 16); err == nil {
			t.Fatal("非法素材被接受")
		}
	}
	if _, err := asset.DecodeFleetSceneSprites(make([]byte, 1536), make([]byte, 1441)); err == nil {
		t.Fatal("非法素材集被接受")
	}
}
