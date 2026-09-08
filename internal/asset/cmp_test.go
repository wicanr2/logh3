package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// 32 張戰場地圖每一張都要剛好展開成 128×102 格——解碼規則錯一點點
// 就會短少，而且不會報錯，只會畫出偏掉的地形。
func TestDecodeCMPAllMaps(t *testing.T) {
	dir := origDir(t)
	n := 0
	for i := 0; i < 32; i++ {
		p := filepath.Join(dir, mapName(i))
		d, err := os.ReadFile(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		cells, err := DecodeCMP(d, BattleMapCols, BattleMapRows)
		if err != nil {
			t.Errorf("%s：%v", mapName(i), err)
			continue
		}
		for _, c := range cells {
			if c >= 128 {
				t.Fatalf("%s 出現圖塊編號 %d", mapName(i), c)
			}
		}
		n++
	}
	if n == 0 {
		t.Skip("找不到原版的 MAP*.CMP")
	}
	if n != 32 {
		t.Errorf("只讀到 %d 張地圖，原版有 32 張", n)
	}
}

func mapName(i int) string {
	return "MAP" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ".CMP"
}
