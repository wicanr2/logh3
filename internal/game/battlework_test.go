package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 這些是原指令契約的邊界回歸；正常完整WORK對拍另在fleetarrival_test。
func TestFleetExpensesMovementStatesAndSignedWords(t *testing.T) {
	g := newTestGame(t)
	kinds, err := rules.LoadShipKindTable(filepath.Join(root, "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.ShipKinds = kinds
	f := g.fleets[0]
	r := g.slot.Fleet[f.Index*rules.FleetRecord:]
	upkeep := int(int16(binary.LittleEndian.Uint16(r[0x0A:])))
	moveCost := int(int16(binary.LittleEndian.Uint16(r[0x0C:])))
	if upkeep == 0 || moveCost == 0 {
		t.Fatal("原版載具費用為0，不能驗證排除規則")
	}
	for state := 0; state <= 5; state++ {
		t.Run(fmt.Sprintf("state%d", state), func(t *testing.T) {
			f.Move, f.Parked = state, state == rules.MoveParked
			g.fleets = []rules.Fleet{f}
			want := moveCost
			if state == rules.MoveGarrison || state == rules.MoveParked {
				want = 0
			}
			if g.fleetUpkeep(f.Faction) != upkeep || g.fleetMoveCost(f.Faction) != want ||
				g.fleetUpkeep(1-f.Faction) != 0 || g.fleetMoveCost(1-f.Faction) != 0 {
				t.Fatal("原版word費用、狀態2／5排除或陣營集合不同")
			}
			g.refreshFleetExpenses(f.Faction)
			work, err := rules.ReadWork(g.workRecord())
			if err != nil || work.Upkeep[f.Faction] != upkeep || work.MoveCost[f.Faction] != want {
				t.Fatal("經費重算沒有保存到原版WORK欄位")
			}
		})
	}
	// 明示數值邊界：三組各65535艘，兩費率255、營運0，公式100268。
	// 低word為34732，原版CWD讀回-30804；不是自然局面的原版收據。
	f.Faction, f.Operations, f.Move, f.Parked = 0, 0, rules.MoveMove, false
	f.Composition = []rules.FleetGroup{{Kind: 0, Ships: 65535}, {Kind: 0, Ships: 65535}, {Kind: 0, Ships: 65535}}
	kinds.Factions[0].Kinds[0].Words[2] = 0xFFFF
	g.fleets = []rules.Fleet{f}
	if g.fleetUpkeep(0) != -30804 || g.fleetMoveCost(0) != -30804 {
		t.Fatal("直接合計公式大整數，沒有經原版signed word")
	}
	g.refreshFleetExpenses(0)
	w, err := rules.ReadWork(g.workRecord())
	if err != nil || w.Upkeep[0] != -30804 || w.MoveCost[0] != -30804 {
		t.Fatal("WORK的signed 32位元經費讀回錯誤")
	}
	g.slot.Upkeep, g.slot.MoveCost = w.Upkeep, w.MoveCost
	dir := t.TempDir()
	if err := g.slot.Save(dir, 0); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
	if err != nil {
		t.Fatal(err)
	}
	readBack, err := rules.ReadWork(saved)
	if err != nil || readBack.Upkeep != w.Upkeep || readBack.MoveCost != w.MoveCost {
		t.Fatal("Slot.Save没有保存signed經費或ReadWork未讀回")
	}
	before := bytes.Clone(g.workRecord())
	g.res.ShipKinds = nil
	g.refreshFleetExpenses(0)
	g.refreshFleetExpenses(-1)
	if !bytes.Equal(before, g.workRecord()) || rules.WriteWorkFleetExpenses(make([]byte, 10), 0, 1, 2) ||
		rules.WriteWorkFleetExpenses(g.mutableWork(), 2, 1, 2) {
		t.Fatal("缺資料／非法陣營改寫未知WORK")
	}
}

func TestBattleWorkSeatCamerasAndInputBoundary(t *testing.T) {
	m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal89-r55-work-a/entry.bin"))
	if err != nil {
		t.Skipf("缺第55輪原版正常RAM：%v", err)
	}
	if fmtHash(sha256.Sum256(m)) != "ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97" {
		t.Fatal("原版RAM身份不同")
	}
	for faction := 0; faction < 2; faction++ {
		t.Run(fmt.Sprintf("faction%d", faction), func(t *testing.T) {
			g := newTestGame(t)
			g.faction, g.battleRegion, g.battleTurn = faction, 26, 1
			g.slot.Fleet = bytes.Clone(m[0x60BA4 : 0x60BA4+rules.FleetCount*rules.FleetRecord])
			g.battleFleets = g.slot.BattleFleetsFor([]int{12, 16, 17, 28, 29, 30, 31})
			w := g.mutableWork()
			w[0x11], w[0x12] = byte(faction), byte(1-faction)
			g.setHumanSeat()
			g.initBattleWork()
			for seat := 0; seat < 2; seat++ {
				o, orig := battleWorkView+4*seat, 0x5DA44+battleWorkView+4*int(w[0x11+seat])
				if !bytes.Equal(g.workRecord()[o:o+4], m[orig:orig+4]) {
					t.Fatal("把座位鏡頭當成固定陣營鏡頭")
				}
			}
			view := screen.BattleView{Col: 12, Row: 5}
			if faction == 1 {
				view = screen.BattleView{Col: 82, Row: 83}
			}
			if g.battleView != view || g.workRecord()[battleWorkActive] != 1 ||
				binary.LittleEndian.Uint16(g.workRecord()[battleWorkTurn:]) != 0 {
				t.Fatal("正式鏡頭／戰場旗標／零起算回合不同")
			}
			before := bytes.Clone(g.workRecord())
			g.state = StateBattle
			g.battleScroll(463, 240)
			if g.battleView == view || !bytes.Equal(before, g.workRecord()) {
				t.Fatal("純捲動未移動視窗，或提前覆寫Input返回前快取")
			}
			g.saveBattleWorkView()
			if binary.LittleEndian.Uint16(g.workRecord()[battleWorkView:]) != uint16(g.battleView.Col*16) ||
				!bytes.Equal(before[battleWorkView+4:battleWorkView+8], g.workRecord()[battleWorkView+4:battleWorkView+8]) {
				t.Fatal("Input返回時沒有只保存當前座位鏡頭")
			}
			g.battleTurn = 2
			g.updateBattleWorkTurn()
			if binary.LittleEndian.Uint16(g.workRecord()[battleWorkTurn:]) != 1 {
				t.Fatal("顯示回合沒有轉成零起算WORK")
			}
			g.clearBattleWork()
			if g.workRecord()[battleWorkActive] != 0 {
				t.Fatal("戰鬥結束仍保留旗標")
			}
			g.setScannedRegion(rules.RegionCount)
			if g.workRecord()[battleWorkRegion] != byte(rules.RegionCount) {
				t.Fatal("完成星域掃描沒有保存32")
			}
			before = bytes.Clone(g.workRecord())
			g.setScannedRegion(-1)
			g.setScannedRegion(33)
			g.saveBattleWorkView()
			if !bytes.Equal(before, g.workRecord()) {
				t.Fatal("非法星域或非戰場改寫WORK")
			}
			// 不模仿原版表尾垃圾值：沒有參戰者或映射非法時保留鏡頭。
			g.battleFleets = nil
			g.initBattleWork()
			if !bytes.Equal(before[battleWorkView:battleWorkView+8], g.workRecord()[battleWorkView:battleWorkView+8]) {
				t.Fatal("没有匹配參戰艦隊卻捏造鏡頭")
			}
			g.mutableWork()[0x11], g.mutableWork()[0x12] = 0xFF, 0xFF
			g.initBattleWork()
			g.saveBattleWorkView()
			if !bytes.Equal(before[battleWorkView:battleWorkView+8], g.workRecord()[battleWorkView:battleWorkView+8]) {
				t.Fatal("非法座位映射改寫鏡頭")
			}
			g.work = make([]byte, 10)
			g.initBattleWork()
			g.saveBattleWorkView()
			g.updateBattleWorkTurn()
			if !bytes.Equal(g.work, make([]byte, 10)) {
				t.Fatal("不完整WORK被控制更新覆寫")
			}
		})
	}
}
