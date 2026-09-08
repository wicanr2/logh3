package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

const root = "../.."

// load 把玩家路徑要用的東西全部備好；缺原版或收據就 skip。
func load(t *testing.T) *Resources {
	t.Helper()
	orig := filepath.Join(root, "org_game/extracted")
	need := []string{
		filepath.Join(orig, "GRAPH.IMG"), filepath.Join(orig, "FACE.AMG"),
		filepath.Join(orig, "ADMIRAL.DTX"), filepath.Join(orig, "STRMAP.BIT"),
		filepath.Join(orig, "STRMAP.RGN"),
	}
	for _, p := range need {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", p)
		}
	}
	fnt, err := font.Load(filepath.Join(orig, "GRAPH.IMG"))
	if err != nil {
		t.Fatal(err)
	}
	faceData, err := os.ReadFile(filepath.Join(orig, "FACE.AMG"))
	if err != nil {
		t.Fatal(err)
	}
	faces, err := asset.OpenAMG(faceData)
	if err != nil {
		t.Fatal(err)
	}
	admiral, err := os.ReadFile(filepath.Join(orig, "ADMIRAL.DTX"))
	if err != nil {
		t.Fatal(err)
	}
	menu, err := screen.LoadMenuSpec(filepath.Join(root, "docs/internals/data/screen-main-menu.json"))
	if err != nil {
		t.Fatal(err)
	}
	cp, err := screen.LoadCopyProtSpec(filepath.Join(root, "docs/internals/data/screen-copy-protection.json"))
	if err != nil {
		t.Fatal(err)
	}
	mg, err := screen.LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-main-game.json"))
	if err != nil {
		t.Fatal(err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	bit, err := os.ReadFile(filepath.Join(orig, "STRMAP.BIT"))
	if err != nil {
		t.Fatal(err)
	}
	rgn, err := os.ReadFile(filepath.Join(orig, "STRMAP.RGN"))
	if err != nil {
		t.Fatal(err)
	}
	colours, err := regionColours(filepath.Join(root, "docs/internals/data/strmap-regions.json"))
	if err != nil {
		t.Fatal(err)
	}
	star, err := asset.BuildStarMap(bit, rgn, colours, 16, 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	so, err := screen.LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-ship-order.json"))
	if err != nil {
		t.Fatal(err)
	}
	bp, err := screen.LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-build-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	bg, err := screen.LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-budget-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	al, err := screen.LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-admiral-list.json"))
	if err != nil {
		t.Fatal(err)
	}
	regions, err := asset.OpenRGN(rgn, 16, 20)
	if err != nil {
		t.Fatal(err)
	}
	screens, maps, err := LoadCommandScreens(
		filepath.Join(root, "docs/internals/data"),
		filepath.Join(root, "docs/internals/data/strmap-regions.json"), bit, rgn)
	if err != nil {
		t.Fatal(err)
	}
	panel, err := screen.LoadRegionPanel(filepath.Join(root, "docs/internals/data/region-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	capital, err := screen.LoadCapitalPanel(filepath.Join(root, "docs/internals/data/capital-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	budget, err := screen.LoadBudgetPanel(filepath.Join(root, "docs/internals/data/budget-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := screen.LoadReportPanel(filepath.Join(root, "docs/internals/data/report-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	flagships, err := screen.LoadFlagships(filepath.Join(root, "docs/internals/data/flagships.json"))
	if err != nil {
		t.Fatal(err)
	}
	admPanel, err := screen.LoadAdmiralPanel(filepath.Join(root, "docs/internals/data/admiral-detail-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	buildPanel, err := screen.LoadReportPanel(filepath.Join(root, "docs/internals/data/build-plan-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	delivery, err := screen.LoadDeliveryPanel(filepath.Join(root, "docs/internals/data/delivery-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	production, err := screen.LoadProductionPanel(filepath.Join(root, "docs/internals/data/production-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	stellarFile, err := os.ReadFile(filepath.Join(orig, "STELLAR.DTA"))
	if err != nil {
		t.Fatal(err)
	}
	stellar, err := LoadRegionRecords(stellarFile,
		filepath.Join(root, "docs/internals/data/month-rollover-796-01.json"))
	if err != nil {
		t.Fatal(err)
	}
	pladat, err := os.ReadFile(filepath.Join(orig, "PLADAT.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	planets, err := rules.LoadPlanets(pladat)
	if err != nil {
		t.Fatal(err)
	}
	growth, err := rules.LoadGrowthTable(filepath.Join(root, "docs/internals/data/growth-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	work, err := rules.LoadWorkTemplate(filepath.Join(root, "docs/internals/data/work-796-01.json"))
	if err != nil {
		t.Fatal(err)
	}
	slotList, err := screen.LoadSlotList(filepath.Join(root, "docs/internals/data/save-slot-list.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saveIndex []rules.SaveIndexEntry
	if b, err := os.ReadFile(filepath.Join(orig, "SAVEDATA.IDX")); err == nil {
		if saveIndex, err = rules.ParseSaveIndex(b); err != nil {
			t.Fatal(err)
		}
	}
	names, err := loadAdmiralNames(filepath.Join("../..", "docs/internals/data/keireki.json"))
	if err != nil {
		t.Fatal(err)
	}
	starOrder, err := loadStarListOrder(filepath.Join("../..", "docs/internals/data/star-list-order.json"))
	if err != nil {
		t.Fatal(err)
	}
	return &Resources{Font: fnt, OrigDir: orig, Faces: faces, Cursor: cur, Menu: menu,
		AdmiralNames: names, StarListOrder: starOrder,
		CopyProt: cp, MainGame: mg, StarMap: star, Admiral: admiral, ShipOrder: so,
		StarBit: bit, StarRGN: rgn, RegionColours: colours,
		BuildPlan: bp, BudgetPlan: bg, AdmiralList: al, Screens: screens, StarMaps: maps,
		Regions: regions, RegionPanel: panel, CapitalPanel: capital,
		BudgetPanel: budget, ProductionPanel: production, DeliveryPanel: delivery,
		ReportPanel: report, BuildPlanPanel: buildPanel, AdmiralPanel: admPanel, Flagships: flagships,
		Stellar: stellar, Planets: planets, Growth: growth,
		SlotList: slotList, SaveIndex: saveIndex, WorkTemplate: work,
		StartMilitary: 50000, ShipPct: 25}
}

// reachCommandPhase 走正常玩家路徑到戰略地圖介面：
// 主選單 → 防拷問答 → 遊戲本體（月初報告）→ 右鍵兩下。
func reachCommandPhase(t *testing.T) *Game {
	t.Helper()
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	if g.State() != StateMainGame {
		t.Fatal("答對了卻沒進遊戲本體")
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	if g.State() != StateCommandPhase {
		t.Fatalf("右鍵兩下之後在 %v，不是戰略地圖介面", g.State())
	}
	// 收據 `main1e` 是用 `-mouse-at` 把滑鼠搬到 (600,300) 之後才擷取的
	// （滑鼠座標；畫面座標 (600,340)）。地圖介面的游標跟著滑鼠跑，
	// 所以要對得上就得搬到同一個地方。
	g.MoveMouse(600, 340)
	return g
}

// TestCommandPhaseNavigation 走命令階段的每一條路：命令列三格各拉出選單，
// 選單每一項各開一張畫面，每一張都要與原版收據逐位元組相同。
//
// **回程不比像素**：右鍵回到地圖介面之後，原版左下角那塊會顯示游標底下的
// 星系（`rcback` 收據），remake 還沒有那一層模型，所以只驗狀態。
func TestCommandPhaseNavigation(t *testing.T) {
	for _, c := range []struct {
		name       string
		menuX, itemX int
		itemY      int
		menuBatch  string
		batch      string
		state      State
	}{
		{"出擊命令", 352, 400, 125, "cmd1e", "sortie1e", StateStandbyAdmirals},
		{"情報收集", 352, 400, 145, "cmd1e", "info1e", StateInfo},
		{"徵稅計劃", 352, 400, 165, "cmd1e", "tax1e", StateTaxPlan},
		{"生產計劃", 352, 400, 185, "cmd1e", "prod1e", StateProductionPlan},
		{"星系資料", 448, 460, 125, "lstmenu1e", "star1e", StateStarList},
		{"艦隊資料", 448, 460, 145, "lstmenu1e", "lst1e", StateFleetList},
		{"提督資料", 448, 460, 165, "lstmenu1e", "lst2e", StateAdmiralListView},
		{"敵將資料", 448, 460, 185, "lstmenu1e", "lst3e", StateEnemyList},
		{"取進度", 544, 560, 145, "sysmenu1e", "load1e", StateLoadGame},
		{"存進度", 544, 560, 165, "sysmenu1e", "save1e", StateSaveGame},
		{"設定", 544, 560, 185, "sysmenu1e", "settings1e", StateSettings},
		{"結束", 544, 560, 205, "sysmenu1e", "quit1e", StateQuitConfirm},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := reachCommandPhase(t)
			matches(t, g, "main1e")

			g.Click(c.menuX, 125)
			matches(t, g, c.menuBatch)

			g.Click(c.itemX, c.itemY)
			if g.State() != c.state {
				t.Fatalf("點了「%s」之後在 %v", c.name, g.State())
			}
			matches(t, g, c.batch)

			g.RightClick(300, 240)
			if g.State() != StateCommandPhase {
				t.Fatalf("右鍵沒有回到地圖介面，在 %v", g.State())
			}
		})
	}
}

// TestExecuteAdvancesThreeDays 釘住「實行」那條規則（`docs/re/20`）。
func TestExecuteAdvancesThreeDays(t *testing.T) {
	g := reachCommandPhase(t)
	if y, m, d := g.Date(); y != 796 || m != 1 || d != 1 {
		t.Fatalf("開局是 %d年%d月%d日，預期 796年1月1日", y, m, d)
	}
	for i, want := range []struct{ month, day int }{
		{1, 4}, {1, 7}, {1, 10},
	} {
		g.Click(544, 125)
		g.Click(560, 125)
		if g.State() != StateCommandPhase {
			t.Fatalf("第 %d 次實行之後在 %v", i+1, g.State())
		}
		if _, m, d := g.Date(); m != want.month || d != want.day {
			t.Fatalf("第 %d 次實行之後是 %d月%d日，預期 %d月%d日",
				i+1, m, d, want.month, want.day)
		}
	}
	// 再按七次就跨月（一個月十回合）。
	for i := 0; i < 7; i++ {
		g.Click(544, 125)
		g.Click(560, 125)
	}
	if _, m, d := g.Date(); m != 2 || d != 1 {
		t.Fatalf("第十次實行之後是 %d月%d日，預期 2月1日", m, d)
	}
}

// TestExecuteTurnMatchesOracle 按一次「實行」之後**整張畫面**要與原版相同。
//
// 這一張逼出了「游標跟不跟著滑鼠」：地圖介面的規格記著量收據那一刻
// `-mouse-at` 搬過去的位置，按完實行回到地圖，原版的箭頭在剛剛點下去的
// 地方，remake 卻還畫在收據那一點——畫面上是兩隻箭頭，223 個像素。
//
// 三份收據是同一個狀態的三個步數（260M／320M／450M），拿來確認畫面在
// 那個區間內是穩定的，不是擷取到畫到一半。
func TestExecuteTurnMatchesOracle(t *testing.T) {
	for _, batch := range []string{"turn260", "turn320", "turn450"} {
		t.Run(batch, func(t *testing.T) {
			g := reachCommandPhase(t)
			g.Click(544, 125)
			g.Click(560, 125)
			if g.State() != StateCommandPhase {
				t.Fatalf("實行之後在 %v", g.State())
			}
			matches(t, g, batch)
		})
	}
}

// TestShipDeliveryMatchesOracle 月底的船艦接收報告。
//
// **只比對話框那一塊**（`layout-only` 不夠、`same-state` 不到）：整張畫面
// 對不上，因為那一刻的地圖已經是結算後的樣子——有兩個星域換了顏色
// （`docs/re/54`），而 remake 還沒有那條規則。對話框本身與原版
// 逐位元組相同，含年月、參謀的姓名軍階與十四個數字。
func TestShipDeliveryMatchesOracle(t *testing.T) {
	g := reachCommandPhase(t)
	for i := 0; i < rules.TurnsPerMonth; i++ {
		g.Click(544, 125) // 系統指令
		g.Click(560, 125) // 實行
	}
	if g.State() != StateShipDelivery {
		t.Fatalf("跨月之後在 %v，不是船艦接收報告", g.State())
	}
	matchesBox(t, g, "turnmonth2", 144, 80, 495, 399)
	// 右側面板的字被清掉、只留「戰略模式處理中」那一行，也對得上。
	matchesBox(t, g, "turnmonth2", 320, 40, 639, 79)

	g.Click(448, 370) // 確認鈕在 (430,360)-(465,379)
	if g.State() != StateCommandPhase {
		t.Fatalf("按了確認之後在 %v", g.State())
	}
}

// TestMenuClickOutsideCloses 點在選單項目外面會把選單關掉（實測四個方向的
// 界外點都回主畫面，`docs/internals/23`）。
func TestMenuClickOutsideCloses(t *testing.T) {
	for _, c := range []struct {
		name       string
		menuX      int
		x, y       int
	}{
		{"戰略命令右界外", 352, 456, 125},
		{"戰略命令下界外", 352, 400, 200},
		{"系統指令左界外", 544, 543, 145},
		{"系統指令右界外", 544, 592, 145},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := reachCommandPhase(t)
			g.Click(c.menuX, 125)
			if g.State() == StateCommandPhase {
				t.Fatal("選單沒開起來")
			}
			g.Click(c.x, c.y)
			if g.State() != StateCommandPhase {
				t.Fatalf("點界外之後在 %v，預期回地圖介面", g.State())
			}
		})
	}
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

// matches 把目前畫面與 oracle 收據逐位元組比。
// matchesBox 只比畫面上的一塊矩形。整張比不了的時候用它，並在呼叫端
// 寫清楚為什麼——**降級要留下理由**，不然下一個人會以為整張都對過。
func matchesBox(t *testing.T, g *Game, batch string, x0, y0, x1, y1 int) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "screen.bin"))
	if err != nil {
		t.Skipf("沒有收據 %s，跳過（tools/oracle.sh 可重建）", batch)
	}
	got := g.Frame().Pix
	bad, total := 0, 0
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			total++
			if want[y*640+x] != got[y*640+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Fatalf("%s (%d,%d)-(%d,%d)：與原版不同的像素 %d / %d",
			batch, x0, y0, x1, y1, bad, total)
	}
}

func matches(t *testing.T, g *Game, batch string) {
	t.Helper()
	p := filepath.Join(root, "workplace/oracle", batch, "screen.bin")
	want, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("沒有收據 %s，跳過（tools/oracle.sh 可重建）", batch)
	}
	got := g.Frame().Pix
	bad := 0
	for i := range want {
		if want[i] != got[i] {
			bad++
		}
	}
	if bad != 0 {
		t.Fatalf("%s：與原版不同的像素 %d / %d", batch, bad, len(want))
	}
}

// TestPlayerPathReachesGame 走完整條玩家路徑，每一步都要與原版收據
// 逐位元組相同：主選單反白 → 點擊 → 防拷問答 → 輸入答案 → 遊戲本體。
//
// **這是「正常玩家路徑」的收據**，不是 direct-entry：每個畫面都是
// 前一個畫面的輸入帶進去的。
func TestPlayerPathReachesGame(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	if g.State() != StateMainMenu {
		t.Fatal("開場不是主選單")
	}
	g.MoveMouse(300, 210)
	matches(t, g, "r0")

	g.Click(300, 210)
	if g.State() != StateCopyProtection {
		t.Fatal("點了主選單沒有進防拷問答")
	}
	matches(t, g, "cb1")

	if g.Answer() != 138 {
		t.Fatalf("答案是 %d，預期 138", g.Answer())
	}
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	if g.State() != StateMainGame {
		t.Fatal("答對了卻沒進遊戲本體")
	}
	matches(t, g, "key1")

	// 主畫面的熱區：點下去開出船艦訂造報告。
	g.Click(448, 330)
	if g.State() != StateShipOrder {
		t.Fatal("點了主畫面的熱區沒有開出船艦訂造報告")
	}
	matches(t, g, "ship1")

	// 訂造報告上的熱區：點下去開出次月份船艦建造計劃。
	g.Click(448, 310)
	if g.State() != StateBuildPlan {
		t.Fatal("點了訂造報告的熱區沒有開出建造計劃")
	}
	matches(t, g, "plan1")
}

// TestWrongAnswerStaysOnQuestion：答錯不會放行。
func TestWrongAnswerStaysOnQuestion(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '2', '3', '\r'} {
		g.Key(c)
	}
	if g.State() != StateCopyProtection {
		t.Fatal("答錯卻進了遊戲")
	}
	if g.Typed() != "" {
		t.Fatalf("答錯之後輸入沒清掉：%q", g.Typed())
	}
}

// TestClickOutsideMenuDoesNothing：點在選單以外不算數。
func TestClickOutsideMenuDoesNothing(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(10, 10)
	if g.State() != StateMainMenu {
		t.Fatal("點在選單外卻換了畫面")
	}
}

// TestBuildPlanArrowsChangePlan：建造計劃上的箭頭要真的改到預算。
//
// 右欄是減、左欄是加，一次 1%（`docs/re/11`）。開局七列加起來剛好
// 100%，所以左欄一開始按不動——那是原版行為，不是壞掉。
func TestBuildPlanArrowsChangePlan(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8'} {
		g.Key(c)
	}
	g.Key('\r')
	g.Click(448, 330)
	g.Click(448, 310)
	if g.State() != StateBuildPlan {
		t.Skip("走不到建造計劃（缺資源），跳過")
	}
	p := g.Plan()
	if p == nil {
		t.Fatal("進了建造計劃卻沒有預算分配")
	}
	if p.Pct[0] != 20 {
		t.Fatalf("開局第 0 列是 %d%%，原版是 20%%", p.Pct[0])
	}
	// 左欄：餘額 0% 時按不動。
	g.Click(216, 170)
	if p.Pct[0] != 20 {
		t.Fatalf("餘額 0%% 還加得上去：%d%%", p.Pct[0])
	}
	// 右欄：一次減 1%。
	g.Click(232, 170)
	if p.Pct[0] != 19 || p.RestPct() != 1 || p.RestAmount() != 188 {
		t.Fatalf("減一次後 %d%% / 餘 %d%% %d，原版是 19%% / 餘 1%% 188",
			p.Pct[0], p.RestPct(), p.RestAmount())
	}
	// 有餘額之後左欄就按得動了。
	g.Click(216, 170)
	if p.Pct[0] != 20 || p.RestPct() != 0 {
		t.Fatalf("加回來變成 %d%% / 餘 %d%%", p.Pct[0], p.RestPct())
	}
	// 第 6 列的箭頭在最後一格。
	g.Click(232, 280)
	if p.Pct[6] != 14 {
		t.Fatalf("點第 6 列的減，變成 %d%%，應該是 14%%", p.Pct[6])
	}
}

// TestBudgetPlanFromMainScreen：主畫面另一個熱區開出次月份預算計劃，
// 畫面要與原版收據逐位元組相同（`docs/re/12`）。
func TestBudgetPlanFromMainScreen(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8'} {
		g.Key(c)
	}
	g.Key('\r')
	if g.State() != StateMainGame {
		t.Fatal("答對了卻沒進遊戲本體")
	}
	g.Click(432, 300)
	if g.State() != StateBudgetPlan {
		t.Fatal("點了 (432,300) 沒有開出預算計劃")
	}
	matches(t, g, "budget1")
}

// TestAdmiralListFromBuildPlan：建造計劃左按鈕開出待命中提督資料，
// 畫面要與原版收據逐位元組相同（`docs/re/12`）。
func TestAdmiralListFromBuildPlan(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8'} {
		g.Key(c)
	}
	g.Key('\r')
	g.Click(448, 330)
	g.Click(448, 310)
	if g.State() != StateBuildPlan {
		t.Fatal("走不到建造計劃")
	}
	g.Click(464, 330)
	if g.State() != StateAdmiralList {
		t.Fatal("點了左按鈕沒有開出待命中提督資料")
	}
	matches(t, g, "admirals1")
}

// TestEndingScreenMatchesOriginalPlanes：結局畫面畫出來之後，
// plane 2 與 plane 3 要與原版收據逐位元組相同。
//
// 只比這兩個 plane，因為原版那一幕是好幾張圖疊出來的，
// plane 0／1 怎麼來的還沒解（`docs/internals/18`）。
func TestEndingScreenMatchesOriginalPlanes(t *testing.T) {
	src := filepath.Join(root, "org_game/extracted/ENDT30.VO5")
	shot := filepath.Join(root, "workplace/oracle/ending-empire/screen.bin")
	for _, p := range []string{src, shot} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過", p)
		}
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	img, err := asset.DecodeVO5(raw)
	if err != nil {
		t.Fatal(err)
	}
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.ShowEnding(img)
	if g.State() != StateEnding {
		t.Fatal("沒有切到結局畫面")
	}
	got := g.Frame().Pix
	want, err := os.ReadFile(shot)
	if err != nil {
		t.Fatal(err)
	}
	for _, plane := range []uint8{2, 3} {
		bad := 0
		for y := 0; y < img.H; y++ {
			for x := 0; x < img.W; x++ {
				i := (screen.EndingImageY+y)*gfx.Width + x
				if (got[i]>>plane)&1 != (want[i]>>plane)&1 {
					bad++
				}
			}
		}
		if bad != 0 {
			t.Errorf("plane %d：與原版差 %d 個位元", plane, bad)
		}
	}
}


// TestRightClickSelectsRegion 按在星圖上會換掉選中的星域。
// 每一筆的預期值都是實測：原版按下去之後左下角那塊寫出來的星系名
// （收據 `hovA`／`hovB`／`hovC`／`rcback`）。
func TestRightClickSelectsRegion(t *testing.T) {
	for _, c := range []struct {
		x, y  int
		want  int
		name  string
	}{
		{60, 100, 2, "亞爾提那"},
		{150, 160, 14, "費沙"},
		{250, 240, 19, "特利布拉"},
		{300, 240, 30, "艾流塞拉"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := reachCommandPhase(t)
			g.RightClick(c.x, c.y)
			if g.SelectedRegion() != c.want {
				t.Fatalf("(%d,%d) 選到星域 %d，預期 %d（%s）",
					c.x, c.y, g.SelectedRegion(), c.want, c.name)
			}
		})
	}
}

// TestRegionPanelOnPlayerPath 走正常玩家路徑到戰略地圖介面，右鍵選一個星域，
// 左下角那塊要與原版逐位元組相同（規格 `docs/spec/001`）。
//
// 只比面板那塊 320×80：這張畫面的游標還是規格裡量到的位置（遊戲目前不會
// 把游標跟著滑鼠搬，那是另一件事），所以整張比會差在游標上。
func TestRegionPanelOnPlayerPath(t *testing.T) {
	for _, c := range []struct {
		batch  string
		x, y   int
		region int
	}{
		{"hovA", 60, 100, 2},
		{"hovB", 150, 160, 14},
		{"hovC", 250, 240, 19},
		{"rcback", 300, 240, 30},
	} {
		t.Run(c.batch, func(t *testing.T) {
			oracle := filepath.Join("../..", "workplace/oracle/"+c.batch+"/screen.bin")
			want, err := os.ReadFile(oracle)
			if err != nil {
				t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", oracle)
			}
			g := reachCommandPhase(t)
			g.RightClick(c.x, c.y)
			if g.SelectedRegion() != c.region {
				t.Fatalf("選到星域 %d，預期 %d", g.SelectedRegion(), c.region)
			}
			fb := g.Frame()
			bad := 0
			for y := 340; y < 420; y++ {
				for x := 0; x < 320; x++ {
					if want[y*640+x] != fb.Pix[y*640+x] {
						bad++
					}
				}
			}
			if bad != 0 {
				t.Fatalf("%s：左下角那塊與原版不同的像素 %d / %d", c.batch, bad, 320*80)
			}
		})
	}
}

// TestMonthEndAdvancesMoneyAndDate 按十次「實行」跨過月底：日期要走到
// 796 年 2 月 1 日，錢要照月底的順序動（`docs/re/26`、`docs/re/33`、`docs/re/35`）：
//
//	原定軍事費 ＝ 上個月的軍事費 ＋ 這次收到的稅
//	軍事費     ＝ 原定軍事費 − 實際花掉的造船費
//
// 796/2 的經濟力目前算出來是 83305、原版是 83304，**差 1 還沒解**
// （`docs/re/33`），所以這裡驗關係，不驗絕對值。
func TestMonthEndAdvancesMoneyAndDate(t *testing.T) {
	g := reachCommandPhase(t)
	before := g.Military()

	for i := 0; i < 10; i++ {
		g.Click(544, 125) // 系統指令
		g.Click(560, 125) // 實行
	}
	y, m, d := g.Date()
	if y != 796 || m != 2 || d != 1 {
		t.Fatalf("十次實行之後是 %d 年 %d 月 %d 日，預期 796/2/1", y, m, d)
	}
	tr := g.Treasury()
	if tr.Budget <= before {
		t.Fatalf("原定軍事費 %d 沒有比結算前的 %d 多", tr.Budget, before)
	}
	if got, want := g.Military(), tr.Budget-tr.ShipPaid; got != want {
		t.Fatalf("軍事費 %d，預期 %d（原定 %d − 已支付 %d）",
			got, want, tr.Budget, tr.ShipPaid)
	}
	if g.TaxTotal() <= 0 {
		t.Fatalf("預估徵稅額 %d，應該是正的", g.TaxTotal())
	}
}

// TestBuildPlanPctReachesProduction：建造計劃上改的百分比要傳到生產計劃。
//
// 兩張畫面是不同月份的同一份分配（`docs/spec/004`）：建造計劃用月初報告
// 那一組建造費、生產計劃用次月份預估的，百分比共用一份。
func TestBuildPlanPctReachesProduction(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.Click(448, 330) // 船艦訂造報告
	g.Click(448, 310) // 建造計劃
	if g.State() != StateBuildPlan {
		t.Fatalf("沒進建造計劃，在 %v", g.State())
	}
	before := g.NextPlan()
	// 第 0 列的右箭頭按兩下：開局各列加起來剛好 100%，只減得動。
	g.Click(232, 170)
	g.Click(232, 170)
	after := g.NextPlan()
	if after.Pct[0] == before.Pct[0] {
		t.Fatalf("按了箭頭之後第 0 列還是 %d%%", after.Pct[0])
	}
	if after.Ships(0) == before.Ships(0) {
		t.Fatalf("百分比變了（%d%% → %d%%）但船艦數沒變（%d）",
			before.Pct[0], after.Pct[0], after.Ships(0))
	}
	if after.RestPct() <= before.RestPct() {
		t.Fatalf("減了預算，餘額百分比卻沒變多（%d%% → %d%%）",
			before.RestPct(), after.RestPct())
	}
}

// TestLoadSlotSetsState 走存檔那條路：載進第 10 格（S3「巴米利恩會戰」）
// 之後，日期與兩筆錢就是那個局面的，不是開局的（規格 `docs/spec/005`）。
func TestLoadSlotSetsState(t *testing.T) {
	res := load(t)
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	orig := filepath.Join("../..", "org_game/extracted")
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(orig, name))
		if err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", name)
		}
		return b
	}
	sl, err := rules.LoadSlot(rules.SlotFiles{
		Stellar:      read("STELLAR.DTA"),
		Planet:       read("PLANET.DTA"),
		Admiral:      read("ADMIRAL.DTA"),
		Fleet:        read("FLEET.DTA"),
		Ship:         read("SHIP.DTA"),
		Work:         read("WORK.DTA"),
		PlanetHeader: read("PLADAT.DAT"),
	})
	if err != nil {
		t.Fatal(err)
	}
	g.LoadSlot(sl)
	if y, m, d := g.Date(); y != 799 || m != 3 || d != 1 {
		t.Errorf("日期 %d/%d/%d，要 799/3/1", y, m, d)
	}
	if g.Military() != 160000 || g.TaxTotal() != 13131 {
		t.Errorf("軍事費 %d 總稅收 %d，要 160000／13131", g.Military(), g.TaxTotal())
	}
	if g.State() != StateCommandPhase {
		t.Errorf("狀態 %v，要命令階段", g.State())
	}
}

// TestLoadGamePlayerPath 走「取進度」那條正常玩家路徑（`docs/spec/007`）：
// 主選單第四列 → 防拷問答 → 存檔畫面 → 點第 9 格（諸神的黃昏）。
func TestLoadGamePlayerPath(t *testing.T) {
	res := load(t)
	orig := filepath.Join(root, "org_game/extracted")
	if _, err := os.Stat(filepath.Join(orig, "WORK.DT9")); err != nil {
		t.Skipf("缺存檔，跳過（不用代用品）")
	}
	res.OrigDir = orig
	graph, err := rules.LoadStarGraph(filepath.Join(root,
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	res.Graph = graph
	res.UnrestParams = rules.UnrestParams{
		Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
		Table: []int{170, 150, 130, 90, 60},
		Base:  []int{50, 75, 100, 150, 200},
	}
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, menuRowY(t, res, menuRowLoad))
	if g.State() != StateCopyProtection {
		t.Fatalf("點第四列之後在 %v，要防拷問答", g.State())
	}
	for _, c := range []byte{'1', '3', '8'} {
		g.Key(c)
	}
	g.Key('\r')
	if g.State() != StateLoadGame {
		t.Fatalf("答對之後在 %v，要存檔畫面", g.State())
	}
	// 第 9 格「諸神的黃昏」：收據量到的落點是 262，畫面座標是 302
	// （dosgolem 的滑鼠座標比畫面少 40，`docs/spec/007`）。
	if n := SlotAt(302); n != 9 {
		t.Fatalf("y=302 落在第 %d 格，要第 9 格", n)
	}
	g.Click(300, 302)
	if g.State() != StateCommandPhase {
		t.Fatalf("點下去之後在 %v，要命令階段", g.State())
	}
	if y, m, d := g.Date(); y != 798 || m != 11 || d != 19 {
		t.Errorf("日期 %d/%d/%d，要 798/11/19", y, m, d)
	}
	if g.Military() != 150000 || g.TaxTotal() != 7280 {
		t.Errorf("軍事費 %d 總稅收 %d，要 150000／7280", g.Military(), g.TaxTotal())
	}
}

// menuRowY 回主選單第 n 列的 y。
func menuRowY(t *testing.T, res *Resources, n int) int {
	t.Helper()
	for y := 0; y < 480; y++ {
		if screen.RowAt(res.Menu, y) == n {
			return y
		}
	}
	t.Fatalf("主選單找不到第 %d 列", n)
	return 0
}

// TestSaveSlotRoundTrip 走「載入 → 存回去」（`docs/spec/009`）：
// 從第 9 格進來、什麼都沒做就存到另一個目錄，六個檔要與原版相同。
func TestSaveSlotRoundTrip(t *testing.T) {
	res := load(t)
	orig := filepath.Join(root, "org_game/extracted")
	if _, err := os.Stat(filepath.Join(orig, "WORK.DT9")); err != nil {
		t.Skipf("缺存檔，跳過（不用代用品）")
	}
	res.OrigDir = orig
	graph, err := rules.LoadStarGraph(filepath.Join(root,
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	res.Graph = graph
	res.UnrestParams = rules.UnrestParams{
		Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
		Table: []int{170, 150, 130, 90, 60},
		Base:  []int{50, 75, 100, 150, 200},
	}
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, menuRowY(t, res, menuRowLoad))
	for _, c := range []byte{'1', '3', '8'} {
		g.Key(c)
	}
	g.Key('\r')
	g.Click(300, 302)
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 9); err != nil {
		t.Fatal(err)
	}
	for _, name := range rules.SlotFileNames(9) {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join(orig, name))
		if err != nil {
			t.Fatal(err)
		}
		bad := 0
		for i := range want {
			if i < len(got) && got[i] != want[i] {
				bad++
			}
		}
		// 未安定度載入時重算過，寫回去自然是重算後的值。
		limit := 0
		if name == "STELLAR.DT9" {
			limit = rules.RegionCount
		}
		if bad > limit {
			t.Errorf("%s：與原版不同的位元組 %d（允許 %d）", name, bad, limit)
		}
	}
}

// TestSaveNewGame 新遊戲也存得了（`docs/spec/009`）：開局走到戰略地圖介面
// 就存出去，六個檔的長度與版面要對，`WORK` 的三個欄位要是目前的狀態。
func TestSaveNewGame(t *testing.T) {
	g := reachCommandPhase(t)
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{
		"STELLAR.DT0": rules.RegionCount * rules.RegionRecord,
		"PLANET.DT0":  81 * rules.PlanetRecord,
		"ADMIRAL.DT0": 82 * rules.AdmiralRecord,
		"FLEET.DT0":   rules.FleetCount * rules.FleetRecord,
		"SHIP.DT0":    rules.ShipCount * rules.ShipRecord,
		"WORK.DT0":    rules.WorkRecord,
	}
	for name, n := range want {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if len(b) != n {
			t.Errorf("%s：%d bytes，要 %d", name, len(b), n)
		}
	}
	w, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
	if err != nil {
		t.Fatal(err)
	}
	turn := int(w[0x04]) | int(w[0x05])<<8
	mil := int(w[0x3E]) | int(w[0x3F])<<8 | int(w[0x40])<<16 | int(w[0x41])<<24
	if turn != g.Turn() || mil != g.Military() {
		t.Errorf("WORK 裡是回合 %d、軍事費 %d，遊戲裡是 %d／%d",
			turn, mil, g.Turn(), g.Military())
	}
}

// TestSaveGameClickWrites 走「存進度」那條路（`docs/spec/010`）：
// 命令階段 → 系統 → 存進度 → 點一列，就寫出那一格的六個檔。
//
// 原版在唯讀的原版目錄上會跳「磁碟片載入時有發生錯誤。」（收據 `save3e`）；
// remake 寫去自己的目錄，所以會成功。
func TestSaveGameClickWrites(t *testing.T) {
	res := load(t)
	dir := t.TempDir()
	res.SaveDir = dir
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	g.Click(544, 125) // 系統
	g.Click(560, 165) // 存進度
	if g.State() != StateSaveGame {
		t.Fatalf("點了存進度之後在 %v", g.State())
	}
	// 第 0 列：畫面 y ＝ 142（一列 20 px）。
	g.Click(300, 145)
	if err := g.SaveErr(); err != nil {
		t.Fatalf("存檔失敗：%v", err)
	}
	for _, name := range rules.SlotFileNames(0) {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("沒寫出 %s", name)
		}
	}
}

// loadAdmiralNames 讀 82 名提督姓名的 Big5 字碼。
func loadAdmiralNames(path string) ([][]uint16, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Records []struct {
			Index     int      `json:"index"`
			NameCodes []uint16 `json:"name_codes"`
		} `json:"records"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make([][]uint16, len(doc.Records))
	for _, r := range doc.Records {
		if r.Index >= 0 && r.Index < len(out) {
			out[r.Index] = r.NameCodes
		}
	}
	return out, nil
}

// loadStarListOrder 讀「星系資料」那張清單的列順序，兩個陣營各一份。
func loadStarListOrder(path string) ([][]int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Orders [][]int `json:"orders"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	return doc.Orders, nil
}
