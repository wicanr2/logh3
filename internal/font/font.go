// Package font 讀原版格式的中文點陣字型。
//
// 格式與索引的證據見 `docs/internals/06-font-graph-img.md`：半形 8×15、
// 全形 16×15、每列 2 bytes 高位在前。
//
// ⚠ **remake 不內嵌原版字型。** 那份字型是倚天中文系統的，路徑由呼叫端
// 傳入；找不到就回錯誤，不用代用品——安靜的替代品會讓「還沒做完」
// 看起來像做完了。
package font

import (
	"fmt"
	"os"
)

const (
	// GlyphW／GlyphH 是全形字的點陣尺寸。
	GlyphW, GlyphH = 16, 15
	// HalfW 是半形字的寬度。
	HalfW = 8

	fullStride = GlyphH * 2 // 全形一個字 30 bytes
	halfStride = GlyphH     // 半形一個字 15 bytes

	// 三個區塊的起點，見 docs/internals/06。
	halfBase   = 0
	symbolBase = 3840
	hanziBase  = 27030
	// hanziIndex 是 Big5 0xA440（第一個漢字）的序號。
	hanziIndex = 471
)

// Font 是一份載入好的字型。
type Font struct {
	data []byte
}

// Load 讀一份原版格式的字型檔。
func Load(path string) (*Font, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("字型讀不到（remake 不內嵌原版字型，請用旗標指到本機的原版檔）: %w", err)
	}
	if len(b) < hanziBase+fullStride {
		return nil, fmt.Errorf("字型檔太小（%d bytes），不像原版格式", len(b))
	}
	return &Font{data: b}, nil
}

// Big5Index 是 Big5 字碼的序號，0xA140 為 0。
func Big5Index(code uint16) int {
	lead, trail := int(code>>8), int(code&0xFF)
	var t int
	if trail <= 0x7E {
		t = trail - 0x40
	} else {
		t = trail - 0xA1 + 63
	}
	return (lead-0xA1)*157 + t
}

// offsetOf 回全形字碼在檔案裡的位移。
func offsetOf(code uint16) int {
	i := Big5Index(code)
	if i < hanziIndex {
		return symbolBase + i*fullStride
	}
	return hanziBase + (i-hanziIndex)*fullStride
}

// Glyph 回全形字的 15 列點陣，每列一個 16 位元值，bit15 在最左。
// 字碼不合法或超出檔案範圍時回 nil——**回空白會讓缺字看起來像空格**，
// 那正是最難發現的一種錯。
func (f *Font) Glyph(code uint16) []uint16 {
	lead, trail := code>>8, code&0xFF
	if lead < 0xA1 || lead > 0xF9 || trail < 0x40 || trail == 0x7F || trail > 0xFE {
		return nil
	}
	o := offsetOf(code)
	if o < 0 || o+fullStride > len(f.data) {
		return nil
	}
	rows := make([]uint16, GlyphH)
	for r := range rows {
		rows[r] = uint16(f.data[o+2*r])<<8 | uint16(f.data[o+2*r+1])
	}
	return rows
}

// Half 回半形字（CP437 排列）的 15 列點陣，每列一個 8 位元值，bit7 在最左。
func (f *Font) Half(code uint8) []uint8 {
	o := halfBase + int(code)*halfStride
	if o+halfStride > len(f.data) {
		return nil
	}
	rows := make([]uint8, GlyphH)
	copy(rows, f.data[o:o+halfStride])
	return rows
}
