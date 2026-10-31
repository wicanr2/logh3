package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"os"
)

// 啟動只驗證PNG標頭，像素與完整雜湊在首次顯示時讀取。
type posterPNG struct {
	path, digest string
	w, h         int
	limit        int64
	failed       bool
}

func registerPosterPNG(path, digest string, w, h int, limit int64) (*posterPNG, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("海報檔案大小無效")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	c, err := png.DecodeConfig(f)
	f.Close()
	if err != nil || c.Width != w || c.Height != h {
		return nil, fmt.Errorf("海報標頭／尺寸或雜湊無效：%s", path)
	}
	if decoded, err := hex.DecodeString(digest); err != nil || len(decoded) != 32 {
		return nil, fmt.Errorf("海報雜湊格式無效")
	}
	return &posterPNG{path: path, digest: digest, w: w, h: h, limit: limit}, nil
}

func (p *posterPNG) image() image.Image {
	if p == nil || p.failed {
		return nil
	}
	im, err := p.decode()
	if err != nil {
		p.failed = true
		fmt.Fprintf(os.Stderr, "海報圖片略過 %s：%v\n", p.path, err)
		return nil
	}
	return im
}

func (p *posterPNG) decode() (image.Image, error) {
	info, err := os.Stat(p.path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > p.limit {
		return nil, fmt.Errorf("海報檔案大小無效")
	}
	raw, err := os.ReadFile(p.path)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != p.digest {
		return nil, fmt.Errorf("海報雜湊不符")
	}
	c, err := png.DecodeConfig(bytes.NewReader(raw))
	if err != nil || c.Width != p.w || c.Height != p.h {
		return nil, fmt.Errorf("海報尺寸已改變")
	}
	return png.Decode(bytes.NewReader(raw))
}
