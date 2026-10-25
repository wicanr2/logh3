package main

import (
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/presentation"
)

func TestPosterLayoutAndMouseBorders(t *testing.T) {
	for _, hd := range []bool{false, true} {
		a := &app{posterMode: true}
		if hd {
			a.hd = &presentation.Renderer{}
		}
		for _, c := range [][4]int{{320, 0, 0, 0}, {2239, 1439, 639, 479}, {319, 0, -1, 0}, {2240, 0, 640, 0}, {0, -1, -107, -1}, {1664, 1050, 448, 350}} {
			x, y := a.gamePosition(c[0], c[1])
			if x != c[2] || y != c[3] {
				t.Fatalf("hd%v %v -> %d,%d", hd, c, x, y)
			}
		}
		if w, h := a.Layout(1280, 720); w != 2560 || h != 1440 || w*9 != h*16 {
			t.Fatal("海報外框比例錯誤", w, h)
		}
		for _, c := range compactControls {
			p := c.Bounds.Min.Add(c.Bounds.Size().Div(2))
			x, y := a.gamePosition(320+p.X*3, p.Y*3)
			if !image.Pt(x, y).In(c.Bounds) {
				t.Fatal("海報模式命中錯位", c.ID)
			}
		}
		if hd {
			if !a.toggleGraphics() || !a.posterMode {
				t.Fatal("F7改了海報模式")
			}
			if w, _ := a.Layout(1280, 720); w != 2560 {
				t.Fatal("F7改了海報外框")
			}
		}
	}
	if posterWindowWidth(720, true) != 1280 || posterWindowWidth(720, false) != 960 {
		t.Fatal("切換沒有保持遊戲高度")
	}
}

func TestPosterPreferenceRoundtripAndFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if enabled, err := readPosterPreference(path); enabled || err != nil {
		t.Fatal("新偏好未預設關閉", enabled, err)
	}
	if err := os.WriteFile(path, []byte(`{"poster_mode":false,"other":17,"non_battleship_flagship":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	art := &posterArt{images: [2]image.Image{image.NewRGBA(image.Rect(0, 0, 1, 1)), image.NewRGBA(image.Rect(0, 0, 1, 1))}}
	a := &app{posters: art, preferencesPath: path}
	for _, enabled := range []bool{true, false} {
		if !a.setPosterMode(enabled) || a.posterMode != enabled {
			t.Fatal("設定海報失敗", a.posterError)
		}
		if got, err := readPosterPreference(path); got != enabled || err != nil {
			t.Fatal("偏好重啟讀回錯誤", got, err)
		}
	}
	b, _ := os.ReadFile(path)
	var value map[string]any
	_ = json.Unmarshal(b, &value)
	if value["other"] != float64(17) || value["non_battleship_flagship"] != true {
		t.Fatal("保存海報改了其他偏好")
	}
	a.preferencesPath = t.TempDir()
	if a.setPosterMode(true) || a.posterMode || a.posterError == "" {
		t.Fatal("保存失敗仍啟用")
	}
	a.posters = nil
	if a.setPosterMode(true) || a.posterMode {
		t.Fatal("缺圖仍啟用")
	}
	if _, err := loadPosterArt(t.TempDir()); err == nil {
		t.Fatal("缺圖未回報")
	}
	_ = os.WriteFile(path, []byte(`{"poster_mode":null}`), 0600)
	if _, err := readPosterPreference(path); err == nil {
		t.Fatal("null被當成合法開關")
	}
}

func TestPosterSetsCycleAndUnavailable(t *testing.T) {
	pair := func() *posterArt {
		return &posterArt{images: [2]image.Image{image.NewRGBA(image.Rect(0, 0, 1, 1)), image.NewRGBA(image.Rect(0, 0, 1, 1))}}
	}
	art := pair()
	art.alternates = [2]*posterArt{pair(), pair()}
	path := filepath.Join(t.TempDir(), "preferences.json")
	a := &app{posters: art, preferencesPath: path}
	for _, want := range []string{"v1", "v2", "v3", "off", "v1"} {
		a.cyclePoster()
		got := a.posters.setName()
		if !a.posterMode {
			got = "off"
		}
		if got != want {
			t.Fatal(want, got)
		}
		if mode, err := readPosterPreference(path); mode != a.posterMode || err != nil {
			t.Fatal(mode, err)
		}
		if set, err := readPosterSetPreference(path); set != a.posters.setName() || err != nil {
			t.Fatal(set, err)
		}
	}
	art.alternates[0] = nil
	a.cyclePoster()
	if art.setName() != "v3" {
		t.Fatal("missing set not skipped")
	}
	if art.selectSet("v2") || art.setName() != "v3" {
		t.Fatal("missing set replaced current set")
	}
	a.preferencesPath = t.TempDir()
	a.cyclePoster()
	if !a.posterMode || art.setName() != "v3" || a.posterError == "" {
		t.Fatal("failed transaction changed set")
	}
}
