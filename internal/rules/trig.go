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
