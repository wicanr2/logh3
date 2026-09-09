package rules

// AI 回合的艦隊整補（原版 `2D76:07BA` ＋ `2D76:0A44`，規格
// `docs/spec/031`、`docs/re/227`–`230`）。
//
// **這是 AI 專屬的**：三個入口各只有一個呼叫者，全在戰略 AI 的段裡
// （`docs/re/229`）。玩家補編制走「艦隊編成」的另一條路。

// refitShort 是「缺編」的判準：現有 × 4 < 滿編 × 3（不到七成五）。
func refitShort(g FleetGroup) bool { return g.Ships*4 < g.Full*3 }

// tornSlot2 是拆掉的那一組在排序前被寫進 `[2]` 的值（原版 `0FFh`），
// 排序時它最大，所以那幾組會被推到最後面再截掉。
const tornSlot2 = 0xFF

// MergeUnderStrength ＝ 原版 `2D76:07BA`：把同一個艦種裡不到七成五的
// 那幾組拆掉、艘數還回星域，收斂成整數組。
//
// 第 0 組（旗艦）不參與。回收斂之後的編制。
func MergeUnderStrength(in BuildInput, groups []FleetGroup) []FleetGroup {
	if len(groups) < 2 {
		return groups
	}
	var sum, short [regionShipKinds]int
	for _, g := range groups[1:] {
		if g.Kind < 0 || g.Kind >= regionShipKinds || !refitShort(g) {
			continue
		}
		sum[g.Kind] += g.Ships
		short[g.Kind]++
	}
	torn := 0
	for k := 0; k < regionShipKinds; k++ {
		if short[k] == 0 {
			continue
		}
		full := in.full(k)
		if full <= 0 {
			continue
		}
		// 差多少湊得滿一組；星域的待命部隊夠才算得進去。
		need := full - sum[k]%full
		if in.stock(k) >= need {
			sum[k] += need
		}
		if sum[k]/full == 0 {
			continue // 湊不出一組，這個艦種不動
		}
		for i := 1; i < len(groups); i++ {
			if groups[i].Kind != k || !refitShort(groups[i]) {
				continue
			}
			if groups[i].Kind != shipKindStandalone {
				ReturnRegionShips(in.Region, in.Faction, groups[i].Kind, groups[i].Ships)
			}
			groups[i] = FleetGroup{Kind: tornSlot2, Slot2: tornSlot2, Slot3: torn}
			torn++
		}
	}
	if torn == 0 {
		return groups
	}
	sortGroups(groups)
	return groups[:len(groups)-torn]
}

// findTemplateSlot ＝ 原版 `2D76:0C4F(艦隊, 艦種, 隊形槽)`：
//
//	找到而且艦種相同 → −1（這一格已經是對的）
//	找到但艦種不同   → 那一槽（要覆蓋它）
//	都沒找到         → 0（要新增一組）
func findTemplateSlot(groups []FleetGroup, kind, slot3 int) int {
	for i := 1; i < len(groups); i++ {
		if groups[i].Slot3 != slot3 {
			continue
		}
		if groups[i].Slot2 == kind {
			return -1
		}
		return i
	}
	return 0
}

// victimSlot ＝ 原版 `2D76:0D40` ＋ `2D76:0DB2`：編制滿了要犧牲一組時，
// 挑**隊形槽已經不在樣板前 n 組裡**的那一組（n ＝ 目前組數）。
// 找不到回 −1（原版 `0FFh`）。第 0 組也在候選裡。
func victimSlot(groups []FleetGroup, tpl *[10][2]int) int {
	n := len(groups)
	for i := 0; i < n; i++ {
		used := false
		for j := 0; j < n && j < len(tpl); j++ {
			if tpl[j][1] == groups[i].Slot3 {
				used = true
				break
			}
		}
		if !used {
			return i
		}
	}
	return -1
}

// TopUpComposition ＝ 原版 `2D76:0A44`：照樣板把組數補回 `want`
// （艦隊記錄的 `+0x0E`）。
//
// **樣板給 0 就用旗艦的艦種，而且是無條件的**——與建艦隊那條
// （`0FF4:11F4` 只在旗艦艦種是 1 時才換）不一樣（`docs/re/228`）。
func TopUpComposition(in BuildInput, groups []FleetGroup, want int) []FleetGroup {
	if in.Tables == nil || in.ClassKey < 0 || in.ClassKey >= len(in.Tables.Template) {
		return groups
	}
	tpl := &in.Tables.Template[in.ClassKey]
	for i := 1; i < want && i < len(tpl); i++ {
		a := tpl[i][0]
		if a == 0 {
			a = in.AdmiralKind
		}
		n := findTemplateSlot(groups, a, tpl[i][1])
		if n < 0 {
			continue // 這一格已經是對的
		}
		kind := a
		if in.groups(kind) == 0 {
			if kind = in.substitute(a, groups); kind < 0 {
				continue
			}
		}
		target := n
		switch {
		case n != 0:
			// 覆蓋：先把原來那一組還掉。
		case len(groups) >= want:
			if target = victimSlot(groups, tpl); target < 0 {
				continue
			}
		default:
			groups = append(groups, FleetGroup{})
			target = len(groups) - 1
		}
		if old := groups[target]; old.Full > 0 && old.Kind != shipKindStandalone &&
			old.Kind >= 0 && old.Kind < regionShipKinds {
			ReturnRegionShips(in.Region, in.Faction, old.Kind, old.Ships)
		}
		groups[target] = in.take(kind, kind, tpl[i][1])
		sortGroups(groups)
	}
	return groups
}

// Refit 跑一次完整的整補：先併殘部，再補回應有組數。
func Refit(in BuildInput, groups []FleetGroup, want int) []FleetGroup {
	return TopUpComposition(in, MergeUnderStrength(in, groups), want)
}
