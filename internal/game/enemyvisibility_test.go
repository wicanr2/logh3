package game

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

func TestEnemyVisibilityPresentationIsolation(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if g.ShowEnemies() {
		t.Fatal("預設不能全顯示")
	}
	g.MoveMouse(600, 430)
	beforeFrame := bytes.Clone(g.Frame().Pix)
	before := append([]rules.BattleUnit(nil), g.battleUnits...)
	dirs := make(map[int]int)
	for k, v := range g.battleSpotting.directions {
		dirs[k] = v
	}
	seed := g.rand94DC().Seed()
	battleSeed := g.battleRand.Seed()
	paths := []string{t.TempDir(), t.TempDir()}
	if err := g.SaveSlot(paths[0], 0); err != nil {
		t.Fatal(err)
	}
	g.ToggleEnemyVisibility()
	units, directions := g.battleDisplayUnits()
	enemies := 0
	for i, u := range before {
		if u.Out || u.Faction == g.faction {
			if units[i] != u {
				t.Fatal("改到己方或出局分隊")
			}
			continue
		}
		enemies++
		v := units[i]
		if v.Detail != rules.SpottingMax || v.MarkerFlags&0xC0 != 0xC0 || v.MarkerX != u.X || v.MarkerY != u.Y {
			t.Fatal("敵軍顯示副本未揭露實際位置")
		}
		if n := g.displayBattleSprite(v); n == rules.BlankSprite {
			t.Fatal("敵軍仍為空圖示")
		}
		if u.Squadron == 0 {
			if f := g.battleFleet(u.Fleet); f != nil && directions[f.Index] != rules.DirectionIndex(f.Facing) {
				t.Fatal("敵軍方向沿舊索敵快取")
			}
		}
	}
	if enemies == 0 {
		t.Fatal("正常S3缺敵軍，測試沒有作用")
	}
	g.Frame()
	g.ToggleEnemyVisibility()
	if !bytes.Equal(beforeFrame, g.Frame().Pix) {
		t.Fatal("關閉後未恢復原畫面")
	}
	if !reflect.DeepEqual(before, g.battleUnits) || !maps.Equal(dirs, g.battleSpotting.directions) || g.rand94DC().Seed() != seed || g.battleRand.Seed() != battleSeed {
		t.Fatal("顯示切換修改規則資料、快取或亂數")
	}
	if err := g.SaveSlot(paths[1], 0); err != nil {
		t.Fatal(err)
	}
	assertVisibilitySavesEqual(t, paths[0], paths[1])
}

func assertVisibilitySavesEqual(t *testing.T, a, b string) {
	t.Helper()
	for _, stem := range []string{"ADMIRAL", "FLEET", "SHIP", "WORK", "STELLAR", "PLANET"} {
		x, e := os.ReadFile(filepath.Join(a, stem+".DT0"))
		if e != nil {
			t.Fatal(e)
		}
		y, e := os.ReadFile(filepath.Join(b, stem+".DT0"))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(x, y) {
			t.Fatal(stem, "因顯示偏好而不同")
		}
	}
}

func TestEnemyVisibilityHoverAndSettings(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	enemy := -1
	for i, u := range g.battleUnits {
		if !u.Out && u.Faction != g.faction {
			enemy = i
			break
		}
	}
	if enemy < 0 {
		t.Fatal("缺敵軍")
	}
	u := g.battleUnits[enemy]
	// 窄幅hover診斷只移鏡頭，不替代正常X11進場驗收。
	g.battleView = screen.BattleView{Col: u.X/16 - 2, Row: u.Y/20 - 2}
	g.battleSpotting.view = g.battleView
	g.Frame()
	x, y := unitScreenPos(g, enemy)
	g.MoveMouse(x+8, y+8)
	if g.battleHoverOn {
		t.Fatal("關閉時洩漏敵軍摘要")
	}
	g.ToggleEnemyVisibility()
	if !g.battleHoverOn || g.battleHoverUnit != enemy {
		t.Fatal("開啟未顯示敵軍摘要")
	}
	g.Click(x+8, y+8)
	if g.CommandMenuOpen() {
		t.Fatal("展示模式允許指揮敵軍")
	}
	g.ToggleEnemyVisibility()
	if g.battleHoverOn || g.battleHoverMinimapLine {
		t.Fatal("關閉後敵軍摘要或目的地殘留")
	}
	g.state = StateSettings
	r := enemyVisibilityButton
	g.Click((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)
	if !g.ShowEnemies() {
		t.Fatal("設定鈕未命中實際繪製位置")
	}
	g.Click(r.Max.X, r.Max.Y)
	if !g.ShowEnemies() {
		t.Fatal("設定鈕邊界外誤觸")
	}
	g.Click(r.Min.X+4, r.Min.Y+4)
	if g.ShowEnemies() {
		t.Fatal("設定鈕不能關閉")
	}
	for _, state := range []State{StateIntro, StateTitle, StateMainMenu, StateCopyProtection} {
		g.state = state
		g.ToggleEnemyVisibility()
		if g.ShowEnemies() {
			t.Fatal("啟動畫面接受敵軍切換")
		}
	}
}

func TestEnemyVisibilityDoesNotChangeNormalTurn(t *testing.T) {
	var games []*Game
	var paths []string
	for _, on := range []bool{false, true} {
		g := commandTestGame(t)
		if g == nil {
			return
		}
		if err := g.SetAILevel(1); err != nil {
			t.Fatal(err)
		}
		g.SetShowEnemies(on)
		g.MoveMouse(600, 430)
		entryTurn := g.battleTurn
		// 相同初始局面與固定正式種子，兩側只差顯示選項。
		g.Click(100, 100)
		g.Click(100, 110)
		completed := false
		for n := 0; n < 30000; n++ {
			if g.State() == StateTurnReport {
				g.Click(240, 240)
			} else {
				g.Tick(time.Second / 60)
			}
			g.Frame()
			if g.State() == StateBattle && g.battleTurn > entryTurn && g.battleInputIdle() && !g.battleMenu {
				completed = true
				break
			}
		}
		if !completed {
			t.Fatal("正常回合未返回輸入")
		}
		path := t.TempDir()
		if err := g.SaveSlot(path, 0); err != nil {
			t.Fatal(err)
		}
		games = append(games, g)
		paths = append(paths, path)
	}
	assertVisibilitySavesEqual(t, paths[0], paths[1])
	if games[0].battleRand.Seed() != games[1].battleRand.Seed() || games[0].rand94DC().Seed() != games[1].rand94DC().Seed() || !reflect.DeepEqual(games[0].battleUnits, games[1].battleUnits) {
		t.Fatal("顯示偏好改變正常戰鬥結果或亂數")
	}
}
