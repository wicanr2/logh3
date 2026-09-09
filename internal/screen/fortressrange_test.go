package screen

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFortressRangeMatchesOracle 用**兩張只差瞄準方位**的收據驗射程橢圓。
//
// 橢圓是 XOR 上去的，所以同一個底圖畫或不畫，差別剛好是那條曲線：
// 兩張收據逐像素比出來的差異集合，必須等於兩條曲線的對稱差。這樣不必
// 知道底圖長什麼樣，也不必把整張戰場畫出來，就能逐位元組驗幾何。
//
//	s3fort   游標在滑鼠 (320,240) → 傾角 18
//	s3fort2  游標在滑鼠 (150,180) → 傾角 140
//
// 兩張的行星都是同一顆（poke 成要塞、座標世界 (1280,780)、防衛力上限
// 3300 → 長半軸 96、短半軸 24），焦點 ＝ 佔位框中心 (1287,787)
// → 畫面 (247,247)。唯一要排除的是**原版自己畫的滑鼠游標**——兩張的
// 游標位置本來就不同（`docs/re/145`）。
func TestFortressRangeMatchesOracle(t *testing.T) {
	root := "../.."
	load := func(name string) []byte {
		b, err := os.ReadFile(
			filepath.Join(root, "workplace/oracle", name, "screen.bin"))
		if err != nil {
			t.Skipf("缺收據 %s，跳過：%v", name, err)
		}
		return b
	}
	a := load("s3fort")
	b := load("s3fort2")
	if len(a) != 640*480 || len(b) != len(a) {
		t.Fatalf("收據大小 %d／%d，該是 %d", len(a), len(b), 640*480)
	}

	const (
		focusX, focusY = 247, 247 // 焦點的畫面座標
		semiMajor      = 96
		semiMinor      = 24
		fieldX, fieldY = 16, 60
		fieldW, fieldH = 28 * 16, 18 * 20
	)
	// 世界座標與畫面座標在這兩張裡差一個固定平移，所以直接把焦點的
	// 畫面座標餵進去，裁切也照畫面座標做。
	odd := func(tilt int) map[[2]int]bool {
		count := map[[2]int]int{}
		for _, p := range FortressRangePoints(focusX, focusY, semiMajor, semiMinor, tilt) {
			if p[0] < fieldX || p[1] < fieldY ||
				p[0] >= fieldX+fieldW || p[1] >= fieldY+fieldH {
				continue
			}
			count[p]++
		}
		out := map[[2]int]bool{}
		for p, n := range count {
			if n%2 == 1 {
				out[p] = true
			}
		}
		return out
	}
	want := map[[2]int]bool{}
	for p := range odd(18) {
		want[p] = !want[p]
	}
	for p := range odd(140) {
		want[p] = !want[p]
	}

	// 兩張收據的游標位置不同，那兩塊要排除。游標吃的是滑鼠座標，
	// 畫面 y ＝ 滑鼠 y ＋ 40。
	skip := func(x, y int) bool {
		for _, m := range [][2]int{{320, 240}, {150, 180}} {
			if x >= m[0] && x < m[0]+24 && y >= m[1]+40 && y < m[1]+40+28 {
				return true
			}
		}
		return false
	}

	got := map[[2]int]bool{}
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			if a[y*640+x] != b[y*640+x] && !skip(x, y) {
				got[[2]int{x, y}] = true
			}
		}
	}
	for p, on := range want {
		if !on || skip(p[0], p[1]) {
			continue
		}
		if !got[p] {
			t.Errorf("預測 (%d,%d) 兩張會不同，實際相同", p[0], p[1])
		}
		delete(got, p)
	}
	for p := range got {
		t.Errorf("(%d,%d) 兩張不同，但不在預測的橢圓上", p[0], p[1])
	}
	n := 0
	for p, on := range want {
		if on && !skip(p[0], p[1]) {
			n++
		}
	}
	if n == 0 {
		t.Fatal("預測出來的差異是空的")
	}
	t.Logf("兩條橢圓的對稱差 %d 個像素，逐一相同", n)
}
