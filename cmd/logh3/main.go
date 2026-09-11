// Command logh3 把一個畫面畫成色號陣列、調色盤與 PNG。
//
// 目前只有主選單。輸出格式刻意與 dosgolem 的 `-dump-ega`／`-dump-palette`
// 相同，這樣 `cmd/oradiff` 可以直接逐位元組比。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

func main() {
	fontPath := flag.String("font", "", "原版字型檔路徑（remake 不內嵌原版素材，必填）")
	framePath := flag.String("frame", "", "原版 FRAME.BLK 路徑（外框圖塊；不給就找字型檔旁邊那一份）")
	specPath := flag.String("spec", "docs/internals/data/screen-main-menu.json", "畫面規格 JSON")
	out := flag.String("out", "workplace/remake", "輸出目錄")
	hl := flag.Int("highlight", -2, "反白哪一列（-1 ＝ 不反白；不給就用規格的預設）")
	flag.Parse()

	if *fontPath == "" {
		fmt.Fprintln(os.Stderr, "缺 -font：remake 不內嵌原版字型，請指到本機的原版檔")
		os.Exit(2)
	}
	fnt, err := font.Load(*fontPath)
	if err != nil {
		die(err)
	}
	spec, err := screen.LoadMenuSpec(*specPath)
	if err != nil {
		die(err)
	}
	// 外框是 `FRAME.BLK` 的圖塊，也是原版素材。缺了就不畫框——remake
	// 不內嵌原版素材，也不用代用品頂替。
	if *framePath == "" {
		*framePath = filepath.Join(filepath.Dir(*fontPath), "FRAME.BLK")
	}
	var tiles *ui.FrameTiles
	if b, err := os.ReadFile(*framePath); err != nil {
		fmt.Fprintf(os.Stderr, "讀不到 %s：不畫視窗外框\n", *framePath)
	} else if tiles, err = ui.LoadFrameTiles(b); err != nil {
		die(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}
	var fb *gfx.Framebuffer
	var missing []uint16
	if *hl == -2 {
		fb, missing = screen.MainMenu(spec, fnt, tiles)
	} else {
		fb, missing = screen.MainMenuWith(spec, fnt, tiles, *hl)
	}
	for _, c := range missing {
		fmt.Fprintf(os.Stderr, "缺字：Big5 %04X\n", c)
	}
	pal := spec.Palette()
	if err := fb.WriteIndex(filepath.Join(*out, "screen.bin")); err != nil {
		die(err)
	}
	if err := pal.WritePalette(filepath.Join(*out, "pal.bin")); err != nil {
		die(err)
	}
	if err := fb.WritePNG(filepath.Join(*out, "screen.png"), pal); err != nil {
		die(err)
	}
	fmt.Printf("寫出 %s 的 screen.bin / pal.bin / screen.png\n", *out)
	if len(missing) > 0 {
		os.Exit(1)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "錯誤:", err)
	os.Exit(1)
}
