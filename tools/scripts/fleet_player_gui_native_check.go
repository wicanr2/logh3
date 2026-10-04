//go:build ignore

// 驗完整X11原生PNG的每個像素，包含二倍視窗每個2x2區塊。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := flag.String("dir", "", "完成GUI收據的目錄")
	flag.Parse()
	var receipt struct{ Scale int }
	b, err := os.ReadFile(filepath.Join(*dir, "receipt.json"))
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(b, &receipt); err != nil {
		panic(err)
	}
	if receipt.Scale != 1 && receipt.Scale != 2 {
		panic("倍率不支援")
	}
	read := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		return img
	}
	paths, err := filepath.Glob(filepath.Join(*dir, "*-x11.png"))
	if err != nil || len(paths) != 3 {
		panic("應有三張完整X11圖")
	}
	for _, path := range paths {
		base := strings.TrimSuffix(path, "-x11.png")
		got, want := read(path), read(base+".png")
		if want.Bounds() != image.Rect(0, 0, 640, 480) || got.Bounds() != image.Rect(0, 0, 640*receipt.Scale, 480*receipt.Scale) {
			panic("形狀不符")
		}
		different, maximum := 0, 0
		bbox := [4]int{got.Bounds().Dx(), got.Bounds().Dy(), -1, -1}
		for y := 0; y < got.Bounds().Dy(); y++ {
			for x := 0; x < got.Bounds().Dx(); x++ {
				a := color.RGBAModel.Convert(got.At(x, y)).(color.RGBA)
				z := color.RGBAModel.Convert(want.At(x/receipt.Scale, y/receipt.Scale)).(color.RGBA)
				if a.R != z.R || a.G != z.G || a.B != z.B {
					different++
					bbox[0] = min(bbox[0], x)
					bbox[1] = min(bbox[1], y)
					bbox[2] = max(bbox[2], x)
					bbox[3] = max(bbox[3], y)
				}
				for _, delta := range []int{int(a.R) - int(z.R), int(a.G) - int(z.G), int(a.B) - int(z.B)} {
					if delta < 0 {
						delta = -delta
					}
					maximum = max(maximum, delta)
				}
			}
		}
		var bounds any
		if different != 0 {
			bounds = bbox
		}
		b, err := json.MarshalIndent(map[string]any{"native_pixels": got.Bounds().Dx() * got.Bounds().Dy(), "native_different": different, "native_bbox": bounds, "max_channel_difference": maximum, "scale": receipt.Scale, "method": "實際原生畫面全部像素，逐點比較邏輯PNG整數最近鄰放大，無排除"}, "", "  ")
		if err != nil {
			panic(err)
		}
		f, err := os.OpenFile(base+"-native-comparison.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			panic(err)
		}
		if _, err = f.Write(append(b, '\n')); err != nil {
			f.Close()
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
		fmt.Println(filepath.Base(base), "原生全圖", different, "／", got.Bounds().Dx()*got.Bounds().Dy())
		if different != 0 {
			os.Exit(1)
		}
	}
}
