package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
)

// TestCopyProtectionMatchesOracle 是防拷問答畫面的同狀態對拍：
// remake 畫出來的色號陣列要與原版收據**逐位元組相同**。
func TestCopyProtectionMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-copy-protection.json")
	oracle := filepath.Join(root, "workplace/oracle/cb1/screen.bin")
	fontPath := filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	facePath := filepath.Join(root, "org_game/extracted/FACE.AMG")
	for _, p := range []string{specPath, oracle, fontPath, facePath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadCopyProtSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	faceData, err := os.ReadFile(facePath)
	if err != nil {
		t.Fatal(err)
	}
	face, err := FaceFrom(faceData, s)
	if err != nil {
		t.Fatal(err)
	}
	fb, missing := CopyProtection(s, fnt, face, "")
	if len(missing) > 0 {
		t.Fatalf("有 %d 個字碼畫不出來：%04X", len(missing), missing)
	}
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != len(fb.Pix) {
		t.Fatalf("收據 %d bytes，畫布 %d bytes", len(want), len(fb.Pix))
	}
	bad := 0
	firstX, firstY := -1, -1
	for i := range want {
		if want[i] != fb.Pix[i] {
			if bad == 0 {
				firstX, firstY = i%640, i/640
			}
			bad++
		}
	}
	if bad != 0 {
		// 差異要看得到才好修：把 remake 這一側倒出來，
		// 用 cmd/oradiff 跟收據疊差分圖。
		_ = os.WriteFile(filepath.Join(root, "workplace/remake-copyprot.bin"), fb.Pix, 0o644)
		t.Fatalf("與原版不同的像素 %d / %d，第一個在 (%d,%d)；bg=%d 視窗=%+v 填色=%d 頭像=(%d,%d)",
			bad, len(want), firstX, firstY, s.Background, s.Window, s.Bevel.Fill, s.Face.X, s.Face.Y)
	}
}

// TestCopyProtectionTypedMatchesOracle 打進去的數字要與原版逐位元組相同。
//
// 那套字型是 GRAPH.IMG 的**全形**數字（`docs/re/56`）。
func TestCopyProtectionTypedMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-copy-protection.json")
	fontPath := filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	facePath := filepath.Join(root, "org_game/extracted/FACE.AMG")
	oracle := filepath.Join(root, "workplace/oracle/typed138/screen.bin")
	for _, p := range []string{specPath, fontPath, facePath, oracle} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadCopyProtSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	faceData, err := os.ReadFile(facePath)
	if err != nil {
		t.Fatal(err)
	}
	face, err := FaceFrom(faceData, s)
	if err != nil {
		t.Fatal(err)
	}
	fb, missing := CopyProtection(s, fnt, face, "138")
	if len(missing) > 0 {
		t.Fatalf("有 %d 個字碼畫不出來：%04X", len(missing), missing)
	}
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	bad, firstX, firstY := 0, -1, -1
	for i := range want {
		if want[i] != fb.Pix[i] {
			if bad == 0 {
				firstX, firstY = i%640, i/640
			}
			bad++
		}
	}
	if bad != 0 {
		t.Fatalf("與原版不同的像素 %d / %d，第一個在 (%d,%d)", bad, len(want), firstX, firstY)
	}
}

// TestCopyProtectionDrawsTyped 釘「打進去的數字與輸入游標落在量到的位置上」。
//
// 打了三位數之後畫面上會變的有兩樣：數字本身，以及**跟著移動的輸入游標**
// （`docs/re/281`）。所以這裡不是只看數字——兩份收據 `cb1`（沒打）與
// `typed138`（打了 138）逐位元組相減，差異的外框就是下面這兩組數字。
func TestCopyProtectionDrawsTyped(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-copy-protection.json")
	fontPath := filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	for _, p := range []string{specPath, fontPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過", p)
		}
	}
	s, err := LoadCopyProtSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	blank, _ := CopyProtection(s, fnt, nil, "")
	typed, _ := CopyProtection(s, fnt, nil, "138")
	bbox := func(keep func(y int) bool) (int, int, int, int) {
		minX, maxX, minY, maxY := 1<<30, -1, 1<<30, -1
		for i := range blank.Pix {
			if blank.Pix[i] == typed.Pix[i] {
				continue
			}
			x, y := i%640, i/640
			if !keep(y) {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
		return minX, minY, maxX, maxY
	}
	// 數字 ＋ 游標：`cb1` 與 `typed138` 相減量到 x[240,303] y[270,284]。
	if x0, y0, x1, y1 := bbox(func(int) bool { return true }); x1 < 0 {
		t.Fatal("打了 138 但畫面沒有任何變化")
	} else if x0 != 240 || y0 != 270 || x1 != 303 || y1 != 284 {
		t.Fatalf("數字＋游標畫在 (%d,%d)-(%d,%d)，原版量到的是 (240,270)-(303,284)", x0, y0, x1, y1)
	}
	// 只看游標那兩列以上的部分 ＝ 數字本身：x[245,285] y[270,282]。
	if x0, y0, x1, y1 := bbox(func(y int) bool { return y < 283 }); x0 != 245 || y0 != 270 || x1 != 285 || y1 != 282 {
		t.Fatalf("數字畫在 (%d,%d)-(%d,%d)，原版量到的是 (245,270)-(285,282)", x0, y0, x1, y1)
	}
}
