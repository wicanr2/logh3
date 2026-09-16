package screen

import (
	"encoding/json"
	"os"
	"strings"

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
	Map        struct {
		Source  string `json:"source"`
		Regions string `json:"regions"`

		// Colours 是星域色號取哪一個欄位（`strmap-regions.json` 裡有好幾組：
		// 勢力色、治安色…）。空的話用勢力色。
		Colours string `json:"colours"`
		X, Y    int
	} `json:"map"`
	Windows map[string]*GameWindow `json:"windows"`

	// ListRows 是清單畫面要畫的列。**不進 JSON**——名字、軍階與幕僚記號
	// 都是從原版資料檔算出來的（`internal/rules`），不是規格裡抄好的。
	ListRows []ListRow `json:"-"`

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

	// MsgSpeed 是設定畫面上訊息速度那個數字畫在哪。原版用 `" %3d"`
	// **靠右**印（`docs/re/286`），所以右邊界固定、位數變了往左長。
	MsgSpeed *TextField `json:"msg_speed"`

	// Order 是疊在地圖之上的視窗畫的順序（先畫的在下面）。
	// 空的話就只畫 report——大多數畫面只有一層。
	Order []string `json:"order"`

	// Hotspots 是可點的區域。原版存的是**相對視窗原點的格子**
	// （`docs/re/09`）；這裡放換算好的像素矩形，另外附上原始格子座標。
	Hotspots []struct {
		X, Y, W, H int
		Grid       []int  `json:"grid"`
		Action     string `json:"action"`

		// Disabled ＝ 原版那一筆的可選旗標（`+4`）是 0，命中測試直接跳過
		// （`docs/internals/23`）。設定畫面的「ＭＩＤＩ」就是這種：出貨的
		// `MUSIC.INI` 沒有 MIDI 介面，那一格點不下去、框也不畫
		// （`docs/re/285`）。**零值是「可選」**——其餘畫面的規格裡沒有
		// 這個欄位，不能反過來寫成 `Enabled`。
		Disabled bool `json:"disabled"`
	} `json:"hotspots"`
	Mouse *struct {
		X, Y int

		// AfterBoxes 指定一個視窗：它的小框在游標之後再畫一次。
		AfterBoxes string `json:"after_boxes"`

		// AfterRules 同理，指定一個視窗的純色塊在游標之後再畫一次。
		AfterRules string `json:"after_rules"`

		// AfterBoxFills 指定一個視窗：它的小框**只有內部**在游標之後填回底色。
		AfterBoxFills string `json:"after_box_fills"`

		// Patches 是游標畫在這個位置時**留在底板上的零星像素**：
		// 游標被重畫的視窗蓋掉一部分之後剩下的那幾點。座標是絕對的，
		// 在游標之後畫（`docs/spec/001`）。
		Patches []struct {
			X, Y   int
			Colour uint8 `json:"colour"`
		} `json:"patches"`

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

		// StaffPost 有值時，`Index` 由遊戲換成那個職位的幕僚本人
		// （頭像的編號就是提督索引，`docs/internals/09`）。規格裡的
		// `Index` 是量收據那一張時的值，只當備援。
		StaffPost *int `json:"staff_post"`
	} `json:"face"`

	// FactionText 是「同一個位置、兩個陣營印不同字」的行：貨幣單位
	// （帝國「單位＝億帝國馬克」、同盟「單位＝億帝納」，`docs/re/64`）
	// 與對戰交棒視窗的標題（`docs/re/76`）。`variants` 依陣營索引。
	//
	// **兩邊不一定同色**：貨幣單位兩邊都是色 7，陣營名用的是陣營色
	// （帝國 6、同盟 5）。`inks` 沒給就兩邊都用 `ink`。
	FactionText []struct {
		X        int        `json:"x"`
		Y        int        `json:"y"`
		Ink      uint8      `json:"ink"`
		Inks     []uint8    `json:"inks"`
		Text     []string   `json:"text"`
		Variants [][]uint16 `json:"variants"`
	} `json:"faction_text"`

	// Staff 是「跟著幕僚走的署名」：規格只記位置與職位，
	// 姓名與軍階由遊戲填（`docs/spec/011`）。
	Staff *struct {
		Post int   `json:"post"`
		X    int   `json:"x"`
		Y    int   `json:"y"`
		Ink  uint8 `json:"ink"`
	} `json:"staff"`
	InputBox *Solid `json:"input_box"`
	Bar      *Solid `json:"bar"`

	// Select 是「這一組小框裡有一個是選中的，它的內部填成另一個顏色」
	// （艦隊組成的七個艦種列，`docs/re/67`）。選取條要畫在**框的字之前**
	// ——字印在條子上面，所以做法是換掉那個框的底色，不是另外畫一條。
	Select *struct {
		First, Count int
		Colour       uint8
	} `json:"select"`

	// selected ＝ 選中的是第幾個 ＋ 1；0 ＝ 都沒選（同 hoverRow 的理由）。
	selected int

	// Grid 是格盤：一格一張 `SPRITE.SPR` 的圖塊。圖塊由呼叫端畫
	// （這一層拿不到素材），這裡只記版面。
	//
	// **`Unit` 是基底不是圖塊編號**：有部隊的那一格畫 `Unit ＋ 艦種`
	// （原版 `ds:984E ＋ 艦種`）。兩個數都分陣營，`0FF4:0098` 依
	// `ds:8F1D` 選一組（`docs/internals/56`）。
	Grid *struct {
		X, Y       int
		Cols, Rows int
		Tile       int
		Source     string
		Empty      int
		Unit       int
		// 同盟那一組；零值代表這張規格沒分陣營，兩邊共用上面那組。
		EmptyAlliance int `json:"empty_alliance"`
		UnitAlliance  int `json:"unit_alliance"`
	} `json:"grid"`

	// Hover 是「滑鼠停在哪一列，那一列就整條刷成反白」的清單
	// （艦隊編號那 16 列、副官那 2 列，`docs/re/67`）。反白條畫在文字
	// **之前**——字是印在條子上面的。
	Hover *struct {
		X, Y, W, H int
		Pitch      int
		Rows       int
		Colour     uint8
	} `json:"hover"`

	// hoverRow ＝ 要刷的那一列 ＋ 1；0 ＝ 都不刷。用「＋1」是為了讓
	// JSON 解出來的零值就是「不刷」，不必另外記「有沒有設過」。
	hoverRow int

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

		// TextX／Codes 是框上的字（按鈕的標籤）。字裁在框的內容區裡：
		// 立體邊往上移那一列會落到框的白色上緣，原版的上緣是乾淨的。
		TextX  int      `json:"text_x"`
		TextY  int      `json:"text_y"`
		Text   string   `json:"text"`
		Codes  []uint16 `json:"codes"`
		Ink    uint8    `json:"ink"`
		Shadow uint8    `json:"shadow"`
	} `json:"boxes"`

	// Tabs 是視窗裡的分頁鈕（一覽表、情報收集那幾張的上方切換鈕）。
	Tabs *ui.TabBar `json:"tabs"`

	// List 是清單的列版面；內容由呼叫端算好放進 MainGameSpec.ListRows。
	List *ListSpec `json:"list"`

	// Rules 是純色的線（報告視窗裡的表格分隔線）。
	Rules []Solid `json:"rules"`
	Lines []struct {
		Y   int   `json:"y"`
		Ink uint8 `json:"ink"`

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

// SetLineCode 把第 line 行第 cell 格的字碼換掉。那一行會先複製出來，所以
// 在視窗的**複本**上呼叫不會動到規格本身（「待命中副官資料」拿「待命中提督資料」
// 那張換兩個字）。
func (w *GameWindow) SetLineCode(line, cell int, code uint16) {
	if line < 0 || line >= len(w.Lines) || cell < 0 || cell >= len(w.Lines[line].Cells) {
		return
	}
	lines := append(w.Lines[:0:0], w.Lines...)
	cells := append(lines[line].Cells[:0:0], lines[line].Cells...)
	cells[cell].Code = code
	lines[line].Cells = cells
	w.Lines = lines
}

// DropLineCellsFrom 把第 line 行 x ≥ from 的格子拿掉（同樣在複本上改）：
// 提督詳細資料那張量死的「：待命中」要換成算出來的狀態時用。
func (w *GameWindow) DropLineCellsFrom(line, from int) {
	if line < 0 || line >= len(w.Lines) {
		return
	}
	lines := append(w.Lines[:0:0], w.Lines...)
	cells := lines[line].Cells[:0:0]
	for _, c := range lines[line].Cells {
		if c.X < from {
			cells = append(cells, c)
		}
	}
	lines[line].Cells = cells
	w.Lines = lines
}

// copyBoxes 把 Boxes 複製一份再改，在視窗的**複本**上呼叫不會動到規格本身
// （與 SetLineCode 同一個理由）。
func (w *GameWindow) copyBoxes() {
	w.Boxes = append(w.Boxes[:0:0], w.Boxes...)
}

// MoveBox 把第 i 個小框的 x 挪到 x（框上的字跟著挪同樣的距離）。
func (w *GameWindow) MoveBox(i, x int) {
	if i < 0 || i >= len(w.Boxes) {
		return
	}
	w.copyBoxes()
	w.Boxes[i].TextX += x - w.Boxes[i].X
	w.Boxes[i].X = x
}

// SetBoxCodes 換掉第 i 個小框上的字。
func (w *GameWindow) SetBoxCodes(i int, codes []uint16) {
	if i < 0 || i >= len(w.Boxes) {
		return
	}
	w.copyBoxes()
	w.Boxes[i].Codes = append([]uint16(nil), codes...)
}

// DropBox 拿掉第 i 個小框。
func (w *GameWindow) DropBox(i int) {
	if i < 0 || i >= len(w.Boxes) {
		return
	}
	w.copyBoxes()
	w.Boxes = append(w.Boxes[:i], w.Boxes[i+1:]...)
}

// TextField 是視窗內容裡一塊會變的字，靠右對齊。與 NumField 的差別是
// 它帶自己的 `Y`——那幾塊在規格裡是 `lines` 的一行，不是 `fields` 的一列。
type TextField struct {
	X, Y, W int
}

// Right 回這一塊的右邊界。
func (f *TextField) Right() int { return f.X + f.W }

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
	w.DrawBoxes(fb, fnt)
	// Rules 畫在分頁鈕**之後**：規格裡的純色塊除了圖例色票，還兼收
	// 「壓在分頁鈕立體邊上的游標尾巴」那種最上層的補丁（`docs/re/53`）。
	// 順序顛倒的話分頁鈕的上緣會把那幾個像素蓋掉。
	w.Tabs.Draw(fb, fnt)
	for _, r := range w.Rules {
		fb.FillRect(r.X, r.Y, r.W, r.H, r.Colour)
	}
	if h := w.Hover; h != nil && w.hoverRow > 0 {
		fb.FillRect(h.X, h.Y+(w.hoverRow-1)*h.Pitch, h.W, h.H, h.Colour)
	}
	return w.drawContent(fb, fnt, shadow, faces)
}

// SetSelected 指定 `Select` 那一組小框裡選中的是第幾個；-1 ＝ 都沒選。
func (w *GameWindow) SetSelected(i int) {
	if w == nil || i < 0 {
		if w != nil {
			w.selected = 0
		}
		return
	}
	w.selected = i + 1
}

// SelectRowAt 回畫面座標 (x,y) 落在 `Select` 那一組小框的第幾個；
// 不在上面回 -1。
//
// 命中的範圍是**熱區**不是畫出來的框：熱區左右各比框窄 2 像素
// （`docs/re/68`）。
func (w *GameWindow) SelectRowAt(x, y int) int {
	sel := w.Select
	if sel == nil {
		return -1
	}
	for i := 0; i < sel.Count; i++ {
		k := sel.First + i
		if k < 0 || k >= len(w.Boxes) {
			continue
		}
		b := w.Boxes[k]
		if x >= b.X+2 && x < b.X+b.W-2 && y >= b.Y && y < b.Y+b.H {
			return i
		}
	}
	return -1
}

// BoxAt 回畫面座標 (x,y) 落在第幾個框上；不在任何一個上面回 -1。
//
// 命中的範圍與 `SelectRowAt` 同一條規則：熱區左右各比畫出來的框窄
// 2 像素（`docs/re/68`）。收據 `vs3` 點在 (400,262)——框裡、熱區外
// ——原版什麼都沒做。
func (w *GameWindow) BoxAt(x, y int) int {
	if w == nil {
		return -1
	}
	for i, b := range w.Boxes {
		if x >= b.X+2 && x < b.X+b.W-2 && y >= b.Y && y < b.Y+b.H {
			return i
		}
	}
	return -1
}

// SetHotspotFill 把 (x,y,wd,ht) 這個熱區對應的小框換成另一個填色，
// 回是否找到。
//
// 熱區與框的對應是量出來的：**框左右各比熱區多 2 像素、上下對齊**
// （`docs/re/68`）；上下箭頭那種元件則與熱區一樣大（`docs/re/285`）。
// 兩種都認，因為判準是「同一個 y 與高度、左右差 0 或 2」——設定畫面上
// 20 個框逐一對得上，沒有第二種解。
//
// 找不到就什麼都不做：可選旗標是 0 的那一筆原版連框都不畫
// （設定畫面的「ＭＩＤＩ」）。
func (w *GameWindow) SetHotspotFill(x, y, wd, ht int, fill uint8) bool {
	if w == nil {
		return false
	}
	for i := range w.Boxes {
		b := &w.Boxes[i]
		if b.Y != y || b.H != ht {
			continue
		}
		if (b.X == x-2 && b.W == wd+4) || (b.X == x && b.W == wd) {
			b.Fill = fill
			return true
		}
	}
	return false
}

// SetFieldText 把 `f` 那一塊的字換成 `text`（半形、靠右對齊），回是否換成了。
//
// 靜態的 `Lines` 已經含了原版開局的值（設定畫面上那個 `90`），所以要跟著
// 狀態變就得改那一行本身，不是另外畫一層蓋上去——疊上去會連立體邊一起疊。
// 位數變了就增減格數：多的截掉，少的複製既有那一格再補（`Cells` 是匿名
// 結構，複製既有元素是唯一不必替它取名的長法）。
func (w *GameWindow) SetFieldText(f *TextField, text string) bool {
	if w == nil || f == nil || text == "" {
		return false
	}
	for i := range w.Lines {
		ln := &w.Lines[i]
		if ln.Y != f.Y || len(ln.Cells) == 0 {
			continue
		}
		// 同一個 y 上還有那一行的中文標題，用「半形 ＋ 落在欄位裡」分開。
		if ln.Cells[0].W != font.HalfW ||
			ln.Cells[0].X < f.X || ln.Cells[0].X >= f.Right() {
			continue
		}
		for len(ln.Cells) < len(text) {
			ln.Cells = append(ln.Cells, ln.Cells[0])
		}
		ln.Cells = ln.Cells[:len(text)]
		for n := range ln.Cells {
			ln.Cells[n].X = f.Right() - (len(text)-n)*font.HalfW
			ln.Cells[n].W = font.HalfW
			ln.Cells[n].Code = uint16(text[n])
		}
		return true
	}
	return false
}

// SetHoverRow 指定這一次要把哪一列刷成反白；-1 ＝ 都不刷。
func (w *GameWindow) SetHoverRow(row int) {
	if w == nil || row < 0 {
		if w != nil {
			w.hoverRow = 0
		}
		return
	}
	w.hoverRow = row + 1
}

// HoverRowAt 回畫面座標 (x,y) 落在清單的第幾列；不在清單上回 -1。
func (w *GameWindow) HoverRowAt(x, y int) int {
	h := w.Hover
	if h == nil || h.Pitch <= 0 {
		return -1
	}
	if x < h.X || x >= h.X+h.W || y < h.Y {
		return -1
	}
	row := (y - h.Y) / h.Pitch
	if row < 0 || row >= h.Rows {
		return -1
	}
	return row
}

// DrawBoxFills 把**游標壓到的**那幾個小框的內部填回底色，不重畫立體邊。
//
// 原版先畫游標、再刷游標底下那一列，而刷的只有列的內容區——
// 游標壓在列與列交界那條邊上的部分留了下來。整個小框重畫會把那幾格也擦掉。
//
// 只刷游標壓到的框，不是全部：同一個視窗裡還有別的小框（清單上方那排
// 標題格、格子裡的小標記），把它們一起填回底色會把標記整個抹掉。
//
// **帶字的框是按鈕，不刷。** 刷回底色模型的是「原版先畫游標、再刷清單
// 那一列」；按鈕在游標之前就畫好了，游標壓上去只是蓋住一角，字還在
// （收據 `fleet1e`：游標停在「確認」上，那兩個字完整）。
func (w *GameWindow) DrawBoxFills(fb *gfx.Framebuffer, cx, cy, cw, ch int) {
	for _, b := range w.Boxes {
		if len(b.Codes) > 0 {
			continue
		}
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
func (w *GameWindow) DrawBoxes(fb *gfx.Framebuffer, fnt *font.Font) {
	for i, b := range w.Boxes {
		fill := b.Fill
		if sel := w.Select; sel != nil && w.selected > 0 &&
			i == sel.First+w.selected-1 && i < sel.First+sel.Count {
			fill = sel.Colour
		}
		f := ui.Frame{Fill: fill}
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
		// fnt ＝ nil 是「游標之後再刷一次」那一趟：只補外框，不重畫字。
		// 原版的游標畫在最上面，字重畫上去會把箭頭咬掉一塊。
		if fnt == nil || len(b.Codes) == 0 {
			continue
		}
		// 只裁上下：立體邊往上移那一列會落到框的白色上緣，原版的上緣是
		// 乾淨的；左右不裁——字的最左一行剛好落在內容區的邊界上。
		n := len(b.Layers)
		cx, cy := 0, b.Y+n
		cw, ch := gfx.Width, b.H-2*n
		ty := b.TextY
		if ty == 0 {
			ty = b.Y + n
		}
		for i, code := range b.Codes {
			ui.DrawCellClipped(fb, fnt, b.TextX+i*font.GlyphW, ty, code,
				b.Ink, b.Shadow, -1, false, cx, cy, cw, ch)
		}
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
	return MainGameWithBoard(s, fnt, mapImg, faces, cur, nil)
}

// MainGameWithBoard 同 MainGame，但在**底板與地圖之後、疊上去的視窗之前**
// 讓呼叫端插一手。從遊戲狀態算出來的那幾塊面板要畫在這個時機——畫太晚
// 會蓋掉疊在上面的視窗（`docs/spec/002`、`003`）。
func MainGameWithBoard(s *MainGameSpec, fnt *font.Font, mapImg *asset.Image, faces *asset.AMG,
	cur *ui.Cursor, afterBoard func(*gfx.Framebuffer)) (*gfx.Framebuffer, []uint16) {
	return MainGameLayered(s, fnt, mapImg, faces, cur, afterBoard, nil)
}

// MainGameLayered 再多一個掛勾：`afterWindow` 在**每一個疊層畫完之後**
// 被呼叫一次，帶著那一層的名字。
//
// 從狀態算出來的內容屬於某一層，得畫在那一層之後、**上面那一層之前**。
// 全部等到最後再畫的話，疊在上面的小視窗會被下面那層的數字戳穿——
// 艦隊編號那張清單右緣就被艦隊編成的能力值畫了一排洞。
func MainGameLayered(s *MainGameSpec, fnt *font.Font, mapImg *asset.Image, faces *asset.AMG,
	cur *ui.Cursor, afterBoard func(*gfx.Framebuffer),
	afterWindow func(string, *gfx.Framebuffer)) (*gfx.Framebuffer, []uint16) {
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
	if afterBoard != nil {
		afterBoard(fb)
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
		if afterWindow != nil {
			afterWindow(name, fb)
		}
	}
	for _, w := range s.Windows {
		if w.List != nil {
			w.List.Draw(fb, fnt, s.ListRows)
		}
	}
	if cur != nil && s.Mouse != nil {
		if c := s.Mouse.Clip; c != nil {
			cur.DrawClipped(fb, s.Mouse.X, s.Mouse.Y, c.X, c.Y, c.W, c.H)
		} else {
			cur.Draw(fb, s.Mouse.X, s.Mouse.Y)
		}
		for _, p := range s.Mouse.Patches {
			fb.Set(p.X, p.Y, p.Colour)
		}
		// 有些畫面在游標之後才刷清單列，那幾格會蓋掉游標的一部分。
		if n := s.Mouse.AfterBoxes; n != "" {
			if w := s.Windows[n]; w != nil {
				w.DrawBoxes(fb, nil)
				if w.List != nil {
					// 列的底板重畫過，字要跟著補回去，游標再蓋一次上去
					// ——原版的游標永遠在最上面。
					w.List.Draw(fb, fnt, s.ListRows)
					cur.Draw(fb, s.Mouse.X, s.Mouse.Y)
				}
			}
		}
		if n := s.Mouse.AfterBoxFills; n != "" {
			if w := s.Windows[n]; w != nil {
				cx, cy := s.Mouse.X+cur.Offset[0], s.Mouse.Y+cur.Offset[1]
				w.DrawBoxFills(fb, cx, cy, cur.Width, cur.Height)
				// 刷回底色的那一列要把字補回去，游標再蓋一次上去。
				if w.List != nil {
					w.List.DrawRowsIn(fb, fnt, s.ListRows, cx, cy, cur.Width, cur.Height)
					cur.Draw(fb, s.Mouse.X, s.Mouse.Y)
				}
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
//
// 可選旗標是 0 的那幾筆**當作不存在**（原版的命中測試就是跳過它們，
// `docs/internals/23`）：點下去與點在空白處一樣，不是「點到但沒反應」。
func (s *MainGameSpec) HitTest(x, y int) string {
	for _, h := range s.Hotspots {
		if h.Disabled {
			continue
		}
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

// DrawStaffLine 畫一行署名：姓名接軍階，全形字。
func DrawStaffLine(fb *gfx.Framebuffer, fnt *font.Font, x, y int, name, rank []uint16, ink uint8) {
	codes := append(append([]uint16(nil), name...), rank...)
	drawFull(fb, fnt, x, y, codes, ink)
}

// DrawFactionText 畫規格宣告的「兩個陣營印不同字」那幾行。
func (w *GameWindow) DrawFactionText(fb *gfx.Framebuffer, fnt *font.Font, faction int) {
	for _, t := range w.FactionText {
		if faction < 0 || faction >= len(t.Variants) {
			continue
		}
		ink := t.Ink
		if faction < len(t.Inks) {
			ink = t.Inks[faction]
		}
		drawFull(fb, fnt, t.X, t.Y, t.Variants[faction], ink)
	}
}

// DrawStarMap 把一張星圖貼在 (x,y)。目的地選取時地圖視窗被提到艦隊視窗
// 之上（`docs/spec/051`），所以要能在視窗之後再貼一次。
func DrawStarMap(fb *gfx.Framebuffer, img *asset.Image, x, y int) {
	if fb == nil || img == nil {
		return
	}
	for j := 0; j < img.H; j++ {
		for i := 0; i < img.W; i++ {
			fb.Set(x+i, y+j, img.Pix[j*img.W+i])
		}
	}
}
