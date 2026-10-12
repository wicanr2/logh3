//go:build ignore

// 規格140：正常主選單讀劇本、實行與戰術命令，不注入局面／亂數。
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
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	out := flag.String("out", "", "新輸出目錄")
	scenario := flag.Int("scenario", 1, "原劇本1或3")
	faction := flag.Int("faction", 0, "玩家陣營")
	level := flag.Int("ai", 1, "AI模式")
	turns := flag.Int("turns", 4, "戰術回合")
	flag.Parse()
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	must(os.Mkdir(*out, 0755))
	r, err := bootstrap.Load("/orig", "docs/internals/data")
	must(err)
	hd, err := presentation.Load("assets/hd", r.Faces)
	must(err)
	r.Faces.ArtworkPrefix = "face"
	g, err := game.New(r)
	must(err)
	must(g.SetAILevel(*level))
	var events []map[string]any
	var states []*game.TacticalObservation
	click := func(x, y int) {
		g.MoveMouse(x, y)
		g.Click(x, y)
		g.LeftRelease()
		g.Frame()
		events = append(events, map[string]any{"kind": "click", "x": x, "y": y})
	}
	snap := func(name string) {
		must(hd.WritePNG(filepath.Join(*out, name+".png"), g.Frame(), g.Palette()))
		events = append(events, map[string]any{"kind": "capture", "name": name})
		states = append(states, g.TacticalObservation())
	}
	click(300, 270)
	for _, key := range []byte("138\r") {
		g.Key(key)
		s := string(key)
		if key == '\r' {
			s = "Return"
		}
		events = append(events, map[string]any{"kind": "key", "key": s})
	}
	snap("load")
	click(300, 262+20*(*scenario))
	click(318, 208+20*(*faction))
	if g.State() != game.StateCommandPhase {
		panic(g.State())
	}
	snap("strategy")
	click(544, 125)
	click(560, 185)
	for i := 0; i < 8; i++ {
		click(424, 310)
	}
	click(368, 330)
	click(320, 350)
	click(240, 370) // 正常設定選單開啟顯示敵軍，方便觀察協同。
	click(416, 370)
	click(544, 125)
	click(560, 125)
	for tick := 0; tick < 30000 && g.State() != game.StateBattle; tick++ {
		switch g.State() {
		case game.StateTurnReport, game.StateOccupationReport, game.StateRebellionReport, game.StateRegionFlipReport, game.StateShipDelivery, game.StateAdmiralFateReport, game.StatePhezzanReport:
			click(240, 240)
		default:
			g.Tick(time.Second / 60)
		}
		g.Frame()
	}
	if g.State() != game.StateBattle {
		panic(g.State())
	}
	events = append(events, map[string]any{"kind": "wait_state", "state": "StateBattle", "expect": map[string]any{"battle_ready": true}, "timeout": 240})
	snap("entry")
	fmt.Printf("S%d faction%d ai%d ENTRY %+v\n", *scenario, *faction, *level, g.TacticalObservation())
	for turn := 0; turn < *turns && g.State() == game.StateBattle; turn++ {
		click(600, 240)
		click(600, 242) // 面板空白處開戰術選單，避免鏡頭移動後點到部隊。
		previous := g.TacticalObservation().Turn
		for tick := 0; tick < 30000; tick++ {
			observation := g.TacticalObservation()
			if observation != nil && observation.CombatInput != "" {
				events = append(events, map[string]any{"kind": "wait_state", "state": "StateBattle", "expect": map[string]any{"tactical.combat_input": observation.CombatInput}, "timeout": 240})
				if observation.CombatInput == "target" {
					click(240, 470) // 由正常空白處選「自動」，不猜可用武器。
				} else {
					click(observation.CombatPoint[0], observation.CombatPoint[1])
				}
			} else if g.State() == game.StateTurnReport {
				click(240, 240)
			} else {
				g.Tick(time.Second / 60)
			}
			g.Frame()
			if g.State() == game.StateCommandPhase || (g.BattleInputReady() && g.TacticalObservation().Turn > previous) {
				break
			}
			if tick == 29999 {
				panic(fmt.Sprintf("戰術續接逾時 state=%v observation=%+v", g.State(), g.TacticalObservation()))
			}
		}
		name := fmt.Sprintf("turn-%02d", turn+1)
		expect := map[string]any{"battle_ready": g.BattleInputReady()}
		if observation := g.TacticalObservation(); observation != nil {
			expect["tactical.turn"] = observation.Turn
		}
		events = append(events, map[string]any{"kind": "wait_state", "state": g.State().String(), "expect": expect, "timeout": 240, "wait": 1.0})
		snap(name)
		fmt.Println(name, g.TacticalObservation())
	}
	for name, value := range map[string]any{"PLAN.json": events, "STATES.json": states} {
		b, err := json.MarshalIndent(value, "", "  ")
		must(err)
		must(os.WriteFile(filepath.Join(*out, name), append(b, '\n'), 0644))
	}
}
