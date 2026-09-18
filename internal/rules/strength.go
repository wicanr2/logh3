package rules

import (
	"encoding/json"
	"os"
)

// 艦艇池一筆 78 bytes（`docs/internals/64`、`docs/re/42`／`344`）。
// **一筆 ＝ 一個分隊**：戰場上的單位與戰略地圖上艦隊的一組編制是同一筆記錄。
const (
	shipFaction = 0x00 // 陣營（抄艦隊 +0x01）
	shipOwner   = 0x01 // 擁有者 ＝ 艦隊表槽號 0–31；0xFF ＝ 空槽／已作廢
	shipRegion  = 0x02 // 所在星域（抄艦隊 +0x06，編制一動就同步）
	shipOrder   = 0x03 // 組序：建立時 0、`297B:03EB` 寫真值；索敵分遣隊固定 0xFF
	shipKind    = 0x04 // 艦種 0–6；7 ＝ 索敵分遣隊；0xFF ＝ 空的
	shipSlot    = 0x05 // 隊形槽 0–40（建立當下不設）
	shipMorale  = 0x06 // 分隊士氣 ＝ (滿編 − 折算損失) × 100 ÷ 滿編（建立當下不設）
	shipAlive   = 0x08 // 現有艘數（word）
	shipFull    = 0x0A // 滿編艘數（word）
	// shipRepairCap 是修理／損失上限：建立時等於現有艘數，戰損時
	// `203A:0593` 用 `Σ實扣² ÷ 滿編` 往下扣（`docs/re/330`）。
	shipRepairCap = 0x0C
	// shipStock1／shipStock2 是兩格「隨艦種變」的存量（**32 位元**，
	// 低字在前）：建立時抄艦種表，戰損時 `新 ＝ 舊 ÷ 原艘數 × (原艘數 − 實扣)`。
	// 艦種 7 兩格都是 0。
	shipStock1 = 0x0E
	shipStock2 = 0x12
	// shipDetail 是索敵等級 0–3（`docs/internals/38`）：戰場上每被看到一輪
	// ＋1、沒被看到就掉一級；佈陣時清 0，戰後留在記錄裡，戰略地圖右下角
	// 「敵艦隊」分頁與交戰演出的艦隊那格都拿它判「看得清不清楚」。
	shipDetail = 0x17
	// shipHidden 的 bit7 ＝ 不進攻擊名單、不被索敵看到。索敵分遣隊在
	// **建立當下**就先設起來（`297B:0066`）。
	shipHidden = 0x1E
	// shipPosX 是分隊位置 x（word，+0x36 是 y）：`297B:0D99` 寫、`297B:0C74` 拿來判佔用。
	shipPosX = 0x34
	// shipIcon 是戰場上那顆圖示物件的遠指標（建立時清 0；作廢時非 0 才釋放）。
	shipIcon = 0x38

	shipKindEmpty = 0xFF
	// ShipKindScout 是索敵分遣隊那一筆的艦種。建新艦隊時 `297B:0066`
	// **無條件多配一筆**，索引寫進艦隊記錄 `+0x38`，不在 `+0x10` 那十個槽裡
	// ——`docs/re/42` 當年「每支艦隊多一筆沒被引用的記錄」就是它。
	ShipKindScout = 7
)

// StrengthWeights 是每個艦種的戰力權重（原版的 `ds:6032`，八個 byte）。
type StrengthWeights []int

// LoadStrengthWeights 讀 `docs/internals/data/ship-strength-weights.json`。
func LoadStrengthWeights(path string) (StrengthWeights, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Weights []int `json:"weights"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	return doc.Weights, nil
}

// ShipStrength 回一筆艦艇記錄的戰力。
//
// 原版的算式（`2E54:1876`）：
//
//	權重[艦種] × 現有艘數 × 1000 ÷ 滿編艘數
//
// 滿編的時候就是 `權重 × 1000`，打殘了按比例往下掉。空的那幾筆
// （`+0x04` ＝ `0xFF`）不算。
func (w StrengthWeights) ShipStrength(record []byte) int {
	if len(record) < ShipRecord {
		return 0
	}
	kind := int(record[shipKind])
	if kind == shipKindEmpty || kind >= len(w) {
		return 0
	}
	full := int(record[shipFull]) | int(record[shipFull+1])<<8
	if full == 0 {
		return 0
	}
	alive := int(record[shipAlive]) | int(record[shipAlive+1])<<8
	return w[kind] * alive * 1000 / full
}

// FleetStrength 回存檔裡第 i 支艦隊的戰力總和。
//
// **只算現在編制裡的那幾組**（`+0x0F`）——原版 `2E54:1876` 的迴圈上界
// 就是它（`docs/internals/54`）。縮編過的艦隊後面幾個槽還留著舊索引，
// 用 `FleetShips` 會多算（`docs/re/194`）。
func (s *Slot) FleetStrength(w StrengthWeights, i int) int {
	total := 0
	for _, n := range s.FleetGroupShips(i) {
		if n < 0 || (n+1)*ShipRecord > len(s.Ship) {
			continue
		}
		total += w.ShipStrength(s.Ship[n*ShipRecord : (n+1)*ShipRecord])
	}
	return total
}

// FleetShipTotals 回第 i 支艦隊的「現有艘數總和」與「主力槽數」。
//
// 戰略 AI 的候選評分要這兩個數（`2E54:1574`，`docs/re/168`）：
// 艘數是各編制槽艦艇 `+0x08` 的和，主力槽是艦種（`+0x04`）為 0 或 1 的
// 那幾槽。
func (s *Slot) FleetShipTotals(i int) (ships, mainSlots int) {
	for _, n := range s.FleetGroupShips(i) {
		if n < 0 || (n+1)*ShipRecord > len(s.Ship) {
			continue
		}
		r := s.Ship[n*ShipRecord : (n+1)*ShipRecord]
		if r[shipKind] == shipKindEmpty {
			continue
		}
		ships += int(r[shipAlive]) | int(r[shipAlive+1])<<8
		if r[shipKind] == 0 || r[shipKind] == 1 {
			mainSlots++
		}
	}
	return ships, mainSlots
}
