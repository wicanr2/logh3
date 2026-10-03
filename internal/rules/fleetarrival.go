package rules

import "encoding/binary"

// PlaceFleetArrival接回113A:0648→297B:08D8，限完整原池與有效入口（規格088）。
// 地形、其他艦隊與槽表沿既有戰略定位規則；現在位置與集結點分開保存。
func (s *Slot) PlaceFleetArrival(f *Fleet, from int, entries *BattleEntries, place *StrategicPlacement) bool {
	if s == nil || f == nil || place == nil || place.Slots == nil ||
		len(s.Fleet) != FleetCount*FleetRecord || len(s.Ship) != ShipCount*ShipRecord ||
		f.Index < 0 || f.Index >= FleetCount {
		return false
	}
	e, ok := entries.EntryFor(f.Region, from)
	if !ok {
		return false // 缺入口的原版fallback未取得動態驗收，不猜補。
	}
	for _, g := range f.Composition {
		if g.Pool < 0 || g.Pool >= ShipCount {
			return false
		}
	}
	facing := placeStrategicFleet(f, PlanetSite{X: e.X, Y: e.Y}, nil, place)
	r := s.fleet(f.Index)
	r[0x3A], r[0x3B], r[0x48], r[0x5A], r[0x5B] = 255, 0, 4, 255, 255
	f.Action = 4
	r[0x4A], r[0x4B], r[0x4C] = byte(facing), byte(facing), byte(facing)
	for _, off := range []int{0x5C, 0x64} {
		binary.LittleEndian.PutUint16(r[off:], uint16(f.X))
		binary.LittleEndian.PutUint16(r[off+2:], uint16(f.Y))
	}
	for i := range f.Composition {
		g := &f.Composition[i]
		sh := s.Ship[g.Pool*ShipRecord : (g.Pool+1)*ShipRecord]
		tx, ty := place.Slots.Station(f.X, f.Y, facing, g.Slot3)
		binary.LittleEndian.PutUint16(sh[0x28:], uint16(tx))
		binary.LittleEndian.PutUint16(sh[0x2A:], uint16(ty))
		binary.LittleEndian.PutUint16(sh[0x34:], uint16(g.X))
		binary.LittleEndian.PutUint16(sh[0x36:], uint16(g.Y))
		sh[0x16], sh[0x17], sh[0x19], sh[0x1E] = 0, 0, 0x3F, 0x7F
		g.Detail = 0
	}
	if f.Scout >= 0 && f.Scout < ShipCount {
		sh := s.Ship[f.Scout*ShipRecord : (f.Scout+1)*ShipRecord]
		for _, off := range []int{0x28, 0x34} {
			binary.LittleEndian.PutUint16(sh[off:], uint16(f.X))
			binary.LittleEndian.PutUint16(sh[off+2:], uint16(f.Y))
		}
		sh[0x16], sh[0x17], sh[0x19], sh[0x1E] = 0, 0, 0x3F, 255
	}
	return true
}
