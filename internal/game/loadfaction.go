package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 088 READY子契約：042F:0006固定位置，繪製／命中共用CommandMenuSpec。
const loadFactionLeft, loadFactionTop = 289, 181

func (g *Game) loadFactionRow(x, y int) int {
	s := g.res.LoadFactionMenu
	if s == nil {
		return -1
	}
	return s.CommandMenuRowAt(screen.CommandMenuLoadFaction, loadFactionLeft, loadFactionTop,
		x, y, s.ItemCount(screen.CommandMenuLoadFaction))
}

func (g *Game) chooseLoadFaction(x, y int) {
	mode := g.loadFactionRow(x, y)
	if mode < 0 || mode > 2 {
		return
	}
	oldFaction, oldVersus := g.faction, g.versus
	g.faction, g.versus = mode, mode == 2
	if g.versus {
		g.faction = 0
	}
	if !g.loadSlotFromDisk(g.pendingLoadSlot) {
		g.faction, g.versus = oldFaction, oldVersus
		return
	}
	// 原始九bytes，042F:009F..0111；保留其他WORK欄位。
	control := [3][9]byte{
		{0, 0, 1, 0, 1, 0, 1, 0, 31},
		{1, 0, 1, 0, 1, 1, 0, 31, 0},
		{2, 0, 0, 0, 0, 0, 1, 0, 31},
	}
	copy(g.mutableWork()[0x0C:0x15], control[mode][:])
	if g.versus {
		for f := 0; f < 2; f++ {
			g.sides[f] = side{
				pct:      g.slot.BuildPct[f],
				treasury: rules.Treasury{Budget: g.slot.BudgetStart[f], ShipPaid: g.slot.ShipPaid[f]},
			}
		}
		g.enterSide(0)
		g.versusLoadedRound = true
		g.state = StateVersusInput
	}
}

func (g *Game) drawLoadFaction(fb *gfx.Framebuffer) {
	if g.state != StateLoadFaction {
		return
	}
	if s := g.res.SlotList; s != nil {
		// 3667選单期間保留所選存檔列，原版Top-2的20px紅條。
		fb.FillRect(s.Save.Bar.X, s.Top-2+g.pendingLoadSlot*s.Height, s.Save.Bar.W, s.Height, s.Save.Bar.Colour)
		s.DrawRow(fb, g.res.Font, g.res.SaveIndex, rules.Epoch, g.pendingLoadSlot)
	}
	if s := g.res.LoadFactionMenu; s != nil {
		screen.DrawCommandMenu(s, fb, g.res.Font, g.res.FrameTiles, screen.CommandMenuLoadFaction,
			nil, loadFactionLeft, loadFactionTop, g.loadFactionRow(g.mouseX, g.mouseY))
	}
}
