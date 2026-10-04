package game

import (
	"encoding/binary"
	"fmt"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 110 READY：8-4返回後，先逐則等待敗方報告，再進勝方占領。
func (g *Game) finishBattleDecision() { g.battleFinishPhase(g.endBattleTurn) }

func (g *Game) battleFinishPhase(next func()) {
	g.battleFinishing = true
	g.battleFinishEnded = false
	var ids []int
	for _, f := range g.battleFleets {
		ids = append(ids, f.Index)
	}
	d, err := g.slot.DecideBattleFinish(ids)
	if err != nil {
		g.saveErr = err
		return
	}
	g.battleFinishEnded = d.Ended
	g.battleLoserReport(d, 0, func() {
		if d.Winner < 0 {
			g.battleFinishing = false
			next()
			return
		}
		if err := g.beginWinnerOccupation(d.Winner); err != nil {
			g.saveErr = err
			return
		}
		g.continueWinnerOccupation(next)
	})
}

func (g *Game) battleLoserReport(d rules.BattleFinishDecision, at int, next func()) {
	for at < len(d.Losers) {
		loser := d.Losers[at]
		seat := loser
		key := "battle_lost_us"
		if !g.humanSeat(loser) {
			if !g.humanSeat(1 - loser) {
				at++
				continue
			}
			key = "battle_lost_ai"
			seat = 1 - loser
		}
		if g.res.TurnWindow != nil {
			resume := at + 1
			// 3826:045C..0472：非FFFF座位改選情報參謀，忽略caller頭像參數0。
			if g.startBattleReports([]turnReport{{face: g.res.Reporter[seat], text: g.res.TurnWindow.ReportFormat(key)}}, func() { g.battleLoserReport(d, resume, next) }) {
				return
			}
		}
		at++
	}
	next()
}

func (g *Game) beginWinnerOccupation(winner int) error {
	if g.slot == nil || g.res.Planets == nil || g.res.Terrain == nil || g.battleRand == nil {
		return fmt.Errorf("勝方占領缺目前存檔、地形或亂數")
	}
	cells := g.battleCells(g.battleRegion)
	if len(cells) == 0 {
		return fmt.Errorf("勝方占領缺戰場地圖")
	}
	var pools []int
	for _, u := range g.battleUnits {
		f := g.fleetByIndex(u.Fleet)
		if f == nil {
			return fmt.Errorf("勝方占領單位沒有艦隊")
		}
		pool := f.Scout
		if u.Squadron >= 0 {
			if u.Squadron >= len(f.Composition) {
				return fmt.Errorf("勝方占領分隊越界")
			}
			pool = f.Composition[u.Squadron].Pool
		}
		pools = append(pools, pool)
	}
	// 前四步已提交分隊原池。行星與兩個計分器仍有結構欄位，先接成同一份現況。
	g.slot.Admiral, g.slot.Stellar, g.slot.Planets = g.mutableAdmirals(), g.res.Stellar, g.res.Planets
	g.slot.Work = g.mutableWork()
	if len(g.slot.Work) != rules.WorkRecord {
		return fmt.Errorf("勝方占領缺WORK")
	}
	g.res.Planets.ApplyBattlePlanets(g.battlePlanets)
	for f := 0; f < 2; f++ {
		binary.LittleEndian.PutUint32(g.slot.Work[0xd0+4*f:], uint32(g.battleScore[f]))
		binary.LittleEndian.PutUint32(g.slot.Work[0xd8+4*f:], uint32(g.campaignScore[f]))
	}
	c, err := rules.NewWinnerOccupation(g.slot, g.battleRegion, winner, g.battleTurn-1, pools,
		g.res.ShipKinds, g.res.Formation, func(x, y int) bool { return g.res.Terrain.Passable(cells, x, y) }, g.battleRand)
	if err != nil {
		return err
	}
	g.battleOccupation = c
	g.battleOccupationScore = [2]int{}
	return nil
}

func (g *Game) continueWinnerOccupation(next func()) {
	for {
		event := g.battleOccupation.Step()
		g.syncWinnerOccupation()
		if event != nil {
			if g.res.TurnWindow != nil && g.startBattleReports([]turnReport{g.battleCaptureReport(*event)}, func() { g.continueWinnerOccupation(next) }) {
				return
			}
			continue
		}
		g.battleOccupation = nil
		g.battleFinishing = false
		next()
		return
	}
}

// 每個報告等待點重讀已提交欄位，後續SyncFleets／ApplyBattleResult不能蓋回舊庫存。
func (g *Game) syncWinnerOccupation() {
	res := g.battleOccupation.Result()
	for f := 0; f < 2; f++ {
		delta := res.Score[f] - g.battleOccupationScore[f]
		g.battleScore[f] = int(uint32(g.battleScore[f]) + uint32(delta))
		g.campaignScore[f] = int(uint32(g.campaignScore[f]) + uint32(delta))
		g.battleFleetScore[f] += delta // 已即時計入整局，EndBattle不再加一次。
	}
	g.battleOccupationScore = res.Score
	g.fleets = g.slot.Fleets(g.res.Weights)
	for i := range g.battleFleets {
		b := &g.battleFleets[i]
		r := g.slot.Fleet[b.Index*rules.FleetRecord : (b.Index+1)*rules.FleetRecord]
		b.X, b.Y = int(binary.LittleEndian.Uint16(r[0x5c:])), int(binary.LittleEndian.Uint16(r[0x5e:]))
		b.DestX, b.DestY = int(binary.LittleEndian.Uint16(r[0x64:])), int(binary.LittleEndian.Uint16(r[0x66:]))
		b.Facing, b.TargetFacing, b.Action = int(r[0x4a]), int(r[0x4c]), int(r[0x48])
		b.Flag0, b.Integrity = r[0]&1 != 0, int(r[0x3c])
		b.ScoutMode, b.ScoutState, b.ScoutOut = int(r[0x3a]), int(r[0x3b]), r[0x3a] != 255 || r[0x3b] != 0
	}
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		f := g.fleetByIndex(u.Fleet)
		if u.Squadron < 0 {
			v, ok := g.slot.ScoutUnit(g.res.Weights, *f, u.X, u.Y)
			if ok {
				r := g.slot.Ship[f.Scout*rules.ShipRecord:]
				v.X, v.Y = int(binary.LittleEndian.Uint16(r[0x34:])), int(binary.LittleEndian.Uint16(r[0x36:]))
				v.TargetX, v.TargetY = int(binary.LittleEndian.Uint16(r[0x28:])), int(binary.LittleEndian.Uint16(r[0x2a:]))
				v.Speed = u.Speed
				*u = v
			}
		} else {
			v := g.slot.BattleUnitsFor(g.res.Weights, *f)[u.Squadron]
			v.Speed = u.Speed
			*u = v
		}
	}
	for i := range g.battlePlanets {
		p := &g.battlePlanets[i]
		p.Owner = g.res.Planets.Owner(p.Index)
		p.Defence = g.res.Planets.Defence(p.Index)
		p.Stock, _ = rules.ReadPlanetStock(g.res.Planets.Record(p.Index))
	}
}

// 185C:0B18只印行星名，與戰略captureReport的星域／行星文字不同。
func (g *Game) battleCaptureReport(event rules.WinnerCapture) turnReport {
	p := g.res.TurnWindow
	f := g.fleetByIndex(event.Fleet)
	var b []byte
	if g.humanSeat(f.Faction) {
		b = append(b, ui.Sprintf(p.ReportFormat("battle_capture_prefix"), g.factionNameColoured(f.Faction))...)
		b = append(b, p.FleetNumber(f.Number)...)
		b = append(b, ui.Sprintf(p.ReportFormat("battle_capture_admiral"), codesBig5(g.admiralNameCodes(f.Admiral)))...)
	} else {
		b = append(b, ui.Sprintf(p.ReportFormat("battle_capture_ai"), g.factionNameColoured(f.Faction))...)
	}
	start, _ := g.res.Planets.Range(g.battleRegion)
	i := start + event.Planet
	b = append(b, p.PlainName(planetNameBytes(g.res.Planets.Record(i)), g.res.Planets.Owner(i))...)
	b = append(b, p.ReportFormat("battle_capture_tail")...)
	// caller傳完整FLEET+04 word；3826只讀低byte。電腦路由用對手情報參謀。
	face := f.Admiral
	if !g.humanSeat(f.Faction) {
		face = g.res.Reporter[1-f.Faction]
	}
	return turnReport{text: b, face: face}
}
