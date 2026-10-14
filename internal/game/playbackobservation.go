package game

import (
	"encoding/hex"
	"fmt"
)

// PlaybackObservation is a read-only view for the optional normal-window trace.
// It does not advance ticks, select commands, apply damage or consume randomness.
type PlaybackObservation struct {
	ID               string            `json:"id"`
	Kind             string            `json:"kind"`
	Tick             int               `json:"tick"`
	Interval         int               `json:"interval"`
	AttackEnd        int               `json:"attack_end"`
	ResultTick       int               `json:"result_tick"`
	Result           bool              `json:"result"`
	DamageApplied    bool              `json:"damage_applied"`
	Lost             int               `json:"lost"`
	Weapon           int               `json:"weapon"`
	Applied          int               `json:"applied"`
	Pictures         []PlaybackPicture `json:"pictures"`
	Sounds           []PlaybackSound   `json:"sounds,omitempty"`
	SoundPlaying     bool              `json:"sound_playing"`
	SoundRemainingNS int64             `json:"sound_remaining_ns"`
}

type PlaybackPicture struct {
	Segment string `json:"segment_hex"`
	Frame   int    `json:"frame"`
	Side    int    `json:"side"`
	Tick    int    `json:"tick"`
}

type PlaybackSound struct {
	ID      int  `json:"id"`
	Tick    int  `json:"tick"`
	Defence bool `json:"defence"`
	Played  bool `json:"played"`
}

// BattleInputReady reports the existing input boundary without changing it.
func (g *Game) BattleInputReady() bool {
	return g.state == StateBattle && !g.battlePreparing && !g.battleMoving && !g.battleSpotting.active && !g.fleetCombat.active && !g.battleFinishing
}

func (g *Game) PlaybackObservation() *PlaybackObservation {
	if g.state != StateEngagement {
		return nil
	}
	v := &PlaybackObservation{SoundPlaying: g.soundEffects.playing, SoundRemainingNS: int64(g.soundEffects.remaining)}
	var pictures []engagementPicture
	if g.fleetCombat.active {
		s := g.fleetCombat.scene
		if s == nil {
			return nil
		}
		v.ID = fmt.Sprintf("fleet/%d/%d/%d", g.battleRegion, g.battleTurn, g.fleetCombat.commandAt)
		v.Kind, v.Tick, v.Interval = "fleet", s.tick, g.fleetSceneInterval()
		v.AttackEnd, v.ResultTick = s.attackEnd, s.resultTick
		v.DamageApplied, v.Result, v.Lost, v.Weapon = s.damaged, s.damaged && s.tick >= s.resultTick, s.hit.Lost, int(s.weapon)
		pictures = s.pictures
		for _, cue := range s.sounds {
			v.Sounds = append(v.Sounds, PlaybackSound{cue.id, cue.tick, cue.defence, cue.played})
		}
	} else {
		if len(g.engQueue) == 0 {
			return nil
		}
		e := g.engQueue[0]
		v.ID = fmt.Sprintf("planet/%d/%d/%d/%d/%d", g.battleRegion, g.battleTurn, e.fleet, e.planet, len(g.engQueue))
		v.Kind, v.Tick, v.Interval = "planet", g.engTick, g.fleetSceneInterval()
		pictures, v.ResultTick = g.engagementTimeline(e)
		v.Result, v.Lost = g.engTick >= v.ResultTick, e.destroyed
		if e.stage != nil {
			for _, cue := range e.stage.sounds {
				v.Sounds = append(v.Sounds, PlaybackSound{cue.id, cue.tick, cue.defence, cue.played})
			}
		}
	}
	for _, p := range pictures {
		v.Pictures = append(v.Pictures, PlaybackPicture{hex.EncodeToString([]byte(p.segment)), p.frame, p.side, p.tick})
		if p.tick <= v.Tick {
			v.Applied++
		}
	}
	return v
}
