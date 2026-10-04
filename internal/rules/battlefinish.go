package rules

import (
	"encoding/binary"
	"fmt"
)

// BattleFinishDecision 是185C:09BF的計數結果，限110 READY契約。
// 決定在占領前固定，後續損失不能逆轉本次勝負。
type BattleFinishDecision struct {
	Able   [2]byte
	Losers []int
	Ended  bool
	Winner int // -1表示雙方可戰或雙方歸零
}

func (s *Slot) DecideBattleFinish(fleetIDs []int) (BattleFinishDecision, error) {
	d := BattleFinishDecision{Winner: -1}
	if s == nil || len(s.Fleet) != FleetCount*FleetRecord || len(fleetIDs) > FleetCount {
		return d, fmt.Errorf("戰場結束缺艦隊表或清單越界")
	}
	var seen [FleetCount]bool
	for _, i := range fleetIDs {
		if i < 0 || i >= FleetCount || seen[i] || !s.FleetUsed(i) || s.fleet(i)[1] > 1 {
			return BattleFinishDecision{Winner: -1}, fmt.Errorf("戰場結束的艦隊身分無效")
		}
		seen[i] = true
		r := s.fleet(i)
		if r[0]&fleetFlagLowMorale == 0 {
			d.Able[r[1]]++
		}
	}
	for f := 0; f < 2; f++ {
		if d.Able[f] == 0 {
			d.Losers = append(d.Losers, f)
		}
	}
	d.Ended = len(d.Losers) != 0
	if len(d.Losers) == 1 {
		d.Winner = 1 - d.Losers[0]
	}
	return d, nil
}

// WinnerCapture 是185C:0B18的等待點。Owner、庫存與得分已提交；
// 此行星的上限功績在下一次Step返回報告後才提交。
type WinnerCapture struct{ Fleet, Planet int }

// WinnerOccupation 保留20B4:1512的局部回合、行星與RND。
// 只供戰場即時續接，不是存檔格式。呼叫期間Slot由此狀態獨占修改。
type WinnerOccupation struct {
	slot                          *Slot
	fleets                        []Fleet
	attackers                     []StrategicFleet
	place                         StrategicPlacement
	kinds                         *ShipKindTable
	poolIDs                       []int
	planets                       []byte
	sites                         []PlanetSite
	rnd                           *Rand
	winner, round, passes, planet int
	pending                       *WinnerCapture
	done                          bool
	result                        StrategicAssaultResult
}

// NewWinnerOccupation 先驗證全部輸入。錯誤不改原池、不消費亂數。
func NewWinnerOccupation(s *Slot, region, winner, round int, poolIDs []int,
	kinds *ShipKindTable, slots *FormationSlots, passable func(int, int) bool, rnd *Rand) (*WinnerOccupation, error) {
	if s == nil || len(s.Fleet) != FleetCount*FleetRecord || len(s.Ship) != ShipCount*ShipRecord ||
		len(s.Admiral) != 82*AdmiralRecord || len(s.Stellar) != 32*RegionRecord || len(s.Work) != WorkRecord ||
		s.Planets == nil || len(s.Planets.PerRegion) != 32 || winner < 0 || winner > 1 || region < 0 || region >= 32 ||
		round < 0 || round > 9 || rnd == nil || kinds == nil || slots == nil || len(slots.Slots) == 0 || passable == nil {
		return nil, fmt.Errorf("勝方占領缺六表、地形、隊形、艦種或合法參數")
	}
	parsed, err := LoadPlanets(s.Planets.Data)
	if err != nil {
		return nil, err
	}
	for i, n := range parsed.PerRegion {
		if n != s.Planets.PerRegion[i] {
			return nil, fmt.Errorf("行星表與索引不同")
		}
	}
	start, n := s.Planets.Range(region)
	if n == 0 {
		return nil, fmt.Errorf("勝方占領星域沒有行星")
	}
	c := &WinnerOccupation{slot: s, kinds: kinds, poolIDs: append([]int(nil), poolIDs...), rnd: rnd,
		winner: winner, round: round, passes: 3, result: StrategicAssaultResult{Merit: map[int]int{}},
		place:   StrategicPlacement{Slots: slots, Passable: passable},
		planets: s.Planets.Data[PlanetHeader+start*PlanetRecord : PlanetHeader+(start+n)*PlanetRecord]}
	for i := 0; i < n; i++ {
		p := c.planets[i*PlanetRecord:]
		if p[planetOwner] > 2 {
			return nil, fmt.Errorf("行星陣營無效")
		}
		c.sites = append(c.sites, PlanetSite{X: int(binary.LittleEndian.Uint16(p[0x46:])), Y: int(binary.LittleEndian.Uint16(p[0x48:]))})
	}
	var seen [ShipCount]bool
	for _, pool := range poolIDs {
		if pool < 0 || pool >= ShipCount || seen[pool] || s.Ship[pool*ShipRecord] > 1 || s.Ship[pool*ShipRecord+4] > 7 {
			return nil, fmt.Errorf("戰場有效池身分無效")
		}
		seen[pool] = true
		sh := s.ShipAt(pool)
		owner := int(sh[shipOwner])
		if owner >= FleetCount || !s.FleetUsed(owner) || s.fleet(owner)[1] != sh[0] || int(s.fleet(owner)[6]) != region {
			return nil, fmt.Errorf("戰場有效池不屬於此星域艦隊")
		}
		f := s.fleet(owner)
		found := int(binary.LittleEndian.Uint16(f[0x38:])) == pool && sh[4] == 7
		for i := 0; i < min(int(f[15]), fleetSlotN); i++ {
			found = found || int(binary.LittleEndian.Uint16(f[16+4*i:])) == pool
		}
		if !found {
			return nil, fmt.Errorf("戰場有效池沒有正式編制引用")
		}
	}
	// 每次戰損重算全陣營費用，連不參戰的艦隊也須可解析。
	for i := 0; i < FleetCount; i++ {
		if !s.FleetUsed(i) {
			continue
		}
		f := s.fleet(i)
		if f[1] > 1 || f[6] >= 32 || f[4] >= 82 || (f[5] != 255 && f[5] >= 82) || f[15] > fleetSlotN {
			return nil, fmt.Errorf("艦隊%d的陣營、提督或編制無效", i)
		}
		var claimed [ShipCount]bool
		for g := 0; g < int(f[15]); g++ {
			pool := int(binary.LittleEndian.Uint16(f[16+4*g:]))
			if pool >= ShipCount || claimed[pool] {
				return nil, fmt.Errorf("艦隊%d池引用越界或重複", i)
			}
			claimed[pool] = true
			sh := s.ShipAt(pool)
			kind, ok := kinds.Kind(int(f[1]), int(sh[4]))
			if sh[0] != f[1] || int(sh[shipOwner]) != i || !ok || len(kind.Words) != 13 {
				return nil, fmt.Errorf("艦隊%d艦種或池身分不同", i)
			}
			if int(f[6]) == region && int(f[1]) == winner {
				if binary.LittleEndian.Uint16(sh[10:]) == 0 || binary.LittleEndian.Uint16(sh[8:]) > 4000 || int(f[19+4*g]) >= len(slots.Slots) {
					return nil, fmt.Errorf("勝方滿編、艘數或隊形無效")
				}
			}
		}
		if int(f[6]) == region && int(f[1]) == winner {
			pool := int(binary.LittleEndian.Uint16(f[0x38:]))
			if pool >= ShipCount || s.Ship[pool*ShipRecord] != f[1] || s.Ship[pool*ShipRecord+4] != 7 || int(s.Ship[pool*ShipRecord+shipOwner]) != i {
				return nil, fmt.Errorf("勝方索敵池身分無效")
			}
		}
	}
	c.fleets = s.Fleets(nil)
	for i := range c.fleets {
		f := &c.fleets[i]
		if f.Region != region {
			continue
		}
		if f.Faction != winner {
			c.place.Others = append(c.place.Others, f)
			continue
		}
		rec := s.fleet(f.Index)
		c.attackers = append(c.attackers, StrategicFleet{Fleet: f, Ability40: int(rec[0x40]), Ability41: int(rec[0x41]), Command: int(s.admiral(f.Admiral)[0xb])})
	}
	if len(c.attackers) == 0 {
		return nil, fmt.Errorf("勝方占領沒有艦隊")
	}
	return c, nil
}

// Step 執行至一個占領報告或完成；等待時不再呼叫本函式。
// 報告返回後再呼叫，才增加上限功績並繼續回復。
func (c *WinnerOccupation) Step() *WinnerCapture {
	if c.done {
		return nil
	}
	if c.pending != nil {
		p := c.planets[c.pending.Planet*PlanetRecord:]
		for _, sf := range c.attackers {
			if sf.Fleet.Index == c.pending.Fleet {
				c.merit(sf.Fleet, int(binary.LittleEndian.Uint16(p[planetDefenceMax:])))
				break
			}
		}
		c.pending = nil
		c.passes = 3
		c.advanceRound()
	}
	for !c.done {
		switch {
		case c.passes > 0:
			c.passes--
			c.advanceRound()
		case int(c.planets[c.planet*PlanetRecord+planetOwner]) == c.winner:
			c.planet++
			c.done = c.planet >= len(c.sites)
		default:
			if event := c.attackRound(); event != nil {
				c.pending = event
				c.syncWork()
				return &WinnerCapture{Fleet: event.Fleet, Planet: event.Planet}
			}
			c.advanceRound()
		}
	}
	c.syncWork()
	return nil
}

func (c *WinnerOccupation) Done() bool                     { return c.done }
func (c *WinnerOccupation) Result() StrategicAssaultResult { return c.result }
func (c *WinnerOccupation) advanceRound() {
	c.round++
	recoverPlanetsOneStep(c.planets)
	c.done = c.round >= strategicRounds
}
func (c *WinnerOccupation) syncWork() {
	w, _ := ReadWork(c.slot.Work)
	c.slot.Upkeep, c.slot.StrategicScore = w.Upkeep, w.StrategicScore
}
func (c *WinnerOccupation) score(faction, n int) {
	c.result.Score[faction] += n
	for _, off := range []int{0xd0, workStrategicScore} {
		o := off + 4*faction
		binary.LittleEndian.PutUint32(c.slot.Work[o:], binary.LittleEndian.Uint32(c.slot.Work[o:])+uint32(n))
	}
}
func (c *WinnerOccupation) merit(f *Fleet, n int) {
	addMerit(&c.result, f, n)
	for _, a := range []struct{ index, n int }{{f.Admiral, n}, {f.Adjutant, n >> 1}} {
		if a.index < 0 {
			continue
		}
		r := c.slot.admiral(a.index)
		binary.LittleEndian.PutUint32(r[2:], binary.LittleEndian.Uint32(r[2:])+uint32(a.n))
	}
}
func (c *WinnerOccupation) attackRound() *WinnerCapture {
	p, site := c.planets[c.planet*PlanetRecord:(c.planet+1)*PlanetRecord], c.sites[c.planet]
	facings := make([]int, len(c.attackers))
	for i, sf := range c.attackers {
		f := sf.Fleet
		f.Target, f.Action = c.planet, 1
		facings[i] = placeStrategicFleet(f, site, c.attackers, &c.place)
		c.placeRaw(f, facings[i])
	}
	for i, sf := range c.attackers {
		f := sf.Fleet
		if f.Flags&fleetFlagLowMorale != 0 {
			continue
		}
		d := FixedDist(f.X, f.Y, site.X, site.Y)
		a := strategicAttack(sf)
		base := assaultBase(a, d)
		amt := min(base>>1+c.rnd.Next()%(base/3), int(binary.LittleEndian.Uint16(p[planetDefence:])))
		defence := int(binary.LittleEndian.Uint16(p[planetDefence:])) - amt
		binary.LittleEndian.PutUint16(p[planetDefence:], uint16(defence))
		c.result.Hits++
		c.score(c.winner, amt/ScoreDivisor)
		c.merit(f, amt/2)
		if a >= defence {
			p[planetOwner] = byte(c.winner)
			clear(p[planetStocks:planetMissiles])
			binary.LittleEndian.PutUint32(p[planetMissiles:], binary.LittleEndian.Uint32(p[planetMissiles:])>>1)
			binary.LittleEndian.PutUint32(p[planetFighters:], 0)
			c.score(c.winner, int(binary.LittleEndian.Uint16(p[planetDefenceMax:]))/captureScoreDiv)
			c.result.Captured = append(c.result.Captured, c.planet)
			c.result.CapturedBy = append(c.result.CapturedBy, f)
			return &WinnerCapture{Fleet: f.Index, Planet: c.planet}
		}
		div := 10
		if p[planetFortress] != 0 {
			div = 20
		}
		base = (c.guard(p) + defence) & 0xffff
		base = assaultBase(base/div, d)
		ang := AtanTable256(site.X-f.X, site.Y-f.Y)
		base = base * assaultDirTable[Direction8(absInt(ang-facings[i]))] / 100
		def := sf.Ability40 * sf.Ability41 / 100
		if def <= 0 {
			continue
		}
		amount := (base*100/def)&0xffff + c.rnd.Next()%100
		strategicSpreadWithLoss(sf, amount, func(i, n int) int { return c.loss(sf, i, n) })
	}
	return nil
}

func (c *WinnerOccupation) guard(p []byte) int {
	total := 0
	radius := FixedFromInt(PlanetQueryRadius(int(binary.LittleEndian.Uint16(p[0x14:]))))
	for _, pool := range c.poolIDs {
		sh := c.slot.ShipAt(pool)
		if sh[0] != p[planetOwner] || sh[0x1e]&0x80 != 0 || sh[4] > 1 {
			continue
		}
		if FixedDist(int(binary.LittleEndian.Uint16(p[0x46:])), int(binary.LittleEndian.Uint16(p[0x48:])), int(binary.LittleEndian.Uint16(sh[0x34:])), int(binary.LittleEndian.Uint16(sh[0x36:]))) > radius {
			continue
		}
		total = min(65535, total+int(binary.LittleEndian.Uint16(sh[8:])))
	}
	return total / 100
}

func (c *WinnerOccupation) placeRaw(f *Fleet, facing int) {
	r := c.slot.fleet(f.Index)
	r[7], r[0x48], r[0x3a], r[0x3b], r[0x5a], r[0x5b] = byte(c.planet), 1, 255, 0, 255, 255
	for _, o := range []int{0x4a, 0x4b, 0x4c} {
		r[o] = byte(facing)
	}
	for _, o := range []int{0x5c, 0x64} {
		binary.LittleEndian.PutUint16(r[o:], uint16(f.X))
		binary.LittleEndian.PutUint16(r[o+2:], uint16(f.Y))
	}
	writePosition := func(sh []byte, tx, ty, x, y int, mode byte) {
		for _, v := range []struct{ o, n int }{{0x28, tx}, {0x2a, ty}, {0x34, x}, {0x36, y}} {
			binary.LittleEndian.PutUint16(sh[v.o:], uint16(v.n))
		}
		sh[0x1e], sh[0x16], sh[0x17], sh[0x19] = mode, 0, 0, 0x3f
	}
	for i := range f.Composition {
		g := &f.Composition[i]
		sh := c.slot.ShipAt(g.Pool)
		tx, ty := c.place.Slots.Station(f.X, f.Y, facing, g.Slot3)
		writePosition(sh, tx, ty, g.X, g.Y, 0x7f)
		if c.slot.Stellar[f.Region*RegionRecord+0x12] != 2 {
			sh[0x17], sh[0x18] = 3, 2
		}
		g.Detail = int(sh[0x17])
	}
	writePosition(c.slot.ShipAt(f.Scout), f.X, f.Y, f.X, f.Y, 255)
}

func (c *WinnerOccupation) loss(sf StrategicFleet, group, amount int) int {
	f := sf.Fleet
	g := &f.Composition[group]
	if g.Kind == ShipKindStandalone {
		return g.Ships
	}
	old := g.Ships
	lost := max(0, min(amount, old))
	g.Ships -= lost
	g.RepairCap = (g.RepairCap - (lost*lost/g.Full)&0xffff) & 0xffff
	sh := c.slot.ShipAt(g.Pool)
	binary.LittleEndian.PutUint16(sh[8:], uint16(g.Ships))
	binary.LittleEndian.PutUint16(sh[12:], uint16(g.RepairCap))
	kind, _ := c.kinds.Kind(f.Faction, g.Kind)
	if old > 0 {
		if g.Kind == 5 || kind.Byte(kindMissileFlag) != 255 {
			g.Missiles = g.Missiles / old * g.Ships
			binary.LittleEndian.PutUint32(sh[14:], uint32(g.Missiles))
		}
		if g.Kind == 5 || kind.Byte(kindFighterFlag) != 255 {
			g.Fighters = g.Fighters / old * g.Ships
			binary.LittleEndian.PutUint32(sh[18:], uint32(g.Fighters))
		}
	}
	strategicMorale(sf)
	for i := range f.Composition {
		g := &f.Composition[i]
		if g.Ships == 0 || g.Full == 0 {
			continue
		}
		t := min((g.Full-g.Ships)*(200-sf.Command)/100, g.Full)
		g.Morale = (g.Full - t) * 100 / g.Full & 255
		c.slot.ShipAt(g.Pool)[6] = byte(g.Morale)
	}
	c.slot.fleet(f.Index)[0], c.slot.fleet(f.Index)[0x3c] = byte(f.Flags), byte(f.Morale)
	var sum uint32
	for i := range c.fleets {
		ff := &c.fleets[i]
		if ff.Faction != f.Faction {
			continue
		}
		u := FleetUpkeep(c.kinds, ff.Faction, ff.Composition, ff.Operations)
		o := FleetOtherCost(c.kinds, ff.Faction, ff.Composition, ff.Operations)
		r := c.slot.fleet(ff.Index)
		binary.LittleEndian.PutUint16(r[10:], uint16(u))
		binary.LittleEndian.PutUint16(r[12:], uint16(o))
		sum += uint32(u)
	}
	binary.LittleEndian.PutUint32(c.slot.Work[workUpkeep+4*f.Faction:], sum)
	c.score(1-f.Faction, lost/ScoreDivisor)
	return lost
}
