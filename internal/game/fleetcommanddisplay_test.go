package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

func fleetDisplayManifest(t *testing.T) autoGameReceipt {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/fleet-command-display-native.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "d56ab5516e494b43987aab0d72210b71d2ff93f1b56c9c3576dfba08c255b1db" {
		t.Fatal("第77輪原版來源不符", err)
	}
	var m autoGameReceipt
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// 明示還原原版命令等待；所有分隊由原池解析，不注入原版像素。
func fleetDisplayFixture(t *testing.T, m autoGameReceipt, ram []byte) *Game {
	t.Helper()
	g := fleetChainRAM(t, m, ram)
	data := filepath.Join(root, "docs/internals/data")
	var err error
	g.res.UnitPanel, err = screen.LoadUnitPanelSpec(filepath.Join(data, "unit-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetDetail, err = screen.LoadFleetDetailSpec(filepath.Join(data, "fleet-detail.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetWindow, err = screen.LoadFleetWindowSpec(filepath.Join(data, "fleet-window.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetCommandDisplay, err = screen.LoadFleetCommandDisplaySpec(filepath.Join(data, "fleet-command-display-spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(root, "org_game/extracted", name))
		if os.IsNotExist(e) {
			t.Skip("缺原版素材", name)
		}
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	g.res.BattlePanel, err = asset.DecodeAMGImage(read("ANI_PNL0.A98"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.BattleTiles, err = asset.DecodeTiles(read(g.res.Battle.Field.Tiles), g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
	if err != nil {
		t.Fatal(err)
	}
	copy(g.slot.Planets.Data[rules.PlanetHeader:], ram[0x53132:0x53132+len(g.slot.Planets.Data)-rules.PlanetHeader])
	g.battlePlanets = g.battleRegionPlanets(26)
	byPool := map[int]rules.BattleUnit{}
	for _, f := range g.fleets {
		for _, u := range g.slot.BattleUnitsFor(g.res.Weights, f) {
			byPool[g.battleUnitPool(u)] = u
		}
		if u, ok := g.slot.ScoutUnit(g.res.Weights, f, 0, 0); ok {
			r := g.slot.ShipAt(f.Scout)
			u.X = int(binary.LittleEndian.Uint16(r[0x34:]))
			u.Y = int(binary.LittleEndian.Uint16(r[0x36:]))
			byPool[f.Scout] = u
		}
	}
	const ds = 0x54b30
	g.battleUnits = nil
	g.battleSpotting.sprites = map[int]int{}
	g.battleSpotting.directions = map[int]int{}
	for _, f := range g.battleFleets {
		g.battleSpotting.directions[f.Index] = int(g.slot.Fleet[f.Index*rules.FleetRecord+0x6a])
	}
	own := -1
	for i := 0; i < int(binary.LittleEndian.Uint16(ram[ds+0x9896:])); i++ {
		pool := int(binary.LittleEndian.Uint16(ram[ds+0x98ba+i*2:]))
		u, ok := byPool[pool]
		if !ok {
			t.Fatal("原版物件不存在", pool)
		}
		if pool == 80 {
			own = i
		}
		// 此診斷父狀態在顯示物件建好後才把其他命令設FF。
		// DS9896／98BA保留原物件；依+2F既有快取還原，不從新Action重建可見性。
		u.Out = false
		g.battleUnits = append(g.battleUnits, u)
		g.battleSpotting.sprites[pool] = int(g.slot.ShipAt(pool)[0x2f])
	}
	if own < 0 {
		t.Fatal("缺目前原池80")
	}
	g.battleView = screen.BattleView{Col: int(binary.LittleEndian.Uint16(ram[ds+0x9bf4:])) / g.res.Battle.Field.TileW, Row: int(binary.LittleEndian.Uint16(ram[ds+0x9bf6:])) / g.res.Battle.Field.TileH}
	g.battleSpotting.view = g.battleView
	g.fleetCombat = fleetCombatState{active: true, stage: fleetCommandTarget, current: own, target: -1}
	g.fleetCombat.candidates, err = g.fleetCombatCandidates(own)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func fleetDisplayCompare(t *testing.T, g *Game, want, originalPNG []byte, name string) {
	t.Helper()
	fb := g.Frame()
	if len(want) != len(fb.Pix) {
		t.Fatal("畫布不符")
	}
	if !bytes.Equal(fb.Pix, g.Frame().Pix) {
		t.Fatal("重畫累積XOR")
	}
	pal := g.Palette()
	im, err := png.Decode(bytes.NewReader(originalPNG))
	if err != nil {
		t.Fatal(err)
	}
	cp, ok := im.ColorModel().(color.Palette)
	if !ok || len(cp) < gfx.Colors {
		t.Fatal("原版有效色盤缺失")
	}
	for i := 0; i < gfx.Colors; i++ {
		r, g, b, _ := cp[i].RGBA()
		rr, gg, bb := pal.RGB8(i)
		if uint8(r>>8) != rr || uint8(g>>8) != gg || uint8(b>>8) != bb {
			t.Fatal("原版有效色盤與正式renderer不同", i)
		}
	}
	stats := func(r [4]int) map[string]any {
		n, maximum := 0, 0
		box := [4]int{640, 480, -1, -1}
		for y := r[1]; y < r[1]+r[3]; y++ {
			for x := r[0]; x < r[0]+r[2]; x++ {
				i := y*640 + x
				if fb.Pix[i] == want[i] {
					continue
				}
				n++
				box = [4]int{min(box[0], x), min(box[1], y), max(box[2], x), max(box[3], y)}
				for c := 0; c < 3; c++ {
					a, b := int(pal[fb.Pix[i]][c]), int(pal[want[i]][c])
					a = a<<2 | a>>4
					b = b<<2 | b>>4
					d := a - b
					if d < 0 {
						d = -d
					}
					maximum = max(maximum, d)
				}
			}
		}
		return map[string]any{"rect": r, "pixels": r[2] * r[3], "same": r[2]*r[3] - n, "different": n, "maximum_channel_difference": maximum, "bbox": box}
	}
	full := stats([4]int{0, 0, 640, 480})
	f := g.res.Battle.Field
	s := g.res.Battle.Status
	mini := g.res.Battle.Minimap
	regions := []map[string]any{stats([4]int{f.X, f.Y, f.Cols * f.TileW, f.Rows * f.TileH}), stats([4]int{s.X, s.Y, s.W, s.H}), stats([4]int{mini.X, mini.Y, mini.W, mini.H}), stats([4]int{0, 0, 640, g.res.Battle.Panel.Y}), stats([4]int{0, g.res.Battle.Panel.Y + g.res.BattlePanel.H, 640, 480 - g.res.Battle.Panel.Y - g.res.BattlePanel.H})}
	full["case"], full["comparison_level"], full["regions"] = name, "same-state controlled command display", regions
	full["whole_frame_gate_passed"] = full["different"].(int) == 0
	full["unapproved_differences"] = full["different"]
	full["effective_palette_equal"] = true
	full["remaining_classification"] = "原版亮標word+4未初始化；#89既有忠實度裁決pending，完整對拍未通過"
	if out := os.Getenv("LOGH3_FLEET_DISPLAY_OUTPUT"); out != "" {
		if err := os.MkdirAll(out, 0755); err != nil {
			t.Fatal(err)
		}
		diff := gfx.NewFramebuffer(3)
		original := gfx.NewFramebuffer(0)
		copy(original.Pix, want)
		for i, p := range fb.Pix {
			if p != want[i] {
				diff.Pix[i] = 15
			}
		}
		for label, pic := range map[string]*gfx.Framebuffer{"original": original, "remake": fb, "diff": diff} {
			if err := pic.WritePNG(filepath.Join(out, name+"-"+label+".png"), pal); err != nil {
				t.Fatal(err)
			}
		}
		if err := fb.WriteIndex(filepath.Join(out, name+"-remake.bin")); err != nil {
			t.Fatal(err)
		}
		b, err := json.MarshalIndent(full, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, name+"-comparison.json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// 回歸檢查僅要求已記錄的未核准差異不增加。完整對拍的零差閘門保持FAIL。
	// 不排除、不修改該像素，也不以這個單元測試的PASS宣稱parity。
	u := g.battleUnits[g.fleetCombat.current]
	x, y := f.X+u.X-g.battleView.Col*f.TileW+15, f.Y+u.Y-g.battleView.Row*f.TileH+15
	if full["different"].(int) != 1 || full["bbox"].([4]int) != [4]int{x, y, x, y} {
		t.Fatalf("出現既有亮標順位以外的差異：全圖%v像素；區域%v", full["different"], regions)
	}
	t.Logf("完整對拍FAIL：%v個未核准像素；其餘307199像素及有效色盤相同", full["different"])
}

func TestFleetCommandDisplayNativeRegression(t *testing.T) {
	m := fleetDisplayManifest(t)
	for _, name := range []string{"wait", "hover1", "hover2", "hover3", "popup1", "popup2", "popup3"} {
		t.Run(name, func(t *testing.T) {
			dir := "workplace/oracle/combat79-r77-" + name + "-a/"
			g := fleetDisplayFixture(t, m, m.read(t, dir+"input-mem.bin"))
			before := bytes.Clone(g.slot.Ship)
			fleet, adml, work := bytes.Clone(g.slot.Fleet), bytes.Clone(g.slot.Admiral), bytes.Clone(g.slot.Work)
			seed := g.battleRand.Seed()
			if name == "wait" {
				g.MoveMouse(240, 270)
			} else {
				g.MoveMouse(348, 248)
			}
			if len(name) > 5 && name[:5] == "popup" {
				g.Click(348, 248)
			}
			fleetDisplayCompare(t, g, m.read(t, dir+"screen.bin"), m.read(t, dir+"screen.png"), name)
			if !bytes.Equal(before, g.slot.Ship) || !bytes.Equal(fleet, g.slot.Fleet) || !bytes.Equal(adml, g.slot.Admiral) || !bytes.Equal(work, g.slot.Work) || g.battleRand.Seed() != seed {
				t.Fatal("命令摘要或Frame變動原池／RND")
			}
		})
	}
}

func TestFleetCommandHoverTransitions(t *testing.T) {
	m := fleetDisplayManifest(t)
	dir := "workplace/oracle/combat79-r77-hover3-a/"
	g := fleetDisplayFixture(t, m, m.read(t, dir+"input-mem.bin"))
	g.MoveMouse(348, 248)
	if g.fleetCombat.target < 0 {
		t.Fatal("純移動未更新敵方摘要")
	}
	before := bytes.Clone(g.Frame().Pix)
	g.Click(348, 248)
	if g.fleetCombat.stage != fleetCommandMenu {
		t.Fatal("武器彈窗未開")
	}
	g.RightClick(348, 248)
	g.RightRelease()
	if g.fleetCombat.stage != fleetCommandTarget || !bytes.Equal(before, g.Frame().Pix) {
		t.Fatal("取消未恢復摘要與射程")
	}
	g.MoveMouse(480, 400)
	if g.fleetCombat.target != -1 {
		t.Fatal("移入面板未清敵方摘要")
	}
	p := g.res.UnitPanel.Cell
	fb := g.Frame()
	for y := p.Y - ui.TextCellDY + 4*p.Row; y < p.Y-ui.TextCellDY+8*p.Row; y++ {
		for x := p.X; x < p.X+8*p.Col; x++ {
			if fb.At(x, y) != 9 {
				t.Fatal("離開候選後殘留敵方資訊", x, y)
			}
		}
	}
	g.MoveMouse(348, 248)
	if !bytes.Equal(before, g.Frame().Pix) {
		t.Fatal("返回候選未恢復同畫面")
	}
}
