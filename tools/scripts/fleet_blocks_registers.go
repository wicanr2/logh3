// 固定原版下方場景入口，非侵入觀測 VGA 平面遮罩。
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/dos"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/internal/state"
	"os"
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
func main() {
	side := os.Args[1]
	if side != "registers-a" && side != "registers-b" {
		panic("side")
	}
	parent := "/out/oracle/blocks48-r70-a/scene-01.state"
	hash := "07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef"
	read(parent, hash)
	m := machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
	d := dos.New(m, "/orig")
	d.Install()
	defer d.Close()
	must(state.Load(parent, m, d))
	start := m.Steps
	events := []map[string]any{}
	calls := map[uint32]bool{cpu.Addr(0x2376, 0x17cb): true, cpu.Addr(0x2376, 0x1852): true,
		cpu.Addr(0x3870, 0x07c9): true, cpu.Addr(0x39d7, 0x0647): true, cpu.Addr(0x321f, 0x080a): true,
		cpu.Addr(0x321f, 0x0818): true, cpu.Addr(0x321f, 0x0828): true}
	for {
		cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
		if cs == 0x1dd6 && ip == 0x0786 {
			break
		}
		if m.Steps-start > 2000000 || d.Exited || m.CPU.Halted {
			panic("not returned")
		}
		if calls[cpu.Addr(cs, ip)] {
			gc, seq, latch := m.VGA.Regs()
			words := make([]uint16, 12)
			for j := range words {
				words[j] = m.Read16(cpu.Addr(m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]+uint16(2*j)))
			}
			events = append(events, map[string]any{"original_location": fmt.Sprintf("%04X:%04X", cs, ip), "step": m.Steps,
				"ax": m.CPU.R[cpu.AX], "gc": gc, "seq": seq, "latch": latch, "stack_words": words})
		}
		must(m.Step())
	}
	screenHash := fmt.Sprintf("%x", sha256.Sum256(m.Indexed()))
	report := map[string]any{"schema": "logh3.fleet-blocks-registers.v1", "parent": parent, "parent_sha256": hash,
		"dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false,
		"command": "bash tools/fleet_blocks_oracle.sh " + side, "new_inputs": []any{}, "new_injections": []any{},
		"start_step": start, "end_step": m.Steps, "irq0_every": m.IRQ0Every, "irq0_pinned": m.IRQ0Pinned,
		"cycle_clock": m.CycleClock, "uid": os.Getuid(), "events": events, "screen_sha256": screenHash}
	b, e := json.MarshalIndent(report, "", "  ")
	must(e)
	p := "/out/oracle/blocks48-r70-" + side + ".json"
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(e)
	_, e = f.Write(append(b, '\n'))
	must(e)
	must(f.Close())
	fmt.Println("register events", len(events), "steps", m.Steps-start, "screen", screenHash)
}
