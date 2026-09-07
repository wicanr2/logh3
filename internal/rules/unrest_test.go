package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestUnrestAtLoad 是未安定度算式的驗收（`docs/spec/006`）：
// 拿 S3 那一格存檔載入，32 個星域算出來的未安定度要與量到的完全相同。
func TestUnrestAtLoad(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("../..",
		"docs/internals/data/unrest-at-load.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Slots []struct {
			Slot    string `json:"slot"`
			Params  struct {
				Turn  int    `json:"turn_word_8F18"`
				Ruler [2]int `json:"ruler_8F4F"`
				Blend [2]int `json:"blend_8FD4"`
				Rate  [2]int `json:"rate_8FF8"`
				Table []int  `json:"table_4B28"`
				Base  []int  `json:"base_147E"`
			} `json:"params"`
			Regions []struct {
				Index  int `json:"index"`
				AtLoad int `json:"at_load"`
			} `json:"regions"`
		} `json:"slots"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	graph, err := LoadStarGraph(filepath.Join("../..",
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, sl := range d.Slots {
		if len(sl.Params.Table) == 0 {
			continue // 那一格的執行期參數還沒量
		}
		t.Run(sl.Slot, func(t *testing.T) {
			slot := loadSlot(t, sl.Slot)
			st := &UnrestState{
				Regions:  slot.Stellar,
				Planets:  slot.Planets,
				Fleets:   slot.Fleet,
				Admirals: slot.Admiral,
				Graph:    graph,
				Params: UnrestParams{
					Turn: sl.Params.Turn, Ruler: sl.Params.Ruler,
					Blend: sl.Params.Blend, Rate: sl.Params.Rate,
					Table: sl.Params.Table, Base: sl.Params.Base,
				},
			}
			for _, r := range sl.Regions {
				if got := st.Unrest(r.Index); got != r.AtLoad {
					t.Errorf("星域 %d：算出來 %d，量到的是 %d",
						r.Index, got, r.AtLoad)
				}
			}
		})
	}
}
