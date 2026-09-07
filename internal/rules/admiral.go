package rules

// 提督表（`ADMIRAL.DT[89AX]`，`docs/internals/08`）：82 筆 × 28 bytes，沒有檔頭。
//
// 一覽表那幾張清單要的是三件事：名字、軍階、以及「在不在幕僚職上」。
// 名字不在提督表裡（在執行檔的資料段，整理成
// `docs/internals/data/keireki.json`），軍階與幕僚旗標都算得出來。

const (
	admiralMerit = 0x06 // 軍階由這個值分段換算

	// 幕僚職不在提督表裡，在 WORK 的 `+0x37`：三個職位各兩個位元組，
	// **帝國那位的提督索引在前、同盟那位在後**（`docs/internals/30`）。
	workStaffOff   = 0x37
	workStaffCount = 3
)

// StaffAdmirals 回某個陣營三個幕僚職的提督索引。
func StaffAdmirals(work []byte, faction int) []int {
	out := make([]int, 0, workStaffCount)
	for post := 0; post < workStaffCount; post++ {
		if i := StaffAdmiral(work, post, faction); i >= 0 {
			out = append(out, i)
		}
	}
	return out
}

// 三個幕僚職的順序（`docs/internals/30`）。`WORK` 的 `+0x37` 起
// **一職兩個位元組：帝國那位的提督索引、同盟那位的**。
const (
	PostMilitary = 0 // 軍事專任參謀：月底的船艦接收報告、次月份建造計劃署名
	PostIntel    = 1 // 情報專任參謀：戰略情報報告署名
	PostDomestic = 2 // 內政專任參謀：月初的會計報告署名
)

// StaffAdmiral 回某個陣營坐某個幕僚職的提督索引，取不到就回 −1。
func StaffAdmiral(work []byte, post, faction int) int {
	if post < 0 || post >= workStaffCount || faction < 0 || faction > 1 {
		return -1
	}
	if len(work) < workStaffOff+workStaffCount*2 {
		return -1
	}
	return int(work[workStaffOff+post*2+faction])
}

// RankNames 是軍階由低到高。`RankOf` 回的是這個陣列的索引。
var RankNames = []string{
	"少尉", "中尉", "上尉", "少校", "中校", "上校", "准將", "少將", "中將", "上將",
}

// rankSteps 是每一階的下限（`+0x06`）。
//
// 量到的對應（27 個資料點，收據 `lst2e`／`sortie1e`／`s3sortie`）：
// 4→上尉、6→少校、12→上校、15→上校、16→准將、20→少將、24→中將、29→上將。
// 15 與 12 同階、29 以上都是上將，所以是**分段**不是查表。
var rankSteps = []int{0, 2, 4, 6, 8, 12, 16, 20, 24, 28}

// Admiral 是提督表一筆解出來的樣子。
type Admiral struct {
	Index   int
	Faction int
	Merit   int
	Record  []byte
}

// ParseAdmirals 解整份提督表。
func ParseAdmirals(data []byte) []Admiral {
	n := len(data) / AdmiralRecord
	out := make([]Admiral, 0, n)
	for i := 0; i < n; i++ {
		r := data[i*AdmiralRecord : (i+1)*AdmiralRecord]
		out = append(out, Admiral{
			Index:   i,
			Faction: int(r[admiralFaction]),
			Merit:   int(r[admiralMerit]),
			Record:  r,
		})
	}
	return out
}

// RankOf 把 `+0x06` 換成軍階索引（`RankNames` 的下標）。
func RankOf(merit int) int {
	rank := 0
	for i, lo := range rankSteps {
		if merit >= lo {
			rank = i
		}
	}
	return rank
}

// Rank 回這名提督的軍階名稱。
func (a Admiral) Rank() string { return RankNames[RankOf(a.Merit)] }

// AdmiralList 回某個陣營的提督，照提督表的順序。
//
// standby ＝ 只要待命中的（把幕僚職濾掉）——「出擊命令」那張清單用這個，
// 「提督資料」那張不濾（幕僚職照列，只是前面標一個 ○）。
func AdmiralList(all []Admiral, faction int, standby bool, staff []int) []Admiral {
	isStaff := make(map[int]bool, len(staff))
	for _, i := range staff {
		isStaff[i] = true
	}
	out := make([]Admiral, 0, len(all))
	for _, a := range all {
		if a.Faction != faction {
			continue
		}
		if standby && isStaff[a.Index] {
			continue
		}
		out = append(out, a)
	}
	return out
}
