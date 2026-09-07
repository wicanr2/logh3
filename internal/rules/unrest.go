package rules

// 未安定度：載入一個局面時整份重算，算式是 `2703:0DA1`（`docs/re/43`）。
//
// 算出來的其實是**治安度**（0–200），記錄裡的 `+0x7D` 是它的補數：
// `未安定度 = 200 − 治安度`。畫面上的治安百分比 ＝ 治安度 / 2。

const (
	planetFlag12 = 0x12 // f4 數的就是這一格不為 0 的行星
	regionFlag13 = 0x13 // bit 2 → −30；f3 看鄰居的 0x14
	regionMorale = 0x7C // 民心：帝國星域取原值、同盟取 200 −
	regionTaxPre = 0x27 // f1 用的稅率（與 +0x26 多半相同）

	fleetFactionOff = 0x01
	fleetUnusedOff  = 0x04
	fleetRegionOff  = 0x06
	fleetAdmiralOff = 0x04
	fleetStateOff   = 0x08
	fleetStateHome  = 5 // 只有這個狀態的艦隊算進治安
	fleetWeightOff  = 0x0F

	admiralFaction = 0x01
	admiralAbilP   = 0x0B
	admiralAbilQ   = 0x13
	admiralRule    = 0x0E // 統治者的這一格拿來乘
)

// UnrestParams 是算式要的執行期參數。都不在原版資料檔裡，只能從記憶體量
// （`docs/internals/data/unrest-at-load.json` 的 `params`）。
type UnrestParams struct {
	// Turn 是 `ds:8F18`；算式只用它除以 10 的餘數。
	Turn int

	// Ruler 是各陣營統治者的提督編號（`ds:8F4F`）。
	Ruler [2]int

	// Blend／Rate 以陣營索引（`ds:8FD4`／`ds:8FF8`），都是查 Table 的下標；
	// Rate 同時是查 Base 的下標。
	Blend, Rate [2]int

	// Table 是 `ds:4B28`，Base 是 `ds:147E`（32 位元陣列的低位字）。
	Table, Base []int
}

// UnrestState 把算一個星域的未安定度要的東西湊在一起。
type UnrestState struct {
	Regions  []byte // 星域表，一筆 RegionRecord
	Planets  *Planets
	Fleets   []byte // 艦隊表，一筆 FleetRecord
	Admirals []byte // 提督表，一筆 AdmiralRecord
	Graph    *StarGraph
	Params   UnrestParams
}

// AdmiralRecord 是提督表一筆的長度。
const AdmiralRecord = 28

// idiv 照 x86 的 `idiv` 往零截斷。Go 的整數除法本來就是往零截斷，
// 這支只是把「這裡不能用往下取整」講清楚——差 1 會讓 f1 差 2。
func idiv(a, b int) int { return a / b }

// Unrest 回一個星域載入時算出來的未安定度。
func (s *UnrestState) Unrest(region int) int {
	r := s.region(region)
	if r == nil {
		return 0
	}
	owner := int(r[regionOwner])
	if owner == 2 { // 無主一律 100，其餘一概不算
		return 100
	}
	v := s.f1(r, owner) - s.f2(r, owner) - s.f3(r, region, owner) + s.f4(region)
	if r[regionFlag13]&4 != 0 {
		v -= 30
	}
	if v < 0 {
		v = 0
	}
	v = idiv(v*int(s.admiral(s.Params.Ruler[owner])[admiralRule]), 100)
	if v > 200 {
		v = 200
	}
	// f5（提督那一項）恆等於 0：它要 `+0x09` 最高位有設，而那一格
	// 載入時會被夾到 127（`docs/re/43`、`docs/internals/08`）。
	v += s.f6(region, owner)
	if v > 200 {
		v = 200
	}
	// 尾巴：離自己首都幾步，一步扣 2。原版另有一條「路徑成本 >= 0x800
	// 就砍半」的分支，但那張成本表載入時的值都遠小於 0x800。
	if d := s.Graph.Hops(CapitalRegion(owner), region); d > 0 {
		v -= 2 * d
	}
	if v < 0 {
		v = 0
	}
	return 200 - v
}

// f1 是稅率、成長率與兩組難度參數揉出來的基底。
func (s *UnrestState) f1(r []byte, owner int) int {
	p := s.Params
	a, b := p.Blend[owner], p.Rate[owner]
	di := p.Base[b]
	growth := int(r[regionGrowth])
	v0 := di
	if growth > 0 {
		v0 = int(r[regionTaxPre]) * 100 / growth
	}
	if v0 <= 25 {
		v0 = di
	}
	d := 100 + idiv(v0*100/di-100, 4)
	m := p.Turn % 10
	blend := ((m+5)*p.Table[a] + (14-m)*p.Table[b]) / 19
	return idiv(blend*100, d)
}

// f2 是民心。帝國星域取原值、同盟取 200 −，再除以 10。
func (s *UnrestState) f2(r []byte, owner int) int {
	v := int(r[regionMorale])
	if owner != 0 {
		v = 200 - v
	}
	return v / 10
}

// f3 是鄰居：敵對的扣得多，而且鄰居的 `+0x13 & 0x14` 再加一份。
func (s *UnrestState) f3(r []byte, region, owner int) int {
	out := 0
	for _, j := range s.Graph.Neighbours(region) {
		n := s.region(j)
		if n == nil {
			continue
		}
		if int(n[regionOwner]) == owner {
			if n[regionFlag13]&0x14 != 0 {
				out += 7
			}
			continue
		}
		if int(n[regionOwner]) == 2 {
			out += 3
		} else {
			out += 5
		}
		if n[regionFlag13]&0x14 != 0 {
			out += 2
		}
	}
	return out
}

// f4 是這個星域裡 `+0x12` 不為 0 的行星數 × 10。
func (s *UnrestState) f4(region int) int {
	if s.Planets == nil {
		return 0
	}
	start, n := s.Planets.Range(region)
	count := 0
	for i := start; i < start+n; i++ {
		if s.Planets.rec(i)[planetFlag12] != 0 {
			count++
		}
	}
	return count * 10
}

// f6 是駐在這個星域、狀態為 5 的同陣營艦隊：`Σ+0x0F × maxP × maxQ / 10000`，
// 其中 maxP／maxQ 取各艦隊指揮官（艦隊 `+0x04`）的 `+0x0B`／`+0x13` 最大值。
func (s *UnrestState) f6(region, owner int) int {
	sum, p, q := 0, 0, 0
	for i := 0; (i+1)*FleetRecord <= len(s.Fleets); i++ {
		f := s.Fleets[i*FleetRecord : (i+1)*FleetRecord]
		if f[fleetUnusedOff] == 0xFF || int(f[fleetRegionOff]) != region ||
			int(f[fleetFactionOff]) != owner || f[fleetStateOff] != fleetStateHome {
			continue
		}
		sum += int(f[fleetWeightOff])
		a := s.admiral(int(f[fleetAdmiralOff]))
		if a == nil {
			continue
		}
		if int(a[admiralAbilP]) > p {
			p = int(a[admiralAbilP])
		}
		if int(a[admiralAbilQ]) > q {
			q = int(a[admiralAbilQ])
		}
	}
	return sum * p * q / 10000
}

func (s *UnrestState) region(i int) []byte {
	if i < 0 || (i+1)*RegionRecord > len(s.Regions) {
		return nil
	}
	return s.Regions[i*RegionRecord : (i+1)*RegionRecord]
}

func (s *UnrestState) admiral(i int) []byte {
	if i < 0 || (i+1)*AdmiralRecord > len(s.Admirals) {
		return nil
	}
	return s.Admirals[i*AdmiralRecord : (i+1)*AdmiralRecord]
}
