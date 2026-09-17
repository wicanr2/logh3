package game

import "testing"

// 一覽表四張的列（規格 055、`docs/re/323`）：從 `fcmd-base` 開一覽表 → 各張 → 點列，
// 與 `tools/fleetcmd_receipts.sh` 第 20 段的 `fcmd-lst-*` 收據整張逐位元組比。
// 座標是畫面座標（滑鼠 y ＋ 40）：一覽表 (448,125)、星系資料 (460,125)／艦隊資料
// (460,145)／提督資料 (460,165)／敵將資料 (460,185)；列 k 在 (460, 90＋20k)；
// ▲ (615,90)、▼ (615,390)；提督詳細資料 經歷 (448,320)／結束 (496,320)。
// 收據裡列與鈕都是先移再按（原版把「移動與按鍵同一瞬間」的那一下丟掉），
// remake 的 Click 本來就先移動再按。

func reachListScreen(t *testing.T, item int) *Game {
	t.Helper()
	g := reachFleetDecided(t)
	g.Click(448, 125)
	if g.State() != StateListMenu {
		t.Fatalf("點一覽表之後在 %v", g.State())
	}
	g.Click(460, 125+20*item)
	return g
}

// 星系資料：點列 ＝ 換選那一格 ＋ 關清單 ＋ 星系情報視窗；分頁與位移每座位記住；
// 游標停在列上換反白與左下角、離開列不動。
func TestListStarRowsMatchOracle(t *testing.T) {
	g := reachListScreen(t, 0)
	if g.State() != StateStarList {
		t.Fatalf("在 %v", g.State())
	}
	matches(t, g, "fcmd-lst-star")
	t.Run("row0", func(t *testing.T) {
		g := reachListScreen(t, 0)
		g.Click(460, 90)
		if g.State() != StateRegionInfo || g.rinfo != 0 || g.listRegion() != 0 {
			t.Fatalf("點瓦爾哈拉那一列之後在 %v、視窗 %d、面板 %d", g.State(), g.rinfo, g.listRegion())
		}
		matchesRegionInfo(t, g, "fcmd-lst-star-row")
		g.Click(270, 390)
		if g.State() != StateCommandPhase {
			t.Fatalf("結束之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-star-row-close")
	})
	t.Run("row2", func(t *testing.T) {
		g := reachListScreen(t, 0)
		g.Click(460, 130)
		if g.State() != StateRegionInfo || g.rinfo != 3 || g.listRegion() != 3 {
			t.Fatalf("點卡斯特羅普那一列之後在 %v、視窗 %d、面板 %d", g.State(), g.rinfo, g.listRegion())
		}
		matchesRegionInfo(t, g, "fcmd-lst-star-row2")
		g.Click(270, 390)
		matches(t, g, "fcmd-lst-star-row2-close")
	})
	t.Run("hover", func(t *testing.T) {
		g := reachListScreen(t, 0)
		g.MoveMouse(460, 290)
		matches(t, g, "fcmd-lst-star-hover")
		g = reachListScreen(t, 0)
		g.MoveMouse(100, 140)
		matches(t, g, "fcmd-lst-star-hover-out")
	})
	t.Run("scroll", func(t *testing.T) {
		g := reachListScreen(t, 0)
		g.Click(615, 390)
		if g.lst.starScroll[0] != 1 {
			t.Fatalf("▼ 之後位移 %d", g.lst.starScroll[0])
		}
		matches(t, g, "fcmd-lst-star-down")
		g.Click(615, 90)
		matches(t, g, "fcmd-lst-star-up")
		g = reachListScreen(t, 0)
		g.Click(615, 390)
		g.Click(460, 90)
		if g.State() != StateRegionInfo || g.rinfo != 4 {
			t.Fatalf("捲一列之後點列 0 在 %v、視窗 %d（原版 4 菲利亞）", g.State(), g.rinfo)
		}
		matchesRegionInfo(t, g, "fcmd-lst-star-down-row")
	})
	t.Run("rclick", func(t *testing.T) {
		g := reachListScreen(t, 0)
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-star-rclick")
	})
}

// 艦隊資料：一列 ＝ 第Ｎ艦隊＋提督＋士氣；點列開艦隊視窗（那一列反白），結束或
// 右鍵回清單；點沒有艦隊的列沒事。
func TestListFleetRowsMatchOracle(t *testing.T) {
	g := reachListScreen(t, 1)
	if g.State() != StateFleetList {
		t.Fatalf("在 %v", g.State())
	}
	matches(t, g, "fcmd-lst-fleet")
	g.Click(460, 90)
	if g.State() != StateFleetWindow || g.fcmd.fleet != 0 || !g.lst.fleetFrom {
		t.Fatalf("點第１艦隊那一列之後在 %v、艦隊 %d", g.State(), g.fcmd.fleet)
	}
	matches(t, g, "fcmd-lst-fleet-row")
	g.Click(480, 380) // 「結束」
	if g.State() != StateFleetList || g.lst.fleetFrom || g.lst.pressed != -1 {
		t.Fatalf("結束之後在 %v", g.State())
	}
	matches(t, g, "fcmd-lst-fleet-row-close")
	t.Run("rclick-window", func(t *testing.T) {
		g := reachListScreen(t, 1)
		g.Click(460, 90)
		g.RightClick(300, 240)
		if g.State() != StateFleetList {
			t.Fatalf("艦隊視窗上右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-fleet-row-rclick")
	})
	t.Run("empty-row", func(t *testing.T) {
		g := reachListScreen(t, 1)
		g.Click(460, 110)
		if g.State() != StateFleetList {
			t.Fatalf("點空列之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-fleet-row1")
	})
	t.Run("rclick", func(t *testing.T) {
		g := reachListScreen(t, 1)
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-fleet-rclick")
	})
}

// 提督資料：點列開提督詳細資料（第１艦隊／現在位置 瓦爾哈拉：出擊中／經歷／結束），
// 經歷疊在上面，結束或右鍵一層一層關；▼ 一列；游標停在列上換反白、離開清單
// 左下角清空。
func TestListAdmiralRowsMatchOracle(t *testing.T) {
	g := reachListScreen(t, 2)
	if g.State() != StateAdmiralListView {
		t.Fatalf("在 %v", g.State())
	}
	matches(t, g, "fcmd-lst-adm")
	g.Click(460, 90)
	if g.lst.detail != 0 || !g.lst.detailOwn {
		t.Fatalf("點萊因哈特那一列之後詳細資料 %d", g.lst.detail)
	}
	matches(t, g, "fcmd-lst-adm-row")
	t.Run("close", func(t *testing.T) {
		g := reachListScreen(t, 2)
		g.Click(460, 90)
		g.Click(496, 320)
		if g.lst.detail != -1 || g.State() != StateAdmiralListView {
			t.Fatalf("結束之後詳細資料 %d、在 %v", g.lst.detail, g.State())
		}
		matches(t, g, "fcmd-lst-adm-row-close")
	})
	t.Run("rclick-detail", func(t *testing.T) {
		g := reachListScreen(t, 2)
		g.Click(460, 90)
		g.RightClick(300, 240)
		if g.lst.detail != -1 || g.State() != StateAdmiralListView {
			t.Fatalf("詳細資料上右鍵之後 %d、在 %v", g.lst.detail, g.State())
		}
		matches(t, g, "fcmd-lst-adm-row-rclick")
	})
	g.Click(448, 320) // 「經歷」
	if !g.lst.bio {
		t.Fatalf("經歷沒開")
	}
	matches(t, g, "fcmd-lst-adm-row-bio")
	g.Click(496, 320) // 經歷只認右鍵
	if !g.lst.bio {
		t.Fatalf("經歷被左鍵關掉了")
	}
	matches(t, g, "fcmd-lst-adm-row-bio-close")
	t.Run("hover", func(t *testing.T) {
		g := reachListScreen(t, 2)
		g.MoveMouse(460, 250)
		matches(t, g, "fcmd-lst-adm-hover")
		g = reachListScreen(t, 2)
		g.Click(100, 140)
		if g.State() != StateAdmiralListView {
			t.Fatalf("點清單外之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-adm-outside")
	})
	t.Run("scroll", func(t *testing.T) {
		g := reachListScreen(t, 2)
		g.Click(615, 390)
		if g.lst.admScroll != 1 {
			t.Fatalf("▼ 之後位移 %d", g.lst.admScroll)
		}
		matches(t, g, "fcmd-lst-adm-down")
		g.Click(460, 90)
		if g.lst.detail != 1 {
			t.Fatalf("捲一列之後點列 0 的詳細資料 %d（原版 1 吉爾菲艾斯）", g.lst.detail)
		}
		matches(t, g, "fcmd-lst-adm-down-row")
	})
	t.Run("rclick", func(t *testing.T) {
		g := reachListScreen(t, 2)
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-adm-rclick")
	})
}

// 敵將資料：詳細資料的現在位置印 ＊＊＊＊＊＊＊＊、沒有狀態；游標不看列。
func TestListEnemyRowsMatchOracle(t *testing.T) {
	g := reachListScreen(t, 3)
	if g.State() != StateEnemyList {
		t.Fatalf("在 %v", g.State())
	}
	matches(t, g, "fcmd-lst-enemy")
	g.Click(460, 90)
	if g.lst.detail < 0 || g.lst.detailOwn {
		t.Fatalf("點亞典波羅那一列之後詳細資料 %d／自軍 %v", g.lst.detail, g.lst.detailOwn)
	}
	matches(t, g, "fcmd-lst-enemy-row")
	g.Click(496, 320)
	if g.lst.detail != -1 {
		t.Fatalf("結束之後詳細資料 %d", g.lst.detail)
	}
	matches(t, g, "fcmd-lst-enemy-row-close")
	t.Run("hover", func(t *testing.T) {
		g := reachListScreen(t, 3)
		g.MoveMouse(460, 250)
		matches(t, g, "fcmd-lst-enemy-hover")
	})
	t.Run("rclick", func(t *testing.T) {
		g := reachListScreen(t, 3)
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-enemy-rclick")
	})
}

// 一覽表的分頁鈕（規格 055 分頁一節、`docs/re/327`）：星系資料 艦隊／配屬／稅金
// （`0D05:02FC`，`ds:8F2D[座位]`）、艦隊資料 士氣／部隊／作戰（`0C7F:0339`，
// `ds:8F31[座位]`）；分頁鈕在 (432|496|560, 60) 48×20，收據點 (520,70)／(584,70)。
func TestListScreenTabsMatchOracle(t *testing.T) {
	t.Run("star-assign", func(t *testing.T) {
		g := reachListScreen(t, 0)
		g.Click(520, 70)
		if g.lst.starTab[0] != 1 {
			t.Fatalf("配屬之後分頁 %d", g.lst.starTab[0])
		}
		matches(t, g, "fcmd-lst-star-tab-assign")
	})
	t.Run("star-tax-reopen", func(t *testing.T) {
		g := reachListScreen(t, 0)
		g.Click(584, 70)
		matches(t, g, "fcmd-lst-star-tab-tax")
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-lst-star-tax-rclick")
		g.Click(448, 125)
		g.Click(460, 125)
		if g.State() != StateStarList || g.lst.starTab[0] != 2 {
			t.Fatalf("再開之後在 %v、分頁 %d（原版記住稅金）", g.State(), g.lst.starTab[0])
		}
		matches(t, g, "fcmd-lst-star-reopen")
	})
	t.Run("fleet-unit", func(t *testing.T) {
		g := reachListScreen(t, 1)
		g.Click(520, 70)
		if g.lst.fleetTab[0] != 1 {
			t.Fatalf("部隊之後分頁 %d", g.lst.fleetTab[0])
		}
		matches(t, g, "fcmd-lst-fleet-tab-unit")
	})
	t.Run("fleet-ops", func(t *testing.T) {
		g := reachListScreen(t, 1)
		g.Click(584, 70)
		matches(t, g, "fcmd-lst-fleet-tab-ops")
	})
}
