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
	"github.com/wicanr2/logh3/internal/screen"
)

func main() {
	fontPath := flag.String("font", "", "原版字型檔路徑（remake 不內嵌原版素材，必填）")
	specPath := flag.String("spec", "docs/internals/data/screen-main-menu.json", "畫面規格 JSON")
	out := flag.String("out", "workplace/remake", "輸出目錄")
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
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}
	fb, missing := screen.MainMenu(spec, fnt)
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
