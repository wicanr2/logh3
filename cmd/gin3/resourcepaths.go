package main

import (
	"os"
	"path/filepath"
)

// Packaged programs resolve their defaults beside the executable. Explicit
// flags keep their meaning relative to the caller's working directory.
func packagedDefaults(executable string, explicit map[string]bool, orig, data, art, music, graphics *string) {
	root := filepath.Dir(executable)
	if info, err := os.Stat(filepath.Join(root, "data")); err != nil || !info.IsDir() {
		return // Source checkout keeps the development defaults.
	}
	for _, item := range []struct {
		flag, directory string
		value           *string
	}{{"orig", "original", orig}, {"data", "data", data}, {"art", "art", art}} {
		if !explicit[item.flag] {
			*item.value = filepath.Join(root, item.directory)
		}
	}
	if !explicit["music"] {
		if info, err := os.Stat(filepath.Join(root, "music")); err == nil && info.IsDir() {
			*music = filepath.Join(root, "music")
		}
	}
	if !explicit["graphics"] {
		if info, err := os.Stat(filepath.Join(*art, "face")); err == nil && info.IsDir() {
			*graphics = "hd"
		}
	}
}
