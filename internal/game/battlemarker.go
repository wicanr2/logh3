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
		}
	}
}
