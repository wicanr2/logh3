package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 交戰演出（規格 056、`docs/re/324`）：S3 第五回合同盟艦隊打帝國行星雷奧米爾那一場，
// 收據 `eng-open`（開場：面板、三個視窗、3300/3300、22000/22000）與 `eng-end`
// （結果：擊毀 290、21710/22000）。
//
// ⚠ **直接進入的診斷比對**：remake 的戰場模擬走不到與原版同一個第五回合（戰術 AI
// 與亂數各走各的），所以這裡把那一場的數直接排進演出佇列，只驗畫面那一層；
// 行星圖、艦隊圖、量表與戰場動畫是事先列出的排除區（`engagement-panel.json`
// 的 `known_diff`，理由在規格 056 已知差異）。

func loadEngagement(t *testing.T, res *Resources) {
	t.Helper()
	if res.Engagement == nil {
		spec, err := screen.LoadEngagementSpec(filepath.Join(root, "docs/internals/data/engagement-panel.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.Engagement = spec
	}
	if res.EngagementPanel == nil {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted", res.Engagement.Panel.File))
		if err != nil {
			t.Skipf("缺 %s，跳過", res.Engagement.Panel.File)
		}
		img, err := asset.DecodeAMGImage(b)
		if err != nil {
			t.Fatal(err)
		}
		res.EngagementPanel = img
	}
}

// 雷奧米爾（Big5）。
var planetLeomir = []byte{0xB9, 0x70, 0xB6, 0xF8, 0xA6, 0xCC, 0xBA, 0xB8}

// reachDashiliBattle 把 S3 載進來，兩邊各搬一支艦隊到達希利（原版要六個回合
// 才會合在那裡），進戰場。
func reachDashiliBattle(t *testing.T) *Game {
	t.Helper()
	g := newTestGame(t)
	loadEngagement(t, g.res)
	const dashili = 26
	moved := [2]bool{}
	for i := range g.fleets {
		f := &g.fleets[i]
		if f.Faction < 0 || f.Faction > 1 || moved[f.Faction] {
			continue
		}
		f.Region, f.Target, f.Move, f.Parked = dashili, rules.NoTarget, rules.MoveParked, true
		moved[f.Faction] = true
	}
	g.StartBattle(dashili)
	if g.state != StateBattle {
		t.Fatalf("進戰場之後在 %v", g.state)
	}
	return g
}

// reachEngagementFixture 進達希利的戰場，再把那一場的數排進演出。
func reachEngagementFixture(t *testing.T) *Game {
	t.Helper()
	g := reachDashiliBattle(t)
	planet := -1
	for i := range g.battlePlanets {
		if string(g.battlePlanetName(i)) == string(planetLeomir) {
			planet = i
		}
	}
	if planet < 0 {
		t.Fatal("達希利的戰場上找不到雷奧米爾")
	}
	// 同盟那一支（收據裡看不清編號：第？艦隊 ？？？？？？？？？）。
	fleet := -1
	for _, f := range g.battleFleets {
		if f.Faction == 1 {
			fleet = f.Index
			break
		}
	}
	if fleet < 0 {
		t.Fatal("戰場上沒有同盟艦隊")
	}
	g.engQueue = []engagement{{
		planet: planet, fleet: fleet,
		defence: 3300, defenceMax: 3300,
		ships: 22000, shipsMax: 22000, destroyed: 290,
		label: screen.EngagementDefence,
	}}
	g.engPhase = 0
	g.state = StateEngagement
	return g
}

func engagementSkips(g *Game) [][4]int {
	var out [][4]int
	for _, r := range g.res.Engagement.KnownDiff {
		out = append(out, [4]int{r.X, r.Y, r.X + r.W - 1, r.Y + r.H - 1})
	}
	return out
}

func TestEngagementMatchesOracle(t *testing.T) {
	g := reachEngagementFixture(t)
	matchesExcept(t, g, "eng-open", engagementSkips(g))
	g.Click(300, 240)
	if g.state != StateEngagement || g.engPhase != 1 {
		t.Fatalf("按一下之後在 %v、幀 %d", g.state, g.engPhase)
	}
	matchesExcept(t, g, "eng-end", engagementSkips(g))
	g.RightClick(300, 240)
	if g.state != StateBattle || len(g.engQueue) != 0 {
		t.Fatalf("看完之後在 %v、佇列 %d", g.state, len(g.engQueue))
	}
}

// 回合處理裡打行星有結果就切演出，看完才做這一回合剩下的（回合數才加）。
func TestEngagementQueuedFromAssault(t *testing.T) {
	g := reachDashiliBattle(t)
	turn := g.battleTurn
	hits := []rules.PlanetAssaultHit{{Fleet: g.battleFleets[0].Index, Planet: 0, Damage: 100, Counter: 50}}
	if !g.queueEngagements(hits) {
		t.Fatal("打行星有結果卻沒排演出")
	}
	if g.state != StateEngagement {
		t.Fatalf("在 %v", g.state)
	}
	e := g.engQueue[0]
	if e.defence != g.battlePlanets[0].Defence+100 || e.destroyed != 50 ||
		e.ships != g.battleFleetShips(e.fleet)+50 || e.label != screen.EngagementLanding {
		t.Fatalf("演出的數 %+v", e)
	}
	g.engagementClick()
	g.engagementClick()
	if g.state != StateBattle || g.battleTurn != turn+1 {
		t.Fatalf("看完之後在 %v、回合 %d（進來時 %d）", g.state, g.battleTurn, turn)
	}
}
