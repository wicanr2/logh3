// Command mobile-preview 在桌面視窗裡跑 Android 版的殼（internal/mobileshell），
// 用來在桌面開發與檢查行動版版面（規格161）。滑鼠左鍵當觸控輕觸無效，請用 -tap。
//
//	go run ./cmd/mobile-preview -orig /orig -data docs/internals/data -shot out.png -after 120
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/logh3/internal/mobileshell"
)

type preview struct {
	*mobileshell.Shell
	shot   string
	after  int
	ticks  int
	w, h   int
	taps   map[int][2]int
	frames int
}

func (p *preview) Update() error {
	if err := p.Shell.Update(); err != nil {
		return err
	}
	p.ticks++
	if t, ok := p.taps[p.ticks]; ok {
		p.Shell.TapLogical(t[0], t[1])
	}
	return nil
}

func (p *preview) Draw(screen *ebiten.Image) {
	p.Shell.Draw(screen)
	p.frames++
	if p.shot != "" && p.ticks >= p.after {
		b := screen.Bounds()
		img := image.NewRGBA(b)
		screen.ReadPixels(img.Pix)
		f, err := os.Create(p.shot)
		if err == nil {
			_ = png.Encode(f, img)
			f.Close()
		}
		fmt.Println("截圖", p.shot, b.Dx(), b.Dy(), "ticks", p.ticks)
		os.Exit(0)
	}
}

func main() {
	orig := flag.String("orig", "org_game/extracted", "原版檔目錄")
	data := flag.String("data", "docs/internals/data", "資料目錄")
	shot := flag.String("shot", "", "跑到 -after 格之後截圖並結束")
	after := flag.Int("after", 120, "截圖前跑幾格")
	tapAt := flag.Int("tap-at", 0, "第幾格送一次輕觸（邏輯座標 -tap-x、-tap-y）")
	tapX := flag.Int("tap-x", 0, "")
	tapY := flag.Int("tap-y", 0, "")
	flag.Parse()
	s := mobileshell.NewShell()
	s.Configure(mobileshell.Config{Orig: *orig, Data: *data, Save: os.TempDir()})
	p := &preview{Shell: s, shot: *shot, after: *after, taps: map[int][2]int{}}
	if *tapAt > 0 {
		p.taps[*tapAt] = [2]int{*tapX, *tapY}
	}
	ebiten.SetWindowSize(1200, 540) // 20:9
	if err := ebiten.RunGame(p); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
