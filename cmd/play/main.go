// Command play 走一遍玩家路徑，把每一步的畫面倒成 PNG。
//
// 這是 remake 的「可操作」入口：沒有顯示層也能一步步把輸入送進去、
// 看每一步畫成什麼。腳本用分號分隔，例如
//
//	move 300 170; click; type 138; enter
//
// 原版素材由玩家自備（`-orig`），本專案不內嵌任何原版檔。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

func main() {
	orig := flag.String("orig", "org_game/extracted", "原版檔案目錄（玩家自備）")
	data := flag.String("data", "docs/internals/data", "畫面規格與資料目錄")
	out := flag.String("out", "workplace/play", "輸出目錄")
	palPath := flag.String("pal", "", "調色盤檔（dosgolem 的 -dump-palette 格式，256×3）；"+
		"不給就用主選單規格裡那 16 色。遊戲本體的調色盤與主選單不同")
	script := flag.String("script", "move 300 170; click; type 138; enter",
		"輸入腳本，分號分隔：move X Y／click [X Y]／type 數字／enter／back")
	flag.Parse()

	res, err := bootstrap.Load(*orig, *data)
	if err != nil {
		die(err)
	}
	g, err := game.New(res)
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}
	pal := res.Menu.Palette()
	if *palPath != "" {
		b, err := os.ReadFile(*palPath)
		if err != nil {
			die(err)
		}
		if len(b) < len(pal)*3 {
			die(fmt.Errorf("%s 只有 %d bytes，要 %d", *palPath, len(b), len(pal)*3))
		}
		// dosgolem 的 -dump-palette 是**8 位元**的值，gfx.Palette 存的是
		// DAC 的 6 位元（畫的時候才用 (v<<2)|(v>>4) 展開）。不換回 6 位元
		// 就會展開兩次，顏色整個偏掉——而色號陣列的對拍看不出來。
		for i := range pal {
			pal[i][0], pal[i][1], pal[i][2] = b[i*3]>>2, b[i*3+1]>>2, b[i*3+2]>>2
		}
	}
	step := 0
	dump := func(label string) {
		fb := g.Frame()
		name := fmt.Sprintf("%02d-%s", step, label)
		if err := fb.WriteIndex(filepath.Join(*out, name+".bin")); err != nil {
			die(err)
		}
		if err := fb.WritePNG(filepath.Join(*out, name+".png"), pal); err != nil {
			die(err)
		}
		fmt.Printf("%2d  %-16s 狀態 %s\n", step, label, stateName(g.State()))
		step++
	}
	dump("start")
	for _, cmd := range strings.Split(*script, ";") {
		f := strings.Fields(strings.TrimSpace(cmd))
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "move":
			if len(f) != 3 {
				die(fmt.Errorf("move 要兩個座標：%q", cmd))
			}
			x, _ := strconv.Atoi(f[1])
			y, _ := strconv.Atoi(f[2])
			g.MoveMouse(x, y)
		case "click":
			g.Click(mouseOf(f, g))
		case "rclick":
			// 右鍵 ＝ 關掉最上面那一層（月初報告要兩下）。
			g.RightClick(mouseOf(f, g))
		case "type":
			if len(f) != 2 {
				die(fmt.Errorf("type 要一串數字：%q", cmd))
			}
			for _, c := range []byte(f[1]) {
				g.Key(c)
			}
		case "enter":
			g.Key('\r')
		case "back":
			g.Key(8)
		default:
			die(fmt.Errorf("看不懂的指令：%q", cmd))
		}
		dump(strings.ReplaceAll(strings.TrimSpace(cmd), " ", "-"))
	}
}

// mouseOf 讓 `click` 可以自己帶座標，也可以沿用上一次 move 的位置。
func mouseOf(f []string, g *game.Game) (int, int) {
	if len(f) == 3 {
		x, _ := strconv.Atoi(f[1])
		y, _ := strconv.Atoi(f[2])
		return x, y
	}
	return g.Mouse()
}

func stateName(s game.State) string {
	switch s {
	case game.StateCopyProtection:
		return "防拷問答"
	case game.StateMainGame:
		return "遊戲本體"
	case game.StateShipOrder:
		return "船艦訂造報告"
	case game.StateAdmiralList:
		return "admiral-list"
	case game.StateBudgetPlan:
		return "budget-plan"
	case game.StateBuildPlan:
		return "船艦建造計劃"
	default:
		return "主選單"
	}
}

// die 印出錯誤就結束。**缺原版檔就明講缺哪一個**，不用代用品。
func die(err error) {
	fmt.Fprintln(os.Stderr, "錯誤：", err)
	os.Exit(1)
}
