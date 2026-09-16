package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 艦隊視窗的「變更」（原版 `0DEA:17A9`，`docs/re/316`、規格 `docs/spec/051`
// 「變更」那一節）。
//
// 按下去先開一個兩項小選單 副官／編組（與作戰選單同一套 `3667`）：
// 「副官」開艦隊編成那張視窗的「副官／結束」變體（`297B:04EF` →
// `07F9:043F`），「編組」開這一支的艦隊組成（`0FF4:000A`）。子視窗開著時
// 「變更」鈕仍反白、小選單留在底下、選的那一項反白。
//
// 這一片只接到兩張子視窗的開與關：副官那張的「副官」鈕（變更／卸任）與
// 艦隊組成的編輯（點艦種、格盤、陣形、決定寫回編制）還沒接
// （issue #42「卡在」）。

// fleetChangeMenuKind 是規格檔裡那張兩項選單的名字。
const fleetChangeMenuKind screen.CommandMenuKind = "strategic-change"

// 兩項的索引（`ds:2232` 的順序）。
const (
	fleetChangeAdjutant = 0
	fleetChangeOrganise = 1
)

// openFleetChangeMenu 開小選單：`3667:0004(…, 欄 0x1D, 列 0x0F, 2 項, …)`，
// 位置與作戰選單同一條「游標所在格 − (1,1)」的公式（收據 `fcmd-change-menu`
// 從 (490,365) 點下去框在格 (29,15)）。
func (g *Game) openFleetChangeMenu(x, y int) {
	m := g.res.FleetActionMenu
	if m == nil || len(m.Menus[fleetChangeMenuKind].Items) == 0 {
		return
	}
	mx, my := m.CommandMenuAt(x, y-mouseDY, len(m.Menus[fleetChangeMenuKind].Items))
	g.fcmd.changeX, g.fcmd.changeY = mx, my+mouseDY
	g.fcmd.changeItem = -1
	g.state = StateFleetChangeMenu
	g.fleetChangeMenuHover(x, y)
}

// fleetChangeMenuHover 更新反白：跟著游標走（`fcmd-change-menu-hover`）。
func (g *Game) fleetChangeMenuHover(x, y int) {
	m := g.res.FleetActionMenu
	if m == nil {
		return
	}
	n := len(m.Menus[fleetChangeMenuKind].Items)
	g.fcmd.changeHover = m.CommandMenuRowAt(fleetChangeMenuKind, g.fcmd.changeX, g.fcmd.changeY, x, y, n)
}

// fleetChangeMenuClick 是小選單開著時的左鍵：兩項各開一張子視窗，點在外面關選單。
func (g *Game) fleetChangeMenuClick(x, y int) {
	g.fleetChangeMenuHover(x, y)
	item := g.fcmd.changeHover
	g.fcmd.changeHover = -1
	switch item {
	case fleetChangeAdjutant:
		g.fcmd.changeItem = item
		g.selectedRow = 0
		g.state = StateFleetChangeAdj
	case fleetChangeOrganise:
		g.fcmd.changeItem = item
		g.selectedRow = 0
		g.state = StateFleetChangeOrg
	default:
		g.closeFleetChangeMenu()
	}
}

// closeFleetChangeMenu 關小選單回艦隊視窗（`3667:00CB`；右鍵或點在外面，
// `fcmd-change-menu-rclick` 與 `fcmd-menu-rclick` 逐位元組相同）。
func (g *Game) closeFleetChangeMenu() {
	g.fcmd.changeHover, g.fcmd.changeItem = -1, -1
	g.state = StateFleetWindow
}

// fleetChangeAdjClick 是「副官」那張變體視窗的左鍵：只有 副官／結束 兩顆鈕
// （艦隊編號那顆沒畫）。「副官」→ 變更／卸任 還沒接；「結束」關窗。
func (g *Game) fleetChangeAdjClick(x, y int) {
	if g.res.FleetPanel == nil {
		return
	}
	switch g.res.FleetPanel.ButtonAt(x, y) {
	case "confirm": // 這張上面那顆印的是「結束」
		g.closeFleetChangeAdj()
	case "adjutant":
		// 變更／卸任小選單（`fcx-change-adj-menu`）：issue #42 後半。
	}
}

// closeFleetChangeAdj 是 `0DEA:1846` 從 `297B:04EF` 回來那一段：副官抄自提督
// 記錄的 `+0x07`、重印副官與費用那幾列、頭像；最後 `073F:0522` 把右下角清單
// 重畫成**沒按下**的樣子（收據 `fcmd-change-adj-close`：那一列底色回 8）。
func (g *Game) closeFleetChangeAdj() {
	f := &g.fleets[g.fcmd.fleet]
	if rec := g.admiralRecord(f.Admiral); rec != nil {
		f.Adjutant = rules.Adjutant(rec)
		f.Operations = g.fleetOperations(f.Admiral, f.Adjutant)
	}
	g.fcmd.listUnpressed = true
	g.fcmd.changeItem = -1
	g.state = StateFleetWindow
}

// fleetChangeOrgClick 是這一支的艦隊組成：中止／決定都回艦隊視窗（沒改編制時
// 決定與中止只差游標，`fcmd-change-org-decide`），陣形與格盤的編輯還沒接；
// 點七個艦種列的哪一列選取條就跳到哪一列（與出擊那條路同一張畫面）。
func (g *Game) fleetChangeOrgClick(x, y int) {
	p := g.res.CompositionPanel
	if p == nil {
		return
	}
	switch p.ButtonAt(x, y) {
	case "cancel", "decide":
		g.fcmd.changeItem = -1
		g.state = StateFleetWindow
		return
	case "formation":
		// 陣形清單：issue #42 後半。
		return
	}
	if spec := g.res.Screens["screen-fleet-composition"]; spec != nil {
		if w := spec.Windows["overlay"]; w != nil {
			if i := w.SelectRowAt(x, y); i >= 0 {
				g.selectedRow = i
			}
		}
	}
}

// drawFleetChange 畫小選單與（開著的話）子視窗。
func (g *Game) drawFleetChange(fb *gfx.Framebuffer) {
	m := g.res.FleetActionMenu
	if m == nil {
		return
	}
	sel := g.fcmd.changeHover
	if g.state != StateFleetChangeMenu {
		sel = g.fcmd.changeItem
	}
	screen.DrawCommandMenu(m, fb, g.res.Font, g.res.FrameTiles, fleetChangeMenuKind,
		nil, g.fcmd.changeX, g.fcmd.changeY, sel)
	switch g.state {
	case StateFleetChangeAdj:
		g.drawFleetChangeAdj(fb)
	case StateFleetChangeOrg:
		g.drawFleetChangeOrg(fb)
	}
}

// 「結束」那顆鈕的字（`結束`，Big5 B5B2 A7F4）。
var fleetChangeCloseCodes = []uint16{0xB5B2, 0xA7F4}

// drawFleetChangeAdj 把艦隊編成那張視窗的變體疊上去：同一張 `screen-fleet-org`
// 的 overlay，少了「艦隊編號」、「確認」印成「結束」（`fcmd-change-adj` 與
// `fleet1e` 只差這兩顆鈕與游標）。
func (g *Game) drawFleetChangeAdj(fb *gfx.Framebuffer) {
	spec := g.res.Screens["screen-fleet-org"]
	if spec == nil || spec.Windows["overlay"] == nil {
		return
	}
	w := *spec.Windows["overlay"]
	boxes := w.Boxes[:0:0]
	for _, b := range spec.Windows["overlay"].Boxes {
		switch b.Text {
		case "艦隊編號":
			continue
		case "確認":
			b.Text, b.Codes = "結束", fleetChangeCloseCodes
		}
		boxes = append(boxes, b)
	}
	w.Boxes = boxes
	w.Draw(fb, g.res.Font, 3, g.res.Faces)
	w.DrawFactionText(fb, g.res.Font, g.faction)
	f := &g.fleets[g.fcmd.fleet]
	g.drawFleetOrgFor(fb, f.Admiral, f.Adjutant, f.Number)
}

// drawFleetChangeOrg 把這一支的艦隊組成疊上去：`screen-fleet-composition` 的
// overlay（第 0 列選取）＋ 這一支的編制與所在星域的待命部隊。
func (g *Game) drawFleetChangeOrg(fb *gfx.Framebuffer) {
	spec := g.res.Screens["screen-fleet-composition"]
	if spec == nil || spec.Windows["overlay"] == nil || g.res.CompositionPanel == nil {
		return
	}
	w := spec.Windows["overlay"]
	w.SetSelected(g.selectedRow)
	w.Draw(fb, g.res.Font, 3, g.res.Faces)
	f := &g.fleets[g.fcmd.fleet]
	rec := g.regionRecord(f.Region)
	if rec == nil {
		return
	}
	v := screen.CompositionView{
		Ships:    rules.RegionShips(rec, f.Faction),
		PerUnit:  rules.ShipsPerUnit[:],
		Mobility: f.Morale,
		Groups:   compositionViewGroups(f.Composition),
		Slots:    f.Expected,
	}
	g.res.CompositionPanel.Draw(fb, g.res.Font, v)
	g.drawCompositionGridWith(fb, spec, v)
}
