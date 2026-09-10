package rules

// 專任參謀的指派（原版 `2D10:0143`／`2D10:0213`，`docs/re/268`）。
//
// 三個專任參謀席位存在 `WORK` 的 `+0x37`（`StaffAdmiral`），值是提督索引。
// 開局那六格是新遊戲初始化寫死的（`0374:028F` 起），之後有兩條路會改：
// 玩家自己任命（`090A:021B`），以及**電腦在自己回合的尾巴重新指派**
// （`2D10:0038` → `0143`／`0213`）。所以那幾格不是常數——796/1 的同盟
// 內政參謀是 60，第一次「實行」之後就換成 63 了。
//
// 這裡只有規則本身；誰在什麼時候呼叫它由上層決定。

const (
	admiralFleetPost = 0x00 // 0xFF ＝ 沒有帶艦隊（現任參謀這一格仍是 0xFF）
	admiralStaffFlag = 0x09 // bit7 ＝ 沒有專任參謀職位（就任清掉、卸任設回）
	admiralRegion    = 0x0A // 所在星域
	admiralField21   = 0x15 // 十進位 +21，名字未知（提督詳細資料沒有印）

	// 候選的等級上限：`28E0:0103` 的第四個參數與 `2D10:0213` 迴圈裡
	// 那道 `cmp es:[si+6], 10h` 是同一個值，兩道重複。
	staffLevelLimit = 0x10
)

// StaffCandidates 回某個陣營可以出任專任參謀的提督索引。
//
// 原版走 `28E0:0103(&清單, 首都星域, 0x65, 0x10)`，旗標展開成四道過濾
// （`docs/re/208`、語意修正見 `docs/re/268`）：同陣營、`+0x00 == 0xFF`
// （沒有帶艦隊）、`+0x09` bit7 有設（沒有專任參謀職位）、`+0x0A` ＝ 首都
// 星域，再加底層那道 `+0x00 != 0xFE`（這一筆不存在）與等級上限。
//
// current 是現任：指派之前原版會先把它寫回「沒有職位」（`2D10:023A`），
// 所以現任自己也在候選裡，可以留任。傳 −1 表示沒有現任。
func StaffCandidates(admirals []byte, faction, capital, current int) []int {
	n := len(admirals) / AdmiralRecord
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		r := admirals[i*AdmiralRecord : (i+1)*AdmiralRecord]
		if r[admiralFleetPost] == 0xFE || r[admiralFleetPost] != 0xFF {
			continue
		}
		if int(r[admiralFaction]) != faction || int(r[admiralRegion]) != capital {
			continue
		}
		// 現任那一格在挑之前已經被寫成「沒有職位」了。
		if i != current && r[admiralStaffFlag]&0x80 == 0 {
			continue
		}
		if int(r[admiralMerit]) >= staffLevelLimit {
			continue
		}
		out = append(out, i)
	}
	return out
}

// PickDomesticStaff 挑內政專任參謀（`2D10:0213`）。
//
// 分數 ＝ 營運（`+0x0E`）× `+0x15`，另外要求 `營運 >= +0x15`；照清單順序
// 掃，**嚴格大於**才換人，所以同分先中先贏。一個候選都沒有就留任。
func PickDomesticStaff(admirals []byte, faction, capital, current int) int {
	best, bestScore := current, 0
	for _, i := range StaffCandidates(admirals, faction, capital, current) {
		r := admirals[i*AdmiralRecord : (i+1)*AdmiralRecord]
		ops, f21 := int(r[admiralOps]), int(r[admiralField21])
		if score := ops * f21; score > bestScore && ops >= f21 {
			best, bestScore = i, score
		}
	}
	return best
}

// PickIntelStaff 挑情報專任參謀（`2D10:0143`）。同一個模子，分數只有營運，
// 也沒有那道比較。
func PickIntelStaff(admirals []byte, faction, capital, current int) int {
	best, bestScore := current, 0
	for _, i := range StaffCandidates(admirals, faction, capital, current) {
		if ops := int(admirals[i*AdmiralRecord+admiralOps]); ops > bestScore {
			best, bestScore = i, ops
		}
	}
	return best
}
