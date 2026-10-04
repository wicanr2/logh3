//go:build ignore

// 正常bootstrap重播公開輸入；overlay只開放測試seed設定。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
	"os"
	"path/filepath"
	"time"
)

func main() {
	input := flag.String("input", "", "正常Game的inputs.json")
	out := flag.String("out", "", "新的bootstrap輸出目錄")
	flag.Parse()
	var plan struct {
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
	if err = json.Unmarshal(b, &plan); err != nil {
		panic(err)
	}
	if _, err = os.Stat(*out); !os.IsNotExist(err) {
		panic("拒絕覆寫輸出")
	}
	if err = os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	r, err := bootstrap.Load("org_game/extracted", "docs/internals/data")
	if err != nil {
		panic(err)
	}
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	if err = g.SetAILevel(-1); err != nil {
		panic(err)
	}
	// bootstrap有開頭動畫；由公開左鍵跳過，再照正常新局入口。
	if g.State() == game.StateIntro {
		g.Click(300, 210)
	}
	g.Click(300, 210+20*plan.Faction)
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	if g.State() != game.StateCommandPhase {
		panic(fmt.Sprint("正常新局未到命令：", g.State()))
	}
	scenes := 0
	for _, in := range plan.Inputs {
		before := g.State()
		switch in.Kind {
		case "click":
			g.MoveMouse(in.X, in.Y)
			g.Click(in.X, in.Y)
			g.Frame()
		case "right":
			g.RightClick(in.X, in.Y)
			g.RightRelease()
		case "tick":
			g.Tick(time.Duration(in.NS))
		case "seed":
			g.FleetChainProbeSeed(in.Seed)
		default:
			panic("未知輸入")
		}
		if g.State() == game.StateEngagement && before != game.StateEngagement {
			scenes++
			if err = g.Frame().WritePNG(filepath.Join(*out, fmt.Sprintf("scene-%d.png", scenes)), g.Palette()); err != nil {
				panic(err)
			}
		}
		if err = g.SaveErr(); err != nil {
			panic(err)
		}
	}
	if g.State() != game.StateBattle || scenes == 0 {
		panic(fmt.Sprint("正常重播未回戰場：", g.State(), " scenes=", scenes))
	}
	if err = g.Frame().WritePNG(filepath.Join(*out, "screen.png"), g.Palette()); err != nil {
		panic(err)
	}
	if err = g.Frame().WriteIndex(filepath.Join(*out, "screen.bin")); err != nil {
		panic(err)
	}
	if err = g.SaveSlot(*out, 0); err != nil {
		panic(err)
	}
	// 直接核對六份正式保存，不以相似畫面取代資料。
	for _, name := range rules.SlotFileNames(0) {
		a, e := os.ReadFile(filepath.Join(filepath.Dir(*input), name))
		if e != nil {
			panic(e)
		}
		b, e := os.ReadFile(filepath.Join(*out, name))
		if e != nil {
			panic(e)
		}
		if string(a) != string(b) {
			panic(name + "與正常Game保存不同")
		}
	}
	meta, _ := json.MarshalIndent(map[string]any{"faction": plan.Faction, "resources": "bootstrap.Load; Game.New", "classification": "normal player input replay; original full-path parity unknown", "seed_setup": "test-only overlay, before each battle execution", "inputs": len(plan.Inputs), "engagement_entries": scenes, "six_save_files_equal": true, "state": g.State()}, "", "  ")
	if err = os.WriteFile(filepath.Join(*out, "replay.json"), append(meta, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Println("正常bootstrap、交戰、續回戰場及六檔保存一致", plan.Faction, len(plan.Inputs), scenes)
}
