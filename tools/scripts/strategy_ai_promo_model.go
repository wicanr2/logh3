//go:build ignore

// 正常新局、實行、存讀及換陣營檢視電腦命令，不注入局面。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	out := flag.String("out", "", "新批次")
	flag.Parse()
	must := func(e error) {
		if e != nil {
			panic(e)
		}
	}
	if *out == "" {
		panic("缺輸出")
	}
	must(os.Mkdir(*out, 0755))
	results := map[string]any{}
	for _, mode := range []struct {
		name  string
		level int
	}{{"original", -1}, {"enhanced", 1}} {
		folder := filepath.Join(*out, mode.name)
		must(os.Mkdir(folder, 0755))
		r, e := bootstrap.Load("/orig", "docs/internals/data")
		must(e)
		must(bootstrap.UseSaveDir(r, filepath.Join(folder, "save")))
		hd, e := presentation.Load("assets/hd", r.Faces)
		must(e)
		r.Faces.ArtworkPrefix = "face"
		g, e := game.New(r)
		must(e)
		must(g.SetAILevel(mode.level))
		var events []map[string]any
		click := func(x, y, b int) {
			g.MoveMouse(x, y)
			if b == 3 {
				g.RightClick(x, y)
			} else {
				g.Click(x, y)
				g.LeftRelease()
			}
			g.Frame()
			g.MoveMouse(x, y)
			g.Frame()
			events = append(events, map[string]any{"kind": "click", "x": x, "y": y, "button": b})
		}
		move := func(x, y int) {
			g.MoveMouse(x, y)
			g.Frame()
			g.MoveMouse(x, y)
			g.Frame()
			events = append(events, map[string]any{"kind": "move", "x": x, "y": y})
		}
		snap := func(name string, save bool) {
			must(hd.WritePNG(filepath.Join(folder, name+".png"), g.Frame(), g.Palette()))
			events = append(events, map[string]any{"kind": "capture", "name": name, "wait": 2})
			if save {
				d := filepath.Join(folder, name)
				must(os.Mkdir(d, 0755))
				must(g.SaveSlot(d, 0))
			}
		}
		click(300, 210, 1)
		for _, k := range []byte("138\r") {
			g.Key(k)
			key := string(k)
			if k == '\r' {
				key = "Return"
			}
			events = append(events, map[string]any{"kind": "key", "key": key})
		}
		click(300, 240, 3)
		click(300, 240, 3)
		move(600, 340)
		click(352, 125, 1)
		click(400, 125, 1)
		click(400, 85, 1)
		click(488, 330, 1)
		click(480, 350, 1)
		click(510, 350, 1)
		if g.State() != game.StateCommandPhase {
			panic(g.State())
		}
		// 新局先經正常UI保存，診斷快照不額外建立只有模型才有的slot底稿。
		click(544, 125, 1)
		click(560, 165, 1)
		click(300, 142, 1)
		must(g.SaveErr())
		click(300, 142, 3)
		move(600, 340)
		snap("initial", true)
		click(544, 125, 1)
		click(560, 125, 1)
		for n := 0; n < 18; n++ {
			g.Tick(g.TimerTick())
			g.Frame()
		}
		if g.State() != game.StateCommandPhase {
			panic(g.State())
		}
		move(600, 340)
		events = append(events, map[string]any{"kind": "wait_image", "reference": filepath.Join("/src", folder, "after-turn.png"), "box": []int{1000, 120, 1880, 420}, "timeout": 60})
		snap("after-turn", true)
		// 正常保存與載入後選同盟，檢視剛才電腦的實際部署，不再推進回合。
		click(544, 125, 1)
		click(560, 165, 1)
		click(300, 142, 1)
		must(g.SaveErr())
		click(300, 142, 3)
		click(544, 125, 1)
		click(560, 145, 1)
		snap("load-list", false)
		click(300, 122, 1)
		events = append(events, map[string]any{"kind": "wait_image", "reference": filepath.Join("/src", folder, "load-faction.png"), "box": []int{0, 0, 1920, 1440}, "timeout": 30})
		snap("load-faction", false)
		click(318, 228, 1)
		if g.State() != game.StateCommandPhase {
			panic(g.State())
		}
		move(600, 340)
		snap("cpu-view", false)
		click(350, 290, 1)
		move(600, 340)
		snap("cpu-fleets", false)
		var commands [][4]int
		f, e := os.ReadFile(filepath.Join(folder, "after-turn", "FLEET.DT0"))
		must(e)
		for i := 0; i < 32; i++ {
			row := f[i*122 : (i+1)*122]
			if row[4] != 255 && row[1] == 1 {
				commands = append(commands, [4]int{i, int(row[6]), int(row[7]), int(row[8])})
			}
		}
		results[mode.name] = map[string]any{"commands": commands, "fleet_sha256": fmt.Sprintf("%x", sha256.Sum256(f))}
		b, e := json.MarshalIndent(events, "", "  ")
		must(e)
		must(os.WriteFile(filepath.Join(folder, "PLAN.json"), append(b, '\n'), 0644))
		fmt.Println(mode.name, commands)
	}
	for _, s := range []string{"ADMIRAL", "FLEET", "SHIP", "WORK", "STELLAR", "PLANET"} {
		a, e := os.ReadFile(filepath.Join(*out, "original", "initial", s+".DT0"))
		must(e)
		b, e := os.ReadFile(filepath.Join(*out, "enhanced", "initial", s+".DT0"))
		must(e)
		if !bytes.Equal(a, b) {
			panic("初始不同 " + s)
		}
	}
	b, e := json.MarshalIndent(map[string]any{"scope": "remake兩模式正常新局戰略命令；以正常存讀換陣營檢視，沒有再次實行", "same_initial_tables": true, "state_injection": false, "seed_injection": false, "results": results}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(*out, "STRATEGY.json"), append(b, '\n'), 0644))
}
