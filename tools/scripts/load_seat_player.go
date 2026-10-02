//go:build ignore

// 088 READY：正式bootstrap／Game正常讀檔三分支；不匯入原版RAM、不注入局面。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

func main() {
	out := flag.String("out", "", "新的本機收據目錄")
	flag.Parse()
	if *out == "" {
		panic("缺out")
	}
	if _, err := os.Stat(*out); !os.IsNotExist(err) {
		panic("輸出已存在")
	}
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	for mode, name := range []string{"empire", "alliance", "versus"} {
		r, err := bootstrap.Load("org_game/extracted", "docs/internals/data")
		if err != nil {
			panic(err)
		}
		g, err := game.New(r)
		if err != nil {
			panic(err)
		}
		inputs := []string{}
		click := func(x, y int) { inputs = append(inputs, fmt.Sprintf("click %d %d", x, y)); g.Click(x, y); g.Frame() }
		click(300, 270)
		if g.State() == game.StateMainMenu {
			click(300, 270)
		}
		for _, k := range []byte("138\r") {
			g.Key(k)
		}
		inputs = append(inputs, "type 138; enter")
		if g.State() != game.StateLoadGame {
			panic(g.State())
		}
		click(300, 322)
		if g.State() != game.StateLoadFaction {
			panic(g.State())
		}
		base := filepath.Join(*out, name)
		if err := os.Mkdir(base, 0755); err != nil {
			panic(err)
		}
		for sel := -1; sel < 3; sel++ {
			x, y := 300, 322
			if sel >= 0 {
				x, y = 318, 208+20*sel
				g.RightClick(x, y)
				inputs = append(inputs, fmt.Sprintf("right %d %d", x, y))
			} else {
				g.MoveMouse(x, y)
			}
			if g.State() != game.StateLoadFaction {
				panic("右鍵關掉陣營選單")
			}
			if err := g.Frame().WriteIndex(filepath.Join(base, fmt.Sprintf("menu%d.bin", sel))); err != nil {
				panic(err)
			}
		}
		stages := []map[string]any{}
		save := func(stage string) {
			p := filepath.Join(base, stage)
			if err := os.Mkdir(p, 0755); err != nil {
				panic(err)
			}
			if err := g.Frame().WriteIndex(filepath.Join(p, "screen.bin")); err != nil {
				panic(err)
			}
			if err := g.SaveSlot(p, 0); err != nil {
				panic(err)
			}
			stages = append(stages, map[string]any{"stage": stage, "state": g.State().String(), "turn": g.Turn(), "faction": g.Faction()})
		}
		click(318, 208+20*mode)
		expect := game.StateCommandPhase
		if mode == 2 {
			expect = game.StateVersusInput
		}
		if g.State() != expect {
			panic(g.State())
		}
		save("loaded")
		if mode == 2 {
			click(432, 270)
			if g.State() != game.StateCommandPhase {
				panic("載入交接後不應顯示月報")
			}
			save("confirmed")
			click(544, 125)
			click(560, 125)
			if g.State() != game.StateVersusInput || g.Faction() != 1 || g.Turn() != 380 {
				panic("首次交接同盟錯誤")
			}
			click(432, 270)
			if g.State() != game.StateCommandPhase {
				panic("同盟交接後不應顯示月報")
			}
			save("handoff")
		}
		b, err := json.MarshalIndent(map[string]any{"classification": "normal player UI; full state parity not established", "inputs": inputs, "stages": stages, "resources": "bootstrap.Load; Game.New; product renderer; no RAM injection"}, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(filepath.Join(base, "probe.json"), append(b, '\n'), 0644); err != nil {
			panic(err)
		}
	}
}
