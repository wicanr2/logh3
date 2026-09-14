package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 回合結束那條鏈（`docs/re/244`、`docs/spec/032`）：行星易主 → 星域易主 →
// 路徑表重建 → `+0x13`／`+0x24` 更新。
//
// 局面是自製的最小樣本，不是收據——收據在 `rules` 那一層
// （`TestPosturesMatchOracle`，32 格逐格相同）。這裡驗的是**接線**。
func newRegionTurnGame(t *testing.T) *Game {
	t.Helper()
	graph, err := rules.LoadStarGraph(filepath.Join("..", "..",
		"docs", "internals", "data", "star-graph.json"))
	if err != nil {
		t.Fatalf("讀星域圖：%v", err)
	}
	// 每個星域一顆行星：0–15 帝國、16–31 同盟。
	pl := &rules.Planets{
		PerRegion: make([]int, rules.RegionCount),
		Data:      make([]byte, rules.PlanetHeader+rules.RegionCount*rules.PlanetRecord),
	}
	for i := range pl.PerRegion {
		pl.PerRegion[i] = 1
		o := byte(0)
		if i >= 16 {
			o = 1
		}
		pl.Data[rules.PlanetHeader+i*rules.PlanetRecord+0x11] = o
	}
	g := &Game{res: &Resources{
		Graph:   graph,
		Planets: pl,
		Stellar: make([]byte, rules.RegionCount*rules.RegionRecord),
	}}
	for i := 0; i < rules.RegionCount; i++ {
		g.regionRecord(i)[0x11] = byte(i)
	}
	return g
}

// regionPass 跑完整條但不跳報告——測純算的那幾支時用。
func regionPass(g *Game) int {
	n := g.recomputeRegionOwners()
	g.pendingFlips = nil
	g.refreshRegionPosture()
	return n
}

func TestAdvanceTurnRefillsRegionSupply(t *testing.T) {
	g := newRegionTurnGame(t)
	g.regionRecord(0)[regionOwnerOff] = 0
	binary.LittleEndian.PutUint32(g.regionRecord(0)[0x1A:], 900)
	binary.LittleEndian.PutUint16(g.regionRecord(0)[0x20:], 100)
	g.res.Planets.SetPlanetEconomy(0, 900)
	stock := rules.PlanetStock{}
	rules.WritePlanetStock(g.res.Planets.Record(0), stock)
	g.advanceTurn()
	got, ok := rules.ReadPlanetStock(g.res.Planets.Record(0))
	if !ok || got.Missiles != 100 || got.Fighters != 100 {
		t.Fatalf("回合 1 行星回充=%+v，預期飛彈／艦載機各 100", got)
	}
}

func newPhezzanTurnGame(t *testing.T) *Game {
	t.Helper()
	g := newRegionTurnGame(t)
	per := make([]int, rules.RegionCount)
	per[0], per[1], per[10], per[14], per[16], per[31] = 1, 32, 1, 1, 45, 1
	g.res.Planets = &rules.Planets{PerRegion: per,
		Data: make([]byte, rules.PlanetHeader+81*rules.PlanetRecord)}
	for i := 0; i < 81; i++ {
		g.res.Planets.Record(i)[0x11] = rules.RegionOwnerNone
	}
	for r := 0; r < rules.RegionCount; r++ {
		g.regionRecord(r)[0x12] = rules.RegionOwnerNone
	}
	setOwner := func(region, owner int) {
		start, n := g.res.Planets.Range(region)
		for i := start; i < start+n; i++ {
			g.res.Planets.Record(i)[0x11] = byte(owner)
		}
		g.regionRecord(region)[0x12] = byte(owner)
	}
	for _, region := range []int{0, 1, 10} {
		setOwner(region, 0)
	}
	setOwner(31, 1)
	g.res.Planets.Record(34)[0x11] = 0 // 費沙剛被帝國拿下；星域仍是無主。
	g.res.Planets.SetPlanetEconomy(0, 1000)
	g.res.Planets.SetPlanetEconomy(33, 900)
	g.res.Planets.SetPlanetEconomy(34, 103100)
	g.res.Planets.SetPlanetEconomy(80, 1000)
	g.res.WorkTemplate = make([]byte, rules.WorkRecord)
	g.faction = 0
	g.turn = 1
	return g
}

func TestPhezzanEventRunsAfterFlipReportOnce(t *testing.T) {
	g := newPhezzanTurnGame(t)
	g.res.UnrestParams = rules.UnrestParams{
		Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
		Table: []int{170, 150, 130, 90, 60}, Base: []int{50, 75, 100, 150, 200},
	}
	if !g.runRegionPass(g.finishTurn) {
		t.Fatal("費沙易主應先停在易主報告")
	}
	if rules.PhezzanOccurred(g.workRecord()) {
		t.Fatal("易主報告還沒關就先跑費沙事件")
	}
	for len(g.PendingFlips()) > 0 {
		g.dismissFlip()
	}
	if !rules.PhezzanOccurred(g.workRecord()) {
		t.Fatal("易主報告關掉後沒有登記 WORK +0x34")
	}
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatalf("保存費沙旗標：%v", err)
	}
	work, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
	if err != nil || len(work) <= 0x34 || work[0x34] != 1 {
		t.Fatalf("存檔沒有保留 WORK +0x34：長度 %d、錯誤 %v", len(work), err)
	}
	if got := g.res.Planets.PlanetEconomy(34); got != 20620 {
		t.Errorf("費沙經濟力 ＝ %d，預期 20620", got)
	}
	if got := g.PhezzanRebellions(); len(got) != 3 || got[0] != 10 || got[1] != 11 || got[2] != 12 {
		t.Errorf("事件叛離星域 ＝ %v，預期 [10 11 12]", got)
	}
	// 即使有人錯誤地再登記，存檔旗標仍要讓事件失敗即關閉。
	g.phezzanPending = true
	g.runPhezzanEvent()
	if got := g.res.Planets.PlanetEconomy(34); got != 20620 {
		t.Errorf("事件重複發生：費沙經濟力又變成 %d", got)
	}
}

func TestPhezzanReportsRunBeforeEventBody(t *testing.T) {
	g := newPhezzanTurnGame(t)
	g.res.PhezzanReport = &screen.PhezzanReportPanel{}
	if !g.runRegionPass(g.finishTurn) {
		t.Fatal("費沙易主應先停在易主報告")
	}
	for len(g.PendingFlips()) > 0 {
		g.dismissFlip()
	}
	if g.state != StatePhezzanReport || rules.PhezzanOccurred(g.workRecord()) {
		t.Fatalf("第一頁狀態=%v occurred=%v", g.state, rules.PhezzanOccurred(g.workRecord()))
	}
	if r, ok := g.CurrentPhezzanReport(); !ok || r.Key() != "owner_0" || r.Page != 0 {
		t.Fatalf("第一頁 %+v,%v", r, ok)
	}
	g.Click(300, 220)
	if r, ok := g.CurrentPhezzanReport(); !ok || r.Page != 1 || rules.PhezzanOccurred(g.workRecord()) {
		t.Fatalf("第二頁 %+v,%v occurred=%v", r, ok, rules.PhezzanOccurred(g.workRecord()))
	}
	g.RightClick(300, 220)
	if !rules.PhezzanOccurred(g.workRecord()) || g.state == StatePhezzanReport {
		t.Fatalf("兩頁關完 occurred=%v state=%v", rules.PhezzanOccurred(g.workRecord()), g.state)
	}
}

func TestPhezzanReportsVersusOwnerThenOpponent(t *testing.T) {
	g := newPhezzanTurnGame(t)
	g.versus = true
	g.res.PhezzanReport = &screen.PhezzanReportPanel{}
	if !g.runRegionPass(g.finishTurn) {
		t.Fatal("費沙易主應先停在易主報告")
	}
	for len(g.PendingFlips()) > 0 {
		g.dismissFlip()
	}
	want := []PhezzanReport{
		{Owner: 0, Viewer: 0, Page: 0},
		{Owner: 0, Viewer: 0, Page: 1},
		{Owner: 0, Viewer: 1, Page: 0},
		{Owner: 0, Viewer: 1, Page: 1},
	}
	if !reflect.DeepEqual(g.pendingPhezzanReports, want) {
		t.Fatalf("費沙報告順序=%+v，預期=%+v", g.pendingPhezzanReports, want)
	}
}

func TestPhezzanEventRecomputesUnrestImmediately(t *testing.T) {
	g := newPhezzanTurnGame(t)
	g.res.UnrestParams = rules.UnrestParams{
		Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
		Table: []int{170, 150, 130, 90, 60}, Base: []int{50, 75, 100, 150, 200},
	}
	g.regionRecord(rules.PhezzanRegion)[regionOwnerOff] = 0
	g.phezzanPending = true
	for i := 0; i < rules.RegionCount; i++ {
		g.regionRecord(i)[rules.RegionUnrest] = 0xFF
	}
	g.runPhezzanEvent()
	changed := 0
	for i := 0; i < rules.RegionCount; i++ {
		if g.regionRecord(i)[rules.RegionUnrest] != 0xFF {
			changed++
		}
	}
	if changed == 0 {
		t.Fatal("費沙事件後 32 格未安定度都沒有立即重算")
	}
}

func TestPhezzanEventNeedsActualOwnerChange(t *testing.T) {
	g := newPhezzanTurnGame(t)
	g.regionRecord(rules.PhezzanRegion)[0x12] = 0
	regionPass(g)
	g.finishTurn()
	if rules.PhezzanOccurred(g.workRecord()) {
		t.Fatal("費沙已屬帝國、這回合沒換手仍觸發事件")
	}
	if got := g.res.Planets.PlanetEconomy(34); got != 103100 {
		t.Errorf("沒換手仍搬走經濟力：%d", got)
	}
}

func TestPhezzanTacticalCaptureRegistersOnlyHumanWinner(t *testing.T) {
	for _, c := range []struct {
		name, winner int
		want         bool
	}{
		{name: 0, winner: 0, want: true},
		{name: 1, winner: 1, want: false},
	} {
		t.Run([]string{"human", "computer"}[c.name], func(t *testing.T) {
			g := newPhezzanTurnGame(t)
			g.res.Planets.Record(34)[0x11] = rules.RegionOwnerNone
			g.regionRecord(rules.PhezzanRegion)[0x12] = rules.RegionOwnerNone
			g.battleRegion = rules.PhezzanRegion
			g.battlePlanets = []rules.BattlePlanet{{
				Index: 34, Owner: c.winner, Economy: 103100,
			}}
			g.EndBattle()
			for len(g.PendingFlips()) > 0 {
				g.dismissFlip()
			}
			if got := rules.PhezzanOccurred(g.workRecord()); got != c.want {
				t.Errorf("戰術戰鬥勝者 %d：費沙事件 ＝ %v，預期 %v", c.winner, got, c.want)
			}
		})
	}
}

func TestRecomputeRegionsFollowsPlanets(t *testing.T) {
	g := newRegionTurnGame(t)
	// 記錄的 `+0x12` 初值是 0，所以只有 16–31 那 16 格算易主。
	if n := regionPass(g); n != 16 {
		t.Fatalf("第一次應該 16 格易主，得到 %d", n)
	}
	if n := regionPass(g); n != 0 {
		t.Errorf("第二次不該再有易主，得到 %d 格", n)
	}
	for i := 0; i < rules.RegionCount; i++ {
		want := 0
		if i >= 16 {
			want = 1
		}
		if got := int(g.regionRecord(i)[0x12]); got != want {
			t.Fatalf("星域 %d 的擁有者 ＝ %d，預期 %d", i, got, want)
		}
	}

	// 首都那一格的成本是 0，bit0 一定沒設。
	for f, cap := range map[int]int{0: 0, 1: 31} {
		rec := g.regionRecord(cap)
		if c := binary.LittleEndian.Uint16(rec[0x24:]); c != 0 {
			t.Errorf("陣營 %d 的首都 %d 的 +0x24 ＝ %d，預期 0", f, cap, c)
		}
		if rec[0x13]&1 != 0 {
			t.Errorf("首都 %d 不該設 bit0", cap)
		}
	}

	// 把星域 20（同盟）的行星換成帝國，它應該易主，而且成本要重算——
	// 路徑表是擁有者的函數，換一格會影響到別格。
	before := binary.LittleEndian.Uint16(g.regionRecord(18)[0x24:])
	g.res.Planets.Data[rules.PlanetHeader+20*rules.PlanetRecord+0x11] = 0
	if n := regionPass(g); n != 1 {
		t.Fatalf("只有星域 20 該易主，得到 %d 格", n)
	}
	if got := int(g.regionRecord(20)[0x12]); got != 0 {
		t.Errorf("星域 20 的擁有者 ＝ %d，預期 0", got)
	}
	if after := binary.LittleEndian.Uint16(g.regionRecord(18)[0x24:]); after == before {
		t.Errorf("星域 18 的 +0x24 沒跟著重算（前後都是 %d）", before)
	}
	// 20 現在是帝國的，它的鄰居（同盟）就變成前線。
	for _, n := range g.res.Graph.Neighbours(20) {
		if g.regionRecord(n)[0x13]&8 == 0 {
			t.Errorf("星域 %d 是 20 的鄰居、擁有者不同，bit3 應該設", n)
		}
	}
}

// 兩支常式合起來管 bit0–bit4（`and 0F2h` ＋ `and 0E9h`），**bit5–bit7
// 兩邊都不動**（`docs/re/245`）。
func TestRecomputeRegionsKeepsOtherFlagBits(t *testing.T) {
	g := newRegionTurnGame(t)
	g.regionRecord(5)[0x13] = 0xFF
	regionPass(g)
	if got := g.regionRecord(5)[0x13] & 0xE0; got != 0xE0 {
		t.Errorf("bit5–bit7 被動到了：%02X", g.regionRecord(5)[0x13])
	}
}

// bit4 ＝ 單方面占據：艦隊停在不屬於它的星域、而對方沒有艦隊。
// 待在自家地盤不算（`056A:0469`，`docs/re/245`）。
func TestRecomputeRegionsMarksOccupation(t *testing.T) {
	g := newRegionTurnGame(t) // 0–15 帝國、16–31 同盟
	// 帝國一支停在同盟的星域 20 → bit4；停在自己的星域 5 → 不設。
	g.fleets = []rules.Fleet{{Faction: 0, Region: 20}, {Faction: 0, Region: 5}}
	regionPass(g)
	if got := g.regionRecord(20)[0x13]; got&0x10 == 0 {
		t.Errorf("星域 20 應該設 bit4，+0x13 ＝ %02X", got)
	}
	if got := g.regionRecord(5)[0x13]; got&0x10 != 0 {
		t.Errorf("星域 5 是自家地盤，不該設 bit4，+0x13 ＝ %02X", got)
	}
	// 雙方都在 → bit2，而且 bit4 不設。
	g.fleets = append(g.fleets, rules.Fleet{Faction: 1, Region: 20})
	regionPass(g)
	if got := g.regionRecord(20)[0x13]; got&0x04 == 0 || got&0x10 != 0 {
		t.Errorf("星域 20 應該只設 bit2，+0x13 ＝ %02X", got)
	}
	// `Flags` bit0 有設的不計（`297B:14F8` 的 `test 1`）。
	g.fleets = []rules.Fleet{{Faction: 0, Region: 20, Flags: 1}}
	regionPass(g)
	if got := g.regionRecord(20)[0x13]; got&0x14 != 0 {
		t.Errorf("bit0 有設的不該算進去，+0x13 ＝ %02X", got)
	}
}

// 打完戰鬥就重算：原版走完回合結束那條鏈（`docs/re/105` 的 `s3win`
// 收據裡 `2703:0944`／`2703:0990` 都在 `2703:0933` 內，而它只有兩個
// 呼叫者），所以 `EndBattle` 也要跑一次。
func TestEndBattleRecomputesRegions(t *testing.T) {
	g := newRegionTurnGame(t)
	regionPass(g)
	if got := int(g.regionRecord(20)[0x12]); got != 1 {
		t.Fatalf("星域 20 一開始應該是同盟，得到 %d", got)
	}

	// 打下星域 20 那顆行星：戰場副本改了，還沒寫回。
	bp := g.res.Planets.BattlePlanetsFor(20, []rules.PlanetSite{{}})
	if len(bp) != 1 {
		t.Fatalf("星域 20 配出 %d 顆行星，預期 1", len(bp))
	}
	bp[0].Owner = 0
	g.battlePlanets = bp
	if got := int(g.regionRecord(20)[0x12]); got != 1 {
		t.Fatalf("還沒 EndBattle 就變了：%d", got)
	}

	before := binary.LittleEndian.Uint16(g.regionRecord(20)[0x24:])
	g.EndBattle()
	if got := g.res.Planets.Owner(20); got != 0 {
		t.Errorf("行星沒寫回：擁有者 ＝ %d，預期 0", got)
	}
	if got := int(g.regionRecord(20)[0x12]); got != 0 {
		t.Errorf("星域 20 的擁有者 ＝ %d，預期 0", got)
	}
	// 易主要先跳報告，而且**後半段還沒跑**——原版就是卡在這裡
	// （`docs/re/246`：`2703:08DE` 零次執行）。
	if g.state != StateRegionFlipReport {
		t.Fatalf("狀態 ＝ %v，預期停在易主報告", g.state)
	}
	if f, ok := g.CurrentFlip(); !ok || f.Region != 20 || f.NewOwner != 0 {
		t.Errorf("報告的內容 ＝ %+v %v，預期星域 20 易主給帝國", f, ok)
	}
	if after := binary.LittleEndian.Uint16(g.regionRecord(20)[0x24:]); after != before {
		t.Errorf("報告還沒按掉就重算了 +0x24：%d → %d", before, after)
	}
	g.RightClick(300, 200) // 右鍵也必須走 callback（收據 chain4；左鍵是 chain5）
	if g.state != StateCommandPhase {
		t.Errorf("按掉之後狀態 ＝ %v，預期回指令畫面", g.state)
	}
	if after := binary.LittleEndian.Uint16(g.regionRecord(20)[0x24:]); after == before {
		t.Errorf("按掉之後 +0x24 沒重算（前後都是 %d）", before)
	}
}

// 報告一則一則來：兩格易主就要按兩次，而且兩次都按完才跑後半段
// （`docs/spec/034`）。
func TestRegionFlipReportsQueueOneByOne(t *testing.T) {
	g := newRegionTurnGame(t)
	regionPass(g)
	// 星域 20 與 21 各翻一顆行星。
	for _, r := range []int{20, 21} {
		g.res.Planets.Data[rules.PlanetHeader+r*rules.PlanetRecord+0x11] = 0
	}
	before := binary.LittleEndian.Uint16(g.regionRecord(18)[0x24:])
	if !g.runRegionPass(nil) {
		t.Fatal("有兩格易主，應該停在報告上")
	}
	if n := len(g.PendingFlips()); n != 2 {
		t.Fatalf("排了 %d 則報告，預期 2", n)
	}
	g.dismissFlip()
	if n := len(g.PendingFlips()); n != 1 {
		t.Errorf("按掉一則之後剩 %d，預期 1", n)
	}
	if after := binary.LittleEndian.Uint16(g.regionRecord(18)[0x24:]); after != before {
		t.Errorf("還有一則沒按掉就重算了後半段")
	}
	g.dismissFlip()
	if after := binary.LittleEndian.Uint16(g.regionRecord(18)[0x24:]); after == before {
		t.Errorf("全部按掉之後後半段沒跑（+0x24 前後都是 %d）", before)
	}
}

// 四句話怎麼挑（`docs/re/247` 的訊息表；後兩句是假設待驗）。
func TestFlipMessageSelection(t *testing.T) {
	for _, c := range []struct {
		old, now, seat int
		want           FlipMessage
	}{
		{1, 0, 0, FlipTakenByUs},      // 敵方的變成我的
		{0, 1, 0, FlipLostToEnemy},    // 我的被打下來
		{0, 2, 0, FlipOurGripLost},    // 我的變成無主
		{1, 2, 0, FlipWePartlyLanded}, // 敵方的變成無主
	} {
		f := RegionFlip{OldOwner: c.old, NewOwner: c.now}
		if got := f.Message(c.seat); got != c.want {
			t.Errorf("%d → %d（座位 %d）＝ %v，預期 %v", c.old, c.now, c.seat, got, c.want)
		}
	}
}

// 一個回合會把**每一個**交戰星域都打完，不是只打編號最小的那一格
// （原版 `0374:03E8` 逐格掃，`docs/re/251`）。
func TestBattleScanContinuesToNextRegion(t *testing.T) {
	g := newRegionTurnGame(t)
	regionPass(g)
	// 星域 5 與 20 各有雙方艦隊。
	g.fleets = []rules.Fleet{
		{Faction: 0, Region: 5}, {Faction: 1, Region: 5},
		{Faction: 0, Region: 20}, {Faction: 1, Region: 20},
	}
	g.rebellionOccupation[5].Contested = true
	g.rebellionOccupation[20].Contested = true
	if got := g.contestedRegion(); got != 5 {
		t.Fatalf("第一格 ＝ %d，預期 5", got)
	}
	if got := g.contestedRegionFrom(6); got != 20 {
		t.Fatalf("從 6 起 ＝ %d，預期 20", got)
	}
	if got := g.contestedRegionFrom(21); got != -1 {
		t.Errorf("從 21 起 ＝ %d，預期 -1", got)
	}

	// 打完星域 5 之後要接著開星域 20，不是收尾。
	g.battleRegion = 5
	g.EndBattle()
	if g.battleRegion != 20 {
		t.Errorf("打完 5 之後 battleRegion ＝ %d，預期 20", g.battleRegion)
	}
	if g.state == StateCommandPhase {
		t.Error("還有一格沒打就回指令畫面了")
	}
	// 再打完 20 就沒有了，這時才走回合結束。
	g.EndBattle()
	if g.state == StateBattle {
		t.Error("都打完了還停在戰場")
	}
}

// 較小編號的戰鬥結束後，原版先經 `090A:0E6D` 重算，再讓逐星域迴圈
// 繼續；後方叛亂效果用重算後的 `+0x7D` 決定翻幾顆（SPEC 042 READY）。
func TestBattleRecomputesUnrestBeforeLaterRebellionEffect(t *testing.T) {
	g := newRegionTurnGame(t)
	const rebelRegion = 10

	// 星域 10 放三顆行星，才能區分戰前未安定度 200（全翻）與戰後重算的
	// <= 80（只翻切片最後一顆）。其餘星域仍各一顆。
	g.res.Planets.PerRegion[rebelRegion] = 3
	total := rules.RegionCount + 2
	g.res.Planets.Data = make([]byte, rules.PlanetHeader+total*rules.PlanetRecord)
	n := 0
	for r, count := range g.res.Planets.PerRegion {
		owner := byte(0)
		if r >= 16 {
			owner = 1
		}
		g.regionRecord(r)[regionOwnerOff] = owner
		for k := 0; k < count; k++ {
			g.res.Planets.Record(n)[0x11] = owner
			n++
		}
	}

	g.res.WorkTemplate = make([]byte, rules.WorkRecord)
	rules.SetStaffAdmiral(g.res.WorkTemplate, rules.PostDomestic, 0, 0)
	rules.SetStaffAdmiral(g.res.WorkTemplate, rules.PostDomestic, 1, 0)
	g.res.Admiral = make([]byte, 82*rules.AdmiralRecord)
	g.res.Admiral[0x0E] = 100
	g.res.UnrestParams = rules.UnrestParams{
		Ruler: [2]int{0, 0}, Blend: [2]int{0, 0}, Rate: [2]int{0, 0},
		Table: []int{170}, Base: []int{50},
	}
	g.turn = 1
	g.battleRegion = 5
	g.regionRecord(rebelRegion)[rules.RegionUnrest] = 200
	g.rebellionRolls[rebelRegion].Rebels = true

	g.EndBattle()

	start, count := g.res.Planets.Range(rebelRegion)
	neutral := 0
	for i := start; i < start+count; i++ {
		if g.res.Planets.Owner(i) == rules.RegionOwnerNone {
			neutral++
		}
	}
	if neutral != 1 {
		t.Fatalf("戰後重算後翻了 %d 顆，預期未安定度 <= 80 只翻 1 顆", neutral)
	}
}

// TestTaxEstimateRefreshesEveryTurn 釘住「預估徵稅額每個回合重算」
// （`docs/spec/035`、`docs/re/255`）。
//
// 原版 `0654:0034` 的共同尾巴 `090A:0E6D` 每個座位回合都跑，而且是冪等的
// ——收據 `monthgate7`／`monthgateB` 各跑三趟，20 筆逐筆累加值完全相同。
// 所以中途打下一個星域之後，當回合的預估值就要跟著變，不必等到月底。
func TestTaxEstimateRefreshesEveryTurn(t *testing.T) {
	g := newRegionTurnGame(t)
	// 每格經濟力 1000、稅率 32‰ → 每格 32。
	for i := 0; i < rules.RegionCount; i++ {
		rec := g.regionRecord(i)
		binary.LittleEndian.PutUint32(rec[0x1A:], 1000)
		rec[0x26] = 32
	}
	regionPass(g)
	g.refreshTaxEstimate()
	before := g.taxTotal
	if before[0] != 16*32 || before[1] != 16*32 {
		t.Fatalf("初值 ＝ %v，預期兩邊各 %d", before, 16*32)
	}

	// 打下一顆行星讓星域 16 從同盟換到帝國，然後跑一個**非月底**的回合。
	g.res.Planets.Data[rules.PlanetHeader+16*rules.PlanetRecord+0x11] = 0
	g.turn = 3 // 3 mod 10 != 0
	g.recomputeRegionOwners()
	g.pendingFlips = nil
	g.finishTurn()

	if got, want := g.taxTotal[0], 17*32; got != want {
		t.Errorf("帝國的預估徵稅額 ＝ %d，預期 %d", got, want)
	}
	if got, want := g.taxTotal[1], 15*32; got != want {
		t.Errorf("同盟的預估徵稅額 ＝ %d，預期 %d", got, want)
	}
}

// TestUpkeepAndMoveCostMatchSave 拿 `WORK.DTA` 記的 `ds:8F72`／`ds:8F82`
// 對 remake 算出來的（`docs/re/263`）。
//
// 這一支同時驗三件本來各自壞掉的事：
//   - `Slot.FleetComposition` 從艦艇池把編制重建回來（本來載進來是空的，
//     於是維持費與移動經費一律算 0）；
//   - `Slot.Fleets` 填 `Operations`（艦隊 `+0x47`），少了它
//     `(200 − 營運)` 那一項會讓兩筆都多算一倍；
//   - 兩者的**集合不同**：維持費算那一方每一支，移動經費只算會動的。
func TestUpkeepAndMoveCostMatchSave(t *testing.T) {
	g := newTestGame(t)
	if g.slot == nil {
		t.Skip("沒有存檔，跳過")
	}
	kinds, err := rules.LoadShipKindTable(filepath.Join(root,
		"docs", "internals", "data", "ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.ShipKinds = kinds
	for f := 0; f < 2; f++ {
		if got, want := g.fleetUpkeep(f), g.slot.Upkeep[f]; got != want {
			t.Errorf("陣營 %d 的維持費 ＝ %d，WORK.DTA 是 %d", f, got, want)
		}
		if got, want := g.fleetMoveCost(f), g.slot.MoveCost[f]; got != want {
			t.Errorf("陣營 %d 的移動經費 ＝ %d，WORK.DTA 是 %d", f, got, want)
		}
	}
}
