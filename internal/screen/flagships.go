package screen

import (
	"encoding/json"
	"os"
)

// Flagships 是兩個陣營各 16 艘旗艦的名字字碼
// （`docs/internals/data/flagships.json`，`docs/re/65`）。
type Flagships struct {
	Tables [][]*struct {
		Slot  int      `json:"slot"`
		Text  string   `json:"text"`
		Codes []uint16 `json:"codes"`
	} `json:"tables"`
}

// LoadFlagships 讀旗艦表。
func LoadFlagships(path string) (*Flagships, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f Flagships
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// Name 回一個陣營第 slot 艘旗艦的字碼；問不到回 nil。
func (f *Flagships) Name(faction, slot int) []uint16 {
	if f == nil || faction < 0 || faction >= len(f.Tables) {
		return nil
	}
	t := f.Tables[faction]
	if slot < 0 || slot >= len(t) || t[slot] == nil {
		return nil
	}
	return t[slot].Codes
}
