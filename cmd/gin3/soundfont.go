package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

var soundFontChoices = []string{"gs", "mt32"}
var soundFontChoiceRect = compactControlBounds("soundfont")
var soundFontLabels = map[string]string{"gs": "GeneralUser GS", "mt32": "Roland MT-32"}

func soundFontMenuRow(i int) image.Rectangle {
	return image.Rect(288, 256+i*16, 448, 272+i*16)
}
func (a *app) selectedSoundFont() string {
	if a.soundFont == "" {
		return "gs"
	}
	return a.soundFont
}
func (a *app) soundFontLibrary(choice string) *audio.Library {
	switch choice {
	case "gs":
		return a.musicHD
	case "mt32":
		return a.musicMT32
	}
	return nil
}
func (a *app) selectSoundFont(choice string, persist bool) error {
	if _, ok := soundFontLabels[choice]; !ok {
		return fmt.Errorf("未知音色庫：%s", choice)
	}
	if a.soundFontLibrary(choice) == nil {
		return fmt.Errorf("%s 曲庫尚未匯入", soundFontLabels[choice])
	}
	if persist && a.preferencesPath != "" {
		if err := writeSoundFontPreference(a.preferencesPath, choice); err != nil {
			return fmt.Errorf("保存音色偏好：%w", err)
		}
	}
	a.soundFont = choice
	a.setHDMusic(true)
	a.soundFontError = ""
	return nil
}

func readSoundFontPreference(path string) (string, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var p struct {
		SoundFont string `json:"soundfont"`
	}
	if err = json.Unmarshal(b, &p); err != nil {
		return "", err
	}
	if p.SoundFont != "" {
		if _, ok := soundFontLabels[p.SoundFont]; !ok {
			return "", fmt.Errorf("未知音色偏好：%s", p.SoundFont)
		}
	}
	return p.SoundFont, nil
}

func writeSoundFontPreference(path, choice string) error {
	if _, ok := soundFontLabels[choice]; !ok {
		return fmt.Errorf("未知音色庫：%s", choice)
	}
	b, _ := json.Marshal(choice)
	return writePreferences(path, map[string]json.RawMessage{"soundfont": b})
}

func writePreferences(path string, updates map[string]json.RawMessage) error {
	p := map[string]json.RawMessage{}
	if b, err := os.ReadFile(path); err == nil {
		if err = json.Unmarshal(b, &p); err != nil {
			return err
		}
		if p == nil {
			return fmt.Errorf("偏好檔不是物件")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	for key, value := range updates {
		p[key] = value
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".preferences-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func (a *app) drawSoundFontChoice(fb *gfx.Framebuffer, text func(image.Rectangle, []uint16), button func(image.Rectangle, uint8)) {
	label := image.Rect(176, 360, 296, 380)
	(ui.Frame{Fill: 9}).Draw(fb, label.Min.X, label.Min.Y, label.Dx(), label.Dy())
	text(label, []uint16{0xADB5, 0xA6E2, 0xAE77}) // 音色庫
	button(soundFontChoiceRect, 9)
	draw := func(r image.Rectangle, label string, ink uint8) {
		x := r.Min.X + (r.Dx()-len(label)*8)/2
		for _, c := range []byte(label) {
			ui.DrawHalfGlyphClipped(fb, a.toolbar.font, x, r.Min.Y+(r.Dy()-15)/2, c, ink, 3, -1, false, r.Min.X+2, r.Min.Y+2, r.Dx()-4, r.Dy()-4)
			x += 8
		}
	}
	short := map[string]string{"gs": "GS", "mt32": "MT-32"}
	draw(soundFontChoiceRect, short[a.selectedSoundFont()], 7)
	if a.soundFontMenu {
		for i, choice := range soundFontChoices {
			r := soundFontMenuRow(i)
			fill, ink := uint8(9), uint8(7)
			if choice == a.selectedSoundFont() {
				fill = 2
			}
			if a.soundFontLibrary(choice) == nil {
				ink = 8
			}
			button(r, fill)
			draw(r, soundFontLabels[choice], ink)
		}
	}
	if a.soundFontError != "" {
		text(image.Rect(176, 380, 448, 400), []uint16{0xB54C, 0xAA6B, 0xA4C1, 0xB4AB, 0xADB5, 0xA6E2})
	} // 無法切換音色
}
