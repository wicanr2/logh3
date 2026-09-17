package bootstrap_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/game"
)

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
	feed(h, "move 300 162; click 300 162") // 取進度第 2 格（存進度那張的列位置低一列）
	if y, m, d := h.Date(); y != y0 || m != m0 || d != d0 {
		t.Fatalf("讀回來的日期 %d/%d/%d，存的時候 %d/%d/%d", y, m, d, y0, m0, d0)
	}
}
