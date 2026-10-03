package game

import "github.com/wicanr2/logh3/internal/rules"

// initBattleMarkers依297B:08D8的佈陣重設標記旗標，快照待實際點上才更新。
func (g *Game) initBattleMarkers() {
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		u.MarkerFlags, u.Marked = 0x3F, false
	}
	// 未放出的索敵隊也重設，仍不加入可點選單位。
	if g.slot != nil {
		for _, bf := range g.battleFleets {
			if f := g.fleetByIndex(bf.Index); f != nil {
				if r := g.slot.ShipAt(f.Scout); len(r) == rules.ShipRecord {
					r[0x19] = 0x3F
				}
			}
		}
	}
	g.refreshBattleMarkers()
}

// 正常單人輸入恢復時依26E5:00AC刷新標記，不改索敵本身的順序或規則。
func (g *Game) refreshBattleMarkers() {
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		forced := u.Faction == g.faction
		if u.Squadron < 0 {
			bf := g.battleFleet(u.Fleet)
			forced = forced && bf != nil && bf.ScoutMode != 0xFF
		}
		rules.UpdateBattleMarker(u, forced, true)
		if g.slot != nil {
			g.slot.ApplyBattleMarker(g.battleUnitPool(*u), *u)
			if r := g.slot.ShipAt(g.battleUnitPool(*u)); len(r) == rules.ShipRecord {
				r[0x2F] = byte(rules.BattleCachedSprite(*u, forced, int(r[0x2F])))
			}
			if u.Squadron == 0 && u.Fleet >= 0 && (u.Fleet+1)*rules.FleetRecord <= len(g.slot.Fleet) {
				d := rules.BlankSprite
				if forced || u.Detail == rules.SpottingMax {
					if bf := g.battleFleet(u.Fleet); bf != nil {
						d = rules.DirectionIndex(bf.Facing)
					}
				}
				g.slot.Fleet[u.Fleet*rules.FleetRecord+0x6A] = byte(d)
			}
		}
	}
	// 26E5:016F..01B6也刷新未出擊偵察池；它不在可點選單位清單。
	if g.slot != nil {
		for _, bf := range g.battleFleets {
			if bf.ScoutMode != 0xFF {
				continue
			}
			f := g.fleetByIndex(bf.Index)
			if f == nil {
				continue
			}
			if u, ok := g.slot.ScoutUnit(g.res.Weights, *f, bf.X, bf.Y); ok {
				r := g.slot.ShipAt(f.Scout)
				r[0x2F] = byte(rules.BattleCachedSprite(u, false, int(r[0x2F])))
			}
		}
	}
}
