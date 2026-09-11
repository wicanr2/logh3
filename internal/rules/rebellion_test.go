package rules

import (
	"bufio"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestRebellionThresholdLadder 驗四段階梯（`docs/re/292` 那張表：營運 100、
// 沒失過格）與兩個除數。
func TestRebellionThresholdLadder(t *testing.T) {
	for _, c := range []struct{ u, want int }{
		{0, 0}, {49, 16}, {50, 25}, {99, 49}, {100, 75},
		{149, 111}, {150, 200}, {200, 266}, {255, 340},
	} {
		if got := RebellionThreshold(c.u, 100, 0); got != c.want {
			t.Errorf("未安定度 %d：門檻 %d，要 %d", c.u, got, c.want)
		}
	}
	// 營運是分母：營運 4 → ×25；已失兩格 → ÷3。
	if got := RebellionThreshold(38, 4, 0); got != 300 {
		t.Errorf("未安定度 38、營運 4：門檻 %d，收據 rbx292e 的星域 1 是 300", got)
	}
	if got := RebellionThreshold(38, 4, 2); got != 100 {
		t.Errorf("已失兩格：門檻 %d，要 300 ÷ 3 ＝ 100", got)
	}
	if got := RebellionThreshold(200, 0, 0); got != 0 {
		t.Errorf("營運 0：門檻 %d，remake 當成 0（原版會除以零）", got)
	}
}

// TestRebelPlanetsKeepsTheUnusedSort 釘住「排了沒用的 qsort」：翻的是切片的
// 第 0 顆或最後一顆，與經濟力無關（`docs/re/292`）。
func TestRebelPlanetsKeepsTheUnusedSort(t *testing.T) {
	for _, c := range []struct {
		unrest, count int
		all           bool
		want          []int
	}{
		{27, 3, false, []int{2}},        // ≤ 80：最後一顆
		{80, 4, false, []int{3}},        // 80 還是最後一顆（原版 `ja`）
		{81, 4, false, []int{0}},        // > 80：第 0 顆
		{120, 4, false, []int{0}},       // 120 還是一顆（原版 `jbe`）
		{121, 4, false, []int{0, 1}},    // > 120：前兩顆
		{121, 1, false, []int{0}},       // 只有一顆就一顆
		{160, 3, false, []int{0, 1}},    // 160 還是兩顆
		{161, 3, false, []int{0, 1, 2}}, // > 160：整格
		{10, 5, true, []int{0, 1, 2, 3, 4}},
	} {
		got := RebelPlanets(c.unrest, c.count, c.all)
		if len(got) != len(c.want) {
			t.Errorf("未安定度 %d、%d 顆、全部 %v：%v，要 %v", c.unrest, c.count, c.all, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("未安定度 %d、%d 顆、全部 %v：%v，要 %v", c.unrest, c.count, c.all, got, c.want)
				break
			}
		}
	}
}

// TestRebelPlanetArithmetic 驗 `2703:1ADB` 的兩條分支，數字取自收據
// `rbx292d` 的行星 6（4299→2866、4400→2934／4400）。
func TestRebelPlanetArithmetic(t *testing.T) {
	p := make([]byte, PlanetRecord)
	put := func(eco, def, limit int) {
		p[planetOwner] = 0
		binary.LittleEndian.PutUint32(p[planetEconomy:], uint32(eco))
		binary.LittleEndian.PutUint16(p[planetDefence:], uint16(def))
		binary.LittleEndian.PutUint16(p[planetGrowthDef:], uint16(limit))
	}
	get := func() (owner, eco, def int) {
		return int(p[planetOwner]), int(binary.LittleEndian.Uint32(p[planetEconomy:])),
			int(binary.LittleEndian.Uint16(p[planetDefence:]))
	}
	put(4299, 4400, 4400)
	RebelPlanet(p)
	if o, e, d := get(); o != 2 || e != 2866 || d != 2934 {
		t.Errorf("行星 6：擁有者 %d、經濟力 %d、防衛力 %d，收據是 2／2866／2934", o, e, d)
	}
	// 防衛力 × 3 < 上限 → 歸零。
	put(9, 1000, 3001)
	RebelPlanet(p)
	if _, e, d := get(); e != 6 || d != 0 {
		t.Errorf("扣不夠：經濟力 %d、防衛力 %d，要 6／0", e, d)
	}
	// 剛好等於：不歸零，扣上限的三分之一。
	put(0, 1000, 3000)
	RebelPlanet(p)
	if _, _, d := get(); d != 0 {
		t.Errorf("1000 × 3 ＝ 3000 不小於上限：防衛力 %d，要 1000 − 1000 ＝ 0", d)
	}
}

// TestDecideRebellionsDrawsTenPerAskedRegion 驗「一格固定抽十次」與閘門：
// 過了三道閘門的格子各吃十個亂數，沒過的一個都不吃。
func TestDecideRebellionsDrawsTenPerAskedRegion(t *testing.T) {
	recs := make([]byte, RegionCount*RegionRecord)
	opening := make([]byte, RegionCount)
	for i := 0; i < RegionCount; i++ {
		recs[i*RegionRecord+regionOwner] = 2
		opening[i] = 2
	}
	set := func(i, owner, unrest int) {
		recs[i*RegionRecord+regionOwner] = byte(owner)
		recs[i*RegionRecord+RegionUnrest] = byte(unrest)
		opening[i] = byte(owner)
	}
	set(0, 0, 200)  // 首都：不問
	set(3, 0, 200)  // 問
	set(5, 1, 200)  // 有自己的艦隊：不問
	set(7, 1, 200)  // 被占據：不問
	set(9, 1, 200)  // 問
	set(31, 1, 200) // 首都：不問
	in := RebellionInput{Records: recs, Opening: opening, Ops: [2]int{1, 1},
		Rand: NewRand(0x12345678)}
	in.Garrison[5][1] = 1
	in.Occupation[7].Occupied = true
	out := DecideRebellions(in)

	want := NewRand(0x12345678)
	want.Advance(2 * RebellionDraws)
	if in.Rand.Seed() != want.Seed() {
		t.Errorf("兩格要問：種子走到 %08X，要走 20 次的 %08X", in.Rand.Seed(), want.Seed())
	}
	for i, r := range out {
		asked := i == 3 || i == 9
		if r.Asked != asked {
			t.Errorf("星域 %d：問了 %v，要 %v", i, r.Asked, asked)
		}
		// 營運 1、未安定度 200 → 門檻 26600，一定叛。
		if r.Rebels != asked {
			t.Errorf("星域 %d：叛亂 %v，要 %v", i, r.Rebels, asked)
		}
	}
}

// ---- 收據 ----
//
// `rbx292e`（`docs/re/292`）：新遊戲、帝國的情報專任參謀（提督 13）營運
// poke 成 4，按一次「實行」。帝國那九格的門檻落在 225–625，**哪幾格叛亂
// 完全由亂數決定**。`rbx292e_pre.bin` 是 `#218,300,000` 的 1 MB 傾印
// （種子 AF5AAD71），在 `2703:0853` 重數艦隊之前（`docs/re/297`）。
//
// `rbx292d`：同一條序列、營運 poke 成 1（門檻 900–2500，與亂數無關），
// 報告有補點擊讓逐格迴圈跑完，另外倒了前後兩份行星表與星域表。
// 判定用的狀態取 `rbx292_c.bin`（`rbx292` 那一次的 `#218,000,000`；
// 兩次到 `#218,600,000` 為止輸入完全相同）。

const (
	dumpDS       = 0x54B30
	dumpRegions  = dumpDS + 0xA79C
	dumpFleets   = dumpDS + 0xC074
	dumpOpening  = dumpDS + 0xB75C
	dumpAdmirals = dumpDS + 0xB77C
	dumpStaff    = dumpDS + 0x8F4D // ds:8F4D[陣營]：情報專任參謀
	dumpSeed     = dumpDS + 0x94DC
)

func readReceipt(t *testing.T, parts ...string) []byte {
	t.Helper()
	p := filepath.Join(append([]string{"..", "..", "workplace"}, parts...)...)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	return b
}

// rebellionInputFromDump 從一份 1 MB 執行期傾印湊出 `056A:0469` 那一刻的
// 判定輸入。艦隊數照 `2703:0853` 自己重數（傾印在它之前）。
func rebellionInputFromDump(t *testing.T, mem []byte) RebellionInput {
	t.Helper()
	if len(mem) != 1<<20 {
		t.Fatalf("傾印 %d bytes，要 1 MB", len(mem))
	}
	recs := append([]byte(nil), mem[dumpRegions:dumpRegions+RegionCount*RegionRecord]...)
	sl := &Slot{Fleet: mem[dumpFleets : dumpFleets+FleetCount*FleetRecord]}
	fleets := sl.Fleets(StrengthWeights{})
	owner := make([]int, RegionCount)
	for i := range owner {
		owner[i] = int(recs[i*RegionRecord+regionOwner])
	}
	g, err := LoadStarGraph(filepath.Join("..", "..", "docs", "internals", "data", "star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	in := RebellionInput{
		Records:    recs,
		Opening:    append([]byte(nil), mem[dumpOpening:dumpOpening+RegionCount]...),
		Occupation: g.Occupations(owner, fleets),
		Rand:       NewRand(binary.LittleEndian.Uint32(mem[dumpSeed:])),
	}
	for r := 0; r < RegionCount; r++ {
		for f := 0; f < 2; f++ {
			in.Garrison[r][f] = CountFleets(fleets, r, f)
		}
	}
	for f := 0; f < 2; f++ {
		in.Ops[f] = AdmiralOps(mem[dumpAdmirals:dumpAdmirals+82*AdmiralRecord], int(mem[dumpStaff+f]))
	}
	return in
}

// rebelRegionsInReport 從收據的 `report.txt` 讀 `2703:1908` 每一次呼叫的
// 星域：呼叫當下 `[bp+0C]` 是 `0374:0556` 手上的星域記錄指標（`ds:A79C ＋
// 編號 × 126`）。
func rebelRegionsInReport(t *testing.T, batch string) []int {
	t.Helper()
	b := readReceipt(t, "oracle", batch, "report.txt")
	var out []int
	in := false
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "2703:1908 執行時的暫存器") {
			in = true
			continue
		}
		if !in {
			continue
		}
		if strings.TrimSpace(line) == "" || !strings.HasPrefix(line, "  ") {
			break
		}
		i := strings.Index(line, " 0C:")
		if !strings.Contains(line, "[bp+]") || i < 0 {
			continue
		}
		v, err := strconv.ParseUint(line[i+4:i+8], 16, 16)
		if err != nil || (v-0xA79C)%RegionRecord != 0 {
			t.Fatalf("%s 的 [bp+0C] 讀不懂：%q", batch, line)
		}
		out = append(out, int(v-0xA79C)/RegionRecord)
	}
	if !in {
		t.Fatalf("%s 的 report.txt 沒有 2703:1908 那一段", batch)
	}
	return out
}

func rebelSet(out [RegionCount]RebellionRoll) []int {
	var s []int
	for _, r := range out {
		if r.Rebels {
			s = append(s, r.Region)
		}
	}
	return s
}

func sameInts(a, b []int) bool {
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

// TestRebellionMatchesReceiptRandom 是亂數那一端的同狀態對拍：`rbx292e`
// 32 格逐格相同，含兩個「差一點」（星域 5 的 748 ≥ 250、星域 8 的
// 407 ≥ 350）。一格十次、只用第 1 與第 6 次這個對齊要是錯的，後面每一格
// 的亂數都會歪掉，不可能連兩個邊界值都對。
func TestRebellionMatchesReceiptRandom(t *testing.T) {
	mem := readReceipt(t, "rbx292e_pre.bin")
	want := rebelRegionsInReport(t, "rbx292e")
	in := rebellionInputFromDump(t, mem)
	if in.Rand.Seed() != 0xAF5AAD71 {
		t.Fatalf("傾印的種子 %08X，收據是 AF5AAD71", in.Rand.Seed())
	}
	out := DecideRebellions(in)
	if got := rebelSet(out); !sameInts(got, want) {
		t.Errorf("叛亂的星域 %v，收據的 2703:1908 是 %v", got, want)
	}
	// `docs/re/292` 那張表：門檻與 r1 ＋ r2 逐格相同。
	for _, c := range []struct{ region, thr, roll int }{
		{1, 300, 297}, {2, 225, 157}, {4, 250, 112}, {5, 250, 748}, {6, 400, 262},
		{7, 350, 347}, {8, 350, 407}, {9, 625, 427}, {13, 400, 201},
	} {
		r := out[c.region]
		if !r.Asked || r.Threshold != c.thr || r.Roll != c.roll {
			t.Errorf("星域 %d：問了 %v、門檻 %d、r1＋r2 %d，收據重播是 %d／%d",
				c.region, r.Asked, r.Threshold, r.Roll, c.thr, c.roll)
		}
	}
	// 同盟的 20／30 兩格有自己剛移進去的艦隊（`2703:0853` 重數之後
	// `+0x29` 是 1／2），不擲亂數；其餘 19 格擲。
	asked := 0
	for _, r := range out {
		if r.Asked {
			asked++
		}
	}
	if out[20].Asked || out[30].Asked || asked != 19 {
		t.Errorf("問了 %d 格（20：%v、30：%v），要 19 格、20 與 30 不問",
			asked, out[20].Asked, out[30].Asked)
	}
	// 收據 `rbx297w`（同一條序列加 `-watch` 監看 `ds:94DC`，`docs/re/297`）
	// 量到整圈 `056A:0469` 跑完、進 `0374:03E8` 時的種子是 936A823B
	// ＝ AF5AAD71 往前 190 步：32 次 `2703:1789` 裡 19 次各抽十個，
	// 20 與 30 兩格一個都沒抽。
	if in.Rand.Seed() != 0x936A823B {
		t.Errorf("整圈之後的種子 %08X，收據 rbx297w 是 936A823B", in.Rand.Seed())
	}
}

// TestRebellionMatchesReceiptRandomFirstRegion 是十次那個數字的正對照：
// `rbx292` 那一次卡在星域 1 的報告上（第 2 格起一次都沒問），`#218,000,000`
// 的種子 AF5AAD71 到 `#220,000,000` 的 BB6CC6CF 正好是十步。
func TestRebellionMatchesReceiptRandomFirstRegion(t *testing.T) {
	pre := readReceipt(t, "rbx292_c.bin")
	post := readReceipt(t, "rbx292_d.bin")
	r := NewRand(binary.LittleEndian.Uint32(pre[dumpSeed:]))
	r.Advance(RebellionDraws)
	if got := binary.LittleEndian.Uint32(post[dumpSeed:]); got != r.Seed() {
		t.Errorf("卡在第一格報告上的種子 %08X，走十步是 %08X", got, r.Seed())
	}
}

// rbx292Planets 把 `rbx292d` 的行星表切片（81 × 82，沒有檔頭）接上
// `PLADAT.DAT` 的檔頭（每星域行星數）。
func rbx292Planets(t *testing.T, slice []byte) *Planets {
	t.Helper()
	hdr, err := os.ReadFile(filepath.Join("..", "..", "org_game", "extracted", "PLADAT.DAT"))
	if err != nil {
		t.Skipf("缺 PLADAT.DAT，跳過（不用代用品）")
	}
	b := append(append([]byte(nil), hdr[:PlanetHeader]...), slice...)
	p, err := LoadPlanets(b)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestRebellionMatchesReceiptEffect 是效果那一端：`rbx292d` 的九格叛亂，
// 被翻的九顆行星（擁有者、經濟力、防衛力）與 81 顆的擁有者逐顆相同，
// 回合尾巴依行星重算之後的 32 格擁有者也逐格相同。
func TestRebellionMatchesReceiptEffect(t *testing.T) {
	mem := readReceipt(t, "rbx292_c.bin")
	want := rebelRegionsInReport(t, "rbx292d")
	regPre := readReceipt(t, "rbx292d_reg_pre.bin")
	regPost := readReceipt(t, "rbx292d_reg_post.bin")
	plaPre := readReceipt(t, "rbx292d_pla_pre.bin")
	plaPost := readReceipt(t, "rbx292d_pla_post.bin")

	out := DecideRebellions(rebellionInputFromDump(t, mem))
	got := rebelSet(out)
	if !sameInts(got, want) {
		t.Fatalf("叛亂的星域 %v，收據的 2703:1908 是 %v", got, want)
	}
	if len(want) != 9 {
		t.Fatalf("收據的叛亂 %d 格，`docs/re/292` 記的是 9 格", len(want))
	}

	planets := rbx292Planets(t, plaPre)
	post := rbx292Planets(t, plaPost)
	recs := append([]byte(nil), regPre...)
	flipped := map[int]bool{}
	for _, r := range got {
		for _, i := range Rebel(recs[r*RegionRecord:(r+1)*RegionRecord], planets, r, false) {
			flipped[i] = true
		}
		if recs[r*RegionRecord+regionOwner] != RegionOwnerNone {
			t.Errorf("星域 %d 的 +0x12 沒寫成 2", r)
		}
	}
	if len(flipped) != 9 {
		t.Errorf("翻了 %d 顆，收據是 9 顆（一格一顆）", len(flipped))
	}
	for i := 0; i < 81; i++ {
		if planets.Owner(i) != post.Owner(i) {
			t.Errorf("行星 %d 的擁有者 %d，收據 %d", i, planets.Owner(i), post.Owner(i))
		}
		if !flipped[i] {
			continue // 其他行星同一回合還會被防衛力回復等常式動到，不在這一份
		}
		if a, b := planets.PlanetEconomy(i), post.PlanetEconomy(i); a != b {
			t.Errorf("行星 %d 的經濟力 %d，收據 %d", i, a, b)
		}
		if a, b := planets.Defence(i), post.Defence(i); a != b {
			t.Errorf("行星 %d 的防衛力 %d，收據 %d", i, a, b)
		}
	}
	// 回合尾巴依行星重算擁有者（`2703:159B`）之後，32 格與收據逐格相同。
	for r := 0; r < RegionCount; r++ {
		if got, want := planets.RegionOwner(r), int(regPost[r*RegionRecord+regionOwner]); got != want {
			t.Errorf("星域 %d 重算的擁有者 %d，收據 %d", r, got, want)
		}
	}
}
