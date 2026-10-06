package game

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// 與正式 CLI 相同：先配置偏好，再從主選單選新局，沒有狀態注入。
func reachConfiguredAI(t *testing.T, level, faction int) *Game {
	t.Helper()
	res := loadWithRules(t)
	loadBattleMaps(t, res)
	weights, err := rules.LoadStrengthWeights(filepath.Join(root, "docs/internals/data/ship-strength-weights.json"))
	if err != nil {
		t.Fatal(err)
	}
	res.Weights = weights // 與正式 bootstrap.Load 相同，舊助手只載畫面／建隊規則。
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.SetAILevel(level); err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210+20*faction)
	for _, c := range []byte("138\r") {
		g.Key(c)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	wantState(t, g, StateCommandPhase, "已配置 AI 的正常新局")
	if g.AILevel() != level || g.faction != faction {
		t.Fatal("新局丟失 AI 偏好或陣營")
	}
	g.MoveMouse(600, 340)
	return g
}

func aiFleetCommands(g *Game, faction int) [][4]int {
	var out [][4]int
	for _, f := range g.fleets {
		if f.Faction == faction {
			out = append(out, [4]int{f.Index, f.Region, f.Target, f.Move})
		}
	}
	return out
}

func TestEnhancedAINormalPlayerPath(t *testing.T) {
	for _, faction := range []int{0, 1} {
		control := reachConfiguredAI(t, -1, faction)
		g := reachConfiguredAI(t, 1, faction)
		for _, game := range []*Game{control, g} {
			if game.rand94DC().Seed() != strategyRandSeed {
				t.Fatal("執行前種子未固定")
			}
			executeTurn(t, game)
		}
		foe := 1 - faction
		orig, enhanced := aiFleetCommands(control, foe), aiFleetCommands(g, foe)
		t.Logf("玩家陣營 %d：原版 %v，增強 %v", faction, orig, enhanced)
		if len(orig) == 0 || len(enhanced) == 0 {
			t.Fatal("正常回合沒有實際建隊")
		}
		if reflect.DeepEqual(orig, enhanced) {
			t.Fatal("增強模式沒有影響正式回合的命令／位置")
		}
		before := g.brain
		g.res.SaveDir = t.TempDir()
		g.Click(544, 125)
		g.Click(560, 165)
		wantState(t, g, StateSaveGame, "AI 局面正常保存")
		g.Click(300, g.res.SlotList.Save.TextTop)
		if err := g.SaveErr(); err != nil {
			t.Fatal(err)
		}
		g.RightClick(300, 142)
		g.Click(544, 125)
		g.Click(560, 145)
		g.Click(300, 122)
		wantState(t, g, StateLoadFaction, "AI局面選陣營")
		g.Click(318, 208+20*faction)
		wantState(t, g, StateCommandPhase, "AI 局面正常載入")
		b, ok := g.brain.(*enhancedBrain)
		if !ok || b == before || g.AILevel() != 1 || b.Difficulty != 1 || b.builder == nil {
			t.Fatal("載入丟失強化偏好、未重接建隊來源或沿用舊策略暫態")
		}
		if !reflect.DeepEqual(enhanced, aiFleetCommands(g, foe)) {
			t.Fatal("正常存讀改變 AI 已下的命令")
		}
		// 按主選單相同方式開另一局時，模式仍在、目標／冷卻物件重建。
		g.startNewGame(faction)
		if g.brain == b || g.AILevel() != 1 {
			t.Fatal("新局未重建強化策略暫態")
		}
		if _, passive := g.passiveComputer(); passive {
			t.Fatal("新局仍套用已被120取代的被動戰場裁決")
		}
	}
}

func TestEnhancedAIPreferenceAndLiveStrength(t *testing.T) {
	g := reachConfiguredAI(t, 0, 0)
	for _, level := range []int{0, 1, 2, 3, -1} {
		if err := g.SetAILevel(level); err != nil || g.AILevel() != level {
			t.Fatalf("合法難度 %d 未保留：%v", level, err)
		}
	}
	if err := g.SetAILevel(1); err != nil {
		t.Fatal(err)
	}
	before := g.brain
	for _, level := range []int{-2, 4} {
		if g.SetAILevel(level) == nil || g.AILevel() != 1 || g.brain != before {
			t.Fatal("不合法難度未拒絕或改變已選模式")
		}
	}
	f := rules.Fleet{Strength: 99999, Composition: []rules.FleetGroup{{Kind: 0, Ships: 50, Full: 100}}}
	want := g.res.Weights[0] * 500
	if got := g.liveFleetStrength(f); got != want {
		t.Fatalf("現況戰力 %d，應為 %d", got, want)
	}
	f.Composition[0].Ships = 0
	if g.liveFleetStrength(f) != 0 {
		t.Fatal("全損編制被舊戰力快取蓋過")
	}
	if err := g.SetAILevel(-1); err != nil {
		t.Fatal(err)
	}
	if _, ok := g.brain.(ai.OriginalStrategy); !ok {
		t.Fatal("選回 -1 沒有回到原版完整策略")
	}
}
