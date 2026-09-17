package game

import (
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 回合裡的報告框（`3826:0307`，規格 059、`docs/re/333`）：攻占訊息框（`20B4:1922`）、
// 士氣報告（`1261:0138`）與全軍覆沒（`1261:0267`）。與過場視窗同一個矩形，每一則清空重印，
// 頭像換人才重畫；等 `ds:8F38` 個 tick 或一次滑鼠左／右鍵按下。

// turnReport 是一則報告：已經組好的文字（`3617:000A` 的 buf）與頭像的提督編號。
type turnReport struct {
	text []byte
	face int
}

// turnReportState 是排隊中的報告與全部看完之後接回哪裡。
type turnReportState struct {
	pending []turnReport
	after   func()
	resume  State
	ticks   int
	rem     time.Duration
}

// startTurnReports 把 reports 排成報告；沒有報告回 false、不切狀態。
func (g *Game) startTurnReports(reports []turnReport, after func()) bool {
	if len(reports) == 0 || g.res.TurnWindow == nil {
		return false
	}
	g.turnRep = turnReportState{pending: reports, after: after, resume: g.state}
	g.state = StateTurnReport
	return true
}

// turnReportTick 讓等待往前走 n 個 tick。
func (g *Game) turnReportTick(n int) {
	for ; n > 0 && g.state == StateTurnReport; n-- {
		g.turnRep.ticks++
		if g.turnRep.ticks >= g.display.MsgSpeed {
			g.dismissTurnReport()
		}
	}
}

// dismissTurnReport 結束一則；全部看完才接回。
func (g *Game) dismissTurnReport() {
	r := &g.turnRep
	r.ticks, r.rem = 0, 0
	if len(r.pending) > 0 {
		r.pending = r.pending[1:]
	}
	if len(r.pending) > 0 {
		return
	}
	after := r.after
	r.after = nil
	g.state = r.resume
	if after != nil {
		after()
	}
}

// drawTurnReport 畫目前這一則。
func (g *Game) drawTurnReport(fb *gfx.Framebuffer) {
	p := g.res.TurnWindow
	if p == nil || len(g.turnRep.pending) == 0 {
		return
	}
	cur := g.turnRep.pending[0]
	var face *asset.Image
	if g.res.Faces != nil {
		if img, err := g.res.Faces.Decode(cur.face); err == nil {
			face = img
		}
	}
	p.DrawReport(fb, g.res.Font, g.res.FrameTiles, cur.text, face)
}

// factionNameColoured 回帶色的陣營名（`ds:60F4[陣營]`）。
func (g *Game) factionNameColoured(faction int) []byte {
	if g.res.FleetWindow == nil || faction < 0 || faction >= len(g.res.FleetWindow.FactionNamesColoured) {
		return nil
	}
	return decodeHexBytes(g.res.FleetWindow.FactionNamesColoured[faction])
}

// captureReport 是 `20B4:1922(艦隊, 星域記錄, 行星)`：人類座位「我軍第Ｎ艦隊／提督／正在攻占…」＋提督頭像；
// 電腦座位「<陣營>艦隊／企圖占領…」＋對手陣營的情報參謀頭像（`ds:8F4D[3982:0250(陣營)]`）。
// 在行星換手之後、同一顆的功績之前印，所以行星名用新擁有者的顏色。
func (g *Game) captureReport(f *rules.Fleet, region, planet int) turnReport {
	p := g.res.TurnWindow
	var b []byte
	r := turnReport{face: f.Admiral}
	if g.humanSeat(f.Faction) {
		b = append(b, p.ReportFormat("capture_us")...)
		b = append(b, p.FleetNumber(f.Number)...)
		b = append(b, ui.Sprintf(p.ReportFormat("capture_us_admiral"), codesBig5(g.admiralNameCodes(f.Admiral)))...)
		b = append(b, p.ReportFormat("capture_us_verb")...)
	} else {
		b = append(b, ui.Sprintf(p.ReportFormat("capture_ai"), g.factionNameColoured(f.Faction))...)
		opp := 0
		if f.Faction == 0 {
			opp = 1
		}
		r.face = g.res.Reporter[opp]
		b = append(b, p.ReportFormat("capture_ai_verb")...)
	}
	b = append(b, p.PlainName(codesBig5(g.regionName(region)), g.regionOwner(region))...)
	start, _ := g.res.Planets.Range(region)
	b = append(b, p.ReportFormat("capture_planet")...)
	b = append(b, p.PlainName(planetNameBytes(g.res.Planets.Record(start+planet)), g.res.Planets.Owner(start+planet))...)
	b = append(b, p.ReportFormat("capture_newline")...)
	b = append(b, ui.Sprintf(p.ReportFormat("capture_tail"), g.factionNameColoured(f.Faction))...)
	r.text = b
	return r
}

// moraleReports 是 `1261:0104(艦隊)`：bit2（士氣跨過 50）且人類座位 → `1261:0138` 低落／恢復；
// bit3（全隊沒有艘數）→ `1261:0267`「已全軍覆沒」（不看座位）；最後清掉 bit2／bit3。
func (g *Game) moraleReports(f *rules.Fleet) []turnReport {
	p := g.res.TurnWindow
	var out []turnReport
	if p == nil {
		f.Flags &^= rules.FleetFlagMoraleFlip | rules.FleetFlagNoShips
		return nil
	}
	name := codesBig5(g.admiralNameCodes(f.Admiral))
	if f.Flags&rules.FleetFlagMoraleFlip != 0 && g.humanSeat(f.Faction) {
		var b []byte
		b = append(b, ui.Sprintf(p.ReportFormat("morale_prefix"), g.factionNameColoured(f.Faction))...)
		b = append(b, p.FleetNumber(f.Number)...)
		b = append(b, ui.Sprintf(p.ReportFormat("morale_admiral"), name)...)
		if f.Flags&rules.FleetFlagLowMorale != 0 {
			b = append(b, p.ReportFormat("morale_low")...)
		} else {
			b = append(b, p.ReportFormat("morale_recovered")...)
		}
		b = append(b, ui.Sprintf(p.ReportFormat("morale_value"), f.Morale)...)
		out = append(out, turnReport{text: b, face: f.Admiral})
	}
	if f.Flags&rules.FleetFlagNoShips != 0 {
		var b []byte
		b = append(b, ui.Sprintf(p.ReportFormat("no_ships_prefix"), g.factionNameColoured(f.Faction))...)
		b = append(b, p.FleetNumber(f.Number)...)
		b = append(b, ui.Sprintf(p.ReportFormat("no_ships_body"), name)...)
		out = append(out, turnReport{text: b, face: f.Admiral})
	}
	f.Flags &^= rules.FleetFlagMoraleFlip | rules.FleetFlagNoShips
	return out
}

// endOfTurnMoraleReports 是回合尾巴的 `1261:00C8`：32 支依序跑 `1261:0104`。
func (g *Game) endOfTurnMoraleReports() []turnReport {
	var out []turnReport
	for i := range g.fleets {
		out = append(out, g.moraleReports(&g.fleets[i])...)
	}
	return out
}
