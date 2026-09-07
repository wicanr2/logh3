package screen

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadRegionColours 讀 `strmap-regions.json` 裡某一組星域色。
//
// 那份檔案的每個星域帶好幾組顏色（勢力色、治安色、星系資料那一組…），
// 畫面規格的 `map.colours` 指定要哪一組；空字串就是勢力色。
// **不能共用同一張地圖**——同一份底圖配不同的色表是不同的畫面。
func LoadRegionColours(path, field string) ([]uint8, error) {
	if field == "" {
		field = "colour_796_01"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Regions []map[string]json.RawMessage `json:"regions"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make([]uint8, len(doc.Regions))
	for i, r := range doc.Regions {
		raw, ok := r[field]
		if !ok {
			return nil, fmt.Errorf("第 %d 個星域沒有 %q 欄位", i, field)
		}
		if err := json.Unmarshal(raw, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
