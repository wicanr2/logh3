package rules

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// 收據 `s3cmdrange`：第 28 艦隊（回轉 15、巡航 30 → 距離 240、朝向 190）
// 停在選點狀態。外形是**色 1 XOR 底色 3 ＝ 色 2**（`docs/re/135`），
// 圓心是旗艦圖塊中心 (298,409)。
//
// 比法用「每個方位的最大半徑」——外形是虛線，逐點比對會被圖樣的洞
// 干擾；而畫面右緣會把外形切掉，所以只比落在戰場裡的那幾點。
func TestManoeuvreBoundaryMatchesOracle(t *testing.T) {
	const (
		w, cx, cy      = 640, 298, 409
		turn, dist, fc = 15, 240, 190
		fx0, fy0       = 16, 40   // 戰場左上
		fx1, fy1       = 464, 400 // 戰場右下
		shapeInk       = 2
	)
	scr, err := os.ReadFile(filepath.Join("..", "..",
		"workplace/oracle/s3cmdrange/screen.bin"))
	if err != nil || len(scr) < w*400 {
		t.Skip("缺收據 s3cmdrange，跳過（不用代用品）")
	}
	// 收據上每個方位的最大半徑。
	maxR := map[int]float64{}
	for y := fy0; y < fy1; y++ {
		for x := fx0; x < fx1; x++ {
			if scr[y*w+x] != shapeInk {
				continue
			}
			r := math.Hypot(float64(x-cx), float64(y-cy))
			a := int(math.Mod(math.Atan2(float64(y-cy), float64(x-cx))*
				128/math.Pi+256, 256))
			if r > maxR[a] {
				maxR[a] = r
			}
		}
	}
	if len(maxR) == 0 {
		t.Fatal("收據上找不到色 2 的外形")
	}

	pts := ManoeuvreBoundary(turn, dist, fc)
	if len(pts) == 0 {
		t.Fatal("算不出外形")
	}
	checked := 0
	for _, p := range pts {
		x, y := cx+p[0], cy+p[1]
		if x < fx0 || x >= fx1 || y < fy0 || y >= fy1 {
			continue // 被畫面切掉的那幾點，收據上本來就沒有
		}
		r := math.Hypot(float64(p[0]), float64(p[1]))
		a := int(math.Mod(math.Atan2(float64(p[1]), float64(p[0]))*
			128/math.Pi+256, 256))
		m, ok := maxR[a]
		if !ok {
			t.Errorf("方位 %d 算出半徑 %.1f，收據上那個方位沒有外形", a, r)
			continue
		}
		if d := r - m; d < -1.5 || d > 1.5 {
			t.Errorf("方位 %d：算出 %.1f、收據 %.1f，差 %+.1f", a, r, m, d)
		}
		checked++
	}
	if checked < 10 {
		t.Fatalf("只比到 %d 個點，收據上看得到的該有十個以上", checked)
	}
	t.Logf("比了 %d 個落在戰場裡的點，全部在 1.5 像素內", checked)
}

// 內圈：距離 ÷ 3 ＋ 8。收據量到圓心 (298.0,408.6)、半徑 88.1
// （`docs/re/120`）。
func TestManoeuvreInnerRadius(t *testing.T) {
	if got := ManoeuvreInnerRadius(240); got != 88 {
		t.Errorf("距離 240 的內圈半徑 %d，收據量到 88.1", got)
	}
}
