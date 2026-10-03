package rules

import "encoding/binary"

// BattleUnit 是戰術畫面上的一支**分隊**。
//
// 一支艦隊進了戰場會散成最多十個分隊，各自有座標、各自移動；分隊就是
// 艦艇池裡的一筆記錄（`docs/internals/38`）：
//
//	+0x34／+0x36  X／Y              現在的位置（研究371）
//	+0x28／+0x2A  TargetX／TargetY  要去的集結點；剛佈陣時與現在位置相同
//
// 艦隊記錄的 `+0x5C` 是第 0 個分隊（旗艦）那一份的複本，右側狀態欄讀它。
//
// 座標空間比看得到的戰場（448×360）大很多：整張戰場 128×102 格、
// 一格 16×20，所以是 2048×2040，右上角小地圖上的白框就是視野。
type BattleUnit struct {
	Fleet       int `json:"fleet"`        // 屬於戰略地圖上的哪一支艦隊
	FleetNumber int `json:"fleet_number"` // 畫面上的艦隊編號（1 起）
	Squadron    int `json:"squadron"`     // 在艦隊裡的第幾個分隊，0 ＝ 旗艦
	Faction     int `json:"faction"`      // 0 帝國、1 同盟
	X           int `json:"x"`            // 現在的位置
	Y           int `json:"y"`
	TargetX     int `json:"target_x"` // 要去的點
	TargetY     int `json:"target_y"`
	Speed       int `json:"speed"`      // 一回合走多遠
	Strength    int `json:"strength"`   // 戰力（`StrengthWeights` 算的）
	Kind        int `json:"kind"`       // 艦種（艦艇記錄 `+0x04`）
	Detail      int `json:"detail"`     // 顯示層級（艦艇記錄 `+0x17`）
	Ships       int `json:"ships"`      // 現有艘數（艦艇記錄 `+0x08`）
	FullShips   int `json:"full_ships"` // 建制滿編艘數（艦艇記錄 `+0x0A`）

	// RepairCap 是「修理」補得回來的上限（艦艇記錄 `+0x0C`）。它與
	// `+0x0A` 不同——戰損會把它扣掉一點（永久損失），`+0x0A` 不動。
	RepairCap int `json:"repair_cap"`
	Attack    int `json:"attack"` // 攻擊力（艦艇記錄 `+0x16`）
	Morale    int `json:"morale"` // 士氣（艦艇記錄 `+0x06`）

	// Slot 是隊形槽索引 0–40（艦艇記錄 `+0x05`）：目的地 ＝ 艦隊位置 ＋
	// `ds:59BA` 第 N 筆偏移依艦隊朝向旋轉。「編隊」改的就是它
	// （`docs/internals/46`）。
	Slot int `json:"slot"`

	// TurnSpeed／CruiseSpeed／FightSpeed 是艦艇記錄的 `+0x1F`／`+0x21`／
	// `+0x22`。艦隊的回轉／巡航／戰鬥速度就是**各分隊取最小值**
	// （`0x2A7AE` 那個迴圈，寫進艦隊記錄 `+0x4D`／`+0x4E`／`+0x4F`；
	// `docs/re/112`）。
	TurnSpeed   int  `json:"turn_speed"`
	CruiseSpeed int  `json:"cruise_speed"`
	FightSpeed  int  `json:"fight_speed"`
	Out         bool `json:"out"` // 出局，不進攻擊名單（艦艇記錄 `+0x1E` bit7）

	// Marked ＝ 艦艇記錄 `+0x19` bit7：這個分隊要在小地圖上點出來。
	// 「索敵」放偵察隊出去時設起來（`1CC5:0966`）；原版接著由
	// `1727:0207` 依陣營挑顏色（帝國 5、同盟 9）XOR 上去，並把 bit6
	// （目前點著）與顏色記回 `+0x19`，位置記進 `+0x1A`／`+0x1C`——
	// 擦的時候 `1727:0298` 要用同一個顏色 XOR 回原處
	// （`docs/re/154`、`docs/internals/38`）。
	Marked bool `json:"marked"`

	// Missiles／Fighters 是艦艇記錄 `+0x0E`／`+0x12` 的 32 位元值：
	// 現有的飛彈與戰鬥機數。分隊面板印的就是它們，顏色由「滿不滿」決定
	// （`docs/re/118`）。
	Missiles int `json:"missiles"`
	Fighters int `json:"fighters"`

	// Action 是艦艇記錄 `+0x1E` 整個位元組：分隊自己的行動
	// （0 移動／1 編隊／2 索敵／3 歸隊，0x7F ＝ 沒有）。
	// 戰場命令選單用它決定「移動」可不可選（`docs/spec/023`）。
	Action int `json:"action"`

	// 規格080：分隊固定扣8.0，不能由整數Speed重建剩餘額度。
	Budget      int `json:"budget,omitempty"`
	MarkerFlags int `json:"marker_flags,omitempty"`
	MarkerX     int `json:"marker_x,omitempty"`
	MarkerY     int `json:"marker_y,omitempty"`

	// Seen／DetailHold 對應艦艇 `+0x16` bit0 與 `+0x18`。Seen 表示這一輪
	// 已被敵方觀測；DetailHold 是顯示層級 3 還要維持幾輪（`docs/re/290`）。
	Seen       bool `json:"seen,omitempty"`
	DetailHold int  `json:"detail_hold,omitempty"`
}

// EngagementRange 回一個分隊的射程（`docs/spec/018`）。
//
// 原版 `2350:0003`：`+0x16 ÷ 200 ＋ 80`。挑攻擊名單時用的門檻是它的兩倍。
func EngagementRange(attack int) int {
	if attack < 0 {
		attack = 0
	}
	return attack/rangeDivisor + rangeBase
}

const (
	rangeDivisor = 200
	rangeBase    = 80
	// RangeFactor 是「攻擊名單的門檻 ＝ 射程 × 2」（`17A0:00AC` 的呼叫端）。
	RangeFactor = 2
)

// InRange 回名單裡「陣營相符、沒出局、而且目的地離 (x,y) 不超過 limit」
// 的分隊索引（原版 `17A0:00AC`，`docs/spec/018`）。
//
// 比的是分隊的**目的地** `TargetX`／`TargetY`，不是現在的位置——原版讀的
// 是艦艇記錄的 `+0x34`／`+0x36`。
func InRange(units []BattleUnit, x, y, limit, faction int) []int {
	var out []int
	for i, u := range units {
		if u.Faction != faction || u.Out {
			continue
		}
		dx, dy := u.TargetX-x, u.TargetY-y
		if dx*dx+dy*dy <= limit*limit {
			out = append(out, i)
		}
	}
	return out
}

// BattleSprite 回一個分隊要畫 `SPRITE.SPR` 的第幾塊。
//
// 照原版 `25DB:03E8` 的算式（`docs/internals/38`）：陣營決定基底
// （帝國 20、同盟 50），顯示層級決定加多少。
func BattleSprite(faction, detail, kind int) int {
	base := AllianceSpriteBase
	if faction == 0 {
		base = EmpireSpriteBase
	}
	if kind == 7 {
		return BlankSprite
	}
	switch detail {
	case 1:
		return base + spriteRoughOffset
	case 2:
		if kind < 0 {
			return base + spriteRoughOffset
		}
		return base + spriteDetailOffset + kind
	}
	return BlankSprite
}

// BattleUnitSprite 是 `25DB:03E8` 的完整戰場分支。己方（viewer）或已到
// Detail 3 時，旗艦畫艦隊編號，其餘畫艦種；敵方 Detail 1／2 才走概略／
// 詳細圖示（`docs/re/85`）。
func BattleUnitSprite(u BattleUnit, viewer int) int {
	// 正式可見清單只含已放出偵察隊，與一般分隊同用己方force（26E5:0194／087）。
	return BattleCachedSprite(u, u.Faction == viewer, BlankSprite)
}

// BattleCachedSprite重算25DB:03E9的SHIP+2F。非完整支線的其他Detail保留舊byte。
// force由caller決定，不能用診斷force1替代正式偵察隊設定。
func BattleCachedSprite(u BattleUnit, force bool, previous int) int {
	base := AllianceSpriteBase
	if u.Faction == 0 {
		base = EmpireSpriteBase
	}
	if force || u.Detail == SpottingMax {
		if u.Squadron == 0 {
			return int(byte(base + u.FleetNumber - 1))
		}
		return int(byte(base + spriteDetailOffset + u.Kind))
	}
	if u.Kind == 7 || u.Detail == 0 {
		return BlankSprite
	}
	switch u.Detail {
	case 1:
		return base + spriteRoughOffset
	case 2:
		return int(byte(base + spriteDetailOffset + u.Kind))
	}
	return int(byte(previous))
}

// `SPRITE.SPR` 裡單位圖示的位置（`docs/re/85`）。
const (
	EmpireSpriteBase   = 20 // 陣營 0 的基底
	AllianceSpriteBase = 50 // 陣營 1 的基底
	BlankSprite        = 9  // 不顯示
	spriteDetailOffset = 16 // 詳細：基底 ＋ 16 ＋ 艦種
	spriteRoughOffset  = 24 // 概略：基底 ＋ 24
)

// DefaultDetail 是佈陣時給的顯示層級。原版佈陣把 `+0x17` 清成 0，
// 之後才由每回合索敵逐級升降（`docs/re/290`）。
const DefaultDetail = 0

// FormationSpacing 是舊的十格近似表使用的分隊間距（`docs/internals/38`）。
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
// 方向本身由 BattleFacing 算：原版量到的十二支帝國艦隊，基底的方向
// 都是「指向戰場中心的角度減 96」（`docs/internals/39`）。同盟那幾支
// 這支只保留給 AI 的合成測試；正式佈陣改由 Deploy 讀每支分隊既有的
// `+0x05` 槽索引，再走原版同一張 41 槽表（`2BE1:0000`／`:0059`）。
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

// BattleStep 是戰場上走一步的長度（`docs/re/102`）。
//
// 原版不是「一步跳到目的地」：`1C03:01B6` 沿方位每次前進 8 像素，
// 每一步都先問地形與碰撞才走。直線通行判定（`1588:0891`）用的也是這個 8。
const BattleStep = 8

// DefaultBattleSpeed 是一回合的行程，單位是像素。
//
// 原版存在艦隊記錄 `+0x52`／`+0x54`（16.16 定點），每回合開頭複製到
// `+0x56`／`+0x58` 當餘額，每走一步扣掉實際距離，**剩不到 8 就停**。
// 那個值 ＝ `+0x4E × 8`——`+0x4E` 是這一支的每回合步數，收據 `s3tgt`
// 十八支全部對得上（`docs/re/102`）。
//
// 30 是收據裡同盟第 16 艦隊量到的步數（240 像素）。**`+0x4E` 自己怎麼
// 算還沒解完**：唯一的寫入點取「各分隊 `艦種列[+6] × 艦隊能力[+0x42]
// ÷ 80` 的最小值」，其中的艦種列還沒對出來。
const DefaultBattleSpeed = 30 * BattleStep

// Moving 回這一支還沒走到目標。
func (u BattleUnit) Moving() bool { return u.X != u.TargetX || u.Y != u.TargetY }

// AdvanceBattle 讓一支分隊走完這一回合的行程，回有沒有動。
//
// 照原版 `1C03:01B6`：沿著指向目標的方位每次走 `BattleStep`，從行程
// 餘額扣掉實際距離，剩不到一步就停；每一步先問地形（`2583:040C` →
// `1588:04EA`），不能站就停在原地不再往前。
//
// terrain／cells 給 nil 就不判地形——測試與還沒載到原版地圖時用。
//
// ⚠ 這是分隊層的近似。原版艦隊層有兩支實行常式（`docs/re/294`）：
// `+0x48` bit4 沒設的 `1C03:0185` 先轉向目的地方位、沿現在的朝向 `+0x4A`
// 走；bit4 有設的 `1C03:0361` 不轉身、每一步直指目的地平移。兩支都在
// 到了之後才轉向抵達朝向 `+0x4C`。艦隊層照原版的那一支是 `AdvanceFleet`；
// 這裡直接朝目標走。
func AdvanceBattle(u *BattleUnit, t *BattleTerrain, cells []uint8) bool {
	if u == nil || !u.Moving() {
		return false
	}
	budget := u.Speed
	if budget <= 0 {
		budget = DefaultBattleSpeed
	}
	// 餘額用 8.8 定點算：一步 (−4,−7) 的長度是 8.06，截成 8 會多走一步
	// ——原版是 29 步，截斷會變 30 步。
	budget *= fixStep
	moved := false
	for budget >= BattleStep*fixStep {
		dx, dy := u.TargetX-u.X, u.TargetY-u.Y
		d := isqrt(dx*dx + dy*dy)
		if d < BattleStep {
			// 原版在這裡把「到了」的旗標設起來就結束這一輪。
			u.X, u.Y = u.TargetX, u.TargetY
			return true
		}
		a := atan256(dx, dy)
		nx := u.X + roundDiv(cos256(a)*BattleStep, 1024)
		ny := u.Y + roundDiv(sin256(a)*BattleStep, 1024)
		if nx == u.X && ny == u.Y {
			break
		}
		if t != nil && cells != nil && !t.Passable(cells, nx, ny) {
			break
		}
		sx, sy := nx-u.X, ny-u.Y
		budget -= isqrt((sx*sx + sy*sy) * fixStep * fixStep)
		u.X, u.Y = nx, ny
		moved = true
	}
	return moved
}

// fixStep 是走一步扣行程時用的定點倍率。
const fixStep = 256

// roundDiv 是四捨五入的整數除法（往絕對值大的方向進位）。
//
// 直接截斷會把每一步縮短：收據裡量到的一步是 (−4,−7)，長度 8.06；
// 截斷會得到 (−3,−7)。
func roundDiv(a, b int) int {
	if a < 0 {
		return -((-a + b/2) / b)
	}
	return (a + b/2) / b
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
func (s *Slot) Deploy(w StrengthWeights, formation *FormationSlots,
	f Fleet, x, y, facing int) []BattleUnit {
	out := s.BattleUnitsFor(w, f)
	for i := range out {
		u := &out[i]
		u.X, u.Y = formation.Station(x, y, facing, u.Slot)
		u.TargetX, u.TargetY = u.X, u.Y
	}
	return out
}

// BattleUnitsFor從活艦艇池取得已配置的兩種座標；正常入口不重新佈陣（088）。
func (s *Slot) BattleUnitsFor(w StrengthWeights, f Fleet) []BattleUnit {
	if s == nil {
		return nil
	}
	ships := s.FleetGroupShips(f.Index)
	if len(ships) == 0 {
		return nil
	}
	if len(ships) > len(formationSlots) {
		ships = ships[:len(formationSlots)]
	}
	out := make([]BattleUnit, 0, len(ships))
	for n, ship := range ships {
		u := BattleUnit{
			Fleet:       f.Index,
			FleetNumber: f.Number,
			Squadron:    n,
			Faction:     f.Faction,
			Speed:       DefaultBattleSpeed,
			Kind:        -1,
			Detail:      DefaultDetail,
		}
		s.fillFromShip(&u, ship, w)
		rec := s.Ship[ship*ShipRecord : (ship+1)*ShipRecord]
		u.X, u.Y = int(binary.LittleEndian.Uint16(rec[0x34:])), int(binary.LittleEndian.Uint16(rec[0x36:]))
		u.TargetX, u.TargetY = int(binary.LittleEndian.Uint16(rec[0x28:])), int(binary.LittleEndian.Uint16(rec[0x2A:]))
		out = append(out, u)
	}
	return out
}

// 戰場中心的參考點與角度的一整圈（`docs/internals/39`）。
//
// 整張戰場是 2048×2040，中心是 (1024,1020)；原版用的是 (1016,1012)，
// 兩邊各少 8。這裡照原版的數字，不改成幾何中心。
const (
	BattleCentreX = 1016
	BattleCentreY = 1012
	BattleTurn    = 256 // 角度的一整圈
	facingOffset  = 96  // 基底方向 ＝ 朝向 − 96（＝ 135°）
)

// BattleFacing 回一支艦隊在 (x,y) 的朝向，一整圈是 256。
//
// 原版是 `2BE1:011D` 的 atan2（`docs/internals/39`）：拿
// (BattleCentreX − x, BattleCentreY − y) 算角度。這裡用整數運算重算，
// 與原版量到的十八支全部相同或差 1。
func BattleFacing(x, y int) int {
	return atan256(BattleCentreX-x, BattleCentreY-y)
}

// FormationBase 回佈陣用的基底向量：長度一個間距、方向是朝向減 96。
func FormationBase(facing int) (int, int) {
	a := ((facing-facingOffset)%BattleTurn + BattleTurn) % BattleTurn
	return cos256(a) * FormationSpacing / 1024, sin256(a) * FormationSpacing / 1024
}

// atan256 是一整圈 256 的 atan2，**只用整數**。
//
// 原版與 remake 都查反正切表；舊二分逼近的差一格已依規格 078 修正。
func atan256(dx, dy int) int {
	// 戰場與戰略都呼叫原版 `2BE1:011D`，以 ds:593A 的截斷比值查表。
	// sin／cos 二分法在 (1,1)、(4,1) 等點差一格；規格 078、研究 366。
	return AtanTable256(dx, dy)
}

func abs32(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// fillFromShip 把艦艇池第 ship 筆的欄位填進一個分隊。ship < 0 或超出
// 範圍就什麼都不做（`docs/internals/38`）。
func (s *Slot) fillFromShip(u *BattleUnit, ship int, w StrengthWeights) {
	if ship < 0 || (ship+1)*ShipRecord > len(s.Ship) {
		return
	}
	rec := s.Ship[ship*ShipRecord : (ship+1)*ShipRecord]
	u.Kind = int(rec[4])
	u.Slot = int(rec[5])
	u.Ships = int(binary.LittleEndian.Uint16(rec[8:]))
	u.FullShips = int(binary.LittleEndian.Uint16(rec[10:]))
	u.RepairCap = int(binary.LittleEndian.Uint16(rec[12:]))
	u.Attack = int(binary.LittleEndian.Uint16(rec[0x16:]))
	u.Seen, u.Detail, u.DetailHold = rec[0x16]&1 != 0, int(rec[0x17]), int(rec[0x18])
	// 士氣是**一個位元組**：`+0x07` 固定 0xFF，整個 word 讀出來
	// 會是 0xFF64。用 -poke 把 0x682EC 寫成 77 驗過，畫面上只有
	// 那一列的士氣跟著變（收據 `s3morale`，`docs/re/111`）。
	u.Morale = int(rec[6])
	u.Missiles = int(binary.LittleEndian.Uint32(rec[0x0E:]))
	u.Fighters = int(binary.LittleEndian.Uint32(rec[0x12:]))
	u.Action = int(rec[0x1E])
	u.Out = u.Action&0x80 != 0
	u.Budget = int(int32(binary.LittleEndian.Uint32(rec[0x24:])))
	u.MarkerFlags = int(rec[0x19])
	u.Marked = u.MarkerFlags&0x80 != 0
	u.MarkerX = int(binary.LittleEndian.Uint16(rec[0x1A:]))
	u.MarkerY = int(binary.LittleEndian.Uint16(rec[0x1C:]))
	u.TurnSpeed = int(rec[0x1F])
	u.CruiseSpeed = int(rec[0x21])
	u.FightSpeed = int(rec[0x22])
	if w != nil {
		u.Strength = w.ShipStrength(rec)
	}
}

// ScoutUnit 做出艦隊的索敵分遣隊（艦隊記錄 `+0x38` 那一筆），擺在
// (x,y)。它平常不在戰場上——戰場的「索敵」才會把它放出來
// （`1CC5:0966`，`docs/re/132`）。第二個回傳值 false ＝ 這一支沒有。
func (s *Slot) ScoutUnit(w StrengthWeights, f Fleet, x, y int) (BattleUnit, bool) {
	if f.Scout == 0xFFFF || f.Scout < 0 {
		return BattleUnit{}, false
	}
	u := BattleUnit{
		Fleet: f.Index, FleetNumber: f.Number, Squadron: scoutSquadron, Faction: f.Faction,
		X: x, Y: y, TargetX: x, TargetY: y,
		Speed: DefaultBattleSpeed, Kind: -1, Detail: DefaultDetail,
	}
	s.fillFromShip(&u, f.Scout, w)
	if u.Kind < 0 {
		return BattleUnit{}, false
	}
	return u, true
}

// scoutSquadron 是索敵分遣隊在 remake 裡的分隊序號。原版沒有這個編號
// ——它不在編制表裡，用一個不會撞到的值標出來。
const scoutSquadron = -1
