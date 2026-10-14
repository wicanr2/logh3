package game

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 145：診斷fixture核狀態不變；正常片頭與X11另由battle_save_gui驗收。
func TestBattleSaveSuccessReturnsWithoutAdvancing(t *testing.T) {
	g := newTestGame(t)
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("沒有首戰")
	}
	g.res.SaveDir = t.TempDir()
	units := append([]rules.BattleUnit(nil), g.battleUnits...)
	fleets := append([]rules.BattleFleet(nil), g.battleFleets...)
	planets := append([]rules.BattlePlanet(nil), g.battlePlanets...)
	view, turn, region := g.battleView, g.battleTurn, g.battleRegion
	score, fleetScore, campaign := g.battleScore, g.battleFleetScore, g.campaignScore
	seed := g.battleRand.Seed()
	openSave := func() {
		g.Click(240, 470)
		x, y := g.res.BattleMenu.BattleMenuAt(240, 470)
		g.Click(x+20, y+20+4*g.res.BattleMenu.Layout.Text.RowPitch)
		if g.State() != StateSaveGame {
			t.Fatalf("未開保存畫面：%v", g.State())
		}
	}
	openSave()
	g.Click(300, 152)
	if g.SaveErr() != nil || g.State() != StateBattle || g.battleMenu {
		t.Fatalf("成功保存未直接返回戰場：%v %v", g.State(), g.SaveErr())
	}
	if g.battleView != view || g.battleTurn != turn || g.battleRegion != region ||
		g.battleRand.Seed() != seed || g.battleScore != score || g.battleFleetScore != fleetScore || g.campaignScore != campaign ||
		!reflect.DeepEqual(units, g.battleUnits) || !reflect.DeepEqual(fleets, g.battleFleets) || !reflect.DeepEqual(planets, g.battlePlanets) {
		t.Fatal("保存返回改動戰場、命令、計分或亂數")
	}
	w, err := os.ReadFile(filepath.Join(g.res.SaveDir, "WORK.DT0"))
	if err != nil || len(w) != rules.WorkRecord ||
		!bytes.Equal(w[0x02:0x1F], g.slot.Work[0x02:0x1F]) ||
		!bytes.Equal(w[0x26:0x2E], g.slot.Work[0x26:0x2E]) {
		t.Fatalf("WORK控制、回合或鏡頭沒有保存：%v", err)
	}
	for _, name := range rules.SlotFileNames(0) {
		if _, err := os.Stat(filepath.Join(g.res.SaveDir, name)); err != nil {
			t.Fatal(err)
		}
	}
	// 用同名檔案擋住輸出目錄，確定失敗不會假稱成功並關窗。
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}
	g.res.SaveDir = blocked
	openSave()
	g.Click(300, 152)
	if g.SaveErr() == nil || g.State() != StateSaveGame || g.battleTurn != turn || g.battleRand.Seed() != seed {
		t.Fatal("保存失敗離開畫面或推進戰鬥")
	}
}
