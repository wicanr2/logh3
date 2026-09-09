package screen

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// 艦隊選點時畫的兩個範圍圖形（原版 `165F:0003`／`165F:004D` 產生、
// `3870:*` 畫，`docs/re/120`、`docs/re/135`）。
//
// **兩個都是 XOR 上去的**：EGA 的 Data Rotate/Function 設成 18h
// （bit 4–3 ＝ XOR，`2376:1810`），所以描述裡的顏色不是畫面上的顏色
// ——內圈的 5 疊在戰場底色 3 上變成 6，外形的 1 變成 2。

// 兩個圖形 XOR 用的顏色（原版描述結構的 `[+5]`）。
const (
	ManoeuvreInnerInk uint8 = 5
	ManoeuvreOuterInk uint8 = 1
)

// DrawManoeuvreRange 把內圈與外形畫在戰場上。
//
// cx／cy 是圓心的**世界座標**（旗艦位置各 ＋8），dist 是這一次的
// 「速度 × 8」，turn 是回轉速度，facing 是艦隊現在的朝向。
func DrawManoeuvreRange(s *BattleSpec, fb *gfx.Framebuffer,
	cx, cy, turn, dist, facing int, v BattleView) {
	if s == nil || fb == nil || dist <= 0 {
		return
	}
	sx := s.Field.X + cx - v.Col*s.Field.TileW
	sy := s.Field.Y + cy - v.Row*s.Field.TileH
	clip := func(x, y int) bool {
		return x >= s.Field.X && y >= s.Field.Y &&
			x < s.Field.X+s.Field.Cols*s.Field.TileW &&
			y < s.Field.Y+s.Field.Rows*s.Field.TileH
	}
	xorCircle(fb, sx, sy, rules.ManoeuvreInnerRadius(dist), ManoeuvreInnerInk, clip)
	pts := rules.ManoeuvreBoundary(turn, dist, facing)
	for i := 1; i < len(pts); i++ {
		xorLine(fb, sx+pts[i-1][0], sy+pts[i-1][1],
			sx+pts[i][0], sy+pts[i][1], ManoeuvreOuterInk, clip)
	}
}

// xorLine 走 Bresenham，把顏色 XOR 上去。
func xorLine(fb *gfx.Framebuffer, x0, y0, x1, y1 int, ink uint8, ok func(int, int) bool) {
	dx, sx := abs(x1-x0), 1
	if x0 > x1 {
		sx = -1
	}
	dy, sy := -abs(y1-y0), 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		if ok(x0, y0) {
			fb.Set(x0, y0, fb.At(x0, y0)^ink)
		}
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// xorCircle 用中點法畫一個圓，同樣是 XOR。
func xorCircle(fb *gfx.Framebuffer, cx, cy, r int, ink uint8, ok func(int, int) bool) {
	if r <= 0 {
		return
	}
	x, y, e := r, 0, 1-r
	put := func(px, py int) {
		if ok(px, py) {
			fb.Set(px, py, fb.At(px, py)^ink)
		}
	}
	for x >= y {
		for _, p := range [8][2]int{
			{cx + x, cy + y}, {cx + y, cy + x}, {cx - y, cy + x}, {cx - x, cy + y},
			{cx - x, cy - y}, {cx - y, cy - x}, {cx + y, cy - x}, {cx + x, cy - y},
		} {
			put(p[0], p[1])
		}
		y++
		if e < 0 {
			e += 2*y + 1
		} else {
			x--
			e += 2*(y-x) + 1
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
