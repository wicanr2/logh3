package ai

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/rules"
)

// 「從首都走得到嗎」（規格 `docs/spec/043`）。
//
// 原版 `2B99:0392(首都, 陣營, &旗標[32])` 對每一格填
//
//	旗標[i] = 2B99:028F(首都, i, 陣營) < 0x800
//
// `2B99:028F` 查的是 EMS 分頁框裡那張 32×32 word 距離表（`docs/re/185`），
// 表由 `2B99:007D` ＋ `2B99:01B1` 從邊權表與星域擁有者重建
// （`docs/re/243`，`rules.StarGraph.BuildTables`）。邊的兩端不全屬於該陣營
// 就加 `0x400`，所以 `< 0x800` 的意思是「**最多踏過一條不全是自己的邊**」
// ——從自己的地盤出發，只走得到自己的地盤與它外面那一圈。
//
// 星圖是全連通的，但這張表**不是全 1**：新遊戲同盟從巴拉特出發，帝國
// 那一側只有伊謝爾倫（13）走得到（收據 `turn260mem` 的 `t260_post.bin`，
// 那一刻分頁框裡四張表與重算逐格相同，`docs/re/298`）。

// ReachLimit 是 `2B99:0392` 拿距離比的門檻（`cmp ax, 0800h`）。
const ReachLimit = 2 * rules.TerritoryPenalty

// CapitalReach 回原版 `2B99:0392(首都, 陣營, …)` 那張 32 格布林表。
//
// 擁有者用**呼叫當下**的：原版的距離表在載入、新遊戲與回合結束
// （`2703:08DE`，重算擁有者之後）重建，而戰略 AI 跑在下一個回合的命令
// 階段裡、任何會改擁有者的步驟（叛亂、戰鬥、回合結束的重算）之前，
// 所以 AI 讀到的表就是「目前擁有者」的表（`docs/spec/043`）。
//
// 距離不對稱（`ds:533A[2][6]` 那一個位元組），但只差在邊權的零頭；
// 跨不跨 `0x800` 只看路上罰了幾條邊，而最少罰幾條是對稱的——所以
// 這張表從首都往外查或往首都查都一樣。
func CapitalReach(g *rules.StarGraph, owner func(region int) int, faction int) [RegionCount]bool {
	if g == nil || len(g.EdgeCost) != RegionCount {
		// `rules.LoadStarGraph` 保證有邊權表；沒有就是呼叫端組錯了圖。
		// 不回「全部走得到」頂替——那正是這張表要修掉的錯誤。
		panic(fmt.Sprintf("CapitalReach：星圖沒有 %d 列的邊權表", RegionCount))
	}
	owners := make([]int, RegionCount)
	for i := range owners {
		owners[i] = owner(i)
	}
	dist, _ := g.BuildTables(owners, faction)
	capital := Capitals[faction]
	var out [RegionCount]bool
	for i := range out {
		out[i] = dist[capital][i] < ReachLimit
	}
	return out
}

// ReachFrom 是同一支 `2B99:0392`，但起點換成任一星域。玩家在戰略地圖上
// 派艦隊時（`113A:00BE`，`docs/re/312`）原版就是拿艦隊**現在所在的星域**
// 呼叫它，再與擁有者一起合成允許表（`docs/spec/051`）。
func ReachFrom(g *rules.StarGraph, owner func(region int) int, faction, from int) [RegionCount]bool {
	if g == nil || len(g.EdgeCost) != RegionCount {
		panic(fmt.Sprintf("ReachFrom：星圖沒有 %d 列的邊權表", RegionCount))
	}
	if from < 0 || from >= RegionCount {
		panic(fmt.Sprintf("ReachFrom：起點 %d 不是星域", from))
	}
	owners := make([]int, RegionCount)
	for i := range owners {
		owners[i] = owner(i)
	}
	dist, _ := g.BuildTables(owners, faction)
	var out [RegionCount]bool
	for i := range out {
		out[i] = dist[from][i] < ReachLimit
	}
	return out
}

// reachFunc 把 `CapitalReach` 包成查詢函式；表只算一次。
func reachFunc(g *rules.StarGraph, owner func(region int) int, faction int) func(int) bool {
	t := CapitalReach(g, owner, faction)
	return func(r int) bool { return r >= 0 && r < RegionCount && t[r] }
}
