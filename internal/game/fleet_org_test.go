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

// reachFleetOrg 走玩家路徑走到萊因哈特那張艦隊編成。
func reachFleetOrg(t *testing.T) *Game {
	t.Helper()
	g := reachCommandPhase(t)
	g.Click(352, 125) // 戰略命令
	g.Click(400, 125) // 出擊命令
	g.Click(400, 85)  // 第 0 列 萊因哈特
	g.Click(488, 330) // 「出擊」
	if g.State() != StateFleetOrg {
		t.Fatalf("走到艦隊編成之前就跑到 %v", g.State())
	}
	return g
}

// TestFleetOrgHotspotsAreNarrowerThanTheBoxes 點在畫出來的框內、熱區外
// 一個像素，原版一個位元組都不動（收據 `fnum2`／`fadj2` 與 `fleet1e`
// 雜湊相同，`docs/re/68`）。
//
// 這兩個是**反例**：正例只說得出「這裡點得到」，反例才定得住邊界。
func TestFleetOrgHotspotsAreNarrowerThanTheBoxes(t *testing.T) {
	for _, c := range []struct {
		x, y  int
		batch string
		what  string
	}{
		{400, 350, "fnum2", "艦隊編號框的右邊緣"},
		{448, 350, "fadj2", "副官框的右邊緣"},
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(c.x, c.y)
			if g.State() != StateFleetOrg {
				t.Fatalf("點 %s (%d,%d) 之後跑到 %v", c.what, c.x, c.y, g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetNumberListMatchesOracle 「艦隊編號」開出的十六列清單
// （收據 `fnum`／`fnum3`，`docs/re/67`）。滑鼠停哪一列，那一列整條反白。
func TestFleetNumberListMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		x     int
		batch string
	}{
		{368, "fnum"},  // 游標壓在「艦」上
		{338, "fnum3"}, // 游標壓在「第」上
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(c.x, 350)
			if g.State() != StateFleetNumber {
				t.Fatalf("點了艦隊編號之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetAdjutantMenuMatchesOracle 「副官」開出的「變更／卸任」兩列
// 小選單（收據 `fadj`／`fadj3`）。
func TestFleetAdjutantMenuMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		x     int
		batch string
	}{
		{432, "fadj"},
		{417, "fadj3"},
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(c.x, 350)
			if g.State() != StateFleetAdjutant {
				t.Fatalf("點了副官之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetCompositionMatchesOracle 艦隊編成按「確認」開出的**艦隊組成**
// （收據 `fok`，`docs/re/67`、`docs/re/69`）。
//
// 七列的「待命部隊」是首都星域的艦艇艘數 ÷ 一單位幾艘，「數量」印的就是
// 一單位幾艘；格盤是 9 × 9 的 `SPRITE.SPR` 圖塊。
func TestFleetCompositionMatchesOracle(t *testing.T) {
	g := reachFleetOrg(t)
	g.Click(480, 350) // 「確認」
	if g.State() != StateFleetComposition {
		t.Fatalf("點了確認之後在 %v", g.State())
	}
	matches(t, g, "fok")
}

// TestFleetCompositionRowSelect 點七個艦種列的哪一列，選取條就跳到哪一列
// （收據 `row3`：點第 2 列巡航艦）。
func TestFleetCompositionRowSelect(t *testing.T) {
	g := reachFleetOrg(t)
	g.Click(480, 350) // 「確認」
	g.Click(360, 210) // 第 2 列 巡航艦
	if g.State() != StateFleetComposition {
		t.Fatalf("點了艦種列之後在 %v", g.State())
	}
	matches(t, g, "row3")
}

// TestFormationListMatchesOracle 艦隊組成按「陣形」開出的八列清單：
// 圓錐／紡錘／半月／Ｕ字／Ｍ字／Ｔ字／分散／密集（`docs/re/70`）。
//
// 三張收據點的是同一顆按鈕的不同位置——游標一次蓋得住兩列，八列要三份
// 收據才讀得齊，而三張的反白條都在同一列（滑鼠都落在那一格裡）。
func TestFormationListMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		x, y  int
		batch string
	}{
		{510, 310, "form1"},
		{497, 300, "form2"},
		{497, 318, "form3"},
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(480, 350) // 「確認」→ 艦隊組成
			g.Click(c.x, c.y) // 「陣形」
			if g.State() != StateFormation {
				t.Fatalf("點了陣形之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetDecideMatchesOracle 艦隊組成按「決定」：艦隊建成，回戰略地圖，
// 右下角清單的「艦隊」分頁多出一列（收據 `dec1`，`docs/re/70`）。
//
// 左下角那塊**不會被擦掉**——艦隊那幾張畫的是自己的首都，回到地圖之後
// 原版就讓它留在那裡。
func TestFleetDecideMatchesOracle(t *testing.T) {
	g := reachFleetOrg(t)
	g.Click(480, 350) // 「確認」→ 艦隊組成
	g.Click(510, 350) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	matches(t, g, "dec1")
}
