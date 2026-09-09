package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// 收據：`docs/internals/data/star-distance.json`（EMS 頁框傾印，`docs/re/241`）
// 與 `star-graph.json` 的兩張中點矩陣。拿同一份傾印裡的星域擁有者重建，
// 四張表 1024×4 格必須逐格相同（`docs/re/243`）。
type starDistanceFile struct {
	Owners   []int   `json:"owners"`
	Empire   [][]int `json:"empire"`
	Alliance [][]int `json:"alliance"`
}

func loadStarDistance(t *testing.T) (*StarGraph, starDistanceFile) {
	t.Helper()
	dir := filepath.Join("..", "..", "docs", "internals", "data")
	g, err := LoadStarGraph(filepath.Join(dir, "star-graph.json"))
	if err != nil {
		t.Fatalf("讀星域圖：%v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "star-distance.json"))
	if err != nil {
		t.Fatalf("讀距離表：%v", err)
	}
	var d starDistanceFile
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatalf("解距離表：%v", err)
	}
	return g, d
}

func TestBuildTablesMatchesOracle(t *testing.T) {
	g, want := loadStarDistance(t)
	for _, c := range []struct {
		name    string
		faction int
		dist    [][]int
		mid     [][]int
	}{
		{"帝國", 0, want.Empire, g.Empire},
		{"同盟", 1, want.Alliance, g.Alliance},
	} {
		dist, mid := g.BuildTables(want.Owners, c.faction)
		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				if dist[i][j] != c.dist[i][j] {
					t.Fatalf("%s 距離[%d][%d] ＝ %d，收據是 %d",
						c.name, i, j, dist[i][j], c.dist[i][j])
				}
				if mid[i][j] != c.mid[i][j] {
					t.Fatalf("%s 中點[%d][%d] ＝ %d，收據是 %d",
						c.name, i, j, mid[i][j], c.mid[i][j])
				}
			}
		}
	}
}

// 值 ＝ 1024 × k ＋ 沿途邊權和，k ＝ 路上兩端不全屬於該陣營的**邊**數
// （`docs/re/242` 消去了「數格子」的十個候選，`docs/re/243` 給出這個定義）。
func TestDistanceDecomposesIntoPenaltyAndEdgeCost(t *testing.T) {
	g, rec := loadStarDistance(t)
	for faction := 0; faction < 2; faction++ {
		dist, mid := g.BuildTables(rec.Owners, faction)
		g2 := &StarGraph{Adjacency: g.Adjacency, EdgeCost: g.EdgeCost,
			Empire: mid, Alliance: mid}
		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				if i == j || dist[i][j] >= StarUnreachable {
					continue
				}
				p := g2.Route(i, j, 0)
				if p == nil {
					t.Fatalf("陣營 %d：%d→%d 展不開", faction, i, j)
				}
				k, cost := 0, 0
				for n := 0; n+1 < len(p); n++ {
					a, b := p[n], p[n+1]
					cost += g.EdgeCost[a][b]
					if rec.Owners[a] != faction || rec.Owners[b] != faction {
						k++
					}
				}
				if got := TerritoryPenalty*k + cost; got != dist[i][j] {
					t.Fatalf("陣營 %d：%d→%d ＝ %d，1024×%d＋%d ＝ %d",
						faction, i, j, dist[i][j], k, cost, got)
				}
			}
		}
	}
}

// 邊權表只有一格不對稱：`ds:533A[2][6] = 15`、`[6][2] = 13`。它是距離矩陣
// 那 186 格不對稱的唯一來源——原始資料的手誤，遊戲照著跑。
func TestEdgeCostHasOneAsymmetricPair(t *testing.T) {
	g, rec := loadStarDistance(t)
	var asym [][2]int
	for i := 0; i < 32; i++ {
		for j := i + 1; j < 32; j++ {
			if g.EdgeCost[i][j] != g.EdgeCost[j][i] {
				asym = append(asym, [2]int{i, j})
			}
		}
	}
	if len(asym) != 1 || asym[0] != [2]int{2, 6} {
		t.Fatalf("不對稱的邊 ＝ %v，預期只有 [2 6]", asym)
	}
	if g.EdgeCost[2][6] != 15 || g.EdgeCost[6][2] != 13 {
		t.Fatalf("2→6 ＝ %d、6→2 ＝ %d，預期 15／13", g.EdgeCost[2][6], g.EdgeCost[6][2])
	}

	// 把它補成對稱，兩張距離矩陣就完全對稱。
	sym := make([][]int, 32)
	for i := range sym {
		sym[i] = append([]int(nil), g.EdgeCost[i]...)
	}
	sym[2][6] = 13
	fixed := &StarGraph{Adjacency: g.Adjacency, EdgeCost: sym}
	for faction := 0; faction < 2; faction++ {
		before, _ := g.BuildTables(rec.Owners, faction)
		after, _ := fixed.BuildTables(rec.Owners, faction)
		n, m := 0, 0
		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				if before[i][j] != before[j][i] {
					n++
				}
				if after[i][j] != after[j][i] {
					m++
				}
			}
		}
		if n == 0 || m != 0 {
			t.Errorf("陣營 %d：原版不對稱 %d 格、補成對稱之後 %d 格，預期 >0 與 0",
				faction, n, m)
		}
	}
}
