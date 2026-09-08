package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 戰術選單第二項「資料」（規格 `docs/spec/021`）。
//
// 版面與繪製在 `internal/screen`，這一層只負責「表上要印什麼」。

// 這張表上會用到的字碼。
const (
	b5Di       = 0xB2C4 // 第
	b5Jian     = 0xC4A5 // 艦
	b5Dui      = 0xB6A4 // 隊
	b5Bu       = 0xB3A1 // 部
	b5Slash    = 0xA1FE // ／（原版用的是 A1FE，字形比 A241 陡）
	b5FullZero = 0xA2AF // ０，全形數字連續十個
	b5FullA    = 0xA2CF // Ａ，全形大寫字母連續二十六個
)

// shipKindLetters 是艦種的代號字母，一種一個（`docs/re/111`）。
//
// 這七個字母在原版記憶體裡就是一條字串 `BHCADTE`（線性 `0x5D94A`），
// 旁邊還有小寫的 `bhcadte`。對得上兩件事：**收據 `s3data` 那一支艦隊
// 五個艦種 0 印Ｂ、三個艦種 2 印Ｃ**，以及七個艦種名稱的英文字首——
// 戰艦 Battleship、高速戰艦 High-speed、巡航艦 Cruiser、航空母艦
// Aircraft carrier、驅逐艦 Destroyer、運補艦 Transport、修護艦
// Engineering。行星那個分頁的第四欄用同一組字母當欄名。
var shipKindLetters = "BHCADTE"

// shipKindLetter 回艦種的代號字母（全形）；超出範圍回空白。
func shipKindLetter(kind int) uint16 {
	if kind < 0 || kind >= len(shipKindLetters) {
		return ' '
	}
	return uint16(b5FullA + int(shipKindLetters[kind]-'A'))
}

// battleDataInk 回最左那一欄的顏色：**擁有者的陣營色**
// （帝國 6、同盟 5，量自收據 `s3data_planet`，`docs/re/111`）。
func battleDataInk(faction int) uint8 {
	if faction == 0 {
		return 6
	}
	return 5
}

// OpenBattleData 打開「資料」，停在第一個分頁。
func (g *Game) OpenBattleData() {
	g.battleData, g.battleDataTab = true, 0
}

// CloseBattleData 關掉「資料」。
func (g *Game) CloseBattleData() { g.battleData = false }

// BattleDataTab 回現在看的是第幾個分頁；沒開回 -1。
func (g *Game) BattleDataTab() int {
	if !g.battleData {
		return -1
	}
	return g.battleDataTab
}

// battleDataClick 處理「資料」開著時的點擊：點標籤換分頁，
// 點別的地方關掉。回 true ＝ 這一下被它吃掉了。
func (g *Game) battleDataClick(x, y int) bool {
	if !g.battleData {
		return false
	}
	if s := g.res.BattleData; s != nil {
		if t := s.BattleDataTabAt(x, y); t >= 0 {
			g.battleDataTab = t
			return true
		}
	}
	g.battleData = false
	return true
}

// battleDataRows 組出表上的每一列。
//
// **士氣那一欄空著**：原版印的是 `%3u`，收據裡三個分頁每一列都是 100，
// 從哪裡算還沒解（`docs/spec/021` 的「還沒解」）。填一個猜的值會讓畫面
// 看起來是對的而其實是錯的，空著比較誠實。
//
// 行星那個分頁還沒接：戰場上的行星只量到座標，對不回星系記錄
// （`docs/re/99`）。
func (g *Game) battleDataRows(tab int) [][]screen.BattleDataRun {
	if tab == 1 {
		return nil
	}
	want := g.faction
	if tab == 2 {
		want = 1 - g.faction
	}
	var out [][]screen.BattleDataRun
	seen := map[int]bool{}
	for _, u := range g.battleUnits {
		if u.Faction != want {
			continue
		}
		row := []screen.BattleDataRun{}
		if !seen[u.Fleet] {
			seen[u.Fleet] = true
			if f := g.fleetByIndex(u.Fleet); f != nil {
				row = append(row, screen.BattleDataRun{
					X: 48, Ink: battleDataInk(u.Faction), Codes: fleetLabel(f.Number)})
			}
		}
		row = append(row,
			screen.BattleDataRun{X: 128, Codes: g.squadronName(u)},
			screen.BattleDataRun{X: 272, Codes: []uint16{shipKindLetter(u.Kind)}},
			screen.BattleDataRun{X: 304, Codes: rightDigits(u.Ships, 4)},
			screen.BattleDataRun{X: 336, Codes: []uint16{b5Slash}},
			screen.BattleDataRun{X: 352, Codes: rightDigits(u.FullShips, 4)})
		out = append(out, row)
	}
	return out
}

// fleetByIndex 回艦隊表第 i 筆；找不到回 nil。
func (g *Game) fleetByIndex(i int) *rules.Fleet {
	for k := range g.fleets {
		if g.fleets[k].Index == i {
			return &g.fleets[k]
		}
	}
	return nil
}

// squadronName 回名字那一欄：旗艦印提督的名字，其餘印「第Ｎ部隊」。
func (g *Game) squadronName(u rules.BattleUnit) []uint16 {
	if u.Squadron == 0 {
		if f := g.fleetByIndex(u.Fleet); f != nil &&
			f.Admiral >= 0 && f.Admiral < len(g.res.AdmiralNames) {
			return g.res.AdmiralNames[f.Admiral]
		}
	}
	return unitLabel(b5Bu, u.Squadron+1)
}

// fleetLabel 回「第<兩個半形數字>艦隊」。
func fleetLabel(n int) []uint16 {
	out := []uint16{b5Di}
	out = append(out, rightDigits(n, 2)...)
	return append(out, b5Jian, b5Dui)
}

// unitLabel 回「第Ｎ部隊」。原版的樣板是「第？部隊」——？ 佔一格，
// 所以個位數用全形數字，兩位數就塞兩個半形（寬度一樣）。
func unitLabel(kind uint16, n int) []uint16 {
	out := []uint16{b5Di}
	if n >= 1 && n <= 9 {
		out = append(out, uint16(b5FullZero+n))
	} else {
		out = append(out, rightDigits(n, 2)...)
	}
	return append(out, kind, b5Dui)
}

// rightDigits 回靠右對齊的 w 個半形字（不足補空白），同原版的 `%Nu`。
func rightDigits(v, w int) []uint16 {
	out := make([]uint16, w)
	for i := range out {
		out[i] = ' '
	}
	if v < 0 {
		v = 0
	}
	for i := w - 1; i >= 0; i-- {
		out[i] = uint16('0' + v%10)
		v /= 10
		if v == 0 {
			break
		}
	}
	return out
}
