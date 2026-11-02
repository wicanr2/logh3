package game

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/rules"
)

// ensureSlot 在新遊戲第一次需要原始表時建出 Slot：保存或進戰場（規格009、159）。
//
// 拿量到的開局 `WORK` 當底稿，其餘的表就是目前手上的。remake 沒有模型的那些欄位
// 會停在 796 年 1 月的初值。
func (g *Game) ensureSlot() error {
	if g.slot != nil {
		return nil
	}
	if len(g.res.WorkTemplate) < rules.WorkRecord || g.res.Planets == nil {
		return fmt.Errorf("新遊戲要建原始表得先給 WorkTemplate（`docs/spec/009`）")
	}
	g.slot = &rules.Slot{
		Stellar: g.res.Stellar, Planets: g.res.Planets,
		Admiral: g.res.Admiral,
		Fleet:   make([]byte, rules.FleetCount*rules.FleetRecord),
		Ship:    make([]byte, rules.ShipCount*rules.ShipRecord),
		Work:    append([]byte(nil), g.res.WorkTemplate...),
	}
	for i := range g.slot.Fleet {
		g.slot.Fleet[i] = 0xFF
	}
	for i := range g.slot.Ship {
		g.slot.Ship[i] = 0xFF
	}
	return nil
}

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
