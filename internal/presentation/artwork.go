// Package presentation 將原版邏輯畫面與核准的 HD 圖片合成。
// 遊戲座標、色號傾印與存檔由原有程式處理。
package presentation

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
)

const Scale = 3

type Renderer struct {
	images         map[string]*image.RGBA
	frame          *image.RGBA
	refine         bool
	font           *fontAtlas
	theme          bool
	classic        bool
	factionThemes  map[string]*factionTheme
	activeTheme    *factionTheme
	themeID        string
	uiMask         []uint8
	material       *image.RGBA
	terrainMasks   map[uint64][]uint8
	planets        [2]*image.RGBA
	mapPlanets     [4]*image.RGBA
	mapEmblems     [2]*image.RGBA
	opaque         map[string]bool
	paletteWords   [16][12]byte
	starCache      map[[32]byte]*image.RGBA
	hull           *image.RGBA
	grid           *image.RGBA
	projected      map[[32]byte]*projectionSprite
	animationCache map[[32]byte]*image.RGBA
	ships          map[string]*image.RGBA
	shipHashes     map[string][32]byte
	frameKey       [32]byte
	frameValid     bool
	generation     uint64
}

// Load 只載入真實 FACE.AMG 索引中的替換圖；缺少單張時保留原圖。
func Load(dir string, faces *asset.AMG) (*Renderer, error) {
	if faces == nil {
		return nil, fmt.Errorf("HD：缺少 FACE.AMG")
	}
	entries, err := os.ReadDir(filepath.Join(dir, "face"))
	if err != nil {
		return nil, fmt.Errorf("HD 圖片目錄：%w", err)
	}
	r := &Renderer{images: make(map[string]*image.RGBA), frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale)), refine: true, classic: true}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".png") {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, "face-"), ".png"))
		if err != nil || name != fmt.Sprintf("face-%02d.png", id) || id < 0 || id >= len(faces.Entries) {
			return nil, fmt.Errorf("HD 人物索引無效：%s，FACE.AMG 共%d張", name, len(faces.Entries))
		}
		original, err := faces.Decode(id)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(dir, "face", name)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.Size() > 32<<20 {
			return nil, fmt.Errorf("HD 圖片過大：%s", name)
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		config, err := png.DecodeConfig(file)
		if err == nil && (config.Width < original.W || config.Height < original.H || config.Width > 4096 || config.Height > 4096) {
			err = fmt.Errorf("尺寸不合理：%dx%d", config.Width, config.Height)
		}
		// 容許生成器將邊長取整，避免把改變構圖的長寬比當作可用原圖。
		if err == nil {
			left, right := config.Width*original.H, config.Height*original.W
			if abs(left-right)*100 > max(left, right) {
				err = fmt.Errorf("長寬比與原圖不同：%dx%d", config.Width, config.Height)
			}
		}
		var img image.Image
		if err == nil {
			_, err = file.Seek(0, 0)
		}
		if err == nil {
			img, err = png.Decode(file)
		}
		closeErr := file.Close()
		if err != nil {
			return nil, fmt.Errorf("HD 圖片 %s：%w", name, err)
		}
		if closeErr != nil {
			return nil, closeErr
		}
		r.images[fmt.Sprintf("face/%02d", id)] = resize(img, original.W*Scale, original.H*Scale)
	}
	if len(r.images) == 0 {
		return nil, fmt.Errorf("HD 目錄沒有可用人物圖：%s", dir)
	}
	r.opaque = make(map[string]bool)
	for key, im := range r.images {
		r.opaque[key] = im.Opaque()
	}
	r.font, err = loadFontAtlas(filepath.Join(dir, "font"))
	if err != nil {
		return nil, err
	}
	r.theme = r.font != nil
	r.factionThemes, err = loadFactionThemes(dir)
	if err != nil {
		return nil, err
	}
	if r.theme {
		r.material, err = loadMapMaterial(filepath.Join(dir, "material"))
		if err != nil {
			return nil, err
		}
		if err = r.loadMapDecor(filepath.Join(dir, "strategy")); err != nil {
			return nil, err
		}
		r.terrainMasks = make(map[uint64][]uint8)
		r.hull, err = loadMaterial(filepath.Join(dir, "material"), "hull-01.png")
		if err != nil {
			return nil, err
		}
		r.ships, r.shipHashes, err = loadShips(filepath.Join(dir, "animation"))
		if err != nil {
			return nil, err
		}
		r.projectionGrid()
	}
	return r, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func (r *Renderer) Count() int { return len(r.images) }

// Render 使用原圖的可見像素覆蓋範圍，後畫內容即使同色也不被 HD 圖覆寫。
func (r *Renderer) Render(fb *gfx.Framebuffer, pal gfx.Palette) *image.RGBA {
	key := r.renderKey(fb, pal)
	if r.frameValid && key == r.frameKey {
		return r.frame
	}
	r.frameKey, r.frameValid = key, true
	r.generation++
	r.themeID = fb.InterfaceTheme
	r.activeTheme = r.factionThemes[r.themeID]
	if r.activeTheme != nil && r.theme {
		if len(r.uiMask) != gfx.Width*gfx.Height {
			r.uiMask = make([]uint8, gfx.Width*gfx.Height)
		}
		clear(r.uiMask)
	} else {
		r.activeTheme = nil
	}
	if r.theme && len(fb.Commands) > 0 {
		for i := range r.paletteWords {
			red, green, blue := pal.RGB8(i)
			for j := 0; j < 3; j++ {
				copy(r.paletteWords[i][j*4:], []byte{red, green, blue, 255})
			}
		}
		r.replay(fb.Commands, pal)
		r.applyFactionUI(fb, pal)
		r.hoverControl(fb, pal)
		return r.frame
	}
	var colours [gfx.Colors][4]uint8
	for i := range colours {
		red, green, blue := pal.RGB8(i)
		colours[i] = [4]uint8{red, green, blue, 255}
	}
	for y := 0; y < gfx.Height; y++ {
		for x := 0; x < gfx.Width; x++ {
			index := fb.Pix[y*gfx.Width+x] & 15
			corners := [4]uint8{index, index, index, index}
			if r.refine {
				corners = refinedCorners(fb, x, y)
			}
			c := colours[index]
			for j := 0; j < Scale; j++ {
				o := (y*Scale+j)*r.frame.Stride + x*Scale*4
				for i := 0; i < Scale; i++ {
					copy(r.frame.Pix[o+i*4:o+i*4+4], c[:])
				}
			}
			for k, p := range [4][2]int{{0, 0}, {Scale - 1, 0}, {0, Scale - 1}, {Scale - 1, Scale - 1}} {
				if corners[k] != index {
					o := (y*Scale+p[1])*r.frame.Stride + (x*Scale+p[0])*4
					v := colours[corners[k]]
					copy(r.frame.Pix[o:o+4], v[:])
				}
			}
		}
	}
	for _, a := range fb.Artwork {
		img := r.images[a.Key]
		if img == nil || img.Bounds().Dx() != a.SourceW*Scale || img.Bounds().Dy() != a.SourceH*Scale {
			continue
		}
		for y := 0; y < a.H; y++ {
			for x := 0; x < a.W; x++ {
				if a.Visible[y*a.W+x] == 0 {
					continue
				}
				for j := 0; j < Scale; j++ {
					for i := 0; i < Scale; i++ {
						src := (y*Scale+j)*img.Stride + (x*Scale+i)*4
						dst := ((a.Y+y)*Scale+j)*r.frame.Stride + ((a.X+x)*Scale+i)*4
						alpha := uint32(img.Pix[src+3])
						for channel := 0; channel < 3; channel++ {
							r.frame.Pix[dst+channel] = uint8(uint32(img.Pix[src+channel]) + uint32(r.frame.Pix[dst+channel])*(255-alpha)/255)
						}
					}
				}
			}
		}
	}
	return r.frame
}

// resize 在載入時以雙線性取樣快取，遊戲每格不重新縮放圖片。
func resize(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	b := src.Bounds()
	for y := 0; y < h; y++ {
		fy := max(0, (float64(y)+0.5)*float64(b.Dy())/float64(h)-0.5)
		y0 := int(fy)
		y1 := min(y0+1, b.Dy()-1)
		wy := fy - float64(y0)
		for x := 0; x < w; x++ {
			fx := max(0, (float64(x)+0.5)*float64(b.Dx())/float64(w)-0.5)
			x0 := int(fx)
			x1 := min(x0+1, b.Dx()-1)
			wx := fx - float64(x0)
			var values [4][4]uint32
			for k, p := range [4]image.Point{{x0, y0}, {x1, y0}, {x0, y1}, {x1, y1}} {
				values[k][0], values[k][1], values[k][2], values[k][3] = src.At(b.Min.X+p.X, b.Min.Y+p.Y).RGBA()
			}
			weights := [4]float64{(1 - wx) * (1 - wy), wx * (1 - wy), (1 - wx) * wy, wx * wy}
			for c := 0; c < 4; c++ {
				v := 0.0
				for k := range weights {
					v += float64(values[k][c]) * weights[k]
				}
				dst.Pix[y*dst.Stride+x*4+c] = uint8(v/257 + 0.5)
			}
		}
	}
	return dst
}

func (r *Renderer) WritePNG(path string, fb *gfx.Framebuffer, pal gfx.Palette) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = png.Encode(f, r.Render(fb, pal))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
