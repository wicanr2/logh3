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

// WorkStaffOffset 是 `WORK` 區塊裡三個幕僚職的起點（`ds:8F4B`）：
// 一職兩個位元組，帝國那位的提督索引在前。
const WorkStaffOffset = workStaffOff

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
	"少尉", "中尉", "上尉", "少校", "中校", "上校", "准將", "少將", "中將", "上將", "一級上將", "元帥",
}

// rankSteps 是每一階的下限（`+0x06`）。
//
// 原版28E0:004F的12級下限，已核原始指令與兩陣營字串（規格108）。
var rankSteps = []int{0, 2, 4, 6, 8, 12, 16, 20, 24, 29, 34, 39}

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
func (a Admiral) Rank() string {
	rank := RankOf(a.Merit)
	if a.Faction == 1 && rank == 10 {
		return RankNames[9]
	}
	return RankNames[rank]
}

// AdmiralList 回某個陣營的提督，照提督表的順序。
//
// standby ＝ 只要**待命中**的——「出擊命令」那張「待命中提督資料」用這個：
// 幕僚職與**已經帶隊（出擊中）的**都濾掉（`+0x09` bit7 沒設就是有職務，
// `AdmiralDutyOf`）。「提督資料」那張不濾（幕僚職照列，只是前面標一個 ○，
// 出擊中的在現在位置後面標「：出擊中」）。
//
// 「出擊中的也要濾掉」是收據 `fcx-r8-65-standby` 量到的：第１艦隊編成之後
// 再開這張清單，萊因哈特不在上面，第一列是吉爾菲艾斯（#65）。
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
		if standby && (isStaff[a.Index] || AdmiralDutyOf(a.Record) == DutySortie) {
			continue
		}
		out = append(out, a)
	}
	return out
}

// AdmiralLevel 回一筆提督記錄的等級（`+0x06`）。軍階由它分段推出來
// （`RankOf`）；提督詳細資料那張畫面把它標成「等級」（`docs/re/65`）。
func AdmiralLevel(record []byte) int {
	if len(record) <= admiralMerit {
		return 0
	}
	return int(record[admiralMerit])
}

// AdmiralFaction 回一筆提督記錄的陣營（`+0x01`：0 帝國、1 同盟）。
func AdmiralFaction(record []byte) int {
	if len(record) <= admiralFaction {
		return 0
	}
	return int(record[admiralFaction])
}

// FlagshipSlot 回一位提督在自己陣營裡的序號——旗艦表就是照這個索引的
// （帝國 ＝ 提督索引、同盟 ＝ 提督索引 − 40，`docs/re/65`）。
//
// 沒有艦隊的提督（記錄 `+0x07` ＝ `0xFF`）畫面上不印旗艦那一行，回 -1。
func FlagshipSlot(record []byte, index int) int {
	if len(record) <= admiralFleet || record[admiralFleet] == 0xFF {
		return -1
	}
	if AdmiralFaction(record) == 0 {
		return index
	}
	return index - AllianceFirstAdmiral
}

// FactionAdmiralIndex 回一位提督在自己陣營裡的序號（帝國 ＝ 提督索引、同盟 ＝
// 提督索引 − 40），不看有沒有艦隊——旗艦名表 `ds:6428` 就是照它索引的。
func FactionAdmiralIndex(record []byte, index int) int {
	if AdmiralFaction(record) == 0 {
		return index
	}
	return index - AllianceFirstAdmiral
}

// AllianceFirstAdmiral 是同盟那一段提督的起點（`+0x01` 在 0–39 是 0、
// 40–81 是 1，`docs/internals/08`）。
const AllianceFirstAdmiral = 40

// admiralFleet 是提督記錄的 `+0x07`：`0xFF` ＝ 沒有艦隊。
const admiralFleet = 0x07

// Adjutant 回這位提督的副官是第幾位提督（記錄 `+0x07`）；`0xFF` 回 -1。
//
// 同一個欄位也決定他有沒有旗艦、能不能單獨帶艦隊（`docs/re/66`）。
func Adjutant(record []byte) int {
	if len(record) <= admiralFleet || record[admiralFleet] == 0xFF {
		return -1
	}
	return int(record[admiralFleet])
}

// admiralOps 是提督記錄的 `+14`（營運）。
const admiralOps = 14

// AdmiralOps 回第 i 位提督的營運（`ADMIRAL` `+14`）。取不到回 0。
//
// 兩個消費者看的是不同的提督：艦隊維持費與移動經費看**那支艦隊的**
// （`Fleet.Operations`，`docs/re/218`）；船艦建造的單價看**軍事專任參謀的**
// （`docs/spec/037`）。
func AdmiralOps(data []byte, i int) int {
	if i < 0 || (i+1)*AdmiralRecord > len(data) {
		return 0
	}
	return int(data[i*AdmiralRecord+admiralOps])
}

// StaffOps 回某個陣營坐某個幕僚職那位提督的營運。取不到回 0。
func StaffOps(work, admiral []byte, post, faction int) int {
	return AdmiralOps(admiral, StaffAdmiral(work, post, faction))
}

// SetStaffAdmiral 把某個陣營的某個幕僚職換人（原版 `2D10:02ED` 那類寫入）。
// 位址超出範圍就什麼都不做。
func SetStaffAdmiral(work []byte, post, faction, admiral int) {
	if post < 0 || post >= workStaffCount || faction < 0 || faction > 1 {
		return
	}
	if len(work) < workStaffOff+workStaffCount*2 || admiral < 0 || admiral > 0xFF {
		return
	}
	work[workStaffOff+post*2+faction] = byte(admiral)
}

// AdjutantCandidates 回可以當副官的提督索引（照提督表的順序）＝ 原版
// `28E0:0103(&清單, 提督[+0x0A], 0x65, 提督[+0x06])`（`docs/re/318`）：
//
//	同一個星域（+0x0A）、同陣營（+0x01）、沒有職務（+0x00 == 0xFF）、
//	閒置（+0x09 bit7）、等級（+0x06）比帶隊提督低
//
// 「閒置」這一項不讀 `+0x09` bit7：remake 的記錄在職幕僚仍帶 bit7
// （原版在職者是 `0x7F`），載入存檔時也可能被夾掉，所以由呼叫端用
// `busy` 回答「這一位在職或在艦隊裡」（幕僚職、艦隊的提督與副官）。
func AdjutantCandidates(all []Admiral, admiral []byte, busy func(int) bool) []int {
	if len(admiral) <= admiralRegion {
		return nil
	}
	faction, region, level := int(admiral[admiralFaction]), int(admiral[admiralRegion]), int(admiral[admiralMerit])
	out := make([]int, 0, len(all))
	for _, a := range all {
		r := a.Record
		if a.Faction != faction || len(r) <= admiralRegion || int(r[admiralRegion]) != region {
			continue
		}
		if r[0] != 0xFF || int(r[admiralMerit]) >= level {
			continue
		}
		if busy != nil && busy(a.Index) {
			continue
		}
		out = append(out, a.Index)
	}
	return out
}

// SetAdjutant 把提督記錄的副官（`+0x07`）換成 adj（-1 ＝ 卸任）＝ 原版
// `07F9:085D` 的副官那一支：舊副官的 `+0x09` 寫 0x80（閒置）、新副官的
// `+0x09` 抄帶隊提督的（同一個艦隊槽）。old／new 是那兩位的記錄，可以是 nil。
func SetAdjutant(admiral, old, new []byte, adj int) {
	if len(admiral) <= admiralStaffFlag {
		return
	}
	if len(old) > admiralStaffFlag {
		old[admiralStaffFlag] = 0x80
	}
	if adj < 0 {
		admiral[admiralFleet] = 0xFF
		return
	}
	if len(new) > admiralStaffFlag {
		new[admiralStaffFlag] = admiral[admiralStaffFlag]
	}
	admiral[admiralFleet] = byte(adj)
}

// IdleAdmiralsInRegion 回某一格裡閒置的提督索引（照提督表的順序）＝ 原版
// `28E0:0103(&清單, 星域, 0x25, 0)`（`docs/re/208`、`docs/re/319`）：
//
//	同陣營（+0x01）、沒有職務（+0x00 == 0xFF）、沒帶艦隊（+0x09 bit7）、
//	所在星域（+0x0A）是那一格
//
// 「沒帶艦隊」與 `AdjutantCandidates` 一樣交給 busy 回答（理由見那一支）。
func IdleAdmiralsInRegion(all []Admiral, faction, region int, busy func(int) bool) []int {
	out := make([]int, 0, len(all))
	for _, a := range all {
		r := a.Record
		if a.Faction != faction || len(r) <= admiralRegion || int(r[admiralRegion]) != region {
			continue
		}
		if r[0] != 0xFF {
			continue
		}
		if busy != nil && busy(a.Index) {
			continue
		}
		out = append(out, a.Index)
	}
	return out
}

// ListPage 是右下角清單那種「五列一頁」的捲動狀態（`docs/re/319`）：
// Count 筆、從 Scroll 起印五列。
type ListPage struct {
	Count, Scroll int
}

// ListPageRows 是右下角清單一頁幾列。
const ListPageRows = 5

// Clamp 照原版每次印之前的夾法：`位移 ＋ 5 > 筆數` 就退到
// `筆數 − 5`（不足五筆就 0）。
func (p *ListPage) Clamp() {
	if p.Scroll+ListPageRows > p.Count {
		if p.Count >= ListPageRows {
			p.Scroll = p.Count - ListPageRows
		} else {
			p.Scroll = 0
		}
	}
	if p.Scroll < 0 {
		p.Scroll = 0
	}
}

// Step 是 ▲（dir −1）／▼（＋1）的一列更新，再夾回範圍。
func (p *ListPage) Step(dir int) {
	p.Scroll += dir
	p.Clamp()
}

// Arrows 回這一頁要不要印 ▲▼（原版只看筆數 > 5）。
func (p ListPage) Arrows() bool { return p.Count > ListPageRows }

// AdmiralRegion 回一筆提督記錄的所在星域（`+0x0A`）；記錄不完整回 −1。
func AdmiralRegion(record []byte) int {
	if len(record) <= admiralRegion {
		return -1
	}
	return int(record[admiralRegion])
}

// AdmiralDuty 是提督詳細資料那張把 `+0x09` 分成的三種（`0D92:00B7`）：
// bit7 有設 ＝ 待命中（沒帶艦隊）、`0x7F` ＝ 執勤中（幕僚職）、其餘 ＝ 出擊中
// （低七位是帶的艦隊槽）。
type AdmiralDuty int

const (
	DutyIdle   AdmiralDuty = iota // 待命中
	DutyStaff                     // 執勤中（參謀）
	DutySortie                    // 出擊中
)

// AdmiralDutyOf 回這一筆的三種之一。
func AdmiralDutyOf(record []byte) AdmiralDuty {
	if len(record) <= admiralStaffFlag {
		return DutyIdle
	}
	switch v := record[admiralStaffFlag]; {
	case v&0x80 != 0:
		return DutyIdle
	case v == 0x7F:
		return DutyStaff
	}
	return DutySortie
}

// AdmiralAlive 回這一筆的 `+0x00`：0 ＝ 死亡、`0x80` ＝ 下落不明，其餘活著
// （詳細資料的「現在位置」印的就是這三種，`0D92:00B7`）。
func AdmiralAlive(record []byte) (dead, missing bool) {
	if len(record) == 0 {
		return false, false
	}
	return record[0] == 0, record[0] == 0x80
}
