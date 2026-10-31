package presentation

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// 圖片索引不持有PNG或解碼像素；只有目前使用的圖留在有界快取。
type artworkSource struct {
	path          string
	width, height int
	w, h          int
	digest        string
	crop          image.Rectangle
	parent        string
}

const artworkCacheLimit = 64 << 20

func (r *Renderer) ensureScene(kind string) {
	if r.artDir == "" || r.sceneLoaded[kind] {
		return
	}
	if r.sceneLoaded == nil {
		r.sceneLoaded = map[string]bool{}
	}
	r.sceneLoaded[kind] = true
	var err error
	switch kind {
	case "material":
		r.material, err = loadMapMaterial(filepath.Join(r.artDir, "material"))
	case "hull":
		r.hull, err = loadMaterial(filepath.Join(r.artDir, "material"), "hull-01.png")
	case "strategy":
		err = r.loadMapDecor(filepath.Join(r.artDir, "strategy"))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "HD場景圖片退回原圖 %s：%v\n", kind, err)
	}
}

func (r *Renderer) artworkImage(key string) *image.RGBA {
	if im := r.images[key]; im != nil {
		r.touchArtwork(key)
		return im
	}
	if im := r.editorFaces[key]; im != nil {
		return im
	}
	source, ok := r.artSources[key]
	if !ok || r.failedArt[key] {
		return nil
	}
	var im *image.RGBA
	var err error
	if source.parent != "" {
		if parent := r.artworkImage(source.parent); parent != nil {
			im = resize(parent, source.w, source.h)
		} else {
			err = fmt.Errorf("人物來源不可用")
		}
	} else {
		im, err = source.decode()
	}
	if err != nil {
		if r.failedArt == nil {
			r.failedArt = make(map[string]bool)
		}
		r.failedArt[key] = true
		fmt.Fprintf(os.Stderr, "HD圖片退回原圖 %s：%v\n", key, err)
		return nil
	}
	if r.images == nil {
		r.images = make(map[string]*image.RGBA)
	}
	if r.opaque == nil {
		r.opaque = make(map[string]bool)
	}
	r.images[key], r.opaque[key] = im, im.Opaque()
	r.touchArtwork(key)
	r.artBytes += len(im.Pix)
	r.artDecodes++
	for r.artBytes > artworkCacheLimit && len(r.artUse) > 1 {
		oldest, age := "", ^uint64(0)
		for candidate, used := range r.artUse {
			if candidate != key && used < age {
				oldest, age = candidate, used
			}
		}
		if oldest == "" {
			break
		}
		r.artBytes -= len(r.images[oldest].Pix)
		delete(r.images, oldest)
		delete(r.opaque, oldest)
		delete(r.artUse, oldest)
	}
	return im
}

func (r *Renderer) touchArtwork(key string) {
	if _, ok := r.artSources[key]; !ok {
		return
	}
	if r.artUse == nil {
		r.artUse = make(map[string]uint64)
	}
	r.artClock++
	r.artUse[key] = r.artClock
}

func (s artworkSource) decode() (*image.RGBA, error) {
	info, err := os.Stat(s.path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 32<<20 {
		return nil, fmt.Errorf("檔案大小無效")
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	if s.digest != "" && fmt.Sprintf("%x", sha256.Sum256(raw)) != s.digest {
		return nil, fmt.Errorf("雜湊不同")
	}
	config, err := png.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if config.Width != s.width || config.Height != s.height {
		return nil, fmt.Errorf("圖片尺寸改變")
	}
	im, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if !s.crop.Empty() {
		if !s.crop.In(im.Bounds()) {
			return nil, fmt.Errorf("圖集裁切越界")
		}
		im = im.(interface {
			SubImage(image.Rectangle) image.Image
		}).SubImage(s.crop)
	}
	return resize(im, s.w, s.h), nil
}

func (r *Renderer) shipImage(key string) *image.RGBA {
	if im := r.ships[key]; im != nil {
		return im
	}
	return r.artworkImage("ship/" + key)
}

// ImageCacheStats只供觀測，索引數與已解碼數分開。
func (r *Renderer) ImageCacheStats() (registered, decoded, residentBytes int) {
	return len(r.artSources), r.artDecodes, r.artBytes
}
