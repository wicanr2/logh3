package asset

import (
	"encoding/binary"
	"fmt"
)

// VO5 是結局畫面的圖（`docs/internals/18-vo5.md`）。
//
// 版面：`[寬(bytes) u8][高 u16]` 之後接壓縮資料，四個 plane 依序解、
// **共用同一個來源指標**。壓縮與 `.AMG` 同源，但以 byte 為單位：
// 兩個相同的 byte 代表重複組，後面跟著「總共幾個」（1 byte；為 0 時
// 改讀一個 word）。填法是直行——寫一個 byte 之後往下一列，走完整個
// 高度就回到頂端往右一格。
func DecodeVO5(d []byte) (*Image, error) {
	if len(d) < 3 {
		return nil, fmt.Errorf("VO5 只有 %d bytes", len(d))
	}
	w := int(d[0])
	h := int(binary.LittleEndian.Uint16(d[1:]))
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("VO5 尺寸不合理：%d×%d", w, h)
	}
	p := 3
	planes := make([][]byte, 4)
	for i := range planes {
		var err error
		planes[i], p, err = unpackVO5Plane(d, p, w, h)
		if err != nil {
			return nil, fmt.Errorf("第 %d 個 plane：%w", i, err)
		}
	}
	img := &Image{W: w * 8, H: h, Pix: make([]uint8, w*8*h)}
	for y := 0; y < h; y++ {
		for xb := 0; xb < w; xb++ {
			b0, b1 := planes[0][y*w+xb], planes[1][y*w+xb]
			b2, b3 := planes[2][y*w+xb], planes[3][y*w+xb]
			o := y*w*8 + xb*8
			for i := 0; i < 8; i++ {
				m := byte(0x80) >> uint(i)
				var v uint8
				if b0&m != 0 {
					v |= 1
				}
				if b1&m != 0 {
					v |= 2
				}
				if b2&m != 0 {
					v |= 4
				}
				if b3&m != 0 {
					v |= 8
				}
				img.Pix[o+i] = v
			}
		}
	}
	return img, nil
}

// VO5Consumed 回「解完四個 plane 用掉幾個位元組」。
// 正確的檔案會剛好用完——這是最便宜的自檢。
func VO5Consumed(d []byte) (int, error) {
	if len(d) < 3 {
		return 0, fmt.Errorf("VO5 只有 %d bytes", len(d))
	}
	w := int(d[0])
	h := int(binary.LittleEndian.Uint16(d[1:]))
	p := 3
	for i := 0; i < 4; i++ {
		var err error
		_, p, err = unpackVO5Plane(d, p, w, h)
		if err != nil {
			return p, err
		}
	}
	return p, nil
}

func unpackVO5Plane(d []byte, p, w, h int) ([]byte, int, error) {
	buf := make([]byte, w*h)
	di, col, left := 0, 0, h
	var prev int = -1
	step := func() bool {
		di += w
		left--
		if left == 0 {
			left = h
			di -= w*h - 1
			col++
			return col == w
		}
		return false
	}
	for {
		if p >= len(d) {
			return nil, p, fmt.Errorf("資料在第 %d 直行用完", col)
		}
		v := d[p]
		p++
		for {
			buf[di] = v
			prev = int(v)
			if step() {
				return buf, p, nil
			}
			if p >= len(d) {
				return nil, p, fmt.Errorf("資料在第 %d 直行用完", col)
			}
			v2 := d[p]
			p++
			if int(v2) != prev {
				v = v2
				continue
			}
			if p >= len(d) {
				return nil, p, fmt.Errorf("重複組沒有次數")
			}
			n := int(d[p])
			p++
			if n == 0 {
				if p+2 > len(d) {
					return nil, p, fmt.Errorf("重複組的長次數被截斷")
				}
				n = int(binary.LittleEndian.Uint16(d[p:]))
				p += 2
			}
			for k := 1; k < n; k++ {
				buf[di] = byte(prev)
				if step() {
					return buf, p, nil
				}
			}
			break
		}
	}
}
