package rules

// `Blend` 與稅率重算（`docs/spec/030`、`docs/internals/55`）。
//
// `Blend`（原版 `ds:8FD4[陣營]`）是未安定度算式的參數，同時是每個星域
// 稅率的級距。動它的只有電腦座位的戰略 AI：回合開頭按軍事費設一次、
// 回合結尾（非月底）歸零，兩次都接著重算該陣營所有星域的稅率
// （`docs/re/222`、`docs/re/223`）。

// blendSteps 是軍事費 → 等級的分段函式（原版 `2D10:0319`）。
// **不是單調的**：50000 以上回到 2，與 10000–40000 那一段同級。
var blendSteps = []struct{ min, level int }{
	{50000, 2}, {40000, 1}, {10000, 2}, {5000, 3}, {0, 4},
}

// BlendForFunds 回那個陣營這一回合的 `Blend`。軍事費是有號的，負數回 2。
func BlendForFunds(funds int) int {
	for _, s := range blendSteps {
		if funds >= s.min {
			return s.level
		}
	}
	return 2
}

// NewGameBlend 是新遊戲開局兩個陣營的 `Blend`（796/1 量到的，
// `docs/spec/006` 的 params 表）。
const NewGameBlend = 2

// BlendTightBudget 是建艦隊的預算會被打七五折的門檻（原版 `2D76:0060`
// 的 `ds:8FD4[陣營] >= 3`，`docs/re/202`）。換算成軍事費就是不到 10000。
const BlendTightBudget = 3

// RegionGrowth 回星域記錄的成長率（`+0x1E`，千分之一為單位）。
// 原版當成**有號 16 位元**讀（`090A:0B93` 的 `cmp word [bx+1Eh], 0`）。
func RegionGrowth(record []byte) int {
	if len(record) < regionGrowth+2 {
		return 0
	}
	v := int(record[regionGrowth]) | int(record[regionGrowth+1])<<8
	if v >= 0x8000 {
		v -= 0x10000
	}
	return v
}

// RecomputeTaxRates 把一個陣營所有星域的稅率（`+0x26`）重算一遍
// （原版 `090A:0B93`）。
//
// `base` ＝ 難度表 `ds:147E` 的第 `Blend[陣營]` 格，`ruler` 是該陣營
// 統治者的提督記錄。擁有者照行星現算（與 `estimateTaxFor` 同一個來源），
// 星域記錄裡那一格在讀存檔時可能是舊的。
func RecomputeTaxRates(records []byte, faction, base int, ruler []byte, owner func(int) int) {
	for i := 0; (i+1)*RegionRecord <= len(records); i++ {
		if owner(i) != faction {
			continue
		}
		rec := records[i*RegionRecord : (i+1)*RegionRecord]
		rate := TaxRatePerMille(RegionGrowth(rec), int(rec[regionMorale]),
			faction, base, ruler)
		rec[regionTaxRate] = byte(rate)
	}
}

// NeutralTaxLevel 是原版重算無主星域稅率時傳的難度下標（`090A:0B5E` 第三段
// `090A:0DC0(2)`），`Base[2]` ＝ 100，所以無主星域的稅率就是它的成長率。
const NeutralTaxLevel = 2

// RecomputeNeutralTaxRates 把無主星域（擁有者 2）的稅率重算一遍（原版
// `090A:0DC0(等級)`，`docs/re/321`）：成長率 ≤ 0 → `base × 5 ÷ 100`，否則
// `成長率 × base ÷ 100`，只寫低位元組、沒有夾限。一個月一次，跟在兩個陣營
// 的 `RecomputeTaxRates` 後面。
func RecomputeNeutralTaxRates(records []byte, base int, owner func(int) int) {
	for i := 0; (i+1)*RegionRecord <= len(records); i++ {
		if owner(i) != RegionOwnerNone {
			continue
		}
		rec := records[i*RegionRecord : (i+1)*RegionRecord]
		v := base * 5 / 100
		if g := RegionGrowth(rec); g > 0 {
			v = g * base / 100
		}
		rec[regionTaxRate] = byte(v)
	}
}

// MaxBlendLevel 是稅率檔位的上限：0–4 對應 微稅／輕稅／普通／重稅／緊急
// （預算計劃那兩顆箭頭 clamp 在這個範圍，`docs/re/343`）。
const MaxBlendLevel = 4
