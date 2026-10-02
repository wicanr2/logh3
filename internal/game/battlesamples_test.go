package game

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 正常載入隨附存檔後實行；不注入艦隊、損傷、戰場、座標或資源。
// 此測試只證明 remake 玩家路徑，原版同狀態收據另驗。
func TestBattleShippedSavesNormalExecuteAndSave(t *testing.T) {
	for _, c := range []struct {
		slot, region, turn, fleets, units int
	}{{8, 8, 94, 16, 103}, {9, 13, 347, 6, 45}, {10, 26, 381, 4, 31}} {
		t.Run(fmt.Sprintf("S%d", c.slot-7), func(t *testing.T) {
			g := newTestGame(t)
			g.state = StateMainMenu
			g.MoveMouse(300, 210)
			g.Click(300, 210)
			for _, key := range []byte("138") {
				g.Key(key)
			}
			g.Key('\r')
			g.RightClick(300, 240)
			g.RightClick(300, 240)
			g.MoveMouse(600, 340)
			for _, p := range [][2]int{{544, 125}, {560, 145}, {300, 122 + c.slot*20}, {318, 208}} {
				g.Click(p[0], p[1])
			}
			if g.state != StateCommandPhase || g.turn != c.turn-1 || g.SaveErr() != nil {
				t.Fatalf("正常載入失敗：state=%v turn=%d err=%v", g.state, g.turn, g.SaveErr())
			}
			g.strategyRand = rules.NewRand(strategyRandSeed) // 實行前固定，與產品預設相同。
			g.Click(544, 125)
			g.Click(560, 125)
			for i := 0; i < 15000 && g.state != StateBattle; i++ {
				switch g.state {
				case StateTurnThinking, StateTurnWindow, StateTurnReport, StateEngagement:
					g.Tick(time.Second / 60)
				case StateOccupationReport, StateRebellionReport, StateRegionFlipReport, StateAdmiralFateReport, StateShipDelivery, StatePhezzanReport:
					g.Click(560, 125)
				case StateMainGame:
					if g.reportLayer > 0 {
						g.RightClick(600, 240)
					} else {
						t.Fatal("一次實行未進戰場")
					}
				default:
					t.Fatalf("進場停在未預期狀態 %v", g.state)
				}
			}
			if g.state != StateBattle || g.battleRegion != c.region || g.turn != c.turn || len(g.battleFleets) != c.fleets || len(g.battleUnits) != c.units {
				t.Fatalf("進場資料不同：state=%v region=%d turn=%d fleets=%d units=%d", g.state, g.battleRegion, g.turn, len(g.battleFleets), len(g.battleUnits))
			}
			loadPreparationPresentation(t, g)
			loadMovementTiles(t, g)
			label := fmt.Sprintf("s%d", c.slot-7)
			movementFrame(t, g, label+"-entry")
			damaged, humanDamaged, supplyKinds := 0, 0, 0
			for _, u := range g.battleUnits {
				if u.Ships < u.FullShips {
					damaged++
					if g.humanSeat(u.Faction) {
						humanDamaged++
					}
				}
				if u.Kind == rules.SupplyKind {
					supplyKinds++
				}
			}
			g.battleRand = rules.NewRand(0x12345678) // 首次戰術實行前固定，不重擲。
			g.MoveMouse(600, 240)
			g.Click(600, 240)
			if !g.battleMenu {
				t.Fatal("正常戰術選單未開啟")
			}
			l := &g.res.BattleMenu.Layout
			x := g.battleMenuX + l.Highlight.X - l.Window.X + 2
			y := g.battleMenuY + l.Highlight.Y - l.Window.Y + 2
			g.MoveMouse(x, y)
			g.Click(x, y)
			supplyFrames, repairFrames, movingFrames := 0, 0, 0
			for i := 0; i < 12000 && (g.battlePreparing || g.battleMoving || g.state == StateTurnReport || g.state == StateEngagement); i++ {
				if g.battlePreparationSupply != nil {
					supplyFrames++
					if g.battlePhase != screen.BattlePhaseSupply {
						t.Fatal("補給結果未顯示補給階段")
					}
				}
				if g.battlePreparationRepair != nil {
					repairFrames++
					if g.battlePhase != screen.BattlePhaseRepair {
						t.Fatal("修理結果未顯示修理階段")
					}
				}
				if g.battleMoving {
					if g.battlePhase != screen.BattlePhaseMovement {
						t.Fatal("移動未顯示移動階段")
					}
					if movingFrames == 0 {
						movementFrame(t, g, label+"-movement")
					}
					movingFrames++
				}
				g.Frame()
				g.Tick(time.Second / 60)
			}
			if g.battlePreparing || g.battleMoving || g.state != StateBattle || g.battleTurn != 2 || g.battlePhase != screen.BattlePhaseInput {
				t.Fatalf("實行未回正常輸入：state=%v battleTurn=%d phase=%v", g.state, g.battleTurn, g.battlePhase)
			}
			if c.slot == 8 && movingFrames == 0 {
				t.Fatal("S1 未顯示已觀察到的移動中間幀")
			}
			movementFrame(t, g, label+"-input")
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			ship, fleet := bytes.Clone(g.slot.Ship), bytes.Clone(g.slot.Fleet)
			g.res.SaveDir = dir
			g.loadSlotFromDisk(0)
			if g.SaveErr() != nil || !bytes.Equal(ship, g.slot.Ship) || !bytes.Equal(fleet, g.slot.Fleet) {
				t.Fatalf("完整 SHIP／FLEET 正式存讀不同：%v", g.SaveErr())
			}
			t.Logf("正常S%d：星域%d、戰略回合%d、%d艦隊／%d分隊；受損%d／玩家受損%d／補給艦%d；結果幀補給%d／修理%d／移動%d；Input還原，SHIP %d／FLEET %d bytes存讀相同；seed戰略%08X、戰術12345678", c.slot-7, c.region, c.turn, c.fleets, c.units, damaged, humanDamaged, supplyKinds, supplyFrames, repairFrames, movingFrames, len(ship), len(fleet), strategyRandSeed)
		})
	}
}
