// 第86輪的DRAFT組合探針。只驗既有rules，沒有正式Game接線。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/rules"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func read(p string) []byte   { b, e := os.ReadFile(p); must(e); return b }
func decode(s string) []byte { b, e := hex.DecodeString(s); must(e); return b }
func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

type checkpoint struct {
	Location string
	Step     int
	Tables   map[string]string
}
type row struct {
	Name           string
	Before, After  map[string]string
	Recalculations []checkpoint
}
type original struct{ Rows []row }

func main() {
	outputPath := flag.String("output", "workplace/promotion79-r86-global-probe-v3.json", "新輸出檔，拒絕覆寫")
	flag.Parse()
	const memoryPath = "workplace/oracle/ownership79-r82-entry-a/mem-in.bin"
	m := read(memoryPath)
	if digest(m) != "b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900" {
		panic("父RAM雜湊不同")
	}
	const ds = 0x54b30
	word := func(o int) int { return int(binary.LittleEndian.Uint16(m[o:])) }
	far := func(o int) int { return word(ds+o) + 16*word(ds+o+2) }
	pdata := append([]byte(nil), m[ds+0xcfb8:ds+0xcfb8+64]...)
	pbase := far(0x5336)
	pdata = append(pdata, m[pbase:pbase+81*82]...)
	_, e := rules.LoadPlanets(pdata)
	must(e)
	graph, e := rules.LoadStarGraph("docs/internals/data/star-graph.json")
	must(e)
	params := rules.UnrestParams{Table: make([]int, 5), Base: make([]int, 5)}
	for i := 0; i < 5; i++ {
		params.Table[i] = int(m[ds+0x4b28+i])
		params.Base[i] = word(ds + 0x147e + 4*i)
	}
	var raw original
	originalBytes := read("workplace/promotion79-r86-global-a.json")
	if !bytes.Equal(originalBytes, read("workplace/promotion79-r86-global-b.json")) {
		panic("原版雙跑不同")
	}
	must(json.Unmarshal(originalBytes, &raw))
	results := []map[string]any{}
	finals := []map[string]any{}
	failures, total := 0, 0
	for _, r := range raw.Rows {
		stellar, work := decode(r.Before["stellar"]), decode(r.Before["work"])
		planetInput := decode(r.Before["planet"])
		planets, err := rules.LoadPlanets(planetInput)
		must(err)
		owners := make([]int, 32)
		for i := range owners {
			owners[i] = rules.RegionOwnerOf(stellar[i*126 : (i+1)*126])
		}
		// 重算前EMS frame仍映射圖像；不能把那段bytes當路徑輸入。
		// 沿042從重算前歸屬獨立建表，重算返回後才核對EMS矩陣。
		paths := graph.UnrestPaths(owners)
		pathModel := make([]byte, 6144)
		for f := 0; f < 2; f++ {
			for i := 0; i < 32; i++ {
				for j := 0; j < 32; j++ {
					binary.LittleEndian.PutUint16(pathModel[f*2048+2*(i*32+j):], uint16(paths.Dist[f][i][j]))
					pathModel[4096+f*1024+i*32+j] = byte(paths.Mid[f][i][j])
				}
			}
		}
		for _, c := range r.Recalculations {
			if !bytes.Equal(planetInput, decode(c.Tables["planet"])) {
				panic("重算改了受控行星輸入")
			}
			if !bytes.Equal(pathModel, decode(c.Tables["paths"])) {
				panic("獨立路徑表與返回後EMS不同")
			}
			admirals, fleets := decode(c.Tables["admiral"]), decode(c.Tables["fleet"])
			wf, err := rules.ReadWork(work)
			must(err)
			params.Turn = wf.Turn
			params.Blend = wf.Blend
			params.Rate = wf.Rate
			for f := 0; f < 2; f++ {
				params.Ruler[f] = int(work[0x3b+f])
			}
			rules.ResumRegionEconomy(stellar, planets)
			u := rules.UnrestState{Regions: stellar, Planets: planets, Fleets: fleets, Admirals: admirals, Graph: graph, Params: params, Paths: paths}
			taxes := [2]int{}
			for i := 0; i < 32; i++ {
				rec := stellar[i*126 : (i+1)*126]
				owner := rules.RegionOwnerOf(rec)
				if owner < 2 {
					taxes[owner] += rules.RegionTax(rec)
				}
				rec[0x7d] = byte(u.Unrest(i))
			}
			upkeep, moves := [2]int{}, [2]int{}
			for i := 0; i < 32; i++ {
				f := fleets[i*122 : (i+1)*122]
				if f[4] == 255 {
					continue
				}
				seat := int(f[1])
				upkeep[seat] += int(int16(binary.LittleEndian.Uint16(f[0xa:])))
				if f[8] != 2 && f[8] != 5 {
					moves[seat] += int(int16(binary.LittleEndian.Uint16(f[0xc:])))
				}
			}
			for f := 0; f < 2; f++ {
				for _, item := range []struct{ o, v int }{{0x56, taxes[f]}, {0x5e, upkeep[f]}, {0x6e, moves[f]}} {
					binary.LittleEndian.PutUint32(work[item.o+4*f:], uint32(item.v))
				}
				available := int(int32(binary.LittleEndian.Uint32(work[0x3e+4*f:]))) + taxes[f] - upkeep[f] - moves[f]
				available = int(int32(available))
				budget := 0
				if available >= 0 {
					if available >= 21474836 {
						budget = (available / 100) * int(work[0x7e+f])
					} else {
						budget = available * int(work[0x7e+f]) / 100
					}
				}
				binary.LittleEndian.PutUint32(work[0x46+4*f:], uint32(budget))
			}
			entry := map[string]any{"name": r.Name, "location": c.Location, "step": c.Step, "tax": taxes, "upkeep": upkeep, "move": moves, "stellar_sha256": digest(stellar), "work_sha256": digest(work), "path_sha256": digest(pathModel), "planet_sha256": digest(planetInput), "admiral_input_sha256": digest(admirals), "fleet_input_sha256": digest(fleets), "ship_input_sha256": digest(decode(c.Tables["ship"]))}
			for n, b := range map[string][]byte{"stellar": stellar, "work": work} {
				expected := decode(c.Tables[n])
				diff := []map[string]int{}
				for i, v := range b {
					if v != expected[i] {
						diff = append(diff, map[string]int{"offset": i, "model": int(v), "original": int(expected[i])})
					}
				}
				entry[n+"_differences"] = diff
				failures += len(diff)
			}
			results = append(results, entry)
			total++
		}
		for n, b := range map[string][]byte{"stellar": stellar, "work": work, "planet": planetInput} {
			if !bytes.Equal(b, decode(r.After[n])) {
				panic("最終表不同：" + r.Name + " " + n)
			}
		}
		finals = append(finals, map[string]any{"name": r.Name, "stellar_sha256": digest(stellar), "work_sha256": digest(work), "planet_sha256": digest(planetInput)})
	}
	output := map[string]any{"schema": "logh3.battle-promotion-global-probe.v3", "scope": "DRAFT診斷組合；已由第85輪核過的admiral/fleet作輸入，行星固定；路徑從重算前歸屬獨立建立並核返回後EMS，不把映射圖像的frame當路徑", "checkpoints": total, "different_bytes": failures, "params_constants": params, "parent_planet_sha256": digest(pdata), "cases": results, "finals": finals}
	b, e := json.MarshalIndent(output, "", "  ")
	must(e)
	f, e := os.OpenFile(*outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	must(e)
	_, e = f.Write(append(b, '\n'))
	must(e)
	must(f.Close())
	fmt.Printf("重算 %d 次，STELLAR／WORK 差異 %d bytes\n", total, failures)
	if failures != 0 {
		os.Exit(1)
	}
}
