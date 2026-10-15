package font

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
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
