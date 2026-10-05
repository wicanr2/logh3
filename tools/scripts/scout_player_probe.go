//go:build ignore

// 正式bootstrap與公開輸入；只讀觀測器由scout_player_prepare.py建立。
package main

import (
	"bytes"
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
	out := flag.String("out", "", "獨立輸出目錄，禁止覆寫")
	flag.Parse()
	if *out == "" {
		panic("缺輸出路徑")
	}
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		panic("拒絕覆寫")
	}
	if e := os.Mkdir(*out, 0755); e != nil {
		panic(e)
	}
	r, e := bootstrap.Load("org_game/extracted", "docs/internals/data")
	if e != nil {
		panic(e)
	}
	if e = bootstrap.LoadIntro(r, "org_game/extracted"); e != nil {
		panic(e)
	}
	if e = bootstrap.UseSaveDir(r, filepath.Join(*out, "player-save")); e != nil {
		panic(e)
	}
	g, e := game.New(r)
	if e != nil {
		panic(e)
	}
	if e = g.SetAILevel(-1); e != nil {
		panic(e)
	}
	var inputs []map[string]any
	var observations []map[string]any
	var actions []map[string]any
	reports := 0
	x, y := 0, 0
	step := func(kind string, px, py int, key byte) {
		before := g.FleetGUIStamp()
		inputs = append(inputs, map[string]any{"kind": kind, "x": px, "y": py, "key": int(key)})
		switch kind {
		case "click":
			x, y = px, py
			g.MoveMouse(x, y)
			g.Click(x, y)
			g.LeftRelease()
		case "right":
			x, y = px, py
			g.RightClick(x, y)
			g.RightRelease()
		case "key":
			g.Key(key)
		case "move":
			x, y = px, py
			g.MoveMouse(x, y)
		case "tick":
			g.Tick(time.Second / 60)
		default:
			panic("未知輸入")
		}
		g.MoveMouse(x, y)
		g.Frame()
		if e = g.SaveErr(); e != nil {
			panic(e)
		}
		if kind != "tick" {
			actions = append(actions, map[string]any{"kind": kind, "x": px, "y": py, "key": int(key), "seed": 0, "before": before, "after": g.FleetGUIStamp()})
		}
	}
	snapshot := func(name string) {
		s := g.ScoutPlayerState()
		observations = append(observations, map[string]any{"name": name, "input": len(inputs), "state": s})
		actions = append(actions, map[string]any{"kind": "wait_scout", "scout": g.ScoutPlayerStamp(), "snapshot": name})
		if e = g.Frame().WritePNG(filepath.Join(*out, name+".png"), g.Palette()); e != nil {
			panic(e)
		}
		if e = g.Frame().WriteIndex(filepath.Join(*out, name+".bin")); e != nil {
			panic(e)
		}
	}
	defer func() {
		b, e := json.MarshalIndent(map[string]any{"inputs": inputs, "observations": observations, "final": g.ScoutPlayerState(), "seed": uint32(0x12345678), "seed_method": "首次戰場實行前設定一次", "classification": "remake正常bootstrap及公開輸入；原版完整流程未驗"}, "", "  ")
		if e != nil {
			panic(e)
		}
		if e = os.WriteFile(filepath.Join(*out, "receipt.json"), append(b, '\n'), 0644); e != nil {
			panic(e)
		}
		b, e = json.MarshalIndent(map[string]any{"classification": "正常bootstrap公開輸入計畫，正式60Hz GUI另驗", "faction": 0, "actions": actions, "scroll_points": scrollPoints, "final": g.FleetGUIStamp()}, "", "  ")
		if e != nil {
			panic(e)
		}
		if e = os.WriteFile(filepath.Join(*out, "plan.json"), append(b, '\n'), 0644); e != nil {
			panic(e)
		}
	}()
	step("click", 300, 210, 0) // 跳過片頭。
	step("click", 300, 210, 0)
	for _, k := range []byte("138\r") {
		step("key", 0, 0, k)
	}
	step("right", 300, 240, 0)
	step("right", 300, 240, 0)
	for _, p := range [][2]int{{544, 125}, {560, 145}, {300, 322}, {318, 208}, {544, 125}, {560, 125}} {
		step("click", p[0], p[1], 0)
	}
	for n := 0; !g.ScoutPlayerState().Idle && n < 20000; n++ {
		step("tick", 0, 0, 0)
	}
	if !g.ScoutPlayerState().Idle {
		panic("正常S3首戰未到達")
	}
	g.ScoutPlayerSeed(0x12345678)
	actions = append(actions, map[string]any{"kind": "seed", "x": 0, "y": 0, "key": 0, "seed": uint32(0x12345678), "before": g.FleetGUIStamp(), "after": g.FleetGUIStamp()})
	inputs = append(inputs, map[string]any{"kind": "seed", "seed": uint32(0x12345678)})
	snapshot("entry")
	fld := r.Battle.Field
	point := func(u rules.BattleUnit) (int, int) {
		v := g.ScoutPlayerState().View
		return fld.X + u.X - v[0]*fld.TileW + 8, fld.Y + u.Y - v[1]*fld.TileH + 8
	}
	leader := -1
	for i, u := range g.ScoutPlayerState().Units {
		if u.Faction == 0 && u.Squadron == 0 && !u.Out {
			leader = i
			break
		}
	}
	if leader < 0 {
		panic("缺己方旗艦")
	}
	u := g.ScoutPlayerState().Units[leader]
	fleet := u.Fleet
	// 用公開右鍵把旗艦放在視野右下角，擴大合法目的地距離。
	want := [2]int{max(0, (u.X-(fld.Cols*fld.TileW-24))/fld.TileW), max(0, (u.Y-(fld.Rows*fld.TileH-24))/fld.TileH)}
	points := map[[2]int][2]int{}
	for py := fld.Y + 1; py < fld.Y+fld.Rows*fld.TileH; py += 16 {
		for px := fld.X + 1; px < fld.X+fld.Cols*fld.TileW; px += 16 {
			col, row, moved := rules.ViewScrollStep(10, 10, rules.ViewScrollDir(px, py-ui.FrameCellOriginY))
			if moved {
				points[[2]int{col - 10, row - 10}] = [2]int{px, py}
			}
		}
	}
	for dir, p := range points {
		scrollPoints[fmt.Sprintf("%d,%d", dir[0], dir[1])] = p
	}
	scroll := func(target [2]int) {
		for n := 0; n < 200; n++ {
			v := g.ScoutPlayerState().View
			if v == target {
				return
			}
			dir := [2]int{}
			for k := 0; k < 2; k++ {
				if target[k] > v[k] {
					dir[k] = 1
				}
				if target[k] < v[k] {
					dir[k] = -1
				}
			}
			p, ok := points[dir]
			if !ok {
				panic("缺捲動方向")
			}
			step("right", p[0], p[1], 0)
		}
		panic("鏡頭未到達")
	}
	scroll(want)
	px, py := point(u)
	step("click", px, py, 0)
	s := g.ScoutPlayerState()
	if !s.Command || s.Selected != leader || len(s.Enabled) < 3 || !s.Enabled[2] {
		panic("索敵選单不可用")
	}
	step("click", s.CommandPoint[0], s.CommandPoint[1]+2*r.CommandMenus.Layout.RowPitch, 0)
	if !g.ScoutPlayerState().Move {
		panic("索敵未派出")
	}
	tx, ty, best := -1, -1, -1
	v := g.ScoutPlayerState().View
	for yy := fld.Y + 8; yy < fld.Y+fld.Rows*fld.TileH; yy += 16 {
		for xx := fld.X + 8; xx < fld.X+fld.Cols*fld.TileW; xx += 16 {
			wx, wy := xx-fld.X+v[0]*fld.TileW, yy-fld.Y+v[1]*fld.TileH
			if !r.Terrain.LineOpen(g.ScoutPlayerCells(), u.X, u.Y, wx, wy) {
				continue
			}
			d := (wx-u.X)*(wx-u.X) + (wy-u.Y)*(wy-u.Y)
			if d > best {
				tx, ty, best = xx, yy, d
			}
		}
	}
	if tx < 0 {
		panic("無合法目的地")
	}
	step("click", tx, ty, 0)
	snapshot("deployed")
	fmt.Println("旗艦", fleet, "起點", u.X, u.Y, "目的地", tx, ty, "距離平方", best)
	for round := 1; round <= 7; round++ {
		if round == 2 {
			// 索敵隊已離開旗艦。公開移動母艦，讓它不能在第四回合前歸隊。
			u = g.ScoutPlayerState().Units[leader]
			scroll([2]int{u.X / fld.TileW, u.Y / fld.TileH})
			px, py = point(u)
			step("click", px, py, 0)
			s = g.ScoutPlayerState()
			if !s.Command || s.Selected != leader {
				panic("移動母艦選單未開")
			}
			step("click", s.CommandPoint[0], s.CommandPoint[1], 0)
			if g.FleetOrderStage() != game.FleetOrderPoint {
				panic("母艦移動未選點")
			}
			v = g.ScoutPlayerState().View
			tx, ty, best = -1, -1, -1
			for yy := fld.Y + 8; yy < fld.Y+fld.Rows*fld.TileH; yy += 16 {
				for xx := fld.X + 8; xx < fld.X+fld.Cols*fld.TileW; xx += 16 {
					wx, wy := xx-fld.X+v[0]*fld.TileW, yy-fld.Y+v[1]*fld.TileH
					if !r.Terrain.LineOpen(g.ScoutPlayerCells(), u.X, u.Y, wx, wy) {
						continue
					}
					d := (wx-u.X)*(wx-u.X) + (wy-u.Y)*(wy-u.Y)
					if d > best {
						tx, ty, best = xx, yy, d
					}
				}
			}
			if tx < 0 {
				panic("無母艦移動目的地")
			}
			step("click", tx, ty, 0)
			if g.FleetOrderStage() != game.FleetOrderFacing {
				panic("母艦朝向未開")
			}
			step("click", tx, ty, 0)
			if g.FleetOrderStage() != game.FleetOrderNone {
				panic("母艦命令未提交")
			}
			snapshot("mother-moving")
		}
		s = g.ScoutPlayerState()
		before := s.Turn
		// 找空白戰場點，打開正常「實行」選單。
		opened := false
		for yy := fld.Y + 8; yy < fld.Y+fld.Rows*fld.TileH && !opened; yy += 32 {
			for xx := fld.X + 8; xx < fld.X+fld.Cols*fld.TileW && !opened; xx += 32 {
				if !g.ScoutPlayerEmptyAt(xx, yy) {
					continue
				}
				step("click", xx, yy, 0)
				opened = g.ScoutPlayerState().Menu
			}
		}
		if !opened {
			panic("戰場選單未開")
		}
		s = g.ScoutPlayerState()
		step("click", s.MenuPoint[0], s.MenuPoint[1], 0)
		for n := 0; n < 20000; n++ {
			s = g.ScoutPlayerState()
			if s.Idle && s.Turn == before+1 {
				break
			}
			if s.ReportText != "" {
				reports++
				if round != 4 || reports != 1 || s.ReportFace != 12 {
					panic("失聯報告路由不符")
				}
				for _, f := range s.Fleets {
					if f.Index == fleet && (f.ScoutMode != 0 || f.ScoutState != 4) {
						panic("報告前已回收")
					}
				}
				step("move", 240, 240, 0)
				snapshot(fmt.Sprintf("report-%d", round))
				step("click", 240, 240, 0)
			} else {
				step("tick", 0, 0, 0)
			}
			if n == 19999 {
				panic("完整實行未返回")
			}
		}
		snapshot(fmt.Sprintf("turn-%d", round))
		for _, f := range g.ScoutPlayerState().Fleets {
			if f.Index == fleet {
				mode, fuel := 0, round
				if round >= 4 {
					mode, fuel = 255, 7-round
				}
				if f.ScoutMode != mode || f.ScoutState != fuel || f.ScoutOut != (round != 7) {
					panic("正常燃料／冷卻不符")
				}
				fmt.Printf("round%d turn%d mode%d fuel%d out%v\n", round, g.ScoutPlayerState().Turn, f.ScoutMode, f.ScoutState, f.ScoutOut)
			}
		}
		if round == 4 {
			if e = os.Mkdir(filepath.Join(*out, "cooldown-save"), 0755); e != nil {
				panic(e)
			}
			if e = g.SaveSlot(filepath.Join(*out, "cooldown-save"), 0); e != nil {
				panic(e)
			}
		}
	}
	if reports != 1 {
		panic("缺正常失聯報告")
	}
	beforeUnits := len(g.ScoutPlayerState().Units)
	u = g.ScoutPlayerState().Units[leader]
	scroll([2]int{u.X / fld.TileW, u.Y / fld.TileH})
	px, py = point(u)
	step("click", px, py, 0)
	s = g.ScoutPlayerState()
	if !s.Command || s.Selected != leader || !s.Enabled[2] {
		panic("冷卻後不能再派出")
	}
	step("click", s.CommandPoint[0], s.CommandPoint[1]+2*r.CommandMenus.Layout.RowPitch, 0)
	if !g.ScoutPlayerState().Move || len(g.ScoutPlayerState().Units) != beforeUnits {
		panic("再次派出新增池副本")
	}
	// 與母艦相同的可通行點，只提交命令，不再實行。
	step("click", px-8, py-8, 0)
	if g.ScoutPlayerState().Move {
		panic("再派出目的地未提交")
	}
	snapshot("redeployed")
	opened := false
	for yy := fld.Y + 8; yy < fld.Y+fld.Rows*fld.TileH && !opened; yy += 32 {
		for xx := fld.X + 8; xx < fld.X+fld.Cols*fld.TileW && !opened; xx += 32 {
			if g.ScoutPlayerEmptyAt(xx, yy) {
				step("click", xx, yy, 0)
				opened = g.ScoutPlayerState().Menu
			}
		}
	}
	if !opened {
		panic("保存前戰術選單未開")
	}
	s = g.ScoutPlayerState()
	step("click", s.MenuPoint[0], s.MenuPoint[1]+4*r.BattleMenu.Layout.Text.RowPitch, 0)
	if g.State() != game.StateSaveGame {
		panic("公開保存選單未開")
	}
	step("click", 300, r.SlotList.Save.TextTop, 0)
	snapshot("saved")
	step("right", 300, 142, 0)
	if g.State() != game.StateCommandPhase {
		panic("保存未返回命令")
	}
	readSaved := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(*out, "player-save", name+".DT0"))
		if e != nil {
			panic(e)
		}
		return b
	}
	expected, e := rules.LoadSlot(rules.SlotFiles{Stellar: readSaved("STELLAR"), Planet: readSaved("PLANET"), Admiral: readSaved("ADMIRAL"), Fleet: readSaved("FLEET"), Ship: readSaved("SHIP"), Work: readSaved("WORK"), PlanetHeader: r.Planets.Data[:rules.PlanetHeader]})
	if e != nil {
		panic(e)
	}
	// 006 CONFORMED：正常載入重算未安定度。40／可見艦隊契約也重數，
	// 因此不能把STELLAR整檔原封不動當成讀檔驗收。
	params := r.UnrestParams
	params.Turn = expected.Turn
	params.Blend = expected.Blend
	params.Rate = expected.Rate
	for side := 0; side < 2; side++ {
		if id := rules.StaffAdmiral(expected.Work, rules.PostDomestic, side); id >= 0 {
			params.Ruler[side] = id
		}
	}
	expected.RecomputeUnrest(r.Graph, params)
	expectedFleets := expected.Fleets(r.Weights)
	rules.RecountRegionFleets(expected.Stellar, expectedFleets)
	rules.RecountRegionVisibleFleets(expected.Stellar, expectedFleets)
	step("click", 544, 125, 0)
	step("click", 560, 145, 0)
	step("click", 300, 122, 0)
	if g.State() != game.StateLoadFaction {
		panic("公開讀檔未選陣營")
	}
	step("click", 318, 208, 0)
	if g.State() != game.StateCommandPhase {
		panic("讀檔未回命令")
	}
	snapshot("loaded")
	if e = g.SaveSlot(*out, 0); e != nil {
		panic(e)
	}
	for _, name := range rules.SlotFileNames(0) {
		a, e := os.ReadFile(filepath.Join(*out, "player-save", name))
		if e != nil {
			panic(e)
		}
		if name == "STELLAR.DT0" {
			a = expected.Stellar
		}
		b, e := os.ReadFile(filepath.Join(*out, name))
		if e != nil || !bytes.Equal(a, b) {
			panic("六檔讀回重存契約不同 " + name)
		}
	}
	fmt.Println("同池再派出、公開保存／讀檔：五檔相同，STELLAR符合既有載入重算契約")
}

var scrollPoints = map[string][2]int{}
