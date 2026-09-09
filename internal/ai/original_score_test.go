package ai

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 這一組測的是 `2E54:1574`／`2E54:143F` 的**形狀**——三組模式常數、
// 各個乘除項有沒有生效。**沒有收據**：`s3aicomp` 那一趟走到的是階段四，
// 沒有經過候選評分，所以這裡釘的是反組譯讀出來的算式，不是原版的輸出
// （`docs/spec/029`、`VERIFICATION-MATRIX.md` 都照這樣記）。

func testGraph(t *testing.T) *rules.StarGraph {
	t.Helper()
	g, err := rules.LoadStarGraph(filepath.Join("..", "..", "docs", "internals", "data", "star-graph.json"))
	if err != nil {
		t.Fatalf("讀星域圖：%v", err)
	}
	return g
}

// base 是一支「什麼加成都沒有」的候選：滿編、士氣 100、不順路。
func base() Candidate {
	return Candidate{
		FleetView: FleetView{Faction: 0, Region: 0, Target: rules.NoTarget, Action: 5, Groups: 10},
		Morale:    100, Expected: 10, Class: 0, Ships: 1000, MainSlots: 5,
	}
}

func TestPickerModeConstants(t *testing.T) {
	// docs/re/168 的三組
	for i, want := range []PickerMode{
		{Weights: [4]int{3, 2, 8, 1}, Integrity: 50, NeedFull: false, MainMin: 0},
		{Weights: [4]int{4, 6, 1, 2}, Integrity: 90, NeedFull: true, MainMin: 0},
		{Weights: [4]int{3, 4, 0, 0}, Integrity: 80, NeedFull: true, MainMin: 4},
	} {
		if PickerModes[i] != want {
			t.Errorf("模式 %d ＝ %+v，預期 %+v", i, PickerModes[i], want)
		}
	}
}

func TestScoreZeroCases(t *testing.T) {
	g := testGraph(t)
	m := PickerModes[0]

	c := base()
	c.Disabled = true
	if got := m.Score(c, 5, g); got != 0 {
		t.Errorf("無法戰鬥的分數 ＝ %d，預期 0", got)
	}

	c = base()
	c.Region = 99 // 圖外 → Hops 回 -1
	if got := m.Score(c, 5, g); got != 0 {
		t.Errorf("走不到的分數 ＝ %d，預期 0", got)
	}

	c = base()
	c.Ships = 0
	if got := m.Score(c, 5, g); got != 1 {
		t.Errorf("算出來是 0 應該夾成 1，得到 %d", got)
	}
}

func TestScoreMultipliers(t *testing.T) {
	g := testGraph(t)
	m := PickerModes[0] // 權重 3 2 8 1、主力門檻 0

	c := base()
	plain := m.Score(c, 5, g)
	if plain <= 0 {
		t.Fatalf("基準分數 ＝ %d", plain)
	}

	// 順路（+0x08 != 5 且 +0x07 == 目標）→ ×2
	c2 := base()
	c2.Action, c2.Target = 0, 5
	if got := m.Score(c2, 5, g); got != plain*2 {
		t.Errorf("順路分數 ＝ %d，預期 %d", got, plain*2)
	}

	// 類別換一個 → 換一個權重（0 是 3、2 是 8）
	c3 := base()
	c3.Class = 2
	if got, want := m.Score(c3, 5, g), plain/3*8; got != want {
		t.Errorf("類別 2 的分數 ＝ %d，預期 %d", got, want)
	}

	// 距離越遠分數越低
	near, far := m.Score(base(), 1, g), m.Score(base(), 20, g)
	if !(near > far) {
		t.Errorf("近的 %d 沒有比遠的 %d 高", near, far)
	}

	// 主力槽不足 → ÷4（模式 2 的門檻是 4）
	m2 := PickerModes[2]
	full := base()
	full.MainSlots, full.Morale = 4, 100
	short := full
	short.MainSlots = 3
	if a, b := m2.Score(full, 5, g), m2.Score(short, 5, g); a != b*4 {
		t.Errorf("主力足 %d 不是不足 %d 的四倍", a, b)
	}
}

func TestEligibleThresholds(t *testing.T) {
	c := base()
	c.Morale = 89
	if PickerModes[1].Eligible(c, 5) {
		t.Error("士氣 89 不該過模式 1 的 90 門檻")
	}
	c.Morale = 90
	if !PickerModes[1].Eligible(c, 5) {
		t.Error("士氣 90 應該過模式 1")
	}

	// 模式 2：門檻 80，但艦隊已經在目標星域時降到 50
	c = base()
	c.Morale, c.Region = 60, 7
	if PickerModes[2].Eligible(c, 5) {
		t.Error("士氣 60、不在目標星域，不該過模式 2 的 80 門檻")
	}
	c.Region = 5
	if !PickerModes[2].Eligible(c, 5) {
		t.Error("士氣 60 但已經在目標星域，應該過模式 2 的 50 門檻")
	}

	// 已經決定過的一律不算候選
	c = base()
	c.Decided = true
	if PickerModes[0].Eligible(c, 5) {
		t.Error("+0x79 != 0xFF 的不該是候選")
	}

	// 模式 1 要求編制不缺
	c = base()
	c.Groups, c.Expected = 8, 10
	if PickerModes[1].Eligible(c, 5) {
		t.Error("編制缺的不該過模式 1")
	}
	if !PickerModes[0].Eligible(c, 5) {
		t.Error("模式 0 不要求編制不缺")
	}
}

func TestPickTakesHighestScore(t *testing.T) {
	g := testGraph(t)
	weak, strong := base(), base()
	weak.Ships, strong.Ships = 500, 2000
	got := PickerModes[0].Pick([]Candidate{weak, strong}, 5, g)
	if got != 1 {
		t.Errorf("挑到第 %d 支，預期第 1 支（分數高的那個）", got)
	}
	if got := PickerModes[0].Pick(nil, 5, g); got != -1 {
		t.Errorf("空清單應該回 -1，得到 %d", got)
	}
}
