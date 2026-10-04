package game

import (
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 星系情報視窗（`073F:05CB` → `060A`，`docs/re/320`、規格 053）。
//
// 戰略地圖介面上左鍵點星圖：`3705:00DC` 找到那一格，**點的是右側面板看的那一格
// （`ds:9588`）就開視窗**，點別的格是換選星域（`073F:02EE`）。右鍵點星圖什麼都
// 不換（收據 `fcmd-map-rclick`）。視窗蓋住整個左半，左鍵只認「結束」、右鍵也關；
// 關掉之後左半重畫。

// rinfoRegion 是視窗看的星域（開著時才有意義）。
func (g *Game) mapRegionAt(x, y int) int {
	if g.res.Regions == nil || x < mapX || x >= mapX+mapW || y < mapY || y >= mapY+mapH {
		return -1
	}
	return asset.RegionAt(g.res.Regions, x-mapX, y-mapY, tileW, tileH)
}

// mapClick 是戰略地圖介面的左鍵落在星圖上。回 true ＝ 吃掉了。
func (g *Game) mapClick(x, y int) bool {
	r := g.mapRegionAt(x, y)
	if r < 0 {
		return false
	}
	if r == g.listRegion() {
		g.openRegionInfo(r)
		return true
	}
	g.selectRegion(r)
	return true
}

// selectRegion 是 `073F:02EE(r)`：右側面板與清單改看星域 r（同一格就不動）——
// 地圖反白換到 r（`3705:0284`）、四個分頁重建（位移歸零）、右側面板那塊摘要
// 重寫成 r 的（`073F:0387`；不是自己的星域治安／成長率／徵稅額／稅率印 `----`）。
// 只有左鍵點別的格走它；右鍵點星圖不動任何東西。
func (g *Game) selectRegion(r int) {
	if r < 0 || r == g.listRegion() {
		return
	}
	g.selectedRegion = r
	g.mapHighlight = r
	for i := range g.tabs.page {
		g.tabs.page[i].Scroll = 0
	}
	g.tabs.pressed = -1
	if rec := g.regionRecord(r); rec != nil && g.res.Planets != nil {
		g.capitalSnap = rules.Summarise(rec, g.res.Planets, r, g.faction)
		g.capitalSnapOK = true
	}
}

// openRegionInfo 開星域 r 的星系情報視窗。
func (g *Game) openRegionInfo(r int) {
	if g.res.RegionInfo == nil || g.regionRecord(r) == nil {
		return
	}
	g.rinfo = r
	g.loadRegionPlanetSites(r)
	g.state = StateRegionInfo
}

// closeRegionInfo 關視窗回戰略地圖介面。
func (g *Game) closeRegionInfo() {
	g.rinfo = -1
	g.state = StateCommandPhase
}

// regionInfoView 把視窗要印的東西從記錄算出來（`073F:06C5`）。
func (g *Game) regionInfoView() *screen.RegionInfoView {
	r := g.rinfo
	rec := g.regionRecord(r)
	if rec == nil || g.res.Planets == nil || g.res.PlanetInfoMap == nil {
		return nil
	}
	owner := int(rec[0x12])
	v := &screen.RegionInfoView{
		Owner:      owner,
		OwnerName:  g.res.PlanetInfoMap.OwnerName(owner),
		RegionName: codesBig5(g.regionName(r)),
		Own:        owner == g.faction,
	}
	if v.Own {
		// 七列是**擁有者陣營**在這一格的待命艘數（`+0x30 ＋ 擁有者 × 0x1C`）與
		// 艘數 ÷ 滿編（`3CC9:0020`）；飛彈／艦載機是星域層那兩格。
		ships := rules.RegionShips(rec, owner)
		for k := range v.Ships {
			if k < len(ships) {
				v.Ships[k] = ships[k]
				if kind, ok := g.res.ShipKinds.Kind(owner, k); ok && kind.FullShips > 0 {
					v.Units[k] = ships[k] / kind.FullShips
				}
			}
		}
		if pool, ok := rules.ReadRegionSupply(rec); ok {
			v.Missiles, v.Fighters = pool.Missiles[owner], pool.Fighters[owner]
		}
	}
	// 行星位置與小圖都從這個星域的戰場地圖來（與指定待命行星那張同一套）。
	var sites []rules.PlanetSite
	if cells := g.battleCells(r); len(cells) > 0 && g.res.Battle != nil {
		v.Cells, v.MapCols, v.MapRows = cells, g.res.Battle.Field.MapCols, g.res.Battle.Field.MapRows
		sites = rules.PlanetSites(cells, g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
	}
	start, n := g.res.Planets.Range(r)
	for k := 0; k < n; k++ {
		p := screen.RegionInfoPlanet{
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

// drawRegionInfo 把視窗疊在戰略地圖介面上（左半整塊）。
func (g *Game) drawRegionInfo(fb *gfx.Framebuffer) {
	if g.res.RegionInfo == nil || g.rinfo < 0 {
		return
	}
	if v := g.regionInfoView(); v != nil {
		g.res.RegionInfo.Draw(fb, g.res.Font, g.res.FrameTiles, v)
	}
}

// occupyAskView 湊齊回合裡「要占領此星系嗎？」那張問話要印的值（`113A:0899`）。
func (g *Game) occupyAskView() (*screen.OccupyAskView, *asset.Image) {
	i := g.occupyAsk.fleet
	if i < 0 || i >= len(g.fleets) || g.res.FleetWindow == nil {
		return nil, nil
	}
	f := &g.fleets[i]
	adm := g.admiralRec(f.Admiral)
	v := &screen.OccupyAskView{
		FactionName: decodeHexBytes(g.res.FleetWindow.FactionNamesColoured[f.Faction]),
		Number:      f.Number,
		AdmiralName: codesBig5(g.admiralNameCodes(f.Admiral)),
		Rank:        codesBig5(g.admiralRankCodes(adm)),
		TargetName:  codesBig5(g.regionName(f.Target)),
		NextName:    codesBig5(g.regionName(g.occupyAsk.next)),
	}
	var face *asset.Image
	if g.res.Faces != nil {
		if img, err := g.res.Faces.Decode(f.Admiral); err == nil {
			face = img
		}
	}
	return v, face
}

// drawOccupyAsk 把問話疊在戰略地圖介面上。
func (g *Game) drawOccupyAsk(fb *gfx.Framebuffer) {
	if g.res.OccupyAsk == nil {
		return
	}
	if v, face := g.occupyAskView(); v != nil {
		g.res.OccupyAsk.Draw(fb, g.res.Font, g.res.FrameTiles, v, face)
	}
}
