package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRegionAtMatchesOracle 釘住「地圖上這一點是哪個星域」。
//
// 每一筆都是實測：在原版的戰略地圖上按右鍵，左下角那塊會寫出星系名
// （收據 main1e／hovA／hovB／hovC／rcback／star1e，掃字見
// `docs/re/22`）。座標是**滑鼠座標**，而星圖畫在畫面的 (0,40)、
// 畫面 y ＝ 滑鼠 y ＋ 40，所以滑鼠座標直接就是地圖內座標。
func TestRegionAtMatchesOracle(t *testing.T) {
	p := filepath.Join("../..", "org_game/extracted/STRMAP.RGN")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺 %s，跳過（不用代用品）", p)
	}
	regions, err := OpenRGN(data, 16, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		x, y  int
		want  int
		name  string
		batch string
	}{
		{60, 60, 2, "亞爾提那", "hovA"},
		{150, 120, 14, "費沙", "hovB"},
		{250, 200, 19, "特利布拉", "hovC"},
		{300, 200, 30, "艾流塞拉", "rcback"},
	} {
		if got := RegionAt(regions, c.x, c.y, 16, 20); got != c.want {
			t.Errorf("(%d,%d) → 星域 %d，預期 %d（%s，收據 %s）",
				c.x, c.y, got, c.want, c.name, c.batch)
		}
	}
}
