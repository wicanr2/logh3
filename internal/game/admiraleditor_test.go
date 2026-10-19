package game

import (
	"bytes"
	"encoding/binary"
	"github.com/wicanr2/logh3/internal/rules"
	"os"
	"path/filepath"
	"testing"
)

func TestArmyBonusPerAdmiralAndLiveProvider(t *testing.T) {
	g := &Game{aiMode: 2, fleets: []rules.Fleet{{Index: 0, Faction: 0, Admiral: 0}, {Index: 1, Faction: 0, Admiral: 1}, {Index: 2, Faction: 1, Admiral: 52}}, battleUnits: []rules.BattleUnit{{Fleet: 0, Faction: 0, Squadron: 0, Ships: 100}, {Fleet: 1, Faction: 0, Squadron: 0, Ships: 100}, {Fleet: 2, Faction: 1, Squadron: 0, Ships: 100}}, admiralPreferences: rules.NewAdmiralPreferences()}
	if g.armyBonus(0, false) != 30 || g.armyBonus(1, false) != 30 || g.armyBonus(0, true) != 50 {
		t.Fatal("兩位名將預設未生效")
	}
	s := rules.AdmiralSpecial{Transfer: 60, Movement: 75, Attack: 40}
	g.admiralPreferences.Admirals[1] = rules.AdmiralOverride{Special: &s}
	if g.armyBonus(0, false) != 40 || g.armyBonus(0, true) != 75 {
		t.Fatal("一般提督不能提供加成或發生累加")
	}
	g.battleUnits[1].Out = true
	if g.armyBonus(0, false) != 30 {
		t.Fatal("離場仍提供")
	}
	g.battleUnits[0].Squadron = 1
	if g.armyBonus(0, false) != 0 {
		t.Fatal("副官或普通分隊提供")
	}
	g.battleUnits[0].Squadron = 0
	g.heroMovementDisabled = true
	if g.armyBonus(0, true) != 0 || g.armyBonus(0, false) != 30 {
		t.Fatal("移動開關影響攻擊")
	}
	g.aiMode = 0
	if g.armyBonus(0, false) != 0 || g.armyBonus(1, true) != 0 {
		t.Fatal("原版模式被加成")
	}
}

func TestFighterBonusCommanderAdjutantAndScope(t *testing.T) {
	g := &Game{aiMode: 2, fleets: []rules.Fleet{{Index: 0, Faction: 0, Admiral: 10, Adjutant: 255}, {Index: 1, Faction: 1, Admiral: 52, Adjutant: 74}, {Index: 2, Faction: 0, Admiral: 1, Adjutant: 255}}, admiralPreferences: rules.NewAdmiralPreferences()}
	if g.fleetFighterBonus(0) != 30 || g.fleetFighterBonus(1) != 30 || g.fleetFighterBonus(2) != 0 || g.fleetFighterBonus(999) != 0 {
		t.Fatal("提督／參謀未觸發，或加成外溢其他艦隊")
	}
	s := rules.DefaultAdmiralSpecial(1)
	s.FighterAttack = 75
	g.admiralPreferences.Admirals[1] = rules.AdmiralOverride{Special: &s}
	g.fleets[0].Adjutant = 1
	if g.fleetFighterBonus(0) != 75 {
		t.Fatal("同艦隊未取較高值或一般參謀不能提供加成")
	}
	g.fleets[0].Admiral = 1
	if g.fleetFighterBonus(0) != 75 {
		t.Fatal("同一人重複累加")
	}
	g.aiMode = 0
	if g.fleetFighterBonus(0) != 0 || g.fleetFighterBonus(1) != 0 {
		t.Fatal("原版模式套用戰鬥機加成")
	}
}

func TestFighterBonusReachesFleetDamageCommand(t *testing.T) {
	g := battleTestGame(t)
	if g == nil {
		t.Skip("缺正常戰場資料")
	}
	g.aiMode = 2
	for _, provider := range []int{10, 74} {
		for _, adjutant := range []bool{false, true} {
			own, target := -1, -1
			for i, u := range g.battleUnits {
				if u.Ships <= 0 || u.Out {
					continue
				}
				if u.Faction == 0 && own < 0 {
					own = i
				}
				if u.Faction == 1 && target < 0 {
					target = i
				}
			}
			if own < 0 || target < 0 {
				t.Fatal("缺交戰單位")
			}
			f := g.fleetByIndex(g.battleUnits[own].Fleet)
			f.Admiral, f.Adjutant = provider, 255
			if adjutant {
				f.Admiral, f.Adjutant = 1, provider
			}
			for weapon := byte(0); weapon < 3; weapon++ {
				c, err := g.fleetDamageCommand(own, target, weapon)
				if err != nil || c.FighterAttackBonus != 30 {
					t.Fatal("正式傷害命令遺漏加成", provider, adjutant, weapon, c, err)
				}
			}
		}
	}
}

func TestGlobalAdmiralOverridesNewGameAndLoad(t *testing.T) {
	g := newTestGame(t)
	before := bytes.Clone(g.res.Admiral)
	p := g.AdmiralPreferences()
	p.Admirals[0] = rules.AdmiralOverride{Abilities: map[int]int{0: 160, 12: 150}}
	if err := g.SetAdmiralPreferences(p); err != nil {
		t.Fatal(err)
	}
	for i, n := range g.res.Admiral {
		want := before[i]
		if i == 11 {
			want = 160
		}
		if i == 22 {
			want = 150
		}
		if n != want {
			t.Fatalf("改動非指定欄位 %d", i)
		}
	}
	g.startNewGame(1)
	if g.res.Admiral[11] != 160 || g.res.Admiral[22] != 150 {
		t.Fatal("新局遺失全局覆寫")
	}
	g.LoadSlot(&rules.Slot{Admiral: bytes.Clone(before), Stellar: bytes.Clone(g.res.Stellar), Planets: g.res.Planets, Fleet: bytes.Repeat([]byte{255}, rules.FleetCount*rules.FleetRecord), Ship: bytes.Repeat([]byte{255}, rules.ShipCount*rules.ShipRecord), Work: bytes.Clone(g.res.WorkTemplate)})
	if g.res.Admiral[11] != 160 || g.res.Admiral[22] != 150 {
		t.Fatal("讀檔遺失全局覆寫")
	}
}

func TestEditedFlagshipChanceCanOverrideMuller(t *testing.T) {
	for _, chance := range []int{0, 100} {
		g, old, _ := flagshipGame(t, 5)
		p := rules.NewAdmiralPreferences()
		s := rules.DefaultAdmiralSpecial(5)
		s.Transfer = chance
		p.Admirals[5] = rules.AdmiralOverride{Special: &s}
		g.admiralPreferences = p
		seed := g.battleRand.Seed()
		if got := g.tryFlagshipTransfer(old); got != (chance == 100) || g.battleRand.Seed() != seed {
			t.Fatal("逐人換艦率或確定性端點錯誤", chance, got)
		}
	}
}

func TestEditorRefreshesFleetAndSavesWithoutMovingShips(t *testing.T) {
	g := newTestGame(t)
	if len(g.fleets) == 0 {
		t.Fatal("S3沒有艦隊")
	}
	f := g.fleets[0]
	before := bytes.Clone(g.slot.Ship)
	p := g.AdmiralPreferences()
	p.Admirals[f.Admiral] = rules.AdmiralOverride{Abilities: map[int]int{3: 180, 4: 170}}
	if err := g.SetAdmiralPreferences(p); err != nil {
		t.Fatal(err)
	}
	r := g.slot.Fleet[f.Index*rules.FleetRecord:]
	ab := rules.FleetAbilities(g.admiralRec(f.Admiral), g.admiralRec(f.Adjutant))
	if !bytes.Equal(r[0x3d:0x48], ab[:]) || g.fleetByIndex(f.Index).Operations != int(ab[10]) {
		t.Fatal("原表與typed艦隊能力未同步")
	}
	groups := g.slot.FleetComposition(f.Index)
	if int(binary.LittleEndian.Uint16(r[0xa:])) != rules.FleetUpkeep(g.res.ShipKinds, f.Faction, groups, int(ab[10])) {
		t.Fatal("營運未更新維持費")
	}
	for pool := 0; pool < rules.ShipCount; pool++ {
		for off := 0; off < rules.ShipRecord; off++ {
			if off == 6 || off == 0x1f || off == 0x21 || off == 0x22 {
				continue
			}
			i := pool*rules.ShipRecord + off
			if before[i] != g.slot.Ship[i] {
				t.Fatalf("能力編輯改動艦艇非派生欄位 pool=%d off=%x", pool, off)
			}
		}
	}
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(filepath.Join(dir, "ADMIRAL.DT0"))
	if err != nil {
		t.Fatal(err)
	}
	if saved[f.Admiral*rules.AdmiralRecord+14] != 180 || saved[f.Admiral*rules.AdmiralRecord+15] != 170 {
		t.Fatal("能力未進公開六表保存")
	}
}
