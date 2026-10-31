package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApprovedPostersRegisterWithoutDecodedImages(t *testing.T) {
	p, err := loadApprovedPosterArt(filepath.Join("..", "..", "assets", "hd"))
	if err != nil {
		t.Fatal(err)
	}
	for _, im := range p.bases {
		if im != nil {
			t.Fatal("啟動解碼全部海報")
		}
	}
	if p.hulls != nil {
		t.Fatal("啟動解碼艦體")
	}
	for _, im := range p.labels {
		if im != nil {
			t.Fatal("啟動解碼全部語系")
		}
	}
	for _, set := range p.rotation.images {
		for _, im := range set {
			if im != nil {
				t.Fatal("啟動解碼全部名言")
			}
		}
	}
	source := p.files["base-v1.png"]
	if im := source.image(); im == nil || im.Bounds().Dx() != posterTotalW {
		t.Fatal("首用不能解碼")
	}
	raw, err := os.ReadFile(source.path)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "changed.png")
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := registerPosterPNG(path, source.digest, source.w, source.h, source.limit)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)/2] ^= 1
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if file.image() != nil || !file.failed {
		t.Fatal("首用未拒絕雜湊不同的圖")
	}
}
