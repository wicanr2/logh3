package rules

import (
	"encoding/json"
	"os"
)

// GrowthTable 是行星經濟力每月的成長率階梯表（`ds:4A68`，`docs/re/33`）。
//
// 查法照原版那支迴圈：**找到第一列 i 使得第 i+1 列的門檻 >= 經濟力**，
// 取第 i 列的成長率。經濟力越大成長越慢，二十萬以上完全不長。
type GrowthTable struct {
	Rows []struct {
		Threshold    int `json:"threshold"`
		RatePerMille int `json:"rate_per_mille"`
	} `json:"rows"`
	DefenceBonus struct {
		Below    int `json:"below"`
		Multiply int `json:"multiply"`
		Divide   int `json:"divide"`
	} `json:"defence_bonus"`
}

// LoadGrowthTable 讀成長率表。
func LoadGrowthTable(path string) (*GrowthTable, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t GrowthTable
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Rate 回一顆行星這個月的成長率（千分之一）。
//
// 防衛力低於門檻（3000）時再 ×3/2。**用的是行星的 `+0x16`，不是畫面上
// 那個 `+0x14`**（`docs/re/44`）——兩格在量到的資料裡數值一樣，
// 所以先前分不出來。
func (t *GrowthTable) Rate(economy, defence int) int {
	if len(t.Rows) == 0 {
		return 0
	}
	i := 0
	for i+1 < len(t.Rows) && t.Rows[i+1].Threshold < economy {
		i++
	}
	r := t.Rows[i].RatePerMille
	if b := t.DefenceBonus; b.Below > 0 && defence < b.Below && b.Divide != 0 {
		r = r * b.Multiply / b.Divide
	}
	return r
}

// planetEconomy 是行星記錄裡的經濟力欄位。
const planetEconomy = 0x1A

// PlanetEconomy 回一顆行星的經濟力。
func (p *Planets) PlanetEconomy(i int) int {
	r := p.rec(i)
	return int(r[planetEconomy]) | int(r[planetEconomy+1])<<8 |
		int(r[planetEconomy+2])<<16 | int(r[planetEconomy+3])<<24
}

// SetPlanetEconomy 寫回一顆行星的經濟力。
func (p *Planets) SetPlanetEconomy(i, v int) {
	r := p.rec(i)
	r[planetEconomy] = byte(v)
	r[planetEconomy+1] = byte(v >> 8)
	r[planetEconomy+2] = byte(v >> 16)
	r[planetEconomy+3] = byte(v >> 24)
}

// MonthEndRegion 跑一個星域的月底結算，回**進國庫的稅**與結算後的
// 星域經濟力（`docs/re/33`、`docs/re/35`）。
//
//	對每一顆行星：
//	    經濟力 += 經濟力 × 成長率(經濟力, 防衛力) / 1000
//	    行星稅  = 經濟力 × 星域稅率 / 1000      （捨去）
//	    經濟力 -= 行星稅
//	星域經濟力 = Σ 該星域行星的經濟力
//	進國庫的稅 = (Σ 成長後、扣稅前的經濟力) × 星域稅率 / 1000   （捨去）
//
// **兩個稅不是同一個數。** 從行星身上扣的是逐顆捨去的和，進國庫的是拿
// 星域總和算一次捨去——星域 0 是 2718 與 2720，十個帝國星域加起來差 9。
// 原版就是這樣：監看總稅收看到的十個增量與「用總和算」逐一相符。
func (p *Planets) MonthEndRegion(t *GrowthTable, region, taxPerMille int) (tax, economy int) {
	p.GrowRegion(t, region)
	return p.TaxRegion(region, taxPerMille)
}

// GrowRegion 是月底的**成長**那一步（原版 `2703:0C26`）：這個星域每顆行星
// 各長一次，回成長前與成長後的經濟力總和。
//
// 原版拿這兩個數重算星域的成長率 `+0x1E`：
//
//	成長率 ＝ (成長後 − 成長前) × 1000 ÷ 成長前
//
// 也就是**用這個月實際長了多少反推**，不是沿用開局那一份
// （`docs/spec/038`、`docs/re/267`）。
func (p *Planets) GrowRegion(t *GrowthTable, region int) (before, after int) {
	start, n := p.Range(region)
	for i := start; i < start+n; i++ {
		e := p.PlanetEconomy(i)
		before += e
		e += e * t.Rate(e, p.Defence(i)) / 1000
		p.SetPlanetEconomy(i, e)
		after += e
	}
	return before, after
}

// TaxRegion 是月底的**收稅**那一步（原版 `2703:0AB8` 旗標 1）：逐行星扣稅，
// 回進國庫的稅與扣完稅的經濟力總和。
//
// **兩個稅不是同一個數**（`docs/re/35`）：從行星身上扣的是逐顆捨去的和，
// 進國庫的是拿成長後、扣稅前的總和算一次捨去。
func (p *Planets) TaxRegion(region, taxPerMille int) (tax, economy int) {
	start, n := p.Range(region)
	grown := 0
	for i := start; i < start+n; i++ {
		e := p.PlanetEconomy(i)
		grown += e
		e -= e * taxPerMille / 1000
		p.SetPlanetEconomy(i, e)
		economy += e
	}
	return grown * taxPerMille / 1000, economy
}

// GrowthFromChange 是原版 `2703:0C26` 尾段那一行：拿實際的成長反推成長率。
func GrowthFromChange(before, after int) int {
	if before <= 0 {
		return 0
	}
	return (after - before) * 1000 / before
}

// SetRegionGrowth 寫回星域記錄的成長率（`+0x1E`，有號 16 位元）。
func SetRegionGrowth(record []byte, v int) {
	if len(record) < regionGrowth+2 {
		return
	}
	record[regionGrowth], record[regionGrowth+1] = byte(v), byte(v>>8)
}

// SetRegionEconomy 寫回星域記錄的經濟力。
func SetRegionEconomy(record []byte, v int) {
	if len(record) < regionEconomy+4 {
		return
	}
	record[regionEconomy] = byte(v)
	record[regionEconomy+1] = byte(v >> 8)
	record[regionEconomy+2] = byte(v >> 16)
	record[regionEconomy+3] = byte(v >> 24)
}

// MonthEnd 跑一次完整的月底結算：每個星域逐行星成長與收稅，
// 經濟力寫回星域記錄，稅率抄一份到 `+0x27`，回玩家陣營收到的總稅收。
//
// 順序照 `docs/re/26` 量到的：總稅收歸零 → 逐星域累加 → 軍事費 += 總稅收
// （加錢那一步由呼叫端做）。
func MonthEnd(records []byte, p *Planets, t *GrowthTable, faction int) int {
	return MonthEndTotals(records, p, t, nil)[faction]
}

// MonthEndTotals 是同一趟結算，但**兩個陣營的徵稅額一起回**。
//
// 成長是就地寫回的，一個月只能跑一次，所以要兩邊的數字就得在同一趟裡拿
// （原版的 `ds:8F6A` 本來就是每個陣營一格，`docs/internals/27`）。
func MonthEndTotals(records []byte, p *Planets, t *GrowthTable,
	recomputeRates func()) [2]int {
	// 第一趟（原版 `0654:0007` → `2703:0BF2`，一個月只跑一次，而且跑在
	// 逐座位之前）：32 格全部成長，並用實際成長重算成長率。
	for i := 0; i*RegionRecord+RegionRecord <= len(records); i++ {
		rec := records[i*RegionRecord : (i+1)*RegionRecord]
		before, after := p.GrowRegion(t, i)
		SetRegionEconomy(rec, after)
		SetRegionGrowth(rec, GrowthFromChange(before, after))
	}
	// 稅率吃成長率，所以在收稅**之前**重算（原版 `090A:0B5E`）。
	if recomputeRates != nil {
		recomputeRates()
	}
	// 第二趟（原版 `2703:0AB8` 旗標 1）：用新稅率收稅。
	var total [2]int
	for i := 0; i*RegionRecord+RegionRecord <= len(records); i++ {
		rec := records[i*RegionRecord : (i+1)*RegionRecord]
		tax, economy := p.TaxRegion(i, RegionTaxRate(rec))
		SetRegionEconomy(rec, economy)
		rec[0x27] = rec[regionTaxRate]
		if o := p.RegionOwner(i); o == 0 || o == 1 {
			total[o] += tax
		}
	}
	return total
}

// RegionEconomyAndGrowth 回一個星域載入時重算出來的經濟力與成長率
// （`docs/re/44`）。
//
// 經濟力 ＝ 該星域行星的經濟力相加；成長率 ＝ 「全部長一次之後的合計」
// 比經濟力多出來的比例（千分之一）。**新遊戲才會重算**——讀存檔時
// 這兩欄照存檔的用，只有未安定度重算（`docs/spec/006`）。
func RegionEconomyAndGrowth(p *Planets, t *GrowthTable, region int) (economy, growth int) {
	start, n := p.Range(region)
	grown := 0
	for i := start; i < start+n; i++ {
		e := p.PlanetEconomy(i)
		economy += e
		grown += e + e*t.Rate(e, p.GrowthDefence(i))/1000
	}
	if economy > 0 {
		growth = (grown - economy) * 1000 / economy
	}
	return economy, growth
}

// planetShare 是行星記錄的 `+0x13`：這顆佔該星域經濟力的整數百分比
// （`docs/internals/29`）。
const planetShare = 0x13

// PayIntoRegion ＝ 原版 `2703:1C11(星域, 金額)`：軍事費花出去的錢**流進那一格
// 行星的經濟力**——逐顆 `經濟力 += 金額 × +0x13 ÷ 100`，除不盡的餘數全給
// 百分比最大的那一顆（相同時取後面那一顆，`jb` 是嚴格小於才跳過）。
//
// 兩個呼叫者（`docs/re/313`）：
//   - 維持費（`056A:09EB`）：艦隊所在星域是自己的才付，付 `+0x0A ÷ 2`；
//   - 移動經費（`113A:06C3`）：走一格時付整筆 `+0x0C`，自己擁有現在這一格就
//     付給它，否則付給下一格。
//
// 星域記錄的經濟力不在這裡動——它是 `090A:0E6D` 每次把行星加總回去的
// （`2703:02E5` → `2703:034E`，`ResumRegionEconomy`）。
func (p *Planets) PayIntoRegion(region, amount int) {
	if p == nil || amount <= 0 {
		return
	}
	start, n := p.Range(region)
	if n == 0 {
		return
	}
	best, bestShare := -1, -1
	left := amount
	for i := start; i < start+n; i++ {
		r := p.rec(i)
		share := int(r[planetShare])
		if share >= bestShare {
			best, bestShare = i, share
		}
		part := amount * share / 100
		p.SetPlanetEconomy(i, p.PlanetEconomy(i)+part)
		left -= part
	}
	if left != 0 && best >= 0 {
		p.SetPlanetEconomy(best, p.PlanetEconomy(best)+left)
	}
}

// ResumRegionEconomy ＝ `2703:02E5` 的第一件事：每一格的經濟力
// `+0x1A` ← 該格行星經濟力相加（`2703:034E`）。`090A:0E6D` 每跑一次都做，
// 所以花進行星的錢下一次重算就反映在星域與徵稅額上（收據 `fcmd-exec`：
// 首都 82302 → 82339、徵稅額 2633 → 2634）。
func ResumRegionEconomy(records []byte, p *Planets) {
	if p == nil {
		return
	}
	n := len(records) / RegionRecord
	for r := 0; r < n; r++ {
		start, cnt := p.Range(r)
		sum := 0
		for i := start; i < start+cnt; i++ {
			sum += p.PlanetEconomy(i)
		}
		SetRegionEconomy(records[r*RegionRecord:(r+1)*RegionRecord], sum)
	}
}
