package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
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

func runRebellionTurn(t *testing.T, ops byte) *Game {
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
	g.Click(544, 125) // 系統指令
	g.Click(560, 125) // 實行
	if g.State() != StateCommandPhase {
		t.Fatalf("實行之後在 %v", g.State())
	}
	return g
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
//   - 同盟那 10 格的未安定度 remake 沒有在 AI 回合裡重算
//     （worklist `region-unrest-recompute`），門檻改驗「代進原版的未安定度
//     就得到原版的門檻」。
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
		// 帝國那 9 格：門檻逐格相同。
		if owner == 0 {
			if roll.Threshold != thr {
				t.Errorf("星域 %d：門檻 %d，原版 %d", i, roll.Threshold, thr)
			}
			continue
		}
		// 同盟那 10 格：**未安定度是上游的落差**——原版在 AI 回合裡重算過
		// （`rbx292_a` → `rbx292_b`，同盟 13 格各降 4–5），remake 載入之後
		// 不再重算（worklist `region-unrest-recompute`）。拿原版的未安定度
		// 代進 remake 的營運與已失格數，要得到原版的門檻——也就是除了
		// 未安定度以外都一樣。
		ops := rules.StaffOps(g.workRecord(), g.res.Admiral, rules.PostIntel, owner)
		if got := rules.RebellionThreshold(orig(i, 0x7D), ops, 0); got != thr {
			t.Errorf("星域 %d：原版的未安定度 %d 代進 remake 的營運 %d 得門檻 %d，原版 %d",
				i, orig(i, 0x7D), ops, got, thr)
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
