package game

import "github.com/wicanr2/logh3/internal/rules"

// recountRegionVisibleFleets 接回04F8:003E→2703:043B及新遊戲初始化。
// 可見數是命令入口的導出快照；移動／編成時不跟實數一起提前改。
// 證據：091，艦艇Detail由既有索敵及到達處理維護。
func (g *Game) recountRegionVisibleFleets() {
	if g.res == nil {
		return
	}
	rules.RecountRegionVisibleFleets(g.res.Stellar, g.fleets)
}

// clearRegionFleetVisibility 接回0374:0556→0AF2，在無戰鬥星域的攻擊／被動效果之前。
// 只清編制內索敵值；獨立偵察隊與其他索敵欄位不在此常式的寫入範圍。
func (g *Game) clearRegionFleetVisibility(region int) {
	for i := range g.fleets {
		f := &g.fleets[i]
		if f.Region != region {
			continue
		}
		f.Action = 4
		if g.slot != nil && f.Index >= 0 && f.Index < rules.FleetCount {
			o := f.Index*rules.FleetRecord + 0x48
			if o < len(g.slot.Fleet) {
				g.slot.Fleet[o] = 4
			}
		}
		for j := range f.Composition {
			gp := &f.Composition[j]
			gp.Detail = 0
			if g.slot != nil && gp.Pool >= 0 && gp.Pool < rules.ShipCount {
				o := gp.Pool*rules.ShipRecord + 0x17
				if o < len(g.slot.Ship) {
					g.slot.Ship[o] = 0
				}
			}
		}
	}
}
