package rules

import "encoding/binary"

// MovementFleet依原版清單順序持有一般分隊，最後一筆為索敵隊（規格080）。
type MovementFleet struct {
	Fleet *BattleFleet
	Units []MovementOccupant
}

// BattleMovement每次Next處理完整一圈，讓正常renderer可畫中間位置。
type BattleMovement struct {
	groups    []MovementFleet
	occupants []MovementOccupant
	formation *FormationSlots
	terrain   *BattleTerrain
	cells     []uint8
}

func NewBattleMovement(groups []MovementFleet, formation *FormationSlots, terrain *BattleTerrain, cells []uint8) *BattleMovement {
	m := &BattleMovement{groups: groups, formation: formation, terrain: terrain, cells: cells}
	for _, group := range groups {
		if group.Fleet == nil {
			continue
		}
		group.Fleet.Budget = group.Fleet.PerTurn
		for _, entry := range group.Units {
			u := entry.Unit
			if u == nil {
				continue
			}
			m.occupants = append(m.occupants, entry)
			if u.Squadron < 0 && group.Fleet.ScoutMode == 0xFF {
				continue
			}
			amount := u.CruiseSpeed << 19
			if u.Budget > 0 && u.Budget < fleetStepCost {
				amount += u.Budget
			}
			u.Budget = int(int32(amount))
		}
	}
	return m
}

func (m *BattleMovement) Next() bool {
	if m == nil {
		return false
	}
	active := false
	for _, group := range m.groups {
		f := group.Fleet
		if f == nil {
			continue
		}
		check := func(x, y int) int { return BattleMovementCheck(m.terrain, m.cells, m.occupants, f.Faction, x, y) }
		var normal []*BattleUnit
		for _, entry := range group.Units {
			if entry.Unit != nil && entry.Unit.Squadron >= 0 {
				normal = append(normal, entry.Unit)
			}
		}
		fm := FleetMover{X: f.X, Y: f.Y, DestX: f.DestX, DestY: f.DestY, Facing: f.Facing, TargetFacing: f.TargetFacing,
			TurnSpeed: f.TurnSpeed, Budget: f.Budget, Action: f.Action, ShortMove: f.ShortMove, MoveFlags: f.MoveFlags}
		if AdvanceFleetStep(&fm, normal, check) {
			active = true
		}
		f.X, f.Y, f.Facing, f.Budget, f.MoveFlags = fm.X, fm.Y, fm.Facing, fm.Budget, fm.MoveFlags
		if fm.Arrived {
			f.Ordered = false
		}
		for _, entry := range group.Units {
			if AdvanceBattleUnitStep(f, entry.Unit, m.formation, check) {
				active = true
			}
		}
	}
	return active
}

// AdvanceBattleUnitStep對應1C61:0153；完整Action分派，不以Out略過旗艦／未啟用索敵隊。
func AdvanceBattleUnitStep(f *BattleFleet, u *BattleUnit, formation *FormationSlots, check func(int, int) int) bool {
	if f == nil || u == nil {
		return false
	}
	if u.Squadron == 0 || (u.Squadron < 0 && f.ScoutMode == 0xFF) {
		if u.X == f.X && u.Y == f.Y {
			return false
		}
		u.X, u.Y, u.TargetX, u.TargetY = f.X, f.Y, f.X, f.Y
		movementMarker(u)
		return true
	}
	switch u.Action {
	case 0:
		return AdvanceBattleUnitOwnStep(f, u, check)
	case 3, 0x7F:
		u.TargetX, u.TargetY = formation.Station(f.X, f.Y, f.Facing, u.Slot)
		if !u.Moving() {
			if u.Action == 3 {
				u.Action = 0x7F
			}
			if u.Kind == 7 && f.ScoutMode != 0xFF {
				f.ScoutMode, f.ScoutState, f.ScoutOut = 0xFF, 0, false
				if u.MarkerFlags&0xC0 == 0xC0 {
					u.MarkerFlags &^= 0x40
				}
				u.MarkerFlags &^= 0x80
				u.Marked = false
				u.Action |= 0x80
				u.Out = true
			}
			return false
		}
		if u.Budget < fleetStepCost {
			return false
		}
		step := BattleStep
		if u.Action == 0x7F {
			if f.MoveFlags&1 != 0 {
				step *= 2
			}
			if u.Attack&2 != 0 {
				step = step * 3 / 2
			}
		}
		MoveBattleUnitStep(f, u, step, check)
		return true
	}
	return false
}

// AdvanceBattleUnitOwnStep對應1C61:03AA；直接常式呼叫不經旗艦／未啟用索敵隊分派。
func AdvanceBattleUnitOwnStep(f *BattleFleet, u *BattleUnit, check func(int, int) int) bool {
	if f == nil || u == nil {
		return false
	}
	if !u.Moving() || (u.Kind != 7 && u.Budget < fleetStepCost) {
		u.Action = 3
		u.Budget = 0
		return false
	}
	if u.Budget < fleetStepCost {
		return false
	}
	MoveBattleUnitStep(f, u, BattleStep, check)
	return true
}

// MoveBattleUnitStep對應1C61:0444；void常式，不對其暫存器殘值作返回值聲明。
func MoveBattleUnitStep(f *BattleFleet, u *BattleUnit, step int, check func(int, int) int) {
	if f == nil || u == nil {
		return
	}
	x, y := u.TargetX, u.TargetY
	if Distance16(u.X, u.Y, x, y) > int64(step*FixOne) {
		x, y = StepAlong(u.X, u.Y, Bearing(u.X, u.Y, x, y), step)
	}
	blocked := 0
	if check != nil {
		blocked = check(x, y)
	}
	if blocked != 0 {
		if u.Kind != 7 || (blocked != MovementTerrainBlocked && u.Action == 3) {
			x, y = AvoidBattleObstacle(u.X, u.Y, x, y, step, check)
		} else {
			if blocked == MovementTerrainBlocked {
				u.Action = 3
			} else {
				f.ScoutState = 4
			}
			u.Budget = fleetStepCost
			x, y = u.X, u.Y
		}
	}
	u.X, u.Y = x, y
	u.Budget = int(int32(u.Budget - fleetStepCost))
	movementMarker(u)
}

func movementMarker(u *BattleUnit) {
	if u.MarkerFlags&0x40 != 0 {
		u.MarkerX, u.MarkerY = u.X, u.Y
	}
}

// ApplyBattleUnitMovement只寫規格080已審查欄位，保留DOS圖示指標與其他原始byte。
func (s *Slot) ApplyBattleUnitMovement(pool int, u BattleUnit) bool {
	if s == nil || pool < 0 || pool >= len(s.Ship)/ShipRecord {
		return false
	}
	r := s.Ship[pool*ShipRecord : (pool+1)*ShipRecord]
	r[5], r[0x16], r[0x19], r[0x1E] = byte(u.Slot), byte(u.Attack), byte(u.MarkerFlags), byte(u.Action)
	binary.LittleEndian.PutUint16(r[0x1A:], uint16(u.MarkerX))
	binary.LittleEndian.PutUint16(r[0x1C:], uint16(u.MarkerY))
	binary.LittleEndian.PutUint32(r[0x24:], uint32(u.Budget))
	for _, p := range []struct{ off, value int }{{0x28, u.TargetX}, {0x2A, u.TargetY}, {0x34, u.X}, {0x36, u.Y}} {
		binary.LittleEndian.PutUint16(r[p.off:], uint16(p.value))
	}
	return true
}

func (s *Slot) ApplyBattleMovementFleet(f BattleFleet) bool {
	if !s.ApplyBattleFleetMovement(f) {
		return false
	}
	r := s.Fleet[f.Index*FleetRecord : (f.Index+1)*FleetRecord]
	action := f.Action
	if f.ShortMove {
		action |= 16
	}
	r[0x3A], r[0x3B], r[0x48], r[0x4C], r[0x4D] = byte(f.ScoutMode), byte(f.ScoutState), byte(action), byte(f.TargetFacing), byte(f.TurnSpeed)
	binary.LittleEndian.PutUint32(r[0x52:], uint32(f.PerTurn))
	for _, p := range []struct{ off, value int }{{0x60, f.OriginX}, {0x62, f.OriginY}, {0x64, f.DestX}, {0x66, f.DestY}} {
		binary.LittleEndian.PutUint16(r[p.off:], uint16(p.value))
	}
	return true
}
