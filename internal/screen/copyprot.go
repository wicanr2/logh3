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

// CopyProtPITHz 與 CopyProtPITDivisor 是 OPEN.EXE 在防拷問答期間使用的
// PIT 設定。原版游標每 320 次中斷翻一次，完整週期 640 次（`docs/re/281`）。
const (
	CopyProtPITHz          = 1193182
	CopyProtPITDivisor     = 0x0800
	CopyProtCaretHalfTicks = 320
	CopyProtCaretPeriod    = CopyProtCaretHalfTicks * 2
)

// CopyProtSpec 是防拷問答畫面的版面規格（量自 oracle，見
// `docs/internals/data/screen-copy-protection.json`）。
type CopyProtSpec struct {
	Background uint8                    `json:"background"`
	Window     struct{ X, Y, W, H int } `json:"window"`
	Bevel      struct {
		Layers  [][4]uint8 `json:"layers"`
		Alt     [][4]uint8 `json:"alt"`
		Fill    uint8      `json:"fill"`
		Patches []struct {
			DX    int   `json:"dx"`
			DY    int   `json:"dy"`
			Color uint8 `json:"color"`
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
	// ButtonText 是輸入框右邊那顆按鈕上的字（「確認」）。
	ButtonText *struct {
		Y      int   `json:"y"`
		Ink    uint8 `json:"ink"`
		Shadow uint8 `json:"shadow"`
		Cells  []struct {
			X    int    `json:"x"`
			Code uint16 `json:"code"`
		} `json:"cells"`
	} `json:"button_text"`

	// InputText 是玩家打進去的數字要畫在哪。**全形數字**，沒有立體邊。
	InputText *struct {
		X, Y  int
		Pitch int   `json:"pitch"`
		Ink   uint8 `json:"ink"`
	} `json:"input_text"`

	// InputCaret 是輸入游標：**下一個**要打的那一格底下的一條 16×2 底線。
	// X 是還沒打字時的位置，每打一位往右一個 Pitch。
	InputCaret *struct {
		X, Y, W, H int
		Pitch      int   `json:"pitch"`
		Ink        uint8 `json:"ink"`
	} `json:"input_caret"`

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
	return CopyProtectionAt(s, fnt, face, typed, 0)
}

// CopyProtectionAt 畫指定 PIT 相位的防拷問答畫面。pitTicks 是進入問答後
// 累積的計時器中斷數；相位 0–319 畫游標，320–639 還原游標底下的底色。
// 以中斷數而不是牆上時間作輸入，讓 dosgolem 與 remake 可固定同一相位。
func CopyProtectionAt(s *CopyProtSpec, fnt *font.Font, face *asset.Image, typed string, pitTicks uint64) (*gfx.Framebuffer, []uint16) {
	fb := gfx.NewFramebuffer(s.Background)
	s.Frame().Draw(fb, s.Window.X, s.Window.Y, s.Window.W, s.Window.H)
	if face != nil {
		face.Draw(fb, s.Face.X, s.Face.Y, face.W, face.H)
	}
	if b := s.InputBox; b != nil {
		f := ui.Frame{Fill: b.Fill, Order: b.Order}
		for _, l := range b.Layers {
			f.Layers = append(f.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
		}
		f.Draw(fb, b.X, b.Y, b.W, b.H)
	}
	if b := s.ButtonText; b != nil {
		// 裁在按鈕內容區：立體邊往上移那一列會落到外框上，原版沒有。
		cx, cy, cw, ch := 0, 0, 640, 480
		if box := s.InputBox; box != nil {
			cx, cy, cw, ch = box.X+2, box.Y+2, box.W-4, box.H-4
		}
		for _, c := range b.Cells {
			ui.DrawGlyphClipped(fb, fnt, c.X, b.Y, c.Code, b.Ink, b.Shadow, -1, false,
				cx, cy, cw, ch)
		}
	}
	var missing []uint16
	for _, line := range s.Text.Lines {
		for _, c := range line.Cells {
			if c.Code == nil {
				continue
			}
			if !ui.DrawGlyph(fb, fnt, c.X, line.Y, *c.Code, s.Text.Ink, s.Text.Shadow, -1, false) {
				missing = append(missing, *c.Code)
			}
		}
	}
	digits := 0
	if t := s.InputText; t != nil {
		// 打進去的數字用**全形**字型（Big5 `A2AF` 起的 ０–９），
		// 不是半形——量出來的字比半形高也寬，而點陣在 GRAPH.IMG 的
		// 全形符號區裡（`docs/re/56`）。
		for i, c := range []byte(typed) {
			if c < '0' || c > '9' {
				continue
			}
			digits++
			code := fullDigit0 + uint16(c-'0')
			if !ui.DrawGlyph(fb, fnt, t.X+i*t.Pitch, t.Y, code, t.Ink, 0, 0, false) {
				missing = append(missing, code)
			}
		}
	}
	if c := s.InputCaret; c != nil && CaretVisible(pitTicks) {
		// 輸入游標畫在**下一個**要打的那一格底下（`docs/re/281`）。
		x0 := c.X + digits*c.Pitch
		for dy := 0; dy < c.H; dy++ {
			for dx := 0; dx < c.W; dx++ {
				fb.Set(x0+dx, c.Y+dy, c.Ink)
			}
		}
	}
	return fb, missing
}

// CaretVisible 回指定 PIT 相位是否應畫游標。
func CaretVisible(pitTicks uint64) bool {
	return pitTicks%CopyProtCaretPeriod < CopyProtCaretHalfTicks
}

// FaceFrom 從自備的 `FACE.AMG` 取規格指定的那一張頭像。
func FaceFrom(data []byte, s *CopyProtSpec) (*asset.Image, error) {
	a, err := asset.OpenAMG(data)
	if err != nil {
		return nil, fmt.Errorf("讀 %s：%w", s.Face.Source, err)
	}
	return a.Decode(s.Face.Index)
}
