package game

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 艦隊編成視窗「副官」→ 變更／卸任（`docs/re/318`、規格 013「副官」一節）。
// 收據 `fcmd-adj-*` 從「變更 → 副官」那張展開、`fcmd-fadj-*` 從出擊命令那條路
// 展開（`tools/fleetcmd_receipts.sh` 第 15、16 段）。座標是畫面座標（滑鼠 y ＋ 40）：
// 小選單 變更 (430,350)、卸任 (430,370)；清單第 i 列 (450, 102 ＋ 20 i)；
// 詳細資料 中止 (448,320)、出擊(決定) (496,320)；結束 (480,340)。

// reachFleetChangeAdj 走到「變更 → 副官」那張視窗（收據 `fcmd-change-adj`）。
func reachFleetChangeAdj(t *testing.T) *Game {
	t.Helper()
	g := reachFleetChangeMenu(t)
	g.Click(497, 372) // 「副官」
	if g.State() != StateFleetChangeAdj {
		t.Fatalf("點了副官之後在 %v", g.State())
	}
	return g
}

// receiptAdmiral 讀收據裡 `-peek` 倒出來的提督記錄（0／2／32 三筆），回第 i 筆。
func receiptAdmiral(t *testing.T, batch string, i int) []byte {
	t.Helper()
	r, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "report.txt"))
	if err != nil {
		t.Skipf("沒有收據 %s，跳過", batch)
	}
	want := map[int]string{0: "lin:602AC", 2: "lin:602E4", 32: "lin:6062C"}[i]
	for _, line := range strings.Split(string(r), "\n") {
		if !strings.Contains(line, want) {
			continue
		}
		// 「  lin:602AC     （執行期 602AC）FF 00 …」：取括號之後的 hex。
		hexPart := line[strings.LastIndex(line, "）")+len("）"):]
		out, err := hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(hexPart), " ", ""))
		if err != nil {
			t.Fatalf("收據 %s 的 %s 讀不成 hex：%v", batch, want, err)
		}
		return out
	}
	t.Fatalf("收據 %s 的 report.txt 沒有 %s", batch, want)
	return nil
}

// wantAdmiralFields 釘住提督記錄的 `+0x07`（副官）與 `+0x09`（艦隊槽／閒置）
// 與收據相同。
func wantAdmiralFields(t *testing.T, g *Game, batch string, indexes ...int) {
	t.Helper()
	for _, i := range indexes {
		want := receiptAdmiral(t, batch, i)
		got := g.admiralRecord(i)
		if len(want) < 10 || got == nil {
			t.Fatalf("%s：提督 %d 的記錄讀不到", batch, i)
		}
		if got[7] != want[7] || got[9] != want[9] {
			t.Fatalf("%s：提督 %d 的 +07/+09 ＝ %02x/%02x，原版 %02x/%02x", batch, i, got[7], got[9], want[7], want[9])
		}
	}
}

// 「副官」開 變更／卸任 小選單（`fcmd-adj-menu`：游標在第一項反白），移到「卸任」
// 反白跟著走（`-hover`），右鍵只關選單（`-rclick`）。
func TestFleetChangeAdjutantMenuMatchesOracle(t *testing.T) {
	g := reachFleetChangeAdj(t)
	g.Click(430, 340) // 「副官」
	if !g.adj.menuOn {
		t.Fatal("點了副官沒有開小選單")
	}
	matches(t, g, "fcmd-adj-menu")
	g.MoveMouse(430, 370)
	matches(t, g, "fcmd-adj-menu-hover")
	g.RightClick(300, 240)
	if g.adj.menuOn || g.State() != StateFleetChangeAdj {
		t.Fatalf("右鍵之後 menuOn=%v state=%v", g.adj.menuOn, g.State())
	}
	matches(t, g, "fcmd-adj-menu-rclick")
}

// 「卸任」：提督記錄 +0x07 ← 0xFF、舊副官 +0x09 ← 0x80，視窗上副官那三格與頭像
// 清掉、能力只看提督（`fcmd-adj-dismiss`）；「結束」抄回艦隊記錄（`+0x05` ＝ 0xFF、
// 營運 100、維持費 30），清單那一列不再反白（`fcmd-adj-dismiss-close`）。
func TestFleetAdjutantDismissMatchesOracle(t *testing.T) {
	g := reachFleetChangeAdj(t)
	g.Click(430, 340)
	g.Click(430, 370) // 「卸任」
	if g.adjOpen() {
		t.Fatal("卸任之後還有東西開著")
	}
	matches(t, g, "fcmd-adj-dismiss")
	wantAdmiralFields(t, g, "fcmd-adj-dismiss", 0, 32)
	if f := g.fleets[0]; f.Adjutant != 32 {
		t.Fatalf("關窗之前艦隊記錄就變了：副官 %d", f.Adjutant)
	}
	g.Click(480, 340) // 「結束」
	if g.State() != StateFleetWindow {
		t.Fatalf("點了結束之後在 %v", g.State())
	}
	matches(t, g, "fcmd-adj-dismiss-close")
	if f := g.fleets[0]; f.Adjutant != -1 || f.Operations != 100 {
		t.Fatalf("結束之後艦隊 副官 %d 營運 %d，原版 -1／100", f.Adjutant, f.Operations)
	}
}

// 「變更」開「待命中副官資料」（`fcmd-adj-change`：同星域、沒有職務、閒置、等級較低的
// 十六列），游標移動不反白（`-hover`），右鍵回視窗（`-rclick`）；點第 1 列 米達麥亞 →
// 詳細資料（`-pick`）：右鍵或中止回視窗不換（`-pick-rclick`／`-pick-cancel`），
// 出擊(決定) 換人（`-pick-confirm`：+0x07 ＝ 2、米達麥亞 +0x09 ＝ 0、修特萊 0x80）；
// 「結束」抄回艦隊（`-confirm-close`：副官 2、營運 102、維持費 29）。
func TestFleetAdjutantChangeMatchesOracle(t *testing.T) {
	reachList := func(t *testing.T) *Game {
		t.Helper()
		g := reachFleetChangeAdj(t)
		g.Click(430, 340)
		g.Click(430, 350) // 「變更」
		if !g.adj.listOn {
			t.Fatal("點了變更沒有開清單")
		}
		return g
	}
	g := reachList(t)
	want := []int{1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 14, 15, 17, 18, 19}
	for i, w := range want {
		if i >= len(g.adj.list) || g.adj.list[i] != w {
			t.Fatalf("候選 %v，原版前十六列 %v", g.adj.list, want)
		}
	}
	matches(t, g, "fcmd-adj-change")
	g.MoveMouse(450, 190)
	matches(t, g, "fcmd-adj-change-hover")
	t.Run("rclick", func(t *testing.T) {
		// 右鍵關清單回視窗（同一個清單元件在出擊命令那條路就是這樣關的）。
		// 收據 `fcmd-adj-change-rclick` 裡原版的清單留在畫面上、之後也不再反應
		// （成因沒讀，規格 013 已知差異），所以這一批不比畫面。
		g := reachList(t)
		g.RightClick(300, 240)
		if g.adj.listOn || g.State() != StateFleetChangeAdj {
			t.Fatalf("右鍵之後 listOn=%v state=%v", g.adj.listOn, g.State())
		}
	})
	t.Run("scroll", func(t *testing.T) {
		// ▼ 一次一頁、到底停在最後十六列；▲ 回上一頁；第二頁點第 1 列是瑞文茲（21）。
		g := reachList(t)
		g.Click(617, 390)
		if g.adj.first != 16 {
			t.Fatalf("▼ 之後 first=%d", g.adj.first)
		}
		matches(t, g, "fcmd-adj-change-down")
		g2 := reachList(t)
		g2.Click(617, 390)
		g2.Click(617, 390)
		if g2.adj.first != len(g2.adj.list)-16 {
			t.Fatalf("▼▼ 之後 first=%d（候選 %d）", g2.adj.first, len(g2.adj.list))
		}
		matches(t, g2, "fcmd-adj-change-down2")
		g3 := reachList(t)
		g3.Click(617, 390)
		g3.Click(617, 90)
		if g3.adj.first != 0 {
			t.Fatalf("▼▲ 之後 first=%d", g3.adj.first)
		}
		matches(t, g3, "fcmd-adj-change-up")
		g.Click(450, 110)
		if !g.adj.detailOn || g.adj.detail != 21 {
			t.Fatalf("第二頁點第 1 列 detailOn=%v detail=%d", g.adj.detailOn, g.adj.detail)
		}
		matches(t, g, "fcmd-adj-change-pick2")
	})
	reachPick := func(t *testing.T) *Game {
		t.Helper()
		g := reachList(t)
		g.Click(450, 110) // 第 1 列 米達麥亞
		if !g.adj.detailOn || g.adj.detail != 2 || g.adj.listOn {
			t.Fatalf("點了一列之後 detailOn=%v detail=%d listOn=%v", g.adj.detailOn, g.adj.detail, g.adj.listOn)
		}
		return g
	}
	g = reachPick(t)
	matches(t, g, "fcmd-adj-change-pick")
	t.Run("pick-rclick", func(t *testing.T) {
		g := reachPick(t)
		g.RightClick(300, 240)
		if g.adjOpen() {
			t.Fatal("右鍵之後還有東西開著")
		}
		matches(t, g, "fcmd-adj-pick-rclick")
		wantAdmiralFields(t, g, "fcmd-adj-pick-rclick", 0, 2, 32)
	})
	t.Run("pick-cancel", func(t *testing.T) {
		g := reachPick(t)
		g.Click(448, 320) // 「中止」
		matches(t, g, "fcmd-adj-pick-cancel")
		wantAdmiralFields(t, g, "fcmd-adj-pick-cancel", 0, 2, 32)
	})
	g.Click(496, 320) // 「出擊」＝ 決定
	if g.adjOpen() {
		t.Fatal("決定之後還有東西開著")
	}
	matches(t, g, "fcmd-adj-pick-confirm")
	wantAdmiralFields(t, g, "fcmd-adj-pick-confirm", 0, 2, 32)
	g.Click(480, 340) // 「結束」
	matches(t, g, "fcmd-adj-confirm-close")
	if f := g.fleets[0]; f.Adjutant != 2 || f.Operations != 102 {
		t.Fatalf("結束之後艦隊 副官 %d 營運 %d，原版 2／102", f.Adjutant, f.Operations)
	}
}

// 出擊命令那條路的同一張小選單（`fcmd-fadj`：艦隊編成按副官）：卸任
// （`-dismiss`）、變更 → 清單（`-change`）→ 米達麥亞（`-change-pick`）→ 出擊(決定)
// 回艦隊編成（`-pick-confirm`）→ 確認 → 決定 建出副官是米達麥亞的艦隊
// （`-confirm-fleet`：`+0x05` ＝ 2、營運 102）。
func TestFleetAdjutantSortiePathMatchesOracle(t *testing.T) {
	reachMenu := func(t *testing.T) *Game {
		t.Helper()
		g := reachFleetOrgFrom(t, reachCommandPhaseOnMapWith(t, loadWithRules(t)))
		g.Click(432, 350) // 「副官」
		if g.State() != StateFleetAdjutant {
			t.Fatalf("點了副官之後在 %v", g.State())
		}
		return g
	}
	g := reachMenu(t)
	matches(t, g, "fcmd-fadj")
	t.Run("dismiss", func(t *testing.T) {
		g := reachMenu(t)
		g.Click(430, 370)
		if g.State() != StateFleetOrg || g.adjOpen() {
			t.Fatalf("卸任之後 state=%v open=%v", g.State(), g.adjOpen())
		}
		matches(t, g, "fcmd-fadj-dismiss")
		wantAdmiralFields(t, g, "fcmd-fadj-dismiss", 0, 32)
	})
	g.Click(430, 350) // 「變更」
	if g.State() != StateFleetOrg || !g.adj.listOn {
		t.Fatalf("變更之後 state=%v listOn=%v", g.State(), g.adj.listOn)
	}
	matches(t, g, "fcmd-fadj-change")
	g.Click(450, 110)
	matches(t, g, "fcmd-fadj-change-pick")
	g.Click(496, 320) // 「出擊」＝ 決定
	matches(t, g, "fcmd-fadj-pick-confirm")
	wantAdmiralFields(t, g, "fcmd-fadj-pick-confirm", 0, 2, 32)
	g.Click(480, 350) // 「確認」→ 艦隊組成
	g.Click(510, 350) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	matches(t, g, "fcmd-fadj-confirm-fleet")
	f := g.fleets[len(g.fleets)-1]
	if f.Adjutant != 2 || f.Operations != 102 {
		t.Fatalf("建出來的艦隊 副官 %d 營運 %d，原版 2／102", f.Adjutant, f.Operations)
	}
	wantAdmiralFields(t, g, "fcmd-fadj-confirm-fleet", 0, 2, 32)
}
