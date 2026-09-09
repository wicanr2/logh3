package asset

import (
	"encoding/binary"
	"fmt"
)

// 開頭動畫（`IBM16.350` ＋ `IBMFLS16.350`，`docs/internals/19`）。
//
// 374 格，一格四個 plane，每個 plane 後面接一個分隔位元組。一個 plane 是
// 80 × 227 ＝ 18,160 bytes 的位元圖，token 解出來的位元組**一欄一欄往下**
// XOR 進去。
const (
	IntroPlaneStride = 80                                // 一列幾個位元組
	IntroPlaneRows   = 227                               // 幾列
	IntroPlaneSize   = IntroPlaneStride * IntroPlaneRows // 18,160
	IntroPlanes      = 4
	IntroWidth       = IntroPlaneStride * 8
)

// IntroFrames 讀 `IBMFLS16.350`：374 個 uint16，每一個是一格的壓縮長度。
func IntroFrames(lengths []byte) ([]int, error) {
	if len(lengths)%2 != 0 || len(lengths) == 0 {
		return nil, fmt.Errorf("長度表 %d bytes 不是 uint16 的整數倍", len(lengths))
	}
	out := make([]int, len(lengths)/2)
	for i := range out {
		out[i] = int(binary.LittleEndian.Uint16(lengths[2*i:]))
	}
	return out, nil
}

// IntroDecoder 一格一格解開頭動畫。每一格是 XOR 上去的，所以 Planes
// 一直是「到目前為止累加的結果」——場景與場景之間會整個歸零。
type IntroDecoder struct {
	Planes [IntroPlanes][]uint8

	data   []byte
	frames []int
	next   int // 下一格是第幾格
	pos    int // 下一格在 data 裡的位移
}

// NewIntroDecoder 建一個解碼器；data 是 `IBM16.350`、lengths 是 `IBMFLS16.350`。
func NewIntroDecoder(data, lengths []byte) (*IntroDecoder, error) {
	frames, err := IntroFrames(lengths)
	if err != nil {
		return nil, err
	}
	total := 0
	for _, n := range frames {
		total += n
	}
	if total != len(data) {
		return nil, fmt.Errorf("長度表加起來 %d，但資料是 %d bytes", total, len(data))
	}
	d := &IntroDecoder{data: data, frames: frames}
	for i := range d.Planes {
		d.Planes[i] = make([]uint8, IntroPlaneSize)
	}
	return d, nil
}

// Frames 回總格數。
func (d *IntroDecoder) Frames() int { return len(d.frames) }

// Next 解下一格。跑完最後一格回 false。
func (d *IntroDecoder) Next() (bool, error) {
	if d.next >= len(d.frames) {
		return false, nil
	}
	used := 0
	for p := 0; p < IntroPlanes; p++ {
		n, err := decodeIntroPlane(d.data, d.pos+used, d.Planes[p])
		if err != nil {
			return false, fmt.Errorf("第 %d 格 plane %d：%w", d.next, p, err)
		}
		used += n + 1 // 每個 plane 後面一個分隔位元組
	}
	if used != d.frames[d.next] {
		return false, fmt.Errorf("第 %d 格吃掉 %d bytes，長度表寫 %d",
			d.next, used, d.frames[d.next])
	}
	d.pos += d.frames[d.next]
	d.next++
	return true, nil
}

// Pixels 把四個 plane 併成一張 640 × 227 的色號圖。
func (d *IntroDecoder) Pixels() *Image {
	img := &Image{W: IntroWidth, H: IntroPlaneRows,
		Pix: make([]uint8, IntroWidth*IntroPlaneRows)}
	for y := 0; y < IntroPlaneRows; y++ {
		for bx := 0; bx < IntroPlaneStride; bx++ {
			o := y*IntroPlaneStride + bx
			for k := 0; k < 8; k++ {
				m := uint8(0x80) >> k
				var c uint8
				for p := 0; p < IntroPlanes; p++ {
					if d.Planes[p][o]&m != 0 {
						c |= 1 << p
					}
				}
				img.Pix[y*IntroWidth+bx*8+k] = c
			}
		}
	}
	return img
}

// decodeIntroPlane 解一個 plane，回吃掉的位元組數（不含分隔位元組）。
//
// 照原版 `041F:024B` 逐條寫的（`docs/re/108` 讀的那一段）：
//
//   - 停的條件是 `di == limit`（`cmp di,cs:[0248] / jz`），**不是 >=**。
//     讀成「大於等於」的話一格只吃 21 個位元組（`docs/re/72`）。
//   - `di` 每格 `+80`，一超過 18,160 就 `-18,159`，所以換欄之後 `di`
//     就是欄號 1、2、3…。
//   - **limit 從 0x100 改成 0x50 的時機是「一個 run token 跑完」**
//     （原版在兩條 run 迴圈的出口各有一次 `mov cs:[0248],50h`），
//     單一 byte 的那條路徑**不動它**。換欄本身不改 limit。
//     兩種寫法在 374 格上的輸出相同，這裡照原版的形狀寫。
func decodeIntroPlane(data []byte, pos int, buf []uint8) (int, error) {
	const wrap = IntroPlaneSize
	di, limit, start := 0, 0x100, pos
	advance := func() {
		di += IntroPlaneStride
		if di >= wrap {
			di -= wrap - 1
		}
	}
	for {
		if di == limit {
			return pos - start, nil
		}
		if pos+1 >= len(data) {
			return 0, fmt.Errorf("資料在位移 %d 用完了", pos)
		}
		a, b := data[pos], data[pos+1]
		pos++
		if b != a {
			buf[di] ^= a
			advance()
			continue
		}
		pos++ // 跳脫：吃掉第二個
		if pos >= len(data) {
			return 0, fmt.Errorf("資料在位移 %d 用完了", pos)
		}
		n := int(data[pos])
		pos++
		count := n & 0x7F
		if n&0x80 == 0 {
			if pos >= len(data) {
				return 0, fmt.Errorf("資料在位移 %d 用完了", pos)
			}
			count = n<<8 | int(data[pos])
			pos++
		}
		for i := 0; i < count; i++ {
			if a != 0 {
				buf[di] ^= a
			}
			advance()
		}
		limit = 0x50 // run 跑完才改（原版兩條 run 迴圈的出口）
	}
}
