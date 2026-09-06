package asset

import "fmt"

// STRMAP 的版面（`docs/internals/25-strmap-bit.md`）。
const (
	strmapCellW  = 16 // 一格的寬（像素），與整個 UI 的格子單位相同
	strmapCellH  = 20
	strmapCols   = 20 // 星圖視窗是 20 × 15 格
	strmapRows   = 15
	strmapCell   = strmapCellW * strmapCellH / 2 // 160 bytes
	STRMAPWidth  = strmapCellW * strmapCols      // 320
	STRMAPHeight = strmapCellH * strmapRows      // 300
)

// DecodeSTRMAP 把 `STRMAP.BIT` 解成主畫面的星圖底圖。
//
// 一格 16 × 20、160 bytes，格子 row-major 排；格內每一列 8 bytes ＝
// **四個平面各 2 bytes 交錯**（`bit0` 的平面在前），每個 word 高位在左。
//
// ⚠ **低兩個位元不是這張圖的最終顏色。** 遊戲載入之後會依勢力與治安
// 重畫 `bit0`／`bit1`，檔案裡只有 `bit2`／`bit3` 的線稿（星系連線、
// 星球、名稱、邊框）是最終的。拿整張去跟原版畫面比只有 77.5% 相同，
// 只比高兩個位元是 92.0%。
func DecodeSTRMAP(d []byte) (*Image, error) {
	need := strmapCols * strmapRows * strmapCell
	if len(d) < need {
		return nil, fmt.Errorf("STRMAP 只有 %d bytes，需要 %d", len(d), need)
	}
	img := &Image{
		W:   STRMAPWidth,
		H:   STRMAPHeight,
		Pix: make([]uint8, STRMAPWidth*STRMAPHeight),
	}
	for cy := 0; cy < strmapRows; cy++ {
		for cx := 0; cx < strmapCols; cx++ {
			base := (cy*strmapCols + cx) * strmapCell
			for y := 0; y < strmapCellH; y++ {
				row := base + y*8
				out := (cy*strmapCellH+y)*STRMAPWidth + cx*strmapCellW
				for p := 0; p < 4; p++ {
					word := uint16(d[row+p*2])<<8 | uint16(d[row+p*2+1])
					for k := 0; k < strmapCellW; k++ {
						if word>>uint(15-k)&1 != 0 {
							img.Pix[out+k] |= 1 << uint(p)
						}
					}
				}
			}
		}
	}
	return img, nil
}
