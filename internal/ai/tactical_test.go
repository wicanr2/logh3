package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// battle 造一場：同盟六支在右下、帝國一支在左上——就是收據 `s3ai6`
// 量到的那個開場（`docs/re/80`）。
func battle() Battlefield {
	return Battlefield{Units: []rules.BattleUnit{
		{Fleet: 12, Faction: 0, X: 416, Y: 280, TargetX: 416, TargetY: 280, Ships: 100, Strength: 30000},
		{Fleet: 16, Faction: 1, X: 1536, Y: 1840, TargetX: 1536, TargetY: 1840, Ships: 100, Strength: 20000},
		{Fleet: 17, Faction: 1, X: 1536, Y: 1808, TargetX: 1536, TargetY: 1808, Ships: 100, Strength: 20000},
		{Fleet: 28, Faction: 1, X: 1568, Y: 1808, TargetX: 1568, TargetY: 1808, Ships: 100, Strength: 20000},
		{Fleet: 29, Faction: 1, X: 1504, Y: 1872, TargetX: 1504, TargetY: 1872, Ships: 100, Strength: 20000},
		{Fleet: 30, Faction: 1, X: 1504, Y: 1808, TargetX: 1504, TargetY: 1808, Ships: 100, Strength: 20000},
		{Fleet: 31, Faction: 1, X: 1536, Y: 1776, TargetX: 1536, TargetY: 1776, Ships: 100, Strength: 20000},
	}}
}

// turn 套用命令再讓所有單位走一步。
func turn(b *Battlefield, t Tactic, faction, n int) []Move {
	ms := t.Moves(*b, faction, n)
	for _, m := range ms {
		b.Units[m.Unit].TargetX, b.Units[m.Unit].TargetY = m.X, m.Y
	}
	for i := range b.Units {
		rules.AdvanceBattle(&b.Units[i], nil, nil)
	}
	return ms
}

// TestOriginalTacticConverges 原版整批指同一個點（`docs/re/80` 量到的行為）。
func TestOriginalTacticConverges(t *testing.T) {
	b := battle()
	ms := turn(&b, OriginalTactic{}, 1, 1)
	if len(ms) != 6 {
		t.Fatalf("下了 %d 道命令，六支都該有", len(ms))
	}
	for _, m := range ms[1:] {
		if m.X != ms[0].X || m.Y != ms[0].Y {
			t.Fatalf("目標點不一致：%v vs %v", m, ms[0])
		}
	}
	// 而且要朝我方壓過來。
	before, after := 1840, b.Units[1].Y
	if after >= before {
		t.Fatalf("第一支的 y 從 %d 變成 %d，沒有朝敵方移動", before, after)
	}
}

// TestEncircleComesFromSeveralSides 強化版要從幾個方位圍上去，
// 不是排成一條線壓過來。
func TestEncircleComesFromSeveralSides(t *testing.T) {
	b := battle()
	e := &Encircle{}
	var ms []Move
	for n := 1; n <= 12; n++ {
		ms = turn(&b, e, 1, n)
	}
	// 六支的目標點要落在目標周圍不同的方位。
	seen := map[[2]int]bool{}
	for _, m := range ms {
		seen[[2]int{sign(m.X - b.Units[0].X), sign(m.Y - b.Units[0].Y)}] = true
	}
	if len(seen) < 3 {
		t.Fatalf("只從 %d 個方位靠過去（%v），沒有包圍", len(seen), seen)
	}
	// 而且真的靠到了。
	near := 0
	for i, u := range b.Units {
		if i == 0 {
			continue
		}
		if dist2(u, b.Units[0]) < 400*400 {
			near++
		}
	}
	if near < 4 {
		t.Fatalf("十二回合只有 %d 支靠到目標旁邊", near)
	}
}

// TestEncircleFocusesTheWeakest 集火挑最好打的那一支。
func TestEncircleFocusesTheWeakest(t *testing.T) {
	b := Battlefield{Units: []rules.BattleUnit{
		{Fleet: 0, Faction: 0, X: 100, Y: 100, Ships: 100, Strength: 50000},
		{Fleet: 1, Faction: 0, X: 900, Y: 900, Ships: 100, Strength: 5000}, // 最弱，雖然比較遠
		{Fleet: 2, Faction: 1, X: 500, Y: 500, Ships: 100, Strength: 20000},
		{Fleet: 3, Faction: 1, X: 520, Y: 500, Ships: 100, Strength: 20000},
	}}
	ms := (&Encircle{}).Moves(b, 1, 1)
	if len(ms) != 2 {
		t.Fatalf("下了 %d 道命令", len(ms))
	}
	for _, m := range ms {
		if dist(m.X, m.Y, 900, 900) > 400 {
			t.Fatalf("命令指向 (%d,%d)，不是最弱那一支旁邊", m.X, m.Y)
		}
	}
}

// TestEncircleRetreatsWhenBeaten 打不動了就撤，不打到全滅。
func TestEncircleRetreatsWhenBeaten(t *testing.T) {
	b := Battlefield{Units: []rules.BattleUnit{
		{Faction: 0, X: 1000, Y: 1000, Ships: 100, Strength: 50000},
		{Fleet: 2, Faction: 1, X: 500, Y: 500, Ships: 100, Strength: 20000},
	}}
	e := &Encircle{}
	e.Moves(b, 1, 1) // 記下開場戰力
	b.Units[1].Strength = 5000
	ms := e.Moves(b, 1, 2)
	if len(ms) != 1 {
		t.Fatalf("下了 %d 道命令", len(ms))
	}
	if dist(ms[0].X, ms[0].Y, 1000, 1000) <= dist(500, 500, 1000, 1000) {
		t.Fatal("戰力只剩四分之一還往敵方走")
	}
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}

func dist2(a, b rules.BattleUnit) int {
	return (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y)
}

func dist(x1, y1, x2, y2 int) int {
	dx, dy := x1-x2, y1-y2
	d := dx*dx + dy*dy
	r := 0
	for r*r < d {
		r++
	}
	return r
}

// squadrons 把 battle() 那一場的每支艦隊展成分隊——戰場上真正在動的
// 是分隊，不是艦隊（`docs/internals/38`）。朝向照原版量到的長度取 23。
func squadrons(n int) Battlefield {
	var out []rules.BattleUnit
	for _, f := range battle().Units {
		ux, uy := 22, 6
		if f.Faction == 1 {
			ux, uy = -22, -6 // 兩邊面對面
		}
		for k := 0; k < n; k++ {
			dx, dy := rules.Formation(k, ux, uy)
			out = append(out, rules.BattleUnit{
				Fleet: f.Fleet, Squadron: k, Faction: f.Faction,
				X: f.X + dx, Y: f.Y + dy,
				TargetX: f.X + dx, TargetY: f.Y + dy,
				Ships: 100, Strength: f.Strength / n,
			})
		}
	}
	return Battlefield{Units: out}
}

// TestEncircleAtSquadronScale 分隊規模下強化版仍然要圍上去：
// 每艦隊只由旗艦收命令、目標點不只一個，而且縮短距離。
func TestEncircleAtSquadronScale(t *testing.T) {
	b := squadrons(6)
	mine := 0
	for _, u := range b.Units {
		if u.Faction == 1 && u.Squadron == 0 {
			mine++
		}
	}
	e := &Encircle{}
	var ms []Move
	for n := 1; n <= 12; n++ {
		ms = turn(&b, e, 1, n)
	}
	if len(ms) != mine {
		t.Fatalf("下了 %d 道命令，%d 個艦隊各一條", len(ms), mine)
	}
	pts := map[[2]int]bool{}
	for _, m := range ms {
		if !m.Formation || b.Units[m.Unit].Squadron != 0 {
			t.Fatal("AI拆分艦隊", m)
		}
		pts[[2]int{m.X, m.Y}] = true
	}
	if len(pts) < 4 {
		t.Fatalf("只用了 %d 個目標點，沒有從幾個方位圍", len(pts))
	}
	// 帝國那一支的分隊還在原地，同盟該壓到它附近。
	var near int
	for _, u := range b.Units {
		if u.Faction == 1 && abs(u.X-416) < 900 && abs(u.Y-280) < 900 {
			near++
		}
	}
	if near == 0 {
		t.Fatal("十二回合過去沒有任何分隊接近敵方")
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// 圍上去的方位要按方向係數排——第一支分到的位置，打過去的係數
// 不能低於最後一支的（`docs/re/88`）。
func TestEncirclePrefersGoodBearings(t *testing.T) {
	b := squadrons(6)
	e := &Encircle{}
	ms := turn(&b, e, 1, 1)
	if len(ms) < 2 {
		t.Fatal("命令太少，看不出方位順序")
	}
	// 驗排序函式本身：回來的順序，係數必須不遞增。
	order := bearingOrder(1000, 1000, 200)
	prev := 1 << 30
	for _, i := range order {
		bb := bearings[i]
		px, py := 1000+bb[0]*20, 1000+bb[1]*20
		f := rules.DamageFactor(rules.BattleFacing(px, py), rules.Bearing(px, py, 1000, 1000))
		if f > prev {
			t.Fatalf("方位沒有按係數排：%v", order)
		}
		prev = f
	}
	if len(order) != len(bearings) {
		t.Fatalf("排出來 %d 個方位，該是 %d 個", len(order), len(bearings))
	}
}
