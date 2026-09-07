package game

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestDiffAgainstOracle 是診斷工具，不是驗收：把某一張畫面與收據的差異
// **逐點列出來**，而不是只報一個數字。
//
//	LOGH3_DIFF="save1e:544,125 560,165" tools/go.sh test ./internal/game/ -run TestDiffAgainstOracle -v
//	LOGH3_DIFF="info1e:352,125 400,145" LOGH3_REGION=356,142,376,164 ... （改列指定矩形的兩邊色號）
//
// 差幾個像素由 `TestCommandPhaseNavigation` 負責判定；這支只負責把
// 「差在哪、原版是什麼色、remake 畫成什麼色」攤開來。逐點看得到顏色，
// 才分得出「位置錯」與「被別的東西蓋掉」——只看總數兩者長得一樣。
func TestDiffAgainstOracle(t *testing.T) {
	spec := os.Getenv("LOGH3_DIFF")
	if spec == "" {
		t.Skip("設 LOGH3_DIFF=<批次>:<x,y> <x,y>… 才跑")
	}
	batch, clicks, _ := strings.Cut(spec, ":")
	if n, rest, ok := strings.Cut(clicks, "x"); ok {
		// "10x544,125 560,125 | 440,368" ＝ 前面那組重複 N 次
		times, _ := strconv.Atoi(n)
		head, tail, _ := strings.Cut(rest, "|")
		clicks = strings.TrimSpace(strings.Repeat(head+" ", times)) + " " + tail
	}
	g := reachCommandPhase(t)
	for _, c := range strings.Fields(clicks) {
		if strings.HasPrefix(c, "@") {
			// @x,y ＝ 只移動滑鼠，不按下去
			x, y, _ := strings.Cut(c[1:], ",")
			mx, _ := strconv.Atoi(x)
			my, _ := strconv.Atoi(y)
			g.MoveMouse(mx, my)
			continue
		}
		x, y, ok := strings.Cut(c, ",")
		if !ok {
			t.Fatalf("點擊要寫成 x,y：%q", c)
		}
		cx, _ := strconv.Atoi(x)
		cy, _ := strconv.Atoi(y)
		g.Click(cx, cy)
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "screen.bin"))
	if err != nil {
		t.Skipf("沒有收據 %s，跳過", batch)
	}
	got := g.Frame().Pix

	if r := os.Getenv("LOGH3_REGION"); r != "" {
		var x0, y0, x1, y1 int
		fmt.Sscanf(r, "%d,%d,%d,%d", &x0, &y0, &x1, &y1)
		for y := y0; y <= y1; y++ {
			var a, b strings.Builder
			for x := x0; x <= x1; x++ {
				fmt.Fprintf(&a, "%X", want[y*640+x])
				fmt.Fprintf(&b, "%X", got[y*640+x])
			}
			t.Logf("y=%3d 原版 %s  remake %s", y, a.String(), b.String())
		}
		return
	}

	bx0, by0, bx1, by1 := 0, 0, 639, 479
	if bb := os.Getenv("LOGH3_BOX"); bb != "" {
		fmt.Sscanf(bb, "%d,%d,%d,%d", &bx0, &by0, &bx1, &by1)
	}
	rows := map[int][]string{}
	n, minY, maxY := 0, 1<<30, -1
	for i := range want {
		if want[i] == got[i] {
			continue
		}
		x, y := i%640, i/640
		if x < bx0 || x > bx1 || y < by0 || y > by1 {
			continue
		}
		rows[y] = append(rows[y], fmt.Sprintf("x%d(原版 %d／remake %d)", x, want[i], got[i]))
		n++
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	if n == 0 {
		t.Logf("%s：逐位元組相同", batch)
		return
	}
	t.Logf("%s：差 %d 個像素，y%d–%d", batch, n, minY, maxY)
	for y := minY; y <= maxY; y++ {
		if len(rows[y]) > 0 {
			t.Logf("y=%d 共 %d：%s", y, len(rows[y]), strings.Join(rows[y], " "))
		}
	}
}
