package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 參謀任命（#62、規格 064、`docs/re/348`）：情報收集的「參謀」鈕開待命中提督清單
// → 選一列開詳細資料 → 「決定」寫專任參謀表、回情報收集換成新的姓名與頭像。
//
// 收據 `fcmd-staff-list`／`-detail`／`-decide`／`-cancel`
// （`tools/fleetcmd_receipts.sh` 第 35 段）。
func TestStaffPickMatchesOracle(t *testing.T) {
	g := reachInfoScreen(t)
	before := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction)

	g.Click(543, 388) // 參謀（畫面 y ＝ 滑鼠 348 ＋ 40）
if !g.staff.listOn {
		t.Fatal("點參謀之後清單沒開")
	}
	matches(t, g, "fcmd-staff-list")

	g.Click(450, 92) // 清單第一列（滑鼠 52 ＋ 40）
if !g.staff.detailOn {
		t.Fatal("點第一列之後詳細資料沒開")
	}
	matches(t, g, "fcmd-staff-detail")
	picked := g.staff.detail

	g.Click(497, 330) // 決定（滑鼠 290 ＋ 40）
if g.staff.on() {
		t.Fatal("按決定之後疊層該整條收掉")
	}
	if g.State() != StateInfo {
		t.Fatalf("按決定之後在 %v，該回情報收集", g.State())
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction); got != picked {
		t.Errorf("情報參謀 ＝ 提督 %d，該是選中的 %d（原本 %d）", got, picked, before)
	}
	matches(t, g, "fcmd-staff-decide")
}

// 「中止」回清單、不是關窗，而且什麼都不寫（收據 `fcmd-staff-cancel`：
// `ds:8F4D` 留在原值）。
func TestStaffPickCancelMatchesOracle(t *testing.T) {
	g := reachInfoScreen(t)
	before := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction)

	g.Click(543, 388)
	g.Click(450, 92)
	g.Click(455, 330) // 中止
if !g.staff.listOn || g.staff.detailOn {
		t.Fatalf("按中止之後 listOn=%v detailOn=%v，原版回清單（不是關窗）", g.staff.listOn, g.staff.detailOn)
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction); got != before {
		t.Errorf("按中止之後情報參謀變成 %d，原版留在 %d", got, before)
	}
	matches(t, g, "fcmd-staff-cancel")
}

// 徵稅計劃那一顆參謀（`090A:0248`）：同一套清單與詳細視窗，換的是**內政**
// 專任參謀。收據 `fcmd-taxstaff-list`／`-decide`（`docs/re/348` §六）。
func TestTaxPlanStaffPickMatchesOracle(t *testing.T) {
	g := reachTaxPlan(t)
	beforeIntel := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction)

	g.Click(415, 370) // 參謀（畫面 y ＝ 滑鼠 330 ＋ 40）
	if !g.staff.listOn {
		t.Fatal("點參謀之後清單沒開")
	}
	matches(t, g, "fcmd-taxstaff-list")

	g.Click(450, 92) // 第一列
	if !g.staff.detailOn {
		t.Fatal("點第一列之後詳細資料沒開")
	}
	picked := g.staff.detail

	g.Click(497, 330) // 決定
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostDomestic, g.faction); got != picked {
		t.Errorf("內政參謀 ＝ 提督 %d，該是選中的 %d", got, picked)
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction); got != beforeIntel {
		t.Errorf("情報參謀被改成 %d，原版只動內政那一格（留在 %d）", got, beforeIntel)
	}
	matches(t, g, "fcmd-taxstaff-decide")
}

// reachMonthlyStart 走開局那一條：主選單 → 防拷答 138 → 月初報告（主畫面）。
func reachMonthlyStart(t *testing.T, faction int) *Game {
	t.Helper()
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210+20*faction) // 主選單第 0 列帝國、第 1 列同盟
	for _, c := range []byte("138\r") {
		g.Key(c)
	}
	if g.State() != StateMainGame || g.faction != faction {
		t.Fatalf("答對防拷之後在 %v、陣營 %d", g.State(), g.faction)
	}
	return g
}

// savedStaff 存進第 0 格之後，從 `WORK.DT0` 讀回某個幕僚職的提督索引。
func savedStaff(t *testing.T, g *Game, post int) int {
	t.Helper()
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
	if err != nil {
		t.Fatal(err)
	}
	return rules.StaffAdmiral(raw, post, g.faction)
}

// pixDiff 數兩張畫面在矩形 [x0,x1)×[y0,y1) 裡不同的像素。
func pixDiff(a, b []uint8, x0, y0, x1, y1 int) int {
	n := 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if a[y*640+x] != b[y*640+x] {
				n++
			}
		}
	}
	return n
}

// 月初報告那條路的「次月份預算計劃」：參謀鈕與徵稅計劃共用熱區 `ds:14A8`
// （`docs/re/343`），換內政專任參謀。這一張走 `Frame` 自己的分支，曾經漏畫
// 參謀疊層：清單開了卻看不見，點擊全被隱形的清單吃掉（使用者回報）。
func TestBudgetPlanStaffPick(t *testing.T) {
	for faction := 0; faction < 2; faction++ {
		t.Run([]string{"帝國", "同盟"}[faction], func(t *testing.T) { budgetPlanStaffPick(t, faction) })
	}
}

func budgetPlanStaffPick(t *testing.T, faction int) {
	g := reachMonthlyStart(t, faction)
	g.Click(432, 340)
	if g.State() != StateBudgetPlan {
		t.Fatalf("沒開出預算計劃，在 %v", g.State())
	}
	before := append([]uint8(nil), g.Frame().Pix...)
	beforeIntel := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction)

	g.Click(415, 370) // 參謀
	if !g.staff.listOn || g.staff.post != rules.PostDomestic {
		t.Fatalf("點參謀之後 listOn=%v post=%d，該開內政參謀的清單", g.staff.listOn, g.staff.post)
	}
	// 清單要真的畫出來：右半前兩列（y 82–121）一定與沒開清單時不同。
	if n := pixDiff(before, g.Frame().Pix, 336, 82, 608, 122); n == 0 {
		t.Fatal("清單開了但畫面右半沒變：參謀疊層沒畫")
	}

	g.Click(450, 92) // 第一列
	if !g.staff.detailOn {
		t.Fatal("點第一列之後詳細資料沒開")
	}
	picked := g.staff.detail
	withList := append([]uint8(nil), g.Frame().Pix...)
	if n := pixDiff(withList, before, 160, 100, 480, 300); n == 0 {
		t.Fatal("詳細資料開了但畫面沒變")
	}

	g.Click(497, 330) // 決定
	if g.staff.on() || g.State() != StateBudgetPlan {
		t.Fatalf("按決定之後 staff=%v state=%v，該收掉疊層、留在預算計劃", g.staff.on(), g.State())
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostDomestic, g.faction); got != picked {
		t.Errorf("內政參謀 ＝ 提督 %d，該是選中的 %d", got, picked)
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction); got != beforeIntel {
		t.Errorf("情報參謀被改成 %d（原本 %d）", got, beforeIntel)
	}
	// 回到預算計劃後「結束」照常可用。
	g.Click(463, 369)
	if g.State() != g.monthlyAccountState() {
		t.Errorf("換完參謀按結束之後在 %v", g.State())
	}
	toCommandPhase(t, g)
	if got := savedStaff(t, g, rules.PostDomestic); got != picked {
		t.Errorf("存檔裡的內政參謀是 %d，該是 %d", got, picked)
	}
}

// toCommandPhase 右鍵逐層關掉月初那幾張報告，直到戰略地圖介面（月初結算在這裡做）。
func toCommandPhase(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 4 && g.State() != StateCommandPhase; i++ {
		g.RightClick(300, 240)
	}
	if g.State() != StateCommandPhase {
		t.Fatalf("右鍵關報告之後停在 %v", g.State())
	}
}

// 指令選單的生產計劃：與建造計劃共用熱區清單 `ds:26E2`（七列箭頭、參謀、結束，
// `docs/re/343` §一）。箭頭動兩張共用的百分比；參謀換軍事專任參謀
// （strong inference，規格 064）。
func TestProductionPlanArrowsAndStaff(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(352, 125)
	g.Click(400, 185) // 生產計劃
	if g.State() != StateProductionPlan {
		t.Fatalf("開生產計劃之後在 %v", g.State())
	}
	pct := g.nextPlan().Pct
	g.Click(232, 210) // 第 0 列「減」（熱區 dec-0，滑鼠 y 160–179）
	if got := g.nextPlan().Pct[0]; got >= pct[0] {
		t.Fatalf("按了第 0 列減之後是 %d%%（原本 %d%%）", got, pct[0])
	}
	g.Click(216, 210) // 第 0 列「加」：剛減出餘額，加得回去
	if got := g.nextPlan().Pct[0]; got != pct[0] {
		t.Fatalf("減一下再加一下是 %d%%，該回到 %d%%", got, pct[0])
	}
	if g.State() != StateProductionPlan {
		t.Fatalf("按箭頭之後跑到 %v", g.State())
	}

	beforeIntel := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction)
	beforeDomestic := rules.StaffAdmiral(g.workRecord(), rules.PostDomestic, g.faction)
	g.Click(464, 370) // 參謀（熱區 admiral-list，滑鼠 (448,320)–(479,339)）
	if !g.staff.listOn || g.staff.post != rules.PostMilitary {
		t.Fatalf("點參謀之後 listOn=%v post=%d，該開軍事參謀的清單", g.staff.listOn, g.staff.post)
	}
	g.Click(450, 92)
	if !g.staff.detailOn {
		t.Fatal("點第一列之後詳細資料沒開")
	}
	picked := g.staff.detail
	g.Click(497, 330) // 決定
	if g.staff.on() || g.State() != StateProductionPlan {
		t.Fatalf("按決定之後 staff=%v state=%v", g.staff.on(), g.State())
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostMilitary, g.faction); got != picked {
		t.Errorf("軍事參謀 ＝ 提督 %d，該是選中的 %d", got, picked)
	}
	if rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction) != beforeIntel ||
		rules.StaffAdmiral(g.workRecord(), rules.PostDomestic, g.faction) != beforeDomestic {
		t.Error("換軍事參謀動到了情報或內政那一格")
	}
	if got, want := g.nextPlan().StaffOps, rules.AdmiralOps(g.res.Admiral, picked); got != want {
		t.Errorf("生產計劃的單價用營運 %d，新參謀是 %d", got, want)
	}
	if got := savedStaff(t, g, rules.PostMilitary); got != picked {
		t.Errorf("存檔裡的軍事參謀是 %d，該是 %d", got, picked)
	}
	g.Click(511, 369) // 結束
	if g.State() != StateCommandPhase {
		t.Errorf("按結束之後在 %v", g.State())
	}
}

// 月初那條路的建造計劃：參謀鈕開「待命中提督資料」（`screen-admiral-list`），
// 點一列疊上詳細資料；中止回清單，決定換軍事專任參謀並回建造計劃。
func TestBuildPlanStaffPick(t *testing.T) {
	for faction := 0; faction < 2; faction++ {
		t.Run([]string{"帝國", "同盟"}[faction], func(t *testing.T) { buildPlanStaffPick(t, faction) })
	}
}

func buildPlanStaffPick(t *testing.T, faction int) {
	g := reachMonthlyStart(t, faction)
	g.Click(448, 370) // 船艦訂造報告
	g.Click(448, 350) // 建造計劃
	if g.State() != StateBuildPlan {
		t.Fatalf("沒進建造計劃，在 %v", g.State())
	}
	before := rules.StaffAdmiral(g.workRecord(), rules.PostMilitary, g.faction)
	g.Click(464, 370) // 參謀
	if g.State() != StateAdmiralList {
		t.Fatalf("點參謀之後在 %v", g.State())
	}
	list := append([]uint8(nil), g.Frame().Pix...)

	g.Click(450, 92) // 第一列
	if !g.staff.detailOn || g.staff.post != rules.PostMilitary {
		t.Fatalf("點第一列之後 detailOn=%v post=%d", g.staff.detailOn, g.staff.post)
	}
	if n := pixDiff(list, g.Frame().Pix, 160, 100, 480, 300); n == 0 {
		t.Fatal("詳細資料開了但畫面沒變：疊層沒畫")
	}
	g.Click(455, 330) // 中止
	if g.staff.on() || g.State() != StateAdmiralList {
		t.Fatalf("按中止之後 staff=%v state=%v，該回清單", g.staff.on(), g.State())
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostMilitary, g.faction); got != before {
		t.Fatalf("按中止之後軍事參謀變成 %d（原本 %d）", got, before)
	}

	g.Click(450, 112) // 第二列
	picked := g.staff.detail
	if !g.staff.detailOn || picked < 0 {
		t.Fatal("點第二列之後詳細資料沒開")
	}
	g.Click(497, 330) // 決定
	if g.staff.on() || g.State() != StateBuildPlan {
		t.Fatalf("按決定之後 staff=%v state=%v，該回建造計劃", g.staff.on(), g.State())
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostMilitary, g.faction); got != picked {
		t.Errorf("軍事參謀 ＝ 提督 %d，該是選中的 %d", got, picked)
	}
	if g.plan == nil || g.plan.StaffOps != rules.AdmiralOps(g.res.Admiral, picked) {
		t.Errorf("建造計劃的單價沒改用新參謀的營運")
	}
	// 換完再開清單，新參謀不在待命名單裡（規格 064 規則 1：不含現任專任參謀）。
	g.Click(464, 370)
	g.setListRows(g.res.AdmiralList)
	for _, i := range g.listIndex {
		if i == picked {
			t.Errorf("新任軍事參謀 %d 仍出現在待命中提督清單", picked)
		}
	}
	toCommandPhase(t, g)
	if got := savedStaff(t, g, rules.PostMilitary); got != picked {
		t.Errorf("存檔裡的軍事參謀是 %d，該是 %d", got, picked)
	}
}
