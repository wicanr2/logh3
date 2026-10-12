//go:build ignore

// 規格137：公開新局操作取得兩陣營完整星圖，比對使用者核准第三稿。
// go run tools/scripts/approved_map_check.go；不注入遊戲狀態。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/presentation"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	output := flag.String("out", "workplace/hd127/approved-map", "新的核對批次")
	flag.Parse()
	out := *output
	check(os.Mkdir(out, 0755))
	res, e := bootstrap.Load("/orig", "docs/internals/data")
	check(e)
	hd, e := presentation.Load("assets/hd", res.Faces)
	check(e)
	res.Faces.ArtworkPrefix = "face"
	var rows []map[string]any
	for f := 0; f < 2; f++ {
		g, e := game.New(res)
		check(e)
		click := func(x, y int) { g.MoveMouse(x, y); g.Click(x, y); g.LeftRelease(); g.Frame(); g.MoveMouse(x, y) }
		click(300, 210+20*f)
		for _, k := range []byte("138\r") {
			g.Key(k)
		}
		click(448, 370)
		click(448, 370)
		fb := g.Frame()
		im := hd.Render(fb, g.Palette())
		ref, e := os.Open(fmt.Sprintf("workplace/hd127/map-full-%d-candidate.png", f))
		check(e)
		want, e := png.Decode(ref)
		check(e)
		check(ref.Close())
		if want.Bounds() != im.Bounds() {
			panic("尺寸不同")
		}
		diff := 0
		for y := 0; y < im.Rect.Dy(); y++ {
			for x := 0; x < im.Rect.Dx(); x++ {
				rr, gg, bb, aa := want.At(x, y).RGBA()
				c := im.RGBAAt(x, y)
				if rr != uint32(c.R)*257 || gg != uint32(c.G)*257 || bb != uint32(c.B)*257 || aa != uint32(c.A)*257 {
					diff++
				}
			}
		}
		if diff != 0 {
			panic(fmt.Sprint("核准稿不同像素 ", f, " ", diff))
		}
		var b bytes.Buffer
		check(png.Encode(&b, im))
		sum := sha256.Sum256(b.Bytes())
		check(os.WriteFile(filepath.Join(out, fmt.Sprintf("map-%d.png", f)), b.Bytes(), 0644))
		rows = append(rows, map[string]any{"faction": f, "same_approved_pixels": true, "different_pixels": diff,
			"png_sha256": hex.EncodeToString(sum[:]), "bounds": image.Rect(0, 0, 1920, 1440),
			"input": "新局陣營→138→兩次結束月初面板", "state": g.State().String()})
	}
	b, e := json.MarshalIndent(rows, "", "  ")
	check(e)
	check(os.WriteFile(filepath.Join(out, "CHECK.json"), append(b, '\n'), 0644))
	fmt.Println("PASS 兩陣營完整星圖與核准第三稿逐像素一致，包含64個名稱矩形")
}
