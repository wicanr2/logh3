package rules

import (
	"bytes"
	"testing"
)

func TestAdmiralPreferencesSparseAndBoundaries(t *testing.T) {
	p := NewAdmiralPreferences()
	p.Admirals[0] = AdmiralOverride{Abilities: map[int]int{0: 150, 12: 140}}
	data := bytes.Repeat([]byte{80}, 82*AdmiralRecord)
	before := bytes.Clone(data)
	p.Apply(data)
	for i, n := range data {
		want := before[i]
		if i == 11 {
			want = 150
		}
		if i == 22 {
			want = 140
		}
		if n != want {
			t.Fatalf("非指定欄位被改：%d", i)
		}
	}
	q := p.Clone()
	q.Admirals[0].Abilities[0] = 120
	if p.Admirals[0].Abilities[0] != 150 {
		t.Fatal("草稿與原設定共用map")
	}
	for _, n := range []int{0, 201} {
		q := p.Clone()
		q.Admirals[0].Abilities[0] = n
		if q.Validate() == nil {
			t.Fatal("接受非法能力", n)
		}
	}
	if DefaultAdmiralSpecial(0).Attack != 30 || DefaultAdmiralSpecial(52).Attack != 30 || DefaultAdmiralSpecial(5).Transfer != 100 || DefaultAdmiralSpecial(1).Attack != 0 {
		t.Fatal("預設值不符裁決")
	}
	if BonusAmount(101, 30) != 131 || BonusAmount(100, 75) != 175 || BonusAmount(0x7fffffff, 200) != 0x7fffffff {
		t.Fatal("百分比取整或飽和錯誤")
	}
}

func TestFleetAttackBonusUsesDamageAndInventoryPath(t *testing.T) {
	for weapon := 0; weapon < 3; weapon++ {
		own := BattleUnit{Faction: 0, Kind: 0, Ships: 1000, FullShips: 1000, X: 100, Y: 100, Missiles: 99999, Fighters: 99999, Squadron: 1}
		for kind := 0; kind < 8; kind++ {
			k, _ := fcKinds(t).Kind(0, kind)
			if k.Byte(10+weapon*4) < 12 && k.Byte(11+weapon*4) > 0 {
				own.Kind = kind
				break
			}
		}
		target := BattleUnit{Faction: 1, Kind: 0, Ships: 10000, FullShips: 10000, X: 130, Y: 100, Squadron: 1}
		c := FleetCombatCommand{Weapon: weapon, AttackAbilities: [3]int{150, 150, 150}, DefendAbilities: [3]int{100, 100, 100}, TargetWeapon: 255, TargetAttribute: 100}
		a, b := own, target
		r := NewRand(123)
		base, err := ApplyFleetCombatDamage(&a, &b, fcKinds(t), c, r)
		if err != nil {
			t.Fatal(err)
		}
		x, y := own, target
		s := NewRand(123)
		c.AttackBonus = 30
		boost, err := ApplyFleetCombatDamage(&x, &y, fcKinds(t), c, s)
		if err != nil {
			t.Fatal(err)
		}
		if boost.Attack != base.Attack*130/100 || boost.Lost < base.Lost || x.Missiles != a.Missiles || x.Fighters != a.Fighters || s.Seed() != r.Seed() {
			t.Fatal("攻擊加成改錯係數、庫存或亂數", weapon, base, boost)
		}
		if boost.Lost == base.Lost {
			t.Fatal("固定局面未呈現實際戰損增加", weapon, base, boost)
		}
	}
}

func TestGlobalWeaponOverrideRejectsZeroDefenceCoefficient(t *testing.T) {
	for _, field := range []int{4, 5, 6} {
		for _, n := range []int{1, 2, 3, 200} {
			p := NewAdmiralPreferences()
			p.Admirals[0] = AdmiralOverride{Abilities: map[int]int{field: n}}
			err := p.ValidateCombat(fcKinds(t))
			if (err != nil) != (n < 3) {
				t.Fatal("實際16艦種／256角度下限錯誤", field, n, err)
			}
		}
	}
}

func TestFighterBonusDamageWeaponsAndBoundaries(t *testing.T) {
	for _, i := range []int{10, 74, 0, 52, 9, 73} {
		want := 0
		if i == 10 || i == 74 {
			want = 30
		}
		if DefaultAdmiralSpecial(i).FighterAttack != want {
			t.Fatal("戰鬥機預設錯誤", i)
		}
	}
	for _, bonus := range []int{-1, 0, 30, 200, 201} {
		p := NewAdmiralPreferences()
		s := DefaultAdmiralSpecial(74)
		s.FighterAttack = bonus
		p.Admirals[74] = AdmiralOverride{Special: &s}
		if (p.Validate() != nil) != (bonus < 0 || bonus > 200) {
			t.Fatal("戰鬥機範圍", bonus)
		}
	}
	for weapon := 0; weapon < 3; weapon++ {
		own := BattleUnit{Faction: 0, Ships: 1000, FullShips: 1000, X: 100, Y: 100, Missiles: 99999, Fighters: 99999, Squadron: 1}
		for kind := 0; kind < 8; kind++ {
			k, _ := fcKinds(t).Kind(0, kind)
			if k.Byte(10+weapon*4) < 12 && k.Byte(11+weapon*4) > 0 {
				own.Kind = kind
				break
			}
		}
		target := BattleUnit{Faction: 1, Ships: 10000, FullShips: 10000, X: 130, Y: 100, Squadron: 1}
		c := FleetCombatCommand{Weapon: weapon, AttackAbilities: [3]int{150, 150, 150}, DefendAbilities: [3]int{100, 100, 100}, TargetWeapon: 255, TargetAttribute: 100, AttackBonus: 30}
		a, b := own, target
		r := NewRand(123)
		base, err := ApplyFleetCombatDamage(&a, &b, fcKinds(t), c, r)
		if err != nil {
			t.Fatal(err)
		}
		x, y := own, target
		r2 := NewRand(123)
		c.FighterAttackBonus = 30
		boost, err := ApplyFleetCombatDamage(&x, &y, fcKinds(t), c, r2)
		if err != nil {
			t.Fatal(err)
		}
		want := base.Attack
		if weapon == 2 {
			want = base.Attack * 130 / 100
		}
		if boost.Attack != want || x.Missiles != a.Missiles || x.Fighters != a.Fighters || r2.Seed() != r.Seed() {
			t.Fatal("武器隔離／庫存／RNG", weapon, base, boost)
		}
		if weapon == 2 && boost.Lost <= base.Lost {
			t.Fatal("戰鬥機實際戰損未增加", base, boost)
		}
		if weapon != 2 && boost != base {
			t.Fatal("其他武器被加成", weapon)
		}
	}
}
