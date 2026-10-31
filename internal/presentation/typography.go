package presentation

import (
	"bytes"
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
	files  []artworkSource
	bounds []image.Rectangle
	failed map[int]bool
	use    map[int]uint64
	clock  uint64
}

// FontStyles 只列出已完整載入的字型；原版點陣與原版強化共用原字庫。
func (r *Renderer) FontStyles() []string {
	styles := []string{"original", "classic"}
	for _, style := range []string{"serif", "sans"} {
		if r.fonts[style] != nil {
			styles = append(styles, style)
		}
	}
	return styles
}

func (r *Renderer) FontStyle() string {
	if r.fontStyle == "" {
		return "classic"
	}
	return r.fontStyle
}

func (r *Renderer) SetFontStyle(style string) error {
	switch style {
	case "original", "classic":
	case "serif", "sans":
		if r.fonts[style] == nil {
			return fmt.Errorf("缺少字型圖集：%s", style)
		}
	default:
		return fmt.Errorf("未知字型：%s", style)
	}
	if style == "serif" || style == "sans" {
		r.font = r.fonts[style]
	}
	r.fontStyle = style
	for _, font := range r.fonts {
		if font != nil && (font != r.font || style == "original" || style == "classic") {
			for i := range font.pages {
				font.pages[i] = nil
			}
			clear(font.use)
		}
	}
	r.frameValid = false
	return nil
}

func (a *fontAtlas) page(index int) *image.Gray {
	if index < 0 || index >= len(a.pages) || a.failed[index] {
		return nil
	}
	if a.use == nil {
		a.use = map[int]uint64{}
	}
	a.clock++
	a.use[index] = a.clock
	if a.pages[index] != nil {
		return a.pages[index]
	}
	if index >= len(a.files) {
		return nil
	}
	source := a.files[index]
	raw, err := checkedFile(filepath.Dir(source.path), filepath.Base(source.path), source.digest)
	var im image.Image
	if err == nil {
		im, err = png.Decode(bytes.NewReader(raw))
	}
	if err == nil && im.Bounds() != a.bounds[index] {
		err = fmt.Errorf("字型頁尺寸改變")
	}
	if err != nil {
		if a.failed == nil {
			a.failed = map[int]bool{}
		}
		a.failed[index] = true
		fmt.Fprintf(os.Stderr, "HD字型頁退回原字庫：%v\n", err)
		return nil
	}
	gray := image.NewGray(im.Bounds())
	for y := 0; y < gray.Rect.Dy(); y++ {
		for x := 0; x < gray.Rect.Dx(); x++ {
			gray.SetGray(x, y, color.GrayModel.Convert(im.At(x, y)).(color.Gray))
		}
	}
	a.pages[index] = gray
	for len(a.use) > 2 {
		oldest, age := -1, ^uint64(0)
		for i, used := range a.use {
			if i != index && used < age {
				oldest, age = i, used
			}
		}
		if oldest < 0 {
			break
		}
		a.pages[oldest] = nil
		delete(a.use, oldest)
	}
	return gray
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
	for _, page := range m.Pages {
		if filepath.Base(page.File) != page.File {
			return nil, fmt.Errorf("HD字型頁檔名無效")
		}
		path := filepath.Join(dir, page.File)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > 32<<20 {
			return nil, fmt.Errorf("字型圖集過大")
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		config, err := png.DecodeConfig(file)
		file.Close()
		if err != nil {
			return nil, err
		}
		if config.Width <= 0 || config.Height <= 0 || config.Width > 4096 || config.Height > 4096 {
			return nil, fmt.Errorf("字型圖集尺寸無效")
		}
		if hash, err := hex.DecodeString(page.SHA256); err != nil || len(hash) != 32 {
			return nil, fmt.Errorf("字型圖集雜湊格式無效")
		}
		a.pages = append(a.pages, nil)
		a.bounds = append(a.bounds, image.Rect(0, 0, config.Width, config.Height))
		a.files = append(a.files, artworkSource{path: path, digest: page.SHA256})
	}
	for _, g := range m.Glyphs {
		w := 16 * Scale
		if g.Half {
			w = 8 * Scale
		}
		if g.Page < 0 || g.Page >= len(a.pages) || g.W != w || g.H != 15*Scale || g.X < 0 || g.Y < 0 ||
			!image.Rect(g.X, g.Y, g.X+g.W, g.Y+g.H).In(a.bounds[g.Page]) || (g.Half && g.Code > 255) {
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
