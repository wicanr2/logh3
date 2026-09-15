package ai

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 可達表（`2B99:0392`，規格 `docs/spec/043`）對兩份原版局面。
//
// 擁有者字串是星域 0–31 的 `+0x12`（0 帝國、1 同盟、2 無主），從 1 MB
// 傾印讀出來的（`tools/scripts/reach_table.py`，`docs/re/298`）：
//
//   - 新遊戲：`t260_post.bin`（收據 `turn260mem`，`#214,000,000`，第一個
//     AI 回合剛跑完）。那一刻 EMS 分頁框裡四張路徑表與照這份擁有者重算
//     的**逐格相同**（兩個陣營的距離與中點各 1024／1024），所以下面的
//     可達表是執行期的值，不只是重算的值。
//   - `s3aicomp.bin`（收據 `s3aicomp`，`docs/re/176`）：同盟跑 AI、
//     `ds:D50E` 拿到分數的四格是 13／15／22／28。
const (
	ownersNewGame  = "00020000002220211111111222112111"
	ownersS3AIComp = "00000000000000001111110010210111"
)

func ownerFromString(s string) func(int) int {
	return func(r int) int { return int(s[r] - '0') }
}

func reachString(t [RegionCount]bool) string {
	b := make([]byte, RegionCount)
	for i, v := range t {
		b[i] = '0'
		if v {
			b[i] = '1'
		}
	}
	return string(b)
}

// TestCapitalReachNewGame 對 `t260_post.bin` 分頁框裡的距離表：同盟從
// 巴拉特出發，帝國那一側只有伊謝爾倫（13，距離 1090）走得到；帝國從
// 瓦爾哈拉出發，同盟那一側一格都走不到。
func TestCapitalReachNewGame(t *testing.T) {
	g := testGraph(t)
	owner := ownerFromString(ownersNewGame)
	for faction, want := range map[int]string{
		0: "11111111111101111000000000000000",
		1: "00000000000001111111111110111111",
	} {
		if got := reachString(CapitalReach(g, owner, faction)); got != want {
			t.Errorf("陣營 %d 的可達表\n  得到 %s\n  原版 %s", faction, got, want)
		}
	}
	// 同盟那一列的距離整列逐格對（分頁框 E0800 起第 31 列）。
	owners := make([]int, RegionCount)
	for i := range owners {
		owners[i] = owner(i)
	}
	dist, _ := g.BuildTables(owners, 1)
	want := []int{4220, 4223, 3181, 3181, 4218, 4209, 2144, 2144, 2128, 3167, 4208, 2127,
		3169, 1090, 1104, 63, 50, 39, 24, 14, 10, 36, 51, 1093, 1070, 2112, 39, 29, 1073,
		34, 16, 0}
	for i, w := range want {
		if dist[31][i] != w {
			t.Errorf("同盟 距離(31, %d) ＝ %d，分頁框裡是 %d", i, dist[31][i], w)
		}
	}
}

// TestAttackBoardNewGameDrawsOnce 是這一條要修的那件事：新遊戲第一個
// AI 回合，同盟的 `ds:D50E` 只有 `[13] = 4096`（`docs/re/209`），
// `AttackNeed` 因此只抽**一次**亂數。可達表當成全 1 的時候是 10 格、
// 抽 10 次（`docs/re/297`）。
func TestAttackBoardNewGameDrawsOnce(t *testing.T) {
	g := testGraph(t)
	routes, err := LoadAdvanceRoutes(filepath.Join("..", "..", "docs", "internals", "data", "ai-advance-routes.json"))
	if err != nil {
		t.Fatalf("讀進軍路線表：%v", err)
	}
	owner := ownerFromString(ownersNewGame)
	state := StateCode(owner(Iserlohn), owner(Phezzan), false)
	r := rules.NewRand(0xD3DC167E)
	b := AttackBoard(TurnInput{Faction: 1, Graph: g, Owner: owner, Rand: r},
		routes, state, [RegionCount]int{})
	for i := 0; i < RegionCount; i++ {
		want := 0
		if i == Iserlohn {
			want = 4096
		}
		if b.Priority[i] != want {
			t.Errorf("ds:D50E[%d] ＝ %d，原版 %d", i, b.Priority[i], want)
		}
	}
	n := 0
	for s := rules.NewRand(0xD3DC167E); s.Seed() != r.Seed() && n < 100; n++ {
		s.Next()
	}
	if n != 1 {
		t.Errorf("階段四抽了 %d 次亂數，原版 1 次", n)
	}
}

// TestCapitalReachMatchesS3AIComp 是正對照：把 `TestAttackScoreMatchesOracle`
// 手填的候選集合換成照擁有者算出來的，要得到同一組四格、同一組分數。
func TestCapitalReachMatchesS3AIComp(t *testing.T) {
	g := testGraph(t)
	owner := ownerFromString(ownersS3AIComp)
	reach := CapitalReach(g, owner, 1)
	enemyCapital := Capitals[0]
	cand := func(i int) bool { return reach[i] && (owner(i) == 0 || i == enemyCapital) }
	var got []int
	for i := 0; i < RegionCount; i++ {
		if cand(i) {
			got = append(got, i)
		}
	}
	if want := []int{13, 15, 22, 28}; len(got) != len(want) ||
		got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		t.Fatalf("候選 ＝ %v，收據 `s3aicomp` 是 %v", got, want)
	}
	routes, err := LoadAdvanceRoutes(filepath.Join("..", "..", "docs", "internals", "data", "ai-advance-routes.json"))
	if err != nil {
		t.Fatalf("讀進軍路線表：%v", err)
	}
	score := AttackScore(routes.For(1, StateCode(0, 0, false)), ThreatMap(s3aicompEmpire, 0, nil), cand)
	want := map[int]int{13: 682, 15: 2048, 22: 2048, 28: 2048}
	for i := 0; i < RegionCount; i++ {
		if score[i] != want[i] {
			t.Errorf("ds:D50E[%d] ＝ %d，收據是 %d", i, score[i], want[i])
		}
	}
}

// TestCapitalReachIsDirectionFree 釘住 `CapitalReach` 註解裡那句「從首都
// 往外查或往首都查都一樣」：距離表不對稱，但跨不跨 `0x800` 是對稱的。
// 階段二（`2E54:0F09`）查的是 `2B99:028F(i, 首都, …)`，方向與可達表相反。
func TestCapitalReachIsDirectionFree(t *testing.T) {
	g := testGraph(t)
	for _, s := range []string{ownersNewGame, ownersS3AIComp} {
		owners := make([]int, RegionCount)
		for i := range owners {
			owners[i] = int(s[i] - '0')
		}
		for faction := 0; faction < 2; faction++ {
			dist, _ := g.BuildTables(owners, faction)
			for i := 0; i < RegionCount; i++ {
				for j := 0; j < RegionCount; j++ {
					if (dist[i][j] < ReachLimit) != (dist[j][i] < ReachLimit) {
						t.Errorf("陣營 %d：%d→%d ＝ %d、%d→%d ＝ %d，跨 0x800 不對稱",
							faction, i, j, dist[i][j], j, i, dist[j][i])
					}
				}
			}
		}
	}
}

// TestReachFromCapitalMatchesCapitalReach 從首都起算的 ReachFrom 就是
// CapitalReach；起點換成別的星域時表會不同（伊謝爾倫出發看得到同盟那邊）。
func TestReachFromCapitalMatchesCapitalReach(t *testing.T) {
	g := testGraph(t)
	owner := ownerFromString(ownersNewGame)
	for faction := 0; faction < 2; faction++ {
		if a, b := ReachFrom(g, owner, faction, Capitals[faction]), CapitalReach(g, owner, faction); a != b {
			t.Errorf("陣營 %d：ReachFrom(首都) %s ≠ CapitalReach %s", faction, reachString(a), reachString(b))
		}
	}
	// 新遊戲帝國從伊謝爾倫（13）出發的表與從首都的**相同**——同盟那一側
	// 緊貼著的 15／16 從首都就已經在 k ≤ 1 裡。起點推進到亞斯提（16）才
	// 看得到差別：傑姆席德（21）從 16 只要罰一條邊，從首都要罰兩條。
	if from13 := ReachFrom(g, owner, 0, 13); from13 != CapitalReach(g, owner, 0) {
		t.Errorf("從伊謝爾倫出發的可達表 %s 與從首都的不同", reachString(from13))
	}
	from16 := ReachFrom(g, owner, 0, 16)
	if !from16[21] || CapitalReach(g, owner, 0)[21] {
		t.Errorf("傑姆席德（21）：從亞斯提出發 %v、從首都 %v；要 true／false", from16[21], CapitalReach(g, owner, 0)[21])
	}
}

// TestReachFromMatchesFleetCommandReceipts 是玩家派艦隊那條路的允許表
// （`docs/spec/051`）：新遊戲帝國第１艦隊在瓦爾哈拉（0），原版讓玩家
// 移動到亞爾提那（2）、香陶（6）、伊謝爾倫（13），對費沙（14）與巴拉特（31）
// 回「這是無法前往的星系。」（收據 `fcmd-confirm`／`fcx-mv-*`／`fcmd-far`）。
// 允許 ＝ 可達 ∧（擁有者是自己 ∨ 現在位置）。
func TestReachFromMatchesFleetCommandReceipts(t *testing.T) {
	g := testGraph(t)
	owner := ownerFromString(ownersNewGame)
	reach := ReachFrom(g, owner, 0, 0)
	allowed := func(r int) bool { return reach[r] && (owner(r) == 0 || r == 0) }
	for r, want := range map[int]bool{2: true, 6: true, 13: true, 14: false, 31: false, 0: true} {
		if allowed(r) != want {
			t.Errorf("星域 %d 允許＝%v，原版是 %v（可達 %v、擁有者 %d）", r, allowed(r), want, reach[r], owner(r))
		}
	}
}
