package rules

import (
	"path/filepath"
	"testing"
)

func loadGraph(t *testing.T) *StarGraph {
	t.Helper()
	g, err := LoadStarGraph(filepath.Join("../..", "docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// TestStarGraphIsSymmetric：相鄰是雙向的。
func TestStarGraphIsSymmetric(t *testing.T) {
	g := loadGraph(t)
	for i, ns := range g.Adjacency {
		for _, j := range ns {
			found := false
			for _, k := range g.Adjacency[j] {
				if k == i {
					found = true
				}
			}
			if !found {
				t.Errorf("星域 %d 說它連到 %d，但 %d 沒有連回來", i, j, j)
			}
		}
	}
}

// TestStarRoutesAreWalkable：照原版的表走，每一步都必須是相鄰的，
// 而且 1024 對都走得出來。兩個陣營各一張表。
func TestStarRoutesAreWalkable(t *testing.T) {
	g := loadGraph(t)
	for _, c := range []struct{ faction, longest int }{{0, 10}, {1, 12}} {
		faction := c.faction
		longest := 0
		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				p := g.Route(i, j, faction)
				if p == nil {
					t.Fatalf("陣營 %d：%d 走不到 %d", faction, i, j)
				}
				for k := 0; k+1 < len(p); k++ {
					ok := false
					for _, nb := range g.Neighbours(p[k]) {
						if nb == p[k+1] {
							ok = true
						}
					}
					if !ok {
						t.Fatalf("陣營 %d：%d→%d 的路經過不相鄰的 %d-%d",
							faction, i, j, p[k], p[k+1])
					}
				}
				if len(p)-1 > longest {
					longest = len(p) - 1
				}
			}
		}
		if longest != c.longest {
			t.Errorf("陣營 %d 最長 %d 步，預期 %d", faction, longest, c.longest)
		}
	}
}

// TestStarRoutesAreNotAlwaysShortest 釘住一件會誤導人的事：
// **原版的路不一定是最短路**。同一張圖上廣度優先最遠 9 步，
// 照原版的表走最遠 12 步——remake 要跟原版一致就得照表走，不能自己算最短路。
func TestStarRoutesAreNotAlwaysShortest(t *testing.T) {
	g := loadGraph(t)
	worst := 0
	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if d := g.Steps(i, j, 0); d > worst {
				worst = d
			}
		}
	}
	if worst <= 9 {
		t.Fatalf("照表走最遠 %d 步，若真的都是最短路就沒有這條測試的必要", worst)
	}
}

// TestCapitalRegions：兩個首都。
func TestCapitalRegions(t *testing.T) {
	if CapitalRegion(0) != 0 || CapitalRegion(1) != 31 {
		t.Fatalf("首都是 %d／%d，預期 0（瓦爾哈拉）／31（巴拉特）",
			CapitalRegion(0), CapitalRegion(1))
	}
}

// TestNextStepMatchesRoute 驗 `NextStep`（原版 `2B99:0302` 的定點迭代）
// 與 `Route` 展開之後的第一步是同一個值——兩張中點矩陣、所有走得通的
// 星域對（`docs/re/183`）。
func TestNextStepMatchesRoute(t *testing.T) {
	g := loadGraph(t)
	pairs := 0
	for faction := 0; faction < 2; faction++ {
		for from := 0; from < 32; from++ {
			for to := 0; to < 32; to++ {
				if from == to {
					continue
				}
				p := g.Route(from, to, faction)
				if len(p) < 2 {
					continue
				}
				if got := g.NextStep(from, to, faction); got != p[1] {
					t.Fatalf("陣營 %d %d→%d：NextStep ＝ %d，Route[1] ＝ %d", faction, from, to, got, p[1])
				}
				pairs++
			}
		}
	}
	if pairs < 1800 {
		t.Fatalf("只比到 %d 對，太少——中點矩陣可能沒載進來", pairs)
	}
}

// TestStarGraphFullyConnected 釘住一個 `internal/ai` 依賴的事實：
// 星圖是全連通的，所以原版 `2B99:0392`（從首都能不能走到）在正常局面下
// 填的是全 1（`docs/re/185`）。
func TestStarGraphFullyConnected(t *testing.T) {
	g := loadGraph(t)
	for _, capital := range []int{0, 31} {
		n := 0
		for i := 0; i < 32; i++ {
			if g.Hops(capital, i) >= 0 {
				n++
			}
		}
		if n != 32 {
			t.Errorf("從星域 %d 只走得到 %d／32 格", capital, n)
		}
	}
	for faction := 0; faction < 2; faction++ {
		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				if i != j && g.Route(i, j, faction) == nil {
					t.Fatalf("陣營 %d 的中點矩陣解不開 %d→%d", faction, i, j)
				}
			}
		}
	}
}
