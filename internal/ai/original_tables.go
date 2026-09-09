package ai

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/rules"
)

// 這一份是原版戰略 AI 每回合要算的那六張「以星域為索引」的表
// （規格 `docs/spec/029`、`docs/internals/54`）。
//
// 都是純函式：吃局面、回表，不改任何狀態，也還沒接到 `advanceTurn`。
// 收據 `s3aicomp`（`docs/re/176`）的數字由 `original_tables_test.go` 釘住。

// FleetView 是算這些表要看的艦隊欄位，對應原版艦隊記錄
// （`docs/internals/33`）。獨立成一個型別，是為了讓這一層可以在
// 艦隊載入器補上 `+0x0F` 之前就先測。
type FleetView struct {
	Faction int // `+0x01`
	Region  int // `+0x06` 所在星域
	Target  int // `+0x07` 要去的星域；rules.NoTarget ＝ 沒有
	Action  int // `+0x08` 移動狀態：5 ＝ 不動、0／1 ＝ 會走
	Groups  int // `+0x0F` 這一支有幾組編制
}

// RegionCount 是星域的筆數。原版是 32 筆（`docs/internals/22`）。
const RegionCount = 32

// ThreatMap 回原版的 `ds:D54E`：**敵方艦隊「下一步會在哪一格」的
// 編制槽總數**（`2E54:1A76`，`docs/re/173`）。
//
// 移動中的艦隊算在**下一格**，所以這是一張預測的威脅圖，不是現況。
//
// 下一格走 `StarGraph.NextStep`，那是原版 `2B99:0302` 的等價物
// （`docs/re/183` 讀完 `2B99:0349` 之後證出來的，不是假設）。
func ThreatMap(fleets []FleetView, enemy int, g *rules.StarGraph) [RegionCount]int {
	var t [RegionCount]int
	for _, f := range fleets {
		if f.Faction != enemy {
			continue
		}
		cell := -1
		switch f.Action {
		case 0, 1:
			if f.Target == rules.NoTarget || g == nil {
				continue
			}
			cell = g.NextStep(f.Region, f.Target, f.Faction)
		case 5:
			cell = f.Region
		default:
			continue
		}
		if cell >= 0 && cell < RegionCount {
			t[cell] += f.Groups
		}
	}
	return t
}

// DefencePriority 回原版的 `ds:D50E` 的**守勢版**
// （`2E54:019F` → `2E54:0222`，`docs/re/171`）：
//
//	只算「與首都連通」而且「擁有者 ＝ 自己」的格子，其餘 0
//	分數 = Σ（相鄰、擁有者 ＝ 敵方的 n）: (n 的旗標 bit0 沒設 或 threat[n] > 4) ? 4 : 2
//
// `edgeFlag` 是星域記錄 `+0x13` 的 bit0（語意未解，`docs/internals/22`）。
func DefencePriority(g *rules.StarGraph, owner func(int) int, edgeFlag func(int) bool,
	reachable func(int) bool, threat [RegionCount]int, faction, enemy int) [RegionCount]int {
	var p [RegionCount]int
	for i := 0; i < RegionCount; i++ {
		if !reachable(i) || owner(i) != faction {
			continue
		}
		sum := 0
		for _, n := range g.Neighbours(i) {
			if n == i || n < 0 || n >= RegionCount || owner(n) != enemy {
				continue
			}
			if !edgeFlag(n) || threat[n] > 4 {
				sum += 4
			} else {
				sum += 2
			}
		}
		p[i] = sum
	}
	return p
}

// AttackScore 回原版的 `ds:D50E` 的**攻勢版**（`2D54:0000`，
// `docs/re/174`、`docs/re/178`）：
//
//	先把進軍路線表裡出現的每一格加 1
//	再對每個候選格：++ 之後乘上 2048 ÷ (threat + 1)
//
// 不是候選的格子**保留路線表的計數、不歸零**——原版就是 `跳過`。
func AttackScore(lists [][]int, threat [RegionCount]int, candidate func(int) bool) [RegionCount]int {
	var s [RegionCount]int
	for _, l := range lists {
		for _, r := range l {
			if r >= 0 && r < RegionCount {
				s[r]++
			}
		}
	}
	for i := 0; i < RegionCount; i++ {
		if !candidate(i) {
			continue
		}
		s[i]++
		s[i] *= 2048 / (threat[i] + 1)
	}
	return s
}

// DefenceNeed 回階段三的 `ds:D58E`（`2E54:0730`，`docs/re/172`）：
//
//	前線 = min(priority ÷ 2, 1)
//	要幾支 = (held < threat ? (threat − held) ÷ 4 : 0) + 前線
//
// 只算擁有者是自己的格子。
func DefenceNeed(priority, threat, held [RegionCount]int, mine func(int) bool) [RegionCount]int {
	var n [RegionCount]int
	for i := 0; i < RegionCount; i++ {
		if !mine(i) {
			continue
		}
		front := priority[i] / 2
		if front > 1 {
			front = 1
		}
		if held[i] < threat[i] {
			n[i] = (threat[i] - held[i]) / 4
		}
		n[i] += front
	}
	return n
}

// AttackNeed 回階段四的 `ds:D58E`（`2E54:08E3`，`docs/re/174`）。
//
// 與 `DefenceNeed` 三處不同：閘門看 `priority != 0` 而不是擁有者、
// 除數是 5 不是 4、加成是 `RNG() % 4` 而不是前線 0／1。
//
// `rng` 要回原版 `3CD5:000C` 的傳回值（0–0x7FFF）。**remake 還沒把
// 原版 RNG 的呼叫次序對上**，所以這一支目前只能用在測試裡
// （`docs/spec/029`「不能做的」）。
func AttackNeed(priority, threat, held [RegionCount]int, rng func() int) [RegionCount]int {
	var n [RegionCount]int
	for i := 0; i < RegionCount; i++ {
		if priority[i] == 0 {
			continue
		}
		if held[i] < threat[i] {
			n[i] = (threat[i] - held[i]) / 5
		}
		n[i] += rng() % 4
	}
	return n
}

// Unsatisfied ＝ 原版 `2E54:088F(星域)`：這一格的滿足率（百分比）。
//
// `need <= assigned` 回 0xFF（夠了）、兩邊的首都回 0，其餘回
// `assigned × 100 ÷ need`。`2E54:0842` 挑的就是這個值最小的一格。
func Unsatisfied(region int, need, assigned [RegionCount]int, capitals [2]int) int {
	if need[region] <= assigned[region] {
		return 0xFF
	}
	if region == capitals[0] || region == capitals[1] {
		return 0
	}
	return assigned[region] * 100 / need[region]
}

// PickCell ＝ 原版 `2E54:0842()`：回「`need != 0` 之中滿足率最低」的那一格，
// 沒有回 -1。原版回 `0xFF`。
func PickCell(need, assigned [RegionCount]int, capitals [2]int) int {
	best, bestV := -1, 0xFF
	for i := 0; i < RegionCount; i++ {
		if need[i] == 0 {
			continue
		}
		if v := Unsatisfied(i, need, assigned, capitals); v < bestV {
			best, bestV = i, v
		}
	}
	return best
}

// StateCode ＝ 原版 `2D54:0163()`：0–11 的戰略態勢碼（`docs/re/179`）。
//
// 只看伊謝爾倫（星域 13）在誰手上、費沙（星域 14）在誰手上，
// 以及「費沙有沒有進入戰局」（原版的 `ds:8F48`）。
func StateCode(iserlohnOwner, phezzanOwner int, phezzanOpen bool) int {
	base := 8
	switch iserlohnOwner {
	case 0:
		base = 0
	case 1:
		base = 4
	}
	if !phezzanOpen {
		return base
	}
	switch phezzanOwner {
	case 0:
		return base + 1
	case 1:
		return base + 2
	}
	return base + 3
}

// Iserlohn／Phezzan 是原版寫死的兩個星域編號（`docs/re/178`）。
const (
	Iserlohn = 13
	Phezzan  = 14
)

// AdvanceRoutes 是原版寫死的進軍路線表（`docs/internals/data/ai-advance-routes.json`）。
// 每個陣營 12 組（一個戰略態勢碼一組），每組數條路線。
type AdvanceRoutes struct {
	Empire   [][][]int
	Alliance [][][]int
}

// LoadAdvanceRoutes 讀 `docs/internals/data/ai-advance-routes.json`。
func LoadAdvanceRoutes(path string) (*AdvanceRoutes, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Routes map[string][]struct {
			N     int `json:"n"`
			Lists []struct {
				Regions []int `json:"regions"`
			} `json:"lists"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := &AdvanceRoutes{}
	for key, entries := range raw.Routes {
		side := make([][][]int, 12)
		for _, e := range entries {
			if e.N < 0 || e.N >= 12 {
				return nil, fmt.Errorf("態勢碼 %d 超出 0–11", e.N)
			}
			for _, l := range e.Lists {
				side[e.N] = append(side[e.N], l.Regions)
			}
		}
		switch {
		case len(key) >= 6 && key[:6] == "帝國":
			out.Empire = side
		default:
			out.Alliance = side
		}
	}
	if len(out.Empire) != 12 || len(out.Alliance) != 12 {
		return nil, fmt.Errorf("路線表要兩個陣營各 12 組")
	}
	return out, nil
}

// For 回某個陣營在某個戰略態勢下的路線。
func (r *AdvanceRoutes) For(faction, state int) [][]int {
	if r == nil || state < 0 || state >= 12 {
		return nil
	}
	if faction == 0 {
		return r.Empire[state]
	}
	return r.Alliance[state]
}
