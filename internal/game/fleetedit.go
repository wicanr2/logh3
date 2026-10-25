package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 艦隊組成編輯器（原版 `0FF4:000A`，`docs/re/317`、規格 `docs/spec/015`
// 「編輯」一節）——出擊命令那條路（`StateFleetComposition`／
// `StateFormation`）與艦隊視窗「變更 → 編組」（`StateFleetChangeOrg`）
// 共用這一份：規則在 `rules.FleetEditor`，這裡只管把點擊換成隊形槽、
// 艦種列、陣形項，以及跳訊息框。
//
// 座標都是畫面座標（收據旗標的滑鼠 y ＋ 40）。格盤一格 16 px，隊形槽
// (欄, 列) 佔 (136 ＋ 16 欄, 168 ＋ 16 列) 起 16 × 16（`0FF4:0C06` 拿
// `ds:28B8` 逐槽比）；七個艦種列的熱區是 (336,160) 起 64 × 140
// （`ds:2A00` 第 0 筆，列號 ＝ 格列 − 2）。

// openFleetEditor 開一次編輯：region 是待命部隊從哪一格扣、expected 是
// 艦隊 `+0x0E`、comp 是進來那一刻的編制（旗艦在槽 20）。
func (g *Game) openFleetEditor(region, faction, expected int, comp []rules.FleetGroup) {
	rec := g.regionRecord(region)
	if rec == nil {
		g.editor = nil
		return
	}
	g.editor = rules.NewFleetEditor(rules.BuildInput{
		Region: rec, Faction: faction, Kinds: g.res.ShipKinds, Tables: g.res.Composition,
	}, g.res.EditorTables, expected, comp)
	g.editor.NonBattleshipFlagship = g.nonBattleshipFlagship
	g.selectedRow = 0
	g.editor.Kind = 0
}

// closeFleetEditor 丟掉工作表（決定或中止之後）。
func (g *Game) closeFleetEditor() {
	g.editor = nil
}

// editorGridSlotAt 回畫面座標 (x,y) 落在格盤的哪一個隊形槽；沒有回 -1
// （`0FF4:0C06`：每一槽拿 `ds:28B8` 的格子比 16 × 16）。
func (g *Game) editorGridSlotAt(spec *screen.MainGameSpec, x, y int) int {
	p := g.res.CompositionPanel
	if p == nil || spec == nil || spec.Windows["overlay"] == nil || spec.Windows["overlay"].Grid == nil {
		return -1
	}
	gr := spec.Windows["overlay"].Grid
	for _, c := range p.GridSlots.Cells {
		cx, cy := gr.X+c.Col*gr.Tile, gr.Y+c.Row*gr.Tile
		if x >= cx && x < cx+gr.Tile && y >= cy && y < cy+gr.Tile {
			return c.Index
		}
	}
	return -1
}

// editorRowAt 回畫面座標落在七個艦種列的哪一列；沒有回 -1
// （熱區 `ds:2A00` 第 0 筆：欄 15–18、列 2–8，比框左右各窄 2 px）。
func (g *Game) editorRowAt(x, y int) int {
	p := g.res.CompositionPanel
	if p == nil {
		return -1
	}
	return p.RowAt(x, y)
}

// fleetEditorClick 是編輯器開著（訊息框沒開、陣形選單沒開）時，落在按鈕
// 以外的左鍵：艦種列換反白（`0FF4:0909`）、格盤放一組（`0FF4:0C80`）。
// 回有沒有吃掉這一下。
func (g *Game) fleetEditorClick(spec *screen.MainGameSpec, x, y int) bool {
	if g.editor == nil {
		return false
	}
	if i := g.editorRowAt(x, y); i >= 0 {
		g.selectedRow = i
		g.editor.Kind = i
		return true
	}
	if slot := g.editorGridSlotAt(spec, x, y); slot >= 0 {
		g.editorMessage(g.editor.Place(slot))
		return true
	}
	return false
}

// fleetEditorRightClick 是編輯器開著時的右鍵（`0FF4:0B85`）：只認格盤，
// 拆掉那一格（`0FF4:0DB1`）；旗艦那一格跳「無法撤離旗艦。」。
func (g *Game) fleetEditorRightClick(spec *screen.MainGameSpec, x, y int) {
	if g.editor == nil {
		return
	}
	if slot := g.editorGridSlotAt(spec, x, y); slot >= 0 {
		g.editorMessage(g.editor.Remove(slot))
	}
}

// editorMessage 把放／拆的結果換成訊息框（樣式 0，只有「確認」）。
func (g *Game) editorMessage(r rules.EditResult) {
	if name := r.MessageName(); name != "" {
		g.showMessage(name)
		if r == rules.EditFlagshipKind && g.nonBattleshipFlagship && g.msgOn {
			// Spec150 remake text: 修護艦與運補艦不能當旗艦。 Big5, not original bytes.
			g.msgRaw = []byte{0xad, 0xd7, 0xc5, 0x40, 0xc4, 0xa5, 0xbb, 0x50, 0xb9, 0x42, 0xb8, 0xc9, 0xc4, 0xa5, 0xa4, 0xa3, 0xaf, 0xe0, 0xb7, 0xed, 0xba, 0x58, 0xc4, 0xa5, 0xa1, 0x43}
		}
	}
}

// editorFormationRowAt 回畫面座標落在陣形清單的哪一項；沒有回 -1
// （`screen-formation` 的 `formation` 視窗那 8 列反白條）。
func (g *Game) editorFormationRowAt(x, y int) int {
	spec := g.res.Screens["screen-formation"]
	if spec == nil || spec.Windows["formation"] == nil {
		return -1
	}
	return spec.Windows["formation"].HoverRowAt(x, y)
}

// fleetEditorFormationPick 是陣形清單開著時的左鍵（`3667:00F2` 回選到的那一項，
// 點在外面回 0xFF）：選到就整批套用（`0FF4:0A4B`）；不管有沒有選到，
// 清單都關掉（`3667:00CB`）。回有沒有選到。
func (g *Game) fleetEditorFormationPick(x, y int) bool {
	i := g.editorFormationRowAt(x, y)
	if i < 0 || g.editor == nil {
		return false
	}
	g.editor.ApplyFormation(i)
	return true
}

// editorView 是畫這張畫面要的一整組值：待命部隊照那一格現在的存量、
// 編隊照工作表、機動力照組數（`297B:083F`）。
func (g *Game) editorView(admiral, adjutant int) screen.CompositionView {
	e := g.editor
	rec := e.In.Region
	groups := e.Groups()
	ops := rules.AdmiralAbility(g.admiralRecord(admiral), rules.AdmiralOperationsAbility)
	if adjutant >= 0 {
		if a := rules.AdmiralAbility(g.admiralRecord(adjutant), rules.AdmiralOperationsAbility); a > ops {
			ops = a
		}
	}
	var table []int
	if e.Tables != nil {
		table = e.Tables.Mobility
	}
	return screen.CompositionView{
		Ships:    rules.RegionShips(rec, e.In.Faction),
		PerUnit:  rules.ShipsPerUnit[:],
		Mobility: rules.FleetMobility(table, ops, len(groups)),
		Groups:   compositionViewGroups(groups),
		Slots:    e.Expected,
	}
}

// drawEditorFormationMenu 把陣形清單疊在編輯器上（`screen-formation` 的
// `formation` 視窗，反白跟游標）。出擊那條路整張畫面規格就是它，這裡給
// 「變更 → 編組」那條用。
func (g *Game) drawEditorFormationMenu(fb *gfx.Framebuffer) {
	spec := g.res.Screens["screen-formation"]
	if spec == nil || spec.Windows["formation"] == nil {
		return
	}
	w := spec.Windows["formation"]
	w.SetHoverRow(w.HoverRowAt(g.mouseX, g.mouseY))
	w.Draw(fb, g.res.Font, 3, g.res.Faces)
}
