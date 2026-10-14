package game

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
	"github.com/wicanr2/logh3/internal/rules"
)

func TestFortressFrameJournalAndHDResult(t *testing.T) {
	g := fortressSceneGame(t, 1)
	g.fortressNoticeClick()
	g.fortressNoticeClick()
	s := g.engQueue[0].gun
	hd, err := presentation.Load(filepath.Join(root, "assets/hd"), g.res.Faces)
	if err != nil {
		t.Fatal(err)
	}
	var previous []byte
	for _, tick := range []int{0, s.schedule.lowerTick, s.schedule.resultTick - 1, s.schedule.resultTick} {
		g.engTick = tick
		fb := g.Frame()
		if !bytes.Equal(gfx.ReplayIndexed(fb.Commands), fb.Pix) {
			t.Fatal("要塞HD紀錄與正式色號不符", tick)
		}
		img := hd.Render(fb, g.Palette())
		var text []byte
		x := 384
		if s.base.Target.Side == 0 {
			x = 16
		}
		for y := 160 * presentation.Scale; y < 200*presentation.Scale; y++ {
			o := y*img.Stride + x*presentation.Scale*4
			text = append(text, img.Pix[o:o+240*presentation.Scale*4]...)
		}
		if tick == s.schedule.resultTick && bytes.Equal(text, previous) {
			t.Fatal("HD結果仍顯示扣損前分隊文字")
		}
		previous = text
	}
}

func TestFortressNoticeRegionMatchesOracle(t *testing.T) {
	native, err := os.ReadFile("../../workplace/oracle/fortress48-r133-r3-hit-a/frame-3765-006c-01-screen.bin")
	if err != nil {
		t.Skip("缺原版通知收據")
	}
	ram, err := os.ReadFile("../../workplace/oracle/fortress48-r133-r3-hit-a/gun-before-mem.bin")
	if err != nil {
		t.Fatal(err)
	}
	g := reachDashiliBattle(t)
	name := ram[0x54560 : 0x54560+17]
	name = name[:bytes.IndexByte(name, 0)]
	fb := gfx.NewFramebuffer(0)
	g.res.Engagement.DrawFortressNotice(fb, g.res.Font, g.res.FrameTiles, 0, name)
	// 原版本次輸入的游標也在通知框內，不能遮罩排除。
	g.res.Cursor.Draw(fb, 240, 200)
	count := 0
	for y := 180; y < 280; y++ {
		for x := 64; x < 416; x++ {
			i := y*640 + x
			if fb.Pix[i] != native[i] {
				count++
			}
		}
	}
	if count != 0 {
		t.Fatalf("發射通知352×100區域%d像素不同", count)
	}
}

// 明示局部診斷；正式玩家證據另存playtest123。
func fortressSceneGame(t *testing.T, animation int) *Game {
	t.Helper()
	g := reachDashiliBattle(t)
	fleetChainAssets(t, g)
	g.passiveBattleComputer = false
	g.fleetCombat = fleetCombatState{}
	p := &g.battlePlanets[0]
	p.Fortress, p.Owner, p.DefenceMax, p.GunState, p.GunAim = true, 0, 12800, 3, 64
	for _, u := range g.battleUnits {
		if u.Ships > 0 && u.Kind != 7 {
			p.X, p.Y = u.X, u.Y-60
			break
		}
	}
	g.battleRand = rules.NewRand(0x12345678)
	g.display.Anim = animation
	g.battleGunHits = g.fortressGunFire()
	if len(g.battleGunHits) < 2 {
		t.Fatalf("診斷未命中多分隊：%v", g.battleGunHits)
	}
	if !g.queueFortressEngagements(g.battleGunScenes) {
		t.Fatalf("缺要塞佇列：%v", g.saveErr)
	}
	if err := g.SaveSlot(t.TempDir(), 0); err == nil {
		t.Fatal("演出中允許公開保存")
	}
	return g
}

func TestFortressSceneNoticeAndOneClickPerResult(t *testing.T) {
	g := fortressSceneGame(t, 1)
	if g.currentFortressNotice() == nil || g.BattleInputReady() {
		t.Fatal("發射通知未阻止戰術輸入")
	}
	g.drawBattle()
	g.fortressNoticeClick()
	if g.currentFortressNotice() == nil || g.currentFortressNotice().noticePhase != 1 {
		t.Fatal("通知點擊跳過射程等待")
	}
	g.fortressNoticeClick()
	if g.state != StateEngagement || g.currentFortressNotice() != nil || g.engClick {
		t.Fatal("未進正式演出或點擊外溢")
	}
	s := g.engQueue[0].gun
	if !s.scene.First || s.scene.Unit < 0 {
		t.Fatal("第一目標遺失")
	}
	count := len(g.engQueue)
	g.engTick = s.schedule.resultTick - 1
	g.engagementClick()
	if len(g.engQueue) != count {
		t.Fatal("演出中點擊跳格")
	}
	g.engagementTick(1)
	if len(g.engQueue) != count-1 || g.engClick || g.engQueue[0].gun.scene.First {
		t.Fatal("一次點擊跳過多個結果")
	}
	v := g.fortressEngagementView(g.engQueue[0].gun)
	if len(v.ClearSides) == 0 {
		t.Fatal("下一分隊未清前目標視窗")
	}
	for _, p := range g.engQueue[0].gun.schedule.pictures {
		if len(p.segment) == 4 && p.segment[0] == 'P' && p.segment[2] == 'A' {
			t.Fatal("重播要塞開火動畫")
		}
	}
	a, b := g.drawEngagement(), g.drawEngagement()
	if a != b || g.saveErr != nil {
		t.Fatal("要塞繪圖或快取失敗", g.saveErr)
	}
}

func TestFortressSceneAnimationOffKeepsResult(t *testing.T) {
	g := fortressSceneGame(t, 0)
	g.fortressNoticeClick()
	g.fortressNoticeClick()
	s := g.engQueue[0].gun
	v := g.fortressEngagementView(s)
	if len(s.schedule.pictures) != 0 || len(s.schedule.sounds) != 0 || !v.ShowResult || v.Scene == nil || v.Target.Ships != uint16(s.scene.After.Units[s.scene.Unit].Ships) {
		t.Fatal("動畫關閉丟失結果或下方")
	}
	g.drawEngagement()
	if g.saveErr != nil {
		t.Fatalf("%v view=%+v", g.saveErr, v)
	}
}

func TestFortressSceneSoundGateUsesOneClock(t *testing.T) {
	g := fortressSceneGame(t, 1)
	fleetSoundAssets(t, g)
	g.display.AnimSpeed = 2
	g.soundEffects.remaining = 2 * time.Second
	g.fortressNoticeClick()
	g.fortressNoticeClick()
	for elapsed := time.Duration(0); elapsed+g.TimerTick() < 2*time.Second; elapsed += g.TimerTick() {
		g.Tick(g.TimerTick())
		if len(g.DrainSoundEffects()) != 0 {
			t.Fatal("音效時鐘重複扣除或提前播cue")
		}
	}
	g.SetSoundEffectPlaying(true)
	for i := 0; i < 4; i++ {
		g.Tick(g.TimerTick())
		if len(g.DrainSoundEffects()) != 0 {
			t.Fatal("未等實際播放端結束")
		}
	}
	g.SetSoundEffectPlaying(false)
	var ids []int
	for i := 0; i < 4 && len(ids) == 0; i++ {
		g.Tick(g.TimerTick())
		ids = g.DrainSoundEffects()
	}
	if len(ids) != 1 {
		t.Fatal("要塞cue沒有單次續播", ids)
	}
}

func TestFortressSceneEmptyShotOnlyNotices(t *testing.T) {
	g := reachDashiliBattle(t)
	g.passiveBattleComputer = false
	p := &g.battlePlanets[0]
	p.Fortress, p.DefenceMax, p.GunState, p.GunAim, p.X, p.Y = true, 12800, 3, 64, 64, 64
	g.battleRand = rules.NewRand(0x12345678)
	seed := g.battleRand.Seed()
	hits := g.fortressGunFire()
	if len(hits) != 0 || len(g.battleGunScenes) != 1 || g.battleGunScenes[0].Unit != -1 || g.battleRand.Seed() != seed {
		t.Fatal("空射改seed或生成目標")
	}
	if !g.queueFortressEngagements(g.battleGunScenes) || g.state != StateBattle {
		t.Fatal("空射沒有通知")
	}
}
