package main

import (
	"bytes"
	"encoding/json"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestAdmiralPreferencesRestartAndUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	p, err := readAdmiralPreferences(path)
	if err != nil || p.Special(52).Attack != 30 {
		t.Fatal(p, err)
	}
	p.Admirals[52] = rules.AdmiralOverride{Abilities: map[int]int{10: 145}, Special: &rules.AdmiralSpecial{Transfer: 40, Movement: 70, Attack: 30}}
	b, _ := json.Marshal(p)
	if err = writePreferences(path, map[string]json.RawMessage{"admiral_editor": b, "other": json.RawMessage(`42`)}); err != nil {
		t.Fatal(err)
	}
	if err = writeSoundFontPreference(path, "mt32"); err != nil {
		t.Fatal(err)
	}
	q, err := readAdmiralPreferences(path)
	if err != nil || q.Special(52).Movement != 70 || q.Admirals[52].Abilities[10] != 145 {
		t.Fatal("重開或其他設定覆寫提督偏好", q, err)
	}
	s := rules.DefaultAdmiralSpecial(74)
	s.FighterAttack = 65
	q.Admirals[74] = rules.AdmiralOverride{Special: &s}
	b, _ = json.Marshal(q)
	if err = writePreferences(path, map[string]json.RawMessage{"admiral_editor": b}); err != nil {
		t.Fatal(err)
	}
	q, err = readAdmiralPreferences(path)
	if err != nil || q.Special(74).FighterAttack != 65 {
		t.Fatal("戰鬥機加成未跨重啟保存", q, err)
	}
	data, _ := os.ReadFile(path)
	var all map[string]json.RawMessage
	json.Unmarshal(data, &all)
	if string(all["other"]) != "42" {
		t.Fatal("未知鍵遺失")
	}
	for _, bad := range []string{`{"admiral_editor":null}`, `{"admiral_editor":{"version":2}}`, `{"admiral_editor":{"version":1,"admirals":{"82":{}}}}`} {
		os.WriteFile(path, []byte(bad), 0600)
		if _, err = readAdmiralPreferences(path); err == nil {
			t.Fatal("接受壞偏好", bad)
		}
	}
}

func TestEditorCancelInputAndFailurePreserveState(t *testing.T) {
	if _, err := os.Stat("../../org_game/extracted/GIN3.COM"); os.IsNotExist(err) {
		t.Skip("缺原版資料")
	}
	res, err := bootstrap.Load("../../org_game/extracted", "../../docs/internals/data")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 82; i++ {
		for _, off := range rules.AdmiralEditableOffsets {
			res.Admiral[i*rules.AdmiralRecord+off] = 100
		}
	}
	g, err := game.New(res)
	if err != nil {
		t.Fatal(err)
	}
	a := &app{g: g, preferencesPath: filepath.Join(t.TempDir(), "preferences.json")}
	a.openAdmiralEditor()
	e := a.admiralEditor
	e.set(0, 150)
	a.editorCancel()
	if e.Confirm != "cancel" || g.AdmiralEditorEntries()[0].Abilities[0] != 100 {
		t.Fatal("草稿修改穿透Game")
	}
	a.editorClick(350, 280)
	if a.admiralEditor != nil {
		t.Fatal("放棄未返回設定")
	}
	a.openAdmiralEditor()
	e = a.admiralEditor
	e.Editing = 0
	e.Input = "0"
	a.editorKey('\r')
	if e.Editing != 0 || e.Error == "" {
		t.Fatal("接受零能力")
	}
	e.Editing = -1
	e.set(0, 150)
	os.WriteFile(a.preferencesPath, []byte(`{"bad":`), 0600)
	a.editorApply()
	if e.Error == "" || g.AdmiralEditorEntries()[0].Abilities[0] != 100 {
		t.Fatal("壞檔仍提交Game")
	}
}

func TestEditorGeometryAndSizedTextReplay(t *testing.T) {
	for f := range editorLabels {
		for _, r := range editorFieldRects(f) {
			if !r.In(editorBounds.Inset(4)) {
				t.Fatal("控制超出編輯窗", f, r)
			}
		}
	}
	a := &app{}
	fb := gfx.NewFramebuffer(9)
	for i, size := range []int{10, 11, 12, 14, 16} {
		r := image.Rect(280, 80+i*28, 590, 100+i*28)
		a.editorText(fb, r, "萊因哈特 楊威利 30%", size, true, 7)
	}
	if !bytes.Equal(fb.Pix, gfx.ReplayIndexed(fb.Commands)) {
		t.Fatal("大小字形與原色重播不一致")
	}
	for _, c := range fb.Commands {
		if c.Kind == gfx.DrawGlyph {
			if c.Key != font.SettingsGlyphKey || c.SourceW != c.W*3 || c.SourceH != c.H*3 {
				t.Fatal("未使用設定字形遮罩")
			}
			if !image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H).In(c.Clip) {
				t.Fatal("文字安全矩形越界")
			}
		}
	}
}
