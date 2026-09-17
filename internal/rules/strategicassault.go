package rules

import "encoding/binary"

// 戰略打行星（`20B4:1497`，規格 `docs/spec/058`、調查 `docs/re/330`）。
//
// 占領命令走到敵方或中立星域、只有一方在那一格（星域 `+0x13` bit4）時，回合裡
// 逐格掃描（`0374:0556`）到這一格就跑一次：攻方的每一支艦隊一回合對「第一顆
// 不是自己的行星」開一火、挨一次反擊，行星每一步回復一次，最多九步。
//
// 與戰場上的打行星（`planetassault.go`，`20B4:0681`）是**不同的算式**：
// 沒有「減守軍 ÷ 6」「＋ rand % 100」、距離是 0（艦隊座標就設成行星座標）、
// 守軍取 0（`17A0:00AC` 掃的是戰場單位表，戰略階段是空的）。

// 行星記錄裡這一支用到的欄位（`docs/internals/29`、`docs/re/330` §7）。
const (
	planetDefenceMax  = 0x16 // u16 防衛力上限
	planetRecoverRate = 0x18 // u16 回復率（百分比）
	planetStocks      = 0x20 // 七個艦種的補充存量，各 u32（28 bytes）
)

// strategicMapCentre 是艦隊定位之後朝向的點（`297B:0ABF`：atan(0x3F8 − x, 0x3F4 − y)）。
const strategicMapCentreX, strategicMapCentreY = 0x3F8, 0x3F4

// strategicRounds 是一次打行星最多幾「回合」（`20B4:1512` 的 `cmp 回合, 9`）。
const strategicRounds = 9

// StrategicFleet 是攻方一支艦隊在這一趟裡要用的東西。
type StrategicFleet struct {
	Fleet *Fleet
	// Ability40／Ability41 是艦隊記錄 `+0x40`／`+0x41`（`FleetAbilities` 的第 3、4 項）：
	// 攻擊力的係數與防禦係數 ＝ `+0x40 × +0x41 ÷ 100`。
	Ability40, Ability41 int
	// Command 是帶隊提督的 `+0x0B`（統率力）：損失折算 `(200 − 統率力)`。
	Command int
}

// StrategicAssaultResult 是這一趟的副作用中不在行星記錄與艦隊結構裡的部分。
type StrategicAssaultResult struct {
	// Merit 是每位提督加的功績（提督記錄 `+0x02`，u32；`203A:0743`）。
	Merit map[int]int
	// Score 是兩個計分器各加了多少（`ds:8FE4`／`ds:8FEC` 同值，索引陣營）。
	Score [2]int
	// Captured 是拿下的行星（切片內序號，依拿下的順序）。
	Captured []int
	// Hits 是開火的次數（除錯用）。
	Hits int
}

// StrategicAssault ＝ `20B4:1512(星域, 回呼, 攻方, 0)`。
//
//   - planets 是這個星域的行星記錄（`PlanetRecord` 一筆、切片內序號 ＝ `ds:CFF8` 的序），就地改寫。
//   - sites 是同序的行星座標（`2AF5:00CB`），只用來定艦隊的朝向。
//   - fleets 是攻方在這一格的艦隊，依艦隊表索引遞增（`297B:131C(…, 星域, 1)`）。
func StrategicAssault(planets []byte, sites []PlanetSite, attacker int,
	fleets []StrategicFleet, r *Rand) StrategicAssaultResult {
	res := StrategicAssaultResult{Merit: map[int]int{}}
	n := len(planets) / PlanetRecord
	if r == nil || n == 0 {
		return res
	}
	round, passes, planet := 0, 3, 0
	for {
		switch {
		case passes > 0:
			passes--
			round++
			recoverPlanetsOneStep(planets)
		case int(planets[planet*PlanetRecord+planetOwner]) == attacker:
			planet++ // 已經是自己的：換下一顆，**不加回合**
		default:
			if strategicAttackRound(planets, planet, sites, attacker, fleets, r, &res) {
				passes = 3 // 拿下之後重新回復三步
			}
			round++
			recoverPlanetsOneStep(planets)
		}
		if round >= strategicRounds || planet >= n {
			return res
		}
	}
}

// recoverPlanetsOneStep ＝ `20B4:15E1`：這個星域**每一顆**行星回復一步
// （防衛力 ＋ 上限 × 回復率 ÷ 100，16 位元加、夾上限）。
func recoverPlanetsOneStep(planets []byte) {
	for o := 0; o+PlanetRecord <= len(planets); o += PlanetRecord {
		p := planets[o : o+PlanetRecord]
		max := int(binary.LittleEndian.Uint16(p[planetDefenceMax:]))
		rate := int(binary.LittleEndian.Uint16(p[planetRecoverRate:]))
		inc := (max * rate / 100) & 0xFFFF
		d := (int(binary.LittleEndian.Uint16(p[planetDefence:])) + inc) & 0xFFFF
		if d >= max {
			d = max
		}
		binary.LittleEndian.PutUint16(p[planetDefence:], uint16(d))
	}
}

// strategicAttackRound ＝ `20B4:1662`：攻方每一支定位到這顆行星，再依序開火；
// 有一支拿下就結束這一回合。回 true ＝ 拿下了。
func strategicAttackRound(planets []byte, pi int, sites []PlanetSite, attacker int,
	fleets []StrategicFleet, r *Rand, res *StrategicAssaultResult) bool {
	p := planets[pi*PlanetRecord : (pi+1)*PlanetRecord]
	// 第一圈：全部定位（含士氣低落的）。`0DEA:1759`：座標 ＝ 行星座標、朝向地圖中心；
	// `+0x07 ← 行星序號`、`+0x48 ← 1`（戰鬥）。
	facing := 0
	if pi < len(sites) {
		facing = AtanTable256(strategicMapCentreX-sites[pi].X, strategicMapCentreY-sites[pi].Y)
	}
	for _, sf := range fleets {
		sf.Fleet.Target = pi
		sf.Fleet.Action = 1
	}
	for _, sf := range fleets {
		f := sf.Fleet
		if f.Flags&fleetFlagLowMorale != 0 {
			continue // `172E`：士氣低落（bit0）不開火、不耗亂數
		}
		defence := int(binary.LittleEndian.Uint16(p[planetDefence:]))
		a := strategicAttack(sf)
		amt := a/2 + r.Next()%(a/3) // `20B4:0D96`：距離 0，沒有衰減
		if amt >= defence {
			amt = defence
		}
		defence -= amt
		binary.LittleEndian.PutUint16(p[planetDefence:], uint16(defence))
		res.Score[attacker] += amt / ScoreDivisor
		res.Hits++
		addMerit(res, f, amt/2)
		if strategicAttack(sf) >= defence { // `20B4:0E2C`
			max := int(binary.LittleEndian.Uint16(p[planetDefenceMax:]))
			p[planetOwner] = byte(attacker)
			for i := planetStocks; i < planetMissiles; i++ {
				p[i] = 0
			}
			binary.LittleEndian.PutUint32(p[planetMissiles:], binary.LittleEndian.Uint32(p[planetMissiles:])>>1)
			binary.LittleEndian.PutUint32(p[planetFighters:], 0)
			res.Score[attacker] += max / captureScoreDiv
			res.Captured = append(res.Captured, pi)
			addMerit(res, f, max)
			return true
		}
		strategicCounter(p, facing, sf, attacker, r, res)
	}
	return false
}

// fleetFlagLowMorale 是艦隊 `+0x00` bit0：士氣低於 50（`1261:032B`）。
const (
	fleetFlagLowMorale  = 0x01
	fleetFlagMoraleFlip = 0x04
	fleetFlagNoShips    = 0x08
)

// strategicAttack ＝ `203A:0424`：艦種 0／1 的艘數合計 × `+0x40` ÷ 100 ÷ 15 ＋ 250。
func strategicAttack(sf StrategicFleet) int {
	main := 0
	for _, g := range sf.Fleet.Composition {
		if g.Kind == 0 || g.Kind == 1 {
			main += g.Ships
		}
	}
	return main*sf.Ability40/100/15 + 250
}

// addMerit ＝ `203A:0743(艦隊, n)`：提督功績 ＋n、副官 ＋n÷2。
func addMerit(res *StrategicAssaultResult, f *Fleet, n int) {
	res.Merit[f.Admiral] += n
	if f.Adjutant >= 0 {
		res.Merit[f.Adjutant] += n >> 1
	}
}

// strategicCounter ＝ `20B4:0E7E(艦隊, 行星, 0)` ＋ `20B4:103D`：行星反擊。
// 距離 0 → 從艦隊看行星的方位是 `atan(0, 0)` ＝ 0x40，方向只看艦隊朝向。
func strategicCounter(p []byte, facing int, sf StrategicFleet, attacker int, r *Rand,
	res *StrategicAssaultResult) {
	div := 10
	if p[planetFortress] != 0 {
		div = 20
	}
	base := int(binary.LittleEndian.Uint16(p[planetDefence:])) / div // 守軍 0
	dir := Direction8(absInt(0x40 - facing))                         // `2BE1:024A`：|a − b|，byte，不折回
	base = base * assaultDirTable[dir] / 100
	def := sf.Ability40 * sf.Ability41 / 100
	if def <= 0 {
		return
	}
	amt := (base*100/def)&0xFFFF + r.Next()%100
	strategicSpread(sf, amt, attacker, res)
}

// strategicSpread ＝ `20B4:103D(艦隊, 量, 0)`：按艘數百分比分攤到各組（兩段捨去、
// 每組至少留一艘），剩下的**整筆**丟給艘數最多那一組——尾段那個「留一艘」寫進
// 了區域變數卻沒拿來用（`11C9` 寫 `[bp+2]`、`11D2` push `[bp+6]`），只受
// `203A:0593` 的 `min(艘數)` 限制。
func strategicSpread(sf StrategicFleet, amount, attacker int, res *StrategicAssaultResult) {
	groups := sf.Fleet.Composition
	total, big, bigShips := 0, -1, 0
	for i, g := range groups {
		if g.Ships == 0 {
			continue
		}
		total += g.Ships
		if g.Ships > bigShips {
			bigShips, big = g.Ships, i
		}
	}
	if total == 0 {
		return
	}
	left := amount
	for i := range groups {
		c := groups[i].Ships
		if c == 0 {
			continue
		}
		this := amount * (c * 100 / total) / 100
		if this >= c {
			this = c - 1
		}
		left -= strategicShipLoss(sf, i, this, attacker, res)
	}
	if left != 0 && big >= 0 && groups[big].Ships != 0 {
		strategicShipLoss(sf, big, left, attacker, res)
	}
}

// strategicShipLoss ＝ `203A:0593(艦艇, 量, NULL)` 在戰略層用得到的部分：扣艘數、
// 對手陣營兩個計分器 ＋損失 ÷ 20、每扣一次就重算士氣（`1261:0067` → `032B`）。
// 艦種 7 不扣、回它現有的艘數。
func strategicShipLoss(sf StrategicFleet, gi, amount, attacker int, res *StrategicAssaultResult) int {
	g := &sf.Fleet.Composition[gi]
	if g.Kind == ShipKindStandalone {
		return g.Ships
	}
	lost := amount
	if lost > g.Ships {
		lost = g.Ships
	}
	if lost < 0 {
		lost = 0
	}
	g.Ships -= lost
	strategicMorale(sf)
	opp := 1
	if sf.Fleet.Faction != 0 {
		opp = 0
	}
	res.Score[opp] += lost / ScoreDivisor
	return lost
}

// strategicMorale ＝ `1261:032B`：士氣 ＝ 各組 `(滿編 − min(損失 × (200 − 統率力) ÷ 100, 滿編))
// × 100 ÷ 滿編` 的平均（只算還有艘數的組）；小於 50 設 bit0，bit0 翻轉時設 bit2，
// 全隊沒艘數設 bit3。
func strategicMorale(sf StrategicFleet) {
	f := sf.Fleet
	sum, n := 0, 0
	for _, g := range f.Composition {
		if g.Ships == 0 || g.Full == 0 {
			continue
		}
		t := (g.Full - g.Ships) * (200 - sf.Command) / 100
		if t > g.Full {
			t = g.Full
		}
		sum += ((g.Full - t) * 100 / g.Full) & 0xFF
		n++
	}
	if n == 0 {
		f.Flags |= fleetFlagNoShips
		return
	}
	f.Flags &^= fleetFlagNoShips
	old := f.Flags
	f.Morale = (sum / n) & 0xFF
	if f.Morale < 50 {
		f.Flags |= fleetFlagLowMorale
	} else {
		f.Flags &^= fleetFlagLowMorale
	}
	if (old^f.Flags)&fleetFlagLowMorale != 0 {
		f.Flags |= fleetFlagMoraleFlip
	}
}

// atanTable 是 `ds:593A`：tan × 1024 截斷，65 項（`2BE1:01E1` 找第一個 `比值 <= 表[k]` 的 k）。
var atanTable = [...]int{0, 25, 50, 75, 100, 126, 151, 177, 203, 229, 256, 283, 310, 338, 366, 395,
	424, 453, 484, 515, 547, 580, 613, 648, 684, 721, 759, 799, 840, 883, 928, 974, 1023, 1075,
	1129, 1187, 1247, 1312, 1380, 1453, 1532, 1617, 1708, 1807, 1915, 2034, 2165, 2310, 2472,
	2654, 2861, 3099, 3375, 3700, 4088, 4560, 5147, 5901, 6903, 8302, 10396, 13882, 20843, 41713}

// AtanTable256 ＝ `2BE1:011D(dx, dy)`：查表版的方位角（一整圈 256）。
//
// 與 `Atan256`（戰場層用的 sin／cos 二分法）不完全等價——|dx|、|dy| 在 0..599 裡約 8,130 組
// 差 1（`docs/re/330` §3.9）；戰略層定朝向照原版查表。
func AtanTable256(dx, dy int) int {
	ratio := 0xFFFF
	if ax := absInt(dx); ax != 0 {
		ratio = (absInt(dy) << 10) / ax
		if ratio > 0xFFFF {
			ratio = 0xFFFF
		}
	}
	a := atanLimit
	for k, v := range atanTable[:atanLimit] {
		if ratio <= v {
			a = k
			break
		}
	}
	switch {
	case dx < 0 && dy >= 0:
		return (0x80 - a) & 0xFF
	case dx < 0 && dy < 0:
		return (a + 0x80) & 0xFF
	case dx >= 0 && dy < 0:
		return (-a) & 0xFF
	}
	return a
}

// atanLimit 是 `2BE1:01E1` 找表的上限（找不到就回 64 ＝ 90°）。
const atanLimit = 64

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
