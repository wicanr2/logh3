package game

import "github.com/wicanr2/logh3/internal/rules"

// placeArrivedFleet在移動過場返回後、Arrive命令收尾前配置戰術資料（規格088）。
// 新遊戲還沒有原池時只記來源星域，等第一場戰鬥建池再配置（規格159）。
func (g *Game) placeArrivedFleet(i, from int) {
	if i < 0 || i >= len(g.fleets) {
		return
	}
	if g.slot == nil {
		if g.arrivalFrom == nil {
			g.arrivalFrom = map[int]int{}
		}
		g.arrivalFrom[g.fleets[i].Index] = from
		return
	}
	g.placeFleet(i, from)
}

// placeFleet 是 `297B:08D8(艦隊, from, 所在星域)`。
//
// 原版建艦隊當下就配好池索引（`297B:0066`）；remake 延到同步才配，所以新編艦隊
// 要先同步一次，否則 `PlaceFleetAt` 會因為池索引是 -1 而什麼都不做。
func (g *Game) placeFleet(i, from int) {
	if g.slot == nil || g.res == nil || g.res.Terrain == nil {
		return
	}
	for _, gp := range g.fleets[i].Composition {
		if gp.Pool < 0 {
			if err := g.syncFleetTables(); err != nil {
				g.saveErr = err
				return
			}
			break
		}
	}
	f := &g.fleets[i]
	cells := g.battleCells(f.Region)
	if len(cells) == 0 {
		return
	}
	site, ok := g.fleetPlacementSite(f, from)
	if !ok {
		return
	}
	place := &rules.StrategicPlacement{Slots: g.res.Formation,
		Passable: func(x, y int) bool { return g.res.Terrain.Passable(cells, x, y) }}
	for j := range g.fleets {
		if j != i && g.fleets[j].Region == f.Region {
			place.Others = append(place.Others, &g.fleets[j])
		}
	}
	g.slot.PlaceFleetAt(f, site, place)
}

// fleetPlacementSite 是 `297B:0ABF` 選中心的順序：入口表（`113A:0D2C`），查不到就用
// 待命行星（`2AF5:060D` → `0DEA:1759`）。只有首都有 from 等於自己的入口，所以在其他
// 星域新編的艦隊走行星這條（規格159）。
//
// 連待命行星都沒有時，原版把 0xFF 當序號去查行星表；remake 不重現這個越界，取第一個入口。
func (g *Game) fleetPlacementSite(f *rules.Fleet, from int) (rules.PlanetSite, bool) {
	if e, ok := g.res.Entries.EntryFor(f.Region, from); ok {
		return rules.PlanetSite{X: e.X, Y: e.Y}, true
	}
	if g.res.Planets != nil {
		if k := g.res.Planets.StandbyPlanet(f.Faction, f.Region); k >= 0 {
			if sites := g.loadRegionPlanetSites(f.Region); k < len(sites) {
				return sites[k], true
			}
		}
	}
	if g.res.Entries != nil && f.Region < len(g.res.Entries.Regions) {
		if es := g.res.Entries.Regions[f.Region].Entries; len(es) > 0 {
			return rules.PlanetSite{X: es[0].X, Y: es[0].Y}, true
		}
	}
	return rules.PlanetSite{}, false
}

// fleetNeverPlaced 判斷這一支還沒有戰術座標。入口與行星都不在 (0,0)，而新編艦隊的
// 中心與各組座標同步之前都是零值。
func fleetNeverPlaced(f *rules.Fleet) bool {
	if f.X != 0 || f.Y != 0 {
		return false
	}
	for _, gp := range f.Composition {
		if gp.X != 0 || gp.Y != 0 {
			return false
		}
	}
	return true
}

// placeUnplacedFleets 補上還沒定位過的艦隊，供正常戰場入口在讀原池之前呼叫。
//
// 原版新編艦隊走 `297B:03EB(槽, 1)` → `297B:08D8(艦隊, 星域, 星域)`，行軍到達走
// `113A:0648`，兩者都在有原池之後才能在 remake 重放（規格159）。
func (g *Game) placeUnplacedFleets() {
	for i := range g.fleets {
		if !fleetNeverPlaced(&g.fleets[i]) {
			continue
		}
		index := g.fleets[i].Index
		from, ok := g.arrivalFrom[index]
		if !ok {
			from = g.fleets[i].Region
		}
		delete(g.arrivalFrom, index)
		g.placeFleet(i, from)
	}
}
