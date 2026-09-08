package rules

import (
	"encoding/json"
	"fmt"
	"os"
)

// 艦種屬性表（`ds:4E3A`，`docs/internals/40`）：一個陣營一筆 208 bytes，
// 內含 8 個艦種各 26 bytes ＝ 13 個 word。

// ShipKind 是一個艦種在表裡的欄位。
type ShipKind struct {
	Kind      int `json:"kind"`
	FullShips int `json:"full_ships"`
	// SemiMajor／SemiMinor 是攻擊橢圓的長短半軸（`+0x16`／`+0x18`，
	// `docs/re/94`）。
	SemiMajor int `json:"semi_major"`
	SemiMinor int `json:"semi_minor"`
}

// ShipKindTable 是整張表。
type ShipKindTable struct {
	Factions []struct {
		Faction int        `json:"faction"`
		Kinds   []ShipKind `json:"kinds"`
	} `json:"factions"`
}

// LoadShipKindTable 讀 `docs/internals/data/ship-kind-table.json`。
func LoadShipKindTable(path string) (*ShipKindTable, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t ShipKindTable
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	if len(t.Factions) == 0 {
		return nil, fmt.Errorf("%s 裡沒有陣營", path)
	}
	return &t, nil
}

// Kind 回某個陣營某個艦種的欄位。
func (t *ShipKindTable) Kind(faction, kind int) (ShipKind, bool) {
	if t == nil {
		return ShipKind{}, false
	}
	for _, f := range t.Factions {
		if f.Faction != faction {
			continue
		}
		for _, k := range f.Kinds {
			if k.Kind == kind {
				return k, true
			}
		}
	}
	return ShipKind{}, false
}

// SkirmishAttack 回單擊算式裡的攻擊係數 `攻擊方 +0x16 ÷ K`
// （`2350:0023`，`docs/re/86`、`docs/re/94`）。
//
// ⚠ **`faction`／`kind`／`ability` 是「目標」的**，不是攻擊方的：K 是
// 目標的方向性防禦（橢圓），只有 `attack`（艦艇記錄的 `+0x16`）取自
// 攻擊方。`dir` 是方向索引。表裡查不到那個艦種就回 0。
func (t *ShipKindTable) SkirmishAttack(faction, kind, dir, ability, attack int) int {
	k, ok := t.Kind(faction, kind)
	if !ok {
		return 0
	}
	den := AttackCoefficient(k.SemiMajor, k.SemiMinor, dir, ability)
	if den <= 0 {
		return 0
	}
	return attack / den
}
