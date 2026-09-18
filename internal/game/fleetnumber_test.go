package game

import (
	"path/filepath"
	"testing"
)

// 兩位數艦隊編號（#74）：`3982:0038` 小於 10 印全形、否則 `%2d` 兩個半形塞一格
// （`docs/re/143`）。四個畫面各一份 S3「巴米利恩會戰」的收據——那一局開局帝國槽 9
// 就是第１０艦隊、同盟槽 28 是第１３艦隊、打雷奧米爾的槽 31 是第１６艦隊，
// 不必假造編號。
//
// 座標是畫面座標（收據旗標的滑鼠 y ＋ 40）。

// reachS3Map 載 S3 存檔，停在戰略地圖的命令階段（收據 `fcmd-s3-base`）。
func reachS3Map(t *testing.T) *Game {
	t.Helper()
	res := loadWithRules(t)
	loadBattleMaps(t, res)
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	g.res.OrigDir = filepath.Join(root, "org_game/extracted")
	g.loadSlotFromDisk(10)
	if g.State() != StateCommandPhase || g.faction != 0 {
		t.Fatalf("S3 載進來在 %v、座位 %d", g.State(), g.faction)
	}
	return g
}

// 艦隊視窗與「指定待命行星」（收據 `fcmd-s3-n10-*`，`tools/fleetcmd_receipts.sh` 第 27 段）：
// 一覽表 → 艦隊資料 → 第 8 列（第１０艦隊）→ 作戰 → 待命。
func TestFleetNumberTwoDigitMatchesOracle(t *testing.T) {
	g := reachS3Map(t)
	g.Click(448, 125) // 一覽表
	g.Click(460, 145) // 艦隊資料
	if g.State() != StateFleetList {
		t.Fatalf("開艦隊資料之後在 %v", g.State())
	}
	matches(t, g, "fcmd-s3-n10-list")

	g.Click(450, 230) // 列 7 ＝ 第１０艦隊
	if g.State() != StateFleetWindow {
		t.Fatalf("點第 8 列之後在 %v", g.State())
	}
	if f := g.fleets[g.fcmd.fleet]; f.Number != 10 {
		t.Fatalf("開出來的是第 %d 艦隊，收據是第１０艦隊", f.Number)
	}
	matches(t, g, "fcmd-s3-n10-win")

	g.Click(490, 345) // 作戰
	matches(t, g, "fcmd-s3-n10-menu")
	g.Click(497, 392) // 待命
	if g.State() != StateFleetStandby {
		t.Fatalf("點待命之後在 %v", g.State())
	}
	// 小圖第 127 列是原版讀到地圖緩衝區後面的垃圾（規格 051 已知差異），
	// 與其他「指定待命行星」的收據同一塊排除區。
	matchesStandby(t, g, "fcmd-s3-n10-standby")
}

// 敵艦隊視窗（收據 `enm-n13-row`，`tools/enemytab_receipts.sh` 第 5 段）：同盟槽 28
// （第１３艦隊，楊威利）搬進瓦爾哈拉、九組索敵等級 3。
func TestEnemyWindowTwoDigitMatchesOracle(t *testing.T) {
	g := reachS3Map(t)
	f := g.fleetBySlot(t, 28)
	if f.Number != 13 {
		t.Fatalf("槽 28 是第 %d 艦隊，收據是第１３艦隊", f.Number)
	}
	f.Region = 0
	for i := range f.Composition {
		f.Composition[i].Detail = 3
	}
	g.Click(490, 290) // 敵艦隊分頁
	if g.listTab() != listTabEnemy {
		t.Fatalf("點了敵艦隊分頁之後在第 %d 頁", g.listTab())
	}
	matches(t, g, "enm-n13")
	g.Click(450, 308)
	if g.tabs.enemy < 0 {
		t.Fatal("點列沒開視窗")
	}
	matches(t, g, "enm-n13-row")
}

// 交戰演出（收據 `eng-known`，`tools/engagement_receipts.sh` 第 6 段）：同一場打雷奧米爾，
// 索敵等級 3 之後右框印「第16艦隊　費雪」而不是「第？艦隊　？？？…」。
func TestEngagementTwoDigitMatchesOracle(t *testing.T) {
	g := reachEngagementFixture(t)
	for i := range g.battleUnits {
		if g.battleUnits[i].Fleet == g.engQueue[0].fleet {
			g.battleUnits[i].Detail = 3
		}
	}
	g.engagementTick(20 - g.engTick)
	if g.state != StateEngagement {
		t.Fatal("tick 20 就收場了")
	}
	matchesExcept(t, g, "eng-known", engagementSkips(g))
}
