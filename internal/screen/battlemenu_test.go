package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// TestBattleMenuMatchesOracle 把戰術選單逐位元組對收據。
//
// 收據 `s3menu` 的 `-dump-at` 在第一下點擊之後、第二下之前取了一張畫面
// （`docs/re/98`、`docs/re/101`），選單開在游標 (240,240) 處。選單那個
// 矩形整塊是選單自己的（框不透明），所以畫在空白畫布上就能直接比。
func TestBattleMenuMatchesOracle(t *testing.T) {
	root := "../.."
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3menu/menu-open.bin"))
	if err != nil || len(want) < gfx.Width*gfx.Height {
		t.Skipf("缺收據 s3menu/menu-open.bin，跳過（不用代用品）")
	}
	s, err := LoadBattleMenuSpec(filepath.Join(root, "docs/internals/data/battle-menu.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺 GRAPH.IMG，跳過：%v", err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	blk, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skipf("缺 FRAME.BLK，跳過：%v", err)
	}
	tiles, err := ui.LoadFrameTiles(blk)
	if err != nil {
		t.Fatal(err)
	}
	x, y := s.BattleMenuAt(240, 240)
	fb := gfx.NewFramebuffer(0)
	if m := DrawBattleMenu(s, fb, fnt, tiles, x, y, 0); len(m) > 0 {
		t.Fatalf("缺字 %04X", m)
	}
	// 收據上有游標——原版自己畫它，壓在第一項上。游標吃的是**滑鼠**座標
	// （`offset_from_mouse` 的 40 就是滑鼠與畫面的差，`docs/internals/31`），
	// 所以這裡傳 200 不是 240。
	cur.Draw(fb, 240, 200)
	w, h := s.Layout.Window.W, s.Layout.Window.H
	diff, first := 0, [2]int{-1, -1}
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			if fb.Pix[yy*gfx.Width+xx] != want[yy*gfx.Width+xx] {
				if diff == 0 {
					first = [2]int{xx, yy}
				}
				diff++
			}
		}
	}
	t.Logf("選單 (%d,%d) %d×%d：%d／%d 個像素不同", x, y, w, h, diff, w*h)
	if diff != 0 {
		t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d",
			first[0], first[1],
			fb.Pix[first[1]*gfx.Width+first[0]], want[first[1]*gfx.Width+first[0]])
	}
}
