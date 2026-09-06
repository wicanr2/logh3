package gfx

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// WritePNG 把畫布依調色盤寫成 PNG。PNG 只是給人看的，
// **驗收依據是色號陣列**——顏色錯與圖形錯是兩件事，混在一張圖裡看不出來。
func (f *Framebuffer) WritePNG(path string, pal Palette) error {
	p := make(color.Palette, Colors)
	for i := range p {
		r, g, b := pal.RGB8(i)
		p[i] = color.RGBA{r, g, b, 255}
	}
	img := image.NewPaletted(image.Rect(0, 0, Width, Height), p)
	copy(img.Pix, f.Pix)
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	return png.Encode(fh, img)
}

// WriteIndex 寫出色號陣列（.bin），格式與 dosgolem 的 -dump-ega 相同。
func (f *Framebuffer) WriteIndex(path string) error {
	return os.WriteFile(path, f.Pix, 0o644)
}

// WritePalette 寫出 256×3 的 8 位元 RGB，格式與 dosgolem 的 -dump-palette 相同。
func (p Palette) WritePalette(path string) error {
	buf := make([]byte, 256*3)
	for i := 0; i < Colors; i++ {
		r, g, b := p.RGB8(i)
		buf[i*3], buf[i*3+1], buf[i*3+2] = r, g, b
	}
	return os.WriteFile(path, buf, 0o644)
}
