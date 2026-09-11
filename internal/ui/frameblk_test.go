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

// 主選單那種框（樣式 9）的最外圈**整圈都是色號 3**。
//
// 這件事是「框可以直接照圖塊貼」的前提：圖塊版畫滿 cols × rows 格，
// 比先前那版線條重建多畫最外圈一像素。收據上那一圈量到的也是 3
// （主選單的背景色），所以多畫的那一圈一個像素都沒有改變
// （`docs/re/289`）。這一圈要是哪天不是 3 了，多畫出來就會被看見。
func TestMenuFrameOuterRingIsInk3(t *testing.T) {
	tiles := loadTiles(t)
	const fill = 9
	for _, s := range []struct{ cols, rows int }{{5, 6}, {4, 4}, {8, 9}, {22, 6}} {
		cellX, cellY := 4, 7
		x, y := cellX*FrameTileW+1, (cellY+FrameCellOriginY/FrameTileH)*FrameTileH+1
		w, h := s.cols*FrameTileW-2, s.rows*FrameTileH-2

		fb := gfx.NewFramebuffer(0)
		if !MenuFrame(tiles, fill).DrawRect(fb, x, y, w, h) {
			t.Fatalf("%d×%d 格對不上格線", s.cols, s.rows)
		}
		// 圖塊版佔的範圍比矩形各外擴一像素。
		rx, ry, rw, rh := x-1, y-1, w+2, h+2
		bad := 0
		for i := 0; i < rw; i++ {
			for _, py := range []int{ry, ry + rh - 1} {
				if fb.At(rx+i, py) != 3 {
					bad++
				}
			}
		}
		for i := 0; i < rh; i++ {
			for _, px := range []int{rx, rx + rw - 1} {
				if fb.At(px, ry+i) != 3 {
					bad++
				}
			}
		}
		if bad != 0 {
			t.Errorf("%d×%d 格：最外圈有 %d 個像素不是 3", s.cols, s.rows, bad)
		}
	}
}
