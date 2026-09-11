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
// 另外還有一支 `056A:0469` 掃同樣 32 格，管 `+0x13` 的 bit1／bit2／bit4
// （`docs/re/245`）。收據 `flagpass` 量到它跑在 `2703:08DE` **之前**，
// 兩支都在回合計數器遞增之前。
//
// **順序不能換**：路徑表是擁有者的函數。這裡做的是第一步與 `2703:0933`
// 那一步；彙總與未安定度 remake 各自現算，不需要這一圈。

// 星域記錄裡這一圈會寫到的三個位移（`docs/internals/22`）。
const (
	regionOwnerOff   = 0x12
	regionFlagsOff   = 0x13
	regionCapCostOff = 0x24
)

// 這一圈分兩步跑，因為中間會被易主報告打斷（`docs/spec/034`）：
//
//	recomputeRegionOwners   依行星重算 +0x12，易主的排進報告佇列
//	（報告全部按掉）
//	refreshRegionPosture    重建路徑表，填 +0x13 與 +0x24
//
// 擁有者照 `docs/re/17`：**全屬同一方才算那一方的，只要混雜就是無主**。
// `+0x13` 的其餘位元（bit5–bit7）原版不動，所以這裡也不動。

// regionsReady 是這一圈要不要跑的前提。
func (g *Game) regionsReady() bool {
	return g.res != nil && g.res.Planets != nil && g.res.Graph != nil &&
		len(g.res.Stellar) >= rules.RegionCount*rules.RegionRecord
}

// recomputeRegionOwners 跑第一步：逐格依行星重算 `+0x12`（`2703:159B`），
// 把易主的排進 `pendingFlips`（原版每一格易主跳一則報告，`docs/spec/034`）。
// 回這一輪有幾格易主。
func (g *Game) recomputeRegionOwners() int {
	if !g.regionsReady() {
		return 0
	}
	changed := 0
	for i := 0; i < rules.RegionCount; i++ {
		rec := g.regionRecord(i)
		old := int(rec[regionOwnerOff])
		o := g.res.Planets.RegionOwner(i)
		if old != o {
			changed++
			g.pendingFlips = append(g.pendingFlips,
				RegionFlip{Region: i, OldOwner: old, NewOwner: o})
		}
		rec[regionOwnerOff] = byte(o)
	}
	return changed
}

// refreshRegionPosture 跑第二步：重建路徑表，再填 `+0x13` 與 `+0x24`。
//
// **原版是在易主報告全部按掉之後才跑這一段**（`docs/re/246`）。
func (g *Game) refreshRegionPosture() {
	if !g.regionsReady() {
		return
	}
	owner := make([]int, rules.RegionCount)
	for i := range owner {
		owner[i] = int(g.regionRecord(i)[regionOwnerOff])
	}
	// `056A:0469` 先跑（收據 `flagpass` 量到它在 `2703:08DE` 之前，
	// 兩支都在回合計數器遞增之前，`docs/re/245`）。
	occ := g.res.Graph.Occupations(owner, g.fleets)
	for i := 0; i < rules.RegionCount; i++ {
		rec := g.regionRecord(i)
		rec[regionFlagsOff] = rules.ApplyOccupation(rec[regionFlagsOff], occ[i])
	}
	for i, p := range g.res.Graph.Postures(owner, g.fleets) {
		rec := g.regionRecord(i)
		rec[regionFlagsOff] = rules.ApplyPosture(rec[regionFlagsOff], p)
		if p.CapitalCost >= 0 {
			binary.LittleEndian.PutUint16(rec[regionCapCostOff:], uint16(p.CapitalCost))
		}
	}
}

// runRegionPass 跑整條回合結束的星域重算，並在有易主時**停在報告上**
// ——後半段（路徑表與態勢）等玩家把報告全部按掉才跑，照原版
// （`docs/re/246`、`docs/spec/034`）。`tail` 是報告收完之後要接回去的那條路。
//
// 回 true ＝ 停在報告上了，呼叫端要立刻 return。
func (g *Game) runRegionPass(tail func()) bool {
	g.recomputeRegionOwners()
	if len(g.pendingFlips) > 0 {
		g.afterFlips = tail
		g.state = StateRegionFlipReport
		return true
	}
	g.refreshRegionPosture()
	return false
}
