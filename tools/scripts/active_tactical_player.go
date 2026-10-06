//go:build ignore

// #93：正常公開S3與CPU活動抽驗，無direct-entry／局面／RNG注入。
package main

import (
	"crypto/sha256"
	"encoding/binary"
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
	orig := flag.String("orig", "/orig", "唯讀原版")
	data := flag.String("data", "docs/internals/data", "畫面規格")
	out := flag.String("out", "", "新輸出")
	flag.Parse()
	if *out == "" {
		panic("缺輸出")
	}
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	results := map[string]any{}
	for _, mode := range []struct {
		name  string
		level int
	}{{"original", -1}, {"enhanced", 1}} {
		r, err := bootstrap.Load(*orig, *data)
		if err != nil {
			panic(err)
		}
		g, err := game.New(r)
		if err != nil {
			panic(err)
		}
		if err := g.SetAILevel(mode.level); err != nil {
			panic(err)
		}
		folder := filepath.Join(*out, mode.name)
		if err := os.Mkdir(folder, 0755); err != nil {
			panic(err)
		}
		click := func(x, y int) { g.MoveMouse(x, y); g.Click(x, y); g.Frame() }
		click(300, 270)
		for _, k := range []byte("138\r") {
			g.Key(k)
		}
		click(300, 322)
		click(318, 208)
		if g.State() != game.StateCommandPhase {
			panic(fmt.Sprint("公開S3載入：", g.State()))
		}
		click(544, 125)
		click(560, 125)
		for n := 0; n < 20000 && g.State() != game.StateBattle; n++ {
			if g.State() == game.StateTurnReport || g.State() == game.StateOccupationReport {
				click(240, 240)
			} else {
				g.Tick(time.Second / 60)
			}
			g.Frame()
		}
		if g.State() != game.StateBattle {
			panic(fmt.Sprint("正常首戰：", g.State()))
		}
		snapshot := func(name string) map[string]any {
			path := filepath.Join(folder, name)
			if err := os.Mkdir(path, 0755); err != nil {
				panic(err)
			}
			if err := g.SaveSlot(path, 0); err != nil {
				panic(err)
			}
			if err := g.Frame().WritePNG(filepath.Join(path, "frame.png"), g.Palette()); err != nil {
				panic(err)
			}
			files := map[string]string{}
			for _, stem := range []string{"ADMIRAL", "FLEET", "SHIP", "WORK", "STELLAR", "PLANET"} {
				b, err := os.ReadFile(filepath.Join(path, stem+".DT0"))
				if err != nil {
					panic(err)
				}
				files[stem] = fmt.Sprintf("%x", sha256.Sum256(b))
			}
			return map[string]any{"name": name, "state": g.State().String(), "files": files}
		}
		rows := []map[string]any{snapshot("entry")}
		initial, err := os.ReadFile(filepath.Join(folder, "entry", "SHIP.DT0"))
		if err != nil {
			panic(err)
		}
		moved := 0
		lost := 0
		commands := []map[string]int{}
		for turn := 1; turn <= 9 && g.State() == game.StateBattle; turn++ {
			click(100, 100)
			click(100, 110)
			poll := filepath.Join(folder, "poll")
			if err := os.MkdirAll(poll, 0755); err != nil {
				panic(err)
			}
			completed := false
			for n := 0; n < 30000; n++ {
				if g.State() == game.StateTurnReport {
					click(240, 240)
				} else {
					g.Tick(time.Second / 60)
				}
				g.Frame()
				if g.SaveErr() != nil {
					panic(g.SaveErr())
				}
				if n%100 == 0 {
					if err := g.SaveSlot(poll, 0); err != nil {
						panic(err)
					}
					w, err := os.ReadFile(filepath.Join(poll, "WORK.DT0"))
					if err != nil {
						panic(err)
					}
					if g.State() == game.StateCommandPhase || g.State() == game.StateBattle && int(binary.LittleEndian.Uint16(w[6:])) >= turn {
						completed = true
						break
					}
				}
			}
			if !completed {
				panic(fmt.Sprintf("%s第%d回合未返回", mode.name, turn))
			}
			rows = append(rows, snapshot(fmt.Sprintf("turn%d", turn)))
		}
		final := filepath.Join(folder, rows[len(rows)-1]["name"].(string))
		ship, err := os.ReadFile(filepath.Join(final, "SHIP.DT0"))
		if err != nil {
			panic(err)
		}
		fleet, err := os.ReadFile(filepath.Join(final, "FLEET.DT0"))
		if err != nil {
			panic(err)
		}
		for i := 0; i < len(ship)/78; i++ {
			a, b := initial[i*78:(i+1)*78], ship[i*78:(i+1)*78]
			if a[0] == 1 && a[1] != 255 && (binary.LittleEndian.Uint16(a[0x34:]) != binary.LittleEndian.Uint16(b[0x34:]) || binary.LittleEndian.Uint16(a[0x36:]) != binary.LittleEndian.Uint16(b[0x36:])) {
				moved++
			}
			if a[0] == 0 && a[1] != 255 {
				lost += int(binary.LittleEndian.Uint16(a[8:])) - int(binary.LittleEndian.Uint16(b[8:]))
			}
		}
		for i := 0; i < len(fleet)/122; i++ {
			b := fleet[i*122 : (i+1)*122]
			if b[1] == 1 && b[4] != 255 {
				commands = append(commands, map[string]int{"fleet": i, "action": int(b[0x48]), "x": int(binary.LittleEndian.Uint16(b[0x64:])), "y": int(binary.LittleEndian.Uint16(b[0x66:]))})
			}
		}
		if moved == 0 {
			panic(mode.name + "CPU沒有移動")
		}
		results[mode.name] = map[string]any{"rows": rows, "cpu_moved_units": moved, "player_lost": lost, "commands": commands}
		fmt.Println(mode.name, "正常回合", len(rows)-1, "CPU移動分隊", moved, "玩家損失", lost)
	}
	b, err := json.MarshalIndent(map[string]any{"scope": "正常公開S3的兩種戰術AI作用；不宣稱原版戰場對拍", "seed": "無注入，正式初始種子與正常續骰", "results": results}, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(*out, "AI.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
}
