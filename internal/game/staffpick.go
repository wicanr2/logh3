package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 參謀任命（#62，`docs/re/342` §5、`docs/re/348`、規格 064）。
//
// 情報收集畫面的「參謀」鈕（`0A02:0AC4`）開的是**與出擊命令同一套**的
// 「待命中提督資料」清單與提督詳細資料視窗（`0C0A:00F0(1)` → `0D92:000D`），
// 差別在**堆疊方式**與第三顆鈕：
//
//   - 出擊那條路開詳細資料時**清單會關掉**（右半回到地圖介面，收據 `adm3e`）；
//     參謀這條路**清單留著**，詳細資料疊在它上面（收據 `fcmd-staff-detail`）。
//     所以這裡用的是副官那條路那種疊層（`drawAdjPick`），底層畫面不切狀態。
//   - 第三顆鈕在這條路上是「決定」：寫 `ds:8F4D[座位] ← 提督編號`，
//     回到情報收集就換成新的姓名與頭像。
//
// 按「中止」回清單、不是關窗（收據 `fcmd-staff-cancel`：`ds:8F4D` 留在原值）。

// staffPickedInk 是詳細資料開著時，清單上被選中那一列的底色（收據量到的色 2；
// 游標經過的反白是規格裡的色 8，兩者不同）。
const (
	staffPickedInk = 2
	// 選中那一列底板的範圍（收據量到露出來的是 x 544–608、高 16）。
	staffRowX     = 336
	staffRowRight = 608
	staffRowH     = 16
	// 詳細資料視窗第三顆鈕在這條路上是「決定」（Big5 A84D A977，與艦隊組成
	// 那張的「決定」同碼），出擊那條路是「出擊」（A558 C0BB）。
	staffDecideBox = 2
)

var staffDecideCodes = []uint16{0xA84D, 0xA977}

// staffPick 是參謀任命這條路上開著的東西。零值 ＝ 沒在任命。
type staffPick struct {
	listOn   bool // 「待命中提督資料」清單開著
	detailOn bool // 詳細資料視窗開著
	// post 是要換哪一個幕僚職（`rules.PostIntel` 等），back 是底層那一張畫面。
	post int
	back State
	// list 是候選提督，detail 是詳細視窗正在看的那一位。
	list   []int
	detail int
	first  int // 捲動起點，保持速度依089。
}

// on 回這條路上有沒有東西開著。
func (s staffPick) on() bool { return s.listOn || s.detailOn }

// beginStaffPick 從情報收集（之後還有徵稅計劃）的「參謀」鈕進來。
// **不切 `g.state`**——清單與詳細是疊在底層畫面上的。
func (g *Game) beginStaffPick(post int, back State) {
	if g.res.Screens["screen-standby-admirals"] == nil {
		return
	}
	staff := rules.StaffAdmirals(g.workRecord(), g.faction)
	list := rules.AdmiralList(rules.ParseAdmirals(g.res.Admiral), g.faction, true, staff)
	idx := make([]int, 0, len(list))
	for _, a := range list {
		idx = append(idx, a.Index)
	}
	g.staff = staffPick{listOn: true, post: post, back: back, list: idx, detail: -1}
}

// staffPickClick 是參謀任命疊層開著時的左鍵。吃掉了回 true。
func (g *Game) staffPickClick(x, y int) bool {
	switch {
	case g.staff.detailOn:
		if g.res.AdmiralPanel == nil {
			return true
		}
		switch g.res.AdmiralPanel.ButtonAt(x, y) {
		case "cancel": // 中止：回清單，什麼都不寫
			g.staff.detailOn = false
		case "sortie": // 這條路上是「決定」
			g.staffPickDecide()
		}
		return true
	case g.staff.listOn:
		if dir := g.adjListScrollAt(x, y); dir != 0 {
			g.staff.first = clampListOffset(g.staff.first+dir, len(g.staff.list), g.adjListPage())
			return true
		}
		if i := g.staffListAt(y); i >= 0 {
			g.staff.detail, g.staff.detailOn = g.staff.list[i], true
		}
		return true
	}
	return false
}

// staffPickRightClick 是右鍵：逐層關掉，最後回底層那一張。
func (g *Game) staffPickRightClick() bool {
	switch {
	case g.staff.detailOn:
		g.staff.detailOn = false
		return true
	case g.staff.listOn:
		g.staff = staffPick{}
		return true
	}
	return false
}

// staffPickDecide 把選中的提督寫進專任參謀表，整條收掉。
//
// 換**內政**專任參謀之後要把稅率那條鏈重跑一次（`090A:0B93` → `090A:0E6D`，
// 與動稅率檔位同一條）：收據 `fcmd-taxstaff-decide` 裡預估徵稅額 10341 → 10763、
// 成長率 3.2% → 3.3%、右上角的稅收與首都摘要都跟著變——因為稅率算式乘的就是
// 那一位的能力（`taxRateAdmiral`）。
func (g *Game) staffPickDecide() {
	post := g.staff.post
	if g.staff.detail >= 0 {
		rules.SetStaffAdmiral(g.mutableWork(), post, g.faction, g.staff.detail)
	}
	g.staff = staffPick{}
	if post == rules.PostDomestic {
		g.refreshBlend(g.faction, g.blend[g.faction])
		g.refreshCapitalSnap()
	}
}

// staffListAt 回滑鼠落在清單第幾列；不在回 −1。列的版面與
// 「待命中提督資料」那張相同（`screen-standby-admirals` 的 overlay）。
func (g *Game) staffListAt(y int) int {
	spec := g.res.Screens["screen-standby-admirals"]
	if spec == nil || spec.Windows["overlay"] == nil || spec.Windows["overlay"].List == nil {
		return -1
	}
	l := spec.Windows["overlay"].List
	if l.Pitch <= 0 || y < l.Y {
		return -1
	}
	row := (y - l.Y) / l.Pitch
	if row < 0 || row >= l.Rows || g.staff.first+row >= len(g.staff.list) {
		return -1
	}
	return g.staff.first + row
}

// drawStaffPick 把清單與詳細資料疊在底層畫面上（副官那條路的 `drawAdjPick`
// 同一種做法）。
func (g *Game) drawStaffPick(fb *gfx.Framebuffer) {
	if g.staff.listOn {
		g.drawStaffList(fb)
	}
	if g.staff.detailOn {
		if spec := g.res.Screens["screen-admiral-detail"]; spec != nil && spec.Windows["overlay"] != nil {
			w := *spec.Windows["overlay"]
			// 第三顆鈕在這條路上是「決定」不是「出擊」（`SetBoxCodes` 自己
			// copy-on-write，不會污染出擊那條路）。
			w.SetBoxCodes(staffDecideBox, staffDecideCodes)
			w.Draw(fb, g.res.Font, 3, g.res.Faces)
			w.DrawFactionText(fb, g.res.Font, g.faction)
			g.drawAdmiralDetailFor(fb, g.staff.detail, true)
		}
	}
}

// drawStaffList 畫「待命中提督資料」：與 `drawAdjList` 同一張規格，
// 但標題不用換（副官那條路才要改成「待命中副官資料」）。
func (g *Game) drawStaffList(fb *gfx.Framebuffer) {
	spec := g.res.Screens["screen-standby-admirals"]
	if spec == nil || spec.Windows["overlay"] == nil {
		return
	}
	w := *spec.Windows["overlay"]
	w.Draw(fb, g.res.Font, 3, g.res.Faces)
	if w.List == nil {
		return
	}
	// 詳細資料開著時，被選中那一列的底板塗**色 2**。這張規格沒有 `hover` 區塊
	// （游標經過不反白），所以自己畫。收據 `fcmd-staff-detail` 只量得到右側
	// 沒被詳細視窗遮住的那一段（x 544–608、高 16），左界被蓋住所以取整列。
	if g.staff.detailOn {
		for k, i := range g.staff.list {
			if i != g.staff.detail || k < g.staff.first {
				continue
			}
			y := w.List.Y + (k-g.staff.first)*w.List.Pitch
			fb.FillRect(staffRowX, y, staffRowRight-staffRowX, staffRowH, staffPickedInk)
			break
		}
	}
	rows := make([]screen.ListRow, 0, len(g.staff.list))
	for k, i := range g.staff.list {
		if k < g.staff.first {
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
