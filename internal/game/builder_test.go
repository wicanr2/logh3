package game

import (
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// reachBrainPhase 與 `reachCommandPhase` 一樣走到戰略地圖，另外把星域圖
// 接上——`runBrain` 缺 `Resources.Graph` 會整段跳過（`World` 帶不到圖，
// 決策層沒有東西可以走）。正式入口在 `internal/bootstrap` 會載它。
func reachBrainPhase(t *testing.T) *Game {
	t.Helper()
	return reachCommandPhaseOnMapWith(t, loadWithRules(t))
}

// loadWithRules ＝ `load` 再補上決策層與艦隊命令要的規則表：星圖、艦種表、
// 兩張編制表、預設副官、訊息框與艦隊視窗的版面。`load()` 只給畫面要的
// 東西，正式入口 `internal/bootstrap` 才會載這幾份；少了艦種表，
// 「出擊命令」編出來的艦隊就沒有編制。
func loadWithRules(t *testing.T) *Resources {
	t.Helper()
	res := load(t)
	graph, err := rules.LoadStarGraph(filepath.Join(root,
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	res.Graph = graph
	data := filepath.Join(root, "docs/internals/data")
	if res.ShipKinds, err = rules.LoadShipKindTable(
		filepath.Join(data, "ship-kind-table.json")); err != nil {
		t.Fatal(err)
	}
	if res.DefaultAdjutants, err = rules.LoadDefaultAdjutants(
		filepath.Join(data, "default-adjutants.json")); err != nil {
		t.Fatal(err)
	}
	if res.Composition, err = rules.LoadCompositionTables(
		filepath.Join(data, "fleet-composition-templates.json"),
		filepath.Join(data, "ship-kind-fallback.json")); err != nil {
		t.Fatal(err)
	}
	if res.MessageBox, err = screen.LoadMessageBoxSpec(
		filepath.Join(data, "message-box.json")); err != nil {
		t.Fatal(err)
	}
	if res.FleetWindow, err = screen.LoadFleetWindowSpec(
		filepath.Join(data, "fleet-window.json")); err != nil {
		t.Fatal(err)
	}
	if res.FleetActionMenu, err = screen.LoadCommandMenuSpec(
		filepath.Join(data, "fleet-command-menu.json")); err != nil {
		t.Fatal(err)
	}
	return res
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
// 維持費那道閘門接上之後建了 **4 支**（原版在同一個局面是 3 支）。
// 差的那一支來自另一個已知不忠實：`Support.Money` 回的是**玩家**那一方
// 的稅收（10341），而原版看的是同盟自己的（10599 → 7852）——remake
// 只追一方的財政（`docs/re/218`）。
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
	// 維持費那道閘門要真的收得了手——不能再一路建到上限。
	if len(g.fleets) >= ai.FleetCap {
		t.Errorf("建到 %d 支（上限 %d）——維持費那道閘門沒收手",
			len(g.fleets), ai.FleetCap)
	}
	// 原版在同一個局面是 3 支；remake 拿玩家的稅收當預算，所以多一支。
	if len(g.fleets) != 4 {
		t.Logf("建了 %d 支（原版 3、目前預期 4）——財政那一側是不是改了？",
			len(g.fleets))
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

// TestOriginalBrainIsDefault 釘住預設就是原版那套戰略 AI
// （`docs/spec/029` READY，使用者裁決 2026-09-10）。
//
// 這一支原本叫 `TestOriginalBrainStillNotDefault`，斷言 `g.brain == nil`
// ——那是「還沒裁決」時的狀態。
func TestOriginalBrainIsDefault(t *testing.T) {
	g := reachBrainPhase(t)
	if _, ok := g.brain.(ai.OriginalStrategy); !ok {
		t.Errorf("預設的 Brain 是 %T，預期 ai.OriginalStrategy", g.brain)
	}
	// `ai.Original`（那個空殼）與 nil 仍然同義：兩者都一道命令都不下。
	// 預設換掉的是 `Game` 接哪一個腦，不是這個型別的行為。
	if got := (ai.Original{}).Orders(ai.World{}, 0, 0); got != nil {
		t.Errorf("ai.Original 回了 %v，預期 nil", got)
	}
}

// TestFirstAITurnMatchesReceipt 拿接好線的原版 AI 跑第一個回合，逐支比
// 收據 `turn260mem`（`docs/re/209`、`docs/re/221`）。
//
// 三支**完全對上**：支數、提督、類別鍵、目的地都與原版相同。
// 支數要接上 `Blend` 與稅率重算才會是三支（`docs/spec/030`）——AI
// 回合開頭把 `Blend` 設成 1、全國稅率重算、稅收從 10599 掉到 7852，
// 閘門就在第三支之後收手。
func TestFirstAITurnMatchesReceipt(t *testing.T) {
	g := reachBrainPhase(t)
	g.SetBrain(g.OriginalBrain())
	g.Click(544, 125)
	g.Click(560, 125)

	if len(g.fleets) < 3 {
		t.Fatalf("只建了 %d 支", len(g.fleets))
	}
	// 收據上三支：提督 47／52／55，類別鍵 0C／04／0B。
	// 建出來的順序是類別 0 → 1 → 2（開局四類平手照 0 1 2 3），
	// 亂數從 D3DC167E 起算得到鍵 04／0B／0C。
	// 運用（`+0x47`）、維持費（`+0x0A`）與編制都是收據上的
	// （`docs/re/218`、`docs/re/224`、`docs/re/225`）。編制逐組寫成
	// `{艦種, 艘數, 隊形槽}`，順序就是原版排序之後的。
	want := []struct {
		admiral, key, target, ops, upkeep int
		comp                              [][3]int
	}{
		// 階段二打無主
		{52, 0x04, 24, 128, 85, [][3]int{
			{0, 3000, 20}, {0, 3000, 10}, {0, 3000, 12}, {6, 1000, 29}}},
		// 同上，湊滿兩支
		{55, 0x0B, 24, 120, 124, [][3]int{
			{1, 3000, 20}, {1, 3000, 10}, {1, 3000, 11}, {1, 3000, 12}}},
		// 階段三的指派
		{47, 0x0C, 15, 102, 217, [][3]int{
			{0, 3000, 20}, {0, 3000, 10}, {3, 2000, 5}, {3, 2000, 8},
			{5, 3000, 29}, {6, 1000, 34}}},
	}
	for i, w := range want {
		f := g.fleets[i]
		if f.Admiral != w.admiral || f.ClassKey != w.key || f.Target != w.target {
			t.Errorf("第 %d 支 ＝ 提督 %d／鍵 %02X／目的地 %d，收據是 %d／%02X／%d",
				i, f.Admiral, f.ClassKey, f.Target, w.admiral, w.key, w.target)
		}
		if f.Morale != rules.NewFleetMorale {
			t.Errorf("第 %d 支的士氣 ＝ %d，新編艦隊應該是 100", i, f.Morale)
		}
		if f.Operations != w.ops {
			t.Errorf("第 %d 支的運用 ＝ %d，收據是 %d", i, f.Operations, w.ops)
		}
		if u := rules.FleetUpkeep(g.res.ShipKinds, 1, f.Composition, f.Operations); u != w.upkeep {
			t.Errorf("第 %d 支的維持費 ＝ %d，收據是 %d", i, u, w.upkeep)
		}
		if len(f.Composition) != len(w.comp) {
			t.Errorf("第 %d 支有 %d 組編制，收據是 %d", i, len(f.Composition), len(w.comp))
			continue
		}
		for j, c := range w.comp {
			got := f.Composition[j]
			if got.Kind != c[0] || got.Ships != c[1] || got.Slot3 != c[2] {
				t.Errorf("第 %d 支第 %d 組 ＝ 艦種 %d／%d 艘／槽 %d，收據是 %d／%d／%d",
					i, j, got.Kind, got.Ships, got.Slot3, c[0], c[1], c[2])
			}
		}
	}
	// 合計就是收據的 `ds:8F72[同盟]`。
	if got, want := g.fleetUpkeep(1), 426; got != want {
		t.Errorf("同盟的艦隊維持費 ＝ %d，收據是 %d", got, want)
	}
	// 這一回合結束會扣兩筆：維持費（原版 `056A:09EB`，逐支 217／85／124）
	// 與移動花費（`113A:0729`，逐支 263／95／134），`docs/re/236`、`237`。
	// 收據 `monthend` 監看到 45848 → 45422 → 44930，而 44930 正是第二個
	// AI 回合那份傾印（`refit2_pre.bin`）裡 `ds:8F52[同盟]` 的值。
	if got, want := g.military[1], 45848-426-492; got != want {
		t.Errorf("扣完維持費與移動花費的同盟軍事費 ＝ %d，收據是 %d", got, want)
	}
	if len(g.fleets) != 3 {
		t.Errorf("建了 %d 支，收據是 3", len(g.fleets))
	}
}

// TestTaxRateRecomputeMatchesReceipt 釘 `090A:0B93`：AI 回合開頭
// 按 `Blend` 重算全國稅率之後，同盟的預估徵稅額要與收據相同
// （`docs/re/223`、`docs/spec/030`）。
//
//	等級 2（開局值）→ 10599      `t260_pre` 的 ds:8F6A[同盟]
//	等級 1（軍事費 45848）→ 7852 `t260_post` 的同一格
func TestTaxRateRecomputeMatchesReceipt(t *testing.T) {
	g := reachBrainPhase(t)
	if got, want := g.military[1], 45848; got != want {
		t.Errorf("同盟的軍事費 ＝ %d，收據是 %d", got, want)
	}
	if got, want := rules.BlendForFunds(g.military[1]), 1; got != want {
		t.Errorf("同盟的 Blend ＝ %d，預期 %d", got, want)
	}
	for _, c := range []struct{ level, tax int }{{2, 10599}, {1, 7852}} {
		g.refreshBlend(1, c.level)
		if got := g.taxTotal[1]; got != c.tax {
			t.Errorf("等級 %d 的總稅收 ＝ %d，收據是 %d", c.level, got, c.tax)
		}
	}
	// 帝國那一邊不跑 AI，重算同盟不該碰到它。
	if got, want := g.taxTotal[0], 10341; got != want {
		t.Errorf("帝國的總稅收 ＝ %d，收據是 %d", got, want)
	}
}

// TestMonthEndKeepsBlend 釘 `2D10:0319` 在 AI 回合結尾的月底分支
// （收據 `monthblend`／`docs/re/238`）。
//
// 平時（`回合數 % 10 != 9`）結尾把 `Blend` 歸零，稅率跟著掉到等級 0；
// 月底那一條照軍事費重算，等級留在 1。同一條序列、同一個時點，
// 原版只差一個回合計數器：對照組 `Blend[同盟] = 0`／總稅收 6696，
// 月底 `Blend[同盟] = 1`／總稅收 10213。`advanceTurn` 開頭已先把
// `g.turn` 加一，所以收據 poke 的原版計數器 9 對應呼叫前的 `g.turn = 9`，
// 不是舊測試寫的 8（`docs/re/299`）。
func TestMonthEndKeepsBlend(t *testing.T) {
	run := func(before int) (turn, blend, tax int) {
		g := reachBrainPhase(t)
		g.SetBrain(g.OriginalBrain())
		g.turn = before
		g.Click(544, 125)
		g.Click(560, 125)
		return g.turn, g.blend[1], g.taxTotal[1]
	}
	plainTurn, plainBlend, plainTax := run(0)
	monthTurn, monthBlend, monthTax := run(9)

	if (plainTurn-1)%rules.TurnsPerMonth == rules.TurnsPerMonth-1 ||
		(monthTurn-1)%rules.TurnsPerMonth != rules.TurnsPerMonth-1 {
		t.Fatalf("原版計數器不對：平時 %d、月底 %d", plainTurn-1, monthTurn-1)
	}
	if plainBlend != 0 {
		t.Errorf("平時結尾的 Blend ＝ %d，原版歸零", plainBlend)
	}
	if monthBlend != 1 {
		t.Errorf("月底結尾的 Blend ＝ %d，收據是 1（軍事費落在 40000–50000）", monthBlend)
	}
	// 等級 0 的稅率比等級 1 低，所以平時那一條的預估徵稅額一定比較小
	// （收據：6696 對 10213）。
	if !(plainTax < monthTax) {
		t.Errorf("平時的總稅收 %d 沒有比月底的 %d 小", plainTax, monthTax)
	}
}

// TestExecuteTurnWithBrainMatchesOracle 是 `TestExecuteTurnMatchesOracle`
// 的**有腦版本**：同一條玩家路徑、同一批收據，但走 `reachBrainPhase`
// ——那個入口有載星域圖，所以戰略 AI 真的會跑。
//
// `docs/spec/029` 記著這個陷阱：`runBrain` 在 `Resources.Graph` 是 nil 時
// 整段跳過，而 `internal/game` 共用的 `load()` 沒有載圖。所以「換掉預設
// 之後 `go test` 還是綠」在那些測試裡**不代表決策層被跑過**。
func TestExecuteTurnWithBrainMatchesOracle(t *testing.T) {
	for _, batch := range []string{"turn260", "turn320", "turn450"} {
		t.Run(batch, func(t *testing.T) {
			g := reachBrainPhase(t)
			if g.brain == nil {
				t.Fatal("這個入口應該有接上 Brain")
			}
			g.Click(544, 125) // 系統指令
			g.Click(560, 125) // 實行
			if g.State() != StateCommandPhase {
				t.Fatalf("實行之後在 %v", g.State())
			}
			matches(t, g, batch)
		})
	}
}
