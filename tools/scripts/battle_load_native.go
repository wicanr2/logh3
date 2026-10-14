// Original public battle save and cold load. No memory or RNG injection.
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
	"strings"

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
func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func main() {
	side := os.Args[1]
	if side != "e" && side != "f" {
		panic("e|f")
	}
	out := "/task/workplace/oracle/battle-load-r136-" + side
	must(os.Mkdir(out, 0755))
	scratch := filepath.Join(out, "scratch")
	must(os.Mkdir(scratch, 0755))
	parent := "/src/workplace/fcmd/fcmd-s3-base.state"
	p, e := os.ReadFile(parent)
	must(e)
	if digest(p) != "0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7" {
		panic("parent identity")
	}
	exe, e := os.ReadFile("/orig/GIN3.COM")
	must(e)
	if digest(exe) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		panic("program identity")
	}
	source, e := os.ReadFile("/task/tools/scripts/battle_load_native.go")
	must(e)
	must(os.WriteFile(filepath.Join(out, "observer-source.go"), source, 0644))
	m := machine.New()
	d := dos.New(m, "/orig")
	d.Install()
	must(state.Load(parent, m, d))
	d.Scratch = scratch
	defer func() { d.Close() }()
	if m.IRQ0Every != 20000 || !m.IRQ0Pinned || m.CycleClock {
		panic("parent timer")
	}
	const ds = uint32(0x54b30)
	var inputs, rows []map[string]any
	var phases []map[string]any
	complete := false
	control := func() map[string]any {
		return map[string]any{"active": m.Read8(ds + 0x8f16), "region": m.Read8(ds + 0x8f17), "strategic_turn": m.Read16(ds + 0x8f18), "battle_turn_zero": m.Read16(ds + 0x8f1a), "mode": m.Read8(ds + 0x8f20), "seat_bytes_hex": hex.EncodeToString(m.Mem[ds+0x8f20 : ds+0x8f29]), "seed_hex": hex.EncodeToString(m.Mem[ds+0x94dc : ds+0x94e0])}
	}
	shot := func(label string) {
		name := fmt.Sprintf("%02d-%s", len(rows), label)
		pix, pal := m.Indexed(), m.Palette()
		w, h := m.VideoSize()
		colors := make(color.Palette, 256)
		for i, c := range pal {
			colors[i] = color.RGBA{c[0], c[1], c[2], 255}
		}
		im := image.NewPaletted(image.Rect(0, 0, w, h), colors)
		for i, v := range pix {
			im.Pix[i] = m.VGA.DACIndex(v)
		}
		f, e := os.OpenFile(filepath.Join(out, name+".png"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		must(e)
		must(png.Encode(f, im))
		must(f.Close())
		must(os.WriteFile(filepath.Join(out, name+".bin"), pix, 0644))
		must(os.WriteFile(filepath.Join(out, name+"-mem.bin"), m.Snapshot().Mem(), 0644))
		b, e := json.Marshal(pal)
		must(e)
		must(os.WriteFile(filepath.Join(out, name+"-palette.json"), b, 0644))
		rows = append(rows, map[string]any{"name": name, "step": m.Steps, "location": fmt.Sprintf("%04X:%04X", m.CPU.Seg[cpu.CS], m.CPU.IP), "control": control(), "screen_sha256": digest(pix)})
		fmt.Println(name, m.Steps, control())
	}
	defer func() {
		result := map[string]any{"schema": "logh3.original-battle-load.v1", "parent_sha256": digest(p), "program_sha256": digest(exe), "observer_source_sha256": digest(source), "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "go_version": runtime.Version(), "uid": os.Getuid(), "irq0_every": m.IRQ0Every, "irq0_pinned": m.IRQ0Pinned, "cycle_clock": m.CycleClock, "memory_injection": false, "seed_injection": false, "inputs": inputs, "rows": rows, "complete": complete, "phases": phases, "opened": d.Opened, "writes": d.Wrote, "file_ops": d.FileOps, "exec_log": d.ExecLog, "missing": d.Missing, "unimplemented": fmt.Sprint(d.Unimplemented)}
		b, e := json.MarshalIndent(result, "", "  ")
		must(e)
		must(os.WriteFile(filepath.Join(out, "report.json"), append(b, '\n'), 0644))
	}()
	type input struct {
		at     uint64
		x, y   int
		button uint16
		word   uint16
		kind   string
	}
	var scheduled []input
	var release uint64
	var releaseButton uint16
	pending := false
	click := func(at uint64, x, y int, button uint16) {
		scheduled = append(scheduled, input{at: at, x: x, y: y, button: button, kind: "mouse"})
	}
	press := func(x, y int, button uint16, hold uint64, why string) {
		d.MoveMouse(x, y)
		d.PressMouse(int(button))
		release, releaseButton = m.Steps+hold, button
		inputs = append(inputs, map[string]any{"kind": "press", "step": m.Steps, "dos_point": []int{x, y}, "display_point": []int{x, y + 40}, "button": button, "hold_steps": hold, "why": why})
	}
	isInput := func() bool { return m.CPU.Seg[cpu.CS] == 0x12aa && m.CPU.IP == 0x02d4 && m.Read8(ds+0x8f16) == 1 }
	advance := func(limit uint64, stop func() bool, ack bool) bool {
		for m.Steps < limit {
			if stop() {
				return true
			}
			if d.Exited || m.CPU.Halted {
				return false
			}
			if release != 0 && m.Steps >= release {
				d.ReleaseMouse(int(releaseButton))
				inputs = append(inputs, map[string]any{"kind": "release", "step": m.Steps, "button": releaseButton})
				release = 0
			}
			for i := 0; i < len(scheduled); i++ {
				v := scheduled[i]
				if v.at != m.Steps {
					continue
				}
				if v.kind == "key" {
					m.PushKey(uint8(v.word >> 8))
					inputs = append(inputs, map[string]any{"kind": "irq1-key", "step": m.Steps, "scan": uint8(v.word >> 8)})
				} else {
					press(v.x, v.y, v.button, 3000000, "public menu")
				}
			}
			if ack && m.CPU.Seg[cpu.CS] == 0x3765 && m.CPU.IP == 0x006c {
				pending = true
			}
			if pending && release == 0 && m.CPU.Seg[cpu.CS] == 0x3765 && m.CPU.IP == 0x0091 {
				pending = false
				press(240, 200, 0, 100000, "public report acknowledgement")
			}
			must(m.Step())
		}
		return stop()
	}
	shot("loaded-S3-parent")
	click(301000000, 544, 85, 0)
	click(304000000, 560, 85, 0)
	click(405000000, 240, 200, 0)
	if !advance(450000000, isInput, false) {
		shot("first-input-not-reached")
		return
	}
	shot("normal-input-turn1")
	start := m.Steps
	click(start+15000000, 100, 60, 0)
	click(start+20000000, 100, 70, 0)
	if !advance(start+120000000, func() bool { return isInput() && m.Read16(ds+0x8f1a) == 1 }, true) {
		shot("second-input-not-reached")
		return
	}
	shot("normal-input-turn2")
	start = m.Steps
	click(start+1000000, 100, 60, 0)
	click(start+6000000, 100, 150, 0)
	advance(start+12000000, func() bool { return false }, false)
	shot("public-save-screen")
	click(m.Steps+1000000, 300, 112, 0)
	saved := func() bool {
		for _, name := range []string{"STELLAR", "PLANET", "ADMIRAL", "FLEET", "SHIP", "WORK"} {
			if _, e := os.Stat(filepath.Join(scratch, strings.ToLower(name)+".dt0")); e != nil {
				return false
			}
		}
		return true
	}
	if !advance(m.Steps+40000000, saved, false) {
		shot("public-save-not-written")
		return
	}
	advance(m.Steps+20000000, isInput, false)
	shot("public-save-written")
	b, e := os.ReadFile(filepath.Join(scratch, "work.dt0"))
	must(e)
	if len(b) != 230 {
		panic("saved WORK length")
	}
	rows[len(rows)-1]["saved_control_hex"] = hex.EncodeToString(b[:32])
	// Saving returns to the tactical shell without another player input.
	// Its polling PC need not equal the first turn's entry marker.
	shot("save-return-battle")
	phases = append(phases, map[string]any{"name": "normal-battle-save", "opened": d.Opened, "writes": d.Wrote, "file_ops": d.FileOps, "exec_log": d.ExecLog, "missing": d.Missing, "unimplemented": fmt.Sprint(d.Unimplemented)})
	now := d.Now
	d.Close()
	m = machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(exe))
	m.IRQ0Every = 20000
	m.IRQ0Pinned = true
	m.KeyEvery = 10000
	d = dos.New(m, "/orig")
	d.Scratch = scratch
	d.Now = now
	d.Install()
	scheduled = nil
	release = 0
	pending = false
	click(40000000, 300, 230, 0)
	click(95000000, 300, 230, 0)
	for i, word := range []uint16{0x0231, 0x0433, 0x0938, 0x1c0d} {
		scheduled = append(scheduled, input{at: 105000000 + uint64(i)*2000000, word: word, kind: "key"})
	}
	click(190000000, 300, 82, 0)
	click(240000000, 318, 168, 0)
	openedSaved := func() bool {
		for _, name := range d.Opened {
			if strings.EqualFold(filepath.Base(name), "WORK.DT0") {
				return true
			}
		}
		return false
	}
	if advance(300000000, func() bool { return openedSaved() && isInput() }, false) {
		shot("cold-load-direct-battle")
		cold := m.Steps
		// Let the public input routine finish drawing before sampling the UI.
		advance(cold+12000000, func() bool { return false }, false)
		shot("cold-load-settled")
		click(cold+15000000, 100, 60, 0)
		click(cold+20000000, 100, 70, 0)
		if advance(cold+120000000, func() bool { return isInput() && m.Read16(ds+0x8f1a) == 2 }, true) {
			shot("cold-load-next-turn")
			complete = true
		} else {
			shot("cold-load-next-turn-budget")
		}
		return
	}
	shot("cold-load-return")
	if !openedSaved() {
		return
	}
	click(301000000, 544, 85, 0)
	click(304000000, 560, 85, 0)
	if advance(500000000, func() bool { return isInput() }, true) {
		shot("cold-load-execute-battle")
		complete = true
		return
	}
	shot("cold-load-execute-budget")
}
