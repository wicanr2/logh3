package rules

import (
	"encoding/hex"
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
			Slot   string `json:"slot"`
			Params struct {
				Turn  int    `json:"turn_word_8F18"`
				Ruler [2]int `json:"ruler_8F4F"`
				Blend [2]int `json:"blend_8FD4"`
				Rate  [2]int `json:"rate_8FF8"`
				Table []int  `json:"table_4B28"`
				Base  []int  `json:"base_147E"`
			} `json:"params"`
			Regions []struct {
				Index  int    `json:"index"`
				AtLoad int    `json:"at_load"`
				Raw    string `json:"raw_at_load"`
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
			var regions, fleets, admirals []byte
			var planets *Planets
			if sl.Slot == "X" {
				// 新遊戲不是存檔：星域表載入時重算過，拿量到的那一份當輸入；
				// 行星讀 `PLADAT.DAT`、提督讀樣板，這個局面還沒有艦隊。
				for _, r := range sl.Regions {
					raw, err := hex.DecodeString(r.Raw)
					if err != nil {
						t.Fatal(err)
					}
					regions = append(regions, raw...)
				}
				planets, err = LoadPlanets(readOrig(t, "PLADAT.DAT"))
				if err != nil {
					t.Fatal(err)
				}
				admirals = readOrig(t, "ADMIRAL.DTX")
			} else {
				slot := loadSlot(t, sl.Slot)
				regions, planets = slot.Stellar, slot.Planets
				fleets, admirals = slot.Fleet, slot.Admiral
			}
			st := &UnrestState{
				Regions:  regions,
				Planets:  planets,
				Fleets:   fleets,
				Admirals: admirals,
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

// TestUnrestSurvivesMissingAdmirals 釘住「資料不全不要 panic」。
//
// `admiral()` 查不到的時候回 nil，而 `Unrest` 原本直接對它取索引。
// 戰鬥中選「取進度」會走到這條路，玩家碰得到。
func TestUnrestSurvivesMissingAdmirals(t *testing.T) {
	s := &UnrestState{
		Regions: make([]byte, RegionRecord),
		Params: UnrestParams{
			Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
			Table: []int{170, 150, 130, 90, 60},
			Base:  []int{50, 75, 100, 150, 200},
		},
	}
	if got := s.Unrest(0); got != 0 {
		t.Errorf("缺提督表時回 %d，應該是 0", got)
	}
}
