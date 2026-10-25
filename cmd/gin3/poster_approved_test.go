package main

import (
	"bytes"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestApprovedPosterInstalledAssetsAndTamper(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "hd")
	p, err := loadPosterArt(dir)
	if err != nil || p == nil || p.approved == nil || !p.available() {
		t.Fatal("已入庫定案素材不可用", err)
	}
	for _, set := range []string{"v1", "v2", "v3"} {
		if !p.selectSet(set) || p.setName() != set {
			t.Fatal("核准組別不可選", set)
		}
	}
	if len(p.approved.labels) != 5 {
		t.Fatal("五語文字不齊")
	}
	bad := t.TempDir()
	path := filepath.Join(bad, "poster", "approved")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(dir, "poster", "approved", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "manifest.json"), manifest, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "base-v1.png"), []byte("changed art"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := loadApprovedPosterArt(bad); got != nil || err == nil || !bytes.Contains([]byte(err.Error()), []byte("雜湊")) {
		t.Fatal("被換掉的定案圖仍載入", got, err)
	}
}

func TestApprovedPosterInputAndControls(t *testing.T) {
	a := &app{posterMode: true, posters: &posterArt{approved: &approvedPosterArt{}}}
	v := a.posterViewport()
	// Frozen approved HTML: 844.8x633.6, centred on a 1280x720 frame.
	for _, pair := range [][2]float64{{v.X, 435.2}, {v.Y, 86.4}, {v.W, 1689.6}, {v.H, 1267.2}} {
		if math.Abs(pair[0]-pair[1]) > 1e-9 {
			t.Fatal("未沿用定案比例", v)
		}
	}
	for _, c := range compactControls {
		p := c.Bounds.Min.Add(c.Bounds.Size().Div(2))
		x, y := a.gamePosition(int(v.X+(float64(p.X)+.5)*v.W/640), int(v.Y+(float64(p.Y)+.5)*v.H/480))
		if !image.Pt(x, y).In(c.Bounds) {
			t.Fatalf("%s 命中錯位：%d,%d", c.ID, x, y)
		}
	}
	for _, p := range []image.Point{{0, 0}, {2560, 0}, {1280, 1440}, {1280, 10}} {
		x, y := a.gamePosition(p.X, p.Y)
		if image.Pt(x, y).In(image.Rect(0, 0, 640, 480)) {
			t.Fatal("海報區誤觸遊戲", p, x, y)
		}
	}
}

func TestApprovedPosterKeepsContentOutsideDetectedBlackBands(t *testing.T) {
	v := (&app{posterMode: true, posters: &posterArt{approved: &approvedPosterArt{}}}).posterViewport()
	for _, width := range []int{640, 1920} {
		scale := width / 640
		top := approvedPosterBand(v, image.Rect(0, 0, width, 32*scale), width, false)
		bottom := approvedPosterBand(v, image.Rect(0, 444*scale, width, 480*scale), width, true)
		if top.Max.Y > int(v.Y+32*v.H/480) || float64(bottom.Min.Y) < v.Y+444*v.H/480 {
			t.Fatal("裝飾侵入非空白遊戲", top, bottom)
		}
		if !approvedPosterBand(v, image.Rectangle{}, width, false).Empty() {
			t.Fatal("非遊戲畫面不得補畫艦艇")
		}
	}
}

func TestApprovedPosterRejectsUnapprovedManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "poster", "approved")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	// No production art is needed to prove a wrong proposal cannot activate the layout.
	if err := os.WriteFile(filepath.Join(path, "manifest.json"), []byte(`{"format":"logh3-approved-poster-v1","canvas":[2560,1440],"game_scale":88,"proposal_sha256":"different"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if p, err := loadApprovedPosterArt(dir); p != nil || err == nil {
		t.Fatal("錯誤定案來源仍啟用", p, err)
	}
}
