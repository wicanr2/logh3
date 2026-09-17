// Package ai 是電腦玩家的決策層（`docs/spec/017`）。
//
// **兩套並存**：`Original` 照原版，`Focused` 是強化版，會挑目標、集結、包圍。
//
// 原版**有**一整套戰略 AI（`docs/re/162`–`docs/re/174`、`docs/internals/54`）：
// 建艦隊的類別配額、敵方兵力圖、每一格的優先度與「要派幾支」、
// 挑候選與評分、指派。**那一套已經接上了**（`OriginalStrategy`，2026-09-10
// 使用者裁決）——`Game` 預設就用它；下面那個 `Original` 空殼留著只是為了
// 「什麼都不做」這個語意。
//
// 25 份記憶體傾印上決定計數器 `ds:D60E` 都是 0，原因是那些傾印停在
// **玩家的座位**上——AI 的進入點 `2D10:0005` 只在 `ds:8F23[ds:8F1E] == 1`
// （這個座位是電腦）時才被呼叫（`docs/re/175`、`docs/re/189`）。
// 所以那不是「原版不下命令」的證據。
//
// 畫面收據 `turn260`／`turn320`／`turn450` 則相反：**原版在那個狀態裡
// 跑過 AI，指派了三支艦隊**（`docs/re/198`）。那三張現在還能逐位元組
// 通過，是因為指派寫的是艦隊記錄的 `+0x07`／`+0x79`，按完「實行」的
// 那一刻畫面上還看不出來。
//
// 預設走 `Original`，強化版由玩家自己開；這樣對拍收據與「與原版逐位元組
// 相同」那個宣稱才站得住。
package ai

import "github.com/wicanr2/logh3/internal/rules"

// Order 是一道命令：叫第 Fleet 支艦隊去 Target 那個星域。
//
// 對應原版的兩個欄位——`+0x07` 寫目的地、`+0x08` 解開移動狀態
// （`docs/internals/33`）。
type Order struct {
	Fleet  int `json:"fleet"`  // 索引，對應 World.Fleets
	Target int `json:"target"` // 目的地星域
	// Move 是寫進艦隊 `+0x08` 的命令（0 移動、1 占領、2 駐留、3 輸送；`docs/internals/33`）。
	// 原版每一條下命令的路都寫這一格（`docs/internals/54`）；零值 ＝ 移動。
	Move int `json:"move"`
	// Direct ＝ 直接寫 `+0x07`／`+0x08`（`2E54:005D` 的原地出擊與退一格）；否則走 `2E54:1B17`：
	// 艦隊已經在目標格就改成停駐（`+0x08 = 5`、`+0x07 = 0xFF`）。
	Direct bool `json:"direct,omitempty"`
}

// World 是 AI 看得到的局面。**只讀**——AI 只回命令，不自己改狀態。
type World struct {
	Fleets []rules.Fleet
	Graph  *rules.StarGraph

	// Owner 回某個星域屬於誰：0 帝國、1 同盟，其餘 ＝ 無主。
	Owner func(region int) int

	// Value 回攻下某個星域值多少（經濟力之類）。nil ＝ 每個都算 1。
	Value func(region int) int

	// Strength 回一支艦隊的戰力。nil ＝ 每支算 1。
	//
	// 原版自己的算式在 `2E54:1876`（`docs/re/79`）：每一組編制
	// `ds:6032[艦種+4] × 一單位幾艘 × 1000 ÷ 艦種+0x0A` 加總。
	// remake 還沒把艦艇池接進來，所以這裡開一個口，接上之後換掉即可。
	Strength func(f rules.Fleet) int
}

func (w World) owner(region int) int {
	if w.Owner == nil {
		return -1
	}
	return w.Owner(region)
}

func (w World) value(region int) int {
	if w.Value == nil {
		return 1
	}
	return w.Value(region)
}

func (w World) strength(f rules.Fleet) int {
	if w.Strength == nil {
		return 1
	}
	return w.Strength(f)
}

// Brain 是一套決策。每回合叫一次，回這一回合要下的命令。
type Brain interface {
	Orders(w World, faction, turn int) []Order
}

// NewFleet 是這一回合要**新編**的一支艦隊。
//
// 原版的階段一（`2D76:0002` 主動建、`2E54:0513` 守首都補）產生的不是
// 命令而是新艦隊，`[]Order` 裝不下，所以另開一條路
// （`docs/internals/54`、`docs/re/201`–`docs/re/204`）。
type NewFleet struct {
	// Admiral 是帶隊提督在提督表裡的索引。
	Admiral int
	// Region 是編在哪一格。原版永遠是自己的首都。
	Region int
	// ClassKey 是艦隊記錄的 `+0x09`（0–15，`ai.PickClassKey` 抽出來的）。
	ClassKey int
	// Adjutant 是副官的提督索引；-1 ＝ 沒有。
	Adjutant int
	// Expected 是艦隊記錄的 `+0x0E`（應有幾組編制）＝ `rules.FleetSlots`
	// 查提督等級（原版 `297B:0066` → `297B:0209`，`docs/re/214`）。
	Expected int

	// Composition／Operations／Upkeep 由 `BuildSupport.Form` 填：編制、
	// 艦隊能力第 10 項（`+0x47` 營運）、以及這一支的維持費（`+0x0A`）。
	//
	// **維持費要在這一趟裡就算出來**：原版每建一支就重算 `ds:8F72`，
	// 下一輪的閘門看的是新值（`docs/re/218`）。
	Composition []rules.FleetGroup
	Operations  int
	Upkeep      int
}

// Builder 是「除了下命令，還會建艦隊」的 `Brain`。
//
// **選配**：`Game` 用型別斷言問一次，沒實作就不建。這樣既有的
// `Original`／`Focused` 不必跟著改。
type Builder interface {
	Build(w World, faction, turn int) []NewFleet
}

// Refitter ＝ 這個腦會跑原版的**整補**（`2D76:0732`，`docs/spec/031`）：
// 回合開頭建完艦隊之後，把自己陣營、停在自己星域而且那一格沒有戰鬥的
// 每一支艦隊的編制併殘部、再補回應有組數。
//
// 整補不需要決策——要補哪一組、犧牲哪一組全部由樣板與存量決定——
// 所以這個介面只是一個標記，`Game` 拿它決定要不要跑那一步。
// 整補是 **AI 專屬**的（三個入口各只有一個呼叫者，`docs/re/229`）。
type Refitter interface {
	OriginalRefit()
}

// Original 照原版。目前回 nil ＝ **還沒接上**原版那套戰略 AI
// （`docs/internals/54`）；在電腦座位真的跑起來的收據做出來之前，
// 這裡不猜。
type Original struct{}

// Orders 回 nil。
func (Original) Orders(World, int, int) []Order { return nil }
