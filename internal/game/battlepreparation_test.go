package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

type preparationBatch struct {
	Mode   string
	In     string `json:"memory_in_sha256"`
	Out    string `json:"memory_out_sha256"`
	Screen string `json:"screen_sha256"`
}

func preparationBatchFor(t *testing.T, mode string) preparationBatch {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-preparation-lifecycle.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != "f8d4848a3e19232a66f92b18c2477647ea375a36379d2b5c792e3f5f48f1e650" {
		t.Fatal("原版生命週期收據身份不符")
	}
	var r struct{ Batches []preparationBatch }
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	for _, c := range r.Batches {
		if c.Mode == mode {
			return c
		}
	}
	t.Fatalf("無原版批次%s", mode)
	return preparationBatch{}
}

func preparationMemory(t *testing.T, mode, name, hash string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "workplace/oracle/battle79-life-"+mode+"-a", name))
	if os.IsNotExist(err) {
		t.Skip("缺原版固定收據，跳過；入口tools/battle_preparation_lifecycle_oracle.sh")
	}
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != hash {
		t.Fatal("原版記憶體／畫面已改變，須重新生成收據")
	}
	return b
}

// preparationGame 明示注入原版第一階段前狀態；不是正常玩家路徑聲明。
// 原始FLEET／SHIP經正式Slot解析與StartBattle，再轉入固定第一階段座標。
func preparationGame(t *testing.T, mem []byte) *Game {
	t.Helper()
	// 先走正常玩家路徑，再載入固定診斷資料；仍明示有狀態注入。
	g := battleTestGame(t)
	if g == nil {
		t.Skip("正常戰場入口受阻")
	}
	loadPreparationPresentation(t, g)
	hydratePreparationGame(t, g, mem)
	return g
}

func loadPreparationPresentation(t *testing.T, g *Game) {
	t.Helper()
	loadKinds(t, g)
	var err error
	panelBytes, err := os.ReadFile(filepath.Join(root, "org_game/extracted/ANI_PNL0.A98"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.BattlePanel, err = asset.DecodeAMGImage(panelBytes)
	if err != nil {
		t.Fatal(err)
	}
	g.res.SupplyPanel, err = screen.LoadSupplyPanelSpec(filepath.Join(root, "docs/internals/data/supply-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetWindow, err = screen.LoadFleetWindowSpec(filepath.Join(root, "docs/internals/data/fleet-window.json"))
	if err != nil {
		t.Fatal(err)
	}
	var names struct {
		Factions []struct{ Bytes string }
		Kinds    []struct{ Bytes string } `json:"ship_kinds"`
	}
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/faction-and-kind-names.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &names); err != nil {
		t.Fatal(err)
	}
	for _, v := range names.Factions {
		b, err := hex.DecodeString(v.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		g.res.FactionNames = append(g.res.FactionNames, b)
	}
	for _, v := range names.Kinds {
		b, err := hex.DecodeString(v.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		g.res.ShipKindNames = append(g.res.ShipKindNames, b)
	}
}

func hydratePreparationGame(t *testing.T, g *Game, mem []byte) {
	t.Helper()
	g.slot.Fleet = append([]byte(nil), mem[0x60BA4:0x60BA4+32*122]...)
	g.slot.Ship = append([]byte(nil), mem[0x668B2:0x668B2+352*78]...)
	copy(g.slot.Planets.Data[rules.PlanetHeader:], mem[0x53132:0x53132+81*82])
	g.LoadSlot(g.slot)
	g.StartBattle(26)
	// 診斷在正常進場後明示還原原版回合初態，包括新進場會重設的標記。
	copy(g.slot.Ship, mem[0x668B2:0x668B2+352*78])
	g.battleTurn = 6
	g.battleRand = rules.NewRand(0x12345678)
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		rec := g.slot.ShipAt(g.battleUnitPool(*u))
		u.X = int(binary.LittleEndian.Uint16(rec[0x34:]))
		u.Y = int(binary.LittleEndian.Uint16(rec[0x36:]))
		u.TargetX = int(binary.LittleEndian.Uint16(rec[0x28:]))
		u.TargetY = int(binary.LittleEndian.Uint16(rec[0x2A:]))
		u.Out = rec[0x1E]&0x80 != 0
		u.MarkerFlags = int(rec[0x19])
		u.Marked = u.MarkerFlags&0x80 != 0
		u.MarkerX = int(binary.LittleEndian.Uint16(rec[0x1A:]))
		u.MarkerY = int(binary.LittleEndian.Uint16(rec[0x1C:]))
	}
	for i := range g.battlePlanets {
		p := &g.battlePlanets[i]
		r := mem[0x53132+p.Index*82:]
		p.Owner = int(r[17])
		p.X = int(binary.LittleEndian.Uint16(r[70:]))
		p.Y = int(binary.LittleEndian.Uint16(r[72:]))
		p.Defence = int(binary.LittleEndian.Uint16(r[20:]))
		p.DefenceMax = int(binary.LittleEndian.Uint16(r[22:]))
		p.Recover = int(binary.LittleEndian.Uint16(r[24:]))
		p.Stock, _ = rules.ReadPlanetStock(r)
	}
	decodeSource := func(a int) rules.PreparationSource {
		off := binary.LittleEndian.Uint16(mem[a:])
		seg := binary.LittleEndian.Uint16(mem[a+2:])
		linear := int(seg)*16 + int(off)
		if linear == 0 {
			return rules.PreparationSource{}
		}
		if linear < 0x668B2 || (linear-0x668B2)%78 != 0 {
			t.Fatal("來源不指向原版池")
		}
		return rules.PreparationSource{Pool: (linear - 0x668B2) / 78, Known: true}
	}
	g.battleSources = rules.PreparationSources{Supply: decodeSource(0x5EF5E), Repair: decodeSource(0x5EF70)}
	refs := g.preparationUnits()
	if len(refs) != 55 || len(g.battlePlanets) != 3 {
		t.Fatalf("原版清單不符：%d／%d", len(refs), len(g.battlePlanets))
	}
	for i, r := range refs {
		want := int(binary.LittleEndian.Uint16(mem[0x5E3EA+i*2:]))
		if r.Pool != want {
			t.Fatalf("原版清單%d池%d，remake%d", i, want, r.Pool)
		}
	}
}

func comparePreparationState(t *testing.T, g *Game, want []byte) {
	t.Helper()
	for _, v := range g.preparationUnits() {
		u := v.Unit
		r := want[0x668B2+v.Pool*78:]
		values := []int{u.Ships, u.FullShips, u.RepairCap, u.Missiles, u.Fighters}
		expected := []int{int(binary.LittleEndian.Uint16(r[8:])), int(binary.LittleEndian.Uint16(r[10:])), int(binary.LittleEndian.Uint16(r[12:])), int(binary.LittleEndian.Uint32(r[14:])), int(binary.LittleEndian.Uint32(r[18:]))}
		if !reflect.DeepEqual(values, expected) {
			t.Errorf("池%d資源%v，原版%v", v.Pool, values, expected)
		}
	}
	for _, p := range g.battlePlanets {
		r := want[0x53132+p.Index*82:]
		s, _ := rules.ReadPlanetStock(r)
		if p.Defence != int(binary.LittleEndian.Uint16(r[20:])) || p.Stock != s {
			t.Errorf("行星%d資源／防衛力不符", p.Index)
		}
	}
	for _, bf := range g.battleFleets {
		f := g.fleetByIndex(bf.Index)
		r := want[0x60BA4+bf.Index*122:]
		if f.Morale != int(r[0x3C]) || f.Flags != int(r[0]) {
			t.Errorf("艦隊%d士氣%d/%02X，原版%d/%02X", bf.Index, f.Morale, f.Flags, r[0x3C], r[0])
		}
	}
	for _, s := range []struct {
		source rules.PreparationSource
		addr   int
	}{{g.battleSources.Supply, 0x5EF5E}, {g.battleSources.Repair, 0x5EF70}} {
		ptr := int(binary.LittleEndian.Uint16(want[s.addr:])) + 16*int(binary.LittleEndian.Uint16(want[s.addr+2:]))
		if !s.source.Known || 0x668B2+s.source.Pool*78 != ptr {
			t.Errorf("來源池%+v，原版線性%X", s.source, ptr)
		}
	}
}

func TestBattlePreparationWholePhaseMatchesOriginal(t *testing.T) {
	for _, mode := range []string{"supply", "repair", "morale"} {
		t.Run(mode, func(t *testing.T) {
			end := preparationBatchFor(t, mode+"-end")
			before := preparationMemory(t, end.Mode, "mem-in.bin", end.In)
			want := preparationMemory(t, end.Mode, "mem-out.bin", end.Out)
			g := preparationGame(t, before)
			completed := 0
			g.beginBattlePreparation(func() { completed++ })
			if mode == "supply" && g.battlePreparationSupply == nil {
				t.Fatal("未停在正式補給結果幀")
			}
			if mode == "repair" && g.battlePreparationRepair == nil {
				t.Fatal("未停在正式修理結果幀")
			}
			if mode == "morale" && (g.state != StateTurnReport || !g.battlePreparing) {
				t.Fatal("士氣報告未保留第一階段續接")
			}
			if mode != "morale" {
				panel := preparationBatchFor(t, mode+"-panel")
				oracle := preparationMemory(t, panel.Mode, "screen.bin", panel.Screen)
				got := g.Frame().Pix
				s := g.res.SupplyPanel
				diff := 0
				y0 := s.CellY*ui.FrameTileH + ui.FrameCellOriginY
				for y := y0; y < y0+s.Rows*ui.FrameTileH; y++ {
					for x := s.CellX * ui.FrameTileW; x < (s.CellX+s.Cols)*ui.FrameTileW; x++ {
						if got[y*640+x] != oracle[y*640+x] {
							diff++
						}
					}
				}
				writePreparationFrame(t, g, mode+"-panel")
				if diff != 0 {
					t.Errorf("結果面板像素不同%d", diff)
				}
				// 玩家點擊不會提早續接／重跑；下一次時間更新自行繼續。
				g.Click(240, 200)
				g.RightClick(240, 200)
				if completed != 0 {
					t.Fatal("結果幀被點擊跳過")
				}
			}
			for i := 0; i < 2000 && g.battlePreparing; i++ {
				g.Tick(time.Second / 60)
			}
			if completed != 1 || g.battlePreparing {
				t.Fatalf("續接完成%d次", completed)
			}
			comparePreparationState(t, g, want)
			if g.battleRand.Seed() != binary.LittleEndian.Uint32(want[0x5E00C:]) {
				t.Fatal("第一階段消費了原版未消費的亂數")
			}
			// 防衛力／資源結束後再移動；測試callback不移動，不宣稱其他七步一致。
			g.Withdraw() // 正式戰後回寫；不是只呼叫孤立的資源helper。
			if g.SaveErr() != nil {
				t.Fatal(g.SaveErr())
			}
			for _, u := range g.battleUnits {
				r := g.slot.ShipAt(g.battleUnitPool(u))
				w := want[0x668B2+g.battleUnitPool(u)*78:]
				if !bytes.Equal(r[8:22], w[8:22]) {
					t.Errorf("池%d戰後回寫不符", g.battleUnitPool(u))
				}
			}
			// 戰後掃完星域可能抵達月初；依正式右鍵路徑關閉訂造報告，
			// 不繞過存檔「待結算」閘門。月初結算後才保存當前完整資料。
			dismissMonthReports(g)
			if g.monthSettlementPending {
				t.Fatalf("戰後月初報告未能結清，狀態%v", g.state)
			}
			ship := append([]byte(nil), g.slot.Ship...)
			planet := append([]byte(nil), g.res.Planets.Data...)
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			g.res.SaveDir = dir
			g.loadSlotFromDisk(0)
			if g.SaveErr() != nil {
				t.Fatal(g.SaveErr())
			}
			if !bytes.Equal(ship, g.slot.Ship) || !bytes.Equal(planet, g.res.Planets.Data) {
				t.Fatal("正式存讀丟失SHIP／PLANET資源")
			}
		})
	}
}

// 固定原版士氣局面，另將防衛力降為0；這是順序診斷，不宣稱新原版同狀態。
func TestBattlePreparationMoraleBlocksFollowingPhase(t *testing.T) {
	end := preparationBatchFor(t, "morale-end")
	g := preparationGame(t, preparationMemory(t, end.Mode, "mem-in.bin", end.In))
	g.battlePlanets[0].Defence, g.battlePlanets[0].DefenceMax, g.battlePlanets[0].Recover = 0, 100, 10
	positions := append([]rules.BattleUnit(nil), g.battleUnits...)
	completed := 0
	g.beginBattlePreparation(func() { completed++ })
	if g.state != StateTurnReport || g.battlePlanets[0].Defence != 0 || completed != 0 {
		t.Fatal("士氣報告尚未關閉就回復防衛力或續接")
	}
	g.AdvanceBattleTurn() // 重複實行不能重跑本階段／移動。
	if g.battleTurn != 6 {
		t.Fatal("等待中推進回合")
	}
	for i, u := range g.battleUnits {
		if u.X != positions[i].X || u.Y != positions[i].Y {
			t.Fatal("等待中移動")
		}
	}
	for i := 0; i < 2000 && g.battlePreparing; i++ {
		g.Tick(time.Second / 60)
	}
	if completed != 1 || g.battlePlanets[0].Defence != 10 {
		t.Fatal("防衛力未於報告後回復一次")
	}
	g.Tick(time.Second)
	if completed != 1 || g.battlePlanets[0].Defence != 10 {
		t.Fatal("續接重複執行")
	}
}

func TestBattlePreparationNormalPlayerExecute(t *testing.T) {
	g := battleTestGame(t)
	if g == nil {
		return
	}
	loadPreparationPresentation(t, g)
	var err error
	g.res.BattleMenu, err = screen.LoadBattleMenuSpec(filepath.Join(root, "docs/internals/data/battle-menu.json"))
	if err != nil {
		t.Fatal(err)
	}
	turn := g.battleTurn
	g.MoveMouse(600, 240)
	g.Click(600, 240)
	if !g.battleMenu {
		t.Fatal("正常戰場未開啟戰術選單")
	}
	l := &g.res.BattleMenu.Layout
	x := g.battleMenuX + l.Highlight.X - l.Window.X + 2
	y := g.battleMenuY + l.Highlight.Y - l.Window.Y + 2
	g.MoveMouse(x, y)
	g.Click(x, y) // 正式選「實行」，不直接呼叫第一階段。
	for i := 0; i < 4000 && (g.battlePreparing || g.battleMoving || g.state == StateTurnReport || g.state == StateEngagement); i++ {
		g.Frame()
		g.Tick(time.Second / 60)
	}
	if !g.battleSources.Supply.Known || g.battlePreparing || g.battleTurn != turn+1 || g.state != StateBattle {
		t.Fatalf("正常實行未完成第一階段及後續：回合%d 狀態%v 來源%+v", g.battleTurn, g.state, g.battleSources)
	}
	writePreparationFrame(t, g, "normal-execute")
}

func writePreparationFrame(t *testing.T, g *Game, label string) {
	t.Helper()
	p := g.Palette()
	var pal []byte
	for _, rgb := range p {
		pal = append(pal, rgb[:]...)
	}
	for _, file := range []struct {
		ext  string
		data []byte
	}{{"bin", g.Frame().Pix}, {"pal", pal}} {
		if err := os.WriteFile(filepath.Join(root, "workplace/battle79-round28-"+label+"."+file.ext), file.data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
