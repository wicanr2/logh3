package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// battle 造一場：同盟六支在右下、帝國一支在左上——就是收據 `s3ai6`
// 量到的那個開場（`docs/re/80`）。
func battle() Battlefield {
	return Battlefield{Units: []rules.BattleUnit{
		{Fleet: 12, Faction: 0, X: 416, Y: 280, TargetX: 416, TargetY: 280, Strength: 30000},
		{Fleet: 16, Faction: 1, X: 1536, Y: 1840, TargetX: 1536, TargetY: 1840, Strength: 20000},
		{Fleet: 17, Faction: 1, X: 1536, Y: 1808, TargetX: 1536, TargetY: 1808, Strength: 20000},
		{Fleet: 28, Faction: 1, X: 1568, Y: 1808, TargetX: 1568, TargetY: 1808, Strength: 20000},
		{Fleet: 29, Faction: 1, X: 1504, Y: 1872, TargetX: 1504, TargetY: 1872, Strength: 20000},
		{Fleet: 30, Faction: 1, X: 1504, Y: 1808, TargetX: 1504, TargetY: 1808, Strength: 20000},
		{Fleet: 31, Faction: 1, X: 1536, Y: 1776, TargetX: 1536, TargetY: 1776, Strength: 20000},
	}}
}

// turn 套用命令再讓所有單位走一步。
func turn(b *Battlefield, t Tactic, faction, n int) []Move {
	ms := t.Moves(*b, faction, n)
	for _, m := range ms {
		b.Units[m.Unit].TargetX, b.Units[m.Unit].TargetY = m.X, m.Y
	}
	for i := range b.Units {
		rules.AdvanceBattle(&b.Units[i])
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
		{Faction: 0, X: 100, Y: 100, Strength: 50000},
		{Faction: 0, X: 900, Y: 900, Strength: 5000}, // 最弱，雖然比較遠
		{Faction: 1, X: 500, Y: 500, Strength: 20000},
		{Faction: 1, X: 520, Y: 500, Strength: 20000},
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
		{Faction: 0, X: 1000, Y: 1000, Strength: 50000},
		{Faction: 1, X: 500, Y: 500, Strength: 20000},
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
