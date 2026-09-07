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
)

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
