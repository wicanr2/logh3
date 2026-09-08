package rules

// BattleUnit 是戰術畫面上的一支部隊。
//
// 對應原版執行期艦隊記錄裡那兩組座標（`docs/internals/33`、`docs/re/80`）：
//
//	+0x5C／+0x5E  X／Y      現在的位置
//	+0x64／+0x66  TargetX／TargetY  要去的點（剛進戰鬥時與現在位置相同）
//
// 座標空間比看得到的戰場（448×360）大很多：量到的值落在 0–2000 之間，
// 右上角小地圖上的白框就是視野。
type BattleUnit struct {
	Fleet    int `json:"fleet"`    // 對應戰略地圖上的哪一支艦隊
	Faction  int `json:"faction"`  // 0 帝國、1 同盟
	X, Y     int `json:"x"`        // 現在的位置
	TargetX  int `json:"target_x"` // 要去的點
	TargetY  int `json:"target_y"`
	Speed    int `json:"speed"`    // 一回合走多遠
	Strength int `json:"strength"` // 戰力（`StrengthWeights` 算的）
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
