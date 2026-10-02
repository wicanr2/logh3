package game

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 收據由 tools/region_supply_receipts.sh 產生；缺原版就跳過，不造替代資料。
func supplyTimingMemory(t *testing.T, stage string) []byte {
	t.Helper()
	m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/supply84-stationed", stage+".bin"))
	if err != nil {
		t.Skipf("缺原版 %s 收據：%v", stage, err)
	}
	if len(m) != 1048576 {
		t.Fatalf("%s 記憶體長度 %d", stage, len(m))
	}
	return m
}

// 使用正式解碼器；艦艇池基底從原版遠指標取得，不能硬編。
func supplyTimingGame(t *testing.T, m []byte) *Game {
	t.Helper()
	res := loadWithRules(t)
	res.Stellar = append([]byte(nil), m[0x5F2CC:0x5F2CC+4032]...)
	copy(res.Planets.Data[rules.PlanetHeader:], m[0x53132:0x53132+6642])
	ptr := m[0x61AE4:0x61AE8]
	base := int(binary.LittleEndian.Uint16(ptr[2:]))*16 + int(binary.LittleEndian.Uint16(ptr))
	if base+rules.ShipCount*rules.ShipRecord > len(m) {
		t.Fatal("原版艦艇池指標越界")
	}
	s := &rules.Slot{Fleet: m[0x60BA4 : 0x60BA4+rules.FleetCount*rules.FleetRecord],
		Ship: m[base : base+rules.ShipCount*rules.ShipRecord]}
	return &Game{res: res, fleets: s.Fleets(nil)}
}

func supplyTimingEqual(t *testing.T, name string, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	n, first := 0, -1
	for i := range got {
		if i >= len(want) || got[i] != want[i] {
			n++
			if first < 0 {
				first = i
			}
		}
	}
	t.Fatalf("%s 不同 %d bytes，第一處 +0x%X（長度 %d/%d）", name, n, first, len(got), len(want))
}

func TestRegionSupplyStagesMatchOriginal(t *testing.T) {
	t.Run("移動後分配整張行星表", func(t *testing.T) {
		pre, post := supplyTimingMemory(t, "move-finished"), supplyTimingMemory(t, "distributed")
		g := supplyTimingGame(t, pre)
		for r := 0; r < rules.RegionCount; r++ {
			for f := 0; f < 2; f++ {
				rules.DistributeRegionSupply(g.regionRecord(r), g.res.Planets, r, f)
			}
		}
		supplyTimingEqual(t, "分配後全 81 顆行星", g.res.Planets.Data[rules.PlanetHeader:], post[0x53132:0x53132+6642])
	})
	t.Run("艦隊補給與重新彙總", func(t *testing.T) {
		g := supplyTimingGame(t, supplyTimingMemory(t, "distributed"))
		ptrs := make([]*rules.Fleet, len(g.fleets))
		for i := range g.fleets {
			ptrs[i] = &g.fleets[i]
		}
		rules.SupplyRegionFleets(g.regionRecord(0), 0, ptrs, g.res.ShipKinds)
		m := supplyTimingMemory(t, "fleet-supplied")
		supplyTimingEqual(t, "艦隊補給後星域 0", g.regionRecord(0), m[0x5F2CC:0x5F2CC+126])
		want := supplyTimingGame(t, m)
		if !reflect.DeepEqual(g.fleets[0].Composition, want.fleets[0].Composition) {
			t.Fatal("飛彈／艦載機補給後編制與原版不同")
		}
		rules.ResupplyRegion(g.regionRecord(0), g.res.Planets, 0, ptrs, g.res.ShipKinds)
		m = supplyTimingMemory(t, "resupplied")
		supplyTimingEqual(t, "艦艇補充後星域 0", g.regionRecord(0), m[0x5F2CC:0x5F2CC+126])
		supplyTimingEqual(t, "艦艇補充後全 81 顆行星", g.res.Planets.Data[rules.PlanetHeader:], m[0x53132:0x53132+6642])
		want = supplyTimingGame(t, m)
		// 後續士氣報告會另改 Morale；只比補給所寫的完整數量欄位。
		for i, got := range g.fleets[0].Composition {
			w := want.fleets[0].Composition[i]
			if got.Ships != w.Ships || got.RepairCap != w.RepairCap || got.Missiles != w.Missiles || got.Fighters != w.Fighters {
				t.Fatalf("組 %d 補充 %+v，原版 %+v", i, got, w)
			}
		}
	})
	t.Run("沒有艦隊仍彙總分配", func(t *testing.T) {
		g := supplyTimingGame(t, supplyTimingMemory(t, "empty-supplied"))
		for _, f := range g.fleets {
			if f.Region == 1 {
				t.Fatal("原版星域 1 樣本應沒有艦隊")
			}
		}
		// empty-filled 停在分配入口；由原版行星彙總可驗整筆星域，
		// 後續全表的 owner-pass 證明無補充時也完成分配。
		rules.ResupplyRegion(g.regionRecord(1), g.res.Planets, 1, nil, g.res.ShipKinds)
		m := supplyTimingMemory(t, "empty-filled")
		a := 0x5F2CC + rules.RegionRecord
		supplyTimingEqual(t, "無艦隊星域 1", g.regionRecord(1), m[a:a+126])
		start, n := g.res.Planets.Range(1)
		post := supplyTimingMemory(t, "owner-pass")
		for i := start; i < start+n; i++ {
			a := 0x53132 + i*rules.PlanetRecord
			supplyTimingEqual(t, "無艦隊行星", g.res.Planets.Record(i), post[a:a+82])
		}
	})
	t.Run("下一座位只回充星域整表", func(t *testing.T) {
		g := supplyTimingGame(t, supplyTimingMemory(t, "refill-before"))
		g.refillRegionSupply()
		m := supplyTimingMemory(t, "refill-after")
		supplyTimingEqual(t, "回充後全 32 格星域", g.res.Stellar, m[0x5F2CC:0x5F2CC+4032])
		supplyTimingEqual(t, "回充不寫全 81 顆行星", g.res.Planets.Data[rules.PlanetHeader:], m[0x53132:0x53132+6642])
	})
}

// 正常入口與編成／實行；注入與原版完全相同，只比較這次已驗的局面。
func TestRegionSupplyNormalTurnMatchesOriginal(t *testing.T) {
	m := supplyTimingMemory(t, "command")
	g := reachFleetDecided(t)
	initial := supplyTimingMemory(t, "initial")
	seed := binary.LittleEndian.Uint32(initial[0x54B30+0x94DC:])
	if seed != strategyRandSeed || g.rand94DC().Seed() != seed {
		t.Fatalf("執行前種子 remake=%08X 原版=%08X", g.rand94DC().Seed(), seed)
	}
	f := &g.fleets[0]
	f.Flags |= 1
	f.Morale = 40
	f.Composition[0].Morale = 40
	f.Composition[0].Ships = 100
	f.Composition[0].Missiles, f.Composition[0].Fighters = 0, 0
	executeTurn(t, g)
	wantState(t, g, StateCommandPhase, "補給回合完成")
	// #84 比較範圍事先限定為全部星域／行星的七種艦艇與兩種補給，
	// 每個陣營的全部欄位都嚴格比較，不能用總量掩蓋分配差異。
	for r := 0; r < rules.RegionCount; r++ {
		a := 0x5F2CC + r*rules.RegionRecord
		supplyTimingEqual(t, "正常回合星域全部庫存", g.regionRecord(r)[0x30:0x78], m[a+0x30:a+0x78])
	}
	for i := 0; i < 81; i++ {
		a := 0x53132 + i*rules.PlanetRecord
		supplyTimingEqual(t, "正常回合行星全部庫存", g.res.Planets.Record(i)[0x20:0x44], m[a+0x20:a+0x44])
	}
	// 保存完整表及全部差異，#85 接續追查，不能宣稱全表一致。
	diff := map[string][]int{}
	for name, tc := range map[string]struct{ got, want []byte }{
		"stellar": {g.res.Stellar, m[0x5F2CC : 0x5F2CC+4032]},
		"planet":  {g.res.Planets.Data[rules.PlanetHeader:], m[0x53132 : 0x53132+6642]},
	} {
		for i, b := range tc.got {
			if b != tc.want[i] {
				diff[name] = append(diff[name], i)
			}
		}
		t.Logf("完整 %s 表差異 %d bytes；未達全表一致，追查見 #85", name, len(diff[name]))
		path := filepath.Join(root, "workplace/oracle/supply84-stationed", "remake-"+name+".bin")
		if err := os.WriteFile(path, tc.got, 0644); err != nil {
			t.Fatal(err)
		}
	}
	b, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workplace/oracle/supply84-stationed", "normal-table-differences.json"), append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	matches(t, g, "supply84-stationed")
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		want []byte
	}{
		{"STELLAR.DT0", g.res.Stellar}, {"PLANET.DT0", g.res.Planets.Data[rules.PlanetHeader:]},
	} {
		got, err := os.ReadFile(filepath.Join(dir, tc.name))
		if err != nil {
			t.Fatal(err)
		}
		supplyTimingEqual(t, "存檔 "+tc.name, got, tc.want)
	}
	planets := append([]byte(nil), g.res.Planets.Data...)
	stellar := append([]byte(nil), g.res.Stellar...)
	g.res.SaveDir = dir
	g.Click(544, 125)
	g.Click(560, 145)
	wantState(t, g, StateLoadGame, "正常取進度入口")
	g.Click(300, 122)
	wantState(t, g, StateLoadFaction, "存檔列後選陣營")
	g.Click(318, 208)
	wantState(t, g, StateCommandPhase, "讀回補給後存檔")
	supplyTimingEqual(t, "讀檔全部行星", g.res.Planets.Data, planets)
	for r := 0; r < rules.RegionCount; r++ {
		a := r * rules.RegionRecord
		supplyTimingEqual(t, "讀檔全部星域庫存", g.regionRecord(r)[0x30:0x78], stellar[a+0x30:a+0x78])
	}
}
