package screen

import (
	"encoding/json"
	"os"
	"sort"

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
		X, Y         int    `json:"-"`
		Cols, Rows   int    `json:"-"`
		TileW, TileH int    `json:"-"`
		Tiles        string `json:"tiles"`
		MapCols      int    `json:"map_cols"`
		MapRows      int    `json:"map_rows"`
		XRaw         int    `json:"x"`
		YRaw         int    `json:"y"`
		ColsRaw      int    `json:"cols"`
		RowsRaw      int    `json:"rows"`
		TileWRaw     int    `json:"tile_w"`
		TileHRaw     int    `json:"tile_h"`
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
	fb := gfx.NewFramebuffer(3) // 原版上下外區是色號3，色號0為藍色（規格080／研究373）。
	if panel != nil {
		for y := 0; y < panel.H; y++ {
			for x := 0; x < panel.W; x++ {
				fb.Set(x, s.Panel.Y+y, panel.Pix[y*panel.W+x])
			}
		}
	}
	drawTerrain(fb, s, tiles, cells, v)
	drawBattleGrid(fb, s, cells, v)
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
// 餘 0 用 Ink、餘 1 用 Shadow（`docs/internals/34`）。**只畫在空格上**：
// 有地形圖塊的格子（編號不是 0）上沒有點（收據 `cam-e`：橫線走到星雲
// 那一格就斷，`docs/re/326`）。
func drawBattleGrid(fb *gfx.Framebuffer, s *BattleSpec, cells []uint8, v BattleView) {
	p := s.Grid.Pitch
	if p <= 0 {
		return
	}
	// 視野左上角在世界座標的位置。
	wx0 := v.Col * s.Field.TileW
	wy0 := v.Row * s.Field.TileH
	w := s.Field.Cols * s.Field.TileW
	h := s.Field.Rows * s.Field.TileH
	empty := func(x, y int) bool {
		if len(cells) == 0 {
			return true
		}
		c := v.Col + (x-s.Field.X)/s.Field.TileW
		r := v.Row + (y-s.Field.Y)/s.Field.TileH
		i := r*s.Field.MapCols + c
		return i < 0 || i >= len(cells) || cells[i] == 0
	}
	dot := func(x, y int) {
		if x%s.Grid.DotX != 0 || !empty(x, y) {
			return
		}
		switch y % s.Grid.DotY {
		case 0:
			fb.Set(x, y, s.Grid.Ink)
		case 1:
			fb.Set(x, y, s.Grid.Shadow)
		}
	}
	// 視野左上角正好落在線上時那條線也要畫（收據 `cam-w`／`cam-n`：x ＝ 16、y ＝ 60）。
	for wx := (wx0 + p - 1) / p * p; wx < wx0+w; wx += p {
		x := s.Field.X + wx - wx0
		for y := s.Field.Y; y < s.Field.Y+h; y++ {
			dot(x, y)
		}
	}
	for wy := (wy0 + p - 1) / p * p; wy < wy0+h; wy += p {
		// 橫線佔兩列——上列 Ink、下列 Shadow，與豎線每 4 列取 2 列同一個相位。
		for y := s.Field.Y + wy - wy0; y < s.Field.Y+wy-wy0+2; y++ {
			for x := s.Field.X; x < s.Field.X+w; x++ {
				dot(x, y)
			}
		}
	}
}

// BattlePhase只表示083已審查的階段文字，不改八步規則或存檔。
type BattlePhase uint8

const (
	BattlePhaseInput BattlePhase = iota
	BattlePhaseSupply
	BattlePhaseRepair
	BattlePhaseMovement
	BattlePhaseSpotting
)

// BattleStatus 是右側狀態欄要顯示的東西。
type BattleStatus struct {
	Phase    BattlePhase
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
		// X／Y／W／H 是**看得見**的那個矩形（框最外圈那一像素與底色
		// 同色），幾何測試用；畫的時候看的是格。
		X, Y, W, H int
		Fill       uint8

		CellX  int `json:"cell_x"`
		CellY  int `json:"cell_y"`
		Cols   int `json:"cols"`
		Rows   int `json:"rows"`
		Style  int `json:"style"`
		Colour int `json:"colour"`
	} `json:"frame"`
	Labels struct {
		Factions      []label `json:"factions"`
		Colon         uint16  `json:"colon"`
		PhaseInput    label   `json:"phase_input"`
		PhaseSupply   label   `json:"phase_supply"`
		PhaseRepair   label   `json:"phase_repair"`
		PhaseMovement label   `json:"phase_movement"`
		PhaseSpotting label   `json:"phase_spotting"`
		Date          label   `json:"date"`
		TurnPrefix    label   `json:"turn_prefix"`
		TurnSuffix    label   `json:"turn_suffix"`
		Victory       label   `json:"victory"`
		FullDigits    label   `json:"full_digits"`
	} `json:"labels"`
}

// DrawBattleStatusFrame 畫狀態欄的視窗框與底色。
//
// 框不是畫線畫出來的，是 `FRAME.BLK` 的圖塊：原版在 `0C0A:6B6D` 用
// `3603:0006(色 0x97, 格 (29,9), 內容 8×8 格, 旗標 1)` 建這個視窗，
// 旗標讓它實際佔 10 × 10 格（`docs/re/142`）。tiles 是 nil 就不畫——
// `FRAME.BLK` 是原版素材，缺檔不用代用品。
func DrawBattleStatusFrame(s *BattleSpec, fb *gfx.Framebuffer, tiles *ui.FrameTiles) {
	f := &s.Status.Frame
	if tiles == nil || f.Cols <= 0 || f.Rows <= 0 {
		return
	}
	(ui.TileFrame{Tiles: tiles, Style: f.Style, Fill: uint8(f.Colour >> 4)}).
		Draw(fb, f.CellX, f.CellY, f.Cols, f.Rows)
}

// DrawBattleStatus 畫右側狀態欄的文字。
//
// 版面全部來自規格（`docs/internals/data/screen-battle.json`）；
// 數字靠右對齊到 ValueRight，日期的年月日各一種顏色
// （`docs/internals/34`）。
func DrawBattleStatus(s *BattleSpec, fb *gfx.Framebuffer, fnt *font.Font, st BattleStatus) {
	p := &s.Status
	// 原版0C先清內容；Game每幀先畫同一視窗框，再只印階段文字。
	if st.Phase != BattlePhaseInput {
		var codes []uint16
		switch st.Phase {
		case BattlePhaseSupply:
			codes = p.Labels.PhaseSupply.Codes
		case BattlePhaseRepair:
			codes = p.Labels.PhaseRepair.Codes
		case BattlePhaseMovement:
			codes = p.Labels.PhaseMovement.Codes
		case BattlePhaseSpotting:
			codes = p.Labels.PhaseSpotting.Codes
		}
		drawFull(fb, fnt, p.LabelX, p.Rows.Phase, codes, p.Ink)
		return
	}
	if len(p.Labels.Factions) < 2 {
		return
	}
	ink, r := p.Ink, p.Rows

	// 第一行：軍名（陣營色）＋ 全形冒號 ＋ 階段。
	viewer := st.Faction & 1
	factionInk := uint8(6)
	if viewer == 1 {
		factionInk = 5
	}
	drawFull(fb, fnt, p.LabelX, r.Phase, p.Labels.Factions[viewer].Codes, factionInk)
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
	// 原版依玩家排列兩方分數；勝利點也是玩家減對手（研究395）。
	scores := [2]int{st.Empire, st.Alliance}
	victory := st.Victory
	if viewer == 1 {
		victory = -victory
	}
	for _, row := range []struct {
		y     int
		codes []uint16
		n     int
	}{
		{r.Victory, p.Labels.Victory.Codes, victory},
		{r.Empire, p.Labels.Factions[viewer].Codes, scores[viewer]},
		{r.Alliance, p.Labels.Factions[1-viewer].Codes, scores[1-viewer]},
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
	// 原版物件的word +4由艦隊與分隊序號組成，由大到小畫（085）。
	// 偵察分隊的-1轉成原版0xFF；排序副本，保留規則與選取的次序。
	order := append([]rules.BattleUnit(nil), units...)
	key := func(u rules.BattleUnit) uint16 {
		return uint16(uint8(u.Fleet)) | uint16(uint8(u.Squadron))<<8
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := uint8(order[i].DisplayPriority), uint8(order[j].DisplayPriority)
		if a != b {
			return a < b
		}
		return key(order[i]) > key(order[j])
	})
	for _, u := range order {
		if u.Out {
			continue
		}
		n := sprite(u)
		if n == rules.BlankSprite || n < 0 || n >= len(sprites) {
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

// DrawBattleFleetDirections沿084畫完整可見旗艦的八向圖示；色號0透明。
func DrawBattleFleetDirections(s *BattleSpec, fb *gfx.Framebuffer, sprites []*asset.Image,
	fleets []rules.BattleFleet, units []rules.BattleUnit, viewer int, v BattleView) {
	DrawBattleFleetDirectionsCached(s, fb, sprites, fleets, units, viewer, v, nil)
}

// 087情報先升級，方向圖示沿FLEET+6A既有快取，不提前套新Detail。
func DrawBattleFleetDirectionsCached(s *BattleSpec, fb *gfx.Framebuffer, sprites []*asset.Image,
	fleets []rules.BattleFleet, units []rules.BattleUnit, viewer int, v BattleView, cached map[int]int) {
	if s == nil || len(sprites) < 8 {
		return
	}
	offsets := [8][2]int{{16, 0}, {8, 8}, {0, 16}, {-8, 8}, {-16, 0}, {-8, -8}, {0, -16}, {8, -8}}
	fx, fy := s.Field.X, s.Field.Y
	fw, fh := s.Field.Cols*s.Field.TileW, s.Field.Rows*s.Field.TileH
	for _, f := range fleets {
		visible := false
		for _, u := range units {
			if u.Fleet == f.Index && u.Squadron == 0 && !u.Out && (u.Faction == viewer || u.Detail == rules.SpottingMax) {
				visible = true
				break
			}
		}
		d := rules.DirectionIndex(f.Facing)
		if n, ok := cached[f.Index]; ok {
			d, visible = n, n >= 0 && n < 8
		}
		if !visible {
			continue
		}
		t := sprites[d]
		if t == nil {
			continue
		}
		sx := fx + f.X - v.Col*s.Field.TileW + offsets[d][0]
		sy := fy + f.Y - v.Row*s.Field.TileH + offsets[d][1]
		for y := 0; y < t.H; y++ {
			for x := 0; x < t.W; x++ {
				px, py := sx+x, sy+y
				ink := t.Pix[y*t.W+x]
				if ink != 0 && px >= fx && py >= fy && px < fx+fw && py < fy+fh {
					fb.Set(px, py, ink)
				}
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

// battleMinimapLookup 是原版 DS:3C8C 的位元投影（規格086）。
// 尾列的非圖塊輸入亦會使用表後的位元；完整來源與雜湊見分級索引。
var battleMinimapLookup = [...]byte{
	0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0x80, 0x40, 0x20, 0x10, 0x08, 0x04, 0x02, 0x01, 0x0C, 0xB8, 0xC9, 0xB5, 0xB9, 0xB6, 0xA5, 0xAC,
	0x71, 0x20, 0x20, 0x20, 0x20, 0x00, 0x0C, 0xAD, 0xD7, 0xB2, 0x7A, 0xB6, 0xA5, 0xAC, 0x71, 0x20,
	0x20, 0x20, 0x20, 0x00, 0x0C, 0xB2, 0xBE, 0xB0, 0xCA, 0xB6, 0xA5, 0xAC, 0x71, 0x20, 0x20, 0x20,
	0x20, 0x00, 0xB8, 0xD1, 0xB0, 0xA3, 0xA6, 0xDB, 0xB0, 0xCA, 0xBC, 0xD2, 0xA6, 0xA1, 0x20, 0x20,
	0x00, 0x0C, 0xAF, 0xC1, 0xBC, 0xC4, 0xB6, 0xA5, 0xAC, 0x71, 0x20, 0x20, 0x20, 0x20, 0x00, 0x0C,
	0xBE, 0xD4, 0xB0, 0xAB, 0xB6, 0xA5, 0xAC, 0x71, 0x20, 0x20, 0x20, 0x20, 0x00, 0x0C, 0xA5, 0x65,
	0xBB, 0xE2, 0xB6, 0xA5, 0xAC, 0x71, 0x20, 0x20, 0x20, 0x20, 0x00, 0x0C, 0xB3, 0xCC, 0xAB, 0xE1,
	0xB6, 0xA5, 0xAC, 0x71, 0x20, 0x20, 0x20, 0x20, 0x00, 0x00, 0x20, 0x20, 0x0A, 0xA4, 0x77, 0xB8,
	0x67, 0xA5, 0xFE, 0xB7, 0xC0, 0xA1, 0x43, 0x20, 0x20, 0x20, 0x20, 0x00, 0x25, 0x73, 0xAC, 0x50,
	0xA8, 0x74, 0x20, 0x20, 0x0A, 0xA4, 0x77, 0xB8, 0x67, 0xA5, 0x65, 0xBB, 0xE2, 0xA4, 0x55, 0xA8,
	0xD3, 0xA1, 0x43, 0x0A, 0x00, 0x25, 0x73, 0xAC, 0x50, 0xA8, 0x74, 0xB3, 0xA1, 0xA4, 0xC0, 0xA6,
	0x61, 0xB0, 0xCF, 0x0A, 0xA4, 0x77, 0xB2, 0xE6, 0xC2, 0xF7, 0xA7, 0xDA, 0xA4, 0xE8, 0xB2, 0xCE,
}

// battleMinimapBitmap 在有界資料內重現原版地圖及鄰接配置取樣。
// 原版最後列讀到標頭與已產生的 bitmap，不能用固定局面遮罩替代。
func battleMinimapBitmap(s *BattleSpec, cells []uint8) []byte {
	m := &s.Minimap
	if s.Field.MapCols != 128 || s.Field.MapRows != 102 || len(cells) != 128*102 ||
		m.W != 128 || m.H != 128 || m.SampleNum != 4 || m.SampleDen != 5 || m.GridPitch != 15 {
		return nil
	}
	data := make([]byte, len(cells)+2+2048)
	copy(data, cells)
	for i, c := range cells {
		if c != 0 {
			continue
		}
		x, y := i%128, i/128
		if x%15 == 0 {
			if y%12 == 0 {
				data[i] = 2
			} else {
				data[i] = 4
			}
		} else if y%12 == 0 {
			data[i] = 3
		}
	}
	// 所驗配置的 2048-byte 區塊有兩個標頭 byte：0208。
	data[len(cells)], data[len(cells)+1] = 2, 8
	out := data[len(cells)+2:]
	for j := 0; j < 128; j++ {
		row := (j + 1) * 4 / 5
		for x := 0; x < 128; x++ {
			out[j*16+x/8] |= battleMinimapLookup[data[row*128+x]] & (0x80 >> uint(x%8))
		}
	}
	return out
}

// DrawMinimap 畫右上角的小地圖。
//
// 原版bitmap從上框下一列起畫128列，尾列亦由地圖資料導出（規格086）。
// 標記、視野框與玩家輸入仍沿用MinimapSpec原點。
func DrawMinimap(s *BattleSpec, fb *gfx.Framebuffer, cells []uint8, v BattleView) {
	m := &s.Minimap
	if m.W <= 0 || m.H <= 0 || len(cells) == 0 {
		return
	}
	num, den := m.SampleNum, m.SampleDen
	if num <= 0 || den <= 0 {
		num, den = s.Field.TileW, s.Field.TileH
	}
	if bits := battleMinimapBitmap(s, cells); bits != nil {
		for x := 0; x < m.W; x++ {
			fb.Set(m.X+x, m.Y, m.TopInk)
		}
		for y := 0; y < m.H; y++ {
			for x := 0; x < m.W; x++ {
				c := m.TopInk
				if bits[y*16+x/8]&(0x80>>uint(x%8)) != 0 {
					c = m.GridInk
				}
				fb.Set(m.X+x, m.Y+1+y, c)
			}
		}
	} else {
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
	}
	// 視野框：只畫邊。
	// 1727:035E將世界鏡頭右移4，沿標記原點加DOS的40列（研究395）。
	// bitmap的128／102比例與世界像素縮放不是同一個座標域。
	scale := m.UnitScale
	if scale <= 0 {
		scale = 16
	}
	vx := m.X + v.Col*s.Field.TileW/scale
	vy := m.Y + m.UnitRowBias + v.Row*s.Field.TileH/scale
	vw, vh := s.Field.Cols*s.Field.TileW/scale, s.Field.Rows*s.Field.TileH/scale
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
// +19的C0均設才畫，以低4位顏色XOR底色（規格082）。原版的來源
// 欄位是艦艇記錄的標記快照 `+0x1A`／`+0x1C`，不取現在位置
// `+0x34`／`+0x36`；值為 `0xFFFF` 的分隊不點。
func DrawMinimapUnits(s *BattleSpec, fb *gfx.Framebuffer, units []rules.BattleUnit) {
	m := &s.Minimap
	if m.W <= 0 || m.H <= 0 || len(m.UnitInk) == 0 {
		return
	}
	for _, u := range units {
		if u.MarkerFlags&0xC0 != 0xC0 {
			continue
		}
		px, py, ok := m.minimapPixel(u.MarkerX, u.MarkerY)
		if !ok {
			continue
		}
		ink := uint8(u.MarkerFlags & 0x0F) // 擦／重畫均沿用已記錄顏色（規格082）。
		fb.Set(px, py, fb.At(px, py)^ink)
	}
}

// DrawMinimapInputHighlight沿088／1727:0677、0736疊摘要快照的集結點。
// 旗艦畫現在至集結點的實線，其餘分隊只點集結點；不改SHIP標記快照。
func DrawMinimapInputHighlight(s *BattleSpec, fb *gfx.Framebuffer, current, target [2]int, fleet bool) {
	if s == nil || fb == nil || target[0] < 0 || target[1] < 0 {
		return
	}
	m := &s.Minimap
	x1, y1, ok := m.minimapPixel(target[0], target[1])
	if !ok {
		return
	}
	const ink uint8 = 4
	if !fleet {
		fb.Set(x1, y1, fb.At(x1, y1)^ink)
		return
	}
	if current[0] < 0 || current[1] < 0 {
		return
	}
	x0, y0, ok := m.minimapPixel(current[0], current[1])
	if !ok {
		return
	}
	r := ClipRect{X0: m.X, Y0: m.Y, X1: m.X + m.W - 1, Y1: m.Y + m.H - 1}
	r.XorLine(fb, x0, y0, x1, y1, ink)
}

// FormationMarkTile 是「編隊」畫面上每個空格子畫的那一塊
// （原版 `2BE1:0564` 寫死 11，`docs/internals/46`）。
const FormationMarkTile = 11

// 25DB:0E78的暫時物件用圖塊10，優先FF，畫在分隊之上。
func DrawBattleSpottingHighlight(s *BattleSpec, fb *gfx.Framebuffer, sprites []*asset.Image, at [2]int, v BattleView) {
	drawBattleMarks(s, fb, sprites, [][2]int{at}, v, 10)
}

// 規格088／25DB:0BD2：集結點優先FE，現在位置FF，兩者使用圖塊10。
func DrawBattleInputHighlight(s *BattleSpec, fb *gfx.Framebuffer, sprites []*asset.Image, current, target [2]int, v BattleView) {
	drawBattleMarks(s, fb, sprites, [][2]int{target, current}, v, 10)
}

// DrawFormationMarks 把「編隊」的格子標記畫上去。
//
// slots 是格子的**世界座標左上角**。色號 0 是透空的——收據 `s3form` 上
// 那 118 個 0 的位置留著戰場底色，沒有被塗掉。
func DrawFormationMarks(s *BattleSpec, fb *gfx.Framebuffer, sprites []*asset.Image,
	slots [][2]int, v BattleView) {
	drawBattleMarks(s, fb, sprites, slots, v, FormationMarkTile)
}

func drawBattleMarks(s *BattleSpec, fb *gfx.Framebuffer, sprites []*asset.Image,
	slots [][2]int, v BattleView, tile int) {
	if s == nil || tile >= len(sprites) {
		return
	}
	t := sprites[tile]
	fx, fy := s.Field.X, s.Field.Y
	fw, fh := s.Field.Cols*s.Field.TileW, s.Field.Rows*s.Field.TileH
	ox, oy := v.Col*s.Field.TileW, v.Row*s.Field.TileH
	for _, p := range slots {
		sx, sy := fx+p[0]-ox, fy+p[1]-oy
		for y := 0; y < t.H; y++ {
			for x := 0; x < t.W; x++ {
				c := t.Pix[y*t.W+x]
				if c == 0 {
					continue
				}
				px, py := sx+x, sy+y
				if px < fx || py < fy || px >= fx+fw || py >= fy+fh {
					continue
				}
				fb.Set(px, py, c)
			}
		}
	}
}
