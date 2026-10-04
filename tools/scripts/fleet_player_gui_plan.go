//go:build ignore

// 第79輪：由正式bootstrap與公開輸入產生每步狀態，供真實X11操作等待。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

func main() {
	input := flag.String("input", "", "第77輪正常公開輸入")
	out := flag.String("out", "", "新的計畫輸出目錄")
	flag.Parse()
	var parent struct {
		Faction int
		Inputs  []struct {
			Kind string
			X, Y int
			NS   int64
			Seed uint32
		}
	}
	b, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(b, &parent); err != nil {
		panic(err)
	}
	if _, err = os.Stat(*out); !os.IsNotExist(err) {
		panic("拒絕覆寫計畫")
	}
	if err = os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	r, err := bootstrap.Load("org_game/extracted", "docs/internals/data")
	if err != nil {
		panic(err)
	}
	if err = bootstrap.LoadIntro(r, "org_game/extracted"); err != nil {
		panic(err)
	}
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	if err = g.SetAILevel(-1); err != nil {
		panic(err)
	}
	var actions []map[string]any
	x, y := 0, 0
	step := func(kind string, px, py int, key byte, seed uint32) {
		a := map[string]any{"kind": kind, "x": px, "y": py, "key": int(key), "seed": seed, "before": g.FleetGUIStamp()}
		switch kind {
		case "click":
			x, y = px, py
			g.MoveMouse(x, y)
			g.Click(x, y)
		case "right":
			x, y = px, py
			g.MoveMouse(x, y)
			g.RightClick(x, y)
			g.RightRelease()
		case "key":
			g.Key(key)
		case "seed":
			g.FleetGUISeed(seed)
		}
		g.MoveMouse(x, y)
		g.Frame()
		a["after"] = g.FleetGUIStamp()
		if g.State() == game.StateBattle {
			stage := g.FleetGUIStamp()["stage"].(int)
			if kind == "click" && (stage == 1 || stage == 2) {
				name := fmt.Sprintf("command-%d-%05d", stage, len(actions))
				if err = g.Frame().WriteIndex(filepath.Join(*out, name+".bin")); err != nil {
					panic(err)
				}
				if err = g.Frame().WritePNG(filepath.Join(*out, name+".png"), g.Palette()); err != nil {
					panic(err)
				}
				a["snapshot"] = name
			}
		}
		actions = append(actions, a)
	}
	if g.State() == game.StateIntro {
		step("click", 300, 210, 0, 0)
	}
	step("click", 300, 210+20*parent.Faction, 0, 0)
	for _, k := range []byte("138\r") {
		step("key", 0, 0, k, 0)
	}
	step("right", 300, 240, 0, 0)
	step("right", 300, 240, 0, 0)
	for _, in := range parent.Inputs {
		if in.Kind == "tick" {
			g.Tick(time.Duration(in.NS))
			g.MoveMouse(x, y)
			g.Frame()
			continue
		}
		step(in.Kind, in.X, in.Y, 0, in.Seed)
	}
	if g.State() != game.StateBattle || !g.FleetGUIStamp()["idle"].(bool) {
		panic("未返回正常Input")
	}
	if err = g.Frame().WriteIndex(filepath.Join(*out, "final.bin")); err != nil {
		panic(err)
	}
	if err = g.Frame().WritePNG(filepath.Join(*out, "final.png"), g.Palette()); err != nil {
		panic(err)
	}
	if err = g.SaveSlot(*out, 0); err != nil {
		panic(err)
	}
	for _, n := range rules.SlotFileNames(0) {
		a, err := os.ReadFile(filepath.Join(*out, n))
		if err != nil {
			panic(err)
		}
		b, err := os.ReadFile(filepath.Join(filepath.Dir(*input), n))
		if err != nil || string(a) != string(b) {
			panic("公開逐Tick滑鼠更新改變保存" + n)
		}
	}
	points := map[string][2]int{}
	fld := r.Battle.Field
	for py := fld.Y + 1; py < fld.Y+fld.Rows*fld.TileH; py += 16 {
		for px := fld.X + 1; px < fld.X+fld.Cols*fld.TileW; px += 16 {
			col, row, moved := rules.ViewScrollStep(10, 10, rules.ViewScrollDir(px, py-ui.FrameCellOriginY))
			if moved {
				key := fmt.Sprintf("%d,%d", col-10, row-10)
				if _, ok := points[key]; !ok {
					points[key] = [2]int{px, py}
				}
			}
		}
	}
	if len(points) != 8 {
		panic("未取得八個正式捲動方向")
	}
	b, err = json.MarshalIndent(map[string]any{"classification": "正常bootstrap計畫；每Tick模擬正式MoveMouse與Frame；GUI另驗", "faction": parent.Faction, "actions": actions, "final": g.FleetGUIStamp(), "six_saves_equal_parent": true, "scroll_points": points}, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(*out, "plan.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Println("陣營", parent.Faction, "真實輸入計畫", len(actions), "六份保存保持")
}
