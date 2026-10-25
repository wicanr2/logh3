package game

// RuleOptions are remake preferences, separate from the original six save tables.
type RuleOptions struct {
	FlagshipTransfer      bool `json:"flagship_transfer"`
	HeroMovement          bool `json:"hero_movement"`
	NonBattleshipFlagship bool `json:"non_battleship_flagship"`
}

func DefaultRuleOptions() RuleOptions { return RuleOptions{FlagshipTransfer: true, HeroMovement: true} }

func (g *Game) RuleOptions() RuleOptions {
	return RuleOptions{FlagshipTransfer: g.flagshipTransfer, HeroMovement: !g.heroMovementDisabled, NonBattleshipFlagship: g.nonBattleshipFlagship}
}

func (g *Game) SetRuleOptions(options RuleOptions) {
	g.flagshipTransfer = options.FlagshipTransfer
	g.heroMovementDisabled = !options.HeroMovement
	g.nonBattleshipFlagship = options.NonBattleshipFlagship
	if g.editor != nil {
		g.editor.NonBattleshipFlagship = options.NonBattleshipFlagship
	}
}

func (g *Game) heroMovementActive(faction int) bool {
	if g.heroMovementDisabled || faction < 0 || faction > 1 {
		return false
	}
	return g.armyBonus(faction, true) > 0
}
