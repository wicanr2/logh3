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
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type posterQuoteFile struct {
	File string `json:"file"`
	SHA  string `json:"sha256"`
}

type posterQuoteManifest struct {
	Format   string `json:"format"`
	Proposal string `json:"proposal_sha256"`
	Source   string `json:"source_sha256"`
	Seconds  int    `json:"interval_seconds"`
	Rect     [4]int `json:"rect"`
	Sets     []struct {
		ID    string                     `json:"id"`
		Files map[string]posterQuoteFile `json:"files"`
	} `json:"sets"`
}

type posterQuoteRotation struct {
	images   [4]map[string]image.Image
	textures map[string]*ebiten.Image
	offset   image.Point
	epoch    time.Time
	index    int
}

func loadPosterQuoteRotation(dir string) (*posterQuoteRotation, error) {
	path := filepath.Join(dir, "quotes.json")
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 32768 {
		return nil, fmt.Errorf("名言清單過大或無效")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m posterQuoteManifest
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Format != "logh3-poster-quotes-v1" || m.Proposal != approvedPosterSHA || m.Seconds != 60 || len(m.Sets) != 4 || m.Rect != [4]int{0, 880, 2560, 1280} || len(m.Source) != 64 {
		return nil, fmt.Errorf("名言輪替清單不相符")
	}
	q := &posterQuoteRotation{offset: image.Pt(m.Rect[0], m.Rect[1]), textures: map[string]*ebiten.Image{}}
	seen := map[string]bool{}
	for i, set := range m.Sets {
		if set.ID == "" || seen[set.ID] || len(set.Files) != 5 {
			return nil, fmt.Errorf("名言組別無效")
		}
		seen[set.ID] = true
		q.images[i] = map[string]image.Image{}
		for _, locale := range []string{"zh-TW", "zh-CN", "ja", "en", "ko"} {
			item := set.Files[locale]
			if item.File != fmt.Sprintf("quote-%d-%s.png", i+2, locale) {
				return nil, fmt.Errorf("名言檔名無效")
			}
			path := filepath.Join(dir, item.File)
			info, err := os.Stat(path)
			if err != nil {
				return nil, err
			}
			if !info.Mode().IsRegular() || info.Size() > 8<<20 {
				return nil, fmt.Errorf("名言圖片過大或無效")
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if fmt.Sprintf("%x", sha256.Sum256(data)) != item.SHA {
				return nil, fmt.Errorf("名言圖片雜湊錯誤：%s", item.File)
			}
			config, err := png.DecodeConfig(bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
			if config.Width != m.Rect[2]-m.Rect[0] || config.Height != m.Rect[3]-m.Rect[1] {
				return nil, fmt.Errorf("名言圖片尺寸錯誤：%s", item.File)
			}
			q.images[i][locale], err = png.Decode(bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
		}
	}
	return q, nil
}

func posterQuoteIndex(elapsed time.Duration, count int) int {
	if elapsed < 0 || count <= 1 {
		return 0
	}
	return int((elapsed / time.Minute) % time.Duration(count))
}

func (q *posterQuoteRotation) indexAt(now time.Time) int {
	if q.epoch.IsZero() {
		q.epoch = now
	}
	q.index = posterQuoteIndex(now.Sub(q.epoch), 5)
	return q.index
}

func (p *approvedPosterArt) drawQuote(screen *ebiten.Image, locale string, now time.Time) {
	if p.labels[locale] == nil {
		locale = "zh-TW"
	}
	index := 0
	if p.rotation != nil {
		index = p.rotation.indexAt(now)
	}
	if index == 0 {
		if p.labelTextures[locale] == nil {
			p.labelTextures[locale] = ebiten.NewImageFromImage(p.labels[locale])
		}
		screen.DrawImage(p.labelTextures[locale], nil)
		return
	}
	q := p.rotation
	key := fmt.Sprintf("%d:%s", index, locale)
	if q.textures[key] == nil {
		q.textures[key] = ebiten.NewImageFromImage(q.images[index-1][locale])
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(q.offset.X), float64(q.offset.Y))
	screen.DrawImage(q.textures[key], op)
}
