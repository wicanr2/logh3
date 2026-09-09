package ai

import "github.com/wicanr2/logh3/internal/rules"

// 把各層串成一個戰略回合（規格 `docs/spec/029`、`docs/internals/54`）。
//
// **這裡不是預設的 `Brain`。** `Original` 仍然回 `nil`——把它換掉會改變
// remake 在「實行」時的行為，那一項要使用者裁決（理由與影響寫在
// `docs/spec/029` 末尾）。這一支是把裁決變成一行的準備。

// StrategyInput 是跑一個回合要的東西。沒給的那幾個回退成保守值，
// 每一個都在欄位註解裡寫明退到什麼。
type StrategyInput struct {
	Faction int
	Fleets  []rules.Fleet
	Graph   *rules.StarGraph

	// Owner 回星域的擁有者（0 帝國、1 同盟、2 無主）。
	Owner func(region int) int

	// Contested 是星域 `+0x13` 的 bit2：這一格有戰鬥（`docs/re/173`）。
	// nil ＝ 都沒有。
	Contested func(region int) bool

	// Blocked 是星域 `+0x13` 的 `& 0x14`（bit2｜bit4）：bit0 有設的那一路
	// 拿它當「不能留在這裡」的第一個條件（`docs/re/197`）。nil ＝ 都沒設。
	Blocked func(region int) bool

	// EnemyFleets 回敵方在那一格有幾支艦隊（星域 `+0x28`／`+0x29`，
	// `docs/re/199`）。nil ＝ 一律 0。
	EnemyFleets func(region int) int

	// Split 是艦隊 `+0x00` 的 bit0（`docs/re/196`）。**bit0 沒設又不在
	// 交戰格的艦隊，逐艦隊決定不碰它們**，才會留給階段二／三／五挑。
	// nil ＝ 都沒設。
	Split func(f rules.Fleet) bool

	// Strength 回一支艦隊的戰力（`rules.Slot.FleetStrength`）。
	// nil ＝ 每支算 1。
	Strength func(f rules.Fleet) int

	// Ships 回一支艦隊的「艘數總和」與「主力槽數」
	// （`rules.Slot.FleetShipTotals`）。nil ＝ 兩個都回 0。
	Ships func(f rules.Fleet) (ships, mainSlots int)

	// Routes 是進軍路線表；nil 就跳過階段四、五。
	Routes *AdvanceRoutes

	// Rand 是原版的亂數（`rules.NewRand`）。nil ＝ 亂數項一律取 0。
	Rand *rules.Rand

	// Reachable 對應 `2B99:0392`；nil ＝ 全部走得到，那**是**原版行為
	// （星圖全連通，`docs/re/185`）。
	Reachable func(region int) bool

	// PhezzanOpen 是原版的 `ds:8F48`（費沙進入戰局，`docs/re/179`）。
	PhezzanOpen bool
}

func (in StrategyInput) owner(r int) int {
	if in.Owner == nil {
		return RegionOwnerNeutral
	}
	return in.Owner(r)
}

func (in StrategyInput) reachable(r int) bool {
	return in.Reachable == nil || in.Reachable(r)
}

func (in StrategyInput) candidates(decided map[int]bool) []Candidate {
	out := make([]Candidate, 0, len(in.Fleets))
	for i, f := range in.Fleets {
		if f.Faction != in.Faction {
			continue
		}
		c := Candidate{
			FleetView: FleetView{Faction: f.Faction, Region: f.Region, Target: f.Target,
				Action: fleetAction(f), Groups: f.Groups},
			Decided: decided[i], Morale: f.Morale, Expected: f.Expected,
			Class: rules.FleetClass(f.ClassKey),
		}
		if in.Ships != nil {
			c.Ships, c.MainSlots = in.Ships(f)
		}
		c.Index = i
		out = append(out, c)
	}
	return out
}

// fleetAction 把 remake 的 `Parked` 換回原版 `+0x08` 的值。
func fleetAction(f rules.Fleet) int {
	if f.Parked {
		return 5
	}
	return 0
}

// Turn 跑一個戰略回合，回這一回合要下的命令。
//
// 目前涵蓋**階段二、三、五**（`docs/spec/029` 的階段表）：
//
//	逐艦隊決定（2E54:005D）
//	階段二：打最近的無主星域（模式 2 湊滿兩支才打）
//	階段三：按缺口補（模式 0）
//	階段五：收尾指派
//
// **沒有涵蓋**階段一（建艦隊——它產生的是新艦隊，不是命令）與階段四
// （`2E54:0AAB` 的三條分支還沒實作）。兩者都在 `docs/spec/029` 標著。
func Turn(in StrategyInput) []Order {
	if in.Graph == nil {
		return nil
	}
	var orders []Order
	decided := map[int]bool{}

	// 逐艦隊決定：填 ds:D5CE，並把「有事做」的那幾支標成已決定。
	dp, held, _ := DecisionPass(DecisionInput{
		Faction: in.Faction, Fleets: in.Fleets, Graph: in.Graph,
		Owner: in.owner, Strength: in.Strength,
		Contested: in.Contested, Split: in.Split,
		Blocked: in.Blocked, EnemyFleets: in.EnemyFleets,
	})
	byIndex := map[int]int{}
	for i, f := range in.Fleets {
		byIndex[f.Index] = i
	}
	// `DecisionPass` 回傳的每一支都算「已經決定」：原版寫進 `+0x79` 的
	// 是 0 或 3，都不再是 `0xFF`，所以 `2E54:143F` 之後挑不到它們
	// （`docs/internals/54`）。**它不會回傳每一支**——`+0x00` bit0 沒設
	// 又不在交戰格的那些留著 `0xFF`，正是後面五個階段的候選來源
	// （`docs/re/196`）。
	for _, d := range dp {
		i, ok := byIndex[d.Index]
		if !ok {
			continue
		}
		decided[i] = true
		if d.Decision == DecisionIdle && d.Target >= 0 {
			orders = append(orders, Order{Fleet: i, Target: d.Target})
		}
	}

	orders = append(orders, in.phaseNeutral(decided)...)
	orders = append(orders, in.phaseDefence(decided, held)...)
	orders = append(orders, in.phaseFinal(decided, held)...)
	return orders
}

// phaseNeutral ＝ 階段二：打最近的、走得到的無主星域，用模式 2 湊滿兩支
// 才打（`docs/re/177`）。
func (in StrategyInput) phaseNeutral(decided map[int]bool) []Order {
	target, best := -1, 1<<30
	capital := Capitals[in.Faction]
	for r := 0; r < RegionCount; r++ {
		if in.owner(r) != RegionOwnerNeutral || !in.reachable(r) {
			continue
		}
		d := in.Graph.Hops(capital, r)
		if d >= 0 && d < best {
			target, best = r, d
		}
	}
	if target < 0 {
		return nil
	}
	m := PickerModes[2]
	var picked []int
	for len(picked) < 2 {
		cs := in.candidates(decided)
		k := m.Pick(cs, target, in.Graph)
		if k < 0 {
			break
		}
		i := cs[k].Index
		decided[i] = true
		picked = append(picked, i)
	}
	if len(picked) < 2 { // 湊不滿就不打；把剛才佔住的放回去
		for _, i := range picked {
			delete(decided, i)
		}
		return nil
	}
	out := make([]Order, 0, 2)
	for _, i := range picked {
		out = append(out, Order{Fleet: i, Target: target})
	}
	return out
}

// phaseDefence ＝ 階段三：算「要派幾支」，反覆挑滿足率最低的格、
// 用模式 0 取艦隊指派（`docs/re/172`）。
func (in StrategyInput) phaseDefence(decided map[int]bool, held [RegionCount]int) []Order {
	views := make([]FleetView, 0, len(in.Fleets))
	for _, f := range in.Fleets {
		views = append(views, FleetView{Faction: f.Faction, Region: f.Region,
			Target: f.Target, Action: fleetAction(f), Groups: f.Groups})
	}
	b := DefenceBoard(TurnInput{
		Faction: in.Faction, Fleets: views, Graph: in.Graph,
		Owner: in.owner, Reachable: in.Reachable,
	}, held)

	var out []Order
	m := PickerModes[0]
	for {
		cell := PickCell(b.Need, b.Assigned, Capitals)
		if cell < 0 {
			return out
		}
		cs := in.candidates(decided)
		k := m.Pick(cs, cell, in.Graph)
		if k < 0 {
			return out
		}
		i := cs[k].Index
		decided[i] = true
		b.Assigned[cell]++
		if in.Fleets[i].Region != cell {
			out = append(out, Order{Fleet: i, Target: cell})
		}
	}
}

// phaseFinal ＝ 階段五：依攻擊目標分挑一格，`2E54:13A8` 找前進基地，
// 把**所有還沒決定**的艦隊掃一遍（`docs/re/177`、`docs/re/179`）。
func (in StrategyInput) phaseFinal(decided map[int]bool, held [RegionCount]int) []Order {
	if in.Routes == nil {
		return nil
	}
	state := StateCode(in.owner(Iserlohn), in.owner(Phezzan), in.PhezzanOpen)
	views := make([]FleetView, 0, len(in.Fleets))
	for _, f := range in.Fleets {
		views = append(views, FleetView{Faction: f.Faction, Region: f.Region,
			Target: f.Target, Action: fleetAction(f), Groups: f.Groups})
	}
	b := AttackBoard(TurnInput{
		Faction: in.Faction, Fleets: views, Graph: in.Graph,
		Owner: in.owner, Reachable: in.Reachable, Rand: in.Rand,
	}, in.Routes, state, held)

	cell, best := -1, 0
	for r := 0; r < RegionCount; r++ {
		if b.Priority[r] > best {
			cell, best = r, b.Priority[r]
		}
	}
	if cell < 0 {
		return nil
	}
	base := ForwardBase(cell, in.Faction, in.Graph, in.owner)
	capital := Capitals[in.Faction]

	var out []Order
	for i, f := range in.Fleets {
		if f.Faction != in.Faction || decided[i] {
			continue
		}
		switch {
		case f.Expected != 0 && f.Groups == f.Expected:
			out = append(out, Order{Fleet: i, Target: capital})
		case base >= 0 && in.Graph.Hops(f.Region, base) >= 0:
			out = append(out, Order{Fleet: i, Target: base})
		case in.Graph.Hops(f.Region, capital) >= 0:
			out = append(out, Order{Fleet: i, Target: capital})
		}
		decided[i] = true
	}
	return out
}

// ForwardBase ＝ 原版 `2E54:13A8(星域)`：在那一格的鄰居裡挑「擁有者 ＝
// 自己、到自己首都步數最少」的一格（`docs/re/179`）。找不到回 -1。
func ForwardBase(region, faction int, g *rules.StarGraph, owner func(int) int) int {
	capital := Capitals[faction]
	best, bestD := -1, 1<<30
	for _, n := range g.Neighbours(region) {
		if n < 0 || n >= RegionCount || owner(n) != faction {
			continue
		}
		if d := g.Hops(n, capital); d >= 0 && d < bestD {
			best, bestD = n, d
		}
	}
	return best
}

// OriginalStrategy 是把 `Turn` 包成 `Brain` 的那一層。
//
// **它不是預設。** `Original` 仍然回 `nil`；要換掉是 `docs/spec/029`
// 末節那個裁決，換的時候只要把建構 `Brain` 的地方改成這一支。
//
// `World` 帶不到的幾樣（交戰旗標、艦隊分類、艦艇數、路線表、亂數）
// 由這個結構自己帶——它們都是原版有、而 `World` 這個介面當初沒開的口。
type OriginalStrategy struct {
	// Contested 是星域 `+0x13` bit2（這一格有戰鬥，`docs/re/173`）。
	Contested func(region int) bool
	// Blocked 是星域 `+0x13` 的 `& 0x14`（`docs/re/197`）。
	Blocked func(region int) bool
	// EnemyFleets 是敵方在那一格的艦隊數（星域 `+0x28`／`+0x29`，
	// `docs/re/199`）。
	EnemyFleets func(region int) int
	// Split 是艦隊 `+0x00` bit0（`docs/re/196`）。
	Split func(f rules.Fleet) bool
	// Ships 回一支艦隊的艘數與主力槽數（`rules.Slot.FleetShipTotals`）。
	Ships func(f rules.Fleet) (ships, mainSlots int)
	// Routes 是進軍路線表（`LoadAdvanceRoutes`）。
	Routes *AdvanceRoutes
	// Rand 是原版的亂數（`rules.NewRand`），跨回合要沿用同一份。
	Rand *rules.Rand
	// PhezzanOpen 是 `ds:8F48`（費沙進入戰局，`docs/re/179`）。
	PhezzanOpen bool
	// Support 是階段一（建艦隊）要的、`World` 帶不到的東西。
	// nil ＝ 不建艦隊（`original_buildpass.go`）。
	Support *BuildSupport
}

// Orders 跑一個戰略回合（`docs/spec/029` 的階段二、三、五）。
func (o OriginalStrategy) Orders(w World, faction, _ int) []Order {
	return Turn(StrategyInput{
		Faction: faction, Fleets: w.Fleets, Graph: w.Graph,
		Owner: w.Owner, Contested: o.Contested, Split: o.Split,
		Blocked: o.Blocked, EnemyFleets: o.EnemyFleets,
		Strength: w.Strength, Ships: o.Ships,
		Routes: o.Routes, Rand: o.Rand, PhezzanOpen: o.PhezzanOpen,
	})
}
