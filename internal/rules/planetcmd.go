package rules

// 行星命令的規則（原版 `1930:0DA2`／`1AAB:0A33`／`2264:0049`，
// `docs/re/134`）。「補給」「補充」與艦艇版同一個骨架，只是中心、
// 陣營與半徑換成行星的，存量也放在行星身上。

// 行星記錄裡的存量欄位（`docs/internals/29`）。
const (
	// planetKindStock 起七個 32 位元，艦種 0–6 各一份：「補充」補艘數
	// 扣的就是它（`1AAB:0D9F`）。艦種 7 那一格會落在飛彈存量上，
	// 但索敵分遣隊不在候選裡，撞不到。
	planetKindStock = 0x20
	planetMissiles  = 0x3C // 飛彈存量（32 位元），「補給」扣它
	planetFighters  = 0x40 // 艦載機存量（32 位元）
	planetDefence2  = 0x16 // 防衛力的第二份，要塞砲的射程查它
	planetGunState  = 0x45 // 要塞砲狀態：0 沒鎖定、3 已鎖定
	planetGunTarget = 0x50 // 要塞砲鎖定的目標，0xFFFF ＝ 沒有
)

// FortressRange 回要塞砲射程那個**橢圓**的長短半軸，單位是像素。
//
//	長半軸 ＝ 防衛力 ÷ 200 ＋ 80、短半軸 ＝ 長半軸 ÷ 4（`2264:0EC3`）
//
// 伊謝爾倫的防衛力 12800 → 144 與 36。圖形種類是 2（`3870:03A3` →
// `39D7:035E`），畫的是一個可以傾斜的橢圓，傾角在描述的 `[+4]`
// （`docs/re/134`、`docs/re/135`）。
func FortressRange(defence int) (semiMajor, semiMinor int) {
	semiMajor = defence/200 + 80
	return semiMajor, semiMajor / 4
}

// FortressGunReady 回要塞砲現在能不能重新指定目標：狀態只有 0（沒鎖定）
// 與 3（已鎖定）兩種可以（`1D9E:0247`）。
func FortressGunReady(state int) bool { return state == 0 || state == 3 }

// 要塞砲的兩種狀態。
const (
	FortressGunIdle   = 0
	FortressGunLocked = 3
)

// PlanetStock 是一顆行星身上的三種存量。欄位對到行星記錄的
// `+0x3C`、`+0x40` 與 `+0x20` 起那七格。
type PlanetStock struct {
	Missiles int    `json:"missiles"`
	Fighters int    `json:"fighters"`
	Ships    [7]int `json:"ships"` // 艦種 0–6 各一份
}

// ReadPlanetStock 從一筆 82 位元組的行星記錄讀出存量。
func ReadPlanetStock(rec []byte) (PlanetStock, bool) {
	var s PlanetStock
	if len(rec) < PlanetRecord {
		return s, false
	}
	s.Missiles = int(le32(rec[planetMissiles:]))
	s.Fighters = int(le32(rec[planetFighters:]))
	for k := range s.Ships {
		s.Ships[k] = int(le32(rec[planetKindStock+k*4:]))
	}
	return s, true
}

// PlanetSupply 用行星的存量補給一個分隊，回實際給出的飛彈與艦載機。
// 滿載量的算法與補給艦那支完全相同（`docs/re/131`）。
func PlanetSupply(t *ShipKindTable, s *PlanetStock, to *BattleUnit) (missiles, fighters int) {
	if t == nil || s == nil || to == nil {
		return 0, 0
	}
	k, ok := t.Kind(to.Faction, to.Kind)
	if !ok {
		return 0, 0
	}
	missiles = give(k.Byte(kindMissilePer)*to.Ships-to.Missiles, s.Missiles)
	fighters = give(k.Byte(kindFighterFull)*to.Ships-to.Fighters, s.Fighters)
	if missiles == 0 && fighters == 0 {
		return 0, 0
	}
	s.Missiles -= missiles
	s.Fighters -= fighters
	to.Missiles += missiles
	to.Fighters += fighters
	return missiles, fighters
}

// PlanetReplenish 用行星的艦種存量補回艘數，回補了幾艘。
//
// 上限是**建制滿編 `+0x0A`**，與艦艇的「修理」補到 `+0x0C` 不一樣。
func PlanetReplenish(s *PlanetStock, to *BattleUnit) int {
	if s == nil || to == nil || to.Kind < 0 || to.Kind >= len(s.Ships) {
		return 0
	}
	n := give(to.FullShips-to.Ships, s.Ships[to.Kind])
	if n == 0 {
		return 0
	}
	to.Ships += n
	s.Ships[to.Kind] -= n
	return n
}

func le32(b []byte) uint32 {
	if len(b) < 4 {
		return 0
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
