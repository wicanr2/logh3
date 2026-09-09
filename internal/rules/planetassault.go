package rules

import "sort"

// 打行星（規格 `docs/spec/019`；算式 `docs/re/146`、閘門 `docs/re/149`、
// 換手條件與反擊 `docs/re/151`）。
//
// 一回合八步裡的第六步：先把每一支在場艦隊的「鎖上哪一顆行星」重算
// （`20B4:000C` 清、`20B4:0091` 標），再逐顆行星把鎖上它的艦隊排序、
// 逐一結算（`20B4:01C1` → `20B4:0486` → `20B4:0503`）。

// assaultNearRange 是第一個距離帶（像素）。第二個是
// `PlanetQueryRadius(行星[+0x14])`。
const assaultNearRange = 150

// assaultDirTable 是 `ds:47D0`：行星反擊依「砲來的方向」打折的百分比。
// 索引 0 是艦隊正面（最耐打 80%），4 是背面（100%），兩側對稱。
var assaultDirTable = [8]int{80, 85, 90, 95, 100, 95, 90, 85}

// Direction8 是 `3982:0022`：把一整圈 256 的角度換成 0–7 的方向索引
// （四捨五入到最近的 32）。
func Direction8(a int) int { return (((a & 0xFF) >> 4) + 1) >> 1 & 7 }

// PlanetAssaultHit 是一支艦隊對一顆行星的一輪。
type PlanetAssaultHit struct {
	Fleet    int  `json:"fleet"`    // 艦隊表索引
	Planet   int  `json:"planet"`   // planets 的索引
	Damage   int  `json:"damage"`   // 打掉的防衛力
	Counter  int  `json:"counter"`  // 挨的反擊，總艘數
	Captured bool `json:"captured"` // 這一下把行星打下來了
	Scorer   int  `json:"scorer"`   // 分數記給誰
	Score    int  `json:"score"`    // 加了多少分
}

// PlanetAssaultTargets 回每一支艦隊這一回合鎖上的行星（−1 ＝ 沒有）與
// 兩個位元：bit 0 ＝ 進了 150 像素而且處於「戰鬥」狀態，
// bit 1 ＝ 還在行星的查詢半徑內、而且那顆不是「要塞砲已瞄準」的要塞。
func PlanetAssaultTargets(planets []BattlePlanet, fleets []BattleFleet,
	units []BattleUnit) (target, bits []int) {
	target = make([]int, len(fleets))
	bits = make([]int, len(fleets))
	for i := range target {
		target[i] = -1
	}
	for pi := range planets {
		p := &planets[pi]
		r := PlanetQueryRadius(p.Defence)
		for fi := range fleets {
			f := &fleets[fi]
			if f.Faction == p.Owner || f.Flag0 || f.Action != 1 {
				continue
			}
			if flagshipShips(f.Index, units) == 0 {
				continue
			}
			d := FixedDist(p.X, p.Y, f.X, f.Y)
			if d > FixedFromInt(assaultNearRange) {
				continue
			}
			bits[fi] |= 1
			target[fi] = pi
			if d > FixedFromInt(r) {
				continue
			}
			if p.Fortress && p.GunState == FortressGunAimed {
				continue
			}
			bits[fi] |= 2
			target[fi] = pi
		}
	}
	return target, bits
}

// flagshipShips 回這一支艦隊旗艦（分隊序 0）的艘數。
func flagshipShips(fleet int, units []BattleUnit) int {
	for i := range units {
		if units[i].Fleet == fleet && units[i].Squadron == 0 {
			return units[i].Ships
		}
	}
	return 0
}

// PlanetAssault 跑完打行星那一步，回逐支的結果。
func PlanetAssault(planets []BattlePlanet, fleets []BattleFleet,
	units []BattleUnit, kinds *ShipKindTable, r *Rand) []PlanetAssaultHit {
	if r == nil {
		return nil
	}
	target, bits := PlanetAssaultTargets(planets, fleets, units)
	order := make([]int, 0, len(fleets))
	for fi := range fleets {
		if target[fi] >= 0 {
			order = append(order, fi)
		}
	}
	// `20B4:02A4`：先依鎖上的行星遞增，再依 `+0x3C` 遞增。
	sort.SliceStable(order, func(x, y int) bool {
		if target[order[x]] != target[order[y]] {
			return target[order[x]] < target[order[y]]
		}
		return fleets[order[x]].Integrity < fleets[order[y]].Integrity
	})
	var out []PlanetAssaultHit
	for pi := range planets {
		any := false
		for _, fi := range order {
			if target[fi] == pi && bits[fi]&1 != 0 {
				any = true
				break
			}
		}
		if !any {
			continue
		}
		for _, fi := range order {
			if target[fi] != pi {
				continue
			}
			if h, ok := assaultOne(&planets[pi], pi, &fleets[fi], bits[fi],
				units, kinds, r); ok {
				out = append(out, h)
			}
		}
	}
	return out
}

// assaultOne 是 `20B4:0503`：一支艦隊對一顆行星的一輪。
func assaultOne(p *BattlePlanet, pi int, f *BattleFleet, bits int,
	units []BattleUnit, kinds *ShipKindTable, r *Rand) (PlanetAssaultHit, bool) {
	h := PlanetAssaultHit{Fleet: f.Index, Planet: pi, Scorer: f.Faction}
	if p.Owner == f.Faction {
		return h, false // 已經是自己的了（前面有人打下來）
	}
	// 門檻 ＝ 防衛力還剩幾成。不低於它就先開火，低於就先挨反擊。
	thr := 0
	if p.DefenceMax > 0 {
		thr = p.Defence * 100 / p.DefenceMax
	}
	if f.Integrity >= thr {
		h.Damage, h.Captured = assaultFire(p, f, bits, units, r)
		if !h.Captured {
			h.Counter = assaultCounter(p, f, bits, units, kinds, r)
		}
	} else {
		h.Counter = assaultCounter(p, f, bits, units, kinds, r)
		h.Damage, h.Captured = assaultFire(p, f, bits, units, r)
	}
	h.Score = h.Damage / ScoreDivisor
	if h.Captured {
		h.Score += p.DefenceMax / captureScoreDiv
		capturePlanet(p, f.Faction)
	}
	return h, true
}

// captureScoreDiv 是換手獎金的除數：`行星[+0x16] ÷ 10`。
const captureScoreDiv = 10

// assaultFire 是 `20B4:0681`：算這一輪打掉多少防衛力並扣下去，
// 回值的第二項 ＝「攻擊力追上剩下的防衛力」（換手的條件，`20B4:0E2C`）。
func assaultFire(p *BattlePlanet, f *BattleFleet, bits int,
	units []BattleUnit, r *Rand) (int, bool) {
	if bits&1 == 0 {
		return 0, false
	}
	a := FleetAttack(f, units)
	b := assaultBase(a, FixedDist(p.X, p.Y, f.X, f.Y))
	b = b/2 + randMod(r, b/3)
	c := PlanetGuard(p, units) / 6
	amt := 0
	if b >= c {
		amt = b - c
	}
	amt += r.Next() % 100
	if amt > p.Defence {
		amt = p.Defence
	}
	p.Defence -= amt
	return amt, a >= p.Defence
}

// assaultBase 套距離衰減：`量 −= 量 × (距離 ÷ 150) ÷ 5`（`20B4:0D96`
// 與 `20B4:0E7E` 共用同一形狀）。dist 是 16.16。
func assaultBase(v int, dist Fixed) int {
	q := FixedDiv(dist, FixedFromInt(assaultNearRange))
	return v - int(int64(v)*int64(q)/(5<<16))
}

// FleetAttack 是 `203A:0424`：這一支艦種 0／1 的艘數合計 ×
// 能力第 3 項 ÷ 100 ÷ 15 ＋ 250。
func FleetAttack(f *BattleFleet, units []BattleUnit) int {
	main := 0
	for i := range units {
		if units[i].Fleet != f.Index {
			continue
		}
		if units[i].Kind == 0 || units[i].Kind == 1 {
			main += units[i].Ships
		}
	}
	return main*f.Ability3/100/15 + 250
}

// PlanetGuard 是 `203A:04D1`：以行星為心、`PlanetQueryRadius` 為半徑，
// **行星自己那一方**艦種 0／1 的艘數合計 ÷ 100。
func PlanetGuard(p *BattlePlanet, units []BattleUnit) int {
	r := PlanetQueryRadius(p.Defence)
	rr := FixedFromInt(r)
	total := 0
	for i := range units {
		u := &units[i]
		if u.Faction != p.Owner || (u.Kind != 0 && u.Kind != 1) || u.Ships == 0 {
			continue
		}
		if FixedDist(p.X, p.Y, u.X, u.Y) <= rr {
			total += u.Ships
		}
	}
	return total / 100
}

// assaultCounter 是 `20B4:0E7E` ＋ `20B4:103D`：行星的反擊。
func assaultCounter(p *BattlePlanet, f *BattleFleet, bits int,
	units []BattleUnit, kinds *ShipKindTable, r *Rand) int {
	if bits&2 == 0 {
		return 0
	}
	div := 10
	if p.Fortress {
		div = 20
	}
	base := (PlanetGuard(p, units) + p.Defence) / div
	base = assaultBase(base, FixedDist(p.X, p.Y, f.X, f.Y))
	// 方向：從行星看艦隊的方位與艦隊朝向的夾角 → 0–7 → 百分比。
	dir := Direction8(AngleDiff256(Atan256(p.X-f.X, p.Y-f.Y), f.Facing))
	base = base * assaultDirTable[dir] / 100
	if f.Defence <= 0 {
		return 0
	}
	amt := base*100/f.Defence + r.Next()%100
	return spreadOverFleet(f, units, kinds, amt)
}

// spreadOverFleet 是 `20B4:103D`：把量按「艘數佔全隊的百分比」分攤到各
// 分隊，餘數丟給艘數最多的那一支，而且**不打光**（每支至少留一艘）。
//
// 兩段整數捨去（先算百分比再乘）是原版的形狀，換成一次乘除會差幾艘。
func spreadOverFleet(f *BattleFleet, units []BattleUnit,
	kinds *ShipKindTable, amount int) int {
	idx := make([]int, 0, 16)
	for i := range units {
		if units[i].Fleet == f.Index {
			idx = append(idx, i)
		}
	}
	sort.SliceStable(idx, func(x, y int) bool {
		return units[idx[x]].Squadron < units[idx[y]].Squadron
	})
	total, big, bigShips := 0, -1, 0
	for _, i := range idx {
		n := units[i].Ships
		if n == 0 {
			continue
		}
		total += n
		if n > bigShips {
			bigShips, big = n, i
		}
	}
	if total == 0 {
		return 0
	}
	left, done := amount, 0
	for _, i := range idx {
		u := &units[i]
		if u.Ships == 0 {
			continue
		}
		share := amount * (u.Ships * 100 / total) / 100
		if share >= u.Ships {
			share = u.Ships - 1
		}
		k, ok := kinds.Kind(u.Faction, u.Kind)
		if !ok {
			continue
		}
		got := ApplyShipLoss(u, k, share)
		left -= got
		if u.Kind != kindScout {
			done += got
		}
	}
	if left > 0 && big >= 0 && units[big].Ships > 0 {
		u := &units[big]
		share := left
		if share >= u.Ships {
			share = u.Ships - 1
		}
		if k, ok := kinds.Kind(u.Faction, u.Kind); ok {
			done += ApplyShipLoss(u, k, share)
		}
	}
	return done
}

// capturePlanet 是換手的效果（`20B4:0503` 尾巴，收據 `s3assault` 逐一相符）。
func capturePlanet(p *BattlePlanet, faction int) {
	p.Owner = faction
	p.Stock.Ships = [7]int{}        // `+0x20`–`+0x3B` ← 0
	p.Stock.Missiles /= 2           // `+0x3C` 右移一位
	p.Stock.Fighters = 0            // `+0x40` ← 0
}

// FleetIntegrity 重算一支艦隊的編制完整度（記錄 `+0x3C`，`docs/re/153`）：
//
//	折算損失 ＝ min(損失 × (200 − 提督[+0x0B]) ÷ 100, 滿編)
//	這一支   ＝ (滿編 − 折算損失) × 100 ÷ 滿編
//	完整度   ＝ 上面那些值的**算術平均**（只算還有艘數的分隊）
//
// 全隊都沒艘數就回 0——原版那時改設 `+0x00` bit 3。
func FleetIntegrity(f *BattleFleet, units []BattleUnit) int {
	fac := 200 - f.AdmiralLossFactor
	sum, n := 0, 0
	for i := range units {
		u := &units[i]
		if u.Fleet != f.Index || u.Ships == 0 || u.FullShips == 0 {
			continue
		}
		v := (u.FullShips - u.Ships) * fac / 100
		if v > u.FullShips {
			v = u.FullShips
		}
		sum += (u.FullShips - v) * 100 / u.FullShips
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / n
}

// RefreshFleetIntegrity 把每一支的 Integrity 重算一次。原版在每次扣完
// 艘數之後就重算（`1261:*`），這裡在打行星之前走一次就夠——那正是唯一
// 讀它的地方。
func RefreshFleetIntegrity(fleets []BattleFleet, units []BattleUnit) {
	for i := range fleets {
		fleets[i].Integrity = FleetIntegrity(&fleets[i], units)
	}
}

// randMod 回 `rand() % n`；n 為 0 時原版不取那一個亂數。
func randMod(r *Rand, n int) int {
	if n <= 0 {
		return 0
	}
	return r.Next() % n
}
