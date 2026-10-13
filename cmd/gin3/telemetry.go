package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
)

// Optional observation of the normal window, with no alternate game entry or
// timing control. Wall timestamps align it with the recording event log.
func (a *app) traceFrame(fb *gfx.Framebuffer, start time.Time, frameTime time.Duration) {
	if a.frameTrace == nil {
		return
	}
	mode := "original"
	if a.hd != nil && !a.original {
		mode = "hd"
	}
	var animations []string
	fontStyle, musicQuality := "original", "original"
	musicTrack := ""
	if a.music != nil {
		musicTrack = a.music.Playing()
	}
	if a.hd != nil {
		fontStyle = a.hd.FontStyle()
	}
	if a.hdMusic {
		musicQuality = "hd"
	}
	for _, c := range fb.Commands {
		if c.Kind == gfx.DrawAnimation {
			h := sha256.Sum256(c.Source)
			animations = append(animations, hex.EncodeToString(h[:]))
		}
	}
	row := struct {
		Wall             int64                     `json:"wall_ns"`
		Ticks            int                       `json:"ticks"`
		State            string                    `json:"state"`
		Mode             string                    `json:"graphics"`
		FrameMS          float64                   `json:"frame_ms"`
		DrawMS           float64                   `json:"draw_ms"`
		Animations       []string                  `json:"animations,omitempty"`
		Toolbar          bool                      `json:"toolbar_visible"`
		Help             bool                      `json:"help_visible"`
		Enemies          bool                      `json:"show_enemies"`
		Playback         *game.PlaybackObservation `json:"playback,omitempty"`
		BattleReady      bool                      `json:"battle_ready"`
		Tactical         *game.TacticalObservation `json:"tactical,omitempty"`
		FontStyle        string                    `json:"font_style"`
		MusicQuality     string                    `json:"music_quality"`
		HDMusicAvailable bool                      `json:"hd_music_available"`
		MusicTrack       string                    `json:"music_track"`
	}{start.UnixNano(), a.ticks, a.g.State().String(), mode, float64(frameTime) / 1e6, float64(time.Since(start)) / 1e6, animations, a.toolbar.visible, a.toolbar.help, a.g.ShowEnemies(), a.g.PlaybackObservation(), a.g.BattleInputReady(), a.g.TacticalObservation(), fontStyle, musicQuality, a.musicHD != nil, musicTrack}
	_ = json.NewEncoder(a.frameTrace).Encode(row)
}
