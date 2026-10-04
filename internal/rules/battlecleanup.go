package rules

import (
	"encoding/binary"
	"fmt"
)

// PruneEmptyBattleGroups依109 READY移除空的一般組，保留未用尾槽與SHIP+03。
// 回傳舊組至新組的映射；-1為作廢。先完整驗證，失敗不改原表。
func (s *Slot) PruneEmptyBattleGroups(index int) ([]int, error) {
	if s == nil || len(s.Fleet) != FleetCount*FleetRecord || len(s.Ship) != ShipCount*ShipRecord || index < 0 || index >= FleetCount {
		return nil, fmt.Errorf("戰場清理的艦隊／艦艇表或索引無效")
	}
	r := s.fleet(index)
	n := int(r[fleetGroups])
	if r[fleetUnused] == 255 || n < 1 || n > fleetSlotN {
		return nil, fmt.Errorf("戰場清理的艦隊無效或組數越界")
	}
	var seen [ShipCount]bool
	mapping := make([]int, n)
	kept := 0
	for i := 0; i < n; i++ {
		pool := int(binary.LittleEndian.Uint16(r[fleetSlots+4*i:]))
		if pool >= ShipCount || seen[pool] || int(s.Ship[pool*ShipRecord+shipOwner]) != index {
			return nil, fmt.Errorf("戰場清理的池引用越界、重複或擁有者不符")
		}
		seen[pool] = true
		if binary.LittleEndian.Uint16(s.Ship[pool*ShipRecord+shipAlive:]) == 0 {
			if i == 0 {
				return nil, fmt.Errorf("空旗艦應走全滅分支")
			}
			mapping[i] = -1
		} else {
			mapping[i] = kept
			kept++
		}
	}
	for i := 0; i < int(r[fleetGroups]); {
		o := fleetSlots + 4*i
		pool := int(binary.LittleEndian.Uint16(r[o:]))
		sh := s.ShipAt(pool)
		if binary.LittleEndian.Uint16(sh[shipAlive:]) != 0 {
			i++
			continue
		}
		sh[shipOwner] = 255
		sh[0x19] &^= 0x40
		clear(sh[0x38:0x3c])
		r[fleetGroups]--
		end := fleetSlots + 4*int(r[fleetGroups])
		copy(r[o:end], r[o+4:end+4])
	}
	return mapping, nil
}
