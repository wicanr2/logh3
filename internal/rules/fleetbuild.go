package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

// stock 回那一格這個艦種的待命部隊存量。
func (in BuildInput) stock(kind int) int {
	if kind < 0 || kind >= regionShipKinds {
		return 0
	}
	s := RegionShips(in.Region, in.Faction)
	if kind >= len(s) {
		return 0
	}
	return s[kind]
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
// 清單裡的 `0xFE` 是「**艦種 5 與 6 挑目前比較少的那個**」
// （`0FF4:1404` 數這支艦隊已經有幾組是 5、幾組是 6，`docs/re/226`）。
// `built` 是已經編好的那幾組（含旗艦）。
func (in BuildInput) substitute(kind int, built []FleetGroup) int {
	if in.Tables == nil || kind < 0 || kind >= len(in.Tables.Fallback) {
		return -1
	}
	for _, c := range in.Tables.Fallback[kind] {
		if c != 0xFE {
			if in.groups(c) != 0 {
				return c
			}
			continue
		}
		for _, k := range balanceFiveSix(built) {
			if in.groups(k) != 0 {
				return k
			}
		}
	}
	return -1
}

// balanceFiveSix 回 `0xFE` 那一格要試的順序：艦種 5 與 6 裡**目前組數
// 少的先試**（原版 `0FF4:1350` 拿 `0FF4:1404` 數出來的兩個數比一次；
// 一樣多的時候先試 6）。
func balanceFiveSix(built []FleetGroup) [2]int {
	n5, n6 := 0, 0
	for _, g := range built {
		switch g.Slot2 {
		case 5:
			n5++
		case 6:
			n6++
		}
	}
	if n5 < n6 {
		return [2]int{5, 6}
	}
	return [2]int{6, 5}
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
// （`0FF4:11F4`，`docs/re/224`）：
//
//	旗艦的艦種 == 1 且 樣板給 0 → a = 1
//	v = 那一格這個艦種編得出幾組
//	(a == 旗艦的艦種 且 v < 4) 或 v == 0 → 照 ds:295C 換一個
//
// **會就地扣掉 `Region` 的艦艇存量。**
//
// 填完之後第 1 組起會排序（原版 `3C6A:000D` 的 qsort 配比較器
// `0FF4:0FF2`：`[2]` 升冪、相同再比 `[3]` 升冪，`docs/re/225`）。
func BuildComposition(in BuildInput) []FleetGroup {
	n := FleetSlots(in.AdmiralLevel)
	out := make([]FleetGroup, 0, n)
	if in.Tables == nil || in.ClassKey < 0 || in.ClassKey >= len(in.Tables.Template) {
		return append(out, in.take(in.AdmiralKind, in.AdmiralKind, 0))
	}
	tpl := in.Tables.Template[in.ClassKey]
	// 旗艦那一組的 `[3]` 取樣板第 0 組的（十六個鍵都是 20，所以收據
	// 分不出它是查表還是寫死）。
	out = append(out, in.take(in.AdmiralKind, in.AdmiralKind, tpl[0][1]))
	for i := 1; i < n && i < len(tpl); i++ {
		a := tpl[i][0]
		if in.AdmiralKind == 1 && a == 0 {
			a = 1
		}
		v := in.groups(a)
		if (a == in.AdmiralKind && v < 4) || v == 0 {
			s := in.substitute(a, out)
			if s < 0 {
				break // 換不到 → 這一支就編到這裡
			}
			a = s
		}
		out = append(out, in.take(a, a, tpl[i][1]))
	}
	sortGroups(out)
	return out
}

// sortGroups 排第 1 組起（旗艦不參與）＝ 原版 `0FF4:0FF2` 那個比較器：
// 先比 `[2]`（艦種）升冪，相同再比 `[3]`（隊形槽）升冪。
//
// **remake-owned**：原版走 `qsort`，兩組的 `[2]`／`[3]` 都相同時它不保證
// 誰在前面。樣板裡沒有重複的 `(slot2, slot3)`，但換過艦種之後可能撞上，
// 所以這裡用穩定排序把次序定下來。
func sortGroups(g []FleetGroup) {
	if len(g) < 3 {
		return
	}
	sort.SliceStable(g[1:], func(i, j int) bool {
		a, b := g[1+i], g[1+j]
		if a.Slot2 != b.Slot2 {
			return a.Slot2 < b.Slot2
		}
		return a.Slot3 < b.Slot3
	})
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

// FleetOtherCost 是**移動一格的花費**（艦隊記錄的 `+0x0C`）：同一支算式，
// 差別只在讀艦種列的 `+0x05`。原版走一格就從那一支陣營的軍事費扣掉它
// （`113A:0729`，`docs/re/237`）；同一個數也累加進 `ds:8F82`。
//
// 收據 `turn260mem` 的三支算出來是 263／95／134，與各自的 `+0x0C`
// 逐一相同，也與收據 `monthend` 監看到的三筆扣款相同。
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
