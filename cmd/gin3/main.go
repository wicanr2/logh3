// Command gin3 是 remake 的可玩入口：Ebiten 開視窗、把 internal/game 的
// 畫面畫出來、把滑鼠與鍵盤送回去。
//
// 遊戲邏輯與繪製都在 internal/，這一層只做三件事：把色號陣列展開成 RGBA、
// 把視窗座標換成遊戲座標、把輸入轉成 game 的動作。**原版素材由玩家自備**
// （`-orig`），本專案不內嵌任何原版檔。
//
//	tools/gui.sh run ./cmd/gin3 -orig /orig
//	tools/gui.sh run ./cmd/gin3 -orig /orig -shot /src/workplace/gui.png -shot-after 8
package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
)

type app struct {
	g      *game.Game
	pal    [gfx.Colors]color.RGBA
	screen *ebiten.Image
	buf    []byte

	shot      string
	shotBin   string
	shotAfter int
	ticks     int

	// script 是開場先送進去的輸入（與 cmd/play 同一套語法）。
	// 給無頭驗收用：同一條玩家路徑走完再倒圖，才比得出顯示層有沒有畫錯。
	script string
	fed    bool
}

func (a *app) Update() error {
	if !a.fed {
		a.fed = true
		if err := feed(a.g, a.script); err != nil {
			return err
		}
	}
	if a.script != "" {
		// 走腳本的時候不吃真的滑鼠鍵盤——不然 Xvfb 的游標會混進來。
		a.ticks++
		if (a.shot != "" || a.shotBin != "") && a.ticks >= a.shotAfter {
			return a.dump()
		}
		return nil
	}
	// 視窗就是 640×480 的邏輯畫布，Game 也吃畫面座標，所以直接傳。
	x, y := ebiten.CursorPosition()
	a.g.MoveMouse(x, y)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		a.g.Click(x, y)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		a.g.RightClick(x, y)
	}
	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			a.g.Key(byte('0' + (k - ebiten.Key0)))
		}
	}
	for k := ebiten.KeyKP0; k <= ebiten.KeyKP9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			a.g.Key(byte('0' + (k - ebiten.KeyKP0)))
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		a.g.Key('\r')
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		a.g.Key(8)
	}
	a.ticks++
	if (a.shot != "" || a.shotBin != "") && a.ticks >= a.shotAfter {
		return a.dump()
	}
	return nil
}

// dump 把畫面倒出來就結束。`.bin` 是色號陣列，格式與 dosgolem 的
// `-dump-ega` 相同，可以直接與收據逐位元組比。
func (a *app) dump() error {
	fb := a.g.Frame()
	if a.shot != "" {
		if err := fb.WritePNG(a.shot, a.paletteSpec()); err != nil {
			return err
		}
	}
	if a.shotBin != "" {
		if err := fb.WriteIndex(a.shotBin); err != nil {
			return err
		}
	}
	return ebiten.Termination
}

// paletteSpec 回畫 PNG 用的調色盤（6 位元，與 gfx 的慣例相同）。
func (a *app) paletteSpec() gfx.Palette {
	var p gfx.Palette
	for i := range p {
		c := a.pal[i]
		p[i] = [3]uint8{c.R >> 2, c.G >> 2, c.B >> 2}
	}
	return p
}

func (a *app) Draw(screen *ebiten.Image) {
	fb := a.g.Frame()
	for i, v := range fb.Pix {
		c := a.pal[v]
		a.buf[i*4], a.buf[i*4+1], a.buf[i*4+2], a.buf[i*4+3] = c.R, c.G, c.B, 255
	}
	a.screen.WritePixels(a.buf)
	screen.DrawImage(a.screen, nil)
}

func (a *app) Layout(int, int) (int, int) { return gfx.Width, gfx.Height }

func main() {
	orig := flag.String("orig", "org_game/extracted", "原版檔案目錄（玩家自備）")
	data := flag.String("data", "docs/internals/data", "畫面規格與資料目錄")
	scale := flag.Float64("scale", 1, "視窗放大倍率")
	shot := flag.String("shot", "", "跑幾格之後把畫面倒成 PNG 就結束（無頭驗收用）")
	script := flag.String("script", "", "開場先送進去的輸入，分號分隔："+
		"move X Y／click [X Y]／rclick [X Y]／type 數字／enter／back")
	shotBin := flag.String("shot-bin", "", "同 -shot，但倒的是色號陣列（可與收據逐位元組比）")
	shotAfter := flag.Int("shot-after", 4, "倒圖前先跑幾格")
	// **預設是原版**：原版按「實行」時電腦不下任何命令（`docs/re/79`），
	// 對拍收據都靠這個前提。強化版要玩家自己開（`docs/spec/017`）。
	aiLevel := flag.Int("ai", -1, "電腦 AI：-1 ＝ 原版（不動）；0–3 ＝ 強化版，"+
		"數字越大越兇（會挑目標、集結、包圍）")
	flag.Parse()

	res, err := bootstrap.Load(*orig, *data)
	if err != nil {
		die(err)
	}
	g, err := game.New(res)
	if err != nil {
		die(err)
	}
	if *aiLevel >= 0 {
		g.SetBrain(&ai.Focused{Difficulty: *aiLevel})
		g.SetTactic(&ai.Encircle{Difficulty: *aiLevel}, false)
	} else {
		// 預設走原版：戰術層整批朝同一點壓上去（`docs/re/80`）。
		g.SetTactic(ai.OriginalTactic{}, false)
	}
	a := &app{
		g:         g,
		screen:    ebiten.NewImage(gfx.Width, gfx.Height),
		buf:       make([]byte, gfx.Width*gfx.Height*4),
		shot:      *shot,
		shotBin:   *shotBin,
		shotAfter: *shotAfter,
		script:    *script,
	}
	pal := res.Menu.Palette()
	for i := range a.pal {
		r, gg, b := pal.RGB8(i)
		a.pal[i] = color.RGBA{r, gg, b, 255}
	}
	ebiten.SetWindowSize(int(float64(gfx.Width)*(*scale)), int(float64(gfx.Height)*(*scale)))
	ebiten.SetWindowTitle("銀河英雄傳說III SP")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(a); err != nil && err != ebiten.Termination {
		die(err)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// feed 把腳本送進遊戲。語法與 cmd/play 相同。
func feed(g *game.Game, script string) error {
	for _, cmd := range strings.Split(script, ";") {
		f := strings.Fields(strings.TrimSpace(cmd))
		if len(f) == 0 {
			continue
		}
		xy := func() (int, int) {
			if len(f) == 3 {
				x, _ := strconv.Atoi(f[1])
				y, _ := strconv.Atoi(f[2])
				return x, y
			}
			return g.Mouse()
		}
		switch f[0] {
		case "move":
			x, _ := strconv.Atoi(f[1])
			y, _ := strconv.Atoi(f[2])
			g.MoveMouse(x, y)
		case "click":
			g.Click(xy())
		case "rclick":
			g.RightClick(xy())
		case "type":
			for _, c := range []byte(f[1]) {
				g.Key(c)
			}
		case "enter":
			g.Key('\r')
		case "back":
			g.Key(8)
		default:
			return fmt.Errorf("看不懂的指令：%q", cmd)
		}
	}
	return nil
}
