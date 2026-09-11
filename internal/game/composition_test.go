package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/screen"
)

// loadCompositionPanel 直接讀版面 JSON，不建整個遊戲。
func loadCompositionPanel(t *testing.T) *screen.CompositionPanel {
	t.Helper()
	p, err := screen.LoadCompositionPanel(filepath.Join(root,
		"docs/internals/data/fleet-composition-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestCompositionLettersMatchShipKindLetters 編隊那一行的字母表
// （原版 `ds:6172`，八筆遠指標各指一個全形字母）與戰場「資料」那張
// 用的半形 `BHCADTE`（線性 `0x5D94A`，`docs/re/111`、規格 021）
// **是同一組**：前七個逐一相同。
//
// 兩張表是從不同地方讀出來的，所以這條測試是**互相驗證**，不是重述：
// `BHCADTE` 那一組有收據（戰場資料畫面逐位元組相同），`ds:6172` 那一組
// 只在艦隊組成畫面出現、收據裡只看得到第 0 格。
func TestCompositionLettersMatchShipKindLetters(t *testing.T) {
	p := loadCompositionPanel(t)
	if got := len(p.UnitLetters.Codes); got != 8 {
		t.Fatalf("字母表 %d 筆，應該是 8 筆（七種艦艇 ＋ 艦種 7 偵察隊）", got)
	}
	for kind := 0; kind < len(shipKindLetters); kind++ {
		want := shipKindLetter(kind)
		got, ok := p.UnitLetter(kind)
		if !ok || got != want {
			t.Errorf("艦種 %d：編隊字母 %04X，BHCADTE 那張是 %04X", kind, got, want)
		}
	}
	// 第八筆是Ｐ ＝ 艦種 7（偵察隊，`docs/re/290`）——`BHCADTE` 沒有它。
	if got, _ := p.UnitLetter(7); got != uint16(b5FullA+'P'-'A') {
		t.Errorf("艦種 7 的字母 %04X，應該是Ｐ %04X", got, b5FullA+'P'-'A')
	}
	if _, ok := p.UnitLetter(8); ok {
		t.Error("艦種 8 不該查得到字母")
	}
}

// TestCompositionStrengthInk 字母的顏色由艘數決定（`0FF4:03FE`）：
// 滿編 7、不到一半 2、其餘 5。收據 `fok` 那一格是滿編的 7。
func TestCompositionStrengthInk(t *testing.T) {
	p := loadCompositionPanel(t)
	for _, c := range []struct {
		ships, full int
		want        uint8
	}{
		{3000, 3000, 7}, // 滿編
		{2999, 3000, 5}, // 差一艘
		{1500, 3000, 5}, // 剛好一半：原版比的是「艘數 × 2 < 滿編」
		{1499, 3000, 2},
		{0, 3000, 2},
		{0, 0, 7}, // 兩個都 0 走的是「相等」那一條
	} {
		if got := p.StrengthInkFor(c.ships, c.full); got != c.want {
			t.Errorf("%d/%d → 色 %d，應該是 %d", c.ships, c.full, got, c.want)
		}
	}
}

// TestCompositionGridSlots 41 個隊形槽落在格盤 9 × 9 的哪一格
// （原版 `ds:28B8`）：每一格 `(欄 ＋ 列)` 都是偶數、沒有兩個槽撞格、
// 中央 (4,4) 是槽 20（旗艦）。
func TestCompositionGridSlots(t *testing.T) {
	p := loadCompositionPanel(t)
	if got := len(p.GridSlots.Cells); got != 41 {
		t.Fatalf("槽位 %d 個，應該是 41 個", got)
	}
	seen := map[[2]int]int{}
	for _, c := range p.GridSlots.Cells {
		if c.Col < 0 || c.Col > 8 || c.Row < 0 || c.Row > 8 {
			t.Fatalf("槽 %d 在 (%d,%d)，超出 9 × 9", c.Index, c.Col, c.Row)
		}
		if (c.Col+c.Row)%2 != 0 {
			t.Errorf("槽 %d 在 (%d,%d)：(欄 ＋ 列) 不是偶數", c.Index, c.Col, c.Row)
		}
		if prev, dup := seen[[2]int{c.Col, c.Row}]; dup {
			t.Errorf("槽 %d 與槽 %d 撞在 (%d,%d)", c.Index, prev, c.Col, c.Row)
		}
		seen[[2]int{c.Col, c.Row}] = c.Index
	}
	col, row, ok := p.GridCell(20)
	if !ok || col != 4 || row != 4 {
		t.Errorf("槽 20 在 (%d,%d)，應該是格盤正中央 (4,4)", col, row)
	}
}
