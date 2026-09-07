package screen

import (
	"strings"
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// MainGameSpec 是遊戲主畫面的版面（量自 oracle，見
// `docs/internals/data/screen-main-game.json`）。
type MainGameSpec struct {
	Background uint8 `json:"background"`
	Map struct {
		Source  string `json:"source"`
		Regions string `json:"regions"`

		// Colours 是星域色號取哪一個欄位（`strmap-regions.json` 裡有好幾組：
		// 勢力色、治安色…）。空的話用勢力色。
		Colours string `json:"colours"`
		X, Y    int
	} `json:"map"`
	Windows map[string]*GameWindow `json:"windows"`

	// Fields 是「會變的數字」畫在哪（建造計劃的百分比、金額、餘額）。
	// 靜態的 Lines 已經含了開局的值，所以要重畫時先把這幾格填回底色。
	Fields *struct {
		Cost *NumField `json:"cost"`
		Rows []struct {
			Y        int       `json:"y"`
			Pct      *NumField `json:"pct"`
			Amount   *NumField `json:"amount"`
			Ships    *NumField `json:"ships"`
			Units    *NumField `json:"units"`
			Building *NumField `json:"building"`
		} `json:"rows"`
		Rest *struct {
			Y      int       `json:"y"`
			Pct    *NumField `json:"pct"`
			Amount *NumField `json:"amount"`
		} `json:"rest"`
	} `json:"fields"`

	// Order 是疊在地圖之上的視窗畫的順序（先畫的在下面）。
	// 空的話就只畫 report——大多數畫面只有一層。
	Order []string `json:"order"`

	// Hotspots 是可點的區域。原版存的是**相對視窗原點的格子**
	// （`docs/re/09`）；這裡放換算好的像素矩形，另外附上原始格子座標。
	Hotspots []struct {
		X, Y, W, H int
		Grid       []int  `json:"grid"`
		Action     string `json:"action"`
	} `json:"hotspots"`
	Mouse *struct {
		X, Y int

		// AfterBoxes 指定一個視窗：它的小框在游標之後再畫一次。
		AfterBoxes string `json:"after_boxes"`

		// AfterRules 同理，指定一個視窗的純色塊在游標之後再畫一次。
		AfterRules string `json:"after_rules"`

		// AfterBoxFills 指定一個視窗：它的小框**只有內部**在游標之後填回底色。
		AfterBoxFills string `json:"after_box_fills"`

		// Clip 是游標的可見矩形。原版的游標是**疊上去**的，視窗一重畫就
		// 把它蓋掉一部分，直到游標下次移動才補回來——所以同一隻游標在
		// 不同畫面上可見的範圍不一樣。
		Clip *struct{ X, Y, W, H int } `json:"clip"`
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

		// Bold ＝ 這一行用粗體畫。原版兩種粗細共用同一份字形，
		// 粗的那一種先把每一列往右加粗一格（`docs/internals/25`）。
		Bold  bool `json:"bold"`
		Cells []struct {
			X    int    `json:"x"`
			W    int    `json:"w"`
			Code uint16 `json:"code"`
		} `json:"cells"`
	} `json:"lines"`
}

// NumField 是一格會變的數字：靠右對齊，右邊界固定。
type NumField struct {
	X, W int
	Text string `json:"text"`
}

// Right 回這一格的右邊界。
func (f *NumField) Right() int { return f.X + f.W }

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
	w.DrawBoxes(fb)
	for _, r := range w.Rules {
		fb.FillRect(r.X, r.Y, r.W, r.H, r.Colour)
	}
	return w.drawContent(fb, fnt, shadow, faces)
}

// DrawBoxFills 把**游標壓到的**那幾個小框的內部填回底色，不重畫立體邊。
//
// 原版先畫游標、再刷游標底下那一列，而刷的只有列的內容區——
// 游標壓在列與列交界那條邊上的部分留了下來。整個小框重畫會把那幾格也擦掉。
//
// 只刷游標壓到的框，不是全部：同一個視窗裡還有別的小框（清單上方那排
// 標題格、格子裡的小標記），把它們一起填回底色會把標記整個抹掉。
func (w *GameWindow) DrawBoxFills(fb *gfx.Framebuffer, cx, cy, cw, ch int) {
	for _, b := range w.Boxes {
		n := len(b.Layers)
		if b.W <= 2*n || b.H <= 2*n {
			continue
		}
		if b.X >= cx+cw || cx >= b.X+b.W || b.Y >= cy+ch || cy >= b.Y+b.H {
			continue
		}
		fb.FillRect(b.X+n, b.Y+n, b.W-2*n, b.H-2*n, b.Fill)
	}
}

// DrawBoxes 只畫視窗裡的小框。拆出來是因為原版有些畫面**先畫游標、
// 再刷清單列**，所以那幾格會把游標蓋掉一部分（`docs/re/12`）。
func (w *GameWindow) DrawBoxes(fb *gfx.Framebuffer) {
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
}

func (w *GameWindow) drawContent(fb *gfx.Framebuffer, fnt *font.Font, shadow uint8,
	faces *asset.AMG) []uint16 {
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
				if !ui.DrawHalfGlyph(fb, fnt, c.X, line.Y, uint8(c.Code), ink, shadow, -1, line.Bold) {
					missing = append(missing, c.Code)
				}
				continue
			}
			if !ui.DrawGlyph(fb, fnt, c.X, line.Y, c.Code, ink, shadow, -1, line.Bold) {
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
	// 疊在地圖上的視窗照 Order 畫；沒寫 Order 就只有 report 一層。
	// 「待命中提督資料」那種畫面是三層：底板 → 建造計劃 → 清單。
	order := s.Order
	if len(order) == 0 {
		order = []string{"report"}
	}
	for _, name := range order {
		if w := s.Windows[name]; w != nil {
			missing = append(missing, w.Draw(fb, fnt, 3, faces)...)
		}
	}
	if cur != nil && s.Mouse != nil {
		if c := s.Mouse.Clip; c != nil {
			cur.DrawClipped(fb, s.Mouse.X, s.Mouse.Y, c.X, c.Y, c.W, c.H)
		} else {
			cur.Draw(fb, s.Mouse.X, s.Mouse.Y)
		}
		// 有些畫面在游標之後才刷清單列，那幾格會蓋掉游標的一部分。
		if n := s.Mouse.AfterBoxes; n != "" {
			if w := s.Windows[n]; w != nil {
				w.DrawBoxes(fb)
			}
		}
		if n := s.Mouse.AfterBoxFills; n != "" {
			if w := s.Windows[n]; w != nil {
				w.DrawBoxFills(fb, s.Mouse.X+cur.Offset[0], s.Mouse.Y+cur.Offset[1],
					cur.Width, cur.Height)
			}
		}
		// 反白條同理：滑鼠移到哪一列，原版就把那一列整條重畫成反白色，
		// 而那一下會把游標蓋掉——只留下沒被條子壓到的那幾列。
		if n := s.Mouse.AfterRules; n != "" {
			if w := s.Windows[n]; w != nil {
				for _, r := range w.Rules {
					fb.FillRect(r.X, r.Y, r.W, r.H, r.Colour)
				}
			}
		}
	}
	return fb, missing
}

// HitTest 回點在哪個熱區上的 action，沒點中回空字串。
func (s *MainGameSpec) HitTest(x, y int) string {
	for _, h := range s.Hotspots {
		if x >= h.X && x < h.X+h.W && y >= h.Y && y < h.Y+h.H {
			return h.Action
		}
	}
	return ""
}

// DrawNumbers 把會變的數字重畫上去：先把原本那幾格填回底色，
// 再靠右畫新的值。pct／amount 各七列，另加餘額那一列。
func (s *MainGameSpec) DrawNumbers(fb *gfx.Framebuffer, fnt *font.Font,
	pct, amount, ships, units []int, restPct, restAmount int) {
	f := s.Fields
	if f == nil {
		return
	}
	w := s.Windows["report"]
	if w == nil {
		return
	}
	fill := w.Frame.Fill
	put := func(nf *NumField, y int, text string) {
		if nf == nil {
			return
		}
		x := nf.Right() - len(text)*font.HalfW
		// 只清「舊值與新值聯集」那一塊。往左多清會吃掉左邊的欄位標題——
		// 那一帶是 3/7/9 的網點，不是底色，清了補不回來。
		// 立體邊往上一列、往左一位元，所以上緣要多留一列。
		clr := nf.X
		if x < clr {
			clr = x
		}
		fb.FillRect(clr, y-1, nf.Right()-clr, font.GlyphH+1, fill)
		for i := 0; i < len(text); i++ {
			ui.DrawHalfGlyph(fb, fnt, x+i*font.HalfW, y, text[i], w.Ink, 3, -1, false)
		}
	}
	for i, r := range f.Rows {
		if i < len(pct) {
			put(r.Pct, r.Y, itoa(pct[i])+"%")
		}
		if i < len(amount) {
			put(r.Amount, r.Y, itoa(amount[i]))
		}
		if i < len(ships) {
			put(r.Ships, r.Y, itoa(ships[i]))
		}
		if i < len(units) {
			put(r.Units, r.Y, itoa(units[i]))
		}
	}
	if f.Rest != nil {
		put(f.Rest.Pct, f.Rest.Y, itoa(restPct)+"%")
		put(f.Rest.Amount, f.Rest.Y, itoa(restAmount))
	}
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}

// InitialBuildPlan 從畫面規格讀出開局的預算分配：建造費與七列的百分比
// 都是規格裡量到的原版初值。
func InitialBuildPlan(s *MainGameSpec) *rules.BuildPlan {
	b := &rules.BuildPlan{}
	if s == nil || s.Fields == nil {
		return b
	}
	if s.Fields.Cost != nil {
		b.Cost = atoi(s.Fields.Cost.Text)
	}
	for i, r := range s.Fields.Rows {
		if i >= rules.BuildPlanRows || r.Pct == nil {
			break
		}
		b.Pct[i] = atoi(strings.TrimSuffix(r.Pct.Text, "%"))
	}
	return b
}

// DrawBuildPlan 把一份預算分配畫上去。
func DrawBuildPlan(s *MainGameSpec, fb *gfx.Framebuffer, fnt *font.Font, b *rules.BuildPlan) {
	n := rules.BuildPlanRows
	pct, amt, ships, units := make([]int, n), make([]int, n), make([]int, n), make([]int, n)
	for i := 0; i < n; i++ {
		pct[i] = b.Pct[i]
		amt[i] = b.Amount(i)
		ships[i] = b.Ships(i)
		units[i] = b.Units(i)
	}
	s.DrawNumbers(fb, fnt, pct, amt, ships, units, b.RestPct(), b.RestAmount())
}

func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			continue
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}
