package rules

import (
	"encoding/binary"
	"fmt"
)

// 星域與行星的記錄版面（`docs/internals/22`、`docs/re/17`、`docs/re/23`）。
const (
	// PlanetRecord 是 `PLADAT.DAT` 一筆的長度；檔頭 64 bytes ＝ 32 個 word，
	// 第 k 個就是第 k 個星域有幾顆行星（合計 81，與筆數相符）。
	PlanetRecord = 82
	PlanetHeader = 64

	planetOwner   = 0x11 // 這顆行星屬於誰：0 帝國、1 同盟、2 其他
	planetDefence = 0x14 // 防衛力

	// RegionRecord 是 `STELLAR.DT*` 一筆的長度。
	RegionRecord = 126

	regionOwner   = 0x12 // 0 帝國、1 同盟、2 無主
	regionCount   = 0x16 // 三個 byte：帝國／同盟／其他 各有幾顆行星
	regionEconomy = 0x1A // 32 位元
	regionDefence = 0x20 // 16 位元，與 +0x22 是同一個值的兩份
	regionUnrest  = 0x7D // 未安定度，記錄的最後一個 byte；治安由它算出來
	regionTaxRate = 0x26 // 稅率，千分之一為單位（畫面上的 3.2% ＝ 32）
)

// RegionEconomy 回星域記錄裡的經濟力（32 位元）。
func RegionEconomy(record []byte) int {
	if len(record) < regionEconomy+4 {
		return 0
	}
	return int(record[regionEconomy]) | int(record[regionEconomy+1])<<8 |
		int(record[regionEconomy+2])<<16 | int(record[regionEconomy+3])<<24
}

// RegionTaxRate 回星域的稅率（千分之一）。**每個星域不一樣**——
// 畫面右側那個 3.2% 是首都的，不是全國統一的。
func RegionTaxRate(record []byte) int {
	if len(record) <= regionTaxRate {
		return 0
	}
	return int(record[regionTaxRate])
}

// RegionTax 回一個星域這個月的徵稅額 ＝ 經濟力 × 稅率 / 1000（捨去）。
func RegionTax(record []byte) int {
	return RegionEconomy(record) * RegionTaxRate(record) / 1000
}

// TotalTax 回一個陣營的總稅收 ＝ **自己擁有的**星域徵稅額相加。
//
// 月底的結算是：總稅收歸零 → 逐星域累加 → 軍事費 += 總稅收
// （線性 `273E0` 起那一支，`docs/re/26`）。
func TotalTax(records [][]byte, faction int) int {
	sum := 0
	for _, r := range records {
		if len(r) > regionOwner && int(r[regionOwner]) == faction {
			sum += RegionTax(r)
		}
	}
	return sum
}

// Security 回畫面上那個「治安 %」。
//
// 原版的算式是 `(200 − 未安定度) / 2`（`docs/re/24`）：
//
//	B8 C8 00       mov ax,00C8        ; 200
//	26 8A 4C 7D    mov cl,es:[si+7D]  ; 未安定度
//	30 ED          xor ch,ch
//	2B C1          sub ax,cx
//	99 2B C2 D1 F8 cwd; sub ax,dx; sar ax,1   ; 帶號除以 2
//
// 六個帝國星域實測全對。**只有自己這一方的星域才顯示數字**，
// 其餘畫面上寫 `----`。
func Security(unrest int) int { return (200 - unrest) / 2 }

// FactionColour 回陣營的文字色號：帝國 6、同盟 5、中立 7。
//
// 原版的表在 `ds:6096`，存的是 `0x96 0x95 0x97`——文字系統的顏色跳脫碼
// 是 `0x90 + 色號`（`docs/re/24`）。與四份收據上量到的墨色一致。
func FactionColour(faction int) uint8 {
	switch faction {
	case 0:
		return 6
	case 1:
		return 5
	default:
		return 7
	}
}

// SecurityColour 回治安那個數字用什麼顏色印：照未安定度分四段。
//
// 程式裡寫的是跳脫碼 `0x96`／`0x97`／`0x95`／`0x92`，減掉 `0x90` 就是色號。
// 實測：未安定度 36／38／27／32 印色 6，44／50 印色 7。
func SecurityColour(unrest int) uint8 {
	return [...]uint8{6, 7, 5, 2}[UnrestBand(unrest)]
}

// 戰略情報報告那張的地圖色（`docs/re/27`）。
var (
	// securityMapColours 是 `ds:1780`：治安段 → 色號。
	// 第 5 個（3）是不屬於自己的星域用的。
	securityMapColours = [5]uint8{6, 1, 5, 2, 3}

	// securityDivisor 是算治安段時的除數。**85 是實測定下來的**：
	// 796 年 1 月的資料只能把它框到 85–88，於是把瓦爾哈拉的未安定度
	// `-poke` 成 85 再看情報報告那張地圖——畫出來是色 5（第 2 段），
	// 只有除數 85 算得出來（86–88 會落在第 1 段、色 1）。收據 `secpoke55`。
	securityDivisor = 85
)

// SecurityMapColour 回戰略情報報告那張地圖上，一個星域該塗什麼顏色。
//
//	段 = 未安定度 × 100 / 除數 / 50，超過 3 夾成 3
//
// 不屬於自己的星域一律色 3（等於底色，看起來就是黑的）。
// 32 個星域與實測的著色逐一相符。
func SecurityMapColour(unrest int, own bool) uint8 {
	if !own {
		return securityMapColours[4]
	}
	band := unrest * 100 / securityDivisor / 50
	if band > 3 {
		band = 3
	}
	return securityMapColours[band]
}

// FleetsUnknown 是「看不到」的艦隊數；畫面上印 `--`。
const FleetsUnknown = 0xFF

// Fleets 回畫面上「艦隊 a：b」那兩個數字：自己的、敵方的。
//
// 記錄的 `+0x28`–`+0x2B` 是四個 byte，**兩邊看到的不是同一組**
// （`docs/re/24`）：帝國玩家看 `+0x28`／`+0x2B`，同盟玩家看
// `+0x2A`／`+0x29`。`0xFF` 代表看不到。
func Fleets(record []byte, faction int) (own, enemy int) {
	if len(record) < 0x2C {
		return FleetsUnknown, FleetsUnknown
	}
	if faction == 0 {
		return int(record[0x28]), int(record[0x2B])
	}
	return int(record[0x2A]), int(record[0x29])
}

// UnrestBand 回未安定度落在哪一段（0 最好、3 最差）。
// 門檻取自同一段程式的三個比較：40、80、120（`docs/re/24`）。
func UnrestBand(unrest int) int {
	switch {
	case unrest <= 40:
		return 0
	case unrest <= 80:
		return 1
	case unrest <= 120:
		return 2
	default:
		return 3
	}
}

// Planets 是 `PLADAT.DAT`：檔頭的每星域行星數 ＋ 81 筆行星記錄。
type Planets struct {
	PerRegion []int
	Data      []byte
}

// LoadPlanets 讀 `PLADAT.DAT`。
func LoadPlanets(b []byte) (*Planets, error) {
	if len(b) < PlanetHeader {
		return nil, fmt.Errorf("PLADAT.DAT 太短：%d bytes", len(b))
	}
	per := make([]int, 32)
	total := 0
	for i := range per {
		per[i] = int(binary.LittleEndian.Uint16(b[i*2:]))
		total += per[i]
	}
	if PlanetHeader+total*PlanetRecord > len(b) {
		return nil, fmt.Errorf("PLADAT.DAT 檔頭說有 %d 顆行星，檔案放不下", total)
	}
	return &Planets{PerRegion: per, Data: b}, nil
}

// Range 回第 region 個星域的行星在檔案裡的起訖序號。
//
// 行星是**照星域順序連續存放**的：檔頭第 k 個 word 就是第 k 個星域佔幾筆。
func (p *Planets) Range(region int) (start, n int) {
	if region < 0 || region >= len(p.PerRegion) {
		return 0, 0
	}
	for i := 0; i < region; i++ {
		start += p.PerRegion[i]
	}
	return start, p.PerRegion[region]
}

func (p *Planets) rec(i int) []byte {
	o := PlanetHeader + i*PlanetRecord
	return p.Data[o : o+PlanetRecord]
}

// Owner 回第 i 顆行星屬於誰。
func (p *Planets) Owner(i int) int { return int(p.rec(i)[planetOwner]) }

// Defence 回第 i 顆行星的防衛力。
func (p *Planets) Defence(i int) int {
	return int(binary.LittleEndian.Uint16(p.rec(i)[planetDefence:]))
}

// RegionDefence 回一個星域的防衛力：該星域所有行星的防衛力相加。
//
// 對過六個星域，與執行期星域記錄的 `+0x20` 完全相同（`docs/re/23`）。
func (p *Planets) RegionDefence(region int) int {
	start, n := p.Range(region)
	sum := 0
	for i := start; i < start+n; i++ {
		sum += p.Defence(i)
	}
	return sum
}

// RegionPlanets 回一個星域內各陣營的行星數：帝國、同盟、其他。
func (p *Planets) RegionPlanets(region int) (empire, alliance, other int) {
	start, n := p.Range(region)
	for i := start; i < start+n; i++ {
		switch p.Owner(i) {
		case 0:
			empire++
		case 1:
			alliance++
		default:
			other++
		}
	}
	return
}

// RegionOwner 判定一個星域屬於誰：**全屬同一方才算那一方的，只要混雜就是無主**
// （`docs/re/17`）。回 0 帝國、1 同盟、2 無主。
func (p *Planets) RegionOwner(region int) int {
	e, a, o := p.RegionPlanets(region)
	switch {
	case a == 0 && o == 0:
		return 0
	case e == 0 && o == 0:
		return 1
	default:
		return 2
	}
}

// Tax 回徵稅額：經濟力 × 稅率（捨去）。稅率以千分之一為單位
// （畫面上的 3.2% ＝ 32）。796 年 1 月與 2 月兩個月都成立（`docs/re/22`）。
func Tax(economy, ratePerMille int) int {
	return economy * ratePerMille / 1000
}

// Treasury 是一個陣營這個月的財政（`docs/internals/27`）。
type Treasury struct {
	Budget   int // 原定軍事費：這個月一開始有多少（ds:8F62）
	ShipPaid int // 已支付的船艦建造費（ds:8F94）
	FleetUpkeep int // 艦隊維持費（ds:8F72）
	Movement    int // 移動經費（ds:8F7A）
	Other       int // 還沒認出來的第三項（ds:8F82）
}

// Remaining 回手上還有多少軍事費 ＝ 原定軍事費 − 已支付的各項。
//
// 玩家那兩個月都成立：`60741 − 15149 ＝ 45592`、`56405 − 14066 ＝ 42339`
// （另外三項都是 0）。**同盟那邊對不上**，還有沒找到的支出項目
// （`docs/re/29`）。
func (t Treasury) Remaining() int {
	return t.Budget - t.ShipPaid - t.FleetUpkeep - t.Movement - t.Other
}
