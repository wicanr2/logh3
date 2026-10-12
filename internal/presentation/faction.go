package presentation

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wicanr2/logh3/internal/gfx"
)

const (
	themeUI uint8 = 1 << iota
	themeGlyph
	themePrimary
	themeEdge
	themeKeepRoles
)

type factionTheme struct {
	Surface, Button, Metal, White, Ink, Selected, SelectedInk, Primary color.RGBA
	MapVoid, MapRoute, MapBorder, MapLabelFill, MapLabelInk            color.RGBA
	MapStates                                                          [16]color.RGBA
	Gradient                                                           int
}

func loadFactionThemes(dir string) (map[string]*factionTheme, error) {
	data, err := os.ReadFile(filepath.Join(dir, "interface-themes.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var source struct {
		Schema int `json:"schema"`
		Themes map[string]struct {
			ApprovedFile                                          string `json:"approved_file"`
			ApprovedSHA                                           string `json:"approved_sha256"`
			Surface, Button, Metal, White, Ink, Selected, Primary string
			SelectedInk                                           string `json:"selected_ink"`
			Gradient                                              int
			MapVoid                                               string   `json:"map_void"`
			MapRoute                                              string   `json:"map_route"`
			MapBorder                                             string   `json:"map_border"`
			MapLabelFill                                          string   `json:"map_label_fill"`
			MapLabelInk                                           string   `json:"map_label_ink"`
			MapStates                                             []string `json:"map_states"`
		} `json:"themes"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		return nil, fmt.Errorf("HD陣營主題：%w", err)
	}
	if source.Schema != 1 || len(source.Themes) != 2 {
		return nil, fmt.Errorf("HD陣營主題格式無效")
	}
	result := make(map[string]*factionTheme)
	for _, id := range []string{"empire", "alliance"} {
		v, ok := source.Themes[id]
		if !ok || len(v.MapStates) != 16 || v.Gradient < 0 || v.Gradient > 16 {
			return nil, fmt.Errorf("HD陣營主題不完整：%s", id)
		}
		if _, err := checkedFile(dir, v.ApprovedFile, v.ApprovedSHA); err != nil {
			return nil, fmt.Errorf("HD定版介面：%w", err)
		}
		t := &factionTheme{Gradient: v.Gradient}
		pairs := []struct {
			src      string
			dst      *color.RGBA
			optional bool
		}{
			{v.Surface, &t.Surface, false}, {v.Button, &t.Button, false}, {v.Metal, &t.Metal, false}, {v.White, &t.White, false},
			{v.Ink, &t.Ink, true}, {v.Selected, &t.Selected, false}, {v.SelectedInk, &t.SelectedInk, true}, {v.Primary, &t.Primary, false},
			{v.MapVoid, &t.MapVoid, false}, {v.MapRoute, &t.MapRoute, false}, {v.MapBorder, &t.MapBorder, false},
			{v.MapLabelFill, &t.MapLabelFill, false}, {v.MapLabelInk, &t.MapLabelInk, false},
		}
		for i := range v.MapStates {
			pairs = append(pairs, struct {
				src      string
				dst      *color.RGBA
				optional bool
			}{v.MapStates[i], &t.MapStates[i], false})
		}
		for _, pair := range pairs {
			if pair.optional && pair.src == "" {
				continue
			}
			if len(pair.src) != 7 || pair.src[0] != '#' {
				return nil, fmt.Errorf("HD主題色格式無效：%s", pair.src)
			}
			n, err := strconv.ParseUint(pair.src[1:], 16, 24)
			if err != nil {
				return nil, err
			}
			*pair.dst = color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}
		}
		seen := map[color.RGBA]bool{}
		for _, c := range t.MapStates {
			seen[c] = true
		}
		if len(seen) != 16 {
			return nil, fmt.Errorf("HD星圖狀態色被合併：%s", id)
		}
		result[id] = t
	}
	return result, nil
}

func (r *Renderer) markUI(c gfx.DrawCommand, flags uint8) {
	if r.activeTheme == nil {
		return
	}
	rect := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H).Intersect(image.Rect(0, 0, gfx.Width, gfx.Height))
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		row := r.uiMask[y*gfx.Width+rect.Min.X : y*gfx.Width+rect.Max.X]
		for x := range row {
			oldRoles := uint8(0)
			// 核准概念的外框是各視窗與按鈕邊緣的聯集。
			// 重疊的UI仍保留亮邊；圖片覆蓋時才一起清除。
			if flags&themeUI != 0 {
				oldRoles = row[x] & themeEdge
			}
			if flags&themeKeepRoles != 0 {
				oldRoles |= row[x] & themePrimary
			}
			row[x] = flags&^themeKeepRoles | oldRoles
		}
	}
}

func (r *Renderer) markUIEdges(c gfx.DrawCommand) {
	if r.activeTheme == nil {
		return
	}
	for _, rect := range []image.Rectangle{
		image.Rect(c.X, c.Y, c.X+c.W, c.Y+2), image.Rect(c.X, c.Y+c.H-2, c.X+c.W, c.Y+c.H),
		image.Rect(c.X, c.Y, c.X+2, c.Y+c.H), image.Rect(c.X+c.W-2, c.Y, c.X+c.W, c.Y+c.H),
	} {
		rect = rect.Intersect(image.Rect(0, 0, gfx.Width, gfx.Height))
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				r.uiMask[y*gfx.Width+x] |= themeEdge
			}
		}
	}
}

func rgbKey(c color.RGBA) uint32 { return uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B) }

// 換色在有序重播完成後執行，保留定版概念的原字模、立體邊與抗鋸齒像素。
// mask依真實繪製及覆蓋更新；人物、地圖、動畫不受UI換色影響。
func (r *Renderer) applyFactionUI(fb *gfx.Framebuffer, p gfx.Palette) {
	t := r.activeTheme
	if t == nil {
		return
	}
	var colours [16]uint32
	for i := range colours {
		colours[i] = rgbKey(paletteColour(p, uint8(i)))
	}
	nearFill := func(i int, selectedOnly bool) bool {
		x, y := i%gfx.Width, i/gfx.Width
		for yy := max(0, y-2); yy < min(gfx.Height, y+3); yy++ {
			for xx := max(0, x-2); xx < min(gfx.Width, x+3); xx++ {
				c := colours[fb.Pix[yy*gfx.Width+xx]&15]
				if c == 0xe32020 || !selectedOnly && (c == 0x005161 || c == 0x415161) {
					return true
				}
			}
		}
		return false
	}
	var lights [gfx.Height * Scale]int
	for y := range lights {
		lights[y] = int(math.RoundToEven(float64(t.Gradient) * (1 - float64(y)/float64(len(lights)))))
	}
	for i, flags := range r.uiMask {
		if flags&themeUI == 0 {
			continue
		}
		x, y := i%gfx.Width, i/gfx.Width
		ink := t.White
		inkKnown := false
		for dy := 0; dy < Scale; dy++ {
			o := (y*Scale+dy)*r.frame.Stride + x*Scale*4
			for dx := 0; dx < Scale; dx++ {
				at := o + dx*4
				key := uint32(r.frame.Pix[at])<<16 | uint32(r.frame.Pix[at+1])<<8 | uint32(r.frame.Pix[at+2])
				var c color.RGBA
				switch key {
				case 0x005161:
					c = t.Surface
					light := lights[y*Scale+dy]
					c.R = uint8(min(255, int(c.R)+light))
					c.G = uint8(min(255, int(c.G)+light))
					c.B = uint8(min(255, int(c.B)+light))
				case 0x415161:
					c = t.Button
					if flags&themePrimary != 0 {
						c = t.Primary
					}
				case 0xb2b2c3:
					c = t.Metal
				case 0xf3f3f3:
					if !inkKnown {
						if flags&themeEdge == 0 {
							if t.Ink.A != 0 && nearFill(i, false) {
								ink = t.Ink
							} else if t.SelectedInk.A != 0 && nearFill(i, true) {
								ink = t.SelectedInk
							}
						}
						inkKnown = true
					}
					c = ink
				case 0xe32020:
					c = t.Selected
				default:
					continue
				}
				r.frame.Pix[at], r.frame.Pix[at+1], r.frame.Pix[at+2] = c.R, c.G, c.B
			}
		}
	}
}

func (r *Renderer) starSurfaceColour(id uint8, y int) color.RGBA {
	if r.activeTheme == nil {
		return surfaceColour(id, y)
	}
	c := r.activeTheme.MapStates[id&15]
	return c
}

func (r *Renderer) starForegroundColour(p gfx.Palette, id uint8, x, y int) color.RGBA {
	if r.activeTheme != nil && !(x < 48 && y < 48) && !(x > 280 && y > 270) {
		switch id & 15 {
		case 8, 9, 11:
			return r.activeTheme.MapRoute
		case 15:
			return r.activeTheme.MapBorder
		}
	}
	return paletteColour(p, id)
}
