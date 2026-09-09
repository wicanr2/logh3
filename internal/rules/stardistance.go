package rules

// 原版的路徑表不是靜態資料，是**擁有者的函數**：星域易主之後整組重建
// （`docs/re/243`、`docs/internals/26`）。這裡是那兩支的等價物。
//
//	2B99:007D(陣營)            用邊權表建初始矩陣
//	2B99:01B1(距離表, 中點矩陣) Floyd–Warshall，中點矩陣是它的見證矩陣
//
// 距離矩陣的值可以分解成
//
//	距離[i][j] = TerritoryPenalty × k + Σ 沿途邊權
//	k = 路徑上「兩端不全屬於該陣營」的邊數
//
// 也就是說路徑是**成本**最短、不是步數最短——會繞遠去走自己的地盤。

// 邊權表與初始矩陣用到的三個常數（`ds:533A`、`2B99:007D`）。
const (
	// StarNoEdge 是邊權表裡「不相鄰」的記法。
	StarNoEdge = 0xFF
	// StarUnreachable 是初始矩陣給不相鄰那一格填的值。原版的鬆弛用**無號**
	// 比較（`jnb`），所以兩個相加溢成 0xFFFE 仍然大於任何真值，不會誤鬆弛。
	StarUnreachable = 0x7FFF
	// TerritoryPenalty 是「這條邊的兩端不全屬於該陣營」加上去的成本。
	TerritoryPenalty = 0x400
)

// InitialCost 回 `2B99:007D(faction)` 建出來的初始矩陣。
//
// `owner` 是 32 格的星域擁有者（星域記錄的 `+0x12`：0 帝國、1 同盟、2 無主）。
func (g *StarGraph) InitialCost(owner []int, faction int) [][]int {
	d := make([][]int, 32)
	for i := range d {
		d[i] = make([]int, 32)
		for j := range d[i] {
			w := g.EdgeCost[i][j]
			if w == StarNoEdge {
				d[i][j] = StarUnreachable
				continue
			}
			if owner[i] != faction || owner[j] != faction {
				w += TerritoryPenalty
			}
			d[i][j] = w
		}
	}
	return d
}

// BuildTables 跑 `2B99:007D` ＋ `2B99:01B1`，回那個陣營的距離矩陣與中點矩陣。
//
// 中點矩陣的記法與 `Route` 讀的一樣：`255` ＝ 直接相連，否則是中繼點。
func (g *StarGraph) BuildTables(owner []int, faction int) (dist, mid [][]int) {
	dist = g.InitialCost(owner, faction)
	mid = make([][]int, 32)
	for i := range mid {
		mid[i] = make([]int, 32)
		for j := range mid[i] {
			mid[i][j] = starDirect
		}
		dist[i][i] = 0
	}
	for k := 0; k < 32; k++ {
		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				// 原版在 16 位元裡相加，所以要照樣截。
				v := (dist[i][k] + dist[k][j]) & 0xFFFF
				if v < dist[i][j] {
					dist[i][j] = v
					mid[i][j] = k
				}
			}
		}
	}
	return dist, mid
}
