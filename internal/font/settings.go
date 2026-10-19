package font

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"sync"
)

// SettingsGlyphKey僅供系統設定使用，不替換遊戲的原版字庫。
const SettingsGlyphKey = "ui/settings-text"

// 開放字庫來源、完整授權及產生方法見settings_font.SOURCE.json與LICENSE.txt。
//
//go:embed settings_font_12.png
var settingsNative []byte

//go:embed settings_font_36.png
var settingsHD []byte

//go:embed settings_font.json
var settingsIndex []byte

// SettingsMask的遮罩已預先載入。呼叫端只能讀取，不得修改切片。
type SettingsMask struct {
	Width  int
	Native []uint8
	HD     []uint8
}

var settingsGlyphs = loadSettingsGlyphs()

var settingsSized sync.Map

// 編輯頁沿核准草稿的10／12／14／16字格。來源仍為同一份設定字型。
func SettingsGlyphSize(char rune, size int) (SettingsMask, bool) {
	base, ok := SettingsGlyph(char)
	if !ok || size < 1 || size > 24 {
		return SettingsMask{}, false
	}
	if size == 12 {
		return base, true
	}
	key := struct {
		C    rune
		Size int
	}{char, size}
	if cached, ok := settingsSized.Load(key); ok {
		return cached.(SettingsMask), true
	}
	w := size
	if char < 128 {
		w = size / 2
	}
	resize := func(src []uint8, sw, sh, dw, dh int) []uint8 {
		dst := make([]uint8, dw*dh)
		for y := 0; y < dh; y++ {
			for x := 0; x < dw; x++ {
				dst[y*dw+x] = src[(y*sh/dh)*sw+x*sw/dw]
			}
		}
		return dst
	}
	mask := SettingsMask{Width: w, Native: resize(base.Native, base.Width, 12, w, size), HD: resize(base.HD, base.Width*3, 36, w*3, size*3)}
	settingsSized.Store(key, mask)
	return mask, true
}

func loadSettingsGlyphs() map[rune]SettingsMask {
	var index struct {
		Chars   string `json:"chars"`
		Columns int    `json:"columns"`
	}
	if err := json.Unmarshal(settingsIndex, &index); err != nil {
		panic(err)
	}
	decode := func(data []byte) image.Image {
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		return im
	}
	lo, hi := decode(settingsNative), decode(settingsHD)
	result := make(map[rune]SettingsMask)
	for i, char := range []rune(index.Chars) {
		width := 12
		if char < 128 {
			width = 6
		}
		mask := func(im image.Image, scale int) []uint8 {
			w, h := width*scale, 12*scale
			x0, y0 := i%index.Columns*h, i/index.Columns*h
			out := make([]uint8, w*h)
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					out[y*w+x] = color.GrayModel.Convert(im.At(x0+x, y0+y)).(color.Gray).Y
				}
			}
			return out
		}
		result[char] = SettingsMask{Width: width, Native: mask(lo, 1), HD: mask(hi, 3)}
	}
	return result
}

func SettingsGlyph(char rune) (SettingsMask, bool) {
	g, ok := settingsGlyphs[char]
	return g, ok
}
