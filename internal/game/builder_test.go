package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// reachBrainPhase 與 `reachCommandPhase` 一樣走到戰略地圖，另外把星域圖
// 接上——`runBrain` 缺 `Resources.Graph` 會整段跳過（`World` 帶不到圖，
// 決策層沒有東西可以走）。正式入口在 `internal/bootstrap` 會載它。
func reachBrainPhase(t *testing.T) *Game {
	t.Helper()
	res := load(t)
	graph, err := rules.LoadStarGraph(filepath.Join(root,
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	res.Graph = graph
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	if g.State() != StateCommandPhase {
		t.Fatalf("右鍵兩下之後在 %v，不是戰略地圖介面", g.State())
	}
	return g
}

// stubBuilder 是一個只會建艦隊、不下命令的 `Brain`。
type stubBuilder struct{ want []ai.NewFleet }

func (stubBuilder) Orders(ai.World, int, int) []ai.Order { return nil }

func (b stubBuilder) Build(ai.World, int, int) []ai.NewFleet { return b.want }

// plainBrain 沒有實作 `ai.Builder`——用來確認選配那條路真的是選配。
type plainBrain struct{}

func (plainBrain) Orders(ai.World, int, int) []ai.Order { return nil }

// TestBuilderAddsFleets 測 `ai.Builder` 那條路：新艦隊會進 `g.fleets`，
// 停在原地、沒有移動命令，類別鍵與副官照抄。
func TestBuilderAddsFleets(t *testing.T) {
	g := reachBrainPhase(t)
	foe := 1 - g.faction
	capital := rules.CapitalRegion(foe)
	g.SetBrain(stubBuilder{want: []ai.NewFleet{
		{Admiral: 41, Region: capital, ClassKey: 0x0A, Adjutant: 75},
		{Admiral: 42, Region: capital, ClassKey: 0x0E, Adjutant: -1},
	}})
	g.Click(544, 125)
	g.Click(560, 125)

	if len(g.fleets) != 2 {
		t.Fatalf("建出 %d 支，預期 2", len(g.fleets))
	}
	for i, f := range g.fleets {
		if f.Faction != foe {
			t.Errorf("第 %d 支的陣營 ＝ %d，預期 %d", i, f.Faction, foe)
		}
		if f.Region != capital {
			t.Errorf("第 %d 支在星域 %d，預期首都 %d", i, f.Region, capital)
		}
		if f.Target != rules.NoTarget {
			t.Errorf("第 %d 支有目的地 %d，新編的不該有", i, f.Target)
		}
		if f.Number != i+1 {
			t.Errorf("第 %d 支的編號 ＝ %d，預期 %d", i, f.Number, i+1)
		}
	}
	if g.fleets[0].ClassKey != 0x0A || g.fleets[1].ClassKey != 0x0E {
		t.Errorf("類別鍵 ＝ %X／%X，預期 0A／0E",
			g.fleets[0].ClassKey, g.fleets[1].ClassKey)
	}
	if g.fleets[0].Adjutant != 75 || g.fleets[1].Adjutant != -1 {
		t.Errorf("副官 ＝ %d／%d，預期 75／-1",
			g.fleets[0].Adjutant, g.fleets[1].Adjutant)
	}
}

// TestBuilderRespectsCapAndOneFleetPerAdmiral 測兩道限制：
// 每個陣營最多 16 支、一位提督只帶一支。
func TestBuilderRespectsCapAndOneFleetPerAdmiral(t *testing.T) {
	g := reachBrainPhase(t)
	foe := 1 - g.faction
	capital := rules.CapitalRegion(foe)
	var want []ai.NewFleet
	for i := 0; i < ai.FleetCap+5; i++ {
		want = append(want, ai.NewFleet{Admiral: 40 + i, Region: capital, Adjutant: -1})
	}
	// 同一位提督再來一次，不該多出一支。
	want = append(want, ai.NewFleet{Admiral: 40, Region: capital, Adjutant: -1})
	g.SetBrain(stubBuilder{want: want})
	g.Click(544, 125)
	g.Click(560, 125)
	if len(g.fleets) != ai.FleetCap {
		t.Fatalf("建出 %d 支，預期停在上限 %d", len(g.fleets), ai.FleetCap)
	}
	seen := map[int]bool{}
	for _, f := range g.fleets {
		if seen[f.Admiral] {
			t.Fatalf("提督 %d 帶了兩支", f.Admiral)
		}
		seen[f.Admiral] = true
	}
}

// TestPlainBrainBuildsNothing 確認沒實作 `ai.Builder` 的 `Brain`
// 不會被叫去建艦隊——`Original` 就是這一類。
func TestPlainBrainBuildsNothing(t *testing.T) {
	g := reachBrainPhase(t)
	g.SetBrain(plainBrain{})
	g.Click(544, 125)
	g.Click(560, 125)
	if len(g.fleets) != 0 {
		t.Errorf("沒實作 Builder 卻建出 %d 支", len(g.fleets))
	}
	if _, ok := ai.Brain(ai.Original{}).(ai.Builder); ok {
		t.Error("ai.Original 不該實作 ai.Builder")
	}
}
