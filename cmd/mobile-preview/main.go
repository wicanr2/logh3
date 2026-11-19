// Command mobile-preview 在桌面視窗裡跑 Android 版的殼（internal/mobileshell），
// 用來在桌面開發與檢查行動版版面（規格161）。輸入用 -taps 指定。
//
//	go run ./cmd/mobile-preview -orig /orig -shot out.png -after 240 -taps 120:533:240,170:513:210
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

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
	mid    map[int]string
	frames int
}

func save(screen *ebiten.Image, path string) {
	b := screen.Bounds()
	img := image.NewRGBA(b)
	screen.ReadPixels(img.Pix)
	if f, err := os.Create(path); err == nil {
		_ = png.Encode(f, img)
		f.Close()
	}
	fmt.Println("截圖", path)
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
	for t, path := range p.mid {
		if p.ticks >= t { // Update 與 Draw 不一一對應，取到了那一格之後的第一次 Draw
			save(screen, path)
			delete(p.mid, t)
		}
	}
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
	music := flag.String("music", "", "原版配樂與音效 OGG 目錄（可空）")
	shot := flag.String("shot", "", "跑到 -after 格之後截圖並結束")
	shots := flag.String("shots", "", "中途截圖「格數:檔名,…」（檔名放在 -shot 同目錄）")
	after := flag.Int("after", 120, "截圖前跑幾格")
	taps := flag.String("taps", "", "輕觸序列「格數:x:y,…」，座標是邏輯畫布（高 480）")
	flag.Parse()
	s := mobileshell.NewShell()
	s.Configure(mobileshell.Config{Orig: *orig, Data: *data, Save: os.TempDir(), Music: *music})
	p := &preview{Shell: s, shot: *shot, after: *after, taps: map[int][2]int{}, mid: map[int]string{}}
	for _, item := range strings.Split(*shots, ",") {
		var t int
		var name string
		if i := strings.IndexByte(item, ':'); i > 0 {
			if _, err := fmt.Sscanf(item[:i], "%d", &t); err == nil {
				name = item[i+1:]
				p.mid[t] = filepath.Join(filepath.Dir(*shot), name)
			}
		}
	}
	for _, item := range strings.Split(*taps, ",") {
		var t, x, y int
		if _, err := fmt.Sscanf(item, "%d:%d:%d", &t, &x, &y); err == nil {
			p.taps[t] = [2]int{x, y}
		}
	}
	ebiten.SetWindowSize(1200, 540) // 20:9
	if err := ebiten.RunGame(p); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
