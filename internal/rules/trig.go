package rules

// sinQuarter 是 sin 的第一象限查表：索引 0–64 對應 0°–90°，
// 值是 sin × 1024（一整圈 256）。
//
// **這是原版 `ds:573A` 那張表倒出來的**（`docs/internals/41`），不是
// 自己算的——原版取的是截斷不是四捨五入，自己算會有幾格差 1，
// 而那足以讓佈陣的座標差一個像素。
var sinQuarter = [65]int{
	0, 25, 50, 75, 100, 125, 150, 175,
	199, 224, 248, 273, 297, 321, 344, 368,
	391, 414, 437, 460, 482, 504, 526, 547,
	568, 589, 609, 629, 649, 668, 687, 706,
	724, 741, 758, 775, 791, 807, 822, 837,
	851, 865, 878, 890, 903, 914, 925, 936,
	946, 955, 964, 972, 979, 986, 993, 999,
	1004, 1008, 1012, 1016, 1019, 1021, 1022, 1023,
	1024,
}

// sin256 回 sin(a) × 1024，a 的一整圈是 256。
func sin256(a int) int {
	a = ((a % BattleTurn) + BattleTurn) % BattleTurn
	switch {
	case a <= 64:
		return sinQuarter[a]
	case a <= 128:
		return sinQuarter[128-a]
	case a <= 192:
		return -sinQuarter[a-128]
	}
	return -sinQuarter[BattleTurn-a]
}

// cos256 回 cos(a) × 1024。原版就是拿同一張表用「角度 ＋ 64」查的
// （`2264:10BA`）。
func cos256(a int) int { return sin256(a + 64) }

// 戰場世界座標的上下限（原版 `2BE1:026F` 走完一步之後夾的值，
// 也是選目的地時的判準，`docs/re/121`）。
const (
	BattleWorldMaxX = 0x7EF // 2031
	BattleWorldMaxY = 0x7E7 // 2023
)

// StepAlong 從 (x,y) 沿方位 a 走 dist 像素，回落點（原版 `2BE1:026F`）。
//
// 原版是 32 位元乘完之後**加 512 再算術右移 10 位**——也就是四捨五入，
// 不是截斷；最後把兩軸各夾進 0–2031／0–2023。差一個像素會讓直線判定
// 走到別的半格上去，所以這裡照抄（`docs/re/123`）。
func StepAlong(x, y, a, dist int) (int, int) {
	x += (cos256(a)*dist + 512) >> 10
	y += (sin256(a)*dist + 512) >> 10
	return clampInt(x, 0, BattleWorldMaxX), clampInt(y, 0, BattleWorldMaxY)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Atan256 回 (dx,dy) 的方位，一整圈 256（原版 `2BE1:011D`）。
// 選抵達朝向與算「往目的地的方位」都用它。
func Atan256(dx, dy int) int { return atan256(dx, dy) }

// Sin256／Cos256 是原版正弦表（`ds:573A`）的公開版本：一整圈 256、
// 值 ＝ sin × 1024（`docs/internals/41`）。
func Sin256(a int) int { return sin256(a) }

// Cos256 見 Sin256。
func Cos256(a int) int { return cos256(a) }
