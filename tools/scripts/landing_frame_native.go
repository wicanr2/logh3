// 第109輪：固定原版S3完整狀態，正常實行後只在第六步注入已記錄的局面。
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
func read(p, h string) []byte {
	b, e := os.ReadFile(p)
	must(e)
	if digest(b) != h {
		panic("來源身份不符 " + p)
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
func save(m *machine.Machine, dir, name string) {
	write(filepath.Join(dir, name+"-mem.bin"), m.Snapshot().Mem())
	px, pal := m.Indexed(), m.Palette()
	write(filepath.Join(dir, name+"-screen.bin"), px)
	b, e := json.Marshal(pal)
	must(e)
	write(filepath.Join(dir, name+"-palette.json"), b)
	colors := make(color.Palette, 256)
	for i, p := range pal {
		colors[i] = color.RGBA{p[0], p[1], p[2], 255}
	}
	w, h := m.VideoSize()
	im := image.NewPaletted(image.Rect(0, 0, w, h), colors)
	for i, p := range px {
		im.Pix[i] = m.VGA.DACIndex(p)
	}
	f, e := os.OpenFile(filepath.Join(dir, name+"-screen.png"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(e)
	must(png.Encode(f, im))
	must(f.Close())
}

func main() {
	if len(os.Args) != 3 {
		panic("probe <counterfirst|empire|multiple> <a|b>")
	}
	kind, side := os.Args[1], os.Args[2]
	if side != "a" && side != "b" {
		panic("不明批次")
	}
	if kind != "counterfirst" && kind != "empire" && kind != "multiple" {
		panic("不明案例")
	}
	revision := "r1"
	if kind == "click" {
		revision = "r6"
	}
	dir := filepath.Join("/task/workplace/oracle", "landing48-r109-branch-"+revision+"-"+kind+"-"+side)
	if _, e := os.Stat(dir); !os.IsNotExist(e) {
		panic("拒絕覆寫")
	}
	parent := "/src/workplace/fcmd/eng-t5.state"
	parentHash := "c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1"
	read(parent, parentHash)
	m := machine.New()
	m.EmptyEnv = true
	must(m.LoadCOM(read("/orig/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")))
	d := dos.New(m, "/orig")
	d.Install()
	defer d.Close()
	must(state.Load(parent, m, d))
	if m.IRQ0Every != 20000 || !m.IRQ0Pinned || m.CycleClock {
		panic("父狀態計时不符")
	}
	must(os.Mkdir(dir, 0755))
	start := m.Steps
	const ds = uint32(0x54b30)
	seed := func() uint32 { return uint32(m.Read16(ds+0x94dc)) | uint32(m.Read16(ds+0x94de))<<16 }
	far := func(a uint32) uint32 { return cpu.Addr(m.Read16(a+2), m.Read16(a)) }
	var events, frames, inputs, injections []map[string]any
	defer func() {
		if r := recover(); r != nil {
			save(m, dir, "failed-end")
			b, e := json.MarshalIndent(map[string]any{"error": fmt.Sprint(r), "step": m.Steps, "events": events, "inputs": inputs, "injections": injections}, "", "  ")
			must(e)
			write(filepath.Join(dir, "failure.json"), append(b, '\n'))
			panic(r)
		}
	}()
	set := func(a uint32, b []byte, why string) {
		before := append([]byte(nil), m.Mem[a:a+uint32(len(b))]...)
		m.WriteBytes(a, b)
		injections = append(injections, map[string]any{"address": fmt.Sprintf("lin:%X", a), "step": m.Steps, "before_hex": hex.EncodeToString(before), "bytes_hex": hex.EncodeToString(b), "reason": why})
	}
	save(m, dir, "parent")
	seenPhase, seeded, clicked, moved, released := false, false, false, false, false
	clickInCall := false
	var releaseAt uint64
	count := map[uint32]int{}
	interesting := map[uint32]bool{}
	for _, ip := range []uint16{0x0001, 0x000c, 0x01c1, 0x0222, 0x0276, 0x0286, 0x038f, 0x0486, 0x0503, 0x0568, 0x05a8, 0x05bf, 0x0670, 0x0681, 0x0711, 0x075b, 0x0772, 0x078f, 0x07a2, 0x07b5, 0x07cb, 0x081f, 0x082f, 0x084c, 0x085f, 0x0875, 0x0886, 0x089c, 0x08b1, 0x0a26, 0x0c4e, 0x0e2c, 0x0e45, 0x1004, 0x121f, 0x1269, 0x1370} {
		interesting[cpu.Addr(0x20b4, ip)] = true
	}
	for _, loc := range [][2]uint16{{0x1dd6, 0x0152}, {0x1dd6, 0x026b}, {0x3958, 0x0041}, {0x3958, 0x01cc}, {0x348a, 0x01dc}, {0x348a, 0x0228}, {0x348a, 0x0281}, {0x3b74, 0x031a}, {0x3765, 0x006c}, {0x321f, 0x008a}, {0x321f, 0x0897}, {0x321f, 0x0d11}, {0x321f, 0x0df7}} {
		interesting[cpu.Addr(loc[0], loc[1])] = true
	}
	returns := map[uint32]string{}
	for m.Steps-start < 200000000 {
		cs, ip := m.CPU.Seg[cpu.CS], m.CPU.IP
		a := cpu.Addr(cs, ip)
		if seenPhase && cs == 0x181b && ip == 0x03f0 {
			break
		}
		if d.Exited || m.CPU.Halted {
			panic("原版提前終止")
		}
		if !moved && m.Steps >= 1139800000 {
			d.MoveMouse(240, 200)
			inputs = append(inputs, map[string]any{"kind": "move", "step": m.Steps, "x": 240, "y": 200})
			moved = true
		}
		if !clicked && m.Steps >= 1140000000 {
			d.PressMouse(0)
			inputs = append(inputs, map[string]any{"kind": "press", "step": m.Steps, "button": 0})
			clicked = true
			releaseAt = m.Steps + 3000000
		}
		if !released && releaseAt != 0 && m.Steps >= releaseAt {
			d.ReleaseMouse(0)
			inputs = append(inputs, map[string]any{"kind": "release", "step": m.Steps, "button": 0})
			released = true
		}
		if cs == 0x20b4 && ip == 0x0486 && !seeded {
			seenPhase, seeded = true, true
			if m.CPU.Seg[cpu.DS] != 0x54b3 {
				panic("第六步DS不符")
			}
			set(ds+0x94dc, []byte{0x78, 0x56, 0x34, 0x12}, "第六步行星驅動前固定seed一次；父狀態已過eligibility計算")
			entry := cpu.Addr(m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP])
			if m.Read16(entry+10) != 1 || m.Read16(entry+4) != 0 {
				panic("父狀態行星／攻方名冊不符")
			}
			list := cpu.Addr(m.Read16(entry+8), m.Read16(entry+6))
			fleet := far(list)
			if fleet != ds+0xc074+31*122 {
				panic("原版本次唯一攻方不是艦隊31")
			}
			if m.Read8(fleet+1) != 1 || m.Read8(fleet+4) == 255 {
				panic("艦隊31身份不符")
			}
			planet := far(ds + 0xcff8)
			if m.Read8(planet+0x11) != 0 {
				panic("第一行星不是帝國")
			}
			if kind != "baseline" {
				set(fleet+0x48, []byte{1}, "本次唯一攻方艦隊31戰鬥命令；不改位置與艘數")
				set(fleet+0x70, []byte{m.Read8(fleet+0x70) | 1}, "明示補登陸eligibility位元；診斷0681呈現，不宣稱正常命令判定")
			}
			if kind == "failure" || kind == "off" || kind == "click" {
				set(planet+0x14, []byte{0x70, 0x17, 0x70, 0x17}, "防衛力與上限6000的明示診斷，不重新擲seed")
			}
			if kind == "capture" {
				set(planet+0x14, []byte{1, 0}, "防衛力1的明示成功邊界")
			}

			set(planet+0x14, []byte{0x70,0x17,0x70,0x17}, "三分支統一防衛力及上限6000；保留固定seed，不挑結果")
			if kind == "counterfirst" {
				set(fleet+0x3c, []byte{50}, "0503以艦隊完整度50小於行星防衛比例100，強制走先反擊分支")
			}
			if kind == "empire" {
				set(fleet+1, []byte{0}, "明示將本次攻方陣營改帝國；不是正常玩家局面")
				set(fleet+4, []byte{0}, "明示帝國提督0，供已存在的帝國旗艦COMANI圖段")
				set(planet+0x11, []byte{1}, "明示敵方行星改同盟")
				ship := far(ds+0xcfb4)
				for j:=0;j<int(m.Read8(fleet+0x0f));j++ {
					idx:=m.Read16(fleet+0x10+uint32(j*4))
					set(ship+uint32(idx)*78, []byte{0}, "完整攻方名冊SHIP陣營同步帝國")
				}
			}
			if kind == "multiple" {
				other := ds+0xc074+30*122
				if m.Read8(other+1)!=1 || m.Read8(other+0x0f)==0 || m.Read8(other+4)==255 {panic("第二攻方30身份不符")}
				set(other+0x48, []byte{1}, "第二攻方30命令登陸")
				set(other+0x70, []byte{m.Read8(other+0x70)|1}, "第二攻方30明示補登陸eligibility")
				off:=uint16(other-ds)
				set(list+4, []byte{byte(off),byte(off>>8),0xb3,0x54}, "父名冊末尾加入現存艦隊30遠指標")
				set(entry+10, []byte{2,0}, "0486名冊長度改2；完整堆疊注入記錄")
			}
			if kind == "off" {
				set(ds+0x8f36, []byte{0, 0}, "公開動畫效果關閉的原版欄位")
			}
			save(m, dir, "phase6-input")
		}
		if interesting[a] {
			if len(events) >= 4096 || len(frames) >= 512 {
				panic("觀測筆數超限，拒絕無界累積")
			}
			words := make([]uint16, 16)
			ss, sp := m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP]
			for j := range words {
				words[j] = m.Read16(cpu.Addr(ss, sp+uint16(j*2)))
			}
			e := map[string]any{"original_location": fmt.Sprintf("%04X:%04X", cs, ip), "step": m.Steps, "seed": seed(), "stack_words": words, "ax": m.CPU.R[cpu.AX], "bp": m.CPU.R[cpu.BP], "ss": ss}
			if cs == 0x3958 && ip == 0x0041 {
				ptr := cpu.Addr(words[3], words[2])
				e["comani_record_hex"] = hex.EncodeToString(m.Mem[ptr : ptr+7])
				_, resume := returns[cpu.Addr(words[1], words[0])]
				e["callback_resume"] = resume
			}
			events = append(events, e)
			if cs == 0x3958 && ip == 0x0041 {
				ret := cpu.Addr(words[1], words[0])
				returns[ret] = "comani-return"
			}
			if cs == 0x348a && ip == 0x0281 {
				// 0281由348A段內near call進入，堆疊只有IP，沒有CS。
				ret := cpu.Addr(cs, words[0])
				returns[ret] = "draw-frame-return"
			}
			if (cs == 0x3765 && ip == 0x006c) || (cs == 0x20b4 && (ip == 0x0711 || ip == 0x075b || ip == 0x078f || ip == 0x07cb || ip == 0x084c || ip == 0x085f || ip == 0x089c || ip == 0x05bf)) || (cs == 0x1dd6 && (ip == 0x0152 || ip == 0x026b)) {
				count[a]++
				name := fmt.Sprintf("frame-%04x-%04x-%02d", cs, ip, count[a])
				save(m, dir, name)
				frames = append(frames, map[string]any{"name": name, "step": m.Steps, "original_location": fmt.Sprintf("%04X:%04X", cs, ip), "seed": seed()})
			}
			if kind == "click" && cs == 0x3958 && ip == 0x0041 && words[5] != 0 && !clickInCall {
				clickInCall = true
				d.PressMouse(0)
				inputs = append(inputs, map[string]any{"kind": "press", "step": m.Steps, "button": 0, "reason": "動畫中排隊，不直接結束"})
				releaseAt = m.Steps + 100000
				released = false
			}
		}
		if name, ok := returns[a]; ok {
			delete(returns, a)
			if name == "comani-return" {
				clickInCall = false
			}
			count[a]++
			stem := fmt.Sprintf("%s-%04x-%04x-%02d", name, cs, ip, count[a])
			save(m, dir, stem)
			frames = append(frames, map[string]any{"name": stem, "step": m.Steps, "original_location": fmt.Sprintf("%04X:%04X", cs, ip), "seed": seed()})
		}
		must(m.Step())
	}
	if !seenPhase || m.CPU.Seg[cpu.CS] != 0x181b || m.CPU.IP != 0x03f0 {
		panic(fmt.Sprintf("未從第六步返回 %04X:%04X step%d", m.CPU.Seg[cpu.CS], m.CPU.IP, m.Steps))
	}
	save(m, dir, "end")
	result := map[string]any{"schema": "logh3.landing-scene-native.v1", "case": kind, "command": "bash tools/landing_frame_branch_oracle.sh " + kind + " " + side, "parent": parent, "parent_sha256": parentHash, "classification": "原版完整S3第六步防衛提示保存自然續行，0486前明示注入；未重算命令eligibility，不是正常玩家整回合同態", "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": false, "go_version": runtime.Version(), "uid": os.Getuid(), "start_step": start, "end_step": m.Steps, "end_original_location": fmt.Sprintf("%04X:%04X", m.CPU.Seg[cpu.CS], m.CPU.IP), "irq0_every": m.IRQ0Every, "irq0_pinned": m.IRQ0Pinned, "cycle_clock": m.CycleClock, "canvas": []int{640, 480}, "seed_once": true, "inputs": inputs, "injections": injections, "events": events, "frames": frames}
	b, e := json.MarshalIndent(result, "", "  ")
	must(e)
	write(filepath.Join(dir, "report.json"), append(b, '\n'))
	fmt.Println(kind, "events", len(events), "frames", len(frames), "steps", m.Steps-start)
}
