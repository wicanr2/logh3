package game

import (
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 戰略地圖介面右下角清單的四個分頁與捲動（`docs/re/319`、`docs/spec/052`）。
//
// 原版：分頁鈕是熱區 `0BEE[1..4]`，點了把 `ds:8F1F`（也抄一份到
// `ds:8F29[座位]`）設成 0–3、叫 `04F8:05ED` 換掉列印與點擊兩個函式指標、
// 重建那一頁的清單（捲動位移歸零）、四顆分頁鈕的核取方塊換色（選中 0x27、
// 其餘 0x97）、印整頁。每個分頁都是「筆數／捲動位移／五列一頁」同一套；
// 欄17的▲▼筆數超過五才印，每次一列（`3765:0255`）；保持速度依089。

// listTabFleet 等是分頁編號（`ds:8F1F`）。
const (
	listTabFleet = iota
	listTabPlanet
	listTabEnemy
	listTabAdmiral
	listTabCount
)

// listTabs 是分頁與清單的執行期狀態。
type listTabs struct {
	// tab 每個陣營各一份（`ds:8F29[座位]`）；目前那一份就是 `ds:8F1F`。
	tab [2]int
	// page 是四個分頁各自的筆數與捲動位移。筆數每次畫之前重算（remake 的
	// 清單是從狀態算的，原版是換分頁／關窗時重建），位移照原版夾進範圍。
	page [listTabCount]rules.ListPage
	// detail 是「提督」分頁點開的詳細資料（提督索引；−1 沒開），
	// detailSortie 是那一張有沒有「出擊」（`ds:8F49[陣營] < 16` 且首都還在）。
	detail       int
	detailSortie bool
	// planet 是「行星」分頁點開的行星情報視窗：那一格裡的第幾顆（−1 沒開）。
	planet       int
	planetRegion int
	// enemy 是「敵艦隊」分頁點開的視窗：那一支在 `g.fleets` 裡的位置
	// （−1 沒開）；enemyMsg ＝ 點的那一列看不清楚、跳的是訊息框，關掉時
	// 要把列的反白一起還原。
	enemy    int
	enemyMsg bool
	// pressed 是點開視窗時反白的那一列（−1 沒有）。
	pressed int
}

func newListTabs() listTabs {
	return listTabs{detail: -1, planet: -1, enemy: -1, pressed: -1}
}

// listTab 回目前這個陣營選著的分頁。
func (g *Game) listTab() int {
	if g.faction < 0 || g.faction >= len(g.tabs.tab) {
		return listTabFleet
	}
	return g.tabs.tab[g.faction]
}

// listRegion 回清單看的那一格（原版 `ds:9588`）：左鍵點星圖換選過的星域，
// 沒有就自己的首都（與「艦隊」分頁的 `fleetRows` 同一條規則）。
func (g *Game) listRegion() int {
	if g.selectedRegion >= 0 {
		return g.selectedRegion
	}
	return rules.CapitalRegion(g.faction)
}

// listTabAt 回畫面座標落在哪一顆分頁鈕上；沒有回 −1。熱區是 `0BEE[1..4]`：
// 視窗 9518 的列 6，欄 0–2／4–6／8–11／13–15（與畫出來的鈕同寬）。
func (g *Game) listTabAt(x, y int) int {
	p := g.res.CapitalPanel
	if p == nil || p.Tabs == nil {
		return -1
	}
	t := p.Tabs
	if y < t.Y || y >= t.Y+t.H {
		return -1
	}
	for i, it := range t.Items {
		if x >= it.X && x < it.X+it.W {
			return i
		}
	}
	return -1
}

// selectListTab 是點了分頁鈕（`04F8:05ED`）：換分頁、那一頁的位移歸零。
func (g *Game) selectListTab(tab int) {
	if tab < 0 || tab >= listTabCount || g.faction < 0 || g.faction >= len(g.tabs.tab) {
		return
	}
	g.tabs.tab[g.faction] = tab
	g.tabs.page[tab].Scroll = 0
	g.tabs.pressed = -1
}

// listRefresh 是原版 `073F:0522`：「艦隊」與「提督」兩頁重建（位移歸零），
// 目前那一頁重印。「變更 → 副官」關窗、出擊那條路建完艦隊都走它。
func (g *Game) listRefresh() {
	g.tabs.page[listTabFleet].Scroll = 0
	g.tabs.page[listTabAdmiral].Scroll = 0
	g.tabs.pressed = -1
}

// listPage 回目前分頁夾好的頁（筆數重算、位移夾進範圍）。
func (g *Game) listPage(tab int) rules.ListPage {
	p := &g.tabs.page[tab]
	switch tab {
	case listTabFleet:
		p.Count = len(g.fleetRows())
	case listTabPlanet:
		_, n := g.res.Planets.Range(g.listRegion())
		p.Count = n
	case listTabEnemy:
		p.Count = len(g.enemyFleetsInRegion())
	case listTabAdmiral:
		p.Count = len(g.idleAdmiralsInRegion())
	}
	p.Clamp()
	return *p
}

// listScroll 是 ▲（−1）／▼（＋1）的一列更新。
func (g *Game) listScroll(dir int) {
	tab := g.listTab()
	g.listPage(tab)
	g.tabs.page[tab].Step(dir)
}

// idleAdmiralsInRegion 回「提督」分頁列的人（`28E0:0103(星域, 0x25)`）。
func (g *Game) idleAdmiralsInRegion() []int {
	return rules.IdleAdmiralsInRegion(rules.ParseAdmirals(g.res.Admiral),
		g.faction, g.listRegion(), g.adjBusy(-1))
}

// enemyFleetsInRegion 回「敵艦隊」分頁列的艦隊（在 `g.fleets` 裡的位置，
// `0B20:0006` → `0B20:003B`）：別的陣營、在這一格、而且索敵等級（艦艇
// `+0x17` 在各組取最大）不是 0。索敵等級是戰場上累積、戰後留在艦艇池裡的
// （`docs/internals/38`）；新編的艦隊沒有艦艇池，永遠 0（收據 `fcmd-tab-enemy`
// 空清單），打過仗或改了記憶體的局面才有列（收據 `enm-*`，`docs/re/325`）。
func (g *Game) enemyFleetsInRegion() []int {
	return rules.EnemyFleetsInRegion(g.fleets, g.faction, g.listRegion())
}

// enemyRows 回「敵艦隊」分頁這一頁的列（`0B20:059B`）。
func (g *Game) enemyRows(page rules.ListPage) []screen.EnemyRow {
	list := g.enemyFleetsInRegion()
	rows := make([]screen.EnemyRow, 0, rules.ListPageRows)
	for i := page.Scroll; i < len(list) && len(rows) < rules.ListPageRows; i++ {
		f := &g.fleets[list[i]]
		r := screen.EnemyRow{Known: f.Spotting() >= rules.SpottingClear,
			Faction: f.Faction, Number: f.Number, Morale: f.Morale}
		if f.Admiral >= 0 && f.Admiral < len(g.res.AdmiralNames) {
			r.Name = g.res.AdmiralNames[f.Admiral]
		}
		rows = append(rows, r)
	}
	return rows
}

// openEnemyWindow 是點了「敵艦隊」分頁的一列（`0B20:02D0` → `0B20:0336`）：
// 那一列反白；看得清楚（索敵等級 3）開視窗，否則跳「情報資料來源不足」。
func (g *Game) openEnemyWindow(idx, row int) {
	list := g.enemyFleetsInRegion()
	if idx < 0 || idx >= len(list) {
		return
	}
	g.tabs.pressed = row
	if g.fleets[list[idx]].Spotting() >= rules.SpottingClear {
		g.tabs.enemy = list[idx]
		return
	}
	name := "enemy_no_intel"
	if s := g.res.EnemyWindow; s != nil && s.NoIntelMessage != "" {
		name = s.NoIntelMessage
	}
	if g.showMessage(name) {
		g.tabs.enemyMsg = true
	} else {
		g.tabs.pressed = -1
	}
}

// enemyWindowView 把開著的那一支排成視窗要印的東西；頭像另外貼。
func (g *Game) enemyWindowView() (*screen.EnemyWindowView, *asset.Image) {
	if g.tabs.enemy < 0 || g.tabs.enemy >= len(g.fleets) {
		return nil, nil
	}
	f := &g.fleets[g.tabs.enemy]
	v := &screen.EnemyWindowView{Faction: f.Faction, Number: f.Number}
	if f.Admiral >= 0 && f.Admiral < len(g.res.AdmiralNames) {
		v.Name = codesBig5(g.res.AdmiralNames[f.Admiral])
	}
	if adm := g.admiralRecord(f.Admiral); adm != nil {
		v.Rank = codesBig5(g.rankCodes(rules.RankOf(int(adm[admiralMeritOff]))))
	}
	for _, gp := range f.Composition {
		v.Groups = append(v.Groups, screen.EnemyGroup{Kind: gp.Kind, Detail: gp.Detail})
	}
	var face *asset.Image
	if g.res.Faces != nil {
		if img, err := g.res.Faces.Decode(f.Admiral); err == nil {
			face = img
		}
	}
	return v, face
}

// listClick 處理戰略地圖介面上落在右下角清單那塊的左鍵：分頁鈕、▲▼、
// 一列。回 true ＝ 吃掉了。
func (g *Game) listClick(x, y int) bool {
	if tab := g.listTabAt(x, y); tab >= 0 {
		g.selectListTab(tab)
		return true
	}
	if g.res.FleetList == nil {
		return false
	}
	tab := g.listTab()
	if dir := g.res.FleetList.ScrollAt(x, y); dir != 0 {
		// 「行星」分頁沒有 ▲▼，它的捲動熱區在列 6 與 11、而且方向相反
		// （`0B99:0220`）；行星最多五顆（`PLADAT.DAT` 檔頭），夾回去永遠是 0，
		// 畫面上看不出來，所以這裡一律不理。
		if tab != listTabPlanet {
			g.listScroll(dir)
		}
		return true
	}
	row := fleetListRowAt(x, y)
	if row < 0 {
		return false
	}
	page := g.listPage(tab)
	idx := page.Scroll + row
	if idx >= page.Count {
		return true
	}
	switch tab {
	case listTabFleet:
		if i := g.fleetListFleetAt(idx); i >= 0 {
			g.openFleetWindow(i)
		}
	case listTabPlanet:
		// `2AF5:0226(星域, 行星)`：那一列反白、開行星情報視窗。
		g.tabs.planet, g.tabs.planetRegion, g.tabs.pressed = idx, g.listRegion(), row
	case listTabAdmiral:
		// `0BCE:02F4`：那一列反白、開提督詳細資料（有沒有「出擊」看艦隊數
		// 與首都）。
		list := g.idleAdmiralsInRegion()
		if idx < len(list) {
			g.tabs.detail, g.tabs.pressed = list[idx], row
			g.tabs.detailSortie = g.sortieAllowed()
		}
	case listTabEnemy:
		g.openEnemyWindow(idx, row)
	}
	return true
}

// sortieAllowed 是提督分頁點開詳細資料時「出擊」鈕有沒有（`0BCE:02F4`）：
// 首都還是自己的、而且艦隊數（`ds:8F49[陣營]`）不到 16。
func (g *Game) sortieAllowed() bool {
	capital := rules.CapitalRegion(g.faction)
	if rec := g.regionRecord(capital); rec == nil || int(rec[0x12]) != g.faction {
		return false
	}
	n := 0
	for _, f := range g.fleets {
		if f.Faction == g.faction {
			n++
		}
	}
	return n < fleetsPerFaction
}

// fleetsPerFaction 是艦隊表每邊的槽數（`ds:8F49[陣營] < 16`，`docs/internals/33`）。
const fleetsPerFaction = 16

// listOverlayOn 回清單點出來的視窗（行星情報／提督詳細資料／敵艦隊）
// 或「情報資料來源不足」那個訊息框有沒有開著。
func (g *Game) listOverlayOn() bool {
	return g.tabs.planet >= 0 || g.tabs.detail >= 0 || g.tabs.enemy >= 0 || g.tabs.enemyMsg
}

// listOverlayClick 處理那幾張視窗開著時的左鍵。回 true ＝ 視窗吃掉了。
func (g *Game) listOverlayClick(x, y int) bool {
	switch {
	case g.tabs.enemyMsg:
		// 訊息框是模態的，只認「確認」（`3826:023B`）；關了那一列的反白還原。
		g.messageClick(x, y)
		if !g.msgOn {
			g.tabs.enemyMsg, g.tabs.pressed = false, -1
		}
		return true
	case g.tabs.enemy >= 0:
		// 左鍵只認「結束」（`0B20:0336` 的事件迴圈）。
		if s := g.res.EnemyWindow; s != nil && s.ButtonAt(x, y) == "close" {
			g.closeListOverlay()
		}
		return true
	case g.tabs.planet >= 0:
		// 左鍵只認「確認」（`2AF5:0226` 的事件迴圈）。
		if s := g.res.PlanetInfoMap; s != nil && s.ButtonAt(x, y) {
			g.closeListOverlay()
		}
		return true
	case g.tabs.detail >= 0:
		if g.res.AdmiralPanel == nil {
			return true
		}
		switch g.tabDetailButtonAt(x, y) {
		case "bio":
			if g.commandSpec(StateAdmiralBio) != nil {
				g.detailAdmiral = g.tabs.detail
				g.state = StateAdmiralBio
			}
		case "cancel", "close":
			g.closeListOverlay()
		case "sortie":
			// `07F9:00D5`：與出擊命令那條路同一支——進艦隊編成就配槽。
			if g.commandSpec(StateFleetOrg) != nil {
				g.detailAdmiral = g.tabs.detail
				g.closeListOverlay()
				g.orgSlot = -1 // 每次出擊重來（#65）
				g.sortieSlots(true)
				g.state = StateFleetOrg
			}
		}
		return true
	}
	return false
}

// tabDetailButtonAt 回提督分頁那張詳細資料的按鈕：有「出擊」時是
// 經歷／中止／出擊（熱區 `1F34`）；沒有時只剩 經歷／結束，而且往右挪到
// 中止與出擊的位置（`1FAE`）。
func (g *Game) tabDetailButtonAt(x, y int) string {
	return g.tabDetailButtonAtFor(x, y, g.tabs.detailSortie)
}

// tabDetailButtonAtFor 是 tabDetailButtonAt 給一覽表那兩張用的版本。
func (g *Game) tabDetailButtonAtFor(x, y int, sortie bool) string {
	act := g.res.AdmiralPanel.ButtonAt(x, y)
	if sortie {
		return act
	}
	switch act {
	case "cancel":
		return "bio"
	case "sortie":
		return "close"
	}
	return ""
}

// listOverlayRightClick 是那幾張視窗的右鍵：都只是關掉；訊息框不理右鍵。
func (g *Game) listOverlayRightClick() bool {
	if !g.listOverlayOn() {
		return false
	}
	if !g.tabs.enemyMsg {
		g.closeListOverlay()
	}
	return true
}

// closeListOverlay 關掉行星情報／提督詳細資料／敵艦隊視窗，那一列的反白還原。
func (g *Game) closeListOverlay() {
	g.tabs.planet, g.tabs.detail, g.tabs.enemy, g.tabs.pressed = -1, -1, -1, -1
}

// drawListRows 畫右下角清單目前分頁的幾列與 ▲▼（畫在底板那一層）。
// 「艦隊」分頁交給 `drawFleetRows`。
func (g *Game) drawListRows(fb *gfx.Framebuffer) {
	l := g.res.FleetList
	if l == nil || l.Tabs == nil || !g.listShown() {
		return
	}
	tab := g.listTab()
	page := g.listPage(tab)
	if tab != listTabPlanet {
		l.DrawScroll(fb, g.res.FrameTiles, page.Arrows())
	}
	switch tab {
	case listTabPlanet:
		start, n := g.res.Planets.Range(g.listRegion())
		rows := make([]screen.PlanetRow, 0, rules.ListPageRows)
		for i := page.Scroll; i < n && len(rows) < rules.ListPageRows; i++ {
			rec := g.res.Planets.Record(start + i)
			if rec == nil {
				break
			}
			rows = append(rows, screen.PlanetRow{
				Owner:    g.res.Planets.Owner(start + i),
				Fortress: rec[0x12] != 0,
				Name:     g.res.Planets.Name(start + i),
				Defence:  g.res.Planets.Defence(start + i),
				Max:      g.res.Planets.GrowthDefence(start + i),
			})
		}
		l.DrawPlanets(fb, g.res.Font, rows, g.tabs.pressed)
	case listTabAdmiral:
		list := g.idleAdmiralsInRegion()
		rows := make([]screen.AdmiralRow, 0, rules.ListPageRows)
		for i := page.Scroll; i < len(list) && len(rows) < rules.ListPageRows; i++ {
			a := list[i]
			var name []uint16
			if a < len(g.res.AdmiralNames) {
				name = g.res.AdmiralNames[a]
			}
			rows = append(rows, screen.AdmiralRow{Name: name,
				Rank: g.rankCodes(rules.RankOf(rules.AdmiralLevel(g.admiralRecord(a))))})
		}
		l.DrawAdmirals(fb, g.res.Font, rows, g.tabs.pressed)
	case listTabEnemy:
		l.DrawEnemies(fb, g.res.Font, g.enemyRows(page), g.tabs.pressed)
	}
}

// listShown 回右下角清單這一刻有沒有印出來：戰略地圖介面與疊在它上面的
// 艦隊命令、出擊那幾張。月初報告那幾張右側面板還是空的（`DrawOpening`），
// 分頁鈕都沒畫，▲▼ 那兩格自然也沒有。
func (g *Game) listShown() bool {
	switch g.state {
	case StateCommandPhase, StateFleetWindow, StateFleetActionMenu, StateFleetTargetPick,
		StateFleetMessage, StateFleetConfirm, StateFleetStandby, StateFleetOccupyWarn,
		StateFleetChangeMenu, StateFleetChangeAdj, StateFleetChangeOrg,
		StateFleetOrg, StateFleetNumber, StateFleetAdjutant, StateFleetComposition, StateFormation:
		return true
	}
	return false
}

// drawListOverlays 畫清單點出來的視窗（在艦隊命令那幾層之後、游標之前）。
func (g *Game) drawListOverlays(fb *gfx.Framebuffer) {
	switch {
	case g.tabs.planet >= 0:
		g.drawMapPlanetInfo(fb)
	case g.tabs.detail >= 0:
		g.drawTabAdmiralDetail(fb)
	case g.tabs.enemy >= 0:
		if v, face := g.enemyWindowView(); v != nil && g.res.EnemyWindow != nil {
			g.res.EnemyWindow.Draw(fb, g.res.Font, g.res.FrameTiles, v, face)
		}
	}
}

// drawMapPlanetInfo 畫戰略地圖那張行星情報（`2AF5:02C3`）。
func (g *Game) drawMapPlanetInfo(fb *gfx.Framebuffer) {
	s := g.res.PlanetInfoMap
	if s == nil || g.res.Planets == nil || g.res.FrameTiles == nil {
		return
	}
	region := g.tabs.planetRegion
	start, n := g.res.Planets.Range(region)
	if g.tabs.planet >= n {
		return
	}
	// 開窗前先把這一格星域層的存量分回行星（`2AF5:0226` 一進來就叫
	// `2703:0540(星域)`，兩個陣營各一趟）。
	if rec := g.regionRecord(region); rec != nil {
		for faction := 0; faction < 2; faction++ {
			rules.DistributeRegionSupply(rec, g.res.Planets, region, faction)
		}
	}
	i := start + g.tabs.planet
	rec := g.res.Planets.Record(i)
	if rec == nil {
		return
	}
	v := screen.PlanetInfo{
		RegionOwner: g.regionOwnerOf(region),
		RegionName:  g.regionNameBytesOf(region),
		PlanetOwner: g.res.Planets.Owner(i),
		PlanetName:  planetNameBytes(rec),
		Fortress:    rec[0x12] != 0,
		Own:         g.res.Planets.Owner(i) == g.faction,
		Economy:     int64(g.res.Planets.PlanetEconomy(i)),
		Defence:     g.res.Planets.Defence(i),
		DefenceMax:  g.res.Planets.GrowthDefence(i),
		Recover:     int(rec[0x18]),
	}
	// 成長率 `2703:0D47(行星)` ＝ 經濟力查階梯表（防衛力上限不到門檻再加成），
	// 除 10 取商與餘。
	if g.res.Growth != nil {
		r := g.res.Growth.Rate(int(v.Economy), v.DefenceMax)
		v.GrowthInt, v.GrowthF = r/10, r%10
	}
	// 艦種那一列印「補充存量 ÷ 那個陣營的滿編艘數」（`3CC9:0020` 長除法，
	// 除數是艦種表 `ds:4E3A` 的 word 0）。
	if stock, ok := rules.ReadPlanetStock(rec); ok {
		v.Missiles, v.Fighters = int64(stock.Missiles), int64(stock.Fighters)
		for k := range v.Kinds {
			if kind, ok := g.res.ShipKinds.Kind(g.faction, k); ok && kind.FullShips > 0 {
				v.Kinds[k] = stock.Ships[k] / kind.FullShips
			}
		}
	}
	screen.DrawPlanetInfo(fb, s, g.res.Font, g.res.FrameTiles, v)
}

// planetNameBytes 回行星記錄開頭那個名字（Big5，`00` 結尾）。
func planetNameBytes(rec []byte) []byte {
	for i := 0; i+1 < len(rec) && i < 0x11; i += 2 {
		if rec[i] == 0 {
			return rec[:i]
		}
	}
	return rec[:0x10]
}

// drawTabAdmiralDetail 畫提督分頁點開的詳細資料：與副官那條路同一張
// overlay；沒有「出擊」時只剩 經歷／結束 兩顆、往右挪（熱區 `1FAE`）。
func (g *Game) drawTabAdmiralDetail(fb *gfx.Framebuffer) {
	g.drawAdmiralDetailOverlay(fb, g.tabs.detail, g.tabs.detailSortie, true)
}

// 「結束」兩個字（Big5 B5B2 A7F4）。
var adminDetailCloseCodes = []uint16{0xB5B2, 0xA7F4}
