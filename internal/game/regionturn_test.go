package game

import (
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
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

func TestRecomputeRegionsFollowsPlanets(t *testing.T) {
	g := newRegionTurnGame(t)
	// 記錄的 `+0x12` 初值是 0，所以只有 16–31 那 16 格算易主。
	if n := g.recomputeRegions(); n != 16 {
		t.Fatalf("第一次應該 16 格易主，得到 %d", n)
	}
	if n := g.recomputeRegions(); n != 0 {
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
	if n := g.recomputeRegions(); n != 1 {
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

// `+0x13` 的其餘位元原版不動，這裡也不動（`and 0F2h` ＋ `or`）。
func TestRecomputeRegionsKeepsOtherFlagBits(t *testing.T) {
	g := newRegionTurnGame(t)
	g.regionRecord(5)[0x13] = 0xFF
	g.recomputeRegions()
	if got := g.regionRecord(5)[0x13] & 0xF2; got != 0xF2 {
		t.Errorf("bit1／bit4–bit7 被動到了：%02X", g.regionRecord(5)[0x13])
	}
}
