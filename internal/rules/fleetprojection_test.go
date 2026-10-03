package rules

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// 期望直接讀原版雙跑凍結收據，包含同座標的梳狀排序結果。
func TestFleetProjectionNative(t *testing.T) {
	data, err := os.ReadFile("../../docs/internals/data/fleet-projection.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "cd191199bf5297fd25d41891dc89fe4108709c6426bde78fa44ce95aa8f471f0" {
		t.Fatal("原版投影收據雜湊不符")
	}
	type record struct {
		ID         int      `json:"id"`
		Raw        string   `json:"raw_hex"`
		Projection []uint16 `json:"projection_words"`
	}
	var receipt struct {
		Native struct {
			Cases []struct {
				Name          string `json:"name"`
				Before, After []record
				Fleets        []struct {
					IDs []int `json:"unit_ids"`
				}
				Checkpoints []struct {
					Array  []int
					Camera []uint16 `json:"camera_words"`
				}
			}
		}
	}
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Native.Cases) != 30 {
		t.Fatal("原版案例數不符")
	}
	for _, c := range receipt.Native.Cases {
		t.Run(c.Name, func(t *testing.T) {
			byID := make(map[int]record)
			for _, r := range c.Before {
				byID[r.ID] = r
			}
			var sides [2][]FleetProjectionUnit
			for side, f := range c.Fleets {
				for _, id := range f.IDs {
					raw, err := hex.DecodeString(byID[id].Raw)
					if err != nil || len(raw) != 78 {
						t.Fatal("原版SHIP形狀不符")
					}
					sides[side] = append(sides[side], FleetProjectionUnit{id, raw[0], binary.LittleEndian.Uint16(raw[0x34:]), binary.LittleEndian.Uint16(raw[0x36:])})
				}
			}
			before, _ := json.Marshal(sides)
			got, err := ProjectFleetScene(sides)
			if err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(sides)
			if string(before) != string(after) {
				t.Fatal("來源被修改")
			}
			last := c.Checkpoints[len(c.Checkpoints)-1]
			camera := []uint16{uint16(got.Angle), uint16(len(got.Units)), uint16(got.References[0].X), uint16(got.References[0].Y), uint16(got.References[1].X), uint16(got.References[1].Y), uint16(got.Anchors[0].X), uint16(got.Anchors[0].Y), uint16(got.Anchors[1].X), uint16(got.Anchors[1].Y)}
			if !reflect.DeepEqual(camera, last.Camera) {
				t.Fatalf("camera %v want %v", camera, last.Camera)
			}
			for i, u := range got.Units {
				visible := uint16(0)
				if u.Visible {
					visible = 1
				}
				words := []uint16{uint16(u.X), uint16(u.Y), visible, uint16(u.Depth)}
				if u.ID != c.After[i].ID || !reflect.DeepEqual(words, c.After[i].Projection) {
					t.Fatalf("pool%d %v want %v", u.ID, words, c.After[i].Projection)
				}
			}
			order := make([]int, len(got.Order))
			for i, idx := range got.Order {
				order[i] = got.Units[idx].ID
			}
			if !reflect.DeepEqual(order, last.Array) {
				t.Fatalf("order %v want %v", order, last.Array)
			}
		})
	}
}

func TestFleetProjectionRejectsInvalidSource(t *testing.T) {
	for name, sides := range map[string][2][]FleetProjectionUnit{
		"empty":       {},
		"duplicate":   {{{ID: 1}}, {{ID: 1, Side: 1}}},
		"wrong-side":  {{{ID: 1, Side: 1}}, {{ID: 2, Side: 1}}},
		"coordinate":  {{{ID: 1, X: 2048}}, {{ID: 2, Side: 1}}},
		"negative-id": {{{ID: -1}}, {{ID: 2, Side: 1}}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ProjectFleetScene(sides); err == nil {
				t.Fatal("錯誤來源被接受")
			}
		})
	}
	var sides [2][]FleetProjectionUnit
	for side := range sides {
		for i := 0; i < 12; i++ {
			sides[side] = append(sides[side], FleetProjectionUnit{ID: side*12 + i, Side: uint8(side)})
		}
	}
	if _, err := ProjectFleetScene(sides); err == nil {
		t.Fatal("超容量來源被接受")
	}
}
