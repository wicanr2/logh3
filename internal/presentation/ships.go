package presentation

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// 圖集只提供基底；來源像素雜湊吻合才用於正式動畫。
func loadShips(dir string) (map[string]artworkSource, map[string][32]byte, error) {
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
	ships := make(map[string]artworkSource)
	hashes := make(map[string][32]byte)
	for _, a := range m.Atlases {
		if filepath.Base(a.File) != a.File || a.Width > 4096 || a.Height > 4096 || a.Width < 1024 || a.Height < 408 || len(a.Cells) > 4 {
			return nil, nil, fmt.Errorf("HD 艦型圖集尺寸無效")
		}
		path := filepath.Join(dir, a.File)
		info, e := os.Stat(path)
		if e != nil {
			return nil, nil, e
		}
		if !info.Mode().IsRegular() || info.Size() > 32<<20 {
			return nil, nil, fmt.Errorf("HD 艦型圖集過大")
		}
		file, e := os.Open(path)
		if e != nil {
			return nil, nil, e
		}
		config, e := png.DecodeConfig(file)
		file.Close()
		if e != nil || config.Width != a.Width || config.Height != a.Height {
			return nil, nil, fmt.Errorf("HD 艦型圖集尺寸不符")
		}
		for _, c := range a.Cells {
			key, e := hex.DecodeString(c.Key)
			_, duplicate := ships[c.Key]
			if e != nil || len(key) != 2 || duplicate {
				return nil, nil, fmt.Errorf("HD 艦型索引無效：%s", c.Key)
			}
			rect := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)
			if c.W < 256 || c.H < 102 || !rect.In(image.Rect(0, 0, a.Width, a.Height)) {
				return nil, nil, fmt.Errorf("HD 艦型裁切越界")
			}
			h, e := hex.DecodeString(c.SHA)
			if e != nil || len(h) != 32 {
				return nil, nil, fmt.Errorf("HD 艦型來源雜湊無效")
			}
			var digest [32]byte
			copy(digest[:], h)
			hashes[c.Key] = digest
			ships[c.Key] = artworkSource{path: path, width: a.Width, height: a.Height,
				w: 256 * Scale, h: 102 * Scale, digest: a.SHA, crop: rect}
		}
	}
	if len(ships) != m.Completed || len(ships) > 45 {
		return nil, nil, fmt.Errorf("HD 艦型數量不符")
	}
	return ships, hashes, nil
}
