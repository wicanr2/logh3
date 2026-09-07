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
	Adjacency [][]int  `json:"adjacency"`
	Empire    [][]int  `json:"midpoint_empire"`
	Alliance  [][]int  `json:"midpoint_alliance"`
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

// Steps 回原版算出來的距離（走幾步）。未安定度的算式用的就是這個
// （`docs/re/28`）。走不出來回 -1。
func (g *StarGraph) Steps(from, to, faction int) int {
	p := g.Route(from, to, faction)
	if p == nil {
		return -1
	}
	return len(p) - 1
}

// CapitalRegion 回一個陣營的首都星域：帝國 0（瓦爾哈拉）、同盟 31（巴拉特）。
// 原版是 `3982:0329`，一支只有兩個出口的函式（`docs/re/28`）。
func CapitalRegion(faction int) int {
	if faction == 0 {
		return 0
	}
	return 31
}
