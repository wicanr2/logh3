package rules

import "testing"

// 原版 s3ai6 那一場在場的十八支艦隊：位置與記錄 +0x4A 的值
// （`tools/scripts/squadrons.py`、`docs/internals/39`）。
var origFacing = []struct{ X, Y, Facing int }{
	{1480, 752, 107}, {1512, 784, 110}, {1480, 784, 109}, {1448, 784, 108},
	{1448, 720, 103}, {936, 1172, 210}, {368, 1640, 224}, {1064, 1132, 177},
	{1544, 688, 105}, {416, 280, 37}, {368, 1608, 225}, {1544, 816, 113},
	{1536, 1840, 170}, {1536, 1808, 169}, {1568, 1808, 168}, {1504, 1872, 171},
	{1504, 1808, 170}, {1536, 1776, 168},
}

// BattleFacing 要與原版填進 +0x4A 的值一致；原版查的是反正切表，
// 所以允許差 1。
func TestBattleFacingMatchesOriginal(t *testing.T) {
	for _, c := range origFacing {
		got := BattleFacing(c.X, c.Y)
		d := got - c.Facing
		if d < 0 {
			d = -d
		}
		if d > BattleTurn/2 {
			d = BattleTurn - d
		}
		if d > 1 {
			t.Errorf("(%d,%d) 算出 %d，原版 %d", c.X, c.Y, got, c.Facing)
		}
	}
}

// 帝國那十二支的佈陣基底：方向是朝向減 96、長度 22–23。
func TestFormationBaseMatchesOriginal(t *testing.T) {
	cases := []struct {
		X, Y, UX, UY int
	}{
		{1480, 752, 22, 6}, {1512, 784, 21, 8}, {1480, 784, 21, 7},
		{1448, 784, 22, 7}, {1448, 720, 22, 4}, {936, 1172, -21, 8},
		{368, 1640, -23, 0}, {1544, 688, 22, 5}, {416, 280, 3, -22},
		{368, 1608, -23, -1}, {1544, 816, 21, 9},
	}
	for _, c := range cases {
		ux, uy := FormationBase(BattleFacing(c.X, c.Y))
		if abs(ux-c.UX) > 2 || abs(uy-c.UY) > 2 {
			t.Errorf("(%d,%d) 算出基底 (%d,%d)，原版 (%d,%d)",
				c.X, c.Y, ux, uy, c.UX, c.UY)
		}
	}
}

// 整數三角表本身：sin²+cos² 要接近 1024²。
func TestTrigTableIsConsistent(t *testing.T) {
	for a := 0; a < BattleTurn; a++ {
		s, c := sin256(a), cos256(a)
		n := s*s + c*c
		if n < 1024*1024*99/100 || n > 1024*1024*101/100 {
			t.Fatalf("角度 %d：sin²+cos² ＝ %d，應接近 %d", a, n, 1024*1024)
		}
	}
}
