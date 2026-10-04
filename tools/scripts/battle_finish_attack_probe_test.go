// 第90輪DRAFT診斷。只在容器暫存rules副本編譯，沒有production接線。
package rules

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestDraftWinnerAssault(t *testing.T) {
	nativePath, expected, outputPath := "/src/workplace/finish79-r89-a.json", "d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477", "/src/workplace/finish79-r90-probe-v2-base.json"
	if os.Getenv("WINNER_PROBE_GUARD") == "1" {
		nativePath = "/src/workplace/finish79-r90-guard-a.json"
		expected = "85fbd698622c307051183167d2315e9977abe5cac8fc8068611b31f4287ee335"
		outputPath = "/src/workplace/finish79-r90-probe-v2-guard.json"
	}
	raw, e := os.ReadFile(nativePath)
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != expected {
		t.Fatal("原版身份不符")
	}
	var original struct {
		Rows []struct {
			Name               string
			Mask, Human, Round int
			Before, After      map[string]string
			SeedAfterHex       string `json:"seed_after_hex"`
		}
	}
	if e = json.Unmarshal(raw, &original); e != nil {
		t.Fatal(e)
	}
	decode := func(s string) []byte {
		b, e := hex.DecodeString(s)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	slots, e := LoadFormationSlots("/src/docs/internals/data/battle-formation-slots.json")
	if e != nil {
		t.Fatal(e)
	}
	output := []map[string]any{}
	for _, row := range original.Rows {
		if row.Mask != 1 && row.Mask != 126 {
			continue
		}
		if row.Round != 0 && row.Round != 1 && row.Round != 5 {
			continue
		}
		winner := 0
		if row.Mask == 1 {
			winner = 1
		}
		before, after := map[string][]byte{}, map[string][]byte{}
		for k, v := range row.Before {
			before[k] = decode(v)
			after[k] = decode(row.After[k])
		}
		slot := &Slot{Fleet: bytes.Clone(before["fleet"]), Ship: bytes.Clone(before["ship"])}
		fleets := slot.Fleets(nil)
		region := int(before["work"][3])
		pall := bytes.Clone(before["planet"])
		first := 0
		for i := 0; i < region; i++ {
			first += int(binary.LittleEndian.Uint16(pall[2*i:]))
		}
		n := int(binary.LittleEndian.Uint16(pall[2*region:]))
		planets := pall[64+first*82 : 64+(first+n)*82]
		sites := make([]PlanetSite, n)
		for i := range sites {
			p := planets[i*82:]
			sites[i] = PlanetSite{X: int(binary.LittleEndian.Uint16(p[0x46:])), Y: int(binary.LittleEndian.Uint16(p[0x48:]))}
		}
		place := &StrategicPlacement{Slots: slots}
		sfs := []StrategicFleet{}
		for i := range fleets {
			f := &fleets[i]
			if f.Region != region {
				continue
			}
			if f.Faction != winner {
				place.Others = append(place.Others, f)
				continue
			}
			rec := slot.Fleet[f.Index*122:]
			adm := before["admiral"][f.Admiral*28:]
			sfs = append(sfs, StrategicFleet{Fleet: f, Ability40: int(rec[0x40]), Ability41: int(rec[0x41]), Command: int(adm[0x0B])})
		}
		r := NewRand(0x12345678)
		// 203A:04D1透過17A0:00AC讀當前戰場池清單；敗方在本階段不承受損失，取固定原池。
		probeGuard = func(p []byte) int {
			total := 0
			list := before["lists"]
			radius := FixedFromInt(PlanetQueryRadius(int(binary.LittleEndian.Uint16(p[0x14:]))))
			for i := 0; i < int(binary.LittleEndian.Uint16(list[2:])); i++ {
				pool := int(binary.LittleEndian.Uint16(list[38+i*2:]))
				sh := slot.Ship[pool*78 : (pool+1)*78]
				if sh[0] != p[0x11] || sh[0x1E]&0x80 != 0 || sh[4] > 1 {
					continue
				}
				if FixedDist(int(binary.LittleEndian.Uint16(p[0x46:])), int(binary.LittleEndian.Uint16(p[0x48:])), int(binary.LittleEndian.Uint16(sh[0x34:])), int(binary.LittleEndian.Uint16(sh[0x36:]))) > radius {
					continue
				}
				total += int(binary.LittleEndian.Uint16(sh[8:]))
				if total > 65535 {
					total = 65535
				}
			}
			return total / 100
		}
		res := probeWinnerAssault(planets, sites, winner, sfs, r, place, row.Round)
		differences := []string{}
		add := func(label string, got, want int) {
			if got != want {
				differences = append(differences, fmt.Sprintf("%s %d/%d", label, got, want))
			}
		}
		if !bytes.Equal(pall, after["planet"]) {
			for i, v := range pall {
				if v != after["planet"][i] {
					differences = append(differences, fmt.Sprintf("planet+%X %02X/%02X", i, v, after["planet"][i]))
					if len(differences) > 30 {
						break
					}
				}
			}
		}
		add("seed", int(r.Seed()), int(binary.LittleEndian.Uint32(decode(row.SeedAfterHex))))
		for _, sf := range sfs {
			f := sf.Fleet
			want := (&Slot{Fleet: after["fleet"], Ship: after["ship"]}).FleetComposition(f.Index)
			rec := after["fleet"][f.Index*122:]
			add(fmt.Sprintf("fleet%d.flags", f.Index), f.Flags, int(rec[0]))
			add(fmt.Sprintf("fleet%d.morale", f.Index), f.Morale, int(rec[0x3C]))
			add(fmt.Sprintf("fleet%d.x", f.Index), f.X, int(binary.LittleEndian.Uint16(rec[0x5C:])))
			add(fmt.Sprintf("fleet%d.y", f.Index), f.Y, int(binary.LittleEndian.Uint16(rec[0x5E:])))
			for i, g := range f.Composition {
				add(fmt.Sprintf("fleet%d.group%d.ships", f.Index, i), g.Ships, want[i].Ships)
				add(fmt.Sprintf("fleet%d.group%d.x", f.Index, i), g.X, want[i].X)
				add(fmt.Sprintf("fleet%d.group%d.y", f.Index, i), g.Y, want[i].Y)
			}
		}
		for a, v := range res.Merit {
			add(fmt.Sprintf("admiral%d.merit", a), int(binary.LittleEndian.Uint32(before["admiral"][a*28+2:]))+v, int(binary.LittleEndian.Uint32(after["admiral"][a*28+2:])))
		}
		for faction := 0; faction < 2; faction++ {
			for _, off := range []int{0xD0, 0xD8} {
				add(fmt.Sprintf("work%X.f%d", off, faction), int(binary.LittleEndian.Uint32(before["work"][off+4*faction:]))+res.Score[faction], int(binary.LittleEndian.Uint32(after["work"][off+4*faction:])))
			}
		}
		output = append(output, map[string]any{"name": row.Name, "differences": differences, "hits": res.Hits, "captured": res.Captured, "seed": r.Seed(), "fleets": fleets, "planet_hex": hex.EncodeToString(pall)})
		t.Log(row.Name, "hits", res.Hits, "captures", res.Captured, "differences", len(differences), differences)
	}
	count := 12
	if os.Getenv("WINNER_PROBE_GUARD") == "1" {
		count = 7
	}
	if len(output) != count {
		t.Fatalf("案例數%d", len(output))
	}
	b, e := json.MarshalIndent(output, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		t.Fatal(e)
	}
}
