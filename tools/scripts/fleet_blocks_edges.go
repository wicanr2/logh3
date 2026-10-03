// 2376:07FA 的受控裁切診斷；完整記錄入口與堆疊注入。
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
	"path/filepath"
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
func call(m *machine.Machine, ip uint16, args []uint16) map[string]any {
	oldCS, oldIP, oldSP := m.CPU.Seg[cpu.CS], m.CPU.IP, m.CPU.R[cpu.SP]
	sp := oldSP - uint16((len(args)+2)*2)
	words := append([]uint16{0, 0x0f00}, args...)
	changes := []map[string]any{}
	for i, w := range words {
		addr := cpu.Addr(m.CPU.Seg[cpu.SS], sp+uint16(2*i))
		changes = append(changes, map[string]any{"linear_hex": fmt.Sprintf("lin:%x", addr), "before": m.Read16(addr), "after": w})
		m.Write16(addr, w)
	}
	m.CPU.Seg[cpu.CS], m.CPU.IP, m.CPU.R[cpu.SP] = 0x2376, ip, sp
	start := m.Steps
	for m.CPU.Seg[cpu.CS] != 0x0f00 || m.CPU.IP != 0 {
		if m.Steps-start > 2000000 {
			panic("call not returned")
		}
		must(m.Step())
	}
	return map[string]any{"entry": fmt.Sprintf("2376:%04X", ip), "before_cpu": []uint16{oldCS, oldIP, oldSP},
		"after_cpu": []uint16{0x2376, ip, sp}, "ss": m.CPU.Seg[cpu.SS], "stack_injections": changes, "steps": m.Steps - start}
}
func main() {
	side := os.Args[1]
	if side != "edges-valid-a" && side != "edges-valid-b" {
		panic("side")
	}
	parent := "/out/oracle/blocks48-r70-a/scene-01.state"
	hash := "07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef"
	read(parent, hash)
	dir := "/out/oracle/blocks48-r70-" + side
	must(os.Mkdir(dir, 0755))
	cases := [][6]int{}
	for x := 8; x < 16; x++ {
		cases = append(cases, [6]int{x, 220, 16, 6, 16, 21790})
	}
	for _, x := range []int{2, 562, 568} {
		for _, y := range []int{160, 168, 385} {
			cases = append(cases, [6]int{x, y, 1, 8, 9, 23326})
		}
	}
	cases = append(cases, [6]int{245, 376, 54, 6, 16, 21790}, [6]int{245, 375, 131, 6, 16, 21790}, [6]int{245, 376, 1, 6, 16, 21790})
	reports := []map[string]any{}
	for i, c := range cases {
		m := machine.New()
		m.EmptyEnv = true
		must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
		d := dos.New(m, "/orig")
		d.Install()
		must(state.Load(parent, m, d))
		beforeSeed := m.Read16(0x54b30 + 0x94dc)
		beforeSeedHi := m.Read16(0x54b30 + 0x94de)
		setup := call(m, 0x0aa5, nil)
		args := []uint16{uint16(c[0]), uint16(c[1]), uint16(c[3]), uint16(c[4]), uint16(c[2]), uint16(c[5]), 15650}
		invoke := call(m, 0x07fa, args)
		if m.Read16(0x54b30+0x94dc) != beforeSeed || m.Read16(0x54b30+0x94de) != beforeSeedHi {
			panic("seed changed")
		}
		name := fmt.Sprintf("edge-%02d", i)
		screen := m.Indexed()
		write(filepath.Join(dir, name+"-screen.bin"), screen)
		reports = append(reports, map[string]any{"name": name, "params": c, "stack_words": append([]uint16{0, 0x0f00}, args...),
			"setup": setup, "invoke": invoke, "map_mask": m.VGA.MapMask(), "seed": uint32(beforeSeed) | uint32(beforeSeedHi)<<16,
			"screen_sha256": fmt.Sprintf("%x", sha256.Sum256(screen))})
		d.Close()
	}
	report := map[string]any{"schema": "logh3.fleet-blocks-edges.v1", "parent": parent, "parent_sha256": hash,
		"dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false,
		"command": "bash tools/fleet_blocks_oracle.sh " + side, "classification": "direct-entry clipping diagnostic; not normal player path",
		"uid": os.Getuid(), "new_inputs": []any{}, "cases": reports}
	b, e := json.MarshalIndent(report, "", "  ")
	must(e)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println("edge cases", len(cases))
}
