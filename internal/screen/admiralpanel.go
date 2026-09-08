package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// AdmiralPanel 是「提督詳細資料」那個視窗裡跟著提督走的東西
// （規格 `docs/spec/012`）：姓名與軍階、頭像、現在位置、旗艦、
// 等級與經驗值，以及十個能力值。
//
// 視窗其餘的字（各欄的標籤、三顆按鈕）是版面，留在畫面規格裡。
type AdmiralPanel struct {
	Ink  uint8 `json:"ink"`
	Name struct {
		X, Y int
	} `json:"name"`
	Rank struct {
		X, Y int
	} `json:"rank"`
	Face struct {
		X, Y int
	} `json:"face"`
	Location struct {
		X, Y int
	} `json:"location"`
	Flagship struct {
		X, Y   int
		LabelX int      `json:"label_x"`
		Label  []uint16 `json:"label"`
	} `json:"flagship"`
	Level struct {
		Y     int `json:"y"`
		Right int `json:"right"`
	} `json:"level"`
	Exp struct {
		Y     int `json:"y"`
		Right int `json:"right"`
	} `json:"exp"`
	Buttons []struct {
		Action     string `json:"action"`
		Text       string `json:"text"`
		X, Y, W, H int
	} `json:"buttons"`
	Stats []struct {
		Y      int    `json:"y"`
		Right  int    `json:"right"`
		Offset int    `json:"offset"`
		Name   string `json:"name"`
	} `json:"stats"`
}

// LoadAdmiralPanel 讀版面。
func LoadAdmiralPanel(path string) (*AdmiralPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p AdmiralPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// AdmiralView 是畫這張畫面要的一整組值。
type AdmiralView struct {
	Name, Rank []uint16 // 姓名與軍階的字碼
	Location   []uint16 // 現在位置的星域名
	Flagship   []uint16 // 旗艦名；空 ＝ 沒有艦隊，那一行不畫
	Level, Exp int
	Record     []byte // ADMIRAL.DT* 的那一筆，能力值直接照位移取
}

// Draw 把算出來的部分畫上去。頭像另外畫（要 `FACE.AMG`）。
//
// `stats` ＝ false 時只畫上半（姓名、軍階、現在位置、旗艦）——經歷那張
// 畫面的視窗蓋住了下半，等級與能力值不該再畫上去。
func (p *AdmiralPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, v AdmiralView, stats bool) {
	// 姓名與軍階**各有自己的欄**，不是接在一起（`adm1e`：姓名 x=224、
	// 軍階 x=368）。
	drawFull(fb, fnt, p.Name.X, p.Name.Y, v.Name, p.Ink)
	drawFull(fb, fnt, p.Rank.X, p.Rank.Y, v.Rank, p.Ink)
	if len(v.Location) > 0 {
		drawFull(fb, fnt, p.Location.X, p.Location.Y, v.Location, p.Ink)
	}
	// 沒有艦隊的提督連「旗艦」那個標籤都不畫（收據 `adm15e`）。
	if len(v.Flagship) > 0 {
		drawFull(fb, fnt, p.Flagship.LabelX, p.Flagship.Y, p.Flagship.Label, p.Ink)
		drawFull(fb, fnt, p.Flagship.X, p.Flagship.Y, v.Flagship, p.Ink)
	}
	if !stats {
		return
	}
	drawRight(fb, fnt, p.Level.Right, p.Level.Y, strconv.Itoa(v.Level), p.Ink)
	drawRight(fb, fnt, p.Exp.Right, p.Exp.Y, strconv.Itoa(v.Exp), p.Ink)
	for _, s := range p.Stats {
		if s.Offset < 0 || s.Offset >= len(v.Record) {
			continue
		}
		drawRight(fb, fnt, s.Right, s.Y, strconv.Itoa(int(v.Record[s.Offset])), p.Ink)
	}
}


// ButtonAt 回畫面座標 (x,y) 落在哪一顆按鈕上；沒有就回空字串。
func (p *AdmiralPanel) ButtonAt(x, y int) string {
	for _, b := range p.Buttons {
		if x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H {
			return b.Action
		}
	}
	return ""
}
