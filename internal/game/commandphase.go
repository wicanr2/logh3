package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// LoadCommandScreens 讀命令階段那一批畫面的規格，順便把它們要的星域地圖
// 各建一份。
//
// 為什麼要建好幾份：情報報告的地圖跟著**治安**換色、星系資料又是另一組
// （選中的星域畫成色 7、首都不重塗），共用一張會整片對不上。
// 哪一張用哪一組寫在規格的 `map.colours`。
func LoadCommandScreens(dataDir, regionsPath string, bit, rgn []byte) (
	map[string]*screen.MainGameSpec, map[string]*asset.Image, error) {
	specs := make(map[string]*screen.MainGameSpec, len(screenSpecName))
	for _, name := range screenSpecName {
		if specs[name] != nil {
			continue
		}
		s, err := screen.LoadMainGameSpec(filepath.Join(dataDir, name+".json"))
		if err != nil {
			return nil, nil, fmt.Errorf("讀 %s：%w", name, err)
		}
		specs[name] = s
	}
	maps := map[string]*asset.Image{}
	for _, s := range specs {
		f := s.Map.Colours
		if f == "" || maps[f] != nil {
			continue
		}
		cols, err := screen.LoadRegionColours(regionsPath, f)
		if err != nil {
			return nil, nil, fmt.Errorf("讀星域色 %s：%w", f, err)
		}
		img, err := asset.BuildStarMap(bit, rgn, cols, 16, 20, 20)
		if err != nil {
			return nil, nil, fmt.Errorf("建地圖 %s：%w", f, err)
		}
		maps[f] = img
	}
	return specs, maps, nil
}

// LoadRegionRecords 湊出 796 年 1 月的執行期星域記錄（一筆 126 bytes）。
//
// 名字、稅率、成長率這些直接從 `STELLAR.DTA` 讀；擁有者、行星數、經濟力
// 原版是載入時依行星重算的（`docs/re/17`、`docs/re/30`），面板要用的那幾格
// 由 `internal/rules` 現算，不動記錄。
//
// **未安定度是例外**：載入時怎麼算出來的還沒解（`docs/re/24`），所以這裡
// 用量到的那一組（`month-rollover-796-01.json`）填回去。解出算式之前，
// 換一個劇本或開局之後的局面這一欄就不對——這是已知的缺口，不是設計。
func LoadRegionRecords(stellar []byte, measuredPath string) ([]byte, error) {
	if len(stellar) < 32*rules.RegionRecord {
		return nil, fmt.Errorf("STELLAR 太短：%d bytes", len(stellar))
	}
	b, err := os.ReadFile(measuredPath)
	if err != nil {
		return nil, err
	}
	var d struct {
		Regions []struct {
			Index  int `json:"index"`
			Unrest int `json:"unrest_796_01"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	out := make([]byte, 32*rules.RegionRecord)
	copy(out, stellar)
	for _, r := range d.Regions {
		if r.Index < 0 || r.Index >= 32 {
			continue
		}
		out[r.Index*rules.RegionRecord+0x7D] = byte(r.Unrest)
	}
	// 艦隊數同樣清掉。796 年 1 月每一份收據上都是 `0：0`，而檔案裡不是
	// 全零（`STELLAR.DTA` 星域 14 的 `+0x28` 是 1）——那一格在載入時
	// 被蓋掉了，檔案裡的值是別的意思。開局沒有艦隊，先當 0。
	for i := 0; i < 32; i++ {
		o := i*rules.RegionRecord + 0x28
		out[o], out[o+1], out[o+2], out[o+3] = 0, 0, 0, 0
	}
	return out, nil
}
