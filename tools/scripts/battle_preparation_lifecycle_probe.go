//go:build ignore

// #79 來源寫入與活指標診斷；入口見 docs/re/369。
// 局部入口使用明示假 far frame，停在寫入後，絕不執行假返回。
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
	output := flag.String("output", "", "全新收據；禁止覆寫")
	flag.Parse()
	if *output == "" {
		panic("必須指定收據")
	}
	raw, err := os.ReadFile("/out/oracle/battle79-preparation-b/mem-in.bin")
	must(err)
	if len(raw) != 1<<20 || digest(raw) != "f70507aaa771a0b97d425037d06696c90385238c6e294a39a103a836960a6808" {
		panic("入口身份不符")
	}
	exe, err := os.ReadFile("/orig/GIN3.COM")
	must(err)
	if digest(exe) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		panic("原版身份不符")
	}
	o, err := oracle.LoadProgram("/orig/GIN3.COM", "/orig", "")
	must(err)
	defer o.Close()
	if o.SetBytes(oracle.Far(0, 0), raw) != len(raw) {
		panic("入口注入不完整")
	}
	o.SetTickRate(0)
	o.CallNear(oracle.Far(0xF000, 0), 0, oracle.CallRegs{DS: 0x54B3, ES: 0x54B3, SetDS: true, SetES: true})
	snap := o.Save()
	ds, pool, planet := uint16(0x54B3), oracle.Far(0x668B, 2), oracle.Far(0x5313, 0x1430)
	rows := []map[string]any{}
	for _, test := range []struct {
		name                     string
		seg, entry, stop, source uint16
		kind                     byte
	}{
		{"manual-supply-write", 0x1930, 0x05EC, 0x0623, 0xA42E, 5},
		{"manual-repair-write", 0x1AAB, 0x0583, 0x05BC, 0xA440, 6},
		{"cpu-repair-write", 0x1AAB, 0x0190, 0x01E7, 0xA440, 6},
	} {
		o.Restore(snap)
		injections := []map[string]any{}
		set := recorder(o, &injections)
		set(oracle.Far(ds, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
		at := oracle.Far(pool.Seg, pool.Off+3*78)
		b := unit(17, 0, int(test.kind))
		set(at, b)
		set(oracle.Far(ds, test.source), make([]byte, 4))
		set(oracle.Far(ds, 0x9896), []byte{1, 0})
		set(oracle.Far(ds, 0x98BA), []byte{3, 0})
		r := o.Regs()
		frame := append([]byte{0, 0xF0}, pointer(at)...)
		if test.name == "cpu-repair-write" {
			frame = []byte{0, 0xF0, 1, 0}
		}
		// 堆疊接近FFFF，參數字必須各自按16位位移回繞；不能連續線性寫跨段。
		for i := 0; i < len(frame); i += 2 {
			set(oracle.Far(r.SS, r.SP+uint16(i)), frame[i:i+2])
		}
		before := o.Steps()
		o.CallNear(oracle.Far(test.seg, test.entry), 0, oracle.CallRegs{})
		must(o.RunUntil(oracle.At(oracle.Far(test.seg, test.stop)), oracle.Budget(10000)))
		if hex.EncodeToString(o.Bytes(oracle.Far(ds, test.source), 4)) != hex.EncodeToString(pointer(at)) {
			panic("來源寫入與指定參數不符；不能接受錯誤堆疊收據")
		}
		rows = append(rows, map[string]any{"name": test.name, "entry": fmt.Sprintf("%04X:%04X", test.seg, test.entry), "stop": fmt.Sprintf("%04X:%04X", test.seg, test.stop), "steps": o.Steps() - before, "source_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds, test.source), 4)), "injections": injections, "rng_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds, 0x94DC), 4)), "scope": "假far frame局部診斷；只到來源寫入後，不執行返回，不證明手動UI或正常玩家路徑。"})
	}
	for _, name := range []string{"retained-and-reused-pool", "null-source-physical-one-zero", "null-source-physical-one-sixteen"} {
		o.Restore(snap)
		injections := []map[string]any{}
		set := recorder(o, &injections)
		set(oracle.Far(ds, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
		set(oracle.Far(ds, 0x9894), []byte{0, 0}) // 明示跳過士氣，僅驗資源排序。
		set(oracle.Far(ds, 0x9896), []byte{3, 0})
		set(oracle.Far(ds, 0x98BA), []byte{0, 0, 1, 0, 2, 0})
		set(oracle.Far(ds, 0xCFFC), []byte{1})
		oldSource := oracle.Far(pool.Seg, pool.Off+3*78)
		ptr := pointer(oldSource)
		firstFleet := 17
		if name != "retained-and-reused-pool" {
			ptr = make([]byte, 4)
			firstFleet = 0
			value := byte(0)
			if name == "null-source-physical-one-sixteen" {
				value = 16
			}
			set(oracle.Far(0, 1), []byte{value}) // IRQ關閉；不得外推正常DOS硬體。
		}
		set(oracle.Far(ds, 0xA440), ptr)
		set(oldSource, unit(17, 0, 6))
		phases := []map[string]any{}
		count := 1
		if name == "retained-and-reused-pool" {
			count = 2
		}
		for phase := 0; phase < count; phase++ {
			if phase == 1 {
				set(oracle.Far(oldSource.Seg, oldSource.Off+1), []byte{16})
			}
			units := [][]byte{unit(firstFleet, 0, 0), unit(16, 0, 0), unit(16, 255, 7)}
			for i := 0; i < 2; i++ {
				put16(units[i], 8, 50)
				put16(units[i], 12, 75)
			}
			for i, b := range units {
				set(oracle.Far(pool.Seg, pool.Off+uint16(i*78)), b)
			}
			p := make([]byte, 82)
			p[17] = 1
			put16(p, 20, 3300)
			put16(p, 22, 3300)
			put16(p, 24, 5)
			put16(p, 70, 1000)
			put16(p, 72, 1000)
			put32(p, 32, 20)
			set(planet, p)
			before := o.Steps()
			_, err = o.Call(oracle.Far(0x1AAB, 0x006D))
			must(err)
			ships := []int{}
			for i := 0; i < 2; i++ {
				ships = append(ships, int(binary.LittleEndian.Uint16(o.Bytes(oracle.Far(pool.Seg, pool.Off+uint16(i*78)+8), 2))))
			}
			phases = append(phases, map[string]any{"phase": phase, "entry": "1AAB:006D", "steps": o.Steps() - before, "target_ships": ships, "source_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds, 0xA440), 4)), "source_fleet_byte": o.Bytes(oracle.Far(oldSource.Seg, oldSource.Off+1), 1)[0], "physical_one": o.Bytes(oracle.Far(0, 1), 1)[0], "rng_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds, 0x94DC), 4))})
		}
		rows = append(rows, map[string]any{"name": name, "phases": phases, "injections": injections, "scope": "常式診斷；本次無種類6，保留來源；池槽重用為明示注入。"})
	}
	result := map[string]any{"schema": "logh3.battle-preparation-lifecycle-oracle.v1", "program_sha256": digest(exe), "memory_sha256": digest(raw), "seed": "0x12345678", "tick": 0, "setup": "完整1MiB入口注入；新LoadProgram堆疊，CallNear F000:0000設DS／ES；每案Restore；所有額外注入順序與bytes逐筆記錄。", "rows": rows}
	data, err := json.MarshalIndent(result, "", "  ")
	must(err)
	data = append(data, '\n')
	f, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(err)
	_, err = f.Write(data)
	must(err)
	must(f.Close())
	fmt.Println("收據", *output, "SHA-256", digest(data), "案例", len(rows))
}

func unit(fleet, squad, kind int) []byte {
	b := make([]byte, 78)
	b[0], b[1], b[2], b[3], b[4], b[30] = 1, byte(fleet), 31, byte(squad), byte(kind), 127
	put16(b, 8, 100)
	put16(b, 10, 100)
	put16(b, 12, 100)
	put16(b, 0x34, 1000)
	put16(b, 0x36, 1000)
	return b
}
func pointer(a oracle.Addr) []byte {
	b := make([]byte, 4)
	put16(b, 0, int(a.Off))
	put16(b, 2, int(a.Seg))
	return b
}
func recorder(o *oracle.Oracle, rows *[]map[string]any) func(oracle.Addr, []byte) {
	return func(a oracle.Addr, b []byte) {
		if o.SetBytes(a, b) != len(b) {
			panic("注入未完整")
		}
		*rows = append(*rows, map[string]any{"linear": a.Linear(), "hex": hex.EncodeToString(b)})
	}
}
func put16(b []byte, off, v int) { binary.LittleEndian.PutUint16(b[off:], uint16(v)) }
func put32(b []byte, off, v int) { binary.LittleEndian.PutUint32(b[off:], uint32(v)) }
func digest(b []byte) string     { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
