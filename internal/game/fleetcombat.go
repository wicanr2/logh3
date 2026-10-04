package game

import (
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

type fleetCombatStage uint8

const (
	fleetCommandHeader fleetCombatStage = iota
	fleetCommandTarget
	fleetCommandMenu
	fleetCommandAutoWait
	fleetCombatScene
)

// 102：只保存生命週期；命令與戰損仍存在原SHIP／FLEET／ADML。
type fleetCombatState struct {
	active                 bool
	stage                  fleetCombatStage
	fleets                 []int
	fleetAt                int
	units                  []int
	unitAt                 int
	current                int
	automatic, headerShown bool
	count, aggregate       byte
	candidates             []rules.FleetCombatCandidate
	target                 int
	menu                   screen.CommandMenuKind
	menuX, menuY           int
	header                 []byte
	ticks                  int
	rem                    time.Duration
	commands               []int
	commandAt              int
	scene                  *fleetCombatSceneState
	err                    error
}

func fleetCommandRosterOrder(fleets []rules.BattleFleet) []int {
	order := make([]int, len(fleets))
	for i := range order {
		order[i] = i
	}
	gap := len(order)
	for {
		gap = max(1, gap*10/13)
		if gap == 9 || gap == 10 {
			gap = 11
		}
		swapped := false
		for i := 0; i+gap < len(order); i++ {
			left, right := fleets[order[i]], fleets[order[i+gap]]
			// 1DD6:0135..013D的原版比較錯誤；同陣營不產生正值。
			if left.Faction > right.Faction {
				order[i], order[i+gap] = order[i+gap], order[i]
				swapped = true
			}
		}
		if gap == 1 && !swapped {
			break
		}
	}
	ids := make([]int, len(order))
	for i, j := range order {
		ids[i] = fleets[j].Index
	}
	return ids
}

func (g *Game) fleetCombatUnits(fleet int) []int {
	var list []int
	for i, u := range g.battleUnits {
		if u.Fleet == fleet && u.Squadron >= 0 {
			list = append(list, i)
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return g.battleUnits[list[i]].Squadron < g.battleUnits[list[j]].Squadron })
	return list
}

func (g *Game) fleetCombatRawFleet(index int) ([]byte, error) {
	if g.slot == nil || index < 0 || (index+1)*rules.FleetRecord > len(g.slot.Fleet) {
		return nil, fmt.Errorf("交戰艦隊原表缺失")
	}
	return g.slot.Fleet[index*rules.FleetRecord : (index+1)*rules.FleetRecord], nil
}

func (g *Game) fleetCombatFail(err error) {
	g.fleetCombat.err = err
	g.saveErr = err
}

// beginFleetCombat是第③步的正式續接，沒有命令則直接進第⑤步。
func (g *Game) beginFleetCombat() {
	g.fleetCombat = fleetCombatState{active: true, current: -1, target: -1}
	g.battleHoverOn, g.battleMenu, g.battleCmd, g.battleScrollHeld = false, false, false, false
	g.battleLosses = nil
	var participants []rules.BattleFleet
	cpu, passive := g.passiveComputer()
	for _, f := range g.battleFleets {
		r, err := g.fleetCombatRawFleet(f.Index)
		if err != nil {
			g.fleetCombatFail(err)
			return
		}
		if r[4] == 255 {
			continue
		}
		for _, i := range g.fleetCombatUnits(f.Index) {
			r, err := g.fleetCombatRecord(i)
			if err != nil {
				g.fleetCombatFail(err)
				return
			}
			r[0x2e] = 255
		}
		if f.Action&15 == 1 && !f.Flag0 && !(passive && f.Faction == cpu) {
			participants = append(participants, f)
		}
	}
	g.fleetCombat.fleets = fleetCommandRosterOrder(participants)
	g.continueFleetCommands()
}

func (g *Game) initFleetAnticipated() error {
	for _, u := range g.battleUnits {
		pool := g.battleUnitPool(u)
		if pool < 0 || (pool+1)*rules.ShipRecord > len(g.slot.Ship) {
			return fmt.Errorf("交戰物件原池缺失")
		}
		r := g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord]
		binary.LittleEndian.PutUint16(r[0x3e:], uint16(u.Ships))
	}
	return nil
}

func (g *Game) continueFleetCommands() {
	c := &g.fleetCombat
	for c.active && c.err == nil {
		if c.fleetAt >= len(c.fleets) {
			if c.aggregate == 0 {
				g.finishFleetCombat()
				return
			}
			g.beginFleetCombatScenes()
			return
		}
		if c.units == nil {
			f := g.battleFleet(c.fleets[c.fleetAt])
			if err := g.initFleetAnticipated(); err != nil {
				g.fleetCombatFail(err)
				return
			}
			c.units = g.fleetCombatUnits(f.Index)
			c.unitAt = 0
			c.count = 0
			c.headerShown = false
			c.automatic = g.autoBattle || !g.humanSeat(f.Faction)
		}
		if c.unitAt >= len(c.units) {
			c.aggregate |= c.count
			c.fleetAt++
			c.units = nil
			continue
		}
		i := c.units[c.unitAt]
		u := g.battleUnits[i]
		list, err := g.fleetCombatCandidates(i)
		if err != nil {
			g.fleetCombatFail(err)
			return
		}
		if u.Out || u.Action != 0x7f || len(list) == 0 {
			c.unitAt++
			continue
		}
		c.current, c.candidates, c.target = i, list, -1
		f := g.battleFleet(u.Fleet)
		if !c.headerShown && g.humanSeat(u.Faction) {
			if g.res.FleetCommand == nil || g.res.TurnWindow == nil {
				g.fleetCombatFail(fmt.Errorf("缺交戰命令格式"))
				return
			}
			c.headerShown = true
			c.stage = fleetCommandHeader
			c.ticks = 0
			c.rem = 0
			c.header = g.res.FleetCommand.Header(g.factionNameColoured(u.Faction), g.res.TurnWindow.FleetNumber(f.Number))
			return
		}
		g.battleView = g.battleViewAt(f.X, f.Y)
		c.count++
		if c.automatic {
			g.applyFleetAutoCommand()
			return
		}
		c.stage = fleetCommandTarget
		g.fleetCombatMouse(g.mouseX, g.mouseY)
		return
	}
}

func (g *Game) applyFleetAutoCommand() {
	c := &g.fleetCombat
	if _, err := g.selectFleetAutoCombatCommand(c.current); err != nil {
		g.fleetCombatFail(err)
		return
	}
	c.stage = fleetCommandAutoWait
	c.ticks = 0
	c.rem = 0
}

func (g *Game) finishFleetUnitCommand() {
	c := &g.fleetCombat
	c.unitAt++
	c.current = -1
	c.target = -1
	c.candidates = nil
	g.continueFleetCommands()
}

func (g *Game) fleetCommandHit(x, y int) []int {
	s := g.res.Battle
	if s == nil {
		return nil
	}
	wx, wy := x-s.Field.X+g.battleView.Col*s.Field.TileW, y-s.Field.Y+g.battleView.Row*s.Field.TileH
	if x < s.Field.X || y < s.Field.Y || x >= s.Field.X+s.Field.Cols*s.Field.TileW || y >= s.Field.Y+s.Field.Rows*s.Field.TileH {
		return nil
	}
	var found []int
	for _, c := range g.fleetCombat.candidates {
		u := g.battleUnits[c.Target]
		if wx >= u.X && wx < u.X+16 && wy >= u.Y && wy < u.Y+16 {
			found = append(found, c.Target)
		}
	}
	return found
}

func (g *Game) fleetCombatMouse(x, y int) {
	c := &g.fleetCombat
	if !c.active || c.stage != fleetCommandTarget {
		return
	}
	hits := g.fleetCommandHit(x, y)
	for _, i := range hits {
		if i == c.target {
			return
		}
	}
	c.target = -1
	if len(hits) > 0 {
		c.target = hits[0]
	}
}

func (g *Game) fleetCombatClick(x, y int, right bool) {
	c := &g.fleetCombat
	if c.err != nil {
		return
	}
	switch c.stage {
	case fleetCommandHeader:
		c.header = nil
		c.ticks = 0
		g.continueFleetCommands()
	case fleetCommandAutoWait:
		g.finishFleetUnitCommand()
	case fleetCombatScene:
		g.fleetSceneClick()
	case fleetCommandMenu:
		row := -1
		if !right {
			row = g.res.CommandMenus.CommandMenuRowAt(c.menu, c.menuX, c.menuY, x, y, g.res.CommandMenus.ItemCount(c.menu))
		}
		if c.menu == screen.CommandMenuCombatChoice {
			if row == 0 {
				c.automatic = true
				g.applyFleetAutoCommand()
				return
			}
			if row == 1 {
				if err := g.commitFleetCombatCommand(c.current, -1, 3); err != nil {
					g.fleetCombatFail(err)
					return
				}
				g.finishFleetUnitCommand()
				return
			}
		} else if row >= 0 && row < 3 {
			if err := g.commitFleetCombatCommand(c.current, c.target, row); err == nil {
				g.finishFleetUnitCommand()
				return
			}
		}
		c.stage = fleetCommandTarget
		c.target = -1
		g.fleetCombatMouse(g.mouseX, g.mouseY)
	case fleetCommandTarget:
		hits := g.fleetCommandHit(x, y)
		if right {
			if len(hits) > 1 {
				n := 0
				for j, i := range hits {
					if i == c.target {
						n = (j + 1) % len(hits)
					}
				}
				c.target = hits[n]
			}
			return
		}
		c.menu = screen.CommandMenuCombatChoice
		if len(hits) > 0 {
			selected := hits[0]
			for _, i := range hits {
				if i == c.target {
					selected = i
				}
			}
			c.target = selected
			c.menu = screen.CommandMenuCombatWeapon
		}
		if g.res.CommandMenus == nil || g.res.FleetCommand == nil {
			g.fleetCombatFail(fmt.Errorf("缺交戰命令選單"))
			return
		}
		c.menuX, c.menuY = g.res.FleetCommand.MenuAt(x, y, c.menu == screen.CommandMenuCombatChoice)
		c.stage = fleetCommandMenu
	}
}

func (g *Game) fleetCombatTick(d time.Duration) {
	c := &g.fleetCombat
	if c.err != nil {
		return
	}
	c.rem += d
	for c.active && c.rem >= g.engagementTickDuration() {
		c.rem -= g.engagementTickDuration()
		switch c.stage {
		case fleetCommandHeader:
			c.ticks++
			if c.ticks >= g.display.MsgSpeed {
				c.header = nil
				g.continueFleetCommands()
			}
		case fleetCommandAutoWait:
			c.ticks++
			if c.ticks >= 6 {
				g.finishFleetUnitCommand()
			}
		case fleetCombatScene:
			g.advanceSoundEffects(g.engagementTickDuration())
			g.fleetSceneTick()
		default:
			c.rem = 0
			return
		}
		if c.err != nil {
			return
		}
	}
}

func (g *Game) drawFleetCombatCommand(fb *gfx.Framebuffer) {
	c := &g.fleetCombat
	if !c.active {
		return
	}
	if c.stage == fleetCommandHeader {
		screen.DrawFleetCommandHeader(fb, g.res.Font, g.res.FrameTiles, c.header)
		return
	}
	if c.current >= 0 && g.res.UnitPanel != nil {
		if v, ok := g.unitPanelFor(c.current); ok {
			if g.res.FleetCommandDisplay != nil {
				var enemy *screen.UnitPanelView
				detail, distance := 0, 0
				var faction, kind []byte
				if c.target >= 0 {
					if e, ok := g.unitPanelFor(c.target); ok {
						enemy = &e
						u := g.battleUnits[c.target]
						a := g.battleUnits[c.current]
						detail = u.Detail
						distance = int(rules.FixedDist(a.X, a.Y, u.X, u.Y) >> 16)
						faction = g.factionNameColoured(u.Faction)
						if u.Kind >= 0 && u.Kind < len(g.res.ShipKindNames) {
							kind = g.res.ShipKindNames[u.Kind]
						}
					}
				}
				g.res.FleetCommandDisplay.DrawPanel(g.res.UnitPanel, fb, g.res.Font, g.res.FrameTiles, v, enemy, detail, distance, faction, kind)
			} else {
				screen.DrawUnitPanelSummary(g.res.UnitPanel, fb, g.res.Font, v)
			}
		}
	}
	if c.stage != fleetCommandMenu {
		return
	}
	u := g.battleUnits[c.current]
	k, _ := g.res.ShipKinds.Kind(u.Faction, u.Kind)
	shots := [3]int{k.Byte(11) * u.Ships, min(k.Byte(15)*u.Ships, u.Missiles), min(k.Byte(19)*u.Ships, u.Fighters)}
	on := []bool{true, true}
	if c.menu == screen.CommandMenuCombatWeapon {
		on = []bool{false, false, false}
		for _, v := range c.candidates {
			if v.Target == c.target {
				for w := range on {
					on[w] = v.Mask&(1<<w) != 0
				}
			}
		}
	}
	labels := g.res.FleetCommand.Labels(shots, c.menu == screen.CommandMenuCombatChoice)
	row := g.res.CommandMenus.CommandMenuRowAt(c.menu, c.menuX, c.menuY, g.mouseX, g.mouseY, len(labels))
	screen.DrawCommandMenuLabels(g.res.CommandMenus, fb, g.res.Font, g.res.FrameTiles, c.menu, labels, on, c.menuX, c.menuY, row)
}

func (g *Game) drawFleetCombatRange(fb *gfx.Framebuffer, view screen.BattleView) {
	c := &g.fleetCombat
	if !c.active || c.current < 0 || c.stage == fleetCommandHeader {
		return
	}
	u := g.battleUnits[c.current]
	screen.DrawBattleSpottingHighlight(g.res.Battle, fb, g.res.Sprites, [2]int{u.X, u.Y}, view)
	if c.stage != fleetCommandTarget {
		return
	}
	f := g.battleFleet(u.Fleet)
	if f == nil {
		return
	}
	r, err := rules.FleetCombatWeaponRanges(g.res.ShipKinds, u)
	if err != nil {
		g.fleetCombatFail(err)
		return
	}
	screen.DrawFleetCombatRange(g.res.Battle, fb, u.X, u.Y, f.Facing, r, view)
}

func (g *Game) finishFleetCombat() {
	g.fleetCombat.active = false
	g.fleetCombat.scene = nil
	g.state = StateBattle
	g.advanceBattleTurnAfterFleetCombat()
}
