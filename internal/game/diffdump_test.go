package game

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpSaveDiff 只在需要時手動跑：把 save1e 的差異列出來。
func TestDumpSaveDiff(t *testing.T) {
	if os.Getenv("LOGH3_DIFF") == "" {
		t.Skip("設 LOGH3_DIFF=1 才跑")
	}
	batch := os.Getenv("LOGH3_DIFF")
	g := reachCommandPhase(t)
	g.Click(544, 125)
	g.Click(560, 165)
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "screen.bin"))
	if err != nil {
		t.Fatal(err)
	}
	got := g.Frame().Pix
	rows := map[int]int{}
	minx, maxx, miny, maxy := 1<<30, -1, 1<<30, -1
	n := 0
	for i := range want {
		if want[i] != got[i] {
			x, y := i%640, i/640
			rows[y]++
			n++
			if x < minx { minx = x }
			if x > maxx { maxx = x }
			if y < miny { miny = y }
			if y > maxy { maxy = y }
		}
	}
	t.Logf("%s 差 %d 像素 bbox x%d-%d y%d-%d", batch, n, minx, maxx, miny, maxy)
	for y := miny; y <= maxy; y++ {
		if rows[y] == 0 {
			continue
		}
		var s string
		for x := 0; x < 640; x++ {
			i := y*640 + x
			if want[i] != got[i] {
				s += fmt.Sprintf(" x%d(want %d got %d)", x, want[i], got[i])
			}
		}
		t.Logf("y=%d 共 %d:%s", y, rows[y], s)
	}
}
