package rules

// sinQuarter 是 sin 的第一象限查表：索引 0–64 對應 0°–90°，
// 值是 sin × 1024（`docs/internals/39` 的角度制，一整圈 256）。
//
// **整數表，不用浮點**：戰場的每一個座標都要能逐位元組重現，
// 浮點在不同平台會差最後一位。
var sinQuarter = [65]int{
	0, 25, 50, 75, 100, 125, 150, 175,
	200, 224, 249, 273, 297, 321, 345, 369,
	392, 415, 438, 460, 483, 505, 526, 548,
	569, 590, 610, 630, 650, 669, 688, 706,
	724, 742, 759, 775, 792, 807, 822, 837,
	851, 865, 878, 891, 903, 915, 926, 936,
	946, 955, 964, 972, 980, 987, 993, 999,
	1004, 1009, 1013, 1016, 1019, 1021, 1023, 1024,
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

// cos256 回 cos(a) × 1024。
func cos256(a int) int { return sin256(a + 64) }
