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

// DecodeSheet 把圖塊庫拼成一張大圖（`STRMAP.BIT` 是 20×15 塊的星域地圖）。
//
// 圖塊在檔案裡是 row-major：先左到右排滿一列，再換下一列。
func DecodeSheet(data []byte, tw, th, cols int) (*Image, error) {
	tiles, err := DecodeTiles(data, tw, th)
	if err != nil {
		return nil, err
	}
	if cols <= 0 || len(tiles)%cols != 0 {
		return nil, fmt.Errorf("%d 張圖塊排不成每列 %d 張", len(tiles), cols)
	}
	rows := len(tiles) / cols
	out := &Image{W: cols * tw, H: rows * th, Pix: make([]uint8, cols*tw*rows*th), Used: len(data)}
	for i, t := range tiles {
		ox, oy := (i%cols)*tw, (i/cols)*th
		for y := 0; y < th; y++ {
			copy(out.Pix[(oy+y)*out.W+ox:], t.Pix[y*tw:(y+1)*tw])
		}
	}
	return out, nil
}

// PaintRegion 在遮罩涵蓋的範圍內，把圖上**留白（色號 0）**的像素塗成
// colour；已經有東西的地方（航路、星系名、星球）原封不動。
//
// 這是原版 `GIN3PS.EXE` 2376:0BDC 那支常式的語意。它在 planar 位元上算：
// 先取原圖四個 plane 的 OR、取反，再與遮罩相 AND，得到「遮罩內、
// 原圖是空的」那些位元，然後依顏色的每個位元 OR 進對應的 plane。
// 在色號陣列上就是這一行條件。
//
// mask 與 img 同尺寸，非 0 表示涵蓋。
func PaintRegion(img *Image, mask []uint8, colour uint8) {
	for i := range img.Pix {
		if i < len(mask) && mask[i] != 0 && img.Pix[i] == 0 {
			img.Pix[i] = colour
		}
	}
}
