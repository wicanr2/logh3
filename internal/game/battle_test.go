package game

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// newTestGame 用測試那份資源開一局，另外把戰場入口表接上。
func newTestGame(t *testing.T) *Game {
	t.Helper()
	res := load(t)
	e, err := rules.LoadBattleEntries(filepath.Join(root, "docs/internals/data/battle-entry-points.json"))
	if err != nil {
		t.Fatal(err)
	}
	res.Entries = e
	if res.Formation == nil {
		formation, err := rules.LoadFormationSlots(filepath.Join(root,
			"docs/internals/data/battle-formation-slots.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.Formation = formation
	}
	if res.Graph == nil {
		graph, err := rules.LoadStarGraph(filepath.Join(root,
			"docs/internals/data/star-graph.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.Graph = graph
	}
	loadBattleMaps(t, res)
	if res.Terrain == nil {
		terrain, err := rules.LoadBattleTerrain(filepath.Join(root,
			"docs/internals/data/battle-terrain.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.Terrain = terrain
	}
	if res.BattleMenu == nil {
		m, err := screen.LoadBattleMenuSpec(filepath.Join(root,
			"docs/internals/data/battle-menu.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.BattleMenu = m
	}
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	// 載 S3「巴米利恩會戰」的存檔，才有艦隊可以佈陣。
	orig := filepath.Join(root, "org_game/extracted")
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(orig, name))
		if err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", name)
		}
		return b
	}
	sl, err := rules.LoadSlot(rules.SlotFiles{
		Stellar: read("STELLAR.DTA"), Planet: read("PLANET.DTA"),
		Admiral: read("ADMIRAL.DTA"), Fleet: read("FLEET.DTA"),
		Ship: read("SHIP.DTA"), Work: read("WORK.DTA"),
		PlanetHeader: read("PLADAT.DAT"),
	})
	if err != nil {
		t.Fatal(err)
	}
	g.LoadSlot(sl)
	return g
}

// StartBattle 要把參戰艦隊放到入口格上，並且把分隊散開。
func TestStartBattleDeploys(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	// 找一個有艦隊的星域。
	region := -1
	for _, f := range g.fleets {
		if len(g.res.Entries.Regions[f.Region].Entries) > 0 {
			region = f.Region
			break
		}
	}
	if region < 0 {
		t.Skip("找不到有艦隊、又在入口表裡的星域")
	}
	g.StartBattle(region)
	if g.state != StateBattle {
		t.Fatalf("狀態是 %v，應該是 StateBattle", g.state)
	}
	if len(g.battleUnits) == 0 {
		t.Fatal("一個分隊都沒佈上去")
	}
	// 旗艦要落在該星域的某一個入口上。
	u := g.battleUnits[0]
	ok := false
	for _, e := range g.res.Entries.Regions[region].Entries {
		if u.X == e.X && u.Y == e.Y {
			ok = true
		}
	}
	if !ok {
		t.Errorf("旗艦在 (%d,%d)，不是星域 %d 的任何一個入口", u.X, u.Y, region)
	}
	// 分隊要散開，不能疊在一起。
	seen := map[[2]int]int{}
	for i, v := range g.battleUnits {
		if prev, ok := seen[[2]int{v.X, v.Y}]; ok {
			p := g.battleUnits[prev]
			t.Fatalf("第 %d 個分隊（艦隊 %d／分隊 %d／槽 %d）與第 %d 個（艦隊 %d／分隊 %d／槽 %d）疊在 (%d,%d)",
				prev, p.Fleet, p.Squadron, p.Slot, i, v.Fleet, v.Squadron, v.Slot, v.X, v.Y)
		}
		seen[[2]int{v.X, v.Y}] = i
	}
}

// 推一個回合，敵方那邊要真的動。
func TestAdvanceBattleTurnMoves(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.battleUnits = []rules.BattleUnit{
		{Fleet: 0, Faction: 0, X: 400, Y: 400, TargetX: 400, TargetY: 400, Strength: 1000, CruiseSpeed: 30, TurnSpeed: 4, Ships: 1000, FullShips: 1000, RepairCap: 1000, Morale: 100},
		{Fleet: 1, Faction: 1, X: 1400, Y: 1400, TargetX: 1400, TargetY: 1400, Strength: 1000, CruiseSpeed: 30, TurnSpeed: 4, Ships: 1000, FullShips: 1000, RepairCap: 1000, Morale: 100},
	}
	g.battleFleets = []rules.BattleFleet{{Index: 0, Faction: 0, X: 400, Y: 400, DestX: 400, DestY: 400, Action: 4, ScoutMode: 255}, {Index: 1, Faction: 1, X: 1400, Y: 1400, DestX: 1400, DestY: 1400, Action: 4, ScoutMode: 255}}
	g.state = StateBattle
	g.battleTurn = 1
	g.faction = 0
	g.SetTactic(ai.OriginalTactic{}, false)
	g.originalBattleAI = true // 驗的是原版戰術層會讓敵方移動（規格 061 在遊戲層關掉它）
	before := g.battleUnits[1]
	g.AdvanceBattleTurn()
	settleBattleExecution(t, g)
	after := g.battleUnits[1]
	if after.X == before.X && after.Y == before.Y {
		t.Fatal("敵方那一支沒動")
	}
	if g.battleTurn != 2 {
		t.Fatalf("回合是 %d，應該是 2", g.battleTurn)
	}
	// 玩家那一支沒下令，不該被 AI 動到目標。
	if g.battleUnits[0].TargetX != 400 || g.battleUnits[0].TargetY != 400 {
		t.Error("玩家那一支的目標被 AI 改了")
	}
}

// 按正式Tick完成已啟動回合；不繞過準備、移動或演出續接。
func settleBattleExecution(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 12000 && (g.battlePreparing || g.battleMoving || g.state == StateTurnReport || g.state == StateEngagement); i++ {
		g.Tick(time.Second / 60)
	}
	if g.battlePreparing || g.battleMoving || g.state == StateTurnReport || g.state == StateEngagement {
		t.Fatalf("回合續接未停止：%v/%d", g.state, g.battleTurn)
	}
}

// loadBattleMaps 補上戰場版面與 32 張星域地圖（缺原版檔就留空）。
// 戰鬥要行星，行星的位置要掃星域地圖（`docs/re/99`）；「指定待命行星」的
// 小圖也從同一張地圖縮出來（`docs/re/314`）。
func loadBattleMaps(t *testing.T, res *Resources) {
	t.Helper()
	if res.Battle == nil {
		spec, err := screen.LoadBattleSpec(filepath.Join(root,
			"docs/internals/data/screen-battle.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.Battle = spec
	}
	if res.BattleMaps != nil {
		return
	}
	maps := make([][]uint8, 32)
	for i := range maps {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted",
			fmt.Sprintf("MAP%02d.CMP", i)))
		if err != nil {
			continue
		}
		cells, err := asset.DecodeCMP(b, res.Battle.Field.MapCols, res.Battle.Field.MapRows)
		if err != nil {
			continue
		}
		maps[i] = cells
	}
	res.BattleMaps = maps
}

// advanceUntilBattle 一直按「實行」並把過場視窗與報告點掉，直到開戰或用完 turns 個回合。
// 回有沒有開戰。
//
// **不要用 `g.advanceTurn()` 直接推進**：回合中途會停在過場視窗與報告框（第六輪加的），
// 沒點掉的話下一次 `advanceTurn` 什麼都不做，於是「四十回合都沒會合」——那就是 #81 那
// 九支戰場測試一直被跳過的真正原因。舊的訊息寫「預設的 ai.Original
// 一道命令都不下」是**錯的**：量過，第一回合就有 17 支艦隊拿到目的地。
//
// S3「巴米利恩會戰」那份存檔用這條路**第一個回合就會在達希利（星域 26）開戰**。
func advanceUntilBattle(t *testing.T, g *Game, turns int) bool {
	t.Helper()
	g.state = StateCommandPhase
	for turn := 0; turn < turns; turn++ {
		pressExecute(g)
		for i := 0; i < 64 && g.State() != StateBattle; i++ {
			switch g.State() {
			case StateOccupationReport, StateRebellionReport, StateRegionFlipReport,
				StateTurnWindow, StateTurnReport, StateShipDelivery:
				g.Click(560, 125)
			default:
				i = 64
			}
		}
		if g.State() == StateBattle {
			return true
		}
	}
	return false
}

// centreViewOn 把戰場視野移到某顆行星上。戰場地圖比視窗大得多，預設視野常常
// 看不到指定的那一顆——測試若靠「挑一個畫面上的點」來驗東西，看不到就會靜靜
// 地 skip 掉（`TestFortressAimStoresBearing` 踩過，#81 收尾時抓到）。
func centreViewOn(g *Game, p rules.BattlePlanet) {
	s := g.res.Battle
	if s == nil {
		return
	}
	col := (p.X+fortressCentreBias)/s.Field.TileW - s.Field.Cols/2
	row := (p.Y+fortressCentreBias)/s.Field.TileH - s.Field.Rows/2
	if col < 0 {
		col = 0
	}
	if row < 0 {
		row = 0
	}
	g.battleView = screen.BattleView{Col: col, Row: row}
}
