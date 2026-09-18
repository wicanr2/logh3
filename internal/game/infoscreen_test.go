package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/screen"
)

// 情報收集畫面的三個分頁與「確認」（#63、`docs/re/342`，收據 `fcmd-info-*`）。
// 座標是畫面座標（收據旗標的滑鼠 y ＋ 40）：治安 (360,168)、勢力 (424,168)、
// 敵艦 (488,168)、參謀 (543,388)、確認 (591,388)。

func reachInfoScreen(t *testing.T) *Game {
	t.Helper()
	g := reachFleetDecided(t)
	var err error
	if g.res.InfoPanel, err = screen.LoadInfoPanel(
		filepath.Join(root, "docs/internals/data/info-panel.json")); err != nil {
		t.Fatal(err)
	}
	g.Click(352, 125) // 戰略命令
	g.Click(400, 145) // 情報收集
	if g.State() != StateInfo {
		t.Fatalf("開情報收集之後在 %v", g.State())
	}
	return g
}

func TestInfoScreenTabsMatchOracle(t *testing.T) {
	g := reachInfoScreen(t)
	if g.infoScreenTab() != infoTabSecurity {
		t.Fatalf("開出來停在第 %d 頁，原版是治安", g.infoScreenTab())
	}
	matches(t, g, "fcmd-info-open")

	g.Click(424, 168) // 勢力
	if g.infoScreenTab() != infoTabPower {
		t.Fatalf("點勢力之後在第 %d 頁", g.infoScreenTab())
	}
	matches(t, g, "fcmd-info-power")

}

// 敵艦那一頁：文字、核取方塊與地圖著色。門檻走 `0A02:0672`＝距離表 `>>10`
// （`docs/re/342` 附錄 A ＋ `docs/re/347`），兩條路的等級照附錄 A 的謂詞。
func TestInfoScreenEnemyTabMatchesOracle(t *testing.T) {
	g := reachInfoScreen(t)
	g.Click(488, 168)
	if g.infoScreenTab() != infoTabEnemy {
		t.Fatalf("點敵艦之後在第 %d 頁", g.infoScreenTab())
	}
	matches(t, g, "fcmd-info-enemy")
}

// 「確認」關窗回戰略地圖；分頁狀態留著，重開還停在同一頁。
func TestInfoScreenConfirmKeepsTab(t *testing.T) {
	g := reachInfoScreen(t)
	g.Click(591, 388)
	if g.State() != StateCommandPhase {
		t.Fatalf("按了確認之後在 %v", g.State())
	}
	matches(t, g, "fcmd-info-confirm")

	g = reachInfoScreen(t)
	g.Click(488, 168) // 敵艦
	g.Click(591, 388) // 確認
	if g.State() != StateCommandPhase {
		t.Fatalf("按了確認之後在 %v", g.State())
	}
	if g.infoScreenTab() != infoTabEnemy {
		t.Errorf("關窗把分頁清掉了（現在第 %d 頁）；原版 ds:8F2F 不清", g.infoScreenTab())
	}
	g.Click(352, 125)
	g.Click(400, 145)
	if g.infoScreenTab() != infoTabEnemy {
		t.Errorf("重開之後在第 %d 頁，原版還停在敵艦（收據 fcmd-info-reopen 的 ds:8F2F ＝ 01）",
			g.infoScreenTab())
	}
}

// 游標經過分頁鈕不反白：只搬滑鼠，畫面只有游標本身變（收據 `fcmd-info-hover`）。
func TestInfoScreenNoHoverHighlight(t *testing.T) {
	g := reachInfoScreen(t)
	g.MoveMouse(424, 168)
	matches(t, g, "fcmd-info-hover")
}
