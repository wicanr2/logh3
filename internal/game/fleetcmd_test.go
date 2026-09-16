package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 戰略地圖上的艦隊命令（`docs/spec/051` 第一片）。每一張都是正常玩家路徑：
// 從 `dec1`（艦隊組成按「決定」回到地圖）接著點下去，與 `tools/fleetcmd_receipts.sh`
// 的 `fcmd-*` 收據逐位元組比。座標是畫面座標（收據旗標的滑鼠 y ＋ 40）。

// reachFleetDecided 走到 `dec1`：第１艦隊剛編成、停在首都。資源要帶規則表
// （`loadWithRules`）：艦隊視窗印的編制、費用與速度都從艦種表算；
// 「指定待命行星」的小圖與行星位置還要掃星域地圖（`loadBattleMaps`）。
func reachFleetDecided(t *testing.T) *Game {
	t.Helper()
	res := loadWithRules(t)
	loadBattleMaps(t, res)
	g := reachCommandPhaseOnMapWith(t, res)
	g.MoveMouse(600, 340)
	g = reachFleetOrgFrom(t, g)
	g.Click(480, 350) // 「確認」→ 艦隊組成
	g.Click(510, 350) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	return g
}

// reachFleetWindow 再點右下角清單第 0 列 → 艦隊視窗（收據 `fcmd-win`）。
func reachFleetWindow(t *testing.T) *Game {
	t.Helper()
	g := reachFleetDecided(t)
	g.Click(450, 308)
	if g.State() != StateFleetWindow {
		t.Fatalf("點了清單第 0 列之後在 %v", g.State())
	}
	return g
}

func TestFleetWindowMatchesOracle(t *testing.T) {
	g := reachFleetWindow(t)
	matches(t, g, "fcmd-win")
}

// 剛編成的第１艦隊已經算進星域記錄的艦隊數：`dec1` 之後只把游標移到
// 瓦爾哈拉，左下角重寫成「艦隊 1：0」（收據 `fcmd-hover-home`）；而回到
// 地圖那一刻沒重寫的那塊（`dec1`）與右側面板整條鏈仍是編成之前寫的 0：0
// ——面板是寫進去那一刻的快照，不是每幀重算（`docs/re/312` §八）。
func TestFleetHoverHomeCountsNewFleet(t *testing.T) {
	g := reachFleetDecided(t)
	if rec := g.regionRecord(rules.CapitalRegion(g.faction)); rules.RegionFleets(rec, g.faction) != 1 {
		t.Fatalf("決定之後首都記錄的艦隊數 = %d，原版 +0x28 是 1（fcx-peek-base）", rules.RegionFleets(rec, g.faction))
	}
	g.MoveMouse(32, 72)
	matches(t, g, "fcmd-hover-home")
}

// 「決定」也把提督與副官記錄的 `+0x09` 從 0x80 改成槽號 0（探索批次
// `fcx-peek-slot-pre`／`fcx-peek-slot`：萊因哈特與副官兩筆 80 → 00）。
func TestFleetDecideSetsAdmiralFleetSlot(t *testing.T) {
	g := reachFleetDecided(t)
	f := g.fleets[len(g.fleets)-1]
	for _, a := range []int{f.Admiral, f.Adjutant} {
		if r := g.admiralRecord(a); r == nil || r[0x09] != 0 {
			t.Fatalf("提督 %d 的 +0x09 = %#x，原版決定之後是 0x00", a, r[0x09])
		}
	}
}

// 清單上沒有艦隊的那一列點了什麼都不會發生（探索批次 `fcx-row0` 第一次：
// 點 (450,350) 畫面只有游標動）。
func TestFleetListEmptyRowDoesNothing(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(450, 350)
	if g.State() != StateCommandPhase {
		t.Fatalf("點了空列之後在 %v", g.State())
	}
}

// 「作戰」開五項選單，反白是游標所在那一項（`fcmd-menu` 停在第三項「駐留」），
// 游標移到別項反白跟著走（`fcmd-menu-hover`：第一項「移動」）。
func TestFleetActionMenuMatchesOracle(t *testing.T) {
	g := reachFleetWindow(t)
	g.Click(490, 345) // 「作戰」
	if g.State() != StateFleetActionMenu {
		t.Fatalf("點了作戰之後在 %v", g.State())
	}
	matches(t, g, "fcmd-menu")
	g.MoveMouse(497, 312)
	matches(t, g, "fcmd-menu-hover")
}

// 選單開著時右鍵只關選單（`fcmd-menu-rclick`）；艦隊視窗右鍵回地圖
// （`fcmd-win-rclick`）；「結束」也回地圖（`fcmd-close`）。
func TestFleetWindowClosesMatchOracle(t *testing.T) {
	t.Run("menu-rclick", func(t *testing.T) {
		g := reachFleetWindow(t)
		g.Click(490, 345)
		g.RightClick(300, 240)
		if g.State() != StateFleetWindow {
			t.Fatalf("選單右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-menu-rclick")
	})
	t.Run("win-rclick", func(t *testing.T) {
		g := reachFleetWindow(t)
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("視窗右鍵之後在 %v", g.State())
		}
		matches(t, g, "fcmd-win-rclick")
	})
	t.Run("close", func(t *testing.T) {
		g := reachFleetWindow(t)
		g.Click(490, 390) // 「結束」
		if g.State() != StateCommandPhase {
			t.Fatalf("點了結束之後在 %v", g.State())
		}
		matches(t, g, "fcmd-close")
	})
}

// reachFleetTargetPick 選「移動」→ 目的地選取（`fcmd-pick`）。
func reachFleetTargetPick(t *testing.T) *Game {
	t.Helper()
	g := reachFleetWindow(t)
	g.Click(490, 345) // 「作戰」
	g.Click(497, 312) // 「移動」
	if g.State() != StateFleetTargetPick {
		t.Fatalf("點了移動之後在 %v", g.State())
	}
	return g
}

// 目的地選取：地圖依允許表重新著色，視窗右半照舊（`fcmd-pick`）；
// 允許表 ＝ 從瓦爾哈拉起算 k ≤ 1 的可達表 ∧ 帝國的星域。
func TestFleetTargetPickMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	allowed := g.FleetTargetsAllowed()
	for r, want := range map[int]bool{0: true, 2: true, 6: true, 13: true, 14: false, 31: false} {
		if allowed[r] != want {
			t.Errorf("星域 %d 允許＝%v，原版是 %v（`docs/re/312` §四）", r, allowed[r], want)
		}
	}
	matches(t, g, "fcmd-pick")
	g.RightClick(300, 240)
	if g.State() != StateFleetWindow {
		t.Fatalf("選取時右鍵之後在 %v", g.State())
	}
	matches(t, g, "fcmd-pick-rclick")
}

// 點亞爾提那 → 「要移動至帝國軍亞爾提那星系嗎？」（`fcmd-confirm`）；
// 點走不到的巴拉特 → 「這是無法前往的星系。」（`fcmd-far`）；
// 點自己所在的瓦爾哈拉 → 動詞換成「待命」（`fcmd-self`）。
func TestFleetConfirmMatchesOracle(t *testing.T) {
	t.Run("confirm", func(t *testing.T) {
		g := reachFleetTargetPick(t)
		g.Click(85, 128)
		if g.State() != StateFleetConfirm {
			t.Fatalf("點了亞爾提那之後在 %v", g.State())
		}
		matches(t, g, "fcmd-confirm")
	})
	t.Run("far", func(t *testing.T) {
		g := reachFleetTargetPick(t)
		g.Click(300, 310)
		if g.State() != StateFleetMessage {
			t.Fatalf("點了巴拉特之後在 %v", g.State())
		}
		matches(t, g, "fcmd-far")
	})
	t.Run("self", func(t *testing.T) {
		g := reachFleetTargetPick(t)
		g.Click(32, 72)
		if g.State() != StateFleetConfirm {
			t.Fatalf("點了瓦爾哈拉之後在 %v", g.State())
		}
		matches(t, g, "fcmd-self")
	})
}

// 「決定」寫命令回地圖：艦隊記錄 `+0x07 ← 2`、`+0x08 ← 0`（`fcmd-done`）；
// 清單第 0 列「第」前面印 ☆（`+0x08` 查 `ds:1A66` 那張記號表，`docs/re/313`）。
func TestFleetMoveOrderWritesFleet(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)  // 亞爾提那
	g.Click(432, 263) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	f := g.fleets[0]
	if f.Target != 2 || f.Move != fleetActionMove || f.Parked {
		t.Fatalf("決定之後艦隊記錄 Target=%d Move=%d Parked=%v，原版 +0x07=2、+0x08=0", f.Target, f.Move, f.Parked)
	}
	// 與原版倒出來的艦隊記錄對：`fcmd-done/fleet0.bin` 的 +0x07／+0x08。
	if rec, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-done/fleet0.bin")); err == nil && len(rec) >= 9 {
		if int(rec[7]) != f.Target || int(rec[8]) != f.Move {
			t.Fatalf("原版艦隊記錄 +0x07=%d +0x08=%d，remake Target=%d Move=%d", rec[7], rec[8], f.Target, f.Move)
		}
	}
	matches(t, g, "fcmd-done")
}

// 下了命令之後再點清單那一列：視窗第 5–7 列換成「移動戰鬥／現在位置／移動目標」
// （`0DEA:03F9`／`04C0` 對 `+0x08 != 5` 的分支，收據 `fcmd-reopen`）。
func TestFleetReopenShowsTarget(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)  // 亞爾提那
	g.Click(432, 263) // 「決定」
	g.Click(450, 308) // 清單第 0 列
	if g.State() != StateFleetWindow {
		t.Fatalf("再點清單之後在 %v", g.State())
	}
	matches(t, g, "fcmd-reopen")
}

// 目的地點自己所在的星域再按「決定」＝ 取消命令：本來就沒有命令時什麼都
// 不改，而且**視窗留著**（`fcmd-self-done` 與 `fcmd-win` 只差游標）；
// 清單那一列沒有記號。
func TestFleetCancelOrderKeepsWindow(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(32, 72)   // 瓦爾哈拉
	g.Click(432, 263) // 「決定」
	if g.State() != StateFleetWindow {
		t.Fatalf("取消命令之後在 %v，原版回到艦隊視窗", g.State())
	}
	if f := g.fleets[0]; f.Target != rules.NoTarget || f.Move != rules.MoveParked || !f.Parked {
		t.Fatalf("取消之後艦隊記錄 %+v", f)
	}
	matches(t, g, "fcmd-self-done")
}

// 駐留（第 2 項）與輸送（第 3 項）走同一個目的地迴圈，差在 `+0x08` 寫 2／3
// （`fcmd-garrison-*`／`fcmd-transport-*`）。
func TestFleetGarrisonAndTransportOrders(t *testing.T) {
	for _, c := range []struct {
		name  string
		y     int
		move  int
		batch string
	}{
		{"駐留", 352, fleetActionGarrison, "fcmd-garrison"},
		{"輸送", 372, fleetActionTransport, "fcmd-transport"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := reachFleetWindow(t)
			g.Click(490, 345) // 「作戰」
			g.Click(497, c.y)
			if g.State() != StateFleetTargetPick {
				t.Fatalf("點了%s之後在 %v", c.name, g.State())
			}
			matches(t, g, c.batch+"-pick")
			g.Click(85, 128) // 亞爾提那
			if g.State() != StateFleetConfirm {
				t.Fatalf("點了亞爾提那之後在 %v", g.State())
			}
			matches(t, g, c.batch+"-confirm")
			g.Click(432, 263) // 「決定」
			if g.State() != StateCommandPhase {
				t.Fatalf("點了決定之後在 %v", g.State())
			}
			if f := g.fleets[0]; f.Target != 2 || f.Move != c.move || f.Parked {
				t.Fatalf("決定之後艦隊記錄 Target=%d Move=%d Parked=%v，原版 +0x07=2、+0x08=%d",
					f.Target, f.Move, f.Parked, c.move)
			}
			matches(t, g, c.batch+"-done")
		})
	}
}

// 下了移動命令之後按「實行」：艦隊走一格到亞爾提那、命令清掉（`+0x07` FF、
// `+0x08` 05），沒有任何出發／到達報告，直接回地圖（`fcmd-exec`）。
func TestFleetOrderThenExecuteMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)  // 亞爾提那
	g.Click(432, 263) // 「決定」
	g.Click(544, 125) // 系統指令
	g.Click(560, 125) // 實行
	if g.State() != StateCommandPhase {
		t.Fatalf("實行之後在 %v", g.State())
	}
	if f := g.fleets[0]; f.Region != 2 || f.Target != rules.NoTarget || f.Move != rules.MoveParked || !f.Parked {
		t.Fatalf("實行之後艦隊記錄 %+v，原版在星域 2、+0x07=FF、+0x08=5", f)
	}
	matches(t, g, "fcmd-exec")
}

func TestFleetConfirmCancelMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)  // 亞爾提那
	g.Click(383, 263) // 「中止」
	if g.State() != StateFleetTargetPick {
		t.Fatalf("點了中止之後在 %v", g.State())
	}
	if f := g.fleets[0]; f.Target != rules.NoTarget || !f.Parked {
		t.Fatalf("中止之後艦隊記錄變了：%+v", f)
	}
	matches(t, g, "fcmd-cancel")
}

// ── 待命（第 4 項）：「指定待命行星」（`docs/re/314`）──

// reachFleetStandby 選「待命」→ 指定待命行星（`fcmd-standby`）。
func reachFleetStandby(t *testing.T) *Game {
	t.Helper()
	g := reachFleetWindow(t)
	g.Click(490, 345) // 「作戰」
	g.Click(497, 392) // 「待命」
	if g.State() != StateFleetStandby {
		t.Fatalf("點了待命之後在 %v", g.State())
	}
	return g
}

// matchesStandby 是整張比，只放過小圖最後一列（y 347、x 152–279）：原版那
// 一列的來源列是 102，超出 102 列的地圖，讀到的是緩衝區後面的東西
// （`docs/spec/051` 已知差異）。
func matchesStandby(t *testing.T, g *Game, batch string) {
	t.Helper()
	matchesExcept(t, g, batch, [][4]int{{152, 347, 279, 347}})
}

// matchesExcept 整張比，但跳過列出的矩形（含端點）；每一塊都要在呼叫端
// 寫清楚理由。
func matchesExcept(t *testing.T, g *Game, batch string, skip [][4]int) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "screen.bin"))
	if err != nil {
		t.Skipf("沒有收據 %s，跳過（tools/oracle.sh 可重建）", batch)
	}
	got := g.Frame().Pix
	bad := 0
	for i := range want {
		x, y := i%640, i/640
		skipped := false
		for _, b := range skip {
			if x >= b[0] && x <= b[2] && y >= b[1] && y <= b[3] {
				skipped = true
				break
			}
		}
		if !skipped && want[i] != got[i] {
			bad++
		}
	}
	if bad != 0 {
		if os.Getenv("LOGH3_DUMP_REMAKE") != "" {
			out := filepath.Join(root, "workplace", "remake-"+batch+".bin")
			if err := os.WriteFile(out, got, 0o644); err != nil {
				t.Logf("倒 remake 畫面失敗：%v", err)
			} else {
				t.Logf("remake 畫面寫到 %s", out)
			}
		}
		t.Fatalf("%s：與原版不同的像素 %d / %d（已跳過 %d 塊）", batch, bad, len(want), len(skip))
	}
}

// 「指定待命行星」開出來：四顆行星四顆鈕、小圖、引線（`fcmd-standby`）；
// 點第 1 顆（布連娜）反白（`fcmd-standby-pick`）。
func TestFleetStandbyMatchesOracle(t *testing.T) {
	g := reachFleetStandby(t)
	matchesStandby(t, g, "fcmd-standby")
	g.Click(350, 250) // 第 1 顆：布連娜
	if g.State() != StateFleetStandby || g.fcmd.standbySel != 1 {
		t.Fatalf("點了行星之後在 %v，反白 %d", g.State(), g.fcmd.standbySel)
	}
	matchesStandby(t, g, "fcmd-standby-pick")
}

// 「決定」寫 `+0x07 ← 1`、`+0x08` 留 5、關視窗回地圖（`fcmd-standby-done`）；
// 再點清單那一列，視窗第 7 列印「待命行星 布連娜」（`fcmd-standby-reopen`）。
func TestFleetStandbyDecideWritesPlanet(t *testing.T) {
	g := reachFleetStandby(t)
	g.Click(350, 250) // 布連娜
	g.Click(400, 350) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	f := g.fleets[0]
	if f.Target != 1 || f.Move != rules.MoveParked || !f.Parked {
		t.Fatalf("決定之後艦隊記錄 Target=%d Move=%d Parked=%v，原版 +0x07=1、+0x08=5", f.Target, f.Move, f.Parked)
	}
	if rec, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-standby-done/fleet0.bin")); err == nil && len(rec) >= 9 {
		if int(rec[7]) != f.Target || int(rec[8]) != f.Move {
			t.Fatalf("原版艦隊記錄 +0x07=%d +0x08=%d，remake Target=%d Move=%d", rec[7], rec[8], f.Target, f.Move)
		}
	}
	matches(t, g, "fcmd-standby-done")
	g.Click(450, 308) // 清單第 0 列
	if g.State() != StateFleetWindow {
		t.Fatalf("再點清單之後在 %v", g.State())
	}
	matches(t, g, "fcmd-standby-reopen")
}

// 「中止」清掉反白、回艦隊視窗，艦隊記錄不動（`fcmd-standby-cancel`）；
// 右鍵等於「決定」：反白的那一顆照樣寫進去（`fcmd-standby-rclick`）。
func TestFleetStandbyCancelAndRightClick(t *testing.T) {
	t.Run("cancel", func(t *testing.T) {
		g := reachFleetStandby(t)
		g.Click(350, 250) // 布連娜
		g.Click(350, 350) // 「中止」
		if g.State() != StateFleetWindow {
			t.Fatalf("點了中止之後在 %v", g.State())
		}
		if f := g.fleets[0]; f.Target != rules.NoTarget || !f.Parked {
			t.Fatalf("中止之後艦隊記錄變了：%+v", f)
		}
		matches(t, g, "fcmd-standby-cancel")
	})
	t.Run("rclick", func(t *testing.T) {
		g := reachFleetStandby(t)
		g.Click(350, 250) // 布連娜
		g.RightClick(300, 240)
		if g.State() != StateCommandPhase {
			t.Fatalf("右鍵之後在 %v", g.State())
		}
		if f := g.fleets[0]; f.Target != 1 || f.Move != rules.MoveParked {
			t.Fatalf("右鍵之後艦隊記錄 %+v，原版 +0x07=1、+0x08=5", f)
		}
		matches(t, g, "fcmd-standby-rclick")
	})
}

// ── 占領（第 1 項）：允許表換成敵方＋中立、費沙先跳參謀提醒（`docs/re/315`）──

// reachFleetOccupyPick 選「占領」→ 目的地選取（`fcmd-occupy-pick`）。
func reachFleetOccupyPick(t *testing.T) *Game {
	t.Helper()
	g := reachFleetWindow(t)
	g.Click(490, 345) // 「作戰」
	g.Click(497, 332) // 「占領」
	if g.State() != StateFleetTargetPick {
		t.Fatalf("點了占領之後在 %v", g.State())
	}
	return g
}

// 占領的允許表：可達 ∧ 擁有者是敵方或中立（`0DEA:12CC` 給 `113A:0003` 的
// (敵方, 2)），地圖照它著色（`fcmd-occupy-pick`）；點自己的亞爾提那 →
// 「這是無法前往的星系。」（`fcmd-occupy-far`）；點同盟的亞斯提 →
// 「要占領同盟軍亞斯提星系嗎？」（`fcmd-occupy-confirm`）→ 決定寫 `+0x08 ← 1`、
// `+0x07 ← 16`，清單那一列印紅☆（`fcmd-occupy-done`）；再開視窗第 5–7 列印
// 「占領戰鬥／現在位置／移動目標 亞斯提」（`fcmd-occupy-reopen`）。
func TestFleetOccupyOrderMatchesOracle(t *testing.T) {
	g := reachFleetOccupyPick(t)
	allowed := g.FleetTargetsAllowed()
	for r, want := range map[int]bool{0: true, 2: false, 3: true, 14: true, 16: true, 31: false} {
		if allowed[r] != want {
			t.Errorf("星域 %d 允許＝%v，原版是 %v", r, allowed[r], want)
		}
	}
	matches(t, g, "fcmd-occupy-pick")
	t.Run("far", func(t *testing.T) {
		g := reachFleetOccupyPick(t)
		g.Click(85, 128) // 亞爾提那（自己的）
		if g.State() != StateFleetMessage {
			t.Fatalf("點了亞爾提那之後在 %v", g.State())
		}
		matches(t, g, "fcmd-occupy-far")
	})
	g.Click(193, 245) // 亞斯提
	if g.State() != StateFleetConfirm {
		t.Fatalf("點了亞斯提之後在 %v", g.State())
	}
	matches(t, g, "fcmd-occupy-confirm")
	g.Click(432, 263) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	if f := g.fleets[0]; f.Target != 16 || f.Move != fleetActionOccupy || f.Parked {
		t.Fatalf("決定之後艦隊記錄 Target=%d Move=%d Parked=%v，原版 +0x07=16、+0x08=1", f.Target, f.Move, f.Parked)
	}
	matches(t, g, "fcmd-occupy-done")
	g.Click(450, 308) // 清單第 0 列
	if g.State() != StateFleetWindow {
		t.Fatalf("再點清單之後在 %v", g.State())
	}
	matches(t, g, "fcmd-occupy-reopen")
}

// 占領目的地點費沙：先兩頁參謀提醒（`1213:0003` → `1213:01B8`，任一鍵翻頁），
// 再「要占領中立軍費沙星系嗎？」，決定寫 `+0x07 ← 14`、`+0x08 ← 1`
// （`fcmd-occupy-phezzan`／`-phezzan2`／`-phezzan3`／`-phezzan-done`）。
func TestFleetOccupyPhezzanWarningMatchesOracle(t *testing.T) {
	g := reachFleetOccupyPick(t)
	g.Click(163, 136) // 費沙
	if g.State() != StateFleetOccupyWarn {
		t.Fatalf("點了費沙之後在 %v", g.State())
	}
	matches(t, g, "fcmd-occupy-phezzan")
	g.Click(300, 240)
	if g.State() != StateFleetOccupyWarn || g.fcmd.warnPage != 1 {
		t.Fatalf("第一頁點了之後在 %v 第 %d 頁", g.State(), g.fcmd.warnPage)
	}
	matches(t, g, "fcmd-occupy-phezzan2")
	g.Click(300, 240)
	if g.State() != StateFleetConfirm {
		t.Fatalf("第二頁點了之後在 %v", g.State())
	}
	matches(t, g, "fcmd-occupy-phezzan3")
	g.Click(432, 263) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	if f := g.fleets[0]; f.Target != rules.PhezzanRegion || f.Move != fleetActionOccupy {
		t.Fatalf("決定之後艦隊記錄 %+v，原版 +0x07=14、+0x08=1", f)
	}
	matches(t, g, "fcmd-occupy-phezzan-done")
}

// ── 作戰選單五項的停用旗標（`docs/re/312` §7.6、`docs/re/315`）──

// reachFleetDecidedBattle 與 `reachFleetDecided` 同一條路，只是在戰略地圖
// 畫好、還沒點「戰略命令」之前把首都的 `+0x13` 設成 4（有戰鬥）——與收據
// `fcmd-battle-base` 在 170M 的 `-poke lin:5F2DF=04` 同一個時點。
func reachFleetDecidedBattle(t *testing.T) *Game {
	t.Helper()
	res := loadWithRules(t)
	loadBattleMaps(t, res)
	g := reachCommandPhaseOnMapWith(t, res)
	g.res.Stellar[rules.CapitalRegion(g.faction)*rules.RegionRecord+0x13] = rules.RegionFlagBattle
	g.MoveMouse(600, 340)
	g = reachFleetOrgFrom(t, g)
	g.Click(480, 350) // 「確認」→ 艦隊組成
	g.Click(510, 350) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	return g
}

// 首都有戰鬥：清單那一列印紅▼、左下角星域名後面「處於戰鬥狀態」且
// 「艦隊 a：b」的「：」跟著變紅（`fcmd-battle-base`）；艦隊視窗「作戰」列印
// 紅字「戰鬥中」（`fcmd-battle-win`）；作戰選單只剩「移動」可選——停用的四項
// 色 8，游標停在停用項上不畫反白（`fcmd-battle-menu`：游標在第三項）、移到
// 「移動」才畫（`fcmd-battle-menu-hover`）；點停用的「駐留」只關選單
// （`fcmd-battle-menu-click`）。
func TestFleetMenuDisabledInBattleRegion(t *testing.T) {
	g := reachFleetDecidedBattle(t)
	matches(t, g, "fcmd-battle-base")
	g.Click(450, 308) // 清單第 0 列
	if g.State() != StateFleetWindow {
		t.Fatalf("點了清單之後在 %v", g.State())
	}
	matches(t, g, "fcmd-battle-win")
	g.Click(490, 345) // 「作戰」
	if g.State() != StateFleetActionMenu {
		t.Fatalf("點了作戰之後在 %v", g.State())
	}
	if on := g.fleetMenuEnabled(); !on[0] || on[1] || on[2] || on[3] || on[4] {
		t.Fatalf("有戰鬥的星域啟用表 %v，原版只剩「移動」", on)
	}
	matches(t, g, "fcmd-battle-menu")
	g.MoveMouse(497, 312)
	matches(t, g, "fcmd-battle-menu-hover")
	g.Click(497, 352) // 停用的「駐留」
	if g.State() != StateFleetWindow {
		t.Fatalf("點了停用的駐留之後在 %v，原版只關選單", g.State())
	}
	if f := g.fleets[0]; f.Target != rules.NoTarget || f.Move != rules.MoveParked {
		t.Fatalf("點了停用的駐留之後艦隊記錄變了：%+v", f)
	}
	matches(t, g, "fcmd-battle-menu-click")
}

// 艦隊 `+0x00` bit0：占領與輸送停用（`fcmd-flag1-menu`，從 `fcmd-win` 把那一個
// byte 改成 01 再點作戰）；點停用的「占領」只關選單（`fcmd-flag1-menu-click`）。
func TestFleetMenuDisabledByFleetFlag(t *testing.T) {
	g := reachFleetWindow(t)
	g.fleets[0].Flags |= 1
	g.Click(490, 345) // 「作戰」
	if g.State() != StateFleetActionMenu {
		t.Fatalf("點了作戰之後在 %v", g.State())
	}
	if on := g.fleetMenuEnabled(); !on[0] || on[1] || !on[2] || on[3] || !on[4] {
		t.Fatalf("+0x00 bit0 的啟用表 %v，原版關掉占領與輸送", on)
	}
	matches(t, g, "fcmd-flag1-menu")
	g.Click(497, 332) // 停用的「占領」
	if g.State() != StateFleetWindow {
		t.Fatalf("點了停用的占領之後在 %v", g.State())
	}
	matches(t, g, "fcmd-flag1-menu-click")
}
