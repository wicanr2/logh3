package game

import (
	"encoding/binary"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

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
// 沒有戰鬥的格子：單方占據（bit4）的跑戰略打行星（`20B4:1497`，
// `docs/re/322` §五）；其餘先依序跑 `2703:1B48`（無主格再跑掉一顆
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

// runRebellions 判這一回合的叛亂並排好報告（`056A:0469` 的 bit1 那一條）。
// 效果不能在這裡先套：原版要等報告結束，再由 `0374:03E8` 逐格掃描；較小
// 編號的戰鬥可能先重算未安定度，較大編號的叛亂效果才讀到新值
// （`docs/spec/042`、`docs/re/310`）。
//
// 讀的狀態都是**艦隊移動之後**的：
//   - bit2／bit4：同一圈先算（`Occupations`，停在自家星域的不算占據）。
//   - `+0x28`／`+0x29`：`2703:0853` 在這一圈的前一步照移動之後的位置
//     重數，remake 直接數 `g.fleets`（`rules.CountFleets`）。
//   - 擁有者 `+0x12` 與未安定度 `+0x7D`：上一個回合尾巴留下來的值。
func (g *Game) runRebellions() []int {
	g.rebellions = nil
	g.rebellionReports = nil
	g.rebellionOccupation = [rules.RegionCount]rules.RegionOccupation{}
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
	g.rebellionOccupation = in.Occupation
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
		if roll.Rebels {
			// `056A:06F3` 在之後的效果把星域改成無主之前取擁有者；報告
			// 必須在判定時保存，不能在 Draw 時重讀（`docs/spec/046`）。
			owner := int(g.regionRecord(r)[regionOwnerOff])
			g.rebellions = append(g.rebellions, r)
			g.rebellionReports = append(g.rebellionReports, RebellionReport{Region: r, Owner: owner})
		}
	}
	return g.rebellions
}

// scanRegionsFrom 跑原版 `0374:03E8` 的逐星域迴圈。碰到判定時已有 bit2 的
// 星域便進戰鬥並暫停；`EndBattle` 從下一格接回。其餘格子的被動效果與叛亂
// 效果只在實際掃到時執行，讓它讀到先前戰鬥之後重算的 `+0x7D`。
func (g *Game) scanRegionsFrom(from int) {
	for r := from; r < rules.RegionCount; r++ {
		g.setScannedRegion(r)
		occ := g.rebellionOccupation[r]
		if occ.Contested {
			// 開戰之前先報告「<星域>星系內／發現敵方艦隊。」（`0374:0489`，規格 059、`docs/re/340`）。
			// `0374:04DA` 先 `3705:0349(0)` 整片還原，再 `3705:03FD(星域, 5)` 把要開戰的那一格塗起來。
			g.mapMark = r
			region := r
			if g.startTurnReports(g.enemyFleetFoundReports(r), func() { g.startPlacedBattle(region) }) {
				return
			}
			g.startPlacedBattle(r)
			return
		}
		// 自家星域的艦艇補充（`1AAB:0FAF`，規格 063、`docs/re/346`）：擁有者停在這一格的艦隊
		// 逐艦種補到滿編，補完逐支 `1261:0067(f, 1)` 可能印「已經恢復」。原版排在這一格的
		// 全滅判定與打行星之前，所以補上去的艘數會進到同一格的耗損判定。
		region := r
		if g.resupplyRegion(r, func() { g.scanRegionAfterResupply(region) }) {
			return
		}
		// 耗損與全滅（`0374:0608`）：停在全滅報告上就等看完，從這一格重跑（被拿掉的那一支已不在表上）。
		if g.attritionFrom(r, 0, func() { g.scanRegionsAfterAttrition(region) }) {
			return
		}
		if g.scanRegionTail(r) {
			return
		}
	}
	g.setScannedRegion(rules.RegionCount)
	g.endOfTurn()
}

// scanRegionAfterResupply 是補充的士氣報告看完之後：接這一格的耗損／全滅判定。
func (g *Game) scanRegionAfterResupply(r int) {
	if g.attritionFrom(r, 0, func() { g.scanRegionsAfterAttrition(r) }) {
		return
	}
	if g.scanRegionTail(r) {
		return
	}
	g.scanRegionsFrom(r + 1)
}

// scanRegionsAfterAttrition 是全滅報告看完之後：這一格剩下的部分，再接下一格。
func (g *Game) scanRegionsAfterAttrition(r int) {
	if g.scanRegionTail(r) {
		return
	}
	g.scanRegionsFrom(r + 1)
}

// scanRegionTail 是 `0374:0556` 在全滅判定之後的部分：單方面占據的打行星，或被動效果與叛亂。
// 停在報告上回 true（報告看完自己接下一格）。
func (g *Game) scanRegionTail(r int) bool {
	if g.rebellionOccupation[r].Occupied {
		// 單方面占據的格子改跑戰略打行星（`20B4:1497`：三次回復、打第一顆不是
		// 自己的行星、反擊、換手，最多九回合；規格 058、`docs/re/330`）。
		// 拿下行星的訊息框（`20B4:1922`）不改狀態、不耗亂數，停在打完之後印。
		// bit4 是 `056A:0469` 先算好的，全滅拿掉攻方之後照樣進來（攻方空了就不打）。
		next := r + 1
		return g.startTurnReports(g.strategicAssault(r), func() { g.scanRegionsFrom(next) })
	}
	rules.DepartNeutralPlanet(g.regionRecord(r), g.res.Planets, r, g.rand94DC())
	rules.RecoverRegionDefences(g.res.Planets, r)
	rules.ResetRegionFortressGuns(g.res.Planets, r)
	if g.rebellionRolls[r].Rebels {
		rules.Rebel(g.regionRecord(r), g.res.Planets, r, false)
	}
	return false
}

// Rebellions 回這一回合叛亂的星域。
func (g *Game) Rebellions() []int { return g.rebellions }

// RebellionReport 保存報告發生的星域與**事件前**擁有者。叛亂效果會先把
// 星域改成無主；延遲到 Draw 才讀 owner 會選錯名字顏色與情報參謀頭像。
type RebellionReport struct {
	Region int
	Owner  int
}

// startRebellionReports 把本回合的每一格叛亂排成模態報告。原版逐格
// `056A:06F3` 一格一則；沒有叛亂時不切 state。
func (g *Game) startRebellionReports(after func()) bool {
	if len(g.rebellionReports) == 0 || g.res == nil || g.res.FlipReport == nil {
		return false
	}
	g.pendingRebellionReports = append(g.pendingRebellionReports[:0], g.rebellionReports...)
	g.afterRebellionReports = after
	g.state = StateRebellionReport
	return true
}

// CurrentRebellionReport 回現在顯示的星域編號。
func (g *Game) CurrentRebellionReport() (RebellionReport, bool) {
	if len(g.pendingRebellionReports) == 0 {
		return RebellionReport{}, false
	}
	return g.pendingRebellionReports[0], true
}

// dismissRebellionReport 收掉一則；全部看完才接回戰鬥掃描。
func (g *Game) dismissRebellionReport() {
	g.keepMapMark()
	if len(g.pendingRebellionReports) > 0 {
		g.pendingRebellionReports = g.pendingRebellionReports[1:]
	}
	if len(g.pendingRebellionReports) > 0 {
		return
	}
	after := g.afterRebellionReports
	g.afterRebellionReports = nil
	if after != nil {
		after()
		return
	}
	g.state = StateCommandPhase
}

// drawRebellionReport 畫 `056A:06F3`：星域名、`發生叛亂。` 與該星域
// 擁有者的情報參謀頭像。文字與幾何共用已由 chain3/6/7/8 證實的 panel，
// rbx292 另釘住 rebellion 這個 message key（`docs/spec/046`）。
func (g *Game) drawRebellionReport(fb *gfx.Framebuffer) {
	report, ok := g.CurrentRebellionReport()
	if !ok || g.res == nil || g.res.FlipReport == nil {
		return
	}
	rec := g.regionRecord(report.Region)
	if rec == nil {
		return
	}
	g.res.FlipReport.Draw(fb, g.res.Font, rules.RegionName(rec), report.Owner, "rebellion")
	g.drawReportFace(fb, report.Owner)
}

// strategicAssault 是 `20B4:1497(星域)`：攻方 ＝ 這一格有帝國艦隊（士氣低落的不算，
// `297B:14F8`）就是帝國、否則同盟；攻方在場的艦隊依艦隊表順序排進 `rules.StrategicAssault`。
// 行星記錄就地改、艦隊的艘數與士氣改在結構上、功績寫提督 `+0x02`、兩個計分器照加。
//
// 回傳拿下行星的訊息框（`20B4:1922`，每拿下一顆一則，依拿下的順序）。
func (g *Game) strategicAssault(region int) []turnReport {
	if g.res.Planets == nil {
		return nil
	}
	attacker := 1
	for _, f := range g.fleets {
		if f.Region == region && f.Faction == 0 && f.Flags&0x01 == 0 {
			attacker = 0
		}
	}
	start, n := g.res.Planets.Range(region)
	if n == 0 {
		return nil
	}
	lo := rules.PlanetHeader + start*rules.PlanetRecord
	planets := g.res.Planets.Data[lo : lo+n*rules.PlanetRecord]
	var sites []rules.PlanetSite
	cells := g.battleCells(region)
	if len(cells) > 0 && g.res.Battle != nil {
		sites = g.loadRegionPlanetSites(region)
	}
	// 每一圈的定位（`0DEA:1759`）：同一格兩支以上時後面的被擠開，距離不是 0（`docs/re/335` §六）。
	place := &rules.StrategicPlacement{Slots: g.res.Formation}
	if g.res.Terrain != nil && len(cells) > 0 {
		place.Passable = func(x, y int) bool { return g.res.Terrain.Passable(cells, x, y) }
	}
	var sfs []rules.StrategicFleet
	for i := range g.fleets {
		f := &g.fleets[i]
		if f.Region != region || f.Faction != attacker {
			continue
		}
		adm := g.admiralRecord(f.Admiral)
		ab := rules.FleetAbilities(adm, g.admiralRecord(f.Adjutant))
		command := 0
		if len(adm) > 0x0B {
			command = int(adm[0x0B])
		}
		sfs = append(sfs, rules.StrategicFleet{Fleet: f, Ability40: int(ab[3]), Ability41: int(ab[4]), Command: command})
	}
	if len(sfs) == 0 {
		return nil
	}
	for i := range g.fleets {
		if f := &g.fleets[i]; f.Region == region && f.Faction != attacker {
			place.Others = append(place.Others, f)
		}
	}
	res := rules.StrategicAssault(planets, sites, attacker, sfs, g.rand94DC(), place)
	g.mutableAdmirals()
	for a, v := range res.Merit {
		if rec := g.admiralRecord(a); len(rec) >= 6 {
			binary.LittleEndian.PutUint32(rec[2:], binary.LittleEndian.Uint32(rec[2:])+uint32(v))
		}
	}
	for f := 0; f < 2; f++ {
		g.battleScore[f] += res.Score[f]
		g.campaignScore[f] += res.Score[f]
	}
	var reports []turnReport
	if g.res.TurnWindow != nil {
		for k, pi := range res.Captured {
			reports = append(reports, g.captureReport(res.CapturedBy[k], region, pi))
		}
	}
	return reports
}
