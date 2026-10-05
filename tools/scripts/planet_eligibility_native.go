// #91原版000C／0091條件矩陣；明示CPU與堆疊注入，不是正常玩家入口。
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(err)
	_, err = f.Write(b)
	must(err)
	must(f.Close())
}
func jsonWrite(path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	must(err)
	write(path, append(b, '\n'))
}

type scenario struct {
	Name                           string
	Baseline                       bool
	Defence, DX, DY                int
	Action, Flag, Owner, Kind, Gun byte
	Dead                           bool
	Second                         string
}

func scenarios(mode string) []scenario {
	base := scenario{Name: "baseline", Baseline: true}
	if mode == "minimum" || mode == "pipeline" {
		return []scenario{base}
	}
	list := []scenario{base}
	add := func(name string, d, dx, dy int, action, flag, owner, kind, gun byte, dead bool, second string) {
		list = append(list, scenario{name, false, d, dx, dy, action, flag, owner, kind, gun, dead, second})
	}
	for _, action := range []byte{0, 1, 4, 0x10, 0x11} {
		add(fmt.Sprintf("command-%02x", action), 6000, 100, 0, action, 0, 0, 0, 0, false, "")
	}
	for _, action := range []byte{0, 1} {
		add(fmt.Sprintf("flag0-command-%d", action), 6000, 100, 0, action, 1, 0, 0, 0, false, "")
	}
	for _, dx := range []int{149, 150, 151, 239, 240, 241} {
		add(fmt.Sprintf("range6000-%d", dx), 6000, dx, 0, 1, 0, 0, 0, 0, false, "")
	}
	for _, dx := range []int{59, 60, 61, 149, 150, 151} {
		add(fmt.Sprintf("range0-%d", dx), 0, dx, 0, 1, 0, 0, 0, 0, false, "")
	}
	for _, dx := range []int{60, 61} {
		add(fmt.Sprintf("range1499-%d", dx), 1499, dx, 0, 0, 0, 0, 0, 0, false, "")
	}
	for _, dx := range []int{240, 241} {
		add(fmt.Sprintf("range65535-%d", dx), 65535, dx, 0, 0, 0, 0, 0, 0, false, "")
	}
	add("diagonal150", 6000, 90, 120, 1, 0, 0, 0, 0, false, "")
	add("diagonal-over150", 6000, 91, 120, 1, 0, 0, 0, 0, false, "")
	for _, gun := range []byte{0, 2, 3} {
		add(fmt.Sprintf("fortress-gun-%d", gun), 6000, 100, 0, 1, 0, 0, 1, gun, false, "")
	}
	add("fortress2-aimed", 6000, 100, 0, 1, 0, 0, 2, 3, false, "")
	add("ordinary-gun3", 6000, 100, 0, 1, 0, 0, 0, 3, false, "")
	add("friendly", 6000, 100, 0, 1, 0, 1, 0, 0, false, "")
	add("neutral", 6000, 100, 0, 1, 0, 2, 0, 0, false, "")
	add("dead-flagship", 6000, 100, 0, 1, 0, 0, 0, 0, true, "")
	add("flag3-only", 6000, 100, 0, 1, 8, 0, 0, 0, false, "")
	add("reset-no-match", 6000, 500, 0, 0, 0, 0, 0, 0, false, "")
	for _, second := range []string{"last-landing", "friendly", "counter-only", "two-counter"} {
		action := byte(1)
		if second == "counter-only" || second == "two-counter" {
			action = 0
		}
		add("two-planets-"+second, 6000, 100, 0, action, 0, 0, 0, 0, false, second)
	}
	return list
}

func cpuRecord(m *machine.Machine) map[string]any {
	return map[string]any{"R": m.CPU.R, "Seg": m.CPU.Seg, "IP": m.CPU.IP, "Flags": m.CPU.Flags}
}
func targets(m *machine.Machine, ds uint32) []map[string]any {
	var out []map[string]any
	for i := 0; i < int(m.Read8(ds+0x9894)); i++ {
		id := int(m.Read8(ds + 0x9898 + uint32(i)))
		f := ds + 0xc074 + uint32(id*122)
		out = append(out, map[string]any{"fleet": id, "target": m.Read8(f + 0x5a), "bits": m.Read16(f + 0x70)})
	}
	return out
}

func runCase(c scenario, dir, mode string) {
	const ds = uint32(0x54b30)
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
		panic("原版父時序不符")
	}
	if m.Read8(cpu.Addr(0x20b4, 0x000c)) != 0x55 || m.Read8(cpu.Addr(0x20b4, 0x0091)) != 0x55 {
		panic("父狀態未載入20B4原碼")
	}
	far := func(a uint32) uint32 { return cpu.Addr(m.Read16(a+2), m.Read16(a)) }
	var injections, events []map[string]any
	set := func(a uint32, b []byte, why string) {
		before := append([]byte(nil), m.Mem[a:a+uint32(len(b))]...)
		m.WriteBytes(a, b)
		injections = append(injections, map[string]any{"address": fmt.Sprintf("lin:%X", a), "step": m.Steps, "before_hex": hex.EncodeToString(before), "bytes_hex": hex.EncodeToString(b), "reason": why})
	}
	setWord := func(a uint32, v uint16, why string) {
		var b [2]byte
		binary.LittleEndian.PutUint16(b[:], v)
		set(a, b[:], why)
	}
	write(filepath.Join(dir, c.Name+"-parent-mem.bin"), m.Snapshot().Mem())
	defer func() {
		if r := recover(); r != nil {
			write(filepath.Join(dir, c.Name+"-failed-mem.bin"), m.Snapshot().Mem())
			jsonWrite(filepath.Join(dir, c.Name+"-failure.json"), map[string]any{"error": fmt.Sprint(r), "step": m.Steps, "cpu": cpuRecord(m), "injections": injections, "events": events})
			panic(r)
		}
	}()
	set(ds+0x94dc, []byte{0x78, 0x56, 0x34, 0x12}, "函式進場前固定seed一次；Targets應不消耗亂數")
	if !c.Baseline {
		set(ds+0x9894, []byte{1}, "固定在場名冊一支現有艦隊31")
		set(ds+0x9898, []byte{31}, "名冊首筆艦隊31")
		n := byte(1)
		if c.Second != "" {
			n = 2
		}
		set(ds+0xcffc, []byte{n}, "固定行星切片長度")
		f := ds + 0xc074 + 31*122
		p := far(ds + 0xcff8)
		set(f, []byte{c.Flag}, "艦隊旗標原byte")
		set(f+0x48, []byte{c.Action}, "命令原byte，原版只讀低四位")
		set(p+0x11, []byte{c.Owner}, "行星擁有者")
		set(p+0x12, []byte{c.Kind}, "行星種類")
		set(p+0x45, []byte{c.Gun}, "要塞砲狀態")
		setWord(p+0x14, uint16(c.Defence), "固定查詢半徑來源防衛力")
		setWord(f+0x5c, m.Read16(p+0x46)+uint16(c.DX), "艦隊參考X與行星的固定距離")
		setWord(f+0x5e, m.Read16(p+0x48)+uint16(c.DY), "艦隊參考Y與行星的固定距離")
		if c.Dead {
			s := far(ds+0xcfb4) + uint32(m.Read16(f+0x10))*78
			setWord(s+8, 0, "旗艦零艘拒絕邊界")
		}
		setWord(f+0x70, 0xffff, "確認000C清除舊bits，不借舊eligibility")
		set(f+0x5a, []byte{17}, "確認000C清除舊行星目標")
		if c.Second != "" {
			p2 := p + 82
			setWord(p2+0x46, m.Read16(p+0x46), "第二行星與第一行星同座標，固定覆寫順序")
			setWord(p2+0x48, m.Read16(p+0x48), "第二行星同Y")
			setWord(p2+0x14, 6000, "第二行星查詢半徑240")
			set(p2+0x11, []byte{0}, "第二行星敵方")
			set(p2+0x12, []byte{0}, "第二行星一般種類")
			set(p2+0x45, []byte{0}, "第二行星砲待命")
			switch c.Second {
			case "last-landing":
				set(p2+0x12, []byte{1}, "第二行星要塞")
				set(p2+0x45, []byte{3}, "第二行星只允許bit0，確認bits累積與last-target")
			case "friendly":
				set(p2+0x11, []byte{1}, "第二行星友軍，不能覆寫目標")
			case "counter-only":
				set(p+0x11, []byte{1}, "第一行星友軍，只有第二行星反擊")
			}
		}
	}
	beforeCPU := cpuRecord(m)
	beforeR, beforeSeg, beforeIP, beforeFlags := m.CPU.R, m.CPU.Seg, m.CPU.IP, m.CPU.Flags
	sp := m.CPU.R[cpu.SP] - 4
	stack := cpu.Addr(m.CPU.Seg[cpu.SS], sp)
	set(stack, []byte{0xfe, 0xff, 0xb4, 0x20}, "000C無參數far返回20B4:FFFE；在返回前停，從不執行該位址")
	m.CPU.R[cpu.SP] = sp
	m.CPU.Seg[cpu.CS] = 0x20b4
	m.CPU.IP = 0x000c
	m.CPU.Seg[cpu.DS] = 0x54b3
	entryCPU := cpuRecord(m)
	start := m.Steps
	write(filepath.Join(dir, c.Name+"-input-mem.bin"), m.Snapshot().Mem())
	interesting := map[uint16]bool{}
	for _, ip := range []uint16{0x000c, 0x0044, 0x004c, 0x0091, 0x0103, 0x011c, 0x0153, 0x015c, 0x0169, 0x016e, 0x0173, 0x018f, 0x0199, 0x01a0, 0x01a5, 0x01ad} {
		interesting[ip] = true
	}
	for m.Steps-start < 1000000 {
		cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
		if cs == 0x20b4 && ip == 0xfffe {
			break
		}
		if d.Exited || m.CPU.Halted {
			panic("原版函式提前終止")
		}
		if cs == 0x20b4 && interesting[ip] {
			if len(events) >= 2048 {
				panic("有界事件超限")
			}
			events = append(events, map[string]any{"original_location": fmt.Sprintf("%04X:%04X", cs, ip), "step": m.Steps, "ax": m.CPU.R[cpu.AX], "bx": m.CPU.R[cpu.BX], "cx": m.CPU.R[cpu.CX], "es": m.CPU.Seg[cpu.ES]})
		}
		must(m.Step())
	}
	if m.CPU.Seg[cpu.CS] != 0x20b4 || m.CPU.IP != 0xfffe {
		panic("000C未在步數上限內返回")
	}
	if m.Read16(ds+0x94dc) != 0x5678 || m.Read16(ds+0x94de) != 0x1234 {
		panic("Targets消耗了亂數")
	}
	write(filepath.Join(dir, c.Name+"-end-mem.bin"), m.Snapshot().Mem())
	eligibilityCPU, eligibilityEnd := cpuRecord(m), m.Steps
	if mode == "pipeline" {
		// 保留新算的Targets，回到完整父狀態的提示等待位置。
		// 原版000C用到的SP以下是呼叫暫存區；不替換正在使用的父堆疊。
		m.CPU.R, m.CPU.Seg, m.CPU.IP, m.CPU.Flags = beforeR, beforeSeg, beforeIP, beforeFlags
		jsonWrite(filepath.Join(dir, "cpu-restore.json"), map[string]any{"step": m.Steps, "cpu": cpuRecord(m), "reason": "新算Targets後恢復父CPU；非正常玩家入口"})
		write(filepath.Join(dir, "continuation-input-mem.bin"), m.Snapshot().Mem())
		var inputs, calls []map[string]any
		moved, pressed, released, phase := false, false, false, false
		seed := func() uint32 { return uint32(m.Read16(ds+0x94dc)) | uint32(m.Read16(ds+0x94de))<<16 }
		for m.Steps-start < 200000000 {
			cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
			if phase && cs == 0x181b && ip == 0x03f0 {
				break
			}
			if d.Exited || m.CPU.Halted {
				panic("原版續行提前終止")
			}
			if !moved && m.Steps >= 1139800000 {
				d.MoveMouse(240, 200)
				moved = true
				inputs = append(inputs, map[string]any{"kind": "move", "step": m.Steps, "x": 240, "y": 200})
			}
			if !pressed && m.Steps >= 1140000000 {
				d.PressMouse(0)
				pressed = true
				inputs = append(inputs, map[string]any{"kind": "press", "step": m.Steps, "button": 0})
			}
			if !released && m.Steps >= 1143000000 {
				d.ReleaseMouse(0)
				released = true
				inputs = append(inputs, map[string]any{"kind": "release", "step": m.Steps, "button": 0})
			}
			if cs == 0x20b4 {
				if ip == 0x0486 && !phase {
					phase = true
					write(filepath.Join(dir, "phase6-input-mem.bin"), m.Snapshot().Mem())
				}
				if ip == 0x0486 || ip == 0x0503 || ip == 0x0681 || ip == 0x08b1 || ip == 0x121f {
					if len(calls) >= 64 {
						panic("原版續行事件超限")
					}
					calls = append(calls, map[string]any{"original_location": fmt.Sprintf("%04X:%04X", cs, ip), "step": m.Steps, "seed": seed()})
				}
			}
			must(m.Step())
		}
		if !phase || m.CPU.Seg[cpu.CS] != 0x181b || m.CPU.IP != 0x03f0 {
			panic("原版續行未返回第六步末")
		}
		write(filepath.Join(dir, "phase6-end-mem.bin"), m.Snapshot().Mem())
		jsonWrite(filepath.Join(dir, "continuation-report.json"), map[string]any{"classification": "fresh 000C direct-entry後恢復父CPU續行第六步；不是正常玩家整回合同態", "inputs": inputs, "calls": calls, "seed_once": true, "final_seed": seed(), "end_step": m.Steps, "end_original_location": "181B:03F0"})
	}
	jsonWrite(filepath.Join(dir, c.Name+"-report.json"), map[string]any{
		"schema": "logh3.planet-eligibility-native.v1", "scenario": c, "parent": parent, "parent_sha256": parentHash,
		"original_program_sha256": "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918",
		"classification":          "明示CPU／堆疊direct-entry重跑000C及0091；非正常玩家整回合同態",
		"cpu_before":              beforeCPU, "cpu_injected": entryCPU, "cpu_return": eligibilityCPU,
		"injections": injections, "events": events, "targets": targets(m, ds), "seed_before": uint32(0x12345678), "seed_after": uint32(0x12345678),
		"start_step": start, "end_step": eligibilityEnd, "irq0_every": m.IRQ0Every, "irq0_pinned": m.IRQ0Pinned, "cycle_clock": m.CycleClock,
		"dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false, "go_version": runtime.Version(), "uid": os.Getuid()})
	fmt.Println(c.Name, "steps", m.Steps-start, "targets", targets(m, ds))
}

func main() {
	if len(os.Args) != 3 || (os.Args[1] != "a" && os.Args[1] != "b") || (os.Args[2] != "minimum" && os.Args[2] != "matrix" && os.Args[2] != "matrix2" && os.Args[2] != "pipeline") {
		panic("native <a|b> <minimum|matrix|pipeline>")
	}
	dir := filepath.Join("/task/workplace/oracle", "planet91-r111-"+os.Args[2]+"-"+os.Args[1])
	must(os.Mkdir(dir, 0755))
	for _, c := range scenarios(os.Args[2]) {
		runCase(c, dir, os.Args[2])
	}
	jsonWrite(filepath.Join(dir, "batch.json"), map[string]any{"command": "bash tools/planet_eligibility_oracle.sh " + os.Args[1] + " " + os.Args[2], "cases": len(scenarios(os.Args[2])), "run": os.Args[1]})
}
