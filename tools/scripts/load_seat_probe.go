//go:build ignore

// 第46輪可丟棄選單呈現；正式產品未接線，入口第46輪goal。
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/screen"
)

func main() {
	out := "workplace/round46-menu-proto2"
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		panic("輸出已存在")
	}
	if err := os.Mkdir(out, 0755); err != nil {
		panic(err)
	}
	r, err := bootstrap.Load("org_game/extracted", "docs/internals/data")
	if err != nil {
		panic(err)
	}
	g, err := game.New(r)
	if err != nil {
		panic(err)
	}
	g.Click(300, 270)
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	if g.State() != game.StateLoadGame {
		panic(g.State())
	}
	s, err := screen.LoadCommandMenuSpec("docs/internals/data/battle-command-menus.json")
	if err != nil {
		panic(err)
	}
	m := s.Menus[screen.CommandMenuFleet]
	// 原始DS:022E四筆6-byte遠指標描述，從本次固定RAM讀取，不猜字碼。
	ram, err := os.ReadFile("workplace/oracle/normal89-r46-menu-a/menu-mem.bin")
	if err != nil {
		panic(err)
	}
	raw := ram[0x54B30+0x22E : 0x54B30+0x22E+24]
	fmt.Printf("menu table raw %x\n", raw)
	// 以字串的兩個Big5字為每列；原始半形空白保留在metadata。
	var labels [][]uint16
	var strings []string
	for row := 0; row < 3; row++ {
		o := row * 6
		lin := int(binary.LittleEndian.Uint16(raw[o+4:]))*16 + int(binary.LittleEndian.Uint16(raw[o+2:]))
		var codes []uint16
		end := lin
		for ram[end] != 0 {
			end++
		}
		strings = append(strings, fmt.Sprintf("%x", ram[lin:end+1]))
		for i := lin; i < end; {
			if ram[i] >= 0x81 {
				codes = append(codes, binary.BigEndian.Uint16(ram[i:]))
				i += 2
			} else {
				i++
			}
		}
		labels = append(labels, codes)
	}
	if len(labels) != 3 {
		panic(labels)
	}
	for i := range labels {
		m.Items[i].Codes = labels[i]
	}
	s.Menus[screen.CommandMenuFleet] = m
	for sel := -1; sel < 3; sel++ {
		x, y := 300, 322
		if sel >= 0 {
			x, y = 318, 208+20*sel
		}
		g.MoveMouse(x, y)
		fb := g.Frame()
		fb.FillRect(48, r.SlotList.Top-2+10*r.SlotList.Height, 544, 20, 2)
		r.SlotList.DrawRow(fb, r.Font, r.SaveIndex, 796, 10)
		screen.DrawCommandMenu(s, fb, r.Font, r.FrameTiles, screen.CommandMenuFleet, nil, 289, 181, sel)
		r.Cursor.Draw(fb, x, y-40)
		if err := fb.WriteIndex(filepath.Join(out, fmt.Sprintf("menu%d.bin", sel))); err != nil {
			panic(err)
		}
	}
	b, _ := json.MarshalIndent(map[string]any{"classification": "disposable presentation prototype; original loaded Game path not implemented", "labels": labels, "raw_table_hex": fmt.Sprintf("%x", raw), "raw_caption_hex": strings}, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "probe.json"), append(b, '\n'), 0644); err != nil {
		panic(err)
	}
}
