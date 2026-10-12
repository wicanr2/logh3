package presentation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

type atlasGlyph struct {
	Code       uint16 `json:"code"`
	Half       bool   `json:"half"`
	Page       int    `json:"page"`
	X, Y, W, H int
}

type atlasManifest struct {
	Schema, Scale int
	LicenseFile   string `json:"license_file"`
	LicenseSHA    string `json:"license_sha256"`
	Pages         []struct {
		File   string
		SHA256 string
	}
	Glyphs []atlasGlyph
}

type fontAtlas struct {
	pages  []*image.Gray
	glyphs map[uint32]atlasGlyph
}

func glyphID(code uint16, half bool) uint32 {
	id := uint32(code)
	if half {
		id |= 1 << 16
	}
	return id
}

func checkedFile(dir, name, digest string) ([]byte, error) {
	if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
		return nil, fmt.Errorf("HD 字型檔名無效：%q", name)
	}
	p := filepath.Join(dir, name)
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 32<<20 {
		return nil, fmt.Errorf("HD 字型檔大小無效：%s", name)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != digest {
		return nil, fmt.Errorf("HD 字型雜湊不同：%s", name)
	}
	return b, nil
}

func loadFontAtlas(dir string) (*fontAtlas, error) {
	b, err := os.ReadFile(filepath.Join(dir, "MANIFEST.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m atlasManifest
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Schema != 1 || m.Scale != Scale || len(m.Pages) == 0 || len(m.Pages) > 8 || len(m.Glyphs) > 16000 {
		return nil, fmt.Errorf("HD 字型圖集格式或索引數無效")
	}
	if _, err = checkedFile(dir, m.LicenseFile, m.LicenseSHA); err != nil {
		return nil, err
	}
	a := &fontAtlas{glyphs: make(map[uint32]atlasGlyph)}
	for _, p := range m.Pages {
		if _, err = checkedFile(dir, p.File, p.SHA256); err != nil {
			return nil, err
		}
		file, err := os.Open(filepath.Join(dir, p.File))
		if err != nil {
			return nil, err
		}
		config, err := png.DecodeConfig(file)
		if err == nil && (config.Width <= 0 || config.Height <= 0 || config.Width > 4096 || config.Height > 4096) {
			err = fmt.Errorf("字型圖集過大")
		}
		var im image.Image
		if err == nil {
			_, err = file.Seek(0, 0)
		}
		if err == nil {
			im, err = png.Decode(file)
		}
		closeErr := file.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		gray := image.NewGray(im.Bounds())
		for y := 0; y < config.Height; y++ {
			for x := 0; x < config.Width; x++ {
				gray.SetGray(x, y, color.GrayModel.Convert(im.At(x, y)).(color.Gray))
			}
		}
		a.pages = append(a.pages, gray)
	}
	for _, g := range m.Glyphs {
		w := 16 * Scale
		if g.Half {
			w = 8 * Scale
		}
		if g.Page < 0 || g.Page >= len(a.pages) || g.W != w || g.H != 15*Scale || g.X < 0 || g.Y < 0 ||
			!image.Rect(g.X, g.Y, g.X+g.W, g.Y+g.H).In(a.pages[g.Page].Bounds()) || (g.Half && g.Code > 255) {
			return nil, fmt.Errorf("HD 字型索引無效：%04X", g.Code)
		}
		id := glyphID(g.Code, g.Half)
		if _, exists := a.glyphs[id]; exists {
			return nil, fmt.Errorf("HD 字型重複：%04X", g.Code)
		}
		a.glyphs[id] = g
	}
	return a, nil
}
