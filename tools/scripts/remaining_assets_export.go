// 保存#70的五類來源素材。以正式asset解碼器匯出PNG，見docs/spec/111及docs/re/430。
// bash tools/go.sh run tools/scripts/remaining_assets_export.go -orig /orig -out workplace/assets-remaining-<run> -manifest workplace/assets-remaining-<run>.json
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
)

type record struct {
	Path         string `json:"path"`
	Source       string `json:"source"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Bytes        int    `json:"bytes"`
	SHA256       string `json:"sha256"`
	PixelsSHA256 string `json:"pixels_sha256"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func hash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func main() {
	orig := flag.String("orig", "/orig", "唯讀原版來源")
	out := flag.String("out", "", "新的輸出目錄，拒絕覆寫")
	manifest := flag.String("manifest", "", "輸入／輸出雜湊JSON")
	data := flag.String("data", "docs/internals/data", "正式畫面palette來源")
	flag.Parse()
	if *out == "" || *manifest == "" {
		panic("須指定out及manifest")
	}
	if _, e := os.Stat(*manifest); !os.IsNotExist(e) {
		panic("manifest已存在或無法核對")
	}
	must(os.Mkdir(*out, 0755))
	sources := map[string]map[string]any{}
	read := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(*orig, name))
		must(e)
		sources[name] = map[string]any{"bytes": len(b), "sha256": hash(b)}
		return b
	}
	menuName := filepath.Join(*data, "screen-main-menu.json")
	m, e := os.ReadFile(menuName)
	must(e)
	sources[menuName] = map[string]any{"bytes": len(m), "sha256": hash(m)}
	var menu struct {
		Palette gfx.Palette `json:"palette_6bit"`
	}
	must(json.Unmarshal(m, &menu))
	var records []record
	emit := func(kind, name, source string, img *asset.Image, pal gfx.Palette) {
		if img == nil || img.W <= 0 || img.H <= 0 || len(img.Pix) != img.W*img.H {
			panic("影像形狀錯誤")
		}
		palette := make(color.Palette, gfx.Colors)
		for i := range palette {
			r, g, b := pal.RGB8(i)
			palette[i] = color.RGBA{r, g, b, 255}
		}
		p := image.NewPaletted(image.Rect(0, 0, img.W, img.H), palette)
		for i, v := range img.Pix {
			if int(v) >= len(palette) {
				panic("色號越界")
			}
			p.Pix[i] = v
		}
		var encoded bytes.Buffer
		must(png.Encode(&encoded, p))
		decoded, e := png.Decode(bytes.NewReader(encoded.Bytes()))
		must(e)
		d, ok := decoded.(*image.Paletted)
		if !ok || d.Bounds() != p.Bounds() || !bytes.Equal(d.Pix, p.Pix) {
			panic("PNG重讀的形狀／索引不同")
		}
		for i := range palette {
			if d.Palette[i] != palette[i] {
				panic("PNG重讀palette不同")
			}
		}
		rel := filepath.Join(kind, name)
		path := filepath.Join(*out, rel)
		must(os.MkdirAll(filepath.Dir(path), 0755))
		f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		must(e)
		_, e = f.Write(encoded.Bytes())
		must(e)
		must(f.Close())
		records = append(records, record{rel, source, img.W, img.H, encoded.Len(), hash(encoded.Bytes()), hash(img.Pix)})
	}
	for k := 0; k < 3; k++ {
		name := fmt.Sprintf("ANI_PNL%d.A98", k)
		b := read(name)
		img, e := asset.DecodeAMGImage(b)
		must(e)
		if img.Used != len(b) {
			panic(fmt.Sprintf("%s消耗%d／%d", name, img.Used, len(b)))
		}
		emit("ani_pnl", fmt.Sprintf("panel-%d.png", k), name, img, menu.Palette)
	}
	for k := 0; k < 3; k++ {
		name := fmt.Sprintf("BLOCK_%d.BLK", k)
		tiles, e := asset.DecodeTiles(read(name), 16, 20)
		must(e)
		for i, img := range tiles {
			emit(fmt.Sprintf("block_%d", k), fmt.Sprintf("spr-%02d.png", i), name, img, menu.Palette)
		}
	}
	tiles, e := asset.DecodeTiles(read("BLOCK_3.BLK"), 16, 20)
	must(e)
	for k := 0; k < 32; k++ {
		name := fmt.Sprintf("MAP%02d.CMP", k)
		cells, e := asset.DecodeCMP(read(name), asset.BattleMapCols, asset.BattleMapRows)
		must(e)
		img := &asset.Image{W: asset.BattleMapCols * 16, H: asset.BattleMapRows * 20, Pix: make([]byte, asset.BattleMapCols*16*asset.BattleMapRows*20)}
		for i, c := range cells {
			if int(c) >= len(tiles) {
				panic(fmt.Sprintf("%s格%d圖塊%d超出%d", name, i, c, len(tiles)))
			}
			x, y := i%asset.BattleMapCols*16, i/asset.BattleMapCols*20
			for row := 0; row < 20; row++ {
				copy(img.Pix[(y+row)*img.W+x:], tiles[c].Pix[row*16:(row+1)*16])
			}
		}
		emit("map", fmt.Sprintf("map-%02d.png", k), name+" + BLOCK_3.BLK", img, menu.Palette)
	}
	intro, e := asset.NewIntroDecoder(read("IBM16.350"), read("IBMFLS16.350"))
	must(e)
	ipal, e := screen.IntroPalette(read("BU16IBM.RGB"))
	must(e)
	for i := 0; ; i++ {
		ok, e := intro.Next()
		must(e)
		if !ok {
			break
		}
		emit("intro", fmt.Sprintf("frame-%03d.png", i), "IBM16.350 + IBMFLS16.350", intro.Pixels(), ipal)
	}
	title, e := asset.DecodeTitle350(read("SPTIT.350"))
	must(e)
	tpal, e := screen.IntroPalette(read("SPTIT.RGB"))
	must(e)
	emit("title", "sptit.png", "SPTIT.350", title, tpal)
	result := map[string]any{"schema": "logh3.remaining-assets.v1", "sources": sources, "images": records, "rights": "原版解碼保存，private repo；不進一般發行包", "intro_scope": "來源374格；第349格保存但原版玩家播放略過，尾段重播另見020", "title_scope": "111圖形解碼；玩家轉場未完成"}
	b, e := json.MarshalIndent(result, "", "  ")
	must(e)
	f, e := os.OpenFile(*manifest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(e)
	_, e = f.Write(append(b, '\n'))
	must(e)
	must(f.Close())
	fmt.Printf("匯出%d張，來源%d檔；全部PNG重讀尺寸、索引及palette相同\n", len(records), len(sources))
}
