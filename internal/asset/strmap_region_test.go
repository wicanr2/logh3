package asset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSTRMAPRegionsConsumeWholeFile(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(origDir(t), "STRMAP.RGN"))
	if err != nil {
		t.Skip("沒有原版 STRMAP.RGN，跳過")
	}
	regions, err := DecodeSTRMAPRegions(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(regions) != StarRegionCount {
		t.Fatalf("解出 %d 個星域，預期 %d", len(regions), StarRegionCount)
	}
	// 32 個星域鋪滿星圖的大部分，彼此不重疊——重疊表示遮罩解錯了。
	owner := make([]int8, STRMAPWidth*STRMAPHeight)
	for i := range owner {
		owner[i] = -1
	}
	covered, overlap := 0, 0
	for i := range regions {
		for y := 0; y < STRMAPHeight; y++ {
			for x := 0; x < STRMAPWidth; x++ {
				if !regions[i].Contains(x, y) {
					continue
				}
				if owner[y*STRMAPWidth+x] >= 0 {
					overlap++
				} else {
					covered++
				}
				owner[y*STRMAPWidth+x] = int8(i)
			}
		}
	}
	total := STRMAPWidth * STRMAPHeight
	if covered*100 < total*80 {
		t.Errorf("只覆蓋 %d / %d（%.1f%%），預期至少 80%%",
			covered, total, float64(covered)*100/float64(total))
	}
	if overlap*100 > total*5 {
		t.Errorf("重疊 %d 像素，太多", overlap)
	}
	t.Logf("覆蓋 %d / %d（%.1f%%），重疊 %d",
		covered, total, float64(covered)*100/float64(total), overlap)
}
