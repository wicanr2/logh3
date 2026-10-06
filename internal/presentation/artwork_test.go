package presentation

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
)

func TestOcclusionClippingAndSourceCrop(t *testing.T) {
	// 左側被裁掉一格，右側只畫出兩格；縮放仍以四格原圖為基準。
	hd := image.NewRGBA(image.Rect(0, 0, 4*Scale, 2*Scale))
	for y := 0; y < hd.Bounds().Dy(); y++ {
		for x := 0; x < hd.Bounds().Dx(); x++ {
			hd.SetRGBA(x, y, color.RGBA{uint8(x + 30), 50, 80, 255})
		}
	}
	r := &Renderer{images: map[string]*image.RGBA{"face/00": hd}, frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale))}
	fb := gfx.NewFramebuffer(0)
	img := &asset.Image{W: 4, H: 2, Pix: make([]uint8, 8), ArtworkKey: "face/00"}
	img.Draw(fb, -1, 10, 3, 2)
	fb.Set(1, 10, 0) // 同色覆蓋也必須遮住 HD。
	pal := gfx.Palette{}
	pal[0] = [3]uint8{0, 0, 0}
	frame := r.Render(fb, pal)
	if got := frame.RGBAAt(0, 10*Scale); got.R != 33 {
		t.Fatalf("cropped source shifted: %v", got)
	}
	if got := frame.RGBAAt(Scale, 10*Scale); got != (color.RGBA{0, 0, 0, 255}) {
		t.Fatalf("same-colour occlusion lost: %v", got)
	}
	if got := frame.RGBAAt(2*Scale, 10*Scale); got.R != 0 {
		t.Fatalf("HD escaped cropped rectangle: %v", got)
	}
	if got := frame.RGBAAt(Scale, 11*Scale); got.R != 36 {
		t.Fatalf("uncovered second row lost: %v", got)
	}
	if frame.Bounds() != image.Rect(0, 0, 1920, 1440) {
		t.Fatal(frame.Bounds())
	}
	for _, v := range fb.Pix {
		if v != 0 {
			t.Fatal("presentation changed indexed game frame")
		}
	}
}

func TestLaterImageAndRowReplaceArtwork(t *testing.T) {
	fb := gfx.NewFramebuffer(0)
	first := &asset.Image{W: 2, H: 2, Pix: make([]uint8, 4), ArtworkKey: "face/00"}
	second := &asset.Image{W: 1, H: 1, Pix: []uint8{0}, ArtworkKey: "face/01"}
	first.Draw(fb, 5, 5, 2, 2)
	second.Draw(fb, 6, 6, 1, 1)
	if fb.Artwork[0].Visible[3] != 0 || fb.Artwork[1].Visible[0] != 1 {
		t.Fatal("image order lost")
	}
	fb.WriteRow(5, 5, []uint8{0, 0})
	if fb.Artwork[0].Visible[0] != 0 || fb.Artwork[0].Visible[1] != 0 {
		t.Fatal("row occlusion lost")
	}
	fb.CopyPixels(make([]uint8, gfx.Width*gfx.Height))
	if len(fb.Artwork) != 0 {
		t.Fatal("copy retained old layer")
	}
	first.Draw(fb, 5, 5, 2, 2)
	fb.Clear(1)
	if len(fb.Artwork) != 0 {
		t.Fatal("clear retained old layer")
	}
}

func TestMissingImageUsesOriginalFrame(t *testing.T) {
	r := &Renderer{images: map[string]*image.RGBA{}, frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale))}
	fb := gfx.NewFramebuffer(3)
	(&asset.Image{W: 1, H: 1, Pix: []uint8{3}, ArtworkKey: "face/04"}).Draw(fb, 10, 10, 1, 1)
	pal := gfx.Palette{}
	pal[3] = [3]uint8{0, 63, 0}
	if got := r.Render(fb, pal).RGBAAt(30, 30); got != (color.RGBA{0, 255, 0, 255}) {
		t.Fatal(got)
	}
}

func TestLoadRejectsMissingCorruptAndOutOfRange(t *testing.T) {
	faces, err := asset.OpenAMG(makeAMG())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "face")
	if err := os.Mkdir(p, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir, faces); err == nil {
		t.Fatal("empty library accepted")
	}
	name := filepath.Join(p, "face-00.png")
	if err := os.WriteFile(name, []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir, faces); err == nil {
		t.Fatal("corrupt PNG accepted")
	}
	writePNG(t, name, image.NewRGBA(image.Rect(0, 0, 16, 4)))
	r, err := Load(dir, faces)
	if err != nil || r.Count() != 1 {
		t.Fatalf("valid library: %v", err)
	}
	if err := os.Rename(name, filepath.Join(p, "face-01.png")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir, faces); err == nil {
		t.Fatal("index outside real archive accepted")
	}
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

// 一筆 8×2 的實際 AMG 編碼格式，用於測試載入器的實際索引上限。
func makeAMG() []byte {
	return []byte{20, 0, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 1, 2, 1, 2, 1, 2, 1, 2}
}
