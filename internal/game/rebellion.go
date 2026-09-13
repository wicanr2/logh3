package game

import "github.com/wicanr2/logh3/internal/rules"

// 叛亂接進回合（`docs/spec/041`、`docs/re/292`、`docs/re/297`）。
//
// 原版在命令階段那一串 `056A:000C` 的最後一支 `056A:0469` 裡逐格判定
// （`2703:1789`），判完之後回到主迴圈的逐格開戰那一圈 `0374:03E8`，
// 沒有戰鬥的格子跑 `0374:0556`，讀到 bit1 就套效果（`2703:1908`）。
//
//	056A:000C:  … 056A:03E0（艦隊移動）→ 2703:051D → 2703:0853（重數艦隊）
//	            → 056A:0469（bit2／bit4／bit1）
//	0374:03E8:  for 星域: bit2 → 開戰；否則 0374:0556（… → 2703:1908）
//
// 沒有戰鬥或單方占據的格子先依序跑 `2703:1B48`（無主格再跑掉一顆
// 行星）、`2703:1CD7`（防衛力回復）、`2703:1D85`（要塞砲重置），最後才
// 套叛亂效果。這個順序很重要：前者會抽一次共用亂數，而叛亂行星是先回復
// 再扣三分之一。

// rand94DC 回原版的全域亂數（`ds:94DC`）。戰略 AI 與叛亂判定抽的是
// 同一個序列（`docs/re/87`、`docs/re/209`），所以只有一份。
func (g *Game) rand94DC() *rules.Rand {
	if g.strategyRand == nil {
		g.strategyRand = rules.NewRand(strategyRandSeed)
	}
	return g.strategyRand
}

// openingOwners 回原版的 `ds:B75C`：796 年 1 月開局那 32 格的擁有者。
//
// 那一份只有新遊戲初始化寫（`2703:0068`），讀存檔不重寫
// （`docs/re/297`：25 份 S3 起跑的傾印全部是開局那一份）。remake 的
// `worldStellar` 正好就是開局值，讀存檔之後也不被換掉。
func (g *Game) openingOwners() []byte {
	src := g.worldStellar
	if len(src) < rules.RegionCount*rules.RegionRecord {
		src = g.res.Stellar
	}
	out := make([]byte, rules.RegionCount)
	for i := range out {
		o := i * rules.RegionRecord
		if o+rules.RegionRecord <= len(src) {
			out[i] = src[o+regionOwnerOff]
		}
	}
	return out
}

// runRebellions 判這一回合的叛亂並套上效果（`056A:0469` 的 bit1 那一條
// ＋ `0374:0556` 的 `2703:1908`）。回叛亂的星域編號（遞增）。
//
// 讀的狀態都是**艦隊移動之後**的：
//   - bit2／bit4：同一圈先算（`Occupations`，停在自家星域的不算占據）。
//   - `+0x28`／`+0x29`：`2703:0853` 在這一圈的前一步照移動之後的位置
//     重數，remake 直接數 `g.fleets`（`rules.CountFleets`）。
//   - 擁有者 `+0x12` 與未安定度 `+0x7D`：上一個回合尾巴留下來的值。
func (g *Game) runRebellions() []int {
	g.rebellions = nil
	if !g.regionsReady() {
		return nil
	}
	owner := make([]int, rules.RegionCount)
	for i := range owner {
		owner[i] = int(g.regionRecord(i)[regionOwnerOff])
	}
	in := rules.RebellionInput{
		Records:    g.res.Stellar,
		Opening:    g.openingOwners(),
		Occupation: g.res.Graph.Occupations(owner, g.fleets),
		Rand:       g.rand94DC(),
	}
	for r := 0; r < rules.RegionCount; r++ {
		for f := 0; f < 2; f++ {
			in.Garrison[r][f] = rules.CountFleets(g.fleets, r, f)
		}
	}
	for f := 0; f < 2; f++ {
		in.Ops[f] = rules.StaffOps(g.workRecord(), g.res.Admiral, rules.PostIntel, f)
	}
	g.rebellionRolls = rules.DecideRebellions(in)
	for r, roll := range g.rebellionRolls {
		// 原版 `0374:03E8` 的 bit2（戰鬥）與 bit4（單方占據）分支都不進
		// `0374:0556`，所以三支月底處理與叛亂效果都只在這個分支執行。
		if in.Occupation[r].Contested || in.Occupation[r].Occupied {
			continue
		}
		rules.DepartNeutralPlanet(g.regionRecord(r), g.res.Planets, r, in.Rand)
		rules.RecoverRegionDefences(g.res.Planets, r)
		rules.ResetRegionFortressGuns(g.res.Planets, r)
		if roll.Rebels {
			rules.Rebel(g.regionRecord(r), g.res.Planets, r, false)
			g.rebellions = append(g.rebellions, r)
		}
	}
	return g.rebellions
}

// Rebellions 回這一回合叛亂的星域（還沒有報告畫面，`strategy-report-screen`）。
func (g *Game) Rebellions() []int { return g.rebellions }
