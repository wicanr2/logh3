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

// StaffIdle 回一位提督的 `+0x09` bit7：**這位提督閒置**（沒帶艦隊，也沒有
// 專任參謀職）。挑參謀的候選清單看的就是它（`docs/re/268`）。
func StaffIdle(record []byte) bool {
	return len(record) > admiralStaffFlag && record[admiralStaffFlag]&0x80 != 0
}

// TakeStaffPost／LeaveStaffPost 是就任與卸任那兩下寫入（`28E0:05CC` 與
// `2D10:023A`）：就任寫 `0x7F`（清掉 bit7，低七位填 127 這個無效艦隊槽），
// 卸任寫回 `0x80`。收據 `staffpick` 裡指派前後整張提督表就只有這兩個
// 位元組不同。
func TakeStaffPost(record []byte) {
	if len(record) > admiralStaffFlag {
		record[admiralStaffFlag] = 0x7F
	}
}

func LeaveStaffPost(record []byte) {
	if len(record) > admiralStaffFlag {
		record[admiralStaffFlag] = 0x80
	}
}

// SetFleetSlot 是原版 `2D76:0611` 那一下：提督配到艦隊槽時把 `+0x09` 寫成
// 「bit7 清掉、低七位是陣營內的槽號」。副官那一筆寫同一個槽
// （`docs/internals/08`）。
//
// 挑專任參謀的候選清單看的就是 bit7（`docs/re/268`），所以帶隊之後
// 那位提督不會再被挑去坐參謀。
// PickFleetSlot 是原版 `07F9:03AC(提督)`：提督 `+0x09` 低七位那一槽空著就沿用，否則本陣營 16 格裡第一個空的
// （`docs/re/181`）——提督換艦隊不會換槽號。used 回陣營內那一槽是不是已經有艦隊；沒有空槽回 −1。
func PickFleetSlot(record []byte, used func(slot int) bool) int {
	if len(record) > admiralStaffFlag {
		if s := int(record[admiralStaffFlag] & 0x7F); s < FleetsPerFaction && !used(s) {
			return s
		}
	}
	for s := 0; s < FleetsPerFaction; s++ {
		if !used(s) {
			return s
		}
	}
	return -1
}

// FleetsPerFaction 是每個陣營在艦隊表裡的槽數（帝國 0–15、同盟 16–31）。
const FleetsPerFaction = 16

// FleetTableIndex 回陣營內槽號在艦隊表（`ds:C074`）的位置。
func FleetTableIndex(faction, slot int) int {
	if faction == 1 {
		return FleetsPerFaction + slot
	}
	return slot
}

func SetFleetSlot(record []byte, slot int) {
	if len(record) > admiralStaffFlag && slot >= 0 {
		record[admiralStaffFlag] = byte(slot & 0x7F)
	}
}
