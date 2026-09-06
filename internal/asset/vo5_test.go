package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// TestVO5ConsumesWholeFile 釘住最便宜的自檢：一個正確的 `.VO5` 解完
// 四個 plane 之後，來源位元組要剛好用完。少一個或多一個都是解錯。
func TestVO5ConsumesWholeFile(t *testing.T) {
	dir := "../../org_game/extracted"
	names, err := filepath.Glob(filepath.Join(dir, "*.VO5"))
	if err != nil || len(names) == 0 {
		t.Skip("缺原版 .VO5，跳過")
	}
	for _, n := range names {
		d, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		used, err := VO5Consumed(d)
		if err != nil {
			t.Errorf("%s：%v", filepath.Base(n), err)
			continue
		}
		if used != len(d) {
			t.Errorf("%s：用掉 %d bytes，檔案有 %d", filepath.Base(n), used, len(d))
		}
	}
	t.Logf("檢查了 %d 個檔", len(names))
}

// TestVO5MatchesOriginalScreen 拿原版畫出來的畫面對拍。
//
// 最終畫面是好幾張圖疊出來的，所以只比 plane 2 與 plane 3——那兩個
// plane 與 `ENDT30.VO5` 單獨解出來的完全相同（`docs/internals/18`）。
func TestVO5MatchesOriginalScreen(t *testing.T) {
	root := "../.."
	src := filepath.Join(root, "org_game/extracted/ENDT30.VO5")
	shot := filepath.Join(root, "workplace/oracle/ending-empire/screen.bin")
	for _, p := range []string{src, shot} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過", p)
		}
	}
	d, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	img, err := DecodeVO5(d)
	if err != nil {
		t.Fatal(err)
	}
	screen, err := os.ReadFile(shot)
	if err != nil {
		t.Fatal(err)
	}
	const W, offY = 640, 80
	for _, plane := range []uint8{2, 3} {
		bad := 0
		for y := 0; y < img.H; y++ {
			for x := 0; x < img.W; x++ {
				a := (img.Pix[y*img.W+x] >> plane) & 1
				b := (screen[(offY+y)*W+x] >> plane) & 1
				if a != b {
					bad++
				}
			}
		}
		if bad != 0 {
			t.Errorf("plane %d：與原版畫面差 %d 個位元", plane, bad)
		}
	}
}
