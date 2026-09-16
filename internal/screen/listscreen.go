package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/gfx"
)

// ListScreenSpec 是一覽表四張共用的清單版面（`docs/internals/data/list-screen-panel.json`，
// `docs/re/323`）：內容區、▲▼、三個分頁鈕的熱區，點開子視窗時那一列的底色帶，
// 以及艦隊資料那張的列。
type ListScreenSpec struct {
	Content Rect   `json:"content"`
	Up      Rect   `json:"up"`
	Down    Rect   `json:"down"`
	Tabs    []Rect `json:"tabs"`
	Rows    int    `json:"rows"`
	Pitch   int    `json:"pitch"`
	Band    struct {
		X, W, Top, Inset int
		Bg               uint8 `json:"bg"`
	} `json:"band"`
	Fleet FleetList `json:"fleet"`
}

// Rect 是一塊熱區（畫面座標）。
type Rect struct{ X, Y, W, H int }

// Contains 回 (x, y) 在不在這一塊裡。
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// LoadListScreenSpec 讀版面。
func LoadListScreenSpec(path string) (*ListScreenSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s ListScreenSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// RowAt 回 (x, y) 落在內容區的第幾列（0 起）；不在內容區回 −1。
func (s *ListScreenSpec) RowAt(x, y int) int {
	if s == nil || s.Pitch <= 0 || !s.Content.Contains(x, y) {
		return -1
	}
	row := (y - s.Content.Y) / s.Pitch
	if row >= s.Rows {
		return -1
	}
	return row
}

// ArrowAt 回 ▲（−1）／▼（＋1）；都不是回 0。
func (s *ListScreenSpec) ArrowAt(x, y int) int {
	if s == nil {
		return 0
	}
	if s.Up.Contains(x, y) {
		return -1
	}
	if s.Down.Contains(x, y) {
		return 1
	}
	return 0
}

// TabAt 回點到第幾個分頁鈕；不是回 −1。
func (s *ListScreenSpec) TabAt(x, y int) int {
	for i, t := range s.Tabs {
		if t.Contains(x, y) {
			return i
		}
	}
	return -1
}

// InHotspots 回 (x, y) 有沒有落在清單的任何熱區（內容區、▲▼、分頁）——
// 艦隊資料的游標處理只看「有沒有命中」，不看命中哪一塊（`0C7F:03CB`）。
func (s *ListScreenSpec) InHotspots(x, y int) bool {
	return s != nil && (s.Content.Contains(x, y) || s.Up.Contains(x, y) ||
		s.Down.Contains(x, y) || s.TabAt(x, y) >= 0)
}

// DrawBand 把第 row 列刷成子視窗開著時的底色（`3603:00DB(…, 0x20, …)`）。
func (s *ListScreenSpec) DrawBand(fb *gfx.Framebuffer, row int) {
	if s == nil || row < 0 || row >= s.Rows || s.Band.W <= 0 {
		return
	}
	top := s.Band.Top + row*s.Pitch
	fb.FillRect(s.Band.X, top+s.Band.Inset, s.Band.W, s.Pitch-2*s.Band.Inset, s.Band.Bg)
}
