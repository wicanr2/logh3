package game

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func flagshipGame(t *testing.T, admiral int) (*Game, int, int) {
	t.Helper()
	g := battleTestGame(t)
	if g == nil {
		t.Skip("缺正常戰場入口資料")
	}
	g.SetPassiveBattleComputer(false)
	g.SetRuleOptions(DefaultRuleOptions())
	for i, u := range g.battleUnits {
		if u.Squadron != 0 || u.Faction != 0 {
			continue
		}
		before := u
		g.battleUnits[i].Ships = 0
		target := rules.FlagshipTransferTarget(g.battleUnits, i)
		if target < 0 {
			g.battleUnits[i] = before
			continue
		}
		f := g.fleetByIndex(u.Fleet)
		f.Admiral = admiral
		g.slot.Fleet[u.Fleet*rules.FleetRecord+4] = byte(admiral)
		return g, i, target
	}
	t.Fatal("S3缺可轉移的帝國艦隊")
	return nil, 0, 0
}

func TestFlagshipTransferSaveCleanupAndSecondLoss(t *testing.T) {
	g, old, target := flagshipGame(t, 5)
	flag, next := g.battleUnits[old], g.battleUnits[target]
	fleet, oldPool, newPool := flag.Fleet, g.battleUnitPool(flag), g.battleUnitPool(next)
	adm, adj := g.fleetByIndex(fleet).Admiral, g.fleetByIndex(fleet).Adjutant
	admirals := bytes.Clone(g.res.Admiral)
	seed := g.battleRand.Seed()
	if !g.tryFlagshipTransfer(old) || g.saveErr != nil {
		t.Fatal("謬拉未成功轉移", g.saveErr)
	}
	if g.battleRand.Seed() != seed || g.battleFlagshipGone(fleet) || g.battleUnitPool(g.battleUnits[target]) != newPool || g.battleUnitPool(g.battleUnits[old]) != oldPool {
		t.Fatal("100%判定或單位／池映射錯誤")
	}
	newState := g.battleUnits[target]
	newState.Squadron, newState.Slot = next.Squadron, next.Slot
	if newState != next || !bytes.Equal(g.res.Admiral, admirals) {
		t.Fatal("轉移改了艦艇資源／位置或提督狀態")
	}
	if g.tryFlagshipTransfer(old) || g.battleRand.Seed() != seed {
		t.Fatal("同一次擊毀重擲")
	}
	if err := g.cleanupBattleFleet(fleet); err != nil {
		t.Fatal(err)
	}
	if g.slot.Ship[oldPool*rules.ShipRecord+1] != 255 || g.slot.FleetGroupShips(fleet)[0] != newPool || g.fleetByIndex(fleet).Admiral != adm || g.fleetByIndex(fleet).Adjutant != adj {
		t.Fatal("末端清理復活舊旗艦或遺失提督")
	}
	dir := t.TempDir()
	g.faction = 0
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal("公開保存", err)
	}
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(dir, name+".DT0"))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	files := rules.SlotFiles{Stellar: read("STELLAR"), Planet: read("PLANET"), Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Work: read("WORK")}
	header, err := os.ReadFile(filepath.Join(root, "org_game/extracted/PLADAT.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	files.PlanetHeader = header
	sl, err := rules.LoadSlot(files)
	if err != nil || sl.FleetGroupShips(fleet)[0] != newPool || sl.Ship[oldPool*rules.ShipRecord+1] != 255 {
		t.Fatal("六表讀回未保存轉移", err)
	}
	// A new destruction may transfer again; the dead original remains dead.
	for i, u := range g.battleUnits {
		if u.Fleet == fleet && u.Squadron == 0 {
			g.battleUnits[i].Ships = 0
			candidate := rules.FlagshipTransferTarget(g.battleUnits, i)
			if candidate < 0 {
				t.Fatal("fixture缺第二艘戰艦分隊")
			}
			if !g.tryFlagshipTransfer(i) || g.battleFlagshipGone(fleet) {
				t.Fatal("新旗艦再被擊毀不能轉移", g.saveErr)
			}
			break
		}
	}
}

func TestFlagshipTransferOrdinaryOneRollAndDisabled(t *testing.T) {
	for _, tc := range []struct {
		seed    uint32
		success bool
	}{{1, false}, {2, true}} {
		g, old, _ := flagshipGame(t, 0)
		g.battleRand = rules.NewRand(tc.seed)
		g.fleetByIndex(g.battleUnits[old].Fleet).Adjutant = 5
		expected := rules.NewRand(tc.seed)
		roll := expected.Next()
		want := roll%100 < 25
		if want != tc.success {
			t.Fatal("固定seed不符預先指定的機率分支")
		}
		before := bytes.Clone(g.slot.Fleet)
		if got := g.tryFlagshipTransfer(old); got != want || g.battleRand.Seed() != expected.Seed() {
			t.Fatalf("seed=%d roll=%d got=%v want=%v", tc.seed, roll, got, want)
		}
		if !want && !bytes.Equal(before, g.slot.Fleet) {
			t.Fatal("失敗轉移改了編制")
		}
		if g.tryFlagshipTransfer(old) {
			t.Fatal("同一次失敗重擲成功")
		}
		if g.battleRand.Seed() != expected.Seed() {
			t.Fatal("同一次失敗重擲取了亂數")
		}
	}
	g, old, _ := flagshipGame(t, 5)
	g.SetRuleOptions(RuleOptions{FlagshipTransfer: false, HeroMovement: true})
	seed := g.battleRand.Seed()
	before := append([]rules.BattleUnit(nil), g.battleUnits...)
	if g.tryFlagshipTransfer(old) || seed != g.battleRand.Seed() || !reflect.DeepEqual(before, g.battleUnits) {
		t.Fatal("關閉仍轉移")
	}
	if !g.battleFlagshipGone(g.battleUnits[old].Fleet) {
		t.Fatal("關閉未保留原全滅判準")
	}
	// No eligible ships must not roll, including when Muller is the adjutant.
	g.SetRuleOptions(DefaultRuleOptions())
	g.fleetByIndex(g.battleUnits[old].Fleet).Admiral = 0
	g.fleetByIndex(g.battleUnits[old].Fleet).Adjutant = 5
	for i := range g.battleUnits {
		if i != old {
			g.battleUnits[i].Out = true
		}
	}
	if g.tryFlagshipTransfer(old) || seed != g.battleRand.Seed() {
		t.Fatal("無候選仍取亂數")
	}
}

func TestFlagshipTransferDamagePaths(t *testing.T) {
	t.Run("fleet", func(t *testing.T) {
		g, old, target := flagshipGame(t, 5)
		g.battleUnits[old].Ships = 1
		own := -1
		for i, u := range g.battleUnits {
			if u.Faction == 1 && u.Kind == 0 && u.Ships > 0 {
				own = i
				break
			}
		}
		if own < 0 {
			t.Fatal("S3缺同盟戰艦")
		}
		g.battleUnits[own].X, g.battleUnits[own].Y = g.battleUnits[old].X+1, g.battleUnits[old].Y
		g.display.Anim = 0
		g.battleRand = rules.NewRand(1)
		s := &fleetCombatSceneState{own: own, target: old, weapon: 0, before: append([]rules.BattleUnit(nil), g.battleUnits...), command: rules.FleetCombatCommand{Weapon: 0, AttackAbilities: [3]int{255, 255, 255}, DefendAbilities: [3]int{1, 1, 1}, TargetWeapon: 0}}
		if err := g.applyFleetSceneDamage(s); err != nil {
			t.Fatal(err)
		}
		if s.hit.Lost != 1 || g.battleUnits[target].Squadron != 0 || g.saveErr != nil {
			t.Fatal("艦隊傷害未轉移", s.hit, g.saveErr)
		}
		if s.after[old].Squadron != 0 || s.after[old].Ships != 0 {
			t.Fatal("傷害演出快照被轉移覆蓋")
		}
	})
	t.Run("fortress", func(t *testing.T) {
		g, old, target := flagshipGame(t, 5)
		for i := range g.battleUnits {
			g.battleUnits[i].X, g.battleUnits[i].Y = 1800, 1800
		}
		g.battleUnits[old].Ships = 1
		g.battleUnits[old].X, g.battleUnits[old].Y = 500, 500
		g.battlePlanets = []rules.BattlePlanet{{Owner: 1, Fortress: true, GunState: rules.FortressGunAimed, GunAim: 0, X: 500, Y: 500, DefenceMax: 12800, Defence: 12800}}
		g.battleRand = rules.NewRand(1)
		hits := g.fortressGunFire()
		if len(hits) != 1 || hits[0].Lost != 1 || g.battleUnits[target].Squadron != 0 || g.saveErr != nil {
			t.Fatal("要塞傷害未轉移", hits, g.saveErr)
		}
		if g.battleGunScenes[1].After.Units[old].Squadron != 0 || g.battleGunScenes[1].After.Units[old].Ships != 0 {
			t.Fatal("要塞快照錯位")
		}
	})
	t.Run("planet", func(t *testing.T) {
		g, old, _ := flagshipGame(t, 5)
		fleet := g.battleUnits[old].Fleet
		for i := range g.battleUnits {
			if g.battleUnits[i].Fleet == fleet {
				g.battleUnits[i].Ships = 1
			}
		}
		g.battleUnits[old].Ships = 200
		bf := *g.battleFleet(fleet)
		bf.X, bf.Y, bf.Action, bf.Defence, bf.Integrity, bf.Flag0 = 500, 500, 1, 1, 50, false
		g.battleFleets = []rules.BattleFleet{bf}
		g.battlePlanets = []rules.BattlePlanet{{Owner: 1, X: 500, Y: 500, DefenceMax: 12800, Defence: 12800}}
		g.battleRand = rules.NewRand(1)
		g.planetAssault()
		if g.battleUnits[old].Ships != 0 || g.battleFlagshipGone(fleet) || g.saveErr != nil {
			t.Fatal("行星反擊未轉移", g.battleUnits[old], g.saveErr)
		}
		if len(g.battleAssaultScenes) == 0 || g.battleAssaultScenes[0].Kind != rules.PlanetCounterScene || g.battleAssaultScenes[0].After.Units[old].Squadron != 0 {
			t.Fatal("反擊前後快照缺失")
		}
	})
}

func TestHeroMovementSwitchKeepsTacticsAndBaseSpeed(t *testing.T) {
	g, flag, _ := flagshipGame(t, 0)
	g.battleUnits[flag].Ships = 100
	g.aiMode = 2
	u := g.battleUnits[flag]
	g.applyTacticalMove(flag, u.X+80, u.Y)
	g.fleetOrderUnit, g.fleetOrderBattle = flag, true
	bf := g.battleFleet(u.Fleet)
	base := bf.PerTurn
	if base <= 0 {
		t.Fatal("移動測試沒有基本額度")
	}
	for _, enabled := range []bool{true, false, true} {
		g.SetRuleOptions(RuleOptions{FlagshipTransfer: true, HeroMovement: enabled})
		if g.battleCommanders()[0] == 0 {
			t.Fatal("關閉移動也關了名將戰術")
		}
		for i := range g.battleUnits {
			g.battleUnits[i].Budget = 0
		}
		g.newBattleMovement()
		want := base
		if enabled {
			want = rules.CommandMovementBudget(base)
		}
		if bf.Budget != want || bf.PerTurn != base {
			t.Fatal("實際新額度錯誤或累乘", bf.Budget, want, bf.PerTurn)
		}
		if r := g.orderedFleetForRange(); r == nil || r.dist != want/rules.FixOne {
			t.Fatal("範圍與額度不同", r)
		}
		if g.battleUnits[flag].CruiseSpeed != u.CruiseSpeed {
			t.Fatal("開關改動基本速度")
		}
	}
	g.aiMode = 0
	if g.heroMovementActive(0) {
		t.Fatal("原版模式啟用名將加成")
	}
}
