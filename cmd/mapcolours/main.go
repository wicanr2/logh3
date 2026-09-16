// Command mapcolours 從一張 oracle 收據的畫面反推戰略地圖每個星域塗的顏色。
//
//	tools/go.sh run ./cmd/mapcolours -orig org_game/extracted 收據1/screen.bin [收據2/screen.bin …]
//
// 原版塗色只碰底圖留白（色號 0）的像素（`docs/internals/12`）；這支把每個
// 星域遮罩內、底圖是 0 的像素在收據上讀出來，統計最多的那個色號。
// 星圖畫在畫面的 (0,40)。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
)

func main() {
	orig := flag.String("orig", "org_game/extracted", "原版素材目錄")
	flag.Parse()
	bit, err := os.ReadFile(filepath.Join(*orig, "STRMAP.BIT"))
	if err != nil {
		die(err)
	}
	rgn, err := os.ReadFile(filepath.Join(*orig, "STRMAP.RGN"))
	if err != nil {
		die(err)
	}
	base, err := asset.DecodeSheet(bit, 16, 20, 20)
	if err != nil {
		die(err)
	}
	regions, err := asset.OpenRGN(rgn, 16, 20)
	if err != nil {
		die(err)
	}
	for _, p := range flag.Args() {
		scr, err := os.ReadFile(p)
		if err != nil {
			die(err)
		}
		fmt.Printf("%s:", p)
		for i, r := range regions {
			var hist [16]int
			for y := 0; y < r.BH*20; y++ {
				gy := r.BY*20 + y
				for x := 0; x < r.BW*16; x++ {
					gx := r.BX*16 + x
					if gx >= base.W || gy >= base.H || !r.Covers(x, y, 16, 20) {
						continue
					}
					if base.Pix[gy*base.W+gx] != 0 {
						continue
					}
					hist[scr[(gy+40)*640+gx]&15]++
				}
			}
			best := 0
			for c := range hist {
				if hist[c] > hist[best] {
					best = c
				}
			}
			fmt.Printf(" %d:%d", i, best)
		}
		fmt.Println()
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
