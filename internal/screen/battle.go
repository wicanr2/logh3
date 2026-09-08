package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
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
