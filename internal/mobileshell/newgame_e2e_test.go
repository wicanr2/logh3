package mobileshell

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/game"
)

// 行動版殼的整場玩家路徑（第177輪發行驗證）。與桌面 tools/scripts/new_game_battle_gui.py
// 同一條路：新遊戲 → 防拷 → 編隊、占領目標 → 實行到開戰 → 打完整場 → 存進度 → 取進度
// → 再實行到下一場戰鬥。差別是每一下都走行動版的觸控：原版座標先換成手機畫面上的
// 位置（戰略版面時地圖與卡片是裁開重排的），戰略版面裁掉的命令列改按側欄，右鍵改按
// 「退回」，數字鍵改按側欄鍵盤。手機上點不到的位置直接判失敗。
//
// 走完整場要數十萬格，只在 LOGH3_MOBILE_E2E=1 時跑；缺原版檔就 skip。
func TestMobileNewGameFullBattle(t *testing.T) {
	if os.Getenv("LOGH3_MOBILE_E2E") != "1" {
		t.Skip("設 LOGH3_MOBILE_E2E=1 才跑整場")
	}
	orig := os.Getenv("LOGH3_ORIG")
	if orig == "" {
		orig = "../../org_game/extracted"
	}
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); err != nil {
		t.Skip("缺原版檔：", err)
	}
	for _, c := range []struct {
		name    string
		faction int
		menuY   int
		target  [2]int
		loadY   int
		width   int
	}{
		{"empire", 0, 210, [2]int{176, 167}, 208, 1066},   // 20:9
		{"alliance", 1, 219, [2]int{151, 216}, 228, 1066}, // 20:9
		{"empire-16x9", 0, 210, [2]int{176, 167}, 208, 853},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := newDriver(t, orig, c.width)
			d.newGameToBattle(c.menuY, c.target)
			d.fullBattle(c.faction)
			d.saveAndLoad(c.loadY)
		})
	}
}

type driver struct {
	t     *testing.T
	s     *Shell
	g     *game.Game
	save  string
	lw    int
	ticks int
	seen  []string
}

func newDriver(t *testing.T, orig string, lw int) *driver {
	s := NewShell()
	s.UseNativeTouch()
	save := t.TempDir()
	s.Configure(Config{Orig: orig, Data: "../../docs/internals/data", Save: save})
	deadline := time.Now().Add(2 * time.Minute)
	for {
		s.mu.Lock()
		g, err, loading := s.g, s.loadErr, s.loading
		s.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		if g != nil && !loading {
			d := &driver{t: t, s: s, g: g, save: save, lw: lw}
			d.relayout()
			return d
		}
		if time.Now().After(deadline) {
			t.Fatal("載入逾時")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// relayout 照 Draw 的算法更新版面（測試沒有 GPU，不呼叫 Draw）。
func (d *driver) relayout() { d.s.lay = computeLayout(d.lw, d.g.PlainStrategic()) }

func (d *driver) step(n int) {
	for i := 0; i < n; i++ {
		if err := d.s.Update(); err != nil {
			d.t.Fatal(err)
		}
		d.ticks++
		d.relayout()
	}
}

func (d *driver) state() string { return d.g.State().String() }

func (d *driver) until(what string, max int, ok func() bool) {
	for i := 0; i < max; i++ {
		if ok() {
			return
		}
		d.step(1)
	}
	d.t.Fatalf("%s：%d 格後仍在 %s", what, max, d.state())
}

// tapLogical 在邏輯座標輕觸一次，再推進幾格讓遊戲處理。
func (d *driver) tapLogical(x, y float64) {
	d.s.QueueTap(x/float64(d.lw), y/480)
	d.step(3)
}

// click 是手機版的「在原版 (gx,gy) 左鍵」：畫面上看得到就點那裡；戰略版面裁掉的
// 命令列改按側欄對應的那一顆。
func (d *driver) click(gx, gy int) {
	d.t.Helper()
	d.relayout()
	if x, y, ok := d.s.lay.fromGame(gx, gy); ok {
		d.tapLogical(x, y)
		return
	}
	for _, b := range d.s.buttons(d.g) {
		if b.act != actGameClick {
			continue
		}
		if abs(gx-b.gx) <= b.gw/2 && abs(gy-b.gy) <= 10 {
			d.tapLogical(float64(b.x+b.w/2), float64(b.y+b.h/2))
			return
		}
	}
	d.t.Fatalf("手機畫面上點不到原版 (%d,%d)，狀態 %s、戰略版面 %v", gx, gy, d.state(), d.s.lay.strategic)
}

// side 按側欄上符合條件的按鈕。
func (d *driver) side(what string, match func(sideButton) bool) {
	d.t.Helper()
	for _, b := range d.s.buttons(d.g) {
		if match(b) {
			d.tapLogical(float64(b.x+b.w/2), float64(b.y+b.h/2))
			return
		}
	}
	d.t.Fatalf("側欄沒有「%s」，狀態 %s", what, d.state())
}

func (d *driver) back() { d.side("退回", func(b sideButton) bool { return b.act == actBack }) }

func (d *driver) newGameToBattle(menuY int, target [2]int) {
	d.until("開頭動畫", 20000, func() bool {
		switch d.g.State() {
		case game.StateIntro, game.StateTitle:
			d.tapLogical(float64(d.lw)/2, 240)
			return false
		}
		return d.g.State() == game.StateMainMenu
	})
	d.click(300, menuY)
	d.until("防拷問答", 600, func() bool { return d.g.State() == game.StateCopyProtection })
	for _, v := range []byte("138") {
		d.side("數字"+string(v), func(b sideButton) bool { return b.act == actDigit && b.value == v })
	}
	d.side("決定", func(b sideButton) bool { return b.act == actEnter })
	for i := 0; i < 8 && d.g.State() != game.StateCommandPhase; i++ {
		d.back()
		d.step(20)
	}
	if d.g.State() != game.StateCommandPhase {
		d.t.Fatalf("月初報告之後在 %s", d.state())
	}
	if want := computeLayout(d.lw, true).strategic; d.s.lay.strategic != want {
		d.t.Fatalf("寬 %d 的戰略地圖版面 %v，預期 %v", d.lw, d.s.lay.strategic, want)
	}
	for _, p := range [][2]int{{352, 125}, {400, 125}, {400, 85}, {488, 330}, {480, 350}, {510, 350}, {450, 308}, {490, 345}, {497, 332}, target, {432, 263}} {
		d.click(p[0], p[1])
	}
	if d.g.State() != game.StateCommandPhase {
		d.t.Fatalf("下令之後在 %s", d.state())
	}
	d.executeUntil("開戰", 400000, func() bool { return d.g.State() == game.StateBattle })
	d.until("戰場就緒", 40000, d.g.BattleInputReady)
}

var reportStates = map[string]bool{"StateOccupationReport": true, "StateRebellionReport": true, "StateRegionFlipReport": true,
	"StateTurnWindow": true, "StateTurnReport": true, "StateShipDelivery": true, "StateAdmiralFateReport": true}

// executeUntil 是戰略回合：實行、點掉報告，done 成立就停。
func (d *driver) executeUntil(what string, max int, done func() bool) {
	d.t.Helper()
	for i := 0; i < max; i++ {
		st := d.state()
		if len(d.seen) == 0 || d.seen[len(d.seen)-1] != st {
			d.seen = append(d.seen, st)
		}
		if done() {
			return
		}
		switch {
		case st == "StateOccupyAsk":
			d.t.Fatalf("%s：未預期的占領問話", what)
		case st == "StateCommandPhase":
			d.click(544, 125)
			d.click(560, 125)
		case st == "StateMainGame" || st == "StateShipOrder":
			d.back()
		case reportStates[st]:
			d.click(560, 125)
		default:
			d.step(1)
		}
	}
	d.t.Fatalf("%s 逾時，最後狀態 %v", what, d.seen[maxInt(0, len(d.seen)-12):])
}

// fullBattle 打完整場：有交火輸入就點，可以下令就開戰術選單按實行，直到離開戰場。
func (d *driver) fullBattle(faction int) {
	obs := d.g.TacticalObservation()
	own := 0
	for _, u := range obs.Units {
		if u.Faction == faction && u.Squadron == 0 {
			own++
		}
	}
	if own == 0 {
		d.t.Fatal("戰場上沒有玩家旗艦")
	}
	d.t.Logf("開戰星域 %d，第 %d 格", obs.Region, d.ticks)
	maxTurn, start := 0, d.ticks
	for i := 0; i < 2000000; i++ {
		st := d.state()
		switch {
		case st == "StateBattle":
			o := d.g.TacticalObservation()
			if o.Turn > maxTurn {
				maxTurn = o.Turn
			}
			if o.CombatInput != "" {
				p := o.CombatPoint
				if o.CombatInput == "target" {
					p = [2]int{240, 470}
				}
				d.click(p[0], p[1])
			} else if d.g.BattleInputReady() {
				d.click(600, 240)
				d.click(600, 242)
			} else {
				d.step(1)
			}
		case reportStates[st]:
			d.click(560, 125)
		case st == "StateEngagement" || st == "StateTurnThinking":
			d.step(1)
		case st == "StateCommandPhase" || st == "StateMainGame" || st == "StateShipOrder":
			d.t.Logf("離開戰場：%s，打到第 %d 回合，戰場 %d 格（%v）", st, maxTurn, d.ticks-start, time.Duration(d.ticks-start)*time.Second/60)
			if maxTurn < 9 {
				d.t.Errorf("戰場只打到第 %d 回合就離開", maxTurn)
			}
			d.executeUntil("戰後回指令畫面", 200000, func() bool { return d.g.State() == game.StateCommandPhase })
			return
		default:
			if st == "StateOccupyAsk" {
				d.t.Fatalf("未預期的畫面 %s", st)
			}
			d.step(1)
		}
	}
	d.t.Fatalf("戰場打不完，在 %s", d.state())
}

// saveAndLoad 正常存進度、取進度，再實行到下一場戰鬥或三回合。
func (d *driver) saveAndLoad(loadY int) {
	d.click(544, 125)
	d.click(560, 165)
	d.until("存進度畫面", 200, func() bool { return d.g.State() == game.StateSaveGame })
	d.click(300, 142)
	d.step(30)
	var tables int
	entries, _ := os.ReadDir(d.save)
	for _, e := range entries {
		switch strings.SplitN(e.Name(), ".", 2)[0] {
		case "STELLAR", "PLANET", "ADMIRAL", "FLEET", "SHIP", "WORK":
			tables++
		}
	}
	if tables < 6 {
		d.t.Fatalf("存檔只有 %d 個表：%v", tables, entries)
	}
	// 存檔雜湊寫出來，給發行驗證與桌面同一條路徑的存檔逐位元組比對。
	if out := os.Getenv("LOGH3_MOBILE_E2E_OUT"); out != "" {
		sums := map[string]string{}
		for _, e := range entries {
			b, err := os.ReadFile(filepath.Join(d.save, e.Name()))
			if err != nil {
				d.t.Fatal(err)
			}
			sums[e.Name()] = fmt.Sprintf("%x", sha256.Sum256(b))
		}
		b, _ := json.MarshalIndent(map[string]any{"subtest": d.t.Name(), "turn": d.g.Turn(), "saved_sha256": sums}, "", "  ")
		name := strings.ReplaceAll(d.t.Name(), "/", "-") + ".json"
		if err := os.WriteFile(filepath.Join(out, name), b, 0o644); err != nil {
			d.t.Fatal(err)
		}
	}
	d.back()
	d.until("存完回指令畫面", 200, func() bool { return d.g.State() == game.StateCommandPhase })
	before := d.g.Turn()
	d.click(544, 125)
	d.click(560, 145)
	d.until("取進度畫面", 200, func() bool { return d.g.State() == game.StateLoadGame })
	d.click(300, 122)
	d.until("選陣營", 200, func() bool { return d.g.State() == game.StateLoadFaction })
	d.click(318, loadY)
	d.until("讀回", 2000, func() bool { return d.g.State() == game.StateCommandPhase })
	if got := d.g.Turn(); got != before {
		d.t.Fatalf("讀回回合 %d，存檔時 %d", got, before)
	}
	for n := 0; n < 3; n++ {
		d.click(544, 125)
		d.click(560, 125)
		d.executeUntil("讀回後回合", 400000, func() bool {
			s := d.g.State()
			return s == game.StateCommandPhase || s == game.StateBattle
		})
		if d.g.State() == game.StateBattle {
			d.until("讀回後戰場就緒", 40000, d.g.BattleInputReady)
			d.t.Logf("讀回後再開戰，共 %d 格", d.ticks)
			return
		}
	}
	d.t.Logf("讀回後三回合沒有開戰，停在 %s", d.state())
}

func abs(x int) int { return int(math.Abs(float64(x))) }

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
