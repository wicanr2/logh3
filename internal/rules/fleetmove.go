package rules

// 艦隊在戰場上怎麼轉、怎麼走（原版 `1C03:01B5`、`1C03:0524`、`3310:0006`）。
//
// 這一層與分隊的移動（`AdvanceBattle`）不是同一支常式：艦隊記錄自己帶
// 一個 16.16 的行程額度，轉向與前進都從裡面扣，而且**扣在落點檢查前面**
// ——被擋掉的一步照樣扣。逐位元的對照見 `docs/re/129`。

// FleetMover 是艦隊移動要用到的那幾格（`docs/internals/33`）。
type FleetMover struct {
	X, Y      int // +0x5C／+0x5E 現在位置
	DestX     int // +0x64
	DestY     int // +0x66
	Facing    int // +0x4A 現在朝向（一整圈 256）
	TurnSpeed int // +0x4D 回轉速度
	Budget    int // +0x56／+0x58 這一回合剩下的行程，16.16
	PerTurn   int // +0x52／+0x54 每回合行程，16.16
	Arrived   bool
}

// RefillFleetBudget 是每回合開頭那一下（`1C03:0040`）：**無條件覆蓋**，
// 不是把剩下的加上去。分隊那份（`1C61:0005`）才有「不足一步就累加」。
func RefillFleetBudget(f *FleetMover) { f.Budget = f.PerTurn }

// Distance16 是原版的兩點距離（`3310:0006`），回 16.16。
//
// 算的是 `M × √(1 + (m/M)²)`，`M` 取兩軸差的大者。中間兩步都截斷
// （`39B5:005C` 除、`39B5:000C` 乘），開根號走 `fixSqrt`。**不是**
// `√(dx²＋dy²)` 的直接寫法——差一個最低位就對不上收據。
func Distance16(x0, y0, x1, y1 int) int64 {
	dx, dy := abs32(x1-x0), abs32(y1-y0)
	m, M := dx, dy
	if m > M {
		m, M = M, m
	}
	if M == 0 {
		return 0
	}
	q := (int64(m) << 16) / int64(M)
	return int64(M) * fixSqrt(FixOne+(q*q)>>16)
}

// TurnFacing256 是帶正負號的最短角差（原版 `2BE1:021B`）：`a − b` 之後
// 正規化到 −128..127。
func TurnFacing256(target, cur int) int {
	d := (target - cur) & 0xFF
	if d >= 128 {
		d -= 256
	}
	return d
}

// fleetStepCost 是額度的門檻與轉向的固定花費，都是 16.16 的 8.0。
const fleetStepCost = BattleStep * FixOne

// turnFleet 轉一步（原版 `1C03:0524`）。回 true 表示「動了」，對應原版
// 的回傳 1；`|角差| ≤ 回轉速度` 那條**一次轉到位而且不扣額度**。
func turnFleet(f *FleetMover, bearing int) bool {
	d := TurnFacing256(bearing, f.Facing)
	if d == 0 {
		return false
	}
	if abs32(d) <= f.TurnSpeed {
		f.Facing = (f.Facing + d) & 0xFF
		return true
	}
	if int64(f.Budget) < fleetStepCost {
		return false
	}
	step := f.TurnSpeed
	if d < 0 {
		step = -step
	}
	f.Facing = (f.Facing + step) & 0xFF
	f.Budget -= fleetStepCost
	return true
}

// AdvanceFleet 走一次「實行」（原版 `1C03:01B5`）：先轉一步，**同一次
// 呼叫接著就前進一步**，所以第一步用的是轉到一半的朝向。
//
// `passable` 回 true 表示那個點站得住（原版 `2583:040C` 回 0）。傳 nil
// 當成處處可站。回 true 表示這一次還有得做，false 表示額度用完了。
//
// ⚠ 這裡只有 `+0x48` bit4 沒設的那一支（`1C03:01B5` ＝ `1C03:0185`）。
// 短程平移（`1C03:0361`）與抵達轉向 `+0x4C` remake 還沒接：原版到了之後
// 用 `1C03:0524` 逐步轉向 `+0x4C`、轉完就回，不再扣額度；這裡到了之後
// 不轉，而且每叫一次扣 1，扣到不足一步才停（`docs/re/294`）。
func AdvanceFleet(f *FleetMover, passable func(x, y int) bool) bool {
	if f == nil {
		return false
	}
	if f.X != f.DestX || f.Y != f.DestY {
		if bearing := atan256(f.DestX-f.X, f.DestY-f.Y); bearing != f.Facing {
			turnFleet(f, bearing)
		}
	}
	if int64(f.Budget) < fleetStepCost {
		return false
	}
	nx, ny := f.DestX, f.DestY
	if Distance16(f.X, f.Y, f.DestX, f.DestY) > fleetStepCost {
		nx, ny = StepAlong(f.X, f.Y, f.Facing, BattleStep)
	}
	d := Distance16(f.X, f.Y, nx, ny)
	if d == 0 {
		d = 1 // 原版：算出來是 0 就當 1，免得原地不動又不扣額度
	}
	f.Budget -= int(d)
	if passable == nil || passable(nx, ny) {
		f.X, f.Y = nx, ny
		if f.X == f.DestX && f.Y == f.DestY {
			f.Arrived = true
		}
	}
	return true
}

// RunFleetTurn 把一回合跑完：補滿額度，然後一直「實行」到額度不足一步。
func RunFleetTurn(f *FleetMover, passable func(x, y int) bool) {
	RefillFleetBudget(f)
	for AdvanceFleet(f, passable) {
	}
}

// StationSquadrons 把一支艦隊的分隊目的地設成各自的隊形位置
// （原版每次移動之後由 `2BE1:0059` 重算，`docs/re/126`）。
//
// units 是整批分隊，idx 是屬於這一支的那幾個。旗艦（槽 20）會被擺回
// 艦隊自己的位置，所以呼叫端不必另外處理。
func StationSquadrons(f *FleetMover, s *FormationSlots, units []BattleUnit, idx []int) {
	if f == nil || s == nil {
		return
	}
	for _, i := range idx {
		if i < 0 || i >= len(units) {
			continue
		}
		u := &units[i]
		x, y := s.Station(f.X, f.Y, f.Facing, u.Slot)
		u.TargetX, u.TargetY = x, y
	}
}

// AdvanceFleetTurn 推進一支艦隊的一個回合：補滿額度、一路「實行」到
// 額度不足一步，然後把分隊的目的地重新擺到隊形位置。
//
// 分隊自己怎麼走仍然由 AdvanceBattle 負責——原版也是兩層各走各的
// （`docs/re/129`、`docs/re/126`）。
func AdvanceFleetTurn(f *FleetMover, s *FormationSlots, units []BattleUnit, idx []int,
	passable func(x, y int) bool) {
	if f == nil {
		return
	}
	RunFleetTurn(f, passable)
	StationSquadrons(f, s, units, idx)
}

// ManoeuvreBoundary 回艦隊選點時畫的那個外形（原版 `165F:004D`，
// `docs/re/135`）。回的是**相對圓心的位移**，圓心是旗艦圖塊的中心。
//
// 形狀的意思是「照回轉速率轉過去、再把剩下的預算直直走完」能到的地方：
//
//	對每個 i（0, 8, 16 …，條件 (i × 16) ÷ 回轉速度 ≤ 距離）：
//	    以每步轉「回轉速度」、走 8 像素積分到轉了 i
//	    補上不足一步的那一段
//	    剩下的預算 距離 − (i × 16) ÷ 回轉速度 沿最後的朝向直直走
//
// 轉彎在預算裡算 16／步、實際只走 8／步——轉向是雙倍代價。
// 算完之後**沿 x 軸鏡射**成兩倍的點，再整批依艦隊朝向旋轉
// （`3A92:0001`）。
//
// 積分那一圈是 **do-while，而且 `rem` 是 16 位元無號**
// （`+0AE1 inc cx` 之前那個 `jnb`）：`i = 0` 時它照樣走一步，
// 然後 `rem` 減成 `0xFFF1` 繼續繞，一直到扣回小於回轉速度為止。
// 繞完的那些整圈正弦幾乎互相抵消，所以看起來只差一兩個像素——
// 但差的地方**每個線段都差**：收據 `s3cmdrange` 上照這樣算的
// 480／501 個像素落在外形上，改成一般的 while 只剩 355
// （`docs/re/155`）。
func ManoeuvreBoundary(turnSpeed, dist, facing int) [][2]int {
	if turnSpeed <= 0 || dist <= 0 {
		return nil
	}
	var half [][2]int
	for i := 0; ; i += manoeuvreStep {
		if (i*16)/turnSpeed > dist {
			// 預算走完：原版把累加值清成 0 再存一點，
			// 也就是**多一個落在圓心的點**，然後收工。
			half = append(half, [2]int{0, 0})
			break
		}
		var sx, sy int
		a, rem := 0, i&0xFFFF
		for {
			sx += cos256(a) << 3
			sy += sin256(a) << 3
			a += turnSpeed
			rem = (rem - turnSpeed) & 0xFFFF
			if rem < turnSpeed {
				break
			}
		}
		frac := (rem << 3) / turnSpeed
		sx += frac * cos256(i)
		sy += frac * sin256(i)
		straight := dist - (i*16)/turnSpeed
		sx += straight * cos256(i)
		sy += straight * sin256(i)
		if sy < 0 {
			break // 原版：y 的累加值變負就停，這一點不存
		}
		half = append(half, [2]int{sx >> 10, sy >> 10})
	}
	out := make([][2]int, 0, len(half)*2)
	out = append(out, half...)
	for k := len(half) - 1; k >= 0; k-- {
		out = append(out, [2]int{half[k][0], -half[k][1]})
	}
	for k := range out {
		out[k][0], out[k][1] = Rotate256(out[k][0], out[k][1], facing)
	}
	return out
}

// manoeuvreStep 是外形取樣的角度間隔（原版 `add word ptr [bp+10h], 8`）。
const manoeuvreStep = 8

// ManoeuvreInnerRadius 回選點時內圈的半徑（原版 `165F:0003`）：
// `距離 ÷ 3 ＋ 8`，圓心同樣是旗艦圖塊的中心（`docs/re/120`）。
func ManoeuvreInnerRadius(dist int) int { return dist/3 + 8 }
