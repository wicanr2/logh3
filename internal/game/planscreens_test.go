package game

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

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

// 五條正常玩家路徑：月初預估不能改經濟力，訂造報告退場才收稅（研究 351）。
func TestBudgetSettlementMatchesOriginal(t *testing.T) {
	for _, tc := range []struct {
		name string
		x, y int
	}{
		{"unchanged", 0, 0}, {"tax-up", 471, 289}, {"tax-down", 487, 289},
		{"pct-up", 471, 329}, {"pct-down", 487, 329},
	} {
		t.Run(tc.name, func(t *testing.T) {
			batch := "planarrow-budget-" + tc.name
			mem, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch+"-post", "memory.bin"))
			if err != nil {
				t.Skipf("缺原版完整記憶體：%v", err)
			}
			if len(mem) != 1048576 {
				t.Fatalf("收據長度 %d", len(mem))
			}
			g, err := New(loadWithRules(t))
			if err != nil {
				t.Fatal(err)
			}
			g.Click(300, 210)
			for _, c := range []byte{'1', '3', '8', '\r'} {
				g.Key(c)
			}
			if !g.monthSettlementPending || g.military[0] != 50000 {
				t.Fatal("月初報告已提前結算")
			}
			pre := append([]byte(nil), g.res.Planets.Data...)
			g.Click(432, 340) // 公開畫面座標；舊熱區使用原版mouse y=300。
			if tc.name == "unchanged" {
				matches(t, g, "planarrow-budget-base")
			} else {
				g.Click(tc.x, tc.y)
				matches(t, g, batch)
			}
			if !bytes.Equal(pre, g.res.Planets.Data) {
				t.Fatal("預算操作提早扣行星經濟力")
			}
			if err := g.SaveSlot(t.TempDir(), 0); err == nil {
				t.Fatal("待結算狀態不應存成無法恢復的局面")
			}
			g.Click(463, 369)
			g.RightClick(300, 240)
			if g.State() != StateShipOrder || !bytes.Equal(pre, g.res.Planets.Data) {
				t.Fatal("會計報告退出已提前扣稅")
			}
			g.RightClick(300, 240)
			if g.State() != StateCommandPhase || g.monthSettlementPending {
				t.Fatal("訂造報告關閉未完成結算")
			}
			matches(t, g, batch+"-post")
			w, err := rules.ReadWork(mem[0x5DA44 : 0x5DA44+rules.WorkRecord])
			if err != nil {
				t.Fatal(err)
			}
			next := g.nextBudget()
			for _, c := range []struct {
				name      string
				got, want int
			}{
				{"軍事費", g.military[0], w.Military[0]}, {"下期稅額", g.taxTotal[0], w.TaxTotal[0]},
				{"原定軍事費", g.treasury.Budget, w.BudgetStart[0]}, {"實付", g.treasury.ShipPaid, w.ShipPaid[0]},
				{"下期造船預算", next.ShipCost(), w.ShipBudget[0]},
				{"稅率快照", g.rate[0], w.Rate[0]}, {"造船比例", g.shipPct[0], w.ShipPct[0]},
			} {
				if c.got != c.want {
					t.Errorf("%s：%d／原版 %d", c.name, c.got, c.want)
				}
			}
			if g.building[0] != w.Building[0] {
				t.Errorf("建造艘數 %v／原版 %v", g.building[0], w.Building[0])
			}
			for _, c := range []struct {
				name      string
				got, want []byte
			}{
				{"星域", g.res.Stellar, mem[0x5F2CC : 0x5F2CC+rules.RegionCount*rules.RegionRecord]},
				{"行星", g.res.Planets.Data[rules.PlanetHeader:], mem[0x53132 : 0x53132+81*rules.PlanetRecord]},
			} {
				bad := 0
				for i, b := range c.got {
					if b != c.want[i] {
						if bad < 8 {
							t.Logf("%s +%X：%02X／原版 %02X", c.name, i, b, c.want[i])
						}
						bad++
					}
				}
				if bad != 0 {
					t.Errorf("%s整表差 %d bytes；規格 065 不允許排除欄位", c.name, bad)
				}
			}
			for i := 0; i < rules.RegionCount; i++ {
				got := g.regionRecord(i)
				want := mem[0x5F2CC+i*rules.RegionRecord : 0x5F2CC+(i+1)*rules.RegionRecord]
				for _, r := range [][2]int{{0x1A, 0x20}, {0x26, 0x28}} {
					if !bytes.Equal(got[r[0]:r[1]], want[r[0]:r[1]]) {
						t.Errorf("星域 %d 財政 +%X：% X／原版 % X", i, r[0], got[r[0]:r[1]], want[r[0]:r[1]])
					}
				}
			}
			for i := 0; i < 81; i++ {
				got := g.res.Planets.Record(i)[0x1A:0x1E]
				want := mem[0x53132+i*rules.PlanetRecord+0x1A : 0x53132+i*rules.PlanetRecord+0x1E]
				if !bytes.Equal(got, want) {
					t.Errorf("行星 %d 經濟力：% X／原版 % X", i, got, want)
				}
			}
			if err := os.WriteFile(filepath.Join(root, "workplace/oracle", batch+"-post", "remake-stellar.bin"), g.res.Stellar, 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "workplace/oracle", batch+"-post", "remake-planet.bin"), g.res.Planets.Data[rules.PlanetHeader:], 0644); err != nil {
				t.Fatal(err)
			}
			post := append([]byte(nil), g.res.Planets.Data...)
			money := g.military
			g.RightClick(300, 240)
			if money != g.military || !bytes.Equal(post, g.res.Planets.Data) {
				t.Fatal("重複右鍵又收稅")
			}
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			planet, err := os.ReadFile(filepath.Join(dir, "PLANET.DT0"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(planet, post[rules.PlanetHeader:]) {
				t.Fatal("存檔行星不是結算後狀態")
			}
			g.res.SaveDir = dir
			g.Click(544, 125)
			g.Click(560, 145)
			if g.State() != StateLoadGame {
				t.Fatal("正常入口未開存檔清單")
			}
			g.Click(300, 122) // 取進度第 0 格。
			wantState(t, g, StateLoadFaction, "結算存檔選陣營")
			g.Click(318, 208+20*g.faction)
			if g.State() != StateCommandPhase || g.monthSettlementPending || money != g.military || !bytes.Equal(post, g.res.Planets.Data) {
				t.Fatal("載入重開待結算或改變資金")
			}
		})
	}
}
