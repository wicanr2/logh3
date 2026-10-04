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

	"github.com/wicanr2/logh3/internal/asset"
)

type finishOracleRow struct {
	Name                   string
	Mask, Human, Round, AX int
	Before, After          map[string]string
	SeedAfterHex           string `json:"seed_after_hex"`
	Calls                  []struct {
		Location string
		Args     []int
		Tables   map[string]string
		SeedHex  string `json:"seed_hex"`
	}
}

func finishOracle(t *testing.T, name, sha string) []finishOracleRow {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../../workplace", name))
	if os.IsNotExist(err) {
		t.Skip("缺固定dosgolem收據", name)
	}
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != sha {
		t.Fatal("原版收據雜湊不同", name)
	}
	var v struct {
		Rows  []finishOracleRow
		Head  string `json:"dosgolem_head"`
		Dirty bool   `json:"dosgolem_dirty"`
	}
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v.Head != "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c" || v.Dirty {
		t.Fatal("原版執行器身分不同")
	}
	return v.Rows
}

func finishSlot(t *testing.T, row finishOracleRow) (*Slot, []byte) {
	t.Helper()
	d := map[string][]byte{}
	for k, v := range row.Before {
		b, e := hex.DecodeString(v)
		if e != nil {
			t.Fatal(e)
		}
		d[k] = b
	}
	p, err := LoadPlanets(d["planet"])
	if err != nil {
		t.Fatal(err)
	}
	w, err := ReadWork(d["work"])
	if err != nil {
		t.Fatal(err)
	}
	return &Slot{Fleet: d["fleet"], Ship: d["ship"], Admiral: d["admiral"], Stellar: d["stellar"], Planets: p, Work: d["work"], WorkFields: w}, d["lists"]
}

func finishTables(s *Slot, list []byte) map[string][]byte {
	return map[string][]byte{"fleet": s.Fleet, "ship": s.Ship, "admiral": s.Admiral, "stellar": s.Stellar, "planet": s.Planets.Data, "work": s.Work, "lists": list}
}

func compareFinishTables(t *testing.T, s *Slot, list []byte, want map[string]string) {
	t.Helper()
	for key, got := range finishTables(s, list) {
		b, e := hex.DecodeString(want[key])
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(got, b) {
			count, first := 0, -1
			for i, v := range got {
				if i >= len(b) || v != b[i] {
					count++
					if first < 0 {
						first = i
					}
				}
			}
			t.Errorf("%s差%d bytes，首差+%X，長度%d/%d", key, count, first, len(got), len(b))
		}
	}
}

func finishCursor(t *testing.T, s *Slot, list []byte, winner, round int) *WinnerOccupation {
	t.Helper()
	kinds, e := LoadShipKindTable("../../docs/internals/data/ship-kind-table.json")
	if e != nil {
		t.Fatal(e)
	}
	slots, e := LoadFormationSlots("../../docs/internals/data/battle-formation-slots.json")
	if e != nil {
		t.Fatal(e)
	}
	terrain, e := LoadBattleTerrain("../../docs/internals/data/battle-terrain.json")
	if e != nil {
		t.Fatal(e)
	}
	region := int(s.Work[3])
	input, e := os.ReadFile(fmt.Sprintf("../../org_game/extracted/MAP%02d.CMP", region))
	if os.IsNotExist(e) {
		t.Skip("缺原版地圖")
	}
	if e != nil {
		t.Fatal(e)
	}
	cells, e := asset.DecodeCMP(input, asset.BattleMapCols, asset.BattleMapRows)
	if e != nil {
		t.Fatal(e)
	}
	var pools []int
	for i := 0; i < int(binary.LittleEndian.Uint16(list[2:])); i++ {
		pools = append(pools, int(binary.LittleEndian.Uint16(list[38+i*2:])))
	}
	c, e := NewWinnerOccupation(s, region, winner, round, pools, kinds, slots, func(x, y int) bool { return terrain.Passable(cells, x, y) }, NewRand(0x12345678))
	if e != nil {
		t.Fatal(e)
	}
	return c
}

// 全67案比完整六表與有效清單。輸入來自固定原版，而非重製公式產生。
func TestBattleFinishMatchesOriginal(t *testing.T) {
	base := finishOracle(t, "finish79-r89-a.json", "d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477")
	guard := finishOracle(t, "finish79-r90-guard-a.json", "85fbd698622c307051183167d2315e9977abe5cac8fc8068611b31f4287ee335")
	if len(base) != 60 || len(guard) != 7 {
		t.Fatal("案例數不同")
	}
	for _, row := range append(base, guard...) {
		t.Run(row.Name, func(t *testing.T) {
			s, list := finishSlot(t, row)
			var ids []int
			for _, i := range list[4 : 4+int(list[0])] {
				ids = append(ids, int(i))
			}
			d, e := s.DecideBattleFinish(ids)
			if e != nil {
				t.Fatal(e)
			}
			if d.Ended != (row.AX == 1) {
				t.Fatal("勝負返回值不同", d, row.AX)
			}
			seed := uint32(0x12345678)
			if d.Winner >= 0 {
				c := finishCursor(t, s, list, d.Winner, row.Round)
				for steps := 0; !c.Done(); steps++ {
					if steps > 10 {
						t.Fatal("續接未結束")
					}
					c.Step()
				}
				seed = c.rnd.Seed()
			}
			compareFinishTables(t, s, list, row.After)
			b, e := hex.DecodeString(row.SeedAfterHex)
			if e != nil {
				t.Fatal(e)
			}
			if seed != binary.LittleEndian.Uint32(b) {
				t.Errorf("seed %08X/%X", seed, b)
			}
		})
	}
}

func TestWinnerOccupationCaptureWaitMatchesOriginal(t *testing.T) {
	rows := finishOracle(t, "finish79-r91-cursor-a.json", "7ae71c1bea82bf045d5083b9326e41c1defdde899355164d1df2d699a82978f8")
	checked := 0
	for _, row := range rows {
		s, list := finishSlot(t, row)
		winner := 0
		if row.Mask == 1 {
			winner = 1
		}
		c := finishCursor(t, s, list, winner, row.Round)
		for !c.Done() {
			event := c.Step()
			if event == nil {
				continue
			}
			for _, call := range row.Calls {
				if call.Location != "185C:0B18" {
					continue
				}
				if event.Fleet != (call.Args[0]-0xc074)/FleetRecord {
					t.Fatal("攻占艦隊不同")
				}
				compareFinishTables(t, s, list, call.Tables)
				seed, _ := hex.DecodeString(call.SeedHex)
				if c.rnd.Seed() != binary.LittleEndian.Uint32(seed) {
					t.Fatal("等待點seed不同")
				}
				checked++
			}
		}
		compareFinishTables(t, s, list, row.After)
		if c.Step() != nil || !c.Done() {
			t.Fatal("已完成狀態重新開始")
		}
		compareFinishTables(t, s, list, row.After)
	}
	if checked != 1 {
		t.Fatal("未核實唯一原版報告", checked)
	}
}

func TestWinnerOccupationRejectsBeforeMutation(t *testing.T) {
	rows := finishOracle(t, "finish79-r89-a.json", "d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477")
	var row finishOracleRow
	for _, r := range rows {
		if r.Mask == 1 && r.Round == 5 {
			row = r
			break
		}
	}
	kinds, e := LoadShipKindTable("../../docs/internals/data/ship-kind-table.json")
	if e != nil {
		t.Fatal(e)
	}
	slots, e := LoadFormationSlots("../../docs/internals/data/battle-formation-slots.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{"pool", "full", "admiral", "formation", "faction", "terrain"} {
		t.Run(bad, func(t *testing.T) {
			s, list := finishSlot(t, row)
			pool := int(binary.LittleEndian.Uint16(s.fleet(16)[16:]))
			pools := []int{pool}
			pass := func(int, int) bool { return true }
			sl := *slots
			switch bad {
			case "pool":
				pools = []int{ShipCount}
			case "full":
				binary.LittleEndian.PutUint16(s.ShipAt(pool)[10:], 0)
			case "admiral":
				s.fleet(16)[4] = 82
			case "formation":
				sl.Slots = nil
			case "faction":
				s.fleet(16)[1] = 2
			case "terrain":
				pass = nil
			}
			before := map[string]string{}
			for k, v := range finishTables(s, list) {
				before[k] = hex.EncodeToString(v)
			}
			r := NewRand(0x12345678)
			if _, err := NewWinnerOccupation(s, int(s.Work[3]), 1, 5, pools, kinds, &sl, pass, r); err == nil {
				t.Fatal("無效輸入未拒絕")
			}
			compareFinishTables(t, s, list, before)
			if r.Seed() != 0x12345678 {
				t.Fatal("拒絕消費亂數")
			}
		})
	}
}
