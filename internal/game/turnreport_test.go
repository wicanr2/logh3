package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 回合裡的報告框（規格 059「報告框」、`docs/re/333`）：正常玩家路徑走到第三回合之後，與
// `tools/fleetcmd_receipts.sh` 第 24 段的收據整張比。
//
// 這幾張要電腦 AI 的走法與原版一致才到得了同一個局面：AI 命令的命令種類、依表格位置排的艦隊與配槽、
// 候選評分讀編制的艘數、進軍路線表、戰略打行星的多支攻方定位（`docs/re/335`）。

// reachPhezzanOrder 走到「占領 → 費沙 → 參謀提醒兩頁 → 決定」（`fcmd-occ-phz-done`）。
func reachPhezzanOrder(t *testing.T) *Game {
	t.Helper()
	g := reachFleetOccupyPick(t)
	g.Click(163, 136) // 費沙
	g.Click(300, 240)
	g.Click(300, 240)
	g.Click(432, 263) // 決定
	return g
}

// clickUntilReport 一直點 (560,125)，直到停在帶 marker 位元組的那一則報告框。
func clickUntilReport(t *testing.T, g *Game, marker []byte) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if g.State() == StateTurnReport && len(g.turnRep.pending) > 0 && bytesContains(g.turnRep.pending[0].text, marker) {
			return
		}
		if g.State() == StateCommandPhase || g.State() == StateSysMenu {
			t.Fatalf("回合跑完停在 %v，沒碰到那一則報告", g.State())
		}
		g.Click(560, 125)
	}
	t.Fatal("點了 64 下還沒到那一則報告")
}

func bytesContains(b, sub []byte) bool {
	for i := 0; i+len(sub) <= len(b); i++ {
		if string(b[i:i+len(sub)]) == string(sub) {
			return true
		}
	}
	return false
}

// Big5 標記：電腦那一路的「企圖占領」、人類那一路的「正在攻占」、士氣報告的「士氣值」、「已全軍覆沒」的「覆沒」。
var (
	markAICapture   = []byte("\xa5\xf8\xb9\xcf\xa5\x65\xbb\xe2")
	markCapture     = []byte("\xa5\xbf\xa6\x62\xa7\xf0\xa5\x65")
	markMorale      = []byte("\xa4\x68\xae\xf0\xad\xc8")
	markNoShips     = []byte("\xc2\xd0\xa8\x53")
	markAnnihilated = []byte("\xa5\xfe\xb7\xc0") // 「全滅」
)

// 拿下行星的訊息框（人類座位）：占領亞斯提那條路第五回合，實行前把巴溫的防衛力改成 100（收據同一個 poke），
// 一打就換手；90 tick 自己關。
func TestTurnReportCaptureMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	for turn := 0; turn < 4; turn++ {
		executeTurn(t, g)
	}
	start, _ := g.res.Planets.Range(16)
	binary.LittleEndian.PutUint16(g.res.Planets.Record(start)[0x14:], 100)
	pressExecute(g)
	clickUntilReport(t, g, markCapture)
	matches(t, g, "fcmd-turn-capture")
	tickTurnWindow(g, 89)
	if g.State() != StateTurnReport {
		t.Fatalf("89 tick 之後在 %v，原版等滿 90", g.State())
	}
	tickTurnWindow(g, 1)
	if g.State() == StateTurnReport && bytesContains(g.turnRep.pending[0].text, markCapture) {
		t.Fatal("90 tick 之後訊息框還開著")
	}
}

// 已全軍覆沒：同一條路第五回合實行前把第 0 組艘數改成 0（收據同一個 poke），走一格之後 `113A:03C5` 印；艦隊留著。
func TestTurnReportNoShipsMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	for turn := 0; turn < 4; turn++ {
		executeTurn(t, g)
	}
	g.fleets[0].Composition[0].Ships = 0
	pressExecute(g)
	clickUntilReport(t, g, markNoShips)
	matches(t, g, "fcmd-turn-no-ships")
	g.Click(560, 125)
	if len(g.fleets) == 0 || g.fleets[0].Faction != 0 || g.fleets[0].Flags&0x08 != 0 {
		t.Fatal("「已全軍覆沒」之後艦隊應該留著、bit3 清掉（原版 `0374:06BC`／`297B:034A` 零次）")
	}
}

// 電腦座位拿下行星：費沙那條第三回合，同盟第 13、16 艦隊在馬爾・亞迪特拿下威爾勃恩（兩支攻方）。
func TestTurnReportAICaptureMatchesOracle(t *testing.T) {
	g := reachPhezzanOrder(t)
	executeTurn(t, g)
	executeTurn(t, g)
	pressExecute(g)
	clickUntilReport(t, g, markAICapture)
	matches(t, g, "fcmd-turn-ai-capture")
}

// 士氣值低落：同一回合打費沙打到士氣 47，印在易主報告之後（回合尾巴 `1261:00C8`），印完清 bit2／bit3。
func TestTurnReportMoraleLowMatchesOracle(t *testing.T) {
	g := reachPhezzanOrder(t)
	executeTurn(t, g)
	executeTurn(t, g)
	pressExecute(g)
	clickUntilReport(t, g, markMorale)
	matches(t, g, "fcmd-turn-morale-low")
	if f := g.fleets[0]; f.Morale != 47 || f.Flags&0x0C != 0 {
		t.Fatalf("艦隊 0 士氣 %d、+00 %02X；原版 47、bit2／bit3 清掉", f.Morale, f.Flags)
	}
}

// 戰鬥中的士氣報告（`3826:0405` 臨時視窗，#72）：收據 `eng-morale-low` 是 S3 交戰第五回合結算的最後一步，
// 第１３艦隊（槽 12，舒坦梅茲）照規格 014 寫成士氣 40、bit0｜bit2。remake 的戰場沒有結算最後一步
// （`181B:03F0` → `185C:001D`），到不了同一個局面，所以這裡**只驗視窗**：拿收據畫面當底、畫上 remake 的
// 報告框再整張比——視窗外的像素本來就來自收據。排除收據游標 (240,240) 那一塊（點實行留下的，
// 不在視窗語意裡）。
func TestBattleMoraleReportWindowMatchesOracle(t *testing.T) {
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/eng-morale-low/screen.bin"))
	if err != nil {
		t.Skip("沒有收據 eng-morale-low，跳過（tools/engagement_receipts.sh 第 5 段）")
	}
	g := newTestGame(t)
	if g.res.FleetWindow, err = screen.LoadFleetWindowSpec(
		filepath.Join(root, "docs/internals/data/fleet-window.json")); err != nil {
		t.Fatal(err)
	}
	g.res.OrigDir = filepath.Join(root, "org_game/extracted")
	g.loadSlotFromDisk(10)
	f := g.fleetBySlot(t, 12)
	f.Flags |= rules.FleetFlagLowMorale | rules.FleetFlagMoraleFlip
	f.Morale = 40
	reps := g.moraleReports(f)
	if len(reps) != 1 {
		t.Fatalf("士氣報告 %d 則", len(reps))
	}
	face, err := g.res.Faces.Decode(reps[0].face)
	if err != nil {
		t.Fatal(err)
	}
	fb := gfx.NewFramebuffer(0)
	copy(fb.Pix, want)
	// 視窗矩形先塗掉：不這樣的話 remake 什麼都沒畫也會通過（底圖就是收據）。
	fb.FillRect(32, 180, 416, 140, 0)
	if miss := g.res.TurnWindow.DrawBattleReport(fb, g.res.Font, g.res.FrameTiles, reps[0].text, face); len(miss) != 0 {
		t.Fatalf("字庫缺字 %04X", miss)
	}
	bad := 0
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			if x >= 240 && x < 256 && y >= 240 && y < 260 {
				continue // 收據的游標
			}
			if fb.Pix[y*640+x] != want[y*640+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		if os.Getenv("LOGH3_DUMP_REMAKE") != "" {
			_ = os.WriteFile(filepath.Join(root, "workplace", "remake-eng-morale-low.bin"), fb.Pix, 0o644)
		}
		t.Fatalf("eng-morale-low：視窗內與原版不同的像素 %d", bad)
	}
}

// 艦隊全滅（#57，`0374:0608` → `06BC` → `073F` → `297B:034A`）：費沙那條第四回合，士氣 47 停在無主的費沙，
// 扣兵之後第 0 組 0 艘 →「…已經全滅。」→ 看完拿掉這一支、提督與副官失蹤（`fcmd-turn-annihilated-report`、
// `fcmd-turn-annihilated` 的 mem-in／mem-out）→ 第五回合命令階段整張（`fcmd-occ-phz-t4`）。
// 報告視窗 (320,160)–(639,319) 是刻意差異：原版格式字串多一個 `\x03` 畫成花紋，使用者裁決修成正常文字。
func TestTurnReportAnnihilatedMatchesOracle(t *testing.T) {
	out, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-turn-annihilated/mem-out.bin"))
	if err != nil {
		t.Skip("沒有收據 fcmd-turn-annihilated，跳過（tools/fleetcmd_receipts.sh 第 26 段）")
	}
	g := reachPhezzanOrder(t)
	executeTurn(t, g)
	executeTurn(t, g)
	// 第三回合照收據的輸入：士氣值低落那一則在地圖上 (300,240) 點掉（`fcmd-occ-phz-t3b`），游標留在艾流塞拉上；
	// 第四回合按系統指令時游標離開地圖，左下角摘要才清掉。
	pressExecute(g)
	clickUntilReport(t, g, markMorale)
	g.Click(300, 240)
	wantState(t, g, StateCommandPhase, "第三回合點掉士氣報告之後")
	matches(t, g, "fcmd-occ-phz-t3b")
	before := g.fleets[0]
	before.Composition = append([]rules.FleetGroup(nil), before.Composition...)
	pressExecute(g)
	clickUntilReport(t, g, markAnnihilated)
	matchesExcept(t, g, "fcmd-turn-annihilated-report", [][4]int{{320, 160, 639, 319}})
	f := g.fleets[0]
	if f.Number != 1 || f.Composition[0].Ships != 0 {
		t.Fatalf("報告開著時艦隊 0 是第 %d 艦隊、第 0 組 %d 艘；原版報告在拿掉之前、第 0 組已經 0 艘", f.Number, f.Composition[0].Ships)
	}
	adm, adj := f.Admiral, f.Adjutant
	in, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-turn-annihilated/mem-in.bin"))
	if err != nil {
		t.Fatal(err)
	}
	// 計分器 `ds:8FE4[對手]`（lin 5DB18）：耗損每組 ÷ 20 加給同盟。
	wantScore := int(binary.LittleEndian.Uint32(out[0x5DB18:])) - int(binary.LittleEndian.Uint32(in[0x5DB18:]))
	if score, gone := rules.StrategicAttrition(&before, int(g.admiralRecord(before.Admiral)[0x0B])); score != [2]int{0, wantScore} || !gone {
		t.Errorf("第四回合前的艦隊 0 單獨耗損：計分 %v、全滅 %v；原版同盟 +%d、全滅", score, gone, wantScore)
	}
	for k := range before.Composition {
		if before.Composition[k].Ships != f.Composition[k].Ships {
			t.Errorf("第 %d 組：單獨耗損 %d 艘、回合裡 %d 艘", k, before.Composition[k].Ships, f.Composition[k].Ships)
		}
	}
	region := f.Region
	g.Click(560, 125)
	if got, want := g.regionRecord(region)[0x28], out[0x5F2CC+region*126+0x28]; got != want {
		t.Errorf("星域 %d +0x28 = %d，原版 %d", region, got, want)
	}
	for _, fl := range g.fleets {
		if fl.Number == 1 && fl.Faction == 0 {
			t.Fatal("看完報告之後第１艦隊還在")
		}
	}
	for _, a := range []int{adm, adj} {
		if a < 0 {
			continue
		}
		got := g.admiralRecord(a)
		want := out[0x602AC+a*rules.AdmiralRecord-0 : 0x602AC+(a+1)*rules.AdmiralRecord]
		if got[0] != want[0] || got[0x09]&0x80 != want[0x09]&0x80 || string(got[2:6]) != string(want[2:6]) {
			t.Errorf("提督 %d：+00 %02X／+09 %02X／+02 % X，原版 %02X／%02X／% X", a, got[0], got[9], got[2:6], want[0], want[9], want[2:6])
		}
	}
	finishTurn(t, g)
	wantState(t, g, StateCommandPhase, "全滅那一回合跑完")
	matches(t, g, "fcmd-occ-phz-t4")
}
