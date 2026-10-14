package rules

import "sort"

// 要塞砲的射程判定與損失（`docs/spec/028`、`docs/re/145`、`docs/re/147`）。

// EllipseRadius 回「以焦點為極點」的圓錐曲線在夾角 delta 上的半徑，
// 回值是 16.16（原版 `2264:047C` 與 `2264:0FC3` 共用同一段）。
//
//	ℓ ＝ b² ÷ a、e ＝ √(1 − (b/a)²)、r ＝ ℓ ÷ (1 − e × cos Δ)
//
// delta 是一整圈 256 的角度差，cos 查的是原版那張表（`ds:573A`，
// 值 × 1024）左移 6 位變成 16.16。
func EllipseRadius(semiMajor, semiMinor, delta int) Fixed {
	if semiMajor <= 0 || semiMinor <= 0 {
		return 0
	}
	l := FixedDiv(FixedFromInt(semiMinor*semiMinor), FixedFromInt(semiMajor))
	e := EllipseEccentricity(semiMajor, semiMinor)
	cos := Fixed(cos256(delta)) << 6
	return FixedDiv(l, FixedOne-FixedMul(e, cos))
}

// EllipseEccentricity 是 `3870:0643`：`√(1 − b² × 65536 ÷ a²)`，16.16。
//
// **要先算 b² 再一次除掉 a²**，不要先算 b/a 再平方——原版是
// `3CC9:0020(b² × 65536, a²)`，兩種寫法的截斷點不同（a ＝ 95、b ＝ 85
// 差 4 個 ulp）。
func EllipseEccentricity(semiMajor, semiMinor int) Fixed {
	if semiMajor <= 0 {
		return 0
	}
	bb := int64(semiMinor) * int64(semiMinor)
	aa := int64(semiMajor) * int64(semiMajor)
	return FixedSqrt(FixedOne - Fixed(bb<<16/aa))
}

// FortressFocalDistance 回焦點到橢圓中心的距離，也就是原版
// `3870:0614` 的 `⌊a × e⌋`。
//
// 數學上等於 `⌊√(a² − b²)⌋`，兩者只在「√(a²−b²) 離整數不到 2⁻¹⁶ × a」
// 時才可能差一格。八顆要塞的長半軸是 112／113／113／115／116／129／
// 133／144，算出來分別是 108／109／109／111／112／124／128／139，
// 最近的一個離整數也有 0.32，所以整數式在這款遊戲裡處處成立
// （`docs/re/147`）。
func FortressFocalDistance(semiMajor, semiMinor int) int {
	d := semiMajor*semiMajor - semiMinor*semiMinor
	if d <= 0 {
		return 0
	}
	c := 0
	for (c+1)*(c+1) <= d {
		c++
	}
	return c
}

// FortressGunInRange 回「這個點在不在射程橢圓裡」（原版 `2264:03D9`
// ＋ `2264:047C`）。座標都是戰場世界座標，aim 是行星記錄 `+0x50`。
func FortressGunInRange(px, py, aim, semiMajor, semiMinor, x, y int) bool {
	d := FixedDist(px, py, x, y)
	delta := AngleDiff256(Atan256(x-px, y-py), aim)
	return d <= EllipseRadius(semiMajor, semiMinor, delta)
}

// AngleDiff256 是 `2BE1:024A`：兩個 byte 角度的**正差**，不折回 0–128。
func AngleDiff256(a, b int) int {
	a &= 0xFF
	b &= 0xFF
	if a >= b {
		return a - b
	}
	return b - a
}

// FortressGunDefence 回一支分隊挨這一發時的防禦係數（`2264:0FC3`）。
//
// 與單擊算式的分母是**同一支**（`AttackCoefficient`，`docs/re/94`）：
// 取艦種自己的裝甲橢圓在「砲來的方向」上的半徑，再乘艦隊能力的第 0 項
// （記錄 `+0x3D`）除以 100。差別只在角度怎麼來——這裡是
// 「從分隊看向行星的方位」與艦隊朝向的夾角。
//
// armourMajor／armourMinor 是艦種表的 word 11／12（`docs/internals/40`）；
// facing 是艦隊的朝向（`+0x4A`）。
func FortressGunDefence(px, py, x, y, facing, armourMajor, armourMinor, ability int) int {
	delta := AngleDiff256(Atan256(px-x, py-y), facing)
	return AttackCoefficient(armourMajor, armourMinor, delta, ability)
}

// FortressGunBase 回損失的基準量（`2264:0EE3`）：
//
//	基準 ＝ 艘數 × 40 × (防衛力上限 ÷ 防禦係數) ÷ 12800
//
// 12800 是伊謝爾倫的防衛力上限，所以括號那一項讀起來是「以伊謝爾倫
// 正面全力一擊為 1.0 的火力比」。
func FortressGunBase(ships, defenceMax, defence int) int {
	if defence <= 0 {
		return 0
	}
	return ships * 40 * (defenceMax / defence) / 12800
}

// FortressGunLoss 回這一發要扣的艘數（還沒夾上限）。r1／r2 是原版連續
// 取的兩個亂數（`基準 ＝ 0` 時第一個不取）。
func FortressGunLoss(base, r1, r2 int) int {
	if base == 0 {
		r1 = 0
	} else {
		r1 %= base
	}
	return base + r1 - base/2 + r2%100
}

// FortressGunHit 是一發打在一支分隊上的結果。
type FortressGunHit struct {
	Planet int `json:"planet"` // 開火的行星（battlePlanets 的索引）
	Unit   int `json:"unit"`   // 被打的分隊（units 的索引）
	Lost   int `json:"lost"`   // 扣掉的艘數
	Scorer int `json:"scorer"` // 分數記給誰（＝ 被打一方的對手）
	Score  int `json:"score"`  // 加了多少分（＝ 扣掉的艘數 ÷ 20，逐支捨去）
}

// 艦種表裡「這個艦種有沒有飛彈／艦載機」的旗標位移（值 0xFF ＝ 沒有）。
// 艦種 5（運補艦）無條件縮減兩格，其他艦種才檢查旗標。
// （`203A:0593`，規格 078／研究 366；研究 147 的反向敘述已勘誤）。
const (
	kindMissileFlag = 0x0E
	kindFighterFlag = 0x12
	kindSupply      = 5
	kindScout       = 7
)

// FortressGunFire 把「已瞄準」的要塞砲全部打出去，回逐支的結果。
//
// 順序照原版（`2264:0248`）：先收射程內的分隊，再依比較函式排序
// （艦隊編號遞減、分隊序遞增、索引遞增），才逐支取兩個亂數算損失。
// **順序會影響結果**，因為亂數是共用的。
//
// 打完之後呼叫端要走一次 FortressGunTickAll。
func FortressGunFire(planets []BattlePlanet, fleets []BattleFleet,
	units []BattleUnit, kinds *ShipKindTable, r *Rand) []FortressGunHit {
	return fortressGunObserved(planets, fleets, units, kinds, r, nil)
}

func fortressGunObserved(planets []BattlePlanet, fleets []BattleFleet,
	units []BattleUnit, kinds *ShipKindTable, r *Rand, observe func(FortressGunScene)) []FortressGunHit {
	if r == nil {
		return nil
	}
	var out []FortressGunHit
	for pi := range planets {
		p := &planets[pi]
		if !p.Fortress || p.GunState != FortressGunAimed {
			continue
		}
		a, b := FortressRange(p.DefenceMax)
		if observe != nil {
			s := PlanetAssaultSnapshot{Planet: *p, Units: append([]BattleUnit(nil), units...), Seed: r.Seed()}
			s.Planet.NameBytes = append([]byte(nil), p.NameBytes...)
			observe(FortressGunScene{Planet: pi, Unit: -1, Before: s, After: s})
		}
		var hits []int
		for i := range units {
			u := &units[i]
			if u.Ships <= 0 || u.Kind == kindScout {
				continue
			}
			if FortressGunInRange(p.X, p.Y, p.GunAim, a, b, u.X, u.Y) {
				hits = append(hits, i)
			}
		}
		sortFortressTargets(hits, units)
		for at, i := range hits {
			var before PlanetAssaultSnapshot
			if observe != nil {
				before = fortressSnapshot(p, units[i].Fleet, fleets, units, r)
			}
			h := fireOne(pi, p, i, fleets, units, kinds, r)
			if observe != nil {
				observe(FortressGunScene{Planet: pi, Unit: i, First: at == 0, Before: before, After: fortressSnapshot(p, units[i].Fleet, fleets, units, r), Hit: h})
			}
			if h.Lost > 0 {
				out = append(out, h)
			}
		}
	}
	return out
}

// FortressGunScene保留144瞬時快照。Unit=-1為發射通知，空射也有。
type FortressGunScene struct {
	Planet, Unit  int
	First         bool
	Before, After PlanetAssaultSnapshot
	Hit           FortressGunHit
}

func FortressGunWithScenes(planets []BattlePlanet, fleets []BattleFleet, units []BattleUnit, kinds *ShipKindTable, r *Rand) ([]FortressGunHit, []FortressGunScene) {
	var scenes []FortressGunScene
	hits := fortressGunObserved(planets, fleets, units, kinds, r, func(s FortressGunScene) { scenes = append(scenes, s) })
	return hits, scenes
}

func fortressSnapshot(p *BattlePlanet, index int, fleets []BattleFleet, units []BattleUnit, r *Rand) PlanetAssaultSnapshot {
	var fleet BattleFleet
	for _, f := range fleets {
		if f.Index == index {
			fleet = f
			break
		}
	}
	// 只在副本計算逐擊完整度；正式FLEET仍由原回合接點刷新。
	for _, u := range units {
		if u.Fleet == index && u.Squadron >= 0 && u.Kind != kindScout && u.Ships > 0 {
			fleet.Integrity = FleetIntegrity(&fleet, units)
			break
		}
	}
	return assaultSnapshot(p, &fleet, units, r)
}

// sortFortressTargets 照帝國那支比較函式（`2264:052F`）排：艦隊編號
// 遞減、分隊序遞增、索引遞增。同盟那支（`2264:0612`）骨架相同。
func sortFortressTargets(idx []int, units []BattleUnit) {
	sort.SliceStable(idx, func(x, y int) bool {
		u, v := &units[idx[x]], &units[idx[y]]
		if u.Fleet != v.Fleet {
			return u.Fleet > v.Fleet
		}
		if u.Squadron != v.Squadron {
			return u.Squadron < v.Squadron
		}
		return idx[x] < idx[y]
	})
}

func fireOne(pi int, p *BattlePlanet, i int, fleets []BattleFleet,
	units []BattleUnit, kinds *ShipKindTable, r *Rand) FortressGunHit {
	u := &units[i]
	h := FortressGunHit{Planet: pi, Unit: i, Scorer: 1 - u.Faction}
	var facing, ability int
	for fi := range fleets {
		if fleets[fi].Index == u.Fleet {
			facing, ability = fleets[fi].Facing, fleets[fi].Ability0
			break
		}
	}
	k, ok := kinds.Kind(u.Faction, u.Kind)
	if !ok {
		return h
	}
	d := FortressGunDefence(p.X, p.Y, u.X, u.Y, facing, k.SemiMajor, k.SemiMinor, ability)
	base := FortressGunBase(u.Ships, p.DefenceMax, d)
	var r1 int
	if base != 0 {
		r1 = r.Next()
	}
	lost := FortressGunLoss(base, r1, r.Next())
	if lost < 0 {
		lost = 0
	}
	if lost > u.Ships {
		lost = u.Ships
	}
	if lost == 0 {
		return h
	}
	h.Lost = ApplyShipLoss(u, k, lost)
	h.Score = h.Lost / ScoreDivisor
	return h
}

// ApplyShipLoss 扣一支分隊的艘數並帶上所有副作用（原版 `203A:0593`，
// `docs/re/147`），回**實際**扣掉的艘數。
//
//	艦種 7（索敵分遣隊）不受損，直接回它現有的艘數
//	實扣 ＝ min(艘數, 量)
//	`+0x0E`／`+0x12`（飛彈／艦載機）按剩餘比例縮減（艦種 5，
//	或艦種表對應欄位不是 0xFF；規格 078）
//	`+0x08` −= 實扣、`+0x0C` −= 實扣² ÷ 滿編
func ApplyShipLoss(u *BattleUnit, k ShipKind, amount int) int {
	if u.Kind == kindScout {
		return u.Ships // 不受損，但回值仍是艘數——分攤時會吃掉那麼多餘額
	}
	lost := amount
	if lost > u.Ships {
		lost = u.Ships
	}
	if lost <= 0 {
		return 0
	}
	before := u.Ships
	u.Ships -= lost
	if u.FullShips > 0 {
		u.RepairCap -= lost * lost / u.FullShips
		if u.RepairCap < 0 {
			u.RepairCap = 0
		}
	}
	if u.Kind == kindSupply || k.Byte(kindMissileFlag) != 0xFF {
		u.Missiles = u.Missiles / before * (before - lost)
	}
	if u.Kind == kindSupply || k.Byte(kindFighterFlag) != 0xFF {
		u.Fighters = u.Fighters / before * (before - lost)
	}
	return lost
}

// FortressGunTickAll 把每一顆有要塞、狀態非 0 的行星減 1
// （`2264:000A`，每回合一次）。
func FortressGunTickAll(planets []BattlePlanet) {
	for i := range planets {
		if planets[i].Fortress {
			planets[i].GunState = FortressGunTick(planets[i].GunState)
		}
	}
}
