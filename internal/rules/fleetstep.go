package rules

// 回合裡一支艦隊走一格的判定（原版 `113A:047D`，`docs/re/322`）。
//
// `056A:03E0` 每回合對 32 筆艦隊各叫一次 `113A:0350`：命令是移動／占領而所在
// 星域已經是目的地就停駐；停駐的走 `113A:03D5`（沒指定待命行星就挑一顆）；
// 其餘走 `113A:047D`：
//
//	next ＝ 路徑表的下一格（`2B99:0302`）
//	現在格與下一格都不是自己的         → 「無法前往敵方的星系」：+0x08 ← 5、+0x07 ← FF
//	下一格不是自己的 且 命令不是占領    → 同上
//	下一格不是自己的 且 占領 且 下一格 ≠ 目的地 → 問「<目的地>無法直接到達此星系。<下一格>要占領此星系嗎？」
//	                                      占領 → 目的地改成下一格再走；待命 → 這回合不動、命令留著
//	軍事費 < 移動經費（`113A:06C3`）     → 「由於移動經費不足無法移動」：不動、命令留著
//	否則                               → 扣經費、走到下一格（到了目的地的處理見 `Arrive`）
//
// 三張報告與「出發前往／已經到達」那張都是回合處理中的過場視窗（`3765:006C`
// 等訊息速度），畫面本身 remake 沒畫。

// 移動狀態 `+0x08` 的動作碼（`docs/re/312` §三、`docs/re/322`）。
const (
	MoveMove      = 0 // 移動
	MoveOccupy    = 1 // 占領
	MoveGarrison  = 2 // 駐留：到了目的地整支解散（`297B:025D`）
	MoveTransport = 3 // 輸送：到了目的地改成 4、掉頭回首都
	MoveReturn    = 4 // 輸送的回程：到了首都停駐
)

// StepOutcome 是 `113A:047D` 走一格的結果。
type StepOutcome int

const (
	StepNone         StepOutcome = iota // 不動（沒命令、停駐、下一格就是現在格）
	StepMoved                           // 走一格（扣經費）
	StepBlockedEnemy                    // 「無法前往敵方的星系」（`113A:0797`）
	StepAskOccupy                       // 要問「要占領此星系嗎？」（`113A:0899`）
	StepNoFunds                         // 「由於移動經費不足無法移動」（`113A:0A4E`）
)

// StepPlan 是判定結果與下一格。
type StepPlan struct {
	Outcome StepOutcome
	Next    int // 下一格；StepNone 時 −1
	Cost    int // 走這一格要扣的移動經費（艦隊 `+0x0C`）
}

// PlanStep 照 `113A:047D` 判這一支這回合怎麼走。owner 回星域擁有者
// （0／1／2），funds 是這一支陣營手上的軍事費，cost 是這一支的移動經費。
func (g *StarGraph) PlanStep(f *Fleet, owner func(int) int, funds, cost int) StepPlan {
	none := StepPlan{Outcome: StepNone, Next: -1}
	if g == nil || f == nil || !f.Moving() {
		return none
	}
	route := g.Route(f.Region, f.Target, f.Faction)
	if len(route) < 2 {
		return none
	}
	next := route[1]
	curOwn := owner(f.Region) == f.Faction
	nextOwn := owner(next) == f.Faction
	if !curOwn && !nextOwn {
		return StepPlan{Outcome: StepBlockedEnemy, Next: next}
	}
	if !nextOwn {
		if f.Move != MoveOccupy {
			return StepPlan{Outcome: StepBlockedEnemy, Next: next}
		}
		if next != f.Target {
			return StepPlan{Outcome: StepAskOccupy, Next: next}
		}
	}
	if next == f.Region {
		return none
	}
	if funds < cost {
		return StepPlan{Outcome: StepNoFunds, Next: next}
	}
	return StepPlan{Outcome: StepMoved, Next: next, Cost: cost}
}

// BlockByEnemy 是 `113A:0797` 對艦隊做的事：命令清掉、停駐。
func BlockByEnemy(f *Fleet) {
	f.Target, f.Move, f.Parked = NoTarget, MoveParked, true
}

// Arrive 是走到目的地那一刻（`113A:047D` 的 `+065B` 起）：輸送改成回程、
// 目的地換成自己的首都；移動／占領／回程停駐（`+0x08 ← 5`、`+0x07 ← FF`）。
// 駐留（2）原版在這裡整支解散（`297B:025D`），回 true 交給呼叫端處理。
func Arrive(f *Fleet) (disband bool) {
	if f.Region != f.Target {
		return false
	}
	switch f.Move {
	case MoveTransport:
		f.Move, f.Target = MoveReturn, CapitalRegion(f.Faction)
	case MoveGarrison:
		return true
	default:
		f.Target, f.Move, f.Parked = NoTarget, MoveParked, true
	}
	return false
}

// StandbyPlanet 是 `2AF5:060D(陣營, 星域)`：這個星域裡該陣營經濟力最高的
// 那一顆（切片內的序號）；一顆都沒有回 −1。停駐而沒指定待命行星的艦隊
// （`+0x07 == FF`）每回合由 `113A:03D5` 用它補上一顆（星域沒在打仗時）。
func (p *Planets) StandbyPlanet(faction, region int) int {
	start, n := p.Range(region)
	best, bestEco := -1, 0
	for k := 0; k < n; k++ {
		if p.Owner(start+k) != faction {
			continue
		}
		if e := p.PlanetEconomy(start + k); best < 0 || e > bestEco {
			best, bestEco = k, e
		}
	}
	return best
}
