package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/rules"
)

// RunNormalSpottingForTest供外部測試用正式bootstrap建立Game。
// 只固定seed並觀測續接，不注入原版RAM、鏡頭、情報或電腦政策。
func RunNormalSpottingForTest(t *testing.T, g *Game) {
	t.Helper()
	mem := map[string][]byte{}
	for name, hash := range map[string]string{
		"label": "a7b02b4d9f603265a7035ae6dad93aa7a7dc473dea7c27d9c93beef3c122f801",
		"end":   "6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4",
	} {
		b, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal79-r61-a", name+"-mem.bin"))
		if os.IsNotExist(err) {
			t.Skip("缺第61輪正常原版收據")
		}
		if err != nil || len(b) != 1048576 || fmtHash(sha256.Sum256(b)) != hash {
			t.Fatalf("%s原版收據身份錯誤：%v", name, err)
		}
		mem[name] = b
	}
	out := os.Getenv("LOGH3_NORMAL_SPOTTING_OUT")
	if out == "" {
		out = filepath.Join(t.TempDir(), "run")
	}
	if err := os.Mkdir(out, 0755); err != nil {
		t.Fatal(err)
	}
	click := func(x, y int) { g.MoveMouse(x, y); g.Click(x, y); g.Frame() }
	click(300, 270)
	if g.state == StateMainMenu {
		click(300, 270)
	}
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	wantState(t, g, StateLoadGame, "正常主選單取進度")
	click(300, 322)
	click(318, 208)
	wantState(t, g, StateCommandPhase, "正常S3帝國")
	g.rand94DC().SetSeed(0xD3DC167E)
	click(544, 125)
	click(560, 125)
	for i := 0; i < 15000 && g.state != StateBattle; i++ {
		switch g.state {
		case StateTurnThinking, StateTurnWindow, StateTurnReport, StateEngagement:
			g.Tick(time.Second / 60)
		case StateOccupationReport, StateRebellionReport, StateRegionFlipReport, StateAdmiralFateReport, StateShipDelivery, StatePhezzanReport:
			click(240, 240)
		default:
			t.Fatalf("正常實行停在%v", g.state)
		}
		g.Frame()
	}
	if g.state != StateBattle || g.turn != 380 || g.battleRegion != 26 || len(g.battleFleets) != 7 || len(g.preparationUnits()) != 55 || g.rand94DC().Seed() != 0xB6E4B78A {
		t.Fatal("正常首戰身份不同")
	}
	g.battleRand = rules.NewRand(0x12345678)
	states := map[string]any{}
	capture := func(name string) {
		p := filepath.Join(out, name)
		if err := os.Mkdir(p, 0755); err != nil {
			t.Fatal(err)
		}
		for n, b := range map[string][]byte{"SHIP.bin": g.slot.Ship, "FLEET.bin": g.slot.Fleet, "WORK.bin": g.slot.Work, "STELLAR.bin": g.res.Stellar, "PLANET.bin": g.slot.Planets.Data[rules.PlanetHeader:]} {
			if err := os.WriteFile(filepath.Join(p, n), b, 0644); err != nil {
				t.Fatal(err)
			}
		}
		f := g.Frame()
		if err := f.WriteIndex(filepath.Join(p, "screen.bin")); err != nil {
			t.Fatal(err)
		}
		if err := f.WritePNG(filepath.Join(p, "screen.png"), g.Palette()); err != nil {
			t.Fatal(err)
		}
		var palette []byte
		for _, rgb := range g.Palette() {
			palette = append(palette, rgb[:]...)
		}
		if err := os.WriteFile(filepath.Join(p, "palette.bin"), palette, 0644); err != nil {
			t.Fatal(err)
		}
		observers := []map[string]any{}
		for _, o := range g.battleSpotting.observers {
			observers = append(observers, map[string]any{"faction": o.faction, "xy": []int{o.x, o.y}, "radius": o.radius, "fleet": o.fleet, "planet": o.planet, "human": o.human})
		}
		states[name] = map[string]any{"battle_turn": g.battleTurn, "phase": g.battlePhase, "seed": fmt.Sprintf("%08X", g.battleRand.Seed()), "camera": []int{g.battleView.Col, g.battleView.Row}, "observers": observers, "next": g.battleSpotting.next, "header": g.battleSpotting.header}
		// 限定情報欄位驗收；全部raw差異由收據保存，不稱全紀錄一致。
		for pool := 0; pool < 352; pool++ {
			r := g.slot.ShipAt(pool)
			w := mem[name][0x668B2+pool*78:]
			if r[0x16]&1 != w[0x16]&1 || !bytes.Equal(r[0x17:0x19], w[0x17:0x19]) {
				t.Errorf("%s池%d Seen／Detail／Hold %x／%x", name, pool, r[0x16:0x19], w[0x16:0x19])
			}
		}
	}
	click(100, 100)
	click(100, 110)
	for i := 0; i < 15000 && !g.battleSpotting.active; i++ {
		g.Tick(time.Second / 60)
		g.Frame()
	}
	if !g.battleSpotting.active {
		t.Fatal("未經正式移動到索敵")
	}
	capture("label")
	if g.battleSpotting.header || len(g.battleSpotting.observers) != 10 {
		t.Fatalf("正常來源清單或無報告入口不同：%d", len(g.battleSpotting.observers))
	}
	completed := 0
	after := g.battleSpotting.after
	g.battleSpotting.after = func() { completed++; capture("end"); after() }
	for i := 0; i < 15000 && (g.battlePreparing || g.battleMoving || g.battleSpotting.active || g.state == StateTurnReport || g.state == StateEngagement); i++ {
		if g.battleSpotting.header {
			t.Fatal("正常首回合出現原版沒有的發現報告")
		}
		g.Tick(time.Second / 60)
		g.Frame()
	}
	if completed != 1 || g.battleTurn != 2 || g.battleSpotting.active {
		t.Fatal("索敵未且僅續接一次")
	}
	// 原版此自然局面第三階段全表未變；兩側各自驗完整增量，不掩蓋入口差異。
	for _, name := range []string{"SHIP", "FLEET", "WORK", "STELLAR", "PLANET"} {
		a, err := os.ReadFile(filepath.Join(out, "label", name+".bin"))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(out, "end", name+".bin"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Errorf("正常索敵階段改變原版未改的完整%s表", name)
		}
	}
	save := filepath.Join(out, "saved")
	if err := os.Mkdir(save, 0755); err != nil {
		t.Fatal(err)
	}
	if err := g.SaveSlot(save, 0); err != nil {
		t.Fatal(err)
	}
	loaded := newTestGame(t)
	loaded.res.SaveDir = save
	loaded.loadSlotFromDisk(0)
	if loaded.SaveErr() != nil || !bytes.Equal(loaded.slot.Ship, g.slot.Ship) || !bytes.Equal(loaded.slot.Fleet, g.slot.Fleet) {
		t.Fatal("正常索敵後SHIP／FLEET存讀不同")
	}
	b, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "states.json"), append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	if !t.Failed() {
		t.Log("正常bootstrap／主選單／S3／首戰實行：352筆Seen／Detail／Hold、無報告續接與正式存讀已核對；完整畫面與raw資料另記差異")
	}
}
