package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 叛亂的同狀態對拍（`docs/spec/041`）：remake 從新遊戲走同一條玩家路徑、
// 在同一個時點把帝國情報專任參謀（提督 13）的營運改成同一個值，按一次
// 「實行」，叛亂的星域要與原版收據逐格相同。
//
// 原版那一端的輸入序列是 `flagpass`（新遊戲 → 系統指令 → 實行），
// 唯一的注入是 `-poke lin:60426@211000000=<營運>`（`docs/re/292`）：
// 提督表 `ds:B77C ＋ 13 × 28 ＋ 0x0E`，在第一個 AI 回合之前。
//
// `rules` 那一層的測試拿收據自己的傾印重播；這一層驗的是 **remake 自己
// 走到那一刻的狀態**——擁有者、未安定度、開局快照、兩位情報參謀、
// 艦隊位置——與原版相同。與亂數無關的那一份（`rbx292d`）逐格比結果；
// 與亂數有關的那一份（`rbx292e`）比進場種子、逐格比 r1 ＋ r2 與叛亂集合
// （見 `TestRebellionSameStateRandom`）。
const intelStaffEmpire = 13 // 帝國的情報專任參謀（`docs/internals/30`）

// dismissAllRebellionReports 讓既有「實行後」收據測試走到原本的比較時點。
// 報告本身的逐則與左右鍵行為另由專門測試釘住。
func dismissAllRebellionReports(g *Game) {
	for g.State() == StateRebellionReport && len(g.pendingRebellionReports) > 0 {
		g.Click(2, 470)
	}
}

func prepareRebellionTurn(t *testing.T, ops byte) *Game {
	t.Helper()
	g := reachBrainPhase(t)
	// 進軍路線表：正式入口（`internal/bootstrap`）會載，`reachBrainPhase`
	// 沒有。少了它 AI 的階段四／五整段不跑，亂數的次數與正式遊戲不同
	// （6 次；載了是 7 次、與原版相同，見 `TestRebellionSameStateRandom`）。
	// 這裡照正式入口載。
	routes, err := ai.LoadAdvanceRoutes(filepath.Join(root, "docs/internals/data/ai-advance-routes.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.AdvanceRoutes = routes
	g.SetBrain(g.OriginalBrain())
	if g.brain == nil {
		t.Fatal("這個入口應該接上原版戰略 AI")
	}
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, 0); got != intelStaffEmpire {
		t.Fatalf("帝國的情報專任參謀是提督 %d，開局應該是 13", got)
	}
	g.mutableAdmirals()[intelStaffEmpire*rules.AdmiralRecord+0x0E] = ops
	return g
}

func runRebellionTurn(t *testing.T, ops byte) *Game {
	t.Helper()
	g := prepareRebellionTurn(t, ops)
	g.Click(544, 125) // 系統指令
	g.Click(560, 125) // 實行
	if len(g.rebellions) > 0 {
		if g.State() != StateRebellionReport {
			t.Fatalf("叛亂後在 %v，預期先進 StateRebellionReport", g.State())
		}
		dismissAllRebellionReports(g)
	}
	if g.State() != StateCommandPhase {
		t.Fatalf("實行之後在 %v", g.State())
	}
	return g
}

// 從主選單／防拷／月初報告走到「實行」，停在原版 rbx292 的第一則叛亂
// 報告；產品 Frame 的文字與頭像區都要逐像素相同（`docs/spec/046`）。
func TestRebellionReportPlayerPathMatchesOriginal(t *testing.T) {
	want, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "rbx292", "screen.bin"))
	if err != nil || len(want) != 640*480 {
		t.Skip("缺收據 rbx292，跳過（不用代用品）")
	}
	g := prepareRebellionTurn(t, 1)
	g.Click(544, 125)
	g.Click(560, 125)
	if g.State() != StateRebellionReport {
		t.Fatalf("實行後在 %v，預期 StateRebellionReport", g.State())
	}
	if r, ok := g.CurrentRebellionReport(); !ok || r.Region != 1 || r.Owner != 0 {
		t.Fatalf("第一則報告 ＝ %+v,%v，原版 rbx292 是帝國的星域 1", r, ok)
	}
	got := g.Frame().Pix
	bad := 0
	for _, box := range [][4]int{{334, 178, 510, 220}, {544, 180, 616, 273}} {
		for y := box[1]; y < box[3]; y++ {
			for x := box[0]; x < box[2]; x++ {
				if got[y*640+x] != want[y*640+x] {
					bad++
				}
			}
		}
	}
	if bad != 0 {
		t.Errorf("正常玩家路徑的叛亂文字＋頭像與 rbx292 差 %d 個像素，預期 0", bad)
	}
}

// TestRebellionSameStateRandom 對收據 `rbx292e`（營運 4，門檻 225–625，
// 結果由亂數決定）。
//
// 驗得到的：
//
//   - 亂數：第一個 AI 回合抽 7 次（建三支艦隊各 2 次 ＋ 階段四伊謝爾倫那一格
//     1 次，`docs/re/209`），叛亂那一圈進場的種子是 AF5AAD71、整圈之後是
//     936A823B（收據 `rbx297w`，`docs/re/297`）。階段四要先過「從首都走得
//     到」（`ai.CapitalReach`，`docs/spec/043`）才只剩伊謝爾倫一格——當成
//     全 1 的時候這裡是 16 次。叛亂這一支不補抽、不遷就（`docs/spec/041`）。
//   - 32 格裡哪幾格擲亂數與原版逐格相同；帝國那 9 格的門檻與 r1 ＋ r2 逐格
//     相同（`docs/re/292` 那張表）；叛亂的星域與收據 `rbx292e` 的
//     `2703:1908` 七次相同。
//   - 19 格的門檻逐格相同，**同盟那 10 格也是**：它們的未安定度原版在 AI
//     回合裡重算過（`docs/spec/042`），remake 接上之後才直接對得上。先前這裡
//     改驗「代進原版的未安定度就得到原版的門檻」，那是上游還沒修時的繞道。
func TestRebellionSameStateRandom(t *testing.T) {
	g := runRebellionTurn(t, 4)

	// 亂數：往回數出 remake 在叛亂之前抽了幾次。叛亂那一圈 19 格 × 10 次
	// （收據 `rbx297w` 量到的，`docs/re/297`）。
	const rebelDraws = 19 * rules.RebellionDraws
	end := g.rand94DC().Seed()
	r := rules.NewRand(strategyRandSeed)
	n := 0
	for ; n < 10000 && r.Seed() != end; n++ {
		r.Next()
	}
	if before := n - rebelDraws; before != 7 {
		t.Fatalf("叛亂之前 remake 抽了 %d 次亂數，原版 7 次（`docs/re/209`）", before)
	}
	entry := rules.NewRand(strategyRandSeed)
	entry.Advance(7)
	if entry.Seed() != 0xAF5AAD71 {
		t.Errorf("叛亂那一圈進場的種子 %08X，原版 AF5AAD71（`docs/re/297`）", entry.Seed())
	}
	if end != 0x936A823B {
		t.Errorf("整圈之後的種子 %08X，收據 rbx297w 是 936A823B", end)
	}
	// 帝國那 9 格的 r1 ＋ r2（`docs/re/292`，`rbx292e_pre.bin` 重播）與叛亂的
	// 星域（收據 `rbx292e` 的 `2703:1908` 呼叫 7 次）。
	wantRoll := map[int]int{1: 297, 2: 157, 4: 112, 5: 748, 6: 262, 7: 347, 8: 407, 9: 427, 13: 201}
	for i, v := range wantRoll {
		if got := g.rebellionRolls[i].Roll; got != v {
			t.Errorf("星域 %d：r1 ＋ r2 ＝ %d，原版 %d", i, got, v)
		}
	}
	wantRebels := map[int]bool{1: true, 2: true, 4: true, 6: true, 7: true, 9: true, 13: true}
	for i, roll := range g.rebellionRolls {
		if roll.Rebels != wantRebels[i] {
			t.Errorf("星域 %d：叛亂 %v，收據 %v（門檻 %d、r1 ＋ r2 %d）",
				i, roll.Rebels, wantRebels[i], roll.Threshold, roll.Roll)
		}
	}

	// 32 格擲不擲與原版逐格相同（`rbx292e_pre.bin` 重播；20／30 兩格有剛
	// 移進去的自家艦隊，`2703:0853` 重數之後不擲）。
	want := map[int]int{
		1: 300, 2: 225, 4: 250, 5: 250, 6: 400, 7: 350, 8: 350, 9: 625, 13: 400,
		15: 19, 16: 11, 17: 9, 18: 9, 19: 8, 21: 3, 22: 10, 26: 10, 27: 9, 29: 10,
	}
	mem := readWorkplace(t, "rbx292e_pre.bin")
	orig := func(i, off int) int { return int(mem[0x54B30+0xA79C+i*rules.RegionRecord+off]) }
	for i, roll := range g.rebellionRolls {
		thr, asked := want[i]
		if roll.Asked != asked {
			t.Errorf("星域 %d：問了 %v，原版 %v", i, roll.Asked, asked)
			continue
		}
		if !asked {
			continue
		}
		// 擁有者看判定當下的（回合結束時叛亂的格子已經變成無主）。
		owner := orig(i, 0x12)
		// 兩個陣營都逐格比門檻。同盟那 10 格的未安定度要靠 AI 回合裡的
		// 重算（`docs/spec/042`）才對得上——拿掉重算，這裡會有 9 格不同。
		if roll.Threshold != thr {
			t.Errorf("星域 %d（陣營 %d）：門檻 %d，原版 %d", i, owner, roll.Threshold, thr)
		}
	}
	// 同盟的情報參謀在 AI 回合裡換成提督 75（營運 128）。
	if got := rules.StaffAdmiral(g.workRecord(), rules.PostIntel, 1); got != 75 {
		t.Errorf("同盟的情報專任參謀是提督 %d，收據是 75", got)
	}
}

// TestRebellionSameStateEffect 對收據 `rbx292d`（營運 1，九格叛亂、與亂數
// 無關）：被翻的九顆行星、81 顆的擁有者、回合結束後 32 格的擁有者都要與
// 原版逐一相同。
func TestRebellionSameStateEffect(t *testing.T) {
	plaPost := readWorkplace(t, "rbx292d_pla_post.bin")
	regPost := readWorkplace(t, "rbx292d_reg_post.bin")
	g := runRebellionTurn(t, 1)
	want := []int{1, 2, 4, 5, 6, 7, 8, 9, 13}
	if got := g.Rebellions(); !equalInts(got, want) {
		t.Fatalf("remake 叛亂的星域 %v，原版是 %v", got, want)
	}
	if len(plaPost) != 81*rules.PlanetRecord {
		t.Fatalf("行星表切片 %d bytes", len(plaPost))
	}
	flipped := map[int]bool{6: true, 8: true, 12: true, 14: true, 17: true,
		20: true, 24: true, 27: true, 33: true} // `docs/re/292` 那張表
	for i := 0; i < 81; i++ {
		got := g.res.Planets.Record(i)
		exp := plaPost[i*rules.PlanetRecord : (i+1)*rules.PlanetRecord]
		if got[0x11] != exp[0x11] {
			t.Errorf("行星 %d 的擁有者 %d，原版 %d", i, got[0x11], exp[0x11])
		}
		if !flipped[i] {
			continue
		}
		ge, ee := binary.LittleEndian.Uint32(got[0x1A:]), binary.LittleEndian.Uint32(exp[0x1A:])
		gd, ed := binary.LittleEndian.Uint16(got[0x14:]), binary.LittleEndian.Uint16(exp[0x14:])
		if ge != ee || gd != ed {
			t.Errorf("行星 %d：經濟力 %d、防衛力 %d，原版 %d／%d", i, ge, gd, ee, ed)
		}
	}
	for r := 0; r < rules.RegionCount; r++ {
		if got, exp := g.regionRecord(r)[0x12], regPost[r*rules.RegionRecord+0x12]; got != exp {
			t.Errorf("星域 %d 回合結束的擁有者 %d，原版 %d", r, got, exp)
		}
	}
	// 翻成無主的格子**不跳易主報告**：`+0x12` 在效果裡就寫成 2，回合尾巴
	// 依行星重算出來也是 2，舊值新值相同。
	if n := len(g.PendingFlips()); n != 0 {
		t.Errorf("回合結束有 %d 則易主報告，原版叛亂之後不跳", n)
	}
}

// TestRebellionOpeningSnapshot 釘 `ds:B75C`：remake 的開局快照與原版新遊戲
// 那一份逐格相同（收據 `rbx292e_pre.bin`，`docs/re/297`）。
func TestRebellionOpeningSnapshot(t *testing.T) {
	mem := readWorkplace(t, "rbx292e_pre.bin")
	g := reachBrainPhase(t)
	want := mem[0x54B30+0xB75C : 0x54B30+0xB75C+rules.RegionCount]
	if got := g.openingOwners(); !equalBytes(got, want) {
		t.Errorf("開局快照 % x，原版 ds:B75C 是 % x", got, want)
	}

	// 讀存檔不重寫它：S3 載進來之後（回合 124，擁有者早就不一樣了）原版的
	// `ds:B75C` 仍然是開局那一份（`s3datamem.bin`，另 24 份 S3 傾印相同，
	// `docs/re/297`）。remake 載同一格存檔之後也要還是那一份。
	s3 := readWorkplace(t, "s3datamem.bin")
	s3want := s3[0x54B30+0xB75C : 0x54B30+0xB75C+rules.RegionCount]
	if !equalBytes(s3want, want) {
		t.Fatalf("s3datamem 的 ds:B75C % x 與新遊戲那一份不同", s3want)
	}
	orig := filepath.Join(root, "org_game/extracted")
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(orig, name))
		if err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", name)
		}
		return b
	}
	sl, err := rules.LoadSlot(rules.SlotFiles{
		Stellar: read("STELLAR.DTA"), Planet: read("PLANET.DTA"),
		Admiral: read("ADMIRAL.DTA"), Fleet: read("FLEET.DTA"),
		Ship: read("SHIP.DTA"), Work: read("WORK.DTA"),
		PlanetHeader: read("PLADAT.DAT"),
	})
	if err != nil {
		t.Fatal(err)
	}
	g.LoadSlot(sl)
	if got := g.openingOwners(); !equalBytes(got, s3want) {
		t.Errorf("讀存檔之後的開局快照 % x，原版仍然是 % x", got, s3want)
	}
	if lost := rules.LostRegions(g.res.Stellar, g.openingOwners()); lost == [2]int{} {
		t.Errorf("讀進 DTA 之後兩邊的已失格數都是 0，快照看起來被換成了存檔那一份")
	}
}

func readWorkplace(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join(root, "workplace", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	return b
}

// 原版 `056A:06F3` 每格一則，報告未關掉前不准進戰鬥掃描；滑鼠任一鍵
// 只收掉目前一則（收據 rbx292/rbx292d，`docs/spec/046`）。
func TestRebellionReportsQueueAndResume(t *testing.T) {
	g := &Game{
		res: &Resources{FlipReport: &screen.FlipReportPanel{}},
		rebellionReports: []RebellionReport{
			{Region: 1, Owner: 0}, {Region: 4, Owner: 1},
		},
		state: StateCommandPhase,
	}
	resumed := 0
	if !g.startRebellionReports(func() { resumed++ }) {
		t.Fatal("兩格叛亂沒有開報告")
	}
	if g.state != StateRebellionReport {
		t.Fatalf("狀態 %v，預期 StateRebellionReport", g.state)
	}
	if r, ok := g.CurrentRebellionReport(); !ok || r.Region != 1 || r.Owner != 0 {
		t.Fatalf("第一則 ＝ %+v,%v，預期帝國星域 1", r, ok)
	}
	g.Click(2, 470)
	if r, ok := g.CurrentRebellionReport(); !ok || r.Region != 4 || r.Owner != 1 || resumed != 0 {
		t.Fatalf("第一下後 ＝ %+v,%v resumed=%d，預期同盟星域 4、尚未續跑", r, ok, resumed)
	}
	g.RightClick(300, 200)
	if _, ok := g.CurrentRebellionReport(); ok || resumed != 1 {
		t.Fatalf("第二下後仍有報告或續跑 %d 次，預期清空且一次", resumed)
	}
	if g.afterRebellionReports != nil {
		t.Fatal("報告清空後仍保留續跑 callback")
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
