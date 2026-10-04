package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

func fleetChainAssets(t *testing.T, g *Game) {
	t.Helper()
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted", name))
		if os.IsNotExist(err) {
			t.Skip("缺原版素材", name)
		}
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var err error
	data := filepath.Join(root, "docs/internals/data")
	if g.res.BattlePanel == nil {
		g.res.BattlePanel, err = asset.DecodeAMGImage(read("ANI_PNL0.A98"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if g.res.BattleTiles == nil {
		g.res.BattleTiles, err = asset.DecodeTiles(read(g.res.Battle.Field.Tiles), g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
		if err != nil {
			t.Fatal(err)
		}
	}
	if g.res.BattleMenu == nil {
		g.res.BattleMenu, err = screen.LoadBattleMenuSpec(filepath.Join(data, "battle-menu.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if g.res.Entries == nil {
		g.res.Entries, err = rules.LoadBattleEntries(filepath.Join(data, "battle-entry-points.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if g.res.Formation == nil {
		g.res.Formation, err = rules.LoadFormationSlots(filepath.Join(data, "battle-formation-slots.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if g.res.Terrain == nil {
		g.res.Terrain, err = rules.LoadBattleTerrain(filepath.Join(data, "battle-terrain.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	var names struct {
		Kinds []struct {
			Bytes string `json:"bytes"`
		} `json:"ship_kinds"`
	}
	b, err := os.ReadFile(filepath.Join(data, "faction-and-kind-names.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &names); err != nil {
		t.Fatal(err)
	}
	g.res.ShipKindNames = nil
	for _, v := range names.Kinds {
		raw, err := hex.DecodeString(v.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		g.res.ShipKindNames = append(g.res.ShipKindNames, raw)
	}
	g.res.FleetCommand, err = screen.LoadFleetCommandSpec(filepath.Join(data, "fleet-command-ui.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetCommandDisplay, err = screen.LoadFleetCommandDisplaySpec(filepath.Join(data, "fleet-command-display-spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	if g.res.UnitPanel == nil {
		g.res.UnitPanel, err = screen.LoadUnitPanelSpec(filepath.Join(data, "unit-panel.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if g.res.FleetDetail == nil {
		g.res.FleetDetail, err = screen.LoadFleetDetailSpec(filepath.Join(data, "fleet-detail.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if g.res.FleetWindow == nil {
		g.res.FleetWindow, err = screen.LoadFleetWindowSpec(filepath.Join(data, "fleet-window.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	g.res.CommandMenus, err = screen.LoadCommandMenuSpec(filepath.Join(data, "battle-command-menus.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetPresentation, err = screen.LoadFleetPresentationSpec(filepath.Join(data, "fleet-presentation-spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.Engagement, err = screen.LoadEngagementSpec(filepath.Join(data, "engagement-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.EngagementPanel, err = asset.DecodeAMGImage(read("ANI_PNL1.A98"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.Comani, err = asset.OpenBMG(read("COMANI.BMG"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetSceneSprites, err = asset.DecodeFleetSceneSprites(read("3DBLOCK.SPR"), read("COM_FLAG.SPR"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestFleetCommandRosterNative(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/fleet-player-chain-native.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "f569a16f863f9e54d3cda5fe0a8fe2c39566ec74864469d05b374fa29017c7eb" {
		t.Fatal("第76輪名冊來源不符", err)
	}
	var m struct {
		Cases []struct {
			Case   string
			Roster []int
			Order  []int `json:"native_order"`
		}
	}
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	numbers := map[int]int{12: 13, 16: 1, 17: 2, 28: 13, 29: 14, 30: 15, 31: 16}
	for _, c := range m.Cases {
		var fleets []rules.BattleFleet
		for _, i := range c.Roster {
			side := 1
			if i == 12 {
				side = 0
			}
			fleets = append(fleets, rules.BattleFleet{Index: i, Faction: side, Number: numbers[i]})
		}
		if got := fleetCommandRosterOrder(fleets); !reflect.DeepEqual(got, c.Order) {
			t.Fatalf("%s名冊%v原版%v", c.Case, got, c.Order)
		}
	}
}

func fleetSceneReceipt(t *testing.T) autoGameReceipt {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/fleet-scene.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "7d7e5b388eebc8348825e91352ff10e9821ce2b0cf7912d6f9b9cf1cafc95379" {
		t.Fatal("第69輪完整場景來源不符", err)
	}
	var m autoGameReceipt
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func fleetChainRAM(t *testing.T, m autoGameReceipt, ram []byte) *Game {
	t.Helper()
	g, _, _ := commandDataFixture(t)
	g.slot.Ship = bytes.Clone(ram[0x668b2 : 0x668b2+rules.ShipCount*rules.ShipRecord])
	g.slot.Fleet = bytes.Clone(ram[0x60ba4 : 0x60ba4+rules.FleetCount*rules.FleetRecord])
	g.slot.Admiral = bytes.Clone(ram[0x602ac : 0x602ac+len(g.slot.Admiral)])
	g.res.Admiral = g.slot.Admiral
	g.admirals = nil
	g.slot.Work = bytes.Clone(ram[0x5da44 : 0x5da44+rules.WorkRecord])
	g.work = g.slot.Work
	for side := 0; side < 2; side++ {
		g.battleScore[side] = int(binary.LittleEndian.Uint32(ram[0x54b30+0x8fe4+side*4:]))
		g.campaignScore[side] = int(binary.LittleEndian.Uint32(ram[0x54b30+0x8fec+side*4:]))
	}
	g.fleets = g.slot.Fleets(g.res.Weights)
	g.battleFleets = g.slot.BattleFleetsFor([]int{12, 16, 17, 28, 29, 30, 31})
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		pool := g.battleUnitPool(*u)
		r := g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord]
		u.Kind, u.Detail, u.Action, u.Out = int(r[4]), int(r[0x17]), int(r[0x1e]), r[0x1e]&128 != 0
		u.Morale = int(r[6])
		u.Ships = int(binary.LittleEndian.Uint16(r[8:]))
		u.FullShips = int(binary.LittleEndian.Uint16(r[10:]))
		u.RepairCap = int(binary.LittleEndian.Uint16(r[12:]))
		u.Missiles, u.Fighters = int(binary.LittleEndian.Uint32(r[14:])), int(binary.LittleEndian.Uint32(r[18:]))
		u.X, u.Y = int(binary.LittleEndian.Uint16(r[0x34:])), int(binary.LittleEndian.Uint16(r[0x36:]))
	}
	for i := range g.battleFleets {
		f := &g.battleFleets[i]
		f.AdmiralLossFactor = int(g.admiralRec(g.fleetAdmiral(f.Index))[0x0b])
	}
	g.battleRand = rules.NewRand(binary.LittleEndian.Uint32(ram[0x5e00c:]))
	fleetChainAssets(t, g)
	g.display.Anim, g.display.AnimSpeed, g.display.MsgSpeed = 1, 0, 90
	return g
}

func TestFleetSceneGameNative(t *testing.T) {
	m := fleetSceneReceipt(t)
	for _, name := range []string{"laser", "missile", "same-target", "death"} {
		t.Run(name, func(t *testing.T) {
			dir := "workplace/oracle/scene48-r69-v2-" + name + "-a/"
			g := fleetChainRAM(t, m, m.read(t, dir+"start-mem.bin"))
			if name == "death" {
				for i, u := range g.battleUnits {
					if g.battleUnitPool(u) == 102 {
						g.battleUnits[i].Ships = 1
						binary.LittleEndian.PutUint16(g.slot.Ship[102*rules.ShipRecord+8:], 1)
					}
				}
			}
			g.fleetCombat = fleetCombatState{active: true, fleets: []int{12}, current: -1}
			g.beginFleetCombatScenes()
			if g.fleetCombat.err != nil {
				t.Fatal(g.fleetCombat.err)
			}
			if name == "laser" || name == "missile" {
				g.Tick(2 * time.Duration(g.fleetSceneInterval()) * g.TimerTick())
				fleetChainCompareFrame(t, g, m, dir, "frame-0786-01", name+"-before")
			}
			end := m.read(t, dir+"end-mem.bin")
			hits := 0
			for ticks := 0; ticks < 3000 && g.fleetCombat.active; ticks++ {
				s := g.fleetCombat.scene
				if s == nil {
					t.Fatal("缺正式場景")
				}
				if s.damaged && s.tick >= s.resultTick {
					hits++
					if g.Frame() == nil || g.fleetCombat.err != nil {
						t.Fatal("場景渲染失敗", g.fleetCombat.err)
					}
					// 最後一擊資料在第⑤步續接前直接比原版；不用helper期望。
					if name == "laser" || name == "missile" || hits == 2 || name == "death" {
						if name == "laser" || name == "missile" {
							fleetChainCompareFrame(t, g, m, dir, "frame-0810-01", name+"-result")
						}
						for _, pool := range []int{80, 82, 102} {
							p := pool * rules.ShipRecord
							for _, r := range [][2]int{{8, 6}, {14, 8}} {
								o, n := r[0], r[1]
								if !bytes.Equal(g.slot.Ship[p+o:p+o+n], end[0x668b2+p+o:0x668b2+p+o+n]) {
									t.Fatalf("pool%d+%X戰損不同：%x 原版%x；hit%+v seed%08X command%+v", pool, o, g.slot.Ship[p+o:p+o+n], end[0x668b2+p+o:0x668b2+p+o+n], s.hit, g.battleRand.Seed(), s.command)
								}
							}
						}
						if g.battleRand.Seed() != binary.LittleEndian.Uint32(end[0x5e00c:]) {
							t.Fatal("逐擊末端seed不符")
						}
						for _, fleet := range []int{12, 16} {
							r, _ := g.fleetCombatRawFleet(fleet)
							if r[0x3c] != end[0x60ba4+fleet*rules.FleetRecord+0x3c] {
								t.Fatal("完整度不同", fleet, r[0x3c], end[0x60ba4+fleet*rules.FleetRecord+0x3c])
							}
						}
						r, _ := g.fleetCombatRawFleet(12)
						for _, id := range []byte{r[4], r[5]} {
							if id == 255 {
								continue
							}
							p := int(id)*rules.AdmiralRecord + 2
							if !bytes.Equal(g.res.Admiral[p:p+4], end[0x602ac+p:0x602ac+p+4]) {
								t.Fatal("提督／副官功績不同", id)
							}
						}
						for side := 0; side < 2; side++ {
							if g.battleScore[side] != int(binary.LittleEndian.Uint32(end[0x54b30+0x8fe4+side*4:])) || g.campaignScore[side] != int(binary.LittleEndian.Uint32(end[0x54b30+0x8fec+side*4:])) {
								t.Fatal("原版兩分數表不符", side)
							}
						}
					}
					// 同狀態核對限第④步；另驗正式續接與六表保存。
					if name == "laser" || name == "missile" || hits == 2 || name == "death" {
						beforeStock := bytes.Clone(g.slot.Ship[80*rules.ShipRecord+14 : 80*rules.ShipRecord+22])
						fleetChainFinishAndSave(t, g)
						if name == "death" && !bytes.Equal(beforeStock, g.slot.Ship[80*rules.ShipRecord+14:80*rules.ShipRecord+22]) {
							t.Fatal("死亡目標的後續命令仍扣庫存")
						}
						return
					}
					g.Click(0, 0)
				} else {
					g.Tick(g.TimerTick())
				}
				if g.fleetCombat.err != nil {
					t.Fatal(g.fleetCombat.err)
				}
			}
			t.Fatal("逐擊結果未到達")
		})
	}
}

func fleetChainFinishAndSave(t *testing.T, g *Game) {
	t.Helper()
	g.Click(0, 0)
	for n := 0; n < 12000 && !g.battleInputIdle() && g.state != StateCommandPhase; n++ {
		if g.fleetCombat.err != nil {
			t.Fatal(g.fleetCombat.err)
		}
		g.Tick(g.TimerTick())
	}
	if g.fleetCombat.active || g.state != StateCommandPhase && !g.battleInputIdle() {
		t.Fatal("正式第⑤步續接沒有回到Input", g.state, g.battlePhase)
	}
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(dir, name+".DT0"))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	loaded, err := rules.LoadSlot(rules.SlotFiles{Stellar: read("STELLAR"), Planet: read("PLANET"), Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Work: read("WORK"), PlanetHeader: g.slot.Planets.Data[:rules.PlanetHeader]})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []struct {
		name      string
		got, want []byte
	}{
		{"SHIP", loaded.Ship, g.slot.Ship}, {"FLEET", loaded.Fleet, g.slot.Fleet},
		{"ADML", loaded.Admiral, g.slot.Admiral},
		{"STELLAR", loaded.Stellar, g.slot.Stellar}, {"PLANET", loaded.Planets.Data, g.slot.Planets.Data},
	} {
		if !bytes.Equal(p.got, p.want) {
			t.Fatal("完整玩家續接存讀遺失", p.name)
		}
	}
	if loaded.StrategicScore != g.campaignScore {
		t.Fatal("保存遺失戰役分數")
	}
	for side := 0; side < 2; side++ {
		if int(binary.LittleEndian.Uint32(loaded.Work[0xd0+side*4:])) != g.battleScore[side] {
			t.Fatal("保存遺失戰術分數")
		}
	}
	// WORK.Save 編碼 typed 欄位，不與未編碼底稿誤比；六表重存必須穩定。
	again := t.TempDir()
	if err := loaded.Save(again, 0); err != nil {
		t.Fatal(err)
	}
	for _, name := range rules.SlotFileNames(0) {
		a, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(again, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Fatal("六表存讀重存不同", name)
		}
	}
}

func fleetChainCompareFrame(t *testing.T, g *Game, m autoGameReceipt, dir, frame, name string) {
	t.Helper()
	want := m.read(t, dir+frame+"-screen.bin")
	fb := g.Frame()
	if len(want) != len(fb.Pix) {
		t.Fatal("原版／Game畫布不符")
	}
	diff := gfx.NewFramebuffer(0)
	different := 0
	maxChannel := 0
	pal := g.Palette()
	channelDiff := func(a, b byte) int {
		m := 0
		for k := 0; k < 3; k++ {
			x, y := int(pal[a][k]), int(pal[b][k])
			x = (x << 2) | (x >> 4)
			y = (y << 2) | (y >> 4)
			d := x - y
			if d < 0 {
				d = -d
			}
			m = max(m, d)
		}
		return m
	}
	bbox := [4]int{640, 480, -1, -1}
	for i, p := range fb.Pix {
		if p != want[i] {
			different++
			maxChannel = max(maxChannel, channelDiff(p, want[i]))
			diff.Pix[i] = 15
			x, y := i%640, i/640
			bbox = [4]int{min(bbox[0], x), min(bbox[1], y), max(bbox[2], x), max(bbox[3], y)}
		}
	}
	if out := os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT"); out != "" {
		original := gfx.NewFramebuffer(0)
		copy(original.Pix, want)
		for label, image := range map[string]*gfx.Framebuffer{"original": original, "remake": fb, "diff": diff} {
			if err := image.WritePNG(filepath.Join(out, name+"-"+label+".png"), g.Palette()); err != nil {
				t.Fatal(err)
			}
		}
		var regions []map[string]any
		for _, r := range [][4]int{{8, 56, 256, 102}, {376, 56, 256, 102}, {16, 160, 240, 40}, {384, 160, 240, 40}, {272, 160, 96, 40}, {298, 63, 11, 66}, {333, 63, 11, 66}, {8, 208, 624, 224}} {
			n := 0
			maximum := 0
			box := [4]int{640, 480, -1, -1}
			for y := r[1]; y < r[1]+r[3]; y++ {
				for x := r[0]; x < r[0]+r[2]; x++ {
					i := y*640 + x
					if fb.Pix[i] != want[i] {
						n++
						maximum = max(maximum, channelDiff(fb.Pix[i], want[i]))
						box = [4]int{min(box[0], x), min(box[1], y), max(box[2], x), max(box[3], y)}
					}
				}
			}
			regions = append(regions, map[string]any{"rect": r, "pixels": r[2] * r[3], "same": r[2]*r[3] - n, "different": n, "maximum_channel_difference": maximum, "bbox": box})
		}
		b, err := json.MarshalIndent(map[string]any{"case": name, "classification": "same-state Game scene adapter; restored command diagnostic", "pixels": len(want), "same": len(want) - different, "different": different, "maximum_channel_difference": maxChannel, "bbox": bbox, "regions": regions}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, name+"-comparison.json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if different != 0 {
		t.Fatalf("Game完整畫面%d像素差，範圍%v", different, bbox)
	}
}

func TestFleetPlayerCommandUI(t *testing.T) {
	for _, weapon := range []int{0, 1, 2, 3, 4} {
		t.Run(fmt.Sprint(weapon), func(t *testing.T) {
			g, own, target := commandDataFixture(t)
			fleetChainAssets(t, g)
			for i := range g.battleFleets {
				g.battleFleets[i].Action = 4
				if g.battleFleets[i].Index == 12 {
					g.battleFleets[i].Action = 1
				}
			}
			g.beginFleetCombat()
			if g.fleetCombat.err != nil || g.fleetCombat.stage != fleetCommandHeader {
				t.Fatal("命令標頭未到達", g.fleetCombat.err)
			}
			if err := g.SaveSlot(t.TempDir(), 0); err == nil {
				t.Fatal("命令中允許保存")
			}
			g.Click(240, 240)
			if g.fleetCombat.current != own || g.fleetCombat.stage != fleetCommandTarget {
				t.Fatal("標頭未接目標等待")
			}
			before := g.battleRand.Seed()
			g.RightClick(0, 0)
			if g.fleetCombat.stage != fleetCommandTarget || g.battleRand.Seed() != before {
				t.Fatal("右鍵退出命令或消費RND")
			}
			u := g.battleUnits[target]
			f := g.res.Battle.Field
			x, y := f.X+u.X-g.battleView.Col*f.TileW+1, f.Y+u.Y-g.battleView.Row*f.TileH+1
			if weapon >= 3 {
				x, y = 16, 56
			}
			g.Click(x, y)
			if g.fleetCombat.stage != fleetCommandMenu {
				t.Fatal("選單未開啟")
			}
			g.Frame()
			if weapon < 3 {
				g.RightClick(x, y)
				if g.fleetCombat.stage != fleetCommandTarget {
					t.Fatal("武器取消未回等待")
				}
				g.Click(x, y)
			}
			c := g.fleetCombat
			row := weapon
			if weapon == 3 {
				row = 1
			}
			if weapon == 4 {
				row = 0
			}
			g.Click(c.menuX+16, c.menuY+20+row*20)
			if g.fleetCombat.err != nil {
				t.Fatal(g.fleetCombat.err)
			}
			if weapon == 4 {
				if g.fleetCombat.stage != fleetCommandAutoWait {
					t.Fatal("自動沒有6tick等待")
				}
				g.Tick(6 * g.TimerTick())
			}
			if weapon == 3 {
				if g.fleetCombat.active {
					t.Fatal("防禦沒有續接第⑤步")
				}
				return
			}
			if g.state != StateEngagement || g.fleetCombat.scene == nil {
				t.Fatal("正式UI未接場景", g.state)
			}
			pool, w, err := g.fleetCombatCommand(own)
			if err != nil || pool != 102 || weapon < 3 && int(w) != weapon {
				t.Fatal("UI原池提交錯誤", pool, w, err)
			}
			if g.battleUnits[target].Ships != u.Ships {
				t.Fatal("動畫未播完即扣戰損")
			}
			scene := g.fleetCombat.scene
			g.AdvanceBattleTurn()
			if g.fleetCombat.scene != scene {
				t.Fatal("重複實行覆蓋命令")
			}
		})
	}
}

// 不注入局面：S3正式讀檔、星域掃描、玩家戰鬥命令、實行、武器自動選擇及保存。
// 不與未知的原版正常全流程作same-state聲明。
func fleetChainRoute(g *Game, x, y int) map[[2]int]int {
	fld := g.res.Battle.Field
	cells := g.battleCells(g.battleRegion)
	start := [2]int{x / fld.TileW, y / fld.TileH}
	cost := map[[2]int]int{start: 0}
	q := [][2]int{start}
	for at := 0; at < len(q); at++ {
		p := q[at]
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				v := [2]int{p[0] + dx, p[1] + dy}
				if v[0] < 0 || v[1] < 0 || v[0] >= asset.BattleMapCols || v[1] >= asset.BattleMapRows {
					continue
				}
				if _, ok := cost[v]; ok {
					continue
				}
				if !g.res.Terrain.Passable(cells, v[0]*fld.TileW, v[1]*fld.TileH) || !g.res.Terrain.LineOpen(cells, p[0]*fld.TileW, p[1]*fld.TileH, v[0]*fld.TileW, v[1]*fld.TileH) {
					continue
				}
				cost[v] = cost[p] + 1
				q = append(q, v)
			}
		}
	}
	return cost
}

func TestNormalFleetPlayerChain(t *testing.T) {
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
	for _, faction := range []int{0, 1} {
		t.Run(fmt.Sprint(faction), func(t *testing.T) {
			g := reachConfiguredAI(t, -1, faction)
			fleetChainAssets(t, g)
			withSound := os.Getenv("LOGH3_FLEET_SOUND") == "1"
			var soundEvents []map[string]any
			if withSound {
				fleetSoundAssets(t, g)
			}
			var inputs []map[string]any
			elapsed := int64(0)
			collectSound := func() {
				for _, id := range g.DrainSoundEffects() {
					soundEvents = append(soundEvents, map[string]any{"id": id, "at_ns": elapsed})
				}
			}
			click := func(x, y int) {
				inputs = append(inputs, map[string]any{"kind": "click", "x": x, "y": y, "at_ns": elapsed})
				g.MoveMouse(x, y)
				g.Click(x, y)
				g.Frame()
				collectSound()
				if out := os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT"); out != "" && g.fleetCombat.active && (g.fleetCombat.stage == fleetCommandTarget || g.fleetCombat.stage == fleetCommandMenu) {
					folder := filepath.Join(out, fmt.Sprintf("normal-%d", faction))
					if err := os.MkdirAll(folder, 0755); err != nil {
						t.Fatal(err)
					}
					name := fmt.Sprintf("command-%d-%05d", g.fleetCombat.stage, len(inputs))
					if err := g.Frame().WritePNG(filepath.Join(folder, name+".png"), g.Palette()); err != nil {
						t.Fatal(err)
					}
					if err := g.Frame().WriteIndex(filepath.Join(folder, name+".bin")); err != nil {
						t.Fatal(err)
					}
				}
			}
			right := func(x, y int) {
				inputs = append(inputs, map[string]any{"kind": "right", "x": x, "y": y, "at_ns": elapsed})
				g.RightClick(x, y)
				g.RightRelease()
			}
			tick := func() {
				d := g.TimerTick()
				inputs = append(inputs, map[string]any{"kind": "tick", "ns": int64(d), "at_ns": elapsed})
				g.Tick(d)
				elapsed += int64(d)
				collectSound()
			}
			seed := func() {
				inputs = append(inputs, map[string]any{"kind": "seed", "seed": uint32(0x12345678), "at_ns": elapsed})
				g.battleRand = rules.NewRand(0x12345678)
			}
			ack := func() {
				if !g.msgOn {
					return
				}
				for y := 0; y < 480; y++ {
					for x := 0; x < 640; x++ {
						if _, ok := g.res.MessageBox.MessageBoxButtonAt(g.msgKind, x, y); ok {
							click(x, y)
							return
						}
					}
				}
				t.Fatal("訊息框沒有可按的確認按鈕")
			}
			click(544, 125)
			click(560, 145)
			click(300, 322)
			click(318, 208+20*faction)
			if g.state != StateCommandPhase {
				t.Fatal("正常S3載入失敗", g.state)
			}
			click(544, 125)
			click(560, 125)
			for n := 0; n < 15000 && g.state != StateBattle; n++ {
				switch g.state {
				case StateTurnThinking, StateTurnWindow, StateTurnReport, StateEngagement:
					tick()
				case StateOccupationReport, StateRebellionReport, StateRegionFlipReport, StateAdmiralFateReport, StateShipDelivery, StatePhezzanReport:
					click(240, 240)
				default:
					t.Fatal("正常首戰停在", g.state)
				}
			}
			if g.state != StateBattle {
				t.Fatal("正常首戰未到達")
			}
			seed() // 明示固定remake測試seed，不作原版整段骰序聲明。
			cpu := g.snapshotSide(1 - faction)
			scenes, orders := 0, 0
			var seen *fleetCombatSceneState
			for turn := 0; turn < rules.MaxBattleTurns*4 && scenes == 0; turn++ {
				if g.state != StateBattle {
					if g.state != StateCommandPhase {
						t.Fatal("戰後未回命令畫面", g.state)
					}
					click(544, 125)
					click(560, 125)
					for n := 0; n < 15000 && g.state != StateBattle; n++ {
						switch g.state {
						case StateTurnThinking, StateTurnWindow, StateTurnReport, StateEngagement:
							tick()
						case StateOccupationReport, StateRebellionReport, StateRegionFlipReport, StateAdmiralFateReport, StateShipDelivery, StatePhezzanReport:
							click(240, 240)
						default:
							t.Fatal("後續正常戰場停在", g.state)
						}
					}
					if g.state != StateBattle {
						t.Fatal("後續正常戰場未到達")
					}
					seed()
					cpu = g.snapshotSide(1 - faction)
				}
				// 依正式右鍵捲動跟上第一支己方旗艦，不直接改鏡頭或單位座標。
				lead := -1
				for i, u := range g.battleUnits {
					if u.Faction == faction && u.Squadron == 0 && !u.Out && (lead < 0 || u.CruiseSpeed > g.battleUnits[lead].CruiseSpeed) {
						lead = i
					}
				}
				for i, u := range g.battleUnits {
					if i != lead {
						continue
					}
					if u.Faction != faction || u.Squadron != 0 || u.Out {
						continue
					}
					want := g.battleViewAt(u.X, u.Y)
					fld := g.res.Battle.Field
					// 把旗艦留在鏡頭靠後一側，能點到完整巡航行程的目的地。
					for _, v := range g.battleUnits {
						if v.Faction == faction || v.Ships == 0 || v.Squadron != 0 {
							continue
						}
						if v.X > u.X {
							want.Col = clampCol(u.X/fld.TileW-1, fld.Cols)
						} else {
							want.Col = clampCol(u.X/fld.TileW-fld.Cols+2, fld.Cols)
						}
						if v.Y > u.Y {
							want.Row = clampRow(u.Y/fld.TileH-1, fld.Rows)
						} else {
							want.Row = clampRow(u.Y/fld.TileH-fld.Rows+2, fld.Rows)
						}
						break
					}
					for n := 0; n < 256 && g.battleView != want; n++ {
						best := abs(g.battleView.Col-want.Col) + abs(g.battleView.Row-want.Row)
						px, py := -1, -1
						for y := fld.Y + 1; y < fld.Y+fld.Rows*fld.TileH; y += 20 {
							for x := fld.X + 1; x < fld.X+fld.Cols*fld.TileW; x += 16 {
								col, row, moved := rules.ViewScrollStep(g.battleView.Col, g.battleView.Row, rules.ViewScrollDir(x, y-40))
								d := abs(col-want.Col) + abs(row-want.Row)
								if moved && d < best {
									px, py, best = x, y, d
								}
							}
						}
						if px < 0 {
							t.Fatal("正常捲動無法跟上旗艦")
						}
						right(px, py)
					}
					break
				}
				// 只下目前鏡頭可點的旗艦。目的地取朝最近敵艦的可見可通行點。
				for i, u := range g.battleUnits {
					if u.Faction != faction || u.Squadron != 0 || u.Out || g.battleFleet(u.Fleet).Flag0 {
						continue
					}
					x, y := unitScreenPos(g, i)
					fld := g.res.Battle.Field
					if x < fld.X || y < fld.Y || x+16 > fld.X+fld.Cols*fld.TileW || y+16 > fld.Y+fld.Rows*fld.TileH {
						continue
					}
					if g.battleUnitAt(x+8, y+8) != i {
						continue
					}
					foe, best := -1, int64(1<<62)
					for j, v := range g.battleUnits {
						if v.Faction != faction && v.Ships > 0 && v.Squadron >= 0 {
							d := rules.Distance16(u.X, u.Y, v.X, v.Y)
							if d < best {
								best, foe = d, j
							}
						}
					}
					if foe < 0 {
						t.Fatal("正常戰場沒有敵方")
					}
					near := best
					// 只規劃測試玩家點擊；沿既有通行表避開星雲，不改產品移動。
					cost := fleetChainRoute(g, g.battleUnits[foe].X, g.battleUnits[foe].Y)
					ox, oy := g.battleView.Col*fld.TileW, g.battleView.Row*fld.TileH
					dx, dy := -1, -1
					bestCost := int(1 << 30)
					best = int64(1 << 62)
					for py := fld.Y + 8; py < fld.Y+fld.Rows*fld.TileH; py += 16 {
						for px := fld.X + 8; px < fld.X+fld.Cols*fld.TileW; px += 16 {
							if wx, wy, ok := g.battleWorldAt(px, py); ok {
								if !g.res.Terrain.LineOpen(g.battleCells(g.battleRegion), u.X, u.Y, wx, wy) {
									continue
								}
								step, ok := cost[[2]int{wx / fld.TileW, wy / fld.TileH}]
								if !ok {
									continue
								}
								v := g.battleUnits[foe]
								d := rules.Distance16(wx, wy, v.X, v.Y)
								if step < bestCost || (step == bestCost && d < best) {
									dx, dy, best, bestCost = px, py, d, step
								}
							}
						}
					}
					if dx < 0 {
						if i != lead {
							continue
						}
						t.Fatal("正常鏡頭沒有目的地", ox, oy, u.X, u.Y, g.res.Terrain.Passable(g.battleCells(g.battleRegion), u.X, u.Y))
					}
					click(x+8, y+8)
					if !g.battleCmd || g.battleCmdKind != screen.CommandMenuFleet {
						t.Fatal("旗艦選單未開")
					}
					l := g.res.CommandMenus.Layout
					row := 0 // 先巡航接近，再用戰鬥速度進射程。
					if near <= int64(300*rules.FixOne) || g.battleTurn == rules.MaxBattleTurns {
						row = 1
					}
					click(g.battleCmdX+l.HighlightDX+8, g.battleCmdY+l.HighlightDY+row*l.RowPitch+8)
					click(dx, dy)
					if g.msgOn {
						ack()
						continue
					}
					click(dx, dy)
					if g.fleetOrder != FleetOrderNone || g.battleFleet(u.Fleet).Action&15 != row {
						t.Fatal("正常UI沒有寫戰鬥命令")
					}
					orders++
				}
				ack()
				click(240, 470) // 戰場外空白，不可能命中分隊或行星。
				if !g.battleMenu {
					t.Fatal("實行選單未開", g.state, g.battleCmd, g.msgOn, g.fleetOrder, g.battlePhase)
				}
				click(g.battleMenuX+20, g.battleMenuY+20)
				for n := 0; n < 20000 && !g.battleInputIdle() && g.state != StateCommandPhase; n++ {
					if g.fleetCombat.err != nil {
						t.Fatal(g.fleetCombat.err)
					}
					if g.fleetCombat.active {
						c := g.fleetCombat
						switch c.stage {
						case fleetCommandHeader:
							click(240, 470)
						case fleetCommandTarget:
							click(240, 470)
						case fleetCommandMenu:
							click(c.menuX+16, c.menuY+20)
						case fleetCombatScene:
							if c.scene != seen {
								scenes++
								seen = c.scene
								g.Frame()
							}
						}
					}
					tick()
				}
				if g.state == StateBattle && !g.battleInputIdle() {
					t.Fatal("正常交戰未續接Input", g.state, g.fleetCombat.stage)
				}
				for _, f := range g.battleFleets {
					if f.Faction == faction {
						t.Logf("turn%d gameTurn%d fleet%d xy%d,%d dest%d,%d action%X speed%d", turn, g.battleTurn, f.Index, f.X, f.Y, f.DestX, f.DestY, f.Action, f.PerTurn)
					}
				}
			}
			if scenes == 0 || orders == 0 {
				t.Fatal("正常UI未進入交戰", orders, scenes)
			}
			if withSound && len(soundEvents) == 0 {
				t.Fatal("正常交戰沒有音效")
			}
			for i, u := range g.battleUnits {
				if before, ok := cpu[i]; ok && (u.X != before.X || u.Y != before.Y) {
					t.Fatal("被動CPU被移動")
				}
			}
			saveDir := t.TempDir()
			if out := os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT"); out != "" {
				saveDir = filepath.Join(out, fmt.Sprintf("normal-%d", faction))
				if err := os.MkdirAll(saveDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := g.Frame().WritePNG(filepath.Join(saveDir, "screen.png"), g.Palette()); err != nil {
					t.Fatal(err)
				}
				if err := g.Frame().WriteIndex(filepath.Join(saveDir, "screen.bin")); err != nil {
					t.Fatal(err)
				}
				b, err := json.MarshalIndent(map[string]any{"classification": "normal Game player path; independent bootstrap replay required", "faction": faction, "orders": orders, "scenes": scenes, "seed": uint32(0x12345678), "inputs": inputs, "cpu_position_unchanged": true}, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(saveDir, "inputs.json"), append(b, '\n'), 0644); err != nil {
					t.Fatal(err)
				}
				if withSound {
					b, err := json.MarshalIndent(soundEvents, "", "  ")
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(saveDir, "sounds.json"), append(b, '\n'), 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := g.SaveSlot(saveDir, 0); err != nil {
				t.Fatal("正常交戰保存", err)
			}
			t.Logf("陣營%d：正常戰鬥命令%d次、交戰%d擊、保存通過", faction, orders, scenes)
		})
	}
}
