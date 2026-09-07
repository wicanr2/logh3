package rules

// 新遊戲載入時整份重算星域表（`docs/re/44`）。
//
// 讀存檔不走這一支——那時除了未安定度以外的欄位都照存檔的用。

// Morale 回一個星域的民心：擁有者決定，帝國 0、同盟 200、無主 100。
func Morale(owner int) int {
	switch owner {
	case 0:
		return 0
	case 1:
		return 200
	}
	return 100
}

// TaxRatePerMille 回載入時算出來的稅率（千分之一）。
//
// `ruler` 是該陣營統治者的提督記錄，`base` 是難度表 `ds:147E` 那一格。
// 原版在 `090A:0CC0` 起。
func TaxRatePerMille(growth, morale, owner, base int, ruler []byte) int {
	if len(ruler) <= admiralTaxB {
		return 1
	}
	v := base * 5 / 100
	if growth > 0 {
		v = growth * base / 100
	}
	v = v * 100 / int(ruler[admiralRule])
	v = v * int(ruler[admiralTaxB]) / 100
	m := morale
	if owner != 0 {
		m = 200 - m
	}
	w := int(ruler[admiralTaxA])
	switch ratio := m * w / 100; {
	case ratio > 150:
		v = v * 30 / w
	case ratio > 100:
		v = v * 50 / w
	case ratio > 50:
		v = v * 75 / w
	}
	if v > 200 {
		v = 200
	}
	if v < 1 {
		v = 1
	}
	return v
}

// 統治者提督記錄裡被稅率算式用到的三格。`+0x0E` 未安定度也用
// （`docs/spec/006`），另外兩格目前只有這裡用得到。
const (
	admiralTaxA = 0x14
	admiralTaxB = 0x15
)

// NewGameRegions 把 `STELLAR.DTX` 重算成新遊戲開局的執行期星域表。
//
// 順序照原版：擁有者與行星數 → 經濟力與成長率 → 民心 → 稅率 →
// `+0x27` 複製 → 未安定度。
func NewGameRegions(stellar []byte, planets *Planets, admirals []byte,
	table *GrowthTable, graph *StarGraph, params UnrestParams) []byte {
	out := make([]byte, len(stellar))
	copy(out, stellar)
	for i := 0; i < RegionCount && (i+1)*RegionRecord <= len(out); i++ {
		r := out[i*RegionRecord : (i+1)*RegionRecord]
		e, a, o := planets.RegionPlanets(i)
		owner := planets.RegionOwner(i)
		r[regionOwner] = byte(owner)
		r[regionCount], r[regionCount+1], r[regionCount+2] = byte(e), byte(a), byte(o)
		r[regionFlag13] = 0
		eco, growth := RegionEconomyAndGrowth(planets, table, i)
		r[regionEconomy] = byte(eco)
		r[regionEconomy+1] = byte(eco >> 8)
		r[regionEconomy+2] = byte(eco >> 16)
		r[regionEconomy+3] = byte(eco >> 24)
		r[regionGrowth], r[regionGrowth+1] = byte(growth), byte(growth>>8)
		r[regionMorale] = byte(Morale(owner))
		// 無主星域沒有統治者可以套，稅率就停在「成長率 × 難度 / 100」
		// 那一步——量到的九個無主星域稅率與成長率一格不差。
		tax := growth
		if owner != 2 {
			tax = TaxRatePerMille(growth, Morale(owner), owner,
				params.Base[params.Rate[owner]],
				admiralRecord(admirals, params.Ruler[owner]))
		}
		if tax > 200 {
			tax = 200
		}
		if tax < 1 {
			tax = 1
		}
		r[regionTaxRate] = byte(tax)
		r[regionTaxPre] = r[regionTaxRate]
	}
	st := &UnrestState{Regions: out, Planets: planets, Admirals: admirals,
		Graph: graph, Params: params}
	unrest := make([]int, RegionCount)
	for i := range unrest {
		unrest[i] = st.Unrest(i)
	}
	for i, v := range unrest {
		out[i*RegionRecord+RegionUnrest] = byte(v)
	}
	return out
}

func admiralRecord(b []byte, i int) []byte {
	if i < 0 || (i+1)*AdmiralRecord > len(b) {
		return nil
	}
	return b[i*AdmiralRecord : (i+1)*AdmiralRecord]
}
