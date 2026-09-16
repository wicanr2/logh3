package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 艦隊組成編輯器（`docs/spec/015` 「編輯」一節、`docs/re/317`）。
// 收據 `fcmd-edit-*` 從「變更 → 編組」展開、`fcmd-fok-*` 從出擊命令那條路
// 展開，兩條路同一支編輯器（`tools/fleetcmd_receipts.sh` 第 13、14 段）。
// 座標是畫面座標（收據旗標的滑鼠 y ＋ 40）：隊形槽 (欄,列) 的中心在
// (144 ＋ 16 欄, 176 ＋ 16 列)、艦種列 i 在 (360, 170 ＋ 20 i)。

// reachFleetChangeOrg 走到「變更 → 編組」的艦隊組成（收據 `fcmd-change-org`）。
func reachFleetChangeOrg(t *testing.T) *Game {
	t.Helper()
	g := reachFleetChangeMenu(t)
	g.Click(497, 392) // 「編組」
	if g.State() != StateFleetChangeOrg {
		t.Fatalf("點了編組之後在 %v", g.State())
	}
	return g
}

// capitalStock 回首都那一格帝國七種艦艇的存量。
func capitalStock(t *testing.T, g *Game) []int {
	t.Helper()
	rec := g.regionRecord(rules.CapitalRegion(g.faction))
	if rec == nil {
		t.Fatal("沒有首都的星域記錄")
	}
	return rules.RegionShips(rec, g.faction)
}

// wantStock 釘住存量：與收據 `-peek lin:5F2FC:28` 倒出來的七個數比。
func wantStock(t *testing.T, g *Game, want [7]int) {
	t.Helper()
	got := capitalStock(t, g)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("首都存量 %v，原版 %v", got, want)
		}
	}
}

// receiptGroups 讀收據的 `fleet0.bin`：`+0x0F` 與 `+0x10` 起每組的 (艦種, 隊形槽)。
func receiptGroups(t *testing.T, batch string) [][2]int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "fleet0.bin"))
	if err != nil {
		t.Skipf("沒有收據 %s 的 fleet0.bin，跳過", batch)
	}
	n := int(b[0x0F])
	out := make([][2]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, [2]int{int(b[0x12+i*4]), int(b[0x13+i*4])})
	}
	return out
}

// wantComposition 把一支艦隊的編制與收據的艦隊記錄比：組數、每組的艦種與隊形槽、
// 順序（旗艦第一，其餘艦種再隊形槽升冪，`0FF4:0F6F`）。
func wantComposition(t *testing.T, f rules.Fleet, batch string) {
	t.Helper()
	want := receiptGroups(t, batch)
	if f.Groups != len(want) || len(f.Composition) != len(want) {
		t.Fatalf("%s：編制 %d 組（Groups %d），原版 +0x0F ＝ %d", batch, len(f.Composition), f.Groups, len(want))
	}
	for i, w := range want {
		if g := f.Composition[i]; g.Slot2 != w[0] || g.Slot3 != w[1] || g.Kind != w[0] {
			t.Fatalf("%s：第 %d 組 (艦種 %d, 槽 %d)，原版 (%d, %d)", batch, i, g.Slot2, g.Slot3, w[0], w[1])
		}
	}
}

// 點格盤放一組：戰艦列反白 → 槽 19 放一組戰艦，存量 −3000、部隊數 2/8、
// 機動力 98％（`fcmd-edit-place1`）；換到巡航艦列（`fcmd-edit-row2`）再放槽 21
// （`fcmd-edit-place2`：3/8、95％）。艦隊記錄要到「決定」才動。
func TestFleetEditPlaceMatchesOracle(t *testing.T) {
	g := reachFleetChangeOrg(t)
	g.Click(176, 240)
	matches(t, g, "fcmd-edit-place1")
	wantStock(t, g, [7]int{72000, 36000, 108000, 10000, 27000, 15000, 5000})
	if len(g.fleets[0].Composition) != 1 {
		t.Fatalf("放了一組艦隊記錄就變了：%d 組", len(g.fleets[0].Composition))
	}
	g.Click(360, 210)
	matches(t, g, "fcmd-edit-row2")
	g.Click(240, 240)
	matches(t, g, "fcmd-edit-place2")
	wantStock(t, g, [7]int{72000, 36000, 104000, 10000, 27000, 15000, 5000})
}

// 點已經有人的格子：先拆再放（`0FF4:0C80`）——槽 19 的戰艦換成巡航艦，
// 戰艦退回 3000、巡航艦再扣 4000（`fcmd-edit-replace`）。
func TestFleetEditReplaceMatchesOracle(t *testing.T) {
	g := reachFleetEditPlace2(t)
	g.Click(176, 240)
	matches(t, g, "fcmd-edit-replace")
	wantStock(t, g, [7]int{75000, 36000, 100000, 10000, 27000, 15000, 5000})
}

// 右鍵拆掉一格（`0FF4:0DB1`）：槽 19 的戰艦退回存量（`fcmd-edit-remove`）；
// 右鍵中央 → 「無法撤離旗艦。」（`fcmd-edit-flag-rclick`）。
func TestFleetEditRemoveMatchesOracle(t *testing.T) {
	t.Run("remove", func(t *testing.T) {
		g := reachFleetEditPlace2(t)
		g.RightClick(176, 240)
		matches(t, g, "fcmd-edit-remove")
		wantStock(t, g, [7]int{75000, 36000, 104000, 10000, 27000, 15000, 5000})
	})
	t.Run("flagship", func(t *testing.T) {
		g := reachFleetEditPlace2(t)
		g.RightClick(208, 240)
		if !g.msgOn {
			t.Fatal("右鍵中央沒有跳訊息框")
		}
		matches(t, g, "fcmd-edit-flag-rclick")
		wantStock(t, g, [7]int{72000, 36000, 104000, 10000, 27000, 15000, 5000})
	})
}

// reachFleetEditPlace2 ＝ 收據 `fcmd-edit-place2`：槽 19 戰艦、槽 21 巡航艦、
// 巡航艦列反白。
func reachFleetEditPlace2(t *testing.T) *Game {
	t.Helper()
	g := reachFleetChangeOrg(t)
	g.Click(176, 240)
	g.Click(360, 210)
	g.Click(240, 240)
	return g
}

// 巡航艦放中央 → 「只有戰艦或者高速戰艦才能當做旗艦。」（`fcmd-edit-centre-msg`），
// 點「確認」關掉回編輯器（`fcmd-edit-centre-ok`）；存量不動。
func TestFleetEditFlagshipKindMessageMatchesOracle(t *testing.T) {
	g := reachFleetEditPlace2(t)
	g.Click(208, 240)
	if !g.msgOn {
		t.Fatal("巡航艦放中央沒有跳訊息框")
	}
	matches(t, g, "fcmd-edit-centre-msg")
	wantStock(t, g, [7]int{72000, 36000, 104000, 10000, 27000, 15000, 5000})
	g.Click(432, 263) // 「確認」
	if g.msgOn || g.State() != StateFleetChangeOrg {
		t.Fatalf("點了確認之後 msgOn=%v state=%v", g.msgOn, g.State())
	}
	matches(t, g, "fcmd-edit-centre-ok")
}

// 「中止」把存量倒回進來那一刻（`0FF4:1127`）、編制不動（`fcmd-edit-cancel`）；
// 「決定」把工作表壓實排序寫回艦隊記錄：`+0x0F` ＝ 3、(0,20)(0,19)(2,21)
// （`fcmd-edit-decide` 的 `fleet0.bin`），存量照扣。
func TestFleetEditCancelAndDecideMatchOracle(t *testing.T) {
	t.Run("cancel", func(t *testing.T) {
		g := reachFleetEditPlace2(t)
		g.Click(510, 320)
		if g.State() != StateFleetWindow {
			t.Fatalf("點了中止之後在 %v", g.State())
		}
		matches(t, g, "fcmd-edit-cancel")
		wantStock(t, g, [7]int{75000, 36000, 108000, 10000, 27000, 15000, 5000})
		wantComposition(t, g.fleets[0], "fcmd-edit-cancel")
	})
	t.Run("decide", func(t *testing.T) {
		g := reachFleetEditPlace2(t)
		g.Click(510, 340)
		if g.State() != StateFleetWindow {
			t.Fatalf("點了決定之後在 %v", g.State())
		}
		matches(t, g, "fcmd-edit-decide")
		wantStock(t, g, [7]int{72000, 36000, 104000, 10000, 27000, 15000, 5000})
		wantComposition(t, g.fleets[0], "fcmd-edit-decide")
	})
}

// 「陣形」開八列清單（`fcmd-edit-formation`），反白跟游標（`-form-hover`），
// 右鍵只關清單（`-form-rclick`）；選「圓錐」整批換（`0FF4:0A4B`）：旗艦以外
// 全拆、照樣板放到 8/8、機動力 81％（`-form-pick`）；再點空格 → 「無法超過
// 目前統率能力以上的船艦數量。」（`-full-msg`）；「決定」寫回八組（`-form-decide`）。
func TestFleetEditFormationMatchesOracle(t *testing.T) {
	// 收據 `fcmd-edit-formation`：從 `-remove`（槽 21 巡航艦、旗艦）按「陣形」。
	reachMenu := func(t *testing.T) *Game {
		t.Helper()
		g := reachFleetEditPlace2(t)
		g.RightClick(176, 240)
		g.Click(510, 310) // 「陣形」
		if !g.fcmd.formOn {
			t.Fatal("點了陣形沒有開清單")
		}
		return g
	}
	g := reachMenu(t)
	matches(t, g, "fcmd-edit-formation")
	g.MoveMouse(505, 390)
	matches(t, g, "fcmd-edit-form-hover")
	t.Run("rclick", func(t *testing.T) {
		g := reachMenu(t)
		g.RightClick(300, 240)
		if g.fcmd.formOn {
			t.Fatal("右鍵沒有關清單")
		}
		matches(t, g, "fcmd-edit-form-rclick")
	})
	g.Click(505, 250) // 「圓錐」
	if g.fcmd.formOn {
		t.Fatal("選了陣形清單沒關")
	}
	matches(t, g, "fcmd-edit-form-pick")
	wantStock(t, g, [7]int{69000, 36000, 100000, 10000, 21000, 12000, 5000})
	t.Run("full", func(t *testing.T) {
		g := reachMenu(t)
		g.Click(505, 250)
		g.Click(144, 176) // 槽 0，8/8 之後
		if !g.msgOn {
			t.Fatal("滿了再放沒有跳訊息框")
		}
		matches(t, g, "fcmd-edit-full-msg")
	})
	g.Click(510, 340) // 「決定」
	matches(t, g, "fcmd-edit-form-decide")
	wantComposition(t, g.fleets[0], "fcmd-edit-form-decide")
}

// 運補艦存量改成 0 再開編組（收據用 `-poke`）：選運補艦列再放 → 「船艦數量不足。」
// （`fcmd-edit-poor-msg`）；陣形 → 圓錐時 (5,29) 那一組照替代表換成修護艦
// （`ds:2986` 第一個有貨的，`fcmd-edit-poor-pick`）。
func TestFleetEditPoorStockMatchesOracle(t *testing.T) {
	reach := func(t *testing.T) *Game {
		t.Helper()
		g := reachFleetChangeMenu(t)
		rec := g.regionRecord(rules.CapitalRegion(g.faction))
		binary.LittleEndian.PutUint32(rec[0x30+5*4:], 0) // 運補艦 ← 0（-poke lin:5F310）
		g.Click(497, 392)                                // 「編組」
		return g
	}
	g := reach(t)
	matches(t, g, "fcmd-edit-poor")
	t.Run("msg", func(t *testing.T) {
		g := reach(t)
		g.Click(360, 270) // 運補艦列
		matches(t, g, "fcmd-edit-poor-row5")
		g.Click(176, 240)
		if !g.msgOn {
			t.Fatal("沒貨的艦種放下去沒有跳訊息框")
		}
		matches(t, g, "fcmd-edit-poor-msg")
	})
	g.Click(510, 310)
	matches(t, g, "fcmd-edit-poor-form")
	g.Click(505, 250) // 「圓錐」
	matches(t, g, "fcmd-edit-poor-pick")
	wantStock(t, g, [7]int{69000, 36000, 100000, 10000, 21000, 0, 4000})
}

// 出擊命令那條路（艦隊編成按「確認」）是同一支編輯器：放槽 19、換列、放槽 21
// 與陣形 → 圓錐，畫面與存量都與「編組」那條一樣（`fcmd-fok-*`）；「決定」建出
// 三組的艦隊回戰略地圖（`fcmd-fok-decide` 的 `fleet0.bin`）。
func TestFleetEditSortiePathMatchesOracle(t *testing.T) {
	reach := func(t *testing.T) *Game {
		t.Helper()
		g := reachFleetOrgFrom(t, reachCommandPhaseOnMapWith(t, loadWithRules(t)))
		g.Click(480, 350) // 「確認」→ 艦隊組成
		if g.State() != StateFleetComposition {
			t.Fatalf("點了確認之後在 %v", g.State())
		}
		return g
	}
	g := reach(t)
	matches(t, g, "fcmd-fok")
	g.Click(176, 240)
	matches(t, g, "fcmd-fok-place1")
	wantStock(t, g, [7]int{72000, 36000, 108000, 10000, 27000, 15000, 5000})
	g.Click(360, 210)
	matches(t, g, "fcmd-fok-row2")
	t.Run("formation", func(t *testing.T) {
		g := reach(t)
		g.Click(176, 240)
		g.Click(360, 210)
		g.Click(510, 310)
		if g.State() != StateFormation {
			t.Fatalf("點了陣形之後在 %v", g.State())
		}
		matches(t, g, "fcmd-fok-form")
		g.Click(505, 250) // 「圓錐」
		if g.State() != StateFleetComposition {
			t.Fatalf("選了陣形之後在 %v", g.State())
		}
		matches(t, g, "fcmd-fok-form-pick")
		wantStock(t, g, [7]int{69000, 36000, 100000, 10000, 21000, 12000, 5000})
	})
	g.Click(240, 240)
	matches(t, g, "fcmd-fok-place2")
	g.Click(510, 340) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	matches(t, g, "fcmd-fok-decide")
	wantStock(t, g, [7]int{72000, 36000, 104000, 10000, 27000, 15000, 5000})
	wantComposition(t, g.fleets[len(g.fleets)-1], "fcmd-fok-decide")
}
