package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 收據 `s3msg`：分隊選單選「移動」之後點一個過不去的格子，原版跳
// 「這是無法行動的地方。」（`ds:4A24`）。這裡把訊息框畫出來，
// 與收據逐位元組比框那一塊（`docs/re/137`）。
func TestMessageBoxGeometryMatchesOracle(t *testing.T) {
	root := "../.."
	l, err := LoadCommandMenuSpec(
		filepath.Join(root, "docs/internals/data/battle-command-menus.json"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadMessageBoxSpec(
		filepath.Join(root, "docs/internals/data/message-box.json"))
	if err != nil {
		t.Fatal(err)
	}
	x, y, w, h, ok := s.Box(2, l)
	if !ok {
		t.Fatal("讀不到種類 2 的樣板")
	}
	// 收據 `s3msg`：分隊選單選「移動」之後點一個過不去的格子，原版跳
	// 「這是無法行動的地方。」。框量到 (65,181) 350×118（`docs/re/137`）。
	if x != 65 || y != 181 || w != 350 || h != 118 {
		t.Fatalf("框是 (%d,%d) %d×%d，收據量到 (65,181) 350×118", x, y, w, h)
	}
	// 位置 ＝ (格x×16+1, (格y+2)×20+1)、大小 ＝ (寬×16+30)×(高×20+38)
	// ——大小的公式與命令選單完全相同，只有 y 多兩格。
	g := &l.Layout
	if w != 20*g.CellW+g.FramePadW || h != 4*g.CellH+g.FramePadH {
		t.Error("大小的公式與命令選單不一致")
	}

}

// 同一份收據，這次連框線的像素一起比。框是 `FRAME.BLK` 的圖塊
// （`docs/re/138`），所以比的是「整個 22 × 6 格」——差異只該落在
// 中間那幾行字上。
func TestMessageBoxFrameMatchesOracle(t *testing.T) {
	root := "../.."
	s, err := LoadMessageBoxSpec(
		filepath.Join(root, "docs/internals/data/message-box.json"))
	if err != nil {
		t.Fatal(err)
	}
	blk, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skipf("缺原版 FRAME.BLK：%v", err)
	}
	tiles, err := ui.LoadFrameTiles(blk)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3msg/screen.bin"))
	if err != nil {
		t.Skipf("缺收據 s3msg：%v", err)
	}

	fb := gfx.NewFramebuffer(0)
	if !DrawMessageBox(fb, s, 2, tiles) {
		t.Fatal("畫不出種類 2 的訊息框")
	}
	var tmpl struct{ CellX, CellY, Cols, Rows int }
	for _, tp := range s.Templates {
		if tp.Kind == 2 {
			tmpl = struct{ CellX, CellY, Cols, Rows int }{tp.CellX, tp.CellY, tp.Cols, tp.Rows}
		}
	}
	cols, rows := tmpl.Cols+2, tmpl.Rows+2
	x0 := tmpl.CellX * ui.FrameTileW
	y0 := tmpl.CellY*ui.FrameTileH + ui.FrameCellOriginY

	// 字在中間那幾格，不比；框那一圈與四周的底色要逐位元組相同。
	var diff int
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if r > 0 && r < rows-1 && c > 0 && c < cols-1 {
				continue
			}
			for y := 0; y < ui.FrameTileH; y++ {
				for x := 0; x < ui.FrameTileW; x++ {
					sx, sy := x0+c*ui.FrameTileW+x, y0+r*ui.FrameTileH+y
					if fb.At(sx, sy) != want[sy*640+sx] {
						diff++
					}
				}
			}
		}
	}
	if diff != 0 {
		t.Errorf("框那一圈有 %d 個像素與收據不同", diff)
	}
}

// 框 ＋ 那行字 ＋ 按鈕一起畫，與收據 `s3msg` 比整個 22 × 6 格
// （42,240 個像素）逐位元組。
func TestMessageBoxMatchesOracle(t *testing.T) {
	root := "../.."
	s, err := LoadMessageBoxSpec(
		filepath.Join(root, "docs/internals/data/message-box.json"))
	if err != nil {
		t.Fatal(err)
	}
	blk, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skipf("缺原版 FRAME.BLK：%v", err)
	}
	tiles, err := ui.LoadFrameTiles(blk)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺原版字型：%v", err)
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3msg/screen.bin"))
	if err != nil {
		t.Skipf("缺收據 s3msg：%v", err)
	}

	fb := gfx.NewFramebuffer(0)
	if !DrawMessageBox(fb, s, 2, tiles) {
		t.Fatal("畫不出種類 2 的訊息框")
	}
	// `ds:4A24`「這是無法行動的地方。」（`docs/re/121`）。
	raw := []byte{
		0xb3, 0x6f, 0xac, 0x4f, 0xb5, 0x4c, 0xaa, 0x6b, 0xa6, 0xe6,
		0xb0, 0xca, 0xaa, 0xba, 0xa6, 0x61, 0xa4, 0xe8, 0xa1, 0x43,
	}
	codes := make([]uint16, 0, len(raw)/2)
	for i := 0; i+1 < len(raw); i += 2 {
		codes = append(codes, uint16(raw[i])<<8|uint16(raw[i+1]))
	}
	if miss := DrawMessageBoxText(fb, s, 2, fnt, 0, codes); len(miss) != 0 {
		t.Fatalf("有缺字：%04X", miss)
	}

	tp, _ := s.template(2)
	x0 := tp.cellX * ui.FrameTileW
	y0 := tp.cellY*ui.FrameTileH + ui.FrameCellOriginY
	w, h := (tp.cols+2)*ui.FrameTileW, (tp.rows+2)*ui.FrameTileH

	if miss := DrawMessageBoxButtons(fb, s, 2, fnt); len(miss) != 0 {
		t.Fatalf("按鈕有缺字：%04X", miss)
	}

	var diff int
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			if fb.At(x, y) != want[y*640+x] {
				diff++
			}
		}
	}
	if diff != 0 {
		t.Errorf("整個訊息框有 %d 個像素與收據不同", diff)
	}
}