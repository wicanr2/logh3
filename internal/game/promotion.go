package game

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

func (g *Game) cacheBattlePromotionPaths() {
	g.battlePromotionPaths = nil
	if g.res.Graph == nil || len(g.res.Stellar) != 32*rules.RegionRecord {
		return
	}
	owners := make([]int, 32)
	for i := range owners {
		owners[i] = rules.RegionOwnerOf(g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord])
	}
	g.battlePromotionPaths = g.res.Graph.UnrestPaths(owners)
}

// 185C:060D，108。每次重算／等待後再試同一人，提督完畢才換副官。
func (g *Game) finishBattlePromotions() { g.battlePromotionPhase(0, false, g.endBattleTurn) }

func (g *Game) battlePromotionPhase(i int, deputy bool, next func()) {
	for i < len(g.battleFleets) {
		index := g.battleFleets[i].Index
		f := g.fleetByIndex(index)
		if f == nil {
			g.saveErr = fmt.Errorf("升級找不到參戰艦隊%d", index)
			return
		}
		admiral := f.Admiral
		if deputy {
			admiral = f.Adjutant
		}
		if admiral >= 0 {
			flags, err := g.promoteBattleAdmiral(index, admiral)
			if err != nil {
				g.saveErr = err
				return
			}
			if flags != 0 {
				rec := g.admiralRec(admiral)
				if g.humanSeat(rules.AdmiralFaction(rec)) {
					printed := uint16(2)
					if deputy {
						printed = flags | 1
					}
					idx, dep := i, deputy
					if g.startBattleReports([]turnReport{g.promotionReport(admiral, printed)}, func() { g.battlePromotionPhase(idx, dep, next) }) {
						return
					}
				}
				continue
			}
		}
		if deputy {
			i++
			deputy = false
		} else {
			deputy = true
		}
	}
	if next != nil {
		next()
	}
}

func (g *Game) promoteBattleAdmiral(index, admiral int) (uint16, error) {
	if g.slot == nil || g.res.EditorTables == nil {
		return 0, fmt.Errorf("升級缺現行存檔或機動力表")
	}
	// 前三步已窄回寫原池；SyncFleets不能在這裡重新配置或覆寫速度。
	g.slot.Admiral, g.slot.Stellar, g.slot.Planets = g.mutableAdmirals(), g.res.Stellar, g.res.Planets
	g.slot.Work = g.mutableWork()
	params := g.res.UnrestParams
	params.Turn, params.Blend, params.Rate = g.turn, g.blend, g.rate
	if len(g.slot.Work) != rules.WorkRecord {
		return 0, fmt.Errorf("升級缺WORK")
	}
	for f := 0; f < 2; f++ {
		params.Ruler[f] = int(g.slot.Work[0x3b+f])
	}
	flags, err := g.slot.PromoteBattleAdmiral(index, admiral, rules.PromotionContext{Spec: g.res.Promotion, Kinds: g.res.ShipKinds, Mobility: g.res.EditorTables.Mobility, Graph: g.res.Graph, Params: params, Paths: g.battlePromotionPaths, Military: g.military}, g.battleRand)
	if err != nil || flags == 0 {
		return flags, err
	}
	g.taxTotal = g.slot.TaxTotal
	g.syncPromotionView(index)
	return flags, nil
}

// 只同步8-4改過的欄位；命令、位置、庫存與每回合移動底稿保留。
func (g *Game) syncPromotionView(index int) {
	r := g.slot.Fleet[index*rules.FleetRecord : (index+1)*rules.FleetRecord]
	f := g.fleetByIndex(index)
	f.Flags, f.Expected, f.Morale, f.Operations = int(r[0]), int(r[14]), int(r[0x3c]), int(r[0x47])
	for i := range f.Composition {
		gp := &f.Composition[i]
		gp.Morale = int(g.slot.Ship[gp.Pool*rules.ShipRecord+6])
	}
	var ab [rules.FleetAbilityCount]byte
	copy(ab[:], r[0x3d:0x48])
	n := rules.NewBattleFleet(index, r, ab)
	for i := range g.battleFleets {
		b := &g.battleFleets[i]
		if b.Index != index {
			continue
		}
		b.Defence, b.Ability0, b.Ability3, b.Integrity, b.Flag0, b.TurnSpeed = n.Defence, n.Ability0, n.Ability3, n.Integrity, n.Flag0, n.TurnSpeed
		b.AdmiralLossFactor = int(g.admiralRec(f.Admiral)[0xb])
	}
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		if u.Fleet != index {
			continue
		}
		pool := f.Scout
		if u.Squadron >= 0 && u.Squadron < len(f.Composition) {
			pool = f.Composition[u.Squadron].Pool
		}
		if pool < 0 || pool >= rules.ShipCount {
			continue
		}
		sh := g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord]
		u.TurnSpeed, u.CruiseSpeed, u.FightSpeed = int(sh[0x1f]), int(sh[0x21]), int(sh[0x22])
		if u.Squadron >= 0 {
			u.Morale = int(sh[6])
		}
	}
}

func (g *Game) promotionReport(admiral int, flags uint16) turnReport {
	p, rec := g.res.Promotion, g.admiralRec(admiral)
	rank := codesBig5(p.RankCodes(rules.AdmiralFaction(rec), rules.RankOf(int(rec[6]))))
	b := ui.Sprintf(p.Format("DS:3DE7"), g.factionNameColoured(rules.AdmiralFaction(rec)), codesBig5(g.admiralNameCodes(admiral)))
	if flags&2 != 0 {
		b = append(b, ui.Sprintf(p.Format("DS:3DEF"), rank)...)
	} else if flags&1 != 0 {
		b = append(b, p.Format("DS:3DFE")...)
	}
	for i, off := range []int{0xb, 0x13, 0x12, 0xf, 0x10, 0x11, 0xc, 0xd, 0xe, 0x17} {
		if flags&(4<<i) != 0 {
			b = append(b, ui.Sprintf(p.Format(fmt.Sprintf("DS:%04X", 0x3e0f+15*i)), int(rec[off]))...)
		}
	}
	return turnReport{text: b, face: int(byte(admiral)), customWait: true, wait: int(uint16(g.display.MsgSpeed) * 2)}
}

// 使用記錄自身陣營，含同盟的一級上將對應名稱。
func (g *Game) admiralRankCodes(rec []byte) []uint16 {
	return g.rankCodes(rules.RankOf(byteAt(rec, admiralMeritOff)), rules.AdmiralFaction(rec))
}
