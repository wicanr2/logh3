package rules

import "testing"

// 八個方向與原點的怪癖（收據 `cam-e`…`cam-nw`、`cam-zero`、`cam-anchor`）。
func TestViewScrollDir(t *testing.T) {
	for _, tc := range []struct{ x, y, dir int }{
		{400, 180, 0}, {400, 350, 1}, {224, 350, 2}, {100, 350, 3},
		{100, 180, 4}, {100, 30, 5}, {224, 30, 6}, {400, 30, 7},
		{224, 180, 2}, // dx ＝ 0 → 64 → 南
		{324, 230, 1}, // 對 (224,180) 是 26.6° 東南；對 (240,200) 只有 19.7°
	} {
		if got := ViewScrollDir(tc.x, tc.y); got != tc.dir {
			t.Errorf("(%d,%d) 方向 %d，要 %d", tc.x, tc.y, got, tc.dir)
		}
	}
}

func TestViewScrollStepClamps(t *testing.T) {
	if c, r, ok := ViewScrollStep(66, 30, 0); c != 67 || r != 30 || !ok {
		t.Errorf("往東：(%d,%d,%v)", c, r, ok)
	}
	if c, r, ok := ViewScrollStep(100, 30, 0); c != 100 || r != 30 || !ok {
		t.Errorf("右邊界往東：(%d,%d,%v)，一軸撞邊還是算走了", c, r, ok)
	}
	if c, r, ok := ViewScrollStep(0, 0, 5); c != 0 || r != 0 || ok {
		t.Errorf("左上角往西北：(%d,%d,%v)，兩軸都撞邊不動", c, r, ok)
	}
	if c, r, ok := ViewScrollStep(100, 84, 1); c != 100 || r != 84 || ok {
		t.Errorf("右下角往東南：(%d,%d,%v)", c, r, ok)
	}
}
