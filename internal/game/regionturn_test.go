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

// 兩支常式合起來管 bit0–bit4（`and 0F2h` ＋ `and 0E9h`），**bit5–bit7
// 兩邊都不動**（`docs/re/245`）。
func TestRecomputeRegionsKeepsOtherFlagBits(t *testing.T) {
	g := newRegionTurnGame(t)
	g.regionRecord(5)[0x13] = 0xFF
	g.recomputeRegions()
	if got := g.regionRecord(5)[0x13] & 0xE0; got != 0xE0 {
		t.Errorf("bit5–bit7 被動到了：%02X", g.regionRecord(5)[0x13])
	}
}

// bit4 ＝ 單方面占據：艦隊停在不屬於它的星域、而對方沒有艦隊。
// 待在自家地盤不算（`056A:0468`，`docs/re/245`）。
func TestRecomputeRegionsMarksOccupation(t *testing.T) {
	g := newRegionTurnGame(t) // 0–15 帝國、16–31 同盟
	// 帝國一支停在同盟的星域 20 → bit4；停在自己的星域 5 → 不設。
	g.fleets = []rules.Fleet{{Faction: 0, Region: 20}, {Faction: 0, Region: 5}}
	g.recomputeRegions()
	if got := g.regionRecord(20)[0x13]; got&0x10 == 0 {
		t.Errorf("星域 20 應該設 bit4，+0x13 ＝ %02X", got)
	}
	if got := g.regionRecord(5)[0x13]; got&0x10 != 0 {
		t.Errorf("星域 5 是自家地盤，不該設 bit4，+0x13 ＝ %02X", got)
	}
	// 雙方都在 → bit2，而且 bit4 不設。
	g.fleets = append(g.fleets, rules.Fleet{Faction: 1, Region: 20})
	g.recomputeRegions()
	if got := g.regionRecord(20)[0x13]; got&0x04 == 0 || got&0x10 != 0 {
		t.Errorf("星域 20 應該只設 bit2，+0x13 ＝ %02X", got)
	}
	// `Flags` bit0 有設的不計（`297B:14F8` 的 `test 1`）。
	g.fleets = []rules.Fleet{{Faction: 0, Region: 20, Flags: 1}}
	g.recomputeRegions()
	if got := g.regionRecord(20)[0x13]; got&0x14 != 0 {
		t.Errorf("bit0 有設的不該算進去，+0x13 ＝ %02X", got)
	}
}

// 打完戰鬥就重算：原版走完回合結束那條鏈（`docs/re/105` 的 `s3win`
// 收據裡 `2703:0944`／`2703:0990` 都在 `2703:0933` 內，而它只有兩個
// 呼叫者），所以 `EndBattle` 也要跑一次。
func TestEndBattleRecomputesRegions(t *testing.T) {
	g := newRegionTurnGame(t)
	g.recomputeRegions()
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

	g.EndBattle()
	if got := g.res.Planets.Owner(20); got != 0 {
		t.Errorf("行星沒寫回：擁有者 ＝ %d，預期 0", got)
	}
	if got := int(g.regionRecord(20)[0x12]); got != 0 {
		t.Errorf("星域 20 的擁有者 ＝ %d，預期 0", got)
	}
	if g.state != StateCommandPhase {
		t.Errorf("狀態 ＝ %v，預期回指令畫面", g.state)
	}
}
