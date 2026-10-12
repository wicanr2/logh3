//go:build ignore

// 核對完整人物索引，並以原版單色空格建立原生矩形。
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	orig := flag.String("orig", "/orig", "原版目錄")
	flag.Parse()
	b, err := os.ReadFile(filepath.Join(*orig, "FACE.AMG"))
	must(err)
	faces, err := asset.OpenAMG(b)
	must(err)
	if len(faces.Entries) != 83 {
		panic("實際FACE.AMG版本不是本次已核對的83格")
	}
	blank, err := faces.Decode(82)
	must(err)
	for _, v := range blank.Pix {
		if v != 9 {
			panic("第82格不是已核對的單色空格")
		}
	}
	path := "assets/original/face/face-82.png"
	f, err := os.Open(path)
	must(err)
	source, err := png.Decode(f)
	must(err)
	must(f.Close())
	c := source.At(0, 0)
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			r, g, b, a := source.At(x, y).RGBA()
			wr, wg, wb, wa := c.RGBA()
			if r != wr || g != wg || b != wb || a != wa {
				panic("原始空白PNG不是單色")
			}
		}
	}
	destination := "assets/hd/face/face-82.png"
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		// 單色UI空格是原生矩形，沒有可以重繪的人物內容。
		panel := image.NewRGBA(image.Rect(0, 0, blank.W*presentation.Scale, blank.H*presentation.Scale))
		draw.Draw(panel, panel.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		must(err)
		must(png.Encode(out, panel))
		must(out.Close())
	}
	hd, err := presentation.Load("assets/hd", faces)
	must(err)
	if hd.Count() != len(faces.Entries) {
		panic(fmt.Sprintf("HD索引未完整：%d/%d", hd.Count(), len(faces.Entries)))
	}
	manifestPath := "assets/hd/MANIFEST.json"
	b, err = os.ReadFile(manifestPath)
	must(err)
	var manifest map[string]any
	must(json.Unmarshal(b, &manifest))
	files := manifest["files"].([]any)
	if len(files) == 82 {
		st, err := os.Stat(destination)
		must(err)
		files = append(files, map[string]any{"index": 82, "name": "空白格", "file": "face/face-82.png", "width": blank.W * presentation.Scale, "height": blank.H * presentation.Scale, "bytes": st.Size(), "sha256": digest(destination), "original_sha256": digest(path), "tool": "原生單色矩形，保持原版空格", "source": "tools/scripts/hd_complete_check.go"})
	}
	if len(files) != len(faces.Entries) {
		panic("manifest索引數量不符")
	}
	seen := map[int]bool{}
	for _, v := range files {
		r := v.(map[string]any)
		index := int(number(r["index"]))
		if seen[index] || index < 0 || index >= len(faces.Entries) {
			panic("重複或越界索引")
		}
		seen[index] = true
		p := filepath.Join("assets/hd", r["file"].(string))
		if digest(p) != r["sha256"].(string) {
			panic("生成PNG雜湊不符")
		}
		if digest(filepath.Join("assets/original/face", filepath.Base(p))) != r["original_sha256"].(string) {
			panic("原圖雜湊不符")
		}
		if index != 82 {
			if _, err := os.Stat(filepath.Join("assets/hd", r["prompt"].(string))); err != nil {
				panic(err)
			}
		}
	}
	manifest["files"] = files
	manifest["completed_portraits"] = 82
	manifest["completed_entries"] = 83
	manifest["status"] = "82 portraits restored; original blank entry preserved; non-portrait graphics refined by presentation layer, not AI redraws"
	manifest["non_portrait_method"] = "1920x1440 presentation; matching-corner refinement with original logical pixel centres and dynamic palette preserved"
	if _, err := os.Stat("assets/hd/font/MANIFEST.json"); err == nil {
		var ships struct {
			Completed int `json:"completed_models"`
		}
		data, err := os.ReadFile("assets/hd/animation/MANIFEST.json")
		must(err)
		must(json.Unmarshal(data, &ships))
		if ships.Completed != 45 {
			panic("HD艦型未完整")
		}
		manifest["status"] = "82 portraits and original blank; production original-layout A interface, semantic maps/symbols, 45 AI ship bases and original-frame animation surfaces"
		manifest["non_portrait_method"] = "1920x1440 ordered semantic presentation; original 640x480 logic and masks; native UI/font/maps/symbols/projection; 45 AI ship bases; original animation timing/events with continuous surfaces and bounded glow, not per-frame AI redraws"
		manifest["component_manifests"] = []string{"font/MANIFEST.json", "material/MANIFEST.json", "animation/MANIFEST.json"}
	}
	b, err = json.MarshalIndent(manifest, "", "  ")
	must(err)
	must(os.WriteFile(manifestPath, append(b, '\n'), 0644))
	fmt.Println("PASS 真實FACE.AMG 83格：82張人物、1張原生空格；全部PNG可解碼、雜湊與提示詞核對")
}
func digest(path string) string {
	b, e := os.ReadFile(path)
	must(e)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
func number(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case float64:
		return n
	}
	panic("無效數字")
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
