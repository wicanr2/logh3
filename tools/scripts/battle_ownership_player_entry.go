//go:build ignore

// 正常主選單讀S3與首戰的同版本Game基準；不注入原池或直接進戰場。
package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

func main() {
	out := flag.String("out", "", "新的Game入口輸出目錄")
	flag.Parse()
	if *out == "" {
		panic("缺少輸出目錄")
	}
	if _, err := os.Stat(*out); !os.IsNotExist(err) {
		panic("拒絕覆寫Game基準")
	}
	r, err := bootstrap.Load("org_game/extracted", "docs/internals/data")
	must(err)
	g, err := game.New(r)
	must(err)
	must(g.SetAILevel(-1))
	click := func(x, y int) { g.MoveMouse(x, y); g.Click(x, y) }
	click(300, 270)
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	for _, p := range [][2]int{{300, 322}, {318, 208}, {544, 125}, {560, 125}} {
		click(p[0], p[1])
	}
	for i := 0; i < 2000 && g.State() != game.StateBattle; i++ {
		g.Tick(time.Second / 60)
		g.Frame()
	}
	if g.State() != game.StateBattle {
		panic("正常首戰未到達：" + g.State().String())
	}
	g.MoveMouse(240, 240)
	must(os.MkdirAll(filepath.Join(*out, "entry"), 0755))
	for _, n := range []string{"screen"} {
		base := filepath.Join(*out, "entry", n)
		must(g.Frame().WriteIndex(base + ".bin"))
		must(g.Frame().WritePNG(base+".png", g.Palette()))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
