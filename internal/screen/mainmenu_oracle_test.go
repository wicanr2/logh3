package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// TestMainMenuMatchesOracle 是主選單的同狀態對拍：remake 畫出來的色號陣列
// 必須與 dosgolem 跑原版產生的收據**逐位元組相同**。
//
// 這是唯一算數的驗收。其他測試全綠只證明 remake 內部自洽；目視相似更不算。
//
// **缺檔就 skip**：字型是原版素材，repo 裡沒有；oracle 收據在 workplace/
// 也不進版控。用代用品頂替會讓「還沒做完」看起來像做完了。
func TestMainMenuMatchesOracle(t *testing.T) {
	// 四個反白狀態各一份收據：滑鼠停在第 r 列上，遊戲反白那一列。
	// **一個狀態對得上不算數**——反白的模型（底色／字色／立體邊都換）
	// 只有在多個狀態同時成立時才算被驗證過。
	//
	// ⚠ 收據要在**選單畫出來之前**就把滑鼠定位（`-mouse-at 30000000`）。
	// 畫完之後才移動的話拿到的是「反白從別列搬過來」的轉換後狀態，
	// 被還原的那一列會留下痕跡——那是轉換效果，不是這個靜態畫面該有的東西。
	cases := []struct {
		batch  string
		mx, my int
	}{
		{"r0", 300, 170}, {"r1", 300, 190}, {"r2", 300, 210}, {"r3", 300, 230},
		{"cur300", 450, 300},  // 滑鼠在選單外：不反白，只有游標
		{"curfill", 310, 241}, // 游標壓在視窗下緣的外框上
	}
	for _, c := range cases {
		t.Run(c.batch, func(t *testing.T) { checkOracle(t, c.batch, c.mx, c.my) })
	}
}

func checkOracle(t *testing.T, batch string, mx, my int) {
	t.Helper()
	root := "../.."
	fontPath := os.Getenv("LOGH3_FONT")
	if fontPath == "" {
		fontPath = filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	}
	oraclePath := filepath.Join(root, "workplace/oracle", batch, "screen.bin")
	palPath := filepath.Join(root, "workplace/oracle", batch, "pal.bin")
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
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	// 表裡是收據的滑鼠座標；RowAt 吃畫面座標，兩者差 ui.MouseDY。
	fb, missing := MainMenuAt(spec, fnt, RowAt(spec, my+ui.MouseDY), cur, mx, my)
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
