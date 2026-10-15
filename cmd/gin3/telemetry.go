package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

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
	musicLibrary := ""
	if a.music != nil {
		musicTrack = a.music.Playing()
		if player, ok := a.music.(interface{ LibraryDir() string }); ok {
			musicLibrary = player.LibraryDir()
		}
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
	var settings map[string][4]int
	if a.g.State() == game.StateSettings {
		settings = compactSettingsObservation()
	}
	row := struct {
		Wall             int64                             `json:"wall_ns"`
		Ticks            int                               `json:"ticks"`
		State            string                            `json:"state"`
		Mode             string                            `json:"graphics"`
		FrameMS          float64                           `json:"frame_ms"`
		DrawMS           float64                           `json:"draw_ms"`
		Animations       []string                          `json:"animations,omitempty"`
		Toolbar          bool                              `json:"toolbar_visible"`
		Help             bool                              `json:"help_visible"`
		Enemies          bool                              `json:"show_enemies"`
		Playback         *game.PlaybackObservation         `json:"playback,omitempty"`
		BattleReady      bool                              `json:"battle_ready"`
		Tactical         *game.TacticalObservation         `json:"tactical,omitempty"`
		Fortress         *game.FortressPlaybackObservation `json:"fortress,omitempty"`
		Planet           *game.PlanetPlaybackObservation   `json:"planet,omitempty"`
		MouseButtons     [2]bool                           `json:"mouse_buttons"`
		Cursor           [2]int                            `json:"cursor"`
		FontStyle        string                            `json:"font_style"`
		MusicQuality     string                            `json:"music_quality"`
		HDMusicAvailable bool                              `json:"hd_music_available"`
		MusicTrack       string                            `json:"music_track"`
		SoundFont        string                            `json:"soundfont"`
		MT32Available    bool                              `json:"mt32_available"`
		SoundFontMenu    bool                              `json:"soundfont_menu"`
		MusicLibrary     string                            `json:"music_library"`
		AILevel          int                               `json:"ai_level"`
		AIMenu           bool                              `json:"ai_menu"`
		Settings         map[string][4]int                 `json:"settings,omitempty"`
		Rules            game.RuleOptions                  `json:"rules"`
	}{start.UnixNano(), a.ticks, a.g.State().String(), mode, float64(frameTime) / 1e6, float64(time.Since(start)) / 1e6, animations, a.toolbar.visible, a.toolbar.help, a.g.ShowEnemies(), a.g.PlaybackObservation(), a.g.BattleInputReady(), a.g.TacticalObservation(), a.g.FortressPlaybackObservation(), a.g.PlanetPlaybackObservation(), [2]bool{ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft), ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)}, [2]int{}, fontStyle, musicQuality, a.musicHD != nil, musicTrack, a.selectedSoundFont(), a.musicMT32 != nil, a.soundFontMenu, musicLibrary, a.g.AILevel(), a.aiSettingsMenu, settings, a.g.RuleOptions()}
	row.Cursor[0], row.Cursor[1] = a.g.MousePosition()
	_ = json.NewEncoder(a.frameTrace).Encode(row)
}
