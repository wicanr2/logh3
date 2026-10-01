//go:build ignore

// #53 原版常式受控診斷。由 dosgolem-logh3/tools/go.sh run 執行，
// 入口與限制見 docs/spec/078-battle-rule-corrections.md。
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

const memoryHash = "3d8d1e7b6a5332253736d31aca9c056c9d2a16255a568fa0498237e5dc954ae6"
const programHash = "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918"

func main() {
	input := flag.String("input", "/out/aspre.bin", "固定原版 1 MiB 記憶體")
	output := flag.String("output", "", "新的收據檔名；禁止覆寫")
	flag.Parse()
	if *output == "" {
		panic("必須指定新收據檔名")
	}
	raw, err := os.ReadFile(*input)
	must(err)
	if len(raw) != 1<<20 || digest(raw) != memoryHash {
		panic("原版檢查點身份不符")
	}
	exe, err := os.ReadFile("/orig/GIN3.COM")
	must(err)
	if digest(exe) != programHash {
		panic("GIN3.COM 身份不符")
	}
	o, err := oracle.LoadProgram("/orig/GIN3.COM", "/orig", "")
	must(err)
	defer o.Close()
	if o.SetBytes(oracle.Far(0, 0), raw) != len(raw) {
		panic("記憶體注入不完整")
	}
	o.SetTickRate(0)
	// 只設定直接呼叫的 DS／ES；CallNear 不執行 trampoline。
	o.CallNear(oracle.Far(0xF000, 0), 0, oracle.CallRegs{DS: 0x54B3, ES: 0x54B3, SetDS: true, SetES: true})
	seed := make([]byte, 4)
	binary.LittleEndian.PutUint32(seed, 0x12345678)
	o.SetBytes(oracle.Far(0x54B3, 0x94DC), seed)
	snap := o.Save()
	rows := []map[string]any{}
	for _, p := range [][2]int{{0, 0}, {1, 1}, {4, 1}, {1, 4}, {-1, 1}, {1, -1}, {-1, -1}, {0, 1}, {0, -1}, {1, 0}, {-1, 0}, {2031, 2023}} {
		o.Restore(snap)
		value, err := o.Call(oracle.Far(0x2BE1, 0x011D), uint16(p[0]), uint16(p[1]))
		must(err)
		rows = append(rows, map[string]any{"dx": p[0], "dy": p[1], "angle": value & 255})
	}
	tableRaw := raw[0x54B30+0x593A : 0x54B30+0x59BA]
	table := []int{}
	for i := 0; i < len(tableRaw); i += 2 {
		table = append(table, int(binary.LittleEndian.Uint16(tableRaw[i:])))
	}
	poolOff := binary.LittleEndian.Uint16(raw[0x54B30+0xCFB4:])
	poolSeg := binary.LittleEndian.Uint16(raw[0x54B30+0xCFB6:])
	unit := oracle.Far(poolSeg, poolOff)
	losses := []map[string]any{}
	for _, c := range [][2]int{{0, 255}, {0, 0}, {5, 255}, {5, 0}, {7, 255}} {
		o.Restore(snap)
		rec := append([]byte(nil), raw[unit.Linear():unit.Linear()+78]...)
		rec[4] = byte(c[0])
		put16(rec, 8, 100)
		put16(rec, 10, 100)
		put16(rec, 12, 100)
		put32(rec, 14, 1003)
		put32(rec, 18, 301)
		o.SetBytes(unit, rec)
		flag1 := oracle.Far(0x54B3, uint16(0x4E3A+c[0]*26+14))
		flag2 := oracle.Far(0x54B3, uint16(0x4E3A+c[0]*26+18))
		o.SetByte(flag1, byte(c[1]))
		o.SetByte(flag2, byte(c[1]))
		value, err := o.Call(oracle.Far(0x203A, 0x0593), unit.Off, unit.Seg, 30, 0, 0)
		must(err)
		losses = append(losses, map[string]any{"kind": c[0], "amount": 30, "return": value & 65535,
			"before_hex": hex.EncodeToString(rec), "after_hex": hex.EncodeToString(o.Bytes(unit, 78)),
			"unit_linear": unit.Linear(), "kind_flag_linear": []uint32{flag1.Linear(), flag2.Linear()}, "kind_flags": c[1],
			"rng_after": o.Bytes(oracle.Far(0x54B3, 0x94DC), 4)})
	}
	spreads := []map[string]any{}
	for _, count := range []int{1, 2} {
		o.Restore(snap)
		f := oracle.Far(0x54B3, 0xC074)
		fleetBefore := o.Bytes(f, 122)
		fleetBefore[15] = byte(count)
		before := []string{}
		after := []string{}
		for i := 0; i < count; i++ {
			at := oracle.Far(poolSeg, poolOff+uint16(78*i))
			rec := append([]byte(nil), raw[at.Linear():at.Linear()+78]...)
			rec[4] = 0
			rec[3] = byte(i)
			put16(rec, 8, 100)
			put16(rec, 10, 100)
			put16(rec, 12, 100)
			put32(rec, 14, 1003)
			put32(rec, 18, 301)
			o.SetBytes(at, rec)
			before = append(before, hex.EncodeToString(rec))
			put16(fleetBefore, 16+4*i, i)
		}
		o.SetBytes(f, fleetBefore)
		amount := 130
		if count == 2 {
			amount = 201
		}
		value, err := o.Call(oracle.Far(0x20B4, 0x103D), f.Off, f.Seg, uint16(amount), 0)
		must(err)
		for i := 0; i < count; i++ {
			after = append(after, hex.EncodeToString(o.Bytes(oracle.Far(poolSeg, poolOff+uint16(78*i)), 78)))
		}
		spreads = append(spreads, map[string]any{"fleet": 0, "unit_start": 0, "amount": amount, "flag": 0,
			"return": value & 65535, "display_percent": o.Byte(oracle.Far(0x54B3, 0xA768)), "before_hex": before, "after_hex": after,
			"fleet_injected_hex": hex.EncodeToString(fleetBefore), "fleet_linear": f.Linear(), "pool_linear": unit.Linear(),
			"rng_after": o.Bytes(oracle.Far(0x54B3, 0x94DC), 4)})
	}
	result := map[string]any{"schema": "logh3.battle-rules-oracle.v1", "input": *input, "input_sha256": digest(raw),
		"exe": "GIN3.COM", "exe_sha256": digest(exe), "seed": "0x12345678", "seed_linear": "lin:5E00C",
		"tick": 0, "register_setup": "新 LoadProgram 的堆疊；完整注入 1 MiB；CallNear F000:0000 設 DS=ES=54B3；每案 Restore",
		"scope": "原版常式同狀態受控診斷；不證明正常玩家路徑或八步呼叫時機",
		"rows":  rows, "atan_table": table, "atan_table_sha256": digest(tableRaw), "losses": losses, "spreads": spreads}
	data, err := json.MarshalIndent(result, "", "  ")
	must(err)
	f, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	must(err)
	_, err = f.Write(append(data, '\n'))
	must(err)
	must(f.Close())
	fmt.Println("收據已寫入", *output, "SHA-256", digest(append(data, '\n')))
}
func put16(b []byte, off, v int) { binary.LittleEndian.PutUint16(b[off:], uint16(v)) }
func put32(b []byte, off, v int) { binary.LittleEndian.PutUint32(b[off:], uint32(v)) }
func digest(b []byte) string     { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
