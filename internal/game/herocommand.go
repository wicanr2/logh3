package game

import (
	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// battleCommanders每次從存活旗艦查帶隊提督，不把名將狀態寫入存檔。
func (g *Game) battleCommanders() (commanders [2]ai.Commander) {
	if g.AILevel() < 0 {
		return
	}
	for _, u := range g.battleUnits {
		if u.Squadron != 0 || u.Out || u.Ships <= 0 || u.Faction < 0 || u.Faction > 1 {
			continue
		}
		f := g.fleetByIndex(u.Fleet)
		if f == nil || f.Faction != u.Faction {
			continue
		}
		// 原資料名稱表對照：keireki.json，規格140。
		switch f.Admiral {
		case 0:
			commanders[u.Faction] = ai.Reinhard
		case 52:
			commanders[u.Faction] = ai.Yang
		}
	}
	return
}

func (g *Game) applyFormationMove(index, x, y int) {
	g.applyTacticalMove(index, x, y)
	fleet := g.battleUnits[index].Fleet
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		if u.Fleet == fleet && u.Squadron > 0 && !u.Out && u.Ships > 0 {
			u.Action = squadronNone // 原有編隊跟隨分派，索敵隊不改。
		}
	}
}

// HeroCommandObservation僅供正常視窗錄影收據讀取，不設定狀態。
func (g *Game) HeroCommandObservation() [2]string {
	c := g.battleCommanders()
	return [2]string{c[0].String(), c[1].String()}
}

type TacticalFleetObservation struct {
	Fleet   int  `json:"fleet"`
	Faction int  `json:"faction"`
	Admiral int  `json:"admiral"`
	X       int  `json:"x"`
	Y       int  `json:"y"`
	TargetX int  `json:"target_x"`
	TargetY int  `json:"target_y"`
	PerTurn int  `json:"base_budget"`
	Budget  int  `json:"remaining_budget"`
	Boost   bool `json:"command_boost"`
}

type TacticalObservation struct {
	Units            []TacticalUnitObservation         `json:"units"`
	CombatChoices    []TacticalCombatChoiceObservation `json:"combat_choices,omitempty"`
	CombatWeaponMask int                               `json:"combat_weapon_mask"`
	FleetOrder       int                               `json:"fleet_order"`
	CommandOpen      bool                              `json:"command_open"`
	CommandKind      string                            `json:"command_kind"`
	CommandPoint     [2]int                            `json:"command_point"`
	Region           int                               `json:"region"`
	Turn             int                               `json:"turn"`
	View             [2]int                            `json:"view"`
	Commanders       [2]string                         `json:"commanders"`
	Fleets           []TacticalFleetObservation        `json:"fleets"`
	CombatInput      string                            `json:"combat_input"`
	CombatPoint      [2]int                            `json:"combat_point"`
	Idle             bool                              `json:"idle"`
	MessageOpen      bool                              `json:"message_open"`
	MenuOpen         bool                              `json:"menu_open"`
	PlanetMenu       bool                              `json:"planet_menu"`
	FortressTarget   bool                              `json:"fortress_target"`
	Planets          []TacticalPlanetObservation       `json:"planets"`
	GunHits          []rules.FortressGunHit            `json:"gun_hits,omitempty"`
	ScoutTarget      bool                              `json:"scout_target"`
	MoveUnit         int                               `json:"move_unit"`
}

// Detached observations only; these fields never set a position, order or RNG.
type TacticalPlanetObservation struct {
	Index      int  `json:"index"`
	Owner      int  `json:"owner"`
	X          int  `json:"x"`
	Y          int  `json:"y"`
	Fortress   bool `json:"fortress"`
	GunState   int  `json:"gun_state"`
	GunAim     int  `json:"gun_aim"`
	DefenceMax int  `json:"defence_max"`
}

func (g *Game) TacticalObservation() *TacticalObservation {
	if g.state != StateBattle && g.state != StateEngagement {
		return nil
	}
	v := &TacticalObservation{Region: g.battleRegion, Turn: g.battleTurn, View: [2]int{g.battleView.Col, g.battleView.Row}, Commanders: g.HeroCommandObservation()}
	v.ScoutTarget, v.MoveUnit = g.battleMove && g.battleMoveRestore, g.battleMoveUnit
	v.Idle, v.MessageOpen, v.MenuOpen = g.battleInputIdle(), g.msgOn, g.battleMenu
	v.PlanetMenu, v.FortressTarget = g.PlanetMenuOpen(), g.FortressTargetPending()
	v.FleetOrder, v.CommandOpen, v.CommandKind = int(g.fleetOrder), g.battleCmd, string(g.battleCmdKind)
	if g.res.CommandMenus != nil {
		l := g.res.CommandMenus.Layout
		v.CommandPoint = [2]int{g.battleCmdX + l.HighlightDX + 8, g.battleCmdY + l.HighlightDY + 8}
	}
	for i, u := range g.battleUnits {
		v.Units = append(v.Units, TacticalUnitObservation{i, u.Fleet, u.Faction, u.Squadron, u.Kind, u.X, u.Y, u.Ships, u.Out})
	}
	for _, p := range g.battlePlanets {
		v.Planets = append(v.Planets, TacticalPlanetObservation{p.Index, p.Owner, p.X, p.Y, p.Fortress, p.GunState, p.GunAim, p.DefenceMax})
	}
	v.GunHits = append([]rules.FortressGunHit(nil), g.battleGunHits...)
	if g.fleetCombat.active {
		for _, candidate := range g.fleetCombat.candidates {
			if candidate.Target == g.fleetCombat.target {
				v.CombatWeaponMask = int(candidate.Mask)
			}
			if candidate.Target < 0 || candidate.Target >= len(g.battleUnits) || g.res.Battle == nil {
				continue
			}
			u, f := g.battleUnits[candidate.Target], g.res.Battle.Field
			x, y := f.X+u.X-g.battleView.Col*f.TileW, f.Y+u.Y-g.battleView.Row*f.TileH
			if x >= f.X && y >= f.Y && x+16 < f.X+f.Cols*f.TileW && y+16 < f.Y+f.Rows*f.TileH {
				v.CombatChoices = append(v.CombatChoices, TacticalCombatChoiceObservation{candidate.Target, int(candidate.Mask), [2]int{x + 8, y + 8}})
			}
		}
		v.CombatPoint = [2]int{240, 470}
		switch g.fleetCombat.stage {
		case fleetCommandHeader:
			v.CombatInput = "header"
		case fleetCommandTarget:
			v.CombatInput = "target"
			if len(g.fleetCombat.candidates) > 0 && g.res.Battle != nil {
				i := g.fleetCombat.candidates[0].Target
				if i >= 0 && i < len(g.battleUnits) {
					u := g.battleUnits[i]
					f := g.res.Battle.Field
					x, y := f.X+u.X-g.battleView.Col*f.TileW, f.Y+u.Y-g.battleView.Row*f.TileH
					if x >= f.X && y >= f.Y && x+16 < f.X+f.Cols*f.TileW && y+16 < f.Y+f.Rows*f.TileH {
						v.CombatPoint = [2]int{x + 8, y + 8}
					}
				}
			}
		case fleetCommandMenu:
			v.CombatInput = "menu"
			v.CombatPoint = [2]int{g.fleetCombat.menuX + 16, g.fleetCombat.menuY + 20}
		}
	}
	for _, bf := range g.battleFleets {
		f := g.fleetByIndex(bf.Index)
		if f == nil {
			continue
		}
		v.Fleets = append(v.Fleets, TacticalFleetObservation{bf.Index, bf.Faction, f.Admiral, bf.X, bf.Y, bf.DestX, bf.DestY, bf.PerTurn, bf.Budget, g.heroMovementActive(bf.Faction)})
	}
	return v
}
