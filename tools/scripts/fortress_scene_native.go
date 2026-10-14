// Original fifth-phase diagnostic, with explicit state injection and one seed.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/dos"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/internal/state"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func read(path, hash string) []byte {
	b, err := os.ReadFile(path)
	must(err)
	if digest(b) != hash {
		panic("來源身份不符 " + path)
	}
	return b
}
func write(path string, b []byte) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(err)
	_, err = f.Write(b)
	must(err)
	must(f.Close())
}
func save(m *machine.Machine, dir, name string) {
	write(filepath.Join(dir, name+"-mem.bin"), m.Snapshot().Mem())
	px, pal := m.Indexed(), m.Palette()
	write(filepath.Join(dir, name+"-screen.bin"), px)
	b, err := json.Marshal(pal)
	must(err)
	write(filepath.Join(dir, name+"-palette.json"), b)
	colors := make(color.Palette, 256)
	for i, p := range pal {
		colors[i] = color.RGBA{p[0], p[1], p[2], 255}
	}
	w, h := m.VideoSize()
	im := image.NewPaletted(image.Rect(0, 0, w, h), colors)
	for i, p := range px {
		im.Pix[i] = m.VGA.DACIndex(p)
	}
	f, err := os.OpenFile(filepath.Join(dir, name+"-screen.png"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(err)
	must(png.Encode(f, im))
	must(f.Close())
}

func main() {
	if len(os.Args) != 3 {
		panic("probe <hit|empty|off> <a|b>")
	}
	kind, side := os.Args[1], os.Args[2]
	if (kind != "hit" && kind != "empty" && kind != "off") || (side != "a" && side != "b") {
		panic("不明案例")
	}
	dir := "/task/workplace/oracle/fortress48-r133-r3-" + kind + "-" + side
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		panic("拒絕覆寫")
	}
	const parent = "/src/workplace/fcmd/eng-t5.state"
	const parentHash = "c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1"
	read(parent, parentHash)
	m := machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
	d := dos.New(m, "/orig")
	d.Install()
	defer d.Close()
	must(state.Load(parent, m, d))
	if m.IRQ0Every != 20000 || !m.IRQ0Pinned || m.CycleClock {
		panic("父狀態計時不符")
	}
	must(os.Mkdir(dir, 0755))
	start := m.Steps
	const ds = uint32(0x54b30)
	seed := func() uint32 { return uint32(m.Read16(ds+0x94dc)) | uint32(m.Read16(ds+0x94de))<<16 }
	far := func(a uint32) uint32 { return cpu.Addr(m.Read16(a+2), m.Read16(a)) }
	var events, frames, inputs, injections []map[string]any
	defer func() {
		if r := recover(); r != nil {
			save(m, dir, "failed-end")
			b, err := json.MarshalIndent(map[string]any{"error": fmt.Sprint(r), "step": m.Steps, "cs": m.CPU.Seg[cpu.CS], "ip": m.CPU.IP, "events": events, "inputs": inputs, "injections": injections}, "", "  ")
			must(err)
			write(filepath.Join(dir, "failure.json"), append(b, '\n'))
			panic(r)
		}
	}()
	set := func(a uint32, b []byte, why string) {
		before := append([]byte(nil), m.Mem[a:a+uint32(len(b))]...)
		m.WriteBytes(a, b)
		injections = append(injections, map[string]any{"address": fmt.Sprintf("lin:%X", a), "step": m.Steps, "before_hex": hex.EncodeToString(before), "bytes_hex": hex.EncodeToString(b), "reason": why})
	}
	word := func(a uint32, v int, why string) { set(a, []byte{byte(v), byte(v >> 8)}, why) }
	interesting := map[uint32]bool{}
	for _, loc := range [][2]uint16{{0x181b, 0x0385}, {0x181b, 0x03ba}, {0x181b, 0x03bf},
		{0x2264, 0x0248}, {0x2264, 0x02bb}, {0x2264, 0x02d9}, {0x2264, 0x06f5}, {0x2264, 0x080b}, {0x2264, 0x0ee3},
		{0x1dd6, 0x0152}, {0x1dd6, 0x026b}, {0x3958, 0x0041}, {0x3958, 0x01cc}, {0x348a, 0x0281},
		{0x3b74, 0x031a}, {0x3765, 0x006c}, {0x321f, 0x008a}, {0x321f, 0x0897}, {0x321f, 0x0d11}, {0x321f, 0x0df7}} {
		interesting[cpu.Addr(loc[0], loc[1])] = true
	}
	returns := map[uint32]string{}
	count := map[uint32]int{}
	injected, called := false, false
	// RE371: release the current occupation notice, then choose Execute in the
	// next tactical menu. These are normal mouse events, not CPU entry changes.
	actions := []struct {
		step uint64
		kind string
	}{
		{1139800000, "move"}, {1140000000, "press"}, {1143000000, "release"},
		{1194800000, "move"}, {1195000000, "press"}, {1198000000, "release"},
		{1209800000, "move"}, {1210000000, "press"}, {1213000000, "release"},
	}
	nextInput := 0
	save(m, dir, "parent")
	for m.Steps-start < 1200000000 {
		cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
		a := cpu.Addr(cs, ip)
		if injected && cs == 0x181b && ip == 0x03d0 {
			break
		}
		if d.Exited || m.CPU.Halted {
			panic("原版提前終止")
		}
		for nextInput < len(actions) && m.Steps >= actions[nextInput].step {
			input := actions[nextInput]
			nextInput++
			switch input.kind {
			case "move":
				d.MoveMouse(240, 200)
			case "press":
				d.PressMouse(0)
			case "release":
				d.ReleaseMouse(0)
			}
			inputs = append(inputs, map[string]any{"kind": input.kind, "step": m.Steps, "button": 0, "x": 240, "y": 200})
		}
		if cs == 0x181b && ip == 0x0385 && !injected {
			if m.CPU.Seg[cpu.DS] != 0x54b3 {
				panic("第五步DS不符")
			}
			injected = true
			planet := far(ds + 0xcff8)
			fleet := ds + 0xc074 + 31*122
			if m.Read8(fleet+1) != 1 || m.Read8(fleet+4) == 255 {
				panic("艦隊31身份不符")
			}
			set(ds+0x94dc, []byte{0x78, 0x56, 0x34, 0x12}, "第五步之前固定seed一次；不重擲")
			set(planet+0x12, []byte{1}, "將局部行星0配置為要塞，呈現診斷")
			set(planet+0x45, []byte{3}, "已知GunState=3觸發第五步要塞砲")
			word(planet+0x16, 12800, "固定已知防衛力上限，決定射程與火力")
			word(planet+0x50, 64, "正下方瞄準")
			// FLEET+74/+76 are not live flagship coordinates. Read the actual
			// SHIP pool record (RE369: 78-byte entries; +34/+36 world pixels).
			pool := far(ds + 0xcfb4)
			x, y := -1, -1
			for i := 0; i < 336; i++ {
				u := pool + uint32(i*78)
				if m.Read8(u+1) == 31 && m.Read8(u+3) == 0 && m.Read16(u+8) > 0 {
					x, y = int(m.Read16(u+0x34)), int(m.Read16(u+0x36))-60
					break
				}
			}
			if x < 0 || x >= 2048 || y < 0 || y >= 2048 {
				panic("缺合法艦隊31旗艦座標")
			}
			if kind == "empty" {
				x, y = 64, 64
			}
			word(planet+0x46, x, "明示要塞X；不改分隊位置、艘數或命令")
			word(planet+0x48, y, "明示要塞Y，已知旗艦在正下方60像素；empty改到64")
			if kind == "off" {
				word(ds+0x8f36, 0, "動畫效果關閉的公開設定欄位")
			}
			save(m, dir, "phase5-input")
		}
		if injected && interesting[a] {
			if len(events) >= 8192 || len(frames) >= 1024 {
				panic("觀測超限")
			}
			words := make([]uint16, 18)
			ss, sp := m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]
			for j := range words {
				words[j] = m.Read16(cpu.Addr(ss, sp+uint16(j*2)))
			}
			e := map[string]any{"original_location": fmt.Sprintf("%04X:%04X", cs, ip), "step": m.Steps, "seed": seed(), "stack_words": words, "ax": m.CPU.R[cpu.AX], "bp": m.CPU.R[cpu.BP], "ss": ss}
			if cs == 0x2264 && ip == 0x0248 {
				called = true
				save(m, dir, "gun-before")
				returns[cpu.Addr(words[1], words[0])] = "gun-return"
			}
			if cs == 0x3958 && ip == 0x0041 {
				ptr := cpu.Addr(words[3], words[2])
				e["comani_record_hex"] = hex.EncodeToString(m.Mem[ptr : ptr+7])
				returns[cpu.Addr(words[1], words[0])] = "comani-return"
			}
			if cs == 0x348a && ip == 0x0281 {
				returns[cpu.Addr(cs, words[0])] = "draw-frame-return"
			}
			if cs == 0x2264 && (ip == 0x06f5 || ip == 0x080b) {
				returns[cpu.Addr(words[1], words[0])] = "fortress-stage-return"
			}
			events = append(events, e)
			if (cs == 0x2264 && (ip == 0x06f5 || ip == 0x080b)) || (cs == 0x1dd6 && (ip == 0x0152 || ip == 0x026b)) || (cs == 0x3765 && ip == 0x006c) {
				count[a]++
				name := fmt.Sprintf("frame-%04x-%04x-%02d", cs, ip, count[a])
				save(m, dir, name)
				frames = append(frames, map[string]any{"name": name, "step": m.Steps, "original_location": fmt.Sprintf("%04X:%04X", cs, ip), "seed": seed()})
			}
		}
		if name, ok := returns[a]; ok {
			delete(returns, a)
			count[a]++
			stem := fmt.Sprintf("%s-%04x-%04x-%02d", name, cs, ip, count[a])
			save(m, dir, stem)
			frames = append(frames, map[string]any{"name": stem, "step": m.Steps, "original_location": fmt.Sprintf("%04X:%04X", cs, ip), "seed": seed()})
		}
		must(m.Step())
	}
	if !injected || !called || m.CPU.Seg[cpu.CS] != 0x181b || m.CPU.IP != 0x03d0 {
		panic(fmt.Sprintf("未觸發並返回第五步 %04X:%04X step%d", m.CPU.Seg[cpu.CS], m.CPU.IP, m.Steps))
	}
	save(m, dir, "end")
	r := map[string]any{"schema": "logh3.fortress-scene-native.v1", "case": kind, "command": "bash tools/fortress_scene_oracle.sh " + kind + " " + side,
		"classification": "full saved original S3 resumed; explicit phase5 fortress diagnostic; not normal player path",
		"parent":         parent, "parent_sha256": parentHash, "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false, "go_version": runtime.Version(), "uid": os.Getuid(),
		"start_step": start, "end_step": m.Steps, "end_original_location": fmt.Sprintf("%04X:%04X", m.CPU.Seg[cpu.CS], m.CPU.IP),
		"irq0_every": m.IRQ0Every, "irq0_pinned": m.IRQ0Pinned, "cycle_clock": m.CycleClock, "seed_once": true, "canvas": []int{640, 480},
		"inputs": inputs, "injections": injections, "events": events, "frames": frames}
	b, err := json.MarshalIndent(r, "", "  ")
	must(err)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println(kind, "events", len(events), "frames", len(frames), "steps", m.Steps-start)
}
