package rules

import "testing"

// 原版 s3ai6 那一場，帝國第 1 艦隊十個分隊相對旗艦的位移
// （`docs/internals/38`、`tools/scripts/squadrons.py`）。
var origFormation = [10][2]int{
	{0, 0}, {22, 6}, {6, -22}, {44, 12}, {12, -44},
	{65, 18}, {50, -10}, {34, -38}, {18, -65}, {28, -16},
}

func TestFormationMatchesOriginal(t *testing.T) {
	// 基底就是第 1 個分隊的位移。
	ux, uy := origFormation[1][0], origFormation[1][1]
	for n, want := range origFormation {
		x, y := Formation(n, ux, uy)
		// a 或 b 到 3 的那兩格原版少 1，其餘要完全一致。
		d := abs(x-want[0]) + abs(y-want[1])
		if d > 1 {
			t.Errorf("分隊 %d：算出 (%d,%d)，原版 (%d,%d)", n, x, y, want[0], want[1])
		}
	}
}

// 法線的方向不能反——反了整支艦隊會排到另一側。
func TestFormationNormalSide(t *testing.T) {
	x, y := Formation(2, 22, 6)
	if x != 6 || y != -22 {
		t.Fatalf("第 2 個分隊算出 (%d,%d)，原版 (6,-22)", x, y)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
