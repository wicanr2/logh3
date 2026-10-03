// 固定完整提交state，觀測傷害前後的整段321F:06FC下方場景。
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
		panic("source differs: " + p)
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
}

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "a" && os.Args[1] != "b") {
		panic("probe <a|b>")
	}
	side := os.Args[1]
	parent := "/out/oracle/combat79-r63-laser-submit-a/selected.state"
	parentHash := "af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d"
	read(parent, parentHash)
	m := machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
	d := dos.New(m, "/orig")
	d.Install()
	defer d.Close()
	must(state.Load(parent, m, d))
	dir := "/out/oracle/blocks48-r70-" + side
	if _, e := os.Stat(dir); !os.IsNotExist(e) {
		panic("refuse overwrite")
	}
	must(os.Mkdir(dir, 0755))
	start := m.Steps
	if m.IRQ0Every != 20000 || !m.IRQ0Pinned || m.CycleClock {
		panic("clock differs")
	}
	ds := uint32(0x54b30)
	frames := []map[string]any{}
	events := []event{}
	scene, serial := 0, 0
	points := map[uint16]string{0x0704: "entry", 0x0713: "background", 0x07da: "block", 0x07f3: "attack-label",
		0x0803: "defence-label", 0x080a: "present", 0x0818: "attacker-highlight", 0x0828: "target-highlight",
		0x083b: "connecting-lines", 0x0862: "attacker-arrows", 0x0889: "target-arrows", 0x088b: "complete"}
	calls := map[uint32]bool{cpu.Addr(0x321f, 0x06fc): true, cpu.Addr(0x2376, 0x07fa): true,
		cpu.Addr(0x2376, 0x0aa5): true, cpu.Addr(0x2376, 0x0b43): true, cpu.Addr(0x3870, 0x07c9): true}
	for {
		cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
		if cs == 0x181b && ip == 0x0385 {
			break
		}
		if m.Steps-start > 60000000 || d.Exited || m.CPU.Halted {
			panic("not returned")
		}
		if cs == 0x321f && ip == 0x06fc {
			scene++
			must(state.Save(filepath.Join(dir, fmt.Sprintf("scene-%02d.state", scene)), m, d))
		}
		if calls[cpu.Addr(cs, ip)] {
			words := make([]uint16, 12)
			for j := range words {
				words[j] = m.Read16(cpu.Addr(m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]+uint16(2*j)))
			}
			events = append(events, event{Location: fmt.Sprintf("%04X:%04X", cs, ip), Step: m.Steps, Seed: seed(m, ds), Args: words, AX: m.CPU.R[cpu.AX]})
		}
		if cs == 0x321f {
			if label, ok := points[ip]; ok {
				serial++
				name := fmt.Sprintf("scene-%02d-%02d-%s", scene, serial, label)
				save(m, dir, name)
				frames = append(frames, map[string]any{"name": name, "scene": scene, "stage": label, "step": m.Steps,
					"original_location": fmt.Sprintf("%04X:%04X", cs, ip), "seed": seed(m, ds)})
			}
		}
		must(m.Step())
	}
	save(m, dir, "end")
	if scene != 2 {
		panic("expected pre/postdamage scenes")
	}
	report := map[string]any{"schema": "logh3.fleet-blocks-native.v1", "command": "bash tools/fleet_blocks_oracle.sh " + side,
		"parent": "combat79-r63-laser-submit-a/selected.state", "parent_sha256": parentHash,
		"dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false,
		"classification": "restored submitted command; no new inputs/injections; not normal remake player path",
		"go_version":     runtime.Version(), "uid": os.Getuid(), "canvas": []int{640, 480}, "start_step": start, "end_step": m.Steps,
		"irq0_every": m.IRQ0Every, "irq0_pinned": m.IRQ0Pinned, "cycle_clock": m.CycleClock,
		"new_inputs": []any{}, "new_injections": []any{}, "events": events, "frames": frames}
	b, e := json.MarshalIndent(report, "", "  ")
	must(e)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println("scenes", scene, "events", len(events), "frames", len(frames), "steps", m.Steps-start)
}
