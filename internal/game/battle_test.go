package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// newTestGame 用測試那份資源開一局，另外把戰場入口表接上。
func newTestGame(t *testing.T) *Game {
	t.Helper()
	res := load(t)
	e, err := rules.LoadBattleEntries(filepath.Join(root, "docs/internals/data/battle-entry-points.json"))
	if err != nil {
		t.Fatal(err)
	}
	res.Entries = e
	if res.Graph == nil {
		graph, err := rules.LoadStarGraph(filepath.Join(root,
			"docs/internals/data/star-graph.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.Graph = graph
	}
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	// 載 S3「巴米利恩會戰」的存檔，才有艦隊可以佈陣。
	orig := filepath.Join(root, "org_game/extracted")
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(orig, name))
		if err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", name)
		}
		return b
	}
	sl, err := rules.LoadSlot(rules.SlotFiles{
		Stellar: read("STELLAR.DTA"), Planet: read("PLANET.DTA"),
		Admiral: read("ADMIRAL.DTA"), Fleet: read("FLEET.DTA"),
		Ship: read("SHIP.DTA"), Work: read("WORK.DTA"),
		PlanetHeader: read("PLADAT.DAT"),
	})
	if err != nil {
		t.Fatal(err)
	}
	g.LoadSlot(sl)
	return g
}

// StartBattle 要把參戰艦隊放到入口格上，並且把分隊散開。
func TestStartBattleDeploys(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	// 找一個有艦隊的星域。
	region := -1
	for _, f := range g.fleets {
		if len(g.res.Entries.Regions[f.Region].Entries) > 0 {
			region = f.Region
			break
		}
	}
	if region < 0 {
		t.Skip("找不到有艦隊、又在入口表裡的星域")
	}
	g.StartBattle(region)
	if g.state != StateBattle {
		t.Fatalf("狀態是 %v，應該是 StateBattle", g.state)
	}
	if len(g.battleUnits) == 0 {
		t.Fatal("一個分隊都沒佈上去")
	}
	// 旗艦要落在該星域的某一個入口上。
	u := g.battleUnits[0]
	ok := false
	for _, e := range g.res.Entries.Regions[region].Entries {
		if u.X == e.X && u.Y == e.Y {
			ok = true
		}
	}
	if !ok {
		t.Errorf("旗艦在 (%d,%d)，不是星域 %d 的任何一個入口", u.X, u.Y, region)
	}
	// 分隊要散開，不能疊在一起。
	seen := map[[2]int]bool{}
	for _, v := range g.battleUnits {
		if seen[[2]int{v.X, v.Y}] {
			t.Fatalf("有兩個分隊疊在 (%d,%d)", v.X, v.Y)
		}
		seen[[2]int{v.X, v.Y}] = true
	}
}

// 推一個回合，敵方那邊要真的動。
func TestAdvanceBattleTurnMoves(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.battleUnits = []rules.BattleUnit{
		{Fleet: 0, Faction: 0, X: 400, Y: 400, TargetX: 400, TargetY: 400, Strength: 1000},
		{Fleet: 1, Faction: 1, X: 1400, Y: 1400, TargetX: 1400, TargetY: 1400, Strength: 1000},
	}
	g.battleTurn = 1
	g.faction = 0
	g.SetTactic(ai.OriginalTactic{}, false)
	before := g.battleUnits[1]
	g.AdvanceBattleTurn()
	after := g.battleUnits[1]
	if after.X == before.X && after.Y == before.Y {
		t.Fatal("敵方那一支沒動")
	}
	if g.battleTurn != 2 {
		t.Fatalf("回合是 %d，應該是 2", g.battleTurn)
	}
	// 玩家那一支沒下令，不該被 AI 動到目標。
	if g.battleUnits[0].TargetX != 400 || g.battleUnits[0].TargetY != 400 {
		t.Error("玩家那一支的目標被 AI 改了")
	}
}
