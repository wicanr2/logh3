package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// MainGameSpec 是遊戲主畫面的版面（量自 oracle，見
// `docs/internals/data/screen-main-game.json`）。
type MainGameSpec struct {
	Background uint8 `json:"background"`
	Map        struct {
		Source  string `json:"source"`
		Regions string `json:"regions"`
		X, Y    int
	} `json:"map"`
	Windows map[string]*GameWindow `json:"windows"`
	Mouse   *struct {
		X, Y int
	} `json:"mouse"`
}

// GameWindow 是主畫面上的一個視窗。
type GameWindow struct {
	X, Y, W, H int
	Frame      struct {
		Layers  [][4]uint8 `json:"layers"`
		Alt     [][4]uint8 `json:"alt"`
		Fill    uint8      `json:"fill"`
		Patches []struct {
			DX    int   `json:"dx"`
			DY    int   `json:"dy"`
			Color uint8 `json:"color"`
		} `json:"patches"`
	} `json:"frame"`
	Ink  uint8 `json:"ink"`
	Face *struct {
		Source string `json:"source"`
		Index  int    `json:"index"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	} `json:"face"`
	InputBox *Solid `json:"input_box"`
	Bar      *Solid `json:"bar"`

	// Boxes 是視窗裡自成一格的小框（輸入框、面板上的橫條）。
	Boxes []struct {
		X, Y, W, H int
		Layers     [][4]uint8 `json:"layers"`
		Alt        [][4]uint8 `json:"alt"`
		Fill       uint8      `json:"fill"`
		Patches    []struct {
			DX    int   `json:"dx"`
			DY    int   `json:"dy"`
			Color uint8 `json:"color"`
		} `json:"patches"`
	} `json:"boxes"`

	// Rules 是純色的線（報告視窗裡的表格分隔線）。
	Rules []Solid `json:"rules"`
	Lines []struct {
		Y     int `json:"y"`
		Ink   uint8 `json:"ink"`
		Cells []struct {
			X    int    `json:"x"`
			W    int    `json:"w"`
			Code uint16 `json:"code"`
		} `json:"cells"`
	} `json:"lines"`
}

// Solid 是一塊純色矩形（輸入框、面板上的橫條）。
type Solid struct {
	X, Y, W, H int
	Colour     uint8 `json:"colour"`
}

// LoadMainGameSpec 讀主畫面規格。
func LoadMainGameSpec(path string) (*MainGameSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s MainGameSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// bevel 把 JSON 的層描述轉成可畫的外框。
func (w *GameWindow) bevel() ui.Frame {
	f := ui.Frame{Fill: w.Frame.Fill}
	for _, l := range w.Frame.Layers {
		f.Layers = append(f.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
	}
	for _, a := range w.Frame.Alt {
		f.Alt = append(f.Alt, ui.Layer{Top: a[0], Right: a[1], Bottom: a[2], Left: a[3]})
	}
	for _, p := range w.Frame.Patches {
		f.Patches = append(f.Patches, ui.Patch{DX: p.DX, DY: p.DY, Color: p.Color})
	}
	return f
}

// Draw 畫這個視窗的外框與文字。shadow 是文字的立體邊色號。
func (w *GameWindow) Draw(fb *gfx.Framebuffer, fnt *font.Font, shadow uint8, faces *asset.AMG) []uint16 {
	w.bevel().Draw(fb, w.X, w.Y, w.W, w.H)
	for _, b := range w.Boxes {
		f := ui.Frame{Fill: b.Fill}
		for _, l := range b.Layers {
			f.Layers = append(f.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
		}
		for _, a := range b.Alt {
			f.Alt = append(f.Alt, ui.Layer{Top: a[0], Right: a[1], Bottom: a[2], Left: a[3]})
		}
		for _, p := range b.Patches {
			f.Patches = append(f.Patches, ui.Patch{DX: p.DX, DY: p.DY, Color: p.Color})
		}
		f.Draw(fb, b.X, b.Y, b.W, b.H)
	}
	for _, r := range w.Rules {
		fb.FillRect(r.X, r.Y, r.W, r.H, r.Colour)
	}
	if w.Face != nil && faces != nil {
		if img, err := faces.Decode(w.Face.Index); err == nil {
			for y := 0; y < img.H; y++ {
				for x := 0; x < img.W; x++ {
					fb.Set(w.Face.X+x, w.Face.Y+y, img.Pix[y*img.W+x])
				}
			}
		}
	}
	var missing []uint16
	for _, line := range w.Lines {
		ink := line.Ink
		if ink == 0 {
			ink = w.Ink
		}
		for _, c := range line.Cells {
			if c.W == font.HalfW {
				if !ui.DrawHalfGlyph(fb, fnt, c.X, line.Y, uint8(c.Code), ink, shadow, -1) {
					missing = append(missing, c.Code)
				}
				continue
			}
			if !ui.DrawGlyph(fb, fnt, c.X, line.Y, c.Code, ink, shadow, -1) {
				missing = append(missing, c.Code)
			}
		}
	}
	return missing
}

// MainGame 畫遊戲主畫面：背景 → 星域地圖 → 右側面板 → 會計報告。
//
// 畫的順序就是原版的層次：會計報告是後來彈出來的，蓋在面板與地圖上。
func MainGame(s *MainGameSpec, fnt *font.Font, mapImg *asset.Image, faces *asset.AMG,
	cur *ui.Cursor) (*gfx.Framebuffer, []uint16) {
	fb := gfx.NewFramebuffer(s.Background)
	var missing []uint16
	// 順序就是原版的層次：底板 → 星域地圖（蓋住底板左上角）→ 會計報告。
	if w := s.Windows["board"]; w != nil {
		missing = append(missing, w.Draw(fb, fnt, 3, faces)...)
	}
	if mapImg != nil {
		for y := 0; y < mapImg.H; y++ {
			for x := 0; x < mapImg.W; x++ {
				fb.Set(s.Map.X+x, s.Map.Y+y, mapImg.Pix[y*mapImg.W+x])
			}
		}
	}
	for _, name := range []string{"report"} {
		if w := s.Windows[name]; w != nil {
			missing = append(missing, w.Draw(fb, fnt, 3, faces)...)
		}
	}
	if cur != nil && s.Mouse != nil {
		cur.Draw(fb, s.Mouse.X, s.Mouse.Y)
	}
	return fb, missing
}
