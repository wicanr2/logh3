package rules

import "encoding/binary"

// RecomputeRegionPlanets 依 `2703:15D5` 重算行星數、防衛力及經濟比例。
// 比例使用這一次歸屬重算時的經濟力；單純重加總經濟力不刷新它。
// 證據及開局接線契約見研究 352／規格 065。
func RecomputeRegionPlanets(records []byte, p *Planets) {
	if p == nil {
		return
	}
	for r := 0; r < len(records)/RegionRecord && r < len(p.PerRegion); r++ {
		rec := records[r*RegionRecord : (r+1)*RegionRecord]
		start, n := p.Range(r)
		var counts [3]byte
		var economy [2]int
		total, defence, limit := 0, 0, 0
		for i := start; i < start+n; i++ {
			e := p.PlanetEconomy(i)
			total += e
			o := p.Owner(i)
			if o == 0 || o == 1 {
				counts[o]++
				economy[o] += e
			} else {
				counts[2]++
			}
			defence += p.Defence(i)
			limit += p.GrowthDefence(i)
		}
		copy(rec[0x16:0x19], counts[:])
		binary.LittleEndian.PutUint16(rec[0x20:], uint16(defence))
		binary.LittleEndian.PutUint16(rec[0x22:], uint16(limit))
		// 原始開局每格總額皆 > 0；空的測試資料不執行除法。
		if total > 0 {
			for f := 0; f < 2; f++ {
				rec[0x14+f] = byte(economy[f] * 100 / total)
			}
			for i := start; i < start+n; i++ {
				p.Record(i)[0x13] = byte(p.PlanetEconomy(i) * 100 / total)
			}
		}
	}
}

// DoubleRegionPlanetShips ＝ `2703:04A3`：首都全部行星的七種艦艇乘二。
// 只由開局電腦座位呼叫；32 位溢位與原版兩個 word 左移一致。
func DoubleRegionPlanetShips(p *Planets, region int) {
	if p == nil {
		return
	}
	start, n := p.Range(region)
	for i := start; i < start+n; i++ {
		rec := p.Record(i)
		for k := 0; k < 7; k++ {
			o := 0x20 + k*4
			binary.LittleEndian.PutUint32(rec[o:], binary.LittleEndian.Uint32(rec[o:])<<1)
		}
	}
}

// AggregateRegionShips ＝ `2703:12EB` 的七種艦艇彙總，按行星擁有者分池。
func AggregateRegionShips(record []byte, p *Planets, region int) {
	if p == nil || len(record) < RegionRecord {
		return
	}
	var ships [2][7]uint32
	start, n := p.Range(region)
	for i := start; i < start+n; i++ {
		f := p.Owner(i)
		if f != 0 && f != 1 {
			continue
		}
		for k := 0; k < 7; k++ {
			ships[f][k] += binary.LittleEndian.Uint32(p.Record(i)[0x20+k*4:])
		}
	}
	for f := 0; f < 2; f++ {
		for k := 0; k < 7; k++ {
			binary.LittleEndian.PutUint32(record[0x30+f*28+k*4:], ships[f][k])
		}
	}
}
