package game

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// newGameFleet 編一支池索引未配的新艦隊，插進艦隊表後回它的 Index。
func newGameFleet(t *testing.T, g *Game, faction, slot, region int, used map[int]bool) int {
	t.Helper()
	admiral := -1
	for i := 0; g.admiralRecord(i) != nil; i++ {
		rec := g.admiralRecord(i)
		if !used[i] && len(rec) >= rules.AdmiralRecord && rules.AdmiralFaction(rec) == faction {
			admiral = i
			break
		}
	}
	if admiral < 0 {
		t.Fatalf("陣營 %d 找不到提督", faction)
	}
	used[admiral] = true
	kind := 0
	full := rules.ShipsPerUnit[kind]
	grp := rules.FleetGroup{Kind: kind, Ships: full, Full: full, Slot2: kind, Slot3: rules.FlagshipSlot3, Pool: -1}
	rules.InitGroupStock(g.res.ShipKinds, faction, &grp)
	index := rules.FleetTableIndex(faction, slot)
	g.insertFleet(rules.Fleet{
		Index: index, Number: slot + 1, Faction: faction, Admiral: admiral, Adjutant: -1,
		Region: region, Target: rules.NoTarget, Move: fleetMoveParked, Parked: true,
		Scout: -1, Morale: rules.NewFleetMorale, Composition: []rules.FleetGroup{grp}, Groups: 1,
	})
	return index
}

// fleetAt 回 Index 那一支在 g.fleets 的位置。
func fleetAt(t *testing.T, g *Game, index int) int {
	t.Helper()
	for i := range g.fleets {
		if g.fleets[i].Index == index {
			return i
		}
	}
	t.Fatalf("艦隊表沒有 %d", index)
	return -1
}

// arriveFrom 走行軍入口把那一支從所在星域帶到 region。
func arriveFrom(t *testing.T, g *Game, index, region int) {
	t.Helper()
	i := fleetAt(t, g, index)
	from := g.moveFleetTo(i, region, 0)
	g.arriveFleet(i, from)
}

// marchEntry 回一個有非首都入口的星域，並回那個入口。
func marchEntry(t *testing.T, g *Game) (int, rules.BattleEntry) {
	t.Helper()
	for r := range g.res.Entries.Regions {
		if len(g.battleCells(r)) == 0 {
			continue
		}
		for _, e := range g.res.Entries.Regions[r].Entries {
			if e.From != r {
				return r, e
			}
		}
	}
	t.Skip("沒有可用的入口")
	return 0, rules.BattleEntry{}
}

// newGameBattleGame 回一局還沒存過檔的新遊戲，艦隊表清空，不等月初報告。
func newGameBattleGame(t *testing.T) *Game {
	t.Helper()
	g := newTestGame(t)
	g.startNewGame(0)
	g.fleets = nil
	g.monthSettlementPending = false
	if g.slot != nil {
		t.Fatal("新遊戲不該已有存檔")
	}
	return g
}

// fleetFacing 回原始表裡那一支的航向（FLEET+4A）。
func fleetFacing(g *Game, index int) int {
	return int(g.slot.Fleet[index*rules.FleetRecord+0x4A])
}

// 新遊戲沒讀過也沒存過檔（`g.slot` 是 nil）就在星域掃描開第一場戰鬥。
// 修正前這裡 nil dereference（`BattleFleetsFor`），堆疊與使用者回報相同。
func TestNewGameFirstBattleEntersWithBothSides(t *testing.T) {
	g := newGameBattleGame(t)
	region, entry := marchEntry(t, g)
	used := map[int]bool{}
	empire := newGameFleet(t, g, 0, 0, entry.From, used)
	alliance := newGameFleet(t, g, 1, 0, region, used)
	arriveFrom(t, g, empire, region)
	if got, ok := g.arrivalFrom[empire]; !ok || got != entry.From {
		t.Fatalf("新遊戲到達應記來源 %d，記到 %d（%v）", entry.From, got, ok)
	}

	g.rebellionOccupation[region] = rules.RegionOccupation{Contested: true}
	g.scanRegionsFrom(region)
	for n := 0; g.state == StateTurnReport && n < 16; n++ {
		g.Click(560, 125)
	}
	if g.state != StateBattle {
		t.Fatalf("狀態是 %v，應該進戰場（saveErr=%v）", g.state, g.saveErr)
	}
	if g.slot == nil {
		t.Fatal("進戰場要建出原池")
	}
	if len(g.battleFleets) != 2 {
		t.Fatalf("參戰艦隊 %d 支，應該 2 支", len(g.battleFleets))
	}
	byFaction := [2]int{}
	cells := g.battleCells(region)
	for _, u := range g.battleUnits {
		byFaction[u.Faction]++
		if !g.res.Terrain.Passable(cells, u.X, u.Y) {
			t.Errorf("艦隊 %d 分隊 %d 落在不可通行的 (%d,%d)", u.Fleet, u.Squadron, u.X, u.Y)
		}
		if u.X == 0 && u.Y == 0 {
			t.Errorf("艦隊 %d 分隊 %d 沒有配置，還在 (0,0)", u.Fleet, u.Squadron)
		}
	}
	if byFaction[0] == 0 || byFaction[1] == 0 {
		t.Fatalf("兩軍分隊數 %v，戰場上要有雙方", byFaction)
	}
	// 行軍到達要用來源星域的入口，不能退成自己星域的規則。
	f := g.fleets[fleetAt(t, g, empire)]
	if f.X != entry.X || f.Y != entry.Y {
		t.Errorf("行軍到達的中心 (%d,%d)，入口是 (%d,%d)", f.X, f.Y, entry.X, entry.Y)
	}
	if len(g.arrivalFrom) != 0 {
		t.Errorf("配置完應清掉來源記號，還有 %v", g.arrivalFrom)
	}
	if g.fleets[fleetAt(t, g, alliance)].X == 0 {
		t.Error("沒行軍的那一支也要配置")
	}
}

// 到達之後、開戰之前存過檔：`g.slot` 已經有了，來源記號要留到開戰才用。
func TestNewGameSaveBetweenArrivalAndBattleKeepsEntry(t *testing.T) {
	g := newGameBattleGame(t)
	region, entry := marchEntry(t, g)
	used := map[int]bool{}
	empire := newGameFleet(t, g, 0, 0, entry.From, used)
	newGameFleet(t, g, 1, 0, region, used)
	arriveFrom(t, g, empire, region)
	if err := g.SaveSlot(t.TempDir(), 0); err != nil {
		t.Fatal(err)
	}
	g.startPlacedBattle(region)
	f := g.fleets[fleetAt(t, g, empire)]
	if f.X != entry.X || f.Y != entry.Y {
		t.Errorf("中心 (%d,%d)，入口是 (%d,%d)", f.X, f.Y, entry.X, entry.Y)
	}
}

// 對戰模式也是新遊戲開局，`g.slot` 同樣是 nil。
func TestVersusFirstBattleEnters(t *testing.T) {
	g := newTestGame(t)
	g.startVersus()
	if g.slot != nil {
		t.Fatal("對戰開局不該已有存檔")
	}
	g.fleets = nil
	region, entry := marchEntry(t, g)
	used := map[int]bool{}
	empire := newGameFleet(t, g, 0, 0, entry.From, used)
	newGameFleet(t, g, 1, 0, region, used)
	arriveFrom(t, g, empire, region)
	g.startPlacedBattle(region)
	if g.state != StateBattle || len(g.battleFleets) != 2 {
		t.Fatalf("狀態 %v、參戰 %d 支（saveErr=%v）", g.state, len(g.battleFleets), g.saveErr)
	}
	byFaction := [2]int{}
	for _, u := range g.battleUnits {
		byFaction[u.Faction]++
	}
	if byFaction[0] == 0 || byFaction[1] == 0 {
		t.Fatalf("兩軍分隊數 %v", byFaction)
	}
}

// 正常按「實行」：人類艦隊走進敵艦隊停駐的星域，把報告點掉就進戰場，打完還能保存。
func TestNewGameFirstBattleFromExecuteSavesAndResumes(t *testing.T) {
	g := newGameBattleGame(t)
	region, entry := -1, rules.BattleEntry{}
	for r := range g.res.Entries.Regions {
		if len(g.battleCells(r)) == 0 || g.regionOwner(r) != 0 {
			continue
		}
		for _, e := range g.res.Entries.Regions[r].Entries {
			if e.From != r && g.regionOwner(e.From) == 0 {
				region, entry = r, e
			}
		}
	}
	if region < 0 {
		t.Skip("帝國沒有相鄰的兩個星域")
	}
	used := map[int]bool{}
	empire := newGameFleet(t, g, 0, 0, entry.From, used)
	newGameFleet(t, g, 1, 0, region, used)
	i := fleetAt(t, g, empire)
	g.fleets[i].Target, g.fleets[i].Move, g.fleets[i].Parked = region, rules.MoveMove, false
	// 電腦不下命令，敵艦隊才會留在原地等人走進來。
	g.SetBrain(nil)

	if !advanceUntilBattle(t, g, 1) {
		t.Fatalf("第一回合沒開戰，停在 %v（saveErr=%v）", g.State(), g.saveErr)
	}
	if g.battleRegion != region {
		t.Fatalf("開戰在星域 %d，應該是 %d", g.battleRegion, region)
	}
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "FLEET.DT0"))
	if err != nil {
		t.Fatal(err)
	}
	r := b[empire*rules.FleetRecord : (empire+1)*rules.FleetRecord]
	if x, y := int(binary.LittleEndian.Uint16(r[0x5C:])), int(binary.LittleEndian.Uint16(r[0x5E:])); x != entry.X || y != entry.Y {
		t.Errorf("存檔裡的中心 (%d,%d)，入口是 (%d,%d)", x, y, entry.X, entry.Y)
	}
	g.EndBattle()
	if g.saveErr != nil {
		t.Fatalf("戰後同步失敗：%v", g.saveErr)
	}
}

// 原版收據：新編、沒行軍的艦隊在首都走自己星域的入口。
//   - `fcmd-base`：帝國第 0 支、星域 0，中心 (336,280)、航向 34。
//   - `t260_post`：同盟第 28 支、星域 31，中心 (1696,1740)、航向 162。
func TestNewFleetAtCapitalUsesOwnRegionEntry(t *testing.T) {
	for _, c := range []struct {
		name                  string
		faction, slot, region int
		index                 int
		x, y, facing          int
	}{
		{"帝國", 0, 0, 0, 0, 336, 280, 34},
		{"同盟", 1, 12, 31, 28, 1696, 1740, 162},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newGameBattleGame(t)
			used := map[int]bool{}
			own := newGameFleet(t, g, c.faction, c.slot, c.region, used)
			if own != c.index {
				t.Fatalf("第 %d 槽算到表位 %d，收據是 %d", c.slot, own, c.index)
			}
			g.startPlacedBattle(c.region)
			if g.state != StateBattle {
				t.Fatalf("狀態 %v（saveErr=%v）", g.state, g.saveErr)
			}
			f := g.fleets[fleetAt(t, g, own)]
			if f.X != c.x || f.Y != c.y {
				t.Errorf("中心 (%d,%d)，原版 (%d,%d)", f.X, f.Y, c.x, c.y)
			}
			if got := fleetFacing(g, own); got != c.facing {
				t.Errorf("航向 %d，原版 %d", got, c.facing)
			}
			if f.Action != 4 {
				t.Errorf("FLEET+48 %d，原版 4", f.Action)
			}
		})
	}
}

// `297B:0ABF`：非首都沒有 from 等於自己的入口，新編艦隊用待命行星的座標。
func TestNewFleetAwayFromCapitalUsesStandbyPlanet(t *testing.T) {
	g := newGameBattleGame(t)
	region, k := -1, -1
	for r := 0; r < rules.RegionCount; r++ {
		if _, ok := g.res.Entries.EntryFor(r, r); ok || len(g.battleCells(r)) == 0 {
			continue
		}
		if p := g.res.Planets.StandbyPlanet(0, r); p >= 0 {
			region, k = r, p
			break
		}
	}
	if region < 0 {
		t.Skip("找不到帝國的非首都行星")
	}
	used := map[int]bool{}
	own := newGameFleet(t, g, 0, 0, region, used)
	sites := g.loadRegionPlanetSites(region)
	if k >= len(sites) {
		t.Skipf("星域 %d 的行星 %d 沒有座標", region, k)
	}
	g.startPlacedBattle(region)
	f := g.fleets[fleetAt(t, g, own)]
	if f.X != sites[k].X || f.Y != sites[k].Y {
		t.Errorf("中心 (%d,%d)，待命行星 (%d,%d)", f.X, f.Y, sites[k].X, sites[k].Y)
	}
}

// 讀檔局新編的艦隊 Pool 是 -1：到達當下要先配好池，否則配置會安靜地沒做，
// 兩軍分隊全部疊在地圖左上角（與新遊戲同一個洞的另一面）。
func TestLoadedGameFreshFleetArrivalIsPlaced(t *testing.T) {
	g := newTestGame(t)
	region, entry := marchEntry(t, g)
	kept := g.fleets[:0]
	used := map[int]bool{}
	first := rules.FleetTableIndex(0, 0)
	for _, f := range g.fleets {
		if f.Region != region && f.Index != first {
			kept = append(kept, f)
			used[f.Admiral], used[f.Adjutant] = true, true
		}
	}
	g.fleets = kept
	if err := g.syncFleetTables(); err != nil {
		t.Fatal(err)
	}
	fresh := newGameFleet(t, g, 0, 0, entry.From, used)
	arriveFrom(t, g, fresh, region)
	if g.saveErr != nil {
		t.Fatal(g.saveErr)
	}
	f := g.fleets[fleetAt(t, g, fresh)]
	if f.X != entry.X || f.Y != entry.Y {
		t.Errorf("中心 (%d,%d)，入口是 (%d,%d)", f.X, f.Y, entry.X, entry.Y)
	}
	if f.Composition[0].Pool < 0 {
		t.Error("到達之後要有池索引")
	}
}

// 已經有座標的艦隊不重新配置：讀檔局的原池位置是原版的，不能被蓋掉。
func TestPlacedFleetsKeepTheirPositions(t *testing.T) {
	g := newTestGame(t)
	type pos struct{ x, y int }
	want := map[int]pos{}
	for _, f := range g.fleets {
		if fleetNeverPlaced(&f) {
			t.Fatalf("S3 的艦隊 %d 不該算沒配置過", f.Index)
		}
		want[f.Index] = pos{f.X, f.Y}
	}
	g.placeUnplacedFleets()
	for _, f := range g.fleets {
		if got := want[f.Index]; got != (pos{f.X, f.Y}) {
			t.Errorf("艦隊 %d 被移動：%v → (%d,%d)", f.Index, got, f.X, f.Y)
		}
	}
}

// 不注入敵方：電腦照原版 AI 自己下命令，人類只編一支，趕在電腦到達時往星域 15 占領，
// 一路按「實行」到第一次開戰。這是使用者回報的路徑：新遊戲、沒存過檔、報告點掉之後在星域 15 進戰場
// （截圖堆疊 `startBattle(…, 0xf, 0x1)`）；原版 AI 開局也往那裡派艦隊（收據 `t260_post`）。
func TestNewGameNaturalFirstBattle(t *testing.T) {
	const contested = 15
	g := newGameBattleGame(t)
	g.SetBrain(g.OriginalBrain())
	from := -1
	for _, e := range g.res.Entries.Regions[contested].Entries {
		if e.From != contested {
			from = e.From
			break
		}
	}
	if from < 0 {
		t.Skip("星域 15 沒有入口")
	}
	used := map[int]bool{}
	human := newGameFleet(t, g, 0, 0, from, used)
	g.state = StateCommandPhase
	// 月初報告與回合中途的視窗照玩家操作點掉；碰到不認得的畫面就停，不靠計數掩蓋。
	for step := 0; step < 200000 && g.State() != StateBattle; step++ {
		switch g.State() {
		case StateCommandPhase:
			if g.turn >= 200 {
				t.Fatalf("200 回合內雙方沒有會合")
			}
			// 電腦的第 23 支第 5 回合走到星域 15；同一回合下命令，兩邊才碰得上。
			if g.turn == 4 {
				i := fleetAt(t, g, human)
				g.fleets[i].Target, g.fleets[i].Move, g.fleets[i].Parked = contested, rules.MoveOccupy, false
			}
			pressExecute(g)
		case StateMainGame, StateShipOrder:
			g.RightClick(300, 240)
		case StateOccupyAsk:
			t.Fatalf("第 %d 回合問要不要占領星域 %d", g.turn, g.occupyAsk.next)
		case StateOccupationReport, StateRebellionReport, StateRegionFlipReport,
			StateTurnWindow, StateTurnReport, StateShipDelivery, StateAdmiralFateReport:
			g.Click(560, 125)
		default:
			t.Fatalf("第 %d 回合停在 %v", g.turn, g.State())
		}
	}
	if g.State() != StateBattle {
		t.Fatalf("沒有開戰，第 %d 回合停在 %v", g.turn, g.State())
	}
	t.Logf("第 %d 回合在星域 %d 開戰，%d 支艦隊、%d 個分隊", g.turn, g.battleRegion, len(g.battleFleets), len(g.battleUnits))
	if g.battleRegion != contested || len(g.battleFleets) != 2 {
		t.Fatalf("開戰在星域 %d、%d 支艦隊，截圖是星域 15、2 支", g.battleRegion, len(g.battleFleets))
	}
	if g.slot == nil {
		t.Fatal("開戰要建出原池")
	}
	byFaction := [2]int{}
	cells := g.battleCells(g.battleRegion)
	for _, u := range g.battleUnits {
		byFaction[u.Faction]++
		if u.X == 0 && u.Y == 0 {
			t.Errorf("艦隊 %d 分隊 %d 還在 (0,0)", u.Fleet, u.Squadron)
		}
		if !g.res.Terrain.Passable(cells, u.X, u.Y) {
			t.Errorf("艦隊 %d 分隊 %d 落在不可通行的 (%d,%d)", u.Fleet, u.Squadron, u.X, u.Y)
		}
	}
	if byFaction[0] == 0 || byFaction[1] == 0 {
		t.Fatalf("兩軍分隊數 %v", byFaction)
	}
	// 打一回合：`startBattle` 之後的戰場層都假設有原池。
	g.AdvanceBattleTurn()
	if g.saveErr != nil {
		t.Fatal(g.saveErr)
	}
}

// 電腦解散一支之後在同一格重編：新艦隊沒帶 Scout（零值），舊列又還留著的話，同步會把零值
// 當成池 #0，回「擁有者／艦種不符」。正常開戰與存檔都會卡在這裡。
func TestDisbandThenRebuildSameSlotSyncs(t *testing.T) {
	g := newTestGame(t)
	if len(g.fleets) == 0 {
		t.Skip("S3 沒有艦隊")
	}
	i := 0
	for j, f := range g.fleets {
		if f.Faction == 1 {
			i = j
			break
		}
	}
	old := g.fleets[i]
	oldPool := old.Composition[0].Pool
	g.disbandFleet(i)
	if g.slot.FleetUsed(old.Index) {
		t.Fatal("解散之後原始表那一列要作廢")
	}
	if r := g.slot.ShipAt(oldPool); r[1] != 0xFF {
		t.Fatalf("解散之後池 %d 的擁有者 %d，應該是 FF", oldPool, r[1])
	}
	full := rules.ShipsPerUnit[0]
	grp := rules.FleetGroup{Kind: 0, Ships: full, Full: full, Slot2: 0, Slot3: rules.FlagshipSlot3, Pool: -1}
	rules.InitGroupStock(g.res.ShipKinds, old.Faction, &grp)
	g.insertFleet(rules.Fleet{Index: old.Index, Number: old.Number, Faction: old.Faction, Admiral: old.Admiral,
		Adjutant: -1, Region: old.Region, Target: rules.NoTarget, Move: fleetMoveParked, Parked: true,
		Morale: rules.NewFleetMorale, Composition: []rules.FleetGroup{grp}, Groups: 1})
	if err := g.syncFleetTables(); err != nil {
		t.Fatal(err)
	}
	f := g.fleets[fleetAt(t, g, old.Index)]
	if f.Composition[0].Pool < 0 || f.Scout < 0 {
		t.Fatalf("重編之後要配好池：%+v", f)
	}
	if r := g.slot.ShipAt(f.Scout); int(r[1]) != old.Index {
		t.Fatalf("索敵隊池 %d 的擁有者 %d，應該是 %d", f.Scout, r[1], old.Index)
	}
}

// playBattleToEnd 照玩家操作把這一場打完：有交火輸入就點（與 GUI 驅動器同一落點），
// 其餘時間按實行，等戰場自己走完九回合的收尾。
func playBattleToEnd(t *testing.T, g *Game) {
	t.Helper()
	for step := 0; g.State() == StateBattle || g.State() == StateTurnReport || g.State() == StateEngagement; step++ {
		if step > 400000 {
			t.Fatalf("戰場打不完：回合 %d、狀態 %v", g.battleTurn, g.State())
		}
		if g.State() == StateBattle && g.saveErr != nil {
			t.Fatalf("戰場中出錯：%v", g.saveErr)
		}
		if g.State() == StateBattle {
			v := g.TacticalObservation()
			switch v.CombatInput {
			case "":
				if !g.battlePreparing && !g.battleMoving && !g.battleSpotting.active && !g.fleetCombat.active &&
					!g.battleFinishing && !g.fortressEngagement {
					g.AdvanceBattleTurn()
					continue
				}
			case "target":
				g.Click(240, 470)
				continue
			default:
				g.Click(v.CombatPoint[0], v.CombatPoint[1])
				continue
			}
		}
		g.Tick(time.Second / 60)
	}
}

// RunNewGameContinuesForTest 全程只用玩家操作，不注入狀態：主選單開新遊戲、編第一支艦隊、
// 下令占領 contested，一路「實行」到開戰，九回合打完，繼續玩到再開戰或第 80 回合，
// 用「存進度」存檔、「取進度」讀回，再繼續玩。沒存過檔所以第一場開戰時 `g.slot` 是 nil。
// 資源由呼叫端用正式 `bootstrap.Load` 給。
func RunNewGameContinuesForTest(t *testing.T, res *Resources, menuRow, contested, loadRowY int) {
	t.Helper()
	dir := t.TempDir()
	res.SaveDir = dir
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	y := menuRowY(t, res, menuRow)
	// 第一下可能只是跳過開頭動畫。
	g.Click(300, y)
	if g.State() == StateMainMenu {
		g.Click(300, y)
	}
	for _, c := range []byte("138\r") {
		g.Key(c)
	}
	for i := 0; i < 4 && g.State() != StateCommandPhase; i++ {
		g.RightClick(300, 240)
	}
	if g.State() != StateCommandPhase {
		t.Fatalf("正常開新遊戲停在 %v", g.State())
	}
	for _, c := range [][2]int{
		{352, 125}, {400, 125}, {400, 85}, {488, 330}, // 戰略命令 → 出擊命令 → 第 0 列 → 出擊
		{480, 350}, {510, 350}, // 確認 → 決定
		{450, 308}, {490, 345}, {497, 332}, // 清單第 0 列 → 作戰 → 占領
	} {
		g.Click(c[0], c[1])
	}
	x, py := -1, -1
	for yy := mapY; yy < mapY+mapH && x < 0; yy++ {
		for xx := mapX; xx < mapX+mapW; xx++ {
			if g.mapRegionAt(xx, yy) == contested {
				x, py = xx, yy
				break
			}
		}
	}
	g.Click(x, py)
	g.Click(432, 263) // 決定
	if g.State() != StateCommandPhase || len(g.fleets) == 0 || g.fleets[0].Target != contested {
		t.Fatalf("編隊下令之後在 %v，艦隊 %+v", g.State(), g.fleets)
	}
	if g.slot != nil {
		t.Fatal("新遊戲不該已有存檔")
	}
	g.SetTactic(ai.OriginalTactic{}, true)
	battles := 0
	// play 一路玩到 done 成立，打到的每一場都走完九回合的正常收尾。
	play := func(done func() bool) {
		t.Helper()
		for step := 0; step < 200000; step++ {
			if g.State() == StateCommandPhase && done() {
				return
			}
			switch g.State() {
			case StateCommandPhase:
				pressExecute(g)
			case StateMainGame, StateShipOrder:
				g.RightClick(300, 240)
			case StateOccupationReport, StateRebellionReport, StateRegionFlipReport,
				StateTurnWindow, StateTurnReport, StateShipDelivery, StateAdmiralFateReport:
				g.Click(560, 125)
			case StateEngagement:
				g.Tick(time.Second / 60)
			case StateTurnThinking:
				thinkThrough(g)
			case StateBattle:
				battles++
				t.Logf("第 %d 場：回合 %d、星域 %d、%d 支艦隊、%d 個分隊", battles, g.turn, g.battleRegion, len(g.battleFleets), len(g.battleUnits))
				if g.slot == nil {
					t.Fatal("開戰要建出原池")
				}
				for _, u := range g.battleUnits {
					if u.X == 0 && u.Y == 0 {
						t.Errorf("艦隊 %d 分隊 %d 還在 (0,0)", u.Fleet, u.Squadron)
					}
				}
				playBattleToEnd(t, g)
				if g.battleTurn <= rules.MaxBattleTurns && !g.battleFinishEnded {
					t.Fatalf("第 %d 場沒有走完就離開戰場：回合 %d、狀態 %v", battles, g.battleTurn, g.State())
				}
				if g.saveErr != nil {
					t.Fatalf("第 %d 場戰後：%v", battles, g.saveErr)
				}
			default:
				t.Fatalf("第 %d 回合停在 %v（第 %d 場之後）", g.turn, g.State(), battles)
			}
		}
		t.Fatalf("玩不到條件，停在第 %d 回合 %v", g.turn, g.State())
	}
	play(func() bool { return battles >= 2 || g.turn >= 80 })
	if battles == 0 {
		t.Fatalf("到第 %d 回合一場都沒開", g.turn)
	}
	// 正常存進度，再正常取進度。
	g.Click(544, 125)
	g.Click(560, 165)
	wantState(t, g, StateSaveGame, "正常存進度")
	g.Click(300, g.res.SlotList.Save.TextTop)
	if err := g.SaveErr(); err != nil {
		t.Fatal(err)
	}
	want := g.slot.Fleets(g.res.Weights)
	fleetSaveEqual(t, want, g.fleets)
	g.RightClick(300, 122)
	wantState(t, g, StateCommandPhase, "存檔後返回")
	g.Click(544, 125)
	g.Click(560, 145)
	wantState(t, g, StateLoadGame, "正常取進度")
	g.Click(300, 122)
	wantState(t, g, StateLoadFaction, "存檔列後選陣營")
	g.Click(318, loadRowY)
	wantState(t, g, StateCommandPhase, "取進度完成")
	fleetSaveEqual(t, g.fleets, want)
	turn := g.turn
	play(func() bool { return g.turn >= turn+20 })
	t.Logf("讀回之後再玩到第 %d 回合，共 %d 場", g.turn, battles)
	if err := g.SaveSlot(t.TempDir(), 0); err != nil {
		t.Fatal(err)
	}
	fleetSaveEqual(t, g.slot.Fleets(g.res.Weights), g.fleets)
}

// 原版新遊戲依主選單寫 WORK+0C..14（`docs/internals/30`）。底稿量在選單之前，這幾格是零；
// 帝國這一列與原版傾印 `t260_post`（回合 0）相同。
func TestNewGameWritesSeatControlBytes(t *testing.T) {
	for _, c := range []struct {
		name string
		want [9]byte
		mk   func(g *Game)
	}{
		{"帝國", [9]byte{0, 0, 1, 0, 1, 0, 1, 0, 31}, func(g *Game) { g.startNewGame(0) }},
		{"同盟", [9]byte{1, 0, 1, 0, 1, 1, 0, 31, 0}, func(g *Game) { g.startNewGame(1) }},
		{"對戰", [9]byte{2, 0, 0, 0, 0, 0, 1, 0, 31}, func(g *Game) { g.startVersus() }},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newTestGame(t)
			g.versus = false
			c.mk(g)
			if got := g.workRecord()[0x0C:0x15]; !bytes.Equal(got, c.want[:]) {
				t.Errorf("WORK+0C..14 = % x，原版 % x", got, c.want)
			}
		})
	}
}

// 進場鏡頭要對著玩家自己的艦隊。同盟新遊戲曾因座位表是零而對到帝國那一支。
func TestNewGameBattleCameraFollowsPlayer(t *testing.T) {
	for _, faction := range []int{0, 1} {
		g := newTestGame(t)
		g.startNewGame(faction)
		g.fleets = nil
		g.monthSettlementPending = false
		region, entry := marchEntry(t, g)
		used := map[int]bool{}
		own := newGameFleet(t, g, faction, 0, entry.From, used)
		newGameFleet(t, g, 1-faction, 0, region, used)
		arriveFrom(t, g, own, region)
		g.startPlacedBattle(region)
		if g.state != StateBattle {
			t.Fatalf("陣營 %d：狀態 %v（saveErr=%v）", faction, g.state, g.saveErr)
		}
		f := g.fleets[fleetAt(t, g, own)]
		if want := g.battleViewAt(f.X, f.Y); g.battleView != want {
			t.Errorf("陣營 %d：鏡頭 %+v，玩家艦隊在 (%d,%d) 應該是 %+v", faction, g.battleView, f.X, f.Y, want)
		}
	}
}
