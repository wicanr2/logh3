package rules

import "fmt"

// FleetWeaponRange共用093的原版武器表；顯示門檻沿規格103。
type FleetWeaponRange struct {
	Major, Minor int
	Enabled      bool
}

func FleetCombatWeaponRanges(kinds *ShipKindTable, u BattleUnit) ([3]FleetWeaponRange, error) {
	var out [3]FleetWeaponRange
	k, err := fleetCombatKind(kinds, &u)
	if err != nil {
		return out, err
	}
	for w := range out {
		i := k.Byte(10 + w*4)
		if i == 255 {
			continue
		}
		if i < 0 || i >= len(fleetWeaponProfiles) {
			return out, fmt.Errorf("未知交戰射程type%d", i)
		}
		p := fleetWeaponProfiles[i]
		out[w] = FleetWeaponRange{Major: p.major, Minor: p.minor, Enabled: w == 0 || w == 1 && u.Missiles > 0 || w == 2 && u.Fighters > 0}
	}
	return out, nil
}
