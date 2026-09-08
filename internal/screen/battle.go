package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// BattleSpec 是戰場畫面的版面（`docs/internals/data/screen-battle.json`）。
type BattleSpec struct {
	Panel struct {
		File string `json:"file"`
		Y    int    `json:"y"`
	} `json:"panel"`
	Field struct {
		X, Y          int    `json:"-"`
		Cols, Rows    int    `json:"-"`
		TileW, TileH  int    `json:"-"`
		Tiles         string `json:"tiles"`
		MapCols       int    `json:"map_cols"`
		MapRows       int    `json:"map_rows"`
		XRaw          int    `json:"x"`
		YRaw          int    `json:"y"`
		ColsRaw       int    `json:"cols"`
		RowsRaw       int    `json:"rows"`
		TileWRaw      int    `json:"tile_w"`
		TileHRaw      int    `json:"tile_h"`
	} `json:"field"`
	Status  StatusSpec  `json:"status"`
	Minimap MinimapSpec `json:"minimap"`

	Grid struct {
		Pitch  int   `json:"pitch"`
		DotX   int   `json:"dot_x"`
		DotY   int   `json:"dot_y"`
		Ink    uint8 `json:"ink"`
		Shadow uint8 `json:"shadow"`
	} `json:"grid"`
}

// LoadBattleSpec 讀戰場畫面的版面。
func LoadBattleSpec(path string) (*BattleSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s BattleSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	s.Field.X, s.Field.Y = s.Field.XRaw, s.Field.YRaw
	s.Field.Cols, s.Field.Rows = s.Field.ColsRaw, s.Field.RowsRaw
	s.Field.TileW, s.Field.TileH = s.Field.TileWRaw, s.Field.TileHRaw
	return &s, nil
}

// BattleView 是視野在整張戰場上的位置，單位是格。
type BattleView struct{ Col, Row int }

// Battle 畫戰場畫面：面板背景 ＋ 地形 ＋ 格線。
//
// panel 是 `ANI_PNL0.A98` 解出來的影像，tiles 是 `BLOCK_2.BLK` 的 128 張
// 圖塊，cells 是 `MAP<nn>.CMP` 展開的 128×102 格。
//
// 還沒畫的：小地圖、狀態欄的內容、以及戰場上的單位——原版在這個時機
// 也沒有把單位畫出來（`docs/re/82`）。
func Battle(s *BattleSpec, panel *asset.Image, tiles []*asset.Image, cells []uint8, v BattleView) *gfx.Framebuffer {
	fb := gfx.NewFramebuffer(0)
	if panel != nil {
		for y := 0; y < panel.H; y++ {
			for x := 0; x < panel.W; x++ {
				fb.Set(x, s.Panel.Y+y, panel.Pix[y*panel.W+x])
			}
		}
	}
	drawTerrain(fb, s, tiles, cells, v)
	drawBattleGrid(fb, s, v)
	return fb
}

// drawTerrain 把視野內的格子照 CMP 的編號貼上去。
func drawTerrain(fb *gfx.Framebuffer, s *BattleSpec, tiles []*asset.Image, cells []uint8, v BattleView) {
	if len(tiles) == 0 {
		return
	}
	for r := 0; r < s.Field.Rows; r++ {
		for c := 0; c < s.Field.Cols; c++ {
			i := (v.Row+r)*s.Field.MapCols + v.Col + c
			if i < 0 || i >= len(cells) {
				continue
			}
			t := tiles[int(cells[i])%len(tiles)]
			ox, oy := s.Field.X+c*s.Field.TileW, s.Field.Y+r*s.Field.TileH
			for y := 0; y < t.H; y++ {
				for x := 0; x < t.W; x++ {
					fb.Set(ox+x, oy+y, t.Pix[y*t.W+x])
				}
			}
		}
	}
}

// drawBattleGrid 疊上格線。
//
// 線在**世界座標**每 Pitch 像素一條，橫豎都是；畫出來是虛線——只有
// x 是 DotX 的倍數、而且 y 除以 DotY 的餘數小於 2 的像素才畫，
// 餘 0 用 Ink、餘 1 用 Shadow（`docs/internals/34`）。
func drawBattleGrid(fb *gfx.Framebuffer, s *BattleSpec, v BattleView) {
	p := s.Grid.Pitch
	if p <= 0 {
		return
	}
	// 視野左上角在世界座標的位置。
	wx0 := v.Col * s.Field.TileW
	wy0 := v.Row * s.Field.TileH
	w := s.Field.Cols * s.Field.TileW
	h := s.Field.Rows * s.Field.TileH
	dot := func(x, y int) {
		if x%s.Grid.DotX != 0 {
			return
		}
		switch y % s.Grid.DotY {
		case 0:
			fb.Set(x, y, s.Grid.Ink)
		case 1:
			fb.Set(x, y, s.Grid.Shadow)
		}
	}
	for wx := (wx0/p + 1) * p; wx < wx0+w; wx += p {
		x := s.Field.X + wx - wx0
		for y := s.Field.Y; y < s.Field.Y+h; y++ {
			dot(x, y)
		}
	}
	for wy := (wy0/p + 1) * p; wy < wy0+h; wy += p {
		// 橫線佔兩列——上列 Ink、下列 Shadow，與豎線每 4 列取 2 列同一個相位。
		for y := s.Field.Y + wy - wy0; y < s.Field.Y+wy-wy0+2; y++ {
			for x := s.Field.X; x < s.Field.X+w; x++ {
				dot(x, y)
			}
		}
	}
}

// BattleStatus 是右側狀態欄要顯示的東西。
type BattleStatus struct {
	Faction  int      // 0 帝國、1 同盟：決定第一行的軍名
	Year     int      // 宇宙曆
	Month    int      // 月
	Day      int      // 日
	Turn     int      // 第幾回合
	Victory  int      // 勝利點
	Empire   int      // 帝國軍的數字
	Alliance int      // 同盟軍的數字
	Region   []uint16 // 星域名（Big5 全形碼）
}

// label 是規格裡的一段文字：原文加上 Big5 碼。
type label struct {
	Text  string   `json:"text"`
	Codes []uint16 `json:"codes"`
}

// StatusSpec 是狀態欄的版面。
type StatusSpec struct {
	X, Y, W, H  int
	LabelX      int `json:"label_x"`
	ColonX      int `json:"colon_x"`
	Ink         uint8
	FactionInk  uint8 `json:"faction_ink"`
	ValueRight  int   `json:"value_right"`
	PhaseX      int   `json:"phase_x"`
	TurnNumberX int   `json:"turn_number_x"`
	Rows        struct {
		Phase, Date, Turn, Victory, Empire, Alliance, Region int
	} `json:"rows"`
	DateX struct {
		Year, Month, Day int
	} `json:"date_x"`
	DateInk struct {
		Year, Month, Day uint8
	} `json:"date_ink"`
	Frame struct {
		X, Y, W, H int
		Fill       uint8
		Order      string
		Layers     [][4]uint8
		Alt        [][4]uint8
		Patches    []struct {
			DX, DY int
			Color  uint8
		} `json:"patches"`
	} `json:"frame"`
	Labels struct {
		Factions    []label `json:"factions"`
		Colon       uint16  `json:"colon"`
		PhaseInput  label   `json:"phase_input"`
		Date        label   `json:"date"`
		TurnPrefix  label   `json:"turn_prefix"`
		TurnSuffix  label   `json:"turn_suffix"`
		Victory     label   `json:"victory"`
		FullDigits  label   `json:"full_digits"`
	} `json:"labels"`
}

// DrawBattleStatusFrame 畫狀態欄的視窗框與底色。
//
// 框逐層量自收據（`docs/internals/34`）：外三層亮邊、左緣是 8／15 的
// 1:1 交錯、中間一層 7、兩層反向的立體邊、一層 3，最後一層只有上緣
// 與右緣是 3——那一層的下緣與左緣填底色，靠 Order 的角落歸屬做出來。
func DrawBattleStatusFrame(s *BattleSpec, fb *gfx.Framebuffer) {
	f := &s.Status.Frame
	if f.W <= 0 || f.H <= 0 {
		return
	}
	fb.FillRect(f.X, f.Y, f.W, f.H, f.Fill)
	var fr ui.Frame
	fr.Fill = f.Fill
	fr.Order = f.Order
	for _, l := range f.Layers {
		fr.Layers = append(fr.Layers, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
	}
	for _, l := range f.Alt {
		fr.Alt = append(fr.Alt, ui.Layer{Top: l[0], Right: l[1], Bottom: l[2], Left: l[3]})
	}
	for _, p := range f.Patches {
		fr.Patches = append(fr.Patches, ui.Patch{DX: p.DX, DY: p.DY, Color: p.Color})
	}
	fr.Draw(fb, f.X, f.Y, f.W, f.H)
}

// DrawBattleStatus 畫右側狀態欄的文字。
//
// 版面全部來自規格（`docs/internals/data/screen-battle.json`）；
// 數字靠右對齊到 ValueRight，日期的年月日各一種顏色
// （`docs/internals/34`）。
func DrawBattleStatus(s *BattleSpec, fb *gfx.Framebuffer, fnt *font.Font, st BattleStatus) {
	p := &s.Status
	if len(p.Labels.Factions) < 2 {
		return
	}
	ink, r := p.Ink, p.Rows

	// 第一行：軍名（陣營色）＋ 全形冒號 ＋ 階段。
	drawFull(fb, fnt, p.LabelX, r.Phase, p.Labels.Factions[st.Faction%2].Codes, p.FactionInk)
	drawFull(fb, fnt, p.ColonX, r.Phase, []uint16{p.Labels.Colon}, ink)
	drawFull(fb, fnt, p.PhaseX, r.Phase, p.Labels.PhaseInput.Codes, ink)

	// 日期：「宇宙曆﹕」之後年月日各一種顏色，年是三位、月日各兩位。
	drawFull(fb, fnt, p.LabelX, r.Date, p.Labels.Date.Codes, ink)
	drawFull(fb, fnt, p.ColonX, r.Date, []uint16{p.Labels.Colon}, ink)
	drawPadded(fb, fnt, p.DateX.Year, r.Date, st.Year, 3, p.DateInk.Year)
	drawPadded(fb, fnt, p.DateX.Month, r.Date, st.Month, 2, p.DateInk.Month)
	drawPadded(fb, fnt, p.DateX.Day, r.Date, st.Day, 2, p.DateInk.Day)

	// 「第Ｎ回合」：中間那個數字是全形的。
	drawFull(fb, fnt, p.LabelX, r.Turn, p.Labels.TurnPrefix.Codes, ink)
	drawFullNumber(fb, fnt, p.TurnNumberX, r.Turn, st.Turn, p.Labels.FullDigits.Codes, ink)
	drawFull(fb, fnt, p.TurnNumberX+font.GlyphW, r.Turn, p.Labels.TurnSuffix.Codes, ink)

	// 三個計數：標籤 ＋ 冒號 ＋ 靠右的半形數字。
	for _, row := range []struct {
		y     int
		codes []uint16
		n     int
	}{
		{r.Victory, p.Labels.Victory.Codes, st.Victory},
		{r.Empire, p.Labels.Factions[0].Codes, st.Empire},
		{r.Alliance, p.Labels.Factions[1].Codes, st.Alliance},
	} {
		drawFull(fb, fnt, p.LabelX, row.y, row.codes, ink)
		drawFull(fb, fnt, p.ColonX, row.y, []uint16{p.Labels.Colon}, ink)
		drawRight(fb, fnt, p.ValueRight, row.y, itoa(row.n), ink)
	}

	drawFull(fb, fnt, p.LabelX, r.Region, st.Region, ink)
}

// drawPadded 畫固定位數的半形數字，不足的補 0。
func drawPadded(fb *gfx.Framebuffer, fnt *font.Font, x, y, n, width int, ink uint8) {
	s := itoa(n)
	for len(s) < width {
		s = "0" + s
	}
	for i := 0; i < len(s); i++ {
		drawHalf(fb, fnt, x+i*font.HalfW, y, s[i], ink)
	}
}

// drawFullNumber 用全形數字畫一個數（「第１回合」中間那一個）。
func drawFullNumber(fb *gfx.Framebuffer, fnt *font.Font, x, y, n int, digits []uint16, ink uint8) {
	if len(digits) < 10 {
		return
	}
	s := itoa(n)
	for i := 0; i < len(s); i++ {
		drawFull(fb, fnt, x+i*font.GlyphW, y, []uint16{digits[s[i]-'0']}, ink)
	}
}

// DrawBattleUnits 把分隊畫到戰場上。
//
// 單位是 `SPRITE.SPR` 的 16×16 圖塊，**左上角對齊分隊的世界座標**
// （`docs/re/85`）。哪個艦種用哪一塊還沒解完，所以圖塊索引由呼叫端給。
//
// 超出戰場那一塊的部分會被裁掉——原版的戰場是有邊界的視窗。
func DrawBattleUnits(s *BattleSpec, fb *gfx.Framebuffer, sprites []*asset.Image,
	units []rules.BattleUnit, sprite func(rules.BattleUnit) int, v BattleView) {
	if len(sprites) == 0 || sprite == nil {
		return
	}
	fx, fy := s.Field.X, s.Field.Y
	fw, fh := s.Field.Cols*s.Field.TileW, s.Field.Rows*s.Field.TileH
	ox, oy := v.Col*s.Field.TileW, v.Row*s.Field.TileH
	for _, u := range units {
		n := sprite(u)
		if n < 0 || n >= len(sprites) {
			continue
		}
		t := sprites[n]
		sx, sy := fx+u.X-ox, fy+u.Y-oy
		for y := 0; y < t.H; y++ {
			for x := 0; x < t.W; x++ {
				px, py := sx+x, sy+y
				if px < fx || py < fy || px >= fx+fw || py >= fy+fh {
					continue
				}
				fb.Set(px, py, t.Pix[y*t.W+x])
			}
		}
	}
}

// MinimapSpec 是小地圖的版面（`docs/internals/34`）。
type MinimapSpec struct {
	X, Y, W, H    int
	GridPitch     int            `json:"grid_pitch"`
	GridInk       uint8          `json:"grid_ink"`
	ViewOr        uint8          `json:"view_or"`
	TileColour    map[string]int `json:"tile_colour"`
	DefaultColour uint8          `json:"default_colour"`
	TopInk        uint8          `json:"top_ink"`
	SampleNum     int            `json:"sample_num"`
	SampleDen     int            `json:"sample_den"`
	UnitScale     int            `json:"unit_scale"`
	UnitRowBias   int            `json:"unit_row_bias"`
	UnitInk       []uint8        `json:"unit_ink"`
	PlanetInk     uint8          `json:"planet_ink"`
}

// DrawMinimap 畫右上角的小地圖。
//
// 三層（`docs/internals/34`）：
//
//   - 地形：像素 (i,j) 顯示格 `(i, j × 4 ÷ 5)`——4/5 就是圖塊 16:20 的比例，
//     色由圖塊編號查表；
//   - 格線：`i % 15 == 0` 或 `j % 15 < 2` 畫成底色，與戰場那層同一組；
//   - 視野框：左上 (視野欄, 視野列 × 高 ÷ 列數 **向上取整**)，大小就是視野的格數。
//     框不是固定色——它把底下那一格的色**或上 ViewOr**（地形 1 變 5、3 變 7）。
//
// 第 0 列是邊框（色 TopInk），不是格線。
func DrawMinimap(s *BattleSpec, fb *gfx.Framebuffer, cells []uint8, v BattleView) {
	m := &s.Minimap
	if m.W <= 0 || m.H <= 0 || len(cells) == 0 {
		return
	}
	num, den := m.SampleNum, m.SampleDen
	if num <= 0 || den <= 0 {
		num, den = s.Field.TileW, s.Field.TileH
	}
	for j := 0; j < m.H; j++ {
		r := j * num / den
		if r >= s.Field.MapRows {
			r = s.Field.MapRows - 1
		}
		for i := 0; i < m.W; i++ {
			c := m.DefaultColour
			if j == 0 {
				c = m.TopInk
			} else if m.GridPitch > 0 && (i%m.GridPitch == 0 || j%m.GridPitch < 2) {
				c = m.GridInk
			} else if k := r*s.Field.MapCols + i; k < len(cells) {
				if n, ok := m.TileColour[itoa(int(cells[k]))]; ok {
					c = uint8(n)
				}
			}
			fb.Set(m.X+i, m.Y+j, c)
		}
	}
	// 視野框：只畫邊。
	vx := m.X + v.Col
	// 上邊向上取整——`s3ai6` 的視野列 30 算出來是 37.35，原版畫在 38。
	vy := m.Y + (v.Row*m.H+s.Field.MapRows-1)/s.Field.MapRows
	vw, vh := s.Field.Cols, s.Field.Rows*m.H/s.Field.MapRows
	mark := func(x, y int) { fb.Set(x, y, fb.At(x, y)|m.ViewOr) }
	for i := 0; i < vw; i++ {
		mark(vx+i, vy)
		mark(vx+i, vy+vh-1)
	}
	for j := 0; j < vh; j++ {
		mark(vx, vy+j)
		mark(vx+vw-1, vy+j)
	}
}

// minimapPixel 把世界座標換成小地圖上的像素（`docs/re/89`）。
func (m *MinimapSpec) minimapPixel(x, y int) (int, int, bool) {
	if m.UnitScale <= 0 {
		return 0, 0, false
	}
	i := x / m.UnitScale
	j := y/m.UnitScale + m.UnitRowBias
	if i < 0 || j < 0 || i >= m.W || j >= m.H {
		return 0, 0, false
	}
	return m.X + i, m.Y + j, true
}

// DrawMinimapPlanets 在小地圖上點出這個星域的行星（`docs/re/89`）。
//
// 畫在分隊之前——兩者疊在同一格時，分隊那層是把底下的值 XOR 掉。
func DrawMinimapPlanets(s *BattleSpec, fb *gfx.Framebuffer, sites []rules.PlanetSite) {
	m := &s.Minimap
	if m.W <= 0 || m.H <= 0 || m.PlanetInk == 0 {
		return
	}
	for _, p := range sites {
		if px, py, ok := m.minimapPixel(p.X, p.Y); ok {
			fb.Set(px, py, m.PlanetInk)
		}
	}
}

// DrawMinimapUnits 在小地圖上點出分隊（`docs/re/89`）。
//
// 每個分隊一個像素，位置是世界座標除以 `UnitScale`（縱向再加一列），
// 顏色是把底下的值 **XOR** 上陣營的墨色（帝國 5、同盟 9）。原版的來源
// 欄位是艦艇記錄的 `+0x1A`／`+0x1C`（不是畫在戰場上用的
// `+0x28`／`+0x2A`），值為 `0xFFFF` 的分隊不點。
func DrawMinimapUnits(s *BattleSpec, fb *gfx.Framebuffer, units []rules.BattleUnit) {
	m := &s.Minimap
	if m.W <= 0 || m.H <= 0 || len(m.UnitInk) == 0 {
		return
	}
	for _, u := range units {
		px, py, ok := m.minimapPixel(u.X, u.Y)
		if !ok {
			continue
		}
		ink := m.UnitInk[0]
		if u.Faction >= 0 && u.Faction < len(m.UnitInk) {
			ink = m.UnitInk[u.Faction]
		}
		fb.Set(px, py, fb.At(px, py)^ink)
	}
}
