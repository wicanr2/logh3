package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/game"
	"image"
	"image/color"
	"image/draw"
	"time"
)

const toolbarHeight = 32

var toolbarButtons = [2]image.Rectangle{image.Rect(8, 5, 152, 27), image.Rect(160, 5, 364, 27)}
var helpBounds = image.Rect(64, 110, 576, 374)
var helpCloseBounds = image.Rect(260, 336, 380, 360)

type toolbar struct {
	visible, help, captured bool
	hover                   int
	font                    *font.Font
	pixels                  *image.RGBA
	texture                 *ebiten.Image
	lastKey                 string
}

func (t *toolbar) move(x, y int) {
	inside := x >= 0 && x < 640 && y >= 0
	t.visible = t.help || inside && (y < 6 || t.visible && y < toolbarHeight)
	t.hover = t.actionAt(x, y)
}
func (t *toolbar) blocks(x, y int) bool {
	return t.help || t.visible && image.Pt(x, y).In(image.Rect(0, 0, 640, toolbarHeight))
}
func (t *toolbar) actionAt(x, y int) int {
	p := image.Pt(x, y)
	if t.help && p.In(helpCloseBounds) {
		return 4
	}
	if t.visible {
		for i, b := range toolbarButtons {
			if p.In(b) {
				return i + 1
			}
		}
	}
	return 0
}
func (a *app) toolbarAction(action int) {
	switch action {
	case 1:
		a.toolbar.help = !a.toolbar.help
	case 2:
		if !a.toggleGraphics() {
			fmt.Println("F7：HD素材不可用，沿用原版顯示")
		}
	case 4:
		a.toolbar.help = false
	}
}
func (a *app) updateInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		a.toolbarAction(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF7) || inpututil.IsKeyJustPressed(ebiten.KeyF6) {
		a.toolbarAction(2)
	}
	closedHelp := a.toolbar.help && inpututil.IsKeyJustPressed(ebiten.KeyEscape)
	if closedHelp {
		a.toolbar.help = false
	}
	x, y := ebiten.CursorPosition()
	x, y = a.gamePosition(x, y)
	t := &a.toolbar
	t.move(x, y)
	blocked := t.blocks(x, y) || t.captured || closedHelp
	if blocked {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			t.captured = true
			a.toolbarAction(t.actionAt(x, y))
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			t.captured = true
		}
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			t.captured = false
		}
		// 功能列攔截輸入，不能點穿到底下；系統指標取代被藏到畫面外的遊戲游標。
		a.g.MoveMouse(-100, -100)
		a.g.LeftRelease()
		a.g.RightRelease()
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
		return
	}
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	a.g.MoveMouse(x, y)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if a.presentationClick(x, y) {
			a.g.LeftRelease()
			return
		}
		a.g.Click(x, y)
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if a.g.State() == game.StateSettings {
			a.g.LeftRelease()
		} else {
			a.g.LeftHold(x, y, time.Second/time.Duration(ebiten.TPS()))
		}
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		a.g.LeftRelease()
	}
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) && (a.soundFontMenu || a.aiSettingsMenu):
		a.soundFontMenu = false
		a.aiSettingsMenu = false
		a.g.RightRelease()
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		a.g.RightClick(x, y)
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight):
		a.g.RightHold(x, y)
	case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight):
		a.g.RightRelease()
	}
	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			a.g.Key(byte('0' + k - ebiten.Key0))
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && a.soundFontMenu {
		a.soundFontMenu = false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && a.aiSettingsMenu {
		a.aiSettingsMenu = false
	}
	for k := ebiten.KeyKP0; k <= ebiten.KeyKP9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			a.g.Key(byte('0' + k - ebiten.KeyKP0))
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		a.g.Key('\r')
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		a.g.Key(8)
	}
}
func (a *app) drawToolbar(screen *ebiten.Image, theme string) {
	t := &a.toolbar
	if !t.visible && !t.help {
		return
	}
	key := fmt.Sprintf("%t/%t/%d/%t/%s", t.visible, t.help, t.hover, a.original, theme)
	if t.pixels == nil {
		t.pixels = image.NewRGBA(image.Rect(0, 0, 640, 480))
		t.texture = ebiten.NewImage(640, 480)
	}
	if t.lastKey != key {
		t.lastKey = key
		clear(t.pixels.Pix)
		base, button, edge, ink := color.RGBA{20, 42, 70, 255}, color.RGBA{35, 63, 96, 255}, color.RGBA{170, 148, 96, 255}, color.RGBA{255, 240, 201, 255}
		if theme == "alliance" {
			base = color.RGBA{45, 31, 22, 255}
			button = color.RGBA{104, 69, 35, 255}
			edge = color.RGBA{185, 144, 80, 255}
			ink = color.RGBA{255, 237, 202, 255}
		}
		fill := func(b image.Rectangle, c color.RGBA) {
			draw.Draw(t.pixels, b, &image.Uniform{C: c}, image.Point{}, draw.Src)
		}
		panel := func(b image.Rectangle, c color.RGBA) { fill(b, edge); fill(b.Inset(1), c) }
		label := func(b image.Rectangle, text string) {
			w := 0
			for _, c := range text {
				if c < 128 {
					w += 8
				} else {
					w += 16
				}
			}
			t.text(b.Min.X+(b.Dx()-w)/2, b.Min.Y+(b.Dy()-15)/2, text, ink)
		}
		buttonAt := func(b image.Rectangle, text string, action int, enabled bool) {
			c := button
			if !enabled {
				c = base
			}
			panel(b, c)
			if enabled && t.hover == action {
				fill(b.Inset(1), color.RGBA{111, 105, 77, 255})
				fill(image.Rect(b.Min.X+1, b.Min.Y+1, b.Max.X-1, b.Min.Y+3), ink)
			}
			label(b, text)
		}
		if t.visible {
			fill(image.Rect(0, 0, 640, toolbarHeight), base)
			fill(image.Rect(0, toolbarHeight-1, 640, toolbarHeight), edge)
			mode := "HD"
			if a.original || a.hd == nil {
				mode = "原版"
			}
			buttonAt(toolbarButtons[0], "F1 說明", 1, true)
			buttonAt(toolbarButtons[1], "F7 原版 / HD: "+mode, 2, a.hd != nil)
		}
		if t.help {
			draw.Draw(t.pixels, image.Rect(0, toolbarHeight, 640, 480), &image.Uniform{C: color.RGBA{0, 0, 0, 150}}, image.Point{}, draw.Over)
			panel(helpBounds, base)
			for i, text := range []string{"操作說明", "F1  開關說明", "F7  原版 / HD", "顯示敵軍: 設定選單", "滑鼠左鍵: 選擇與確認", "滑鼠右鍵: 返回或取消", "滑鼠移到頂端: 顯示功能列", "敵軍顯示只影響畫面"} {
				t.text(92, 126+i*25, text, ink)
			}
			buttonAt(helpCloseBounds, "關閉 Esc", 4, true)
		}
		t.texture.WritePixels(t.pixels.Pix)
	}
	op := &ebiten.DrawImageOptions{}
	if a.hd != nil {
		op.GeoM.Scale(3, 3)
	}
	screen.DrawImage(t.texture, op)
}

// 字碼由Python的Big5編碼表產生；字模仍取玩家自己的原版資料。
var toolbarRuneCodes = map[rune]uint16{'設': 0xB35D, '定': 0xA977, '單': 0xB3E6, '不': 0xA4A3, '作': 0xA740, '列': 0xA643, '到': 0xA8EC, '功': 0xA55C, '原': 0xADEC, '取': 0xA8FA, '只': 0xA575, '可': 0xA569, '右': 0xA56B, '回': 0xA65E, '左': 0xA5AA, '影': 0xBC76, '或': 0xA9CE, '擇': 0xBEDC, '操': 0xBEDE, '敵': 0xBCC4, '明': 0xA9FA, '材': 0xA7F7, '沿': 0xAA75, '消': 0xAEF8, '滑': 0xB7C6, '版': 0xAAA9, '用': 0xA5CE, '畫': 0xB565, '確': 0xBD54, '示': 0xA5DC, '移': 0xB2BE, '端': 0xBADD, '素': 0xAFC0, '能': 0xAFE0, '與': 0xBB50, '認': 0xBB7B, '說': 0xBBA1, '軍': 0xAD78, '返': 0xAAF0, '選': 0xBFEF, '鍵': 0xC1E4, '閉': 0xB3AC, '開': 0xB67D, '關': 0xC3F6, '面': 0xADB1, '響': 0xC554, '頂': 0xB3BB, '顯': 0xC5E3, '鼠': 0xB9AB}

func (t *toolbar) text(x, y int, text string, ink color.RGBA) {
	if t.font == nil {
		return
	}
	for _, c := range text {
		width := 16
		var rows []uint16
		if c < 128 {
			width = 8
			for _, v := range t.font.Half(byte(c)) {
				rows = append(rows, uint16(v)<<8)
			}
		} else {
			rows = t.font.Glyph(toolbarRuneCodes[c])
		}
		for dy, bits := range rows {
			for dx := 0; dx < width; dx++ {
				if bits&(1<<uint(15-dx)) != 0 {
					t.pixels.SetRGBA(x+dx, y+dy, ink)
				}
			}
		}
		x += width
	}
}
