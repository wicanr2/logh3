//go:build ignore

// #79 子步8-2原版常式診斷。新堆疊／冷VGA，不作正常路徑或畫面證據。
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/dosgolem/oracle"
)

func main() {
	output := flag.String("output", "", "新的本機收據，禁止覆寫")
	flag.Parse()
	if *output == "" {
		panic("須指定新輸出")
	}
	raw, err := os.ReadFile("/task/workplace/oracle/battle79-move-entry-a/mem-in.bin")
	must(err)
	if len(raw) != 1<<20 || digest(raw) != "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d" {
		panic("固定RAM身份不符")
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
	for _, a := range []oracle.Addr{oracle.Far(0x185C, 0x03C6), oracle.Far(0x1213, 0x007B), oracle.Far(0x3826, 0x0405)} {
		a := a
		o.OnCall(a, func(o *oracle.Oracle) {
			r := o.Regs()
			args := []uint16{}
			for i := 0; i < 6; i++ {
				args = append(args, o.Word(oracle.Far(r.SS, r.SP+4+uint16(i*2))))
			}
			row := map[string]any{"entry": a.String(), "args": args, "step": o.Steps()}
			rp := oracle.Far(0x54B3, 0xA79C+uint16(o.Byte(oracle.Far(0x54B3, 0x8F17)))*126)
			row["region_record_hex"] = hex.EncodeToString(o.Bytes(rp, 126))
			if a == oracle.Far(0x3826, 0x0405) {
				b := o.Bytes(oracle.Far(args[1], args[0]), 256)
				for i, v := range b {
					if v == 0 {
						b = b[:i]
						break
					}
				}
				row["text_hex"] = hex.EncodeToString(b)
				row["pending_before_display"] = o.Byte(oracle.Far(0x54B3, 0x988E))
			}
			calls = append(calls, row)
		})
	}
	rows := []map[string]any{}
	patterns := [][]byte{{}, {0}, {1}, {2}, {0, 1}, {0, 2}, {1, 2}, {0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {0, 1, 2}}
	run := func(name string, region, old, humans, occurred int, owners []byte) {
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
		set(oracle.Far(0x54B3, 0x8F17), []byte{byte(region)})
		// 座位0為帝國、座位1為同盟；控制值0人類／1電腦。速度0只取消診斷等待。
		set(oracle.Far(0x54B3, 0x8F25), []byte{0})
		controls := []byte{1, 1}
		for f := 0; f < 2; f++ {
			if humans&(1<<f) != 0 {
				controls[f] = 0
			}
		}
		set(oracle.Far(0x54B3, 0x8F21), controls)
		set(oracle.Far(0x54B3, 0x8F38), []byte{0, 0})
		set(oracle.Far(0x54B3, 0x8F48), []byte{byte(occurred)})
		set(oracle.Far(0x54B3, 0x988E), []byte{0})
		set(oracle.Far(0x54B3, 0xCFFC), []byte{byte(len(owners))})
		rp := oracle.Far(0x54B3, 0xA79C+uint16(region*126))
		set(oracle.Far(rp.Seg, rp.Off+0x12), []byte{byte(old)})
		set(oracle.Far(rp.Seg, rp.Off+0x16), []byte{17, 18, 19})
		pp := oracle.Far(o.Word(oracle.Far(0x54B3, 0xCFFA)), o.Word(oracle.Far(0x54B3, 0xCFF8)))
		for i, own := range owners {
			set(oracle.Far(pp.Seg, pp.Off+uint16(i*82+0x11)), []byte{own})
		}
		before := o.Bytes(rp, 126)
		steps := o.Steps()
		_, err := o.CallBudget(20_000_000, oracle.Far(0x185C, 0x02DA))
		must(err)
		rows = append(rows, map[string]any{"name": name, "region": region, "old_owner": old,
			"human_mask": humans, "occurred": occurred, "planet_owners": owners,
			"injections": injections, "before_record_hex": hex.EncodeToString(before),
			"after_record_hex": hex.EncodeToString(o.Bytes(rp, 126)),
			"pending_after": o.Byte(oracle.Far(0x54B3, 0x988E)),
			"seed_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(0x54B3, 0x94DC), 4)),
			"steps": o.Steps() - steps, "calls": calls})
	}
	for i, owners := range patterns {
		for old := 0; old < 3; old++ {
			for humans := 0; humans < 4; humans++ {
				run(fmt.Sprintf("region26-pattern%d-old%d-humans%d", i, old, humans), 26, old, humans, 0, owners)
			}
		}
	}
	for own := 0; own < 3; own++ {
		for humans := 0; humans < 4; humans++ {
			for occurred := 0; occurred < 2; occurred++ {
				run(fmt.Sprintf("phezzan-own%d-humans%d-occurred%d", own, humans, occurred), 14, 2, humans, occurred, []byte{byte(own)})
			}
		}
	}
	result := map[string]any{"schema": "logh3.battle-ownership-oracle.v1", "program_sha256": digest(exe),
		"memory_sha256": digest(raw), "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c",
		"seed": "0x12345678", "tick": 0,
		"setup": "原版完整RAM；新LoadProgram堆疊，CallNear F000:0000設DS／ES；每案Restore。無替身、不改指令；訊息速度明示0，冷VGA，僅核常式分支與文字，不核畫面、正常入口或時序。",
		"rows": rows}
	b, err := json.MarshalIndent(result, "", "  ")
	must(err)
	f, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(err)
	_, err = f.Write(append(b, '\n'))
	must(err)
	must(f.Close())
	fmt.Println("原版案例", len(rows), "輸出", *output)
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
