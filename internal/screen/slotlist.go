package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// SlotList 是存檔畫面那 11 列的版面（規格 `docs/spec/008`）。
//
// 每一列的內容都從 `SAVEDATA.IDX` 算出來，不是量好畫死的：沒用到的格子
// 印破折號與「未使用」，有內容的印存檔時間、序號、遊戲內日期與名稱。
type SlotList struct {
	Rows     int   `json:"rows"`
	Top      int   `json:"top"`
	Height   int   `json:"height"`
	UsedInk  uint8 `json:"used_ink"`
	EmptyInk uint8 `json:"empty_ink"`

	StampX int `json:"stamp_x"`
	TimeX  int `json:"time_x"`
	HashX  int `json:"hash_x"`
	IndexX int `json:"index_x"`
	DateX  int `json:"date_x"`
	NameX  int `json:"name_x"`
	EmptyX int `json:"empty_x"`

	HashCode   uint16   `json:"hash_code"`
	IndexCodes []uint16 `json:"index_codes"`
	EmptyCodes []uint16 `json:"empty_codes"`
	EmptyStamp string   `json:"empty_stamp"`
	EmptyTime  string   `json:"empty_time"`
}

// LoadSlotList 讀存檔畫面的版面。
func LoadSlotList(path string) (*SlotList, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s SlotList
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &s, nil
}

// Draw 把存檔目錄畫成那 11 列。`epoch` 是遊戲內日期的起始年（796）。
func (s *SlotList) Draw(fb *gfx.Framebuffer, fnt *font.Font,
	entries []rules.SaveIndexEntry, epoch int) {
	for n := 0; n < s.Rows && n < len(entries); n++ {
		e := entries[n]
		y := s.Top + n*s.Height
		if !e.Used {
			drawASCII(fb, fnt, s.StampX, y, s.EmptyStamp, s.EmptyInk)
			drawASCII(fb, fnt, s.TimeX, y, s.EmptyTime, s.EmptyInk)
			drawFull(fb, fnt, s.EmptyX, y, s.EmptyCodes, s.EmptyInk)
			continue
		}
		yy, mm, dd, hh, mi, ss := e.StampFields()
		drawASCII(fb, fnt, s.StampX, y,
			fmt.Sprintf("%02d/%02d/%02d", yy, mm, dd), s.UsedInk)
		drawASCII(fb, fnt, s.TimeX, y,
			fmt.Sprintf("%02d:%02d:%02d", hh, mi, ss), s.UsedInk)
		drawFull(fb, fnt, s.HashX, y, []uint16{s.HashCode}, s.UsedInk)
		if i := e.Num - 1; i >= 0 && i < len(s.IndexCodes) {
			drawFull(fb, fnt, s.IndexX, y, s.IndexCodes[i:i+1], s.UsedInk)
		}
		gy, gm, gd := rules.Date(e.Turn, epoch)
		drawASCII(fb, fnt, s.DateX, y,
			fmt.Sprintf("%3d/%02d/%02d", gy, gm, gd), s.UsedInk)
		drawFull(fb, fnt, s.NameX, y, e.Name, s.UsedInk)
	}
}

// drawASCII 從 x 起一格一格畫半形字（一格 8 px）。
func drawASCII(fb *gfx.Framebuffer, fnt *font.Font, x, y int, s string, ink uint8) {
	for i := 0; i < len(s); i++ {
		drawHalf(fb, fnt, x+i*font.HalfW, y, s[i], ink)
	}
}
