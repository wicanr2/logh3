package game

import "github.com/wicanr2/logh3/internal/rules"

// placeArrivedFleet在移動過場返回後、Arrive命令收尾前配置戰術資料（規格088）。
func (g *Game) placeArrivedFleet(i, from int) {
	if i < 0 || i >= len(g.fleets) || g.slot == nil || g.res == nil || g.res.Terrain == nil {
		return
	}
	f := &g.fleets[i]
	cells := g.battleCells(f.Region)
	if len(cells) == 0 {
		return
	}
	place := &rules.StrategicPlacement{Slots: g.res.Formation,
		Passable: func(x, y int) bool { return g.res.Terrain.Passable(cells, x, y) }}
	for j := range g.fleets {
		if j != i && g.fleets[j].Region == f.Region {
			place.Others = append(place.Others, &g.fleets[j])
		}
	}
	g.slot.PlaceFleetArrival(f, from, g.res.Entries, place)
}
