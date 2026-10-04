// 第91輪DRAFT診斷。只編譯於容器暫存rules副本。
package rules

import (
	"bytes"
	"encoding/binary"
)

var probeAfterPlace = func(sf StrategicFleet, planet, facing int) {}
var probeAfterLoss = func(sf StrategicFleet, group, lost int) {}

func configureFinishRaw(slot *Slot, fleets []Fleet, before map[string][]byte, kinds *ShipKindTable, slots *FormationSlots) map[string][]byte {
	raw := map[string][]byte{}
	for k, v := range before {
		raw[k] = bytes.Clone(v)
	}
	raw["fleet"], raw["ship"] = slot.Fleet, slot.Ship
	w16 := func(b []byte, o, v int) { binary.LittleEndian.PutUint16(b[o:], uint16(v)) }
	w32 := func(b []byte, o int, v uint32) { binary.LittleEndian.PutUint32(b[o:], v) }
	probeAfterPlace = func(sf StrategicFleet, planet, facing int) {
		f := sf.Fleet
		rec := raw["fleet"][f.Index*122:]
		rec[0x07], rec[0x48] = byte(planet), 1
		rec[0x3A], rec[0x3B], rec[0x5A], rec[0x5B] = 0xFF, 0, 0xFF, 0xFF
		for _, o := range []int{0x4A, 0x4B, 0x4C} {
			rec[o] = byte(facing)
		}
		for _, o := range []int{0x5C, 0x64} {
			w16(rec, o, f.X)
			w16(rec, o+2, f.Y)
		}
		for i := range f.Composition {
			g := &f.Composition[i]
			sh := raw["ship"][g.Pool*78:]
			// 2BE1:0000先寫原隊形槽目標；297B:0D99避障只調現在位置。
			destX, destY := slots.Station(f.X, f.Y, facing, g.Slot3)
			w16(sh, 0x28, destX)
			w16(sh, 0x2A, destY)
			w16(sh, 0x34, g.X)
			w16(sh, 0x36, g.Y)
			sh[0x1E], sh[0x16], sh[0x17], sh[0x19] = 0x7F, 0, 0, 0x3F
			if raw["stellar"][f.Region*126+0x12] != 2 {
				sh[0x17], sh[0x18] = 3, 2
			}
			g.Detail = int(sh[0x17])
		}
		if f.Scout >= 0 && f.Scout < len(raw["ship"])/78 {
			sh := raw["ship"][f.Scout*78:]
			w16(sh, 0x28, f.X)
			w16(sh, 0x2A, f.Y)
			w16(sh, 0x34, f.X)
			w16(sh, 0x36, f.Y)
			sh[0x1E], sh[0x16], sh[0x17], sh[0x19] = 0xFF, 0, 0, 0x3F
		}
	}
	probeAfterLoss = func(sf StrategicFleet, group, lost int) {
		f := sf.Fleet
		g := &f.Composition[group]
		sh := raw["ship"][g.Pool*78:]
		old := int(binary.LittleEndian.Uint16(sh[8:]))
		kind, ok := kinds.Kind(f.Faction, g.Kind)
		if !ok {
			panic("缺艦種")
		}
		if old != g.Ships+lost || g.Full <= 0 {
			panic("戰損前後資料不連續")
		}
		g.RepairCap = (g.RepairCap - (lost*lost/g.Full)&0xFFFF) & 0xFFFF
		w16(sh, 0x0C, g.RepairCap)
		if old > 0 {
			if g.Kind == 5 || kind.Byte(kindMissileFlag) != 0xFF {
				g.Missiles = g.Missiles / old * g.Ships
				w32(sh, 0x0E, uint32(g.Missiles))
			}
			if g.Kind == 5 || kind.Byte(kindFighterFlag) != 0xFF {
				g.Fighters = g.Fighters / old * g.Ships
				w32(sh, 0x12, uint32(g.Fighters))
			}
		}
		w16(sh, 8, g.Ships)
		for i := range f.Composition {
			gp := &f.Composition[i]
			if gp.Ships == 0 || gp.Full == 0 {
				continue
			}
			t := (gp.Full - gp.Ships) * (200 - sf.Command) / 100
			if t > gp.Full {
				t = gp.Full
			}
			gp.Morale = (gp.Full - t) * 100 / gp.Full & 0xFF
			raw["ship"][gp.Pool*78+6] = byte(gp.Morale)
		}
		raw["fleet"][f.Index*122] = byte(f.Flags)
		raw["fleet"][f.Index*122+0x3C] = byte(f.Morale)
		// 1261:0067每次重算全陣營費用；不是只減當下分隊的費用。
		var sum uint32
		for i := range fleets {
			ff := &fleets[i]
			if ff.Faction != f.Faction {
				continue
			}
			upkeep := FleetUpkeep(kinds, ff.Faction, ff.Composition, ff.Operations)
			other := FleetOtherCost(kinds, ff.Faction, ff.Composition, ff.Operations)
			w16(raw["fleet"], ff.Index*122+0x0A, upkeep)
			w16(raw["fleet"], ff.Index*122+0x0C, other)
			sum += uint32(upkeep)
		}
		w32(raw["work"], 0x5E+4*f.Faction, sum)
	}
	return raw
}
