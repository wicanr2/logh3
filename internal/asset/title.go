package asset

import "fmt"

const TitleWidth, TitleRows = 640, 350

// DecodeTitle350 解SPTIT.350四個直行XOR-RLE plane，見docs/spec/111。
// 每個plane後有01分隔byte；不接受部分圖或額外尾端資料。
func DecodeTitle350(data []byte) (*Image, error) {
	const stride, cells = TitleWidth / 8, TitleWidth / 8 * TitleRows
	var planes [4][]byte
	pos := 0
	for pl := range planes {
		planes[pl] = make([]byte, cells)
		for written := 0; written < cells; {
			if pos+1 >= len(data) {
				return nil, fmt.Errorf("SPTIT plane %d在%d截斷", pl, pos)
			}
			a, b := data[pos], data[pos+1]
			pos++
			run := 1
			if a == b {
				pos++
				if pos >= len(data) {
					return nil, fmt.Errorf("SPTIT run在%d截斷", pos)
				}
				n := int(data[pos])
				pos++
				run = n & 127
				if n&128 == 0 {
					if pos >= len(data) {
						return nil, fmt.Errorf("SPTIT長run在%d截斷", pos)
					}
					run = n<<8 | int(data[pos])
					pos++
				}
			}
			if run == 0 || run > cells-written {
				return nil, fmt.Errorf("SPTIT plane %d run %d超出剩餘%d", pl, run, cells-written)
			}
			if a != 0 {
				for i := written; i < written+run; i++ {
					planes[pl][(i%TitleRows)*stride+i/TitleRows] ^= a
				}
			}
			written += run
		}
		if pos >= len(data) || data[pos] != 1 {
			return nil, fmt.Errorf("SPTIT plane %d缺01分隔byte，位移%d", pl, pos)
		}
		pos++
	}
	if pos != len(data) {
		return nil, fmt.Errorf("SPTIT消耗%d，來源%d bytes", pos, len(data))
	}
	img := &Image{W: TitleWidth, H: TitleRows, Used: pos, Pix: make([]byte, TitleWidth*TitleRows)}
	for y := 0; y < TitleRows; y++ {
		for x := 0; x < TitleWidth; x++ {
			mask := byte(128 >> (x % 8))
			for pl := range planes {
				if planes[pl][y*stride+x/8]&mask != 0 {
					img.Pix[y*TitleWidth+x] |= 1 << pl
				}
			}
		}
	}
	return img, nil
}
