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
// 防衛力低於門檻（3000）時再 ×3/2——這一條是從程式讀出來的，
// 還沒實測到（`docs/re/33`）。
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

// MonthEndRegion 跑一個星域的月底結算：逐行星成長、逐行星扣稅，
// 回這個星域收到的稅與結算後的星域經濟力（`docs/re/33`）。
//
//	對每一顆行星：
//	    經濟力 += 經濟力 × 成長率(經濟力, 防衛力) / 1000
//	    稅      = 經濟力 × 星域稅率 / 1000      （捨去）
//	    經濟力 -= 稅
//	星域經濟力 = Σ 該星域行星的經濟力
//
// **捨去發生在每一顆行星上**，不是在星域的總和上——差別看得出來：
// 星域 0 逐行星是 2718，用總和算是 2720。
func (p *Planets) MonthEndRegion(t *GrowthTable, region, taxPerMille int) (tax, economy int) {
	start, n := p.Range(region)
	for i := start; i < start+n; i++ {
		e := p.PlanetEconomy(i)
		e += e * t.Rate(e, p.Defence(i)) / 1000
		levy := e * taxPerMille / 1000
		e -= levy
		p.SetPlanetEconomy(i, e)
		tax += levy
		economy += e
	}
	return tax, economy
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
	total := 0
	for i := 0; i*RegionRecord+RegionRecord <= len(records); i++ {
		rec := records[i*RegionRecord : (i+1)*RegionRecord]
		tax, economy := p.MonthEndRegion(t, i, RegionTaxRate(rec))
		SetRegionEconomy(rec, economy)
		rec[0x27] = rec[regionTaxRate]
		if p.RegionOwner(i) == faction {
			total += tax
		}
	}
	return total
}
