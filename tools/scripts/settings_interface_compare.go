//go:build ignore

// 完整原生640×480色號／RGB與規格所定矩形，不使用排除區。
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
	type rect struct{ X, Y, W, H int }
	var spec struct {
		Windows  map[string]rect
		Hotspots []struct {
			X, Y, W, H int
			Action     string
		}
	}
	data, err := os.ReadFile("docs/internals/data/screen-settings.json")
	if err != nil || json.Unmarshal(data, &spec) != nil {
		panic("無法讀設定矩形")
	}
	window := spec.Windows["overlay"]
	bug := rect{}
	for _, h := range spec.Hotspots {
		if h.Action == "bug" {
			bug = rect{h.X, h.Y, h.W, h.H}
		}
	}
	regions := []struct {
		Name string
		R    rect
	}{{"full", rect{0, 0, 640, 480}}, {"settings", window}, {"bug", bug}}
	read := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		im, err := png.Decode(f)
		if err != nil || im.Bounds() != image.Rect(0, 0, 640, 480) {
			panic("PNG不是原生640×480")
		}
		return im
	}
	write := func(path string, im image.Image) {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
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
	for _, c := range []struct{ Name, Native string }{{"settings", "settings66-r98-base-c"}, {"bug", "settings66-r98-bug-b"}} {
		nativeDir := filepath.Join("workplace", "oracle", c.Native)
		native := read(filepath.Join(nativeDir, "screen.png"))
		got := read(filepath.Join(player, c.Name+".png"))
		nb, err := os.ReadFile(filepath.Join(nativeDir, "screen.bin"))
		if err != nil {
			panic(err)
		}
		pb, err := os.ReadFile(filepath.Join(player, c.Name+".bin"))
		if err != nil || len(nb) != 307200 || len(pb) != len(nb) {
			panic("色號形狀錯誤")
		}
		idx := 0
		for i, v := range nb {
			if v != pb[i] {
				idx++
			}
		}
		diff := image.NewRGBA(got.Bounds())
		var stats []map[string]any
		for _, region := range regions {
			r := region.R
			if r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 || r.X+r.W > 640 || r.Y+r.H > 480 {
				panic("規格矩形無效")
			}
			n, maximum := 0, 0
			bbox := [4]int{640, 480, -1, -1}
			for y := r.Y; y < r.Y+r.H; y++ {
				for x := r.X; x < r.X+r.W; x++ {
					a := color.RGBAModel.Convert(native.At(x, y)).(color.RGBA)
					b := color.RGBAModel.Convert(got.At(x, y)).(color.RGBA)
					abs := func(a, b uint8) uint8 {
						if a > b {
							return a - b
						}
						return b - a
					}
					dr, dg, db := abs(a.R, b.R), abs(a.G, b.G), abs(a.B, b.B)
					if region.Name == "full" {
						diff.SetRGBA(x, y, color.RGBA{dr, dg, db, 255})
					}
					maximum = max(maximum, int(dr), int(dg), int(db))
					if dr|dg|db != 0 {
						n++
						bbox[0], bbox[1] = min(bbox[0], x), min(bbox[1], y)
						bbox[2], bbox[3] = max(bbox[2], x), max(bbox[3], y)
					}
				}
			}
			var bounds any
			if n != 0 {
				bounds = bbox
			}
			stats = append(stats, map[string]any{"name": region.Name, "rect": []int{r.X, r.Y, r.W, r.H}, "pixels": r.W * r.H,
				"same": r.W*r.H - n, "different": n, "max_channel_difference": maximum, "bbox": bounds})
			if n != 0 {
				panic(fmt.Sprintf("%s/%s差%d像素", c.Name, region.Name, n))
			}
		}
		write(filepath.Join(out, c.Name+"-original.png"), native)
		write(filepath.Join(out, c.Name+"-remake.png"), got)
		write(filepath.Join(out, c.Name+"-diff.png"), diff)
		reports = append(reports, map[string]any{"name": c.Name, "level": "same-state; 限設定靜態狀態及BUG不可選輸入",
			"normalization": "原生640×480，沒有裁切、縮放或排除區", "index_different": idx, "regions": stats})
		if idx != 0 {
			panic("色號不同")
		}
	}
	b, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(out, "comparison.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Println("設定／BUG!各307200色號及RGB零差，規格矩形亦零差")
}
