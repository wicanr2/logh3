package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// CopyProtSpec 是防拷問答畫面的版面規格（量自 oracle，見
// `docs/internals/data/screen-copy-protection.json`）。
type CopyProtSpec struct {
	Background uint8 `json:"background"`
	Window     struct{ X, Y, W, H int } `json:"window"`
	Bevel struct {
		Layers [][4]uint8 `json:"layers"`
		Alt    [][4]uint8 `json:"alt"`
		Fill   uint8      `json:"fill"`
		Patches []struct {
			DX int `json:"dx"`
			DY int `json:"dy"`
			Color  uint8 `json:"color"`
		} `json:"patches"`
	} `json:"frame"`
	Face struct {
		Source string `json:"source"`
		Index  int    `json:"index"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	} `json:"face"`
	InputBox *struct {
		X, Y, W, H int
		Layers     [][4]uint8 `json:"layers"`
		Fill       uint8      `json:"fill"`
		Order      string     `json:"order"`
	} `json:"input_box"`
	Text struct {
		Ink    uint8 `json:"ink"`
		Shadow uint8 `json:"shadow"`
		Lines  []struct {
			Y     int `json:"y"`
			Cells []struct {
				X    int     `json:"x"`
				Code *uint16 `json:"code"`
			} `json:"cells"`
		} `json:"lines"`
	} `json:"text"`
	// InputText 是玩家打進去的數字要畫在哪（半形字型，沒有立體邊）。
	InputText *struct {
		X, Y  int
		Pitch int   `json:"pitch"`
		Ink   uint8 `json:"ink"`
	} `json:"input_text"`

	Palette6bit [][3]uint8 `json:"palette_6bit"`
}

// LoadCopyProtSpec 讀防拷問答畫面的規格。
func LoadCopyProtSpec(path string) (*CopyProtSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s CopyProtSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Palette 回規格裡的 16 色調色盤。
func (s *CopyProtSpec) Palette() gfx.Palette {
	var p gfx.Palette
	for i := 0; i < gfx.Colors && i < len(s.Palette6bit); i++ {
		p[i] = s.Palette6bit[i]
	}
	return p
}

// Frame 回這個視窗的外框。
func (s *CopyProtSpec) Frame() ui.Frame {
	f := ui.Frame{Fill: s.Bevel.Fill}
	for _, l := range s.Bevel.Layers {
		f.Layers = append(f.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
	}
	for _, a := range s.Bevel.Alt {
		f.Alt = append(f.Alt, ui.Layer{Top: a[0], Right: a[1], Bottom: a[2], Left: a[3]})
	}
	for _, p := range s.Bevel.Patches {
		f.Patches = append(f.Patches, ui.Patch{DX: p.DX, DY: p.DY, Color: p.Color})
	}
	return f
}

// CopyProtection 畫防拷問答畫面。face 是 `FACE.AMG` 解出來的頭像
// （nil 就不畫，用來單獨驗版面），typed 是玩家已經打進去的數字。
// 回傳畫不出來的字碼。
func CopyProtection(s *CopyProtSpec, fnt *font.Font, face *asset.Image, typed string) (*gfx.Framebuffer, []uint16) {
	fb := gfx.NewFramebuffer(s.Background)
	s.Frame().Draw(fb, s.Window.X, s.Window.Y, s.Window.W, s.Window.H)
	if face != nil {
		for y := 0; y < face.H; y++ {
			for x := 0; x < face.W; x++ {
				fb.Set(s.Face.X+x, s.Face.Y+y, face.Pix[y*face.W+x])
			}
		}
	}
	if b := s.InputBox; b != nil {
		f := ui.Frame{Fill: b.Fill, Order: b.Order}
		for _, l := range b.Layers {
			f.Layers = append(f.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
		}
		f.Draw(fb, b.X, b.Y, b.W, b.H)
	}
	var missing []uint16
	for _, line := range s.Text.Lines {
		for _, c := range line.Cells {
			if c.Code == nil {
				continue
			}
			if !ui.DrawGlyph(fb, fnt, c.X, line.Y, *c.Code, s.Text.Ink, s.Text.Shadow, -1) {
				missing = append(missing, *c.Code)
			}
		}
	}
	if t := s.InputText; t != nil {
		for i, c := range []byte(typed) {
			ui.DrawHalfGlyph(fb, fnt, t.X+i*t.Pitch, t.Y, c, t.Ink, 0, 0)
		}
	}
	return fb, missing
}

// FaceFrom 從自備的 `FACE.AMG` 取規格指定的那一張頭像。
func FaceFrom(data []byte, s *CopyProtSpec) (*asset.Image, error) {
	a, err := asset.OpenAMG(data)
	if err != nil {
		return nil, fmt.Errorf("讀 %s：%w", s.Face.Source, err)
	}
	return a.Decode(s.Face.Index)
}
