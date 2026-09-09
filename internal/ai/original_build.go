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

// levelScale ＝ 原版 `ds:5FCA`：同盟缺錢時拿等級換算出來的「基底」
// （`2D76:0434`，`docs/re/205`）。
//
// **前 16 格才是那張表本身**（`1 1 2 2 3 3 4 4 5 5 5 5 6 6 6 6`）。
// 原版用等級直接當 word 索引、沒有夾限，而提督的等級到 41——
// 索引 16 以後讀的是隔壁的 `ds:5FEA`，值從傾印量出來就是下面這些。
// 這裡照量到的重現，不修掉。
var levelScale = [44]int{
	1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 5, 5, 6, 6, 6, 6, // ds:5FCA 本體
	30, 31, 32, 33, 40, 41, 42, 43, // 以下讀出表尾，落在 ds:5FEA
	29, 28, 27, 26, 25, 24, 23, 22,
	21, 20, 19, 18, 17, 16, 15, 14,
	13, 12, 11, 10,
}

// AdmiralView 是建艦隊評分要看的提督欄位（`docs/internals/08`）。
type AdmiralView struct {
	Index   int    // 提督表裡的索引
	Level   int    // `+0x06` 等級
	Command int    // `+0x0B` 統率力
	Ability [4]int // 四個類別各自對應的能力：+0x14／+0x13／+0x12／+0x17
	Kind    int    // `+0x08 & 0x0F` 艦種
}

// AllianceBrokeBelow 是 `2D76:0434` 那條「同盟缺錢」分支的門檻：
// 總稅收低於它才改查 `levelScale`。
const AllianceBrokeBelow = 20000

// AdmiralBuildScore ＝ 原版 `2D76:0434(提督, 類別)`：算一位提督在某個
// 艦隊類別的分數（原版寫進提督 `+0x1A`，排序器 `2D76:04FD` 比的就是它）。
//
//	faction 目前在跑 AI 的陣營（0 帝國、1 同盟）
//	tax     那個陣營的總稅收（ds:8F6A[陣營 × 4]）
//
// 等級超過 43 的沒有觀察到，夾在 43——**那是 remake 自己的收斂**，
// 原版會繼續往下讀（`docs/re/205`）。
func AdmiralBuildScore(a AdmiralView, class, faction, tax int) int {
	base := a.Level * 2
	if faction == 1 && tax < AllianceBrokeBelow {
		i := a.Level
		if i < 0 {
			i = 0
		}
		if i >= len(levelScale) {
			i = len(levelScale) - 1
		}
		base = levelScale[i]
	}
	score := base * a.Command / 100
	if class >= 0 && class < len(a.Ability) {
		score *= a.Ability[class]
	}
	if a.Kind == 1 {
		score = score * 3 / 2
	}
	return score
}
