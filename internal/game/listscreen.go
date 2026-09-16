package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 一覽表四張的列（`docs/re/323`、規格 055）。
//
// 四張都是同一種「清單迴圈」：星系資料 `0D05:00D7`、艦隊資料 `0C7F:00DE`、
// 提督／敵將資料 `0C0A:00F0`。左鍵點列開那一張視窗——星系 → 換選那一格
// （`073F:02EE`）＋ 關清單 ＋ 星系情報視窗（`073F:05CB`）；艦隊 → 艦隊視窗
// （`0DEA:0003`，與右下角清單同一支，關了回清單）；提督／敵將 → 提督詳細
// 資料（`0D92:000D(r, 0, 自軍, 0)`：經歷／結束，關了回清單）。右鍵關清單
// （子視窗開著就只關子視窗）。▲▼ 一下一列（原版按住連捲）。游標停在列上
// 換地圖反白與左下角摘要（`3705:0591`）：星系那張只在列上才換、離開不動；
// 艦隊與提督那兩張離開列就取消反白、左下角清空；敵將那張不看游標。
//
// 一覽表的分頁鈕還沒接（#51）：星系資料的 艦隊／配屬／稅金 與艦隊資料的
// 士氣／部隊／作戰 畫面上是量死的第一頁，點了沒事。

type listScreen struct {
	// starScroll 是星系資料的位移，每座位記住（`ds:1DF2[座位]`，關了再開還在）。
	starScroll [2]int
	// fleetScroll／admScroll 是艦隊資料與提督／敵將資料的位移（`ds:96C3`／
	// `ds:96A1`），開清單歸零。
	fleetScroll, admScroll int
	// hl 是清單開著時地圖反白的那一格（`ds:F182`），−1 沒有；hlDrawn ＝
	// 開清單之後地圖重畫過沒有——`3705:0585` 只把反白從緩衝區拿掉、沒有
	// 重畫，要到第一次 `0591` 才畫上去。
	hl      int
	hlDrawn bool
	// pressed 是子視窗開著時反白的那一列，−1 沒有。
	pressed int
	// detail 是提督／敵將資料點開的詳細資料（提督索引），−1 沒開；
	// detailOwn ＝ `0D92:000D` 的第三個參數（印現在位置與狀態，敵將印 ＊）。
	detail    int
	detailOwn bool
	bio       bool
	// fleetFrom ＝ 艦隊視窗是從艦隊資料開的，關了回清單而不是回地圖。
	fleetFrom bool
}

func newListScreen() listScreen {
	return listScreen{hl: -1, pressed: -1, detail: -1}
}

// isListScreen 回這個狀態是不是一覽表的四張之一。
func isListScreen(s State) bool {
	switch s {
	case StateStarList, StateFleetList, StateAdmiralListView, StateEnemyList:
		return true
	}
	return false
}

// openListScreen 開一覽表的一張（一覽表選單點了項目）。
func (g *Game) openListScreen(s State) {
	g.lst.fleetScroll, g.lst.admScroll = 0, 0
	g.lst.hl, g.lst.hlDrawn = -1, false
	g.lst.pressed, g.lst.detail, g.lst.bio, g.lst.fleetFrom = -1, -1, false, false
	g.state = s
	g.listScreenHover()
}

// closeListScreen 關清單回地圖介面。
func (g *Game) closeListScreen() {
	g.lst.pressed, g.lst.detail, g.lst.bio, g.lst.fleetFrom = -1, -1, false, false
	g.state = StateCommandPhase
}

// listScreenScroll 回目前這張的位移。
func (g *Game) listScreenScroll() int {
	switch g.state {
	case StateStarList:
		if g.faction >= 0 && g.faction < len(g.lst.starScroll) {
			return g.lst.starScroll[g.faction]
		}
	case StateFleetList:
		return g.lst.fleetScroll
	case StateAdmiralListView, StateEnemyList:
		return g.lst.admScroll
	}
	return 0
}

func (g *Game) setListScreenScroll(v int) {
	switch g.state {
	case StateStarList:
		if g.faction >= 0 && g.faction < len(g.lst.starScroll) {
			g.lst.starScroll[g.faction] = v
		}
	case StateFleetList:
		g.lst.fleetScroll = v
	case StateAdmiralListView, StateEnemyList:
		g.lst.admScroll = v
	}
}

// listScreenFleets 回艦隊資料那張列的艦隊（自軍全部，照艦隊表順序；
// `297B:131C(清單, 0xFF, 1)`）。
func (g *Game) listScreenFleets() []int {
	var out []int
	for i, f := range g.fleets {
		if f.Faction == g.faction {
			out = append(out, i)
		}
	}
	return out
}

// listScreenCount 回目前這張的筆數。
func (g *Game) listScreenCount() int {
	switch g.state {
	case StateStarList:
		return len(g.starListOrder())
	case StateFleetList:
		return len(g.listScreenFleets())
	case StateAdmiralListView, StateEnemyList:
		if spec := g.commandSpec(g.state); spec != nil {
			g.setListRows(spec)
		}
		return g.listScreenAdmiralCount
	}
	return 0
}

// listScreenStep 是 ▲（−1）／▼（＋1）按一下：一列，夾在 `0 … 筆數 − 16`
// （`0D05:01CD`／`0225`、`0C7F:01D4`／`022C`、`0C0A:01D7`／`0235`）。
func (g *Game) listScreenStep(dir int) {
	spec := g.res.ListScreen
	if spec == nil {
		return
	}
	n := g.listScreenCount()
	v := g.listScreenScroll() + dir
	if max := n - spec.Rows; v > max {
		v = max
	}
	if v < 0 {
		v = 0
	}
	g.setListScreenScroll(v)
}

// listScreenRowRegion 回游標在第 row 列時該反白的星域（−1 ＝ 取消反白；
// −2 ＝ 不動）。
func (g *Game) listScreenRowRegion(row int) int {
	idx := g.listScreenScroll() + row
	switch g.state {
	case StateStarList:
		// `0D05:0388`：只有命中列才叫 `0591`，離開列不動。
		if order := g.starListOrder(); row >= 0 && idx < len(order) {
			return order[idx]
		}
		return -2
	case StateFleetList:
		// `0C7F:03CB`：命中清單的任何熱區就算列，沒有艦隊的列取消反白；
		// 清單外不動。
		if fl := g.listScreenFleets(); row >= 0 && idx < len(fl) {
			return g.fleets[fl[idx]].Region
		}
		return -1
	case StateAdmiralListView:
		// `0C0A:0351`（第三個參數 1）：列上是那位提督的所在星域，其餘
		// （含清單外）取消反白。`listIndex` 是切掉位移之後的那十六列。
		if row >= 0 && row < len(g.listIndex) {
			return rules.AdmiralRegion(g.admiralRecord(g.listIndex[row]))
		}
		return -1
	}
	return -2
}

// listScreenHover 是游標動了之後清單那一圈做的事（`3705:0591`）。
func (g *Game) listScreenHover() {
	spec := g.res.ListScreen
	if spec == nil || !isListScreen(g.state) || g.lst.detail >= 0 {
		return
	}
	if g.state == StateEnemyList {
		return
	}
	if spec := g.commandSpec(g.state); spec != nil {
		g.setListRows(spec)
	}
	x, y := g.mouseX, g.mouseY
	row := spec.RowAt(x, y)
	if g.state != StateStarList {
		// 艦隊與提督那兩張的游標處理不看命中的是哪一塊，命中就用
		// `341B:044A` 換算列：▲ 那一格算列 0、▼ 算列 15、分頁那一列算 −1
		// （不是列 → 取消反白）。星系那張只認內容區（`0D05:0388`）。
		switch {
		case spec.Up.Contains(x, y):
			row = 0
		case spec.Down.Contains(x, y):
			row = spec.Rows - 1
		case spec.TabAt(x, y) >= 0:
			row = -1
		case g.state == StateFleetList && !spec.InHotspots(x, y):
			return
		}
	}
	if r := g.listScreenRowRegion(row); r != -2 {
		g.listHighlight(r)
	}
}

// listHighlight 是 `3705:0591(r)`：與現在反白的同一格就不動；否則換反白
// （−1 ＝ 取消）、左下角寫那一格的摘要（−1 ＝ 清空）、地圖重畫。
func (g *Game) listHighlight(r int) {
	if r == g.lst.hl {
		return
	}
	g.lst.hl, g.lst.hlDrawn = r, true
	g.setHoverRegion(r)
}

// listScreenMapHighlight 回清單開著時地圖該反白的那一格：重畫過就是 hl，
// 還沒重畫過畫面上仍是開清單前的樣子。
func (g *Game) listScreenMapHighlight() (int, bool) {
	if !isListScreen(g.state) && !g.lst.fleetFrom {
		return 0, false
	}
	if !g.lst.hlDrawn {
		return 0, false
	}
	return g.lst.hl, true
}

// listScreenClick 是四張清單開著時的左鍵。
func (g *Game) listScreenClick(x, y int) {
	spec := g.res.ListScreen
	if spec == nil {
		return
	}
	// 詳細資料（與它上面的經歷）開著：經歷只認右鍵；詳細資料認 經歷／結束。
	if g.lst.detail >= 0 {
		if g.lst.bio || g.res.AdmiralPanel == nil {
			return
		}
		switch g.tabDetailButtonAtFor(x, y, false) {
		case "bio":
			g.lst.bio = true
		case "close":
			g.closeListDetail()
		}
		return
	}
	if dir := spec.ArrowAt(x, y); dir != 0 {
		g.listScreenStep(dir)
		return
	}
	row := spec.RowAt(x, y)
	if row < 0 {
		return
	}
	idx := g.listScreenScroll() + row
	switch g.state {
	case StateStarList:
		order := g.starListOrder()
		if idx >= len(order) {
			return
		}
		// `0ABD:0000`：`3705:0284`＋`073F:02EE(r)` 換選那一格、`0D05:089F(0)`
		// 關清單（位移每座位記住）、`073F:05CB(ds:9588)` 開星系情報視窗。
		r := order[idx]
		g.listHighlight(r)
		g.selectRegion(r)
		g.closeListScreen()
		g.openRegionInfo(r)
	case StateFleetList:
		fl := g.listScreenFleets()
		if idx >= len(fl) {
			return
		}
		// `0ABD:0061`：那一列反白（`3603:00DB` 0x20）→ `0DEA:0003(艦隊)`。
		g.lst.pressed, g.lst.fleetFrom = row, true
		g.openFleetWindow(fl[idx])
	case StateAdmiralListView, StateEnemyList:
		if row >= len(g.listIndex) {
			return
		}
		// `0ABD:00D6`：那一列反白 → `0D92:000D(r, 0, 自軍, 0)`。
		g.lst.pressed, g.lst.detail = row, g.listIndex[row]
		g.lst.detailOwn = g.state == StateAdmiralListView
	}
}

// closeListDetail 關詳細資料回清單，那一列的反白還原（`0C0A:06A7(…, 0x87)`）；
// 清單迴圈接著就看游標在哪（`0C0A:0351`）。
func (g *Game) closeListDetail() {
	g.lst.detail, g.lst.bio, g.lst.pressed = -1, false, -1
	g.listScreenHover()
}

// listScreenRightClick 是四張清單開著時的右鍵：經歷 → 詳細資料 → 清單 → 地圖，
// 一次關一層。
func (g *Game) listScreenRightClick() {
	switch {
	case g.lst.bio:
		g.lst.bio = false
	case g.lst.detail >= 0:
		g.closeListDetail()
	default:
		g.closeListScreen()
	}
}

// fleetWindowReturn 是艦隊視窗那一族關掉時回哪裡：從艦隊資料開的回清單
// （`0ABD:0061` 從 `0DEA:0003` 回來重印那一列、繼續清單迴圈），其餘回地圖。
func (g *Game) fleetWindowReturn() State {
	if g.lst.fleetFrom {
		g.lst.fleetFrom, g.lst.pressed = false, -1
		// 回到清單時地圖沒有反白、左下角空白（收據 `fcmd-lst-fleet-row-close`／
		// `-row-rclick`：`ds:F182` 都是 FF，游標在清單外也一樣）。
		g.listHighlight(-1)
		return StateFleetList
	}
	return StateCommandPhase
}

// resumeListScreen 是子視窗關掉回到清單那一刻：清單迴圈接著就看游標在哪
// （`0C7F:03CB`／`0C0A:0351`）。
func (g *Game) resumeListScreen() {
	if isListScreen(g.state) {
		g.listScreenHover()
	}
}

// fleetWindowFamily 回這個狀態是不是艦隊視窗那一族（從艦隊資料開的時候
// 底下畫的是艦隊資料那張）。
func fleetWindowFamily(s State) bool {
	switch s {
	case StateFleetWindow, StateFleetActionMenu, StateFleetTargetPick, StateFleetMessage,
		StateFleetConfirm, StateFleetStandby, StateFleetOccupyWarn, StateFleetChangeMenu,
		StateFleetChangeAdj, StateFleetChangeOrg:
		return true
	}
	return false
}

// listScreenFleetRows 回艦隊資料「士氣」分頁要印的列（從位移起十六列）。
func (g *Game) listScreenFleetRows() []screen.FleetRow {
	spec := g.res.ListScreen
	if spec == nil {
		return nil
	}
	fl := g.listScreenFleets()
	var rows []screen.FleetRow
	for i := g.lst.fleetScroll; i < len(fl) && len(rows) < spec.Rows; i++ {
		f := g.fleets[fl[i]]
		var name []uint16
		if f.Admiral >= 0 && f.Admiral < len(g.res.AdmiralNames) {
			name = g.res.AdmiralNames[f.Admiral]
		}
		rows = append(rows, screen.FleetRow{Number: f.Number, Name: name, Value: f.Morale})
	}
	return rows
}

// drawListScreenBand 在清單那一層畫反白列與艦隊資料的列（在列的字之前、
// 疊上去的視窗之前）。
func (g *Game) drawListScreenBand(fb *gfx.Framebuffer) {
	spec := g.res.ListScreen
	if spec == nil {
		return
	}
	if !isListScreen(g.state) && !(g.lst.fleetFrom && fleetWindowFamily(g.state)) {
		return
	}
	spec.DrawBand(fb, g.lst.pressed)
	if g.state == StateFleetList || g.lst.fleetFrom {
		spec.Fleet.Draw(fb, g.res.Font, g.listScreenFleetRows(), -1)
	}
}

// drawListScreenOverlays 畫清單點出來的詳細資料與經歷（整張畫完、游標之後；
// 原版的游標永遠在最上面，所以再畫一次）。
func (g *Game) drawListScreenOverlays(fb *gfx.Framebuffer) {
	if !isListScreen(g.state) || g.lst.detail < 0 {
		return
	}
	g.drawAdmiralDetailOverlay(fb, g.lst.detail, false, g.lst.detailOwn)
	if g.lst.bio {
		if spec := g.res.Screens["screen-admiral-bio"]; spec != nil && spec.Windows["bio"] != nil && g.res.BioPanel != nil {
			spec.Windows["bio"].Draw(fb, g.res.Font, 3, g.res.Faces)
			if i := g.lst.detail; i < len(g.res.Biographies) {
				g.res.BioPanel.Draw(fb, g.res.Font, g.res.Biographies[i])
			}
		}
	}
	if g.res.Cursor != nil && g.mouseY >= 0 {
		g.res.Cursor.Draw(fb, g.mouseX, g.mouseY-mouseDY)
	}
}

// drawAdmiralDetailOverlay 畫疊在清單上的提督詳細資料（`0D92:000D`）：
// sortie ＝ 有「出擊」那顆（經歷／中止／出擊），否則 經歷／結束 兩顆往右挪
// （熱區 `1FAE`）；own ＝ 印現在位置與狀態（`：待命中／執勤中／出擊中`），
// 否則現在位置印 ＊＊＊＊＊＊＊＊、沒有狀態（敵將那張）。
func (g *Game) drawAdmiralDetailOverlay(fb *gfx.Framebuffer, admiral int, sortie, own bool) {
	spec := g.res.Screens["screen-admiral-detail"]
	if spec == nil || spec.Windows["overlay"] == nil {
		return
	}
	v, ok := g.admiralView(admiral)
	if !ok {
		return
	}
	w := *spec.Windows["overlay"]
	if !sortie {
		// 三顆鈕（經歷／中止／出擊）變兩顆：經歷挪到中止的位置、出擊那顆
		// 改印「結束」。
		w.MoveBox(0, w.Boxes[1].X)
		w.SetBoxCodes(2, adminDetailCloseCodes)
		w.DropBox(1)
	}
	// 規格裡量死的「：待命中」換成算出來的（`admiralDetailExtras`）。
	if p := g.res.AdmiralPanel; p != nil && p.Status.X > 0 {
		for i, l := range w.Lines {
			if l.Y == p.Status.Y {
				w.DropLineCellsFrom(i, p.Status.X)
			}
		}
	}
	rec := g.admiralRecord(admiral)
	w.Draw(fb, g.res.Font, 3, g.res.Faces)
	w.DrawFactionText(fb, g.res.Font, rules.AdmiralFaction(rec))
	v.Prefix, v.Status, v.Location = g.admiralDetailExtras(admiral, own, v.Location)
	g.drawAdmiralView(fb, v, admiral, true)
}

// 「第」「艦隊」「參謀」「：」與三種狀態的字碼（Big5）。
var (
	detailFleetPrefix = []uint16{0xB2C4}
	detailFleetSuffix = []uint16{0xC4A5, 0xB6A4}
	detailStaff       = []uint16{0xB0D1, 0xBFD1}
	detailColon       = uint16(0xA147)
	detailIdle        = []uint16{0xABDD, 0xA952, 0xA4A4} // 待命中
	detailOnDuty      = []uint16{0xB0F5, 0xB6D4, 0xA4A4} // 執勤中
	detailSortie      = []uint16{0xA558, 0xC0BB, 0xA4A4} // 出擊中
	detailMasked      = []uint16{0xA1AF, 0xA1AF, 0xA1AF, 0xA1AF, 0xA1AF, 0xA1AF, 0xA1AF, 0xA1AF}
	detailDead        = []uint16{0xA6BA, 0xA460}                 // 死亡
	detailMissing     = []uint16{0xA455, 0xB8A8, 0xA4A3, 0xA9FA} // 下落不明
)

// admiralDetailExtras 是 `0D92:00B7` 跟著 `+0x09`／`+0x00` 走的三段：姓名前
// 那一格（第Ｎ艦隊／參謀／空）、現在位置（own 才印星域名，死亡／下落不明
// 另印；不是自軍印 ＊＊＊＊＊＊＊＊）與它後面的狀態。
func (g *Game) admiralDetailExtras(admiral int, own bool, location []uint16) (prefix, status, loc []uint16) {
	rec := g.admiralRecord(admiral)
	if rec == nil {
		return nil, nil, location
	}
	duty := rules.AdmiralDutyOf(rec)
	dead, missing := rules.AdmiralAlive(rec)
	if !dead && !missing {
		switch duty {
		case rules.DutyStaff:
			prefix = detailStaff
		case rules.DutySortie:
			// 帶的艦隊：`+0x09` 低七位是艦隊槽，remake 直接找帶隊的那一支。
			for _, f := range g.fleets {
				if f.Admiral == admiral {
					prefix = append(append(append([]uint16(nil), detailFleetPrefix...),
						screen.FullDigits(f.Number)...), detailFleetSuffix...)
					break
				}
			}
		}
	}
	if !own {
		return prefix, nil, detailMasked
	}
	switch {
	case dead:
		return prefix, nil, detailDead
	case missing:
		return prefix, nil, detailMissing
	}
	status = append([]uint16{detailColon}, map[rules.AdmiralDuty][]uint16{
		rules.DutyIdle: detailIdle, rules.DutyStaff: detailOnDuty, rules.DutySortie: detailSortie,
	}[duty]...)
	return prefix, status, location
}
