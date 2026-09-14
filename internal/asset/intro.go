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

// IntroPlayer 是開頭動畫播放端需要的最小介面。
//
// 來源解碼器與「照 OPEN.EXE 實際播放順序」的包裝器都實作這個介面，
// 讓播放器不必知道素材是怎麼分塊、何時清除 plane 或重播尾段。
type IntroPlayer interface {
	Frames() int
	Next() (bool, error)
	Pixels() *Image
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

// introChunkFrames 是 OPEN.EXE 的每塊格數表（`docs/internals/data/
// intro-chunks.json`）。第 23 塊的表值是 10，雖然檔案實際能放下第 349
// 格；原版照表換塊，所以第 349 格不會被解碼。這些數字是 dosgolem
// `introfull` 收據量到的遊戲資料，不是依目前 remake 的畫面猜出來的。
var introChunkFrames = [...]int{
	117, 15, 9, 7, 7, 7, 7, 8, 9, 7,
	6, 6, 5, 5, 5, 5, 7, 34, 12, 11,
	13, 17, 20, 10, 5, 4, 3, 3, 4, 5,
}

// IntroPlaybackDecoder 把來源資料解碼成原版 OPEN.EXE 實際播出的序列。
//
// 原版先播 0–348，略過 349，再把 350–373 播三趟；每趟尾段開始前
// 四個 plane 都清成零。來源 IntroDecoder 仍保留「每個檔案格都解一次」
// 的純解碼契約，這層只負責玩家可見的播放順序。
type IntroPlaybackDecoder struct {
	data    []byte
	lengths []byte
	raw     *IntroDecoder
	seq     []int
	next    int
}

// NewIntroPlaybackDecoder 建立與原版播放順序相同的開頭動畫解碼器。
func NewIntroPlaybackDecoder(data, lengths []byte) (*IntroPlaybackDecoder, error) {
	raw, err := NewIntroDecoder(data, lengths)
	if err != nil {
		return nil, err
	}
	total := 0
	for _, n := range introChunkFrames {
		total += n
	}
	// 第 23 塊的表少算一格（第 349 格），所以表值總和比檔案長度
	// 對應的來源格數少一格；這個差異正是原版略過第 349 格的證據。
	if total != raw.Frames()-1 {
		return nil, fmt.Errorf("分塊格數加總 %d，但來源扣除漏播格後有 %d 格", total, raw.Frames()-1)
	}
	seq := make([]int, 0, 349+24*3)
	for i := 0; i < 349; i++ {
		seq = append(seq, i)
	}
	for pass := 0; pass < 3; pass++ {
		for i := 350; i < raw.Frames(); i++ {
			seq = append(seq, i)
		}
	}
	return &IntroPlaybackDecoder{data: data, lengths: lengths, raw: raw, seq: seq}, nil
}

// Frames 回玩家可見的格數：349 格前段＋24 格尾段的三趟。
func (d *IntroPlaybackDecoder) Frames() int { return len(d.seq) }

// Next 解並顯示下一格。尾段每一趟都先清掉 XOR 累加的 plane。
func (d *IntroPlaybackDecoder) Next() (bool, error) {
	if d.next >= len(d.seq) {
		return false, nil
	}
	target := d.seq[d.next]
	if target < 350 {
		// 前段依序走到 348；349 只在第一次進尾段時消耗，永不顯示。
		if target != d.raw.next {
			return false, fmt.Errorf("播放序列要第 %d 格，來源下一格是 %d", target, d.raw.next)
		}
		if _, err := d.raw.Next(); err != nil {
			return false, err
		}
		d.next++
		return true, nil
	}

	if target == 350 {
		if d.next == 349 {
			// 第一次進尾段：把原版漏掉的第 349 格吃掉，再清 plane。
			if _, err := d.raw.Next(); err != nil {
				return false, err
			}
		} else {
			// 後兩趟從檔案開頭重建到第 349 格；清除後只保留
			// 350 格開始的 XOR 增量，正是原版的載入器行為。
			raw, err := NewIntroDecoder(d.data, d.lengths)
			if err != nil {
				return false, err
			}
			for i := 0; i < 350; i++ {
				if _, err := raw.Next(); err != nil {
					return false, err
				}
			}
			d.raw = raw
		}
		for p := range d.raw.Planes {
			for i := range d.raw.Planes[p] {
				d.raw.Planes[p][i] = 0
			}
		}
	}
	if target != d.raw.next {
		return false, fmt.Errorf("播放序列要第 %d 格，來源下一格是 %d", target, d.raw.next)
	}
	if _, err := d.raw.Next(); err != nil {
		return false, err
	}
	d.next++
	return true, nil
}

// Pixels 回目前播放格的色號圖。
func (d *IntroPlaybackDecoder) Pixels() *Image { return d.raw.Pixels() }

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
