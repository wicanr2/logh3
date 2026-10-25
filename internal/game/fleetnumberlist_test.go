package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 艦隊編號清單與結束確認（#65，`tools/fleetcmd_receipts.sh` 第 29 段）。
// 座標是畫面座標（收據旗標的滑鼠 y ＋ 40）：清單列 k 在 (368, 90＋20k)，
// 艦隊編號鈕 (368,350)、確認 (480,350)、決定 (510,350)；
// 結束確認的中止 (383,263)／決定 (432,263)。

// reachFleetNumberList 走到艦隊編成再按「艦隊編號」（收據 `fcmd-fnum`）。
func reachFleetNumberList(t *testing.T) *Game {
	t.Helper()
	g := reachFleetOrgFrom(t, reachCommandPhaseOnMapWith(t, loadWithRules(t)))
	g.Click(368, 350)
	if g.State() != StateFleetNumber {
		t.Fatalf("按了艦隊編號之後在 %v", g.State())
	}
	return g
}

// 點第 3 列 → 清單關掉、標題變第３艦隊；確認 → 決定，那一支進的是**槽 2**。
func TestFleetNumberListPicksSlot(t *testing.T) {
	g := reachFleetNumberList(t)
	matches(t, g, "fcmd-fnum")

	g.Click(368, 130) // 列 2 ＝ 第03艦隊
	if g.State() != StateFleetOrg {
		t.Fatalf("點了列之後在 %v", g.State())
	}
	if g.orgSlot != 2 {
		t.Fatalf("挑到的槽是 %d，預期 2", g.orgSlot)
	}
	matches(t, g, "fcmd-fnum-row")

	g.Click(480, 350) // 確認 → 艦隊組成
	g.Click(510, 350) // 決定
	if g.State() != StateCommandPhase {
		t.Fatalf("決定之後在 %v", g.State())
	}
	if len(g.fleets) != 1 {
		t.Fatalf("建出 %d 支", len(g.fleets))
	}
	if f := g.fleets[0]; f.Index != rules.FleetTableIndex(g.faction, 2) || f.Number != 3 {
		t.Fatalf("建出來的在槽 %d、編號 %d；原版是槽 2、第３艦隊（收據 fcmd-fnum-decide 的艦隊表）",
			f.Index, f.Number)
	}
	if slot := int(g.admiralRecord(g.fleets[0].Admiral)[9]) & 0x7f; slot != 2 {
		t.Fatalf("提督指向槽%d，選取的是槽2", slot)
	}
	matches(t, g, "fcmd-fnum-decide")
}

// 已經有第１艦隊時：列 0 畫成暗色（ink 8），點它只關清單、不換編號。
func TestFleetNumberListUsedRow(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(352, 125) // 戰略命令
	g.Click(400, 125) // 出擊
	g.Click(400, 85)
	g.Click(488, 330)
	if g.State() != StateFleetOrg {
		t.Skipf("走不到艦隊編成（現在 %v），跳過", g.State())
	}
	g.Click(368, 350)
	if g.State() != StateFleetNumber {
		t.Fatalf("按了艦隊編號之後在 %v", g.State())
	}
	if !g.fleetSlotUsed(0) || g.fleetSlotUsed(1) {
		t.Fatalf("槽 0 在用 %v、槽 1 在用 %v；預期 true／false", g.fleetSlotUsed(0), g.fleetSlotUsed(1))
	}
	matches(t, g, "fcmd-fnum2")

	g.Click(368, 90) // 列 0 ＝ 已經在用的第01艦隊
	if g.State() != StateFleetOrg {
		t.Fatalf("點了已用的列之後在 %v", g.State())
	}
	if g.orgSlot != -1 {
		t.Errorf("點已用的列把槽換成 %d 了；原版標題仍是第２艦隊", g.orgSlot)
	}
	matches(t, g, "fcmd-fnum2-used")
}

// 系統指令 → 結束 →「您確定要結束這個遊戲嗎？」：中止回地圖、決定結束程式。
func TestQuitConfirmButtons(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(544, 125) // 系統指令
	g.Click(560, 205) // 結束
	if g.State() != StateQuitConfirm {
		t.Fatalf("點了結束之後在 %v", g.State())
	}
	matches(t, g, "fcmd-quit")

	g.Click(300, 240) // 框內別處：不動
	if g.State() != StateQuitConfirm {
		t.Fatalf("點框內別處之後在 %v", g.State())
	}
	g.Click(383, 263) // 中止
	if g.State() != StateCommandPhase || g.Quit() {
		t.Fatalf("中止之後在 %v、quit %v", g.State(), g.Quit())
	}
	matches(t, g, "fcmd-quit-cancel")

	g.Click(544, 125)
	g.Click(560, 205)
	g.Click(432, 263) // 決定 → 原版走 DOS 的 AH=4C
	if !g.Quit() {
		t.Error("按了決定但沒有要結束")
	}
}

// 新艦隊進哪一格：走原版 `07F9:03AC`（提督 `+0x09` 低七位那一槽空著就沿用），
// **不是**「下一個空槽」。收據 `fcmd-sortie-adm3-done`：第１艦隊已經在的情況下
// 讓羅嚴塔爾（提督 3）出擊，建出來的在槽 3（第４艦隊），槽 1／2 還空著。
func TestSortieUsesAdmiralSlot(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(352, 125) // 戰略命令
	g.Click(400, 125) // 出擊
	g.Click(400, 125) // 清單第 3 列（羅嚴塔爾）
	g.Click(488, 330) // 出擊
	if g.State() != StateFleetOrg || g.detailAdmiral != 3 {
		t.Fatalf("在 %v、看的是提督 %d；預期艦隊編成、提督 3（羅嚴塔爾）", g.State(), g.detailAdmiral)
	}
	if k := g.orgFleetSlot(); k != 3 {
		t.Fatalf("配到槽 %d，原版是槽 3（第４艦隊）", k)
	}
	matches(t, g, "fcmd-sortie-adm3")

	g.Click(480, 350) // 確認
	g.Click(510, 350) // 決定
	if g.State() != StateCommandPhase {
		t.Fatalf("決定之後在 %v", g.State())
	}
	var got *rules.Fleet
	for i := range g.fleets {
		if g.fleets[i].Admiral == 3 {
			got = &g.fleets[i]
		}
	}
	if got == nil {
		t.Fatal("羅嚴塔爾那一支沒建出來")
	}
	if got.Index != rules.FleetTableIndex(g.faction, 3) || got.Number != 4 {
		t.Fatalf("建出來的在槽 %d、編號 %d；原版是槽 3、第４艦隊", got.Index, got.Number)
	}
	if g.fleetSlotUsed(1) || g.fleetSlotUsed(2) {
		t.Error("槽 1／2 被占掉了；原版那兩格還空著")
	}
	matches(t, g, "fcmd-sortie-adm3-done")
}
