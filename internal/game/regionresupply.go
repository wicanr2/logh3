package game

import "github.com/wicanr2/logh3/internal/rules"

// 自家星域的艦艇補充（`1AAB:0FAF`，#71、`docs/re/346`、規格 063）。
//
// 逐格掃描（`0374:0556`）到這一格時的第二步，排在耗損／全滅之前。規則層在
// `rules.ResupplyRegion`；這裡只負責補完之後逐支 `1261:0067(艦隊, 1)`——
// 士氣重算 ＋ 當場印報告（「已經恢復」就是從這裡出來的）。

// resupplyRegion 跑這一格的補充。停在報告上回 true（報告看完自己接 next）。
func (g *Game) resupplyRegion(r int, next func()) bool {
	if g.res == nil || g.res.Planets == nil {
		return false
	}
	rec := g.regionRecord(r)
	if rec == nil {
		return false
	}
	ptrs := make([]*rules.Fleet, len(g.fleets))
	at := make(map[*rules.Fleet]int, len(g.fleets))
	for i := range g.fleets {
		ptrs[i] = &g.fleets[i]
		at[ptrs[i]] = i
	}
	mine := rules.ResupplyRegion(rec, g.res.Planets, r, ptrs, g.res.ShipKinds)
	if len(mine) == 0 {
		return false
	}
	// 原版逐支「重算 → 立刻印」；報告本來就照艦隊順序排，所以這裡收完再一起排隊。
	var reports []turnReport
	for _, f := range mine {
		reports = append(reports, g.fleetStepMorale(at[f])...)
	}
	if len(reports) == 0 {
		return false
	}
	return g.startTurnReports(reports, next)
}
