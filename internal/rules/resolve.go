package rules

// 戰鬥結算（`docs/spec/018`）。
//
// 這一版走的是**艦隊層**那條算式（`21A0:0004`，`docs/re/86`）：一支艦隊
// 一回合算一次總損失，再按比例分攤給對方的分隊。原版每一擊走的是
// `2350:0023`，那條的分母 K 還沒解完（`docs/re/86`），解出來之後換掉。

// BattleFleet 是戰鬥期間一支艦隊要用到的欄位（艦隊記錄，`docs/internals/33`）。
type BattleFleet struct {
	Index   int  `json:"index"`   // 艦隊表的索引
	Faction int  `json:"faction"` // 0 帝國、1 同盟
	Bonus   int  `json:"bonus"`   // `+0x14`：基本量的加成
	Halved  bool `json:"halved"`  // `+0x12` 非 0：除數從 10 變 20
	Defence int  `json:"defence"` // `+0x40 × +0x41 ÷ 100`
	Facing  int  `json:"facing"`  // `+0x4A`：朝向
}

// NewBattleFleet 從艦隊記錄與算好的能力做出一支戰鬥用的艦隊。
func NewBattleFleet(index int, rec []byte, ab [FleetAbilityCount]byte) BattleFleet {
	f := BattleFleet{Index: index}
	if len(rec) > 0x14 {
		f.Faction = int(rec[1])
		f.Bonus = int(rec[0x14])
		f.Halved = rec[0x12] != 0
	}
	f.Defence = int(ab[defenceAbilityA]) * int(ab[defenceAbilityB]) / 100
	return f
}

// 防禦係數的兩項就是艦隊記錄的 `+0x40`／`+0x41`，也就是能力的第 3、4 項。
const (
	defenceAbilityA = 0x40 - FleetAbilityOff
	defenceAbilityB = 0x41 - FleetAbilityOff
)

// BattleLoss 是一支艦隊這一回合被扣掉的量。
type BattleLoss struct {
	Attacker int   `json:"attacker"` // 打的人（艦隊索引）
	Defender int   `json:"defender"` // 被打的人
	Total    int   `json:"total"`    // 總損失
	PerUnit  []int `json:"per_unit"` // 分攤到目標各分隊的量，順序同 units
}

// ResolveBattleTurn 跑一回合的結算：每一支艦隊各打一次，扣數直接寫回
// `units` 的 `Ships`。回這一回合發生的每一次交火。
//
// 交火條件照 `docs/spec/018`：以艦隊旗艦分隊的位置為中心，射程取旗下
// 分隊 `Attack` 的最大值，門檻是射程 × `RangeFactor`；範圍內有敵方分隊
// 才開火，目標是最近的那一支所屬的艦隊。
func ResolveBattleTurn(fleets []BattleFleet, units []BattleUnit, rnd *Rand) []BattleLoss {
	var out []BattleLoss
	for _, f := range fleets {
		mine := unitsOf(units, f.Index)
		if len(mine) == 0 {
			continue
		}
		cx, cy := units[mine[0]].X, units[mine[0]].Y
		limit := EngagementRange(maxAttack(units, mine)) * RangeFactor
		near := InRange(units, cx, cy, limit, 1-f.Faction)
		if len(near) == 0 {
			continue
		}
		target := nearestFleet(units, near, cx, cy)
		def := fleetByIndex(fleets, target)
		theirs := unitsOf(units, target)
		if len(theirs) == 0 {
			continue
		}
		tx, ty := units[theirs[0]].X, units[theirs[0]].Y
		total := BattleDamage(DamageInput{
			Kind1Ships: kindOneShips(units, mine),
			Bonus:      f.Bonus,
			Halved:     f.Halved,
			Distance:   isqrt((tx-cx)*(tx-cx) + (ty-cy)*(ty-cy)),
			Facing:     f.Facing,
			Bearing:    Bearing(cx, cy, tx, ty),
			Defence:    def.Defence,
			Rand:       rnd.Next(),
		})
		alive := make([]int, len(theirs))
		kinds := make([]int, len(theirs))
		for i, u := range theirs {
			alive[i], kinds[i] = units[u].Ships, units[u].Kind
		}
		per := DistributeLosses(alive, kinds, total, true)
		for i, u := range theirs {
			units[u].Ships -= per[i]
			if units[u].Ships < 0 {
				units[u].Ships = 0
			}
		}
		out = append(out, BattleLoss{
			Attacker: f.Index, Defender: target, Total: total, PerUnit: per,
		})
	}
	return out
}

func unitsOf(units []BattleUnit, fleet int) []int {
	var out []int
	for i, u := range units {
		if u.Fleet == fleet && !u.Out && u.Ships > 0 {
			out = append(out, i)
		}
	}
	return out
}

func maxAttack(units []BattleUnit, idx []int) int {
	best := 0
	for _, i := range idx {
		if units[i].Attack > best {
			best = units[i].Attack
		}
	}
	return best
}

func kindOneShips(units []BattleUnit, idx []int) int {
	sum := 0
	for _, i := range idx {
		if units[i].Kind == 1 {
			sum += units[i].Ships
		}
	}
	return sum
}

func nearestFleet(units []BattleUnit, idx []int, x, y int) int {
	best, bestD := -1, 0
	for _, i := range idx {
		dx, dy := units[i].X-x, units[i].Y-y
		d := dx*dx + dy*dy
		if best < 0 || d < bestD {
			best, bestD = units[i].Fleet, d
		}
	}
	return best
}

func fleetByIndex(fleets []BattleFleet, index int) BattleFleet {
	for _, f := range fleets {
		if f.Index == index {
			return f
		}
	}
	return BattleFleet{Index: index, Defence: 100}
}

// BattleFleetsFor 從存檔的艦隊表與提督表做出戰鬥用的艦隊資料。
//
// 能力自己算（`FleetAbilities`），不讀艦隊記錄裡現成的——remake 換了
// 提督或副官之後那份會過期（`docs/re/91`）。
func (s *Slot) BattleFleetsFor(indexes []int) []BattleFleet {
	out := make([]BattleFleet, 0, len(indexes))
	for _, i := range indexes {
		rec := s.fleet(i)
		if len(rec) < FleetRecord {
			continue
		}
		var adj []byte
		if rec[5] != 0xFF {
			adj = s.admiral(int(rec[5]))
		}
		f := NewBattleFleet(i, rec, FleetAbilities(s.admiral(int(rec[4])), adj))
		out = append(out, f)
	}
	return out
}

func (s *Slot) admiral(i int) []byte {
	if i < 0 || (i+1)*AdmiralRecord > len(s.Admiral) {
		return nil
	}
	return s.Admiral[i*AdmiralRecord : (i+1)*AdmiralRecord]
}
