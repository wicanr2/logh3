package rules

import "testing"

// TestStepAlongRoundsAndClamps 釘住 `2BE1:026F` 的算式（`docs/re/123`）。
//
// 原版是「32 位元乘完加 512 再算術右移 10 位」＝ 四捨五入，不是截斷；
// 這裡挑的方位就是兩者會差一個像素的。**這是照反組譯釘的規則，
// 不是對收據**——原版真正走一步的落點還沒取到 same-state 樣本。
func TestStepAlongRoundsAndClamps(t *testing.T) {
	rounded := 0
	for a := 0; a < BattleTurn; a++ {
		for _, d := range []int{8, 16, 24} {
			gx, gy := StepAlong(1000, 1000, a, d)
			wx := 1000 + (cos256(a)*d+512)>>10
			wy := 1000 + (sin256(a)*d+512)>>10
			if gx != wx || gy != wy {
				t.Fatalf("方位 %d 距離 %d：走到 (%d,%d)，應該是 (%d,%d)",
					a, d, gx, gy, wx, wy)
			}
			// 截斷版本與四捨五入版本差多少格
			if tx := 1000 + cos256(a)*d/1024; tx != wx {
				rounded++
			}
		}
	}
	if rounded == 0 {
		t.Error("挑不出四捨五入與截斷不同的樣本，測試等於沒在測")
	}
	t.Logf("768 組樣本裡有 %d 組 x 分量與截斷不同", rounded)

	// 夾範圍：往正方向走過頭要被夾住，負方向夾到 0。
	if x, y := StepAlong(BattleWorldMaxX, BattleWorldMaxY, 0, 100); x != BattleWorldMaxX {
		t.Errorf("x 沒夾住：%d,%d", x, y)
	}
	if x, _ := StepAlong(0, 0, 128, 100); x != 0 {
		t.Errorf("x 沒夾到 0：%d", x)
	}
	if _, y := StepAlong(0, 0, 192, 100); y != 0 {
		t.Errorf("y 沒夾到 0：%d", y)
	}
}
