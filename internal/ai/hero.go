package ai

import (
	"sort"

	"github.com/wicanr2/logh3/internal/rules"
)

// Commander是規格140的remake指揮風格，不是原版規則。
type Commander uint8

const (
	NoCommander Commander = iota
	Reinhard
	Yang
)

func (c Commander) String() string {
	switch c {
	case Reinhard:
		return "萊因哈特"
	case Yang:
		return "楊威利"
	}
	return ""
}

// heroMoves按艦隊分配角色，避免把同一編隊拆向不同方位。
func (e *Encircle) heroMoves(b Battlefield, faction int, commander Commander, turn int) []Move {
	var mine, foes []int
	for i, u := range b.Units {
		if u.Out || u.Ships <= 0 || u.Squadron != 0 {
			continue
		}
		if u.Faction == faction {
			mine = append(mine, i)
		} else if u.Faction == 1-faction {
			foes = append(foes, i)
		}
	}
	if len(mine) == 0 || len(foes) == 0 {
		return nil
	}
	// 穩定角色與單位池排列無關，不為指揮風格多抽亂數。
	sort.SliceStable(mine, func(i, j int) bool { return b.Units[mine[i]].Fleet < b.Units[mine[j]].Fleet })
	cx, cy := centre(b, mine)
	target, best := foes[0], int64(-1)
	for _, i := range foes {
		u := b.Units[i]
		strength := 0
		for _, part := range b.Units {
			if part.Fleet == u.Fleet && !part.Out && part.Ships > 0 && part.Squadron >= 0 {
				strength += part.Strength
			}
		}
		// 附近的薄弱敵艦隊優先，避免追逐遠端零戰力索敵隊。
		score := int64(strength)*256 + int64(u.X-cx)*int64(u.X-cx) + int64(u.Y-cy)*int64(u.Y-cy)
		if best < 0 || score < best || (score == best && u.Fleet < b.Units[target].Fleet) {
			target, best = i, score
		}
	}
	t := b.Units[target]
	if turn <= 1 || e.commandAxis[faction] == [2]int{} {
		e.commandAxis[faction] = [2]int{cx - t.X, cy - t.Y}
	}
	dx, dy := e.commandAxis[faction][0], e.commandAxis[faction][1]
	length := int(rules.Distance16(0, 0, dx, dy) / rules.FixOne)
	if length == 0 {
		dx, dy, length = 1, 0, 1
	}
	radius := max(96, 224-e.Difficulty*24)
	front := 1
	if commander == Reinhard {
		front = (len(mine) + 1) / 2
	}
	var out []Move
	for n, i := range mine {
		u := b.Units[i]
		along, side := radius, 0
		if n >= front {
			wing := n - front
			sign := 1
			if wing%2 != 0 {
				sign = -1
			}
			side = sign * (radius + 64 + wing/2*48)
			along = -radius / 2
			if commander == Reinhard {
				along = -radius
			}
			// 遠處先展開翼側；跨過側方後才深入敵後，減少正面穿越。
			projection := ((u.X-t.X)*dx + (u.Y-t.Y)*dy) / length
			lateral := ((u.X-t.X)*(-dy) + (u.Y-t.Y)*dx) / length
			if projection > radius && lateral*sign < radius {
				along = radius
			}
		} else if commander == Reinhard {
			along = radius / 2
			side = (2*n - front + 1) * 32
		}
		x := t.X + (dx*along-dy*side)/length
		y := t.Y + (dy*along+dx*side)/length
		x, y = reachableHeroPoint(b, u.X, u.Y, x, y)
		out = append(out, Move{Unit: i, X: x, Y: y, Formation: true})
	}
	return out
}

func reachableHeroPoint(b Battlefield, x, y, tx, ty int) (int, int) {
	tx, ty = max(16, min(2032, tx)), max(16, min(2024, ty))
	open := func(px, py int) bool {
		return b.Terrain == nil || len(b.Cells) == 0 ||
			(b.Terrain.Passable(b.Cells, px, py) && b.Terrain.LineOpen(b.Cells, x, y, px, py))
	}
	if open(tx, ty) {
		return tx, ty
	}
	best, distance := -1, int64(1<<62)
	for i, p := range b.Rally {
		if !open(p.X, p.Y) {
			continue
		}
		d := rules.Distance16(p.X, p.Y, tx, ty)
		if d < distance {
			best, distance = i, d
		}
	}
	if best >= 0 {
		return b.Rally[best].X, b.Rally[best].Y
	}
	for step := 15; step > 0; step-- {
		px, py := x+(tx-x)*step/16, y+(ty-y)*step/16
		if open(px, py) {
			return px, py
		}
	}
	return x, y
}
