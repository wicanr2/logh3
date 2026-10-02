//go:build ignore

// #79小地圖標記受控常式診斷；不代替正常玩家路徑。
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
	output := flag.String("output", "", "新收據，禁止覆寫")
	flag.Parse()
	if *output == "" {
		panic("必須指定新收據")
	}
	raw, err := os.ReadFile("/out/oracle/battle79-move-entry-a/mem-in.bin")
	mustMarker(err)
	if len(raw) != 1<<20 || markerDigest(raw) != "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d" {
		panic("初態身份不同")
	}
	exe, err := os.ReadFile("/orig/GIN3.COM")
	mustMarker(err)
	if markerDigest(exe) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		panic("原版身份不同")
	}
	o, err := oracle.LoadProgram("/orig/GIN3.COM", "/orig", "")
	mustMarker(err)
	defer o.Close()
	if o.SetBytes(oracle.Far(0, 0), raw) != len(raw) {
		panic("RAM未完整載入")
	}
	o.SetTickRate(0)
	o.CallNear(oracle.Far(0xF000, 0), 0, oracle.CallRegs{DS: 0x54B3, ES: 0x54B3, SetDS: true, SetES: true})
	ds := oracle.Far(0x54B3, 0)
	word := func(off int) uint16 { return binary.LittleEndian.Uint16(raw[int(ds.Linear())+off:]) }
	pool := oracle.Far(word(0xCFB6), word(0xCFB4))
	snap := o.Save()
	rows := []map[string]any{}
	for _, enabled := range []byte{0, 1} {
		for _, forced := range []uint16{0, 1} {
			for _, detail := range []byte{0, 2} {
				for _, faction := range []byte{0, 1, 2} {
					for _, flags := range []byte{0x0F, 0x85, 0xC5, 0x45} {
						o.Restore(snap)
						b := make([]byte, 78)
						b[0], b[0x17], b[0x19] = faction, detail, flags
						binary.LittleEndian.PutUint16(b[0x1A:], 111)
						binary.LittleEndian.PutUint16(b[0x1C:], 222)
						binary.LittleEndian.PutUint16(b[0x34:], 1000)
						binary.LittleEndian.PutUint16(b[0x36:], 900)
						injections := []map[string]any{}
						set := func(at oracle.Addr, bytes []byte) {
							if o.SetBytes(at, bytes) != len(bytes) {
								panic("注入不完整")
							}
							injections = append(injections, map[string]any{"linear": at.Linear(), "hex": hex.EncodeToString(bytes)})
						}
						set(pool, b)
						set(oracle.Far(ds.Seg, 0xA426), []byte{enabled})
						set(oracle.Far(ds.Seg, 0x94DC), []byte{0x78, 0x56, 0x34, 0x12})
						_, err = o.Call(oracle.Far(0x1727, 0x02E2), pool.Off, pool.Seg, forced)
						mustMarker(err)
						rows = append(rows, map[string]any{"name": fmt.Sprintf("switch-%d-%d-%d-%d-%02X", enabled, forced, detail, faction, flags), "entry": "1727:02E2", "enabled": enabled, "forced": forced, "detail": detail, "faction": faction, "before_hex": hex.EncodeToString(b), "after_hex": hex.EncodeToString(o.Bytes(pool, 78)), "injections": injections, "seed_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(ds.Seg, 0x94DC), 4))})
					}
				}
			}
		}
	}
	result := map[string]any{"schema": "logh3.battle-marker-probe.v1", "input_sha256": markerDigest(raw), "program_sha256": markerDigest(exe), "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "seed": "0x12345678", "tick": 0, "scope": "原版常式診斷；每案Restore完整RAM、新堆疊、DS=ES=54B3，注入均列出。只驗標記狀態及快照，不驗VGA畫面或正常呼叫時機。", "rows": rows}
	data, err := json.MarshalIndent(result, "", "  ")
	mustMarker(err)
	data = append(data, '\n')
	f, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	mustMarker(err)
	_, err = f.Write(data)
	mustMarker(err)
	mustMarker(f.Close())
	fmt.Println("收據", *output, "SHA-256", markerDigest(data), "案例", len(rows))
}

func markerDigest(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func mustMarker(err error) {
	if err != nil {
		panic(err)
	}
}
