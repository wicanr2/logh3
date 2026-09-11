package screen

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ui"
)

// 視窗外框是 `FRAME.BLK` 的圖塊，只貼得進 16 × 20 的格子裡，所以框的
// 矩形一定是 `(格x×16+1, (格y+2)×20+1)` 起、`(格數×16−2) × (列數×20−2)` 大
// （`docs/internals/47`）。對不上就整個框不畫——**而畫面上只是少一個框，
// 不會報錯**。這支測試把產品裡用得到的每一個矩形都跑一遍。
func cellAligned(x, y, w, h int) bool {
	return (x-1)%ui.FrameTileW == 0 && (y-1)%ui.FrameTileH == 0 &&
		(w+2)%ui.FrameTileW == 0 && (h+2)%ui.FrameTileH == 0
}

func TestMenuRectsAreCellAligned(t *testing.T) {
	root := "../.."
	data := filepath.Join(root, "docs/internals/data")

	t.Run("主選單", func(t *testing.T) {
		s, err := LoadMenuSpec(filepath.Join(data, "screen-main-menu.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !cellAligned(s.Window.X, s.Window.Y, s.Window.W, s.Window.H) {
			t.Errorf("主選單的框 (%d,%d) %d×%d 對不上格線",
				s.Window.X, s.Window.Y, s.Window.W, s.Window.H)
		}
	})

	// 戰術選單與命令選單開在游標處。游標可以停在任何一個像素上，
	// 所以要掃過一整格的每一個落點——**(240,240) 剛好是 16 與 20 的
	// 整數倍**，只拿那一點試的話，換成像素位移的錯模型也會過
	// （`docs/re/289`）。
	t.Run("戰術選單", func(t *testing.T) {
		s, err := LoadBattleMenuSpec(filepath.Join(data, "battle-menu.json"))
		if err != nil {
			t.Fatal(err)
		}
		for my := 40; my < 440; my++ {
			for dx := 0; dx < ui.FrameTileW; dx++ {
				mx := 224 + dx
				x, y := s.BattleMenuAt(mx, my)
				if !cellAligned(x, y, s.Layout.Window.W, s.Layout.Window.H) {
					t.Fatalf("游標 (%d,%d) → 框 (%d,%d) %d×%d 對不上格線",
						mx, my, x, y, s.Layout.Window.W, s.Layout.Window.H)
				}
			}
		}
		// 收據 s3menu 的落點：格 (14,11) → (225,221)（`docs/re/117`）。
		if x, y := s.BattleMenuAt(240, 240); x != 225 || y != 221 {
			t.Errorf("游標 (240,240) → (%d,%d)，收據上是 (225,221)", x, y)
		}
	})

	t.Run("命令選單", func(t *testing.T) {
		s, err := LoadCommandMenuSpec(filepath.Join(data, "battle-command-menus.json"))
		if err != nil {
			t.Fatal(err)
		}
		for kind := range s.Menus {
			n := s.ItemCount(kind)
			w, h := s.Size(kind, n)
			for my := 40; my < 440; my++ {
				for dx := 0; dx < ui.FrameTileW; dx++ {
					mx := 224 + dx
					x, y := s.CommandMenuAt(mx, my, n)
					if !cellAligned(x, y, w, h) {
						t.Fatalf("%s 選單：游標 (%d,%d) → 框 (%d,%d) %d×%d 對不上格線",
							kind, mx, my, x, y, w, h)
					}
				}
			}
		}
	})
}
