package rules

import (
	"encoding/json"
	"fmt"
	"os"
)

// StarGraph 是 32 個星域的相鄰關係與原版的路徑表（`docs/internals/26`）。
//
// 原版把路徑存成兩張 32×32 的**中點矩陣**放在 EMS 裡（每個陣營一張）：
// `T[i][j] == 255` 代表兩者直接相連，否則 `T[i][j]` 是中繼點，
// `path(i,j) = path(i,T) + path(T,j)`。
//
// **兩張表的相鄰關係相同，但走出來的路不一定是最短路**：同一張圖上廣度優先
// 最遠 9 步，照原版的表展開帝國那張最遠 10 步、同盟那張 12 步。
// 要跟原版一致就得照表走。
type StarGraph struct {
	Adjacency [][]int `json:"adjacency"`
	Empire    [][]int `json:"midpoint_empire"`
	Alliance  [][]int `json:"midpoint_alliance"`
}

const starDirect = 255 // 中點矩陣裡「直接相連」的記法

// LoadStarGraph 讀 `docs/internals/data/star-graph.json`。
func LoadStarGraph(path string) (*StarGraph, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g StarGraph
	if err := json.Unmarshal(b, &g); err != nil {
		return nil, err
	}
	for name, m := range map[string][][]int{
		"adjacency": g.Adjacency, "midpoint_empire": g.Empire, "midpoint_alliance": g.Alliance,
	} {
		if len(m) != 32 {
			return nil, fmt.Errorf("%s 有 %d 筆，預期 32", name, len(m))
		}
	}
	return &g, nil
}

// Neighbours 回與這個星域直接相連的星域。
func (g *StarGraph) Neighbours(i int) []int {
	if i < 0 || i >= len(g.Adjacency) {
		return nil
	}
	return g.Adjacency[i]
}

func (g *StarGraph) midpoint(faction int) [][]int {
	if faction == 0 {
		return g.Empire
	}
	return g.Alliance
}

// Route 回原版那個陣營從 from 走到 to 的路（含頭尾）；走不出來回 nil。
func (g *StarGraph) Route(from, to, faction int) []int {
	t := g.midpoint(faction)
	if from < 0 || from >= 32 || to < 0 || to >= 32 {
		return nil
	}
	var walk func(i, j, depth int) []int
	walk = func(i, j, depth int) []int {
		if i == j {
			return []int{i}
		}
		if depth > 40 {
			return nil
		}
		v := t[i][j]
		if v == starDirect {
			return []int{i, j}
		}
		a, b := walk(i, v, depth+1), walk(v, j, depth+1)
		if a == nil || b == nil {
			return nil
		}
		return append(a[:len(a)-1:len(a)-1], b...)
	}
	return walk(from, to, 0)
}

// NextStep 回從 from 往 to 走的**下一格**；走不出來回 -1。
//
// 這是原版 `2B99:0302(to, from, 陣營)` 的等價物（`docs/re/183`）。原版
// 走的是定點迭代——`a ← T[from][a]`，起點 a ＝ to，直到 `T[from][a]`
// 是 255（直接相連）就回 a：
//
//	2B99:0349(a, b, m): v = m[b][a]; v == 0FFh ? a : 2B99:0349(v, b, m)
//
// 與 `Route(from, to)[1]` 是同一個值：`Route` 的 `walk(i, v) ++ walk(v, j)`
// 展開之後，第一步同樣是 `T[i][·]` 這條鏈的定點。
func (g *StarGraph) NextStep(from, to, faction int) int {
	t := g.midpoint(faction)
	if from < 0 || from >= 32 || to < 0 || to >= 32 {
		return -1
	}
	a := to
	for i := 0; i <= 64; i++ {
		v := t[from][a]
		if v == starDirect {
			return a
		}
		if v < 0 || v >= 32 {
			return -1
		}
		a = v
	}
	return -1
}

// Steps 回把中點矩陣遞迴展開之後的路徑長度（走幾步）。走不出來回 -1。
//
// **未安定度用的不是這個**：量到的 31 個步數（`docs/re/43`）與 Hops
// 相同，而 Steps 在星域 15、22 兩處多一步——遞迴展開挑的中繼點不見得
// 落在最短路上。要「幾步」用 Hops，要「經過哪些星域」才用 Route。
func (g *StarGraph) Steps(from, to, faction int) int {
	p := g.Route(from, to, faction)
	if p == nil {
		return -1
	}
	return len(p) - 1
}

// Hops 回相鄰圖上的最短步數（BFS）。未安定度算式尾巴扣的就是這個
// ——對 `2703:0ED6` 設 -regs-at 量到的 31 個回值與它逐一相同
// （`docs/re/43`）。走不出來回 -1。
func (g *StarGraph) Hops(from, to int) int {
	if from == to {
		return 0
	}
	seen := map[int]bool{from: true}
	frontier := []int{from}
	for d := 1; len(frontier) > 0; d++ {
		var next []int
		for _, u := range frontier {
			for _, v := range g.Neighbours(u) {
				if seen[v] {
					continue
				}
				if v == to {
					return d
				}
				seen[v] = true
				next = append(next, v)
			}
		}
		frontier = next
	}
	return -1
}

// CapitalRegion 回一個陣營的首都星域：帝國 0（瓦爾哈拉）、同盟 31（巴拉特）。
// 原版是 `3982:0329`，一支只有兩個出口的函式（`docs/re/28`）。
func CapitalRegion(faction int) int {
	if faction == 0 {
		return 0
	}
	return 31
}
