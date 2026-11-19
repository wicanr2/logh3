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
	actNone      sideAction = iota
	actBack                 // 右鍵
	actDigit                // 數字鍵，value 是 '0'–'9'
	actBackspace            // 退格
	actEnter                // 決定
	actGameClick            // 在原版座標 (gx,gy) 左鍵：側欄的命令列三鈕
	actScroll               // 戰場視野往 (gx,gy) 那一側捲三格
)

type sideButton struct {
	x, y, w, h int // 邏輯畫布座標
	codes      []uint16
	act        sideAction
	value      byte
	gx, gy     int
	gw         int // 原版按鈕寬（命令列三鈕）
}

// sideCmd 是原版命令列的一顆按鈕：字碼與它在原版畫面上的中心點。
type sideCmd struct {
	codes  []uint16
	gx, gy int
	w      int
}

const (
	codeTui  = 0xB068 // 退（battle-menu.json）
	codeHui  = 0xA65E // 回（screen-battle.json）
	codeGe   = 0xAEE6 // 格（flagships.json）
	codeJue  = 0xA84D // 決（screen-formation.json）
	codeDing = 0xA977 // 定（screen-formation.json）
	codeZero = 0xA2AF // ０（screen-battle.json，１–９ 依序 +1）
	// 方向箭頭：規格資料裡沒有，以原版 GRAPH.IMG 的字形核對（第176輪畫出點陣確認）。
	codeUp    = 0xA1F4 // ↑
	codeDown  = 0xA1F5 // ↓
	codeLeft  = 0xA1F6 // ←
	codeRight = 0xA1F7 // →
)

// 戰場視野 16..464 × 60..420（screen-battle.json 的 field：28×18 格、每格 16×20），
// 中心 (240,240)。捲動方向由按下點相對中心的方位決定（rules.ViewScrollDir），
// 所以四顆箭頭各取靠邊的一點。
var scrollPoints = map[uint16][2]int{
	codeUp: {240, 68}, codeDown: {240, 412}, codeLeft: {25, 240}, codeRight: {455, 240},
}

// scrollSteps 是一下箭頭走幾格（右鍵按下算一格，之後每次 RightHold 再一格）。
const scrollSteps = 3

const (
	colX  = 4  // 左欄按鈕左緣
	colW  = 80 // 左欄按鈕寬（四個全形字 64＋留白）
	btnH  = 30
	backY = 400 // 「退回」在兩種版面都放同一個位置
)

// sideButtons 回目前要顯示的側欄按鈕。
func sideButtons(lay screenLayout, state interface {
	isCopyProt() bool
	isBattle() bool
}, cmds []sideCmd) []sideButton {
	if lay.lw < 640+sidebarW { // 4:3 或更窄：沒有側欄空間
		return nil
	}
	out := []sideButton{{x: colX, y: backY, w: colW, h: btnH, codes: []uint16{codeTui, codeHui}, act: actBack}}
	if lay.strategic {
		for i, c := range cmds {
			out = append(out, sideButton{x: colX, y: 60 + i*(btnH+12), w: colW, h: btnH,
				codes: c.codes, act: actGameClick, gx: c.gx, gy: c.gy, gw: c.w})
		}
	}
	if state.isBattle() {
		const aw, ah, gap = 32, 30, 4 // 十字排列總寬 104，16:9 的右側帶（106）也放得下
		rightX := gameOrigin(lay.lw) + 640
		cx := rightX + (lay.lw-rightX)/2
		cy := 240
		for _, a := range []struct {
			code   uint16
			dx, dy int
		}{{codeUp, 0, -1}, {codeDown, 0, 1}, {codeLeft, -1, 0}, {codeRight, 1, 0}} {
			p := scrollPoints[a.code]
			out = append(out, sideButton{x: cx - aw/2 + a.dx*(aw+gap), y: cy - ah/2 + a.dy*(ah+gap), w: aw, h: ah,
				codes: []uint16{a.code}, act: actScroll, gx: p[0], gy: p[1]})
		}
	}
	if state.isCopyProt() {
		const kw, kh, gap = 32, 28, 4
		padW := 3*kw + 2*gap
		rightX := gameOrigin(lay.lw) + 640
		x0 := rightX + (lay.lw-rightX-padW)/2
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

// stateView 讓 sideButtons 不必依賴整個 Game（測試直接給值）。
type stateView struct{ copyProt, battle bool }

func (v stateView) isCopyProt() bool { return v.copyProt }
func (v stateView) isBattle() bool   { return v.battle }

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
		backClick(g)
	case actDigit:
		g.Key(b.value)
	case actBackspace:
		g.Key(8)
	case actEnter:
		g.Key('\r')
	case actGameClick:
		g.Click(b.gx, b.gy)
	case actScroll:
		// 照原版的「按住右鍵、游標在邊緣」走：按下一格，之後每次 RightHold 一格。
		// 選單開著或正在移動時，右鍵由遊戲照原規則處理（關選單等），不會誤捲。
		g.RightClick(b.gx, b.gy)
		for i := 1; i < scrollSteps; i++ {
			g.RightHold(b.gx, b.gy)
		}
		g.RightRelease()
	}
}

// sideStyle 是畫按鈕要的原版樣式。
type sideStyle struct {
	frame       ui.Frame
	ink, shadow uint8
	fnt         *font.Font
	cmds        []sideCmd
	menuPal     gfx.Palette // 還沒看過遊戲本體調色盤時（開頭動畫）先用主選單的
}

// newSideStyle 取原版命令列按鈕的樣式；資料缺了就回 nil（不畫側欄，不用代用品）。
func newSideStyle(res *game.Resources) *sideStyle {
	if res == nil || res.Font == nil || res.CapitalPanel == nil || res.CapitalPanel.Buttons == nil {
		return nil
	}
	b := res.CapitalPanel.Buttons
	st := &sideStyle{frame: ui.FrameOf(b.Fill, b.Order, b.Layers), ink: b.Ink, shadow: b.Shadow, fnt: res.Font}
	for _, it := range b.Items {
		st.cmds = append(st.cmds, sideCmd{codes: it.Codes, gx: it.X + b.W/2, gy: b.Y + b.H/2, w: b.W})
	}
	if res.Menu != nil {
		st.menuPal = res.Menu.Palette()
	}
	return st
}

// drawSide 把側欄按鈕畫到畫布上（u＝畫布像素／邏輯像素）。按鈕先以原版色號畫進
// 640×480 暫存畫面（gfx 只有這個尺寸），色號 0 當透明，轉成貼圖後依「字碼＋尺寸」
// 快取，調色盤一變就整批重建。
func (s *Shell) drawSide(screen *ebiten.Image, buttons []sideButton, pal [gfx.Colors]color.RGBA, u float64) {
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
		s.sharp.draw(screen, tex, tex.Bounds(), float64(b.x)*u, float64(b.y)*u, u)
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

// backClick 是一次「退回」＝右鍵按下再放開。
//
// 戰場上沒有任何選單或視窗開著時不送：原版在戰場空按右鍵是捲視野（第176輪預覽實測
// 會捲一格），不是退回；這時沒有東西可關。有東西開著時，右鍵照原規則關上一層，
// 位置取戰場中心。其他畫面用目前的游標位置（原版右鍵在哪裡按都是關上一層）。
// 一定補放開：只按不放，遊戲會一直認為右鍵按著（戰場的捲動狀態留著）。
func backClick(g *game.Game) {
	x, y := g.MousePosition()
	if g.State() == game.StateBattle {
		if g.BattleIdle() {
			return
		}
		x, y = 224, 220
	}
	g.RightClick(x, y)
	g.RightRelease()
}
