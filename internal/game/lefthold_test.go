package game

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 同結果輸入映射：089清單速度已核准不同，舊收據仍保存原保持指令數。
func holdListSteps(g *Game, x, y, steps int) {
	g.Click(x, y)
	if steps > 1 {
		g.LeftHold(x, y, 350*time.Millisecond+time.Duration(steps-2)*100*time.Millisecond)
	}
	g.LeftRelease()
}

func TestLeftHoldListTimingAndCancellation(t *testing.T) {
	g := reachListScreen(t, 0)
	g.Click(615, 390)
	if g.listScreenScroll() != 1 {
		t.Fatal("短按未移一列")
	}
	g.LeftHold(100, 100, 349*time.Millisecond)
	if g.listScreenScroll() != 1 {
		t.Fatal("初始延遲前就捲動")
	}
	g.LeftHold(100, 100, time.Millisecond)
	if g.listScreenScroll() != 2 {
		t.Fatal("350ms未更新第二列，或移出清單錯誤取消")
	}
	g.LeftHold(100, 100, 99*time.Millisecond)
	if g.listScreenScroll() != 2 {
		t.Fatal("100ms間隔前就更新")
	}
	g.LeftHold(100, 100, time.Millisecond)
	if g.listScreenScroll() != 3 {
		t.Fatal("100ms間隔未更新")
	}
	g.LeftHold(100, 100, 10*time.Second)
	if g.listScreenScroll() != 16 {
		t.Fatal("沒有夾在一覽表上界16")
	}
	g.LeftRelease()
	g.LeftHold(615, 90, time.Second)
	if g.listScreenScroll() != 16 {
		t.Fatal("放開後保持仍更新")
	}
	g.Click(615, 90)
	g.LeftHold(615, 90, 10*time.Second)
	if g.listScreenScroll() != 0 {
		t.Fatal("沒有夾在下界0")
	}
	g.Click(615, 390)
	g.msgOn = true // 窄幅診斷：同畫面上的模態也必須取消。
	g.LeftHold(615, 390, time.Second)
	if g.listScreenScroll() != 1 || g.leftHeld.action.kind != leftRepeatNone {
		t.Fatal("同畫面模態未取消保持")
	}
	g.msgOn = false
	g.Click(615, 390)
	g.RightClick(100, 100)
	if g.leftHeld.action.kind != leftRepeatNone {
		t.Fatal("右鍵未清除保持")
	}
}

func TestLeftHoldDoesNotRepeatSelectionOrCrossTabs(t *testing.T) {
	g := reachListTab(t, listTabAdmiral)
	g.Click(617, 388)
	g.Click(415, 290)
	g.LeftHold(617, 388, time.Second)
	if g.listPage(listTabAdmiral).Scroll != 1 || g.leftHeld.action.kind != leftRepeatNone {
		t.Fatal("保持跨分頁")
	}
	g.Click(560, 290)
	g.Click(450, 308)
	g.LeftHold(496, 320, 10*time.Second)
	if !g.listOverlayOn() || g.State() != StateCommandPhase {
		t.Fatal("按住選擇自動確認或出擊")
	}
}

func TestLeftHoldFourListDomains(t *testing.T) {
	for item := 0; item < 4; item++ {
		g := reachListScreen(t, item)
		count := g.listScreenCount()
		last := count - g.res.ListScreen.Rows
		if last < 0 {
			last = 0
		}
		g.Click(615, 390)
		g.LeftHold(100, 100, 350*time.Millisecond)
		want := 2
		if want > last {
			want = last
		}
		if g.listScreenScroll() != want {
			t.Fatalf("%v的保持沒有路由到自身位移：%d，要%d", g.State(), g.listScreenScroll(), want)
		}
		g.LeftRelease()
	}
}

func TestLeftHoldStandbyAndStaffRows(t *testing.T) {
	g := reachFleetDecided(t)
	g.Click(352, 125)
	g.Click(400, 125)
	if g.state != StateStandbyAdmirals {
		t.Fatal("正常命令未開待命清單")
	}
	g.setListRows(g.standbySpec())
	want := g.listIndex[1]
	g.Click(617, 390)
	if g.standbyScroll != 1 || g.state != StateStandbyAdmirals {
		t.Fatal("箭頭誤選清單列，或未移一列")
	}
	g.Click(450, 90)
	if g.state != StateAdmiralDetail || g.detailAdmiral != want {
		t.Fatal("偏移後列選擇沒有用同一資料起點")
	}
	g = reachTaxPlan(t)
	g.Click(416, 369)
	if !g.staff.listOn {
		t.Fatal("正常內政參謀清單未開")
	}
	holdListSteps(g, 617, 390, 2)
	if g.staff.first != 2 || g.staff.detailOn {
		t.Fatal("參謀箭頭沒有捲動或誤開詳細資料")
	}
	g.RightClick(100, 100)
	if g.staff.listOn || g.leftHeld.action.kind != leftRepeatNone {
		t.Fatal("參謀清單關閉未取消保持")
	}
}

func TestLeftHoldEconomicTimingBoundsAndMove(t *testing.T) {
	for _, state := range []State{StateTaxPlan, StateBudgetPlan} {
		t.Run(state.String(), func(t *testing.T) {
			g := reachTaxPlan(t)
			g.state = state // 共用處理端診斷；正常月初入口由既有budget測試覆蓋。
			g.Click(471, 329)
			if g.shipPct[g.faction] != 26 {
				t.Fatal("短按未加一點")
			}
			g.LeftHold(472, 329, leftPlanPeriod-time.Nanosecond)
			if g.shipPct[g.faction] != 26 {
				t.Fatal("八倒數前就更新")
			}
			g.LeftHold(472, 329, time.Nanosecond)
			if g.shipPct[g.faction] != 27 {
				t.Fatal("在同一熱區移一像素後未保持")
			}
			g.LeftHold(100, 100, 5*leftPlanPeriod)
			if g.shipPct[g.faction] != 27 {
				t.Fatal("移出熱區仍更新")
			}
			g.LeftHold(471, 329, leftPlanPeriod)
			if g.shipPct[g.faction] != 28 {
				t.Fatal("返回熱區未繼續，或累積補發")
			}
			g.LeftHold(471, 329, 20*time.Second)
			if g.shipPct[g.faction] != 100 {
				t.Fatal("百分比上界未夾住")
			}
			g.Click(487, 329)
			g.LeftHold(487, 329, 20*time.Second)
			if g.shipPct[g.faction] != 0 {
				t.Fatal("百分比下界未夾住")
			}
			g.Click(471, 289)
			g.LeftHold(471, 289, 20*time.Second)
			if g.blend[g.faction] != 4 {
				t.Fatal("稅率上界未夾住")
			}
			g.Click(487, 289)
			g.LeftHold(487, 289, 20*time.Second)
			if g.blend[g.faction] != 0 {
				t.Fatal("稅率下界未夾住")
			}
			g.LeftRelease()
			out := t.TempDir()
			if err := g.SaveSlot(out, 0); err != nil {
				t.Fatal(err)
			}
			work, err := os.ReadFile(filepath.Join(out, "WORK.DT0"))
			if err != nil || len(work) != 230 || work[0xC0+g.faction] != 0 || work[0x7E+g.faction] != 0 {
				t.Fatalf("正式保存未保留保持後的稅率／百分比：%v", err)
			}
			g.Click(471, 329)
			g.LoadSlot(g.slot)
			if g.leftHeld.action.kind != leftRepeatNone {
				t.Fatal("讀檔沿用舊保持")
			}
		})
	}
}

func TestLeftHoldMonthlyAdmiralList(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, key := range []byte("138\r") {
		g.Key(key)
	}
	g.Click(448, 330)
	g.Click(448, 310)
	g.Click(464, 330)
	if g.State() != StateAdmiralList {
		t.Fatal("正常月初建造計劃未開待命提督清單")
	}
	g.Click(608, 390) // 原版16像素熱區的左緣，不能只認8像素圖塊。
	g.LeftHold(100, 100, leftListDelay)
	if g.standbyScroll != 2 || g.State() != StateAdmiralList {
		t.Fatal("月初清單保持沒有移兩列")
	}
	g.LeftRelease()
	g.MoveMouse(617, 390)
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/hold60-r58-month-clean-a/screen.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺月初原版保持收據")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(want)) != "45c79a94d9a81231198081a05041c81e3bb22a65dddbd31202729e81882a532f" {
		t.Fatalf("月初原版收據漂移：%v", err)
	}
	if base := os.Getenv("LOGH3_LEFT_HOLD_MONTH_OUT"); base != "" {
		if err := os.Mkdir(base, 0755); err != nil {
			t.Fatal(err)
		}
		if err := g.Frame().WriteIndex(filepath.Join(base, "screen.bin")); err != nil {
			t.Fatal(err)
		}
		if err := g.Frame().WritePNG(filepath.Join(base, "screen.png"), g.res.Menu.Palette()); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(g.Frame().Pix, want) {
		t.Fatal("月初清單保持後的完整畫面不同")
	}
	g.Click(608, 90)
	g.RightClick(100, 100)
	if g.State() != StateAdmiralList || g.leftHeld.action.kind != leftRepeatNone {
		t.Fatal("月初原版右鍵保留清單，但必須取消左键捕獲")
	}
	if err := g.SaveSlot(t.TempDir(), 0); err == nil {
		t.Fatal("月初清單保持不應繞過結算保存閘門")
	}
}
