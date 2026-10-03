// 固定已提交命令的完整原版state，觀測完整單擊scene owner；不注入新局面。
package main

import (
	"crypto/sha256"
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

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func seed(m *machine.Machine, ds uint32) uint32 {
	return uint32(m.Read16(ds+0x94dc)) | uint32(m.Read16(ds+0x94de))<<16
}
func read(p, h string) []byte {
	b, e := os.ReadFile(p)
	must(e)
	if fmt.Sprintf("%x", sha256.Sum256(b)) != h {
		panic("來源不符：" + p)
	}
	return b
}
func write(p string, b []byte) {
	f, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(e)
	_, e = f.Write(b)
	must(e)
	must(f.Close())
}
func save(m *machine.Machine, dir, name string) {
	write(filepath.Join(dir, name+"-mem.bin"), m.Snapshot().Mem())
	write(filepath.Join(dir, name+"-screen.bin"), m.Indexed())
	pal := m.Palette()
	b, e := json.Marshal(pal)
	must(e)
	write(filepath.Join(dir, name+"-palette.json"), b)
	colors := make(color.Palette, 256)
	for i, p := range pal {
		colors[i] = color.RGBA{p[0], p[1], p[2], 255}
	}
	w, h := m.VideoSize()
	im := image.NewPaletted(image.Rect(0, 0, w, h), colors)
	for i, p := range m.Indexed() {
		im.Pix[i] = m.VGA.DACIndex(p)
	}
	f, e := os.OpenFile(filepath.Join(dir, name+"-screen.png"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(e)
	must(png.Encode(f, im))
	must(f.Close())
}

type event struct {
	Location string   `json:"original_location"`
	Step     uint64   `json:"step"`
	Seed     uint32   `json:"seed"`
	Args     []uint16 `json:"stack_words"`
	AX       uint16   `json:"ax"`
	Record   []uint8  `json:"comani_record,omitempty"`
}

func main() {
	if len(os.Args) != 3 {
		panic("probe <laser|missile|same-target|death> <a|b>")
	}
	kind, side := os.Args[1], os.Args[2]
	if side != "a" && side != "b" {
		panic("run must be a/b")
	}
	parents := map[string][2]string{
		"laser":       {"combat79-r63-laser-submit-a/selected.state", "af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d"},
		"missile":     {"combat79-r63-missile-submit-a/selected.state", "1b84dd2f0bf76cfd5b526d4df18d7dcbc1f7330478c35e2d97844f6dfaf0457a"},
		"same-target": {"sequence79-r66-same-target-second-a/submitted.state", "5722e421d007a050c37788577d31119b19876b8e27e3405a674f355a39a30d82"},
		"death":       {"sequence79-r66-same-target-second-a/submitted.state", "5722e421d007a050c37788577d31119b19876b8e27e3405a674f355a39a30d82"},
	}
	p, ok := parents[kind]
	if !ok {
		panic("unknown case")
	}
	parent := filepath.Join("/out/oracle", p[0])
	read(parent, p[1])
	m := machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
	d := dos.New(m, "/orig")
	d.Install()
	defer d.Close()
	must(state.Load(parent, m, d))
	dir := filepath.Join("/out/oracle", "scene48-r69-v2-"+kind+"-"+side)
	if _, e := os.Stat(dir); !os.IsNotExist(e) {
		panic("refuse overwrite")
	}
	must(os.Mkdir(dir, 0755))
	start := m.Steps
	if m.IRQ0Every != 20000 || !m.IRQ0Pinned || m.CycleClock {
		panic("父狀態的虛擬計時與原版收據不符")
	}
	ds := uint32(0x54b30)
	injections := []map[string]any{}
	classification := "restored submitted-command scene; no new injection; not normal player start"
	if kind == "death" {
		classification = "restored submitted-command scene; explicit death injection; not normal player start"
		base := cpu.Addr(m.Read16(ds+0xcfb6), m.Read16(ds+0xcfb4))
		addr := base + 102*78 + 8
		before := m.Read16(addr)
		if before != 3000 {
			panic("死亡反例父狀態的目標艘數不符")
		}
		m.Write16(addr, 1)
		injections = append(injections, map[string]any{"address": fmt.Sprintf("lin:%x", addr), "step": start,
			"before_hex": "b80b", "bytes_hex": "0100", "reason": "固定目標102為1艘，核對首擊死亡後第二命令跳過；不重擲seed"})
	}
	var events []event
	var frames []map[string]any
	save(m, dir, "start")
	interesting := map[uint32]bool{}
	for _, ip := range []uint16{0x0152, 0x026a, 0x026b, 0x030f, 0x05d7, 0x0686, 0x06a0, 0x06db, 0x0775, 0x0786, 0x079e, 0x07b1, 0x07d8, 0x0810, 0x0819, 0x082f, 0x0a17, 0x0b90, 0x0c54, 0x0f6d} {
		interesting[cpu.Addr(0x1dd6, ip)] = true
	}
	for _, ip := range []uint16{0x000a, 0x01ca, 0x0223, 0x0392, 0x0513, 0x06fc, 0x0897, 0x0a08, 0x0af1, 0x0b64, 0x0c37, 0x0d11} {
		interesting[cpu.Addr(0x321f, ip)] = true
	}
	for _, loc := range [][2]uint16{{0x3765, 0x006c}, {0x3958, 0x0041}, {0x3958, 0x01cc}, {0x3cd5, 0x000c}, {0x3b74, 0x031a}} {
		interesting[cpu.Addr(loc[0], loc[1])] = true
	}
	framePoints := map[uint16]bool{}
	for _, ip := range []uint16{0x026a, 0x05d7, 0x06a0, 0x06db, 0x0775, 0x0786, 0x079e, 0x07b1, 0x07d8, 0x0810, 0x0819, 0x026b} {
		framePoints[ip] = true
	}
	counts := map[uint16]int{}
	budget := uint64(60_000_000)
	if kind == "same-target" || kind == "death" {
		budget = 120_000_000 // 既有原版收據590146747，父state507999999。
	}
	for {
		cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
		if cs == 0x1dd6 && ip == 0x05d7 && m.CPU.Seg[cpu.DS] != 0x54b3 {
			panic("scene owner DS differs")
		}
		if cs == 0x181b && ip == 0x0385 {
			break
		}
		if m.Steps-start > budget || d.Exited || m.CPU.Halted {
			panic(fmt.Sprintf("scene not returned %04X:%04X step%d", cs, ip, m.Steps))
		}
		if interesting[cpu.Addr(cs, ip)] {
			words := make([]uint16, 12)
			for j := range words {
				words[j] = m.Read16(cpu.Addr(m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]+uint16(2*j)))
			}
			ev := event{Location: fmt.Sprintf("%04X:%04X", cs, ip), Step: m.Steps, Seed: seed(m, ds), Args: words, AX: m.CPU.R[cpu.AX]}
			if cs == 0x3958 && ip == 0x0041 {
				ptr := cpu.Addr(words[3], words[2])
				for j := uint32(0); j < 7; j++ {
					ev.Record = append(ev.Record, m.Read8(ptr+j))
				}
			}
			events = append(events, ev)
		}
		if cs == 0x1dd6 && framePoints[ip] {
			counts[ip]++
			name := fmt.Sprintf("frame-%04x-%02d", ip, counts[ip])
			save(m, dir, name)
			frames = append(frames, map[string]any{"name": name, "step": m.Steps, "original_location": fmt.Sprintf("%04X:%04X", cs, ip), "seed": seed(m, ds)})
		}
		must(m.Step())
	}
	save(m, dir, "end")
	b, e := json.MarshalIndent(map[string]any{"schema": "logh3.fleet-scene-native.v1", "case": kind,
		"command": "bash tools/fleet_scene_oracle.sh " + kind + " " + side, "parent": p[0], "parent_sha256": p[1],
		"classification": classification,
		"dosgolem_head":  "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false,
		"go_version": runtime.Version(), "uid": os.Getuid(), "start_step": start, "end_step": m.Steps,
		"irq0_every": m.IRQ0Every, "irq0_pinned": m.IRQ0Pinned, "cycle_clock": m.CycleClock,
		"canvas": []int{640, 480}, "new_inputs": []any{}, "new_injections": injections, "events": events, "frames": frames}, "", "  ")
	must(e)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println(kind, "events", len(events), "frames", len(frames), "steps", m.Steps-start)
}
