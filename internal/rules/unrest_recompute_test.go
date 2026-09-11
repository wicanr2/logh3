package rules

import (
	"encoding/binary"
	"testing"
)

// 遊戲進行中的未安定度重算（`docs/spec/042`）。算式本身與載入時那一支相同
// （`2703:0DA1`，`docs/spec/006`），差在兩個輸入：
//
//   - 尾巴讀的是 `2B99:0003` 照擁有者建好的路徑表（成本 >= 0x800 就砍半、
//     否則沿中點矩陣數步數），不是載入當下那一種；
//   - 回合計數器、Blend、`ds:8FF8`、內政專任參謀都是那一刻的值。
//
// 收據的傾印是線性 0 起的 1 MB，版面與 `rebellion_test.go` 相同。

// unrestDump 從一份 1 MB 傾印湊出算未安定度要的輸入（星域表另外給，
// 行星表在 `5313:0002`，`docs/internals/29`）。
type unrestDump struct {
	mem    []byte
	params UnrestParams
}

func readUnrestDump(t *testing.T, name string) unrestDump {
	t.Helper()
	mem := readReceipt(t, name)
	if len(mem) != 1<<20 {
		t.Fatalf("%s 有 %d bytes，要 1 MB", name, len(mem))
	}
	u16 := func(a int) int { return int(binary.LittleEndian.Uint16(mem[a:])) }
	p := UnrestParams{Turn: u16(dumpDS + 0x8F18)}
	for f := 0; f < 2; f++ {
		p.Ruler[f] = int(mem[dumpDS+0x8F4F+f]) // 內政專任參謀
		p.Blend[f] = int(mem[dumpDS+0x8FD4+f])
		p.Rate[f] = int(mem[dumpDS+0x8FF8+f])
	}
	for i := 0; i < 5; i++ {
		p.Table = append(p.Table, int(mem[dumpDS+0x4B28+i]))
		p.Base = append(p.Base, u16(dumpDS+0x147E+4*i)) // 32 位元陣列的低位字
	}
	return unrestDump{mem: mem, params: p}
}

const dumpPlanets = 0x53132 // 5313:0002，81 筆 × 82 bytes，沒有檔頭

func (d unrestDump) planets(t *testing.T) *Planets {
	return rbx292Planets(t, d.mem[dumpPlanets:dumpPlanets+81*PlanetRecord])
}

func (d unrestDump) regions() []byte {
	return append([]byte(nil), d.mem[dumpRegions:dumpRegions+RegionCount*RegionRecord]...)
}

func (d unrestDump) state(graph *StarGraph, regions []byte, planets *Planets) *UnrestState {
	return &UnrestState{
		Regions: regions, Planets: planets,
		Fleets:   d.mem[dumpFleets : dumpFleets+FleetCount*FleetRecord],
		Admirals: d.mem[dumpAdmirals : dumpAdmirals+82*AdmiralRecord],
		Graph:    graph, Params: d.params,
		Paths: graph.UnrestPaths(regionOwners(regions)),
	}
}

func regionOwners(regions []byte) []int {
	owner := make([]int, RegionCount)
	for i := range owner {
		owner[i] = int(regions[i*RegionRecord+regionOwner])
	}
	return owner
}

// checkUnrest 逐格比 32 格，回有幾格不符。
func checkUnrest(t *testing.T, st *UnrestState, want []byte) int {
	t.Helper()
	bad := 0
	for i := 0; i < RegionCount; i++ {
		got, w := st.Unrest(i), int(want[i*RegionRecord+RegionUnrest])
		if got != w {
			bad++
			t.Errorf("星域 %d：算出來 %d，收據是 %d", i, got, w)
		}
	}
	return bad
}

// TestUnrestRecomputeAfterAITurn 是 worklist `region-unrest-recompute` 的
// 判準：收據 `rbx292` 在第一個 AI 回合跑完的那一刻（`rbx292_b.bin`，
// `#216,000,000`），拿那一刻的局面重算，32 格與傾印逐格相同。
//
// 正對照：AI 回合之前那一份（`rbx292_a.bin`）與它比有 14 格不同——同盟 13 格
// （Blend 2 → 0、內政參謀 60 → 63）與被 poke 成 255 的星域 1（`-poke
// lin:5F3C7@211000000=FF`，重算之後回到 38）。
func TestUnrestRecomputeAfterAITurn(t *testing.T) {
	graph := loadGraph(t)
	before := readUnrestDump(t, "rbx292_a.bin")
	after := readUnrestDump(t, "rbx292_b.bin")
	if after.params.Blend != [2]int{2, 0} || after.params.Ruler != [2]int{16, 63} || after.params.Turn != 0 {
		t.Fatalf("rbx292_b 的參數 %+v，預期 Blend 2/0、內政參謀 16/63、回合 0", after.params)
	}
	regs := after.regions()
	if bad := checkUnrest(t, after.state(graph, regs, after.planets(t)), regs); bad != 0 {
		t.Fatalf("%d 格不符", bad)
	}
	changed := 0
	old := before.regions()
	for i := 0; i < RegionCount; i++ {
		if old[i*RegionRecord+RegionUnrest] != regs[i*RegionRecord+RegionUnrest] {
			changed++
		}
	}
	if changed != 14 {
		t.Errorf("AI 回合前後 %d 格的未安定度不同，收據是 14 格（同盟 13 ＋ 星域 1）", changed)
	}
	if old[1*RegionRecord+RegionUnrest] != 0xFF || regs[1*RegionRecord+RegionUnrest] != 38 {
		t.Errorf("星域 1：AI 回合前 %d、後 %d，收據是 255 → 38",
			old[1*RegionRecord+RegionUnrest], regs[1*RegionRecord+RegionUnrest])
	}
}

// TestUnrestPathsAreTheLiveTables 釘「遊戲進行中尾巴讀的表 ＝ 照現在的
// 擁有者建的表」：`rbx292_b.bin` 那一刻 EMS 分頁框 `D000` 正映著四張路徑表
// （`ds:D062`–`ds:D06E` 都指向 `D000:xxxx`），與 `UnrestPaths` 逐格相同。
func TestUnrestPathsAreTheLiveTables(t *testing.T) {
	graph := loadGraph(t)
	d := readUnrestDump(t, "rbx292_b.bin")
	for k := 0; k < 4; k++ {
		off := binary.LittleEndian.Uint16(d.mem[dumpDS+0xD062+4*k:])
		seg := binary.LittleEndian.Uint16(d.mem[dumpDS+0xD062+4*k+2:])
		if want := [4]uint16{0, 0x800, 0x1000, 0x1400}[k]; seg != 0xD000 || off != want {
			t.Fatalf("ds:%04X 指向 %04X:%04X，預期 D000:%04X", 0xD062+4*k, seg, off, want)
		}
	}
	p := graph.UnrestPaths(regionOwners(d.regions()))
	for f := 0; f < 2; f++ {
		for i := 0; i < RegionCount; i++ {
			for j := 0; j < RegionCount; j++ {
				dist := int(binary.LittleEndian.Uint16(d.mem[0xD0000+f*0x800+(i*32+j)*2:]))
				mid := int(d.mem[0xD1000+f*0x400+i*32+j])
				if dist != p.Dist[f][i][j] || mid != p.Mid[f][i][j] {
					t.Fatalf("陣營 %d [%d][%d]：傾印 %d／%d，建出來 %d／%d",
						f, i, j, dist, mid, p.Dist[f][i][j], p.Mid[f][i][j])
				}
			}
		}
	}
}

// TestUnrestRecomputeAfterTurnEnd 是同一條序列跑完整個回合：收據 `rbx292d`
// （營運 poke 成 1，九格叛亂）在 `#239,000,000` 倒的星域表與行星表。九格
// 被翻成無主（未安定度 100），其餘 23 格照新的鄰居、新的路徑表、**進位之後
// 的回合計數器 1** 重算——那一次是下一個座位回合開頭 `0654:0034` 的共同
// 尾巴（`docs/spec/035`）。艦隊、提督與四個全域取 `rbx292_e.bin`（同一條
// 輸入、`#225,000,000`，移動早就跑完）。
func TestUnrestRecomputeAfterTurnEnd(t *testing.T) {
	graph := loadGraph(t)
	d := readUnrestDump(t, "rbx292_e.bin")
	regs := readReceipt(t, "rbx292d_reg_post.bin")
	planets := rbx292Planets(t, readReceipt(t, "rbx292d_pla_post.bin"))
	d.params.Turn = 1
	if bad := checkUnrest(t, d.state(graph, regs, planets), regs); bad != 0 {
		t.Fatalf("%d 格不符", bad)
	}
	// 反對照：計數器用進位之前的 0，同盟那 13 格就全部對不上（f1 的
	// 「回合 % 10」那一項）。
	d.params.Turn = 0
	st := d.state(graph, regs, planets)
	off := 0
	for i := 0; i < RegionCount; i++ {
		if st.Unrest(i) != int(regs[i*RegionRecord+RegionUnrest]) {
			off++
		}
	}
	if off != 13 {
		t.Errorf("回合計數器用 0 時 %d 格不符，預期同盟 13 格", off)
	}
}

// TestUnrestRecomputeHalvesFarRegions 是「路徑成本 >= 0x800 就把治安度砍半」
// 那一條分支的收據：`urx299c`（`chain4` 那條序列在 dosgolem `e0b6a03` 上重做，
// `docs/re/246`、`docs/re/299`）把星域 4 唯一那顆行星 poke 成同盟，回合尾巴它
// 易主，到同盟首都的成本是 4218（`0x400 × 4 ＋ 122`）。
//
// 監看星域 4 的 `+0x7D` 抓到兩次寫入，都是 `2703:0DA1`：回合尾巴
// `2703:08DE` 那一次寫 128（計數器還是 0）、下一個座位回合開頭
// `090A:0E6D` 那一次寫 127（計數器 1）。傾印（`#231,000,000`）32 格逐格
// 相同；若照載入當下的讀法（一步扣 2、不砍半）星域 4 會算成 71。
func TestUnrestRecomputeHalvesFarRegions(t *testing.T) {
	graph := loadGraph(t)
	d := readUnrestDump(t, "urx299c_post.bin")
	if d.params.Turn != 1 {
		t.Fatalf("回合計數器 %d，預期 1", d.params.Turn)
	}
	regs := d.regions()
	st := d.state(graph, regs, d.planets(t))
	if c := st.Paths.Dist[1][CapitalRegion(1)][4]; c != 4218 {
		t.Fatalf("星域 4 到同盟首都的成本 %d，預期 4218", c)
	}
	if bad := checkUnrest(t, st, regs); bad != 0 {
		t.Fatalf("%d 格不符", bad)
	}
	st.Params.Turn = 0
	if got := st.Unrest(4); got != 128 {
		t.Errorf("計數器 0 時星域 4 算出 %d，監看到 `2703:08DE` 那一次寫 128", got)
	}
	st.Params.Turn = 1
	st.Paths = nil
	if got := st.Unrest(4); got != 71 {
		t.Errorf("星域 4 不砍半時算出 %d，預期 71（反對照）", got)
	}
}

// TestRouteStepsWalksTheMidpointMatrix 對 `2B99:0450` 的步數：沿中點矩陣
// 一格一格走，走的步數要等於 `Route` 展開的邊數（兩張出貨的表 × 32 × 32）。
func TestRouteStepsWalksTheMidpointMatrix(t *testing.T) {
	graph := loadGraph(t)
	for f := 0; f < 2; f++ {
		mid := graph.midpoint(f)
		for i := 0; i < RegionCount; i++ {
			for j := 0; j < RegionCount; j++ {
				want := graph.Steps(i, j, f)
				if got := RouteSteps(mid, i, j); got != want {
					t.Errorf("陣營 %d：%d → %d 走 %d 步，Route 是 %d", f, i, j, got, want)
				}
			}
		}
	}
	if got := RouteSteps(nil, 0, 1); got != -1 {
		t.Errorf("沒有表時回 %d，應該是 -1", got)
	}
}

// TestRulerScaleKeepsLow16Bits 釘 `2703:0E3F` 的 `mul cx / cwd / idiv`：
// 乘積只留低 16 位元、當有號數除。
func TestRulerScaleKeepsLow16Bits(t *testing.T) {
	for _, c := range []struct{ v, ops, want int }{
		{100, 118, 118},
		{200, 100, 200},
		{246, 133, 327},  // 32718：還在有號範圍內
		{247, 133, -326}, // 32851 → 0x8053 → −32685 → −326.85 往零截斷
		{0, 133, 0},
	} {
		if got := rulerScale(c.v, c.ops); got != c.want {
			t.Errorf("rulerScale(%d, %d) ＝ %d，應該是 %d", c.v, c.ops, got, c.want)
		}
	}
}
