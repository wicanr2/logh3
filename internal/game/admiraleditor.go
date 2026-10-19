package game

import (
	"bytes"
	"fmt"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

type AdmiralEditorEntry struct {
	Index, Faction, Level int
	Name                  []uint16
	Label                 string
	Abilities             [13]int
	Defaults              [13]int
	Special               rules.AdmiralSpecial
}

func (g *Game) AdmiralPreferences() rules.AdmiralPreferences { return g.admiralPreferences.Clone() }

func (g *Game) AdmiralEditorEntries() []AdmiralEditorEntry {
	var out []AdmiralEditorEntry
	for i := 0; (i+1)*rules.AdmiralRecord <= len(g.res.Admiral); i++ {
		r := g.res.Admiral[i*rules.AdmiralRecord:]
		e := AdmiralEditorEntry{Index: i, Faction: int(r[1]), Level: int(r[6]), Special: g.admiralPreferences.Special(i)}
		if i < len(g.res.AdmiralNames) {
			e.Name = append([]uint16(nil), g.res.AdmiralNames[i]...)
		}
		if i < len(g.res.AdmiralLabels) {
			e.Label = g.res.AdmiralLabels[i]
		}
		for f, off := range rules.AdmiralEditableOffsets {
			e.Abilities[f] = int(r[off])
			e.Defaults[f] = e.Abilities[f]
			if (i+1)*rules.AdmiralRecord <= len(g.admiralDefaults) {
				e.Defaults[f] = int(g.admiralDefaults[i*rules.AdmiralRecord+off])
			}
		}
		out = append(out, e)
	}
	return out
}

func (g *Game) DrawEditorPortrait(fb *gfx.Framebuffer, index, x, y, w, h int) {
	if g.res.Faces != nil {
		if img, err := g.res.Faces.Decode(index); err == nil {
			pix := make([]uint8, w*h)
			for py := 0; py < h; py++ {
				for px := 0; px < w; px++ {
					pix[py*w+px] = img.Pix[(py*img.H/h)*img.W+px*img.W/w]
				}
			}
			key := fmt.Sprintf("editor-face/%02d", index)
			fb.DrawSource(key, x, y, w, h, w, h, pix)
			fb.MarkArtwork(key, x, y, w, h, w, h)
		}
	}
}

func (g *Game) applyAdmiralOverrides() {
	if len(g.admiralPreferences.Admirals) == 0 {
		return
	}
	g.admiralPreferences.Apply(g.mutableAdmirals())
}

// Validate before the caller persists; a failed transaction must not touch Game.
func (g *Game) ValidateAdmiralPreferences(p rules.AdmiralPreferences) error {
	_, _, err := g.prepareAdmiralPreferences(p)
	return err
}

func (g *Game) prepareAdmiralPreferences(p rules.AdmiralPreferences) ([]byte, *rules.Slot, error) {
	if err := p.Validate(); err != nil {
		return nil, nil, err
	}
	if err := p.ValidateCombat(g.res.ShipKinds); err != nil {
		return nil, nil, err
	}
	if len(g.res.Admiral) < 82*rules.AdmiralRecord {
		return nil, nil, fmt.Errorf("提督資料不足82筆")
	}
	data := bytes.Clone(g.res.Admiral)
	p.Apply(data)
	var work *rules.Slot
	if g.slot != nil {
		x := *g.slot
		work = &x
		work.Admiral = data
		work.Fleet = bytes.Clone(g.slot.Fleet)
		work.Ship = bytes.Clone(g.slot.Ship)
		for _, f := range g.fleets {
			own, deputy := p.Admirals[f.Admiral], p.Admirals[f.Adjutant]
			if len(own.Abilities) > 0 || len(deputy.Abilities) > 0 {
				if g.res.EditorTables == nil {
					return nil, nil, fmt.Errorf("缺少能力重算資料")
				}
				if err := work.RefreshAdmiralFleet(f.Index, g.res.ShipKinds, g.res.EditorTables.Mobility); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	return data, work, nil
}

func (g *Game) SetAdmiralPreferences(p rules.AdmiralPreferences) error {
	data, work, err := g.prepareAdmiralPreferences(p)
	if err != nil {
		return err
	}
	g.admiralPreferences = p.Clone()
	g.admirals = data
	g.res.Admiral = data
	if work != nil {
		g.slot.Admiral, g.slot.Fleet, g.slot.Ship = work.Admiral, work.Fleet, work.Ship
		for _, f := range g.fleets {
			if len(p.Admirals[f.Admiral].Abilities) > 0 || len(p.Admirals[f.Adjutant].Abilities) > 0 {
				g.syncPromotionView(f.Index)
			}
		}
	}
	g.reconnectAI()
	return nil
}

func (g *Game) armyBonus(faction int, movement bool) int {
	if g.AILevel() < 0 || faction < 0 || faction > 1 || movement && g.heroMovementDisabled {
		return 0
	}
	bonus := 0
	for _, u := range g.battleUnits {
		if u.Faction != faction || u.Squadron != 0 || u.Out || u.Ships <= 0 {
			continue
		}
		if f := g.fleetByIndex(u.Fleet); f != nil && f.Faction == faction {
			s := g.admiralPreferences.Special(f.Admiral)
			if movement {
				bonus = max(bonus, s.Movement)
			} else {
				bonus = max(bonus, s.Attack)
			}
		}
	}
	return bonus
}

// 僅攻擊艦隊的提督／參謀提供戰鬥機增幅；同艦隊取較高值。
func (g *Game) fleetFighterBonus(fleet int) int {
	if g.AILevel() < 0 {
		return 0
	}
	f := g.fleetByIndex(fleet)
	if f == nil {
		return 0
	}
	bonus := 0
	for _, i := range []int{f.Admiral, f.Adjutant} {
		if i >= 0 && i < 82 {
			bonus = max(bonus, g.admiralPreferences.Special(i).FighterAttack)
		}
	}
	return bonus
}
