package rules

import "sort"

// 自家星域的艦艇補充（`1AAB:0FAF`，#71、`docs/re/346`、規格 063）。
//
// 逐格掃描（`0374:0556`）到這一格時的第二步，排在耗損／全滅之前：擁有者停在
// 這一格的艦隊，逐艦種從那一格的待命存量補到編制的滿編為止。

// resupplyRepairKind 是修護艦。`1AAB:1286` 每回合把它的飛彈那一格重設成
// 現有艘數——艦種表的兩個係數都是 0，建立時是 0（`docs/re/345`）。
const resupplyRepairKind = 6

// ResetRepairShips ＝ `1AAB:1235` → `1AAB:1286`：把停在這一格的**所有陣營**
// 艦隊裡的修護艦，飛彈那一格重設成現有艘數。原版在補充的最前面跑。
func ResetRepairShips(fleets []*Fleet, region int) {
	for _, f := range fleets {
		if f == nil || f.Region != region {
			continue
		}
		for i := range f.Composition {
			g := &f.Composition[i]
			if g.Kind != resupplyRepairKind {
				continue
			}
			g.Missiles = g.Ships
		}
	}
}

// resupplyCandidate 是排序用的一組編制。
type resupplyCandidate struct {
	fleet *Fleet
	group *FleetGroup
	slot  int // 分隊序號（編制陣列裡的位置）＝ 艦艇記錄的 `+0x03`
}

// ResupplyRegion ＝ `1AAB:0FAF`：這一格的補充。回**擁有者停在這一格的艦隊**，
// 呼叫端照原版逐支跑 `1261:0067(f, 1)`（士氣重算 ＋ 當場印報告）。
//
// 無主星域（擁有者 2）直接回 nil，什麼都不做。
//
// 挑艦隊只看「所在星域」與「陣營」，**不看移動狀態**（`297B:13C3`／`142A`
// 兩道濾網都沒有碰 `+0x08`）——回合跑到這一步時移動階段已經結束，所以條件
// 實際上是「這一回合結束時人在這一格的擁有者艦隊」。
func ResupplyRegion(record []byte, planets *Planets, region int, fleets []*Fleet,
	kinds *ShipKindTable) []*Fleet {
	if len(record) < RegionRecord {
		return nil
	}
	owner := int(record[regionOwner])
	if owner < 0 || owner > 1 {
		return nil // 無主星域直接回
	}
	ResetRepairShips(fleets, region)

	var mine []*Fleet
	for _, f := range fleets {
		if f != nil && f.Region == region && f.Faction == owner {
			mine = append(mine, f)
		}
	}
	// 原版在這裡先跑 `2703:12EB`：它把星域的 `+0x30`／`+0x68`／`+0x70` 清掉再從
	// 每顆行星重算（`docs/re/349` §2.3）。remake 已經在每回合開頭的
	// `refillRegionSupply` 做過同一件事（`2B5F:02B2` ＋ `2703:0590`），在這裡
	// 再做一次會把行星存量整批重分配，所以略過。
	// （收據上這一支前後的位元組相同，那只是因為重算的結果與原值相同，
	// 不是因為它什麼都沒做——規格 063 原本寫錯了理由。）
	gave := false
	for kind := 0; kind < regionShipKinds; kind++ {
		cands := resupplyCandidates(mine, kind)
		if len(cands) == 0 {
			continue
		}
		sortResupplyCandidates(cands)
		for _, c := range cands {
			if regionShipStock(record, owner, kind) == 0 {
				break // 這個艦種的存量見底，後面排隊的拿不到
			}
			need := c.group.Full - c.group.Ships
			if need <= 0 {
				continue
			}
			give := TakeRegionShips(record, owner, kind, need)
			if give <= 0 {
				continue
			}
			addGroupShips(kinds, c.fleet.Faction, c.group, give)
			gave = true
		}
	}
	// `2703:0540`：扣完之後把這一格的存量照經濟力比例分回行星（兩陣營各一趟），
	// 扣掉的量才會進到下一回合的基數。**沒有真的補到東西就不跑**——分配本身
	// 會重寫行星的存量欄位，白跑一趟等於平白把行星之間的分布洗一次。
	if gave {
		for faction := 0; faction < 2; faction++ {
			DistributeRegionSupply(record, planets, region, faction)
		}
	}
	return mine
}

// addGroupShips 把 give 艘加進一組編制，並照原版連帶更新修理上限與兩格存量。
//
//   - `+0x08` 現有艘數 ＋= give；
//   - `+0x0C` 修理上限 ＋= min(give, 滿編 − 修理上限)；
//   - `+0x0E`／`+0x12` ＋= 艦種表係數 × give（`2703:14CF`，`docs/re/345`）。
//
// 收據上兩格各多的那一塊 108600 是**排在前面**的 `SupplyRegionFleets`
// （`1930:1572`）補的，不屬於這一支（`docs/re/349` §三逐值對過）。
// 這裡只加係數那一項。
func addGroupShips(kinds *ShipKindTable, faction int, g *FleetGroup, give int) {
	g.Ships += give
	if room := g.Full - g.RepairCap; room > 0 {
		if give < room {
			g.RepairCap += give
		} else {
			g.RepairCap = g.Full
		}
	}
	k, ok := kinds.Kind(faction, g.Kind)
	if !ok {
		return
	}
	g.Missiles += k.Byte(kindMissilePer) * give
	g.Fighters += k.Byte(kindFighterFull) * give
}

// resupplyCandidates ＝ `1AAB:13BE`：這些艦隊裡艦種等於 kind 的編制組。
// **不篩掉已經滿編的**（那些算出來的 give 是 0），也碰不到索敵分遣隊
// （它不在編制陣列裡）。
func resupplyCandidates(fleets []*Fleet, kind int) []resupplyCandidate {
	var out []resupplyCandidate
	for _, f := range fleets {
		for i := range f.Composition {
			if f.Composition[i].Kind != kind {
				continue
			}
			out = append(out, resupplyCandidate{fleet: f, group: &f.Composition[i], slot: i})
		}
	}
	return out
}

// sortResupplyCandidates ＝ `1AAB:145F` 的比較器：旗艦那一組先拿，其次越殘破
// 的先補，最後用分隊序號 tie-break。（原版還比一次陣營，同一次呼叫裡必然
// 相同，是死代碼。）
func sortResupplyCandidates(c []resupplyCandidate) {
	sort.SliceStable(c, func(i, j int) bool {
		if a, b := c[i].slot == 0, c[j].slot == 0; a != b {
			return a
		}
		if c[i].group.Ships != c[j].group.Ships {
			return c[i].group.Ships < c[j].group.Ships
		}
		return c[i].slot < c[j].slot
	})
}

// regionShipStock 回那一格某陣營某艦種的待命存量（`+0x30 ＋ 陣營×28 ＋ 艦種×4`）。
func regionShipStock(record []byte, faction, kind int) int {
	if faction < 0 || faction > 1 || kind < 0 || kind >= regionShipKinds {
		return 0
	}
	o := regionShips + faction*regionShipKinds*4 + kind*4
	if o+4 > len(record) {
		return 0
	}
	return int(le32(record[o:]))
}

// SupplyRegionFleets ＝ `1930:1572` → `1930:15F4`（`docs/re/349` §二）：把這一格的
// 飛彈與艦載機存量（星域 `+0x68`／`+0x70`）補進擁有者艦隊的每一組編制。
//
// 排在艦艇補充（`ResupplyRegion`）**之前**，逐格掃描的序列是
// `1930:1572` → `1AAB:0FAF` → `0374:0608`（`docs/internals/62`）。
//
// 規則（都照原版）：
//
//   - 挑艦隊的條件與艦艇補充相同（所在星域 ＋ 陣營，`297B:131C` 同一支同旗標）；
//   - 逐組取 `min(星域存量, 補到滿的缺口)`，**先到先拿**——前面的組拿光，後面的拿 0；
//   - **不篩艦種**，補給艦與修護艦也一起補；碰不到索敵分遣隊（只走編制陣列）。
//
// **一處刻意偏離**（使用者裁決 2026-09-19，`docs/re/349` §四）：原版算缺口用帶符號
// 減、卻用無號比較，所以目標超編時缺口會變成接近 2³²、把整格存量抽乾。這裡
// **缺口為負就當 0**（不補）。代價是玩家局面真的走到「某一組超編」時會與原版分岔。
func SupplyRegionFleets(record []byte, region int, fleets []*Fleet, kinds *ShipKindTable) {
	if len(record) < RegionRecord {
		return
	}
	owner := int(record[regionOwner])
	if owner < 0 || owner > 1 {
		return // 無主星域
	}
	for _, f := range fleets {
		if f == nil || f.Region != region || f.Faction != owner {
			continue
		}
		for i := range f.Composition {
			g := &f.Composition[i]
			k, ok := kinds.Kind(f.Faction, g.Kind)
			if !ok {
				continue
			}
			supplyOne(record, owner, regionMissiles, &g.Missiles,
				k.Byte(kindMissilePer)*g.Ships)
			supplyOne(record, owner, regionFighters, &g.Fighters,
				k.Byte(kindFighterFull)*g.Ships)
		}
	}
}

// supplyOne 把一格存量補到 have 上，補到 full 為止；回補了多少。
// 缺口為負時當 0——不照抄原版那個無號比較的溢位（見 SupplyRegionFleets）。
func supplyOne(record []byte, faction, off int, have *int, full int) {
	need := full - *have
	if need <= 0 {
		return
	}
	o := off + faction*4
	if o+4 > len(record) {
		return
	}
	stock := int(le32(record[o:]))
	give := need
	if stock < give {
		give = stock
	}
	if give <= 0 {
		return
	}
	*have += give
	putLE32(record[o:], stock-give)
}
