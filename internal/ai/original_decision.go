package ai

import "github.com/wicanr2/logh3/internal/rules"

// 逐艦隊決定那一趟（原版 `2E54:005D` → `2E54:16F5`／`2E54:1955`，
// `docs/re/79`、`docs/internals/54`）。它是每個 AI 回合的第一件事，
// 產出 `ds:D5CE`（這一格有幾支決定留守）。
//
// **回傳的 `decided` 不是原版的 `ds:D60E`。** 那個計數器數的是整個回合
// ——逐艦隊決定、守首都湊人、新編艦隊、階段二一次派兩支、階段三的指派
// 都會加它（十一處遞增點，`docs/re/197`、`docs/re/198`）。

// Decision 是一支艦隊這回合的決定，對應原版艦隊記錄的 `+0x79`。
type Decision int

// 原版目前見到的四個值（`docs/internals/54`）。
const (
	DecisionIdle    Decision = 0 // 沒事做：走 2E54:1955 往首都方向退
	DecisionEngaged Decision = 3 // 有事做／就地待命
	DecisionPending Decision = 4 // 候選
	DecisionArrived Decision = 6 // 已經到了
)

// 移動狀態 `+0x08` 這一趟用得到的兩個值（`docs/internals/33`）。
const (
	moveSkip   = 4 // 這一趟跳過它
	moveSortie = 3 // 原地出擊：`+0x07` 填自己所在的星域
)

// RegionOwnerNeutral ＝ 星域記錄 `+0x12` 的「無主」（帝國 0、同盟 1）。
const RegionOwnerNeutral = 2

// Engaged 回原版 `2E54:16F5(艦隊)` 的判斷：這一支在它現在這一格
// **有沒有事做**。
//
//	在自己陣營的首都星域           → 無條件是
//	這一格是自己的                 → 我方 × 4 >= 敵方，或 士氣 == 100
//	這一格無主（擁有者 2）         → 我方 × 2 > 敵方
//	這一格是敵方的                 → 我方 > 敵方
//
// `own`／`enemy` 是那一格雙方的戰力總和（原版 `2E54:1876`，
// remake 用 `rules.Slot.FleetStrength` 加總）。
func Engaged(region, regionOwner, faction, own, enemy, morale int) bool {
	if region == Capitals[faction] {
		return true
	}
	switch regionOwner {
	case faction:
		return own*4 >= enemy || morale == 100
	case RegionOwnerNeutral:
		return own*2 > enemy
	default:
		return own > enemy
	}
}

// RetreatTarget 回原版 `2E54:1955(艦隊)` 挑的那一格：**往自己首都的方向
// 退一格**——在相鄰星域裡挑「擁有者 ＝ 自己」而且到首都步數最少的那一格。
//
// 找不到回 -1（原版寫 `+0x08 = 5`、`+0x07 = 0xFF`，也就是不動）。
func RetreatTarget(region, faction int, g *rules.StarGraph, owner func(int) int) int {
	capital := Capitals[faction]
	best, bestD := -1, 1<<30
	for _, n := range g.Neighbours(region) {
		if n < 0 || n >= RegionCount || owner(n) != faction {
			continue
		}
		d := g.Hops(n, capital)
		if d < 0 {
			continue
		}
		if d < bestD {
			best, bestD = n, d
		}
	}
	return best
}

// FleetDecision 是一支艦隊跑完 `2E54:005D` 之後的結果。
type FleetDecision struct {
	Index    int      // 艦隊表的槽號
	Decision Decision // 寫進 `+0x79` 的值
	Target   int      // 寫進 `+0x07`；-1 ＝ 0xFF（沒有命令）
	Parked   bool     // `+0x08` 是 5
	// Move 是這一趟寫進 `+0x08` 的值；-1 ＝ 沒有寫（`+0x08 == 4` 那條
	// 只改 `+0x79`）。`Parked` ＝ `Move == 5`。
	Move int
}

// DecisionInput 是跑一趟逐艦隊決定要的東西。
type DecisionInput struct {
	Faction  int
	Fleets   []rules.Fleet
	Graph    *rules.StarGraph
	Owner    func(region int) int
	Strength func(f rules.Fleet) int // 一支艦隊的戰力（rules.FleetStrength）
	// Contested 是星域記錄 `+0x13` 的 bit2：這一格有戰鬥
	// （`docs/re/173`、`docs/internals/22`）。
	Contested func(region int) bool

	// Split 是艦隊記錄 `+0x00` 的 bit0：把艦隊分成兩類
	// （`docs/internals/33`）。**bit0 沒設而且不在交戰格的那些艦隊，
	// 這一趟完全不碰**——`+0x79` 留著 `0xFF`，於是後面五個階段挑得到
	// 它們（`docs/re/196`）。nil ＝ 都沒設。
	Split func(f rules.Fleet) bool

	// Blocked 是星域記錄 `+0x13` 的 `& 0x14`（bit2｜bit4）：bit0 有設的
	// 那一路拿它當「不能留在這裡」的第一個條件（`docs/re/197`）。
	// nil ＝ 都沒設。
	Blocked func(region int) bool

	// EnemyFleets 回**敵方**在那一格有幾支艦隊，＝ 星域記錄 `+0x28`
	// （帝國）／`+0x29`（同盟）（`rules.RegionFleets`、`docs/re/199`）。
	// nil ＝ 一律 0。
	EnemyFleets func(region int) int
}

// DecisionPass 跑一趟 `2E54:005D`：回每一支的決定與 `ds:D5CE`（這一格有
// 幾支決定留守）。
//
// 三條分支照原版的順序（`docs/internals/54`、`docs/re/79`、`docs/re/196`、
// `docs/re/197`）：
//
//	+0x08 == 4          → +0x79 = 0，其餘什麼都不寫
//	+0x00 bit0 有設     → 當場處理完：原地出擊，或往首都退一格
//	+0x00 bit0 沒設     → 只有所在星域「有戰鬥」才處理，否則整支不碰
//
// 只處理 `Faction` 那個陣營的艦隊。**回傳的清單只含這一趟真的碰過的
// 那幾支**；第三條裡沒有戰鬥的那些不在裡面，因為原版把它們的 `+0x79`
// 留在 `0xFF`，交給後面五個階段（`docs/re/196`）。
//
// `decided` 是**這一趟**碰了幾支，不等於 `ds:D60E`（見檔首）。
func DecisionPass(in DecisionInput) (out []FleetDecision, held [RegionCount]int, decided int) {
	strength := func(f rules.Fleet) int {
		if in.Strength == nil {
			return 1
		}
		return in.Strength(f)
	}
	contested := func(r int) bool { return in.Contested != nil && in.Contested(r) }
	blocked := func(r int) bool { return in.Blocked != nil && in.Blocked(r) }
	enemyFleets := func(r int) int {
		if in.EnemyFleets == nil {
			return 0
		}
		return in.EnemyFleets(r)
	}
	split := func(f rules.Fleet) bool { return in.Split != nil && in.Split(f) }

	// 每一格雙方的戰力總和。
	var own, enemy [RegionCount]int
	for _, f := range in.Fleets {
		if f.Region < 0 || f.Region >= RegionCount {
			continue
		}
		if f.Faction == in.Faction {
			own[f.Region] += strength(f)
		} else {
			enemy[f.Region] += strength(f)
		}
	}

	// retreat 把一支填成「往首都方向退一格」；退不了就原地不動
	// （原版寫 `+0x08 = 5`、`+0x07 = 0xFF`）。
	retreat := func(d *FleetDecision, region int) {
		if t := RetreatTarget(region, in.Faction, in.Graph, in.Owner); t >= 0 {
			d.Target, d.Move, d.Parked = t, 0, false
			return
		}
		d.Move, d.Parked = 5, true
	}

	for _, f := range in.Fleets {
		if f.Faction != in.Faction || f.Region < 0 || f.Region >= RegionCount {
			continue
		}
		d := FleetDecision{Index: f.Index, Target: rules.NoTarget,
			Decision: DecisionIdle, Move: -1, Parked: false}

		switch {
		case f.Move == moveSkip:
			// 只寫 `+0x79 = 0`，移動狀態與目的地都不動。

		case split(f):
			// 這一格待不住就退：有戰鬥（`+0x13 & 0x14`）、有敵方艦隊
			// （`+0x28`／`+0x29`）、或這一格不是自己的。
			if blocked(f.Region) || enemyFleets(f.Region) != 0 || in.Owner(f.Region) != in.Faction {
				retreat(&d, f.Region)
				break
			}
			d.Move, d.Target = moveSortie, f.Region // 原地出擊

		default:
			// bit0 沒設：沒有戰鬥就整支不碰，`+0x79` 留 `0xFF`，
			// 後面五個階段挑得到它（`docs/re/196`）。
			if !contested(f.Region) {
				continue
			}
			if Engaged(f.Region, in.Owner(f.Region), in.Faction,
				own[f.Region], enemy[f.Region], f.Morale) {
				d.Decision = DecisionEngaged
				held[f.Region]++
				break
			}
			retreat(&d, f.Region)
		}

		decided++
		out = append(out, d)
	}
	return out, held, decided
}
