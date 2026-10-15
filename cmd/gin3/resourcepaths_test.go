package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackagedDefaults(t *testing.T) {
	root := filepath.Join(t.TempDir(), "銀河 英雄")
	for _, dir := range []string{"data", "original", "art/face", "music"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	orig, data, art, music, graphics := "org_game/extracted", "docs/internals/data", "assets/hd", "", "original"
	packagedDefaults(filepath.Join(root, "gin3.exe"), nil, &orig, &data, &art, &music, &graphics)
	if orig != filepath.Join(root, "original") || data != filepath.Join(root, "data") || art != filepath.Join(root, "art") || music != filepath.Join(root, "music") || graphics != "hd" {
		t.Fatalf("package paths: %q %q %q %q %q", orig, data, art, music, graphics)
	}
	// An explicit missing directory must fail there, never silently load another game.
	orig, music, graphics = "my-missing-game", "", "original"
	packagedDefaults(filepath.Join(root, "gin3.exe"), map[string]bool{"orig": true, "music": true, "graphics": true}, &orig, &data, &art, &music, &graphics)
	if orig != "my-missing-game" || music != "" || graphics != "original" {
		t.Fatal("explicit flags were overridden")
	}
	// Engine-only packages still name their documented import directory.
	if err := os.Remove(filepath.Join(root, "original")); err != nil {
		t.Fatal(err)
	}
	packagedDefaults(filepath.Join(root, "gin3.exe"), nil, &orig, &data, &art, &music, &graphics)
	if orig != filepath.Join(root, "original") {
		t.Fatal("missing import fell back to checkout")
	}
}

func TestDevelopmentDefaults(t *testing.T) {
	orig, data, art, music, graphics := "org_game/extracted", "docs/internals/data", "assets/hd", "", "original"
	packagedDefaults(filepath.Join(t.TempDir(), "gin3"), nil, &orig, &data, &art, &music, &graphics)
	if orig != "org_game/extracted" || data != "docs/internals/data" || art != "assets/hd" || music != "" || graphics != "original" {
		t.Fatal("development defaults changed")
	}
}
