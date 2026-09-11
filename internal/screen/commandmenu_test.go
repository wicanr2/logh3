package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// TestCommandMenuMatchesOracle 把三張命令選單逐位元組對收據
// （`docs/spec/023`）。
//
// 三份收據走同一條前置序列，只差最後一下點在哪個分隊上：
//
//	s3cmdsq3  分隊選單（移動／編隊），游標在第一項上
//	s3cmdfl3  艦隊選單（移動／戰鬥／索敵），被夾到上面、游標在框外
//	s3cmdoff  分隊行動改成 0 讓「移動」不可選
//
// 框不透明，所以畫在空白畫布上就能直接比那個矩形。收據上有原版自己畫的
// 游標，壓在框裡面，所以這裡也要畫——它吃的是**滑鼠**座標
// （畫面 y ＝ 滑鼠 y ＋ 40，`docs/re/52`）。
func TestCommandMenuMatchesOracle(t *testing.T) {
	root := "../.."
	s, err := LoadCommandMenuSpec(
		filepath.Join(root, "docs/internals/data/battle-command-menus.json"))
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
	cases := []struct {
		batch   string
		kind    CommandMenuKind
		mx, my  int // 滑鼠座標
		enabled []bool
	}{
		{"s3cmdsq3", CommandMenuSquadron, 261, 258, []bool{true, true}},
		{"s3cmdfl3", CommandMenuFleet, 298, 369, []bool{true, true, true}},
		{"s3cmdoff", CommandMenuSquadron, 261, 258, []bool{false, true}},
	}
	for _, c := range cases {
		t.Run(c.batch, func(t *testing.T) {
			want, err := os.ReadFile(
				filepath.Join(root, "workplace/oracle", c.batch, "screen.bin"))
			if err != nil || len(want) < gfx.Width*gfx.Height {
				t.Skipf("缺收據 %s，跳過（不用代用品）", c.batch)
			}
			n := s.ItemCount(c.kind)
			sx, sy := c.mx, c.my+ui.MouseDY
			x, y := s.CommandMenuAt(sx, sy, n)
			w, h := s.Size(c.kind, n)
			fb := gfx.NewFramebuffer(0)
			sel := s.CommandMenuRowAt(c.kind, x, y, sx, sy, n)
			if m := DrawCommandMenu(s, fb, fnt, tiles, c.kind, c.enabled, x, y, sel); len(m) > 0 {
				t.Fatalf("缺字 %04X", m)
			}
			cur.Draw(fb, c.mx, c.my)
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
			t.Logf("選單 (%d,%d) %d×%d 反白第 %d 列：%d／%d 個像素不同",
				x, y, w, h, sel, diff, w*h)
			if diff != 0 {
				t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d",
					first[0], first[1],
					fb.Pix[first[1]*gfx.Width+first[0]],
					want[first[1]*gfx.Width+first[0]])
			}
		})
	}
}
