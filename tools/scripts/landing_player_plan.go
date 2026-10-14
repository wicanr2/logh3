//go:build ignore

// Normal bootstrap and public input only. Plans player orders, never state/RNG.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
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
	r.SaveDir = filepath.Join(out, "save")
	if err := os.Mkdir(r.SaveDir, 0755); err != nil {
		panic(err)
	}
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	var actions, stages, saveActions []map[string]any
	passed := false
	defer func() {
		b, e := json.MarshalIndent(map[string]any{"actions": actions, "save_actions": saveActions, "stages": stages, "passed": passed, "state": g.State().String(), "tactical": g.TacticalObservation(), "state_injection": false, "seed_injection": false, "passive_override": false}, "", "  ")
		if e != nil {
			panic(e)
		}
		if e = os.WriteFile(filepath.Join(out, "plan.json"), append(b, '\n'), 0644); e != nil {
			panic(e)
		}
	}()
	stamp := func() map[string]any {
		s := map[string]any{"state": g.State().String()}
		if t := g.TacticalObservation(); t != nil {
			s["region"], s["turn"], s["idle"] = t.Region, t.Turn, t.Idle
			s["view"], s["fleet_order"], s["command_open"], s["menu_open"] = t.View, t.FleetOrder, t.CommandOpen, t.MenuOpen
			s["combat_input"] = t.CombatInput
		}
		return s
	}
	click := func(x, y int) {
		before := stamp()
		a := map[string]any{"kind": "click", "point": [2]int{x, y}, "before": before}
		if t := g.TacticalObservation(); t != nil {
			f := r.Battle.Field
			if t.CommandOpen {
				a["command_row"] = (y - t.CommandPoint[1]) / r.CommandMenus.Layout.RowPitch
			} else if t.CombatInput == "menu" {
				a["weapon_row"] = (y - t.CombatPoint[1]) / r.CommandMenus.Layout.RowPitch
			} else if t.CombatInput == "" && !t.MenuOpen && x >= f.X && x < f.X+f.Cols*f.TileW && y >= f.Y && y < f.Y+f.Rows*f.TileH {
				a["world"] = [2]int{x - f.X + t.View[0]*f.TileW, y - f.Y + t.View[1]*f.TileH}
			}
		}
		g.MoveMouse(x, y)
		g.Click(x, y)
		g.LeftRelease()
		g.Frame()
		a["after"] = stamp()
		actions = append(actions, a)
	}
	right := func(x, y int) {
		before := stamp()
		g.MoveMouse(x, y)
		g.RightClick(x, y)
		g.RightRelease()
		g.Frame()
		actions = append(actions, map[string]any{"kind": "right", "point": [2]int{x, y}, "before": before, "after": stamp()})
	}
	snap := func(name string) {
		fb := g.Frame()
		if e := fb.WritePNG(filepath.Join(out, name+".png"), g.Palette()); e != nil {
			panic(e)
		}
		if e := fb.WriteIndex(filepath.Join(out, name+".bin")); e != nil {
			panic(e)
		}
		actions = append(actions, map[string]any{"kind": "snapshot", "name": name, "after": stamp()})
	}
	captured := false
	stageSeen := map[string]bool{}
	step := func() {
		if p := g.PlanetPlaybackObservation(); p != nil {
			pb := g.PlaybackObservation()
			key := fmt.Sprintf("%s/%d", pb.ID, p.Kind)
			if !stageSeen[key] {
				stageSeen[key] = true
				stages = append(stages, map[string]any{"id": key, "stage": p})
				fmt.Println("STAGE", key, *p)
			}
			if p.Kind == 2 && p.Captured {
				captured = true
			}
		}
		switch g.State() {
		case game.StateRegionInfo:
			right(240, 240)
		case game.StateTurnReport, game.StateOccupationReport, game.StateRebellionReport, game.StateRegionFlipReport, game.StateAdmiralFateReport, game.StateShipDelivery, game.StatePhezzanReport:
			click(240, 240)
		case game.StateBattle:
			t := g.TacticalObservation()
			if t.MessageOpen {
				snap("unhandled-modal")
				panic("unhandled modal; inspect the actual public button")
			} else if t.CombatInput != "" {
				x, y := t.CombatPoint[0], t.CombatPoint[1]
				if t.CombatInput == "target" {
					x, y = 240, 470
					for _, c := range t.CombatChoices {
						if c.Mask != 0 {
							x, y = c.Point[0], c.Point[1]
							break
						}
					}
				} else if t.CombatInput == "menu" && t.CombatWeaponMask != 0 {
					row := 1
					if t.CombatWeaponMask&2 == 0 {
						row = 2
						if t.CombatWeaponMask&4 == 0 {
							row = 0
						}
					}
					y += row * r.CommandMenus.Layout.RowPitch
				}
				click(x, y)
			}
		}
		g.Tick(g.TimerTick())
		if e := g.SaveErr(); e != nil {
			panic(e)
		}
	}
	wait := func() {
		for n := 0; n < 30000; n++ {
			if g.State() == game.StateCommandPhase || g.BattleInputReady() && g.TacticalObservation().Idle {
				return
			}
			step()
		}
		panic(fmt.Sprint("wait exceeded ", g.State(), g.TacticalObservation()))
	}
	click(600, 440)
	click(600, 440)
	if g.State() != game.StateMainMenu {
		panic(g.State())
	}
	click(300, 270)
	for _, k := range []byte("138") {
		g.Key(k)
		actions = append(actions, map[string]any{"kind": "key", "key": string(k), "after": stamp()})
	}
	click(432, 270)
	click(300, 322)
	click(318, 228)
	if g.State() != game.StateCommandPhase {
		panic(g.State())
	}
	snap("s3-alliance")
	click(544, 125)
	click(560, 185)
	for i := 0; i < 3; i++ {
		click(424, 310)
	}
	click(416, 370)
	field := r.Battle.Field
	clamp := func(n, hi int) int { return max(0, min(hi, n)) }
	for turn := 0; turn < 64 && !captured; turn++ {
		if g.State() == game.StateCommandPhase {
			click(544, 125)
			click(560, 125)
			wait()
		}
		if g.State() != game.StateBattle {
			panic(g.State())
		}
		t := g.TacticalObservation()
		fmt.Println("TURN", turn, t.Region, t.Turn, t.Planets)
		actions = append(actions, map[string]any{"kind": "ready", "after": stamp()})
		for _, u := range t.Units {
			if u.Faction != 1 || u.Squadron != 0 || u.Out || u.Ships == 0 {
				continue
			}
			var goal *game.TacticalPlanetObservation
			best := int64(1 << 62)
			for _, p := range t.Planets {
				if p.Owner == 1 {
					continue
				}
				d := rules.Distance16(u.X, u.Y, p.X, p.Y)
				if d < best {
					best = d
					v := p
					goal = &v
				}
			}
			if goal == nil {
				continue
			}
			want := [2]int{u.X/field.TileW - 1, u.Y/field.TileH - 1}
			if goal.X < u.X {
				want[0] = u.X/field.TileW - field.Cols + 2
			}
			if goal.Y < u.Y {
				want[1] = u.Y/field.TileH - field.Rows + 2
			}
			want[0], want[1] = clamp(want[0], field.MapCols-field.Cols), clamp(want[1], field.MapRows-field.Rows)
			for n := 0; n < 256 && g.TacticalObservation().View != want; n++ {
				v := g.TacticalObservation().View
				if v[0] != want[0] {
					if v[0] < want[0] {
						right(400, 220)
					} else {
						right(64, 220)
					}
				} else if v[1] < want[1] {
					right(224, 400)
				} else {
					right(224, 80)
				}
			}
			v := g.TacticalObservation().View
			x, y := field.X+u.X-v[0]*field.TileW, field.Y+u.Y-v[1]*field.TileH
			cells := r.BattleMaps[t.Region]
			start := [2]int{goal.X / field.TileW, goal.Y / field.TileH}
			cost := map[[2]int]int{start: 0}
			q := [][2]int{start}
			for at := 0; at < len(q); at++ {
				p := q[at]
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						w := [2]int{p[0] + dx, p[1] + dy}
						if w[0] < 0 || w[1] < 0 || w[0] >= field.MapCols || w[1] >= field.MapRows {
							continue
						}
						if _, ok := cost[w]; ok {
							continue
						}
						if !r.Terrain.Passable(cells, w[0]*field.TileW, w[1]*field.TileH) || !r.Terrain.LineOpen(cells, p[0]*field.TileW, p[1]*field.TileH, w[0]*field.TileW, w[1]*field.TileH) {
							continue
						}
						cost[w] = cost[p] + 1
						q = append(q, w)
					}
				}
			}
			dx, dy := -1, -1
			bestCost := int(1 << 30)
			bestDist := int64(1 << 62)
			for py := field.Y + 8; py < field.Y+field.Rows*field.TileH; py += 16 {
				for px := field.X + 8; px < field.X+field.Cols*field.TileW; px += 16 {
					wx, wy := v[0]*field.TileW+px-field.X, v[1]*field.TileH+py-field.Y
					if !r.Terrain.LineOpen(cells, u.X, u.Y, wx, wy) {
						continue
					}
					c, ok := cost[[2]int{wx / field.TileW, wy / field.TileH}]
					if !ok {
						continue
					}
					d := rules.Distance16(wx, wy, goal.X, goal.Y)
					if c < bestCost || c == bestCost && d < bestDist {
						dx, dy, bestCost, bestDist = px, py, c, d
					}
				}
			}
			if dx < 0 {
				continue
			}
			click(x+8, y+8)
			cur := g.TacticalObservation()
			if !cur.CommandOpen || cur.CommandKind != "fleet" {
				right(240, 240)
				continue
			}
			row := 0
			if best <= int64(300*rules.FixOne) || t.Turn == rules.MaxBattleTurns {
				row = 1
			}
			click(cur.CommandPoint[0], cur.CommandPoint[1]+row*r.CommandMenus.Layout.RowPitch)
			click(dx, dy)
			if g.TacticalObservation().MessageOpen {
				step()
				continue
			}
			click(dx, dy)
			if g.FleetOrderStage() != game.FleetOrderNone {
				panic("order incomplete")
			}
		}
		click(240, 470)
		mx, my := r.BattleMenu.BattleMenuAt(240, 470)
		click(mx+20, my+20)
		wait()
		if captured {
			snap("after-capture")
		}
	}
	if !captured {
		panic("no natural capture")
	}
	boundary := len(actions)
	click(240, 470)
	mx, my := r.BattleMenu.BattleMenuAt(240, 470)
	click(mx+20, my+20+4*r.BattleMenu.Layout.Text.RowPitch)
	if g.State() != game.StateSaveGame {
		panic("public save failed")
	}
	click(300, 152)
	if e := g.SaveErr(); e != nil {
		panic(e)
	}
	snap("public-saved")
	if g.State() != game.StateBattle {
		panic("save did not return to battle")
	}
	click(240, 470)
	click(mx+20, my+20+3*r.BattleMenu.Layout.Text.RowPitch)
	click(300, 132)
	click(318, 228)
	if g.State() != game.StateCommandPhase {
		panic("public load failed")
	}
	click(544, 125)
	click(560, 165)
	click(300, 172)
	if e := g.SaveErr(); e != nil {
		panic(e)
	}
	right(400, 240)
	snap("public-loaded")
	saveActions = append(saveActions, actions[boundary:]...)
	actions = actions[:boundary]
	passed = true
}
