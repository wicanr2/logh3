package rules

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// SyncFleets 以現況編制更新原始存檔表（規格 067）。未知位元組保留；
// 索引配置與更新都在副本完成，失敗不改動輸入。回傳的編制帶已配置池索引。
func (s *Slot) SyncFleets(fleets []Fleet, kinds *ShipKindTable, mobility []int) ([]Fleet, error) {
	if len(s.Fleet) != FleetCount*FleetRecord || len(s.Ship) != ShipCount*ShipRecord {
		return nil, fmt.Errorf("艦隊／艦艇表長度不符原版格式")
	}
	fb, sb := bytes.Clone(s.Fleet), bytes.Clone(s.Ship)
	out := append([]Fleet(nil), fleets...)
	var live [FleetCount]bool
	var claimed [ShipCount]bool
	var oldReferenced [ShipCount]bool
	for i := 0; i < FleetCount; i++ {
		if !s.FleetUsed(i) {
			continue
		}
		for _, p := range s.FleetGroupShips(i) {
			if p >= 0 && p < ShipCount {
				oldReferenced[p] = true
			}
		}
		r := s.fleet(i)
		p := int(binary.LittleEndian.Uint16(r[fleetScout:]))
		if p < ShipCount {
			oldReferenced[p] = true
		}
	}
	claim := func(index, owner, kind int) error {
		if index < 0 || index >= ShipCount || claimed[index] {
			return fmt.Errorf("艦艇池索引 %d 越界或重複", index)
		}
		r := sb[index*ShipRecord : (index+1)*ShipRecord]
		if int(r[shipOwner]) != owner || int(r[shipKind]) != kind {
			return fmt.Errorf("艦艇池索引 %d 的擁有者／艦種不符現況", index)
		}
		claimed[index] = true
		return nil
	}
	for i := range out {
		f := &out[i]
		if f.Index < 0 || f.Index >= FleetCount || live[f.Index] || f.Faction < 0 || f.Faction > 1 ||
			f.Admiral < 0 || f.Admiral >= 0xFF || len(f.Composition) > fleetSlotN {
			return nil, fmt.Errorf("艦隊索引、陣營、提督或編制不合法：%d", f.Index)
		}
		live[f.Index] = true
		f.Composition = append([]FleetGroup(nil), f.Composition...)
		fresh := fb[f.Index*FleetRecord+fleetUnused] == 0xFF
		for _, gp := range f.Composition {
			if gp.Kind < 0 || gp.Kind >= 8 || gp.Pool < -1 || gp.Ships < 0 || gp.Ships > 65535 || gp.Full < 0 || gp.Full > 65535 {
				return nil, fmt.Errorf("艦隊 %d 的艦艇欄位超出格式範圍", f.Index)
			}
			if gp.Pool >= 0 {
				if err := claim(gp.Pool, f.Index, gp.Kind); err != nil {
					return nil, err
				}
			}
		}
		if fresh {
			f.Scout = -1
		} // 新艦隊尚無原始表；不能把 int 零值誤認為池 #0。
		if f.Scout >= 0 && f.Scout != 0xFFFF {
			if err := claim(f.Scout, f.Index, shipKindStandalone); err != nil {
				return nil, err
			}
		}
	}
	// 被拆掉的組與艦隊作廢；保留其餘 bytes，與 2AD2:01EF 的契約相同。
	for i := 0; i < ShipCount; i++ {
		r := sb[i*ShipRecord : (i+1)*ShipRecord]
		if oldReferenced[i] && r[shipOwner] != 0xFF && !claimed[i] {
			r[shipOwner] = 0xFF
			clear(r[0x38:0x3C])
		}
	}
	for i := 0; i < FleetCount; i++ {
		if !live[i] {
			fb[i*FleetRecord+fleetUnused] = 0xFF
		}
	}
	allocate := func(f *Fleet, kind int) (int, error) {
		for i := 0; i < ShipCount; i++ {
			r := sb[i*ShipRecord : (i+1)*ShipRecord]
			if r[shipOwner] == 0xFF && !claimed[i] {
				claimed[i] = true
				r[shipFaction], r[shipOwner], r[2], r[shipKind] = byte(f.Faction), byte(f.Index), byte(f.Region), byte(kind)
				clear(r[0x38:0x3C])
				return i, nil
			}
		}
		return -1, fmt.Errorf("艦艇池已滿（%d 筆）", ShipCount)
	}
	for i := range out {
		f := &out[i]
		r := fb[f.Index*FleetRecord : (f.Index+1)*FleetRecord]
		fresh := r[fleetUnused] == 0xFF
		changed := fresh || int(r[fleetGroups]) != len(f.Composition) || int(r[fleetUnused]) != f.Admiral || int(r[fleetAdjutant]) != int(byte(f.Adjutant))
		costChanged := changed || int(r[fleetOps]) != f.Operations
		moved := fresh || int(r[fleetRegion]) != f.Region
		for k, gp := range f.Composition {
			o := fleetSlots + 4*k
			if int(binary.LittleEndian.Uint16(r[o:])) != gp.Pool || int(r[o+2]) != gp.Slot2 || int(r[o+3]) != gp.Slot3 {
				changed = true
			}
			if gp.Pool >= 0 && int(binary.LittleEndian.Uint16(sb[gp.Pool*ShipRecord+shipAlive:])) != gp.Ships {
				costChanged = true
			}
		}
		if fresh {
			for j := range r {
				r[j] = 0xFF
			}
			r[0x3A], r[0x3B] = 0xFF, 0
		}
		// 配置旗艦、索敵隊、其餘編制。既有組的池索引不依排序重新配置。
		for k := range f.Composition {
			if k == 1 && f.Scout < 0 {
				idx, err := allocate(f, shipKindStandalone)
				if err != nil {
					return nil, err
				}
				f.Scout = idx
			}
			gp := &f.Composition[k]
			if gp.Pool < 0 {
				idx, err := allocate(f, gp.Kind)
				if err != nil {
					return nil, err
				}
				gp.Pool = idx
			}
		}
		if f.Scout < 0 {
			idx, err := allocate(f, shipKindStandalone)
			if err != nil {
				return nil, err
			}
			f.Scout = idx
		}
		ab := FleetAbilities(s.admiral(f.Admiral), s.admiral(f.Adjutant))
		if changed {
			copy(r[FleetAbilityOff:], ab[:])
		}
		if (changed || costChanged) && kinds != nil {
			binary.LittleEndian.PutUint16(r[0x0A:], uint16(FleetUpkeep(kinds, f.Faction, f.Composition, f.Operations)))
			binary.LittleEndian.PutUint16(r[0x0C:], uint16(FleetOtherCost(kinds, f.Faction, f.Composition, f.Operations)))
		}
		r[fleetFlags], r[fleetFaction], r[fleetNumber], r[3] = byte(f.Flags), byte(f.Faction), byte(f.Number), byte(f.Index)
		r[fleetUnused], r[fleetAdjutant], r[fleetRegion], r[fleetTarget] = byte(f.Admiral), byte(f.Adjutant), byte(f.Region), byte(f.Target)
		r[fleetState], r[fleetClassKey], r[fleetExpected], r[fleetGroups] = byte(f.Move), byte(f.ClassKey), byte(f.Expected), byte(len(f.Composition))
		f.Groups = len(f.Composition)
		r[fleetMorale], r[fleetOps] = byte(f.Morale), byte(f.Operations)
		r[fleetAction] = r[fleetAction]&0xF0 | byte(f.Action)&0x0F
		if fresh {
			r[fleetAction] = byte(f.Action) & 0x0F
		}
		binary.LittleEndian.PutUint16(r[fleetScout:], uint16(f.Scout))
		binary.LittleEndian.PutUint16(r[fleetPosX:], uint16(f.X))
		binary.LittleEndian.PutUint16(r[fleetPosX+2:], uint16(f.Y))
		minSpeed := [3]byte{255, 255, 255}
		for k, gp := range f.Composition {
			o := fleetSlots + 4*k
			binary.LittleEndian.PutUint16(r[o:], uint16(gp.Pool))
			r[o+2], r[o+3] = byte(gp.Slot2), byte(gp.Slot3)
			sh := sb[gp.Pool*ShipRecord : (gp.Pool+1)*ShipRecord]
			sh[shipMorale], sh[shipDetail] = byte(gp.Morale), byte(gp.Detail)
			if changed {
				sh[3], sh[5] = byte(k), byte(gp.Slot3)
			}
			if moved {
				sh[2] = byte(f.Region)
			}
			for _, v := range []struct{ off, value int }{{shipAlive, gp.Ships}, {shipFull, gp.Full}, {shipRepairCap, gp.RepairCap}, {shipPosX, gp.X}, {shipPosX + 2, gp.Y}} {
				binary.LittleEndian.PutUint16(sh[v.off:], uint16(v.value))
			}
			binary.LittleEndian.PutUint32(sh[shipMissiles:], uint32(gp.Missiles))
			binary.LittleEndian.PutUint32(sh[shipFighters:], uint32(gp.Fighters))
			if changed {
				kind, ok := kinds.Kind(f.Faction, gp.Kind)
				if !ok || len(mobility) <= len(f.Composition) {
					return nil, fmt.Errorf("缺艦種／機動力表，不能初始化現況艦艇")
				}
				p := (270 - 2*int(ab[9])) * mobility[len(f.Composition)] / 100
				base := [3]int{kind.Byte(8) * int(ab[6]) / 100, kind.Byte(6) * int(ab[5]) / 80, kind.Byte(7) * int(ab[5]) / 60}
				for n, off := range []int{0x1F, 0x21, 0x22} {
					sh[off] = byte(base[n] - base[n]*p/100)
				}
			}
			for n, off := range []int{0x1F, 0x21, 0x22} {
				if sh[off] < minSpeed[n] {
					minSpeed[n] = sh[off]
				}
			}
		}
		if changed {
			copy(r[0x4D:], minSpeed[:])
		}
		if fresh && f.Scout >= 0 && f.Scout < ShipCount {
			sh := sb[f.Scout*ShipRecord : (f.Scout+1)*ShipRecord]
			kind, ok := kinds.Kind(f.Faction, shipKindStandalone)
			if !ok {
				return nil, fmt.Errorf("缺索敵分遣隊艦種表")
			}
			sh[3], sh[5], sh[6] = 0xFF, 20, 100
			sh[0x1E] |= 0x80
			sh[0x1F], sh[0x21], sh[0x22] = byte(kind.Byte(8)), byte(kind.Byte(6)), byte(kind.Byte(7))
			for _, off := range []int{shipAlive, shipFull, shipRepairCap} {
				binary.LittleEndian.PutUint16(sh[off:], uint16(kind.FullShips))
			}
			clear(sh[shipMissiles : shipFighters+4])
		}
		if moved && f.Scout >= 0 && f.Scout < ShipCount {
			sb[f.Scout*ShipRecord+2] = byte(f.Region)
		}
	}
	s.Fleet, s.Ship = fb, sb
	return out, nil
}
