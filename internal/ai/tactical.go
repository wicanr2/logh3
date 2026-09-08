package ai

import "github.com/wicanr2/logh3/internal/rules"

// Move 是戰術層的一道命令：叫第 Unit 支部隊往 (X,Y) 走。
//
// 對應原版的 `+0x64`／`+0x66`（`docs/re/80`）。
type Move struct {
	Unit int `json:"unit"`
	X    int `json:"x"`
	Y    int `json:"y"`
}

// Battlefield 是戰術層看得到的局面。**只讀**。
type Battlefield struct {
	Units   []rules.BattleUnit
	Planets []rules.BattlePlanet
}

// enemyPlanet 回離 (x,y) 最近的敵方行星，沒有就回 false。
//
// 原版量到的目標點裡有一個是 **(1320,812)**——那正是雷奧米爾的座標
// （`docs/re/80` 的艦隊 28、`docs/re/99` 的行星表）。所以戰術層會把
// 艦隊帶向敵方的行星，不是只追著敵艦跑。
func enemyPlanet(b Battlefield, faction, x, y int) (rules.BattlePlanet, bool) {
	best, found, bestD := rules.BattlePlanet{}, false, 0
	for _, p := range b.Planets {
		if p.Owner == faction || (p.Owner != 0 && p.Owner != 1) {
			continue
		}
		dx, dy := p.X-x, p.Y-y
		d := dx*dx + dy*dy
		if !found || d < bestD {
			best, found, bestD = p, true, d
		}
	}
	return best, found
}

// Tactic 是戰術層的一套決策。
type Tactic interface {
	Moves(b Battlefield, faction, turn int) []Move
}

// OriginalTactic 照原版：**整批朝同一個點壓過去**。
//
// 這不是猜的——`docs/re/80` 量到一場六對一的戰鬥裡，敵方六支有四支的
// 目標點是同一個座標，而且整批朝我方移動；**另一支指向 (1320,812)，
// 那是行星雷奧米爾的座標**（`docs/re/99`）。所以目標優先取敵方的行星，
// 沒有行星才退回最近的敵方單位。
type OriginalTactic struct{}

// Moves 回這一回合的命令。
func (OriginalTactic) Moves(b Battlefield, faction, turn int) []Move {
	mine, foes := split(b, faction)
	if len(mine) == 0 {
		return nil
	}
	tx, ty, ok := 0, 0, false
	if len(mine) > 0 {
		u := b.Units[mine[0]]
		if p, found := enemyPlanet(b, faction, u.X, u.Y); found {
			tx, ty, ok = p.X, p.Y, true
		}
	}
	if !ok {
		if len(foes) == 0 {
			return nil
		}
		t := b.Units[nearestFoe(b, mine, foes)]
		tx, ty = t.X, t.Y
	}
	var out []Move
	for _, i := range mine {
		out = append(out, Move{Unit: i, X: tx, Y: ty})
	}
	return out
}

// Encircle 是強化版的戰術：**集火 ＋ 繞側後 ＋ 撤退門檻**
// （`docs/spec/017`）。
//
// 與原版的差別只有三件事：
//
//   - 集火挑的是**最好打的**那一支（戰力最低），不是最近的；
//   - 每一支分到目標周圍**不同方位**的一個點，從幾個方向同時接觸，
//     而不是排成一路壓上去；
//   - 己方戰力掉到開場的一定比例以下就撤——不打到全滅。
type Encircle struct {
	// Difficulty 0–3：越大圍得越緊、越晚撤。
	Difficulty int

	// Radius 是圍上去時停在目標周圍多遠。0 ＝ 用預設。
	Radius int

	start map[int]int // 開場時各方的戰力，用來算撤退門檻
}

// 圍上去的八個方位（8 分之一圓，避免浮點數）。
var bearings = [8][2]int{
	{10, 0}, {7, 7}, {0, 10}, {-7, 7}, {-10, 0}, {-7, -7}, {0, -10}, {7, -7},
}

const defaultRadius = 200

// retreatAt 回「己方戰力掉到開場的幾成就撤」。難度越高撐越久。
func (e *Encircle) retreatAt() int {
	n := 40 - e.Difficulty*10
	if n < 10 {
		n = 10
	}
	return n
}

// Moves 回這一回合的命令。
func (e *Encircle) Moves(b Battlefield, faction, turn int) []Move {
	if e.start == nil {
		e.start = map[int]int{}
	}
	if _, ok := e.start[faction]; !ok {
		e.start[faction] = rules.BattleStrength(b.Units, faction)
	}

	mine, foes := split(b, faction)
	if len(mine) == 0 || len(foes) == 0 {
		return nil
	}

	// 撐不住就撤：往遠離敵方重心的方向走，不打到全滅。
	now := rules.BattleStrength(b.Units, faction)
	if e.start[faction] > 0 && now*100 < e.start[faction]*e.retreatAt() {
		return e.retreat(b, mine, foes)
	}

	// 集火：挑戰力最低的那一支，整場只打它。
	t := b.Units[weakestFoe(b, foes)]
	radius := e.Radius
	if radius <= 0 {
		radius = defaultRadius
	}
	// 難度越高圍得越緊。
	radius -= e.Difficulty * 30
	if radius < 40 {
		radius = 40
	}

	// 方位不是隨便挑的：站在不同方位打過去，原版會乘一個 80–100 的
	// 方向係數（`docs/re/88`）。先把八個方位按係數由高到低排，
	// 分隊依序去佔——**先佔到划算的位置**。
	order := bearingOrder(t.X, t.Y, radius)

	var out []Move
	for n, i := range mine {
		// 每一支分到不同方位——**這就是包圍**：一路壓上去只會排成
		// 一條線，繞到不同方位才會同時從幾個方向接觸。
		b := bearings[order[n%len(order)]]
		out = append(out, Move{
			Unit: i,
			X:    t.X + b[0]*radius/10,
			Y:    t.Y + b[1]*radius/10,
		})
	}
	return out
}

// bearingOrder 把八個方位按「從那裡打目標拿到的方向係數」由高到低排。
//
// 係數相同的維持原本的順序，這樣同一個局面每次算出來都一樣——
// 決定性是對拍的前提。
func bearingOrder(tx, ty, radius int) []int {
	order := make([]int, len(bearings))
	factor := make([]int, len(bearings))
	for i, b := range bearings {
		px, py := tx+b[0]*radius/10, ty+b[1]*radius/10
		factor[i] = rules.DamageFactor(rules.BattleFacing(px, py),
			rules.Bearing(px, py, tx, ty))
		order[i] = i
	}
	// 插入排序，穩定且不需要額外套件。
	for i := 1; i < len(order); i++ {
		for j := i; j > 0 && factor[order[j]] > factor[order[j-1]]; j-- {
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	return order
}

// retreat 讓整批往遠離敵方重心的方向退。
func (e *Encircle) retreat(b Battlefield, mine, foes []int) []Move {
	cx, cy := centre(b, foes)
	var out []Move
	for _, i := range mine {
		u := b.Units[i]
		dx, dy := u.X-cx, u.Y-cy
		if dx == 0 && dy == 0 {
			dx = 1
		}
		out = append(out, Move{Unit: i, X: u.X + dx, Y: u.Y + dy})
	}
	return out
}

func split(b Battlefield, faction int) (mine, foes []int) {
	for i, u := range b.Units {
		if u.Faction == faction {
			mine = append(mine, i)
		} else {
			foes = append(foes, i)
		}
	}
	return mine, foes
}

func centre(b Battlefield, idx []int) (int, int) {
	if len(idx) == 0 {
		return 0, 0
	}
	sx, sy := 0, 0
	for _, i := range idx {
		sx += b.Units[i].X
		sy += b.Units[i].Y
	}
	return sx / len(idx), sy / len(idx)
}

func nearestFoe(b Battlefield, mine, foes []int) int {
	cx, cy := centre(b, mine)
	best, bestD := foes[0], 1<<62
	for _, i := range foes {
		u := b.Units[i]
		d := (u.X-cx)*(u.X-cx) + (u.Y-cy)*(u.Y-cy)
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

func weakestFoe(b Battlefield, foes []int) int {
	best := foes[0]
	for _, i := range foes {
		if b.Units[i].Strength < b.Units[best].Strength {
			best = i
		}
	}
	return best
}
