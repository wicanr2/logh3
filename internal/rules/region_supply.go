package rules

import "encoding/binary"

const (
	regionFlags       = 0x13
	regionMissiles    = 0x68 // 32 位元 × 陣營
	regionFighters    = 0x70 // 32 位元 × 陣營
	regionSupplyLimit = 0x78
)

// RegionSupplyPool 是星域層的兩種戰術補給存量。
type RegionSupplyPool struct {
	Missiles [2]int
	Fighters [2]int
}

// ReadRegionSupply 讀星域記錄的 `+0x68`／`+0x70`。
func ReadRegionSupply(record []byte) (RegionSupplyPool, bool) {
	var p RegionSupplyPool
	if len(record) < regionSupplyLimit {
		return p, false
	}
	for f := 0; f < 2; f++ {
		p.Missiles[f] = int(binary.LittleEndian.Uint32(record[regionMissiles+f*4:]))
		p.Fighters[f] = int(binary.LittleEndian.Uint32(record[regionFighters+f*4:]))
	}
	return p, true
}

func writeRegionSupply(record []byte, p RegionSupplyPool) bool {
	if len(record) < regionSupplyLimit {
		return false
	}
	for f := 0; f < 2; f++ {
		binary.LittleEndian.PutUint32(record[regionMissiles+f*4:], uint32(p.Missiles[f]))
		binary.LittleEndian.PutUint32(record[regionFighters+f*4:], uint32(p.Fighters[f]))
	}
	return true
}

// AggregateRegionSupply 照 `2703:12EB` 從行星彙總兩陣營的飛彈與艦載機。
func AggregateRegionSupply(record []byte, planets *Planets, region int) bool {
	if planets == nil || len(record) < RegionRecord {
		return false
	}
	var pool RegionSupplyPool
	start, n := planets.Range(region)
	for i := start; i < start+n; i++ {
		owner := planets.Owner(i)
		if owner < 0 || owner > 1 {
			continue
		}
		stock, ok := ReadPlanetStock(planets.Record(i))
		if !ok {
			return false
		}
		pool.Missiles[owner] += stock.Missiles
		pool.Fighters[owner] += stock.Fighters
	}
	return writeRegionSupply(record, pool)
}

// RefillRegionSupply 照 `2B5F:02B2` 回充星域擁有者的兩種存量。
func RefillRegionSupply(record []byte) bool {
	pool, ok := ReadRegionSupply(record)
	if !ok || len(record) < RegionRecord {
		return false
	}
	owner := RegionOwnerOf(record)
	if owner < 0 || owner > 1 || record[regionFlags]&1 != 0 {
		return true
	}
	limit := RegionEconomy(record) + int(binary.LittleEndian.Uint16(record[regionDefence:]))
	add := limit / 10
	capAdd := func(v int) int {
		if v >= limit-add {
			return limit
		}
		return v + add
	}
	pool.Missiles[owner] = capAdd(pool.Missiles[owner])
	pool.Fighters[owner] = capAdd(pool.Fighters[owner])
	return writeRegionSupply(record, pool)
}

// DistributeRegionSupply 照 `2703:0590` 把某陣營的星域層存量分回行星：
// 七種艦艇的待命艘數（`+0x30 ＋ 陣營 × 0x1C`）進行星的補充存量 `+0x20`，
// 飛彈／艦載機（`+0x68`／`+0x70`）進 `+0x3C`／`+0x40`。
// 每顆先取整數百分比；逐顆捨去後的餘額全給最大經濟力行星，同值後者勝
// （`docs/re/273`；艦艇那七輪在 `+06E0`，餘額在 `+0800`）。
func DistributeRegionSupply(record []byte, planets *Planets, region, faction int) bool {
	pool, ok := ReadRegionSupply(record)
	if !ok || planets == nil || faction < 0 || faction > 1 {
		return false
	}
	ships := RegionShips(record, faction)
	start, n := planets.Range(region)
	total, maxEconomy, maxPlanet := 0, 0, -1
	for i := start; i < start+n; i++ {
		if planets.Owner(i) != faction {
			continue
		}
		e := planets.PlanetEconomy(i)
		total += e
		if e >= maxEconomy {
			maxEconomy, maxPlanet = e, i
		}
	}
	if total == 0 || maxPlanet < 0 {
		return true
	}
	leftM, leftF := pool.Missiles[faction], pool.Fighters[faction]
	var leftS [7]int
	copy(leftS[:], ships)
	for i := start; i < start+n; i++ {
		if planets.Owner(i) != faction {
			continue
		}
		pct := planets.PlanetEconomy(i) * 100 / total
		m := pool.Missiles[faction] * pct / 100
		f := pool.Fighters[faction] * pct / 100
		stock, ok := ReadPlanetStock(planets.Record(i))
		if !ok {
			return false
		}
		for k := range stock.Ships {
			v := 0
			if k < len(ships) {
				v = ships[k] * pct / 100
			}
			stock.Ships[k] = v
			leftS[k] -= v
		}
		stock.Missiles, stock.Fighters = m, f
		WritePlanetStock(planets.Record(i), stock)
		leftM, leftF = leftM-m, leftF-f
	}
	stock, ok := ReadPlanetStock(planets.Record(maxPlanet))
	if !ok {
		return false
	}
	for k := range stock.Ships {
		stock.Ships[k] += leftS[k]
	}
	stock.Missiles += leftM
	stock.Fighters += leftF
	return WritePlanetStock(planets.Record(maxPlanet), stock)
}
