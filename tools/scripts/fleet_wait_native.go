// 第72輪：沿實際05D7場景量測輸入排隊及3765:006C等待返回。
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
func read(p, h string) []byte {
	b, e := os.ReadFile(p)
	must(e)
	if fmt.Sprintf("%x", sha256.Sum256(b)) != h {
		panic(p)
	}
	return b
}
func write(p string, b []byte) {
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(e)
	_, e = f.Write(b)
	must(e)
	must(f.Close())
}
func load(p string) (*machine.Machine, *dos.DOS) {
	m := machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
	d := dos.New(m, "/orig")
	d.Install()
	must(state.Load(p, m, d))
	return m, d
}
func main() {
	side := os.Args[1]
	if side != "a" && side != "b" {
		panic("a|b")
	}
	parent := "/out/oracle/blocks48-r70-a/scene-01.state"
	ph := "07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef"
	read(parent, ph)
	dir := "/out/oracle/wait48-r72-v2-" + side
	must(os.Mkdir(dir, 0755))
	type scenario struct {
		Name, Input string
		Phase       string
		Delay       uint64
		Duration    int
	}
	cases := []scenario{{"timeout", "", "wait", 0, -1}, {"duration-0", "", "wait", 0, 0}, {"duration-1", "", "wait", 0, 1}}
	for _, input := range []string{"left", "right", "space", "escape"} {
		cases = append(cases, scenario{input + "-queued", input, "animation", 50000, -1}, scenario{input + "-wait", input, "wait", 200000, -1})
	}
	reports := []map[string]any{}
	for ci, c := range cases {
		m, d := load(parent)
		ds := uint32(0x54b30)
		start := m.Steps
		if m.IRQ0Every != 20000 || !m.IRQ0Pinned || m.CycleClock {
			panic("clock")
		}
		inputEvents := []map[string]any{}
		send := func(release bool) {
			action := "press"
			if release {
				action = "release"
			}
			inputEvents = append(inputEvents, map[string]any{"step": m.Steps, "relative_parent_step": m.Steps - start, "kind": c.Input, "action": action, "mouse_xy": []uint16{d.Mouse.X, d.Mouse.Y}})
			switch c.Input {
			case "left":
				if release {
					d.ReleaseMouse(0)
				} else {
					d.PressMouse(0)
				}
			case "right":
				if release {
					d.ReleaseMouse(1)
				} else {
					d.PressMouse(1)
				}
			case "space":
				m.QueueScanCodes(0x39, 0xb9)
			case "escape":
				m.QueueScanCodes(0x01, 0x81)
			}
		}
		injected := false
		for m.CPU.Seg[cpu.CS] != 0x3765 || m.CPU.IP != 0x006c {
			elapsed := m.Steps - start
			if c.Phase == "animation" && c.Input != "" {
				if elapsed == c.Delay {
					send(false)
					injected = true
				}
				if elapsed == c.Delay+50000 && (c.Input == "left" || c.Input == "right") {
					send(true)
				}
			}
			if elapsed > 60_000_000 || d.Exited || m.CPU.Halted {
				panic("wait entry not reached")
			}
			must(m.Step())
		}
		waitStart := m.Steps
		ss, sp := m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]
		args := []uint16{}
		for j := 0; j < 6; j++ {
			args = append(args, m.Read16(cpu.Addr(ss, sp+uint16(2*j))))
		}
		changes := []map[string]any{}
		if c.Duration >= 0 {
			a := cpu.Addr(ss, sp+4)
			changes = append(changes, map[string]any{"linear_hex": fmt.Sprintf("lin:%x", a), "before": m.Read16(a), "after": c.Duration})
			m.Write16(a, uint16(c.Duration))
		}
		name := fmt.Sprintf("case-%02d", ci)
		write(filepath.Join(dir, name+"-entry-mem.bin"), m.Snapshot().Mem())
		write(filepath.Join(dir, name+"-entry-screen.bin"), m.Indexed())
		if ci == 0 {
			must(state.Save(filepath.Join(dir, "wait-entry.state"), m, d))
		}
		polls := []map[string]any{}
		seen := map[string]bool{}
		for m.CPU.Seg[cpu.CS] != 0x3765 || m.CPU.IP != 0x00be {
			elapsed := m.Steps - waitStart
			if c.Phase == "wait" && c.Input != "" {
				if elapsed == c.Delay {
					send(false)
					injected = true
				}
				if elapsed == c.Delay+50000 && (c.Input == "left" || c.Input == "right") {
					send(true)
				}
			}
			if m.CPU.Seg[cpu.CS] == 0x3765 && m.CPU.IP == 0x0091 {
				v := m.Read16(cpu.Addr(ss, m.CPU.R[cpu.BP]))
				timer := m.Read16(ds + 0x8efe)
				key := fmt.Sprintf("%d/%d", v, timer)
				if !seen[key] {
					polls = append(polls, map[string]any{"original_location": "3765:0091", "step": m.Steps, "event_word": v, "timer_word": timer})
					seen[key] = true
				}
			}
			if elapsed > 20_000_000 {
				panic(fmt.Sprintf("wait not returned %04X:%04X", m.CPU.Seg[cpu.CS], m.CPU.IP))
			}
			must(m.Step())
		}
		if c.Input != "" && !injected {
			panic("input was not sent")
		}
		write(filepath.Join(dir, name+"-return-mem.bin"), m.Snapshot().Mem())
		write(filepath.Join(dir, name+"-return-screen.bin"), m.Indexed())
		reports = append(reports, map[string]any{"name": name, "scenario": c, "parent_start_step": start, "wait_entry_step": waitStart, "return_step": m.Steps, "wait_steps": m.Steps - waitStart, "entry_ss": ss, "entry_sp": sp, "entry_stack_words": args, "injections": changes, "inputs": inputEvents, "polls": polls, "return_ax": m.CPU.R[cpu.AX], "timer_after": m.Read16(ds + 0x8efe), "mouse_buttons_after": d.Mouse.Buttons})
		d.Close()
	}
	b, e := json.MarshalIndent(map[string]any{"schema": "logh3.fleet-wait-native.v1", "parent": parent, "parent_sha256": ph, "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false, "command": "bash tools/fleet_wait_oracle.sh " + side, "go_version": runtime.Version(), "uid": os.Getuid(), "classification": "fixed original scene with explicit queued or wait inputs; duration0/1 are direct diagnostics; not normal player path", "irq0_every": 20000, "irq0_pinned": true, "cycle_clock": false, "cases": reports}, "", "  ")
	must(e)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println("wait cases", len(cases))
}
