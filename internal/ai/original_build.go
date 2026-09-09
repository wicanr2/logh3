package ai

import "sort"

// 建艦隊那條路的前半（原版 `2D76:*`，`docs/internals/54`、`docs/re/202`）。
//
// 這裡只有**閘門與配額**兩段，都是純函式。真正把艦隊建出來的
// `2D76:02FA` 還有一段沒讀完（排序之後往下的挑法，`docs/re/191`），
// 所以 `ai.Turn` 目前不呼叫這裡——`docs/spec/029` 的階段一仍標「未做」。

// FleetCap 是一個陣營的艦隊上限。艦隊表每邊 16 個槽，`2D76:0060` 與
// `2E54:0513` 的迴圈條件都是 `ds:8F49[陣營] < 16`（`docs/internals/33`）。
const FleetCap = 16

// ClassQuota 是四個艦隊類別的目標比例（百分比，`2D76:01A5`）。
// 0 未命名 40％、1 占領 45％、2 防禦 5％、3 運用 10％。
var ClassQuota = [4]int{40, 45, 5, 10}

// BuildGate ＝ 原版 `2D76:0060()`：這一回合能不能建艦隊。
//
//	existing 現有艦隊數（ds:8F49[陣營]，同時是 2D76:011C 的回傳值）
//	tax      總稅收（ds:8F6A[陣營 × 4]）
//	upkeep   艦隊維持費（ds:8F72[陣營 × 4]）
//	tight    ds:8FD4[陣營] >= 3（語意未解，效果是預算打七五折）
//
// 回 `ok == false` 就是原版的 `0xFFFF`（不建）。三道關卡依序是
// **上限 16 → 少於三支直接放行 → 維持費 × 20 追上稅收就不建**
// （`docs/re/202`）。
func BuildGate(existing, tax, upkeep int, tight bool) (n int, ok bool) {
	if existing >= FleetCap {
		return 0, false
	}
	if existing < 3 {
		return existing, true // 少於三支不看錢
	}
	budget := tax
	if tight {
		budget = budget * 3 / 4
	}
	if upkeep*20 >= budget {
		return 0, false
	}
	return existing, true
}

// ClassQuotaOrder ＝ 原版 `2D76:011C` ＋ `2D76:01A5` ＋ `2D76:0251`：
// 回四個艦隊類別「最缺的排前面」的順序（`docs/internals/54`）。
//
//	counts 各類別現有幾支（原版填進 ds:D506 的四個 word）
//	total  艦隊總數
//
// 每個類別減掉 `total × 配額 ÷ 100`（整數除法捨去），再依這個差**升冪**
// 排序。差最小（最負）的就是最缺的那一類，原版照這個順序逐一試建，
// 成功一個就停。
func ClassQuotaOrder(counts [4]int, total int) [4]int {
	var short [4]int
	for i, pct := range ClassQuota {
		short[i] = counts[i] - total*pct/100
	}
	order := [4]int{0, 1, 2, 3}
	s := order[:]
	sort.SliceStable(s, func(a, b int) bool { return short[s[a]] < short[s[b]] })
	return order
}

// admiralPickTiers 是 `2D76:02FA` 的四層門檻（`docs/re/203`）。
// 亂數值域 0–0x7FFF，所以區間寬度是 0x1000／0x1000／0x2000／0x4000
// ——**一半的機率仍然挑第一名**。
var admiralPickTiers = []struct {
	threshold int // r 要 >= 這個值
	rank      int // 挑排序後的第幾名（0 起）
}{
	{0x7000, 3},
	{0x6000, 2},
	{0x4000, 1},
}

// AdmiralPickMinLevel 是被降格選中的提督自己要有的等級（`+0x06`）。
const AdmiralPickMinLevel = 16

// PickBuildAdmiral ＝ 原版 `2D76:02FA` 排序之後那一段：從**已經依分數
// 排好**的提督清單裡挑一位（`docs/re/203`）。
//
//	r      3CD5:000C() 的傳回值（0–0x7FFF）
//	count  清單長度
//	level  回排序後第 k 名的等級（`+0x06`）
//
// 回排序後的名次（0 ＝ 分數最高）。清單是空的回 -1。
//
// 每一層要「人數夠」而且「那個人等級 ≥ 16」，任一不成立就**往下一層試**
// ——不是直接落到第一名。
func PickBuildAdmiral(r, count int, level func(rank int) int) int {
	if count <= 0 {
		return -1
	}
	for _, t := range admiralPickTiers {
		if r >= t.threshold && count > t.rank && level(t.rank) >= AdmiralPickMinLevel {
			return t.rank
		}
	}
	return 0
}

// classKeys 是四個艦隊類別各自可以抽到的**類別鍵**（原版 `ds:5F62`／
// `ds:5F70`／`ds:5F6C`／`ds:5F74`，`docs/re/204`）。
//
// 這四張表是 `rules.FleetClass` 那張 `ds:5F52` 的反表——每個鍵
// `FleetClass(鍵)` 回的就是它所在的這一格，`TestClassKeysAreInverse`
// 逐項對過。
var classKeys = [4][]int{
	0: {0x00, 0x01, 0x02, 0x03, 0x04, 0x06, 0x07, 0x08, 0x09},
	1: {0x0A, 0x0B},
	2: {0x05, 0x0C, 0x0D},
	3: {0x0E, 0x0F},
}

// ClassKeys 回某一個類別可以抽到的類別鍵（唯讀）。
func ClassKeys(class int) []int {
	if class < 0 || class >= len(classKeys) {
		return nil
	}
	return classKeys[class]
}

// PickClassKey ＝ 原版 `2D76:0576`：從那一類的鍵裡等機率抽一個
// （`r` 是 `3CD5:000C()` 的傳回值）。原版把它寫進提督 `+0x08` 的高四位，
// 之後搬進艦隊記錄的 `+0x09`（`docs/internals/33`）。
//
// 類別不合法回 -1（原版沒有防守，呼叫端只給 0–3）。
func PickClassKey(class, r int) int {
	keys := ClassKeys(class)
	if len(keys) == 0 {
		return -1
	}
	return keys[r%len(keys)]
}
