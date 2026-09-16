package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 「指定待命行星」（`0DEA:1320`，`docs/re/314`、規格 `docs/spec/051` 待命那一節）。
//
// 作戰選單第 4 項「待命」不走目的地迴圈，而是開一張自己的視窗：這個星域
// 的行星一顆一顆鈕，點了反白，「決定」（或右鍵）把反白那一顆寫進艦隊記錄
// `+0x07`、`+0x08` 留 5，關艦隊視窗；「中止」清掉反白回艦隊視窗。
// 沒反白就按「決定」也只是回艦隊視窗（`1320` 回 0xFF，不算「有改」）。

// openFleetStandby 開視窗。已經有待命行星的（`+0x08`＝5 且 `+0x07`≠0xFF）
// 那一顆一開就反白（`0DEA:14D2` 結尾）。
func (g *Game) openFleetStandby() {
	if g.res.FleetStandby == nil {
		g.state = StateFleetWindow
		return
	}
	f := &g.fleets[g.fcmd.fleet]
	g.fcmd.standbySel = -1
	g.fcmd.standbyRepainted = [5]bool{}
	if f.Move == fleetMoveParked && f.Target >= 0 {
		g.fcmd.standbySel = f.Target
	}
	g.state = StateFleetStandby
}

// fleetStandbyClick 是視窗開著時的左鍵（`0DEA:16AB`）：行星鈕換反白、
// 「決定」寫命令、「中止」清反白；點在別處什麼都不做。
func (g *Game) fleetStandbyClick(x, y int) {
	_, n := g.standbyPlanets()
	switch code := g.res.FleetStandby.HotspotAt(x, y, n); code {
	case screen.StandbyHotspotNone:
	case screen.StandbyHotspotDecide:
		g.finishFleetStandby()
	case screen.StandbyHotspotCancel:
		g.fcmd.standbySel = -1
		g.finishFleetStandby()
	default:
		// 舊反白改回 0x87、新的改成 0x27（`37A9:02B1` 兩次）——兩顆都算
		// 「換過顏色」，重畫時蓋在引線上。
		if prev := g.fcmd.standbySel; prev >= 0 && prev < len(g.fcmd.standbyRepainted) {
			g.fcmd.standbyRepainted[prev] = true
		}
		if code < len(g.fcmd.standbyRepainted) {
			g.fcmd.standbyRepainted[code] = true
		}
		g.fcmd.standbySel = code
	}
}

// finishFleetStandby 是迴圈結束那一段（`0DEA:1037` 區塊 C）：有反白的就
// `+0x08 ← 5`、`+0x07 ← 那一顆`，算「有改」關視窗；沒有就只回艦隊視窗。
// 兩條路都以 `090A:0E6D` 重算未安定度收尾。
//
// 原版另外把行星的位置抄進艦隊 `+0x5C`／`+0x5E`（`0DEA:1759`）；remake 的
// 艦隊記錄沒有戰場位置欄位，這一筆不模型化（規格的已知差異）。
func (g *Game) finishFleetStandby() {
	f := &g.fleets[g.fcmd.fleet]
	sel := g.fcmd.standbySel
	g.fcmd.standbySel = -1
	if sel < 0 {
		g.recomputeUnrest(g.turn)
		g.state = StateFleetWindow
		return
	}
	f.Target, f.Move, f.Parked = sel, fleetMoveParked, true
	g.recomputeUnrest(g.turn)
	g.closeFleetWindow()
}

// standbyPlanets 回艦隊所在星域的行星在行星表裡的起訖（`2AF5:0036(+0x06)`
// 之後的 `ds:CFF8`／`ds:CFFC`）。
func (g *Game) standbyPlanets() (start, n int) {
	if g.res.Planets == nil || g.fcmd.fleet < 0 || g.fcmd.fleet >= len(g.fleets) {
		return 0, 0
	}
	return g.res.Planets.Range(g.fleets[g.fcmd.fleet].Region)
}

// standbyPlanetIndex 回這一支待命的那顆行星在行星表裡的序號（−1 ＝ 沒有）。
func (g *Game) standbyPlanetIndex(f *rules.Fleet) int {
	if g.res.Planets == nil || f.Move != fleetMoveParked || f.Target < 0 {
		return -1
	}
	start, n := g.res.Planets.Range(f.Region)
	if f.Target >= n {
		return -1
	}
	return start + f.Target
}

// fleetStandbyView 把視窗要印的值湊齊。
func (g *Game) fleetStandbyView() *screen.FleetStandbyView {
	i := g.fcmd.fleet
	if i < 0 || i >= len(g.fleets) || g.res.FleetWindow == nil {
		return nil
	}
	f := &g.fleets[i]
	v := &screen.FleetStandbyView{
		FactionName: decodeHexBytes(g.res.FleetWindow.FactionNamesColoured[f.Faction]),
		Number:      f.Number,
		AdmiralName: codesBig5(g.admiralNameCodes(f.Admiral)),
		RegionName:  codesBig5(g.regionName(f.Region)),
		Selected:    g.fcmd.standbySel,
		Repainted:   g.fcmd.standbyRepainted[:],
	}
	// 行星位置是掃這個星域的戰場地圖算的（`rules.PlanetSites`，`docs/re/89`），
	// 小圖也從同一張地圖縮出來。
	var sites []rules.PlanetSite
	if cells := g.battleCells(f.Region); len(cells) > 0 && g.res.Battle != nil {
		v.Cells, v.MapCols, v.MapRows = cells, g.res.Battle.Field.MapCols, g.res.Battle.Field.MapRows
		sites = rules.PlanetSites(cells, g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
	}
	start, n := g.standbyPlanets()
	for k := 0; k < n; k++ {
		p := screen.FleetStandbyPlanet{
			Name:  codesBig5(g.res.Planets.Name(start + k)),
			Owner: g.res.Planets.Owner(start + k),
		}
		if k < len(sites) {
			p.X, p.Y = sites[k].X, sites[k].Y
		}
		v.Planets = append(v.Planets, p)
	}
	return v
}

// drawFleetStandby 把視窗疊在艦隊視窗上。
func (g *Game) drawFleetStandby(fb *gfx.Framebuffer) {
	if g.res.FleetStandby == nil {
		return
	}
	if v := g.fleetStandbyView(); v != nil {
		g.res.FleetStandby.Draw(fb, g.res.Font, g.res.FrameTiles, v)
	}
}
