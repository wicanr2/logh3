package ai

import (
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func heroBattle() Battlefield {
	b := Battlefield{Commanders: [2]Commander{NoCommander, Yang}}
	for fleet := 0; fleet < 5; fleet++ {
		for squad := 0; squad < 3; squad++ {
			b.Units = append(b.Units, rules.BattleUnit{Fleet: fleet, Squadron: squad, Faction: 1, Ships: 100, Strength: 1000, X: 300, Y: 800 + fleet*30})
		}
	}
	b.Units = append(b.Units, rules.BattleUnit{Fleet: 10, Squadron: 0, Faction: 0, Ships: 100, Strength: 1000, X: 1100, Y: 860})
	return b
}

func TestHeroFleetRolesAndDeterminism(t *testing.T) {
	b := heroBattle()
	for _, hero := range []Commander{Yang, Reinhard} {
		b.Commanders[1] = hero
		a, c := (&Encircle{Difficulty: 1}).Moves(b, 1, 1), (&Encircle{Difficulty: 1}).Moves(b, 1, 1)
		if !reflect.DeepEqual(a, c) || len(a) != 5 {
			t.Fatalf("%v 缺少整隊命令或不決定：%v", hero, a)
		}
		left, right := false, false
		for _, m := range a {
			if !m.Formation || b.Units[m.Unit].Squadron != 0 {
				t.Fatal("把編隊拆散")
			}
			left = left || m.Y < 860
			right = right || m.Y > 860
		}
		if !left || !right {
			t.Fatal("缺少兩翼", a)
		}
	}
	b.Commanders[1] = Yang
	y := (&Encircle{}).Moves(b, 1, 1)
	b.Commanders[1] = Reinhard
	r := (&Encircle{}).Moves(b, 1, 1)
	if reflect.DeepEqual(y, r) {
		t.Fatal("兩位名將沒有不同戰術")
	}
}

func TestHeroIgnoresDeadAndScoutsAndClamps(t *testing.T) {
	b := heroBattle()
	b.Units[15].Out = true
	b.Units = append(b.Units, rules.BattleUnit{Fleet: 11, Squadron: -1, Faction: 0, Ships: 1, X: 50, Y: 50})
	if m := (&Encircle{}).Moves(b, 1, 1); len(m) != 0 {
		t.Fatal("追逐已離場或索敵隊", m)
	}
	b.Units[15].Out = false
	b.Units[15].X, b.Units[15].Y = 20, 20
	for _, m := range (&Encircle{}).Moves(b, 1, 1) {
		if m.X < 16 || m.X > 2032 || m.Y < 16 || m.Y > 2024 {
			t.Fatal("目標越界", m)
		}
	}
}

func TestHeroWithoutCommanderPreservesGeneric(t *testing.T) {
	b := heroBattle()
	b.Commanders = [2]Commander{}
	for _, m := range (&Encircle{}).Moves(b, 1, 1) {
		if m.Formation {
			t.Fatal("無名將也套名將命令")
		}
	}
}

func TestHeroWaypointsRespectTerrainAndRandomness(t *testing.T) {
	b := heroBattle()
	b.Rand = rules.NewRand(0x12345678)
	before := b.Rand.Seed()
	b.Terrain = &rules.BattleTerrain{BlockMask: []uint8{0, 15}, QuadrantBits: []uint8{1, 2, 4, 8}, HalfCellW: 8, HalfCellH: 10, Footprint: 16, Step: 8}
	b.Cells = make([]uint8, 128*102)
	for row := 0; row < 102; row++ {
		b.Cells[row*128+60] = 1
	}
	for _, m := range (&Encircle{}).Moves(b, 1, 1) {
		u := b.Units[m.Unit]
		if !b.Terrain.Passable(b.Cells, m.X, m.Y) || !b.Terrain.LineOpen(b.Cells, u.X, u.Y, m.X, m.Y) {
			t.Fatal("命令穿越地形", m)
		}
	}
	if b.Rand.Seed() != before {
		t.Fatal("名將命令消耗亂數")
	}
}
