//go:build ignore

// 正常bootstrap／S3／戰術實行與保存；不注入功績、位置或亂數。
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

func main() {
	out := flag.String("out", "", "新的正常玩家輸出目錄")
	flag.Parse()
	if *out == "" {
		panic("缺輸出目錄")
	}
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		panic("拒絕覆寫輸出")
	}
	must(os.MkdirAll(*out, 0755))
	r, e := bootstrap.Load("org_game/extracted", "docs/internals/data")
	must(e)
	r.SaveDir = *out
	g, e := game.New(r)
	must(e)
	must(g.SetAILevel(-1))
	inputs := []map[string]any{}
	click := func(x, y int) {
		inputs = append(inputs, map[string]any{"kind": "click", "x": x, "y": y})
		g.MoveMouse(x, y)
		g.Click(x, y)
		g.LeftRelease()
		g.Frame()
		must(g.SaveErr())
	}
	tick := func(n int) {
		inputs = append(inputs, map[string]any{"kind": "tick", "count": n, "ns": int64(time.Second / 60)})
		for i := 0; i < n; i++ {
			g.Tick(time.Second / 60)
			g.Frame()
			must(g.SaveErr())
		}
	}
	click(300, 270)
	for _, k := range []byte("138\r") {
		g.Key(k)
		inputs = append(inputs, map[string]any{"kind": "key", "byte": k})
	}
	for _, p := range [][2]int{{300, 322}, {318, 208}, {544, 125}, {560, 125}} {
		click(p[0], p[1])
	}
	for i := 0; i < 2000 && g.State() != game.StateBattle; i++ {
		tick(1)
	}
	if g.State() != game.StateBattle {
		panic("未到正常首戰")
	}
	shot := func(name string) {
		g.MoveMouse(240, 240)
		must(g.Frame().WriteIndex(filepath.Join(*out, name+".bin")))
		must(g.Frame().WritePNG(filepath.Join(*out, name+".png"), g.Palette()))
	}
	shot("entry")
	menu := func(row int) {
		click(160, 280)
		if g.CommandMenuOpen() {
			panic("戰術選單落點命中單位")
		}
		x, y := r.BattleMenu.BattleMenuAt(160, 280)
		l := r.BattleMenu.Layout
		click(x+l.Highlight.X-l.Window.X+8, y+l.Highlight.Y-l.Window.Y+row*l.Text.RowPitch+8)
	}
	menu(0)
	tick(5000)
	if g.State() != game.StateBattle {
		panic(fmt.Sprint("實行未返回：", g.State()))
	}
	shot("after-turn")
	menu(4)
	if g.State() != game.StateSaveGame {
		panic("正常保存選單未開")
	}
	shot("save-menu")
	click(200, 150)
	w, e := os.ReadFile(filepath.Join(*out, "WORK.DT0"))
	must(e)
	if len(w) != 230 || binary.LittleEndian.Uint16(w[6:]) != 1 {
		panic("未完成首回合結算")
	}
	b, _ := json.MarshalIndent(map[string]any{"classification": "normal bootstrap/public input; no original full-path parity claim", "no_state_injection": true, "rng_setup": "product default; no seed injection", "strategic_turn": g.Turn(), "battle_work_turn": binary.LittleEndian.Uint16(w[6:]), "inputs": inputs, "state": g.State().String(), "promotion_scope": "first normal turn continuation; no natural promotion claim"}, "", "  ")
	must(os.WriteFile(filepath.Join(*out, "replay.json"), append(b, '\n'), 0644))
	fmt.Println("正常首戰實行、8-4續接與公開存進度通過")
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
