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

// TestCopyProtectionDrawsTyped 只釘「打進去的數字有畫出來、而且落在量到的
// 位置上」。**這一格還沒對拍通過**：畫面上那套數字字型不是 GRAPH.IMG 的
// 半形字型（比它高也寬），來源還沒找到，所以不能宣稱逐位元組一致。
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
	var minX, maxX, minY, maxY = 1 << 30, -1, 1 << 30, -1
	for i := range blank.Pix {
		if blank.Pix[i] != typed.Pix[i] {
			x, y := i%640, i/640
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
	}
	if maxX < 0 {
		t.Fatal("打了 138 但畫面沒有任何變化")
	}
	// 原版量到的範圍是 x 244–285、y 270–283（收據 typed138）。
	if minX < 243 || maxX > 286 || minY < 266 || maxY > 284 {
		t.Fatalf("數字畫在 (%d,%d)-(%d,%d)，超出原版量到的範圍", minX, minY, maxX, maxY)
	}
}
