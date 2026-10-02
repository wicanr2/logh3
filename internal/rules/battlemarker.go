package rules

import "encoding/binary"

// UpdateBattleMarker對應1727:02E2；完整狀態及快照依規格082。
func UpdateBattleMarker(u *BattleUnit, forced, enabled bool) {
	if u == nil {
		return
	}
	if !forced && u.Detail == 0 {
		if enabled && u.MarkerFlags&0xC0 == 0xC0 {
			u.MarkerFlags &^= 0x40
		}
		u.MarkerFlags &^= 0x80
	} else {
		u.MarkerFlags |= 0x80
		if enabled && u.MarkerFlags&0x40 == 0 {
			ink := 4
			if u.Faction == 0 {
				ink = 5
			} else if u.Faction == 1 {
				ink = 9
			}
			u.MarkerFlags = 0xC0 | ink
			u.MarkerX, u.MarkerY = u.X, u.Y
		}
	}
	u.Marked = u.MarkerFlags&0x80 != 0
}

// ApplyBattleMarker只回寫標記五個bytes，不覆蓋資源、行動或移動額度。
func (s *Slot) ApplyBattleMarker(pool int, u BattleUnit) {
	r := s.ShipAt(pool)
	if len(r) != ShipRecord {
		return
	}
	r[0x19] = byte(u.MarkerFlags)
	binary.LittleEndian.PutUint16(r[0x1A:], uint16(u.MarkerX))
	binary.LittleEndian.PutUint16(r[0x1C:], uint16(u.MarkerY))
}
