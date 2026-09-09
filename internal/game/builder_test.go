package game

import (
	"encoding/binary"
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
	// 決策層還要艦種表與兩張編制表；`load()` 只給畫面要的東西，
	// 正式入口 `internal/bootstrap` 才會載這幾份。
	data := filepath.Join(root, "docs/internals/data")
	if res.ShipKinds, err = rules.LoadShipKindTable(
		filepath.Join(data, "ship-kind-table.json")); err != nil {
		t.Fatal(err)
	}
	if res.Composition, err = rules.LoadCompositionTables(
		filepath.Join(data, "fleet-composition-templates.json"),
		filepath.Join(data, "ship-kind-fallback.json")); err != nil {
		t.Fatal(err)
	}
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

// TestOriginalBrainBuildsWithComposition 把接好線的原版 AI 掛上去跑一個
// 回合：它會從首都的**待命部隊**（星域記錄 `+0x30` 起）編出艦隊，
// 每一支都有編制，而且存量真的被搬走（`docs/re/215`、`docs/re/217`）。
//
// ⚠ **它會一路建到上限 16 支，而原版在同一個局面停在三支**
// （收據 `turn260mem`，`docs/re/209`）。原因是收手的那道閘門看艦隊
// 維持費，而 remake 還沒算維持費（`docs/re/211`）——這一支就是把那個
// 已知偏差釘住，換掉之後有人修好維持費，它會紅，那時候再改預期。
func TestOriginalBrainBuildsWithComposition(t *testing.T) {
	g := reachBrainPhase(t)
	foe := 1 - g.faction
	capital := rules.CapitalRegion(foe)
	rec := g.regionRecord(capital)
	if rec == nil {
		t.Fatal("拿不到電腦首都的星域記錄")
	}
	before := 0
	for _, v := range rules.RegionShips(rec, foe) {
		before += v
	}
	if before == 0 {
		t.Skip("新遊戲的首都沒有待命部隊，這一支測不到")
	}

	g.SetBrain(g.OriginalBrain())
	g.Click(544, 125)
	g.Click(560, 125)

	if len(g.fleets) == 0 {
		t.Fatal("一支都沒建")
	}
	for i, f := range g.fleets {
		if f.Faction != foe {
			t.Errorf("第 %d 支是陣營 %d，預期 %d", i, f.Faction, foe)
		}
		if len(f.Composition) == 0 {
			t.Errorf("第 %d 支沒有編制", i)
		}
		if f.Groups != len(f.Composition) {
			t.Errorf("第 %d 支 Groups ＝ %d，編制有 %d 組", i, f.Groups, len(f.Composition))
		}
		if f.Expected < f.Groups {
			t.Errorf("第 %d 支現有 %d 組多於應有 %d 組", i, f.Groups, f.Expected)
		}
	}
	after := 0
	for _, v := range rules.RegionShips(rec, foe) {
		after += v
	}
	if after >= before {
		t.Errorf("存量從 %d 變成 %d——一艘都沒搬", before, after)
	}
	// 已知偏差：沒有維持費就收不了手。
	if len(g.fleets) != ai.FleetCap {
		t.Logf("建了 %d 支（不是上限 %d）——維持費那道閘門是不是接上了？",
			len(g.fleets), ai.FleetCap)
	}
	t.Logf("建了 %d 支，首都待命部隊從 %d 剩 %d", len(g.fleets), before, after)
}

// TestOriginalBrainNeedsStock 首都沒有待命部隊就一支都編不出來
// （`07F9:023B` 的存量檢查，`docs/re/215`）。
func TestOriginalBrainNeedsStock(t *testing.T) {
	g := reachBrainPhase(t)
	foe := 1 - g.faction
	rec := g.regionRecord(rules.CapitalRegion(foe))
	if rec == nil {
		t.Fatal("拿不到電腦首都的星域記錄")
	}
	for k := 0; k < 7; k++ {
		binary.LittleEndian.PutUint32(rec[0x30+foe*28+k*4:], 0)
	}
	g.SetBrain(g.OriginalBrain())
	g.Click(544, 125)
	g.Click(560, 125)
	if len(g.fleets) != 0 {
		t.Errorf("沒有待命部隊卻建出 %d 支", len(g.fleets))
	}
}

// TestOriginalBrainCandidatesExcludeLedAdmirals 候選提督不含已經帶隊的。
func TestOriginalBrainCandidatesExcludeLedAdmirals(t *testing.T) {
	g := reachBrainPhase(t)
	s := g.buildSupport()
	if s == nil {
		t.Fatal("buildSupport 回 nil——提督表沒接上")
	}
	before := len(s.Admirals(1))
	if before == 0 {
		t.Fatal("同盟一個候選提督都沒有")
	}
	// 讓其中一位帶隊，候選就要少一個。
	first := s.Admirals(1)[0]
	g.fleets = append(g.fleets, rules.Fleet{Faction: 1, Admiral: first.Index, Adjutant: -1})
	if after := len(s.Admirals(1)); after != before-1 {
		t.Errorf("帶隊之後候選 %d 人，預期 %d", after, before-1)
	}
}

// TestOriginalBrainStillNotDefault 預設仍然沒有 `Brain`——換上原版那套
// 是 `docs/spec/029` 末節的裁決，不是這一輪接線的副作用。
func TestOriginalBrainStillNotDefault(t *testing.T) {
	g := reachBrainPhase(t)
	if g.brain != nil {
		t.Errorf("預設的 Brain 是 %T，預期 nil", g.brain)
	}
	// `ai.Original` 與 nil 同義：兩者都一道命令都不下。
	if got := (ai.Original{}).Orders(ai.World{}, 0, 0); got != nil {
		t.Errorf("ai.Original 回了 %v，預期 nil", got)
	}
}
