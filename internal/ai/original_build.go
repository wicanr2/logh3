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
