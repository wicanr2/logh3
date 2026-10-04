package game

import (
	"encoding/binary"
	"fmt"

	"github.com/wicanr2/logh3/internal/rules"
)

// 規格100的資料介面。完整092命令UI與第④步生命週期仍待接線。
// Target是battleUnits索引，提交時才映射原池；原存檔格式不新增欄位。
func (g *Game) fleetCombatCandidates(own int) ([]rules.FleetCombatCandidate, error) {
	if own < 0 || own >= len(g.battleUnits) {
		return nil, fmt.Errorf("交戰候選分隊不存在")
	}
	f := g.battleFleet(g.battleUnits[own].Fleet)
	if f == nil {
		return nil, fmt.Errorf("交戰候選艦隊不存在")
	}
	return rules.FleetCombatCandidates(g.battleUnits, own, f.Facing, g.res.ShipKinds, g.res.Terrain, g.battleCells(g.battleRegion))
}

func (g *Game) fleetCombatRecord(unit int) ([]byte, error) {
	if g.slot == nil || unit < 0 || unit >= len(g.battleUnits) || g.battleUnits[unit].Squadron < 0 {
		return nil, fmt.Errorf("交戰命令必須指向一般分隊原池")
	}
	pool := g.battleUnitPool(g.battleUnits[unit])
	if pool < 0 || pool >= len(g.slot.Ship)/rules.ShipRecord {
		return nil, fmt.Errorf("交戰命令分隊原池不存在")
	}
	return g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord], nil
}

// fleetCombatCommand讀原池現有+2C／+2E；未初始化的資料不自動改寫。
func (g *Game) fleetCombatCommand(unit int) (pool int, weapon byte, err error) {
	rec, err := g.fleetCombatRecord(unit)
	if err != nil {
		return 0, 0, err
	}
	return int(binary.LittleEndian.Uint16(rec[0x2c:])), rec[0x2e], nil
}

// commitFleetCombatCommand只提交資料，不結算傷害或選擇自動命令。
// 候選重新查詢成功後才寫入；FF／防禦3均保持既有目標word。
func (g *Game) commitFleetCombatCommand(own, target, weapon int) error {
	if weapon != 255 && (weapon < 0 || weapon > 3) {
		return fmt.Errorf("交戰命令武器無效")
	}
	rec, err := g.fleetCombatRecord(own)
	if err != nil {
		return err
	}
	u := g.battleUnits[own]
	cpu, passive := g.passiveComputer()
	if u.Action != 0x7f || u.Out || passive && u.Faction == cpu {
		return fmt.Errorf("交戰命令分隊目前不能下令")
	}
	pool := 0
	if weapon < 3 {
		candidates, err := g.fleetCombatCandidates(own)
		if err != nil {
			return err
		}
		accepted := false
		for _, c := range candidates {
			if c.Target == target && c.Mask&(1<<weapon) != 0 {
				accepted = true
				break
			}
		}
		if !accepted {
			return fmt.Errorf("交戰目標或武器已不在可用候選")
		}
		if _, err := g.fleetCombatRecord(target); err != nil {
			return err
		}
		pool = g.battleUnitPool(g.battleUnits[target])
		if pool > 0xffff {
			return fmt.Errorf("交戰目標原池超出word範圍")
		}
	}
	rec[0x2e] = byte(weapon)
	if weapon < 3 {
		binary.LittleEndian.PutUint16(rec[0x2c:], uint16(pool))
	}
	return nil
}
