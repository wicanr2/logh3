package rules

import (
	"encoding/binary"
	"fmt"
)

// These are user-selected remake rules (spec146), not original-game parity.
func FlagshipTransferTarget(units []BattleUnit, destroyed int) int {
	return FlagshipTransferTargetWithPolicy(units, destroyed, false)
}

// FlagshipKindAllowed follows user-selected spec150. T/E and non-ship kinds
// remain ineligible even when non-battleship flagships are enabled.
func FlagshipKindAllowed(kind int, nonBattleship bool) bool {
	return kind >= 0 && (kind <= 1 || nonBattleship && kind <= 4)
}

func FlagshipTransferTargetWithPolicy(units []BattleUnit, destroyed int, nonBattleship bool) int {
	if destroyed < 0 || destroyed >= len(units) {
		return -1
	}
	flag := units[destroyed]
	if flag.Squadron != 0 || flag.Ships != 0 || flag.Out {
		return -1
	}
	best := -1
	for i, u := range units {
		if u.Fleet != flag.Fleet || u.Faction != flag.Faction || u.Squadron <= 0 || u.Out || u.Ships <= 0 || !FlagshipKindAllowed(u.Kind, nonBattleship) {
			continue
		}
		if best < 0 || u.Ships > units[best].Ships || u.Ships == units[best].Ships && u.Squadron < units[best].Squadron {
			best = i
		}
	}
	return best
}

func FlagshipTransferSucceeds(admiral, roll int) bool {
	return admiral == 5 || roll >= 0 && roll%100 < 25
}

// TransferFlagship swaps pool references without moving ships or reviving losses.
// All validation precedes mutation; the caller supplies current damage first.
func (s *Slot) TransferFlagship(index, squadron int) error {
	return s.TransferFlagshipWithPolicy(index, squadron, false)
}

func (s *Slot) TransferFlagshipWithPolicy(index, squadron int, nonBattleship bool) error {
	if s == nil || len(s.Fleet) != FleetCount*FleetRecord || len(s.Ship) != ShipCount*ShipRecord || index < 0 || index >= FleetCount {
		return fmt.Errorf("旗艦轉移缺現行艦隊／艦艇表")
	}
	r := s.fleet(index)
	n := int(r[fleetGroups])
	if r[fleetUnused] == 255 || n < 2 || n > fleetSlotN || squadron <= 0 || squadron >= n {
		return fmt.Errorf("旗艦轉移的艦隊或候選分隊無效")
	}
	var seen [ShipCount]bool
	for i := 0; i < n; i++ {
		pool := int(binary.LittleEndian.Uint16(r[fleetSlots+4*i:]))
		if pool >= ShipCount || seen[pool] || int(s.Ship[pool*ShipRecord+shipOwner]) != index {
			return fmt.Errorf("旗艦轉移的池引用越界、重複或擁有者不符")
		}
		seen[pool] = true
	}
	a, b := fleetSlots, fleetSlots+4*squadron
	oldPool, newPool := int(binary.LittleEndian.Uint16(r[a:])), int(binary.LittleEndian.Uint16(r[b:]))
	old, next := s.ShipAt(oldPool), s.ShipAt(newPool)
	if binary.LittleEndian.Uint16(old[shipAlive:]) != 0 || binary.LittleEndian.Uint16(next[shipAlive:]) == 0 || !FlagshipKindAllowed(int(next[shipKind]), nonBattleship) || next[0x1e]&0x80 != 0 {
		return fmt.Errorf("旗艦轉移的死亡狀態或候選艦種無效")
	}
	var group [4]byte
	newSlot := r[b+3]
	copy(group[:], r[a:a+4])
	copy(r[a:a+4], r[b:b+4])
	copy(r[b:b+4], group[:])
	r[a+3], r[b+3] = byte(FlagshipFormationSlot), newSlot
	old[3], old[5] = byte(squadron), r[b+3]
	next[3], next[5] = 0, byte(FlagshipFormationSlot)
	copy(r[fleetPosX:fleetPosX+4], next[shipPosX:shipPosX+4])
	return nil
}
