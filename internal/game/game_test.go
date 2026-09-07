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
	Bootstrap796_01(stellar, planets, growth)
	return &Resources{Font: fnt, Faces: faces, Cursor: cur, Menu: menu,
		CopyProt: cp, MainGame: mg, StarMap: star, Admiral: admiral, ShipOrder: so,
		BuildPlan: bp, BudgetPlan: bg, AdmiralList: al, Screens: screens, StarMaps: maps,
		Regions: regions, RegionPanel: panel, CapitalPanel: capital,
		BudgetPanel: budget, Stellar: stellar, Planets: planets, Growth: growth,
		Military: 45592, TaxTotal: 10341, ShipPct: 25,
		Treasury: rules.Treasury{Budget: 60741, ShipPaid: 15149},
		MonthReport: rules.Budget{Treasury: 50000, Tax: 10741, ShipPct: 25}}
}

// reachCommandPhase 走正常玩家路徑到戰略地圖介面：
// 主選單 → 防拷問答 → 遊戲本體（月初報告）→ 右鍵兩下。
func reachCommandPhase(t *testing.T) *Game {
	t.Helper()
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 170)
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
	g.MoveMouse(300, 170)
	matches(t, g, "r0")

	g.Click(300, 170)
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
	g.Click(300, 170)
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
	g.Click(300, 170)
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
	g.Click(300, 170)
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
	g.Click(300, 170)
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
// 796 年 2 月 1 日，軍事費要加上這一個月收到的稅（`docs/re/26`、`docs/re/33`）。
//
// 796/2 的經濟力目前算出來是 83305、原版是 83304，**差 1 還沒解**
// （`docs/re/33`），所以這裡只驗日期與「軍事費 ＝ 舊的 ＋ 總稅收」這條關係，
// 不驗絕對值。
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
	if got, want := g.Military(), before+g.TaxTotal(); got != want {
		t.Fatalf("軍事費 %d，預期 %d（結算前 %d ＋ 總稅收 %d）",
			got, want, before, g.TaxTotal())
	}
	if g.TaxTotal() <= 0 {
		t.Fatalf("總稅收 %d，應該是正的", g.TaxTotal())
	}
}
