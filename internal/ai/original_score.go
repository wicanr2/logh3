package ai

import "github.com/wicanr2/logh3/internal/rules"

// 候選艦隊的挑選與評分（原版 `2E54:143F`／`2E54:1574`，`docs/re/168`、
// 規格 `docs/spec/029`）。

// PickerMode 是 `2E54:143F` 的三種模式。原版三個呼叫點各用一種：
// 模式 0 是階段三的指派、模式 1 是 `2E54:0AAB` 的增援、
// 模式 2 是階段二打無主星域（`docs/re/177`）。
type PickerMode struct {
	Weights   [4]int // 依艦隊類別選一個，乘進分數
	Integrity int    // 完整度（`+0x3C`）門檻
	NeedFull  bool   // 要求 `+0x0F == +0x0E`
	MainMin   int    // 主力槽數不到這個數，分數 ÷ 4
}

// PickerModes 是原版寫死的三組（`docs/re/168`）。
var PickerModes = [3]PickerMode{
	{Weights: [4]int{3, 2, 8, 1}, Integrity: 50, NeedFull: false, MainMin: 0},
	{Weights: [4]int{4, 6, 1, 2}, Integrity: 90, NeedFull: true, MainMin: 0},
	{Weights: [4]int{3, 4, 0, 0}, Integrity: 80, NeedFull: true, MainMin: 4},
}

// Candidate 是評分要看的艦隊欄位，對應原版艦隊記錄（`docs/internals/33`）。
type Candidate struct {
	FleetView

	// Index ＝ 這是呼叫端那份艦隊清單的第幾筆。評分不看它，
	// 挑完之後要拿它回頭指到艦隊。
	Index int
	Disabled  bool // `+0x00` bit0：無法戰鬥
	Decided   bool // `+0x79` 不是 0xFF：這回合已經決定過
	Morale    int  // `+0x3C`
	Expected  int  // `+0x0E`；0 ＝ 當成與 Groups 相同（存檔還沒定位到這一格）
	Class     int  // ds:5F52[`+0x09`] 換出來的 0–3（`docs/re/190`）
	Ships     int  // Σ 各編制槽艦艇的 `+0x08`
	MainSlots int  // 艦種 0 或 1 的槽數
}

// Score 回原版 `2E54:1574` 的分數。
//
//	無法戰鬥或走不到 → 0
//	分數 = (艘數 ÷ 10) × 士氣 ÷ 100 × 編制槽數 ÷ 應有數 × 類別權重
//	     × (順路 ? 2 : 1) ÷ (距離 × 2 + 1) ÷ (主力不足 ? 4 : 1)
//	算完是 0 就夾成 1
//
// 距離用 `StarGraph.Hops`——原版是查 EMS 裡的距離表（`docs/re/185`），
// 而 `docs/re/43` 量過那個表與相鄰圖的最短步數相同。
func (m PickerMode) Score(c Candidate, target int, g *rules.StarGraph) int {
	if c.Disabled {
		return 0
	}
	d := g.Hops(c.Region, target)
	if d < 0 {
		return 0
	}
	expected := c.Expected
	if expected == 0 {
		expected = c.Groups
	}
	if expected == 0 {
		return 0
	}
	n := c.Ships / 10 * c.Morale / 100
	n = n * c.Groups / expected
	if c.Class >= 0 && c.Class < len(m.Weights) {
		n *= m.Weights[c.Class]
	}
	if c.Action != 5 && c.Target == target {
		n *= 2
	}
	n /= d*2 + 1
	if c.MainSlots < m.MainMin {
		n /= 4
	}
	if n == 0 {
		n = 1
	}
	return n
}

// Eligible 回這一支過不過得了 `2E54:143F` 的過濾。
//
// 「在用」與「陣營」由呼叫端先篩掉；這裡看的是決定狀態、完整度門檻與
// 「編制不缺」。模式 2 另有一條：艦隊已經在目標星域時門檻降到 50。
func (m PickerMode) Eligible(c Candidate, target int) bool {
	if c.Decided {
		return false
	}
	threshold := m.Integrity
	if m.MainMin == 4 && c.Region == target { // 模式 2 的特例
		threshold = 50
	}
	if c.Morale < threshold {
		return false
	}
	if m.NeedFull && c.Expected != 0 && c.Groups != c.Expected {
		return false
	}
	return true
}

// Pick 回分數最高的那一支的索引；一支都沒有回 -1（原版回 0:0）。
func (m PickerMode) Pick(cs []Candidate, target int, g *rules.StarGraph) int {
	best, bestScore := -1, 0
	for i, c := range cs {
		if !m.Eligible(c, target) {
			continue
		}
		if s := m.Score(c, target, g); s > bestScore {
			best, bestScore = i, s
		}
	}
	return best
}
