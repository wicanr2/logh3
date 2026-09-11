package rules

// 未安定度：算式是 `2703:0DA1`（`docs/re/43`、`docs/spec/006`）。載入一個局面
// 時整份重算，遊戲進行中每一次 `090A:0E6D`（稅率或參謀一變、每次「實行」的
// 扣維持費、每個座位回合開頭）與回合尾巴的 `2703:08DE` 也都整份重算
// （`docs/spec/042`）。
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

	// FleetList 是 remake 自己的艦隊表。有給就取代 Fleets（f6 只讀陣營、
	// 所在星域、`+0x08 == 5`、`+0x0F` 與指揮官，`Fleet` 都有）。
	FleetList []Fleet

	// Paths 是尾巴那一段讀的路徑表。nil ＝ **載入當下**那一種：距離一律
	// 小於 `0x800`、步數用相鄰圖的最短步數（`docs/spec/006` 量到的）。
	// 遊戲進行中原版讀的是 `2B99:0003` 照擁有者建好的那兩張表
	// （`docs/spec/042`），用 `StarGraph.UnrestPaths` 建。
	Paths *UnrestPaths
}

// UnrestPaths 是未安定度尾巴那一段讀的兩組路徑表，以陣營索引：
// `Dist` 給 `2B99:028F`（距離矩陣 `ds:D062`／`ds:D066`），`Mid` 給
// `2B99:0450` 數步數（中點矩陣 `ds:D06A`／`ds:D06E`）。
type UnrestPaths struct {
	Dist [2][][]int
	Mid  [2][][]int
}

// UnrestPaths 照 `2B99:0003` 的做法（`2B99:007D` ＋ `2B99:01B1`，兩個陣營
// 各一次）用現在的擁有者建路徑表。`owner` 是 32 格的 `+0x12`。
func (g *StarGraph) UnrestPaths(owner []int) *UnrestPaths {
	p := &UnrestPaths{}
	for f := 0; f < 2; f++ {
		p.Dist[f], p.Mid[f] = g.BuildTables(owner, f)
	}
	return p
}

// unrestHalveCost 是 `2703:0E9B` 的 `cmp ax, 800h / jb`：到首都的路徑成本
// （無號）不到它就一步扣 2，到了就把治安度砍半。
const unrestHalveCost = 0x800

// RouteSteps 是 `2B99:0450(星域, 首都, 陣營)`：從 `from` 一格一格往 `to`
// 走（每一格都是 `2B99:0302` 的下一格），數走了幾步。`mid` 是那個陣營的
// 中點矩陣（`255` ＝ 直接相連）。走不到回 -1。
func RouteSteps(mid [][]int, from, to int) int {
	if from < 0 || from >= len(mid) || to < 0 || to >= len(mid) {
		return -1
	}
	n, cur := 0, from
	for cur != to {
		// 2B99:0302 → 2B99:0349：a ← to；a ← mid[cur][a]，直到 255 回 a。
		a := to
		for k := 0; ; k++ {
			if k > 64 || a < 0 || a >= len(mid[cur]) {
				return -1
			}
			v := mid[cur][a]
			if v == starDirect {
				break
			}
			a = v
		}
		cur = a
		n++
		if n > 64 {
			return -1
		}
	}
	return n
}

// AdmiralRecord 是提督表一筆的長度。
const AdmiralRecord = 28

// idiv 照 x86 的 `idiv` 往零截斷。Go 的整數除法本來就是往零截斷，
// 這支只是把「這裡不能用往下取整」講清楚——差 1 會讓 f1 差 2。
func idiv(a, b int) int { return a / b }

// Unrest 回一個星域的未安定度（`2703:0DA1` 寫進 `+0x7D` 的那個值）。
func (s *UnrestState) Unrest(region int) int {
	r := s.region(region)
	if r == nil {
		return 0
	}
	owner := int(r[regionOwner])
	if owner == 2 { // 無主一律 100，其餘一概不算
		return 100
	}
	if owner < 0 || owner >= len(s.Params.Ruler) {
		return 0
	}
	v := s.f1(r, owner) - s.f2(r, owner) - s.f3(r, region, owner) + s.f4(region)
	if r[regionFlag13]&4 != 0 {
		v -= 30
	}
	if v < 0 {
		v = 0
	}
	// 缺提督表就算不出來——回 0。原本這裡直接對 admiral() 的結果取索引，
	// 而它在查不到的時候回 nil，於是「資料不全」變成 panic；那是玩家走得到
	// 的路（戰鬥中選「取進度」就會經過這裡）。
	ruler := s.admiral(s.Params.Ruler[owner])
	if len(ruler) <= admiralRule {
		return 0
	}
	v = rulerScale(v, int(ruler[admiralRule]))
	if v > 200 {
		v = 200
	}
	// f5（提督那一項，`2703:10FC`）恆等於 0：它把 `n × p × q ÷ 1000` 的商
	// 放在 bx，卻把清成 0 的那一格推給 `3310:0130`（整數開根號），
	// 開根號 0 回 0（`docs/spec/042`）。與提督的旗標無關。
	v += s.f6(region, owner)
	if v > 200 {
		v = 200
	}
	v = s.tail(v, region, owner)
	if v < 0 {
		v = 0
	}
	return 200 - v
}

// tail 是 `2703:0E7F` 起那一段：到自己首都的路徑成本 >= 0x800 就把治安度
// 砍半，否則一步扣 2。
func (s *UnrestState) tail(v, region, owner int) int {
	capital := CapitalRegion(owner)
	if s.Paths == nil {
		// 載入當下（`docs/spec/006`）：量到的成本都小於 0x800、步數與
		// 相鄰圖的最短步數相同。
		if d := s.Graph.Hops(capital, region); d > 0 {
			v -= 2 * d
		}
		return v
	}
	dist, mid := s.Paths.Dist[owner], s.Paths.Mid[owner]
	if capital >= len(dist) || region >= len(dist[capital]) {
		return v
	}
	if uint16(dist[capital][region]) >= unrestHalveCost {
		return idiv(v, 2)
	}
	if d := RouteSteps(mid, capital, region); d > 0 {
		v -= 2 * d
	}
	return v
}

// rulerScale 是 `2703:0E33`–`0E47`：`v × 統治者 +0x0E ÷ 100`。乘法是
// `mul cx`（16 × 16 → dx:ax），接著 `cwd / idiv 100`——**高位的 dx 被 cwd
// 蓋掉**，只留低 16 位元、再當有號數除（`docs/spec/042`）。乘積到 0x8000 以上
// 就變成負的；呼叫端的 `> 200` 夾限是有號比較，擋不住它——加上 f6 之後
// 仍是負的，就由尾巴的 `< 0 → 0` 收成治安度 0（未安定度 200）。
func rulerScale(v, ops int) int {
	return idiv(int(int16(uint16(v*ops))), 100)
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
	if s.Graph == nil {
		return 0
	}
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
	take := func(weight, admiral int) {
		sum += weight
		a := s.admiral(admiral)
		if a == nil {
			return
		}
		if int(a[admiralAbilP]) > p {
			p = int(a[admiralAbilP])
		}
		if int(a[admiralAbilQ]) > q {
			q = int(a[admiralAbilQ])
		}
	}
	if s.FleetList != nil {
		for _, f := range s.FleetList {
			if f.Region != region || f.Faction != owner || !f.Parked {
				continue
			}
			take(f.Groups, f.Admiral)
		}
		return sum * p * q / 10000
	}
	for i := 0; (i+1)*FleetRecord <= len(s.Fleets); i++ {
		f := s.Fleets[i*FleetRecord : (i+1)*FleetRecord]
		if f[fleetUnusedOff] == 0xFF || int(f[fleetRegionOff]) != region ||
			int(f[fleetFactionOff]) != owner || f[fleetStateOff] != fleetStateHome {
			continue
		}
		take(int(f[fleetWeightOff]), int(f[fleetAdmiralOff]))
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
