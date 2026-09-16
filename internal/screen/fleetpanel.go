package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// FleetPanel 是「艦隊編成」那個視窗裡跟著提督與副官走的東西
// （規格 `docs/spec/013`）：艦隊編號、提督與副官各自的姓名軍階、
// 等級、經驗值與旗艦，以及九格艦隊能力。
//
// 艦隊能力每一格是 `max(提督, 副官)`；**副官贏的那幾格印成 AltInk**
// ——原版自己把「這一格是副官撐起來的」標了出來（`docs/re/66`）。
type FleetPanel struct {
	Ink    uint8 `json:"ink"`
	AltInk uint8 `json:"alt_ink"`
	Title  struct {
		Y       int `json:"y"`
		PrefixX int `json:"prefix_x"`
		DigitsX int `json:"digits_x"`
	} `json:"title"`
	Admiral  fleetSide `json:"admiral"`
	Adjutant fleetSide `json:"adjutant"`
	Stats    []struct {
		Y      int    `json:"y"`
		Right  int    `json:"right"`
		Offset int    `json:"offset"`
		Name   string `json:"name"`
	} `json:"stats"`
	// NoAdjutant 是沒有副官時要蓋成底色的那一塊（等級／經驗值那一列連標籤，
	// `07F9:0C24` 印空白；`docs/re/318`）。
	NoAdjutant struct {
		X, Y, W, H int
		Fill       uint8
	} `json:"no_adjutant"`

	Buttons []struct {
		Action     string `json:"action"`
		Text       string `json:"text"`
		X, Y, W, H int
	} `json:"buttons"`
}

type fleetSide struct {
	Name struct {
		X, Y int
	} `json:"name"`
	Face struct {
		X, Y int
	} `json:"face"`
	Flagship struct {
		X, Y int
	} `json:"flagship"`
	Level struct {
		Y     int `json:"y"`
		Right int `json:"right"`
	} `json:"level"`
	Exp struct {
		Y     int `json:"y"`
		Right int `json:"right"`
	} `json:"exp"`
}

// LoadFleetPanel 讀版面。
func LoadFleetPanel(path string) (*FleetPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p FleetPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// FleetView 是畫這張畫面要的一整組值。
type FleetView struct {
	Number                     int      // 第幾艦隊
	AdmiralName                []uint16 // 姓名 ＋ 軍階，接在一起
	AdmiralRank                []uint16
	AdmiralLevel, AdmiralExp   int
	Flagship                   []uint16
	AdjutantName               []uint16
	AdjutantRank               []uint16
	AdjutantLevel, AdjutantExp int
	AdmiralRecord              []byte
	AdjutantRecord             []byte
	// NoAdjutant ＝ 沒有副官（提督記錄 `+0x07` ＝ 0xFF，「變更 → 副官 → 卸任」
	// 之後）：副官那三格不印，九格能力只看提督。
	NoAdjutant bool
}

// Draw 把算出來的部分畫上去。兩張頭像另外畫（要 `FACE.AMG`）。
func (p *FleetPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font, v FleetView) {
	drawFull(fb, fnt, p.Title.DigitsX, p.Title.Y, fullDigits(v.Number), p.Ink)

	drawFull(fb, fnt, p.Admiral.Name.X, p.Admiral.Name.Y,
		append(append([]uint16(nil), v.AdmiralName...), v.AdmiralRank...), p.Ink)
	drawRight(fb, fnt, p.Admiral.Level.Right, p.Admiral.Level.Y,
		strconv.Itoa(v.AdmiralLevel), p.Ink)
	drawRight(fb, fnt, p.Admiral.Exp.Right, p.Admiral.Exp.Y,
		strconv.Itoa(v.AdmiralExp), p.Ink)
	drawFull(fb, fnt, p.Admiral.Flagship.X, p.Admiral.Flagship.Y, v.Flagship, p.Ink)

	if v.NoAdjutant {
		if b := p.NoAdjutant; b.W > 0 && b.H > 0 {
			fb.FillRect(b.X, b.Y, b.W, b.H, b.Fill)
		}
	} else {
		drawFull(fb, fnt, p.Adjutant.Name.X, p.Adjutant.Name.Y,
			append(append([]uint16(nil), v.AdjutantName...), v.AdjutantRank...), p.Ink)
		drawRight(fb, fnt, p.Adjutant.Level.Right, p.Adjutant.Level.Y,
			strconv.Itoa(v.AdjutantLevel), p.Ink)
		drawRight(fb, fnt, p.Adjutant.Exp.Right, p.Adjutant.Exp.Y,
			strconv.Itoa(v.AdjutantExp), p.Ink)
	}

	for _, s := range p.Stats {
		a, b := statAt(v.AdmiralRecord, s.Offset), statAt(v.AdjutantRecord, s.Offset)
		ink, val := p.Ink, a
		if b > a {
			ink, val = p.AltInk, b
		}
		drawRight(fb, fnt, s.Right, s.Y, strconv.Itoa(val), ink)
	}
}

func statAt(record []byte, off int) int {
	if off < 0 || off >= len(record) {
		return 0
	}
	return int(record[off])
}

// ButtonAt 回畫面座標 (x,y) 落在哪一顆按鈕上；沒有就回空字串。
func (p *FleetPanel) ButtonAt(x, y int) string {
	for _, b := range p.Buttons {
		if x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H {
			return b.Action
		}
	}
	return ""
}
