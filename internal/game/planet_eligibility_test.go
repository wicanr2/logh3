package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 原版第六步typed資料注入只作Game消費與061診斷，不稱正常玩家入口。
func planetCounterOnlyGame(t *testing.T) (*Game, []rules.BattleUnit) {
	t.Helper()
	var receipt struct {
		Pipeline struct {
			Input, Expected struct {
				Planets []rules.BattlePlanet
				Fleets  []rules.BattleFleet
				Units   []rules.BattleUnit
				Seed    uint32
			}
		}
	}
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/planet-eligibility.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	g := newTestGame(t)
	g.battlePlanets, g.battleFleets, g.battleUnits = receipt.Pipeline.Input.Planets, receipt.Pipeline.Input.Fleets, receipt.Pipeline.Input.Units
	g.battleRand = rules.NewRand(receipt.Pipeline.Input.Seed)
	g.battleRegion, g.battleTurn, g.state = 26, 5, StateBattle
	return g, receipt.Pipeline.Expected.Units
}

func TestPlanetCounterOnlyPassivePolicy(t *testing.T) {
	for _, faction := range []int{0, 1} {
		for _, original := range []bool{false, true} {
			t.Run(fmt.Sprintf("faction%d-original%t", faction, original), func(t *testing.T) {
				g, expected := planetCounterOnlyGame(t)
				g.faction, g.originalBattleAI = faction, original
				before := append([]rules.BattleUnit(nil), g.battleUnits...)
				actions := append([]rules.BattleFleet(nil), g.battleFleets...)
				hits := g.planetAssault()
				if len(hits) != 1 || hits[0].Counter != 253 || hits[0].Damage != 0 || hits[0].Captured || hits[0].Score != 0 {
					t.Fatalf("結果%+v", hits)
				}
				if !original && faction == 1 {
					expected = before
				}
				if !reflect.DeepEqual(g.battleUnits, expected) {
					t.Fatal("原版正對照或被動CPU零傷害不符")
				}
				for i, f := range g.battleFleets {
					if f.Action != actions[i].Action {
						t.Fatal("CPU命令未恢復")
					}
				}
				if g.battleRand.Seed() != 0x0b719151 {
					t.Fatal("被動CPU改變亂數結算")
				}
			})
		}
	}
}

// 普通／快速／最快及關動畫共用規則結算，防衛Hit須保留既有演出消費。
func TestPlanetCounterOnlyAnimationModes(t *testing.T) {
	for _, anim := range []int{0, 1} {
		for _, speed := range []int{0, 1, 2} {
			t.Run(fmt.Sprintf("anim%d-speed%d", anim, speed), func(t *testing.T) {
				g, expected := planetCounterOnlyGame(t)
				g.originalBattleAI = true
				g.display.Anim, g.display.AnimSpeed = anim, speed
				loadEngagement(t, g.res)
				hits := g.planetAssault()
				if !g.queueEngagements(hits) || len(g.engQueue) != 1 || g.engQueue[0].label != screen.EngagementDefence || g.engQueue[0].destroyed != 253 {
					t.Fatal("只受反擊未接防衛演出")
				}
				if !reflect.DeepEqual(g.battleUnits, expected) || g.battleRand.Seed() != 0x0b719151 {
					t.Fatal("動畫設定改變原池或seed")
				}
				v := g.engagementView()
				if v.Gauges == nil || (anim == 0 && !v.ShowResult) {
					t.Fatal("反擊量表／關動畫結果缺失")
				}
				if fb := g.drawEngagement(); len(fb.Pix) != 640*480 {
					t.Fatal("正式renderer畫布錯誤")
				}
			})
		}
	}
}
