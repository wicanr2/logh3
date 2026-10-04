package game

import (
	"encoding/binary"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 185C:0491，規格107。8-2全部返回後才開始，8-4／8-5另續。
func (g *Game) finishBattleScoutFuel() {
	g.battleScoutFuelPhase(0, g.endBattleTurn)
}

func (g *Game) battleScoutFuelPhase(i int, next func()) {
	for ; i < len(g.battleFleets); i++ {
		bf := &g.battleFleets[i]
		if bf.ScoutMode == 0xFF {
			if bf.ScoutState != 0 {
				bf.ScoutState--
			}
			g.writeBattleScoutFuel(bf)
			continue
		}
		bf.ScoutState = int(byte(bf.ScoutState + 1))
		g.writeBattleScoutFuel(bf)
		if bf.ScoutState < 4 {
			continue
		}
		if f := g.fleetByIndex(bf.Index); f != nil && g.humanSeat(bf.Faction) {
			idx := i
			if g.startBattleReports([]turnReport{g.scoutLostReport(f)}, func() {
				g.recoverBattleScout(&g.battleFleets[idx])
				g.battleScoutFuelPhase(idx+1, next)
			}) {
				return
			}
		}
		g.recoverBattleScout(bf)
	}
	next()
}

// 只同步原版8-3寫入的兩個byte，其他FLEET欄位保留。
func (g *Game) writeBattleScoutFuel(bf *rules.BattleFleet) {
	bf.ScoutOut = bf.ScoutMode != 0xFF || bf.ScoutState != 0
	if g.slot != nil && bf.Index >= 0 && (bf.Index+1)*rules.FleetRecord <= len(g.slot.Fleet) {
		r := g.slot.Fleet[bf.Index*rules.FleetRecord:]
		r[0x3A], r[0x3B] = byte(bf.ScoutMode), byte(bf.ScoutState)
	}
}

func scoutRecoveredFlags(flags int) int {
	if flags&0xC0 == 0xC0 {
		flags &^= 0x40 // 1727:0298擦既有小地圖點。
	}
	return flags &^ 0x80
}

// 報告返回才回收；保留索敵原池身份，下次beginScout重用。
func (g *Game) recoverBattleScout(bf *rules.BattleFleet) {
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		if u.Fleet == bf.Index && u.Squadron < 0 {
			u.MarkerFlags = scoutRecoveredFlags(u.MarkerFlags)
			u.Marked, u.Out, u.Action = false, true, 0xFF
			u.X, u.Y, u.TargetX, u.TargetY = bf.X, bf.Y, bf.X, bf.Y
		}
	}
	if g.slot != nil {
		if f := g.fleetByIndex(bf.Index); f != nil {
			if r := g.slot.ShipAt(f.Scout); len(r) == rules.ShipRecord {
				r[0x19], r[0x1E] = byte(scoutRecoveredFlags(int(r[0x19]))), 0xFF
				for _, off := range []int{0x28, 0x34} {
					binary.LittleEndian.PutUint16(r[off:], uint16(bf.X))
					binary.LittleEndian.PutUint16(r[off+2:], uint16(bf.Y))
				}
			}
		}
	}
	bf.ScoutMode, bf.ScoutState = 0xFF, 3
	g.writeBattleScoutFuel(bf)
}

func (g *Game) scoutLostReport(f *rules.Fleet) turnReport {
	p := g.res.TurnWindow
	if p == nil {
		return turnReport{face: f.Admiral & 255}
	}
	b := ui.Sprintf(p.ReportFormat("battle_scout_prefix"), g.factionNameColoured(f.Faction))
	b = append(b, p.FleetNumber(f.Number)...)
	b = append(b, ui.Sprintf(p.ReportFormat("battle_scout_admiral"), codesBig5(g.admiralNameCodes(f.Admiral)))...)
	b = append(b, p.ReportFormat("battle_scout_tail")...)
	return turnReport{text: b, face: f.Admiral & 255}
}
