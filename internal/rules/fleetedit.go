package rules

import (
	"fmt"
	"sort"
)

// 艦隊組成編輯器（原版 `0FF4:000A`，`docs/re/317`、`docs/internals/56`）。
//
// 出擊命令那條路（艦隊編成按「確認」）與艦隊視窗「變更 → 編組」開的是
// **同一支**：進來時把編制抄成「隊形槽 → 一組」的工作表（`ds:977A`），
// 之後點格盤放、右鍵拆、「陣形」整批換，每一步都直接動那一格星域的
// 待命存量；「決定」把工作表壓實排序寫回艦隊記錄，「中止」把存量倒回
// 進來那一刻的樣子（`ds:987E` 的備份）。

// EditorSlots 是格盤的隊形槽數（`ds:977A` 41 筆，索引 ＝ `ds:28B8` 的槽）。
const EditorSlots = 41

// EditorTables 是編輯器的三張常數表（`docs/internals/data/fleet-editor.json`）。
type EditorTables struct {
	// Templates ＝ `ds:2D06`：八個陣形 × 10 組 (艦種, 隊形槽)；第 0 組
	// 固定 (0, 20) 是旗艦，套的時候從第 1 組起。
	Templates [8][10][2]int
	// Substitute ＝ `ds:2986`：那一種編不出一組時照這個順序換
	// （**不是**新編艦隊用的 `ds:295C`，這裡沒有 0xFE）。
	Substitute [7][]int
	// Mobility ＝ `ds:4E2E`：機動力的「組數 → 係數」（`297B:083F`）。
	Mobility []int
}

// LoadEditorTables 讀 `fleet-editor.json`。
func LoadEditorTables(path string) (*EditorTables, error) {
	var raw struct {
		Templates struct {
			Rows []struct {
				Index int      `json:"index"`
				Pairs [][2]int `json:"pairs"`
			} `json:"rows"`
		} `json:"templates"`
		Substitute struct {
			Kinds []struct {
				Kind  int   `json:"kind"`
				Order []int `json:"order"`
			} `json:"kinds"`
		} `json:"substitute"`
		Mobility struct {
			Table []int `json:"table"`
		} `json:"mobility"`
	}
	if err := readJSON(path, &raw); err != nil {
		return nil, err
	}
	t := &EditorTables{Mobility: raw.Mobility.Table}
	for _, r := range raw.Templates.Rows {
		if r.Index < 0 || r.Index >= len(t.Templates) {
			return nil, fmt.Errorf("%s：陣形 %d 超出 0–7", path, r.Index)
		}
		for i, p := range r.Pairs {
			if i >= len(t.Templates[r.Index]) {
				break
			}
			t.Templates[r.Index][i] = p
		}
	}
	for _, k := range raw.Substitute.Kinds {
		if k.Kind < 0 || k.Kind >= len(t.Substitute) {
			return nil, fmt.Errorf("%s：艦種 %d 超出 0–6", path, k.Kind)
		}
		t.Substitute[k.Kind] = k.Order
	}
	return t, nil
}

// FleetMobility 是艦隊組成畫面的「機動力」＝ 原版 `297B:083F(艦隊, 組數)`：
//
//	100 − (270 − 2 × ops) × table[groups] ÷ 100      ; 整數除法
//
// `ops` 是提督與副官 `+23`（運用）取大（沒有副官就只看提督），
// `table` 是 `ds:4E2E`（索引 ＝ 現有組數）。萊因哈特那一支 2／3／8 組
// 印 98／95／81％（收據 `fcmd-edit-place1`／`-place2`／`-form-pick`）。
func FleetMobility(table []int, ops, groups int) int {
	if groups < 0 || groups >= len(table) {
		return 100
	}
	return 100 - (270-2*ops)*table[groups]/100
}

// EditResult 是放一組／拆一組的結果：`EditOK` 改了工作表，其餘四種是
// 原版跳的那四句訊息（`3826:0005(0, 0x97, …)`，`message-box.json` 的
// `edit_*`），`EditNoop` 是什麼都沒發生（右鍵點在空格）。
type EditResult int

const (
	EditOK             EditResult = iota
	EditNoop                      // 右鍵點在空格：什麼都不做
	EditFlagshipKind              // 「只有戰艦或者高速戰艦才能當做旗艦。」（`ds:2B4E`）
	EditNoShips                   // 「船艦數量不足。」（`ds:2B71`）
	EditOverCapacity              // 「無法超過目前統率能力以上的船艦數量。」（`ds:2B80`）
	EditFlagshipRemove            // 「無法撤離旗艦。」（`ds:2BA5`）
)

// MessageName 回這個結果對應的訊息名（`message-box.json`）；沒有訊息回空字串。
func (r EditResult) MessageName() string {
	switch r {
	case EditFlagshipKind:
		return "edit_flagship_kind"
	case EditNoShips:
		return "edit_no_ships"
	case EditOverCapacity:
		return "edit_over_capacity"
	case EditFlagshipRemove:
		return "edit_flagship_remove"
	}
	return ""
}

// FleetEditor 是開著的一次編輯。
type FleetEditor struct {
	// In 帶星域記錄（**會被就地改寫**：每放一組就從那一格扣存量）、
	// 陣營與艦種表；Tables 是三張常數表。
	In     BuildInput
	Tables *EditorTables

	// Expected ＝ 艦隊 `+0x0E`：編制槽的應有數 ＝ 部隊數的分母
	// （`ds:984B ＋ ds:984C`）。
	Expected int

	// Kind ＝ `ds:984A`：反白的那一列艦種，放格子時放這一種。
	Kind int
	// User-selected remake extension, spec150. Zero value preserves original kinds.
	NonBattleshipFlagship bool

	// Slots ＝ `ds:977A`：索引是隊形槽，nil 是空位。
	Slots [EditorSlots]*FleetGroup

	// stock 是進來那一刻那一格存量的備份（`ds:987E`，28 bytes）。
	stock []byte
}

// NewFleetEditor 開一次編輯：把 comp 照隊形槽放進工作表、備份存量
// （`0FF4:0094`／`00F5`／`06C3`）。
func NewFleetEditor(in BuildInput, tables *EditorTables, expected int, comp []FleetGroup) *FleetEditor {
	e := &FleetEditor{In: in, Tables: tables, Expected: expected}
	for i := range comp {
		g := comp[i]
		if g.Slot3 < 0 || g.Slot3 >= EditorSlots {
			continue
		}
		e.Slots[g.Slot3] = &g
	}
	off := regionShips + in.Faction*regionShipKinds*4
	if off+regionShipKinds*4 <= len(in.Region) {
		e.stock = append([]byte(nil), in.Region[off:off+regionShipKinds*4]...)
	}
	return e
}

// Used ＝ `ds:984B`：已編成幾組。
func (e *FleetEditor) Used() int {
	n := 0
	for _, s := range e.Slots {
		if s != nil {
			n++
		}
	}
	return n
}

// Free ＝ `ds:984C`：還能放幾組。
func (e *FleetEditor) Free() int {
	if f := e.Expected - e.Used(); f > 0 {
		return f
	}
	return 0
}

// full 回艦種的滿編艘數；沒帶艦種表就用量到的那七個定值。
func (e *FleetEditor) full(kind int) int {
	if f := e.In.full(kind); f > 0 {
		return f
	}
	if kind >= 0 && kind < len(ShipsPerUnit) {
		return ShipsPerUnit[kind]
	}
	return 0
}

// Available ＝ 原版 `0FF4:0E84`：那一格這個艦種編得出幾組 ＝ 存量 ÷ 滿編。
func (e *FleetEditor) Available(kind int) int {
	f := e.full(kind)
	if f <= 0 {
		return 0
	}
	return e.In.stock(kind) / f
}

// Groups ＝ `ds:981E`（`0FF4:0F6F`）：旗艦那一組排第一（隊形槽寫成 20），
// 其餘照艦種、再照隊形槽升冪。
func (e *FleetEditor) Groups() []FleetGroup {
	out := make([]FleetGroup, 0, e.Used())
	if s := e.Slots[FlagshipSlot3]; s != nil {
		g := *s
		g.Slot3 = FlagshipSlot3
		out = append(out, g)
	}
	rest := make([]FleetGroup, 0, e.Used())
	for i, s := range e.Slots {
		if s == nil || i == FlagshipSlot3 {
			continue
		}
		rest = append(rest, *s)
	}
	sort.SliceStable(rest, func(i, j int) bool {
		if rest[i].Slot2 != rest[j].Slot2 {
			return rest[i].Slot2 < rest[j].Slot2
		}
		return rest[i].Slot3 < rest[j].Slot3
	})
	return append(out, rest...)
}

// free 把那一格的組拆掉、艘數退回存量（`2AD2:018C`）。
func (e *FleetEditor) free(slot int) {
	s := e.Slots[slot]
	if s == nil {
		return
	}
	if s.Kind != shipKindStandalone {
		ReturnRegionShips(e.In.Region, e.In.Faction, s.Kind, s.Ships)
	}
	e.Slots[slot] = nil
}

// Place ＝ 原版 `0FF4:0C80(隊形槽)`：把反白那一種放到那一格。
//
//	槽 20 只收艦種 0／1                      → EditFlagshipKind
//	這一種編不出一組                         → EditNoShips
//	那一格已經有人 → 先拆掉（艘數退回）
//	沒有空位（ds:984C == 0）                → EditOverCapacity
//	放：艘數 ＝ min(滿編, 存量)，從存量扣掉（2AD2:0029）
func (e *FleetEditor) Place(slot int) EditResult {
	if slot < 0 || slot >= EditorSlots {
		return EditNoop
	}
	if slot == FlagshipSlot3 && !FlagshipKindAllowed(e.Kind, e.NonBattleshipFlagship) {
		return EditFlagshipKind
	}
	if e.Available(e.Kind) == 0 {
		return EditNoShips
	}
	e.free(slot)
	if e.Free() == 0 {
		return EditOverCapacity
	}
	g := e.take(e.Kind, slot)
	e.Slots[slot] = &g
	return EditOK
}

// take ＝ `BuildInput.take`，但滿編用這裡的退路。
func (e *FleetEditor) take(kind, slot int) FleetGroup {
	f := e.full(kind)
	g := FleetGroup{Kind: kind, Full: f, Slot2: kind, Slot3: slot, Pool: -1}
	if kind == shipKindStandalone {
		g.Ships = f
		InitGroupStock(e.In.Kinds, e.In.Faction, &g)
		return g
	}
	g.Ships = TakeRegionShips(e.In.Region, e.In.Faction, kind, f)
	InitGroupStock(e.In.Kinds, e.In.Faction, &g)
	return g
}

// Remove ＝ 原版 `0FF4:0DB1(隊形槽)`（右鍵）：空格什麼都不做、槽 20 拒絕、
// 其餘拆掉退回存量。
func (e *FleetEditor) Remove(slot int) EditResult {
	if slot < 0 || slot >= EditorSlots || e.Slots[slot] == nil {
		return EditNoop
	}
	if slot == FlagshipSlot3 {
		return EditFlagshipRemove
	}
	e.free(slot)
	return EditOK
}

// substitute ＝ 原版 `0FF4:0B38`：照 `ds:2986` 的順序找第一個編得出一組的；
// 找不到回 -1。
func (e *FleetEditor) substitute(kind int) int {
	if e.Tables == nil || kind < 0 || kind >= len(e.Tables.Substitute) {
		return -1
	}
	for _, c := range e.Tables.Substitute[kind] {
		if e.Available(c) != 0 {
			return c
		}
	}
	return -1
}

// ApplyFormation ＝ 原版 `0FF4:0A4B(陣形)`：先把旗艦以外全部拆掉
// （`0FF4:0AFB`），再照樣板第 1 組起放，放到沒有空位為止；旗艦是
// 高速戰艦時樣板的戰艦改成高速戰艦；那一種編不出一組就換，換不到
// 整批停下來。反白的艦種列套完復原。
func (e *FleetEditor) ApplyFormation(t int) {
	if e.Tables == nil || t < 0 || t >= len(e.Tables.Templates) {
		return
	}
	for i := range e.Slots {
		if i != FlagshipSlot3 {
			e.free(i)
		}
	}
	flag := 0
	if s := e.Slots[FlagshipSlot3]; s != nil {
		flag = s.Kind
	}
	saved := e.Kind
	defer func() { e.Kind = saved }()
	for _, p := range e.Tables.Templates[t][1:] {
		if e.Free() == 0 {
			return
		}
		k := p[0]
		if flag == 1 && k == 0 {
			k = 1
		}
		e.Kind = k
		if e.Available(k) == 0 {
			s := e.substitute(k)
			if s < 0 {
				return
			}
			e.Kind = s
		}
		if e.Place(p[1]) != EditOK {
			return
		}
	}
}

// Cancel ＝ 原版 `0FF4:1127`：存量倒回進來那一刻的備份。編制本身由
// 呼叫端保留原本那一份（原版是把艦艇記錄的備份抄回去）。
func (e *FleetEditor) Cancel() {
	if e.stock == nil {
		return
	}
	off := regionShips + e.In.Faction*regionShipKinds*4
	copy(e.In.Region[off:off+len(e.stock)], e.stock)
}
