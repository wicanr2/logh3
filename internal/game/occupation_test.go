package game

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 占領命令走到之後（`docs/re/322`、規格 054）：與 `tools/fleetcmd_receipts.sh`
// 第 19 段的 `fcmd-occ-*` 收據對。每回合按 系統指令 (544,125) → 實行 (560,125)，
// 收據的游標停在 (560,125)。

// reachOccupyAstarte 走到 `fcmd-occupy-done`：第１艦隊在首都、占領亞斯提（16）。
func reachOccupyAstarte(t *testing.T) *Game {
	t.Helper()
	g := reachFleetOccupyPick(t)
	g.Click(193, 245) // 亞斯提
	g.Click(432, 263) // 決定
	if g.State() != StateCommandPhase || g.fleets[0].Target != 16 {
		t.Fatalf("到不了 fcmd-occupy-done：%v、目的地 %d", g.State(), g.fleets[0].Target)
	}
	return g
}

// executeTurn 按一次「實行」，把回合裡跳出來的過場視窗與模態報告（占據／叛亂／士氣）
// 用同一個落點關掉，回到命令階段。
func executeTurn(t *testing.T, g *Game) {
	t.Helper()
	g.Click(544, 125)
	g.Click(560, 125)
	finishTurn(t, g)
}

// finishTurn 從回合中途停下來的地方（過場視窗或報告）一路點到回合跑完。
func finishTurn(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		switch g.State() {
		case StateOccupationReport, StateRebellionReport, StateRegionFlipReport, StateTurnWindow, StateTurnReport:
			g.Click(560, 125)
		default:
			return
		}
	}
}

// 亞斯提那條路：0 → 2 → 7 → 8 → 13 一回合一格（路徑表，全在自己的星域），
// 第五回合走進亞斯提（同盟、目的地）→ 停駐、`+0x07 ← FF`；同一回合星域 16
// 設 bit4（單方面占據）、跳「亞斯提星系已被攻入 目前轉移至占領狀態」；
// 收據 `fcmd-occ-t1`–`t5`。畫面前兩回合整張比（第三回合起 remake 的叛亂在
// 別的格子分岔），之後只比艦隊記錄與星域旗標。
func TestOccupyAstarteArrivalMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	want := []int{2, 7, 8, 13, 16}
	for turn, region := range want {
		executeTurn(t, g)
		f := g.fleets[0]
		if f.Region != region {
			t.Fatalf("第 %d 回合之後艦隊在星域 %d，原版 %d（`fcmd-occ-t%d`）", turn+1, f.Region, region, turn+1)
		}
		if turn+1 < len(want) {
			if f.Move != rules.MoveOccupy || f.Target != 16 || f.Parked {
				t.Fatalf("第 %d 回合之後艦隊記錄 %+v，原版 +0x07=16、+0x08=1", turn+1, f)
			}
			// 第三回合起 remake 的叛亂判定在別的格子分岔（電腦那一方不下命令，
			// 亂數序列與原版對不齊；`docs/spec/046` 的證據限制），畫面只比前兩回合。
			if turn < 2 {
				matches(t, g, fmt.Sprintf("fcmd-occ-t%d", turn+1))
			}
			continue
		}
		if !f.Parked || f.Move != rules.MoveParked {
			t.Fatalf("到了亞斯提之後艦隊記錄 %+v，原版 +0x08=5", f)
		}
		if rec := g.regionRecord(16); rec == nil || rec[regionFlagsOff]&0x10 == 0 || rec[0x28] != 1 {
			t.Fatalf("到了亞斯提之後星域 16 的旗標 %02X／帝國艦隊數 %d，原版 18／1", rec[regionFlagsOff], rec[0x28])
		}
	}
}

// 費沙那條路：0 → 2 → 6 → 14 三回合到達（`fcmd-occ-phz-t1`–`t3`）。
func TestOccupyPhezzanArrivalMatchesOracle(t *testing.T) {
	g := reachFleetOccupyPick(t)
	g.Click(163, 136) // 費沙
	g.Click(300, 240)
	g.Click(300, 240)
	g.Click(432, 263) // 決定
	if g.State() != StateCommandPhase || g.fleets[0].Target != rules.PhezzanRegion {
		t.Fatalf("到不了 fcmd-occ-phz-done：%v", g.State())
	}
	for turn, region := range []int{2, 6, 14} {
		executeTurn(t, g)
		if f := g.fleets[0]; f.Region != region {
			t.Fatalf("第 %d 回合之後艦隊在星域 %d，原版 %d（`fcmd-occ-phz-t%d`）", turn+1, f.Region, region, turn+1)
		}
	}
	f := g.fleets[0]
	if !f.Parked || f.Move != rules.MoveParked {
		t.Fatalf("到了費沙之後艦隊記錄 %+v，原版 +0x08=5", f)
	}
	if rec := g.regionRecord(14); rec == nil || rec[regionFlagsOff]&0x10 == 0 {
		t.Fatalf("到了費沙之後星域 14 的旗標 %02X，原版 18", rec[regionFlagsOff])
	}
}

// `113A:047D` 的四條分支（`docs/re/322`）：兩格都不是自己的 → 「無法前往敵方的
// 星系」清命令；下一格不是自己的且不是占領 → 同上；占領且下一格不是目的地 →
// 問話；經費不夠 → 不動。
func TestPlanStepBranches(t *testing.T) {
	g := reachOccupyAstarte(t)
	owner := g.res.Planets.RegionOwner
	f := g.fleets[0]
	f.Region, f.Target, f.Move, f.Parked = 13, 16, rules.MoveOccupy, false
	if p := g.res.Graph.PlanStep(&f, owner, 1000, 26); p.Outcome != rules.StepMoved || p.Next != 16 {
		t.Errorf("伊謝爾倫 → 亞斯提（目的地）：%+v，預期走進去", p)
	}
	f.Move = rules.MoveMove
	if p := g.res.Graph.PlanStep(&f, owner, 1000, 26); p.Outcome != rules.StepBlockedEnemy {
		t.Errorf("移動命令走進同盟星域：%+v，預期「無法前往敵方的星系」", p)
	}
	f.Move, f.Target = rules.MoveOccupy, 19 // 特利布拉：下一格亞斯提不是目的地
	if p := g.res.Graph.PlanStep(&f, owner, 1000, 26); p.Outcome != rules.StepAskOccupy || p.Next != 16 {
		t.Errorf("占領特利布拉、下一格亞斯提：%+v，預期問話", p)
	}
	f.Region, f.Target = 16, 19 // 已在同盟的亞斯提，下一格也不是自己的
	if p := g.res.Graph.PlanStep(&f, owner, 1000, 26); p.Outcome != rules.StepBlockedEnemy {
		t.Errorf("兩格都不是自己的：%+v，預期「無法前往敵方的星系」", p)
	}
	f.Region, f.Target = 0, 2
	if p := g.res.Graph.PlanStep(&f, owner, 10, 26); p.Outcome != rules.StepNoFunds {
		t.Errorf("軍事費 10 < 移動經費 26：%+v，預期不動", p)
	}
	g.fleets[0].Region, g.fleets[0].Target, g.fleets[0].Move, g.fleets[0].Parked = 13, 19, rules.MoveOccupy, false
	executeTurn(t, g)
	if g.State() != StateOccupyAsk || g.occupyAsk.next != 16 {
		t.Fatalf("回合裡該問「亞斯提要占領此星系嗎？」，在 %v（next %d）", g.State(), g.occupyAsk.next)
	}
	g.Click(480, 270) // 待命（內容格 (21,5)）
	if g.State() == StateOccupyAsk || g.fleets[0].Region != 13 || g.fleets[0].Target != 19 {
		t.Fatalf("按了待命之後 %v、艦隊 %+v，預期不動、命令留著", g.State(), g.fleets[0])
	}
}

// 占據那一格的戰略打行星接進回合（規格 058）：remake 的回合亂數第三回合起就與原版分岔，
// 同狀態的逐擊對拍在 `rules.TestStrategicAssaultMatchesOracle`（`fcmd-assault-*` 的
// mem-in／mem-out）。這裡釘三件事：行星座標與原版 `2AF5:00CB` 算的相同（mem-out 的
// `+0x46`／`+0x48`）、走進亞斯提那一回合巴溫真的被打（防衛力下降、艦隊士氣下降、提督有功績）、
// 同一套回合流程照常收尾。
func TestOccupyAstarteRunsStrategicAssault(t *testing.T) {
	out, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-assault-ast/mem-out.bin"))
	if err != nil || len(out) != 1<<20 {
		t.Skip("缺收據 fcmd-assault-ast 的 mem-out.bin")
	}
	g := reachOccupyAstarte(t)
	start, n := g.res.Planets.Range(16)
	sites := rules.PlanetSites(g.battleCells(16), g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
	if len(sites) < n {
		t.Fatalf("亞斯提的行星座標 %d 筆，行星 %d 顆", len(sites), n)
	}
	for i := 0; i < n; i++ {
		o := 0x53132 + (start+i)*rules.PlanetRecord
		x, y := int(binary.LittleEndian.Uint16(out[o+0x46:])), int(binary.LittleEndian.Uint16(out[o+0x48:]))
		if sites[i].X != x || sites[i].Y != y {
			t.Fatalf("行星 %d 座標 (%d,%d)，原版 (%d,%d)", start+i, sites[i].X, sites[i].Y, x, y)
		}
	}
	for turn := 0; turn < 5; turn++ {
		executeTurn(t, g)
	}
	baw := g.res.Planets.Record(start)
	if d := int(binary.LittleEndian.Uint16(baw[0x14:])); d >= 4200 || baw[0x11] != 1 {
		t.Fatalf("第五回合之後巴溫防衛力 %d、擁有者 %d，應該被打過（原版 3768）", d, baw[0x11])
	}
	f := g.fleets[0]
	if f.Morale >= 100 || f.Morale < 50 {
		t.Fatalf("第五回合之後艦隊士氣 %d，應該掉一些（原版 67）", f.Morale)
	}
	if m := binary.LittleEndian.Uint32(g.admiralRecord(f.Admiral)[2:]); m == 0 {
		t.Fatal("提督功績沒有加（原版 845）")
	}
}
