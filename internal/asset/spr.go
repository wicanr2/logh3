package asset

import (
	"encoding/binary"
	"fmt"
)

// DecodeTiles 解未壓縮的 planar 圖塊庫（`SPRITE.SPR`）。
//
// 版面照原版 `GIN3PS.EXE` 2376:0D25 那支 blit：**每一列依序放 plane 0–3
// 各一個 word**，走完一列再往下。所以一張 w×h 的圖塊佔 (w/8)×h×4 bytes，
// 圖塊在檔案裡一張接一張，沒有檔頭也沒有索引。
//
// w 必須是 16 的倍數（原版的 blit 一次搬一個 word）。
func DecodeTiles(data []byte, w, h int) ([]*Image, error) {
	if w <= 0 || h <= 0 || w%16 != 0 {
		return nil, fmt.Errorf("圖塊尺寸不合理：%d×%d", w, h)
	}
	words := w / 16
	stride := words * 4 * 2 // 每列 4 個 plane 各一個 word
	size := stride * h
	if len(data)%size != 0 {
		return nil, fmt.Errorf("%d bytes 不是 %d×%d 圖塊（每張 %d bytes）的整數倍",
			len(data), w, h, size)
	}
	var out []*Image
	for off := 0; off+size <= len(data); off += size {
		img := &Image{W: w, H: h, Pix: make([]uint8, w*h), Used: size}
		for y := 0; y < h; y++ {
			row := data[off+y*stride:]
			for word := 0; word < words; word++ {
				for p := 0; p < 4; p++ {
					v := binary.LittleEndian.Uint16(row[(word*4+p)*2:])
					// 小端存放，但畫面上的位元是高位在左。
					be := v<<8 | v>>8
					for bit := 0; bit < 16; bit++ {
						if be&(0x8000>>bit) != 0 {
							img.Pix[y*w+word*16+bit] |= 1 << p
						}
					}
				}
			}
		}
		out = append(out, img)
	}
	return out, nil
}
