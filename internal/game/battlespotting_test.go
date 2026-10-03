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
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

type spottingRecord struct {
	Index int
	Hex   string
}
type spottingFrame struct {
	Name                   string
	Hashes                 map[string]string
	Units, Fleets, Planets []spottingRecord
	CameraGrid             []int `json:"camera_grid"`
	CameraPixels           []int `json:"camera_pixels"`
	Seed                   string
}

func (f spottingFrame) view() screen.BattleView {
	if len(f.CameraGrid) == 2 {
		return screen.BattleView{Col: f.CameraGrid[0], Row: f.CameraGrid[1]}
	}
	return screen.BattleView{Col: f.CameraPixels[0] / 16, Row: f.CameraPixels[1] / 20}
}

type spottingCase struct {
	Scenario string
	Frames   []spottingFrame
	Runs     []struct{ Path string }
}

func spottingFixture(t *testing.T, c spottingCase) (*Game, string) {
	t.Helper()
	base := filepath.Join(root, c.Runs[0].Path)
	mem, err := os.ReadFile(filepath.Join(base, "entry-mem.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺原版索敵收據；docs/playtest/046-battle-spotting-phase.md")
	}
	if err != nil {
		t.Fatal(err)
	}
	wantHash := c.Frames[0].Hashes["-mem.bin"]
	if wantHash == "" {
		wantHash = c.Frames[0].Hashes["entry-mem.bin"]
	}
	if fmtHash(sha256.Sum256(mem)) != wantHash {
		t.Fatal("原版入口身份不同")
	}
	g := newTestGame(t)
	g.loadSlotFromDisk(8)
	loadPreparationPresentation(t, g)
	loadMovementTiles(t, g)
	g.res.PlanetInfo, err = screen.LoadPlanetInfoSpec(filepath.Join(root, "docs/internals/data/planet-info.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.slot.Fleet = bytes.Clone(mem[0x60BA4 : 0x60BA4+32*122])
	g.slot.Ship = bytes.Clone(mem[0x668B2 : 0x668B2+352*78])
	g.slot.Work = bytes.Clone(mem[0x5DA44 : 0x5DA44+rules.WorkRecord])
	copy(g.slot.Planets.Data[rules.PlanetHeader:], mem[0x53132:0x53132+81*82])
	g.LoadSlot(g.slot)
	g.battleRegion = 26
	if c.Scenario == "s1" {
		g.battleRegion = 8
	}
	g.state = StateBattle
	g.battleUnits, g.battleFleets = nil, nil
	for _, r := range c.Frames[0].Fleets {
		raw, _ := hex.DecodeString(r.Hex)
		bf := rules.NewBattleFleet(r.Index, raw, [rules.FleetAbilityCount]byte{})
		g.battleFleets = append(g.battleFleets, bf)
		f := g.fleetByIndex(r.Index)
		if f == nil {
			t.Fatal("缺原版艦隊")
		}
		g.battleUnits = append(g.battleUnits, g.slot.Deploy(g.res.Weights, g.res.Formation, *f, bf.X, bf.Y, bf.Facing)...)
		if bf.ScoutMode != 0xFF {
			if u, ok := g.slot.ScoutUnit(g.res.Weights, *f, bf.X, bf.Y); ok {
				g.battleUnits = append(g.battleUnits, u)
			}
		}
	}
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		r := g.slot.ShipAt(g.battleUnitPool(*u))
		u.X, u.Y = int(binary.LittleEndian.Uint16(r[52:])), int(binary.LittleEndian.Uint16(r[54:]))
		u.TargetX, u.TargetY = int(binary.LittleEndian.Uint16(r[40:])), int(binary.LittleEndian.Uint16(r[42:]))
	}
	g.battlePlanets = g.battleRegionPlanets(g.battleRegion)
	for i := range g.battlePlanets {
		p := &g.battlePlanets[i]
		r := mem[0x53132+p.Index*82:]
		p.Owner, p.X, p.Y = int(r[17]), int(binary.LittleEndian.Uint16(r[70:])), int(binary.LittleEndian.Uint16(r[72:]))
		p.Defence = int(binary.LittleEndian.Uint16(r[20:]))
	}
	g.battleView = c.Frames[0].view()
	g.battleTurn = int(binary.LittleEndian.Uint16(mem[0x5DA4A:])) + 1
	g.battleRand = rules.NewRand(binary.LittleEndian.Uint32(mem[0x5E00C:]))
	g.battlePhase = screen.BattlePhaseMovement
	g.MoveMouse(240, 240)
	if c.Scenario == "s1" {
		g.MoveMouse(100, 150)
	}
	// 唯一快照還含顯示物件舊相對座標。從該狀態重建最後顯示鏡頭，
	// 由正式renderer畫入口背景；沒有複製原版framebuffer當期望或產品素材。
	u := g.battleUnits[0]
	raw := g.slot.ShipAt(g.battleUnitPool(u))
	obj := int(binary.LittleEndian.Uint16(raw[0x38:])) + 16*int(binary.LittleEndian.Uint16(raw[0x3A:]))
	if obj > 0 && obj+40 <= len(mem) {
		actual := g.battleView
		g.battleView = screen.BattleView{Col: (u.X - int(int16(binary.LittleEndian.Uint16(mem[obj+6:])))) / 16,
			Row: (u.Y - int(int16(binary.LittleEndian.Uint16(mem[obj+8:])))) / 20}
		g.Frame()
		g.battleView = actual
	}
	// 唯一入口注入完成。原版其餘邊界只作斷言，不再寫回Game。
	return g, base
}

// 正式Game第三階段逐觀測者、等待與renderer；只在唯一入口讀原版RAM。
func TestBattleSpottingWholePhaseMatchesOriginal(t *testing.T) {
	testBattleSpottingCases(t, []string{"battle-spotting-phase.json", "battle-discovery-evidence.json"}, 9, "round42-spot-")
}

func TestBattleSpottingSourceAndScrollMatchesOriginal(t *testing.T) {
	output := os.Getenv("LOGH3_SPOTTING_OUTPUT")
	if output == "" {
		output = "round43-spot-"
	}
	testBattleSpottingCases(t, []string{"battle-discovery-scroll.json"}, 5, output)
}

func testBattleSpottingCases(t *testing.T, names []string, count int, output string) {
	t.Helper()
	var cases []spottingCase
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join(root, "docs/internals/data", name))
		if err != nil {
			t.Fatal(err)
		}
		var data struct{ Cases []spottingCase }
		if err = json.Unmarshal(b, &data); err != nil {
			t.Fatal(err)
		}
		cases = append(cases, data.Cases...)
	}
	if len(cases) != count {
		t.Fatal("原版案例數不同")
	}
	for _, c := range cases {
		t.Run(c.Scenario, func(t *testing.T) {
			g, base := spottingFixture(t, c)
			completed := 0
			g.beginBattleSpotting(func() { completed++ })
			// 後續所有可比邊界以正式狀態演算，輸出存於既有workplace。
			check := func(f spottingFrame) {
				t.Helper()
				for _, r := range f.Units {
					want, _ := hex.DecodeString(r.Hex)
					if !bytes.Equal(g.slot.ShipAt(r.Index), want) {
						t.Errorf("%s池%d全78 bytes不同：%x／%x", f.Name, r.Index, g.slot.ShipAt(r.Index), want)
					}
				}
				for _, r := range f.Fleets {
					want, _ := hex.DecodeString(r.Hex)
					if !bytes.Equal(g.slot.Fleet[r.Index*122:(r.Index+1)*122], want) {
						t.Errorf("%s艦隊%d全122 bytes不同", f.Name, r.Index)
					}
				}
				if g.battleView != f.view() {
					t.Errorf("%s鏡頭%v／%v", f.Name, g.battleView, f.view())
				}
				want, err := os.ReadFile(filepath.Join(base, f.Name+".bin"))
				if err != nil {
					t.Fatal(err)
				}
				wantHash := f.Hashes[".bin"]
				if wantHash == "" {
					wantHash = f.Hashes[f.Name+".bin"]
				}
				if fmtHash(sha256.Sum256(want)) != wantHash {
					t.Fatal("原版畫面身份不同")
				}
				got := g.Frame().Pix
				diff := 0
				for i := range want {
					if want[i] != got[i] {
						diff++
					}
				}
				prefix := filepath.Join(root, "workplace", output+c.Scenario+"-"+f.Name)
				if err = os.WriteFile(prefix+".bin", got, 0644); err != nil {
					t.Fatal(err)
				}
				var palette []byte
				for _, rgb := range g.Palette() {
					palette = append(palette, rgb[:]...)
				}
				if err = os.WriteFile(prefix+".pal", palette, 0644); err != nil {
					t.Fatal(err)
				}
				if diff != 0 {
					t.Errorf("%s全畫面差%d／307200", f.Name, diff)
				}
				t.Logf("%s全畫面差%d，階段%d／觀測者%d／當筆%d", f.Name, diff, g.battleSpotting.stage, g.battleSpotting.next, g.battleSpotting.item)
			}
			for _, f := range c.Frames {
				s := &g.battleSpotting
				switch {
				case f.Name == "entry":
					continue // 原版在寫索敵文字之前，另以入口身份及RAM核對。
				case f.Name == "label":
				case f.Name == "discovered-0" || f.Name == "discovered":
					for steps := 0; s.stage != 1 && s.active && steps < 200; steps++ {
						g.continueBattleSpotting()
					}
				case f.Name == "header":
					for steps := 0; !s.header && s.active && steps < 200; steps++ {
						g.continueBattleSpotting()
					}
				case f.Name == "camera-start":
					continue // 呼叫內部邊界，正式完整置鏡另核highlight入口。
				case len(f.Name) > 7 && f.Name[:7] == "scroll-":
					// 每次正式更新必須保留下一個原版重畫。不能跳到期望鏡頭，
					// 也不能迴圈尋找碰巧相同的終態。
					g.continueBattleSpotting()
				case f.Name == "highlight-start":
					for steps := 0; s.stage != 4 && s.active && steps < 200; steps++ {
						g.continueBattleSpotting()
					}
				case len(f.Name) > 5 && f.Name[:5] == "unit-":
					var index int
					fmt.Sscanf(f.Name, "unit-%d", &index)
					for steps := 0; (s.stage != 5 || s.item != index) && s.active && steps < 200; steps++ {
						g.continueBattleSpotting()
					}
				case f.Name == "seat1":
					continue // CPU無UI；終態全紀錄另核。
				case f.Name == "end":
					for steps := 0; s.active && steps < 200; steps++ {
						g.continueBattleSpotting()
					}
				}
				check(f)
			}
			if completed != 1 || g.battleSpotting.active {
				t.Fatal("第三階段沒有且僅續接一次")
			}
			if g.battleRand.Seed() != binary.LittleEndian.Uint32(mustSpottingMem(t, base)[0x5E00C:]) {
				t.Fatal("索敵消費了原版未消費的亂數")
			}
			// 實際Tick等待與提前左右鍵由正常玩家測試另驗，不能以以上邊界直走代替。
		})
	}
}

func mustSpottingMem(t *testing.T, base string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(base, "entry-mem.bin"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBattleSpottingScrollTicks(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-discovery-scroll.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data struct{ Cases []spottingCase }
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	for _, c := range data.Cases {
		if c.Scenario != "near-source" && c.Scenario != "scout13" && c.Scenario != "offscreen" {
			continue
		}
		t.Run(c.Scenario, func(t *testing.T) {
			g, base := spottingFixture(t, c)
			completed := 0
			g.beginBattleSpotting(func() { completed++ })
			for i := 0; !g.battleSpotting.header && i < 100; i++ {
				g.Tick(time.Second / 60)
			}
			if !g.battleSpotting.header {
				t.Fatal("未到達原版來源標頭")
			}
			for _, f := range c.Frames {
				if len(f.Name) <= 7 || f.Name[:7] != "scroll-" {
					continue
				}
				if f.Name == "scroll-0" {
					g.Tick(time.Minute) // 只結束目前等待並呈現第一步，不吞掉其他鏡頭幀。
				} else {
					view, turn := g.battleView, g.battleTurn
					g.Click(100, 100)
					g.RightClick(100, 100)
					g.AdvanceBattleTurn()
					if g.battleView != view || g.battleTurn != turn {
						t.Fatal("移鏡期間輸入越過續接閘門")
					}
					g.Tick(time.Second / 60)
				}
				want, err := os.ReadFile(filepath.Join(base, f.Name+".bin"))
				if err != nil {
					t.Fatal(err)
				}
				// 固定收據的游標位置；指令阻擋與滑鼠移動為不同契約。
				g.MoveMouse(240, 240)
				if g.battleView != f.view() || !bytes.Equal(g.Frame().Pix, want) || completed != 0 {
					t.Fatalf("%s正式Tick漏過原版重畫：%v／%v", f.Name, g.battleView, f.view())
				}
			}
			for i := 0; g.battleSpotting.active && i < 10000; i++ {
				g.Tick(time.Second / 60)
			}
			if completed != 1 || g.battleSpotting.active {
				t.Fatal("移鏡結束沒有僅續接一次")
			}
		})
	}
}

func TestBattleSpottingNormalWaitAndSave(t *testing.T) {
	g := battleTestGame(t)
	if g == nil {
		t.Skip("缺正常戰場入口")
	}
	loadPreparationPresentation(t, g)
	loadMovementTiles(t, g)
	g.battleRand = rules.NewRand(0x12345678)
	// 正常實行，不注入情報或報告；確認索敵先於戰鬥、一次續接以及正式存讀。
	g.Click(100, 100)
	g.Click(100, 110)
	saw, header := false, false
	for i := 0; i < 10000 && (g.battlePreparing || g.battleMoving || g.battleSpotting.active || g.state == StateTurnReport); i++ {
		if g.battleSpotting.active {
			saw = true
			turn := g.battleTurn
			g.AdvanceBattleTurn()
			if g.battleTurn != turn {
				t.Fatal("索敵期間重複實行")
			}
		}
		if g.battleSpotting.header {
			if !header {
				t.Logf("正常首份報告來源：陣營%d、艦隊%d、行星%d", g.battleSpotting.source.faction, g.battleSpotting.source.fleet, g.battleSpotting.source.planet)
				movementFrame(t, g, "spotting-header")
			}
			header = true
		}
		g.Tick(time.Second / 60)
	}
	if !saw || g.battleTurn != 2 {
		t.Fatalf("正常實行未完成第三階段：%v／%d", saw, g.battleTurn)
	}
	t.Logf("正常實行：索敵=%v、發現標頭=%v、完成回合=%d", saw, header, g.battleTurn)
	// 第50輪原版正常S3：到達配置後，首回合兩座位索敵均完成，
	// 2526:033B／04B1、3826:0298、3765:006C及25DB:0E78零次。
	// 舊近似佈陣造成的發現報告不再是正常入口判準；等待另由固定入口驗。
	if header {
		t.Fatal("正常S3首回合出現原版沒有的發現報告；研究388")
	}
	movementFrame(t, g, "spotting-completed")
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	ship, fleets := bytes.Clone(g.slot.Ship), bytes.Clone(g.slot.Fleet)
	loaded := newTestGame(t)
	loaded.res.SaveDir = dir
	loaded.loadSlotFromDisk(0)
	if loaded.SaveErr() != nil || !bytes.Equal(ship, loaded.slot.Ship) || !bytes.Equal(fleets, loaded.slot.Fleet) {
		t.Fatal("索敵後完整SHIP／FLEET存讀不同")
	}
	// 下一次實行先清Seen。移動寫回同一byte不能把上一輪的Seen設回。
	g.AdvanceBattleTurn()
	for _, u := range g.battleUnits {
		raw := g.slot.ShipAt(g.battleUnitPool(u))
		if len(raw) != rules.ShipRecord || (raw[0x16]&1 != 0) != u.Seen || (u.Attack&1 != 0) != u.Seen {
			t.Fatal("下輪衰減後Seen／移動旗標／原池不同步")
		}
	}
	settleBattleExecution(t, g)
	if g.battleTurn != 3 {
		t.Fatal("第二次正常實行沒有完成")
	}
}

// 固定入口只用於等待診斷；與上面的正常實行證據分開。
func TestBattleSpottingTimedWaitAndClicks(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-spotting-phase.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data struct{ Cases []spottingCase }
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	var c spottingCase
	for _, candidate := range data.Cases {
		if candidate.Scenario == "s3" {
			c = candidate
		}
	}
	if c.Scenario == "" {
		t.Fatal("缺固定S3入口")
	}
	for _, mode := range []string{"ticks", "left", "right"} {
		t.Run(mode, func(t *testing.T) {
			g, _ := spottingFixture(t, c)
			g.display.MsgSpeed = 90
			completed := 0
			g.beginBattleSpotting(func() { completed++ })
			for i := 0; i < 100 && g.battleSpotting.stage != 1; i++ {
				g.Tick(g.TimerTick())
			}
			g.Click(600, 240)
			g.RightClick(600, 240)
			if g.battleSpotting.stage != 1 {
				t.Fatal("非等待邊界被點擊跳過")
			}
			g.Tick(g.TimerTick())
			if !g.battleSpotting.header || g.battleSpotting.ticks != 90 {
				t.Fatal("MsgSpeed標頭等待未開始")
			}
			view := g.battleView
			g.RightHold(600, 240)
			if view != g.battleView || g.battleMenu || completed != 0 {
				t.Fatal("標頭期間允許命令或捲動")
			}
			if mode == "ticks" {
				g.Tick(89 * g.TimerTick())
				if !g.battleSpotting.header || g.battleSpotting.ticks != 1 {
					t.Fatal("標頭提早結束")
				}
			} else if mode == "left" {
				g.Click(600, 240)
			} else {
				g.RightClick(600, 240)
			}
			g.Tick(g.TimerTick())
			if g.battleSpotting.header || completed != 0 {
				t.Fatal("標頭等待沒有結束，或提前結算戰鬥")
			}
			for i := 0; i < 100 && g.battleSpotting.stage != 5; i++ {
				g.Tick(g.TimerTick())
			}
			if g.battleSpotting.stage != 5 || g.battleSpotting.ticks != 6 {
				t.Fatal("候選6 ticks等待未開始")
			}
			if mode == "ticks" {
				g.Tick(5 * g.TimerTick())
				if g.battleSpotting.stage != 5 || g.battleSpotting.ticks != 1 {
					t.Fatal("候選提早結束")
				}
			} else if mode == "left" {
				g.Click(600, 240)
			} else {
				g.RightClick(600, 240)
			}
			g.Tick(g.TimerTick())
			if g.battleSpotting.stage == 5 || completed != 0 {
				t.Fatal("候選等待沒有結束，或跳過其餘報告")
			}
			for i := 0; i < 10000 && g.battleSpotting.active; i++ {
				g.Tick(g.TimerTick())
			}
			if g.battleSpotting.active || completed != 1 {
				t.Fatal("第三階段未且僅完成一次")
			}
			g.Tick(g.TimerTick())
			if completed != 1 {
				t.Fatal("重複續接戰鬥")
			}
		})
	}
}
