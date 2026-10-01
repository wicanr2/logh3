package game

import "testing"

// 徵稅計劃／生產計劃的「結束」與預算計劃的兩組箭頭（#64、`docs/re/343`，
// 收據 `fcmd-plan-*`）。座標是畫面座標（收據旗標的滑鼠 y ＋ 40）：
// 稅率檔位 ＋(471,289) －(487,289)、建造費％ ＋(471,329)、結束 (463,369)；
// 生產計劃的結束在 (511,369)。

func reachTaxPlan(t *testing.T) *Game {
	t.Helper()
	g := reachFleetDecided(t)
	g.Click(352, 125) // 戰略命令
	g.Click(400, 165) // 徵稅計劃
	if g.State() != StateTaxPlan {
		t.Fatalf("開徵稅計劃之後在 %v", g.State())
	}
	return g
}

// 「結束」關掉疊圖回命令階段，**不寫回任何東西**。
func TestTaxPlanEndClosesOverlay(t *testing.T) {
	g := reachTaxPlan(t)
	matches(t, g, "fcmd-plan-tax")
	blend, pct := g.blend, g.shipPct
	g.Click(463, 369)
	if g.State() != StateCommandPhase {
		t.Fatalf("按了結束之後在 %v", g.State())
	}
	if g.blend != blend || g.shipPct != pct {
		t.Errorf("結束改了稅率檔位／造船％（%v→%v、%v→%v）；原版純粹關窗",
			blend, g.blend, pct, g.shipPct)
	}
	matches(t, g, "fcmd-plan-tax-end")
}

// 生產計劃的「結束」是同一條規則，鈕在另一個位置。
func TestProductionPlanEndClosesOverlay(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(352, 125)
	g.Click(400, 185) // 生產計劃
	if g.State() != StateProductionPlan {
		t.Fatalf("開生產計劃之後在 %v", g.State())
	}
	matches(t, g, "fcmd-plan-prod")
	g.Click(511, 369)
	if g.State() != StateCommandPhase {
		t.Fatalf("按了結束之後在 %v", g.State())
	}
	matches(t, g, "fcmd-plan-prod-end")
}

// 稅率檔位箭頭：短按一次一階，clamp 在 0–4；改完同一張畫面就重算。
func TestTaxPlanArrows(t *testing.T) {
	g := reachTaxPlan(t)
	if g.blend[g.faction] != 2 {
		t.Fatalf("開局的稅率檔位是 %d，收據是 2（普通）", g.blend[g.faction])
	}
	g.Click(471, 289) // ＋
	if g.blend[g.faction] != 3 {
		t.Fatalf("加一下之後是 %d，收據 fcmd-plan-tax-up 的 ds:8FD4 ＝ 03", g.blend[g.faction])
	}
	matches(t, g, "fcmd-plan-tax-up")

	g = reachTaxPlan(t)
	g.Click(487, 289) // －
	if g.blend[g.faction] != 1 {
		t.Fatalf("減一下之後是 %d，收據 fcmd-plan-tax-down 的 ds:8FD4 ＝ 01", g.blend[g.faction])
	}
	matches(t, g, "fcmd-plan-tax-down")

	// clamp：連按到底不會溢出。
	for i := 0; i < 8; i++ {
		g.Click(471, 289)
	}
	if g.blend[g.faction] != 4 {
		t.Errorf("一直加之後是 %d，原版 clamp 在 4（緊急）", g.blend[g.faction])
	}
	for i := 0; i < 8; i++ {
		g.Click(487, 289)
	}
	if g.blend[g.faction] != 0 {
		t.Errorf("一直減之後是 %d，原版 clamp 在 0（微稅）", g.blend[g.faction])
	}
}

// 建造費％箭頭：短按 +1，只連動建造費、不動稅收。
func TestTaxPlanShipPctArrow(t *testing.T) {
	g := reachTaxPlan(t)
	before := g.shipPct[g.faction]
	tax := g.taxTotal
	g.Click(471, 329) // ＋
	if g.shipPct[g.faction] != before+1 {
		t.Fatalf("加一下之後是 %d%%，收據 fcmd-plan-pct-up 的 ds:8F92 ＝ 26（開局 25）",
			g.shipPct[g.faction])
	}
	if g.taxTotal != tax {
		t.Errorf("建造費％動到了稅收（%v→%v）；原版兩顆箭頭各自獨立重算", tax, g.taxTotal)
	}
	matches(t, g, "fcmd-plan-pct-up")
}

// 四顆箭頭及呈現狀態生命週期都以原版整張收據核對（規格 003、研究 350）。
func TestPlanArrowRenderingMatchesOriginal(t *testing.T) {
	for _, tc := range []struct {
		batch string
		x, y  int
	}{
		{"tax-up", 471, 289}, {"tax-down", 487, 289},
		{"pct-up", 471, 329}, {"pct-down", 487, 329},
	} {
		t.Run(tc.batch, func(t *testing.T) {
			g := reachTaxPlan(t)
			g.Click(tc.x, tc.y)
			matches(t, g, "planarrow-"+tc.batch)
		})
	}
	t.Run("懸停", func(t *testing.T) {
		g := reachTaxPlan(t)
		g.MoveMouse(471, 289)
		matches(t, g, "planarrow-hover")
	})
	t.Run("移開仍保留", func(t *testing.T) {
		g := reachTaxPlan(t)
		g.Click(471, 289)
		g.MoveMouse(600, 380)
		matches(t, g, "planarrow-moved")
	})
	t.Run("兩顆各自保留", func(t *testing.T) {
		g := reachTaxPlan(t)
		g.Click(471, 289)
		g.Click(487, 289)
		matches(t, g, "planarrow-updown")
	})
	t.Run("關窗重開", func(t *testing.T) {
		g := reachTaxPlan(t)
		g.Click(471, 289)
		g.Click(463, 369)
		g.Click(352, 125)
		g.Click(400, 165)
		matches(t, g, "planarrow-reopen")
	})
	t.Run("上界仍轉色", func(t *testing.T) {
		g := reachTaxPlan(t)
		// 與原版明示注入相同：只改檔位，不預先跑重算。
		g.blend[g.faction] = 4
		g.Click(471, 289)
		matches(t, g, "planarrow-clamp")
	})
}
