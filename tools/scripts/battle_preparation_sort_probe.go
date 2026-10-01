//go:build ignore

// #79 排序及行星沿用來源指標的原版診斷；入口見 docs/re/368。
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
	output := flag.String("output", "", "新的本機收據；禁止覆寫")
	flag.Parse()
	if *output == "" {
		panic("必須指定新收據")
	}
	raw, err := os.ReadFile("/out/oracle/battle79-preparation-b/mem-in.bin")
	must(err)
	if len(raw) != 1<<20 || digest(raw) != "f70507aaa771a0b97d425037d06696c90385238c6e294a39a103a836960a6808" {
		panic("入口記憶體身份不符")
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
	ds := oracle.Far(0x54B3, 0)
	word := func(off int) uint16 { return binary.LittleEndian.Uint16(raw[ds.Linear()+uint32(off):]) }
	pool := oracle.Far(word(0xCFB6), word(0xCFB4))
	planet := oracle.Far(word(0xCFFA), word(0xCFF8))
	snap := o.Save()
	rows := []map[string]any{}
	// 包含9／10轉11的間距規則，並以戰場55筆量級確認平手交換。
	for _, n := range []int{0, 1, 2, 3, 7, 8, 12, 13, 14, 15, 16, 55} {
		for _, comparator := range []struct {
			seg, off uint16
			name     string
			donor    int
		}{
			{0x1930, 0x0510, "supply", 0xA42E}, {0x1AAB, 0x04BD, "repair", 0xA440},
		} {
			for _, variant := range []string{"mixed", "reverse", "all-ties"} {
				o.Restore(snap)
				injections := []map[string]any{}
				set := recorder(o, &injections)
				set(oracle.Far(ds.Seg, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
				donor := unit(16, 0, 6)
				donorAt := oracle.Far(pool.Seg, pool.Off+uint16(n*78))
				set(donorAt, donor)
				set(oracle.Far(ds.Seg, uint16(comparator.donor)), pointer(donorAt))
				before, identities := make([]int, n), []map[string]int{}
				for i := 0; i < n; i++ {
					f, squad := 17+i%3, (n-i)%9
					if i%4 == 0 {
						f = 16
					}
					if variant == "all-ties" {
						f, squad = 17+i%3, 0
					}
					set(oracle.Far(pool.Seg, pool.Off+uint16(i*78)), unit(f, squad, 0))
					before[i] = i
					if variant == "reverse" {
						before[i] = n - i - 1
					}
					identities = append(identities, map[string]int{"index": i, "fleet": f, "squadron": squad})
				}
				array := make([]byte, 2*n)
				for i, v := range before {
					put16(array, i*2, v)
				}
				set(oracle.Far(ds.Seg, 0xE000), array)
				_, err = o.Call(oracle.Far(0x3C6A, 0x000D), 0xE000, ds.Seg, uint16(n), 2, comparator.off, comparator.seg)
				must(err)
				after := make([]int, n)
				b := o.Bytes(oracle.Far(ds.Seg, 0xE000), 2*n)
				for i := range after {
					after[i] = int(binary.LittleEndian.Uint16(b[2*i:]))
				}
				rows = append(rows, map[string]any{"name": fmt.Sprintf("%s-%s-%d", comparator.name, variant, n), "entry": "3C6A:000D", "comparator": fmt.Sprintf("%04X:%04X", comparator.seg, comparator.off), "units": identities, "donor_fleet": 16, "before": before, "after": after, "injections": injections, "rng_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0x94DC), 4))})
			}
		}
	}
	// 行星排序：運補階段總是用最後一筆艦艇；修理階段只在遇到種類6時更新。
	for _, name := range []string{"planet-supply-last-unit", "planet-repair-retained-source", "planet-repair-last-donor"} {
		o.Restore(snap)
		injections := []map[string]any{}
		set := recorder(o, &injections)
		set(oracle.Far(ds.Seg, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
		set(oracle.Far(ds.Seg, 0x9894), []byte{0, 0}) // 跳過士氣，限制在資源診斷。
		set(oracle.Far(ds.Seg, 0x9896), []byte{3, 0})
		set(oracle.Far(ds.Seg, 0x98BA), []byte{0, 0, 1, 0, 2, 0})
		set(oracle.Far(ds.Seg, 0xCFFC), []byte{1})
		units := [][]byte{unit(17, 0, 0), unit(16, 0, 0), unit(16, 255, 7), unit(17, 0, 6)}
		for i := 0; i < 2; i++ {
			put16(units[i], 8, 50)
			put16(units[i], 12, 75)
		}
		if name == "planet-repair-last-donor" {
			units[2] = unit(16, 8, 6)
		}
		for i, b := range units {
			set(oracle.Far(pool.Seg, pool.Off+uint16(i*78)), b)
		}
		set(oracle.Far(ds.Seg, 0xA42E), pointer(oracle.Far(pool.Seg, pool.Off+3*78)))
		set(oracle.Far(ds.Seg, 0xA440), pointer(oracle.Far(pool.Seg, pool.Off+3*78)))
		p := make([]byte, 82)
		p[17] = 1
		put16(p, 20, 3300)
		put16(p, 22, 3300)
		put16(p, 24, 5)
		put16(p, 70, 1000)
		put16(p, 72, 1000)
		entry := oracle.Far(0x1AAB, 0x006D)
		if name == "planet-supply-last-unit" {
			put32(p, 60, 100)
			entry = oracle.Far(0x1930, 0x019D)
		} else {
			put32(p, 32, 20)
		}
		set(planet, p)
		_, err = o.Call(entry)
		must(err)
		after := []string{}
		before := []string{}
		for i, b := range units {
			before = append(before, hex.EncodeToString(b))
			after = append(after, hex.EncodeToString(o.Bytes(oracle.Far(pool.Seg, pool.Off+uint16(i*78)), 78)))
		}
		rows = append(rows, map[string]any{"name": name, "entry": fmt.Sprintf("%04X:%04X", entry.Seg, entry.Off), "units_before_hex": before, "units_after_hex": after, "planet_before_hex": hex.EncodeToString(p), "planet_after_hex": hex.EncodeToString(o.Bytes(planet, 82)), "source_pointers_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0xA42E), 4)) + "/" + hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0xA440), 4)), "injections": injections, "rng_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0x94DC), 4))})
	}
	result := map[string]any{"schema": "logh3.battle-preparation-sort-oracle.v1", "input_sha256": digest(raw), "program_sha256": digest(exe), "seed": "0x12345678", "seed_linear": "lin:5E00C", "tick": 0, "setup": "完整1MiB注入；新LoadProgram堆疊；CallNear F000:0000設DS=ES=54B3；每案Restore；完整注入逐筆記錄。", "scope": "排序／資源常式診斷；士氣艦隊數清零；不證明正常玩家路徑或完整第一階段。", "rows": rows}
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
