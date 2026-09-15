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
// （`loadWithRules`）：艦隊視窗印的編制、費用與速度都從艦種表算。
func reachFleetDecided(t *testing.T) *Game {
	t.Helper()
	g := reachCommandPhaseOnMapWith(t, loadWithRules(t))
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
// 「中止」回目的地選取（`fcmd-cancel`）。
//
// `fcmd-done` 那張整張比不了：清單第 0 列左端多一個移動圖示，那張圖與規則
// 還沒查（規格 051 的已知差異），所以比的是**那一列以外**的整張畫面。
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
	// 清單第 0 列 (336,300) 288×20 之外整張相同。
	matchesBox(t, g, "fcmd-done", 0, 0, 639, 299)
	matchesBox(t, g, "fcmd-done", 0, 320, 639, 479)
	matchesBox(t, g, "fcmd-done", 0, 300, 335, 319)
	matchesBox(t, g, "fcmd-done", 624, 300, 639, 319)
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
