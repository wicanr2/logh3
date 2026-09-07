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
