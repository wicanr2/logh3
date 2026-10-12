//go:build ignore

// 規格122：正式Game與HD合成器產生正常操作的預期畫面。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
)

type event struct {
	Kind    string   `json:"kind"`
	X       int      `json:"x,omitempty"`
	Y       int      `json:"y,omitempty"`
	Button  int      `json:"button,omitempty"`
	Key     string   `json:"key,omitempty"`
	Name    string   `json:"name,omitempty"`
	State   string   `json:"state,omitempty"`
	Artwork []string `json:"artwork,omitempty"`
	Cursors [][4]int `json:"cursors,omitempty"`
}

func main() {
	orig := flag.String("orig", "/orig", "原版目錄")
	data := flag.String("data", "docs/internals/data", "規格目錄")
	out := flag.String("out", "", "新輸出目錄")
	faction := flag.Int("faction", 0, "陣營")
	art := flag.String("art", "assets/hd", "HD目錄")
	graphics := flag.String("graphics", "hd", "original或hd")
	flag.Parse()
	if *out == "" || (*faction < 0 || *faction > 2) || (*graphics != "hd" && *graphics != "original") {
		panic("參數")
	}
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	r, err := bootstrap.Load(*orig, *data)
	if err != nil {
		panic(err)
	}
	var hd *presentation.Renderer
	if *graphics == "hd" {
		hd, err = presentation.Load(*art, r.Faces)
		if err != nil {
			panic(err)
		}
		r.Faces.ArtworkPrefix = "face"
	}
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	var events []event
	click := func(x, y, b int) {
		if b == 3 {
			g.RightClick(x, y)
		} else {
			g.Click(x, y)
			g.LeftRelease()
		}
		g.Frame()
		// 正式app每次Update都接收目前位置，即使按下後滑鼠未再移動。
		g.MoveMouse(x, y)
		g.Frame()
		events = append(events, event{Kind: "click", X: x, Y: y, Button: b})
	}
	move := func(x, y int) { g.MoveMouse(x, y); g.Frame(); events = append(events, event{Kind: "move", X: x, Y: y}) }
	snap := func(name string, state game.State) {
		if g.State() != state {
			panic(fmt.Sprintf("%s: %v != %v", name, g.State(), state))
		}
		var err error
		if hd != nil {
			err = hd.WritePNG(filepath.Join(*out, name+".png"), g.Frame(), g.Palette())
		} else {
			err = g.Frame().WritePNG(filepath.Join(*out, name+".png"), g.Palette())
		}
		if err != nil {
			panic(err)
		}
		var artwork []string
		for _, layer := range g.Frame().Artwork {
			artwork = append(artwork, layer.Key)
		}
		var cursors [][4]int
		for _, command := range g.Frame().Commands {
			if command.Kind == gfx.DrawCursor {
				cursors = append(cursors, [4]int{command.X, command.Y, command.W, command.H})
			}
		}
		events = append(events, event{Kind: "capture", Name: name, State: state.String(), Artwork: artwork, Cursors: cursors})
	}
	click(300, 210+20**faction, 1)
	for _, c := range []byte("138") {
		g.Key(c)
		key := string(c)
		if c == '\r' {
			key = "Return"
		}
		events = append(events, event{Kind: "key", Key: key})
	}
	// 正常玩家實際點擊第二畫面的確認，不能用Enter避開漏接。
	button := r.CopyProt.InputBox
	click(button.X+button.W/2, button.Y+button.H/2, 1)
	account, order := game.StateMainGame, game.StateShipOrder
	if *faction == 2 {
		snap("00-handoff", game.StateVersusInput)
		click(432, 270, 1)
		account, order = game.StateVersusAccount, game.StateVersusShipOrder
	}
	snap("01-account", account)
	click(448, 350, 1)
	snap("02-budget", game.StateBudgetPlan)
	click(487, 329, 1)
	snap("03-budget-changed", game.StateBudgetPlan)
	click(463, 370, 1)
	snap("04-account-return", account)
	click(448, 370, 1)
	snap("05-order", order)
	move(600, 420)
	snap("06-order-cursor", order)
	click(448, 350, 1)
	snap("07-build", game.StateBuildPlan)
	click(232, 210, 1)
	snap("08-build-row0", game.StateBuildPlan)
	click(232, 230, 1)
	snap("09-build-row1", game.StateBuildPlan)
	move(600, 420)
	snap("10-build-cursor", game.StateBuildPlan)
	click(512, 370, 1)
	snap("11-order-return", order)
	click(448, 370, 1)
	snap("12-command", game.StateCommandPhase)
	for _, m := range []struct {
		name         string
		x            int
		menu, target game.State
	}{{"strategy", 352, game.StateCmdMenu, game.StateTaxPlan}, {"lists", 448, game.StateListMenu, game.StateAdmiralListView}, {"system", 544, game.StateSysMenu, game.StateSettings}} {
		click(m.x, 125, 1)
		snap(m.name+"-open", m.menu)
		move(m.x+12, 165)
		snap(m.name+"-moved", m.menu)
		y := 165
		if m.x == 544 {
			y = 185
		}
		click(m.x+16, y, 1)
		snap(m.name+"-selected", m.target)
		if m.target == game.StateSettings {
			r := game.EnemyVisibilityButtonBounds()
			click((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2, 1)
			if !g.ShowEnemies() {
				panic("設定鈕未開啟敵軍顯示")
			}
			snap("system-enemies-on", game.StateSettings)
			g.ToggleEnemyVisibility()
			events = append(events, event{Kind: "key", Key: "F7"})
			if g.ShowEnemies() {
				panic("F7未關閉敵軍顯示")
			}
			snap("system-enemies-off", game.StateSettings)
		}
		click(600, 400, 3)
		snap(m.name+"-closed", game.StateCommandPhase)
	}
	click(448, 125, 1)
	click(464, 165, 1)
	snap("13-admiral-list", game.StateAdmiralListView)
	rowY := 90
	if *faction == 1 {
		rowY = 330
	}
	click(400, rowY, 1)
	snap("14-admiral-detail", game.StateAdmiralListView)
	move(465, 145)
	snap("15-portrait-cursor", game.StateAdmiralListView)
	click(600, 400, 3)
	snap("16-detail-return", game.StateAdmiralListView)
	// 新批次的真實人物也必須從正常一覽路徑到達。
	secondY := 130
	expectedKey := "face/02"
	if *faction == 1 {
		secondY = 110
		expectedKey = "face/41"
	}
	click(400, secondY, 1)
	snap("17-new-portrait", game.StateAdmiralListView)
	found := false
	for _, layer := range g.Frame().Artwork {
		if layer.Key == expectedKey {
			found = true
		}
	}
	if !found && hd != nil {
		panic("正常一覽未到達新人物：" + expectedKey)
	}
	move(465, 145)
	snap("18-new-portrait-cursor", game.StateAdmiralListView)
	click(600, 400, 3)
	snap("19-new-detail-return", game.StateAdmiralListView)
	click(600, 400, 3)
	snap("20-command-return", game.StateCommandPhase)
	b, err := json.MarshalIndent(map[string]any{"faction": *faction, "events": events, "scope": "remake正常輸入与renderer一致性，不宣稱原版parity"}, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(*out, "PLAN.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
}
