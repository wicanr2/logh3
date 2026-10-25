package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPosterQuoteMinuteBoundaries(t *testing.T) {
	for _, c := range []struct {
		elapsed time.Duration
		index   int
	}{
		{-time.Second, 0}, {0, 0}, {time.Minute - time.Nanosecond, 0},
		{time.Minute, 1}, {2 * time.Minute, 2}, {3 * time.Minute, 3}, {4 * time.Minute, 4},
		{5*time.Minute - time.Nanosecond, 4}, {5 * time.Minute, 0}, {24 * time.Hour, 0},
	} {
		if got := posterQuoteIndex(c.elapsed, 5); got != c.index {
			t.Fatalf("%s -> %d，預期%d", c.elapsed, got, c.index)
		}
	}
	for _, count := range []int{-1, 0, 1} {
		if posterQuoteIndex(time.Minute, count) != 0 {
			t.Fatal("缺輪替不得越界")
		}
	}
	q := &posterQuoteRotation{}
	start := time.Unix(1234567, 0)
	if q.indexAt(start) != 0 || q.indexAt(start.Add(61*time.Second)) != 1 || q.indexAt(start.Add(301*time.Second)) != 0 {
		t.Fatal("實際顯示時鐘未沿60秒循環")
	}
	if q.epoch != start {
		t.Fatal("讀取組別重設了計時")
	}
}

func TestPosterQuoteInstalledAssetsAndInvalidExtension(t *testing.T) {
	source := filepath.Join("..", "..", "assets", "hd", "poster", "approved")
	q, err := loadPosterQuoteRotation(source)
	if err != nil || q == nil {
		t.Fatal("五組名言未正確入庫", err)
	}
	for _, set := range q.images {
		if len(set) != 5 {
			t.Fatal("新增組別未齊五語")
		}
	}
	manifest, err := os.ReadFile(filepath.Join(source, "quotes.json"))
	if err != nil {
		t.Fatal(err)
	}
	var original posterQuoteManifest
	if err = json.Unmarshal(manifest, &original); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"interval", "locale", "path", "hash", "dimensions"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			var m posterQuoteManifest
			if err := json.Unmarshal(manifest, &m); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "interval":
				m.Seconds = 1
			case "locale":
				delete(m.Sets[0].Files, "ko")
			case "path":
				item := m.Sets[0].Files["zh-TW"]
				item.File = "../escape.png"
				m.Sets[0].Files["zh-TW"] = item
			case "hash", "dimensions":
				var b bytes.Buffer
				if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
					t.Fatal(err)
				}
				item := m.Sets[0].Files["zh-TW"]
				if scenario == "dimensions" {
					item.SHA = fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
					m.Sets[0].Files["zh-TW"] = item
				}
				if err := os.WriteFile(filepath.Join(dir, item.File), b.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
			}
			b, _ := json.Marshal(m)
			if err := os.WriteFile(filepath.Join(dir, "quotes.json"), b, 0600); err != nil {
				t.Fatal(err)
			}
			if got, err := loadPosterQuoteRotation(dir); got != nil || err == nil {
				t.Fatal("損毀擴充仍啟用", scenario)
			}
		})
	}
	// An invalid optional extension keeps the entire approved frame and first pair.
	art := t.TempDir()
	dir := filepath.Join(art, "poster", "approved")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	var main struct {
		Files map[string]any `json:"files"`
	}
	baseManifest, err := os.ReadFile(filepath.Join(source, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(baseManifest, &main); err != nil {
		t.Fatal(err)
	}
	for name := range main.Files {
		b, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.WriteFile(filepath.Join(dir, "manifest.json"), baseManifest, 0600); err != nil {
		t.Fatal(err)
	}
	if p, err := loadPosterArt(art); err != nil || p == nil || p.approved == nil || p.approved.rotation != nil {
		t.Fatal("舊素材沒有安全沿用首組", err)
	}
	if err = os.WriteFile(filepath.Join(dir, "quotes.json"), []byte(`{"interval_seconds":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := loadPosterArt(art)
	if err == nil || p == nil || !p.available() || p.approved == nil || p.approved.rotation != nil || len(p.approved.labels) != 5 {
		t.Fatal("損毀名言不應關閉核准海報", err)
	}
}
