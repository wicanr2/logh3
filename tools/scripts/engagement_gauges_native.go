// 第68輪原版量表局部診斷。複製到dosgolem的gitignore workplace後編譯，
// 沿原版完整state到自然D11入口；只對兩個已定位的far常式作明示直接呼叫。
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

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func read(path, digest string) []byte {
	b, err := os.ReadFile(path)
	must(err)
	if fmt.Sprintf("%x", sha256.Sum256(b)) != digest {
		panic("輸入雜湊不符：" + path)
	}
	return b
}
func write(path string, data []byte) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(err)
	_, err = f.Write(data)
	must(err)
	must(f.Close())
}
func savePNG(m *machine.Machine, path string) {
	w, h := m.VideoSize()
	px, pal := m.Indexed(), m.Palette()
	colors := make(color.Palette, 256)
	for i, p := range pal {
		colors[i] = color.RGBA{p[0], p[1], p[2], 255}
	}
	img := image.NewPaletted(image.Rect(0, 0, w, h), colors)
	for i, p := range px {
		img.Pix[i] = m.VGA.DACIndex(p)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(err)
	must(png.Encode(f, img))
	must(f.Close())
}

type rectCall struct {
	Location string   `json:"original_location"`
	Step     uint64   `json:"step"`
	Args     []uint16 `json:"args"`
}
type gaugeCase struct {
	Name       string           `json:"case"`
	Kind       string           `json:"kind"`
	Faction    int              `json:"fleet_faction"`
	Value      int              `json:"value"`
	Divisor    int              `json:"divisor"`
	Height     int              `json:"native_height"`
	Args       []uint16         `json:"args"`
	CallSteps  uint64           `json:"call_steps"`
	Rects      []rectCall       `json:"rectangle_calls"`
	Injections []map[string]any `json:"injections"`
}

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "a" && os.Args[1] != "b") {
		panic("用法：probe <a|b>")
	}
	out := filepath.Join("/out/oracle", "gauges48-r68-"+os.Args[1])
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		panic("拒絕覆寫既有原版量表輸出")
	}
	parent := "/out/oracle/combat79-r63-laser-submit-a/selected.state"
	read(parent, "af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d")
	exe := read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")
	m := machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(exe))
	d := dos.New(m, "/orig")
	d.Install()
	defer d.Close()
	must(state.Load(parent, m, d))
	start := m.Steps
	for m.CPU.Seg[cpu.CS] != 0x321f || m.CPU.IP != 0x0d11 {
		if m.Steps-start >= 10_000_000 || d.Exited || m.CPU.Halted {
			panic(fmt.Sprintf("未到自然量表入口：%04X:%04X step%d", m.CPU.Seg[cpu.CS], m.CPU.IP, m.Steps))
		}
		must(m.Step())
	}
	if w, h := m.VideoSize(); w != 640 || h != 480 || m.CPU.Seg[cpu.DS] != 0x54b3 {
		panic("原版畫布或DS不符")
	}
	ss, sp := m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]
	fleetOff, fleetSeg := m.Read16(cpu.Addr(ss, sp+4)), m.Read16(cpu.Addr(ss, sp+6))
	fleet := cpu.Addr(fleetSeg, fleetOff)
	if fleet != 0x6115c {
		panic(fmt.Sprintf("自然D11參數不是艦隊12：%x", fleet))
	}
	// 原星域PLANET切片指標，取第一筆作DF7局部診斷輸入。
	ds := uint32(m.CPU.Seg[cpu.DS]) * 16
	planetOff, planetSeg := m.Read16(ds+0xcff8), m.Read16(ds+0xcffa)
	planet := cpu.Addr(planetSeg, planetOff)
	if planet == 0 || planet+82 >= 0xa0000 {
		panic("原版PLANET切片無效")
	}
	must(os.Mkdir(out, 0755))
	base := m.Snapshot()
	write(filepath.Join(out, "base-mem.bin"), base.Mem())
	write(filepath.Join(out, "base-screen.bin"), m.Indexed())
	palBytes, err := json.Marshal(m.Palette())
	must(err)
	write(filepath.Join(out, "base-palette.json"), palBytes)
	savePNG(m, filepath.Join(out, "base-screen.png"))
	var cases []gaugeCase
	for faction := 0; faction < 2; faction++ {
		for _, value := range []int{0, 1, 15, 33, 50, 99, 100} {
			cases = append(cases, gaugeCase{Name: fmt.Sprintf("fleet%d-%d", faction, value), Kind: "fleet", Faction: faction, Value: value})
		}
		for _, pair := range [][2]int{{0, 1000}, {1, 1000}, {31, 1000}, {1000, 1000}, {1999, 1000}, {2000, 1000}, {2500, 1000}, {0, 32768}, {1, 32768}} {
			cases = append(cases, gaugeCase{Name: fmt.Sprintf("planet%d-%d-%d", faction, pair[0], pair[1]), Kind: "planet", Faction: faction, Value: pair[0], Divisor: pair[1]})
		}
	}
	for i := range cases {
		c := &cases[i]
		m.Restore(base)
		m.Write8(fleet+1, byte(c.Faction))
		c.Injections = []map[string]any{{"linear_hex": fmt.Sprintf("0x%x", fleet+1), "bytes": []int{c.Faction}}}
		entry := uint16(0x0d11)
		c.Args = []uint16{fleetOff, fleetSeg}
		if c.Kind == "fleet" {
			m.Write8(fleet+0x3c, byte(c.Value))
			c.Injections = append(c.Injections, map[string]any{"linear_hex": fmt.Sprintf("0x%x", fleet+0x3c), "bytes": []int{c.Value}})
		} else {
			entry = 0x0df7
			m.Write16(planet+0x14, uint16(c.Value))
			c.Injections = append(c.Injections, map[string]any{"linear_hex": fmt.Sprintf("0x%x", planet+0x14), "bytes": []int{c.Value & 255, c.Value >> 8}})
			c.Args = []uint16{planetOff, planetSeg, fleetOff, fleetSeg, uint16(c.Divisor)}
		}
		// 此處為明示direct-call診斷，不更改原版code或原始資料檔。
		sp := m.CPU.R[cpu.SP]
		push := func(v uint16) { sp -= 2; m.Write16(cpu.Addr(m.CPU.Seg[cpu.SS], sp), v) }
		for j := len(c.Args) - 1; j >= 0; j-- {
			push(c.Args[j])
		}
		push(0xfff0)
		push(0)
		m.CPU.R[cpu.SP], m.CPU.Seg[cpu.CS], m.CPU.IP = sp, 0x321f, entry
		callStart := m.Steps
		for m.CPU.Seg[cpu.CS] != 0xfff0 || m.CPU.IP != 0 {
			if m.Steps-callStart >= 1_000_000 || d.Exited || m.CPU.Halted {
				panic("量表常式未返回：" + c.Name)
			}
			cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
			if cs == 0x321f && ip == 0x0d6d {
				c.Height = int(m.CPU.R[cpu.AX])
			} else if cs == 0x321f && ip == 0x0e82 {
				c.Height = int(m.Read16(cpu.Addr(m.CPU.Seg[cpu.SS], m.CPU.R[cpu.BP]+8)))
			}
			if cs == 0x39d7 && (ip == 0x0854 || ip == 0x0647) {
				args := make([]uint16, 6)
				for j := range args {
					args[j] = m.Read16(cpu.Addr(m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]+4+uint16(j*2)))
				}
				c.Rects = append(c.Rects, rectCall{fmt.Sprintf("%04X:%04X", cs, ip), m.Steps, args})
			}
			must(m.Step())
		}
		c.CallSteps = m.Steps - callStart
		write(filepath.Join(out, c.Name+"-screen.bin"), m.Indexed())
		savePNG(m, filepath.Join(out, c.Name+"-screen.png"))
	}
	result := map[string]any{"schema": "logh3.engagement-gauges-native.v1", "classification": "direct native gauge calls; not normal-player or whole-scene parity",
		"dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "go_version": runtime.Version(), "uid": os.Getuid(),
		"parent": parent, "parent_sha256": "af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d",
		"command": "bash tools/engagement_gauges_oracle.sh " + os.Args[1], "canvas": []int{640, 480}, "dosgolem_dirty": false,
		"start_step": start, "base_step": baseStep(base, m), "fleet_linear_hex": fmt.Sprintf("0x%x", fleet),
		"planet_linear_hex": fmt.Sprintf("0x%x", planet), "cases": cases}
	b, err := json.MarshalIndent(result, "", "  ")
	must(err)
	write(filepath.Join(out, "report.json"), append(b, '\n'))
	fmt.Println("native cases", len(cases), "output", out)
}

func baseStep(base *machine.Snapshot, m *machine.Machine) uint64 {
	m.Restore(base)
	return m.Steps
}
