package presentation

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/logh3/internal/asset"
)

// CheckReleasePortraits逐張核對發行複本與舊runtime的實際RGBA，不只比較檔案大小。
func CheckReleasePortraits(master, output string, faces *asset.AMG) error {
	r, err := Load(master, faces)
	if err != nil {
		return err
	}
	for key, source := range r.artSources {
		if !strings.HasPrefix(key, "face/") {
			continue
		}
		want, err := source.decode()
		if err != nil {
			return err
		}
		packed := artworkSource{path: filepath.Join(output, filepath.Base(source.path)), width: source.w, height: source.h, w: source.w, h: source.h}
		got, err := packed.decode()
		if err != nil {
			return err
		}
		if !bytes.Equal(want.Pix, got.Pix) {
			return fmt.Errorf("發行圖片像素不同：%s", key)
		}
	}
	return nil
}

// WriteReleasePortraits只建立新發行目錄，不修改原HD圖或既有封包。
func WriteReleasePortraits(dir, output string, faces *asset.AMG) error {
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		return fmt.Errorf("發行圖輸出已存在或不可存取：%s", output)
	}
	r, err := Load(dir, faces)
	if err != nil {
		return err
	}
	if err = os.Mkdir(output, 0755); err != nil {
		return err
	}
	type entry struct {
		File                    string `json:"file"`
		SourceSHA               string `json:"source_sha256"`
		SHA                     string `json:"sha256"`
		Pixels                  string `json:"rgba_sha256"`
		Size                    [2]int `json:"size"`
		BytesBefore, BytesAfter int64
	}
	var keys []string
	for key := range r.artSources {
		if strings.HasPrefix(key, "face/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var entries []entry
	for _, key := range keys {
		source := r.artSources[key]
		im, err := source.decode()
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(source.path)
		if err != nil {
			return err
		}
		name := filepath.Base(source.path)
		f, err := os.OpenFile(filepath.Join(output, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(f, im)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		packed, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			return err
		}
		entries = append(entries, entry{File: name, SourceSHA: fmt.Sprintf("%x", sha256.Sum256(raw)), SHA: fmt.Sprintf("%x", sha256.Sum256(packed)), Pixels: fmt.Sprintf("%x", sha256.Sum256(im.Pix)), Size: [2]int{im.Rect.Dx(), im.Rect.Dy()}, BytesBefore: int64(len(raw)), BytesAfter: int64(len(packed))})
	}
	b, err := json.MarshalIndent(struct {
		Format string  `json:"format"`
		Scale  int     `json:"scale"`
		Faces  []entry `json:"faces"`
	}{"logh3-release-portraits-v1", Scale, entries}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(output, "RELEASE.json"), append(b, '\n'), 0644)
}
