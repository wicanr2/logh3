package game

import (
	"fmt"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
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
