// Command play 走一遍玩家路徑，把每一步的畫面倒成 PNG。
//
// 這是 remake 的「可操作」入口：沒有顯示層也能一步步把輸入送進去、
// 看每一步畫成什麼。腳本用分號分隔，例如
//
//	move 300 170; click; type 138; enter
//
// 原版素材由玩家自備（`-orig`），本專案不內嵌任何原版檔。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

func main() {
	orig := flag.String("orig", "org_game/extracted", "原版檔案目錄（玩家自備）")
	data := flag.String("data", "docs/internals/data", "畫面規格與資料目錄")
	out := flag.String("out", "workplace/play", "輸出目錄")
	script := flag.String("script", "move 300 170; click; type 138; enter",
		"輸入腳本，分號分隔：move X Y／click [X Y]／type 數字／enter／back")
	flag.Parse()

	res, err := load(*orig, *data)
	if err != nil {
		die(err)
	}
	g, err := game.New(res)
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}
	pal := res.Menu.Palette()
	step := 0
	dump := func(label string) {
		fb := g.Frame()
		name := fmt.Sprintf("%02d-%s", step, label)
		if err := fb.WriteIndex(filepath.Join(*out, name+".bin")); err != nil {
			die(err)
		}
		if err := fb.WritePNG(filepath.Join(*out, name+".png"), pal); err != nil {
			die(err)
		}
		fmt.Printf("%2d  %-16s 狀態 %s\n", step, label, stateName(g.State()))
		step++
	}
	dump("start")
	for _, cmd := range strings.Split(*script, ";") {
		f := strings.Fields(strings.TrimSpace(cmd))
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "move":
			if len(f) != 3 {
				die(fmt.Errorf("move 要兩個座標：%q", cmd))
			}
			x, _ := strconv.Atoi(f[1])
			y, _ := strconv.Atoi(f[2])
			g.MoveMouse(x, y)
		case "click":
			g.Click(mouseOf(f, g))
		case "type":
			if len(f) != 2 {
				die(fmt.Errorf("type 要一串數字：%q", cmd))
			}
			for _, c := range []byte(f[1]) {
				g.Key(c)
			}
		case "enter":
			g.Key('\r')
		case "back":
			g.Key(8)
		default:
			die(fmt.Errorf("看不懂的指令：%q", cmd))
		}
		dump(strings.ReplaceAll(strings.TrimSpace(cmd), " ", "-"))
	}
}

// mouseOf 讓 `click` 可以自己帶座標，也可以沿用上一次 move 的位置。
func mouseOf(f []string, g *game.Game) (int, int) {
	if len(f) == 3 {
		x, _ := strconv.Atoi(f[1])
		y, _ := strconv.Atoi(f[2])
		return x, y
	}
	return g.Mouse()
}

func stateName(s game.State) string {
	switch s {
	case game.StateCopyProtection:
		return "防拷問答"
	case game.StateMainGame:
		return "遊戲本體"
	case game.StateShipOrder:
		return "船艦訂造報告"
	case game.StateAdmiralList:
		return "admiral-list"
	case game.StateBudgetPlan:
		return "budget-plan"
	case game.StateBuildPlan:
		return "船艦建造計劃"
	default:
		return "主選單"
	}
}

func load(orig, data string) (*game.Resources, error) {
	fnt, err := font.Load(filepath.Join(orig, "GRAPH.IMG"))
	if err != nil {
		return nil, err
	}
	faceData, err := os.ReadFile(filepath.Join(orig, "FACE.AMG"))
	if err != nil {
		return nil, err
	}
	faces, err := asset.OpenAMG(faceData)
	if err != nil {
		return nil, err
	}
	admiral, err := os.ReadFile(filepath.Join(orig, "ADMIRAL.DTX"))
	if err != nil {
		return nil, err
	}
	menu, err := screen.LoadMenuSpec(filepath.Join(data, "screen-main-menu.json"))
	if err != nil {
		return nil, err
	}
	cp, err := screen.LoadCopyProtSpec(filepath.Join(data, "screen-copy-protection.json"))
	if err != nil {
		return nil, err
	}
	mg, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-main-game.json"))
	if err != nil {
		return nil, err
	}
	so, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-ship-order.json"))
	if err != nil {
		return nil, err
	}
	bp, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-build-plan.json"))
	if err != nil {
		return nil, err
	}
	bg, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-budget-plan.json"))
	if err != nil {
		return nil, err
	}
	al, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-admiral-list.json"))
	if err != nil {
		return nil, err
	}
	cur, err := ui.LoadCursor(filepath.Join(data, "cursor.json"))
	if err != nil {
		return nil, err
	}
	bit, err := os.ReadFile(filepath.Join(orig, "STRMAP.BIT"))
	if err != nil {
		return nil, err
	}
	rgn, err := os.ReadFile(filepath.Join(orig, "STRMAP.RGN"))
	if err != nil {
		return nil, err
	}
	colours, err := regionColours(filepath.Join(data, "strmap-regions.json"))
	if err != nil {
		return nil, err
	}
	star, err := asset.BuildStarMap(bit, rgn, colours, 16, 20, 20)
	if err != nil {
		return nil, err
	}
	return &game.Resources{Font: fnt, Faces: faces, Cursor: cur, Menu: menu,
		CopyProt: cp, MainGame: mg, StarMap: star, Admiral: admiral,
		ShipOrder: so, BuildPlan: bp, BudgetPlan: bg, AdmiralList: al}, nil
}

func regionColours(path string) ([]uint8, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Regions []struct {
			Colour uint8 `json:"colour_796_01"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make([]uint8, len(doc.Regions))
	for i, r := range doc.Regions {
		out[i] = r.Colour
	}
	return out, nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "錯誤：", err)
	os.Exit(1)
}
