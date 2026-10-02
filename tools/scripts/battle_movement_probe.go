//go:build ignore

// #79 移動初始化與自主移動停止條件；入口／限制見研究371。
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
	output := flag.String("output", "", "新的本機收據，禁止覆寫")
	flag.Parse()
	if *output == "" {
		panic("必須指定新收據")
	}
	raw, err := os.ReadFile("/out/oracle/battle79-move-entry-a/mem-in.bin")
	must(err)
	if len(raw) != 1<<20 || digest(raw) != "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d" {
		panic("移動階段輸入身份不符")
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
		panic("輸入未完整載入")
	}
	o.SetTickRate(0)
	o.CallNear(oracle.Far(0xF000, 0), 0, oracle.CallRegs{DS: 0x54B3, ES: 0x54B3, SetDS: true, SetES: true})
	ds := oracle.Far(0x54B3, 0)
	word := func(off int) uint16 { return binary.LittleEndian.Uint16(raw[int(ds.Linear())+off:]) }
	pool := oracle.Far(word(0xCFB6), word(0xCFB4))
	fleet := oracle.Far(ds.Seg, 0xC62C) // 原版清單第0支，表槽12；明示診斷注入。
	snap := o.Save()
	rows := []map[string]any{}
	for _, active := range []bool{false, true} {
		for _, budget := range []int{-1, 0, 1, 8*65536 - 1, 8 * 65536, 8*65536 + 1, 1000 * 65536} {
			o.Restore(snap)
			injections := []map[string]any{}
			set := func(at oracle.Addr, b []byte) {
				if o.SetBytes(at, b) != len(b) {
					panic("注入不完整")
				}
				injections = append(injections, map[string]any{"linear": at.Linear(), "hex": hex.EncodeToString(b)})
			}
			f := o.Bytes(fleet, 122)
			f[15] = 2
			put16(f, 0x10, 0)
			put16(f, 0x14, 1)
			put16(f, 0x38, 2)
			f[0x3A] = 0xFF
			if active {
				f[0x3A] = 0
			}
			f[0x48] = 1 // 艦隊戰鬥狀態仍只讀分隊巡航速度。
			put32(f, 0x52, 200*65536)
			put32(f, 0x56, budget)
			set(fleet, f)
			set(oracle.Far(ds.Seg, 0x9894), []byte{1, 0})
			set(oracle.Far(ds.Seg, 0x9898), []byte{12})
			set(oracle.Far(ds.Seg, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
			before := []string{}
			for i := 0; i < 3; i++ {
				b := make([]byte, 78)
				b[1] = 12
				b[4] = 0
				b[0x21] = 9
				b[0x22] = 31
				if i == 2 {
					b[4] = 7
					b[0x21] = 40
				}
				put32(b, 0x24, budget)
				set(oracle.Far(pool.Seg, pool.Off+uint16(i*78)), b)
				before = append(before, hex.EncodeToString(b))
			}
			ret, err := o.Call(oracle.Far(0x1C03, 0x0006))
			must(err)
			after := []string{}
			for i := 0; i < 3; i++ {
				after = append(after, hex.EncodeToString(o.Bytes(oracle.Far(pool.Seg, pool.Off+uint16(i*78)), 78)))
			}
			rows = append(rows, map[string]any{"name": fmt.Sprintf("refill-%t-%d", active, budget), "entry": "1C03:0006", "budget_before": budget, "scout_active": active, "fleet_before_hex": hex.EncodeToString(f), "fleet_after_hex": hex.EncodeToString(o.Bytes(fleet, 122)), "units_before_hex": before, "units_after_hex": after, "injections": injections, "return_low_byte": ret & 255, "seed_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0x94DC), 4))})
		}
	}
	for _, c := range []struct {
		name         string
		kind, budget int
		arrived      bool
	}{
		{"own-normal-zero", 0, 0, false}, {"own-normal-positive-tail", 0, 8*65536 - 1, false},
		{"own-normal-arrived-full", 0, 8 * 65536, true}, {"own-scout-zero", 7, 0, false},
		{"own-scout-positive-tail", 7, 8*65536 - 1, false}, {"own-scout-arrived", 7, 8 * 65536, true},
	} {
		o.Restore(snap)
		b := make([]byte, 78)
		b[1] = 12
		b[4] = byte(c.kind)
		put32(b, 0x24, c.budget)
		put16(b, 0x28, 1000)
		put16(b, 0x2A, 1000)
		put16(b, 0x34, 1000)
		put16(b, 0x36, 1000)
		if !c.arrived {
			put16(b, 0x28, 1100)
		}
		if o.SetBytes(pool, b) != len(b) {
			panic("注入不完整")
		}
		if o.SetBytes(oracle.Far(ds.Seg, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12}) != 4 {
			panic("seed注入失敗")
		}
		ret, err := o.Call(oracle.Far(0x1C61, 0x03AA), pool.Off, pool.Seg, fleet.Off, fleet.Seg)
		must(err)
		rows = append(rows, map[string]any{"name": c.name, "entry": "1C61:03AA", "units_before_hex": []string{hex.EncodeToString(b)}, "units_after_hex": []string{hex.EncodeToString(o.Bytes(pool, 78))}, "injections": []map[string]any{{"linear": pool.Linear(), "hex": hex.EncodeToString(b)}, {"linear": ds.Linear() + 0x94DC, "hex": "78563412"}}, "return_low_byte": ret & 255, "seed_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0x94DC), 4))})
	}
	result := map[string]any{"schema": "logh3.battle-movement-probe.v1", "input_sha256": digest(raw), "program_sha256": digest(exe), "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "seed": "0x12345678", "seed_linear": "lin:5E00C", "tick": 0, "scope": "原版常式診斷；全RAM重設、新堆疊、DS=ES=54B3，每案Restore。只驗初始化及不繪圖的停止分支，不證明完整移動／碰撞／正常玩家UI。", "rows": rows}
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

func put16(b []byte, off, v int) { binary.LittleEndian.PutUint16(b[off:], uint16(v)) }
func put32(b []byte, off, v int) { binary.LittleEndian.PutUint32(b[off:], uint32(v)) }
func digest(b []byte) string     { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
