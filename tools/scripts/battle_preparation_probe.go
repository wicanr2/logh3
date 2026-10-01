//go:build ignore

// #79 的受控原版常式診斷；入口見 docs/re/367-battle-preparation.md。
// 完整注入 bytes 留在本機收據，不能冒稱正常玩家路徑。
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

const memoryHash = "f70507aaa771a0b97d425037d06696c90385238c6e294a39a103a836960a6808"

func main() {
	output := flag.String("output", "", "新的本機收據；禁止覆寫")
	flag.Parse()
	if *output == "" {
		panic("必須指定新收據")
	}
	raw, err := os.ReadFile("/out/oracle/battle79-preparation-b/mem-in.bin")
	must(err)
	if len(raw) != 1<<20 || digest(raw) != memoryHash {
		panic("階段入口記憶體身份不符")
	}
	exe, err := os.ReadFile("/orig/GIN3.COM")
	must(err)
	if digest(exe) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		panic("原版程式身份不符")
	}
	o, err := oracle.LoadProgram("/orig/GIN3.COM", "/orig", "")
	must(err)
	defer o.Close()
	if o.SetBytes(oracle.Far(0, 0), raw) != len(raw) {
		panic("記憶體未完整注入")
	}
	o.SetTickRate(0)
	o.CallNear(oracle.Far(0xF000, 0), 0, oracle.CallRegs{DS: 0x54B3, ES: 0x54B3, SetDS: true, SetES: true})
	ds := oracle.Far(0x54B3, 0)
	word := func(off int) uint16 { return binary.LittleEndian.Uint16(raw[ds.Linear()+uint32(off):]) }
	pool := oracle.Far(word(0xCFB6), word(0xCFB4))
	planet := oracle.Far(word(0xCFFA), word(0xCFF8))
	snap := o.Save()
	rows := []map[string]any{}
	for _, name := range []string{"supply-priority", "supply-radius-192", "supply-radius-193", "repair-priority", "repair-supply-kind", "repair-radius-180", "repair-radius-181", "planet-repair-cap", "planet-repair-partial", "planet-repair-empty", "planet-repair-full", "planet-supply-kind", "defence-low-product", "defence-product-wrap", "defence-add-wrap", "defence-already-above"} {
		o.Restore(snap)
		injections := []map[string]any{}
		set := func(at oracle.Addr, b []byte) {
			if o.SetBytes(at, b) != len(b) {
				panic("注入未完整寫入")
			}
			injections = append(injections, map[string]any{"linear": at.Linear(), "hex": hex.EncodeToString(b)})
		}
		set16 := func(off, value int) {
			b := make([]byte, 2)
			put16(b, 0, value)
			set(oracle.Far(ds.Seg, uint16(off)), b)
		}
		set16(0x9894, 0) // 只驗資源，不進艦隊士氣報告；這是明示診斷限制。
		set16(0x9896, 3)
		set16(0x98BA, 0)
		set16(0x98BC, 1)
		set16(0x98BE, 2)
		set(oracle.Far(ds.Seg, 0xCFFC), []byte{0})
		set(oracle.Far(ds.Seg, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
		units := make([][]byte, 3)
		for i := range units {
			b := make([]byte, 78)
			b[0], b[1], b[2], b[3], b[4], b[30] = 1, 16, 31, byte(2-i), 0, 127
			put16(b, 8, 100)
			put16(b, 10, 100)
			put16(b, 12, 100)
			put16(b, 0x34, 1000)
			put16(b, 0x36, 1000)
			units[i] = b
		}
		entry := oracle.Far(0x1930, 0x019D)
		p := make([]byte, 82)
		p[17] = 1
		put16(p, 20, 3300)
		put16(p, 22, 3300)
		put16(p, 24, 5)
		put16(p, 70, 1000)
		put16(p, 72, 1000)
		switch name {
		case "supply-priority", "supply-radius-192", "supply-radius-193":
			units[0][4] = 5
			put32(units[0], 14, 450)
			put32(units[0], 18, 950)
			if name != "supply-priority" {
				set16(0x9896, 2)
				distance := 192
				if name == "supply-radius-193" {
					distance = 193
				}
				put16(units[1], 0x34, 1000+distance)
			}
		case "repair-priority", "repair-supply-kind", "repair-radius-180", "repair-radius-181":
			entry = oracle.Far(0x1AAB, 0x006D)
			units[0][4] = 6
			put32(units[0], 14, 70)
			put16(units[1], 8, 50)
			put16(units[2], 8, 50)
			if name == "repair-supply-kind" {
				units[2][4] = 5
			}
			if name == "repair-radius-180" || name == "repair-radius-181" {
				set16(0x9896, 2)
				distance := 180
				if name == "repair-radius-181" {
					distance = 181
				}
				put16(units[1], 0x34, 1000+distance)
			}
		case "planet-repair-cap", "planet-repair-partial", "planet-repair-empty", "planet-repair-full", "planet-supply-kind":
			entry = oracle.Far(0x1AAB, 0x006D)
			set(oracle.Far(ds.Seg, 0xCFFC), []byte{1})
			set16(0x9896, 1)
			put16(units[0], 8, 50)
			put16(units[0], 10, 100)
			put16(units[0], 12, 75)
			put32(p, 32, 100)
			if name == "planet-repair-partial" {
				put32(p, 32, 20)
			}
			if name == "planet-repair-empty" {
				put32(p, 32, 0)
			}
			if name == "planet-repair-full" {
				put16(units[0], 8, 100)
			}
			if name == "planet-supply-kind" {
				units[0][4] = 5
				put32(p, 32, 0)
				put32(p, 32+5*4, 20)
			}
		default:
			entry = oracle.Far(0x181B, 0x0083)
			set(oracle.Far(ds.Seg, 0xCFFC), []byte{1})
			p[17] = 2 // 無主仍回復。
			put16(p, 20, 1000)
			put16(p, 22, 3300)
			put16(p, 24, 5)
			if name == "defence-product-wrap" {
				put16(p, 22, 12800)
				put16(p, 24, 10)
			}
			if name == "defence-add-wrap" {
				put16(p, 20, 65400)
				put16(p, 22, 65535)
				put16(p, 24, 1)
			}
			if name == "defence-already-above" {
				put16(p, 20, 4000)
			}
		}
		for i, b := range units {
			set(oracle.Far(pool.Seg, pool.Off+uint16(78*i)), b)
		}
		set(planet, p)
		_, err := o.Call(entry)
		must(err)
		before, after := []string{}, []string{}
		for i, b := range units {
			before = append(before, hex.EncodeToString(b))
			after = append(after, hex.EncodeToString(o.Bytes(oracle.Far(pool.Seg, pool.Off+uint16(78*i)), 78)))
		}
		rows = append(rows, map[string]any{"name": name, "entry": fmt.Sprintf("%04X:%04X", entry.Seg, entry.Off),
			"injections": injections, "units_before_hex": before, "units_after_hex": after,
			"planet_before_hex": hex.EncodeToString(p), "planet_after_hex": hex.EncodeToString(o.Bytes(planet, 82)),
			"rng_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0x94DC), 4))})
	}
	result := map[string]any{"schema": "logh3.battle-preparation-oracle.v1", "input_sha256": digest(raw),
		"program_sha256": digest(exe), "seed": "0x12345678", "seed_linear": "lin:5E00C", "tick": 0,
		"setup": "完整 1 MiB 記憶體；新 LoadProgram 堆疊；CallNear F000:0000 設 DS=ES=54B3；每案 Restore；完整額外注入逐筆列出。",
		"scope": "原版常式受控診斷；戰場艦隊數清零跳過士氣報告；不證明正常玩家路徑或完整八步。", "rows": rows}
	data, err := json.MarshalIndent(result, "", "  ")
	must(err)
	data = append(data, '\n')
	f, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(err)
	_, err = f.Write(data)
	must(err)
	must(f.Close())
	fmt.Println("收據", *output, "SHA-256", digest(data))
}

func put16(b []byte, off, v int) { binary.LittleEndian.PutUint16(b[off:], uint16(v)) }
func put32(b []byte, off, v int) { binary.LittleEndian.PutUint32(b[off:], uint32(v)) }
func digest(b []byte) string     { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
