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
// `2B99:028F(首都, i, 陣營) < 0x800`（`docs/re/185`）。那張距離表帶著
// 領土罰則，**不是全 1**——從自己的地盤只走得到自己的地盤與外面那一圈
// （`CapitalReach`，規格 `docs/spec/043`）。`nil` ＝ 照 `Graph` 與 `Owner`
// 當場算，那就是原版行為；要注入別的表才填它。
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
	return in.Reachable(r)
}

// withReach 在 `Reachable` 沒給時照原版算一張（`CapitalReach`）。
// 兩張表的入口都先過它，一回合只算一次。
func (in TurnInput) withReach() TurnInput {
	if in.Reachable == nil {
		in.Reachable = reachFunc(in.Graph, in.Owner, in.Faction)
	}
	return in
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
	in = in.withReach()
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
// ——次序是量出來的（`docs/re/184`）。候選要先過「從首都走得到」
// （`CapitalReach`）：新遊戲第一個 AI 回合同盟只有伊謝爾倫一格非零、
// 抽一次，與原版相同（`docs/re/298`）。叛亂判定與它共用同一個序列
// （`docs/spec/041`），這裡多抽或少抽一次，之後每一條擲骰的規則都錯位。
func AttackBoard(in TurnInput, routes *AdvanceRoutes, state int, held [RegionCount]int) Board {
	in = in.withReach()
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
