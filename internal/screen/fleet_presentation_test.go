package screen

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFleetPresentationNative(t *testing.T) {
	raw, err := os.ReadFile("../../docs/internals/data/fleet-presentation.json")
	if err != nil {
		t.Fatal(err)
	}
	// 由原版凍結器產生，後續收據改變須先審查其來源。
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "2dbff381163ce7d512af633e496bc5349dc2597210c0131f7f0d0874b0943851" {
		t.Fatal("凍結呈現收據雜湊不符")
	}
	var receipt struct {
		Base        string            `json:"base_zlib_base64"`
		Palette     [16][3]uint8      `json:"logical_palette_rgb8"`
		AssetInputs map[string]string `json:"asset_inputs"`
		Cases       []struct {
			Metadata struct {
				Name   string
				Params [5]int
			}
			Input struct {
				Side, Squadron, Detail, Mode, Weapon uint8
				PlayerSide                           uint8    `json:"player_side"`
				SideControl                          [2]uint8 `json:"side_control"`
				FleetNumber                          uint16   `json:"fleet_number"`
				AdmiralName                          string   `json:"admiral_name"`
				KindName                             string   `json:"kind_name"`
				Ships, Destroyed                     uint16
				ShipsMax                             uint16   `json:"ships_max"`
				WeaponTypes                          [3]uint8 `json:"weapon_types"`
			}
			Cells  string   `json:"native_cells_hex"`
			Cursor [2]uint8 `json:"native_cursor"`
			Colour uint8    `json:"native_colour"`
			Native string   `json:"native_zlib_base64"`
		}
	}
	if err = json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Cases) != 92 {
		t.Fatal("原版案例數不符")
	}
	s, err := LoadFleetPresentationSpec("../../docs/internals/data/fleet-presentation-spec.json")
	if err != nil {
		t.Fatal(err)
	}
	decode := func(b64 string) []byte {
		t.Helper()
		b, e := base64.StdEncoding.DecodeString(b64)
		if e != nil {
			t.Fatal(e)
		}
		z, e := zlib.NewReader(bytes.NewReader(b))
		if e != nil {
			t.Fatal(e)
		}
		v, e := io.ReadAll(z)
		if e != nil {
			t.Fatal(e)
		}
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		if len(v) != gfx.Width*gfx.Height {
			t.Fatal("原版畫布形狀錯誤")
		}
		return v
	}
	unhex := func(s string) []byte {
		b, e := hex.DecodeString(s)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	base := decode(receipt.Base)
	var fnt *font.Font
	fontPath := "../../org_game/extracted/GRAPH.IMG"
	if _, err = os.Stat(fontPath); err == nil {
		b, e := os.ReadFile(fontPath)
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != receipt.AssetInputs["GRAPH.IMG"] {
			t.Fatal("原版字型來源不符")
		}
		fnt, err = font.Load(fontPath)
		if err != nil {
			t.Fatal(err)
		}
	}
	var stats []map[string]any
	for _, c := range receipt.Cases {
		t.Run(c.Metadata.Name, func(t *testing.T) {
			v := c.Input
			fb := gfx.NewFramebuffer(0)
			copy(fb.Pix, base)
			var buffer *ui.TextBuffer
			var e error
			window := s.Windows[2]
			if c.Metadata.Params[3] < 0 {
				view := FleetUnitView{Side: v.Side, Squadron: v.Squadron, Detail: v.Detail, FleetNumber: v.FleetNumber, AdmiralName: unhex(v.AdmiralName), KindName: unhex(v.KindName), Ships: v.Ships, ShipsMax: v.ShipsMax, Visibility: FleetVisibility{Mode: v.Mode, PlayerSide: v.PlayerSide, SideControl: v.SideControl}}
				buffer, e = s.UnitBuffer(view)
				if e == nil && fnt != nil {
					e = s.DrawUnit(fb, fnt, nil, view)
				}
				window = s.Windows[v.Side]
			} else {
				view := FleetWeaponView{Side: v.Side, Weapon: v.Weapon, WeaponTypes: v.WeaponTypes, ShowResult: true, Destroyed: v.Destroyed}
				buffer, e = s.MiddleBuffer(view)
				if e == nil && fnt != nil {
					e = s.DrawMiddle(fb, fnt, nil, view)
				}
			}
			if e != nil {
				t.Fatal(e)
			}
			cells := make([]byte, len(buffer.Cells)*4)
			for i, cell := range buffer.Cells {
				binary.LittleEndian.PutUint16(cells[i*4:], cell.Code)
				cells[i*4+2] = cell.Colour
				cells[i*4+3] = cell.Flags
			}
			if !bytes.Equal(cells, unhex(c.Cells)) || [2]uint8{uint8(buffer.Col), uint8(buffer.Row)} != c.Cursor || buffer.Colour != c.Colour {
				t.Fatalf("內容格／游標／字色不同: %x 原版 %s，游標%d,%d 原版%v", cells, c.Cells, buffer.Col, buffer.Row, c.Cursor)
			}
			if fnt == nil {
				t.Skip("缺原版字型；內容格已核，逐像素驗證跳過")
			}
			want := decode(c.Native)
			different := 0
			var first string
			for i, p := range fb.Pix {
				if p != want[i] {
					different++
					if first == "" {
						first = fmt.Sprintf("%d,%d %d/%d", i%640, i/640, p, want[i])
					}
				}
			}
			for i, rgb := range receipt.Palette {
				r, g, b := s.Palette.RGB8(i)
				if [3]uint8{r, g, b} != rgb {
					t.Fatal("logical→DAC色盤不同")
				}
			}
			regionDifferent := 0
			for y := window.Y*20 + 40; y < window.Y*20+40+window.H*20; y++ {
				for x := window.X * 16; x < (window.X+window.W)*16; x++ {
					i := y*640 + x
					if fb.Pix[i] != want[i] {
						regionDifferent++
					}
				}
			}
			regionPixels := window.W * 16 * window.H * 20
			stats = append(stats, map[string]any{"case": c.Metadata.Name, "pixels": len(want), "same": len(want) - different, "different": different, "max_channel_difference": 0, "difference_bbox": nil, "region": [4]int{window.X * 16, window.Y*20 + 40, window.W * 16, window.H * 20}, "region_pixels": regionPixels, "region_same": regionPixels - regionDifferent, "region_different": regionDifferent, "classification": "same-state local presentation; other layers borrowed from native entry"})
			if out := os.Getenv("LOGH3_PRESENTATION_OUTPUT"); out != "" {
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
					path := filepath.Join(out, c.Metadata.Name+"-"+label)
					f, e := os.OpenFile(path+".png", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
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
					if e = os.WriteFile(path+".bin", p, 0644); e != nil {
						t.Fatal(e)
					}
				}
			}
			if different != 0 {
				t.Fatalf("%d像素不同，首差%s", different, first)
			}
		})
	}
	if out := os.Getenv("LOGH3_PRESENTATION_OUTPUT"); out != "" {
		b, e := json.MarshalIndent(stats, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, "comparison.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

func TestFleetPresentationRejectsInvalidInput(t *testing.T) {
	s, err := LoadFleetPresentationSpec("../../docs/internals/data/fleet-presentation-spec.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []FleetUnitView{{Side: 2}, {Visibility: FleetVisibility{Mode: 3}}, {Visibility: FleetVisibility{PlayerSide: 2}}, {KindName: []byte("ship")}} {
		fb := gfx.NewFramebuffer(7)
		if err = s.DrawUnit(fb, nil, nil, v); err == nil {
			t.Fatal("非法或缺名稱狀態被接受")
		}
		if !bytes.Equal(fb.Pix, bytes.Repeat([]byte{7}, 640*480)) {
			t.Fatal("錯誤仍改畫布")
		}
	}
	if _, err = s.MiddleBuffer(FleetWeaponView{Side: 2}); err == nil {
		t.Fatal("非法陣營被接受")
	}
	b, err := os.ReadFile("../../docs/internals/data/fleet-presentation-spec.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"windows", "palette", "weapon_names", "formats"} {
		var obj map[string]any
		if err = json.Unmarshal(b, &obj); err != nil {
			t.Fatal(err)
		}
		delete(obj, field)
		payload, e := json.Marshal(obj)
		if e != nil {
			t.Fatal(e)
		}
		p := filepath.Join(t.TempDir(), "broken.json")
		if e = os.WriteFile(p, payload, 0644); e != nil {
			t.Fatal(e)
		}
		if _, e = LoadFleetPresentationSpec(p); e == nil {
			t.Fatalf("缺%s仍載入", field)
		}
	}
}
