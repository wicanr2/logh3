package rules

import "encoding/binary"

// 戰鬥結算（`docs/spec/018`）。
//
// 這一版走的是**艦隊層**那條算式（`21A0:0004`，`docs/re/86`）：一支艦隊
// 一回合算一次總損失，再按比例分攤給對方的分隊。原版每一擊走的是
// `2350:0023`，那條的分母 K 還沒解完（`docs/re/86`），解出來之後換掉。

// MaxBattleTurns 是一場戰鬥最多幾個回合（`docs/re/92`）。
//
// 原版的主迴圈是 `cmp word ptr ds:8F1A, 9` ／ `jnb 離開`，回合從 0 起算，
// 所以畫面上印的是第 1 到第 9 回合。
const MaxBattleTurns = 9

// BattleFleet 是戰鬥期間一支艦隊要用到的欄位（艦隊記錄，`docs/internals/33`）。
type BattleFleet struct {
	Index   int  `json:"index"`   // 艦隊表的索引
	Number  int  `json:"number"`  // `+0x02`：畫面上「第 N 艦隊」的那個 N
	Faction int  `json:"faction"` // 0 帝國、1 同盟
	Bonus   int  `json:"bonus"`   // `+0x14`：基本量的加成
	Halved  bool `json:"halved"`  // `+0x12` 非 0：除數從 10 變 20
	Defence int  `json:"defence"` // `+0x40 × +0x41 ÷ 100`
	Facing  int  `json:"facing"`  // `+0x4A`：朝向

	// Ability0 是能力的第 0 項（記錄 `+0x3D`）。單擊算式與要塞砲的
	// 防禦係數都乘它再除 100（`2264:0FC3`，`docs/re/147`）。
	Ability0 int `json:"ability0"`

	// Flag0 是 `+0x00` bit0。語意未解（`docs/internals/33`），但兩件事
	// 靠它：戰力加總不算這種艦隊，戰場的艦隊選單把「戰鬥」與「索敵」
	// 關掉（`docs/spec/023`）。
	Flag0 bool `json:"flag0"`
	// ScoutOut ＝ `+0x3A` 不是 0xFF 或 `+0x3B` 不是 0：已經放了索敵隊
	// 出去，「索敵」不能再選。
	ScoutOut bool `json:"scout_out"`

	// 下面五格是艦隊層的「移動／戰鬥」命令寫進去的（`2583:0003`，
	// `docs/re/120`–`122`）。原版對應 `+0x60`／`+0x62`（下令前的位置
	// 備份）、`+0x64`／`+0x66`（要去的點）、`+0x4C`（抵達朝向）。
	OriginX, OriginY int `json:"origin_x,omitempty"`
	DestX, DestY     int `json:"dest_x,omitempty"`
	TargetFacing     int `json:"target_facing,omitempty"`

	// Ordered ＝ 這一支收到過移動／戰鬥命令（上面五格才有意義）。
	Ordered bool `json:"ordered,omitempty"`

	// X／Y 是艦隊自己的位置（`+0x5C`／`+0x5E`，＝ 旗艦的位置），
	// Budget／PerTurn 是行程額度（`+0x56`/`+0x58` 與 `+0x52`/`+0x54`，
	// 16.16）。艦隊層的移動引擎用它們（`docs/re/129`）。
	X, Y            int `json:"x,omitempty"`
	Budget, PerTurn int `json:"budget,omitempty"`
	TurnSpeed       int `json:"turn_speed,omitempty"`

	// Action 是艦隊記錄 `+0x48` 的低四位。「戰鬥」成功之後寫 1
	// （`1CC5:08BD`）。
	Action int `json:"action,omitempty"`
}

// NewBattleFleet 從艦隊記錄與算好的能力做出一支戰鬥用的艦隊。
func NewBattleFleet(index int, rec []byte, ab [FleetAbilityCount]byte) BattleFleet {
	f := BattleFleet{Index: index}
	if len(rec) > 0x14 {
		f.Faction = int(rec[1])
		f.Number = int(rec[2])
		f.Bonus = int(rec[0x14])
		f.Halved = rec[0x12] != 0
		f.Flag0 = rec[0]&1 != 0
	}
	if len(rec) > 0x3B {
		f.ScoutOut = rec[0x3A] != 0xFF || rec[0x3B] != 0
	}
	f.Defence = int(ab[defenceAbilityA]) * int(ab[defenceAbilityB]) / 100
	f.Ability0 = int(ab[0])
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
	Scorer   int   `json:"scorer"`   // 分數記給哪一邊（＝ 被扣一方的對手）
	Score    int   `json:"score"`    // 這一次交火加了多少分
}

// ScoreDivisor 是分數的除數：每扣一個分隊，對手加「扣掉的艘數 ÷ 20」
// （`203A:06C0`，`docs/re/93`）。
const ScoreDivisor = 20

// BattleVictory 回勝利點 ＝ 帝國的分數 − 同盟的分數（`docs/re/93`）。
//
// 原版沒有把它存起來，畫的時候現算（`13581` 那一段的 32 位元減法）。
func BattleVictory(score [2]int) int { return score[0] - score[1] }

// ResolveBattleTurn 跑一回合的結算，扣數直接寫回 `units` 的 `Ships`。
// 回這一回合發生的每一次交火。
//
// **以行星為中心**（`docs/re/99`）：一顆行星對每一支敵方艦隊算一次。
//
//	量     ＝ Σ{艦種 0 或 1、陣營 ＝ 行星擁有者、距行星 ≤ 半徑} 現有艘數
//	基本量 ＝ (量 ÷ 100 ＋ 行星防衛力) ÷ (10 或 20)
//	總損失 ＝ (基本量 − 距離項) × 方向係數 ÷ 100 × 100 ÷ 防禦係數 ＋ 亂數 %100
//
// 半徑是 `PlanetQueryRadius(防衛力)`，距離量的是行星到目標艦隊旗艦。
// **哪些（行星, 艦隊）配對會交火、順序如何，原版還沒逆向**——這裡對
// 每一顆有主的行星、每一支敵方艦隊各算一次。
func ResolveBattleTurn(planets []BattlePlanet, fleets []BattleFleet,
	units []BattleUnit, rnd *Rand) []BattleLoss {
	var out []BattleLoss
	for _, p := range planets {
		if p.Owner != 0 && p.Owner != 1 {
			continue // 無主的行星不開火
		}
		ships := shipsNearPlanet(units, p)
		for _, f := range fleets {
			if f.Faction == p.Owner {
				continue
			}
			theirs := unitsOf(units, f.Index)
			if len(theirs) == 0 {
				continue
			}
			tx, ty := units[theirs[0]].X, units[theirs[0]].Y
			total := BattleDamage(DamageInput{
				Ships01:  ships,
				Bonus:    p.Defence,
				Halved:   p.Fortress, // 有要塞：除數從 10 變 20
				Distance: isqrt((tx-p.X)*(tx-p.X) + (ty-p.Y)*(ty-p.Y)),
				Facing:   f.Facing,
				Bearing:  Bearing(p.X, p.Y, tx, ty),
				Defence:  f.Defence,
				Rand:     rnd.Next(),
			})
			out = append(out, resolveOne(units, theirs, f, total, p.Owner))
		}
	}
	return out
}

// shipsNearPlanet ＝ `203A:04D1`：行星附近、擁有者那一方、艦種 0 或 1 的
// 現有艘數總和，再除 100。比的是分隊的**目的地**（`17A0:00AC`）。
func shipsNearPlanet(units []BattleUnit, p BattlePlanet) int {
	r := PlanetQueryRadius(p.Defence)
	sum := 0
	for _, i := range InRange(units, p.X, p.Y, r, p.Owner) {
		if units[i].Kind == 0 || units[i].Kind == 1 {
			sum += units[i].Ships
		}
	}
	return sum
}

// resolveOne 把總損失分攤到一支艦隊的各分隊，扣回去，回這一次交火。
func resolveOne(units []BattleUnit, theirs []int, f BattleFleet,
	total, scorer int) BattleLoss {
	alive := make([]int, len(theirs))
	kinds := make([]int, len(theirs))
	for i, u := range theirs {
		alive[i], kinds[i] = units[u].Ships, units[u].Kind
	}
	per := DistributeLosses(alive, kinds, total, true)
	score := 0
	for i, u := range theirs {
		units[u].Ships -= per[i]
		if units[u].Ships < 0 {
			units[u].Ships = 0
		}
		// 分數是**逐個分隊**除 20 再加，不是總量除一次——原版的
		// 善後常式每扣一筆就加一次（`docs/re/93`）。
		score += per[i] / ScoreDivisor
	}
	return BattleLoss{
		Attacker: -1, Defender: f.Index, Total: total, PerUnit: per,
		Scorer: scorer, Score: score,
	}
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

// shipsKind01 加總艦種 0 與 1 的現有艘數（`203A:04D1`，`docs/re/99`）。
func shipsKind01(units []BattleUnit, idx []int) int {
	sum := 0
	for _, i := range idx {
		if units[i].Kind == 0 || units[i].Kind == 1 {
			sum += units[i].Ships
		}
	}
	return sum
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

// ApplyBattleResult 把戰場上剩下的艘數寫回艦艇池。
//
// 只寫 `+0x08`（現有艘數）。原版戰後還會動什麼（撤退、星域易主、
// 分數怎麼結算）還沒解，所以這裡就只做這一件確定的事。
func (s *Slot) ApplyBattleResult(units []BattleUnit) int {
	changed := 0
	for _, u := range units {
		ships := s.FleetShips(u.Fleet)
		if u.Squadron < 0 || u.Squadron >= len(ships) {
			continue
		}
		n := ships[u.Squadron]
		if n < 0 || (n+1)*ShipRecord > len(s.Ship) {
			continue
		}
		rec := s.Ship[n*ShipRecord : (n+1)*ShipRecord]
		if int(binary.LittleEndian.Uint16(rec[8:])) == u.Ships {
			continue
		}
		binary.LittleEndian.PutUint16(rec[8:], uint16(u.Ships))
		changed++
	}
	return changed
}
