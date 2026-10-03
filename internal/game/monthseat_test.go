package game

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func monthSeatFlags(g *Game) []byte {
	return g.workRecord()[workMonthSeatFlag : workMonthSeatFlag+2]
}

func TestMonthSeatNormalLifecycleAndSave(t *testing.T) {
	g := reachFleetDecided(t)
	for _, stage := range []string{"initial", "command"} {
		if stage == "command" {
			executeTurn(t, g)
		}
		m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal89-r48-clock-a", stage+".bin"))
		if err != nil {
			t.Skipf("缺原版正常月底旗標收據：%v", err)
		}
		if len(m) != 1048576 || !bytes.Equal(monthSeatFlags(g), m[0x5DB24:0x5DB26]) {
			t.Fatalf("正常%s旗標不同：%v", stage, monthSeatFlags(g))
		}
		dir := t.TempDir()
		if err := g.SaveSlot(dir, 0); err != nil {
			t.Fatal(err)
		}
		w, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
		if err != nil || !bytes.Equal(w[0xE0:0xE2], monthSeatFlags(g)) {
			t.Fatalf("正常%s旗標未保存：%v", stage, err)
		}
	}
}

// 正常UI載入S3後才注入單一旗標；邊界案例不冒充自然S3。
func TestMonthSeatLoadedBothAIModesAndFactions(t *testing.T) {
	for _, level := range []int{-1, 1} {
		for faction := 0; faction < 2; faction++ {
			for _, flag := range []byte{0, 1, 255} {
				t.Run(fmt.Sprintf("level%d/faction%d/flag%d", level, faction, flag), func(t *testing.T) {
					g := reachConfiguredAI(t, level, faction)
					g.Click(544, 125)
					g.Click(560, 145)
					g.Click(300, 322)
					g.Click(318, 208+20*faction)
					wantState(t, g, StateCommandPhase, "正常S3讀檔")
					if g.turn != 380 || !bytes.Equal(monthSeatFlags(g), []byte{0, 0}) {
						t.Fatal("S3原始回合／旗標被替換")
					}
					foe := 1 - faction
					g.mutableWork()[workMonthSeatFlag+foe] = flag
					money, building := g.military, g.building
					stellar, planets := bytes.Clone(g.res.Stellar), bytes.Clone(g.res.Planets.Data)
					g.beginTurn()
					unchanged := money == g.military && building == g.building &&
						bytes.Equal(stellar, g.res.Stellar) && bytes.Equal(planets, g.res.Planets.Data)
					if unchanged != (flag == 0) || !bytes.Equal(monthSeatFlags(g), []byte{0, 0}) {
						t.Fatal("未按原始非零閘門結算／清旗標")
					}
					money, building = g.military, g.building
					stellar, planets = bytes.Clone(g.res.Stellar), bytes.Clone(g.res.Planets.Data)
					g.beginTurn()
					if money != g.military || building != g.building ||
						!bytes.Equal(stellar, g.res.Stellar) || !bytes.Equal(planets, g.res.Planets.Data) {
						t.Fatal("同一回合重複收稅或造船")
					}
					dir := t.TempDir()
					if err := g.SaveSlot(dir, 0); err != nil {
						t.Fatal(err)
					}
					g.res.SaveDir = dir
					g.Click(544, 125)
					g.Click(560, 145)
					g.Click(300, 122)
					g.Click(318, 208+20*faction)
					wantState(t, g, StateCommandPhase, "旗標消費後正常存讀")
					if !bytes.Equal(monthSeatFlags(g), []byte{0, 0}) || g.AILevel() != level {
						t.Fatal("存讀恢復旗標或丟失AI模式")
					}
					if money != g.military || building != g.building {
						t.Fatal("正常存讀改變已結算財政／建造")
					}
					// 正常讀檔會重算未安定度及艦隊數。比較點取讀檔完成後，
					// 再確認第二次座位閘門不會重新扣行星經濟力。
					stellar, planets = bytes.Clone(g.res.Stellar), bytes.Clone(g.res.Planets.Data)
					g.beginTurn()
					if money != g.military || building != g.building ||
						!bytes.Equal(stellar, g.res.Stellar) || !bytes.Equal(planets, g.res.Planets.Data) {
						t.Fatal("存讀後重複收稅或造船")
					}
				})
			}
		}
	}
}

func TestMonthSeatNonMonthlyClearsWithoutSettlement(t *testing.T) {
	g := reachConfiguredAI(t, -1, 0)
	g.turn = 3 // 明示的mod 10邊界，不當作原版正常局面。
	g.mutableWork()[0xE1] = 255
	stellar, planets := bytes.Clone(g.res.Stellar), bytes.Clone(g.res.Planets.Data)
	money, building, paid := g.military, g.building, g.paidUpkeep
	g.beginTurn()
	if monthSeatFlags(g)[1] != 0 || money != g.military || building != g.building || paid != g.paidUpkeep ||
		!bytes.Equal(stellar, g.res.Stellar) || !bytes.Equal(planets, g.res.Planets.Data) {
		t.Fatal("非月初未走共同清旗標尾巴，或錯做結算")
	}
}

func TestMonthSeatNormalMonthReportsRearmAndConsume(t *testing.T) {
	g := reachCommandPhase(t)
	for i := 0; i < 10; i++ {
		g.Click(544, 125)
		g.Click(560, 125)
		thinkThrough(g)
	}
	if g.turn != 10 || !g.monthSettlementPending || !bytes.Equal(monthSeatFlags(g), []byte{1, 1}) {
		t.Fatalf("正常9→10月報前旗標錯誤：turn=%d pending=%v flags=%v", g.turn, g.monthSettlementPending, monthSeatFlags(g))
	}
	if err := g.SaveSlot(t.TempDir(), 0); err == nil {
		t.Fatal("尚未結算月報允許保存")
	}
	dismissMonthReports(g)
	if !bytes.Equal(monthSeatFlags(g), []byte{0, 1}) {
		t.Fatal("關閉訂造報告未只消費玩家旗標")
	}
	money, building := g.military, g.building
	g.settlePlayerMonth()
	if money != g.military || building != g.building {
		t.Fatal("關閉報告後再次結算玩家")
	}
	g.beginTurn() // 明示CPU取得控制邊界。
	if !bytes.Equal(monthSeatFlags(g), []byte{0, 0}) {
		t.Fatal("月初CPU旗標未消費")
	}
	money, building = g.military, g.building
	g.beginTurn()
	if money != g.military || building != g.building {
		t.Fatal("月初CPU重複結算")
	}
}

func TestMonthSeatMissingWorkAndUnknownPreserved(t *testing.T) {
	w := bytes.Repeat([]byte{0xA5}, rules.WorkRecord)
	g := &Game{res: &Resources{WorkTemplate: w}}
	want := bytes.Clone(w)
	want[0xE0], want[0xE1] = 1, 1
	g.armMonthSeats()
	g.clearMonthSeat(0)
	want[0xE0] = 0
	if !bytes.Equal(g.workRecord(), want) || !bytes.Equal(w, bytes.Repeat([]byte{0xA5}, rules.WorkRecord)) {
		t.Fatal("原始模板或未知欄位被覆寫")
	}
	for _, f := range []int{-1, 2} {
		g.clearMonthSeat(f)
		if g.monthSeatDue(f) || !bytes.Equal(g.workRecord(), want) {
			t.Fatal("非法陣營修改旗標")
		}
	}
	for _, broken := range []*Game{{}, {res: &Resources{WorkTemplate: w[:225]}}} {
		broken.armMonthSeats()
		broken.clearMonthSeat(0)
		if broken.monthSeatDue(0) || len(broken.work) != 0 {
			t.Fatal("缺WORK仍捏造旗標")
		}
	}
	g.versus = true
	g.armMonthSeats()
	g.clearMonthSeat(1)
	if !bytes.Equal(g.workRecord(), want) {
		t.Fatal("修改未取證的雙人旗標生命週期")
	}
}
