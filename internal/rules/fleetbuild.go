package rules

import (
	"encoding/json"
	"fmt"
	"os"
)

// 新編一支艦隊的編制（原版 `297B:0066` 的旗艦那一組 ＋ `0FF4:11F4` 的
// 第 1 組起，`docs/re/212`–`docs/re/217`）。
//
// **玩家「艦隊編成」與 AI 建艦隊走的是同一支** `297B:0066`
// （只有兩個呼叫者，`docs/re/217`），所以這裡的規則對兩條路都成立。

// FleetGroup 是編制裡的一組 ＝ 一筆艦艇。
//
// 原版把艦艇放在艦艇池（`ds:CFB4`，一筆 78 bytes），艦隊記錄 `+0x10`
// 起每組 4 bytes 指過去；remake 直接把解出來的欄位帶在艦隊上。
type FleetGroup struct {
	Kind  int `json:"kind"`  // 艦種（艦艇 `+0x04`）
	Ships int `json:"ships"` // 現有艘數（艦艇 `+0x08`）
	Full  int `json:"full"`  // 滿編艘數（艦艇 `+0x0A`）
	Slot2 int `json:"slot2"` // 編制槽 `[2]`，抄自樣板（語意未解）
	Slot3 int `json:"slot3"` // 編制槽 `[3]`，抄自樣板 ＝ 隊形槽索引
}

// CompositionTables 是新編艦隊要用的兩張原版資料表。
type CompositionTables struct {
	// Template ＝ `ds:2BC6`：16 個類別鍵 × 10 組 × (slot2, slot3)。
	Template [16][10][2]int
	// Fallback ＝ `ds:295C`：艦種編不出來時的替代順序（已去掉結尾的
	// `0xFF`；`0xFE` 留著，它是清單裡的特例標記，`docs/re/216`）。
	Fallback [8][]int
}

// LoadCompositionTables 讀 `docs/internals/data/` 那兩份 JSON。
func LoadCompositionTables(templatePath, fallbackPath string) (*CompositionTables, error) {
	var tpl struct {
		Keys []struct {
			Key   int `json:"key"`
			Slots []struct {
				Slot2 int `json:"slot2"`
				Slot3 int `json:"slot3"`
			} `json:"slots"`
		} `json:"keys"`
	}
	if err := readJSON(templatePath, &tpl); err != nil {
		return nil, err
	}
	var fb struct {
		Kinds []struct {
			Kind  int   `json:"kind"`
			Order []int `json:"order"`
		} `json:"kinds"`
	}
	if err := readJSON(fallbackPath, &fb); err != nil {
		return nil, err
	}
	t := &CompositionTables{}
	for _, k := range tpl.Keys {
		if k.Key < 0 || k.Key >= len(t.Template) {
			return nil, fmt.Errorf("%s：類別鍵 %d 超出 0–15", templatePath, k.Key)
		}
		for i, s := range k.Slots {
			if i >= len(t.Template[k.Key]) {
				break
			}
			t.Template[k.Key][i] = [2]int{s.Slot2, s.Slot3}
		}
	}
	for _, k := range fb.Kinds {
		if k.Kind < 0 || k.Kind >= len(t.Fallback) {
			return nil, fmt.Errorf("%s：艦種 %d 超出 0–7", fallbackPath, k.Kind)
		}
		t.Fallback[k.Kind] = k.Order
	}
	return t, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s：%w", path, err)
	}
	return nil
}

// shipKindStandalone 是不從星域存量扣的那個艦種（`2AD2:0029` 對它有
// 獨立分支，直接拿滿編艘數，`docs/re/215`）。
const shipKindStandalone = 7

// BuildInput 是新編一支艦隊要的東西。
type BuildInput struct {
	// Region 是**執行期星域記錄**，一筆 `RegionRecord`。
	// **會被就地改寫**——每一組都從這裡扣存量（`docs/re/215`）。
	Region []byte
	// RegionOwner 是那一格的擁有者（`+0x12`）。樣板給 0 而擁有者是 1 時
	// 改用 1（`docs/re/216`）。
	RegionOwner int
	// Faction 是艦隊的陣營，Kinds 是艦種屬性表（滿編艘數的來源）。
	Faction int
	Kinds   *ShipKindTable
	Tables  *CompositionTables

	// AdmiralLevel 決定應有組數，AdmiralKind 是提督 `+0x08` 的低四位
	// （旗艦那一組的艦種），ClassKey 選樣板。
	AdmiralLevel int
	AdmiralKind  int
	ClassKey     int
}

// full 回某個艦種的滿編艘數；查不到回 0。
func (in BuildInput) full(kind int) int {
	if in.Kinds == nil {
		return 0
	}
	k, ok := in.Kinds.Kind(in.Faction, kind)
	if !ok {
		return 0
	}
	return k.FullShips
}

// groups 回「那一格這個艦種編得出幾組」＝ 原版 `0FF4:0E84`：
// 存量 ÷ 滿編艘數（`docs/re/216`）。
func (in BuildInput) groups(kind int) int {
	f := in.full(kind)
	if f <= 0 || kind < 0 || kind >= regionShipKinds {
		return 0
	}
	stock := RegionShips(in.Region, in.Faction)
	if kind >= len(stock) {
		return 0
	}
	return stock[kind] / f
}

// substitute 回替代艦種 ＝ 原版 `0FF4:1350`：照 `ds:295C` 的順序找第一個
// 編得出至少一組的；找不到回 -1（原版回 `0xFF`）。
//
// 清單裡的 `0xFE` 在原版有一段獨立分支，**還沒讀**（`docs/re/216`）；
// 這裡直接跳過它。
func (in BuildInput) substitute(kind int) int {
	if in.Tables == nil || kind < 0 || kind >= len(in.Tables.Fallback) {
		return -1
	}
	for _, c := range in.Tables.Fallback[kind] {
		if c == 0xFE {
			continue // 未讀的特例
		}
		if in.groups(c) != 0 {
			return c
		}
	}
	return -1
}

// take 產生一組編制：算滿編、從星域存量搬走艘數（`2AD2:0029` ＋
// `2703:1433`，`docs/re/215`）。
func (in BuildInput) take(kind, slot2, slot3 int) FleetGroup {
	f := in.full(kind)
	g := FleetGroup{Kind: kind, Full: f, Slot2: slot2, Slot3: slot3}
	if kind == shipKindStandalone {
		g.Ships = f // 不扣存量
		return g
	}
	g.Ships = TakeRegionShips(in.Region, in.Faction, kind, f)
	return g
}

// BuildComposition 產生一支新艦隊的編制。
//
// 第 0 組是旗艦（`297B:0066` 填，艦種取提督的 `+0x08` 低四位），
// 第 1 組起照 `ds:2BC6` 的樣板、艦種不夠就照 `ds:295C` 換
// （`0FF4:11F4`）。**會就地扣掉 `Region` 的艦艇存量。**
//
// **沒有做排序**：原版填完會用 `0FF4:0FF2` 對第 1 組起排一次，
// 那個比較器還沒讀（`docs/re/213`）。組數與艘數不受影響。
//
// **沒有做 `a == 星域擁有者 且 v < 4` 那一段**：`0FF4:11F4` 裡還有一個
// 未讀的分支（`docs/re/216`）。
func BuildComposition(in BuildInput) []FleetGroup {
	n := FleetSlots(in.AdmiralLevel)
	out := make([]FleetGroup, 0, n)
	out = append(out, in.take(in.AdmiralKind, 0, 0)) // 旗艦
	if in.Tables == nil || in.ClassKey < 0 || in.ClassKey >= len(in.Tables.Template) {
		return out
	}
	tpl := in.Tables.Template[in.ClassKey]
	for i := 1; i < n && i < len(tpl); i++ {
		a := tpl[i][0]
		if in.RegionOwner == 1 && a == 0 {
			a = 1
		}
		if in.groups(a) == 0 {
			s := in.substitute(a)
			if s < 0 {
				break // 換不到 → 這一支就編到這裡
			}
			a = s
		}
		out = append(out, in.take(a, a, tpl[i][1]))
	}
	return out
}

// 艦種列裡的兩個每艘費用（`ds:4E3A` 那一列，`docs/re/218`）。
const (
	kindUpkeepPer = 0x04 // 每艘的維持費（進 ds:8F72）
	kindOtherPer  = 0x05 // 每艘的另一項費用（進 ds:8F82）
)

// FleetUpkeep 回一支艦隊的維持費 ＝ 原版 `297B:116C` 寫進艦隊 `+0x0A`
// 的那個值（`docs/re/218`）：
//
//	Σ(艦種列[+0x04] × 那一組的艘數) ÷ 1000 × (200 − operations) ÷ 100
//
// `operations` 是艦隊能力的第 10 項（`+0x47`）＝ `max(提督, 副官)` 的
// `ADMIRAL` `+14`（營運）。**營運越高、維持費越低。**
//
// 收據 `turn260mem` 的三支算出來是 217／85／124，與各自的 `+0x0A`
// 逐一相同，合計 426 ＝ `ds:8F72`。
func FleetUpkeep(kinds *ShipKindTable, faction int, comp []FleetGroup, operations int) int {
	return fleetCost(kinds, faction, comp, operations, kindUpkeepPer)
}

// FleetOtherCost 是同一支算出來的另一項（進 `ds:8F82`），
// 差別只在讀艦種列的 `+0x05`。項目名稱還沒解（`docs/internals/27`）。
func FleetOtherCost(kinds *ShipKindTable, faction int, comp []FleetGroup, operations int) int {
	return fleetCost(kinds, faction, comp, operations, kindOtherPer)
}

func fleetCost(kinds *ShipKindTable, faction int, comp []FleetGroup,
	operations, off int) int {
	if kinds == nil {
		return 0
	}
	sum := 0
	for _, g := range comp {
		k, ok := kinds.Kind(faction, g.Kind)
		if !ok {
			continue
		}
		sum += k.Byte(off) * g.Ships
	}
	return sum / 1000 * (200 - operations) / 100
}
