package main

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
	"image"
	"os"
	"strconv"
)

var editorBounds = image.Rect(24, 40, 616, 440)
var editorTabs = [3]image.Rectangle{image.Rect(204, 168, 328, 193), image.Rect(332, 168, 456, 193), image.Rect(460, 168, 584, 193)}
var editorActions = [3]image.Rectangle{image.Rect(36, 402, 129, 427), image.Rect(414, 402, 507, 427), image.Rect(515, 402, 608, 427)}
var editorLabels = []string{"統率力", "移動", "回轉", "營運", "光線砲", "飛彈", "戰鬥機", "防禦", "占領", "運用", "內政／應變參數", "稅務／決策權重", "目標選擇判斷", "旗艦擊毀換艦率", "全軍移動加成", "全軍攻擊加成", "本艦隊戰鬥機攻擊"}
var editorDescriptions = []string{"影響稅率、治安分段、AI艦隊類別", "及低完整度時的戰術判斷。", "影響內政參謀評分、稅率", "與戰術AI權重。", "愈高愈容易選中評分最高的攻擊目標。", "", "旗艦被擊毀時的換艦成功機率。", "", "參戰時，同陣營全部艦隊的移動增幅。", "", "參戰時，同陣營全部艦隊的攻擊增幅。", "", "任提督或參謀時，所在艦隊的戰鬥機攻擊增幅。", ""}

type admiralEditorState struct {
	Entries                        []game.AdmiralEditorEntry
	Draft                          rules.AdmiralPreferences
	Faction, Selected, Tab, Scroll int
	Dirty                          bool
	Editing                        int
	Input, Error, Notice           string
	Confirm                        string
	Changed                        map[int]bool
}

func readAdmiralPreferences(path string) (rules.AdmiralPreferences, error) {
	p := rules.NewAdmiralPreferences()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	var all map[string]json.RawMessage
	if json.Unmarshal(b, &all) != nil || all == nil {
		return p, fmt.Errorf("提督偏好檔格式無效")
	}
	if raw, ok := all["admiral_editor"]; ok {
		if string(raw) == "null" {
			return p, fmt.Errorf("提督偏好不得為空")
		}
		if err = json.Unmarshal(raw, &p); err != nil {
			return rules.NewAdmiralPreferences(), err
		}
		if err = p.Validate(); err != nil {
			return rules.NewAdmiralPreferences(), err
		}
	}
	return p, nil
}

func (a *app) openAdmiralEditor() {
	e := &admiralEditorState{Entries: a.g.AdmiralEditorEntries(), Draft: a.g.AdmiralPreferences(), Faction: a.g.Faction(), Editing: -1}
	if len(e.Entries) == 0 {
		return
	}
	a.admiralEditor = e
	e.selectFaction(e.Faction)
}
func (e *admiralEditorState) indices() []int {
	var out []int
	for i, v := range e.Entries {
		if v.Faction == e.Faction {
			out = append(out, i)
		}
	}
	return out
}
func (e *admiralEditorState) selectFaction(f int) {
	e.Faction = f
	e.Scroll = 0
	e.Selected = 0
	ids := e.indices()
	if len(ids) > 0 {
		e.Selected = ids[0]
	}
	if f == 1 && len(e.Entries) > 52 {
		e.Selected = 52
		e.Scroll = max(0, 12-7)
	}
}
func (e *admiralEditorState) value(field int) int {
	v := e.Entries[e.Selected]
	if field < 13 {
		return v.Abilities[field]
	}
	switch field {
	case 13:
		return v.Special.Transfer
	case 14:
		return v.Special.Movement
	case 15:
		return v.Special.Attack
	default:
		return v.Special.FighterAttack
	}
}
func (e *admiralEditorState) set(field, value int) bool {
	lo, hi := 1, 200
	if field >= 13 {
		lo = 0
	}
	if field == 13 {
		hi = 100
	}
	if value < lo || value > hi {
		e.Error = fmt.Sprintf("請輸入%d至%d", lo, hi)
		return false
	}
	v := &e.Entries[e.Selected]
	p := e.Draft.Admirals[e.Selected]
	if p.Abilities == nil {
		p.Abilities = map[int]int{}
	}
	if field < 13 {
		v.Abilities[field] = value
		p.Abilities[field] = value
	} else {
		switch field {
		case 13:
			v.Special.Transfer = value
		case 14:
			v.Special.Movement = value
		case 15:
			v.Special.Attack = value
		case 16:
			v.Special.FighterAttack = value
		}
		s := v.Special
		p.Special = &s
	}
	e.Draft.Admirals[e.Selected] = p
	e.Dirty = true
	if e.Changed == nil {
		e.Changed = map[int]bool{}
	}
	e.Changed[e.Selected] = true
	e.Error = ""
	e.Notice = "尚未套用"
	return true
}
func editorFieldRects(field int) [3]image.Rectangle {
	x, y := 0, 0
	if field < 10 {
		x = 204 + (field%2)*202
		y = 205 + (field/2)*33
	} else {
		x = 204
		i := field - 10
		if field >= 13 {
			i = field - 13
		}
		y = 205 + i*49
		if field >= 13 {
			y = 205 + i*40
		}
	}
	right := 390
	if field >= 10 {
		right = 592
	} else if field%2 == 1 {
		right += 202
	}
	w := 43
	if field >= 13 {
		w = 67
	}
	x = right - 21 - 3 - w - 3 - 21
	return [3]image.Rectangle{image.Rect(x, y, x+21, y+23), image.Rect(x+24, y, x+24+w, y+23), image.Rect(x+27+w, y, x+48+w, y+23)}
}
func (e *admiralEditorState) fields() []int {
	if e.Tab == 0 {
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	}
	if e.Tab == 1 {
		return []int{10, 11, 12}
	}
	return []int{13, 14, 15, 16}
}

func (a *app) editorApply() {
	e := a.admiralEditor
	if err := a.g.ValidateAdmiralPreferences(e.Draft); err != nil {
		e.Error = err.Error()
		return
	}
	b, err := json.Marshal(e.Draft)
	if err == nil && a.preferencesPath != "" {
		err = writePreferences(a.preferencesPath, map[string]json.RawMessage{"admiral_editor": b})
	}
	if err != nil {
		e.Error = "無法保存提督設定"
		return
	}
	if err = a.g.SetAdmiralPreferences(e.Draft); err != nil {
		e.Error = err.Error()
		return
	}
	e.Dirty = false
	e.Changed = nil
	e.Notice = "已保存全局設定"
	e.Error = ""
}
func (a *app) editorCancel() {
	e := a.admiralEditor
	if e == nil {
		return
	}
	if e.Editing >= 0 {
		e.Editing = -1
		e.Error = ""
		return
	}
	if e.Confirm != "" {
		e.Confirm = ""
		return
	}
	if e.Dirty {
		e.Confirm = "cancel"
	} else {
		a.admiralEditor = nil
	}
}
func (a *app) editorKey(c byte) {
	e := a.admiralEditor
	if e == nil {
		return
	}
	if c == 27 {
		a.editorCancel()
		return
	}
	if e.Editing < 0 {
		return
	}
	switch {
	case c >= '0' && c <= '9':
		if len(e.Input) < 3 {
			e.Input += string(c)
		}
	case c == 8:
		if len(e.Input) > 0 {
			e.Input = e.Input[:len(e.Input)-1]
		}
	case c == '\r':
		n, err := strconv.Atoi(e.Input)
		if err == nil && e.set(e.Editing, n) {
			e.Editing = -1
		} else {
			e.Error = "數值超出允許範圍"
		}
	}
}
func (a *app) editorClick(x, y int) {
	e := a.admiralEditor
	p := image.Pt(x, y)
	if e.Confirm != "" {
		if p.In(image.Rect(204, 267, 309, 292)) {
			e.Confirm = ""
		}
		if p.In(image.Rect(315, 267, 420, 292)) {
			if e.Confirm == "cancel" {
				a.admiralEditor = nil
				return
			}
			for f, n := range e.Entries[e.Selected].Defaults {
				e.set(f, n)
			}
			s := rules.DefaultAdmiralSpecial(e.Selected)
			e.set(13, s.Transfer)
			e.set(14, s.Movement)
			e.set(15, s.Attack)
			e.set(16, s.FighterAttack)
			e.Confirm = ""
		}
		return
	}
	if e.Editing >= 0 {
		if p.In(image.Rect(220, 268, 300, 292)) {
			e.Editing = -1
			e.Error = ""
		}
		if p.In(image.Rect(310, 268, 390, 292)) {
			a.editorKey('\r')
		}
		return
	}
	for f := 0; f < 2; f++ {
		if p.In(image.Rect(36+f*80, 76, 112+f*80, 99)) {
			e.selectFaction(f)
			return
		}
	}
	ids := e.indices()
	for i := 0; i < 8 && e.Scroll+i < len(ids); i++ {
		if p.In(image.Rect(40, 110+i*29, 184, 136+i*29)) {
			e.Selected = ids[e.Scroll+i]
			return
		}
	}
	if p.In(image.Rect(36, 362, 112, 382)) {
		e.Scroll = max(0, e.Scroll-3)
		return
	}
	if p.In(image.Rect(116, 362, 192, 382)) {
		e.Scroll = min(max(0, len(ids)-8), e.Scroll+3)
		return
	}
	for i, r := range editorTabs {
		if p.In(r) {
			e.Tab = i
			return
		}
	}
	for _, f := range e.fields() {
		for action, r := range editorFieldRects(f) {
			if !p.In(r) {
				continue
			}
			if action == 1 {
				e.Editing = f
				e.Input = ""
				e.Error = ""
			} else {
				step := 1
				if f >= 13 {
					step = 5
				}
				if action == 0 {
					step = -step
				}
				e.set(f, e.value(f)+step)
			}
			return
		}
	}
	if p.In(editorActions[0]) {
		e.Confirm = "reset"
	}
	if p.In(editorActions[1]) {
		a.editorCancel()
	}
	if p.In(editorActions[2]) {
		a.editorApply()
	}
}

func (a *app) drawAdmiralEditor(fb *gfx.Framebuffer) {
	e := a.admiralEditor
	if e == nil {
		return
	}
	panel := func(r image.Rectangle) {
		(ui.Frame{Role: "ui/window", Layers: []ui.Layer{{Top: 15, Right: 15, Bottom: 15, Left: 15}, {Top: 7, Right: 7, Bottom: 3, Left: 3}}, Fill: 9}).Draw(fb, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
	}
	text := func(r image.Rectangle, s string) { a.compactText(fb, r, s, 7) }
	button := func(r image.Rectangle, s string, active bool) {
		fill, ink := uint8(9), uint8(7)
		if active {
			fill = 2
			ink = 3
		}
		(ui.Frame{Role: "ui/button", Layers: []ui.Layer{{Top: 7, Right: 7, Bottom: 3, Left: 3}}, Fill: fill}).Draw(fb, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
		if r.Min.X == 40 && r.Dx() == 144 {
			labelBounds := r
			labelBounds.Min.X += 8
			a.editorText(fb, labelBounds.Inset(1), s, 12, true, ink)
		} else {
			a.compactText(fb, r.Inset(1), s, ink)
		}
		if image.Pt(a.g.MousePosition()).In(r) {
			fb.HoverControl = r
		}
	}
	panel(editorBounds)
	a.editorText(fb, image.Rect(36, 48, 604, 66), "提督能力編輯", 14, false, 7)
	for f, s := range []string{"銀河帝國", "自由同盟"} {
		button(image.Rect(36+f*80, 76, 112+f*80, 99), s, e.Faction == f)
	}
	ids := e.indices()
	for i := 0; i < 8 && e.Scroll+i < len(ids); i++ {
		n := ids[e.Scroll+i]
		prefix := "  "
		if e.Changed[n] {
			prefix = "* "
		}
		button(image.Rect(40, 110+i*29, 184, 136+i*29), prefix+e.Entries[n].Label, n == e.Selected)
	}
	button(image.Rect(36, 362, 112, 382), "▲", false)
	button(image.Rect(116, 362, 192, 382), "▼", false)
	v := e.Entries[e.Selected]
	a.g.DrawEditorPortrait(fb, e.Selected, 204, 77, 64, 80)
	a.editorText(fb, image.Rect(282, 80, 592, 100), v.Label, 16, true, 7)
	faction := "銀河帝國"
	if v.Faction == 1 {
		faction = "自由行星同盟"
	}
	a.editorText(fb, image.Rect(282, 104, 592, 122), fmt.Sprintf("%s　等級 %d", faction, v.Level), 12, true, 7)
	a.editorText(fb, image.Rect(282, 126, 592, 144), "全局共用，新遊戲也套用", 11, true, 7)
	for i, s := range []string{"基本能力", "內部能力", "特殊能力"} {
		button(editorTabs[i], s, e.Tab == i)
	}
	for _, f := range e.fields() {
		rs := editorFieldRects(f)
		x := 204
		if f < 10 {
			x += (f % 2) * 202
		}
		a.editorText(fb, image.Rect(x, rs[0].Min.Y, rs[0].Min.X-4, rs[0].Max.Y), editorLabels[f], 12, true, 7)
		n := fmt.Sprint(e.value(f))
		if f >= 13 {
			n += "%"
			if f > 13 {
				n = "+" + n
			}
		}
		button(rs[0], "-", false)
		button(rs[1], n, false)
		button(rs[2], "+", false)
		if f >= 10 {
			start := (f - 10) * 2
			if f >= 13 {
				start = 6 + (f-13)*2
			}
			a.editorText(fb, image.Rect(204, rs[0].Max.Y+2, 604, rs[0].Max.Y+18), editorDescriptions[start]+editorDescriptions[start+1], 10, true, 7)
		}
	}
	help := "點選數值輸入，或使用 -／+ 微調。"
	if e.Tab == 1 {
		help = "以上為用途說明，原版欄名未知。"
	}
	if e.Tab == 2 {
		help = "強化模式：全軍與戰鬥機加成各自乘算。"
	}
	a.editorText(fb, image.Rect(204, 371, 604, 395), help, 10, true, 7)
	notice := e.Notice
	if e.Error != "" {
		notice = e.Error
	}
	a.editorText(fb, image.Rect(36, 384, 192, 400), notice, 10, true, 7)
	for i, s := range []string{"還原本提督", "取消", "套用變更"} {
		button(editorActions[i], s, i == 2 && e.Dirty)
	}
	if e.Editing >= 0 {
		panel(image.Rect(203, 168, 438, 306))
		text(image.Rect(215, 179, 426, 198), editorLabels[e.Editing])
		button(image.Rect(270, 209, 371, 237), e.Input+"_", false)
		button(image.Rect(220, 268, 300, 292), "取消", false)
		button(image.Rect(310, 268, 390, 292), "確定", false)
		text(image.Rect(210, 242, 432, 258), e.Error)
	}
	if e.Confirm != "" {
		panel(image.Rect(180, 186, 460, 310))
		s := "還原本提督的預設能力？"
		if e.Confirm == "cancel" {
			s = "放棄尚未套用的修改？"
		}
		text(image.Rect(190, 213, 450, 236), s)
		button(image.Rect(204, 267, 309, 292), "返回編輯", false)
		button(image.Rect(315, 267, 420, 292), "確定", false)
	}
}

func (a *app) editorText(fb *gfx.Framebuffer, r image.Rectangle, text string, size int, left bool, ink uint8) {
	group := fb.BeginGroup()
	defer func() {
		fb.EndGroup(group, gfx.DrawCommand{Kind: gfx.DrawText, Text: text, X: r.Min.X, Y: r.Min.Y, W: r.Dx(), H: r.Dy(), Clip: r, Colour: ink, TextLeft: left})
	}()
	width := 0
	for _, c := range text {
		if c < 128 {
			width += size / 2
		} else {
			width += size
		}
	}
	x, y := r.Min.X, r.Min.Y+(r.Dy()-size)/2
	if !left {
		x += (r.Dx() - width) / 2
	}
	for _, c := range text {
		glyph, ok := font.SettingsGlyphSize(c, size)
		if !ok {
			continue
		}
		if x+glyph.Width > r.Max.X {
			break
		}
		start := fb.BeginGroup()
		for py := 0; py < size; py++ {
			for px := 0; px < glyph.Width; px++ {
				if glyph.Native[py*glyph.Width+px] >= 88 && image.Pt(x+px, y+py).In(r) {
					fb.Set(x+px, y+py, ink)
				}
			}
		}
		fb.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, Key: font.SettingsGlyphKey, X: x, Y: y, W: glyph.Width, H: size, Colour: ink, Clip: r, SourceW: glyph.Width * 3, SourceH: size * 3, Source: glyph.HD})
		x += glyph.Width
	}
}

func (a *app) admiralEditorObservation() any {
	e := a.admiralEditor
	if e == nil {
		return nil
	}
	controls := map[string][4]int{}
	add := func(name string, r image.Rectangle) { controls[name] = [4]int{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y} }
	for i, s := range []string{"basic", "hidden", "special"} {
		add(s, editorTabs[i])
	}
	for i, s := range []string{"restore", "cancel", "apply"} {
		add(s, editorActions[i])
	}
	for f := 0; f < 2; f++ {
		add(fmt.Sprint("faction-", f), image.Rect(36+f*80, 76, 112+f*80, 99))
	}
	ids := e.indices()
	for i := 0; i < 8 && e.Scroll+i < len(ids); i++ {
		add(fmt.Sprintf("admiral-%d", ids[e.Scroll+i]), image.Rect(40, 110+i*29, 184, 136+i*29))
	}
	add("scroll-down", image.Rect(116, 362, 192, 382))
	for _, f := range e.fields() {
		for i, s := range []string{"dec", "value", "inc"} {
			add(fmt.Sprintf("field-%d-%s", f, s), editorFieldRects(f)[i])
		}
	}
	add("input-ok", image.Rect(310, 268, 390, 292))
	add("input-cancel", image.Rect(220, 268, 300, 292))
	add("confirm-yes", image.Rect(315, 267, 420, 292))
	add("confirm-no", image.Rect(204, 267, 309, 292))
	return map[string]any{"selected": e.Selected, "faction": e.Faction, "tab": e.Tab, "dirty": e.Dirty, "editing": e.Editing, "error": e.Error, "notice": e.Notice, "confirm": e.Confirm, "entry": e.Entries[e.Selected], "controls": controls, "game_entry": a.g.AdmiralEditorEntries()[e.Selected]}
}
