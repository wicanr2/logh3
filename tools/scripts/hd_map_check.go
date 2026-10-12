//go:build ignore

// 規格131：使用正式Battle消費端驗32張地圖。這是離線診斷，非正常玩家入口。
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
	"sort"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func main() {
	out := flag.String("out", "workplace/hd125/maps-r1", "新輸出目錄")
	flag.Parse()
	if err := os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	res, err := bootstrap.Load("/orig", "docs/internals/data")
	if err != nil {
		panic(err)
	}
	hd, err := presentation.Load("assets/hd", res.Faces)
	if err != nil {
		panic(err)
	}
	used := map[int]bool{}
	var results []map[string]any
	quartets := [2]int{}
	for n, cells := range res.BattleMaps {
		if len(cells) != 128*102 {
			panic("格數不同")
		}
		for i, v := range cells {
			used[int(v)] = true
			if v == 6 || v == 22 {
				if i%128 == 127 || i+129 >= len(cells) || cells[i+1] != v+1 || cells[i+128] != v+8 || cells[i+129] != v+9 {
					panic("四格天體不一致")
				}
				k := 0
				if v == 22 {
					k = 1
				}
				quartets[k]++
			}
		}
		for _, view := range []screen.BattleView{{Col: 0, Row: 0}, {Col: 100, Row: 84}} {
			fb := screen.Battle(res.Battle, res.BattlePanel, res.BattleTiles, cells, view)
			screen.DrawMinimap(res.Battle, fb, cells, view)
			if !bytes.Equal(fb.Pix, gfx.ReplayIndexed(fb.Commands)) {
				panic("原色重播不一致")
			}
			rendered := hd.Render(fb, res.Menu.Palette())
			entry := map[string]any{"region": n, "camera": view, "cells_sha256": digest(cells), "indexed_sha256": digest(fb.Pix), "hd_rgba_sha256": digest(rendered.Pix), "indexed_replay_equal": true}
			if n == 0 || n == 13 || n == 26 || n == 31 {
				file := fmt.Sprintf("map-%02d-%d-%d.png", n, view.Col, view.Row)
				if err := hd.WritePNG(filepath.Join(*out, file), fb, res.Menu.Palette()); err != nil {
					panic(err)
				}
				entry["screenshot"] = file
			}
			results = append(results, entry)
		}
	}
	var ids []int
	for id := range used {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	if len(ids) != 53 {
		panic(fmt.Sprintf("實際使用圖塊數%d，不同於已核53",len(ids)))
	}
	atlas := gfx.NewFramebuffer(3)
	for i, id := range ids {
		x, y := 16+(i%16)*38, 60+(i/16)*80
		im := res.BattleTiles[id]
		start := atlas.BeginGroup()
		im.Draw(atlas, x, y, im.W, im.H)
		atlas.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawTerrain, X: x, Y: y, W: im.W, H: im.H, SourceW: im.W, SourceH: im.H, Source: im.Pix, Code: uint16(id), WorldX: 480, WorldY: 560})
		for k, c := range []byte(fmt.Sprintf("%02d", id)) {
			ui.DrawHalfGlyph(atlas, res.Font, x+8*k, y+23, c, 15, 0, 0, false)
		}
	}
	if err := hd.WritePNG(filepath.Join(*out, "used-tiles.png"), atlas, res.Menu.Palette()); err != nil {
		panic(err)
	}
	var sources []map[string]string
	for _, name := range []string{"BLOCK_2.BLK", "STRMAP.BIT", "STRMAP.RGN"} {
		b, err := os.ReadFile(filepath.Join("/orig", name))
		if err != nil {
			panic(err)
		}
		sources = append(sources, map[string]string{"file": name, "sha256": digest(b)})
	}
	for n := range res.BattleMaps {
		name := fmt.Sprintf("MAP%02d.CMP", n)
		b, err := os.ReadFile(filepath.Join("/orig", name))
		if err != nil {
			panic(err)
		}
		sources = append(sources, map[string]string{"file": name, "sha256": digest(b)})
	}
	metadata := map[string]any{"classification": "正式Battle消費端離線診斷；非正常玩家或原版same-state證據", "maps": len(res.BattleMaps), "cameras": len(results), "used_tiles": ids, "planet_quartets": quartets, "runs": results, "sources": sources}
	b, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(*out, "MAPS.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("PASS 32張／64鏡頭，53圖塊，天體%d／%d組，原色重播零差\n", quartets[0], quartets[1])
}
