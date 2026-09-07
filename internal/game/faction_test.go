package game

import "testing"

// TestAllianceStartMatchesOracle 主選單第 1 列是同盟（`docs/re/61`）。
//
// 同一條玩家路徑、只換點的那一列：防拷問答的畫面與帝國那張逐位元組相同
// （收據 `cbAlly` vs `cb1`，0 個像素），答案也一樣；月初報告則不同——
// 署名換成同盟的內政專任參謀邱吾權，錢也是同盟星域收的。
func TestAllianceStartMatchesOracle(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 230) // 主選單第 1 列「同盟」
	if g.State() != StateCopyProtection {
		t.Fatalf("點了同盟之後在 %v", g.State())
	}
	matches(t, g, "cbAlly")

	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	if g.State() != StateMainGame {
		t.Fatalf("答對了卻在 %v", g.State())
	}
	if g.faction != 1 {
		t.Fatalf("陣營是 %d，預期 1（同盟）", g.faction)
	}
	matches(t, g, "keyAlly")
}
