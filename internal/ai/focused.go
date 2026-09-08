package ai

import (
	"sort"

	"github.com/wicanr2/logh3/internal/rules"
)

// Focused 是強化版：**一次投入大量兵力，包圍、集中、殲滅**
// （`docs/spec/017`）。
//
// 原版那一層只會把艦隊往家裡收（`docs/re/79`），從不主動挑目標；
// 這裡補的就是「選目標 → 集結 → 同回合壓上去」這三件事。
//
// 「適度」調的**不是拳頭大小，是節奏**：一次攻勢能集中的都上，
// 但兩次之間要隔幾回合、家裡要留守軍、同時只打一個地方。
type Focused struct {
	// Difficulty 0–3，越大越兇。它同時調三件事：兩次攻勢的間隔、
	// 家裡留幾支、要多大的兵力比才敢動手。
	Difficulty int

	// 內部狀態。**目標要黏著**——每回合換一個就變成到處推一點。
	target     int // -1 ＝ 還沒選
	targetTurn int // 選在第幾回合
	lastAttack int // 上一次壓上去是第幾回合
	started    bool
}

// 這幾個是初值，`docs/spec/017` 的表。難度往上調就變兇。
const (
	holdTarget   = 4  // 選定的目標至少維持幾回合
	baseCooldown = 3  // 兩次攻勢之間至少隔幾回合
	baseGarrison = 1  // 首都與前線各留幾支
	baseRatio    = 15 // 兵力比門檻 ×10（15 ＝ 1.5 倍）
	maxWait      = 3  // 等包圍成形最多等幾回合
)

func (b *Focused) cooldown() int {
	n := baseCooldown - b.Difficulty
	if n < 1 {
		n = 1
	}
	return n
}

func (b *Focused) garrison() int {
	n := baseGarrison
	if b.Difficulty >= 3 {
		n = 0
	}
	return n
}

func (b *Focused) ratio() int {
	n := baseRatio - b.Difficulty*2
	if n < 10 {
		n = 10
	}
	return n
}

// Orders 回這一回合要下的命令。
func (b *Focused) Orders(w World, faction, turn int) []Order {
	if !b.started {
		b.target, b.lastAttack, b.started = -1, -100, true
	}
	mine := b.myFleets(w, faction)
	if len(mine) == 0 || w.Graph == nil {
		return nil
	}

	b.pickTarget(w, faction, turn)
	if b.target < 0 {
		return nil
	}
	staging := b.stagingRegions(w, faction, b.target)
	if len(staging) == 0 {
		// 沒有自己的星域與目標相鄰——先往目標的方向靠，靠到相鄰為止。
		return b.marchToward(w, mine, b.target)
	}

	ready, spare := b.split(w, faction, mine, staging)
	if b.canAttack(w, ready, turn) {
		// **目標要先抄下來再清**：`b.target` 是壓上去的目的地，
		// 清完再讀就變成 -1，整批艦隊會收到一道無效的命令。
		tgt := b.target
		b.lastAttack, b.target = turn, -1
		return orders(ready, tgt)
	}
	return b.gather(w, spare, staging)
}

// myFleets 回這一方**動得了**的艦隊索引。
func (b *Focused) myFleets(w World, faction int) []int {
	var out []int
	for i, f := range w.Fleets {
		if f.Faction == faction {
			out = append(out, i)
		}
	}
	return out
}

// pickTarget 選一個目標星域：相鄰、守軍弱、值錢、離自己近。
//
// **選定就黏著**：沒被打下來、也還在敵手上的話，至少維持 `holdTarget`
// 個回合再考慮換。每回合重選就是「到處推一點」，正是要避免的。
func (b *Focused) pickTarget(w World, faction, turn int) {
	if b.target >= 0 && turn-b.targetTurn < holdTarget && w.owner(b.target) != faction {
		return
	}
	if b.target >= 0 && w.owner(b.target) == faction {
		b.target = -1 // 打下來了
	}

	best, bestScore := -1, 0
	for r := 0; r < 32; r++ {
		if w.owner(r) == faction || !b.frontier(w, faction, r) {
			continue
		}
		// 分數 ＝ 價值 ÷（守軍 ＋ 1）。守軍是對方在那一格的艦隊戰力。
		score := w.value(r) * 10 / (b.defenders(w, faction, r) + 1)
		if score > bestScore {
			best, bestScore = r, score
		}
	}
	if best >= 0 && best != b.target {
		b.target, b.targetTurn = best, turn
	}
}

// frontier ＝ 這個星域與我方的某個星域相鄰。
func (b *Focused) frontier(w World, faction, region int) bool {
	for _, n := range w.Graph.Neighbours(region) {
		if w.owner(n) == faction {
			return true
		}
	}
	return false
}

// defenders 回對方在某一格的戰力總和。
func (b *Focused) defenders(w World, faction, region int) int {
	n := 0
	for _, f := range w.Fleets {
		if f.Region == region && f.Faction != faction {
			n += w.strength(f)
		}
	}
	return n
}

// stagingRegions 回可以當進攻位的我方星域：與目標相鄰。
func (b *Focused) stagingRegions(w World, faction, target int) []int {
	var out []int
	for _, n := range w.Graph.Neighbours(target) {
		if w.owner(n) == faction {
			out = append(out, n)
		}
	}
	sort.Ints(out)
	return out
}

// split 把艦隊分成「已經在進攻位上」與「還沒到」兩堆，並先扣掉守軍。
func (b *Focused) split(w World, faction int, mine, staging []int) (ready, spare []int) {
	at := map[int]bool{}
	for _, r := range staging {
		at[r] = true
	}
	home := rules.CapitalRegion(faction)
	kept := map[int]int{} // 每個要留守的星域已經留了幾支
	for _, i := range mine {
		f := w.Fleets[i]
		if at[f.Region] {
			ready = append(ready, i)
			continue
		}
		// 首都留守：不會為了進攻把家清空。
		if f.Region == home && kept[home] < b.garrison() {
			kept[home]++
			continue
		}
		spare = append(spare, i)
	}
	return ready, spare
}

// canAttack ＝ 包圍成形（至少兩個方向）而且兵力比達標。
//
// 等太久也不行：`maxWait` 到了就照現有兵力打，不然會永遠在等。
func (b *Focused) canAttack(w World, ready []int, turn int) bool {
	if len(ready) == 0 || turn-b.lastAttack < b.cooldown() {
		return false
	}
	sides := map[int]bool{}
	force := 0
	for _, i := range ready {
		sides[w.Fleets[i].Region] = true
		force += w.strength(w.Fleets[i])
	}
	def := b.defenders(w, w.Fleets[ready[0]].Faction, b.target)
	strong := force*10 >= def*b.ratio()
	if !strong {
		return false
	}
	if len(sides) >= 2 {
		return true
	}
	// 只有一個方向：等，但不要等到天荒地老。
	return turn-b.targetTurn >= maxWait
}

// gather 把還沒到位的艦隊派往**最近的**進攻位。
func (b *Focused) gather(w World, spare, staging []int) []Order {
	var out []Order
	for _, i := range spare {
		f := w.Fleets[i]
		best, bestHops := -1, 1<<30
		for _, r := range staging {
			h := w.Graph.Hops(f.Region, r)
			if h >= 0 && h < bestHops {
				best, bestHops = r, h
			}
		}
		if best >= 0 && best != f.Region {
			out = append(out, Order{Fleet: i, Target: best})
		}
	}
	return out
}

// marchToward 沒有進攻位時，先整批往目標靠。
func (b *Focused) marchToward(w World, mine []int, target int) []Order {
	var out []Order
	for _, i := range mine {
		if w.Fleets[i].Region != target {
			out = append(out, Order{Fleet: i, Target: target})
		}
	}
	return out
}

func orders(fleets []int, target int) []Order {
	var out []Order
	for _, i := range fleets {
		out = append(out, Order{Fleet: i, Target: target})
	}
	return out
}
