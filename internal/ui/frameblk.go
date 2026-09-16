package ui

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 視窗外框的圖塊（`FRAME.BLK`，`docs/re/138`）。
//
// 原版的框不是用線畫的，是**格**。`34CB:094A` 在畫面格表 `ds:D81E` 的
// 外圈填八個字碼 `0xFF00 + 旗標 + n`，`2376:0E0D` 看到高位元組 `0xFF`
// 就轉進 `2376:0E4A`，從 `ds:6BF4 + 0xA0 × (低位元組 − 1)` 取一張
// 16 × 20 的四平面圖塊貼上去。`ds:6BF4` 那塊就是 `FRAME.BLK` 的內容，
// 逐位元組相同，33 張。
//
// 圖塊裡的色號 0 不是黑，是**留給視窗底色**：原版每一列先算
// `~(平面0|平面1|平面2|平面3)`，再把視窗顏色的高四位（一平面一位元）
// 填進那些位元。所以同一組圖塊換個顏色就是另一種框。

// FrameTileW／FrameTileH 是一張框線圖塊的大小，也就是一格。
const (
	FrameTileW = 16
	FrameTileH = 20
)

// 樣式的起始碼就是視窗的 `[+3]`。`FRAME.BLK` 有 33 張圖塊：四組外框
// （每組八張）加一張落單的。第三、四組沒有色號 0，視窗底色透不出來。
const (
	PanelFrameStyle    = 1  // 圖塊 0–7：戰場視窗、訊息框、清單面板
	MainMenuFrameStyle = 9  // 圖塊 8–15：主選單與小選單
	FrameStyle3        = 17 // 圖塊 16–23
	FrameStyle4        = 25 // 圖塊 24–31
)

// FrameRing 是外圈八張圖塊相對於樣式起點的順序，取自 `34CB:094A` 的
// 寫入次序：左上、上、右上、左、右、左下、下、右下。
const (
	FrameTopLeft = iota
	FrameTop
	FrameTopRight
	FrameLeft
	FrameRight
	FrameBottomLeft
	FrameBottom
	FrameBottomRight
	FrameRingLen
)

// FrameTiles 是 `FRAME.BLK` 解出來的圖塊庫。
type FrameTiles struct{ Tiles []*asset.Image }

// LoadFrameTiles 解 `FRAME.BLK`。版面與 `SPRITE.SPR` 同一種（每列四個
// 平面各一個 word），所以直接走 asset.DecodeTiles。
func LoadFrameTiles(data []byte) (*FrameTiles, error) {
	imgs, err := asset.DecodeTiles(data, FrameTileW, FrameTileH)
	if err != nil {
		return nil, err
	}
	if len(imgs) < FrameRingLen {
		return nil, fmt.Errorf("FRAME.BLK 只有 %d 張圖塊，不足一組外框", len(imgs))
	}
	return &FrameTiles{Tiles: imgs}, nil
}

// TileFrame 是一個用圖塊畫的視窗外框。
//
// Style 是原版視窗的 `[+3]`（訊息框與命令選單都是 1）；外圈第 n 張用的
// 圖塊索引是 `Style − 1 + n`。Fill 是視窗顏色 `[+1]` 的**高四位**，
// 圖塊裡色號 0 的像素都填這個。
type TileFrame struct {
	Tiles *FrameTiles
	Style int
	Fill  uint8
}

// MenuFrame 是主選單、戰術選單與命令選單共用的那個框：`FRAME.BLK` 的
// 第二組圖塊（起始碼 9）。fill 是視窗顏色 `[+1]` 的高四位，圖塊裡色號 0
// 的像素都填它。
//
// **四個角不是算出來的，是圖塊 8／10／13／15 本身的像素**（`docs/re/289`）。
// 先前那版用七層線條重建整個框，四個角要另外列七個修正像素才對得上；
// 那七個像素就是線條模型與圖塊資料的差，不是遊戲裡的東西——原版從頭到尾
// 只是把八張圖塊貼進畫面格表。
func MenuFrame(tiles *FrameTiles, fill uint8) TileFrame {
	return TileFrame{Tiles: tiles, Style: MainMenuFrameStyle, Fill: fill}
}

// Draw 把 cols × rows 格的框畫在畫面格 (cellX, cellY)。
//
// 格座標的原點在畫面上方 40 像素處——`2376:040E` 的畫面位移從 `0x0C80`
// （＝ 2 × 1600 bytes ＝ 兩格列）起算，滑鼠的 `screen_y = mouse_y + 40`
// 是同一個偏移（`docs/re/52`、`docs/re/137`）。
func (f TileFrame) Draw(fb *gfx.Framebuffer, cellX, cellY, cols, rows int) {
	if f.Tiles == nil || cols < 2 || rows < 2 {
		return
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x := (cellX + c) * FrameTileW
			y := (cellY+r)*FrameTileH + FrameCellOriginY
			n, edge := ringTile(c, r, cols, rows)
			if !edge {
				fb.FillRect(x, y, FrameTileW, FrameTileH, f.Fill)
				continue
			}
			f.drawTile(fb, x, y, f.Style-1+n)
		}
	}
}

// FrameCellOriginY 是格座標系原點與畫面頂端的距離。
const FrameCellOriginY = 40

func ringTile(c, r, cols, rows int) (int, bool) {
	switch {
	case r == 0 && c == 0:
		return FrameTopLeft, true
	case r == 0 && c == cols-1:
		return FrameTopRight, true
	case r == 0:
		return FrameTop, true
	case r == rows-1 && c == 0:
		return FrameBottomLeft, true
	case r == rows-1 && c == cols-1:
		return FrameBottomRight, true
	case r == rows-1:
		return FrameBottom, true
	case c == 0:
		return FrameLeft, true
	case c == cols-1:
		return FrameRight, true
	}
	return 0, false
}

func (f TileFrame) drawTile(fb *gfx.Framebuffer, x, y, idx int) {
	f.Tiles.DrawTile(fb, x, y, idx, f.Fill)
}

// DrawTile 把第 idx 張圖塊貼在像素 (x, y)，色號 0 填 fill。文字引擎的
// `0xFF nn` 字碼走的就是這一條（`2376:0E4A`，索引 ＝ nn − 1）——右下角
// 清單的 ▲▼ 是 `\xFF\x11`／`\xFF\x15`（第 16／20 張），空白是第 32 張。
func (t *FrameTiles) DrawTile(fb *gfx.Framebuffer, x, y, idx int, fill uint8) {
	if t == nil || idx < 0 || idx >= len(t.Tiles) {
		return
	}
	img := t.Tiles[idx]
	for ty := 0; ty < img.H; ty++ {
		for tx := 0; tx < img.W; tx++ {
			v := img.Pix[ty*img.W+tx]
			if v == 0 {
				v = fill
			}
			fb.Set(x+tx, y+ty, v)
		}
	}
}

// DrawRect 是給既有呼叫端用的入口：參數與 Frame.Draw 相同（像素矩形），
// 內部換算回格。矩形一定是 `(格x×16＋1, (格y＋2)×20＋1)` 起、
// `(格數×16−2) × (列數×20−2)` 大——那兩格與那兩像素都是格座標系的原點
// 造成的（`docs/re/138`）。對不上格就不畫，寧可少一個框也不要畫錯位置。
func (f TileFrame) DrawRect(fb *gfx.Framebuffer, x, y, w, h int) bool {
	if (x-1)%FrameTileW != 0 || (y-1)%FrameTileH != 0 ||
		(w+2)%FrameTileW != 0 || (h+2)%FrameTileH != 0 {
		return false
	}
	cellY := (y-1)/FrameTileH - FrameCellOriginY/FrameTileH
	f.Draw(fb, (x-1)/FrameTileW, cellY, (w+2)/FrameTileW, (h+2)/FrameTileH)
	return true
}

// 內容格的旗標（視窗內容緩衝區一格的 `[+3]`，搬到畫面格表就是顏色的
// 高位元組）。`2376:1144`／`2376:1180`／`2376:1193` 依這些位元加裝飾，
// 位置與色號都寫死在程式裡（`docs/re/138`）。
const (
	CellUnderline   = 0x01 // 格子最後一列整條色 15（`docs/internals/25`）
	CellDither      = 0x02 // 網點
	CellBold        = 0x04 // 加粗
	CellButtonBody  = 0x08 // 按鈕的上下兩條邊
	CellButtonLeft  = 0x10 // 按鈕左端（畫在格內最右邊兩個像素）
	CellButtonRight = 0x20 // 按鈕右端（畫在格內最左邊兩個像素）
)

// 按鈕邊用的兩個色號：`2376:1270` 是 7、`2376:1278` 是 3。
const (
	ButtonLightInk = 7
	ButtonDarkInk  = 3
	// UnderlineInk 是底線的色號（`2376:1180` 設 15 再寫 `0xFFFF`）。
	UnderlineInk = 15
)

// DrawCellDecor 把一格的按鈕裝飾畫上去，(x, y) 是那一格的左上角像素。
//
// 三段的像素位置直接照 `2376:11B9`／`2376:11F1`／`2376:121B`：上下邊
// 各兩列、左端在格內第 14–15 行、右端在第 0–1 行，而且**最後一列與其他
// 列不同色**——左端最後一列的最右邊那個像素是暗色，右端第一列的最左邊
// 那個像素是亮色。四個角就是靠這兩個像素接起來的。
func DrawCellDecor(fb *gfx.Framebuffer, x, y int, flags uint8) {
	// 底線 `2376:1180`：格子最後一列整條填 15（`mov word es:[di-0x50], FFFF`，
	// `docs/internals/25`；收據 `fcmd-win` 的分隔列量到 y ＝ 格頂 ＋ 19）。
	if flags&CellUnderline != 0 {
		for i := 0; i < FrameTileW; i++ {
			fb.Set(x+i, y+FrameTileH-1, UnderlineInk)
		}
	}
	if flags&CellButtonBody != 0 {
		for i := 0; i < FrameTileW; i++ {
			fb.Set(x+i, y, ButtonLightInk)
			fb.Set(x+i, y+1, ButtonLightInk)
			fb.Set(x+i, y+FrameTileH-2, ButtonDarkInk)
			fb.Set(x+i, y+FrameTileH-1, ButtonDarkInk)
		}
	}
	if flags&CellButtonLeft != 0 {
		for r := 0; r < FrameTileH-1; r++ {
			fb.Set(x+FrameTileW-2, y+r, ButtonLightInk)
			fb.Set(x+FrameTileW-1, y+r, ButtonLightInk)
		}
		fb.Set(x+FrameTileW-2, y+FrameTileH-1, ButtonLightInk)
		fb.Set(x+FrameTileW-1, y+FrameTileH-1, ButtonDarkInk)
	}
	if flags&CellButtonRight != 0 {
		for r := 1; r < FrameTileH; r++ {
			fb.Set(x, y+r, ButtonDarkInk)
			fb.Set(x+1, y+r, ButtonDarkInk)
		}
		fb.Set(x, y, ButtonLightInk)
		fb.Set(x+1, y, ButtonDarkInk)
	}
}
