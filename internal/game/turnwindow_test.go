package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 回合裡的過場視窗（規格 059、`docs/re/332`）：與 `tools/fleetcmd_receipts.sh` 第 23 段的
// `fcmd-turn-*` 收據整張比。收據的游標停在「實行」(560,125)。

// pressExecute 按 系統指令 → 實行，回合跑到第一個停下來的地方。
func pressExecute(g *Game) {
	g.Click(544, 125)
	g.Click(560, 125)
}

// tickTurnWindow 讓過場視窗走 n 個計時器 tick。
func tickTurnWindow(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.Tick(g.engagementTickDuration())
	}
}

func wantState(t *testing.T, g *Game, want State, when string) {
	t.Helper()
	if g.State() != want {
		t.Fatalf("%s在 %v，應該是 %v", when, g.State(), want)
	}
}

// 移動到亞爾提那：出發前往（瓦爾哈拉塗色 5）→ 22 tick → 已經到達（亞爾提那塗色 5）→ 90 tick →
// 回命令階段（`fcmd-exec`）。
func TestTurnWindowDepartArriveMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)  // 亞爾提那
	g.Click(432, 263) // 「決定」
	pressExecute(g)
	wantState(t, g, StateTurnWindow, "實行之後")
	if f := g.fleets[0]; f.Region != 2 {
		t.Fatalf("過場開著時艦隊在 %d；原版 `+05C4` 先改位置才印（應為 2）", f.Region)
	}
	matches(t, g, "fcmd-turn-depart")
	tickTurnWindow(g, 21)
	matches(t, g, "fcmd-turn-depart")
	tickTurnWindow(g, 1)
	wantState(t, g, StateTurnWindow, "22 tick 之後")
	matches(t, g, "fcmd-turn-arrive")
	tickTurnWindow(g, 89)
	matches(t, g, "fcmd-turn-arrive")
	tickTurnWindow(g, 1)
	wantState(t, g, StateCommandPhase, "再 90 tick 之後")
	matches(t, g, "fcmd-exec")
}

// 22 tick 那段點一下就換第二段；一下只結束一段（`fcmd-turn-clk22`）。右鍵同樣結束。
func TestTurnWindowClickEndsOneStage(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)
	g.Click(432, 263)
	pressExecute(g)
	g.Click(560, 125)
	wantState(t, g, StateTurnWindow, "點一下之後")
	matches(t, g, "fcmd-turn-clk22")
	g.RightClick(560, 125)
	wantState(t, g, StateCommandPhase, "右鍵之後")
	matches(t, g, "fcmd-exec")
}

// 駐留到亞爾提那：第二段是「已經駐留在星系上。」，解散在那一張之後。
func TestTurnWindowGarrisonMatchesOracle(t *testing.T) {
	g := reachFleetWindow(t)
	g.Click(490, 345) // 「作戰」
	g.Click(497, 352) // 「駐留」
	g.Click(85, 128)  // 亞爾提那
	g.Click(432, 263) // 「決定」
	pressExecute(g)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-depart")
	g.Click(560, 125)
	if len(g.fleets) == 0 || g.fleets[0].Faction != 0 {
		t.Fatalf("「已經駐留」那一張還開著，艦隊已經解散；原版在等完之後才 `297B:025D`")
	}
	matches(t, g, "fcmd-turn-garrison")
	g.Click(560, 125)
	wantState(t, g, StateCommandPhase, "關掉之後")
	matches(t, g, "fcmd-garrison-t1")
}

// 占領往亞斯提的第一回合：沒到目的地，第二段不印到達句。
func TestTurnWindowOccupyOnwardMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	pressExecute(g)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-occ-depart")
	g.Click(560, 125)
	matches(t, g, "fcmd-turn-occ-onward")
}

// 經費不足：實行之前把帝國軍事費改成 32（收據 poke `ds:8F52[0]`），扣完維持費 22 剩 10 < 移動費 26。
func TestTurnWindowNoFundsMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)
	g.Click(432, 263)
	g.Click(544, 125)
	g.military[0] = 32
	g.Click(560, 125)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-nofund")
	tickTurnWindow(g, 89)
	matches(t, g, "fcmd-turn-nofund")
	tickTurnWindow(g, 1)
	wantState(t, g, StateCommandPhase, "90 tick 之後")
	if f := g.fleets[0]; f.Region != 0 || f.Target != 2 || f.Move != rules.MoveMove || g.military[0] != 10 {
		t.Fatalf("經費不足之後艦隊 %+v、軍事費 %d；原版不動、命令留著、剩 10", f, g.military[0])
	}
}

// 被擋：實行之前把亞爾提那（星域 2）的擁有者改成同盟（`fcmd-turn-block-pre`，實行前 poke，
// 地圖跟著重上色）。
func TestTurnWindowBlockedMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)
	g.Click(432, 263)
	g.Click(544, 125)
	g.regionRecord(2)[regionOwnerOff] = 1
	g.Click(560, 125)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-block-pre")
	g.Click(560, 125)
	if f := g.fleets[0]; f.Region != 0 || f.Target != rules.NoTarget || f.Move != rules.MoveParked {
		t.Fatalf("被擋之後艦隊 %+v；原版 +0x07 FF、+0x08 5、不動", f)
	}
}

// 問話：占領往亞斯提走到亞爾提那之後，把亞爾維斯（星域 7）改成同盟再實行 →
// 「亞斯提無法直接到達此星系。亞爾維斯要占領此星系嗎？」→ 占領 → 出發前往（亞爾提那 → 亞爾維斯）→ 沒有到達句。
func TestTurnWindowOccupyAskMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	executeTurn(t, g)
	if g.fleets[0].Region != 2 {
		t.Fatalf("第一回合之後艦隊在 %d，應在亞爾提那", g.fleets[0].Region)
	}
	g.Click(544, 125)
	g.regionRecord(7)[regionOwnerOff] = 1
	g.Click(560, 125)
	wantState(t, g, StateOccupyAsk, "實行之後")
	matches(t, g, "fcmd-turn-ask")
	g.Click(424, 270) // 「占領」
	wantState(t, g, StateTurnWindow, "點了占領之後")
	matches(t, g, "fcmd-turn-ask-depart")
	g.Click(424, 270)
	matches(t, g, "fcmd-turn-ask-onward")
}

// clickThroughUntil 在回合停下來的畫面上一直點 (560,125)，直到停在 want；回合跑完還沒碰到就失敗。
func clickThroughUntil(t *testing.T, g *Game, want State) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if g.State() == want {
			return
		}
		switch g.State() {
		case StateTurnWindow, StateTurnReport, StateOccupationReport, StateRebellionReport, StateRegionFlipReport:
			g.Click(560, 125)
		default:
			t.Fatalf("回合跑完停在 %v，沒碰到 %v", g.State(), want)
		}
	}
	t.Fatalf("點了 64 下還沒到 %v（在 %v）", want, g.State())
}

// 拿下行星的訊息框（`20B4:1922`，人類座位）：占領亞斯提那條路第五回合，實行前把巴溫的防衛力改成 100
// （收據 `fcmd-turn-capture` 同一個 poke），一打就換手。
//
// 第三回合起 remake 的叛亂亂數與原版分岔（規格 054 已知差異），這一刻亞斯提的擁有者 remake 是無主、原版是
// 同盟——框裡的星域名用擁有者色。實行前改擁有者會讓 remake 走進另一條分岔的回合，所以正常路徑先確認停在
// 這一則，比框時才把亞斯提的 `+0x12` 照原版進 `20B4:1497` 那一刻的傾印（`fcmd-assault-ast/mem-in.bin`，
// 同一條輸入）改回來、重組這一則（規格 014）。其餘地圖與左下角是不同的局面，只比訊息框與右側標題。
func TestTurnReportCaptureMatchesOracle(t *testing.T) {
	mem, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-assault-ast/mem-in.bin"))
	if err != nil || len(mem) != 1<<20 {
		t.Skip("缺收據 fcmd-assault-ast 的 mem-in.bin")
	}
	g := reachOccupyAstarte(t)
	for turn := 0; turn < 4; turn++ {
		executeTurn(t, g)
	}
	start, _ := g.res.Planets.Range(16)
	binary.LittleEndian.PutUint16(g.res.Planets.Record(start)[0x14:], 100)
	pressExecute(g)
	clickThroughUntil(t, g, StateTurnReport)
	if len(g.turnRep.pending) != 1 || g.res.Planets.Owner(start) != 0 {
		t.Fatalf("排隊 %d 則、巴溫擁有者 %d；應該是帝國拿下巴溫一則", len(g.turnRep.pending), g.res.Planets.Owner(start))
	}
	g.regionRecord(16)[regionOwnerOff] = mem[0x5F2CC+16*rules.RegionRecord+regionOwnerOff]
	g.turnRep.pending[0] = g.captureReport(&g.fleets[0], 16, 0)
	matchesBox(t, g, "fcmd-turn-capture", 320, 160, 639, 319)
	matchesBox(t, g, "fcmd-turn-capture", 320, 40, 639, 119)
	tickTurnWindow(g, 89)
	if g.State() != StateTurnReport {
		t.Fatalf("89 tick 之後在 %v，原版等滿 90", g.State())
	}
	tickTurnWindow(g, 1)
	if g.State() == StateTurnReport {
		t.Fatal("90 tick 之後訊息框還開著")
	}
}

// 士氣值低落（`1261:0138`，回合尾巴 `1261:00C8`）：費沙那條第三回合打完 → 印在易主報告之後。
//
// remake 走到這裡士氣是 48、原版 47：同一個算式（`TestStrategicAssaultMatchesOracle` 以原版傾印驅動是 47），
// 差在第三回合起分岔的亂數（規格 054 已知差異）。報告的字由那一刻的 `+0x3C` 組，所以正常路徑先確認
// 會停在這一則、旗標清掉，比框時把士氣改成原版的 47 重組這一則（`fleet0.bin` 的 `+0x3C`）。
func TestTurnReportMoraleLowMatchesOracle(t *testing.T) {
	g := reachFleetOccupyPick(t)
	g.Click(163, 136) // 費沙
	g.Click(300, 240)
	g.Click(300, 240)
	g.Click(432, 263) // 決定
	executeTurn(t, g)
	executeTurn(t, g)
	pressExecute(g)
	clickThroughUntil(t, g, StateTurnReport)
	fleet0, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fcmd-turn-morale-low/fleet0.bin"))
	if err != nil {
		t.Skip("缺收據 fcmd-turn-morale-low")
	}
	f := &g.fleets[0]
	if f.Flags&rules.FleetFlagLowMorale == 0 || len(g.turnRep.pending) != 1 {
		t.Fatalf("艦隊 +0x00 = %02X、排隊 %d 則；應該是士氣低落一則", f.Flags, len(g.turnRep.pending))
	}
	f.Morale = int(fleet0[0x3C])
	f.Flags |= rules.FleetFlagMoraleFlip
	g.turnRep.pending = g.moraleReports(f)
	matchesBox(t, g, "fcmd-turn-morale-low", 320, 160, 639, 319)
	matchesBox(t, g, "fcmd-turn-morale-low", 320, 40, 639, 119)
	if f := g.fleets[0]; f.Flags&(rules.FleetFlagMoraleFlip|rules.FleetFlagNoShips) != 0 {
		t.Fatalf("報告印完艦隊 +0x00 = %02X，`1261:0104` 會清掉 bit2／bit3", f.Flags)
	}
}

// 電腦座位拿下行星（`fcmd-turn-ai-capture`：費沙那條第三回合，同盟艦隊在馬爾・亞迪特拿下威爾勃恩）。
// remake 走到第三回合時電腦的行軍已經與原版分岔，碰不到同一場，所以這一則是**直接進入的診斷比對**：
// 同盟的艦隊、馬爾・亞迪特（24）的擁有者照收據的星域表、威爾勃恩標成同盟，組 `20B4:1922` 的電腦那一路
// （「同盟軍艦隊／企圖占領…」＋帝國情報參謀頭像），只比訊息框。
func TestTurnReportAICaptureLayout(t *testing.T) {
	reg := receiptPeek(t, "fcmd-turn-ai-capture", "lin:5F2CC")
	g := reachFleetDecided(t)
	// 電腦那一路不印編號與提督，只看陣營：拿第１艦隊改成同盟就夠。
	f := g.fleets[0]
	f.Faction = 1
	g.regionRecord(24)[regionOwnerOff] = reg[24*rules.RegionRecord+regionOwnerOff]
	start, n := g.res.Planets.Range(24)
	pi := -1
	for i := 0; i < n; i++ {
		if string(planetNameBytes(g.res.Planets.Record(start+i))) == "\xab\xc2\xba\xb8\xab\x6b\xae\xa6" {
			pi = i
		}
	}
	if pi < 0 {
		t.Fatal("馬爾・亞迪特找不到威爾勃恩")
	}
	g.res.Planets.Record(start + pi)[0x11] = 1
	g.startTurnReports([]turnReport{g.captureReport(&f, 24, pi)}, nil)
	matchesBox(t, g, "fcmd-turn-ai-capture", 320, 160, 639, 319)
}

// 已全軍覆沒（`1261:0267`）：占領亞斯提那條路第五回合實行前把第 0 組艘數改成 0（收據 `fcmd-turn-no-ships`
// 同一個 poke），走完一格之後 `113A:03C5` → `1261:0067(f, 1)` 設 bit3 → 印在過場視窗上面。艦隊留著。
// 只比報告框與標題（第三回合起的分岔，理由同上）。
func TestTurnReportNoShipsMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	for turn := 0; turn < 4; turn++ {
		executeTurn(t, g)
	}
	if len(g.fleets[0].Composition) != 1 {
		t.Fatalf("第１艦隊 %d 組；收據只改了第 0 組", len(g.fleets[0].Composition))
	}
	g.fleets[0].Composition[0].Ships = 0
	pressExecute(g)
	clickThroughUntil(t, g, StateTurnReport)
	matchesBox(t, g, "fcmd-turn-no-ships", 320, 160, 639, 319)
	matchesBox(t, g, "fcmd-turn-no-ships", 320, 40, 639, 119)
	g.Click(560, 125)
	if len(g.fleets) == 0 || g.fleets[0].Faction != 0 || g.fleets[0].Flags&rules.FleetFlagNoShips != 0 {
		t.Fatal("「已全軍覆沒」之後艦隊應該留著、bit3 清掉（原版 `0374:06BC`／`297B:034A` 零次）")
	}
}
