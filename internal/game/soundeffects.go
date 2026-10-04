package game

import (
	"fmt"
	"time"
)

type soundEffectsState struct {
	enabled   bool
	remaining time.Duration
	events    []int
}

type fleetSceneSound struct {
	id, tick, pictureAt int
	defence, played     bool
}

// EnableSoundEffects只在完整素材與實際播放端可用時開啟（規格104）。
func (g *Game) EnableSoundEffects(enabled bool) error {
	if enabled && (g.res.FleetSound == nil || g.res.SoundSamples == nil || len(g.res.SoundSamples.Samples) != 28) {
		return fmt.Errorf("音效需要完整28段樣本與艦隊對照表")
	}
	g.soundEffects = soundEffectsState{enabled: enabled}
	return nil
}

// DrainSoundEffects回傳自上次更新以來的一次性播放事件。
func (g *Game) DrainSoundEffects() []int {
	ids := g.soundEffects.events
	g.soundEffects.events = nil
	return ids
}

func (g *Game) advanceSoundEffects(d time.Duration) {
	g.soundEffects.remaining = max(0, g.soundEffects.remaining-d)
}

func (g *Game) addFleetSceneSound(s *fleetCombatSceneState, unit, column, tick, pictureAt int, defence bool) error {
	if !g.soundEffects.enabled {
		return nil
	}
	u := g.battleUnits[unit]
	id, err := g.res.FleetSound.Cue(u.Faction, u.Kind, column, u.Squadron == 0, defence)
	if err != nil {
		return err
	}
	if id >= 0 {
		s.sounds = append(s.sounds, fleetSceneSound{id: id, tick: tick, pictureAt: pictureAt, defence: defence})
	}
	return nil
}

func (g *Game) processFleetSceneSounds(s *fleetCombatSceneState) {
	if !g.soundEffects.enabled {
		return
	}
	for i := range s.sounds {
		cue := &s.sounds[i]
		if cue.played || cue.tick > s.tick {
			continue
		}
		if g.soundEffects.remaining > 0 {
			d := g.engagementTickDuration()
			delay := int((g.soundEffects.remaining + d - 1) / d)
			for p := cue.pictureAt; p < len(s.pictures); p++ {
				s.pictures[p].tick += delay
			}
			for j := i; j < len(s.sounds); j++ {
				s.sounds[j].tick += delay
			}
			if cue.defence {
				s.resultTick += delay
			} else {
				s.attackEnd += delay
			}
			continue
		}
		cue.played = true
		g.soundEffects.events = append(g.soundEffects.events, cue.id)
		g.soundEffects.remaining = g.res.SoundSamples.Samples[cue.id].Duration()
	}
}
