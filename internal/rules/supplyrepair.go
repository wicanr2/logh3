package rules

// 戰場上的「補給」與「修理」（原版 `1930:05EC`、`1AAB:0583`，
// `docs/re/131`）。兩支的骨架一模一樣：先算出一個作用半徑，收集半徑內
// 的候選分隊讓玩家點一個，點下去就在兩筆艦艇記錄之間搬數量。

// 補給艦與修理艦的艦種編號（作用半徑查的是**這兩個艦種**那一列，
// 不是發令分隊自己的那一列）。
const (
	SupplyKind = 5
	RepairKind = 6
)

// 半徑的倍率。原版**同一個 6.4 算了兩次，而且兩次不一樣**：
//
//   - 收集候選（`1930:05CF`／`1AAB:05CF`）：速度左移 21 位再除以 5，
//     30 → 12,582,912 ＝ 正好 192.0。
//   - 按下去確認（`1930:08D6`）：速度乘 `0x00066666`（16.16 的 6.39999），
//     30 → 12,582,900 ＝ 191.99982。
//
// 差 12／65536 個像素。照抄兩邊，不合成一個——落在那條縫上的目標
// 列得出來卻按不下去，那是原版的行為。
const (
	supplyRangeShift = 21
	supplyRangeDiv   = 5
	supplyRangeMul   = 0x66666
)

// SupplyListRange 回「收集候選」用的半徑，16.16。
func SupplyListRange(t *ShipKindTable, faction, kind int) int64 {
	return int64(kindSpeed(t, faction, kind)) << supplyRangeShift / supplyRangeDiv
}

// SupplyRange 回「按下去確認」用的半徑，16.16。比 SupplyListRange 小一點。
func SupplyRange(t *ShipKindTable, faction, kind int) int64 {
	return int64(kindSpeed(t, faction, kind)) * supplyRangeMul
}

func kindSpeed(t *ShipKindTable, faction, kind int) int {
	if t == nil {
		return 0
	}
	k, ok := t.Kind(faction, kind)
	if !ok {
		return 0
	}
	return k.Byte(kindTopSpeed)
}

// InSupplyRange 回目標在不在半徑內。距離走原版的 `3310:0006`。
func InSupplyRange(sx, sy, tx, ty int, r int64) bool {
	return Distance16(sx, sy, tx, ty) <= r
}

// Supply 把補給艦身上的飛彈與艦載機補給目標，回實際給出的兩個量。
//
// 目標的滿載量 ＝ 艦種列 `+0x10`（飛彈每艘幾發）與 `+0x14`（艦載機
// 滿量）各乘上**目標現有的艘數** `+0x08`；缺多少就給多少，補給艦身上
// 不夠就給到見底。兩邊都是 32 位元。
func Supply(t *ShipKindTable, from, to *BattleUnit) (missiles, fighters int) {
	if t == nil || from == nil || to == nil {
		return 0, 0
	}
	k, ok := t.Kind(to.Faction, to.Kind)
	if !ok {
		return 0, 0
	}
	missiles = give(k.Byte(kindMissilePer)*to.Ships-to.Missiles, from.Missiles)
	fighters = give(k.Byte(kindFighterFull)*to.Ships-to.Fighters, from.Fighters)
	if missiles == 0 && fighters == 0 {
		return 0, 0
	}
	from.Missiles -= missiles
	from.Fighters -= fighters
	to.Missiles += missiles
	to.Fighters += fighters
	return missiles, fighters
}

// Repair 把修理艦身上的存量換成目標的艘數，回實際補回幾艘。
//
// 上限是目標記錄的 `+0x0C`（**不是**分隊面板印的滿編 `+0x0A`），
// 修理艦的存量放在它自己的 `+0x0E`——與補給艦的飛彈同一格。
func Repair(from, to *BattleUnit) int {
	if from == nil || to == nil {
		return 0
	}
	n := give(to.RepairCap-to.Ships, from.Missiles)
	if n == 0 {
		return 0
	}
	to.Ships += n
	from.Missiles -= n
	return n
}

// give 回「缺多少」與「拿得出多少」的較小者，負的當 0。
func give(need, have int) int {
	if need <= 0 || have <= 0 {
		return 0
	}
	if need < have {
		return need
	}
	return have
}

// kindFighterFull 是艦種列的「艦載機滿量」（`+0x14`）。分隊面板拿它
// 判斷艦載機滿不滿（`docs/re/118`），補給也拿它當上限。
const kindFighterFull = 0x14

// SupplyRangePixels 回「補給」／「修理」選對象時畫在戰場上那個圈的半徑
// （像素）。
//
// `165F:09D9` 直接算 `最高速度 × 32 ÷ 5`（五次 `shl`／`rcl` 之後除 5），
// 不是把 16.16 的 SupplyListRange 截下來——兩者目前每一個速度都給同一個
// 數，但**算式不同就分開寫**，哪天發現差一個數才知道是誰錯（`docs/re/140`）。
func SupplyRangePixels(t *ShipKindTable, faction, kind int) int {
	return kindSpeed(t, faction, kind) * 32 / 5
}
