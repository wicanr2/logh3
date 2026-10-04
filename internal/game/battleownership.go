package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// finishBattleOwnership 接在8-1之後，再續107的索敵燃料。
func (g *Game) finishBattleOwnership() {
	g.battleOwnershipPhase(g.finishBattleScoutFuel)
}

// battleOwnershipPhase 是185C:02DA，契約與完整原版證據見規格106。
// 報告期間三方行星數已更新，星域擁有者留到兩個通知都返回才寫。
func (g *Game) battleOwnershipPhase(next func()) {
	rec := g.regionRecord(g.battleRegion)
	if rec == nil {
		next()
		return
	}
	var counts [3]byte
	for _, p := range g.battlePlanets {
		counts[p.Owner]++ // 正式行星Owner域為0、1、2。
	}
	copy(rec[0x16:0x19], counts[:])
	owner := rules.RegionOwnerNone
	if counts[1] == 0 && counts[2] == 0 {
		owner = 0
	} else if counts[0] == 0 && counts[2] == 0 {
		owner = 1
	}
	if int(rec[regionOwnerOff]) == owner {
		next()
		return
	}
	g.battleOwnershipNotify(0, owner, func() {
		rec[regionOwnerOff] = byte(owner)
		if g.battleRegion == rules.PhezzanRegion {
			g.phezzanBattleResolved = true
		}
		next()
	})
}

// battleOwnershipNotify 是185C:03C6。逐座位續接，讓費沙登記發生在
// 取得方自己的報告前，不能提早到對手丟失報告之前。
func (g *Game) battleOwnershipNotify(faction, owner int, next func()) {
	for ; faction < 2; faction++ {
		if !g.humanSeat(faction) {
			continue
		}
		key := "battle_region_lost"
		if owner == faction {
			key = "battle_region_taken"
			if g.battleRegion == rules.PhezzanRegion && !rules.PhezzanOccurred(g.workRecord()) {
				g.phezzanPending = true
			}
		} else if owner == rules.RegionOwnerNone {
			key = "battle_region_partly_landed"
			if g.regionOwner(g.battleRegion) == faction {
				key = "battle_region_grip_lost"
			}
		}
		if g.res.TurnWindow == nil {
			continue
		}
		report := turnReport{face: g.res.Reporter[faction], text: ui.Sprintf(g.res.TurnWindow.ReportFormat(key),
			g.regionNameBytesOf(g.battleRegion))}
		resume := faction + 1
		if g.startBattleReports([]turnReport{report}, func() {
			g.battleOwnershipNotify(resume, owner, next)
		}) {
			return
		}
	}
	next()
}
