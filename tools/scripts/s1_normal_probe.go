//go:build ignore

// S1正常公開輸入探針。沒有StartBattle、局面注入或直接呼叫回合常式。
package main

import (
	"fmt"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
	"os"
	"path/filepath"
	"time"
)

func main() {
	out := os.Args[1]
	if err := os.Mkdir(out, 0755); err != nil {
		panic(err)
	}
	r, err := bootstrap.Load("/orig", "docs/internals/data")
	if err != nil {
		panic(err)
	}
	hd, err := presentation.Load("assets/hd", r.Faces)
	if err != nil {
		panic(err)
	}
	r.Faces.ArtworkPrefix = "face"
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	click := func(x, y int) {
		g.MoveMouse(x, y)
		g.Click(x, y)
		g.LeftRelease()
		g.Frame()
		g.MoveMouse(x, y)
		fmt.Println("click", x, y, g.State(), "menu", g.CommandMenuOpen(), "stage", g.FleetOrderStage())
	}
	snap := func(name string) {
		if err := hd.WritePNG(filepath.Join(out, name+".png"), g.Frame(), g.Palette()); err != nil {
			panic(err)
		}
	}
	click(300, 270)
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	snap("load")
	click(300, 282)
	click(318, 208)
	click(544, 125)
	click(560, 185)
	for i := 0; i < 8; i++ {
		click(424, 310)
	}
	click(368, 330)
	click(320, 350)
	fmt.Println("settings", g.DisplaySettings())
	snap("settings")
	click(416, 370)
	click(544, 125)
	click(560, 125)
	for i := 0; i < 30000 && g.State() != game.StateBattle; i++ {
		switch g.State() {
		case game.StateTurnWindow, game.StateTurnReport, game.StateOccupationReport, game.StateRebellionReport, game.StateRegionFlipReport, game.StateShipDelivery:
			click(560, 125)
		}
		g.Tick(time.Second / 60)
		g.Frame()
	}
	if g.State() != game.StateBattle {
		panic(g.State())
	}
	snap("battle")
	click(248, 250)
	snap("fleet-menu")
	click(250, 262)
	if g.FleetOrderStage() != game.FleetOrderPoint {
		panic("未進入目的地")
	}
	click(320, 290)
	if g.FleetOrderStage() != game.FleetOrderFacing {
		panic("未進入朝向")
	}
	click(350, 290)
	if g.FleetOrderStage() != game.FleetOrderNone {
		panic("命令未提交")
	}
	g.MoveMouse(500, 200)
	g.Frame()
	g.MoveMouse(248, 250)
	g.Frame()
	snap("ordered")
	combat, turn := 0, 0
	for tick := 0; tick < 36000; tick++ {
		if tick%900 == 0 {
			turn++
			click(600, 240)
			click(600, 242)
			fmt.Println("execute", turn, "tick", tick, g.State())
		}
		g.Tick(time.Second / 60)
		fb := g.Frame()
		anim := false
		for _, c := range fb.Commands {
			if c.Kind == gfx.DrawAnimation && c.Code == 0 {
				anim = true
			}
		}
		if anim {
			combat++
			if combat == 1 || combat%100 == 0 {
				fmt.Println("COMBAT", tick, combat)
				snap(fmt.Sprintf("combat-%06d", tick))
			}
			if combat >= 600 {
				break
			}
		}
		if tick%600 == 0 {
			snap(fmt.Sprintf("phase-%06d", tick))
		}
	}
	fmt.Println("combat_frames", combat)
	if combat == 0 {
		panic("未觀測到正常戰鬥動畫")
	}
}
