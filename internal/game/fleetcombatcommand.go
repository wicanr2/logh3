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

// selectFleetAutoCombatCommand沿101提交自動預估及命令資料。
// 原版沒有beam時可能沿最後候選的武器，不使用手動100的mask拒絕。
// 完整092仍未呼叫本介面；此處不啟用061被動CPU或結算戰損。
func (g *Game) selectFleetAutoCombatCommand(own int) (rules.FleetAutoSelection, error) {
	var empty rules.FleetAutoSelection
	rec, err := g.fleetCombatRecord(own)
	if err != nil {
		return empty, err
	}
	u := g.battleUnits[own]
	cpu, passive := g.passiveComputer()
	if u.Action != 0x7f || u.Out || passive && u.Faction == cpu {
		return empty, fmt.Errorf("自動交戰分隊目前不能下令")
	}
	rawFleet := func(index int) ([]byte, error) {
		if index < 0 || index >= len(g.slot.Fleet)/rules.FleetRecord {
			return nil, fmt.Errorf("自動交戰艦隊原表缺失")
		}
		return g.slot.Fleet[index*rules.FleetRecord : (index+1)*rules.FleetRecord], nil
	}
	fleet, err := rawFleet(u.Fleet)
	if err != nil {
		return empty, err
	}
	candidates, err := g.fleetCombatCandidates(own)
	if err != nil {
		return empty, err
	}
	var input []rules.FleetAutoCandidate
	var records [][]byte
	pools := map[int]bool{}
	for _, c := range candidates {
		target, err := g.fleetCombatRecord(c.Target)
		if err != nil {
			return empty, err
		}
		tf, err := rawFleet(g.battleUnits[c.Target].Fleet)
		if err != nil {
			return empty, err
		}
		pool := g.battleUnitPool(g.battleUnits[c.Target])
		if pool > 65535 || pools[pool] {
			return empty, fmt.Errorf("自動交戰目標原池重複或超出word")
		}
		pools[pool] = true
		command := rules.FleetCombatCommand{TargetWeapon: int(target[0x2e]), TargetFacing: int(tf[0x4a])}
		for i := range command.AttackAbilities {
			command.AttackAbilities[i], command.DefendAbilities[i] = int(fleet[0x3d+i]), int(tf[0x3d+i])
		}
		input = append(input, rules.FleetAutoCandidate{Target: c.Target, Unit: g.battleUnits[c.Target],
			Mask: c.Mask, Anticipated: binary.LittleEndian.Uint16(target[0x3e:]), Command: command})
		records = append(records, target)
	}
	// originalBattleAI與對戰保留原版電腦權重；單人CPU已在上方拒絕。
	computer := !g.versus && u.Faction != g.faction
	selected, err := rules.SelectFleetAutoCommand(u, binary.LittleEndian.Uint16(rec[0x3e:]), input,
		g.res.ShipKinds, computer, int(fleet[0x45]), g.battleRand)
	if err != nil {
		return empty, err
	}
	rec[0x2e] = selected.Weapon
	if selected.Target < 0 {
		return selected, nil
	}
	binary.LittleEndian.PutUint16(rec[0x2c:], uint16(g.battleUnitPool(g.battleUnits[selected.Target])))
	binary.LittleEndian.PutUint16(rec[0x3e:], selected.Anticipated)
	for i, e := range selected.Evaluations {
		target := records[i]
		target[0x3c] = e.Mask
		binary.LittleEndian.PutUint16(target[0x46:], e.Score)
		for w, v := range e.WeaponScores {
			binary.LittleEndian.PutUint16(target[0x48+w*2:], v)
		}
	}
	return selected, nil
}
