//go:build ignore

// HD 圖片核對表。僅建立診斷圖，不改任何素材。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	out := flag.String("out", "", "既有工作目錄")
	flag.Parse()
	var manifest struct {
		Files []struct {
			Index int
			File  string
			Name  string
		}
	}
	b, err := os.ReadFile("assets/hd/MANIFEST.json")
	must(err)
	must(json.Unmarshal(b, &manifest))
	for start := 0; start < len(manifest.Files); start += 16 {
		end := min(start+16, len(manifest.Files))
		sheet := image.NewRGBA(image.Rect(0, 0, 1000, ((end-start+3)/4)*210))
		draw.Draw(sheet, sheet.Bounds(), image.NewUniform(color.RGBA{55, 55, 55, 255}), image.Point{}, draw.Src)
		for j, r := range manifest.Files[start:end] {
			x, y := (j%4)*250, (j/4)*210
			for k, path := range []string{filepath.Join("assets/original/face", filepath.Base(r.File)), filepath.Join("assets/hd", r.File)} {
				f, e := os.Open(path)
				must(e)
				im, e := png.Decode(f)
				must(e)
				must(f.Close())
				w, h := 80, 100
				if k == 1 {
					w, h = 160, 200
				}
				dx := x
				if k == 1 {
					dx = x + 85
				}
				bounds := im.Bounds()
				for py := 0; py < h; py++ {
					for px := 0; px < w; px++ {
						sheet.Set(dx+px, y+py, im.At(bounds.Min.X+px*bounds.Dx()/w, bounds.Min.Y+py*bounds.Dy()/h))
					}
				}
			}
			fmt.Printf("%s %02d %s\n", fmt.Sprintf("review-%02d.png", start/16), r.Index, r.Name)
		}
		f, e := os.OpenFile(filepath.Join(*out, fmt.Sprintf("review-%02d.png", start/16)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		must(e)
		must(png.Encode(f, sheet))
		must(f.Close())
	}
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
