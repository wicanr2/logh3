package ai

import "github.com/wicanr2/logh3/internal/rules"

// Board 是原版戰略 AI 一個回合裡那六張「以星域為索引」的表
// （`docs/internals/54`、規格 `docs/spec/029`）。
//
// 名字沿用原版的位址，免得兩邊對照時還要換一次名。
type Board struct {
	Priority [RegionCount]int // ds:D50E
	Threat   [RegionCount]int // ds:D54E
	Need     [RegionCount]int // ds:D58E
	Assigned [RegionCount]int // ds:D5AE
	Held     [RegionCount]int // ds:D5CE
	Decided  int              // ds:D60E
}

// Capitals 是兩邊的首都星域：帝國 0（瓦爾哈拉）、同盟 31（巴拉特）。
// `2E54:0342`、`2D54:0000` 都直接用這兩個常數（`docs/re/171`、`docs/re/179`）。
var Capitals = [2]int{0, 31}

// TurnInput 是算一回合那六張表要的東西。
//
// `Reachable` 對應原版的 `2B99:0392(首都, 陣營, &旗標)`：它對每一格填
// `2B99:028F(首都, i, 陣營) < 0x800`，也就是「走不走得到」（`docs/re/185`）。
//
// 星圖是全連通的（`rules.TestStarGraphFullyConnected`），所以正常局面下
// 那張表是全 1——`nil` 的預設「全部走得到」就是原版行為。原版的距離表
// 放在 EMS 裡、數值還沒取到，所以真有哪一格被標成走不到的話，
// 這裡要換成注入的版本。
type TurnInput struct {
	Faction   int
	Fleets    []FleetView
	Graph     *rules.StarGraph
	Owner     func(region int) int
	EdgeFlag  func(region int) bool // 星域 `+0x13` bit0
	Reachable func(region int) bool
	Rand      *rules.Rand
}

func (in TurnInput) reachable(r int) bool {
	if in.Reachable == nil {
		return true
	}
	return in.Reachable(r)
}

func (in TurnInput) edgeFlag(r int) bool {
	if in.EdgeFlag == nil {
		return false
	}
	return in.EdgeFlag(r)
}

// Enemy 回對面的陣營。
func Enemy(faction int) int { return 1 - faction }

// DefenceBoard 算到階段三為止的那幾張表（`2E54:000C` ＋ `2E54:0730`）：
// 敵方兵力圖、守勢優先度、要派幾支。
//
// `held` 是 `ds:D5CE`（這一格有幾支決定留守），由逐艦隊決定那一趟填
// （`2E54:005D`，`docs/re/173`）；這裡當輸入收，因為那一趟還沒實作。
func DefenceBoard(in TurnInput, held [RegionCount]int) Board {
	var b Board
	b.Held = held
	b.Threat = ThreatMap(in.Fleets, Enemy(in.Faction), in.Graph)
	b.Priority = DefencePriority(in.Graph, in.Owner, in.edgeFlag, in.reachable,
		b.Threat, in.Faction, Enemy(in.Faction))
	b.Need = DefenceNeed(b.Priority, b.Threat, b.Held,
		func(r int) bool { return in.Owner(r) == in.Faction })
	return b
}

// AttackBoard 算階段四那兩張重算過的表（`2D54:0000` ＋ `2E54:08E3`）。
//
// `state` 是戰略態勢碼（`StateCode`），`routes` 是進軍路線表。
// **會消耗亂數**：每個 `Priority != 0` 的格子取一次，依星域編號由小到大
// ——次序是量出來的（`docs/re/184`）。
//
// ⚠ 新遊戲第一個 AI 回合這裡算出 10 格非零（星域 0–2、4–9 與 13），原版
// 那一刻只有 13 非零，所以這一支多抽 9 次亂數（remake 16 次、原版 7 次，
// `docs/re/297`；worklist `ai-attack-reachability`）。最可能是 `Reachable`
// 在那個局面不該是全 1。叛亂判定與它共用同一個序列（`docs/spec/041`）。
func AttackBoard(in TurnInput, routes *AdvanceRoutes, state int, held [RegionCount]int) Board {
	var b Board
	b.Held = held
	b.Threat = ThreatMap(in.Fleets, Enemy(in.Faction), in.Graph)
	enemy := Enemy(in.Faction)
	enemyCapital := Capitals[enemy]
	b.Priority = AttackScore(routes.For(in.Faction, state), b.Threat, func(r int) bool {
		return in.reachable(r) && (in.Owner(r) == enemy || r == enemyCapital)
	})
	b.Need = AttackNeed(b.Priority, b.Threat, b.Held, func() int {
		if in.Rand == nil {
			return 0
		}
		return in.Rand.Next()
	})
	return b
}
