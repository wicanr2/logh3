package rules

// BattleUnit 是戰術畫面上的一支**分隊**。
//
// 一支艦隊進了戰場會散成最多十個分隊，各自有座標、各自移動；分隊就是
// 艦艇池裡的一筆記錄（`docs/internals/38`）：
//
//	+0x28／+0x2A  X／Y              現在的位置
//	+0x34／+0x36  TargetX／TargetY  要去的點；剛佈陣時與現在位置相同
//
// 艦隊記錄的 `+0x5C` 是第 0 個分隊（旗艦）那一份的複本，右側狀態欄讀它。
//
// 座標空間比看得到的戰場（448×360）大很多：整張戰場 128×102 格、
// 一格 16×20，所以是 2048×2040，右上角小地圖上的白框就是視野。
type BattleUnit struct {
	Fleet    int `json:"fleet"`     // 屬於戰略地圖上的哪一支艦隊
	Squadron int `json:"squadron"`  // 在艦隊裡的第幾個分隊，0 ＝ 旗艦
	Faction  int `json:"faction"`   // 0 帝國、1 同盟
	X        int `json:"x"`         // 現在的位置
	Y        int `json:"y"`
	TargetX  int `json:"target_x"` // 要去的點
	TargetY  int `json:"target_y"`
	Speed    int `json:"speed"`     // 一回合走多遠
	Strength int `json:"strength"`  // 戰力（`StrengthWeights` 算的）
}

// FormationSpacing 是分隊之間的間距（`docs/internals/38`）。
//
// 量到的基底向量長度都落在 22–23。
const FormationSpacing = 23

// formationSlots 是分隊佔的格點，(a,b) ＝ a 個間距沿朝向、b 個沿它的法線。
//
// 分隊少的艦隊取前面幾個。最後兩個格點的順序在觀察到的局面裡會對調，
// 分配規則還沒解出來（`docs/internals/38`）。
var formationSlots = [10][2]int{
	{0, 0}, {1, 0}, {0, 1}, {2, 0}, {0, 2},
	{3, 0}, {2, 1}, {1, 2}, {0, 3}, {1, 1},
}

// Formation 回第 n 個分隊相對旗艦的位移。
//
// (ux,uy) 是艦隊的朝向，長度就是一個間距——原版量到的值都是 22 或 23
// （`docs/internals/38`）。法線取 (uy,−ux)，與原版第 2 個分隊的位移同一側。
//
// 位移是基底的整數倍。原版在 a 或 b 到 3 的時候會比整數倍少 1，
// 看得出它是沿同一個角度取距離、每一步各自取整；十個格點裡有八個
// 用整數倍就完全對得上，剩下兩個差 1。
func Formation(n, ux, uy int) (int, int) {
	if n <= 0 || n >= len(formationSlots) {
		return 0, 0
	}
	a, b := formationSlots[n][0], formationSlots[n][1]
	return a*ux + b*uy, a*uy - b*ux
}

// DefaultBattleSpeed 是一回合走多遠的暫定值。
//
// **還沒釘死**：`docs/re/80` 量到兩支分別是每回合約 233 與 325，
// 看起來跟機動力（艦隊記錄 `+0x3C`）有關，但只有兩個資料點。
const DefaultBattleSpeed = 256

// Moving 回這一支還沒走到目標。
func (u BattleUnit) Moving() bool { return u.X != u.TargetX || u.Y != u.TargetY }

// AdvanceBattle 把一支部隊往目標推進一步，回有沒有動。
//
// 走的是直線：不夠一步就直接到位。原版的路徑細節還沒逆向，
// 但量到的位移確實指著目標的方向（`docs/re/80`）。
func AdvanceBattle(u *BattleUnit) bool {
	if u == nil || !u.Moving() {
		return false
	}
	speed := u.Speed
	if speed <= 0 {
		speed = DefaultBattleSpeed
	}
	dx, dy := u.TargetX-u.X, u.TargetY-u.Y
	d := isqrt(dx*dx + dy*dy)
	if d <= speed {
		u.X, u.Y = u.TargetX, u.TargetY
		return true
	}
	u.X += dx * speed / d
	u.Y += dy * speed / d
	return true
}

// isqrt 是整數平方根。**故意不用浮點數**：這一層的每一個數字都要能
// 逐位元組重現，浮點在不同平台會有最後一位的差異。
func isqrt(n int) int {
	if n <= 0 {
		return 0
	}
	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	if x == 0 {
		return 1
	}
	return x
}

// BattleStrength 回某一方在場上的戰力總和。
func BattleStrength(units []BattleUnit, faction int) int {
	n := 0
	for _, u := range units {
		if u.Faction == faction {
			n += u.Strength
		}
	}
	return n
}

// Deploy 把一支艦隊佈到戰場上，回它散出來的分隊。
//
// 旗艦放在 (x,y)，其餘分隊照 Formation 排在它周圍；(dx,dy) 是艦隊的
// 朝向。分隊數 ＝ 艦隊指到的艦艇筆數，最多十個（`docs/internals/38`）。
//
// 戰力平均分給每個分隊——原版是每個分隊各自帶自己那筆艦艇的艘數，
// 這裡先用艦隊總戰力除以分隊數，等分隊層的戰鬥結算解出來再換掉。
func (s *Slot) Deploy(w StrengthWeights, f Fleet, x, y, dx, dy int) []BattleUnit {
	ships := s.FleetShips(f.Index)
	if len(ships) == 0 {
		return nil
	}
	if len(ships) > len(formationSlots) {
		ships = ships[:len(formationSlots)]
	}
	out := make([]BattleUnit, 0, len(ships))
	for n, ship := range ships {
		ox, oy := Formation(n, dx, dy)
		u := BattleUnit{
			Fleet:    f.Index,
			Squadron: n,
			Faction:  f.Faction,
			X:        x + ox,
			Y:        y + oy,
			TargetX:  x + ox,
			TargetY:  y + oy,
			Speed:    DefaultBattleSpeed,
		}
		if w != nil && ship >= 0 && (ship+1)*ShipRecord <= len(s.Ship) {
			u.Strength = w.ShipStrength(s.Ship[ship*ShipRecord : (ship+1)*ShipRecord])
		}
		out = append(out, u)
	}
	return out
}
