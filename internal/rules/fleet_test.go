package rules

import "testing"

func testGraph(t *testing.T) *StarGraph {
	t.Helper()
	g, err := LoadStarGraph("../../docs/internals/data/star-graph.json")
	if err != nil {
		t.Skipf("讀不到星域圖：%v", err)
	}
	return g
}

// TestAdvanceMatchesS3 拿收據 s3ai0／s3ai6 量到的三組移動當案例
// （`docs/re/77`）：三組都是一步之遙，六回合之內走完，到了之後
// 命令要被清掉。
func TestAdvanceMatchesS3(t *testing.T) {
	g := testGraph(t)
	cases := []struct{ from, to, faction int }{
		{23, 25, 0}, // 帝國第 8／14 支：波列多 → 干達巴爾
		{28, 26, 0}, // 帝國第 12 支：巴米利恩 → 達希利
		{27, 26, 1}, // 同盟六支：萊加爾 → 達希利
	}
	for _, c := range cases {
		f := Fleet{Faction: c.faction, Region: c.from, Target: c.to}
		if !g.Advance(&f) {
			t.Fatalf("%d→%d 沒有動", c.from, c.to)
		}
		if f.Region != c.to {
			t.Fatalf("%d→%d 走一格之後在 %d", c.from, c.to, f.Region)
		}
		if f.Target != NoTarget {
			t.Fatalf("到了 %d 命令沒清掉（還是 %d）", c.to, f.Target)
		}
		if g.Advance(&f) {
			t.Fatal("沒有命令卻還在動")
		}
	}
}

// TestAdvanceWalksTheOriginalRoute 走多格時要照原版那張路徑表走，
// 不是相鄰圖上的最短路（`docs/internals/26`）。
func TestAdvanceWalksTheOriginalRoute(t *testing.T) {
	g := testGraph(t)
	const from, to = 0, 31 // 瓦爾哈拉 → 巴拉特，橫跨整張圖
	want := g.Route(from, to, 0)
	if len(want) < 3 {
		t.Fatalf("路徑只有 %d 格，測不出多格移動", len(want))
	}
	f := Fleet{Faction: 0, Region: from, Target: to}
	var got []int
	for got = append(got, f.Region); g.Advance(&f); {
		got = append(got, f.Region)
		if len(got) > 40 {
			t.Fatal("走不完")
		}
	}
	if len(got) != len(want) {
		t.Fatalf("走了 %v，路徑表是 %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 格走到 %d，路徑表是 %d", i, got[i], want[i])
		}
	}
	if f.Target != NoTarget {
		t.Fatal("到了目的地命令沒清掉")
	}
}

// TestCountFleetsMatchesS3 星域記錄的艦隊數是重數出來的（`docs/re/40`）。
// 收據 s3ai6 量到達希利是帝國 1 支、同盟 6 支。
func TestCountFleetsMatchesS3(t *testing.T) {
	fleets := []Fleet{
		{Faction: 0, Region: 26, Target: NoTarget},
		{Faction: 1, Region: 26, Target: NoTarget},
		{Faction: 1, Region: 26, Target: NoTarget},
		{Faction: 1, Region: 26, Target: NoTarget},
		{Faction: 1, Region: 26, Target: NoTarget},
		{Faction: 1, Region: 26, Target: NoTarget},
		{Faction: 1, Region: 26, Target: NoTarget},
		{Faction: 0, Region: 25, Target: NoTarget},
	}
	if n := CountFleets(fleets, 26, 0); n != 1 {
		t.Fatalf("達希利的帝國艦隊數 %d，預期 1", n)
	}
	if n := CountFleets(fleets, 26, 1); n != 6 {
		t.Fatalf("達希利的同盟艦隊數 %d，預期 6", n)
	}
	if n := CountFleets(fleets, 25, 0); n != 1 {
		t.Fatalf("干達巴爾的帝國艦隊數 %d，預期 1", n)
	}
}
