package game

import "testing"

// 星系情報視窗（`docs/spec/053`、`docs/re/320`）：從 `fcmd-base` 左鍵點星圖，與
// `tools/fleetcmd_receipts.sh` 第 18 段的 `fcmd-rinfo-*` 收據整張逐位元組比。
// 座標是畫面座標（滑鼠 y ＋ 40）：瓦爾哈拉 (32,72)、費沙 (163,136)、巴拉特 (300,310)、
// 「結束」(270,390)。

// 點右側面板看的那一格（首都）→ 視窗：星系名、小圖與四顆行星的引線、七列待命
// 船艦（艦數／部隊）、飛彈與戰鬥機；「結束」或右鍵關、點別處不動。
func TestRegionInfoMatchesOracle(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(32, 72)
	if g.State() != StateRegionInfo || g.rinfo != 0 {
		t.Fatalf("點了瓦爾哈拉之後在 %v、星域 %d", g.State(), g.rinfo)
	}
	matchesRegionInfo(t, g, "fcmd-rinfo")
	t.Run("outside", func(t *testing.T) {
		g := reachFleetDecided(t)
		g.Click(32, 72)
		g.Click(100, 140)
		if g.State() != StateRegionInfo {
			t.Fatalf("點在視窗裡不是結束的地方之後在 %v", g.State())
		}
		matchesRegionInfo(t, g, "fcmd-rinfo-outside")
	})
	t.Run("rclick", func(t *testing.T) {
		g := reachFleetDecided(t)
		g.Click(32, 72)
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-rinfo-rclick")
	})
	g.Click(270, 390) // 「結束」
	if g.State() != StateCommandPhase {
		t.Fatalf("結束之後在 %v", g.State())
	}
	matches(t, g, "fcmd-rinfo-close")
}

// 右鍵點星圖什麼都不換：右側面板仍是瓦爾哈拉、清單仍是第１艦隊那一列、地圖反白
// 不動，只有左下角那塊跟著游標寫成巴拉特（收據 `fcmd-map-rclick`：`ds:9588` 仍是
// 0；與左鍵同一個落點的 `fcmd-rinfo-ballat` 是對照）。
func TestMapRightClickKeepsPanel(t *testing.T) {
	g := reachFleetDecided(t)
	g.RightClick(300, 310)
	if g.State() != StateCommandPhase || g.listRegion() != 0 || g.hoverRegion != 31 {
		t.Fatalf("右鍵巴拉特之後在 %v、面板看 %d、左下角 %d", g.State(), g.listRegion(), g.hoverRegion)
	}
	matches(t, g, "fcmd-map-rclick")
}

// 兩塊事先列出的排除區：
//   - 小圖最後一列（畫面 y 212、x 24–151）：原版那一列的來源列超出地圖，讀到的是
//     緩衝區後面的東西（與指定待命行星那張同一個已知差異，規格 051／053）；
//   - 右側面板「經濟力」那個數字（x 432–479、y 224–233）：remake 的月底結算把同盟
//     與中立星域的稅在同一趟扣掉，原版是各座位在自己的回合扣、中立不扣，所以
//     796/01 帝國看到的別家經濟力比 remake 多一點（巴拉特 68282 對 65826、費沙
//     103100 對 99904；規格 053 已知差異）。
var (
	skipMinimapLastRow = [4]int{24, 212, 151, 212}
	skipForeignEconomy = [4]int{432, 224, 479, 233}
)

func matchesRegionInfo(t *testing.T, g *Game, batch string) {
	t.Helper()
	matchesExcept(t, g, batch, [][4]int{skipMinimapLastRow})
}

func matchesForeignRegionInfo(t *testing.T, g *Game, batch string) {
	t.Helper()
	matchesExcept(t, g, batch, [][4]int{skipMinimapLastRow, skipForeignEconomy})
}

func matchesForeignPanel(t *testing.T, g *Game, batch string) {
	t.Helper()
	matchesExcept(t, g, batch, [][4]int{skipForeignEconomy})
}

// 點別的格是換選星域（右側面板與清單跟著換），再點一次同一格才開視窗：
// 中立的費沙與同盟的巴拉特七列都印 `--------`（第六行是原版的錯字「Ｓ」）。
func TestRegionInfoOtherRegionsMatchOracle(t *testing.T) {
	t.Run("fezzan", func(t *testing.T) {
		g := reachFleetDecided(t)
		g.Click(163, 136)
		if g.State() != StateCommandPhase || g.selectedRegion != 14 {
			t.Fatalf("點了費沙之後在 %v、選中 %d", g.State(), g.selectedRegion)
		}
		matchesForeignPanel(t, g, "fcmd-rinfo-fezzan")
		g.Click(163, 136)
		if g.State() != StateRegionInfo || g.rinfo != 14 {
			t.Fatalf("再點一次費沙之後在 %v、星域 %d", g.State(), g.rinfo)
		}
		matchesForeignRegionInfo(t, g, "fcmd-rinfo-fezzan2")
	})
	g := reachFleetDecided(t)
	g.Click(300, 310)
	if g.State() != StateCommandPhase || g.selectedRegion != 31 {
		t.Fatalf("點了巴拉特之後在 %v、選中 %d", g.State(), g.selectedRegion)
	}
	matchesForeignPanel(t, g, "fcmd-rinfo-ballat")
	g.Click(300, 310)
	if g.State() != StateRegionInfo || g.rinfo != 31 {
		t.Fatalf("再點一次巴拉特之後在 %v、星域 %d", g.State(), g.rinfo)
	}
	matchesForeignRegionInfo(t, g, "fcmd-rinfo-ballat2")
	g.Click(270, 390)
	if g.State() != StateCommandPhase {
		t.Fatalf("結束之後在 %v", g.State())
	}
	matchesForeignPanel(t, g, "fcmd-rinfo-ballat2-close")
}
