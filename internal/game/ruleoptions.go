package game

// RuleOptions are remake preferences, separate from the original six save tables.
type RuleOptions struct {
	FlagshipTransfer bool `json:"flagship_transfer"`
	HeroMovement     bool `json:"hero_movement"`
}

func DefaultRuleOptions() RuleOptions { return RuleOptions{FlagshipTransfer: true, HeroMovement: true} }

func (g *Game) RuleOptions() RuleOptions {
	return RuleOptions{FlagshipTransfer: g.flagshipTransfer, HeroMovement: !g.heroMovementDisabled}
}

func (g *Game) SetRuleOptions(options RuleOptions) {
	g.flagshipTransfer = options.FlagshipTransfer
	g.heroMovementDisabled = !options.HeroMovement
}

func (g *Game) heroMovementActive(faction int) bool {
	if g.heroMovementDisabled || faction < 0 || faction > 1 {
		return false
	}
	return g.battleCommanders()[faction] != 0
}
