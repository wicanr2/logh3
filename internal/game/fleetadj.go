package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 艦隊編成視窗的「副官」→ 變更／卸任（原版 `07F9:085D` 的熱區 0 →
// `07F9:0A80`，`docs/re/318`）。出擊命令那條路（`StateFleetOrg` 上的
// `StateFleetAdjutant` 小選單）與艦隊視窗「變更 → 副官」（`StateFleetChangeAdj`）
// 共用這一份：
//
//	卸任 → 提督記錄 +0x07 ← 0xFF、舊副官 +0x09 ← 0x80，視窗重印
//	變更 → 「待命中副官資料」清單（`0C0A:000F(星域, 0x65, 等級)`：同星域、同陣營、
//	       沒有職務、閒置、等級比提督低）→ 點一列 → 提督詳細資料（`0D92:000D`）
//	       → 出擊(決定) 換人、中止／右鍵不換；兩種都回艦隊編成視窗
//
// 兩條路都只改**提督記錄**：出擊那條路「決定」建艦隊時從 +0x07 抄，
// 「變更 → 副官」那張關窗時 `297B:04EF` 抄回艦隊 +0x05 並重算費用。

// adjPick 是副官那條路上開著的東西。
type adjPick struct {
	menuOn   bool  // 變更／卸任 小選單開著（只有「變更 → 副官」那條路用；出擊那條路是 StateFleetAdjutant）
	listOn   bool  // 「待命中副官資料」清單開著
	detailOn bool  // 詳細資料視窗開著
	detail   int   // 詳細資料顯示這一位
	list     []int // 候選（提督索引，照提督表的順序）
	first    int   // 清單捲到第幾筆（`ds:96A1`；▲▼每次一列）
}

// adjAdmiral 回這條路上帶隊的提督：出擊那條路是點進來的那一位，
// 「變更 → 副官」是那一支艦隊的提督。
func (g *Game) adjAdmiral() int {
	if g.state == StateFleetChangeAdj {
		return g.fleets[g.fcmd.fleet].Admiral
	}
	return g.detailAdmiral
}

// adjMenuRowAt 回畫面座標落在 變更／卸任 的哪一列（`screen-fleet-adjutant` 的
// `menu` 視窗那兩列反白條）；沒有回 -1。
func (g *Game) adjMenuRowAt(x, y int) int {
	spec := g.res.Screens["screen-fleet-adjutant"]
	if spec == nil || spec.Windows["menu"] == nil {
		return -1
	}
	return spec.Windows["menu"].HoverRowAt(x, y)
}

// adjMenuPick 是小選單選了第 row 項（`3667:00F2` 回來）：0 變更開清單、
// 1 卸任、其餘（點在外面）只關選單。
func (g *Game) adjMenuPick(row int) {
	g.adj.menuOn = false
	switch row {
	case 0:
		g.openAdjList()
	case 1:
		g.setAdjutant(-1)
	}
}

// adjBusy 回這一位在職或在艦隊裡（原版讀 `+0x09` bit7；見 `rules.AdjutantCandidates`）。
func (g *Game) adjBusy(admiral int) func(int) bool {
	busy := map[int]bool{admiral: true}
	for _, i := range rules.StaffAdmirals(g.workRecord(), g.faction) {
		busy[i] = true
	}
	for _, f := range g.fleets {
		busy[f.Admiral] = true
		if f.Adjutant >= 0 {
			busy[f.Adjutant] = true
		}
	}
	if rec := g.admiralRecord(admiral); rec != nil {
		if a := rules.Adjutant(rec); a >= 0 {
			busy[a] = true
		}
	}
	return func(i int) bool { return busy[i] }
}

// openAdjList 開「待命中副官資料」。
func (g *Game) openAdjList() {
	admiral := g.adjAdmiral()
	g.adj.list = rules.AdjutantCandidates(rules.ParseAdmirals(g.res.Admiral),
		g.admiralRecord(admiral), g.adjBusy(admiral))
	g.adj.listOn, g.adj.detailOn, g.adj.first = true, false, 0
}

// adjListPage 是清單一頁幾列（規格的 `list.rows`，原版 16）。
func (g *Game) adjListPage() int {
	spec := g.res.Screens["screen-standby-admirals"]
	if spec == nil || spec.Windows["overlay"] == nil || spec.Windows["overlay"].List == nil {
		return 16
	}
	return spec.Windows["overlay"].List.Rows
}

// adjListScrollAt 回畫面座標落在清單右緣的 ▲（-1）或 ▼（+1）上；沒有回 0
// （那兩顆是 overlay 的 14 × 20 小框，最上面那顆 ▲、最下面那顆 ▼）。
func (g *Game) adjListScrollAt(x, y int) int {
	spec := g.res.Screens["screen-standby-admirals"]
	return admiralScrollAt(spec, x, y)
}

func admiralScrollAt(spec *screen.MainGameSpec, x, y int) int {
	if spec == nil || spec.Windows["overlay"] == nil {
		return 0
	}
	top, bottom := -1, -1
	for i, b := range spec.Windows["overlay"].Boxes {
		// 月初規格存8像素箭頭圖塊，待命規格存14像素內框。
		// 原版1BF2熱區寬16，兩種圖塊右緣相同（規格089）。
		if (b.W != 14 && b.W != 8) || len(b.Codes) > 0 {
			continue
		}
		if top < 0 || b.Y < spec.Windows["overlay"].Boxes[top].Y {
			top = i
		}
		if bottom < 0 || b.Y > spec.Windows["overlay"].Boxes[bottom].Y {
			bottom = i
		}
	}
	for dir, i := range map[int]int{-1: top, 1: bottom} {
		if i < 0 {
			continue
		}
		b := spec.Windows["overlay"].Boxes[i]
		if x >= b.X+b.W-16 && x < b.X+b.W && y >= b.Y && y < b.Y+b.H {
			return dir
		}
	}
	return 0
}

// adjListScroll 每次一列（`0C0A:01D7`／`0235`），保持速度依規格089。
func (g *Game) adjListScroll(dir int) {
	g.personnelBase = nil
	page := g.adjListPage()
	last := len(g.adj.list) - page
	if last < 0 {
		last = 0
	}
	g.adj.first += dir
	if g.adj.first > last {
		g.adj.first = last
	}
	if g.adj.first < 0 {
		g.adj.first = 0
	}
}

// adjListRowAt 回畫面座標落在清單的第幾列（`screen-standby-admirals` 的 list
// 那十六列）；不在清單上回 -1。
func (g *Game) adjListRowAt(x, y int) int {
	spec := g.res.Screens["screen-standby-admirals"]
	if spec == nil || spec.Windows["overlay"] == nil || spec.Windows["overlay"].List == nil {
		return -1
	}
	l := spec.Windows["overlay"].List
	if l.Pitch <= 0 || y < l.Y {
		return -1
	}
	row := (y - l.Y) / l.Pitch
	if row < 0 || row >= l.Rows || g.adj.first+row >= len(g.adj.list) {
		return -1
	}
	return g.adj.first + row
}

// adjClick 是清單或詳細資料開著時的左鍵。回有沒有吃掉這一下。
func (g *Game) adjClick(x, y int) bool {
	switch {
	case g.adj.detailOn:
		// 詳細資料：中止關窗不換、出擊(決定) 換人；經歷還沒接。
		if p := g.res.AdmiralPanel; p != nil {
			switch p.ButtonAt(x, y) {
			case "cancel":
				g.adj.detailOn = false
			case "sortie":
				g.setAdjutant(g.adj.detail)
				g.adj.detailOn = false
			}
		}
		return true
	case g.adj.listOn:
		// ▲▼每次一列；點一列開那一位的詳細資料，清單同時關掉（`0C0A:0742`在開詳細
		// 資料之前）；點在別處什麼都不做（右鍵也不關，收據 `fcmd-adj-change-rclick`）。
		if dir := g.adjListScrollAt(x, y); dir != 0 {
			g.adjListScroll(dir)
			return true
		}
		if row := g.adjListRowAt(x, y); row >= 0 {
			g.adj.detail, g.adj.detailOn = g.adj.list[row], true
			g.closeAdjList()
		}
		return true
	case g.adj.menuOn:
		g.adjMenuPick(g.adjMenuRowAt(x, y))
		return true
	}
	return false
}

// closeAdjList 關清單（`0C0A:0742`）。清單蓋住的東西從各自的文字緩衝區重畫，
// 直接貼上去的頭像不會回來——「變更 → 副官」那條路上艦隊視窗右上角的提督頭像
// 從此不見，關窗才重貼（`fleetcmd.faceErased`）。
func (g *Game) closeAdjList() {
	g.adj.listOn = false
	if g.state == StateFleetChangeAdj {
		g.fcmd.faceErased = true
	}
}

// adjRightClick 是小選單、清單或詳細資料開著時的右鍵：關最上面那一層（詳細資料
// 關掉就回視窗，清單不會再開）。回有沒有吃掉這一下。
//
// 清單的右鍵在原版是「回 0xFF、關清單」（出擊命令的待命中提督清單就是這樣關的，
// 探索批次 `fcx-s3list-rclick`），但從「副官」進來的這一份在收據裡右鍵之後
// 清單留在畫面上、之後的點擊也沒有反應（`fcmd-adj-change-rclick`、
// `fcx-adj-list-rr*`：事件沒有進到清單的迴圈，成因沒讀）。remake 照同一個
// 元件的規則關清單；那一批收據不比（規格 013 已知差異）。
func (g *Game) adjRightClick() bool {
	switch {
	case g.adj.detailOn:
		g.adj.detailOn = false
	case g.adj.listOn:
		g.closeAdjList()
	case g.adj.menuOn:
		g.adj.menuOn = false
	default:
		return false
	}
	return true
}

// adjOpen 回這條路上有沒有東西開著。
func (g *Game) adjOpen() bool {
	return g.adj.menuOn || g.adj.listOn || g.adj.detailOn
}

// closeAdj 全部關掉（離開視窗時）。
func (g *Game) closeAdj() {
	g.adj = adjPick{}
}

// setAdjutant 換副官（adj ＝ -1 卸任）：改的是提督記錄（`rules.SetAdjutant`）。
func (g *Game) setAdjutant(adj int) {
	admiral := g.adjAdmiral()
	g.mutableAdmirals()
	rec := g.admiralRecord(admiral)
	if rec == nil {
		return
	}
	rules.SetAdjutant(rec, g.admiralRecord(rules.Adjutant(rec)), g.admiralRecord(adj), adj)
}

// drawAdjPick 把小選單、清單、詳細資料疊在艦隊編成視窗上。
func (g *Game) drawAdjPick(fb *gfx.Framebuffer) {
	if g.adj.menuOn {
		if spec := g.res.Screens["screen-fleet-adjutant"]; spec != nil && spec.Windows["menu"] != nil {
			w := spec.Windows["menu"]
			w.SetHoverRow(w.HoverRowAt(g.mouseX, g.mouseY))
			w.Draw(fb, g.res.Font, 3, g.res.Faces)
		}
	}
	if g.adj.listOn {
		g.drawAdjList(fb)
	}
	if g.adj.detailOn {
		if spec := g.res.Screens["screen-admiral-detail"]; spec != nil && spec.Windows["overlay"] != nil {
			w := spec.Windows["overlay"]
			w.Draw(fb, g.res.Font, 3, g.res.Faces)
			w.DrawFactionText(fb, g.res.Font, g.faction)
			g.drawAdmiralDetailFor(fb, g.adj.detail, true)
		}
	}
}

// 「待命中副官資料」標題裡換掉的兩個字（副官，Big5 B0C6 A978）。
var adjListTitleCodes = []uint16{0xB0C6, 0xA978}

// drawAdjList 畫「待命中副官資料」：`screen-standby-admirals` 的 overlay，
// 標題「提督」換成「副官」，列是候選，從first起顯示最多十六列。
func (g *Game) drawAdjList(fb *gfx.Framebuffer) {
	spec := g.res.Screens["screen-standby-admirals"]
	if spec == nil || spec.Windows["overlay"] == nil {
		return
	}
	w := *spec.Windows["overlay"]
	// 標題那一行是 lines[0] 的七格：待命中提督資料 → 待命中副官資料。
	w.SetLineCode(0, 3, adjListTitleCodes[0])
	w.SetLineCode(0, 4, adjListTitleCodes[1])
	w.Draw(fb, g.res.Font, 3, g.res.Faces)
	if w.List == nil {
		return
	}
	rows := make([]screen.ListRow, 0, len(g.adj.list))
	for k, i := range g.adj.list {
		if k < g.adj.first {
			continue
		}
		rec := g.admiralRecord(i)
		if rec == nil {
			continue
		}
		var name []uint16
		if i < len(g.res.AdmiralNames) {
			name = g.res.AdmiralNames[i]
		}
		rows = append(rows, screen.ListRow{Name: name, Rank: rules.RankOf(rules.AdmiralLevel(rec)), RankCodes: g.admiralRankCodes(rec)})
	}
	w.List.Draw(fb, g.res.Font, rows)
}
