//go:build ignore

// #79 子步8-3原版常式診斷。明示注入，不作正常GUI或時間證據。
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/dosgolem/oracle"
)

func main() {
	output := flag.String("output", "", "新的本機收據")
	flag.Parse()
	if *output == "" {
		panic("缺新輸出路徑")
	}
	raw, err := os.ReadFile("/task/workplace/oracle/ownership79-r82-entry-a/mem-in.bin")
	must(err)
	if len(raw) != 1<<20 || digest(raw) != "b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900" {
		panic("父RAM身份不符")
	}
	exe, err := os.ReadFile("/orig/GIN3.COM")
	must(err)
	if digest(exe) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		panic("原版身份不符")
	}
	o, err := oracle.LoadProgram("/orig/GIN3.COM", "/orig", "")
	must(err)
	defer o.Close()
	o.WriteBytes(oracle.Far(0, 0), raw)
	o.SetTickRate(0)
	o.CallNear(oracle.Far(0xF000, 0), 0, oracle.CallRegs{DS: 0x54B3, ES: 0x54B3, SetDS: true, SetES: true})
	snap := o.Save()
	var calls []map[string]any
	var fleet oracle.Addr
	var ship oracle.Addr
	for _, a := range []oracle.Addr{oracle.Far(0x185C, 0x05A9), oracle.Far(0x3826, 0x0405),
		oracle.Far(0x25DB, 0x0535), oracle.Far(0x1727, 0x0298),
		oracle.Far(0x25DB, 0x0714), oracle.Far(0x25DB, 0x0759)} {
		a := a
		o.OnCall(a, func(o *oracle.Oracle) {
			r := o.Regs()
			args := []uint16{}
			for i := 0; i < 6; i++ {
				args = append(args, o.Word(oracle.Far(r.SS, r.SP+4+uint16(i*2))))
			}
			row := map[string]any{"entry": a.String(), "args": args, "step": o.Steps(),
				"fleet_hex": hex.EncodeToString(o.Bytes(fleet, 122)),
				"ship_hex":  hex.EncodeToString(o.Bytes(ship, 78))}
			if a == oracle.Far(0x3826, 0x0405) {
				b := o.Bytes(oracle.Far(args[1], args[0]), 256)
				for i, v := range b {
					if v == 0 {
						b = b[:i]
						break
					}
				}
				row["text_hex"] = hex.EncodeToString(b)
			}
			calls = append(calls, row)
		})
	}
	rows := []map[string]any{}
	run := func(index, mode, fuel, humans, count int) {
		o.Restore(snap)
		calls = nil
		injections := []map[string]any{}
		set := func(a oracle.Addr, b []byte) {
			o.WriteBytes(a, b)
			if hex.EncodeToString(o.Bytes(a, len(b))) != hex.EncodeToString(b) {
				panic("注入回讀不符")
			}
			injections = append(injections, map[string]any{"linear": fmt.Sprintf("lin:%X", a.Linear()), "hex": hex.EncodeToString(b)})
		}
		set(oracle.Far(0x54B3, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
		set(oracle.Far(0x54B3, 0x9894), []byte{byte(count)})
		set(oracle.Far(0x54B3, 0x9898), []byte{byte(index)})
		set(oracle.Far(0x54B3, 0x8F25), []byte{0})
		controls := []byte{1, 1}
		for f := 0; f < 2; f++ {
			if humans&(1<<f) != 0 {
				controls[f] = 0
			}
		}
		set(oracle.Far(0x54B3, 0x8F21), controls)
		set(oracle.Far(0x54B3, 0x8F38), []byte{0, 0})
		fleet = oracle.Far(0x54B3, 0xC074+uint16(index*122))
		pool := o.Word(oracle.Far(fleet.Seg, fleet.Off+0x38))
		ship = oracle.Far(o.Word(oracle.Far(0x54B3, 0xCFB6)), o.Word(oracle.Far(0x54B3, 0xCFB4))+pool*78)
		set(oracle.Far(fleet.Seg, fleet.Off+0x3A), []byte{byte(mode), byte(fuel)})
		// 明示放出位置與標記，避免回收恰與原位置相同而漏驗寫入。
		set(oracle.Far(ship.Seg, ship.Off+0x19), []byte{0xC5})
		set(oracle.Far(ship.Seg, ship.Off+0x1E), []byte{2})
		xy := make([]byte, 4)
		binary.LittleEndian.PutUint16(xy, 560)
		binary.LittleEndian.PutUint16(xy[2:], 360)
		set(oracle.Far(ship.Seg, ship.Off+0x28), xy)
		binary.LittleEndian.PutUint16(xy, 600)
		binary.LittleEndian.PutUint16(xy[2:], 400)
		set(oracle.Far(ship.Seg, ship.Off+0x34), xy)
		beforeF, beforeS := o.Bytes(fleet, 122), o.Bytes(ship, 78)
		steps := o.Steps()
		_, err := o.CallBudget(20_000_000, oracle.Far(0x185C, 0x0491))
		must(err)
		rows = append(rows, map[string]any{"name": fmt.Sprintf("fleet%d-mode%d-fuel%d-human%d-count%d", index, mode, fuel, humans, count),
			"fleet": index, "pool": pool, "mode": mode, "fuel": fuel, "human_mask": humans, "fleet_count": count,
			"injections": injections, "before_fleet_hex": hex.EncodeToString(beforeF), "before_ship_hex": hex.EncodeToString(beforeS),
			"after_fleet_hex": hex.EncodeToString(o.Bytes(fleet, 122)), "after_ship_hex": hex.EncodeToString(o.Bytes(ship, 78)),
			"seed_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(0x54B3, 0x94DC), 4)),
			"steps":          o.Steps() - steps, "calls": calls})
	}
	for _, f := range []int{12, 16} {
		for _, mode := range []int{0, 2, 0x7F, 0xFE, 0xFF} {
			for _, fuel := range []int{0, 1, 2, 3, 4, 254, 255} {
				for mask := 0; mask < 4; mask++ {
					run(f, mode, fuel, mask, 1)
				}
			}
		}
	}
	run(12, 0, 3, 3, 0)
	result := map[string]any{"schema": "logh3.scout-fuel-oracle.v1", "program_sha256": digest(exe),
		"memory_sha256": digest(raw), "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "seed": "0x12345678", "tick": 0,
		"setup": "原版RAM；新LoadProgram堆疊，CallNear F000:0000設DS／ES；每案Restore。無替身、不改指令；訊息速度0、冷VGA，僅核常式分支／資料／文字，不核正常入口或時間。", "rows": rows}
	b, err := json.MarshalIndent(result, "", "  ")
	must(err)
	f, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(err)
	_, err = f.Write(append(b, '\n'))
	must(err)
	must(f.Close())
	fmt.Println("原版索敵燃料案例", len(rows), "輸出", *output)
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
