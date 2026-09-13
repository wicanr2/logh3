package game

import "github.com/wicanr2/logh3/internal/rules"

// runPhezzanEvent 在回合尾巴消費「費沙這一回合易主」的登記。
// 規則層不碰 UI；之後 #8 的戰略報告畫面可直接使用 phezzanRebellions。
func (g *Game) runPhezzanEvent() {
	if !g.phezzanPending || rules.PhezzanOccurred(g.workRecord()) || !g.regionsReady() {
		return
	}
	owner := rules.RegionOwnerOf(g.regionRecord(rules.PhezzanRegion))
	if owner < 0 || owner > 1 {
		return
	}
	target := 0
	if owner == 0 {
		target = 80
	}
	if g.res.Planets.Record(rules.PhezzanPlanet) == nil || g.res.Planets.Record(target) == nil {
		return
	}
	rules.SetPhezzanOccurred(g.mutableWork())
	g.phezzanPending = false
	turn := g.turn - 1 // 原版在這一刻尚未做回合計數器 ++。
	result := rules.RunPhezzanEvent(g.res.Stellar, g.res.Planets, owner,
		g.versus || owner == g.faction, turn,
		func(region, faction int) int {
			return rules.CountFleets(g.fleets, region, faction)
		})
	// 原版費沙包裝器 `1213:023B` 在本體回來後於 `1213:0256`
	// 立即呼叫 `090A:0E6D`（`docs/re/299`、`docs/re/304`）。事件會改
	// 經濟、民心及叛亂格，不能留到下一個座位回合才重算。
	g.recomputeUnrest(turn)
	g.phezzanRebellions = append(g.phezzanRebellions[:0], result.Rebellions...)
}

// PhezzanRebellions 回最近一次費沙事件翻成無主的星域。
func (g *Game) PhezzanRebellions() []int {
	return append([]int(nil), g.phezzanRebellions...)
}
