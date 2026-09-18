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
	if g.State() != StateStandbyAdmirals {
		t.Fatalf("點參謀之後在 %v，該開待命中提督清單", g.State())
	}
	matches(t, g, "fcmd-staff-list")

	g.Click(450, 92) // 清單第一列（滑鼠 52 ＋ 40）
	if g.State() != StateAdmiralDetail {
		t.Fatalf("點第一列之後在 %v，該開提督詳細資料", g.State())
	}
	// **`fcmd-staff-detail` 那張還不比**：remake 的提督詳細資料自帶整張背景
	// （戰略地圖介面），而原版這條路是**疊在待命清單上面**，另外第三顆鈕的字
	// 也還是「出擊」不是「決定」。疊層與按鈕文字是 #62 剩下的部分，
	// 收據 `fcmd-staff-detail` 已經錄好在等（差 20,017 個像素，52 塊，
	// 最大一塊 88×222 就是被蓋掉的清單區）。
	picked := g.detailAdmiral

	g.Click(497, 330) // 決定（滑鼠 290 ＋ 40）
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
	if g.State() != StateStandbyAdmirals {
		t.Fatalf("按中止之後在 %v，原版回清單（不是關窗）", g.State())
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction); got != before {
		t.Errorf("按中止之後情報參謀變成 %d，原版留在 %d", got, before)
	}
	matches(t, g, "fcmd-staff-cancel")
}
