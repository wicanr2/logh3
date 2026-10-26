//go:build ignore

// 產生人事單次點擊的正常入口劇本與正式renderer完整畫面。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/presentation"
)

type event struct {
	Kind         string
	X, Y, Button int
	Key, Name    string
}

func main() {
	out := flag.String("out", "", "新輸出目錄")
	faction := flag.Int("faction", 1, "陣營")
	flag.Parse()
	if _, err := os.Stat(*out); !os.IsNotExist(err) {
		panic("輸出必須是新目錄")
	}
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	res, err := bootstrap.Load("/orig", "docs/internals/data")
	must(err)
	g, err := game.New(res)
	must(err)
	r, err := presentation.Load("assets/hd", res.Faces)
	must(err)
	res.Faces.ArtworkPrefix = "face"
	var events []event
	click := func(x, y, button int) {
		if button == 1 {
			g.Click(x, y)
		} else {
			g.RightClick(x, y)
		}
		events = append(events, event{Kind: "click", X: x, Y: y, Button: button})
	}
	snap := func(name string) {
		g.MoveMouse(600, 440)
		must(r.WritePNG(filepath.Join(*out, name+".png"), g.Frame(), g.Palette()))
		events = append(events, event{Kind: "capture", Name: name})
	}
	click(300, 210+20**faction, 1)
	for _, k := range []byte("138\r") {
		g.Key(k)
		name := string(k)
		if k == '\r' {
			name = "Return"
		}
		events = append(events, event{Kind: "key", Key: name})
	}
	click(300, 240, 3)
	click(300, 240, 3)
	click(352, 125, 1)
	click(400, 145, 1)
	click(543, 388, 1)
	snap("staff-list")
	for i := 0; i < 5; i++ {
		click(450, 92, 1)
		snap(fmt.Sprintf("staff-detail-%d", i))
		click(448, 330, 1)
		snap(fmt.Sprintf("staff-cancel-%d", i))
	}
	click(450, 112, 1)
	click(496, 330, 1)
	snap("staff-decided")
	click(300, 240, 3)
	click(352, 125, 1)
	click(400, 125, 1)
	click(400, 85, 1)
	click(488, 330, 1)
	for i := 0; i < 5; i++ {
		click(430, 350, 1)
		click(430, 350, 1)
		snap(fmt.Sprintf("adj-list-%d", i))
		click(450, 102, 1)
		snap(fmt.Sprintf("adj-detail-%d", i))
		click(448, 330, 1)
		snap(fmt.Sprintf("adj-cancel-%d", i))
	}
	click(430, 350, 1)
	click(430, 350, 1)
	click(450, 102, 1)
	click(496, 330, 1)
	snap("adj-decided")
	encoded, err := json.MarshalIndent(events, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(*out, "PLAN.json"), encoded, 0644))
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
