// 固定原版入口，觀測082F／0B90／0F6D文字及視窗。診斷不取代玩家路徑。
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
	"runtime"
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
	old := []uint16{m.CPU.Seg[cpu.CS], m.CPU.IP, m.CPU.R[cpu.SP]}
	sp := old[2] - uint16(2*(len(args)+2))
	words := append([]uint16{0, 0x0f00}, args...)
	changes := []map[string]any{}
	for i, w := range words {
		a := cpu.Addr(m.CPU.Seg[cpu.SS], sp+uint16(2*i))
		changes = append(changes, map[string]any{"linear_hex": fmt.Sprintf("lin:%x", a), "before": m.Read16(a), "after": w})
		m.Write16(a, w)
	}
	m.CPU.Seg[cpu.CS], m.CPU.IP, m.CPU.R[cpu.SP] = 0x1dd6, ip, sp
	start := m.Steps
	for m.CPU.Seg[cpu.CS] != 0x0f00 || m.CPU.IP != 0 {
		if m.Steps-start > 2000000 {
			panic("not returned")
		}
		must(m.Step())
	}
	return map[string]any{"entry": fmt.Sprintf("1DD6:%04X", ip), "before_cpu": old, "injected_cpu": []uint16{0x1dd6, ip, sp}, "ss": m.CPU.Seg[cpu.SS], "stack_injections": changes, "steps": m.Steps - start}
}
func main() {
	side := os.Args[1]
	if side != "a" && side != "b" {
		panic("side")
	}
	parent := "/out/oracle/blocks48-r70-a/scene-01.state"
	hash := "07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef"
	read(parent, hash)
	dir := "/out/oracle/presentation48-r71-" + side
	must(os.Mkdir(dir, 0755))
	cases := [][5]int{}
	for _, unit := range []int{80, 82, 102, 104} {
		for _, mode := range []int{0, 1, 2} {
			for detail := 0; detail < 5; detail++ {
				cases = append(cases, [5]int{unit, mode, detail, -1, 0})
			}
		}
	}
	for weapon := 0; weapon < 4; weapon++ {
		for _, loss := range []int{0, 1, 726, 65535} {
			for _, unit := range []int{80, 102} {
				cases = append(cases, [5]int{unit, 0, 0, weapon, loss})
			}
		}
	}
	reports := []map[string]any{}
	for i, c := range cases {
		m := machine.New()
		m.EmptyEnv = true
		must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
		d := dos.New(m, "/orig")
		d.Install()
		must(state.Load(parent, m, d))
		ds := uint32(0x54b30)
		pool := cpu.Addr(m.Read16(ds+0xcfb6), m.Read16(ds+0xcfb4))
		a := pool + uint32(c[0]*78)
		seed := uint32(m.Read16(ds+0x94dc)) | uint32(m.Read16(ds+0x94de))<<16
		changes := []map[string]any{}
		poke := func(addr uint32, v byte) {
			changes = append(changes, map[string]any{"linear_hex": fmt.Sprintf("lin:%x", addr), "before": m.Read8(addr), "after": v})
			m.Write8(addr, v)
		}
		poke(ds+0x8f20, byte(c[1]))
		poke(a+0x17, byte(c[2]))
		invocations := []map[string]any{}
		if c[3] < 0 {
			invocations = append(invocations, call(m, 0x082f, []uint16{uint16(c[0])}))
		} else {
			poke(a+0x2e, byte(c[3]))
			invocations = append(invocations, call(m, 0x0b90, []uint16{uint16(a & 15), uint16(a >> 4)}))
			invocations = append(invocations, call(m, 0x0f6d, []uint16{uint16(c[4]), uint16(m.Read8(a))}))
		}
		if uint32(m.Read16(ds+0x94dc))|uint32(m.Read16(ds+0x94de))<<16 != seed {
			panic("seed changed")
		}
		name := fmt.Sprintf("case-%02d", i)
		write(filepath.Join(dir, name+"-screen.bin"), m.Indexed())
		write(filepath.Join(dir, name+"-mem.bin"), m.Snapshot().Mem())
		pal := m.Palette()
		logical := [][3]uint8{}
		for j := byte(0); j < 16; j++ {
			logical = append(logical, pal[m.VGA.DACIndex(j)])
		}
		b, e := json.Marshal(logical)
		must(e)
		write(filepath.Join(dir, name+"-palette.json"), b)
		reports = append(reports, map[string]any{"name": name, "params": c, "injections": changes, "calls": invocations, "seed": seed, "ship_linear_hex": fmt.Sprintf("lin:%x", a), "side": m.Read8(a), "cpu_mode": []byte{m.Read8(ds + 0x8f21), m.Read8(ds + 0x8f22), m.Read8(ds + 0x8f25)}, "screen_sha256": fmt.Sprintf("%x", sha256.Sum256(m.Indexed()))})
		d.Close()
	}
	b, e := json.MarshalIndent(map[string]any{"schema": "logh3.fleet-presentation-native.v1", "parent": parent, "parent_sha256": hash, "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false, "command": "bash tools/fleet_presentation_oracle.sh " + side, "go_version": runtime.Version(), "uid": os.Getuid(), "classification": "direct-entry presentation diagnostic; not normal player path", "new_inputs": []any{}, "cases": reports}, "", "  ")
	must(e)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println("cases", len(cases))
}
