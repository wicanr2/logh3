package game

import "testing"

// TestFleetOrgMatchesOracle 提督詳細資料按「出擊」開出艦隊編成
// （收據 `fleet1e`／`fleet2e`，`docs/spec/013`）。
//
// 兩張收據點的是不同提督：萊因哈特的副官是修特萊、吉爾菲艾斯的是錦茲
// （記錄 `+0x07`）。艦隊能力十項逐項取大，副官贏的那幾格印成色 5
// ——兩張的墨色分布不同，是這條規則最直接的證據（`docs/re/66`）。
func TestFleetOrgMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		y     int
		batch string
	}{
		{85, "fleet1e"},  // 第 0 列 萊因哈特，副官修特萊
		{105, "fleet2e"}, // 第 1 列 吉爾菲艾斯，副官錦茲
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachCommandPhase(t)
			g.Click(352, 125) // 戰略命令
			g.Click(400, 125) // 出擊命令
			g.Click(400, c.y)
			if g.State() != StateAdmiralDetail {
				t.Fatalf("點了 y=%d 之後在 %v", c.y, g.State())
			}
			g.Click(488, 330) // 「出擊」
			if g.State() != StateFleetOrg {
				t.Fatalf("點了出擊之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}
