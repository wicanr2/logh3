package rules

// 原版的 16.16 定點數（`docs/re/147`）。
//
// 距離、離心率、橢圓半徑那一整串都走這一套：`39B5:000C` 乘、
// `39B5:005C` 除、`3310:014B` 開方、`3310:0006` 兩點距離。**乘除都是
// 先取絕對值算完再補號**，也就是往零截斷，不是往下取整；照抄才不會在
// 邊界上差一格。

// Fixed 是 16.16 定點數：整數部分在高 16 位。
type Fixed int32

// FixedOne 是 1.0。
const FixedOne Fixed = 1 << 16

// FixedFromInt 把整數變成 16.16。
func FixedFromInt(v int) Fixed { return Fixed(v) << 16 }

// Int 回整數部分（往零截斷，與原版取高位字一致）。
func (f Fixed) Int() int {
	if f < 0 {
		return -int((-f) >> 16)
	}
	return int(f >> 16)
}

func fixedAbs(f Fixed) (Fixed, bool) {
	if f < 0 {
		return -f, true
	}
	return f, false
}

// FixedMul 是 `39B5:000C`：64 位元乘積取中間 32 位。
func FixedMul(a, b Fixed) Fixed {
	x, na := fixedAbs(a)
	y, nb := fixedAbs(b)
	v := Fixed((int64(x) * int64(y)) >> 16)
	if na != nb {
		return -v
	}
	return v
}

// FixedDiv 是 `39B5:005C`：移位相減的除法，等於 `(a << 16) / b`。
func FixedDiv(a, b Fixed) Fixed {
	if b == 0 {
		return 0
	}
	x, na := fixedAbs(a)
	y, nb := fixedAbs(b)
	v := Fixed((int64(x) << 16) / int64(y))
	if na != nb {
		return -v
	}
	return v
}

// FixedSqrt 是 `3310:014B`：先把 1.0 一路左移到不小於被開方數，再跑
// 牛頓法，**收斂時回的是上一個值**（原版的 `if x ≥ t: return t`）。
func FixedSqrt(v Fixed) Fixed {
	if v <= 0 {
		return 0
	}
	x, t := FixedOne, v
	for x < t {
		x <<= 1
		t >>= 1
	}
	for {
		t = x
		x = (FixedDiv(v, x) + x) >> 1
		if x >= t {
			return t
		}
	}
}

// FixedDist 是 `3310:0006`：兩點距離，回 16.16。
//
// 原版把兩軸的差先左移 16 位再算 `長邊 × √(1 ＋ (短邊/長邊)²)`，
// 所以誤差的形狀與「先平方再開方」不同。
func FixedDist(x1, y1, x2, y2 int) Fixed {
	dx, dy := x1-x2, y1-y2
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	a, b := FixedFromInt(dx), FixedFromInt(dy)
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a < b {
		a, b = b, a
	}
	t := FixedDiv(b, a)
	return FixedMul(FixedSqrt(FixedMul(t, t)+FixedOne), a)
}
