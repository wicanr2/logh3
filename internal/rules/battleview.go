package rules

// 戰場視野（鏡頭 `ds:9BF4`／`ds:9BF6`，格座標 `ds:9BA3`／`ds:9BA4`）的滑鼠捲動
// （`1588:043F` → `1588:03A9`，`docs/re/326`）。
//
// 原版在戰場視窗上按住右鍵：每一圈重讀游標（int 33h，滑鼠座標 0–639 × 0–399）、
// 算它相對於 (224,180) 的方位（`2BE1:011D` 的 atan2 → `3982:0022` 八個方向）、
// 視野往那邊走一格，放開右鍵才停。那個原點是戰場的半寬半高 (448/2, 360/2)，
// 忘了加戰場在滑鼠座標裡的左上角 (16,20)——所以它不是戰場正中央 (240,200)
// （收據 `cam-anchor`：(324,230) 往東南走，對正中央算是東）。
const (
	ViewScrollAnchorX = 224 // 滑鼠座標
	ViewScrollAnchorY = 180
	// ViewMaxCol／ViewMaxRow 是視野左上角格的上限：整張 128×102、看得到 28×18
	// （`1588:04A0` 夾 0–100 × 0–84）。
	ViewMaxCol = 100
	ViewMaxRow = 84
)

// viewStep 是 `ds:3C1E`：八個方向各走幾格（0 東、1 東南、2 南、…、7 東北）。
var viewStep = [8][2]int{{1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}}

// ViewScrollDir 回按住右鍵時視野該走的方向 0–7；(mx,my) 是滑鼠座標。
// 正好在原點上（dx ＝ 0）照 atan2 的那個怪癖算成 64 ＝ 往南（收據 `cam-zero`）。
func ViewScrollDir(mx, my int) int {
	return Direction8(atan256(mx-ViewScrollAnchorX, my-ViewScrollAnchorY))
}

// ViewScrollStep 把視野 (col,row) 往 dir 走一格，夾進 0–ViewMaxCol × 0–ViewMaxRow。
// 回 false ＝ 兩軸都撞到邊（`1588:04A0` 夾了兩次以上），原版那一圈不重畫。
func ViewScrollStep(col, row, dir int) (int, int, bool) {
	d := viewStep[dir&7]
	col, row = col+d[0], row+d[1]
	clamped := 0
	clamp := func(v, hi int) int {
		if v < 0 {
			clamped++
			return 0
		}
		if v > hi {
			clamped++
			return hi
		}
		return v
	}
	col, row = clamp(col, ViewMaxCol), clamp(row, ViewMaxRow)
	return col, row, clamped < 2
}
