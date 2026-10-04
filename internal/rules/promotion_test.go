package rules

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type promotionOracle struct {
	Rows []struct {
		Name          string
		Case          []int
		Before, After map[string]string
		SeedAfter     string `json:"seed_after_hex"`
		Calls         []struct {
			Location string
			Args     []int
			AX       int
		}
		Recalculations []struct{ Tables map[string]string }
	}
}

func promotionHex(t *testing.T, s string) []byte {
	t.Helper()
	b, e := hex.DecodeString(s)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func promotionInputs(t *testing.T) (promotionOracle, PromotionContext) {
	t.Helper()
	read := func(path, hash string) []byte {
		b, e := os.ReadFile(filepath.Join("../..", path))
		if os.IsNotExist(e) {
			t.Skip("缺原版升級收據；重生入口tools/battle_promotion_global_oracle.sh")
		}
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("升級來源雜湊不符", path, e)
		}
		return b
	}
	b := read("workplace/promotion79-r86-global-a.json", "2bc8fa7bf8dc4081fad414d23e7bfeabb76cd8df510d53f3442136283cb1a435")
	if !bytes.Equal(b, read("workplace/promotion79-r86-global-b.json", "2bc8fa7bf8dc4081fad414d23e7bfeabb76cd8df510d53f3442136283cb1a435")) {
		t.Fatal("原版雙跑不同")
	}
	var raw promotionOracle
	if e := json.Unmarshal(b, &raw); e != nil {
		t.Fatal(e)
	}
	c := PromotionContext{}
	var e error
	c.Spec, e = LoadPromotionSpec("../../docs/internals/data/admiral-promotion-spec.json")
	if e != nil {
		t.Fatal(e)
	}
	c.Kinds, e = LoadShipKindTable("../../docs/internals/data/ship-kind-table.json")
	if e != nil {
		t.Fatal(e)
	}
	editor, e := LoadEditorTables("../../docs/internals/data/fleet-editor.json")
	if e != nil {
		t.Fatal(e)
	}
	c.Mobility = editor.Mobility
	c.Graph, e = LoadStarGraph("../../docs/internals/data/star-graph.json")
	if e != nil {
		t.Fatal(e)
	}
	m := read("workplace/oracle/ownership79-r82-entry-a/mem-in.bin", "b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900")
	for i := 0; i < 5; i++ {
		c.Params.Table = append(c.Params.Table, int(m[0x54b30+0x4b28+i]))
		c.Params.Base = append(c.Params.Base, int(binary.LittleEndian.Uint16(m[0x54b30+0x147e+4*i:])))
	}
	return raw, c
}

func promotionSlot(t *testing.T, before map[string]string, c *PromotionContext) *Slot {
	t.Helper()
	p := promotionHex(t, before["planet"])
	s, e := LoadSlot(SlotFiles{Admiral: promotionHex(t, before["admiral"]), Fleet: promotionHex(t, before["fleet"]), Ship: promotionHex(t, before["ship"]), Stellar: promotionHex(t, before["stellar"]), Work: promotionHex(t, before["work"]), PlanetHeader: p[:64], Planet: p[64:]})
	if e != nil {
		t.Fatal(e)
	}
	c.Params.Turn, c.Params.Blend, c.Params.Rate = s.Turn, s.Blend, s.Rate
	c.Military = s.Military
	for f := 0; f < 2; f++ {
		c.Params.Ruler[f] = int(s.Work[0x3b+f])
	}
	owners := make([]int, 32)
	for i := range owners {
		owners[i] = int(s.Stellar[i*RegionRecord+0x12])
	}
	c.Paths = c.Graph.UnrestPaths(owners)
	return s
}

func TestBattlePromotionOriginalTables(t *testing.T) {
	raw, base := promotionInputs(t)
	attempts, upgrades, randoms, checkpoints := 0, 0, 0, 0
	for _, row := range raw.Rows {
		t.Run(row.Name, func(t *testing.T) {
			c := base
			s := promotionSlot(t, row.Before, &c)
			rnd, expected := NewRand(0x12345678), NewRand(0x12345678)
			var flags uint16
			cursor := 0
			compare := func(tables map[string]string) {
				for name, b := range map[string][]byte{"admiral": s.Admiral, "fleet": s.Fleet, "ship": s.Ship, "stellar": s.Stellar, "work": s.Work} {
					want := promotionHex(t, tables[name])
					if !bytes.Equal(b, want) {
						for i, v := range b {
							if v != want[i] {
								t.Fatalf("%s首個差異+%X：%02X／原版%02X", name, i, v, want[i])
							}
						}
					}
				}
			}
			for _, call := range row.Calls {
				switch call.Location {
				case "28E0:05F9":
					var e error
					flags, e = s.PromoteBattleAdmiral(row.Case[0], call.Args[0]&255, c, rnd)
					if e != nil {
						t.Fatal(e)
					}
					attempts++
				case "28E0:098D", "28E0:09A0", "28E0:070D", "28E0:078A", "28E0:0857", "28E0:08E2":
					if expected.Next() != call.AX {
						t.Fatal("原版RND與固定LCG不同")
					}
					randoms++
				case "185C:065B", "185C:06A8":
					if int(flags) != call.AX || rnd.Seed() != expected.Seed() {
						t.Fatalf("成長旗標／seed不符：%X／%X，%X／%X", flags, call.AX, rnd.Seed(), expected.Seed())
					}
					if flags != 0 {
						upgrades++
					}
				case "185C:066E", "185C:06BB":
					compare(row.Recalculations[cursor].Tables)
					cursor++
					checkpoints++
				}
			}
			compare(row.After)
			if rnd.Seed() != binary.LittleEndian.Uint32(promotionHex(t, row.SeedAfter)) || cursor != len(row.Recalculations) {
				t.Fatal("最終seed或重算次數不符")
			}
			if !bytes.Equal(promotionHex(t, row.Before["planet"]), promotionHex(t, row.After["planet"])) {
				t.Fatal("原版改行星輸入")
			}
		})
	}
	if len(raw.Rows) != 66 || attempts != 121 || upgrades != 53 || randoms != 473 || checkpoints != 53 {
		t.Fatalf("案例／嘗試／成長／RND／重算：%d/%d/%d/%d/%d", len(raw.Rows), attempts, upgrades, randoms, checkpoints)
	}
	t.Log("66案、121次嘗試、53次完整五表及473個RND符合原版")
}

func TestBattlePromotionInvalidAtomic(t *testing.T) {
	raw, c := promotionInputs(t)
	for _, bad := range []string{"pool", "scout", "full", "faction", "paths", "spec", "admiral"} {
		t.Run(bad, func(t *testing.T) {
			c := c
			s := promotionSlot(t, raw.Rows[1].Before, &c)
			f := s.fleet(raw.Rows[1].Case[0])
			adm := int(f[4])
			switch bad {
			case "pool":
				binary.LittleEndian.PutUint16(f[0x10:], 352)
			case "scout":
				binary.LittleEndian.PutUint16(f[0x38:], 352)
			case "full":
				pool := int(binary.LittleEndian.Uint16(f[0x10:]))
				binary.LittleEndian.PutUint16(s.Ship[pool*78+8:], 1)
				binary.LittleEndian.PutUint16(s.Ship[pool*78+10:], 0)
			case "faction":
				f[1] = 2
			case "paths":
				c.Paths.Dist[0] = nil
			case "spec":
				c.Spec = nil
			case "admiral":
				adm = 82
			}
			before := [][]byte{bytes.Clone(s.Admiral), bytes.Clone(s.Fleet), bytes.Clone(s.Ship), bytes.Clone(s.Stellar), bytes.Clone(s.Work)}
			rnd := NewRand(0x12345678)
			if _, e := s.PromoteBattleAdmiral(raw.Rows[1].Case[0], adm, c, rnd); e == nil {
				t.Fatal("不合法資料未報錯")
			}
			for i, b := range [][]byte{s.Admiral, s.Fleet, s.Ship, s.Stellar, s.Work} {
				if !bytes.Equal(b, before[i]) {
					t.Fatal("失敗改了資料表", i)
				}
			}
			if rnd.Seed() != 0x12345678 {
				t.Fatal("失敗消耗RND")
			}
		})
	}
}
