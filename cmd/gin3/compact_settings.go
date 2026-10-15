package main

import (
	"fmt"
	"image"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

var compactSettingsBounds = image.Rect(160, 41, 464, 400)

type compactControl struct {
	ID, Label string
	Bounds    image.Rectangle
	Legacy    image.Point
}

// 規格147：同一份資料驅動繪製、點擊及唯讀GUI觀測。
var compactControls = func() []compactControl {
	var controls []compactControl
	add := func(id, label string, r image.Rectangle, x, y int) {
		controls = append(controls, compactControl{id, label, r, image.Pt(x, y)})
	}
	add("font", "", image.Rect(288, 68, 448, 84), 0, 0)
	add("hd", "HD", image.Rect(288, 86, 338, 102), 0, 0)
	add("fm", "FM", image.Rect(342, 86, 392, 102), 368, 110)
	add("off", "關閉", image.Rect(396, 86, 448, 102), 416, 110)
	for i, s := range []string{"波琍璐", "英雄交響曲", "悲愴交響曲", "來自新世界", "四季．夏", "沈默的戰場"} {
		add(fmt.Sprintf("bgm-%d", i), s, image.Rect(288, 104+i*14, 448, 118+i*14), 352, 130+i*20)
	}
	for i, s := range []string{"全艦攻擊", "決戰前夕", "使用原有的BGM"} {
		add(fmt.Sprintf("battle-%d", i), s, image.Rect(288, 190+i*14, 448, 204+i*14), 352, 250+i*20)
	}
	add("msg-inc", "+", image.Rect(400, 234, 422, 250), 408, 310)
	add("msg-dec", "-", image.Rect(426, 234, 448, 250), 424, 310)
	add("anim-on", "開啟", image.Rect(288, 252, 366, 268), 368, 330)
	add("anim-off", "關閉", image.Rect(370, 252, 448, 268), 416, 330)
	for i, s := range []string{"普通", "快速", "最快"} {
		add(fmt.Sprintf("speed-%d", i), s, image.Rect(288+i*54, 270, 340+i*54, 286), 320+i*48, 350)
	}
	for i, s := range []string{"soundfont", "enemies", "ai", "flagship", "movement"} {
		add(s, "", image.Rect(288, 288+i*18, 448, 304+i*18), 0, 0)
	}
	add("close", "結束", image.Rect(360, 378, 448, 394), 0, 0)
	return controls
}()

func compactControlBounds(id string) image.Rectangle {
	for _, c := range compactControls {
		if c.ID == id {
			return c.Bounds
		}
	}
	return image.Rectangle{}
}

func aiSettingsRow(level int) image.Rectangle {
	return image.Rect(288, 244+(level+1)*16, 448, 260+(level+1)*16)
}

var compactRuneCodes = map[rune]uint16{
	'世': 0xA540, '交': 0xA5E6, '介': 0xA4B6, '使': 0xA8CF, '來': 0xA8D3, '保': 0xAB4F, '全': 0xA5FE, '則': 0xAB68, '前': 0xAB65, '加': 0xA55B,
	'動': 0xB0CA, '化': 0xA4C6, '原': 0xADEC, '名': 0xA657, '啟': 0xB1D2, '四': 0xA57C, '型': 0xABAC, '場': 0xB3F5, '夏': 0xAE4C, '夕': 0xA469,
	'字': 0xA672, '存': 0xA673, '季': 0xA975, '宋': 0xA7BA, '定': 0xA977, '將': 0xB14E, '平': 0xA5AD, '度': 0xABD7, '庫': 0xAE77, '強': 0xB16A,
	'快': 0xA7D6, '思': 0xABE4, '息': 0xAEA7, '悲': 0xB464, '愴': 0xB75B, '成': 0xA6A8, '戰': 0xBED4, '戲': 0xC0B8, '擊': 0xC0BB, '攻': 0xA7F0,
	'效': 0xAEC4, '敵': 0xBCC4, '新': 0xB773, '旗': 0xBA58, '時': 0xAEC9, '普': 0xB4B6, '曲': 0xA6B1, '最': 0xB3CC, '有': 0xA6B3, '束': 0xA7F4,
	'果': 0xAA47, '樂': 0xBCD6, '決': 0xA84D, '沈': 0xA848, '法': 0xAA6B, '波': 0xAA69, '源': 0xB7BD, '無': 0xB54C, '版': 0xAAA9, '琍': 0xB27C,
	'璐': 0xED69, '用': 0xA5CE, '界': 0xACC9, '畫': 0xB565, '的': 0xAABA, '示': 0xA5DC, '移': 0xB2BE, '結': 0xB5B2, '自': 0xA6DB, '艦': 0xC4A5,
	'色': 0xA6E2, '英': 0xAD5E, '規': 0xB357, '訊': 0xB054, '設': 0xB35D, '軍': 0xAD78, '轉': 0xC2E0, '通': 0xB371, '速': 0xB374, '遊': 0xB943,
	'配': 0xB074, '閉': 0xB3AC, '開': 0xB67D, '關': 0xC3F6, '陣': 0xB07D, '雄': 0xB6AF, '面': 0xADB1, '音': 0xADB5, '響': 0xC554, '顯': 0xC5E3,
	'體': 0xC5E9, '鬥': 0xB0AB, '黑': 0xB6C2, '默': 0xC071, '點': 0xC249, '．': 0xA144,
}

func compactTextWidth(text string) int {
	w := 0
	for _, c := range text {
		if c < 128 {
			w += 6
		} else {
			w += 12
		}
	}
	return w
}

func (a *app) compactText(fb *gfx.Framebuffer, r image.Rectangle, text string, ink uint8) {
	x, y := r.Min.X+(r.Dx()-compactTextWidth(text))/2, r.Min.Y+(r.Dy()-12)/2
	for _, c := range text {
		if glyph, ok := font.SettingsGlyph(c); ok {
			start := fb.BeginGroup()
			for py := 0; py < 12; py++ {
				for px := 0; px < glyph.Width; px++ {
					if glyph.Native[py*glyph.Width+px] >= 88 && image.Pt(x+px, y+py).In(r) {
						fb.Set(x+px, y+py, ink)
					}
				}
			}
			fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, Key: font.SettingsGlyphKey,
				X: x, Y: y, W: glyph.Width, H: 12, Colour: ink, Clip: r,
				SourceW: glyph.Width * 3, SourceH: 36, Source: glyph.HD})
			x += glyph.Width
			continue
		}
		code, width, sourceW := uint16(c), 12, 16
		var rows []uint16
		if c < 128 {
			width, sourceW = 6, 8
			for _, b := range a.toolbar.font.Half(byte(c)) {
				rows = append(rows, uint16(b)<<8)
			}
		} else {
			code = compactRuneCodes[c]
			rows = a.toolbar.font.Glyph(code)
		}
		start := fb.BeginGroup()
		paint := func(dx, dy int, colour uint8) {
			for py := 0; py < 12; py++ {
				if len(rows) == 0 {
					continue
				}
				bits := rows[py*len(rows)/12]
				for px := 0; px < width; px++ {
					if bits&(1<<uint(15-px*sourceW/width)) != 0 && image.Pt(x+px+dx, y+py+dy).In(r) {
						fb.Set(x+px+dx, y+py+dy, colour)
					}
				}
			}
		}
		paint(-1, -1, 3)
		paint(0, 0, ink)
		fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: x, Y: y, W: width, H: 12, Code: code, Half: c < 128, Colour: ink, Shadow: 3, ShadowDY: -1, Clip: r})
		x += width
	}
}

func aiSettingsLabel(level int) string {
	if level < 0 {
		return "原版規則"
	}
	return fmt.Sprintf("強化 %d", level)
}

func (a *app) compactSettingsClick(x, y int) bool {
	p := image.Pt(x, y)
	if a.aiSettingsMenu {
		for level := -1; level <= 3; level++ {
			if p.In(aiSettingsRow(level)) {
				_ = a.g.SetAILevel(level)
				break
			}
		}
		a.aiSettingsMenu = false
		return true
	}
	if a.soundFontMenu {
		for i, choice := range soundFontChoices {
			if p.In(soundFontMenuRow(i)) && a.soundFontLibrary(choice) != nil {
				if err := a.selectSoundFont(choice, true); err != nil {
					a.soundFontError = err.Error()
				} else {
					a.soundFontMenu = false
				}
				return true
			}
		}
		a.soundFontMenu = false
		return true
	}
	for _, c := range compactControls {
		if !p.In(c.Bounds) {
			continue
		}
		switch c.ID {
		case "font":
			if a.hd != nil {
				styles := a.hd.FontStyles()
				for i, s := range styles {
					if s == a.hd.FontStyle() {
						_ = a.hd.SetFontStyle(styles[(i+1)%len(styles)])
						break
					}
				}
			}
		case "hd":
			if a.soundFontLibrary(a.selectedSoundFont()) != nil {
				if a.g.MusicSettings().Device == 2 {
					a.g.Click(368, 110)
					_ = a.setHDMusic(true)
				} else {
					_ = a.setHDMusic(!a.hdMusic)
				}
			}
		case "soundfont":
			a.soundFontMenu = true
		case "enemies":
			a.g.ToggleEnemyVisibility()
		case "ai":
			a.aiSettingsMenu = true
		case "flagship", "movement":
			a.ruleSettingsClick(p)
		case "close":
			a.g.RightClick(x, y)
		default:
			if c.ID == "fm" {
				_ = a.setHDMusic(false)
			}
			a.g.Click(c.Legacy.X, c.Legacy.Y)
		}
		a.g.MoveMouse(x, y)
		return true
	}
	return true // 設定頁不讓新座標穿透舊熱區。
}

func (a *app) drawCompactSettings(fb *gfx.Framebuffer) {
	// 金屬外緣讓深色介面也能辨識左框，內層保留右上光源的陰影。
	(ui.Frame{Role: "ui/window", Layers: []ui.Layer{
		{Top: 15, Right: 15, Bottom: 15, Left: 15},
		{Top: 7, Right: 7, Bottom: 3, Left: 3},
	}, Fill: 9}).Draw(fb, compactSettingsBounds.Min.X, compactSettingsBounds.Min.Y, compactSettingsBounds.Dx(), compactSettingsBounds.Dy())
	a.compactText(fb, image.Rect(176, 48, 448, 64), "遊戲設定", 7)
	for _, l := range []struct {
		text string
		y    int
	}{{"介面字型", 68}, {"音源", 86}, {"平時配樂", 104}, {"戰鬥配樂", 190}, {"訊息速度", 234}, {"動畫效果", 252}, {"動畫速度", 270}, {"音色庫", 288}, {"顯示敵軍", 306}, {"AI強化", 324}, {"轉移旗艦", 342}, {"名將移動加成", 360}} {
		a.compactText(fb, image.Rect(176, l.y, 284, l.y+16), l.text, 7)
	}
	music, display, options := a.g.MusicSettings(), a.g.DisplaySettings(), a.g.RuleOptions()
	button := func(r image.Rectangle, label string, selected, enabled bool) {
		fill, ink := uint8(9), uint8(7)
		if selected {
			fill = 2
			ink = 3
		}
		if !enabled {
			ink = 8
		}
		(ui.Frame{Role: "ui/button", Layers: []ui.Layer{{Top: 7, Right: 7, Bottom: 3, Left: 3}}, Fill: fill}).Draw(fb, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
		a.compactText(fb, r.Inset(1), label, ink)
		if enabled && image.Pt(a.g.MousePosition()).In(r) {
			fb.HoverControl = r
		}
	}
	for _, c := range compactControls {
		label, selected, enabled := c.Label, false, true
		switch c.ID {
		case "font":
			style := "original"
			if a.hd != nil {
				style = a.hd.FontStyle()
			}
			label = map[string]string{"original": "原版點陣", "classic": "原版強化", "serif": "思源宋體", "sans": "思源黑體"}[style]
			enabled = a.hd != nil
		case "hd":
			selected = a.hdMusic && music.Device != 2
			enabled = a.soundFontLibrary(a.selectedSoundFont()) != nil
		case "fm":
			selected = !a.hdMusic && music.Device == 1
		case "off":
			selected = music.Device == 2
		case "anim-on":
			selected = display.Anim == 1
		case "anim-off":
			selected = display.Anim == 0
		case "soundfont":
			label = soundFontLabels[a.selectedSoundFont()]
		case "enemies":
			selected = a.g.ShowEnemies()
		case "ai":
			label = aiSettingsLabel(a.g.AILevel())
		case "flagship":
			selected = options.FlagshipTransfer
		case "movement":
			selected = options.HeroMovement
		default:
			for i := 0; i < 6; i++ {
				if c.ID == fmt.Sprintf("bgm-%d", i) {
					selected = music.BGM == i
				}
			}
			for i := 0; i < 3; i++ {
				if c.ID == fmt.Sprintf("battle-%d", i) {
					selected = music.Battle == i
				}
				if c.ID == fmt.Sprintf("speed-%d", i) {
					selected = display.AnimSpeed == i
				}
			}
		}
		if c.ID == "enemies" || c.ID == "flagship" || c.ID == "movement" {
			label = "關閉"
			if selected {
				label = "開啟"
			}
		}
		button(c.Bounds, label, selected, enabled)
	}
	a.compactText(fb, image.Rect(288, 234, 396, 250), fmt.Sprint(display.MsgSpeed), 7)
	if a.ruleSettingsError != "" || a.soundFontError != "" {
		a.compactText(fb, image.Rect(176, 378, 352, 394), "無法設定", 7)
	}
	if a.soundFontMenu {
		for i, s := range soundFontChoices {
			button(soundFontMenuRow(i), soundFontLabels[s], s == a.selectedSoundFont(), a.soundFontLibrary(s) != nil)
		}
	}
	if a.aiSettingsMenu {
		for level := -1; level <= 3; level++ {
			button(aiSettingsRow(level), aiSettingsLabel(level), level == a.g.AILevel(), true)
		}
	}
}

func compactSettingsObservation() map[string][4]int {
	result := map[string][4]int{"bounds": {160, 41, 464, 400}, "glyph": {12, 12, 6, 12}}
	for _, c := range compactControls {
		r := c.Bounds
		result[c.ID] = [4]int{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y}
	}
	for level := -1; level <= 3; level++ {
		r := aiSettingsRow(level)
		result[fmt.Sprintf("ai-choice-%d", level)] = [4]int{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y}
	}
	for i := range soundFontChoices {
		r := soundFontMenuRow(i)
		result[fmt.Sprintf("soundfont-choice-%d", i)] = [4]int{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y}
	}
	return result
}
