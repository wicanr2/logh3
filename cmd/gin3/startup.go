package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/localize"
	"github.com/wicanr2/logh3/internal/presentation"
)

// 來源及完整授權見 startup_font.LICENSE.txt、tools/scripts/startup_font.py。
//
//go:embed startup_font.png
var startupFontPNG []byte

//go:embed startup_font.json
var startupFontJSON []byte

type startup struct {
	load     func(func(int, string)) (*app, error)
	scale    float64
	once     sync.Once
	mu       sync.Mutex
	step     int
	label    string
	loaded   *app
	err      error
	closed   bool
	game     *app
	texture  *ebiten.Image
	last     string
	font     image.Image
	glyphs   map[string][4]int
	language *localize.Catalog
}

func newStartup(load func(func(int, string)) (*app, error), scale float64) *startup {
	s := &startup{load: load, scale: scale, label: "準備載入"}
	s.font, _ = png.Decode(bytes.NewReader(startupFontPNG))
	_ = json.Unmarshal(startupFontJSON, &s.glyphs)
	ebiten.SetWindowSize(560, 220)
	ebiten.SetWindowTitle("銀河英雄傳說III SP 載入中")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetCursorMode(ebiten.CursorModeVisible)
	return s
}

func (s *startup) report(step int, label string) {
	s.mu.Lock()
	s.step, s.label = step, label
	s.mu.Unlock()
	fmt.Fprintf(os.Stderr, "載入 %d/7：%s\n", step, label)
}

func (s *startup) begin() {
	go func() {
		a, err := s.load(s.report)
		if err != nil {
			fmt.Fprintln(os.Stderr, "載入失敗：", err)
		}
		s.mu.Lock()
		s.loaded, s.err = a, err
		closed := s.closed
		s.mu.Unlock()
		if closed && a != nil {
			a.close()
		}
	}()
}

func (s *startup) Update() error {
	if s.game != nil {
		return s.game.Update()
	}
	s.once.Do(s.begin)
	s.mu.Lock()
	a, err := s.loaded, s.err
	s.mu.Unlock()
	if a != nil {
		s.game = a
		a.prepareWindow(s.scale)
		return nil // 完成當格不把載入期間的按鍵或點擊送進遊戲。
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if err != nil {
			return err
		}
		return ebiten.Termination
	}
	if err != nil && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if image.Pt(x, y).In(image.Rect(440, 175, 530, 205)) {
			return err
		}
	}
	return nil
}

func (s *startup) close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	a := s.loaded
	s.mu.Unlock()
	if a != nil {
		a.close()
	}
}

func (a *app) close() {
	if a.effects != nil {
		a.effects.Stop()
	}
	if a.music != nil {
		a.music.Stop()
	}
	if a.frameTrace != nil {
		_ = a.frameTrace.Close()
	}
}

func (a *app) prepareWindow(scale float64) {
	a.screen = ebiten.NewImage(gfx.Width, gfx.Height)
	if a.hd != nil {
		ebiten.SetScreenFilterEnabled(false)
		a.hdScreen = ebiten.NewImage(gfx.Width*presentation.Scale, gfx.Height*presentation.Scale)
	}
	w, h := int(float64(gfx.Width)*scale), int(float64(gfx.Height)*scale)
	if a.posterMode {
		w = posterWindowWidth(h, true)
	}
	ebiten.SetWindowSize(w, h)
	a.windowReady = true
	ebiten.SetWindowTitle("銀河英雄傳說III SP")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
}

func (s *startup) Layout(w, h int) (int, int) {
	if s.game != nil {
		return s.game.Layout(w, h)
	}
	return 560, 220
}

func (s *startup) text(dst *image.RGBA, x, y int, text string, ink color.RGBA, maxWidth int) {
	if s.language != nil && s.language.Locale != "zh-TW" {
		if localize.DrawRGBA(dst, s.language.Locale, s.language.Translate(text), image.Rect(x, y, x+maxWidth, y+23), ink, 16) {
			return
		}
	}
	start := x
	for _, char := range text {
		glyph, ok := s.glyphs[string(char)]
		if !ok {
			glyph = s.glyphs["□"]
		}
		if x+glyph[2] > start+maxWidth {
			break
		}
		bounds := image.Rect(x, y, x+glyph[2], y+glyph[3])
		draw.DrawMask(dst, bounds, &image.Uniform{C: ink}, image.Point{}, s.font, image.Pt(glyph[0], glyph[1]), draw.Over)
		x += glyph[2]
	}
}

func (s *startup) frame(step int, label string, err error) *image.RGBA {
	fb := image.NewRGBA(image.Rect(0, 0, 560, 220))
	base, panel, gold := color.RGBA{15, 29, 49, 255}, color.RGBA{27, 51, 79, 255}, color.RGBA{172, 147, 91, 255}
	ink := color.RGBA{255, 239, 204, 255}
	fill := func(r image.Rectangle, c color.RGBA) { draw.Draw(fb, r, &image.Uniform{C: c}, image.Point{}, draw.Src) }
	fill(fb.Bounds(), base)
	fill(image.Rect(12, 12, 548, 208), gold)
	fill(image.Rect(13, 13, 547, 207), panel)
	s.text(fb, 30, 25, "銀河英雄傳說 III SP", ink, 500)
	if err != nil {
		s.text(fb, 30, 65, "載入失敗，請檢查資料與路徑", ink, 500)
		runes := []rune(err.Error())
		for i := 0; i < 3 && len(runes) > 0; i++ {
			n, width := 0, 0
			for n < len(runes) && width+22 <= 500 {
				w := 22
				if runes[n] < 128 {
					w = 11
				}
				width += w
				n++
			}
			s.text(fb, 30, 92+i*24, string(runes[:n]), ink, 500)
			runes = runes[n:]
		}
		fill(image.Rect(440, 175, 530, 202), gold)
		fill(image.Rect(441, 176, 529, 201), base)
		s.text(fb, 462, 176, "結束", ink, 60)
		return fb
	}
	s.text(fb, 30, 70, label, ink, 500)
	s.text(fb, 30, 108, fmt.Sprintf("階段 %d / 7", step), ink, 500)
	fill(image.Rect(30, 146, 530, 160), base)
	fill(image.Rect(31, 147, 31+498*min(max(step-1, 0), 7)/7, 159), gold)
	s.text(fb, 30, 175, "正在載入，請稍候", ink, 500)
	return fb
}

func (s *startup) Draw(screen *ebiten.Image) {
	if s.game != nil {
		s.game.Draw(screen)
		return
	}
	s.mu.Lock()
	step, label, err := s.step, s.label, s.err
	s.mu.Unlock()
	key := fmt.Sprintf("%d/%s/%v", step, label, err)
	if s.texture == nil {
		s.texture = ebiten.NewImage(560, 220)
	}
	if key != s.last {
		s.last = key
		s.texture.WritePixels(s.frame(step, label, err).Pix)
	}
	screen.DrawImage(s.texture, nil)
}
