package asset

import (
	"encoding/binary"
	"fmt"
)

// Region 是一個星域的著色範圍（`STRMAP.RGN` 的一筆）。
type Region struct {
	BX, BY, BW, BH int // 以圖塊為單位的位置與大小
	Mask           []byte
}

// RGNHeader 是 `STRMAP.RGN` 檔頭的長度；每筆 6 bytes，所以有 32 個區域。
const RGNHeader = 192

// OpenRGN 解析星域區域遮罩。
//
// 檔頭每筆 6 bytes：`{塊x, 塊y, 塊寬, 塊高, u16 位移}`，位移相對資料區
// （檔頭之後）。遮罩是 1bpp，**以圖塊為單位排列**：每塊 40 bytes
// （20 列 × 2 bytes），塊之間 row-major。
//
// ⚠ 「逐像素列橫跨整個區域」那種排法在一個圖塊列上同樣是 200 bytes，
// 光看指標的步進分不出來——解錯的話區域邊緣會歪掉，而地圖整體
// 看起來仍然正常。
func OpenRGN(data []byte, tw, th int) ([]Region, error) {
	if len(data) < RGNHeader {
		return nil, fmt.Errorf("STRMAP.RGN 太短：%d bytes", len(data))
	}
	per := tw * th / 8 // 每塊的遮罩位元組數
	var out []Region
	for i := 0; i+6 <= RGNHeader; i += 6 {
		bx, by := int(data[i]), int(data[i+1])
		bw, bh := int(data[i+2]), int(data[i+3])
		off := int(binary.LittleEndian.Uint16(data[i+4:]))
		n := bw * bh * per
		if bw <= 0 || bh <= 0 || RGNHeader+off+n > len(data) {
			return nil, fmt.Errorf("第 %d 個區域超出檔案：%d×%d @%d", i/6, bw, bh, off)
		}
		out = append(out, Region{BX: bx, BY: by, BW: bw, BH: bh,
			Mask: data[RGNHeader+off : RGNHeader+off+n]})
	}
	return out, nil
}

// Covers 回這個區域內的相對座標有沒有被遮罩涵蓋。
func (r Region) Covers(x, y, tw, th int) bool {
	if x < 0 || y < 0 || x >= r.BW*tw || y >= r.BH*th {
		return false
	}
	per := tw * th / 8
	bi := (y/th)*r.BW + x/tw
	i := bi*per + (y%th)*(tw/8) + (x%tw)/8
	return r.Mask[i]&(0x80>>(x%8)) != 0
}

// Paint 把這個區域塗成 colour：只塗 base 上留白（色號 0）的像素。
// base 是原始底圖，dst 是工作副本——原版也是這樣分開的，
// 所以同一塊可以反覆重塗而不會越塗越髒。
func (r Region) Paint(dst, base *Image, colour uint8, tw, th int) {
	for y := 0; y < r.BH*th; y++ {
		gy := r.BY*th + y
		if gy >= dst.H {
			break
		}
		for x := 0; x < r.BW*tw; x++ {
			gx := r.BX*tw + x
			if gx >= dst.W || !r.Covers(x, y, tw, th) {
				continue
			}
			if base.Pix[gy*base.W+gx] == 0 {
				dst.Pix[gy*dst.W+gx] = colour
			}
		}
	}
}
