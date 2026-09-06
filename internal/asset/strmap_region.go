package asset

import (
	"encoding/binary"
	"fmt"
)

// StarRegion 是星圖上的一個星域（`docs/internals/25-strmap-bit.md`）。
//
// 形狀存成 1bpp 遮罩，單位與星圖底圖相同：16 × 20 的格子，一格 40 bytes、
// 每列 2 bytes、word 高位在左。遮罩只在 `GX,GY` 起 `GW×GH` 格的範圍內。
type StarRegion struct {
	GX, GY, GW, GH int
	Mask           []byte
}

// Contains 回答星圖上的某個像素屬不屬於這個星域。
func (r *StarRegion) Contains(x, y int) bool {
	cx, cy := x/strmapCellW-r.GX, y/strmapCellH-r.GY
	if cx < 0 || cy < 0 || cx >= r.GW || cy >= r.GH {
		return false
	}
	cell := (cy*r.GW + cx) * strmapMaskCell
	row := cell + y%strmapCellH*2
	word := uint16(r.Mask[row])<<8 | uint16(r.Mask[row+1])
	return word>>uint(15-x%strmapCellW)&1 != 0
}

const (
	strmapMaskCell  = strmapCellW * strmapCellH / 8 // 40 bytes
	strmapRegionRec = 6                             // 格x, 格y, 寬, 高, 資料位移(u16)
	StarRegionCount = 32
)

// DecodeSTRMAPRegions 解 `STRMAP.RGN`：32 個星域的形狀。
//
// 表在檔案最前面，每筆 6 bytes；位移是相對**表結束之後**的資料區。
// 32 筆與 `STELLAR.DT*` 的 32 筆一一對應（該檔 `+17` 就是 0…31 的編號）。
func DecodeSTRMAPRegions(d []byte) ([]StarRegion, error) {
	head := StarRegionCount * strmapRegionRec
	if len(d) < head {
		return nil, fmt.Errorf("STRMAP.RGN 只有 %d bytes", len(d))
	}
	out := make([]StarRegion, 0, StarRegionCount)
	want := 0
	for i := 0; i < StarRegionCount; i++ {
		p := i * strmapRegionRec
		r := StarRegion{
			GX: int(d[p]), GY: int(d[p+1]),
			GW: int(d[p+2]), GH: int(d[p+3]),
		}
		off := int(binary.LittleEndian.Uint16(d[p+4:]))
		if r.GW == 0 || r.GH == 0 {
			return nil, fmt.Errorf("第 %d 個星域大小是 %d×%d", i, r.GW, r.GH)
		}
		// 位移必須正好接在前一個之後——對不上就表示版面解錯了，
		// 不要繼續解出一堆看起來像形狀的垃圾。
		if off != want {
			return nil, fmt.Errorf("第 %d 個星域的位移是 %04X，預期 %04X", i, off, want)
		}
		n := r.GW * r.GH * strmapMaskCell
		if head+off+n > len(d) {
			return nil, fmt.Errorf("第 %d 個星域的遮罩超出檔案", i)
		}
		r.Mask = d[head+off : head+off+n]
		out = append(out, r)
		want += n
	}
	if head+want != len(d) {
		return nil, fmt.Errorf("表 %d ＋ 遮罩 %d ≠ 檔案 %d", head, want, len(d))
	}
	return out, nil
}
