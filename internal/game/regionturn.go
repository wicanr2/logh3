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
// **順序不能換**：路徑表是擁有者的函數。這裡做的是第一步、`2703:0933`
// 與 `2703:0DA1`（未安定度，`recomputeUnrest`，`docs/spec/042`）；彙總
// remake 各自現算，不需要這一圈。

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

// refillRegionSupply 是原版 `0654:0007` 在月底閘門之前每回合都跑的
// `2B5F:02B2` ＋ `2703:0590`。先由行星彙總、回充擁有者，再照整數百分比
// 分回行星；因此戰鬥／補給命令消耗的行星存量會進入下一回合的基數。
func (g *Game) refillRegionSupply() {
	if !g.regionsReady() {
		return
	}
	for region := 0; region < rules.RegionCount; region++ {
		rec := g.regionRecord(region)
		if !rules.AggregateRegionSupply(rec, g.res.Planets, region) ||
			!rules.RefillRegionSupply(rec) {
			continue
		}
		for faction := 0; faction < 2; faction++ {
			rules.DistributeRegionSupply(rec, g.res.Planets, region, faction)
		}
	}
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
			if i == rules.PhezzanRegion && !g.phezzanBattleResolved &&
				!rules.PhezzanOccurred(g.workRecord()) {
				g.phezzanPending = true
			}
		}
		rec[regionOwnerOff] = byte(o)
	}
	g.phezzanBattleResolved = false
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
	// `056A:0469`（bit1／bit2／bit4）**不在這裡**：它在命令階段那一串走完艦隊之後、
	// 逐格掃描之前跑（`prepareOccupationReports`；`docs/re/245` 的斷點其實是 `056A:03E0`
	// 的返回）。收據 `fcmd-occ-phz-t3`：戰略打行星把艦隊士氣打到 bit0 之後，回合
	// 結束時費沙的 `+0x13` 仍是 18——這一圈若重算 bit4 就會清掉（`docs/re/330`）。
	for i, p := range g.res.Graph.Postures(owner, g.fleets) {
		rec := g.regionRecord(i)
		rec[regionFlagsOff] = rules.ApplyPosture(rec[regionFlagsOff], p)
		if p.CapitalCost >= 0 {
			binary.LittleEndian.PutUint16(rec[regionCapCostOff:], uint16(p.CapitalCost))
		}
	}
	// `2703:08DE` 逐格的第三支 `2703:0DA1`。回合計數器還沒進位
	// （remake 的 `g.turn` 在 advanceTurn 開頭就先加了）。
	//
	// 原版是逐格「態勢 → 彙總 → 未安定度」交錯跑，所以第 i 格讀到的是
	// 編號比它大的鄰居**還沒更新**的 `+0x13`；這裡全部更新完才算。兩者只在
	// 這一圈改了鄰居的 bit2 時才會不同，而下一個座位回合開頭的那一次
	// （`finishTurn`）會照更新完的旗標再算一次，玩家看得到的是那一次。
	g.recomputeUnrest(g.turn - 1)
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
