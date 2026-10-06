package game

import (
	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰略地圖上的艦隊命令（`docs/re/312`、規格 `docs/spec/051` 第一片）。
//
// 入口在戰略地圖介面右下角「艦隊」分頁：點清單一列開那支艦隊的視窗
// （作戰／變更／結束），「作戰」拉出五項選單，「移動」把地圖露出來讓玩家
// 點目的地，確認框按「決定」就寫進艦隊記錄——原版 `+0x07` 目的地、
// `+0x08` 動作碼，remake 是 `Fleet.Target`／`Fleet.Move`／`Fleet.Parked`。
//
// 第一片只接「移動」；第二輪補上駐留／輸送／占領（同一個目的地迴圈，
// 占領的允許表換成敵方＋中立、對費沙先跳參謀提醒）與待命（另一張
// 「指定待命行星」，`fleetstandby.go`）。「變更」點了還是只關選單
// （規格的排除範圍，issue #42）。

// fleetCmd 是這條路上跨狀態要記的東西。
type fleetCmd struct {
	fleet   int // g.fleets 的索引
	hover   int // 作戰選單游標所在那一項（-1 ＝ 不在任何一項上）
	menuX   int // 選單框左上角（畫面座標）
	menuY   int
	action  int                  // 選了第幾項（0 移動）
	allowed [ai.RegionCount]bool // 目的地選取的允許表（`113A:00BE`）
	target  int                  // 確認框在問的星域
	msgKind int                  // 開著的訊息框種類（-1 ＝ 沒有）
	msgRaw  []byte               // 訊息框的內容
	msgConf bool                 // true ＝ 確認框（決定／中止），false ＝ 只有確認
	// probe 是目的地選取迴圈（`113A:0003`）自己那一份「游標底下那一格」
	// ——與地圖介面的 `Game.hoverProbe` 分開。進選取時從 −1 起算，所以
	// 剛進去游標在星圖外那塊不動（`fcmd-pick`）；點了星域再按「中止」回到
	// 選取，游標在框上（星圖外）那塊就清空（`fcmd-cancel`）；而按「決定」
	// 回地圖介面走的是地圖介面自己的探針，那塊留著剛寫的目的地摘要
	// （`fcmd-done`）。
	probe int
	// standbySel 是「指定待命行星」反白的那一顆（區內序號，−1 ＝ 沒有；
	// 原版 `ds:971C` 指著的熱區）；standbyRepainted 記視窗顯示之後換過
	// 顏色的鈕（只有那幾顆會蓋在引線上，`docs/re/314`）。
	standbySel       int
	standbyRepainted [5]bool
	// warnPage 是占領費沙前那兩頁參謀提醒翻到第幾頁（`1213:0003`，
	// `StateFleetOccupyWarn`）。
	warnPage int
	// 「變更」那條路（`fleetchange.go`）：小選單的位置與游標所在項、選了哪一項
	// （子視窗開著時選單留在底下、那一項反白），以及副官那張關掉之後清單
	// 那一列不再反白（`073F:0522` 把清單重畫回未按下的樣子）。
	changeX, changeY int
	changeHover      int
	changeItem       int
	listUnpressed    bool
	// formOn ＝ 「變更 → 編組」那張的陣形清單開著（出擊那條路用
	// `StateFormation`，這條路疊在艦隊視窗上沒有自己的畫面規格）。
	formOn bool
	// faceErased ＝ 艦隊視窗右上角的頭像被疊上來的清單抹掉了（`fleetadj.go`），
	// 關掉「變更 → 副官」那張才重貼。
	faceErased bool
}

// 作戰選單的五項（原版 `+0x08` 的動作碼就是項目索引，`docs/re/312` §三）。
const (
	fleetActionMove = iota
	fleetActionOccupy
	fleetActionGarrison
	fleetActionTransport
	fleetActionStandby
)

// fleetMoveParked 是艦隊記錄 `+0x08` 的「不動」（`docs/internals/33`）。
const fleetMoveParked = rules.MoveParked

// fleetListRowAt 回畫面座標 (x,y) 落在右下角清單的第幾列（0–4），不在清單上
// 回 -1。熱區是 `0BEE[0]`：視窗 9518 的格 (0,7) 起 18×5 ＝ 畫面 x 336–623、
// y 300–399（`docs/re/312` §一）。
func fleetListRowAt(x, y int) int {
	const left, top, w, rowH, rows = 336, 300, 288, 20, 5
	if x < left || x >= left+w || y < top || y >= top+rows*rowH {
		return -1
	}
	return (y - top) / rowH
}

// fleetListFleetAt 回清單第 row 列是 g.fleets 的哪一支；那一列沒有艦隊回 -1。
// 清單列的內容與 `fleetRows` 同一個來源（面板那一格的自軍艦隊，依 g.fleets 順序）。
func (g *Game) fleetListFleetAt(row int) int {
	if row < 0 {
		return -1
	}
	region := g.selectedRegion
	if region < 0 {
		region = rules.CapitalRegion(g.faction)
	}
	n := 0
	for i, f := range g.fleets {
		if f.Faction != g.faction || f.Region != region {
			continue
		}
		if n == row {
			return i
		}
		n++
	}
	return -1
}

// fleetListRowOf 是 `fleetListFleetAt` 的反向：g.fleets 的第 i 支在清單第幾列，
// 不在清單上回 -1。
func (g *Game) fleetListRowOf(i int) int {
	for row := 0; row < len(g.fleets); row++ {
		if g.fleetListFleetAt(row) == i {
			return row
		}
	}
	return -1
}

// openFleetWindow 開第 i 支艦隊的視窗。
func (g *Game) openFleetWindow(i int) {
	if g.res.FleetWindow == nil || i < 0 || i >= len(g.fleets) {
		return
	}
	g.fcmd = fleetCmd{fleet: i, hover: -1, target: -1, msgKind: -1}
	g.state = StateFleetWindow
}

// fleetWindowClick 是艦隊視窗開著時的左鍵：作戰／變更／結束三顆鈕。
func (g *Game) fleetWindowClick(x, y int) {
	switch g.res.FleetWindow.ButtonAt(x, y) {
	case "ops":
		g.openFleetActionMenu(x, y)
	case "close":
		g.closeFleetWindow()
	case "change":
		// 星域有戰鬥或不是自己的時這顆鈕根本沒畫（`fleetChangeHidden`）。
		if !g.fleetChangeHidden(&g.fleets[g.fcmd.fleet]) {
			g.openFleetChangeMenu(x, y)
		}
	}
}

// fleetChangeHidden 是 `0DEA:0040` 開頭那道判斷：所在星域有戰鬥或不是自己的，
// 「變更」鈕不畫也不能點。
func (g *Game) fleetChangeHidden(f *rules.Fleet) bool {
	if g.regionFlags(f.Region)&rules.RegionFlagBattle != 0 {
		return true
	}
	return g.res.Planets == nil || g.res.Planets.RegionOwner(f.Region) != f.Faction
}

// closeFleetWindow 回到戰略地圖介面（從艦隊資料開的回那張清單）。
func (g *Game) closeFleetWindow() {
	g.fcmd = fleetCmd{fleet: -1, hover: -1, target: -1, msgKind: -1}
	g.state = g.fleetWindowReturn()
	g.resumeListScreen()
}

// openFleetActionMenu 開「作戰」的五項選單。框的位置照戰場命令選單那條公式
// （游標所在格 − (1,1)，列夾在板子裡；`docs/re/312` §三），
// 座標系是 640×400 的滑鼠座標，所以 y 先扣 40。
func (g *Game) openFleetActionMenu(x, y int) {
	m := g.res.FleetActionMenu
	if m == nil {
		return
	}
	mx, my := m.CommandMenuAt(x, y-mouseDY, len(m.Menus[fleetActionMenuKind].Items))
	g.fcmd.menuX, g.fcmd.menuY = mx, my+mouseDY
	g.state = StateFleetActionMenu
	g.fleetMenuHover(x, y)
}

// fleetActionMenuKind 是規格檔裡那張選單的名字。
const fleetActionMenuKind screen.CommandMenuKind = "strategic-action"

// fleetMenuHover 更新選單的反白：跟著游標走（收據 `fcmd-menu-hover`）。
func (g *Game) fleetMenuHover(x, y int) {
	m := g.res.FleetActionMenu
	if m == nil {
		return
	}
	n := len(m.Menus[fleetActionMenuKind].Items)
	g.fcmd.hover = m.CommandMenuRowAt(fleetActionMenuKind, g.fcmd.menuX, g.fcmd.menuY, x, y, n)
}

// fleetMenuEnabled 是五項的啟用旗標（`0DEA:1037` 開選單前算的，`docs/re/312`
// §7.6）：所在星域有戰鬥（`+0x13 & 6`）或不是自己的 → 只剩「移動」；
// 艦隊 `+0x00` bit0 → 占領與輸送關掉。停用的印色 8、游標停在上面不畫反白、
// 點了只關選單（收據 `fcmd-battle-menu*`、`fcmd-flag1-menu*`，`docs/re/315`）。
func (g *Game) fleetMenuEnabled() []bool {
	on := []bool{true, true, true, true, true}
	if g.fcmd.fleet < 0 || g.fcmd.fleet >= len(g.fleets) {
		return on
	}
	f := &g.fleets[g.fcmd.fleet]
	if g.regionFlags(f.Region)&6 != 0 || g.res.Planets == nil || g.res.Planets.RegionOwner(f.Region) != f.Faction {
		on[fleetActionOccupy], on[fleetActionGarrison], on[fleetActionTransport], on[fleetActionStandby] = false, false, false, false
	}
	if f.Flags&1 != 0 {
		on[fleetActionOccupy], on[fleetActionTransport] = false, false
	}
	return on
}

// fleetMenuClick 是選單開著時的左鍵：點在項目上就選它，點在外面或點到停用的
// 關選單（探索批次 `fcx-pop-255`：點在第一項上方只把選單收掉；
// `fcmd-battle-menu-click`／`fcmd-flag1-menu-click`：點停用的項也是）。
func (g *Game) fleetMenuClick(x, y int) {
	g.fleetMenuHover(x, y)
	item := g.fcmd.hover
	g.fcmd.hover = -1
	if on := g.fleetMenuEnabled(); item >= 0 && item < len(on) && !on[item] {
		item = -1
	}
	switch item {
	case fleetActionMove, fleetActionOccupy, fleetActionGarrison, fleetActionTransport:
		g.startFleetTargetPick(item)
	case fleetActionStandby:
		g.openFleetStandby()
	default:
		g.state = StateFleetWindow
	}
}

// startFleetTargetPick 進入目的地選取（`113A:0003`）：建允許表，地圖露出來。
func (g *Game) startFleetTargetPick(action int) {
	f := &g.fleets[g.fcmd.fleet]
	g.fcmd.action = action
	g.fcmd.allowed = g.fleetTargetsAllowed(f, action)
	g.fcmd.probe = -1
	g.state = StateFleetTargetPick
}

// fleetTargetsAllowed 是 `113A:00BE` 的允許表：從艦隊所在星域起算的可達表
// （`2B99:0392`，k ≤ 1）∧ 擁有者在允許的那兩個裡（或就是現在位置）。
// 移動／駐留／輸送允許的擁有者是 (自己, 0xFF)，占領是 (敵方, 中立)。
func (g *Game) fleetTargetsAllowed(f *rules.Fleet, action int) [ai.RegionCount]bool {
	var out [ai.RegionCount]bool
	if g.res.Graph == nil || g.res.Planets == nil {
		return out
	}
	owner := func(r int) int { return g.res.Planets.RegionOwner(r) }
	reach := ai.ReachFrom(g.res.Graph, owner, f.Faction, f.Region)
	a, b := f.Faction, -1
	if action == fleetActionOccupy {
		a, b = 1-f.Faction, 2
	}
	for i := range out {
		if !reach[i] {
			continue
		}
		o := owner(i)
		out[i] = o == a || o == b || i == f.Region
	}
	return out
}

// fleetPickClick 是目的地選取時的左鍵（`113A:01AB`）：只認地圖上的星域。
func (g *Game) fleetPickClick(x, y int) {
	if g.res.Regions == nil || x < mapX || x >= mapX+mapW || y < mapY || y >= mapY+mapH {
		return
	}
	r := asset.RegionAt(g.res.Regions, x-mapX, y-mapY, tileW, tileH)
	if r < 0 || r >= ai.RegionCount {
		return
	}
	spec := g.res.FleetWindow
	if !g.fcmd.allowed[r] {
		g.fcmd.msgKind, g.fcmd.msgConf = spec.UnreachableKind, false
		g.fcmd.msgRaw = decodeHexBytes(spec.Unreachable)
		g.state = StateFleetMessage
		return
	}
	g.fcmd.target = r
	// `1213:0003`：目的地是費沙而且費沙事件還沒發生過（`ds:8F48 == 0`）
	// → 先給兩頁參謀提醒（與費沙事件的報告同一支 `1213:01B8`），
	// 每頁一次點擊（或訊息速度 × 2 的逾時）才往下走到確認框
	// （收據 `fcmd-occupy-phezzan`／`-phezzan2`／`-phezzan3`，`docs/re/315`）。
	if r == rules.PhezzanRegion && !rules.PhezzanOccurred(g.workRecord()) && g.res.PhezzanReport != nil {
		g.fcmd.warnPage = 0
		g.state = StateFleetOccupyWarn
		return
	}
	g.openFleetConfirm()
}

// openFleetConfirm 是 `113A:023B(星域)`：動詞是「待命」（點的是現在位置）或
// 這一項的動詞；陣營名用 `ds:6092[擁有者] | 0x90` 的顏色，星域名用 0x97。
func (g *Game) openFleetConfirm() {
	spec := g.res.FleetWindow
	f := &g.fleets[g.fcmd.fleet]
	r := g.fcmd.target
	verb := decodeHexBytes(spec.Verbs[g.fcmd.action])
	if r == f.Region {
		verb = decodeHexBytes(spec.VerbStandby)
	}
	owner := g.res.Planets.RegionOwner(r)
	g.fcmd.msgKind, g.fcmd.msgConf = spec.ConfirmKind, true
	g.fcmd.msgRaw = ui.Sprintf(decodeHexBytes(spec.ConfirmFormat),
		verb, 0x90|int(rules.FactionColour(owner)), decodeHexBytes(spec.FactionNames[owner]),
		0x97, codesBig5(g.regionName(r)))
	g.state = StateFleetConfirm
}

// fleetOccupyWarnClick 是參謀提醒開著時的任一鍵（`3765:006C` 只認事件碼 1／2，
// 不看座標）：翻頁，第二頁翻過去就開確認框。
func (g *Game) fleetOccupyWarnClick() {
	g.fcmd.warnPage++
	if g.fcmd.warnPage < 2 {
		return
	}
	g.openFleetConfirm()
}

// occupyWarnKey 是這一頁在 `phezzan-report-panel.json` 裡的鍵：帝國看 `ds:3016`
// 那兩頁、同盟看 `ds:301E` 那兩頁。
func (g *Game) occupyWarnKey() string {
	if g.faction == 1 {
		return "occupy_1"
	}
	return "occupy_0"
}

// fleetMessageClick 是目的地選取跳出來的兩種訊息框的左鍵。
//
// 「這是無法前往的星系。」只有「確認」，關掉回選取；確認框「決定」寫命令
// 回地圖、「中止」回選取（收據 `fcmd-far`／`fcmd-done`／`fcmd-cancel`）。
// 訊息框是模態的：點在按鈕外什麼都不做（`docs/re/137`）。
func (g *Game) fleetMessageClick(x, y int) {
	mb := g.res.MessageBox
	if mb == nil {
		g.state = StateFleetTargetPick
		return
	}
	i, hit := mb.MessageBoxButtonAt(g.fcmd.msgKind, x, y)
	if !hit {
		return
	}
	g.fcmd.msgRaw = nil
	if !g.fcmd.msgConf || !mb.ButtonIsDecide(g.fcmd.msgKind, i) {
		g.state = StateFleetTargetPick
		return
	}
	g.applyFleetOrder()
}

// applyFleetOrder 是確認框按「決定」之後那一段（`0DEA:11xx`，`docs/re/312` §三）：
// 點的是現在位置 → 取消既有命令（`+0x08 ← 5`、`+0x07 ← 0xFF`）；
// 否則 `+0x08 ← 動作碼`、`+0x07 ← 目的地`。結尾 `090A:0E6D` 重算未安定度。
//
// 原版 `0DEA:1037` 的三個區塊（`docs/re/313`）：移動／占領／輸送走同一段——
// 目的地是現在位置就取消既有命令（`+0x08 ← 5`、`+0x07 ← 0xFF`），而且
// **不算「有改」**，視窗留著（收據 `fcmd-self-done`：與 `fcmd-win` 只差游標）；
// 別的星域才寫命令、關窗。駐留（2）是另一段：不管目的地是不是現在位置一律
// 寫 `+0x08 ← 2`、`+0x07 ← r` 並關窗（`fcmd-garrison-done`）。
func (g *Game) applyFleetOrder() {
	f := &g.fleets[g.fcmd.fleet]
	r := g.fcmd.target
	changed := true
	switch {
	case g.fcmd.action == fleetActionGarrison:
		f.Target, f.Move, f.Parked = r, fleetActionGarrison, false
	case r == f.Region:
		f.Target, f.Move, f.Parked = rules.NoTarget, fleetMoveParked, true
		changed = false
	default:
		f.Target, f.Move, f.Parked = r, g.fcmd.action, false
	}
	g.recomputeUnrest(g.turn)
	// 目的地迴圈結束時地圖畫回原樣、再反白清單看的那一格
	// （`3705:0349(0)`、`3705:0284(ds:9588)`，`docs/re/312` §四）。
	g.mapHighlight = g.fleetListRegion()
	if changed {
		g.closeFleetWindow()
		return
	}
	g.state = StateFleetWindow
}

// fleetListRegion 是右下角清單看的那一格（原版 `ds:9588`）：左鍵點星圖
// 換選過的星域，沒選過就是自己的首都。
func (g *Game) fleetListRegion() int {
	if g.selectedRegion >= 0 {
		return g.selectedRegion
	}
	return rules.CapitalRegion(g.faction)
}

// fleetCmdRightClick 是這條路上每一層的右鍵：關最上面那一層
// （收據 `fcmd-win-rclick`／`fcmd-menu-rclick`／`fcmd-pick-rclick`）；
// 座標只有編組那張用得到（拆格盤的哪一格）。
func (g *Game) fleetCmdRightClick(x, y int) {
	switch g.state {
	case StateFleetWindow:
		g.closeFleetWindow()
	case StateFleetActionMenu:
		g.fcmd.hover = -1
		g.state = StateFleetWindow
	case StateFleetTargetPick:
		g.mapHighlight = g.fleetListRegion()
		g.state = StateFleetWindow
	case StateFleetMessage, StateFleetConfirm:
		// 訊息框只認左鍵（`3826:023B`）。
	case StateFleetOccupyWarn:
		g.fleetOccupyWarnClick()
	case StateFleetStandby:
		// 右鍵＝「決定」：有反白的就寫進去（`0DEA:165A` 的事件 2 直接跳出
		// 迴圈，`1320` 回目前反白那一筆；收據 `fcmd-standby-rclick`）。
		g.finishFleetStandby()
	case StateFleetChangeMenu:
		g.closeFleetChangeMenu()
	case StateFleetChangeAdj:
		// 小選單／清單／詳細資料開著就只關那一層；否則右鍵 ＝ 結束
		// （`fcmd-change-adj-rclick`）。
		if g.adjRightClick() {
			return
		}
		g.closeFleetChangeAdj()
	case StateFleetChangeOrg:
		// 右鍵不關窗（`fcmd-change-org-rclick`），只拆格盤那一格
		// （`fcmd-edit-remove`）或關陣形清單。
		g.fleetChangeOrgRightClick(x, y)
	}
}

// fleetCmdMouseMove 是這條路上的游標移動：選單的反白跟著走。
func (g *Game) fleetCmdMouseMove(x, y int) {
	switch g.state {
	case StateFleetActionMenu:
		g.fleetMenuHover(x, y)
	case StateFleetChangeMenu:
		g.fleetChangeMenuHover(x, y)
	}
}

// FleetCommandFleet 回艦隊視窗正在看哪一支（-1 ＝ 沒開）。
func (g *Game) FleetCommandFleet() int {
	switch g.state {
	case StateFleetWindow, StateFleetActionMenu, StateFleetTargetPick, StateFleetMessage, StateFleetConfirm,
		StateFleetStandby, StateFleetOccupyWarn, StateFleetChangeMenu, StateFleetChangeAdj, StateFleetChangeOrg:
		return g.fcmd.fleet
	}
	return -1
}

// FleetTargetsAllowed 回目的地選取的允許表（測試與診斷用）。
func (g *Game) FleetTargetsAllowed() [ai.RegionCount]bool { return g.fcmd.allowed }

// ── 畫面 ──

// fleetWindowView 把視窗要印的值湊齊（`docs/re/312` §二、§七）。
func (g *Game) fleetWindowView() (*screen.FleetWindowView, *asset.Image) {
	i := g.fcmd.fleet
	if i < 0 || i >= len(g.fleets) {
		return nil, nil
	}
	f := &g.fleets[i]
	adm := g.admiralRec(f.Admiral)
	adj := g.admiralRec(f.Adjutant)
	v := &screen.FleetWindowView{
		Faction: f.Faction, Number: f.Number,
		CmdName:  codesBig5(g.admiralNameCodes(f.Admiral)),
		CmdRank:  codesBig5(g.admiralRankCodes(adm)),
		CmdLevel: byteAt(adm, admiralMeritOff), CmdExp: u32At(adm, admiralExpOff),
		Move: f.Move, Target: f.Target,
		RegionName: codesBig5(g.regionName(f.Region)), RegionOwner: g.res.Planets.RegionOwner(f.Region),
		RegionFlags: g.regionFlags(f.Region),
		Expected:    f.Expected,
		Morale:      f.Morale,
		Command:     byteAt(adm, admiralCommandOff),
		OpsPressed: g.state == StateFleetActionMenu || g.state == StateFleetTargetPick ||
			g.state == StateFleetMessage || g.state == StateFleetConfirm || g.state == StateFleetStandby ||
			g.state == StateFleetOccupyWarn,
		ChangeHidden:  g.fleetChangeHidden(f),
		ChangePressed: g.state == StateFleetChangeMenu || g.state == StateFleetChangeAdj || g.state == StateFleetChangeOrg,
	}
	if f.Move != fleetMoveParked && f.Target >= 0 {
		v.TargetName = codesBig5(g.regionName(f.Target))
		v.TargetOwner = g.res.Planets.RegionOwner(f.Target)
	}
	// 停駐且 `+0x07` 不是 0xFF ＝ 待命行星：所在星域的第幾顆
	// （`0DEA:04C0` 先 `2AF5:0036(星域)` 再 `ds:CFF8 + 0x52×+0x07`，收據 `fcmd-standby-reopen`）。
	if f.Move == fleetMoveParked && f.Target >= 0 && g.res.Planets != nil {
		if p := g.standbyPlanetIndex(f); p >= 0 {
			v.StandbyPlanetName = codesBig5(g.res.Planets.Name(p))
			v.StandbyPlanetOwner = g.res.Planets.Owner(p)
		}
	}
	if adj != nil {
		v.AdjPresent = true
		v.AdjName = codesBig5(g.admiralNameCodes(f.Adjutant))
		v.AdjRank = codesBig5(g.admiralRankCodes(adj))
		v.AdjLevel, v.AdjExp = byteAt(adj, admiralMeritOff), u32At(adj, admiralExpOff)
	}
	// 編制、飛彈／戰鬥機存量、速度：有艦艇池就讀艦艇記錄，沒有（新遊戲
	// 編成的）就從編制與艦種表算——新編艦隊的存量就是滿的。
	ab := rules.FleetAbilities(adm, adj)
	for _, gp := range f.Composition {
		v.Groups = append(v.Groups, screen.FleetWindowGroup{Kind: gp.Kind, Ships: gp.Ships, Full: gp.Full})
		if k, ok := g.res.ShipKinds.Kind(f.Faction, gp.Kind); ok {
			if k.Byte(0x0E) != 0xFF {
				v.MissileMax += k.Byte(0x10) * gp.Ships
				v.MissileCur += k.Byte(0x10) * gp.Ships
			}
			if k.Byte(0x12) != 0xFF {
				v.FighterMax += k.Byte(0x14) * gp.Ships
				v.FighterCur += k.Byte(0x14) * gp.Ships
			}
			// `297B:0E8B`：三個速度是各組的最小值。
			t, c, b := g.res.FleetWindow.ShipSpeeds(k.Byte(0x08), k.Byte(0x06), k.Byte(0x07),
				int(ab[6]), int(ab[5]), int(ab[9]), len(f.Composition))
			if v.Turn == 0 || t < v.Turn {
				v.Turn = t
			}
			if v.Cruise == 0 || c < v.Cruise {
				v.Cruise = c
			}
			if v.Fight == 0 || b < v.Fight {
				v.Fight = b
			}
		}
	}
	// 費用（`docs/internals/33`：`+0x0A` 維持費、`+0x0C` 移動費）。
	if g.res.ShipKinds != nil {
		v.Upkeep = rules.FleetUpkeep(g.res.ShipKinds, f.Faction, f.Composition, f.Operations)
		v.MoveCost = rules.FleetOtherCost(g.res.ShipKinds, f.Faction, f.Composition, f.Operations)
	}
	// 五個能力：艦隊 +0x3D–+0x41 ＝ 光線砲、飛彈、戰鬥機、占領、防禦。
	v.Abilities = [5]int{int(ab[0]), int(ab[1]), int(ab[2]), int(ab[3]), int(ab[4])}
	// 「副官比提督高」看提督記錄的哪一格（`docs/re/312` §七）。
	better := func(off int) bool { return adj != nil && byteAt(adj, off) > byteAt(adm, off) }
	v.CostAdjBetter = better(0x0E)
	v.SpeedAdjBetter = [3]bool{better(0x0D), better(0x0C), better(0x0C)}
	v.AbilAdjBetter = [5]bool{better(0x0F), better(0x10), better(0x11), better(0x13), better(0x12)}

	// 頭像是直接貼上去的，不在視窗的文字緩衝區裡：疊在上面的清單關掉時
	// 視窗從緩衝區重畫，頭像就不見了，要等關掉「變更 → 副官」那張時
	// `3958:0215` 再貼（收據 `fcmd-adj-change-pick`／`-pick-cancel`／`-confirm-close`）。
	var face *asset.Image
	if g.res.Faces != nil && !g.fcmd.faceErased {
		if img, err := g.res.Faces.Decode(f.Admiral); err == nil {
			face = img
		}
	}
	return v, face
}

// regionFlags 回星域記錄的 `+0x13`。
func (g *Game) regionFlags(i int) int {
	if i < 0 || (i+1)*rules.RegionRecord > len(g.res.Stellar) {
		return 0
	}
	return int(g.res.Stellar[i*rules.RegionRecord+0x13])
}

// admiralNameCodes 回提督的名字（字碼）。
func (g *Game) admiralNameCodes(i int) []uint16 {
	if i < 0 || i >= len(g.res.AdmiralNames) {
		return nil
	}
	return g.res.AdmiralNames[i]
}

// codesBig5 把字碼串還原成 Big5 位元組（每個字碼是「高位元組 ＋ 低位元組」，
// 半形兩個一組的格子也一樣拆成兩個）。
func codesBig5(codes []uint16) []byte {
	out := make([]byte, 0, 2*len(codes))
	for _, c := range codes {
		out = append(out, byte(c>>8), byte(c))
	}
	return out
}

// pickStarMap 是目的地選取時那張重新著色的地圖（`113A:00BE`）：
// 艦隊所在 → `FactionColour(擁有者)`、不在允許表 → 3、其餘 → `{1, 12, 7}[擁有者]`。
func (g *Game) pickStarMap() *asset.Image {
	if g.res.StarBit == nil || g.res.StarRGN == nil || g.res.Planets == nil {
		return nil
	}
	n := len(g.res.Stellar) / rules.RegionRecord
	if n == 0 || g.fcmd.fleet < 0 || g.fcmd.fleet >= len(g.fleets) {
		return nil
	}
	f := &g.fleets[g.fcmd.fleet]
	cols := make([]uint8, n)
	for i := range cols {
		owner := g.res.Planets.RegionOwner(i)
		switch {
		case i == f.Region:
			cols[i] = rules.FactionColour(owner)
		case i >= ai.RegionCount || !g.fcmd.allowed[i]:
			cols[i] = 3
		default:
			cols[i] = pickReachableColour(owner)
		}
	}
	img, err := asset.BuildStarMap(g.res.StarBit, g.res.StarRGN, cols, 16, 20, 20)
	if err != nil {
		return nil
	}
	return img
}

// pickReachableColour 是 `ds:2DAA`：帝國 1、同盟 12、中立 7。
func pickReachableColour(owner int) uint8 {
	switch owner {
	case 0:
		return 1
	case 1:
		return 12
	default:
		return 7
	}
}

// drawFleetCommand 把這條路的各層畫在戰略地圖介面上。
//
// 目的地選取時原版把地圖視窗與左下角摘要提到最上層（`341B:02A3`），
// 艦隊視窗的右半照舊露著（收據 `fcmd-pick`）；這裡就照那個順序畫：
// 先視窗，再把地圖與左下角那塊蓋上去。
func (g *Game) drawFleetCommand(fb *gfx.Framebuffer) {
	if g.FleetCommandFleet() < 0 || g.res.FleetWindow == nil {
		return
	}
	v, face := g.fleetWindowView()
	if v == nil {
		return
	}
	// 目的地選取（`113A:0003`）不關視窗，而是把星圖與左下角那塊**整個左半**
	// 重畫在視窗上面：視窗在 x < 320 的那幾格回到底板（地圖換成著色過的
	// 那張、左下角摘要連框一起），x ≥ 320 那一半留著（收據 `fcmd-pick`：
	// 「光線砲／飛彈／…」那幾列仍在，「回轉速度／巡航速度」那幾列不在）。
	// remake 先把畫視窗之前的左半留一份，畫完視窗放回去。
	var left []uint8
	picking := g.state == StateFleetTargetPick || g.state == StateFleetMessage ||
		g.state == StateFleetConfirm || g.state == StateFleetOccupyWarn
	if picking {
		left = append([]uint8(nil), fb.Pix...)
	}
	g.res.FleetWindow.Draw(fb, g.res.Font, g.res.FrameTiles, v, face)
	switch g.state {
	case StateFleetChangeMenu, StateFleetChangeAdj, StateFleetChangeOrg:
		g.drawFleetChange(fb)
	case StateFleetStandby:
		g.drawFleetStandby(fb)
	case StateFleetActionMenu:
		if m := g.res.FleetActionMenu; m != nil {
			screen.DrawCommandMenu(m, fb, g.res.Font, g.res.FrameTiles, fleetActionMenuKind,
				g.fleetMenuEnabled(), g.fcmd.menuX, g.fcmd.menuY, g.fcmd.hover)
		}
	case StateFleetTargetPick, StateFleetMessage, StateFleetConfirm, StateFleetOccupyWarn:
		for y := 0; y < gfx.Height; y++ {
			fb.WriteRow(0, y, left[y*gfx.Width:y*gfx.Width+mapX+mapW])
		}
		if m := g.pickStarMap(); m != nil {
			screen.DrawStarMap(fb, m, mapX, mapY)
		}
		switch {
		case g.state == StateFleetOccupyWarn:
			// 參謀提醒與費沙事件的報告同一張版面、同一張頭像（觀看陣營的
			// 軍事專任參謀，`ds:8F4B[陣營]`）。
			var face *asset.Image
			if idx := rules.StaffAdmiral(g.workRecord(), rules.PostMilitary, g.faction); g.res.Faces != nil && idx >= 0 {
				face, _ = g.res.Faces.Decode(idx)
			}
			g.res.PhezzanReport.Draw(fb, g.res.Font, g.res.FrameTiles, g.occupyWarnKey(), g.fcmd.warnPage, face)
		case g.state != StateFleetTargetPick && g.res.MessageBox != nil:
			screen.DrawMessageBox(fb, g.res.MessageBox, g.fcmd.msgKind,
				g.res.FrameTiles, g.res.Font, g.fcmd.msgRaw)
		}
	}
}

// decodeHexBytes 把規格檔裡的十六進位字串還原成位元組（原版資料段的 Big5）。
func decodeHexBytes(s string) []byte {
	out := make([]byte, 0, len(s)/2)
	for i := 0; i+1 < len(s); i += 2 {
		out = append(out, hexNibble(s[i])<<4|hexNibble(s[i+1]))
	}
	return out
}

func hexNibble(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
