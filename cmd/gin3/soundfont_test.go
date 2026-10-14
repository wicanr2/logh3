package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/audio"
)

func TestSoundFontPlayerPreferences(t *testing.T) {
	dir := t.TempDir()
	for _, track := range audio.Tracks {
		if err := os.WriteFile(filepath.Join(dir, track+".ogg"), []byte{1}, 0600); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "preferences.json")
	a := app{musicHD: audio.NewHDLibrary(dir), musicMT32: audio.NewHDLibrary(dir), preferencesPath: path}
	if err := a.selectSoundFont("mt32", true); err != nil {
		t.Fatal(err)
	}
	if choice, err := readSoundFontPreference(path); err != nil || choice != "mt32" || !a.hdMusic {
		t.Fatalf("preference=%q error=%v", choice, err)
	}
	if err := a.selectSoundFont("gs", true); err != nil {
		t.Fatal(err)
	}
	if choice, _ := readSoundFontPreference(path); choice != "gs" {
		t.Fatal("second selection did not persist")
	}
	before, _ := os.ReadFile(path)
	a.musicMT32 = nil
	if err := a.selectSoundFont("mt32", true); err == nil || a.soundFont != "gs" {
		t.Fatal("missing library changed the selection")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatal("missing library overwrote preference")
	}
	if err := os.WriteFile(path, []byte("broken json"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.selectSoundFont("gs", true); err == nil {
		t.Fatal("corrupt preference overwritten")
	}
	after, _ = os.ReadFile(path)
	if string(after) != "broken json" {
		t.Fatal("corrupt preference was destroyed")
	}
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	a.musicMT32 = audio.NewHDLibrary(dir)
	if err := os.Chmod(filepath.Dir(path), 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(filepath.Dir(path), 0700)
	if err := a.selectSoundFont("mt32", true); err == nil || a.soundFont != "gs" {
		t.Fatal("failed preference write changed the active source")
	}
	after, _ = os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatal("failed preference write destroyed the old preference")
	}
}
