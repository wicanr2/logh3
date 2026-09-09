package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
)

// introDump 讀一份 oracle 倒出來的動畫格。`-dump-at` 除了 PNG 也會順手寫
// 一份同名的 `.bin`——那是 640×350 的色號陣列，**與調色盤無關**，所以
// 逐位元組對拍用它，不看 PNG（動畫那一段的 PNG 因為 DAC 還沒換成
// `BU16IBM.RGB` 看起來像雜訊，`docs/re/261`）。
func introDump(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "workplace", "oracle", "introsched", "frames", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("沒有收據 %s，跳過（tools/oracle.sh introsched 可重建）", name)
	}
	if len(b) != IntroPageCols*IntroPageRows {
		t.Fatalf("%s 有 %d bytes，要 %d", name, len(b), IntroPageCols*IntroPageRows)
	}
	return b
}

// TestIntroPageGeometryMatchesOracle 用原版倒出來的畫面釘住頁面版面
// （`docs/re/261`）：227 列的動畫畫在 640×350 頁面的第 **61** 列，上下
// 留白全 0，而且畫出來的是真正的 16 色（不是只有 0 與 15）。
//
// `IntroTopRow` 本來是從「227 列在 480 列裡置中」推出來的
// （`docs/spec/020`）；這一支把它換成量到的。
func TestIntroPageGeometryMatchesOracle(t *testing.T) {
	for _, name := range []string{"if000.bin", "if001.bin", "if003.bin", "if004.bin"} {
		want := introDump(t, name)
		first, last, inks := -1, -1, map[uint8]int{}
		for y := 0; y < IntroPageRows; y++ {
			row := want[y*IntroPageCols : (y+1)*IntroPageCols]
			any := false
			for _, v := range row {
				if v != 0 {
					any = true
					inks[v]++
				}
			}
			if any {
				if first < 0 {
					first = y
				}
				last = y
			}
		}
		if first != IntroTopRow {
			t.Errorf("%s：第一列有內容的是 %d，IntroTopRow 是 %d", name, first, IntroTopRow)
		}
		if got := last - first + 1; got != asset.IntroPlaneRows {
			t.Errorf("%s：有內容的列數 %d，一格是 %d 列", name, got, asset.IntroPlaneRows)
		}
		if len(inks) < 4 {
			t.Errorf("%s：只用到 %d 種色號，不像 16 色畫面", name, len(inks))
		}
	}
}
