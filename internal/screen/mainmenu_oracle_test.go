package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// TestMainMenuMatchesOracle 是主選單的同狀態對拍：remake 畫出來的色號陣列
// 必須與 dosgolem 跑原版產生的收據**逐位元組相同**。
//
// 這是唯一算數的驗收。其他測試全綠只證明 remake 內部自洽；目視相似更不算。
//
// **缺檔就 skip**：字型是原版素材，repo 裡沒有；oracle 收據在 workplace/
// 也不進版控。用代用品頂替會讓「還沒做完」看起來像做完了。
func TestMainMenuMatchesOracle(t *testing.T) {
	root := "../.."
	fontPath := os.Getenv("LOGH3_FONT")
	if fontPath == "" {
		fontPath = filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	}
	oraclePath := filepath.Join(root, "workplace/oracle/hilite/screen.bin")
	palPath := filepath.Join(root, "workplace/oracle/hilite/pal.bin")
	for _, p := range []string{fontPath, oraclePath, palPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}

	fnt, err := font.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := LoadMenuSpec(filepath.Join(root, "docs/internals/data/screen-main-menu.json"))
	if err != nil {
		t.Fatal(err)
	}
	fb, missing := MainMenu(spec, fnt)
	if len(missing) != 0 {
		t.Fatalf("有缺字：%04X", missing)
	}

	want, err := os.ReadFile(oraclePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != gfx.Width*gfx.Height {
		t.Fatalf("oracle 收據長度 %d，預期 %d", len(want), gfx.Width*gfx.Height)
	}
	var diff, firstX, firstY int
	firstX, firstY = -1, -1
	for i := range want {
		if want[i] != fb.Pix[i] {
			if firstX < 0 {
				firstX, firstY = i%gfx.Width, i/gfx.Width
			}
			diff++
		}
	}
	if diff != 0 {
		t.Errorf("與 oracle 差 %d 個像素（%.4f%%），第一個在 (%d,%d)：oracle=%d remake=%d",
			diff, float64(diff)*100/float64(len(want)), firstX, firstY,
			want[firstY*gfx.Width+firstX], fb.Pix[firstY*gfx.Width+firstX])
	}

	// 調色盤也要對：圖形對了顏色錯是兩件事，混在一起看不出來。
	wantPal, err := os.ReadFile(palPath)
	if err != nil {
		t.Fatal(err)
	}
	pal := spec.Palette()
	for i := 0; i < gfx.Colors; i++ {
		r, g, b := pal.RGB8(i)
		if wantPal[i*3] != r || wantPal[i*3+1] != g || wantPal[i*3+2] != b {
			t.Errorf("色號 %d：oracle=(%d,%d,%d) remake=(%d,%d,%d)", i,
				wantPal[i*3], wantPal[i*3+1], wantPal[i*3+2], r, g, b)
		}
	}
}
