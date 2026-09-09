package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func loadTiles(t *testing.T) *FrameTiles {
	t.Helper()
	blk, err := os.ReadFile(filepath.Join("../..", "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skipf("缺原版 FRAME.BLK：%v", err)
	}
	tiles, err := LoadFrameTiles(blk)
	if err != nil {
		t.Fatal(err)
	}
	return tiles
}

// FRAME.BLK 是 33 張 16×20 的四平面圖塊，外框用的是前八張。
func TestFrameTilesShape(t *testing.T) {
	tiles := loadTiles(t)
	if len(tiles.Tiles) != 33 {
		t.Errorf("圖塊數 %d，預期 33", len(tiles.Tiles))
	}
	for i, img := range tiles.Tiles[:FrameRingLen] {
		if img.W != FrameTileW || img.H != FrameTileH {
			t.Errorf("圖塊 %d 是 %d×%d", i, img.W, img.H)
		}
	}
}

// `ui.MenuFrame` 是從主選單量出來的線條版，主選單用的是第二種樣式
// （圖塊 8–15，起始碼 9）。兩者的關係應該是：**線條版畫的每一個像素都
// 與圖塊版相同，差別只有圖塊版多畫最外圈那一像素**——線條版的矩形從
// `格x×16＋1` 起算，正好把外圈讓出去。原版主選單看不出差別，是因為
// 那一圈的圖塊色號就是底色 3（`docs/re/138`）。
func TestMenuFrameMatchesTiles(t *testing.T) {
	tiles := loadTiles(t)
	// 主選單 5×6 格擺第一個；其餘尺寸用來檢驗線條版的角落修正
	// 換尺寸還成不成立。
	sizes := []struct{ cols, rows int }{
		{5, 6}, {4, 4}, {4, 5}, {10, 10}, {22, 6},
	}
	for _, s := range sizes {
		cellX, cellY := 4, 7
		x, y := cellX*FrameTileW+1, (cellY+FrameCellOriginY/FrameTileH)*FrameTileH+1
		w, h := s.cols*FrameTileW-2, s.rows*FrameTileH-2

		a := gfx.NewFramebuffer(0)
		f := MenuFrame()
		f.Fill = 9
		f.Draw(a, x, y, w, h)

		b := gfx.NewFramebuffer(0)
		if !(TileFrame{Tiles: tiles, Style: MainMenuFrameStyle, Fill: 9}).DrawRect(b, x, y, w, h) {
			t.Fatalf("%d×%d 格對不上格線", s.cols, s.rows)
		}

		// 線條版畫得到的範圍內，兩者必須逐位元組相同。
		var inner int
		for py := y; py < y+h; py++ {
			for px := x; px < x+w; px++ {
				if a.At(px, py) != b.At(px, py) {
					inner++
				}
			}
		}
		if inner != 0 {
			t.Errorf("%d×%d 格：矩形內差 %d 個像素", s.cols, s.rows, inner)
		}

		// 差的只該是最外圈那一像素，數量等於周長。
		var total int
		for i := range a.Pix {
			if a.Pix[i] != b.Pix[i] {
				total++
			}
		}
		perim := 2*s.cols*FrameTileW + 2*s.rows*FrameTileH - 4
		if total != perim {
			t.Errorf("%d×%d 格：整體差 %d 個像素，周長是 %d", s.cols, s.rows, total, perim)
		}
	}
}
