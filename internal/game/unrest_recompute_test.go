package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// 未安定度的同狀態對拍（`docs/spec/042`）：remake 從新遊戲走同一條玩家
// 路徑、在同一個時點做同樣的注入，`+0x7D` 要與原版收據逐格相同。
//
// 原版那一端是收據 `rbx292`／`rbx292d`／`rbx292e`（`docs/re/292`、`297`）：
// 新遊戲 → 系統指令 → 實行，`#211,000,000`（第一個 AI 回合之前）注入
//
//	rbx292   提督 13 的營運 ← 1，星域 1 的 `+0x7D` ← 0xFF
//	rbx292d  提督 13 的營運 ← 1
//	rbx292e  提督 13 的營運 ← 4
//
// `rules` 那一層拿收據自己的傾印驗算式；這一層驗的是 **remake 自己算出來的
// 輸入**（Blend、`ds:8FF8`、內政參謀、艦隊、旗標、路徑表、回合計數器）
// 在原版重算的那幾個時點與原版相同。

const (
	unrestDumpDS      = 0x54B30
	unrestDumpRegions = unrestDumpDS + 0xA79C
)

func readUnrestReceipt(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join(root, "workplace", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	return b
}

// unrestTurnSetup 走到第一個「實行」之前，照正式入口接上原版戰略 AI
// （含進軍路線表），再做收據的注入。
func unrestTurnSetup(t *testing.T, ops byte, pokeRegion1 bool) *Game {
	t.Helper()
	g := reachBrainPhase(t)
	routes, err := ai.LoadAdvanceRoutes(filepath.Join(root, "docs/internals/data/ai-advance-routes.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.AdvanceRoutes = routes
	g.SetBrain(g.OriginalBrain())
	if g.brain == nil {
		t.Fatal("這個入口應該接上原版戰略 AI")
	}
	const intelEmpire = 13 // 帝國的情報專任參謀（`docs/internals/30`）
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, 0); got != intelEmpire {
		t.Fatalf("帝國的情報專任參謀是提督 %d，開局應該是 13", got)
	}
	g.mutableAdmirals()[intelEmpire*rules.AdmiralRecord+0x0E] = ops
	if pokeRegion1 {
		g.regionRecord(1)[rules.RegionUnrest] = 0xFF
	}
	return g
}

func compareUnrest(t *testing.T, g *Game, want []byte) {
	t.Helper()
	bad := 0
	for i := 0; i < rules.RegionCount; i++ {
		got := g.regionRecord(i)[rules.RegionUnrest]
		w := want[i*rules.RegionRecord+rules.RegionUnrest]
		if got != w {
			bad++
			t.Errorf("星域 %d：remake %d，原版 %d", i, got, w)
		}
	}
	if bad > 0 {
		t.Fatalf("32 格裡 %d 格不符", bad)
	}
}

// TestUnrestSameStateAfterAITurn 是 worklist `region-unrest-recompute` 的判準：
// 第一個 AI 回合跑完的那一刻，32 格的 `+0x7D` 與 `rbx292_b.bin`
// （`#216,000,000`）逐格相同。
//
// 這一刻夾在 `advanceTurn` 的第二步（`runBrain`）與第三步（扣維持費）之間，
// 正式入口停不下來，所以照 `advanceTurn` 的前兩步手動跑。星域 1 被 poke 成
// 255，重算之後要回到 38（正對照：真的整份重算過）；同盟 13 格各降 4–5。
func TestUnrestSameStateAfterAITurn(t *testing.T) {
	g := unrestTurnSetup(t, 1, true)
	mem := readUnrestReceipt(t, "rbx292_b.bin")
	before := append([]byte(nil), g.res.Stellar...)

	g.turn++ // advanceTurn 的第一步
	g.runBrain()

	// 輸入先對：Blend、內政專任參謀（原版 `ds:8FD4`／`ds:8F4F`）。
	if g.blend != [2]int{int(mem[unrestDumpDS+0x8FD4]), int(mem[unrestDumpDS+0x8FD5])} {
		t.Errorf("Blend ＝ %v，收據是 %v", g.blend, mem[unrestDumpDS+0x8FD4:unrestDumpDS+0x8FD6])
	}
	for f := 0; f < 2; f++ {
		if got, want := g.taxRateAdmiral(f), int(mem[unrestDumpDS+0x8F4F+f]); got != want {
			t.Errorf("陣營 %d 的內政專任參謀是提督 %d，收據是 %d", f, got, want)
		}
	}
	compareUnrest(t, g, mem[unrestDumpRegions:])
	changed := 0
	for i := 0; i < rules.RegionCount; i++ {
		o := i*rules.RegionRecord + rules.RegionUnrest
		if before[o] != g.res.Stellar[o] {
			changed++
		}
	}
	if changed != 14 {
		t.Errorf("AI 回合前後 %d 格變了，收據是 14 格（同盟 13 ＋ 星域 1）", changed)
	}
}

// TestUnrestSameStateAfterTurn 是同一條序列跑完整個「實行」：`rbx292d`
// （營運 1，九格叛亂）在 `#239,000,000` 倒的星域表。叛亂、依行星重算擁有者、
// 路徑表重建都走過之後，32 格逐格相同——九格無主是 100，其餘 23 格照新的
// 鄰居、新的路徑表與**進位之後的回合計數器**重算。
func TestUnrestSameStateAfterTurn(t *testing.T) {
	g := unrestTurnSetup(t, 1, false)
	want := readUnrestReceipt(t, "rbx292d_reg_post.bin")
	g.Click(544, 125) // 系統指令
	g.Click(560, 125) // 實行
	dismissAllRebellionReports(g)
	if g.State() != StateCommandPhase {
		t.Fatalf("實行之後在 %v", g.State())
	}
	if g.turn != 1 {
		t.Fatalf("回合數 %d，應該是 1", g.turn)
	}
	compareUnrest(t, g, want)
}

// TestUnrestSameStateRebellionReads 是讀的那一端：叛亂的門檻讀的是命令階段
// 那一串開頭（`056A:09EB` → `090A:0E6D`）重算的值。收據 `rbx292e`（營運 4）
// 問了 19 格，remake 每一格的門檻都要等於拿原版那一刻的未安定度
// （`rbx292e_pre.bin`）與原版的營運算出來的門檻——**同盟那 10 格也是**。
func TestUnrestSameStateRebellionReads(t *testing.T) {
	g := unrestTurnSetup(t, 4, false)
	mem := readUnrestReceipt(t, "rbx292e_pre.bin")
	g.Click(544, 125)
	g.Click(560, 125)
	dismissAllRebellionReports(g)
	if g.State() != StateCommandPhase {
		t.Fatalf("實行之後在 %v", g.State())
	}
	adm := mem[unrestDumpDS+0xB77C:]
	var ops [2]int
	for f := 0; f < 2; f++ {
		ops[f] = rules.AdmiralOps(adm[:82*rules.AdmiralRecord], int(mem[unrestDumpDS+0x8F4D+f]))
	}
	asked := [2]int{}
	for i, roll := range g.rebellionRolls {
		if !roll.Asked {
			continue
		}
		rec := mem[unrestDumpRegions+i*rules.RegionRecord:]
		owner := int(rec[0x12])
		if owner > 1 {
			t.Errorf("星域 %d：原版那一刻是無主，remake 卻問了", i)
			continue
		}
		asked[owner]++
		want := rules.RebellionThreshold(int(rec[rules.RegionUnrest]), ops[owner], 0)
		if roll.Threshold != want {
			t.Errorf("星域 %d（陣營 %d）：門檻 %d，原版的未安定度 %d 與營運 %d 算出 %d",
				i, owner, roll.Threshold, rec[rules.RegionUnrest], ops[owner], want)
		}
	}
	if asked != [2]int{9, 10} {
		t.Errorf("問了帝國 %d 格、同盟 %d 格，收據是 9／10", asked[0], asked[1])
	}
}
