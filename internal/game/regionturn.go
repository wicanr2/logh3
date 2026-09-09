package game

import (
	"encoding/binary"

	"github.com/wicanr2/logh3/internal/rules"
)

// 回合結束的星域重算（原版 `2703:08DE`，`docs/re/244`、`docs/spec/032`）。
//
//	逐格重算擁有者（2703:159B → 2703:15D5）；易主的另外處理
//	2B99:0003        重建四張路徑表
//	for 星域 i:  2703:0933(態勢旗標 + 到首都的成本)
//	             2703:12EB(艦艇彙總)、2703:0DA1(未安定度)
//	判勝負
//
// **順序不能換**：路徑表是擁有者的函數。這裡做的是第一步與 `2703:0933`
// 那一步；彙總與未安定度 remake 各自現算，不需要這一圈。

// 星域記錄裡這一圈會寫到的三個位移（`docs/internals/22`）。
const (
	regionOwnerOff   = 0x12
	regionFlagsOff   = 0x13
	regionCapCostOff = 0x24
)

// recomputeRegions 跑一輪回合結束的星域重算，回這一輪有幾格易主。
//
// 擁有者照 `docs/re/17`：**全屬同一方才算那一方的，只要混雜就是無主**。
// 之後重建路徑表，把 `+0x13` 的 bit0／bit2／bit3 與 `+0x24` 填回去
// （其餘位元原版不動，所以這裡也不動）。
func (g *Game) recomputeRegions() int {
	if g.res == nil || g.res.Planets == nil || g.res.Graph == nil ||
		len(g.res.Stellar) < rules.RegionCount*rules.RegionRecord {
		return 0
	}
	owner := make([]int, rules.RegionCount)
	changed := 0
	for i := 0; i < rules.RegionCount; i++ {
		rec := g.regionRecord(i)
		o := g.res.Planets.RegionOwner(i)
		if int(rec[regionOwnerOff]) != o {
			changed++
		}
		rec[regionOwnerOff] = byte(o)
		owner[i] = o
	}
	for i, p := range g.res.Graph.Postures(owner, g.fleets) {
		rec := g.regionRecord(i)
		rec[regionFlagsOff] = rules.ApplyPosture(rec[regionFlagsOff], p)
		if p.CapitalCost >= 0 {
			binary.LittleEndian.PutUint16(rec[regionCapCostOff:], uint16(p.CapitalCost))
		}
	}
	return changed
}
