package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 右下角清單的四個分頁與捲動（`docs/spec/052`、`docs/re/319`）。每一張都從
// `fcmd-base`（第１艦隊剛編成、停在首都）接著點下去，與 `tools/fleetcmd_receipts.sh`
// 第 17 段的 `fcmd-tab-*` 收據整張逐位元組比。座標是畫面座標（滑鼠 y ＋ 40）。
//
// 分頁鈕：艦隊 (350,290)、行星 (415,290)、敵艦隊 (490,290)、提督 (560,290)；
// 列 k (450, 308＋20k)；▲ (617,308)、▼ (617,388)。

func reachListTab(t *testing.T, tab int) *Game {
	t.Helper()
	g := reachFleetDecided(t)
	x := []int{350, 415, 490, 560}[tab]
	g.Click(x, 290)
	if g.State() != StateCommandPhase || g.listTab() != tab {
		t.Fatalf("點了分頁 %d 之後在 %v、分頁 %d", tab, g.State(), g.listTab())
	}
	return g
}

// 行星分頁：瓦爾哈拉四顆行星，防衛力都滿（色 7）；點一列開行星情報視窗
// （帝國軍星系／行星、經濟力與成長率、防衛力與回復率、七個艦種的補充部隊數、
// 飛彈與戰鬥機、「確認」），確認或右鍵關掉、那一列反白還原。欄 17 的捲動
// 熱區點了沒有事（行星最多五顆）。
func TestListTabPlanetMatchesOracle(t *testing.T) {
	g := reachListTab(t, listTabPlanet)
	matches(t, g, "fcmd-tab-planet")
	if p := g.listPage(listTabPlanet); p.Count != 4 || p.Scroll != 0 {
		t.Fatalf("行星分頁 %+v，原版 ds:CFFC 4、ds:95D6 0", p)
	}
	t.Run("scroll", func(t *testing.T) {
		g := reachListTab(t, listTabPlanet)
		g.Click(617, 290)
		matches(t, g, "fcmd-tab-planet-scroll")
	})
	g.Click(450, 308)
	if g.tabs.planet != 0 {
		t.Fatalf("點了第 0 列之後行星情報視窗開的是 %d", g.tabs.planet)
	}
	matches(t, g, "fcmd-tab-planet-row")
	t.Run("rclick", func(t *testing.T) {
		g := reachListTab(t, listTabPlanet)
		g.Click(450, 308)
		g.RightClick(300, 240)
		if g.listOverlayOn() {
			t.Fatal("右鍵之後行星情報視窗還開著")
		}
		matches(t, g, "fcmd-tab-planet-rclick")
	})
	g.Click(462, 329) // 「確認」
	if g.listOverlayOn() {
		t.Fatal("點了確認之後行星情報視窗還開著")
	}
	matches(t, g, "fcmd-tab-planet-ok")
}

// 敵艦隊分頁：新遊戲第一回合首都沒有敵艦隊，整頁空；點空列沒有事。
func TestListTabEnemyMatchesOracle(t *testing.T) {
	g := reachListTab(t, listTabEnemy)
	matches(t, g, "fcmd-tab-enemy")
	g.Click(450, 308)
	matches(t, g, "fcmd-tab-enemy-row")
}

// 提督分頁：首都閒置的提督 35 位（同陣營、沒有職務、沒帶艦隊、在那一格），
// 姓名與軍階；▲▼ 一下一頁（原版按住 0.3M 步剛好五列），到底夾在最後五列；
// 點一列開提督詳細資料（經歷／中止／出擊）、中止或右鍵回來；出擊 → 艦隊編成
// → 確認 → 艦隊組成 → 決定建出第２艦隊，回來提督分頁還選著、那一位不在了。
func TestListTabAdmiralMatchesOracle(t *testing.T) {
	g := reachListTab(t, listTabAdmiral)
	matches(t, g, "fcmd-tab-admiral")
	if p := g.listPage(listTabAdmiral); p.Count != 35 || p.Scroll != 0 {
		t.Fatalf("提督分頁 %+v，原版 ds:95D8 35、ds:95D9 0", p)
	}
	t.Run("scroll", func(t *testing.T) {
		g := reachListTab(t, listTabAdmiral)
		holdListSteps(g, 617, 388, 5)
		matches(t, g, "fcmd-tab-admiral-down")
		if p := g.listPage(listTabAdmiral); p.Scroll != 5 {
			t.Fatalf("▼ 一下之後位移 %d，原版 5", p.Scroll)
		}
		t.Run("down2", func(t *testing.T) {
			g := reachListTab(t, listTabAdmiral)
			holdListSteps(g, 617, 388, 5)
			holdListSteps(g, 617, 388, 5)
			matches(t, g, "fcmd-tab-admiral-down2")
		})
		holdListSteps(g, 617, 308, 5)
		matches(t, g, "fcmd-tab-admiral-up")
	})
	t.Run("end", func(t *testing.T) {
		// 原版按住3M步到30；089改用固定速度的等價保持，不宣稱同時間。
		g := reachListTab(t, listTabAdmiral)
		holdListSteps(g, 617, 388, 40)
		if p := g.listPage(listTabAdmiral); p.Scroll != 30 {
			t.Fatalf("捲到底位移 %d，原版 30", p.Scroll)
		}
		matches(t, g, "fcmd-tab-admiral-end")
	})
	g.Click(450, 308)
	if g.tabs.detail != 1 || !g.tabs.detailSortie {
		t.Fatalf("點了第 0 列之後詳細資料 %d（出擊 %v），要 1（吉爾菲艾斯）", g.tabs.detail, g.tabs.detailSortie)
	}
	matches(t, g, "fcmd-tab-admiral-row")
	t.Run("cancel", func(t *testing.T) {
		g := reachListTab(t, listTabAdmiral)
		g.Click(450, 308)
		g.Click(448, 320) // 「中止」
		if g.listOverlayOn() {
			t.Fatal("中止之後詳細資料還開著")
		}
		matches(t, g, "fcmd-tab-admiral-cancel")
	})
	t.Run("rclick", func(t *testing.T) {
		g := reachListTab(t, listTabAdmiral)
		g.Click(450, 308)
		g.RightClick(300, 240)
		matches(t, g, "fcmd-tab-admiral-rclick")
	})
	g.Click(496, 320) // 「出擊」→ 艦隊編成
	if g.State() != StateFleetOrg || g.detailAdmiral != 1 {
		t.Fatalf("點了出擊之後在 %v、提督 %d", g.State(), g.detailAdmiral)
	}
	matches(t, g, "fcmd-tab-admiral-sortie")
	t.Run("rclick-is-confirm", func(t *testing.T) {
		// 出擊那條路的艦隊編成按右鍵 ＝ 確認（`07F9:00D5`），直接進艦隊組成。
		g := reachListTab(t, listTabAdmiral)
		g.Click(450, 308)
		g.Click(496, 320)
		g.RightClick(300, 240)
		if g.State() != StateFleetComposition {
			t.Fatalf("艦隊編成按右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-tab-sortie-rclick")
	})
	g.Click(480, 350) // 「確認」→ 艦隊組成
	g.Click(510, 350) // 「決定」
	if g.State() != StateCommandPhase || len(g.fleets) != 2 || g.listTab() != listTabAdmiral {
		t.Fatalf("決定之後在 %v、%d 支、分頁 %d", g.State(), len(g.fleets), g.listTab())
	}
	matches(t, g, "fcmd-tab-fleet2")
	if p := g.listPage(listTabAdmiral); p.Count != 33 {
		t.Fatalf("出擊之後提督分頁 %d 位，原版 33（提督與副官各少一位）", p.Count)
	}
	g.Click(350, 290) // 艦隊分頁：兩支
	matches(t, g, "fcmd-tab-fleet2-list")
}

// 出擊命令那條路的艦隊編成按右鍵也是確認（冷啟動同 `fcmd-fok`，320M 那一下換成右鍵）。
func TestFleetOrgRightClickIsConfirm(t *testing.T) {
	g := reachFleetOrgFrom(t, reachCommandPhaseOnMapWith(t, loadWithRules(t)))
	g.RightClick(300, 240)
	if g.State() != StateFleetComposition {
		t.Fatalf("艦隊編成按右鍵之後在 %v", g.State())
	}
	matches(t, g, "fcmd-forg-rclick")
	wantStock(t, g, [7]int{75000, 36000, 108000, 10000, 27000, 15000, 5000})
}

// 六支艦隊：從提督分頁再出擊四位（列 0 → 出擊 → 確認 → 決定），艦隊分頁出現 ▲▼；
// ▼ 捲到第 1–5 列（位移只能到 1），▲ 回來；捲著的時候點最後一列開的是第６艦隊，
// 結束回來清單留在捲過的位置。
func TestListTabFleetScrollMatchesOracle(t *testing.T) {
	reach := func(t *testing.T) *Game {
		t.Helper()
		g := reachListTab(t, listTabAdmiral)
		for i := 0; i < 5; i++ {
			g.Click(450, 308)
			g.Click(496, 320)
			g.Click(480, 350)
			g.Click(510, 350)
			if g.State() != StateCommandPhase {
				t.Fatalf("第 %d 次出擊之後在 %v", i+2, g.State())
			}
		}
		if len(g.fleets) != 6 {
			t.Fatalf("六次出擊之後 %d 支", len(g.fleets))
		}
		return g
	}
	g := reach(t)
	matches(t, g, "fcmd-tab-fleet6")
	if p := g.listPage(listTabAdmiral); p.Count != 25 {
		t.Fatalf("六支之後提督分頁 %d 位，原版 25", p.Count)
	}
	g.Click(350, 290)
	matches(t, g, "fcmd-tab-fleet6-list")
	g.Click(617, 388)
	if p := g.listPage(listTabFleet); p.Count != 6 || p.Scroll != 1 {
		t.Fatalf("▼ 之後艦隊分頁 %+v，原版 6／1", p)
	}
	matches(t, g, "fcmd-tab-fleet6-down")
	t.Run("up", func(t *testing.T) {
		g := reach(t)
		g.Click(350, 290)
		g.Click(617, 388)
		g.Click(617, 308)
		matches(t, g, "fcmd-tab-fleet6-up")
	})
	g.Click(450, 388) // 捲過之後的最後一列 ＝ 第６艦隊
	if g.State() != StateFleetWindow || g.FleetCommandFleet() != 5 {
		t.Fatalf("點了最後一列之後在 %v、艦隊 %d", g.State(), g.FleetCommandFleet())
	}
	matches(t, g, "fcmd-tab-fleet6-row")
	g.Click(490, 390) // 「結束」
	if g.State() != StateCommandPhase {
		t.Fatalf("結束之後在 %v", g.State())
	}
	matches(t, g, "fcmd-tab-fleet6-close")
	if p := g.listPage(listTabFleet); p.Scroll != 1 {
		t.Fatalf("關窗之後位移 %d，原版 ds:9593 仍是 1", p.Scroll)
	}
}

// 分頁是每個陣營各一份（`ds:8F29[座位]`）。
func TestListTabPerFaction(t *testing.T) {
	g := reachListTab(t, listTabAdmiral)
	g.faction = 1
	if g.listTab() != listTabFleet {
		t.Fatalf("換到同盟之後分頁 %d，要 0", g.listTab())
	}
	g.faction = 0
	if g.listTab() != listTabAdmiral {
		t.Fatalf("換回帝國之後分頁 %d，要 3", g.listTab())
	}
	_ = rules.ListPageRows
}
