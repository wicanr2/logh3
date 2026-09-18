package game

import (
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
