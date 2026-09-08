package rules

import (
	"encoding/json"
	"os"
)

// 艦艇池一筆 78 bytes（`docs/re/42`）。戰力算式用到的三個欄位
// （`2E54:1876`，`docs/re/79`）：
const (
	shipKind  = 0x04 // 艦種 0–6；7 ＝ 每支艦隊多出來的那一筆；0xFF ＝ 空的
	shipAlive = 0x08 // 現有艘數（word）
	shipFull  = 0x0A // 滿編艘數（word）

	shipKindEmpty = 0xFF
)

// StrengthWeights 是每個艦種的戰力權重（原版的 `ds:6032`，八個 byte）。
type StrengthWeights []int

// LoadStrengthWeights 讀 `docs/internals/data/ship-strength-weights.json`。
func LoadStrengthWeights(path string) (StrengthWeights, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Weights []int `json:"weights"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	return doc.Weights, nil
}

// ShipStrength 回一筆艦艇記錄的戰力。
//
// 原版的算式（`2E54:1876`）：
//
//	權重[艦種] × 現有艘數 × 1000 ÷ 滿編艘數
//
// 滿編的時候就是 `權重 × 1000`，打殘了按比例往下掉。空的那幾筆
// （`+0x04` ＝ `0xFF`）不算。
func (w StrengthWeights) ShipStrength(record []byte) int {
	if len(record) < ShipRecord {
		return 0
	}
	kind := int(record[shipKind])
	if kind == shipKindEmpty || kind >= len(w) {
		return 0
	}
	full := int(record[shipFull]) | int(record[shipFull+1])<<8
	if full == 0 {
		return 0
	}
	alive := int(record[shipAlive]) | int(record[shipAlive+1])<<8
	return w[kind] * alive * 1000 / full
}

// FleetStrength 回存檔裡第 i 支艦隊的戰力總和。
//
// 十個槽指到的艦艇各算一次（`docs/re/42`）。
func (s *Slot) FleetStrength(w StrengthWeights, i int) int {
	total := 0
	for _, n := range s.FleetShips(i) {
		if n < 0 || (n+1)*ShipRecord > len(s.Ship) {
			continue
		}
		total += w.ShipStrength(s.Ship[n*ShipRecord : (n+1)*ShipRecord])
	}
	return total
}
