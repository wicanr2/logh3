package bootstrap_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
)

// 個人索引尾三格空白時，公開讀檔仍須能選取出貨的 S1／S2／S3。
func TestBundledScenariosSurvivePersonalIndex(t *testing.T) {
	orig := filepath.Join(root, "org_game/extracted")
	b, err := os.ReadFile(filepath.Join(orig, "SAVEDATA.IDX"))
	if err != nil {
		t.Skip("缺原版索引，不用代用品")
	}
	for i := 8 * rules.SaveEntry; i < len(b); i++ {
		b[i] = 0
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SAVEDATA.IDX"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ row, year, month, day int }{
		{8, 796, 10, 10}, {9, 798, 11, 19}, {10, 799, 3, 1},
	} {
		t.Run(fmt.Sprintf("S%d", c.row-7), func(t *testing.T) {
			g := loadGameSaveDir(t, dir)
			feed(g, fmt.Sprintf("click 300 270;type 138;enter;click 300 %d", 122+c.row*20))
			if g.State() != game.StateLoadFaction {
				t.Fatalf("額外劇本不能選取：%v", g.State())
			}
			feed(g, "click 318 208")
			if y, m, d := g.Date(); g.State() != game.StateCommandPhase || y != c.year || m != c.month || d != c.day {
				t.Fatalf("額外劇本未載入：%v %d/%d/%d", g.State(), y, m, d)
			}
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			// 同一個遊戲保存後重新開讀檔，三個出貨劇本仍可見。
			feed(g, "click 544 125;click 560 145")
			for row := 8; row < 11; row++ {
				feed(g, fmt.Sprintf("click 300 %d", 122+row*20))
				if g.State() != game.StateLoadFaction {
					t.Fatalf("保存後 S%d 消失：%v", row-7, g.State())
				}
				feed(g, "rclick 300 240")
			}
		})
	}
}

// 「存進度」存下來的那一格要讀得回來（使用者 2026-09-17：所有功能選項都要有作用）。
// 取進度第 k 格在 y ＝ 122 ＋ 20k（第 10 格 S3 在 (300,322)，`battledata_test.go`）；存進度那張低一列。
func TestSaveThenLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	g := loadGameSaveDir(t, dir)
	feed(g, monkeyStarts[0].script+"; move 600 340")
	apply(g, "execute") // 推進一回合，日期不是開局那一天
	y0, m0, d0 := g.Date()
	feed(g, "click 544 125; click 560 165")
	if g.State() != game.StateSaveGame {
		t.Fatalf("開存進度之後在 %v", g.State())
	}
	feed(g, "move 300 182; click 300 182") // 存進度第 2 格
	for _, f := range []string{"SAVEDATA.IDX", "WORK.DT2", "STELLAR.DT2"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatalf("存檔目錄沒有 %s：%v", f, err)
		}
	}

	h := loadGameSaveDir(t, dir)
	feed(h, monkeyStarts[0].script+"; move 600 340; click 544 125; click 560 145")
	if h.State() != game.StateLoadGame {
		t.Fatalf("開取進度之後在 %v", h.State())
	}
	feed(h, "move 300 162; click 300 162")
	if h.State() != game.StateLoadFaction {
		t.Fatal("缺讀檔陣營選單")
	}
	feed(h, "click 318 208") // 取進度第 2 格（存進度那張的列位置低一列）
	if y, m, d := h.Date(); y != y0 || m != m0 || d != d0 {
		t.Fatalf("讀回來的日期 %d/%d/%d，存的時候 %d/%d/%d", y, m, d, y0, m0, d0)
	}
}
