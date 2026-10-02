//go:build ignore

// #79 移動單步診斷；計畫含完整注入，入口與限制見研究372。
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

type injection struct {
	Linear uint32 `json:"linear"`
	Hex    string `json:"hex"`
}
type observation struct {
	Label  string `json:"label"`
	Linear uint32 `json:"linear"`
	Length int    `json:"length"`
}
type call struct {
	Segment uint16   `json:"segment"`
	Offset  uint16   `json:"offset"`
	Args    []uint16 `json:"args"`
}
type scenario struct {
	Name       string      `json:"name"`
	Injections []injection `json:"injections"`
	Calls      []call      `json:"calls"`
}
type plan struct {
	Common       []injection   `json:"common"`
	Observations []observation `json:"observations"`
	Cases        []scenario    `json:"cases"`
}

func main() {
	input := flag.String("plan", "", "完整注入計畫")
	output := flag.String("output", "", "新收據，禁止覆寫")
	flag.Parse()
	if *input == "" || *output == "" {
		panic("必須指定計畫與新輸出")
	}
	planRaw, err := os.ReadFile(*input)
	must(err)
	var p plan
	must(json.Unmarshal(planRaw, &p))
	raw, err := os.ReadFile("/out/oracle/battle79-move-entry-a/mem-in.bin")
	must(err)
	if digest(raw) != "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d" || len(raw) != 1<<20 {
		panic("來源記憶體身份不符")
	}
	exe, err := os.ReadFile("/orig/GIN3.COM")
	must(err)
	if digest(exe) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		panic("原版身份不符")
	}
	o, err := oracle.LoadProgram("/orig/GIN3.COM", "/orig", "")
	must(err)
	defer o.Close()
	o.SetBytes(oracle.Far(0, 0), raw)
	o.SetTickRate(0)
	o.CallNear(oracle.Far(0xF000, 0), 0, oracle.CallRegs{DS: 0x54B3, ES: 0x54B3, SetDS: true, SetES: true})
	set := func(list []injection) {
		for _, v := range list {
			b, err := hex.DecodeString(v.Hex)
			must(err)
			if len(b) == 0 || v.Linear+uint32(len(b)) > 1<<20 {
				panic("無效注入範圍")
			}
			o.SetBytes(addr(v.Linear), b)
			if hex.EncodeToString(o.Bytes(addr(v.Linear), len(b))) != v.Hex {
				panic("注入回讀不符")
			}
		}
	}
	set(p.Common)
	snap := o.Save()
	calls := []map[string]any{}
	for _, a := range []oracle.Addr{oracle.Far(0x2583, 0x040C), oracle.Far(0x2583, 0x0451), oracle.Far(0x25DB, 0x0714), oracle.Far(0x25DB, 0x0535)} {
		a := a
		o.OnCall(a, func(o *oracle.Oracle) {
			r := o.Regs()
			args := []uint16{}
			for i := 0; i < 8; i++ {
				args = append(args, o.Word(oracle.Far(r.SS, r.SP+4+uint16(i*2))))
			}
			calls = append(calls, map[string]any{"entry": a.String(), "args": args})
		})
	}
	observe := func() map[string]string {
		out := map[string]string{}
		for _, v := range p.Observations {
			out[v.Label] = hex.EncodeToString(o.Bytes(addr(v.Linear), v.Length))
		}
		return out
	}
	rows := []map[string]any{}
	for _, c := range p.Cases {
		o.Restore(snap)
		set(c.Injections)
		before := observe()
		calls = nil
		returns := []uint32{}
		for _, x := range c.Calls {
			ret, err := o.CallBudget(2_000_000, oracle.Far(x.Segment, x.Offset), x.Args...)
			must(err)
			returns = append(returns, ret)
		}
		rows = append(rows, map[string]any{"name": c.Name, "before": before, "after": observe(), "returns": returns, "calls": calls, "seed_after_hex": hex.EncodeToString(o.Bytes(oracle.Far(0x54B3, 0x94DC), 4))})
	}
	result := map[string]any{"schema": "logh3.battle-step-oracle.v1", "program_sha256": digest(exe), "input_sha256": digest(raw), "plan_sha256": digest(planRaw), "plan": p, "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "seed": "0x12345678", "seed_linear": "lin:5E00C", "tick": 0, "scope": "原版常式診斷；完整RAM／新堆疊、DS=ES=54B3，每案Restore。沒有替身或改原版指令；冷VGA及新DOS檔案服務，畫面不作正式UI證據。", "rows": rows}
	b, err := json.MarshalIndent(result, "", "  ")
	must(err)
	b = append(b, '\n')
	f, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	must(err)
	_, err = f.Write(b)
	must(err)
	must(f.Close())
	fmt.Println("案例", len(rows), "收據", *output, "SHA-256", digest(b))
}
func addr(linear uint32) oracle.Addr { return oracle.Far(uint16(linear>>4), uint16(linear&15)) }
func digest(b []byte) string         { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
