//go:build ignore

// #79正常玩家入口探測；與cmd/gin3共用資源及Game，不匯入原版RAM。
// 先依battle_normal_probe_seed.py建立覆蓋，再以go run -overlay執行。
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
)

func main() {
	out := flag.String("out", "workplace/round44-normal", "新的探測輸出目錄")
	skipLoaded := flag.Bool("skip-loaded-save", false, "省略實行前診斷存檔，以排除存檔副作用")
	inputDisplay := flag.Bool("input-display", false, "首戰正常Input的四個純移動收據")
	flag.Parse()
	if _, err := os.Stat(*out); !os.IsNotExist(err) {
		panic("輸出已存在或無法核對")
	}
	if err := os.Mkdir(*out, 0755); err != nil {
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
	inputs := []string{}
	click := func(x, y int) {
		inputs = append(inputs, fmt.Sprintf("click %d %d", x, y))
		g.MoveMouse(x, y)
		g.Click(x, y)
		g.Frame()
	}
	// 第一個點擊跳過原版開頭；第二個點擊主選單取進度。
	click(300, 270)
	if g.State() == game.StateMainMenu {
		click(300, 270)
	}
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	inputs = append(inputs, "type 138; enter")
	if g.State() != game.StateLoadGame {
		panic(fmt.Sprintf("未到取進度：%v", g.State()))
	}
	click(300, 322)
	click(318, 208)
	if g.State() != game.StateCommandPhase {
		panic(fmt.Sprintf("未到命令階段：%v", g.State()))
	}
	save := func(name string) {
		dir := filepath.Join(*out, name)
		if err := os.Mkdir(dir, 0755); err != nil {
			panic(err)
		}
		if err := g.Frame().WriteIndex(filepath.Join(dir, "screen.bin")); err != nil {
			panic(err)
		}
		if err := g.Frame().WritePNG(filepath.Join(dir, "screen.png"), r.Menu.Palette()); err != nil {
			panic(err)
		}
		if err := g.SaveSlot(dir, 0); err != nil {
			panic(err)
		}
	}
	if !*skipLoaded {
		save("loaded")
	}
	// 測試期覆蓋只提供亂數觀測／固定，不改正常輸入、資產或局面。
	seedBefore := g.NormalProbeStrategySeed()
	g.NormalProbeSetStrategySeed(0xD3DC167E)
	seedFixed := g.NormalProbeStrategySeed()
	click(544, 125)
	click(560, 125)
	ticks := 0
	for ; ticks < 15000 && g.State() != game.StateBattle; ticks++ {
		switch g.State() {
		case game.StateTurnThinking, game.StateTurnWindow, game.StateTurnReport, game.StateEngagement:
			g.Tick(time.Second / 60)
		case game.StateOccupationReport, game.StateRebellionReport, game.StateRegionFlipReport, game.StateAdmiralFateReport, game.StateShipDelivery, game.StatePhezzanReport:
			click(240, 240)
		default:
			panic(fmt.Sprintf("實行停在%v", g.State()))
		}
		g.Frame()
	}
	if g.State() != game.StateBattle {
		panic("未進戰場")
	}
	g.MoveMouse(240, 240)
	save("entry")
	if *inputDisplay {
		for _, c := range []struct {
			name string
			x, y int
		}{{"flagship", 248, 248}, {"squad", 249, 224}, {"empty", 100, 140}, {"panel", 500, 290}} {
			g.MoveMouse(c.x, c.y)
			inputs = append(inputs, fmt.Sprintf("move %d %d", c.x, c.y))
			save(c.name)
		}
	}
	y, m, d := g.Date()
	meta := map[string]any{"classification": "normal-player-entry-probe; parity not established", "strategy_seed_observed_before": fmt.Sprintf("%08X", seedBefore), "strategy_seed_before": fmt.Sprintf("%08X", seedFixed), "strategy_seed_after": fmt.Sprintf("%08X", g.NormalProbeStrategySeed()), "seed_setup": "測試期Go覆蓋只固定實行前D3DC167E，正式產品不變", "inputs": inputs, "ticks": ticks, "date": []int{y, m, d}, "state": g.State(), "resources": "bootstrap.Load; Game.New", "palette_note": "screen.png暫用主選單色盤；正式RGB比較必須讀原版有效色盤"}
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(*out, "probe.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("正常正式入口已擷取，日期%d/%d/%d，ticks=%d；尚未宣稱same-state\n", y, m, d, ticks)
}
