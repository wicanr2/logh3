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
	r := FieldClip(s)
	xorCircle(fb, sx, sy, rules.ManoeuvreInnerRadius(dist), ManoeuvreInnerInk, r.Inside)
	pts := rules.ManoeuvreBoundary(turn, dist, facing)
	for i := 1; i < len(pts); i++ {
		r.XorLine(fb, sx+pts[i-1][0], sy+pts[i-1][1],
			sx+pts[i][0], sy+pts[i][1], ManoeuvreOuterInk)
	}
}

// ClipRect 是原版畫線之前的裁切矩形（`ds:8D6A`／`8D6C`／`8D6E`／`8D70`，
// 兩端都含）。戰場那一組就是 `FieldClip`。
type ClipRect struct{ X0, Y0, X1, Y1 int }

// FieldClip 回戰場的裁切矩形。
func FieldClip(s *BattleSpec) ClipRect {
	return ClipRect{
		X0: s.Field.X, Y0: s.Field.Y,
		X1: s.Field.X + s.Field.Cols*s.Field.TileW - 1,
		Y1: s.Field.Y + s.Field.Rows*s.Field.TileH - 1,
	}
}

// Inside 回這個點在不在矩形裡。
func (r ClipRect) Inside(x, y int) bool {
	return x >= r.X0 && x <= r.X1 && y >= r.Y0 && y <= r.Y1
}

// outcode 是 `3870:09F1` 的四個位元：1 左、2 右、8 上、4 下。
func (r ClipRect) outcode(x, y int) int {
	c := 0
	switch {
	case x < r.X0:
		c |= 1
	case x > r.X1:
		c |= 2
	}
	switch {
	case y < r.Y0:
		c |= 8
	case y > r.Y1:
		c |= 4
	}
	return c
}

// move 是 `3870:0A30`：把落在外面的那一端拉到邊界上。除法**往零截斷**
// （`3CC9:70`），少了這一條端點會差一格。
func (r ClipRect) move(code, x, y, ox, oy int) (int, int) {
	switch {
	case code&8 != 0:
		x += (r.Y0 - y) * (x - ox) / (y - oy)
		y = r.Y0
	case code&4 != 0:
		x += (r.Y1 - y) * (x - ox) / (y - oy)
		y = r.Y1
	case code&1 != 0:
		y += (r.X0 - x) * (y - oy) / (x - ox)
		x = r.X0
	case code&2 != 0:
		y += (r.X1 - x) * (y - oy) / (x - ox)
		x = r.X1
	}
	return x, y
}

// Clip 是 `3870:08FF` 的 Cohen–Sutherland。整段在外面就回 false。
//
// **裁切換掉的是端點，不是「畫的時候跳過畫面外的點」**：接下來的
// Bresenham 從新端點起算，斜率累加的相位跟著變，差一格就整段錯開
// （`docs/re/155`）。
func (r ClipRect) Clip(x0, y0, x1, y1 int) (int, int, int, int, bool) {
	c0, c1 := r.outcode(x0, y0), r.outcode(x1, y1)
	for n := 0; n < 8; n++ {
		if c0&c1 != 0 {
			return 0, 0, 0, 0, false
		}
		if c0 == 0 && c1 == 0 {
			return x0, y0, x1, y1, true
		}
		if c0 != 0 {
			x0, y0 = r.move(c0, x0, y0, x1, y1)
			c0 = r.outcode(x0, y0)
			if c0 == 0 && c1 == 0 {
				return x0, y0, x1, y1, true
			}
		}
		if c0&c1 != 0 {
			return 0, 0, 0, 0, false
		}
		if c1 != 0 {
			x1, y1 = r.move(c1, x1, y1, x0, y0)
			c1 = r.outcode(x1, y1)
		}
	}
	return 0, 0, 0, 0, false
}

// XorLine 裁切之後照原版 `39D7:0936` 走一條線，把顏色 XOR 上去。
//
// 原版分四支：|Δx| > |Δy|（橫向為主）、|Δx| ＝ |Δy|（正對角）、
// Δx ＝ 0（垂直）、其餘（縱向為主）。呼叫端 `3870:0833` 先把兩端
// **按 x 排好**，所以次要軸的累加器是 0.16 定點、起始值 `0x8000`
// （等於四捨五入），每畫完一點才累加。
//
// 共用端點會被畫兩次 ——XOR 之後就消失，原版的折線在每個轉折上
// 都少一個像素，那不是漏畫。
func (r ClipRect) XorLine(fb *gfx.Framebuffer, x0, y0, x1, y1 int, ink uint8) {
	cx0, cy0, cx1, cy1, ok := r.Clip(x0, y0, x1, y1)
	if !ok {
		return
	}
	put := func(x, y int) { fb.Set(x, y, fb.At(x, y)^ink) }
	if cx1 < cx0 {
		cx0, cy0, cx1, cy1 = cx1, cy1, cx0, cy0
	}
	dx, dy := cx1-cx0, cy1-cy0
	ady, sy := abs(dy), 1
	if dy < 0 {
		sy = -1
	}
	switch {
	case dx > ady: // 橫向為主
		step, acc, y := (ady<<16)/dx, 0x8000, cy0
		for k := 0; k <= dx; k++ {
			put(cx0+k, y)
			if acc += step; acc >= 0x10000 {
				acc -= 0x10000
				y += sy
			}
		}
	case dx == ady: // 正對角
		for k := 0; k <= dx; k++ {
			put(cx0+k, cy0+k*sy)
		}
	case dx == 0: // 垂直
		for k := 0; k <= ady; k++ {
			put(cx0, cy0+k*sy)
		}
	default: // 縱向為主
		step, acc, x := (dx<<16)/ady, 0x8000, cx0
		for k := 0; k <= ady; k++ {
			put(x, cy0+k*sy)
			if acc += step; acc >= 0x10000 {
				acc -= 0x10000
				x++
			}
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

// 選抵達朝向時，目的地上那個指向性的十字（原版描述 `[+6] ＝ 4`、
// `3870:042F`，`docs/re/122`）。
//
// 兩條線，都依朝向旋轉：沿朝向從後方 FacingCrossBack 拉到前方
// FacingCrossFront，加上垂直方向 ±FacingCrossSide。遮罩是 0xFFFF
// ——**實線**；畫面上看起來斷斷續續是 XOR 疊在不同底色上的結果。
const (
	FacingCrossInk   uint8 = 4
	FacingCrossFront       = 0x40 // 描述 [+8]
	FacingCrossBack        = 0x20 // 描述 [+0A]
	FacingCrossSide        = 0x20 // 描述 [+0C]
	FacingCrossBias        = 7    // 中心 ＝ 選到的點各 ＋7
)

// DrawFacingCross 把那個十字畫在目的地上。tx／ty 是**世界座標**。
func DrawFacingCross(s *BattleSpec, fb *gfx.Framebuffer,
	tx, ty, facing int, v BattleView) {
	if s == nil || fb == nil {
		return
	}
	cx := s.Field.X + tx + FacingCrossBias - v.Col*s.Field.TileW
	cy := s.Field.Y + ty + FacingCrossBias - v.Row*s.Field.TileH
	r := FieldClip(s)
	for _, seg := range FacingCrossSegments(facing) {
		r.XorLine(fb, cx+seg[0], cy+seg[1], cx+seg[2], cy+seg[3], FacingCrossInk)
	}
}

// FacingCrossSegments 回兩條線的端點（相對中心的位移）。
func FacingCrossSegments(facing int) [2][4]int {
	ax, ay := rules.Rotate256(-FacingCrossBack, 0, facing)
	bx, by := rules.Rotate256(FacingCrossFront, 0, facing)
	cx, cy := rules.Rotate256(0, -FacingCrossSide, facing)
	dx, dy := rules.Rotate256(0, FacingCrossSide, facing)
	return [2][4]int{{ax, ay, bx, by}, {cx, cy, dx, dy}}
}

// 要塞砲的射程是一個橢圓（描述 `[+6] ＝ 2`，`docs/re/134`）。
const FortressRangeInk uint8 = 4

// DrawFortressRange 把要塞砲的射程橢圓畫在行星上。
//
// fx／fy 是**焦點**的世界座標——要塞站在焦點上，不在中心，所以射程
// 在瞄準的那一側長、背面短。中心 ＝ 焦點 ＋ 沿 tilt 走
// `rules.FortressFocalDistance` 個像素（`165F:0A71`）。
//
// 曲線本身是**中點圓演算法**（`39D7:035E`）：拿長半軸畫一個整數圓，
// 每個點的另一軸乘 短/長 壓扁，旋轉 tilt 之後畫四組對稱點的正負兩端。
// 一圈只採 256 個角度會在長軸兩端留下空隙，與原版對不上
// （收據 `s3fort`／`s3fort2`，`docs/re/145`）。
func DrawFortressRange(s *BattleSpec, fb *gfx.Framebuffer,
	fx, fy, semiMajor, semiMinor, tilt int, v BattleView) {
	if s == nil || fb == nil {
		return
	}
	ox := s.Field.X - v.Col*s.Field.TileW
	oy := s.Field.Y - v.Row*s.Field.TileH
	for _, p := range FortressRangePoints(fx, fy, semiMajor, semiMinor, tilt) {
		x, y := p[0]+ox, p[1]+oy
		if x < s.Field.X || y < s.Field.Y ||
			x >= s.Field.X+s.Field.Cols*s.Field.TileW ||
			y >= s.Field.Y+s.Field.Rows*s.Field.TileH {
			continue
		}
		// XOR：同一個像素被畫兩次就等於沒畫，原版就是靠這個
		// 讓橢圓疊在地形上還能擦乾淨。
		fb.Set(x, y, fb.At(x, y)^FortressRangeInk)
	}
}

// FortressRangePoints 回橢圓要畫的**世界座標**點，順序與原版一致；
// 同一個點出現兩次就是被 XOR 兩次（等於沒畫），所以呼叫端要照次數
// 的奇偶決定畫不畫。
func FortressRangePoints(fx, fy, semiMajor, semiMinor, tilt int) [][2]int {
	if semiMajor <= semiMinor || semiMinor <= 0 {
		return nil
	}
	c := rules.FortressFocalDistance(semiMajor, semiMinor)
	ox, oy := rules.Rotate256(c, 0, tilt)
	cx, cy := fx+ox, fy+oy
	var pts [][2]int
	var last [4][2]int
	var seen [4]bool
	x, err, y := semiMajor, semiMajor, 0
	for x >= y {
		px := x * semiMinor / semiMajor
		py := y * semiMinor / semiMajor
		for i, u := range [4][2]int{{x, py}, {x, -py}, {y, px}, {y, -px}} {
			rx, ry := rules.Rotate256(u[0], u[1], tilt)
			if seen[i] && last[i] == [2]int{rx, ry} {
				continue
			}
			seen[i], last[i] = true, [2]int{rx, ry}
			pts = append(pts, [2]int{cx + rx, cy + ry}, [2]int{cx - rx, cy - ry})
		}
		err -= 2*y - 1
		y++
		if err < 0 {
			err += 2 * (x - 1)
			x--
		}
	}
	return pts
}

// SupplyRangeInk 是「補給」／「修理」選對象時那個圈的色號。
//
// 圈是 `165F:09D9` 掛上去的圖形描述：中心 ＝ 單位 `+0x34`／`+0x36` 各 ＋8
// （佔位框中心）、`[+5] = 5`（色）、`[+6] = 1`（圓）、半徑 ＝ 艦種列
// `+0x06`（最高速度）× 32 ÷ 5，也就是 `rules.SupplyRange` 的那個 6.4
// （`docs/re/140`）。
const SupplyRangeInk = 5

// DrawSupplyRange 把補給／修理的查詢範圍畫在戰場上。cx／cy 是單位的
// 世界座標（不含 ＋8），r 是半徑。
func DrawSupplyRange(s *BattleSpec, fb *gfx.Framebuffer, cx, cy, r int, v BattleView) {
	if s == nil || fb == nil || r <= 0 {
		return
	}
	sx := s.Field.X + cx + 8 - v.Col*s.Field.TileW
	sy := s.Field.Y + cy + 8 - v.Row*s.Field.TileH
	clip := func(x, y int) bool {
		return x >= s.Field.X && y >= s.Field.Y &&
			x < s.Field.X+s.Field.Cols*s.Field.TileW &&
			y < s.Field.Y+s.Field.Rows*s.Field.TileH
	}
	xorCircle(fb, sx, sy, r, SupplyRangeInk, clip)
}
