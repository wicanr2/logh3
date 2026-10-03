// 第72輪：固定原版投影入口與有界幾何診斷，不取代正常玩家路徑。
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
	f, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
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
func ptr(m *machine.Machine, a uint32) uint32 { return cpu.Addr(m.Read16(a+2), m.Read16(a)) }
func bytes(m *machine.Machine, a uint32, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = m.Read8(a + uint32(i))
	}
	return b
}
func words(m *machine.Machine, a uint32, n int) []uint16 {
	b := make([]uint16, n)
	for i := range b {
		b[i] = m.Read16(a + uint32(i*2))
	}
	return b
}
func main() {
	side := os.Args[1]
	if side != "a" && side != "b" {
		panic("a|b")
	}
	parent := "/out/oracle/combat79-r63-laser-submit-a/selected.state"
	ph := "af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d"
	read(parent, ph)
	dir := "/out/oracle/projection48-r72-" + side
	must(os.Mkdir(dir, 0755))
	m, d := load(parent)
	start := m.Steps
	for m.CPU.Seg[cpu.CS] != 0x321f || m.CPU.IP != 0x000a {
		if m.Steps-start > 30_000_000 || d.Exited || m.CPU.Halted {
			panic("projection entry not reached")
		}
		must(m.Step())
	}
	entry := filepath.Join(dir, "entry.state")
	must(state.Save(entry, m, d))
	write(filepath.Join(dir, "entry-mem.bin"), m.Snapshot().Mem())
	entryStep := m.Steps
	ss, sp := m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]
	args := words(m, cpu.Addr(ss, sp), 6)
	fleets := [2]uint32{cpu.Addr(args[3], args[2]), cpu.Addr(args[5], args[4])}
	d.Close()
	type scenario struct {
		Name   string
		Delta  [2]int
		Offset [2]int
		Inject bool
	}
	cases := []scenario{{Name: "canonical"}}
	for _, v := range []int{0, 1, 20, 199, 200, 201} {
		cases = append(cases, scenario{fmt.Sprintf("distance-%d", v), [2]int{v, 0}, [2]int{13, 17}, true})
	}
	for i, v := range [][2]int{{20, 20}, {0, 20}, {-20, 20}, {-20, 0}, {-20, -20}, {0, -20}, {20, -20}, {201, 201}, {25, 17}, {199, 1}, {200, 1}} {
		cases = append(cases, scenario{fmt.Sprintf("bearing-%02d", i), v, [2]int{13, 17}, true})
	}
	for i, v := range [][2]int{{-199, 0}, {-198, 0}, {175, 0}, {176, 0}, {0, -106}, {0, -105}, {0, -104}, {0, 104}, {0, 105}, {0, 106}, {-300, -200}, {300, 200}} {
		cases = append(cases, scenario{fmt.Sprintf("visibility-%02d", i), [2]int{20, 0}, v, true})
	}
	reports := []map[string]any{}
	for ci, c := range cases {
		m, d = load(entry)
		ds := uint32(m.CPU.Seg[cpu.DS]) * 16
		pool := ptr(m, ds+0xcfb4)
		unitIDs := []uint16{}
		fleetRecords := []map[string]any{}
		for _, f := range fleets {
			ids := []uint16{}
			for j := 0; j < int(m.Read8(f+15)); j++ {
				ids = append(ids, m.Read16(f+16+uint32(j*4)))
			}
			unitIDs = append(unitIDs, ids...)
			fleetRecords = append(fleetRecords, map[string]any{"linear_hex": fmt.Sprintf("lin:%x", f), "raw_hex": fmt.Sprintf("%x", bytes(m, f, 122)), "unit_ids": ids})
		}
		changes := []map[string]any{}
		poke := func(a uint32, v uint16) {
			changes = append(changes, map[string]any{"linear_hex": fmt.Sprintf("lin:%x", a), "before": m.Read16(a), "after": v})
			m.Write16(a, v)
		}
		if c.Inject {
			for fi, f := range fleets {
				fx, fy := 1000, 1000
				if fi == 1 {
					fx += c.Delta[0]
					fy += c.Delta[1]
				}
				for j := 0; j < int(m.Read8(f+15)); j++ {
					id := m.Read16(f + 16 + uint32(j*4))
					ux, uy := fx+j*10, fy+j*8
					if fi == 0 && j == 1 {
						ux, uy = fx+c.Offset[0], fy+c.Offset[1]
					}
					poke(pool+uint32(id)*78+0x34, uint16(ux))
					poke(pool+uint32(id)*78+0x36, uint16(uy))
				}
			}
		}
		seed := words(m, ds+0x94dc, 2)
		before := []map[string]any{}
		for _, id := range unitIDs {
			a := pool + uint32(id)*78
			before = append(before, map[string]any{"id": id, "raw_hex": fmt.Sprintf("%x", bytes(m, a, 78))})
		}
		name := fmt.Sprintf("case-%02d", ci)
		write(filepath.Join(dir, name+"-before-mem.bin"), m.Snapshot().Mem())
		checkpoints := []map[string]any{}
		runstart := m.Steps
		for {
			cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
			if cs == 0x321f && (ip == 0x002a || ip == 0x0047 || ip == 0x005e || ip == 0x0069 || ip == 0x0083) {
				n := int(m.Read16(ds + 0xd7ec))
				array := []uint16{}
				for j := 0; j < n; j++ {
					a := ptr(m, ds+0xd78e+uint32(j*4))
					if a < pool || (a-pool)%78 != 0 {
						panic("array pointer")
					}
					array = append(array, uint16((a-pool)/78))
				}
				checkpoints = append(checkpoints, map[string]any{"original_location": fmt.Sprintf("321F:%04X", ip), "step": m.Steps, "array": array, "camera_words": words(m, ds+0xd7ea, 10)})
				if ip == 0x0083 {
					break
				}
			}
			if m.Steps-runstart > 2_000_000 {
				panic(fmt.Sprintf("projection not returned %04X:%04X", cs, ip))
			}
			must(m.Step())
		}
		if fmt.Sprint(seed) != fmt.Sprint(words(m, ds+0x94dc, 2)) {
			panic("seed changed")
		}
		after := []map[string]any{}
		for _, id := range unitIDs {
			a := pool + uint32(id)*78
			after = append(after, map[string]any{"id": id, "raw_hex": fmt.Sprintf("%x", bytes(m, a, 78)), "projection_words": words(m, a+0x3e, 4)})
		}
		write(filepath.Join(dir, name+"-after-mem.bin"), m.Snapshot().Mem())
		reports = append(reports, map[string]any{"name": name, "scenario": c, "injections": changes, "before": before, "after": after, "fleets": fleetRecords, "checkpoints": checkpoints, "seed_words": seed, "steps": m.Steps - runstart})
		d.Close()
	}
	b, e := json.MarshalIndent(map[string]any{"schema": "logh3.fleet-projection-native.v1", "parent": parent, "parent_sha256": ph, "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false, "command": "bash tools/fleet_projection_oracle.sh " + side, "go_version": runtime.Version(), "uid": os.Getuid(), "classification": "fixed submitted-state projection and explicit geometry diagnostics; not normal player path", "start_step": start, "entry_step": entryStep, "entry_ss": ss, "entry_sp": sp, "entry_stack_words": args, "new_inputs": []any{}, "cases": reports}, "", "  ")
	must(e)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println("projection cases", len(cases), "entry steps", entryStep-start)
}
