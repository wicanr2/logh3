package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func TestCopyProtectionVisibleConfirm(t *testing.T) {
	for _, row := range []int{menuRowEmpire, menuRowAlliance, menuRowVersus, menuRowLoad} {
		g, err := New(load(t))
		if err != nil {
			t.Fatal(err)
		}
		g.Click(300, 210+20*row)
		b := g.res.CopyProt.InputBox
		if b == nil {
			t.Fatal("缺少確認按鈕")
		}
		g.Click(b.X+b.W/2, b.Y+b.H/2)
		if g.State() != StateCopyProtection {
			t.Fatal("空答案被送出")
		}
		for _, c := range []byte("138") {
			g.Key(c)
		}
		g.Click(b.X-1, b.Y+b.H/2)
		if g.State() != StateCopyProtection {
			t.Fatal("按鈕外送出答案")
		}
		g.Click(b.X+b.W/2, b.Y+b.H/2)
		want := StateMainGame
		if row == menuRowVersus {
			want = StateVersusInput
		}
		if row == menuRowLoad {
			want = StateLoadGame
		}
		if g.State() != want {
			t.Fatalf("選項%d確認未送出：%v != %v", row, g.State(), want)
		}
	}
}

func TestVersusVisibleMonthlyButtons(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 250)
	for _, c := range []byte("138\r") {
		g.Key(c)
	}
	w := g.commandSpec(StateVersusInput).Windows["overlay"].Boxes[0]
	g.Click(w.X+w.W/2, w.Y+w.H/2)
	if g.State() != StateVersusAccount {
		t.Fatal(g.State())
	}
	g.Click(448, 350)
	if g.State() != StateBudgetPlan {
		t.Fatalf("雙人財政變更失效：%v", g.State())
	}
	g.Click(463, 370)
	if g.State() != StateVersusAccount {
		t.Fatalf("預算未回雙人財政：%v", g.State())
	}
	g.Click(448, 370)
	if g.State() != StateVersusShipOrder {
		t.Fatalf("雙人財政確認失效：%v", g.State())
	}
	g.Click(448, 350)
	if g.State() != StateBuildPlan {
		t.Fatalf("雙人訂造變更失效：%v", g.State())
	}
	g.Click(512, 370)
	if g.State() != StateVersusShipOrder {
		t.Fatalf("建造未回雙人訂造：%v", g.State())
	}
	g.Click(448, 370)
	if g.State() != StateCommandPhase {
		t.Fatalf("雙人訂造確認失效：%v", g.State())
	}
}

// 畫面上的按鈕必須可以由 Game 公開畫面座標操作，不能沿用舊熱區座標。
func TestVisibleMonthlyReportButtons(t *testing.T) {
	for _, faction := range []int{0, 1} {
		g, err := New(load(t))
		if err != nil {
			t.Fatal(err)
		}
		g.Click(300, 210+20*faction)
		for _, c := range []byte{'1', '3', '8', '\r'} {
			g.Key(c)
		}
		g.Click(448, 310) // 可見按鈕上方不能觸發變更。
		if g.State() != StateMainGame {
			t.Fatalf("上方空白觸發報告按鈕：%v", g.State())
		}
		g.Click(448, 350)
		if g.State() != StateBudgetPlan {
			t.Fatalf("可見變更未開預算：%v", g.State())
		}
		g.Click(463, 370)
		if g.State() != StateMainGame {
			t.Fatalf("預算結束未返回會計：%v", g.State())
		}
		g.Click(448, 370)
		if g.State() != StateShipOrder {
			t.Fatalf("可見確認未開訂造：%v", g.State())
		}
		g.Click(448, 350)
		if g.State() != StateBuildPlan {
			t.Fatalf("訂造變更未開建造：%v", g.State())
		}
		before := g.plan.Pct
		for row := range before {
			g.Click(232, 210+row*20)
			if g.plan.Pct[row] != before[row]-1 {
				t.Fatalf("可見第%d列減鈕未改正確列：%v", row, g.plan.Pct)
			}
		}
		want := g.plan.Pct
		g.Click(512, 370)
		if g.State() != StateShipOrder || g.pct != want {
			t.Fatalf("建造結束未保留計畫：%v %v", g.State(), g.pct)
		}
		g.Click(448, 370)
		if g.State() != StateCommandPhase || g.pct != want {
			t.Fatalf("訂造確認未完成月初：%v %v", g.State(), g.pct)
		}
		saved := t.TempDir()
		if err := g.SaveSlot(saved, 0); err != nil {
			t.Fatal(err)
		}
		read := func(name string) []byte {
			b, err := os.ReadFile(filepath.Join(saved, name+".DT0"))
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		header, err := os.ReadFile(filepath.Join(root, "org_game/extracted/PLADAT.DAT"))
		if err != nil {
			t.Fatal(err)
		}
		slot, err := rules.LoadSlot(rules.SlotFiles{Stellar: read("STELLAR"), Planet: read("PLANET"), Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Work: read("WORK"), PlanetHeader: header})
		if err != nil {
			t.Fatal(err)
		}
		g.LoadSlot(slot)
		if g.pct != want {
			t.Fatalf("保存再讀回丟失百分比：%v != %v", g.pct, want)
		}
	}
}

func TestMonthlyBuildPlanRightClickReturnsReport(t *testing.T) {
	g := reachBuildPlan(t)
	g.Click(232, 210)
	want := g.plan.Pct
	g.RightClick(600, 400)
	if g.State() != StateShipOrder || g.pct != want {
		t.Fatalf("右鍵未保留計畫返回訂造：%v %v", g.State(), g.pct)
	}
}

func TestReportAndDropdownVisibleCursor(t *testing.T) {
	for _, state := range []State{StateCmdMenu, StateListMenu, StateSysMenu, StateShipOrder, StateBuildPlan} {
		g, err := New(load(t))
		if err != nil {
			t.Fatal(err)
		}
		g.state = state
		g.MoveMouse(600, 420)
		g.Frame()
		var spec = g.commandSpec(state)
		if state == StateShipOrder {
			spec = g.res.ShipOrder
		}
		if state == StateBuildPlan {
			spec = g.res.BuildPlan
		}
		if spec == nil || spec.Mouse == nil {
			t.Fatalf("%v缺游標", state)
		}
		if spec.Mouse.X != 600 || spec.Mouse.Y != 380 {
			t.Fatalf("%v游標停在舊位置：%v", state, spec.Mouse)
		}
		g.MoveMouse(550, 185)
		g.Frame()
		if spec.Mouse.X != 550 || spec.Mouse.Y != 145 {
			t.Fatalf("%v移動未跟隨", state)
		}
	}
}
