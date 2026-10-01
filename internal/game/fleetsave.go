package game

// syncFleetTables 在原始表消費端同步現況（規格 067），不從底稿覆蓋活的局面。
func (g *Game) syncFleetTables() error {
	if g.slot == nil {
		return nil
	}
	g.slot.Admiral, g.slot.Stellar, g.slot.Planets = g.res.Admiral, g.res.Stellar, g.res.Planets
	var mobility []int
	if g.res.EditorTables != nil {
		mobility = g.res.EditorTables.Mobility
	}
	fleets, err := g.slot.SyncFleets(g.fleets, g.res.ShipKinds, mobility)
	if err != nil {
		return err
	}
	g.fleets = fleets
	return nil
}
