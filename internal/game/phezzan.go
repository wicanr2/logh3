package game

import (
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

type PhezzanReport struct {
	Owner, Viewer, Page int
}

func (r PhezzanReport) Key() string {
	if r.Viewer == r.Owner {
		if r.Owner == 0 {
			return "owner_0"
		}
		return "owner_1"
	}
	if r.Owner == 0 {
		return "opponent_0"
	}
	return "opponent_1"
}

// startPhezzanReports 重現 `1213:0096` 在事件本體前依序給人類座位看的
// 兩頁訊息。全部關完才執行經濟轉移、邊境叛離與懲罰。
func (g *Game) startPhezzanReports(after func()) bool {
	if !g.phezzanPending || rules.PhezzanOccurred(g.workRecord()) ||
		!g.regionsReady() || g.res.PhezzanReport == nil {
		return false
	}
	owner := rules.RegionOwnerOf(g.regionRecord(rules.PhezzanRegion))
	if owner < 0 || owner > 1 {
		return false
	}
	human := [2]bool{}
	if g.versus {
		human = [2]bool{true, true}
	} else if g.faction >= 0 && g.faction < 2 {
		human[g.faction] = true
	}
	g.pendingPhezzanReports = nil
	for _, viewer := range []int{owner, 1 - owner} {
		if !human[viewer] {
			continue
		}
		for page := 0; page < 2; page++ {
			g.pendingPhezzanReports = append(g.pendingPhezzanReports,
				PhezzanReport{Owner: owner, Viewer: viewer, Page: page})
		}
	}
	if len(g.pendingPhezzanReports) == 0 {
		return false
	}
	g.afterPhezzanReports = after
	g.state = StatePhezzanReport
	return true
}

func (g *Game) CurrentPhezzanReport() (PhezzanReport, bool) {
	if len(g.pendingPhezzanReports) == 0 {
		return PhezzanReport{}, false
	}
	return g.pendingPhezzanReports[0], true
}

func (g *Game) dismissPhezzanReport() {
	if len(g.pendingPhezzanReports) > 0 {
		g.pendingPhezzanReports = g.pendingPhezzanReports[1:]
	}
	if len(g.pendingPhezzanReports) > 0 {
		return
	}
	g.runPhezzanEvent()
	after := g.afterPhezzanReports
	g.afterPhezzanReports = nil
	if after != nil {
		after()
		return
	}
	g.state = StateCommandPhase
}

func (g *Game) drawPhezzanReport(fb *gfx.Framebuffer) {
	r, ok := g.CurrentPhezzanReport()
	if !ok || g.res == nil || g.res.PhezzanReport == nil {
		return
	}
	var faceIndex = rules.StaffAdmiral(g.workRecord(), rules.PostMilitary, r.Viewer)
	var img = (*asset.Image)(nil)
	if g.res.Faces != nil && faceIndex >= 0 {
		img, _ = g.res.Faces.Decode(faceIndex)
	}
	g.res.PhezzanReport.Draw(fb, g.res.Font, g.res.FrameTiles, r.Key(), r.Page, img)
}

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
	turn := g.turn // 原版在這一刻尚未做回合計數器 ++。
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
