// Package asset 讀原版資料檔。**本套件不內嵌任何原版資料**——
// 解碼器只描述格式，檔案由玩家自備。
package asset

import (
	"encoding/binary"
	"fmt"
)

// AMG 是 .AMG 圖庫（`FACE.AMG` 是人物頭像）：一張索引表加上連續的影像資料。
type AMG struct {
	Entries []AMGEntry
	data    []byte
	// ArtworkPrefix 只在 HD 顯示模式指定；原版解碼保持。
	ArtworkPrefix string
}

// AMGEntry 是索引表的一筆：影像的位元組長度與檔內位移。
type AMGEntry struct {
	Size   int
	Offset int
}

// Image 是解碼後的一張圖：4 位元色號，一像素一 byte。
type Image struct {
	W, H       int
	Pix        []uint8
	ArtworkKey string
	StarMap    *StarMapLayers

	// Used 是解碼實際吃掉的位元組數。**應該剛好等於索引表寫的長度**——
	// 對不上就是解碼走偏了，而畫面多半還是看得出人形，不會報錯。
	Used int
}

// OpenAMG 解析索引表。索引表在檔頭，長度就是第一筆的位移
// （每筆 6 bytes，所以筆數 ＝ 第一筆位移 ÷ 6）。
func OpenAMG(data []byte) (*AMG, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("AMG 太短：%d bytes", len(data))
	}
	first := int(binary.LittleEndian.Uint32(data[2:]))
	if first%6 != 0 || first == 0 || first > len(data) {
		return nil, fmt.Errorf("AMG 索引表長度不合理：%d", first)
	}
	a := &AMG{data: data}
	for i := 0; i < first/6; i++ {
		sz := int(binary.LittleEndian.Uint16(data[i*6:]))
		off := int(binary.LittleEndian.Uint32(data[i*6+2:]))
		if off+sz > len(data) {
			return nil, fmt.Errorf("第 %d 筆超出檔案：%d+%d", i, off, sz)
		}
		a.Entries = append(a.Entries, AMGEntry{Size: sz, Offset: off})
	}
	return a, nil
}

// Decode 解出第 i 張圖。演算法照原版 `GIN3PS.EXE` 348A:02F9 那支常式：
//
//   - 影像是**直行掃描**的 planar 資料。一次取兩個 byte，寫進同一個 x 的
//     相鄰兩列（原版是 `add di,80`，mode 12h 的每列位元組數）。
//   - 兩個 byte **相同**時是重複組：後面跟著「第二個值」與「次數」，
//     那一對值重複寫這麼多次。
//   - 走完每直行的 h 個 word（＝ 2h 列）就換下一直行；走完 w 個直行
//     就換下一個 plane（原版是把 map mask 左移一位）。
//
// **四個 plane 共用同一個狀態機**：重複組的次數可以跨過 plane 邊界。
// 把每個 plane 當成獨立一段解，跨界的那幾個 word 會落在錯的位置——
// 症狀是畫面對得起來但 bit0 有幾百個像素不同，看起來像色盤問題。
func (a *AMG) Decode(i int) (*Image, error) {
	if i < 0 || i >= len(a.Entries) {
		return nil, fmt.Errorf("沒有第 %d 張（共 %d 張）", i, len(a.Entries))
	}
	e := a.Entries[i]
	img, err := DecodeAMGImage(a.data[e.Offset : e.Offset+e.Size])
	if err != nil {
		return nil, fmt.Errorf("第 %d 張：%w", i, err)
	}
	if a.ArtworkPrefix != "" {
		img.ArtworkKey = fmt.Sprintf("%s/%02d", a.ArtworkPrefix, i)
	}
	return img, nil
}

// DecodeAMGImage 解一張 `.AMG` 影像。
//
// `ANI_PNL?.A98`（戰鬥畫面的面板）**整個檔就是一張這種影像**，
// 沒有前面那張索引表——所以解碼的部分要能單獨用（`docs/internals/37`）。
func DecodeAMGImage(img []byte) (*Image, error) {
	if len(img) < 13 {
		return nil, fmt.Errorf("太短：%d bytes", len(img))
	}
	w := int(binary.LittleEndian.Uint16(img[8:]))  // 直行數，每行 8 像素
	h := int(binary.LittleEndian.Uint16(img[10:])) // 每直行的 word 數
	if w <= 0 || h <= 0 || w > 256 || h > 256 {
		return nil, fmt.Errorf("尺寸不合理：%d×%d", w, h)
	}
	rows := h * 2
	planes := make([][]uint8, 4)
	for k := range planes {
		planes[k] = make([]uint8, w*rows)
	}
	pl, col, di, left, p := 0, 0, 0, h, 12
	put := func(x, y uint8) {
		if pl > 3 {
			return
		}
		planes[pl][di] = x
		di += w
		planes[pl][di] = y
		di += w
		if left--; left == 0 {
			col++
			di = col
			left = h
			if col == w {
				pl, col, di = pl+1, 0, 0
			}
		}
	}
	for pl <= 3 && p+1 < len(img) {
		lo, hi := img[p], img[p+1]
		p += 2
		if lo != hi {
			put(lo, hi)
			continue
		}
		if p+1 >= len(img) {
			break
		}
		v2, cnt := img[p], int(img[p+1])
		p += 2
		if cnt == 0 {
			cnt = 65536
		}
		for k := 0; k < cnt && pl <= 3; k++ {
			put(lo, v2)
		}
	}
	out := &Image{W: w * 8, H: rows, Pix: make([]uint8, w*8*rows), Used: p}
	for y := 0; y < rows; y++ {
		for bx := 0; bx < w; bx++ {
			for bit := 0; bit < 8; bit++ {
				var v uint8
				for k := 0; k < 4; k++ {
					if planes[k][y*w+bx]&(0x80>>bit) != 0 {
						v |= 1 << k
					}
				}
				out.Pix[y*w*8+bx*8+bit] = v
			}
		}
	}
	return out, nil
}
