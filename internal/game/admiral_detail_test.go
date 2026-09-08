package game

import "testing"

// TestAdmiralDetailMatchesOracle 從「出擊命令」的待命中提督清單點第一列，
// 開出那位提督的詳細資料（收據 `adm1e`，`docs/re/65`）。
func TestAdmiralDetailMatchesOracle(t *testing.T) {
	g := reachCommandPhase(t)
	g.Click(352, 125) // 戰略命令
	g.Click(400, 125) // 出擊命令
	if g.State() != StateStandbyAdmirals {
		t.Fatalf("點了出擊命令之後在 %v", g.State())
	}
	g.Click(400, 85) // 清單第 0 列
	if g.State() != StateAdmiralDetail {
		t.Fatalf("點了第一列之後在 %v", g.State())
	}
	matches(t, g, "adm1e")
}
