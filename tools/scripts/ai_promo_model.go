//go:build ignore

// 第126輪：正常公開S3兩種AI的模型、完整六表與錄影輸入。不注入局面或seed。
package main

import (
	"bytes"
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
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	out := flag.String("out", "", "新的模型批次")
	turns := flag.Int("turns", 9, "正常戰術回合數")
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
	stems := []string{"ADMIRAL", "FLEET", "SHIP", "WORK", "STELLAR", "PLANET"}
	results := map[string]any{}
	for _, mode := range []struct {
		name  string
		level int
	}{{"original", -1}, {"enhanced", 1}} {
		folder := filepath.Join(*out, mode.name)
		must(os.Mkdir(folder, 0755))
		r, e := bootstrap.Load("/orig", "docs/internals/data")
		must(e)
		hd, e := presentation.Load("assets/hd", r.Faces)
		must(e)
		r.Faces.ArtworkPrefix = "face"
		g, e := game.New(r)
		must(e)
		must(g.SetAILevel(mode.level))
		var events []map[string]any
		click := func(x, y int) {
			g.MoveMouse(x, y)
			g.Click(x, y)
			g.LeftRelease()
			g.Frame()
			g.MoveMouse(x, y)
			g.Frame()
			events = append(events, map[string]any{"kind": "click", "x": x, "y": y})
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
			events = append(events, map[string]any{"kind": "capture", "name": name})
			if save {
				dir := filepath.Join(folder, name)
				must(os.Mkdir(dir, 0755))
				must(g.SaveSlot(dir, 0))
			}
		}
		click(300, 270)
		for _, k := range []byte("138\r") {
			g.Key(k)
			key := string(k)
			if k == '\r' {
				key = "Return"
			}
			events = append(events, map[string]any{"kind": "key", "key": key})
		}
		snap("load", false)
		click(300, 322)
		click(318, 208)
		if g.State() != game.StateCommandPhase {
			panic(g.State())
		}
		snap("strategy", true)
		click(544, 125)
		click(560, 185)
		for i := 0; i < 8; i++ {
			click(424, 310)
		}
		click(368, 330)
		click(320, 350)
		click(416, 370)
		click(544, 125)
		click(560, 125)
		for n := 0; n < 30000 && g.State() != game.StateBattle; n++ {
			if g.State() == game.StateTurnReport || g.State() == game.StateOccupationReport {
				click(240, 240)
			} else {
				g.Tick(time.Second / 60)
			}
			g.Frame()
		}
		if g.State() != game.StateBattle {
			panic(g.State())
		}
		move(600, 180)
		phase := func(f *gfx.Framebuffer) []byte {
			var p []byte
			for y := 240; y < 258; y++ {
				p = append(p, f.Pix[y*640+576:y*640+608]...)
			}
			return p
		}
		input := phase(g.Frame())
		events = append(events, map[string]any{"kind": "wait_image", "reference": filepath.Join("/src", folder, "entry-off.png"), "box": []int{1728, 720, 1824, 774}, "timeout": 180})
		snap("entry-off", true)
		for _, name := range []string{"entry-on", "entry-off-restored", "entry-on-final"} {
			g.ToggleEnemyVisibility()
			g.Frame()
			events = append(events, map[string]any{"kind": "key", "key": "F8"})
			snap(name, false)
		}
		initial, e := os.ReadFile(filepath.Join(folder, "entry-off", "WORK.DT0"))
		must(e)
		firstTurn := int(binary.LittleEndian.Uint16(initial[6:]))
		rows := []string{"entry-off"}
		for turn := 1; turn <= *turns && g.State() == game.StateBattle; turn++ {
			click(100, 100)
			click(100, 110)
			move(600, 180)
			completed := false
			for n := 0; n < 30000; n++ {
				if g.State() == game.StateTurnReport {
					g.Click(240, 240)
					g.LeftRelease()
				} else {
					g.Tick(time.Second / 60)
				}
				fb := g.Frame()
				if n%120 == 0 && g.State() == game.StateBattle && bytes.Equal(phase(fb), input) {
					dir := filepath.Join(folder, "poll")
					must(os.MkdirAll(dir, 0755))
					must(g.SaveSlot(dir, 0))
					w, e := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
					must(e)
					if int(binary.LittleEndian.Uint16(w[6:])) >= firstTurn+turn {
						completed = true
						break
					}
				}
				if g.State() == game.StateCommandPhase {
					completed = true
					break
				}
			}
			if !completed {
				panic(fmt.Sprintf("%s回合%d未返回輸入", mode.name, turn))
			}
			name := fmt.Sprintf("turn-%02d", turn)
			box := []int{1728, 720, 1824, 774}
			if g.State() == game.StateCommandPhase {
				box = []int{1000, 120, 1880, 420}
			}
			events = append(events, map[string]any{"kind": "wait_image", "reference": filepath.Join("/src", folder, name+".png"), "box": box, "timeout": 240, "wait": 1.0})
			move(600, 180)
			snap(name, true)
			rows = append(rows, name)
			fmt.Println(mode.name, name, g.State())
		}
		files := map[string]any{}
		for _, row := range rows {
			f := map[string]string{}
			for _, stem := range stems {
				b, e := os.ReadFile(filepath.Join(folder, row, stem+".DT0"))
				must(e)
				f[stem] = fmt.Sprintf("%x", sha256.Sum256(b))
			}
			files[row] = f
		}
		write := func(name string, v any) {
			b, e := json.MarshalIndent(v, "", "  ")
			must(e)
			must(os.WriteFile(filepath.Join(folder, name), append(b, '\n'), 0644))
		}
		write("PLAN.json", events)
		write("STATES.json", map[string]any{"mode": mode.name, "level": mode.level, "completed_turns_at_entry": firstTurn, "rows": rows, "files": files, "normal_public_input": true, "state_injection": false, "seed_injection": false})
		results[mode.name] = files
	}
	for _, stage := range []string{"strategy", "entry-off"} {
		for _, stem := range stems {
			a, e := os.ReadFile(filepath.Join(*out, "original", stage, stem+".DT0"))
			must(e)
			b, e := os.ReadFile(filepath.Join(*out, "enhanced", stage, stem+".DT0"))
			must(e)
			if !bytes.Equal(a, b) {
				panic("初始資料不同 " + stage + "/" + stem)
			}
		}
	}
	b, e := json.MarshalIndent(map[string]any{"scope": "remake兩種AI；正常S3，同初始六表、正式固定初始seed；不宣稱DOS整場同態", "strategy_seed": "0xD3DC167E，產品預設，無重擲", "same_initial_tables": true, "results": results}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(*out, "AI.json"), append(b, '\n'), 0644))
}
