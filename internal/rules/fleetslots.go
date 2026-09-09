package rules

// fleetSlotSteps 是原版 `297B:0209` 的九段表：**提督等級 → 艦隊應有組數**
// （艦隊記錄的 `+0x0E`，`docs/re/214`）。
//
// 每一列是「等級低於 lo 就回 n」，由上往下第一個成立的算數；
// 全部不成立回 10。**4 之後直接跳到 6**，沒有 5——原版就是這樣寫的。
var fleetSlotSteps = []struct{ lo, n int }{
	{2, 1}, {6, 2}, {16, 3}, {20, 4}, {24, 6}, {29, 7}, {34, 8}, {39, 9},
}

// FleetSlots 回一位提督帶的艦隊**應該**有幾組編制（`+0x0E`）。
//
// 收據 `turn260mem` 上三支的 `+0x0F` 剛好等於這個值：等級 20 → 6、
// 18 → 4、16 → 4（`docs/re/209`、`docs/re/214`）。
func FleetSlots(level int) int {
	for _, s := range fleetSlotSteps {
		if level < s.lo {
			return s.n
		}
	}
	return 10
}

// NewFleetMorale 是新編艦隊的 `+0x3C`：原版 `297B:0066` 寫死 100
// （`docs/re/214`）。那一格在艦隊組成畫面叫「機動力」、在戰術選單的
// 明細叫「艦隊士氣值」（`docs/internals/33`）。
const NewFleetMorale = 100
