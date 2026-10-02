package game

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func runtimeFieldsMemory(t *testing.T, batch, stage string) []byte {
	t.Helper()
	m, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, stage+".bin"))
	if err != nil {
		t.Skipf("缺 #85 原版收據：%v", err)
	}
	if len(m) != 1048576 {
		t.Fatal("原版全記憶體收據長度錯誤")
	}
	return m
}

func TestRuntimeFieldsNormalTurnMatchesOriginal(t *testing.T) {
	for _, batch := range []string{"runtime85-normal", "runtime85-stationed"} {
		t.Run(batch, func(t *testing.T) {
			m := runtimeFieldsMemory(t, batch, "command")
			initial := runtimeFieldsMemory(t, batch, "initial")
			g := reachFleetDecided(t)
			seed := binary.LittleEndian.Uint32(initial[0x5E00C:])
			if seed != strategyRandSeed || g.rand94DC().Seed() != seed {
				t.Fatal("執行前兩側固定種子不相同")
			}
			if batch == "runtime85-stationed" {
				f := &g.fleets[0]
				f.Flags |= 1
				f.Morale, f.Composition[0].Morale, f.Composition[0].Ships = 40, 40, 100
				f.Composition[0].Missiles, f.Composition[0].Fighters = 0, 0
			}
			executeTurn(t, g)
			wantState(t, g, StateCommandPhase, "正常完整回合")
			supplyTimingEqual(t, "完整星域 4032 bytes", g.res.Stellar, m[0x5F2CC:0x5F2CC+4032])
			supplyTimingEqual(t, "完整行星 6642 bytes", g.res.Planets.Data[rules.PlanetHeader:], m[0x53132:0x53132+6642])
			matches(t, g, batch)
			for name, raw := range map[string][]byte{"stellar": g.res.Stellar, "planet": g.res.Planets.Data[rules.PlanetHeader:]} {
				if err := os.WriteFile(filepath.Join(root, "workplace/oracle", batch, "remake-"+name+".bin"), raw, 0644); err != nil {
					t.Fatal(err)
				}
			}
			meta, err := json.MarshalIndent(map[string]any{"seed": seed, "seed_setup": "執行前固定 strategyRandSeed",
				"entry": "reachFleetDecided → 正常系統實行 → 命令階段", "stellar_bytes": 4032, "planet_bytes": 6642,
				"different": 0, "injected": batch == "runtime85-stationed"}, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "workplace/oracle", batch, "remake-tables.json"), meta, 0644); err != nil {
				t.Fatal(err)
			}
			stellar, planets := bytes.Clone(g.res.Stellar), bytes.Clone(g.res.Planets.Data)
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			for _, tc := range []struct {
				name string
				want []byte
			}{{"STELLAR.DT0", stellar}, {"PLANET.DT0", planets[rules.PlanetHeader:]}} {
				raw, err := os.ReadFile(filepath.Join(dir, tc.name))
				if err != nil {
					t.Fatal(err)
				}
				supplyTimingEqual(t, "正常存檔全表 "+tc.name, raw, tc.want)
			}
			g.res.SaveDir = dir
			g.Click(544, 125)
			g.Click(560, 145)
			wantState(t, g, StateLoadGame, "正常取進度")
			g.Click(300, 122)
			wantState(t, g, StateLoadFaction, "存檔列後選陣營")
			g.Click(318, 208)
			wantState(t, g, StateCommandPhase, "取進度完成")
			// 載入會依檔內艦隊重數計數、重算未安定度（研究 40／規格 006）。
			// #86 另追現況艦隊未同步；此處只聲明本輪新增距離的讀回保存。
			for r := 0; r < rules.RegionCount; r++ {
				o := r*rules.RegionRecord + 0x78
				supplyTimingEqual(t, "取進度 AI 距離", g.res.Stellar[o:o+2], stellar[o:o+2])
			}
			supplyTimingEqual(t, "取進度全部行星", g.res.Planets.Data, planets)
		})
	}
}

func TestPlanetCoordinatesLoadAtStandbyOnly(t *testing.T) {
	g := reachFleetDecided(t)
	before := bytes.Clone(g.res.Planets.Data)
	g.Frame() // 畫面不應提前寫入地圖座標。
	supplyTimingEqual(t, "繪製不改行星", g.res.Planets.Data, before)
	g.stepFleet(0)
	m := runtimeFieldsMemory(t, "runtime85-normal", "map-after")
	start, n := g.res.Planets.Range(0)
	for i := 0; i < 81; i++ {
		rec := g.res.Planets.Record(i)
		if i >= start && i < start+n {
			a := 0x53132 + i*rules.PlanetRecord
			supplyTimingEqual(t, "待命時載入行星座標", rec[0x46:0x4A], m[a+0x46:a+0x4A])
		} else {
			a := rules.PlanetHeader + i*rules.PlanetRecord
			supplyTimingEqual(t, "其他星域不提前載入", rec, before[a:a+rules.PlanetRecord])
		}
	}
}
