package ai

import "github.com/wicanr2/logh3/internal/rules"

// 逐艦隊決定那一趟（原版 `2E54:005D` → `2E54:16F5`／`2E54:1955`，
// `docs/re/79`、`docs/internals/54`）。它是每個 AI 回合的第一件事，
// 產出 `ds:D5CE`（這一格有幾支決定留守）。
//
// **回傳的 `decided` 不是原版的 `ds:D60E`。** 那個計數器數的是整個回合
// ——逐艦隊決定、守首都湊人、新編艦隊、階段三的指派都會加它
// （十處遞增點，`docs/re/197`）。

// Decision 是一支艦隊這回合的決定，對應原版艦隊記錄的 `+0x79`。
type Decision int

// 原版目前見到的四個值（`docs/internals/54`）。
const (
	DecisionIdle    Decision = 0 // 沒事做：走 2E54:1955 往首都方向退
	DecisionEngaged Decision = 3 // 有事做／就地待命
	DecisionPending Decision = 4 // 候選
	DecisionArrived Decision = 6 // 已經到了
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
}

// DecisionPass 跑一趟 `2E54:005D`：回每一支的決定、`ds:D5CE`（這一格有
// 幾支決定留守）與 `ds:D60E`（決定了幾支）。
//
// 只處理 `Faction` 那個陣營的艦隊。**回傳的清單只含這一趟真的決定了的
// 那幾支**；`+0x00` bit0 沒設又不在交戰格的不在裡面，因為原版把它們的
// `+0x79` 留在 `0xFF`，交給後面五個階段（`docs/re/196`）。
// `decided` 只是**這一趟**處理了幾支，不等於 `ds:D60E`（見上）。
func DecisionPass(in DecisionInput) (out []FleetDecision, held [RegionCount]int, decided int) {
	strength := func(f rules.Fleet) int {
		if in.Strength == nil {
			return 1
		}
		return in.Strength(f)
	}
	contested := func(r int) bool { return in.Contested != nil && in.Contested(r) }

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

	split := func(f rules.Fleet) bool { return in.Split != nil && in.Split(f) }

	for _, f := range in.Fleets {
		if f.Faction != in.Faction || f.Region < 0 || f.Region >= RegionCount {
			continue
		}
		// 那道過濾：不是「分成兩類」的那一類、又不在交戰格 → 這一趟不碰，
		// `+0x79` 留 `0xFF`（`docs/re/196`）。
		if !split(f) && !contested(f.Region) {
			continue
		}
		d := FleetDecision{Index: f.Index, Target: rules.NoTarget, Parked: true}
		if contested(f.Region) &&
			Engaged(f.Region, in.Owner(f.Region), in.Faction, own[f.Region], enemy[f.Region], f.Morale) {
			d.Decision = DecisionEngaged
			held[f.Region]++
		} else {
			d.Decision = DecisionIdle
			if t := RetreatTarget(f.Region, in.Faction, in.Graph, in.Owner); t >= 0 {
				d.Target, d.Parked = t, false
			}
		}
		decided++
		out = append(out, d)
	}
	return out, held, decided
}
