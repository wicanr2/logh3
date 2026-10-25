//go:build ignore

// Generate a normal-input model for the garrison promo. No memory/seed injection.
// The model's rule API is mapped to the actual settings button by the GUI driver.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
)

func main() {
	out := flag.String("out", "", "new output directory")
	flag.Parse()
	if _, err := os.Stat(*out); !os.IsNotExist(err) {
		panic("output exists")
	}
	must(os.Mkdir(*out, 0755))
	r, err := bootstrap.Load("/orig", "docs/internals/data")
	must(err)
	g, err := game.New(r)
	must(err)
	regions, err := asset.OpenRGN(r.StarRGN, 16, 20)
	must(err)
	point := func(region int) (int, int) {
		var points [][2]int
		for y := 0; y < 280; y += 2 {
			for x := 8; x < 310; x += 2 {
				if asset.RegionAt(regions, x, y, 16, 20) == region {
					points = append(points, [2]int{x, y + 40})
				}
			}
		}
		if len(points) == 0 {
			panic("no map point")
		}
		return points[len(points)/2][0], points[len(points)/2][1]
	}
	var events []map[string]any
	add := func(kind string, x, y int, key string, name string) {
		switch kind {
		case "click":
			g.Click(x, y)
		case "right":
			g.RightClick(x, y)
			g.RightRelease()
		case "key":
			g.Key(key[0])
		case "capture":
			g.Frame()
		}
		g.Frame()
		e := map[string]any{"kind": kind, "x": x, "y": y, "key": key, "name": name, "state": g.State().String(), "strategic": g.StrategicObservation()}
		events = append(events, e)
		if name != "" {
			fmt.Println(name, g.State(), g.StrategicObservation())
		}
	}
	click := func(x, y int) { add("click", x, y, "", "") }
	right := func() { add("right", 300, 240, "", "") }
	snap := func(name string) { add("capture", 0, 0, "", name) }
	want := func(s game.State) {
		if g.State() != s {
			panic(fmt.Sprintf("state %v, want %v", g.State(), s))
		}
	}
	selectRegion := func(region int) {
		if view := g.StrategicObservation(); view != nil && view.Region == region {
			return
		}
		x, y := point(region)
		click(x, y)
	}
	advance := func() {
		click(544, 125)
		click(560, 125)
		for n := 0; g.State() != game.StateCommandPhase && n < 12000; n++ {
			switch g.State() {
			case game.StateTurnThinking, game.StateTurnWindow, game.StateTurnReport, game.StateEngagement:
				g.Tick(time.Second / 60)
			case game.StateShipDelivery, game.StateMainGame, game.StateOccupationReport, game.StateRebellionReport, game.StateRegionFlipReport, game.StateAdmiralFateReport, game.StatePhezzanReport:
				g.RightClick(300, 240)
			default:
				panic(fmt.Sprintf("advance stopped %v", g.State()))
			}
		}
		want(game.StateCommandPhase)
		events = append(events, map[string]any{"kind": "wait_command", "state": g.State().String(), "strategic": g.StrategicObservation()})
	}
	click(300, 230)
	for _, key := range []string{"1", "3", "8", "\r"} {
		add("key", 0, 0, key, "")
	}
	right()
	right()
	want(game.StateCommandPhase)
	capital := rules.CapitalRegion(1)
	front := -1
	for _, n := range r.Graph.Neighbours(capital) {
		if r.Planets.RegionOwner(n) == 1 {
			front = n
			break
		}
	}
	if front < 0 {
		panic("no allied garrison neighbour")
	}
	snap("new-alliance")
	for n := 0; n < 2; n++ {
		click(352, 125)
		click(400, 125)
		click(400, 85)
		want(game.StateAdmiralDetail)
		click(488, 330)
		want(game.StateFleetOrg)
		click(480, 350)
		want(game.StateFleetComposition)
		click(360, 210)
		click(176, 240)
		click(360, 250)
		click(240, 240)
		click(510, 350)
		want(game.StateCommandPhase)
		snap(fmt.Sprintf("initial-fleet-%d", n))
	}
	before := g.StrategicObservation()
	if len(before.Fleets) != 2 {
		panic("two initial fleets required")
	}
	admirals := []int{before.Fleets[0].Admiral, before.Fleets[1].Admiral}
	for row := 0; row < 2; row++ {
		selectRegion(capital)
		click(350, 290)
		click(450, 308+row*20)
		want(game.StateFleetWindow)
		click(490, 345)
		click(497, 352)
		want(game.StateFleetTargetPick)
		if !g.FleetTargetsAllowed()[front] {
			panic("garrison target not allowed")
		}
		x, y := point(front)
		click(x, y)
		want(game.StateFleetConfirm)
		click(432, 263)
		want(game.StateCommandPhase)
	}
	snap("garrison-orders")
	advance()
	selectRegion(front)
	snap("garrison-arrived")
	if len(g.StrategicObservation().Fleets) != 0 {
		panic("garrison did not disband both fleets")
	}
	click(544, 125)
	click(560, 185)
	want(game.StateSettings)
	snap("option-default-off")
	options := g.RuleOptions()
	if options.NonBattleshipFlagship {
		panic("new option must default off")
	}
	options.NonBattleshipFlagship = true
	g.SetRuleOptions(options)
	events = append(events, map[string]any{"kind": "rule", "id": "non-battleship", "enabled": true, "state": g.State().String()})
	snap("option-enabled")
	right()
	want(game.StateCommandPhase)
	selectRegion(front)
	click(560, 290)
	idle := g.StrategicObservation().Idle
	pick := func(admiral int) {
		for row, id := range idle {
			if id == admiral && row < 5 {
				click(450, 308+20*row)
				return
			}
		}
		panic("commander missing from garrison list")
	}
	pick(admirals[0])
	snap("first-commander")
	right()
	pick(admirals[1])
	snap("changed-commander")
	click(488, 330)
	want(game.StateFleetOrg)
	click(368, 350)
	click(368, 130)
	want(game.StateFleetOrg)
	snap("fleet-number-three")
	click(480, 350)
	want(game.StateFleetComposition)
	snap("local-ship-stock")
	messageConfirm := func() {
		minX, minY, maxX, maxY := 640, 480, -1, -1
		for y := 0; y < 480; y++ {
			for x := 0; x < 640; x++ {
				if _, hit := r.MessageBox.MessageBoxButtonAt(0, x, y); hit {
					if x < minX {
						minX = x
					}
					if y < minY {
						minY = y
					}
					if x > maxX {
						maxX = x
					}
					if y > maxY {
						maxY = y
					}
				}
			}
		}
		if maxX < 0 {
			panic("no message button")
		}
		click((minX+maxX)/2, (minY+maxY)/2)
	}
	for _, k := range []int{5, 6} {
		click(360, 170+k*20)
		click(208, 240)
		if !g.StrategicObservation().Message {
			panic("E/T accepted as flagship")
		}
		snap(fmt.Sprintf("support-rejected-%d", k))
		messageConfirm()
	}
	click(360, 210)
	click(208, 240)
	if g.StrategicObservation().Editor[0].Kind != 2 {
		panic("cruiser flagship failed")
	}
	snap("cruiser-flagship")
	click(360, 250)
	click(176, 240)
	click(360, 170)
	click(240, 240)
	snap("formed-at-garrison")
	click(510, 350)
	want(game.StateCommandPhase)
	snap("fleet-ready")
	f := g.StrategicObservation().Fleets
	if len(f) != 1 || f[0].Region != front || f[0].Admiral != admirals[1] || f[0].Number != 3 {
		panic("garrison fleet identity wrong")
	}
	to := capital
	for _, n := range r.Graph.Neighbours(front) {
		if n != capital && r.Planets.RegionOwner(n) == 1 {
			to = n
			break
		}
	}
	selectRegion(front)
	click(350, 290)
	click(450, 308)
	want(game.StateFleetWindow)
	click(490, 345)
	click(497, 312)
	want(game.StateFleetTargetPick)
	if !g.FleetTargetsAllowed()[to] {
		panic("departure target not allowed")
	}
	x, y := point(to)
	click(x, y)
	click(432, 263)
	want(game.StateCommandPhase)
	snap("departure-order")
	advance()
	selectRegion(to)
	snap("departed")
	if fleets := g.StrategicObservation().Fleets; len(fleets) != 1 || fleets[0].Region != to {
		panic("fleet did not depart")
	}
	data, err := json.MarshalIndent(map[string]any{"capital": capital, "garrison": front, "destination": to, "admirals": admirals, "normal_gui_required": true, "events": events}, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(*out, "PLAN.json"), data, 0644))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
