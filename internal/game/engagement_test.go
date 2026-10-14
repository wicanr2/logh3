package game

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 交戰演出（規格 056、`docs/re/324`、`docs/re/331`）：S3 第五回合同盟艦隊打帝國行星雷奧米爾
// 那一場。收據步數換成計時器 tick：底圖在 #1140.1406M 畫下去、設 8 tick，第 n 次遞減在
// #1140.145018M ＋ (n−1) × 200,000 步（`-tick 20000`），所以步數 s 的 tick ＝
// ⌊(s − 1140.145018M) ÷ 0.2M⌋ ＋ 1：`eng-open` #1141.5M ＝ 7（還沒有圖）、`eng-text` #1144M ＝ 20
// （兩張靜圖）、`eng-mid` #1150M ＝ 50（PPAB 四格）、`eng-mid2` #1156M ＝ 80（八格）、
// `eng-end` #1169M ＝ 145（艦隊那段畫完、擊毀 290、21710/22000）、`eng-after` #1172M ＝ 160（停頓中）。
//
// ⚠ **直接進入的診斷比對**：remake 的戰場模擬走不到與原版同一個第五回合（戰術 AI
// 與亂數各走各的），所以這裡把那一場的數直接排進演出佇列，只驗畫面那一層；
// 量表與戰場方塊是事先列出的排除區（`engagement-panel.json` 的 `known_diff`，理由在
// 規格 056 已知差異）；兩個視窗的圖（`COMANI.BMG`）整塊比。

func loadEngagement(t *testing.T, res *Resources) {
	t.Helper()
	if res.FleetPresentation == nil {
		spec, err := screen.LoadFleetPresentationSpec(filepath.Join(root, "docs/internals/data/fleet-presentation-spec.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.FleetPresentation = spec
	}
	if res.Engagement == nil {
		spec, err := screen.LoadEngagementSpec(filepath.Join(root, "docs/internals/data/engagement-panel.json"))
		if err != nil {
			t.Fatal(err)
		}
		res.Engagement = spec
	}
	if res.EngagementPanel == nil {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted", res.Engagement.Panel.File))
		if err != nil {
			t.Skipf("缺 %s，跳過", res.Engagement.Panel.File)
		}
		img, err := asset.DecodeAMGImage(b)
		if err != nil {
			t.Fatal(err)
		}
		res.EngagementPanel = img
	}
	if res.Comani == nil {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted", res.Engagement.Animation.File))
		if err != nil {
			t.Skipf("缺 %s，跳過", res.Engagement.Animation.File)
		}
		bmg, err := asset.OpenBMG(b)
		if err != nil {
			t.Fatal(err)
		}
		res.Comani = bmg
	}
	if res.FleetSceneSprites == nil {
		blocks, err := os.ReadFile(filepath.Join(root, "org_game/extracted/3DBLOCK.SPR"))
		if err != nil {
			t.Skipf("缺行星方塊素材：%v", err)
		}
		flags, err := os.ReadFile(filepath.Join(root, "org_game/extracted/COM_FLAG.SPR"))
		if err != nil {
			t.Skipf("缺行星旗標素材：%v", err)
		}
		res.FleetSceneSprites, err = asset.DecodeFleetSceneSprites(blocks, flags)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// 正式Game的交戰色盤對原版actual RGB；此入口仍是交戰診斷fixture。
func TestEngagementPaletteMatchesNative(t *testing.T) {
	g := reachEngagementFixture(t)
	b, err := os.ReadFile(filepath.Join(root, "workplace/oracle/eng-open/pal.bin"))
	if err != nil {
		t.Skipf("缺原版色盤收據：%v", err)
	}
	if len(b) < 16*3 {
		t.Fatal("原版色盤形狀錯誤")
	}
	for i := 0; i < 16; i++ {
		r, gc, bl := g.Palette().RGB8(i)
		if [3]uint8{r, gc, bl} != [3]uint8{b[3*i], b[3*i+1], b[3*i+2]} {
			t.Fatalf("色號%d的Game RGB不符原版", i)
		}
	}
	fb := g.drawEngagement()
	for _, p := range fb.Pix {
		r, gc, bl := g.Palette().RGB8(int(p))
		if [3]uint8{r, gc, bl} != [3]uint8{b[3*int(p)], b[3*int(p)+1], b[3*int(p)+2]} {
			t.Fatal("renderer色號未正確映射RGB")
		}
	}
	if out := os.Getenv("LOGH3_PRESENTATION_GAME_OUTPUT"); out != "" {
		if err = fb.WritePNG(filepath.Join(out, "engagement.png"), g.Palette()); err != nil {
			t.Fatal(err)
		}
	}
	g.engagementClick()
	g.engagementTick(1000)
	if g.state != StateBattle {
		t.Fatal("未回戰場")
	}
	if g.Palette() != g.res.Menu.Palette() {
		t.Fatal("退出交戰未恢復原色盤")
	}
}

// 雷奧米爾（Big5）。
var planetLeomir = []byte{0xB9, 0x70, 0xB6, 0xF8, 0xA6, 0xCC, 0xBA, 0xB8}

// reachDashiliBattle 把 S3 載進來，兩邊各搬一支艦隊到達希利（原版要六個回合
// 才會合在那裡），進戰場。
func reachDashiliBattle(t *testing.T) *Game {
	t.Helper()
	g := newTestGame(t)
	loadEngagement(t, g.res)
	const dashili = 26
	moved := [2]bool{}
	for i := range g.fleets {
		f := &g.fleets[i]
		if f.Faction < 0 || f.Faction > 1 || moved[f.Faction] {
			continue
		}
		// 同盟那一支挑費雪（提督 55，艦隊表第 31 筆）：原版那一場打雷奧米爾的就是他
		// （`fcx-engrec-*` 的記錄 `+5` ＝ 55 → `ds:8E34` ＝ C0 → 段 FÀD0，`docs/re/331`）。
		if f.Faction == 1 && f.Admiral != 55 {
			continue
		}
		f.Region, f.Target, f.Move, f.Parked = dashili, rules.NoTarget, rules.MoveParked, true
		moved[f.Faction] = true
	}
	g.StartBattle(dashili)
	if g.state != StateBattle {
		t.Fatalf("進戰場之後在 %v", g.state)
	}
	return g
}

// reachEngagementFixture 進達希利的戰場，再把那一場的數排進演出。
func reachEngagementFixture(t *testing.T) *Game {
	t.Helper()
	g := reachDashiliBattle(t)
	planet := -1
	for i := range g.battlePlanets {
		if string(g.battlePlanetName(i)) == string(planetLeomir) {
			planet = i
		}
	}
	if planet < 0 {
		t.Fatal("達希利的戰場上找不到雷奧米爾")
	}
	// 同盟那一支（收據裡看不清編號：第？艦隊 ？？？？？？？？？）。
	fleet := -1
	for _, f := range g.battleFleets {
		if f.Faction == 1 && g.fleetAdmiral(f.Index) == 55 {
			fleet = f.Index
			break
		}
	}
	if fleet < 0 {
		t.Fatal("戰場上沒有同盟艦隊")
	}
	g.engQueue = []engagement{{
		planet: planet, fleet: fleet,
		defence: 3300, defenceMax: 3300,
		ships: 22000, shipsMax: 22000, destroyed: 290,
		label: screen.EngagementDefence,
	}}
	g.fillEngagementNames(&g.engQueue[0])
	g.startEngagement()
	g.state = StateEngagement
	return g
}

func engagementSkips(g *Game) [][4]int {
	var out [][4]int
	for _, r := range g.res.Engagement.KnownDiff {
		out = append(out, [4]int{r.X, r.Y, r.X + r.W - 1, r.Y + r.H - 1})
	}
	return out
}

func TestEngagementMatchesOracle(t *testing.T) {
	g := reachEngagementFixture(t)
	e := g.engQueue[0]
	if e.planetKind != 'P' || e.planetSide != 0 || e.fleetSide != 1 || e.fleetName != [2]byte{'F', 0xC0} {
		t.Fatalf("段名與左右 %+v，原版 PPD0 左、FÀD0 右（費雪 → ds:8E34[55] ＝ C0）", e)
	}
	for _, c := range []struct {
		batch string
		tick  int
	}{{"eng-open", 7}, {"eng-text", 20}, {"eng-mid", 50}, {"eng-mid2", 80}, {"eng-end", 145}, {"eng-after", 160}} {
		g.engagementTick(c.tick - g.engTick)
		if g.state != StateEngagement {
			t.Fatalf("tick %d 就收場了", c.tick)
		}
		matchesExcept(t, g, c.batch, engagementSkips(g))
	}
	// 沒有按鍵：擊毀那一刻（tick 144）起停頓滿 90 tick 收場。原版 #1187.0M ＝ tick 235：
	// 它畫完方塊（`321F:0897` 約 26 萬步的 CPU 時間）才開始停頓，晚了一個 tick（規格 056 已知差異）。
	g.engagementTick(233 - g.engTick)
	if g.state != StateEngagement {
		t.Fatalf("tick 233 就收場了")
	}
	g.engagementTick(1)
	if g.state != StateBattle || len(g.engQueue) != 0 {
		t.Fatalf("tick 234 之後在 %v、佇列 %d", g.state, len(g.engQueue))
	}
}

// 播放中按鍵不跳格：排進佇列，到擊毀之後的停頓才讀、立刻收場（`fcx-engrec-click`：#1146M
// 左鍵 → #1169.5M 收場；停頓中 #1175M 點一下 → #1175.5M）。
func TestEngagementClickQueuesUntilPause(t *testing.T) {
	g := reachEngagementFixture(t)
	g.engagementTick(29)
	g.Click(300, 240)
	if g.state != StateEngagement {
		t.Fatal("播放中按一下就收場了")
	}
	g.engagementTick(80 - g.engTick)
	matchesExcept(t, g, "eng-mid2", engagementSkips(g)) // 畫面照走
	g.engagementTick(144 - g.engTick)
	if g.state != StateBattle {
		t.Fatalf("擊毀之後在 %v，原版停頓一開始就讀到按鍵收場", g.state)
	}
	g = reachEngagementFixture(t)
	g.engagementTick(174)
	g.RightClick(300, 240)
	if g.state != StateBattle {
		t.Fatalf("停頓中右鍵之後在 %v", g.state)
	}
}

// 回合處理裡打行星有結果就切演出，看完才做這一回合剩下的（回合數才加）。
func TestEngagementQueuedFromAssault(t *testing.T) {
	g := reachDashiliBattle(t)
	turn := g.battleTurn
	hits := []rules.PlanetAssaultHit{{Fleet: g.battleFleets[0].Index, Planet: 0, Damage: 100, Counter: 50}}
	if !g.queueEngagements(hits) {
		t.Fatal("打行星有結果卻沒排演出")
	}
	if g.state != StateEngagement {
		t.Fatalf("在 %v", g.state)
	}
	e := g.engQueue[0]
	if e.defence != g.battlePlanets[0].Defence+100 || e.destroyed != 50 ||
		e.ships != g.battleFleetShips(e.fleet)+50 || e.label != screen.EngagementLanding {
		t.Fatalf("演出的數 %+v", e)
	}
	g.engagementClick()
	g.engagementTick(1000)
	if g.state != StateBattle || g.battleTurn != turn+1 {
		t.Fatalf("看完之後在 %v、回合 %d（進來時 %d）", g.state, g.battleTurn, turn)
	}
}

// 動畫效果關掉（`ds:8F36 ＝ 0`，收據 `eng-skip-mid` #1150M）：四次畫圖都跳過，兩個視窗只剩
// 底圖，擊毀與停頓照舊、一開始就印結果（`docs/re/331` §2.3）。原版那一趟的停頓在 #1141.602M
// （tick 8）才開始、#1159.545M（tick 98）收場；remake 從 tick 0 起算、tick 90 收場——前面那
// 7 tick 花在哪沒讀（規格 056 已知差異）。
func TestEngagementAnimationOffMatchesOracle(t *testing.T) {
	g := reachEngagementFixture(t)
	g.display.Anim = 0
	g.engagementTick(50)
	if g.state != StateEngagement {
		t.Fatal("動畫關掉之後 tick 50 就收場了（停頓 90 tick）")
	}
	matchesExcept(t, g, "eng-skip-mid", engagementSkips(g))
	g.engagementTick(40)
	if g.state != StateBattle {
		t.Fatalf("tick 90 之後在 %v，原版停頓滿 90 tick 收場", g.state)
	}
}

// 094接線診斷：從typed Hit進正式佇列及renderer；不宣稱正常玩家路徑。
func TestEngagementGaugesQueuedAndDrawn(t *testing.T) {
	for _, faction := range []int{0, 1} {
		g := reachDashiliBattle(t)
		fleet := -1
		for _, f := range g.battleFleets {
			if f.Faction == faction {
				fleet = f.Index
				break
			}
		}
		if fleet < 0 {
			t.Fatal("缺艦隊")
		}
		g.battlePlanets[0].Defence = 1000
		snap := rules.PlanetDefenceGauges{BeforeIntegrity: 50, AfterIntegrity: 15, BeforeAttack: 1000, AfterAttack: 250}
		if !g.queueEngagements([]rules.PlanetAssaultHit{{Fleet: fleet, Planet: 0, Counter: 50, DefenceGauges: &snap}}) {
			t.Fatal("未排演出")
		}
		// 佇列擁有獨立快照，之後規則報告的變更不能污染當前演出。
		snap.BeforeIntegrity = 1
		for _, after := range []bool{false, true} {
			if after {
				_, g.engTick = g.engagementTimeline(g.engQueue[0])
			}
			v := g.engagementView()
			wantIntegrity, wantAttack := 50, 1000
			if after {
				wantIntegrity, wantAttack = 15, 250
			}
			if v.Gauges == nil || v.Gauges.Integrity != wantIntegrity || v.Gauges.Attack != wantAttack || v.Gauges.FleetFaction != faction {
				t.Fatalf("快照錯誤 %+v", v.Gauges)
			}
			fb := g.drawEngagement()
			fg := screen.EngagementGauge{FleetFaction: faction, Value: wantIntegrity}
			r, _ := fg.Rect()
			h, _ := fg.Height()
			fill := uint8(14)
			if faction == 1 {
				fill = 12
			}
			if fb.At(r.X+1, r.Y+r.H-h) != fill || fb.At(r.X+1, r.Y+r.H-h-1) != 0 {
				t.Fatal("正式renderer未畫當前完整度")
			}
			pg := screen.EngagementGauge{Planet: true, FleetFaction: faction, Value: 1000, Divisor: wantAttack}
			pr, _ := pg.Rect()
			ph, _ := pg.Height()
			pfill := uint8(12)
			if faction == 1 {
				pfill = 14
			}
			if fb.At(pr.X+1, pr.Y+pr.H-ph) != pfill {
				t.Fatal("正式renderer未畫行星量表")
			}
			if out := os.Getenv("LOGH3_GAUGE_GAME_OUTPUT"); out != "" {
				name := fmt.Sprintf("faction%d-after%t.png", faction, after)
				if err := fb.WritePNG(filepath.Join(out, name), g.Palette()); err != nil {
					t.Fatal(err)
				}
			}
		}
		g.engTick = 0
		g.display.Anim = 0
		if v := g.engagementView(); !v.ShowResult || v.Gauges.Integrity != 15 || v.Gauges.Attack != 250 {
			t.Fatal("動畫關閉未立即切到反擊後量表")
		}
	}
}
