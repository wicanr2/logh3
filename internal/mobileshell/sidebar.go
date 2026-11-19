package mobileshell

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 側欄控制（規格161 M1）：畫在原版 640×480 畫面左右兩條，中央原版畫面不疊任何東西。
//
// 外觀沿用原版命令列按鈕（capital-panel.json 的 buttons：底色 8、兩層立體邊 7／3、
// 字色 7、陰影 3），字用原版點陣字。字碼都取自已核對的規格資料：
// 全形數字 ０–９ 0xA2AF–0xA2B8（screen-battle.json）、決定（screen-formation.json）、
// 退（battle-menu.json）、回（screen-battle.json）、格（flagships.json）。

type sideAction int

const (
	actNone  sideAction = iota
	actBack             // 右鍵
	actDigit            // 數字鍵，value 是 '0'–'9'
	actBackspace
	actEnter
)

type sideButton struct {
	x, y, w, h int // 邏輯畫布座標
	codes      []uint16
	act        sideAction
	value      byte
}

const (
	codeTui  = 0xB068 // 退（battle-menu.json）
	codeHui  = 0xA65E // 回（screen-battle.json）
	codeGe   = 0xAEE6 // 格（flagships.json）
	codeJue  = 0xA84D // 決（screen-formation.json）
	codeDing = 0xA977 // 定（screen-formation.json）
	codeZero = 0xA2AF // ０（screen-battle.json，１–９ 依序 +1）
)

// sideButtons 回目前要顯示的側欄按鈕。bandW 是單側寬度。
func sideButtons(logicalW int, state game.State) []sideButton {
	bandW := gameOrigin(logicalW)
	if bandW < 40 {
		return nil // 4:3 或更窄的螢幕沒有側欄
	}
	rightX := gameOrigin(logicalW) + gfx.Width
	var out []sideButton
	// 左側：退回（＝右鍵），放在下方拇指好按的位置。
	bw := 64
	out = append(out, sideButton{x: (bandW - bw) / 2, y: 400, w: bw, h: 28,
		codes: []uint16{codeTui, codeHui}, act: actBack})
	// 右側：防拷問答時出數字鍵盤。
	if state == game.StateCopyProtection {
		const kw, kh, gap = 32, 28, 4
		padW := 3*kw + 2*gap
		x0 := rightX + (bandW-padW)/2
		y0 := 150
		for i := 1; i <= 9; i++ {
			r, c := (i-1)/3, (i-1)%3
			out = append(out, sideButton{x: x0 + c*(kw+gap), y: y0 + r*(kh+gap), w: kw, h: kh,
				codes: []uint16{codeZero + uint16(i)}, act: actDigit, value: byte('0' + i)})
		}
		out = append(out, sideButton{x: x0 + kw + gap, y: y0 + 3*(kh+gap), w: kw, h: kh,
			codes: []uint16{codeZero}, act: actDigit, value: '0'})
		ww := (padW - gap) / 2
		y := y0 + 4*(kh+gap) + gap
		out = append(out,
			sideButton{x: x0, y: y, w: ww, h: kh, codes: []uint16{codeTui, codeGe}, act: actBackspace},
			sideButton{x: x0 + ww + gap, y: y, w: ww, h: kh, codes: []uint16{codeJue, codeDing}, act: actEnter})
	}
	return out
}

// sideButtonAt 回落在 (x,y)（邏輯畫布座標）的按鈕。
func sideButtonAt(buttons []sideButton, x, y int) (sideButton, bool) {
	for _, b := range buttons {
		if x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h {
			return b, true
		}
	}
	return sideButton{}, false
}

// runSide 執行一顆側欄按鈕。
func runSide(g *game.Game, b sideButton) {
	switch b.act {
	case actBack:
		x, y := g.MousePosition()
		g.RightClick(x, y)
	case actDigit:
		g.Key(b.value)
	case actBackspace:
		g.Key(8)
	case actEnter:
		g.Key('\r')
	}
}

// sideStyle 是畫按鈕要的原版樣式。
type sideStyle struct {
	frame       ui.Frame
	ink, shadow uint8
	fnt         *font.Font
}

// newSideStyle 取原版命令列按鈕的樣式；資料缺了就回 nil（不畫側欄，不用代用品）。
func newSideStyle(res *game.Resources) *sideStyle {
	if res == nil || res.Font == nil || res.CapitalPanel == nil || res.CapitalPanel.Buttons == nil {
		return nil
	}
	b := res.CapitalPanel.Buttons
	return &sideStyle{frame: ui.FrameOf(b.Fill, b.Order, b.Layers), ink: b.Ink, shadow: b.Shadow, fnt: res.Font}
}

// drawSide 把側欄按鈕畫到 screen 上。按鈕先以原版色號畫進一張 640×480 的暫存
// 畫面（gfx 只有這個尺寸），色號 0 當透明，再依調色盤轉色成貼圖。
// 貼圖依「字碼＋尺寸」快取，調色盤一變就整批重建。
func (s *Shell) drawSide(screen *ebiten.Image, buttons []sideButton, pal [gfx.Colors]color.RGBA) {
	if len(buttons) == 0 || s.style == nil {
		return
	}
	if s.sideCache == nil || s.sidePal != pal {
		for _, t := range s.sideCache {
			t.Deallocate()
		}
		s.sideCache, s.sidePal = map[string]*ebiten.Image{}, pal
	}
	for _, b := range buttons {
		key := fmt.Sprint(b.codes, b.w, b.h)
		tex := s.sideCache[key]
		if tex == nil {
			tex = s.renderButton(b, pal)
			s.sideCache[key] = tex
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(b.x), float64(b.y))
		screen.DrawImage(tex, op)
	}
}

func (s *Shell) renderButton(b sideButton, pal [gfx.Colors]color.RGBA) *ebiten.Image {
	fb := gfx.NewFramebuffer(0)
	s.style.frame.Draw(fb, 0, 0, b.w, b.h)
	textW := len(b.codes) * font.GlyphW
	cx, cy := (b.w-textW)/2, (b.h-font.GlyphH)/2
	for i, code := range b.codes {
		ui.DrawGlyphClipped(fb, s.style.fnt, cx+i*font.GlyphW, cy, code,
			s.style.ink, s.style.shadow, -1, false, 0, 0, b.w, b.h)
	}
	img := image.NewRGBA(image.Rect(0, 0, b.w, b.h))
	for y := 0; y < b.h; y++ {
		for x := 0; x < b.w; x++ {
			if v := fb.Pix[y*gfx.Width+x]; v != 0 {
				img.SetRGBA(x, y, pal[v])
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}
