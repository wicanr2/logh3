package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/rules"
)

func TestNormalArrivalCoordinatesAndSave(t *testing.T) {
	m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal89-r50-arrival-a/entry.bin"))
	if err != nil {
		t.Skipf("缺第50輪原版完整RAM：%v", err)
	}
	if len(m) != 1048576 {
		t.Fatal("原版RAM長度錯誤")
	}
	if fmtHash(sha256.Sum256(m)) != "ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97" {
		t.Fatal("第50輪原版入口RAM身份不同")
	}
	for _, level := range []int{-1, 1} {
		t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) {
			g := reachConfiguredAI(t, level, 0)
			// 舊AI助手只載戰略及地圖；補正式bootstrap已有的配置資料。
			data := filepath.Join(root, "docs/internals/data")
			if g.res.Entries, err = rules.LoadBattleEntries(filepath.Join(data, "battle-entry-points.json")); err != nil {
				t.Fatal(err)
			}
			if g.res.Formation, err = rules.LoadFormationSlots(filepath.Join(data, "battle-formation-slots.json")); err != nil {
				t.Fatal(err)
			}
			if g.res.Terrain, err = rules.LoadBattleTerrain(filepath.Join(data, "battle-terrain.json")); err != nil {
				t.Fatal(err)
			}
			click := func(x, y int) {
				g.MoveMouse(x, y)
				g.Click(x, y)
				g.Frame()
			}
			click(544, 125)
			click(560, 145)
			click(300, 322)
			click(318, 208)
			wantState(t, g, StateCommandPhase, "正常S3載入")
			g.rand94DC().SetSeed(0xD3DC167E)
			before := bytes.Clone(g.slot.Fleet)
			click(544, 125)
			click(560, 125)
			modal := false
			for ticks := 0; ticks < 15000 && g.state != StateBattle; ticks++ {
				if g.state == StateTurnWindow && !modal {
					f := g.fleets[g.turnWin.fleet]
					o := f.Index * rules.FleetRecord
					if f.Index != 8 || f.Region != 25 || g.turnWin.from != 23 ||
						!bytes.Equal(g.slot.Fleet[o+0x4A:o+0x4D], before[o+0x4A:o+0x4D]) ||
						!bytes.Equal(g.slot.Fleet[o+0x5C:o+0x60], before[o+0x5C:o+0x60]) {
						t.Fatal("移動過場尚未結束就更新戰術配置，或來源遺失")
					}
					modal = true
				}
				switch g.state {
				case StateTurnThinking, StateTurnWindow, StateTurnReport, StateEngagement:
					g.Tick(time.Second / 60)
				case StateOccupationReport, StateRebellionReport, StateRegionFlipReport,
					StateAdmiralFateReport, StateShipDelivery, StatePhezzanReport:
					click(240, 240)
				default:
					t.Fatalf("實行停在%v", g.state)
				}
				g.Frame()
			}
			if !modal || g.state != StateBattle || g.battleRegion != 26 || g.turn != 380 ||
				g.rand94DC().Seed() != 0xB6E4B78A || len(g.battleFleets) != 7 {
				t.Fatal("正常首戰身份／固定seed結果不符")
			}
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			work, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(work, m[0x5DA44:0x5DA44+rules.WorkRecord]) {
				var offsets []int
				for i, b := range work {
					if b != m[0x5DA44+i] {
						offsets = append(offsets, i)
					}
				}
				t.Fatalf("正常首戰完整WORK不同，位移：%X", offsets)
			}
			fleets, err := os.ReadFile(filepath.Join(dir, "FLEET.DT0"))
			if err != nil {
				t.Fatal(err)
			}
			ships, err := os.ReadFile(filepath.Join(dir, "SHIP.DT0"))
			if err != nil {
				t.Fatal(err)
			}
			pools := map[int]bool{}
			for _, index := range []int{8, 12, 14, 16, 17, 28, 29, 30, 31} {
				o := index * rules.FleetRecord
				orig := m[0x60BA4+o : 0x60BA4+o+rules.FleetRecord]
				for _, field := range [][2]int{{0x4A, 3}, {0x5C, 4}, {0x64, 4}} {
					a, n := field[0], field[1]
					if !bytes.Equal(fleets[o+a:o+a+n], orig[a:a+n]) {
						t.Fatalf("艦隊%d欄位+%02X不符原版", index, a)
					}
				}
				for j := 0; j < int(orig[0x0F]); j++ {
					pools[int(binary.LittleEndian.Uint16(orig[0x10+4*j:]))] = true
				}
				pools[int(binary.LittleEndian.Uint16(orig[0x38:]))] = true
			}
			if len(pools) != 70 {
				t.Fatalf("移動池身份不同：%d", len(pools))
			}
			for index := range pools {
				o := index * rules.ShipRecord
				orig := m[0x668B2+o : 0x668B2+o+rules.ShipRecord]
				for _, field := range [][2]int{{2, 1}, {0x28, 4}, {0x34, 4}, {0x16, 4}, {0x1E, 1}} {
					a, n := field[0], field[1]
					if !bytes.Equal(ships[o+a:o+a+n], orig[a:a+n]) {
						t.Fatalf("池%d欄位+%02X不符原版", index, a)
					}
				}
			}
			for _, u := range g.battleUnits {
				pool := g.slot.FleetGroupShips(u.Fleet)[u.Squadron]
				o := 0x668B2 + pool*rules.ShipRecord
				if u.X != int(binary.LittleEndian.Uint16(m[o+0x34:])) ||
					u.Y != int(binary.LittleEndian.Uint16(m[o+0x36:])) ||
					u.TargetX != int(binary.LittleEndian.Uint16(m[o+0x28:])) ||
					u.TargetY != int(binary.LittleEndian.Uint16(m[o+0x2A:])) {
					t.Fatal("戰場入口重新覆寫已配置的現在／集結座標")
				}
			}
			// 原版26E5:00AC刷新全部55筆，包含不在battleUnits的七筆偵察池。
			cached := map[int]bool{}
			for _, u := range g.battleUnits {
				cached[g.battleUnitPool(u)] = true
			}
			for _, bf := range g.battleFleets {
				f := g.fleetByIndex(bf.Index)
				if f == nil || bf.ScoutMode != 0xFF {
					t.Fatal("正常首戰的未出擊偵察隊身份不同")
				}
				cached[f.Scout] = true
				for _, u := range g.battleUnits {
					if g.battleUnitPool(u) == f.Scout {
						t.Fatal("未出擊偵察隊進入可點選單位清單")
					}
				}
			}
			if len(cached) != 55 || len(g.battleUnits) != 48 {
				t.Fatal("正常首戰的完整池與可點選單位數不同")
			}
			for pool := range cached {
				want := m[0x668B2+pool*rules.ShipRecord+0x2F]
				if got := ships[pool*rules.ShipRecord+0x2F]; got != want {
					t.Fatalf("原池%d正式保存圖示快取%d，原版%d", pool, got, want)
				}
				if got, want := ships[pool*rules.ShipRecord+0x30], m[0x668B2+pool*rules.ShipRecord+0x30]; got != want {
					t.Fatalf("原池%d正式保存顯示順位%d，原版%d", pool, got, want)
				}
			}
			for _, u := range g.battleUnits {
				pool := g.battleUnitPool(u)
				if u.DisplayPriority != int(m[0x668B2+pool*rules.ShipRecord+0x30]) {
					t.Fatalf("原池%d的typed順位未同步", pool)
				}
			}
			loaded := newTestGame(t)
			loaded.res.SaveDir = dir
			loaded.loadSlotFromDisk(0)
			if loaded.SaveErr() != nil || loaded.slot == nil || !bytes.Equal(loaded.slot.Ship, ships) {
				t.Fatal("正常首戰完整SHIP正式存讀不同")
			}
			if !bytes.Equal(loaded.slot.Work, work) || loaded.slot.MoveCost != [2]int{} ||
				loaded.slot.Upkeep != [2]int{2342, 951} {
				t.Fatal("正常首戰完整WORK／typed經費正式重讀不同")
			}
			t.Log("55筆快取與七筆未出擊偵察池符合原版，48筆可點選單位及完整SHIP存讀保持")
			t.Log("正常S3九艦隊／七十筆池、過場前後及正式保存符合原版")
		})
	}
}
