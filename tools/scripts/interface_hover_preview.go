//go:build ignore

// 規格138：從正常Game入口產生介面配色試樣。
package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	art := flag.String("art", "assets/hd", "試樣素材")
	out := flag.String("out", "", "新輸出目錄")
	flag.Parse()
	if *out == "" {
		panic("缺少out")
	}
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	r, err := bootstrap.Load("/orig", "docs/internals/data")
	if err != nil {
		panic(err)
	}
	hd, err := presentation.Load(*art, r.Faces)
	if err != nil {
		panic(err)
	}
	r.Faces.ArtworkPrefix = "face"
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	click := func(x, y int) { g.Click(x, y); g.LeftRelease(); g.Frame(); g.MoveMouse(x, y) }
	snap := func(name string) {
		g.MoveMouse(600, 420)
		if err := hd.WritePNG(filepath.Join(*out, name+".png"), g.Frame(), g.Palette()); err != nil {
			panic(err)
		}
		if err := g.Frame().WritePNG(filepath.Join(*out, name+"-original.png"), g.Palette()); err != nil {
			panic(err)
		}
	}
	click(300, 210)
	for _, c := range []byte("138") {
		g.Key(c)
	}
	b := r.CopyProt.InputBox
	click(b.X+b.W/2, b.Y+b.H/2)
	snap("account")
	click(448, 350)
	snap("budget")
	click(463, 370)
	click(448, 370)
	click(448, 370)
	snap("strategy")
}
