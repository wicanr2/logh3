package game

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/wicanr2/logh3/internal/rules"
)

// cleanupBattleFleet在8-1提交當下戰損及四byte編制引用，供後續階段直接取池。
// 109 READY：存活分隊保留活的命令，Squadron只更新remake索引。
func (g *Game) cleanupBattleFleet(index int) error {
	f := g.fleetByIndex(index)
	if f == nil || g.slot == nil || index < 0 || (index+1)*rules.FleetRecord > len(g.slot.Fleet) {
		return fmt.Errorf("戰場清理缺現行艦隊或原表")
	}
	r := g.slot.Fleet[index*rules.FleetRecord : (index+1)*rules.FleetRecord]
	n := int(r[15])
	if n < 1 || n > 10 || len(f.Composition) != n {
		return fmt.Errorf("戰場清理的結構化編制與原表不符")
	}
	for i, gp := range f.Composition {
		if gp.Pool != int(binary.LittleEndian.Uint16(r[16+4*i:])) {
			return fmt.Errorf("戰場清理的結構化池引用與原表不符")
		}
	}
	var units []rules.BattleUnit
	seen := make([]bool, n)
	for _, u := range g.battleUnits {
		if u.Fleet != index || u.Squadron < 0 {
			continue
		}
		if u.Squadron >= n || seen[u.Squadron] {
			return fmt.Errorf("戰場清理的分隊索引越界或重複")
		}
		seen[u.Squadron] = true
		units = append(units, u)
	}
	for _, yes := range seen {
		if !yes {
			return fmt.Errorf("戰場清理缺現行分隊")
		}
	}
	// 行星砲等路徑亦可留在BattleUnit；不能依戰前Composition判定死亡。
	work := *g.slot
	work.Fleet, work.Ship = bytes.Clone(g.slot.Fleet), bytes.Clone(g.slot.Ship)
	work.ApplyBattleResult(units)
	mapping, err := work.PruneEmptyBattleGroups(index)
	if err != nil {
		return err
	}
	composition := work.FleetComposition(index)
	out := make([]rules.BattleUnit, 0, len(g.battleUnits))
	for _, u := range g.battleUnits {
		if u.Fleet == index && u.Squadron >= 0 {
			i := mapping[u.Squadron]
			if i < 0 {
				continue
			}
			u.Squadron = i
		}
		out = append(out, u)
	}
	g.slot.Fleet, g.slot.Ship = work.Fleet, work.Ship
	f.Composition, f.Groups = composition, len(composition)
	g.battleUnits = out
	return nil
}
