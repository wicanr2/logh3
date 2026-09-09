package ai

import (
	"sort"

	"github.com/wicanr2/logh3/internal/rules"
)

// 階段一（建艦隊）整條串起來：原版 `2D76:0002()` driven 的那一趟
// （`docs/internals/54`、`docs/re/202`–`docs/re/205`）。
//
//	2D76:0060()          能不能建（上限／少於三支／維持費 × 20）
//	2D76:011C／01A5／0251 各類別 現有 − 應有，升冪，最缺的先試
//	2D76:02FA(類別)      全部提督評分 → 降冪 → 四層挑一位
//	2D76:0576(提督, 類別) 抽類別鍵
//	2D76:0611(提督)      真的編出來（編不成就換下一類）
//
// **一個 AI 回合可以建很多支**：`2D76:0002` 是一個迴圈——閘門過了就
// 呼叫 `2D76:01A5`（配額排序 ＋ 依序試建，成功一類就 break），
// 回非 0（有建成）就再繞一輪，直到閘門擋下來或四類都編不成
// （`docs/re/210`）。收據 `turn260mem` 上同盟第一個回合建了**三支**，
// 第四輪被錢的閘門擋掉（`docs/re/209`）。

// BuildInput 是跑一趟建艦隊要的東西。
type BuildInput struct {
	Faction int
	// CapitalOwner 是自己首都那一格的擁有者。不是自己就整趟不做
	// （`2D76:0002` 的第一道檢查）。
	CapitalOwner int
	// Fleets 是場上所有艦隊；這一趟只數自己那一邊。
	Fleets []rules.Fleet

	// Tax／Upkeep／Tight 是 `2D76:0060` 看的三個量：總稅收
	// （`ds:8F6A[陣營 × 4]`）、艦隊維持費（`ds:8F72[陣營 × 4]`）、
	// `ds:8FD4[陣營] >= 3`（語意未解，效果是預算打七五折）。
	Tax    int
	Upkeep int
	Tight  bool

	// Rand 是原版的亂數。一次類別嘗試會用掉**兩個**：`2D76:02FA` 挑人選
	// 一個、`2D76:0576` 抽類別鍵一個。nil ＝ 一律取 0。
	Rand *rules.Rand

	// Candidates 是候選提督。
	//
	// ⚠ 原版走 `28E0:0103(&清單, 0FFh, 25h, 0)` 收全部提督，而那支清單
	// 產生器的 `0x25` 還沒讀（`docs/internals/54` 的未解清單）。所以
	// **這個過濾目前是 remake 自己決定的**，不是量出來的原版行為。
	Candidates []AdmiralView

	// CanForm ＝ 原版 `07F9:023B(提督)`：那一位所在星域裡、他那個艦種的
	// 艦艇存量夠不夠編一支（`07F9:02CB`，門檻是艦種表 `ds:4E3A` 那一列的
	// 滿編艘數）。nil ＝ 一律可以。
	CanForm func(a AdmiralView) bool

	// Adjutant 回預設副官的提督索引（原版 `ds:5F78[編號]`），-1 ＝ 沒有。
	// nil ＝ 都沒有。
	Adjutant func(admiral int) int
}

func (in BuildInput) next() int {
	if in.Rand == nil {
		return 0
	}
	return in.Rand.Next()
}

// BuildPass 跑一趟階段一，回這一回合要新編的艦隊。
//
// ⚠ **閘門的「錢」那一項要靠呼叫端給對的維持費。** 原版是靠
// `維持費 × 20 >= 總稅收` 收手的（收據 `turn260mem` 建到第三支就停，
// `docs/re/209`）；`Upkeep` 給 0 的話那一項永遠通過，迴圈只會被
// `FleetCap` 與 `CanForm` 擋住——那是與原版不同的行為。
func BuildPass(in BuildInput) []NewFleet {
	if in.CapitalOwner != in.Faction {
		return nil // 首都不是自己的就不建
	}
	// 這一趟建出來的也要算進「現有幾支」與類別計數，下一輪才對；
	// 用掉的提督（與他的副官）也不再是候選。
	var out []NewFleet
	used := map[int]bool{}
	for {
		n := len(out)
		var counts [4]int
		for _, f := range out {
			if c := rules.FleetClass(f.ClassKey); c >= 0 && c < len(counts) {
				counts[c]++
			}
		}
		for _, f := range in.Fleets {
			if f.Faction != in.Faction {
				continue
			}
			n++
			if c := rules.FleetClass(f.ClassKey); c >= 0 && c < len(counts) {
				counts[c]++
			}
		}
		if _, ok := BuildGate(n, in.Tax, in.Upkeep, in.Tight); !ok {
			return out
		}
		built := false
		for _, class := range ClassQuotaOrder(counts, n) {
			nf, ok := in.tryClass(class, used)
			if !ok {
				continue // 這一類編不成 → 換下一類（2D76:01A5 的迴圈）
			}
			out = append(out, nf)
			// 原版靠 `2D76:0611` 清掉提督 `+0x09` 的 bit7，下一輪重建清單
			// 時他就不合格了（`docs/re/208` 的 bit5）；副官也被綁進同一支
			// 艦隊，一樣不再是候選。
			used[nf.Admiral] = true
			if nf.Adjutant >= 0 {
				used[nf.Adjutant] = true
			}
			built = true
			break // 成功一類就跳出，回到 2D76:0002 的外圈
		}
		if !built {
			return out // 四類都編不成 → 2D76:01A5 回 0 → 外圈結束
		}
	}
}

// tryClass ＝ `2D76:02FA(類別)` ＋ `2D76:0576` ＋ `2D76:0611` 的前半。
//
// **亂數照原版的次序用掉**：先挑人選、再抽類別鍵。編不成也已經用掉了，
// 與原版一致（`2D76:0576` 排在 `2D76:0611` 之前）。
func (in BuildInput) tryClass(class int, used map[int]bool) (NewFleet, bool) {
	// 評分 → 降冪（`2D76:0434` ＋ `2D76:04FD`）。用掉的先濾掉。
	ranked := make([]AdmiralView, 0, len(in.Candidates))
	for _, a := range in.Candidates {
		if !used[a.Index] {
			ranked = append(ranked, a)
		}
	}
	if len(ranked) == 0 {
		return NewFleet{}, false
	}
	score := make(map[int]int, len(ranked))
	for _, a := range ranked {
		score[a.Index] = AdmiralBuildScore(a, class, in.Faction, in.Tax)
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		return score[ranked[i].Index] > score[ranked[j].Index]
	})

	k := PickBuildAdmiral(in.next(), len(ranked), func(rank int) int {
		return ranked[rank].Level
	})
	if k < 0 {
		return NewFleet{}, false
	}
	a := ranked[k]
	key := PickClassKey(class, in.next())
	if key < 0 {
		return NewFleet{}, false
	}
	if in.CanForm != nil && !in.CanForm(a) {
		return NewFleet{}, false // 編不成 → 換下一類
	}
	adj := -1
	if in.Adjutant != nil {
		adj = in.Adjutant(a.Index)
	}
	return NewFleet{Admiral: a.Index, Region: Capitals[in.Faction],
		ClassKey: key, Adjutant: adj}, true
}

// BuildSupport 是建艦隊那條路要的、`World` 帶不到的東西。
// `OriginalStrategy.Support` 是 nil 就不建艦隊。
type BuildSupport struct {
	// Money 回那個陣營的總稅收、艦隊維持費，與 `ds:8FD4[陣營] >= 3`。
	Money func(faction int) (tax, upkeep int, tight bool)
	// Admirals 回候選提督（過濾目前 remake-owned，見 `BuildInput.Candidates`）。
	Admirals func(faction int) []AdmiralView
	// CanForm ＝ `07F9:023B`。
	CanForm func(faction int, a AdmiralView) bool
	// Adjutant ＝ `ds:5F78[編號]`。
	Adjutant func(admiral int) int
}

// Build 讓 `OriginalStrategy` 也是一個 `ai.Builder`。
//
// `Support` 是 nil 就回 nil——**不建艦隊是安全的預設**，因為原版那一段
// 要的資料（提督表、錢、艦艇存量）`World` 帶不到。
func (o OriginalStrategy) Build(w World, faction, _ int) []NewFleet {
	s := o.Support
	if s == nil || w.Owner == nil {
		return nil
	}
	in := BuildInput{
		Faction:      faction,
		CapitalOwner: w.Owner(Capitals[faction]),
		Fleets:       w.Fleets,
		Rand:         o.Rand,
	}
	if s.Money != nil {
		in.Tax, in.Upkeep, in.Tight = s.Money(faction)
	}
	if s.Admirals != nil {
		in.Candidates = s.Admirals(faction)
	}
	if s.CanForm != nil {
		in.CanForm = func(a AdmiralView) bool { return s.CanForm(faction, a) }
	}
	in.Adjutant = s.Adjutant
	return BuildPass(in)
}
