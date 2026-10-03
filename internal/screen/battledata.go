package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰術選單第二項「資料」開出來的畫面（`docs/re/109`、`docs/re/111`、
// 規格 `docs/spec/021`）。
//
// 它蓋掉戰場那一塊（x 16–462、y 60–419），右邊的小地圖與狀態欄留著。
// 上排三個標籤切三張表：艦隊／行星／敵艦隊。
//
// **文字不是逐欄擺位的。** 原版把整行組成一條字串再印，全形進 16、
// 半形進 8，起點固定 x ＝ 48——三張表頭都是記憶體裡寫死的 48 位元組
// 字串，連空白都算在內（`docs/re/111`）。所以這裡也照字串排，
// 欄位的 x 是算出來的，不是規格裡抄的。

// BattleDataFrame 是一個框的層模型（與 `GameWindow.Frame` 同一套）。
type BattleDataFrame struct {
	Layers  [][4]uint8 `json:"layers"`
	Alt     [][4]uint8 `json:"alt"`
	Fill    uint8      `json:"fill"`
	Patches []struct {
		DX    int   `json:"dx"`
		DY    int   `json:"dy"`
		Color uint8 `json:"color"`
	} `json:"patches"`
}

// Frame 回這個框的 `ui.Frame`。
func (f BattleDataFrame) Frame() ui.Frame { return f.frame() }

func (f BattleDataFrame) frame() ui.Frame {
	out := ui.Frame{Fill: f.Fill}
	for _, l := range f.Layers {
		out.Layers = append(out.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
	}
	for _, a := range f.Alt {
		out.Alt = append(out.Alt, ui.Layer{Top: a[0], Right: a[1], Bottom: a[2], Left: a[3]})
	}
	for _, p := range f.Patches {
		out.Patches = append(out.Patches, ui.Patch{DX: p.DX, DY: p.DY, Color: p.Color})
	}
	return out
}

// BattleDataCell 是一格字：`w` ＝ 8 是半形、16 是全形。
type BattleDataCell struct {
	X    int    `json:"x"`
	W    int    `json:"w"`
	Code uint16 `json:"code"`
}

// BattleDataLine 是一整行字。
type BattleDataLine struct {
	Text  string           `json:"text"`
	Cells []BattleDataCell `json:"cells"`
}

// BattleDataSpec 是這張畫面的版面。
type BattleDataSpec struct {
	Window struct {
		X, Y, W, H int
		Frame      BattleDataFrame `json:"frame"`
	} `json:"window"`
	Tab struct {
		Y, W, H int
		Xs      []int           `json:"xs"`
		Frame   BattleDataFrame `json:"frame"`
		Mark    struct {
			DX, DY, W, H int
			Frame        BattleDataFrame `json:"frame"`
			On           uint8           `json:"on"`
		} `json:"mark"`
		Labels []BattleDataLine `json:"labels"`
	} `json:"tab"`
	Row struct {
		X, Y, W, H int
		Count      int             `json:"count"`
		Frame      BattleDataFrame `json:"frame"`
	} `json:"row"`
	// Scroll 是表右邊那一欄：**三塊**，上下各 20（與表列同高）、中間 240。
	Scroll struct {
		X, W  int
		Boxes []struct{ Y, H int } `json:"boxes"`
		Frame BattleDataFrame      `json:"frame"`
	} `json:"scroll"`
	Text struct {
		X       int   `json:"x"`
		HeaderY int   `json:"header_y"`
		RowDY   int   `json:"row_dy"`
		Ink     uint8 `json:"ink"`
		LeadInk uint8 `json:"lead_ink"`
		Shadow  uint8 `json:"shadow"`
	} `json:"text"`
	Headers []BattleDataLine `json:"headers"`
}

// LoadBattleDataSpec 讀 `battle-data.json`。
func LoadBattleDataSpec(path string) (*BattleDataSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s BattleDataSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(s.Tab.Xs) != 3 || len(s.Headers) != 3 {
		return nil, fmt.Errorf("%s: 要三個標籤與三條表頭", path)
	}
	return &s, nil
}

// BattleDataRun 是一行裡的一段字：同一段共用顏色。
//
// 原版的字串裡夾著 `\x01[a,b,` 的屬性跳脫，所以一行可以換色——
// 最左邊那一欄是色 6、其餘色 7（`docs/re/111`）。
type BattleDataRun struct {
	// Codes 是一段字的字碼：**小於 0x100 當半形（進 8）、其餘當全形
	// （進 16）**。Big5 的前導位元組一定 ≥ 0xA1，所以這條規則不會誤判。
	Codes []uint16
	Ink   uint8
	X     int // 0 ＝ 接著上一段
}

// BattleDataRowAt 回畫面座標落在第幾列；不在表上回 -1。
func (s *BattleDataSpec) BattleDataRowAt(x, y int) int {
	if x < s.Row.X || x >= s.Row.X+s.Row.W {
		return -1
	}
	r := (y - s.Row.Y) / s.Row.H
	if y < s.Row.Y || r >= s.Row.Count {
		return -1
	}
	return r
}

// BattleDataTabAt 回畫面座標落在第幾個標籤；不在上面回 -1。
func (s *BattleDataSpec) BattleDataTabAt(x, y int) int {
	if y < s.Tab.Y || y >= s.Tab.Y+s.Tab.H {
		return -1
	}
	for i, tx := range s.Tab.Xs {
		if x >= tx && x < tx+s.Tab.W {
			return i
		}
	}
	return -1
}

// BattleDataScrollAt 回捲軸按鈕方向：上是 -1、下是 1；中間槽或外面是 0。
// 原版熱區是 Scroll.Boxes 的第一與最後一格，中間長槽不接受點擊。
func (s *BattleDataSpec) BattleDataScrollAt(x, y int) int {
	// 畫框是 x=434..447；原版熱區表用一個 16 像素格，從 x=432 開始。
	// 左邊多出的 2 像素是刻意保留的互動邊界（`docs/re/303`）。
	if x < s.Scroll.X-2 || x >= s.Scroll.X+s.Scroll.W || len(s.Scroll.Boxes) < 2 {
		return 0
	}
	for _, b := range []struct {
		y, h, dir int
	}{
		{s.Scroll.Boxes[0].Y, s.Scroll.Boxes[0].H, -1},
		{s.Scroll.Boxes[len(s.Scroll.Boxes)-1].Y, s.Scroll.Boxes[len(s.Scroll.Boxes)-1].H, 1},
	} {
		if y >= b.y && y < b.y+b.h {
			return b.dir
		}
	}
	return 0
}

// DrawBattleData 畫這張畫面：外框 → 三個標籤 → 十四列 ＋ 捲軸槽 →
// 表頭 → 每一列的字。回畫不出來的字碼。
//
// rows 的長度不必等於 Row.Count；多的不畫，少的留空列。
func DrawBattleData(s *BattleDataSpec, fb *gfx.Framebuffer, fnt *font.Font,
	tab int, rows [][]BattleDataRun) []uint16 {
	if s == nil || fb == nil {
		return nil
	}
	var missing []uint16
	s.Window.Frame.frame().Draw(fb, s.Window.X, s.Window.Y, s.Window.W, s.Window.H)
	// 字元窗右側陰影另占一列；舊樣本的底圖恰好也是3（研究395）。
	fb.FillRect(s.Window.X+s.Window.W, s.Window.Y, 1, s.Window.H, 3)

	tf, mf := s.Tab.Frame.frame(), s.Tab.Mark.Frame.frame()
	for i, tx := range s.Tab.Xs {
		tf.Draw(fb, tx, s.Tab.Y, s.Tab.W, s.Tab.H)
		f := mf
		if i == tab {
			f.Fill = s.Tab.Mark.On
		}
		f.Draw(fb, tx+s.Tab.Mark.DX, s.Tab.Y+s.Tab.Mark.DY, s.Tab.Mark.W, s.Tab.Mark.H)
		if i < len(s.Tab.Labels) {
			// 只裁上下：立體邊往上移那一列會落到框的亮色上緣，
			// 原版的上緣是乾淨的（同 `GameWindow.DrawBoxes`）。
			n := len(s.Tab.Frame.Layers)
			missing = append(missing, drawCells(fb, fnt, s.Tab.Labels[i].Cells,
				s.Tab.Y+2, s.Text.Ink, s.Text.Shadow, false,
				s.Tab.Y+n, s.Tab.H-2*n)...)
		}
	}

	rf := s.Row.Frame.frame()
	for r := 0; r < s.Row.Count; r++ {
		rf.Draw(fb, s.Row.X, s.Row.Y+r*s.Row.H, s.Row.W, s.Row.H)
	}
	sf := s.Scroll.Frame.frame()
	for _, b := range s.Scroll.Boxes {
		sf.Draw(fb, s.Scroll.X, b.Y, s.Scroll.W, b.H)
	}

	if tab >= 0 && tab < len(s.Headers) {
		missing = append(missing, drawCells(fb, fnt, s.Headers[tab].Cells,
			s.Text.HeaderY, s.Text.Ink, s.Text.Shadow, true, 0, 0)...)
	}
	for r, runs := range rows {
		if r >= s.Row.Count {
			break
		}
		y := s.Row.Y + r*s.Row.H + s.Text.RowDY
		x := s.Text.X
		for _, run := range runs {
			if run.X > 0 {
				x = run.X
			}
			ink := run.Ink
			if ink == 0 {
				ink = s.Text.Ink
			}
			cs, next := layoutCells(run.Codes, x)
			rn := len(s.Row.Frame.Layers)
			missing = append(missing, drawCells(fb, fnt, cs, y, ink, s.Text.Shadow, false,
				s.Row.Y+r*s.Row.H+rn, s.Row.H-2*rn)...)
			x = next
		}
	}
	return missing
}

// DrawBattleDataScrollArrows依原版FF11／FF15字碼畫圖塊，判斷完整清單長度。
func DrawBattleDataScrollArrows(s *BattleDataSpec, fb *gfx.Framebuffer, tiles *ui.FrameTiles, total int) {
	if s == nil || fb == nil || tiles == nil || total <= s.Row.Count || len(s.Scroll.Boxes) < 2 {
		return
	}
	tiles.DrawTile(fb, s.Scroll.X-2, s.Scroll.Boxes[0].Y, 16, 9)
	tiles.DrawTile(fb, s.Scroll.X-2, s.Scroll.Boxes[len(s.Scroll.Boxes)-1].Y, 20, 9)
}

// layoutCells 把一串字碼攤成格。
//
// **一格永遠 16 像素**：字碼 ≥ 0x100 是一個全形字，其餘是兩個半形字
// 併成一格（`docs/internals/31`）。併格不只是省事——立體邊與加粗都是
// 整格一起做的，拆成兩格畫，右半那個字的立體邊會蓋掉左半那個字的
// 最後一行（`docs/re/111`）。
func layoutCells(codes []uint16, x int) ([]BattleDataCell, int) {
	out := make([]BattleDataCell, 0, len(codes))
	for i := 0; i < len(codes); {
		c := codes[i]
		if c >= 0x100 {
			out = append(out, BattleDataCell{X: x, W: 16, Code: c})
			x += 16
			i++
			continue
		}
		lo := uint16(' ')
		if i+1 < len(codes) && codes[i+1] < 0x100 {
			lo = codes[i+1]
			i++
		}
		out = append(out, BattleDataCell{X: x, W: 16, Code: c<<8 | lo})
		x += 16
		i++
	}
	return out, x
}

// drawCells 畫一行字。cy／ch 是上下的裁切範圍，ch ＝ 0 ＝ 不裁。
func drawCells(fb *gfx.Framebuffer, fnt *font.Font, cells []BattleDataCell,
	y int, ink, shadow uint8, bold bool, cy, ch int) []uint16 {
	if ch == 0 {
		cy, ch = 0, gfx.Height
	}
	var missing []uint16
	for _, c := range cells {
		if c.Code == ' ' || c.Code == 0x2020 {
			continue
		}
		if c.W == font.HalfW {
			if !ui.DrawHalfGlyphClipped(fb, fnt, c.X, y, uint8(c.Code), ink, shadow, -1, bold,
				0, cy, gfx.Width, ch) {
				missing = append(missing, c.Code)
			}
			continue
		}
		// 一格 16 像素：字碼高位元組 ≥ 0x80 是全形，否則是兩個半形
		// 塞在同一格（`ui.DrawCellClipped` 自己分路）。
		if !ui.DrawCellClipped(fb, fnt, c.X, y, c.Code, ink, shadow, -1, bold,
			0, cy, gfx.Width, ch) {
			missing = append(missing, c.Code)
		}
	}
	return missing
}
