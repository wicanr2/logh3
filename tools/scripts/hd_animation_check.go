//go:build ignore

// 正式解碼器及HD呈現的動畫診斷；不是正常玩家或原版same-state收據。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
	"github.com/wicanr2/logh3/internal/screen"
)

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func main() {
	out := flag.String("out", "workplace/hd125/animations-r1", "新目錄")
	flag.Parse()
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	res, err := bootstrap.Load("/orig", "docs/internals/data")
	if err != nil {
		panic(err)
	}
	if err := bootstrap.LoadIntro(res, "/orig"); err != nil {
		panic(err)
	}
	hd, err := presentation.Load("assets/hd", res.Faces)
	if err != nil {
		panic(err)
	}
	var records []map[string]any
	var manifest struct {
		Completed int `json:"completed_models"`
		Atlases   []struct {
			Cells []struct {
				Key string
				SHA string `json:"pixels_sha256"`
			}
		}
	}
	b, err := os.ReadFile("assets/hd/animation/MANIFEST.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(b, &manifest); err != nil {
		panic(err)
	}
	baseHashes := map[string]string{}
	for _, a := range manifest.Atlases {
		for _, c := range a.Cells {
			prefix, e := hex.DecodeString(c.Key)
			if e != nil {
				panic(e)
			}
			im, e := res.Comani.StaticPicture(string(prefix))
			if e != nil {
				panic(e)
			}
			if hash(im.Pix) != c.SHA {
				panic("原D0基底雜湊不符：" + c.Key)
			}
			baseHashes[c.Key] = c.SHA
		}
	}
	if len(baseHashes) != 45 || manifest.Completed != 45 {
		panic("艦型未全覆蓋")
	}
	total, segments := 0, 0
	for _, name := range res.Comani.Names() {
		frames, err := res.Comani.Frames(name)
		if err != nil {
			panic(err)
		}
		planes := asset.PlanesFrom(make([]byte, gfx.Width*gfx.Height))
		segments++
		for i, f := range frames {
			if err := f.Draw(planes, 0); err != nil {
				panic(err)
			}
			total++
			// 全段解碼；取每族首段及原始特殊艦型的首／中／尾做HD合成。
			chosen := name[1] == 'A' || name[1] == 0xDD || name[0] == 'P' || (name[2:] == "D0" && i == 0)
			if !chosen || (i != 0 && i != len(frames)/2 && i != len(frames)-1) {
				continue
			}
			pixels := planes.Pixels()
			source := make([]byte, 256*102)
			for y := 0; y < 102; y++ {
				copy(source[y*256:], pixels[y*gfx.Width:y*gfx.Width+256])
			}
			fb := gfx.NewFramebuffer(0)
			base, err := res.Comani.StaticPicture(name[:2])
			if err != nil {
				panic(err)
			}
			start := fb.BeginGroup()
			fb.DrawSource("", 8, 56, 256, 102, 256, 102, source)
			fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawAnimation, X: 8, Y: 56, W: 256, H: 102, SourceW: 256, SourceH: 102, Source: source, Key: base.ArtworkKey, Base: base.Pix})
			if !bytes.Equal(fb.Pix, gfx.ReplayIndexed(fb.Commands)) {
				panic("原色重播差異")
			}
			image := hd.Render(fb, res.FleetPresentation.Palette)
			file := fmt.Sprintf("comani-%X-%02d.png", []byte(name), i)
			if err := hd.WritePNG(filepath.Join(*out, file), fb, res.FleetPresentation.Palette); err != nil {
				panic(err)
			}
			records = append(records, map[string]any{"segment_hex": fmt.Sprintf("%X", []byte(name)), "frame": i, "source_sha256": hash(source), "hd_sha256": hash(image.Pix), "png": file, "indexed_replay_equal": true})
		}
	}
	intro := 0
	for {
		ok, err := res.Intro.Next()
		if err != nil {
			panic(err)
		}
		if !ok {
			break
		}
		if intro == 0 || intro == 108 || intro == 156 || intro == 216 || intro == 276 || intro == 348 || intro == 349 || intro == 420 {
			fb := screen.IntroScreen(screen.IntroPage(res.Intro))
			pal := res.IntroPal
			if intro >= 349 && res.IntroTailPal != nil {
				pal = *res.IntroTailPal
			}
			if !bytes.Equal(fb.Pix, gfx.ReplayIndexed(fb.Commands)) {
				panic("片頭原色重播差異")
			}
			file := fmt.Sprintf("intro-%03d.png", intro)
			if err := hd.WritePNG(filepath.Join(*out, file), fb, pal); err != nil {
				panic(err)
			}
			records = append(records, map[string]any{"intro_frame": intro, "png": file, "indexed_replay_equal": true})
		}
		intro++
	}
	sources := map[string]string{}
	for _, f := range []string{"COMANI.BMG", "IBM16.350", "IBMFLS16.350"} {
		b, err := os.ReadFile(filepath.Join("/orig", f))
		if err != nil {
			panic(err)
		}
		sources[f] = hash(b)
	}
	report := map[string]any{"classification": "原解碼器及正式HD呈現離線診斷，不替代正常遊玩", "segments": segments, "comani_frames": total, "intro_playback_frames": intro, "samples": records, "sources": sources, "hd_ship_models": len(baseHashes), "verified_base_hashes": baseHashes}
	b, err = json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(*out, "ANIMATIONS.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("PASS COMANI %d段%d格，片頭%d播放格，HD抽樣%d\n", segments, total, intro, len(records))
}
