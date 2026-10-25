package localize

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"sync"
)

const GlyphKey = "ui/localized/"

type GlyphRect struct{ X, Y, W, H int }

var fontOnce sync.Once
var fontErr error
var glyphs map[string]map[string]GlyphRect
var pages map[int]*image.Gray
var big5 map[uint16]rune

func prepareFont() {
	fontOnce.Do(func() {
		var config struct {
			Glyphs map[string]map[string]GlyphRect
		}
		b, err := embedded.ReadFile("font.json")
		if err == nil {
			err = json.Unmarshal(b, &config)
		}
		if err != nil {
			fontErr = err
			return
		}
		glyphs = config.Glyphs
		b, err = embedded.ReadFile("big5.json")
		if err == nil {
			err = json.Unmarshal(b, &big5)
		}
		if err != nil {
			fontErr = err
			return
		}
		pages = map[int]*image.Gray{}
		for _, size := range []int{12, 36} {
			name := "font-12.png"
			if size == 36 {
				name = "font-36.png"
			}
			b, err = embedded.ReadFile(name)
			if err != nil {
				fontErr = err
				return
			}
			im, err := png.Decode(bytes.NewReader(b))
			if err != nil {
				fontErr = err
				return
			}
			pages[size] = im.(*image.Gray)
		}
	})
}

func Mask(locale string, ch rune, size int) (*image.Gray, GlyphRect, bool) {
	prepareFont()
	if fontErr != nil {
		return nil, GlyphRect{}, false
	}
	key := locale + "/" + decimal(ch)
	entry, ok := glyphs[key]
	if !ok {
		return nil, GlyphRect{}, false
	}
	s := "12"
	if size == 36 {
		s = "36"
	}
	r, ok := entry[s]
	return pages[size], r, ok
}

func decimal(ch rune) string {
	if ch == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for ch > 0 {
		i--
		b[i] = byte(ch%10) + '0'
		ch /= 10
	}
	return string(b[i:])
}

func Decode(code uint16, half bool, width int) string {
	prepareFont()
	if half {
		if width == 16 {
			return string([]rune{rune(code >> 8), rune(code & 255)})
		}
		return string(rune(code))
	}
	if ch, ok := big5[code]; ok {
		return string(ch)
	}
	return ""
}
