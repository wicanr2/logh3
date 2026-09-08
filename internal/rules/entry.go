package rules

import (
	"encoding/json"
	"fmt"
	"os"
)

// BattleEntry 是戰場上的一個入口：從 From 那個星域打進來的艦隊，
// 開場放在格 (Col,Row)，也就是座標 (X,Y)。
//
// From 等於星域自己的編號時，那是**本來就在這裡**的守方（`docs/internals/39`）。
type BattleEntry struct {
	From int `json:"from"`
	Col  int `json:"col"`
	Row  int `json:"row"`
	X    int `json:"x"`
	Y    int `json:"y"`
}

// BattleEntries 是 32 個星域各一組入口。
type BattleEntries struct {
	TileW   int `json:"tile_w"`
	TileH   int `json:"tile_h"`
	Regions []struct {
		Region  int           `json:"region"`
		Entries []BattleEntry `json:"entries"`
	} `json:"regions"`
}

// LoadBattleEntries 讀 `docs/internals/data/battle-entry-points.json`。
func LoadBattleEntries(path string) (*BattleEntries, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var e BattleEntries
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, err
	}
	if len(e.Regions) != 32 {
		return nil, fmt.Errorf("入口表有 %d 個星域，預期 32", len(e.Regions))
	}
	return &e, nil
}

// EntryFor 回「在 region 打，從 from 進來」的入口；找不到回 false。
//
// 原版的查法是照表的順序線性找第一個鍵相同的（`113A:0D2C`），
// 這裡照做——同一個鍵出現兩次時要跟原版挑同一筆。
func (e *BattleEntries) EntryFor(region, from int) (BattleEntry, bool) {
	if e == nil || region < 0 || region >= len(e.Regions) {
		return BattleEntry{}, false
	}
	for _, x := range e.Regions[region].Entries {
		if x.From == from {
			return x, true
		}
	}
	return BattleEntry{}, false
}
