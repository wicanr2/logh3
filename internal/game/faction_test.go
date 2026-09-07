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

	// 右鍵兩下關掉報告，滑鼠搬到與收據同一個地方（`main1eAlly` 的
	// `-mouse-at 600,300`，畫面座標 (600,340)）。
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	if g.State() != StateCommandPhase {
		t.Fatalf("右鍵兩下之後在 %v，不是戰略地圖介面", g.State())
	}
	g.MoveMouse(600, 340)
	matches(t, g, "main1eAlly")
}

// TestAllianceBuildPlanMatchesOracle 同盟的船艦訂造報告與次月份建造計劃。
//
// 七列的單價**兩個陣營不一樣**（`docs/re/62`）：畫面上的金額與船艦數
// 把每一列的單價各夾成唯一的整數，同盟是 9／14／5／23／3／7／24。
func TestAllianceBuildPlanMatchesOracle(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 230)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.Click(448, 330)
	if g.State() != StateShipOrder {
		t.Fatalf("點了主畫面的熱區之後在 %v", g.State())
	}
	matches(t, g, "shipAlly")
	g.Click(448, 310)
	if g.State() != StateBuildPlan {
		t.Fatalf("點了訂造報告的熱區之後在 %v", g.State())
	}
	matches(t, g, "planAlly")
}

// reachCommandPhaseAlly 走同盟那條玩家路徑到戰略地圖介面。
func reachCommandPhaseAlly(t *testing.T) *Game {
	t.Helper()
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 230)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	if g.State() != StateCommandPhase {
		t.Fatalf("右鍵兩下之後在 %v", g.State())
	}
	g.MoveMouse(600, 340)
	return g
}

// TestAllianceCommandPhaseScreens 同盟的命令階段子畫面。
//
// 這幾張的規格是照帝國的收據量的，底板右側那幾行也在裡面；
// `drawBoardPanels` 會蓋過去，但**蓋不掉比較長的舊字**——這幾張就是
// 在驗那件事有沒有發生。
func TestAllianceCommandPhaseScreens(t *testing.T) {
	for _, c := range []struct {
		name         string
		menuX, itemX int
		itemY        int
		menuBatch    string
		batch        string
		state        State
	}{
		{"情報收集", 352, 400, 145, "cmd1eAlly", "info1eAlly", StateInfo},
		{"提督資料", 448, 460, 165, "lstmenu1eAlly", "lst2eAlly", StateAdmiralListView},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := reachCommandPhaseAlly(t)
			matches(t, g, "main1eAlly")

			g.Click(c.menuX, 125)
			matches(t, g, c.menuBatch)

			g.Click(c.itemX, c.itemY)
			if g.State() != c.state {
				t.Fatalf("點了「%s」之後在 %v", c.name, g.State())
			}
			matches(t, g, c.batch)
		})
	}
}
