package rules

import "fmt"

// FleetCombatCandidate是規格100的單次查詢結果。Target是caller物件索引，
// Mask的bit0／1／2為光線砲／飛彈／戰鬥機；不保存SHIP+3C暫存殘值。
type FleetCombatCandidate struct {
	Target int
	Mask   uint8
}

// FleetCombatCandidates依1F6F:02D2查目前位置。原版查詢先於自身Action=7F
// 的命令門檻，因此此函式不先刪除Out或零艘數的攻方。
func FleetCombatCandidates(units []BattleUnit, own, facing int, kinds *ShipKindTable,
	terrain *BattleTerrain, cells []uint8) ([]FleetCombatCandidate, error) {
	return fleetCombatCandidates(units, own, facing, kinds, terrain, cells, fleetWeaponProfiles[:])
}

// 接受獨立的range表以驗證原版受控表值；正式入口共用093凍結表。
func fleetCombatCandidates(units []BattleUnit, own, facing int, kinds *ShipKindTable,
	terrain *BattleTerrain, cells []uint8, profiles []fleetWeaponProfile) ([]FleetCombatCandidate, error) {
	if own < 0 || own >= len(units) || facing < 0 || facing > 255 {
		return nil, fmt.Errorf("交戰候選攻方或朝向無效")
	}
	if len(profiles) != 12 || terrain == nil || len(terrain.BlockMask) == 0 || len(terrain.QuadrantBits) != 4 || len(cells) != 128*102 {
		return nil, fmt.Errorf("交戰候選地形或地圖缺失")
	}
	for _, p := range profiles {
		if p.major <= 0 || p.minor <= 0 || p.minor > p.major || p.major > 0x7fff {
			return nil, fmt.Errorf("交戰候選射程半軸無效")
		}
	}
	for _, c := range cells {
		if int(c) >= len(terrain.BlockMask) {
			return nil, fmt.Errorf("交戰候選圖塊索引超出地形表")
		}
	}
	for _, u := range units {
		if u.Faction < 0 || u.Faction > 1 || u.Kind < 0 || u.Kind > 7 || u.Detail < 0 || u.Detail > 255 ||
			u.Action < 0 || u.Action > 255 || u.X < 0 || u.X > 2047 || u.Y < 0 || u.Y > 2039 ||
			u.Missiles < 0 || u.Fighters < 0 || uint64(u.Missiles) > 0xffffffff || uint64(u.Fighters) > 0xffffffff {
			return nil, fmt.Errorf("交戰候選分隊欄位無效")
		}
	}
	a := units[own]
	k, err := fleetCombatKind(kinds, &a)
	if err != nil {
		return nil, err
	}
	var types [3]int
	for w := range types {
		types[w] = k.Byte(10 + w*4)
		if types[w] != 255 && types[w] >= len(fleetWeaponProfiles) {
			return nil, fmt.Errorf("交戰候選未知武器type%d", types[w])
		}
	}
	var out []FleetCombatCandidate
	for i, b := range units {
		if b.Faction == a.Faction || b.Out || b.Action&128 != 0 || b.Detail == 0 || b.Kind == 7 {
			continue
		}
		distance := FixedDist(a.X, a.Y, b.X, b.Y)
		if distance > FixedFromInt(300) {
			continue
		}
		delta := AngleDiff256(Atan256(b.X-a.X, b.Y-a.Y), facing)
		mask := uint8(0)
		for w, kind := range types {
			if kind == 255 || w == 1 && a.Missiles == 0 || w == 2 && a.Fighters == 0 {
				continue
			}
			p := profiles[kind]
			if distance > EllipseRadius(p.major, p.minor, delta) {
				continue
			}
			if w == 0 && !fleetBeamClear(units, own, i, terrain, cells) {
				continue
			}
			mask |= 1 << w
		}
		if mask != 0 {
			out = append(out, FleetCombatCandidate{Target: i, Mask: mask})
		}
	}
	return out, nil
}

// fleetBeamClear沿091C每第8個整數射線步驗地形與分隊。不是LineOpen。
func fleetBeamClear(units []BattleUnit, own, target int, terrain *BattleTerrain, cells []uint8) bool {
	a, b := units[own], units[target]
	var blocks []BattleUnit
	for i, u := range units {
		if i == own || i == target || u.Out || u.Action&128 != 0 {
			continue
		}
		if u.X >= min(a.X, b.X) && u.X <= max(a.X, b.X) && u.Y >= min(a.Y, b.Y) && u.Y <= max(a.Y, b.Y) {
			blocks = append(blocks, u)
		}
	}
	x1, y1, x2, y2 := a.X+7, a.Y+7, b.X+7, b.Y+7
	dx, dy := abs32(x2-x1), abs32(y2-y1)
	clear := func(x, y, count int) bool {
		if count&7 != 0 {
			return true
		}
		if terrain.blocked(cells, x/8, y/10) {
			return false
		}
		for _, u := range blocks {
			if x >= u.X && x < u.X+16 && y >= u.Y && y < u.Y+16 {
				return false
			}
		}
		return true
	}
	if dx > dy {
		if x1 > x2 {
			x1, x2, y1, y2 = x2, x1, y2, y1
		}
		sign := 1
		if y1 >= y2 {
			sign = -1
		}
		err := dx / 2
		for x, count := x1+1, 1; x <= x2; x, count = x+1, count+1 {
			err -= dy
			if err < 0 {
				err += dx
				y1 += sign
			}
			if !clear(x, y1, count) {
				return false
			}
		}
	} else {
		if y1 > y2 {
			x1, x2, y1, y2 = x2, x1, y2, y1
		}
		sign := 1
		if x1 >= x2 {
			sign = -1
		}
		err := dy / 2
		for y, count := y1+1, 1; y <= y2; y, count = y+1, count+1 {
			err -= dx
			if err < 0 {
				err += dy
				x1 += sign
			}
			if !clear(x1, y, count) {
				return false
			}
		}
	}
	return true
}
