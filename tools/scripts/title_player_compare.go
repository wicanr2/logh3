//go:build ignore

// 以Go標準PNG解碼器核原版350頁與正常Game的480画布，保留完整差分。
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		panic("用法：玩家收據目錄 比較輸出目錄")
	}
	player, out := os.Args[1], os.Args[2]
	if err := os.Mkdir(out, 0755); err != nil {
		panic(err)
	}
	read := func(p string) image.Image {
		f, err := os.Open(p)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		im, err := png.Decode(f)
		if err != nil {
			panic(err)
		}
		return im
	}
	write := func(p string, im image.Image) {
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			panic(err)
		}
		if err := png.Encode(f, im); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
	var reports []map[string]any
	for _, c := range []struct{ Name, Native string }{{"last", "title68-r97-last-a"}, {"title", "title68-r97-waits-b"}} {
		native := read(filepath.Join("workplace", "oracle", c.Native, "screen.png"))
		got := read(filepath.Join(player, c.Name+".png"))
		if native.Bounds() != image.Rect(0, 0, 640, 350) || got.Bounds() != image.Rect(0, 0, 640, 480) {
			panic("圖形形狀不符")
		}
		norm := image.NewRGBA(image.Rect(0, 0, 640, 480))
		diff := image.NewRGBA(norm.Bounds())
		ns, full, maxDiff := 0, 0, 0
		for y := 0; y < 480; y++ {
			for x := 0; x < 640; x++ {
				want := color.RGBA{0, 0, 0, 255}
				if y >= 65 && y < 415 {
					want = color.RGBAModel.Convert(native.At(x, y-65)).(color.RGBA)
				}
				norm.SetRGBA(x, y, want)
				v := color.RGBAModel.Convert(got.At(x, y)).(color.RGBA)
				abs := func(a, b uint8) uint8 {
					if a > b {
						return a - b
					}
					return b - a
				}
				r, g, b := abs(v.R, want.R), abs(v.G, want.G), abs(v.B, want.B)
				diff.SetRGBA(x, y, color.RGBA{r, g, b, 255})
				maxDiff = max(maxDiff, int(r), int(g), int(b))
				if r|g|b != 0 {
					full++
					if y >= 65 && y < 415 {
						ns++
					}
				}
			}
		}
		write(filepath.Join(out, c.Name+"-native-normalized.png"), norm)
		write(filepath.Join(out, c.Name+"-diff.png"), diff)
		nb, err := os.ReadFile(filepath.Join("workplace", "oracle", c.Native, "screen.bin"))
		if err != nil {
			panic(err)
		}
		pb, err := os.ReadFile(filepath.Join(player, c.Name+".bin"))
		if err != nil {
			panic(err)
		}
		if len(nb) != 224000 || len(pb) != 307200 {
			panic("色號形狀不符")
		}
		idx := 0
		for i, v := range nb {
			if v != pb[65*640+i] {
				idx++
			}
		}
		var bbox any
		if full != 0 {
			bbox = "nonzero; comparison failed"
		}
		report := map[string]any{"name": c.Name, "level": "same-state; 指定350頁與動畫末格", "normalization": "350頁置於第65列；上下補邊逐像素核0，無排除",
			"pixels": 307200, "same": 307200 - full, "different": full, "bbox": bbox, "max_channel_difference": maxDiff,
			"page":    map[string]any{"rect": []int{0, 65, 640, 350}, "pixels": 224000, "different": ns, "index_different": idx},
			"padding": map[string]any{"rectangles": [][]int{{0, 0, 640, 65}, {0, 415, 640, 65}}, "pixels": 83200, "different": full - ns}}
		reports = append(reports, report)
		if full != 0 || idx != 0 {
			panic(fmt.Sprintf("%s RGB差%d／色號差%d", c.Name, full, idx))
		}
	}
	b, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(out, "comparison.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Println("末格／標題：每頁224000色號、完整307200 RGB與補邊零差")
}
