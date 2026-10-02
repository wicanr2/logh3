package rules

// 第一階段：規格079，原版1930:019D／1AAB:006D。清單包含未放出的索敵隊。
// 原版來源是持續遠指標；以池身份表示，與未設定來源分開。
type PreparationSource struct {
	Pool  int
	Known bool
}

type PreparationSources struct {
	Supply PreparationSource
	Repair PreparationSource
}

type PreparationUnit struct {
	Pool int
	Unit *BattleUnit
}

// PreparationResult 是一次已完成的正數資源轉移。行星轉移的From為nil，無艦艇結果面板。
type PreparationResult struct {
	From, To *BattleUnit
	Repair   bool
}

// BattlePreparation 逐次完成資源轉移，讓正式顯示層能呈現結果後自動繼續。
// 不消費亂數，不在這裡執行士氣報告或回復防衛力。
type BattlePreparation struct {
	kinds      *ShipKindTable
	units      []PreparationUnit
	planets    []BattlePlanet
	sources    *PreparationSources
	poolFleet  func(int) int
	phase      int
	cursor     int
	candidates []int
	target     int
	started    bool
}

func NewBattlePreparation(kinds *ShipKindTable, units []PreparationUnit, planets []BattlePlanet,
	sources *PreparationSources, poolFleet func(int) int) *BattlePreparation {
	if sources == nil {
		sources = &PreparationSources{}
	}
	return &BattlePreparation{kinds: kinds, units: units, planets: planets, sources: sources, poolFleet: poolFleet}
}

// SortPreparationCandidates 保留原版梳狀排序的交換順序，包含平手的遠距交換。
// 原版3C6A:000D及1930:0510／1AAB:04BD；不能換成穩定排序。
func SortPreparationCandidates(list []int, unit func(int) *BattleUnit, sourceFleet func() int) {
	if len(list) <= 1 {
		return
	}
	gap := len(list)
	for {
		gap = gap * 10 / 13
		if gap == 0 {
			gap = 1
		}
		if gap == 9 || gap == 10 {
			gap = 11
		}
		swapped := false
		for i := 0; i+gap < len(list); i++ {
			a, b := unit(list[i]), unit(list[i+gap])
			cmp := 0
			if a.Fleet == b.Fleet {
				cmp = int(byte(a.Squadron)) - int(byte(b.Squadron))
			} else if f := sourceFleet(); a.Fleet == f {
				cmp = -1
			} else if b.Fleet == f {
				cmp = 1
			}
			if cmp > 0 {
				list[i], list[i+gap] = list[i+gap], list[i]
				swapped = true
			}
		}
		if gap == 1 && !swapped {
			return
		}
	}
}

func (p *BattlePreparation) sourceFleet(repair bool) int {
	s := p.sources.Supply
	if repair {
		s = p.sources.Repair
	}
	if !s.Known {
		return 0 // 固定oracle的零遠指標讀實體byte1=0；不是池0。
	}
	if p.poolFleet == nil {
		return 0xFF
	}
	return p.poolFleet(s.Pool)
}

func (p *BattlePreparation) collect(x, y, faction int, initial, fine int64, supply bool) []int {
	var out []int
	for i, v := range p.units {
		u := v.Unit
		if u == nil || u.Faction != faction || u.Out || u.Kind < 0 || u.Kind > RepairKind {
			continue
		}
		if supply && (u.Kind == SupplyKind || u.Kind == RepairKind) {
			continue
		}
		if !supply && p.phase == 2 && u.Kind == RepairKind {
			continue
		}
		d := Distance16(x, y, u.X, u.Y) // SHIP+34/+36：現在位置。
		if d <= initial && d <= fine {
			out = append(out, i)
		}
	}
	return out
}

// Next 回一項正數轉移，或false表示資源階段完成。來源在找候選前更新。
func (p *BattlePreparation) Next() (PreparationResult, bool) {
	for p.phase < 4 {
		if p.started {
			for p.target < len(p.candidates) {
				to := p.units[p.candidates[p.target]].Unit
				p.target++
				r := PreparationResult{To: to, Repair: p.phase >= 2}
				n := 0
				if p.phase == 0 || p.phase == 2 {
					r.From = p.units[p.cursor].Unit
					if r.Repair {
						n = Repair(r.From, to)
					} else {
						m, f := Supply(p.kinds, r.From, to)
						n = m + f
					}
				} else {
					stock := &p.planets[p.cursor].Stock
					if r.Repair {
						n = PlanetReplenish(stock, to)
					} else {
						m, f := PlanetSupply(p.kinds, stock, to)
						n = m + f
					}
				}
				if n > 0 {
					return r, true
				}
			}
			p.started = false
			p.cursor++
		}
		count := len(p.units)
		if p.phase == 1 || p.phase == 3 {
			count = len(p.planets)
		}
		if p.cursor >= count {
			p.phase++
			p.cursor = 0
			continue
		}
		p.started, p.target, p.candidates = true, 0, nil
		if p.phase == 0 || p.phase == 2 {
			v := p.units[p.cursor]
			u := v.Unit
			if u == nil {
				continue
			}
			s := PreparationSource{Pool: v.Pool, Known: v.Pool >= 0}
			kind := SupplyKind
			if p.phase == 0 {
				p.sources.Supply = s // 每筆都更新，包含出局／未放出索敵隊。
			} else {
				kind = RepairKind
				if u.Kind == kind {
					p.sources.Repair = s
				}
			}
			if u.Kind != kind || p.kinds == nil {
				continue
			}
			fine := SupplyListRange(p.kinds, u.Faction, kind)
			initial := fine >> 16 << 16
			if p.phase == 2 {
				initial = int64(kindSpeed(p.kinds, u.Faction, kind)*6) << 16
			}
			p.candidates = p.collect(u.X, u.Y, u.Faction, initial, fine, p.phase == 0)
		} else {
			planet := p.planets[p.cursor]
			if planet.Owner != 0 && planet.Owner != 1 {
				continue
			}
			r := int64(PlanetQueryRadius(planet.Defence)) << 16
			p.candidates = p.collect(planet.X, planet.Y, planet.Owner, r, r, p.phase == 1)
		}
		SortPreparationCandidates(p.candidates,
			func(i int) *BattleUnit { return p.units[i].Unit },
			func() int { return p.sourceFleet(p.phase >= 2) })
	}
	return PreparationResult{}, false
}

// RecoverBattleDefence 是原版181B:0083：乘積低16位除100，再以WORD加法回繞。
func RecoverBattleDefence(planets []BattlePlanet) {
	for i := range planets {
		p := &planets[i]
		if p.Defence >= p.DefenceMax {
			continue
		}
		inc := int(uint16(p.DefenceMax*p.Recover)) / 100
		p.Defence = int(uint16(p.Defence + inc))
		if p.Defence > p.DefenceMax {
			p.Defence = p.DefenceMax
		}
	}
}
