//go:build ignore

// 全素材呈現層抽樣。原圖PNG僅讀取；輸出只作HD取樣診斷。
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	out := flag.String("out", "", "既有工作目錄")
	flag.Parse()
	b, e := os.ReadFile("/orig/FACE.AMG")
	must(e)
	faces, e := asset.OpenAMG(b)
	must(e)
	r, e := presentation.Load("assets/hd", faces)
	must(e)
	var receipt []map[string]any
	for _, group := range []string{"strmap", "sprite", "block_0", "block_1", "block_2", "block_3", "ani_pnl", "intro", "title", "vo5", "comani", "3dblock", "comflag"} {
		var paths []string
		must(filepath.WalkDir(filepath.Join("assets/original", group), func(p string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if !d.IsDir() && filepath.Ext(p) == ".png" {
				paths = append(paths, p)
			}
			return nil
		}))
		sort.Strings(paths)
		if len(paths) == 0 {
			panic("圖片分類沒有原始PNG：" + group)
		}
		// 每類首、中、尾圖，涵蓋動畫來源的不同位置。
		indices := []int{0}
		if len(paths) > 1 {
			indices = append(indices, len(paths)/2, len(paths)-1)
		}
		for _, index := range indices {
			p := paths[index]
			f, e := os.Open(p)
			must(e)
			im, e := png.Decode(f)
			must(e)
			must(f.Close())
			pal := gfx.Palette{}
			colours := map[[3]uint8]uint8{}
			fb := gfx.NewFramebuffer(0)
			bounds := im.Bounds()
			width, height := min(gfx.Width, bounds.Dx()), min(gfx.Height, bounds.Dy())
			// 原生尺寸窗口，不縮放地圖來掩蓋內容或布局。
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					red, green, blue, _ := im.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
					key := [3]uint8{uint8(red >> 8), uint8(green >> 8), uint8(blue >> 8)}
					code, ok := colours[key]
					if !ok {
						if len(colours) >= 16 {
							panic("原圖不是16色")
						}
						code = uint8(len(colours))
						colours[key] = code
						pal[code] = [3]uint8{key[0] >> 2, key[1] >> 2, key[2] >> 2}
					}
					fb.Set(x, y, code)
				}
			}
			before := fmt.Sprintf("%x", sha256.Sum256(fb.Pix))
			frame := r.Render(fb, pal)
			changed := 0
			for y := 0; y < gfx.Height; y++ {
				for x := 0; x < gfx.Width; x++ {
					red, green, blue := pal.RGB8(int(fb.At(x, y)))
					centre := frame.RGBAAt(x*3+1, y*3+1)
					if centre.R != red || centre.G != green || centre.B != blue {
						panic("原始邏輯像素中心改變")
					}
					for dy := 0; dy < 3; dy++ {
						for dx := 0; dx < 3; dx++ {
							c := frame.RGBAAt(x*3+dx, y*3+dy)
							if c.R != red || c.G != green || c.B != blue {
								changed++
							}
						}
					}
				}
			}
			if before != fmt.Sprintf("%x", sha256.Sum256(fb.Pix)) {
				panic("呈現層改動邏輯畫布")
			}
			name := fmt.Sprintf("refine-%s-%03d", group, index)
			must(fb.WritePNG(filepath.Join(*out, name+"-original.png"), pal))
			must(r.WritePNG(filepath.Join(*out, name+"-hd.png"), fb, pal))
			receipt = append(receipt, map[string]any{"group": group, "archive_png_count": len(paths), "source": p, "source_sha256": digest(p), "source_size": []int{bounds.Dx(), bounds.Dy()}, "logical_sha256": before, "hd_changed_corners": changed, "unchanged_centres": gfx.Width * gfx.Height, "display": []int{frame.Bounds().Dx(), frame.Bounds().Dy()}, "normal_player_path": false, "scope": "source-artwork presentation sampling, not gameplay parity"})
		}
	}
	b, e = json.MarshalIndent(receipt, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(*out, "REFINEMENT.json"), append(b, '\n'), 0644))
	fmt.Println("PASS", len(receipt), "原始素材取樣：地圖、艦艇、動畫、標題與面板，邏輯中心及色號不變")
}
func digest(p string) string {
	b, e := os.ReadFile(p)
	must(e)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
