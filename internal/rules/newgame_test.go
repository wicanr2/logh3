package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestNewGameRegions 是新遊戲載入的驗收（`docs/re/44`、`docs/spec/006`）：
// 從 `STELLAR.DTX` ＋ `PLADAT.DAT` ＋ `ADMIRAL.DTX` 重建 796 年 1 月的
// 星域表，每個星域的擁有者、行星數、經濟力、成長率、防衛力以外的欄位
// 都要與量到的相同。
func TestNewGameRegions(t *testing.T) {
	root := "../.."
	stellar := readOrig(t, "STELLAR.DTX")
	planets, err := LoadPlanets(readOrig(t, "PLADAT.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	admirals := readOrig(t, "ADMIRAL.DTX")
	table, err := LoadGrowthTable(filepath.Join(root,
		"docs/internals/data/growth-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	graph, err := LoadStarGraph(filepath.Join(root,
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root,
		"docs/internals/data/region-table-796-01.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Regions []struct {
			Index   int   `json:"index"`
			Owner   int   `json:"owner"`
			Planets []int `json:"planets"`
			Economy int   `json:"economy"`
			Growth  int   `json:"growth_per_mille"`
			Tax     int   `json:"tax_rate_per_mille"`
			Morale  int   `json:"morale"`
			Unrest  int   `json:"unrest"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	// 難度參數量自同一輪（收據 `newload`）。
	params := UnrestParams{
		Turn: 0, Ruler: [2]int{16, 60}, Blend: [2]int{2, 2}, Rate: [2]int{2, 2},
		Table: []int{170, 150, 130, 90, 60},
		Base:  []int{50, 75, 100, 150, 200},
	}
	got := NewGameRegions(stellar, planets, admirals, table, graph, params)
	for _, w := range d.Regions {
		r := got[w.Index*RegionRecord : (w.Index+1)*RegionRecord]
		if int(r[0x12]) != w.Owner {
			t.Errorf("星域 %d 擁有者 %d，要 %d", w.Index, r[0x12], w.Owner)
		}
		for k, n := range w.Planets {
			if int(r[0x16+k]) != n {
				t.Errorf("星域 %d 行星數[%d] %d，要 %d", w.Index, k, r[0x16+k], n)
			}
		}
		if e := RegionEconomy(r); e != w.Economy {
			t.Errorf("星域 %d 經濟力 %d，要 %d", w.Index, e, w.Economy)
		}
		for _, c := range []struct {
			name string
			off  int
			want int
		}{
			{"成長率", 0x1E, w.Growth}, {"稅率", 0x26, w.Tax},
			{"民心", 0x7C, w.Morale}, {"未安定度", 0x7D, w.Unrest},
		} {
			if int(r[c.off]) != c.want {
				t.Errorf("星域 %d %s %d，要 %d", w.Index, c.name, r[c.off], c.want)
			}
		}
	}
}
