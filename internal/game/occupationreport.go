package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// 單方面占據報告（原版 `056A:0469` → `056A:053F` → `056A:0571` →
// `056A:059E`，`docs/spec/047`、`docs/re/306`）。
//
// bit4 的條件是「只有一方艦隊停在不屬於它的星域」。bit4 每回合重算，
// 所以報告不是只在邊緣觸發：舊 bit4 決定用「持續」句，新的 bit4 決定
// 這回合是否顯示。

// OccupationReport 保存一則報告畫面需要的事件時狀態。
type OccupationReport struct {
	Region     int
	Owner      int
	Occupier   int
	Seat       int
	Continuing bool
}

// MessageKey 照 `056A:059E` 選四句：先看占據者是不是本座位，再看舊 bit4。
func (r OccupationReport) MessageKey() string {
	ours := r.Occupier == r.Seat
	switch {
	case ours && r.Continuing:
		return "occupation_continuing_us"
	case ours:
		return "occupation_new_us"
	case r.Continuing:
		return "occupation_continuing_enemy"
	default:
		return "occupation_new_enemy"
	}
}

// occupationFaction 回該格唯一仍有效的艦隊陣營。呼叫端只在
// RegionOccupation.Occupied 時使用，因此正常結果必為 0 或 1。
func (g *Game) occupationFaction(region int) int {
	for _, f := range g.fleets {
		if f.Flags&1 == 0 && f.Region == region && (f.Faction == 0 || f.Faction == 1) {
			return f.Faction
		}
	}
	return rules.RegionOwnerNone
}

// prepareOccupationReports 重現 `056A:0469` 的 bit2／bit4 重算，並把每一格
// 新 bit4 排成報告。現有 remake 以整圈批次計算；逐格暫停造成的 RNG 時序
// 限制明列在 SPEC 047，不在 presentation 裡猜補。
func (g *Game) prepareOccupationReports() {
	g.occupationReports = nil
	if !g.regionsReady() {
		return
	}
	owner := make([]int, rules.RegionCount)
	for r := range owner {
		owner[r] = int(g.regionRecord(r)[regionOwnerOff])
	}
	occ := g.res.Graph.Occupations(owner, g.fleets)
	for r, o := range occ {
		rec := g.regionRecord(r)
		old := rec[regionFlagsOff]
		rec[regionFlagsOff] = rules.ApplyOccupation(old, o)
		if !o.Occupied || g.faction < 0 || g.faction > 1 {
			continue
		}
		g.occupationReports = append(g.occupationReports, OccupationReport{
			Region: r, Owner: owner[r], Occupier: g.occupationFaction(r),
			Seat: g.faction, Continuing: old&0x10 != 0,
		})
	}
}

func (g *Game) startOccupationReports(after func()) bool {
	if len(g.occupationReports) == 0 || g.res == nil || g.res.FlipReport == nil {
		return false
	}
	g.pendingOccupationReports = append(g.pendingOccupationReports[:0], g.occupationReports...)
	g.afterOccupationReports = after
	g.state = StateOccupationReport
	return true
}

func (g *Game) CurrentOccupationReport() (OccupationReport, bool) {
	if len(g.pendingOccupationReports) == 0 {
		return OccupationReport{}, false
	}
	return g.pendingOccupationReports[0], true
}

func (g *Game) dismissOccupationReport() {
	if len(g.pendingOccupationReports) > 0 {
		g.pendingOccupationReports = g.pendingOccupationReports[1:]
	}
	if len(g.pendingOccupationReports) > 0 {
		return
	}
	after := g.afterOccupationReports
	g.afterOccupationReports = nil
	if after != nil {
		after()
		return
	}
	g.state = StateCommandPhase
}

func (g *Game) drawOccupationReport(fb *gfx.Framebuffer) {
	report, ok := g.CurrentOccupationReport()
	if !ok || g.res == nil || g.res.FlipReport == nil {
		return
	}
	rec := g.regionRecord(report.Region)
	if rec == nil {
		return
	}
	g.res.FlipReport.DrawInline(fb, g.res.Font, rules.RegionName(rec),
		report.Owner, report.MessageKey())
	g.drawReportFace(fb, report.Seat)
}
