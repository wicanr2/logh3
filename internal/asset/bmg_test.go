package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBMGWholeFileParses：254 段、1,438 格，每一格都剛好接到下一格（`docs/re/328`）。
func TestBMGWholeFileParses(t *testing.T) {
	b := openBMG(t)
	if n := len(b.Names()); n != 254 {
		t.Fatalf("目錄 %d 筆，原版 254", n)
	}
	frames := 0
	for _, name := range b.Names() {
		fr, err := b.Frames(name)
		if err != nil {
			t.Fatal(err)
		}
		frames += len(fr)
		// 每一格畫在左視窗都要剛好吃完自己的資料、不越界。
		p := PlanesFrom(make([]byte, 640*480))
		for i, f := range fr {
			used, err := f.draw(p, BMGWindowBase(1, 4))
			if err != nil {
				t.Fatalf("%x 第 %d 格：%v", name, i, err)
			}
			if used != len(f.data) {
				t.Fatalf("%x 第 %d 格吃了 %d／%d bytes", name, i, used, len(f.data))
			}
		}
	}
	if frames != 1438 {
		t.Fatalf("共 %d 格，原版 1438", frames)
	}
}

// TestBMGEngagementMatchesOracle 用交戰演出的收據逐格重畫（`tools/engagement_receipts.sh`，
// S3 第五回合打行星那一場）：`eng-text` 已經畫好 PPD0（左）與 FÀD0（右）的第 0 格；
// PPAB 第 1–4 格疊在左邊 ＝ `eng-mid`、第 5–8 格 ＝ `eng-mid2`、第 9–10 格之後
// FÀD0 整段（第 1–6 格）畫在右邊 ＝ `eng-end`。只比兩個 256×102 視窗。
func TestBMGEngagementMatchesOracle(t *testing.T) {
	b := openBMG(t)
	start := receipt(t, "eng-text")
	p := PlanesFrom(start)
	left, right := BMGWindowBase(1, 4), BMGWindowBase(47, 4)

	// 第 0 格（覆寫）畫回同一個位置不會改變 eng-text。
	for _, c := range []struct {
		name string
		base int
	}{{"PPD0", left}, {"F\xc0D0", right}} {
		fr, err := b.Frames(c.name)
		if err != nil {
			t.Fatal(err)
		}
		if err := fr[0].Draw(p, c.base); err != nil {
			t.Fatal(err)
		}
	}
	compareWindows(t, p.Pixels(), start, "eng-text（第 0 格）")

	ppab, err := b.Frames("PPAB")
	if err != nil {
		t.Fatal(err)
	}
	if len(ppab) != 10 {
		t.Fatalf("PPAB %d 格，原版 10", len(ppab))
	}
	drawN := func(fr []BMGFrame, base int) {
		t.Helper()
		for _, f := range fr {
			if err := f.Draw(p, base); err != nil {
				t.Fatal(err)
			}
		}
	}
	drawN(ppab[0:4], left)
	compareWindows(t, p.Pixels(), receipt(t, "eng-mid"), "eng-mid")
	drawN(ppab[4:8], left)
	compareWindows(t, p.Pixels(), receipt(t, "eng-mid2"), "eng-mid2")
	drawN(ppab[8:], left)
	fad0, err := b.Frames("F\xc0D0")
	if err != nil {
		t.Fatal(err)
	}
	drawN(fad0[1:], right)
	compareWindows(t, p.Pixels(), receipt(t, "eng-end"), "eng-end")
}

func openBMG(t *testing.T) *BMG {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(origDir(t), "COMANI.BMG"))
	if err != nil {
		t.Skipf("讀不到 COMANI.BMG：%v", err)
	}
	b, err := OpenBMG(data)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func receipt(t *testing.T, batch string) []byte {
	t.Helper()
	scr, err := os.ReadFile(filepath.Join("../..", "workplace/oracle", batch, "screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skipf("沒有收據 %s，跳過", batch)
	}
	return scr
}

// compareWindows 比交戰演出左右兩個視窗（(8,56)／(376,56)，各 256×102）。
func compareWindows(t *testing.T, got, want []byte, what string) {
	t.Helper()
	bad := 0
	for _, x0 := range []int{8, 376} {
		for y := 56; y < 56+102; y++ {
			for x := x0; x < x0+256; x++ {
				if got[y*640+x] != want[y*640+x] {
					bad++
				}
			}
		}
	}
	if bad != 0 {
		t.Fatalf("%s：兩個視窗裡不同的像素 %d / %d", what, bad, 2*256*102)
	}
}
