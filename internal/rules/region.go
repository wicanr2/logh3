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

	planetOwner     = 0x11 // 這顆行星屬於誰：0 帝國、1 同盟、2 其他
	planetDefence   = 0x14 // 防衛力（畫面上的那個，星域防衛力 ＝ 這一欄相加）
	planetGrowthDef = 0x16 // 成長率的 ×3/2 加成看的是這一欄（`docs/re/44`）

	// RegionRecord 是 `STELLAR.DT*` 一筆的長度，RegionCount 是星域數。
	RegionRecord = 126
	RegionCount  = 32

	regionOwner   = 0x12 // 0 帝國、1 同盟、2 無主
	regionFleets  = 0x28 // +0x28 帝國、+0x29 同盟：這一格各有幾支艦隊
	regionCount   = 0x16 // 三個 byte：帝國／同盟／其他 各有幾顆行星
	regionEconomy = 0x1A // 32 位元
	regionDefence = 0x20 // 16 位元，與 +0x22 是同一個值的兩份
	// RegionUnrest 是未安定度，記錄的最後一個 byte；治安由它算出來。
	// 存檔裡也有這一欄，但載入時會重算（`docs/re/41`）。
	RegionUnrest  = 0x7D
	regionTaxRate = 0x26 // 稅率，千分之一為單位（畫面上的 3.2% ＝ 32）
	regionGrowth  = 0x1E // 成長率，千分之一為單位（畫面上的 4.5% ＝ 45）
)

// RegionFleets 回星域記錄裡「這一格某個陣營有幾支艦隊」：`+0x28` 是帝國、
// `+0x29` 是同盟（`docs/re/199`）。
//
// 這一格是遊戲每次重數出來的：`2703:08CE` 先把 32 格的 `+0x28`／`+0x29`
// 清 0，再掃艦隊表對 `+0x28 ＋ 陣營` 遞增（`docs/re/40`）。
// 戰略 AI 的 `2E54:005D` 拿**敵方**那一格判斷要不要退。
func RegionFleets(record []byte, faction int) int {
	if faction < 0 || faction > 1 || len(record) <= regionFleets+faction {
		return 0
	}
	return int(record[regionFleets+faction])
}

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

// RegionColour 回星圖上一個星域的勢力色：帝國 1、同盟 12、無主 3。
//
// 這是**地圖上的星點顏色**，跟 FactionColour（文字色 6／5／7）不是同一套。
// 量自 `strmap-regions.json` 的 `colour_796_01`：六個帝國星域全是 1、
// 同盟那批全是 12、其餘 3。首都那一格是例外，見 `docs/re/61`。
func RegionColour(faction int) uint8 {
	switch faction {
	case 0:
		return 1
	case 1:
		return 12
	default:
		return 3
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

	// DefaultSecurityDivisor 是算治安段時的除數，問不到情報參謀時用它。
	//
	// **85 是實測定下來的**：796 年 1 月的帝國局面只能把它框到 85–88，
	// 於是把瓦爾哈拉的未安定度 `-poke` 成 85 再看情報報告那張地圖——
	// 畫出來是色 5（第 2 段），只有 85 算得出來（86–88 會落在第 1 段、
	// 色 1）。收據 `secpoke55`。
	//
	// 那個 85 其實不是常數，是**帝國情報專任參謀（提督 13）的 `+0x14`**
	// ——同盟那位（提督 58）是 82，畫出來的段真的不一樣
	// （`docs/re/63`）。留這個常數只當備援。
	DefaultSecurityDivisor = 85
)

// admiralSecurity 是提督記錄裡當除數用的那一格。與稅率算式用的
// `admiralTaxA` 是同一個位元組（`+0x14`）。
const admiralSecurity = admiralTaxA

// SecurityDivisor 回情報專任參謀的 `+0x14`，那就是治安分段的除數。
//
// `admirals` 是 `ADMIRAL.DT*` 整份，`i` 是提督索引；問不到就回
// DefaultSecurityDivisor。
func SecurityDivisor(admirals []byte, i int) int {
	r := admiralRecord(admirals, i)
	if len(r) <= admiralSecurity || r[admiralSecurity] == 0 {
		return DefaultSecurityDivisor
	}
	return int(r[admiralSecurity])
}

// SecurityMapColour 回戰略情報報告那張地圖上，一個星域該塗什麼顏色。
//
//	段 = 未安定度 × 100 / 除數 / 50，超過 3 夾成 3
//
// 除數是**情報專任參謀的 `+0x14`**（`SecurityDivisor`）：帝國那位 85、
// 同盟那位 82，所以同一個未安定度在兩邊可能落到不同的段
// （未安定度 42：帝國第 0 段、同盟第 1 段，`docs/re/63`）。
//
// 不屬於自己的星域一律色 3（等於底色，看起來就是黑的）。
// 兩個陣營各 32 個星域與實測的著色逐一相符。
func SecurityMapColour(unrest int, own bool, divisor int) uint8 {
	if !own {
		return securityMapColours[4]
	}
	if divisor <= 0 {
		divisor = DefaultSecurityDivisor
	}
	band := unrest * 100 / divisor / 50
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

// 補給欄那三組欄位（`docs/re/111`，反組譯自 `0x140A5` 起那一段）。
const (
	planetYard     = 0x20 // ＋ 艦種 × 4：對該艦種的補給量（32 位元）
	planetSupplyM  = 0x3C // 補給欄的 M（32 位元）
	planetSupplyF  = 0x40 // 補給欄的 F（32 位元）
	planetYardKind = 7    // 七種艦艇
)

// u32 讀記錄裡的 32 位元值。
func (p *Planets) u32(i, off int) uint32 {
	return binary.LittleEndian.Uint32(p.rec(i)[off:])
}

// SupplyM／SupplyF 回補給欄那兩格是不是 `o`。
//
// 原版只看「非零」：`mov ax,es:[bx+3Ch] / or ax,es:[bx+3Eh] / jz`
// ——**兩個 word 或起來**，等於整個 32 位元非零。
func (p *Planets) SupplyM(i int) bool { return p.u32(i, planetSupplyM) != 0 }

// SupplyF 同 SupplyM，讀的是 `+0x40`。
func (p *Planets) SupplyF(i int) bool { return p.u32(i, planetSupplyF) != 0 }

// Yard 回這顆行星對第 kind 種艦艇能不能補給（畫面上那七格 `o`／`x`）。
//
// 欄位是 `+0x20 + 艦種 × 4` 的 32 位元值，判準同樣是「非零」。
// 七格的順序就是表頭 `BHCADTE` 的順序，也就是艦種 0–6。
func (p *Planets) Yard(i, kind int) bool {
	if kind < 0 || kind >= planetYardKind {
		return false
	}
	return p.u32(i, planetYard+kind*4) != 0
}

// Name 回第 i 顆行星的名字：記錄 `+0x00` 起的 Big5 字碼，`00` 結尾
// （最多四個全形字，`docs/re/111`）。
func (p *Planets) Name(i int) []uint16 {
	r := p.rec(i)
	var out []uint16
	for k := 0; k+1 < PlanetRecord && r[k] != 0; k += 2 {
		out = append(out, uint16(r[k])<<8|uint16(r[k+1]))
	}
	return out
}

// Owner 回第 i 顆行星屬於誰。
func (p *Planets) Owner(i int) int { return int(p.rec(i)[planetOwner]) }

// Defence 回第 i 顆行星的防衛力。
func (p *Planets) Defence(i int) int {
	return int(binary.LittleEndian.Uint16(p.rec(i)[planetDefence:]))
}

// GrowthDefence 回成長率加成看的那一欄（`+0x16`）。與 Defence 多半同值，
// 但原版讀的是這一格（`docs/re/44`）。
func (p *Planets) GrowthDefence(i int) int {
	return int(binary.LittleEndian.Uint16(p.rec(i)[planetGrowthDef:]))
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

// RegionOwnerOf 從星域記錄的 `+0x12` 讀擁有者（0 帝國、1 同盟、2 無主）。
func RegionOwnerOf(record []byte) int {
	if len(record) <= regionOwner {
		return 2
	}
	return int(record[regionOwner])
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
	Budget      int // 原定軍事費：這個月一開始有多少（ds:8F62）
	ShipPaid    int // 已支付的船艦建造費（ds:8F94）
	// 底下四格照畫面上的兩組（「已支付」與「本次預定」），
	// 對應的全域由收據 `fiscalrows5` 逐格釘出來（`docs/re/239`）。
	FleetUpkeep  int // 已支付 艦隊維持費（ds:8F7A ＝ 累計已付）
	Movement     int // 已支付 移動經費（ds:8F8A，`docs/re/240`）
	NextUpkeep   int // 本次預定 艦隊維持費（ds:8F72 ＝ Σ 艦隊 +0x0A）
	NextMovement int // 本次預定 移動經費（ds:8F82 ＝ Σ 會動的艦隊 +0x0C）
}

// Remaining 回手上還有多少軍事費（＝ 原版 `ds:8F52`）：
//
//	軍事費 ＝ 原定軍事費 − 已付建造費 − 累計已付維持費 − 已付移動經費
//
// **扣的是三個「已付」的累計欄位**，不含「本次預定」的 `ds:8F72`／`ds:8F82`
// ——後兩個是這一回合還沒發生的預估，混進來會重複算（`docs/re/258`）。
//
// 三格存檔 × 兩個陣營六組全中（`docs/internals/30`）：`WORK.DT8` 帝國
// `120533 − 8801 − 15155 − 6577 ＝ 90000`、同盟
// `47942 − 9030 − 12412 − 11500 ＝ 15000`，其餘四組同樣精確。
// 玩家開局那兩個月（另外兩項是 0）也還成立：`60741 − 15149 ＝ 45592`、
// `56405 − 14066 ＝ 42339`。
func (t Treasury) Remaining() int {
	return t.Budget - t.ShipPaid - t.FleetUpkeep - t.Movement
}

// Paid 是畫面上那四格加起來 ＝ 預估軍事費要扣掉的部分
// （`docs/re/239`：收據 `fiscalrows5` 把 60741 扣成 −5925，
// 22222 ＋ 11111 ＋ 33333 ＝ 66666）。
func (t Treasury) Paid() int {
	return t.FleetUpkeep + t.Movement + t.NextUpkeep + t.NextMovement
}

// 星域記錄裡兩個陣營的艦艇艘數各佔七個 32 位元（`docs/re/69`）。
const (
	regionShips     = 0x30
	regionShipKinds = 7
)

// RegionShips 回這個星域裡某一個陣營的七種艦艇艘數。
func RegionShips(record []byte, faction int) []int {
	off := regionShips + faction*regionShipKinds*4
	out := make([]int, regionShipKinds)
	for i := range out {
		o := off + i*4
		if o+4 > len(record) {
			return out
		}
		out[i] = int(binary.LittleEndian.Uint32(record[o:]))
	}
	return out
}

// 行星在星域地圖上的位置（`docs/re/89`）。
//
// `PLADAT.DAT` 裡沒有座標，是載入星域地圖時算出來的：地圖上圖塊編號
// 為 `PlanetTile` 的格子依列序對到該星域的行星，位置取格子的中心偏移。
const (
	PlanetTile   = 6  // 星域地圖上代表行星的圖塊
	planetOffX   = 8  // 一格 16 寬，取中央
	planetOffY   = 12 // 一格 20 高
	planetMapCol = 128
)

// PlanetSite 是一顆行星在星域地圖上的位置。
type PlanetSite struct {
	Col int `json:"col"`
	Row int `json:"row"`
	X   int `json:"x"`
	Y   int `json:"y"`
}

// PlanetSites 掃一張星域地圖，回上面的行星位置（依列序）。
func PlanetSites(cells []uint8, tileW, tileH int) []PlanetSite {
	if tileW <= 0 || tileH <= 0 {
		return nil
	}
	var out []PlanetSite
	for i, c := range cells {
		if c != PlanetTile {
			continue
		}
		col, row := i%planetMapCol, i/planetMapCol
		out = append(out, PlanetSite{
			Col: col, Row: row,
			X: col*tileW + planetOffX,
			Y: row*tileH + planetOffY,
		})
	}
	return out
}

// TakeRegionShips ＝ 原版 `2703:1433(星域, 陣營, 艦種)`：新編一組編制時
// **從星域的艦艇存量搬走一單位**（`docs/re/215`）。
//
//	n = min(存量, full)
//	存量 -= n
//	回 n
//
// `full` 是艦種表那一列的滿編艘數（`ds:4E3A` 的 word 0，
// `rules.ShipKind.FullShips`）。存量是 32 位元，就地改寫 `record`。
//
// 陣營或艦種超出範圍、或記錄太短時回 0 且不動任何東西。
func TakeRegionShips(record []byte, faction, kind, full int) int {
	if faction < 0 || faction > 1 || kind < 0 || kind >= regionShipKinds || full <= 0 {
		return 0
	}
	o := regionShips + faction*regionShipKinds*4 + kind*4
	if o+4 > len(record) {
		return 0
	}
	have := int(binary.LittleEndian.Uint32(record[o : o+4]))
	n := full
	if have < n {
		n = have
	}
	binary.LittleEndian.PutUint32(record[o:o+4], uint32(have-n))
	return n
}

// ReturnRegionShips ＝ 原版 `2703:155B(星域, 艦艇)`：拆掉一組編制時
// **把現有艘數整批還回星域的待命部隊**（`docs/re/229`）。
//
// 是 `TakeRegionShips` 的反向。艦種 7 不走這一支（產生它的時候也不扣
// 存量，`docs/re/215`）——那一條在呼叫端擋掉。
func ReturnRegionShips(record []byte, faction, kind, n int) {
	if faction < 0 || faction > 1 || kind < 0 || kind >= regionShipKinds || n <= 0 {
		return
	}
	o := regionShips + faction*regionShipKinds*4 + kind*4
	if o+4 > len(record) {
		return
	}
	have := int(binary.LittleEndian.Uint32(record[o : o+4]))
	binary.LittleEndian.PutUint32(record[o:o+4], uint32(have+n))
}
