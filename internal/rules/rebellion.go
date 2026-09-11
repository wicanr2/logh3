package rules

import "encoding/binary"

// 叛亂（`docs/spec/041`、`docs/re/292`、`docs/re/297`）。
//
// 原版一次「實行」裡分兩段：
//
//	056A:0469 那一圈   每一格判一次（2703:1789），叛亂的設 +0x13 bit1
//	0374:03E8 那一圈   逐格讀 bit1，設了就跑效果（2703:1908）再清掉
//
// 判定全部跑完才輪到效果，所以這裡拆成 `DecideRebellions`（整圈判定）
// 與 `Rebel`（一格的效果）兩支，呼叫端先判完 32 格再逐格套效果。

// 叛亂用到的幾個欄位與常數。
const (
	// rebellionRollMod 是 `2703:1789` 的 `idiv 190h`：亂數 % 400。
	rebellionRollMod = 400
	// RebellionDraws 是**一格固定抽幾次亂數**：第 1 次當 r1、第 6 次當 r2，
	// 其餘八次抽了就丟（`2703:181E`–`185D` 十個 `call far 3CD5:000C`）。
	// 少抽一次，後面每一格、以及這一圈之後每一個用到亂數的規則都會錯位。
	RebellionDraws = 10
)

// RebellionThreshold 回 `2703:1789` 拿 `r1 ＋ r2` 去比的那個門檻。
//
//	f = 未安定度 < 50  ? u ÷ 3
//	  : 未安定度 < 100 ? u ÷ 2
//	  : 未安定度 < 150 ? u × 3 ÷ 4
//	  :                  u × 4 ÷ 3
//	門檻 = f × 100 ÷ 營運 ÷ (已失格數 ＋ 1)
//
// 全部是無號 16 位元整數除法（`div`，往下取整）。`f × 100` 原版是
// `mul` 之後把 DX 清掉再 `div`，只留低 16 位元；未安定度是 byte，
// f 最大 340，乘 100 是 34000，不會溢位，所以這裡不必模擬截斷。
//
// 營運為 0 時原版的 `div` 會觸發除以零（INT 0）。出貨資料裡沒有營運 0 的
// 提督，remake 在這裡回 0（＝ 不叛），不讓程式當掉。
func RebellionThreshold(unrest, ops, lost int) int {
	u := unrest & 0xFF
	var f int
	switch {
	case u < 50:
		f = u / 3
	case u < 100:
		f = u / 2
	case u < 150:
		f = u * 3 / 4
	default:
		f = u * 4 / 3
	}
	if ops <= 0 {
		return 0
	}
	return f * 100 / ops / (lost + 1)
}

// LostRegions 是 `2703:09DD`：各陣營**從開局到現在**丟掉了幾格。
//
//	對每一格：現在的擁有者 == 2 → 跳過
//	          開局的擁有者 == 2 → 跳過
//	          開局 == 現在     → 跳過
//	          否則 lost[開局的擁有者]++
//
// `opening` 是原版的 `ds:B75C`：32 格的擁有者快照，唯一的寫入點是
// 新遊戲初始化（`2703:0068`）。**讀存檔不會重寫它**（`docs/re/297`），
// 所以它永遠是 796 年 1 月開局那一份，不是「這一局載入當下」的。
//
// 被翻成無主的格子不算「丟掉」——只有落到對方手上的才算。
func LostRegions(records, opening []byte) [2]int {
	var lost [2]int
	for i := 0; i < RegionCount && i < len(opening); i++ {
		o := i * RegionRecord
		if o+RegionRecord > len(records) {
			break
		}
		now := int(records[o+regionOwner])
		was := int(opening[i])
		if now == RegionOwnerNone || was == RegionOwnerNone || was == now {
			continue
		}
		if was == 0 || was == 1 {
			lost[was]++
		}
	}
	return lost
}

// RebellionInput 是 `056A:0469` 那一圈判叛亂時讀到的狀態。
type RebellionInput struct {
	// Records 是 32 格星域記錄：讀 `+0x12`（擁有者）與 `+0x7D`（未安定度）。
	Records []byte
	// Opening 是 `ds:B75C`（見 `LostRegions`）。
	Opening []byte
	// Occupation 是同一圈先算的 bit2／bit4：兩個都沒設的格子才會問
	// `2703:1789`（`docs/re/245`）。
	Occupation [RegionCount]RegionOccupation
	// Garrison 是各格兩個陣營的艦隊數，＝ 星域記錄 `+0x28`／`+0x29`。
	// **要用這一圈之前剛重數的那一份**：`056A:000C` 在 `056A:0469` 的
	// 前一步呼叫 `2703:0853`，照艦隊移動之後的位置重數（`docs/re/297`）。
	Garrison [RegionCount][2]int
	// Ops 是 `ds:8F4D[陣營]`（情報專任參謀）那位提督的營運 `+0x0E`。
	Ops [2]int
	// Rand 是原版的全域亂數（`ds:94DC`），判定會就地往前走。
	Rand *Rand
}

// RebellionRoll 是一格的判定結果（給測試與報告看）。
type RebellionRoll struct {
	Region    int
	Asked     bool // 有沒有進到 `2703:1789` 的亂數那一段
	Threshold int
	Roll      int // r1 ＋ r2
	Rebels    bool
}

// DecideRebellions 跑一整圈 `056A:0469` 的 bit1 那一條，回每一格的結果。
//
// 一格要過四道閘門才擲亂數（`2703:1789`）：
//
//	bit2 或 bit4 有設（這一格有戰鬥、或被單方面占據）→ 根本不問
//	編號 0（瓦爾哈拉）或 31（巴拉特）              → 不叛
//	擁有者 == 2（無主）                            → 不叛
//	這一格有擁有者自己的艦隊（+0x28 ＋ 擁有者 != 0）→ 不叛
//
// 過關之後照順序抽十次亂數（第 1 次、第 6 次各 % 400），`r1 ＋ r2 < 門檻`
// 就叛亂。首都判斷看的是**編號**，不看擁有者——首都被對方拿下也一樣不問。
//
// `已失格數` 在整圈開頭算一次（`2703:09DD`），這一圈裡不會因為前面的
// 格子叛亂而改變——叛亂的效果要等下一圈才套上去。
func DecideRebellions(in RebellionInput) [RegionCount]RebellionRoll {
	var out [RegionCount]RebellionRoll
	lost := LostRegions(in.Records, in.Opening)
	for i := 0; i < RegionCount; i++ {
		out[i].Region = i
		o := i * RegionRecord
		if o+RegionRecord > len(in.Records) {
			break
		}
		if in.Occupation[i].Contested || in.Occupation[i].Occupied {
			continue
		}
		if i == 0 || i == 31 {
			continue
		}
		owner := int(in.Records[o+regionOwner])
		if owner == RegionOwnerNone {
			continue
		}
		// 原版查 `+0x28 ＋ 擁有者` 那個位元組；擁有者只會是 0／1／2，
		// 2 已經在上一道擋掉了。
		if owner < 0 || owner > 1 || in.Garrison[i][owner] != 0 {
			continue
		}
		r1 := in.Rand.Next() % rebellionRollMod
		in.Rand.Advance(4)
		r2 := in.Rand.Next() % rebellionRollMod
		in.Rand.Advance(4)
		t := RebellionThreshold(int(in.Records[o+RegionUnrest]), in.Ops[owner], lost[owner])
		out[i].Asked = true
		out[i].Threshold = t
		out[i].Roll = r1 + r2
		out[i].Rebels = r1+r2 < t // 無號比較；兩邊都不會是負的
	}
	return out
}

// RebelPlanets 回 `2703:1908` 會翻切片裡的哪幾顆（切片序號，照呼叫
// `2703:1ADB` 的先後）。
//
//	全部旗標 != 0        → 整格
//	未安定度 > 160        → 整格
//	未安定度 > 120        → 前兩顆（只有一顆就一顆）
//	未安定度 > 80         → 第 0 顆
//	其餘                  → 最後一顆
//
// **原版在這之前把行星依經濟力遞減排過序，但兩條迴圈都沒用那個結果**
// （`docs/re/292`「排序排了沒有用」）：迴圈走的是 `ds:CFF8` 那個原始
// 切片，排好的指標陣列一個位元組都沒被讀。所以翻的是切片的第 0 顆或
// 最後一顆，不是最有錢的那顆——這是 1994 年那份原始碼的 bug，照原樣
// 保存。
func RebelPlanets(unrest, count int, all bool) []int {
	if count <= 0 {
		return nil
	}
	n, forward := 1, true
	u := unrest & 0xFF
	switch {
	case all || u > 160:
		n = count
	case u > 120:
		if count >= 2 {
			n = 2
		}
	case u <= 80:
		forward = false
	}
	out := make([]int, 0, n)
	for k := 0; k < n; k++ {
		if forward {
			out = append(out, k)
		} else {
			out = append(out, count-1-k)
		}
	}
	return out
}

// Rebel 跑 `2703:1908(星域, 全部旗標)`：星域 `+0x12` 寫成無主，再照
// `RebelPlanets` 挑的那幾顆各跑一次 `RebelPlanet`。
//
// 每回合的叛亂 `全部旗標` 是 0；費沙事件那一個呼叫端傳 1（整格脫離，
// `docs/re/292` 末節，remake 還沒有費沙事件）。
//
// `+0x12 = 2` 是暫時的：回合尾巴會依行星重算擁有者（`2703:159B`）。
// 開局那種「一格全是同一方」的星域掉一顆之後混雜，重算出來也是 2，
// 所以**不會**另外跳一則易主報告（舊值已經是 2）。
func Rebel(record []byte, planets *Planets, region int, all bool) []int {
	if len(record) <= RegionUnrest || planets == nil {
		return nil
	}
	record[regionOwner] = RegionOwnerNone
	start, count := planets.Range(region)
	picked := RebelPlanets(int(record[RegionUnrest]), count, all)
	out := make([]int, 0, len(picked))
	for _, k := range picked {
		if p := planets.Record(start + k); p != nil {
			RebelPlanet(p)
			out = append(out, start+k)
		}
	}
	return out
}

// RebelPlanet 是 `2703:1ADB`：一顆行星脫離。
//
//	擁有者 +0x11 ← 2
//	經濟力 +0x1A −= 經濟力 ÷ 3                      （32 位元）
//	防衛力 +0x14：若 (防衛力 × 3) 的低 16 位元 < 上限 +0x16 → 0
//	              否則 −= 上限 ÷ 3
//
// 防衛力那一段是 16 位元無號：`mul cx` 之後只拿 AX 去比（`cmp ax,dx /
// jb`），所以 `防衛力 × 3` 超過 65535 時會繞回——出貨資料的上限最多
// 一萬多，碰不到，但照原版寫。
func RebelPlanet(p []byte) {
	if len(p) < planetEconomy+4 {
		return
	}
	p[planetOwner] = RegionOwnerNone
	eco := int32(binary.LittleEndian.Uint32(p[planetEconomy:]))
	eco -= eco / 3
	binary.LittleEndian.PutUint32(p[planetEconomy:], uint32(eco))
	def := binary.LittleEndian.Uint16(p[planetDefence:])
	limit := binary.LittleEndian.Uint16(p[planetGrowthDef:])
	if def*3 < limit {
		def = 0
	} else {
		def -= limit / 3
	}
	binary.LittleEndian.PutUint16(p[planetDefence:], def)
}
