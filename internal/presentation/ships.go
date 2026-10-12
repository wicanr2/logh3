package presentation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// 圖集只提供基底；來源像素雜湊吻合才用於正式動畫。
func loadShips(dir string) (map[string]*image.RGBA, map[string][32]byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, "MANIFEST.json"))
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var m struct {
		Schema      int
		LogicalCell []int `json:"logical_cell"`
		Completed   int   `json:"completed_models"`
		Atlases     []struct {
			File          string
			Width, Height int
			SHA           string `json:"sha256"`
			Cells         []struct {
				Key        string
				X, Y, W, H int
				SHA        string `json:"pixels_sha256"`
			}
		}
	}
	if len(data) > 1<<20 {
		return nil, nil, fmt.Errorf("HD 艦型清單過大")
	}
	if err = json.Unmarshal(data, &m); err != nil {
		return nil, nil, err
	}
	if m.Schema != 1 || len(m.LogicalCell) != 2 || m.LogicalCell[0] != 256 || m.LogicalCell[1] != 102 || len(m.Atlases) > 12 {
		return nil, nil, fmt.Errorf("HD 艦型清單格式無效")
	}
	ships := make(map[string]*image.RGBA)
	hashes := make(map[string][32]byte)
	for _, a := range m.Atlases {
		if filepath.Base(a.File) != a.File || a.Width > 4096 || a.Height > 4096 || a.Width < 1024 || a.Height < 408 || len(a.Cells) > 4 {
			return nil, nil, fmt.Errorf("HD 艦型圖集尺寸無效")
		}
		raw, e := os.ReadFile(filepath.Join(dir, a.File))
		if e != nil {
			return nil, nil, e
		}
		if len(raw) > 32<<20 {
			return nil, nil, fmt.Errorf("HD 艦型圖集過大")
		}
		sum := sha256.Sum256(raw)
		if fmt.Sprintf("%x", sum) != a.SHA {
			return nil, nil, fmt.Errorf("HD 艦型圖集雜湊不符：%s", a.File)
		}
		config, e := png.DecodeConfig(bytes.NewReader(raw))
		if e != nil || config.Width != a.Width || config.Height != a.Height {
			return nil, nil, fmt.Errorf("HD 艦型圖集尺寸不符")
		}
		im, e := png.Decode(bytes.NewReader(raw))
		if e != nil {
			return nil, nil, e
		}
		if im.Bounds().Dx() != a.Width || im.Bounds().Dy() != a.Height {
			return nil, nil, fmt.Errorf("HD 艦型圖集尺寸不符")
		}
		for _, c := range a.Cells {
			key, e := hex.DecodeString(c.Key)
			if e != nil || len(key) != 2 || ships[c.Key] != nil {
				return nil, nil, fmt.Errorf("HD 艦型索引無效：%s", c.Key)
			}
			rect := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)
			if c.W < 256 || c.H < 102 || !rect.In(im.Bounds()) {
				return nil, nil, fmt.Errorf("HD 艦型裁切越界")
			}
			h, e := hex.DecodeString(c.SHA)
			if e != nil || len(h) != 32 {
				return nil, nil, fmt.Errorf("HD 艦型來源雜湊無效")
			}
			var digest [32]byte
			copy(digest[:], h)
			hashes[c.Key] = digest
			// resize接受非零原點的子圖；不改寫生成圖集。
			sub := im.(interface {
				SubImage(image.Rectangle) image.Image
			}).SubImage(rect)
			ships[c.Key] = resize(sub, 256*Scale, 102*Scale)
		}
	}
	if len(ships) != m.Completed || len(ships) > 45 {
		return nil, nil, fmt.Errorf("HD 艦型數量不符")
	}
	return ships, hashes, nil
}
