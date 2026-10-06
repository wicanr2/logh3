//go:build ignore

// 規格121：正式Game產生可見操作的預期畫面，供獨立X11入口核對。
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

type event struct {
	Kind   string `json:"kind"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Button int    `json:"button,omitempty"`
	Key    string `json:"key,omitempty"`
	Name   string `json:"name,omitempty"`
	State  string `json:"state,omitempty"`
}

func main() {
	orig := flag.String("orig", "/orig", "原版目錄")
	data := flag.String("data", "docs/internals/data", "規格目錄")
	out := flag.String("out", "", "新輸出目錄")
	faction := flag.Int("faction", 0, "陣營")
	flag.Parse()
	if *out == "" || (*faction != 0 && *faction != 1) {
		panic("參數")
	}
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	r, err := bootstrap.Load(*orig, *data)
	if err != nil {
		panic(err)
	}
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	var events []event
	click := func(x, y, b int) {
		if b == 3 {
			g.RightClick(x, y)
		} else {
			g.Click(x, y)
			g.LeftRelease()
		}
		g.Frame()
		// 正式app每次Update都接收目前位置，即使按下後滑鼠未再移動。
		g.MoveMouse(x, y)
		g.Frame()
		events = append(events, event{Kind: "click", X: x, Y: y, Button: b})
	}
	move := func(x, y int) { g.MoveMouse(x, y); g.Frame(); events = append(events, event{Kind: "move", X: x, Y: y}) }
	snap := func(name string, state game.State) {
		if g.State() != state {
			panic(fmt.Sprintf("%s: %v != %v", name, g.State(), state))
		}
		if err := g.Frame().WritePNG(filepath.Join(*out, name+".png"), g.Palette()); err != nil {
			panic(err)
		}
		events = append(events, event{Kind: "capture", Name: name, State: state.String()})
	}
	click(300, 210+20**faction, 1)
	for _, c := range []byte("138\r") {
		g.Key(c)
		key := string(c)
		if c == '\r' {
			key = "Return"
		}
		events = append(events, event{Kind: "key", Key: key})
	}
	snap("01-account", game.StateMainGame)
	click(448, 350, 1)
	snap("02-budget", game.StateBudgetPlan)
	click(487, 329, 1)
	snap("03-budget-changed", game.StateBudgetPlan)
	click(463, 370, 1)
	snap("04-account-return", game.StateMainGame)
	click(448, 370, 1)
	snap("05-order", game.StateShipOrder)
	move(600, 420)
	snap("06-order-cursor", game.StateShipOrder)
	click(448, 350, 1)
	snap("07-build", game.StateBuildPlan)
	click(232, 210, 1)
	snap("08-build-row0", game.StateBuildPlan)
	click(232, 230, 1)
	snap("09-build-row1", game.StateBuildPlan)
	move(600, 420)
	snap("10-build-cursor", game.StateBuildPlan)
	click(512, 370, 1)
	snap("11-order-return", game.StateShipOrder)
	click(448, 370, 1)
	snap("12-command", game.StateCommandPhase)
	for _, m := range []struct {
		name         string
		x            int
		menu, target game.State
	}{{"strategy", 352, game.StateCmdMenu, game.StateTaxPlan}, {"lists", 448, game.StateListMenu, game.StateAdmiralListView}, {"system", 544, game.StateSysMenu, game.StateSettings}} {
		click(m.x, 125, 1)
		snap(m.name+"-open", m.menu)
		move(m.x+12, 165)
		snap(m.name+"-moved", m.menu)
		y := 165
		if m.x == 544 {
			y = 185
		}
		click(m.x+16, y, 1)
		snap(m.name+"-selected", m.target)
		click(600, 400, 3)
		snap(m.name+"-closed", game.StateCommandPhase)
	}
	b, err := json.MarshalIndent(map[string]any{"faction": *faction, "events": events, "scope": "remake正常輸入与renderer一致性，不宣稱原版parity"}, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(*out, "PLAN.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
}
