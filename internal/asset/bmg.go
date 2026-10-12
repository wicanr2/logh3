package asset

import (
	"encoding/binary"
	"fmt"
)

// BMG 是 `COMANI.BMG`：交戰演出的動畫庫（`docs/internals/20`、`docs/re/328`）。
//
// 檔頭 u16 是目錄長度（10 bytes 一筆），目錄每筆 `char[4]` 名稱、u16 長度、u32 位移，
// **位移從檔案 +2 算起**（原版 seek 到「位移＋2」）。每一段是一串格：
//
//	+0  u16 格數
//	格：+0 u8 格種（0 覆寫、1 XOR 差分）、+1 u8（全 0，程式碼不讀）、
//	    +2 u16 到下一格的距離（從 +4 起算）、+4 u16 VRAM 位移（列×80＋位元組欄，
//	    相對於視窗原點）、+6 u16 直行數、+8 u16 高（格種 0 是組數、格種 1 是列數）、
//	    +10 壓縮資料
//
// 畫在 VRAM 的 planar 位元組上（`348A:01DC` 靜圖只畫第 0 格、`348A:0228` 整段）。
type BMG struct {
	data    []byte
	entries map[string]bmgEntry
	names   []string
	static  map[string]*Image
}

type bmgEntry struct{ size, off int }

// BMGFrame 是一段裡的一格。
type BMGFrame struct {
	Kind   int // 0 覆寫、1 XOR
	Offset int // VRAM 位移（列 × 80 ＋ 位元組欄）
	Cols   int // 直行數（一欄 8 像素）
	Height int // 格種 0：組數（× 2 列）；格種 1：列數
	data   []byte
}

// OpenBMG 解析目錄。
func OpenBMG(data []byte) (*BMG, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("COMANI.BMG 太短")
	}
	dirLen := int(binary.LittleEndian.Uint16(data))
	if dirLen%10 != 0 || 2+dirLen > len(data) {
		return nil, fmt.Errorf("COMANI.BMG 目錄長度 %d 不合理", dirLen)
	}
	b := &BMG{data: data, entries: map[string]bmgEntry{}}
	for p := 2; p < 2+dirLen; p += 10 {
		name := string(data[p : p+4])
		size := int(binary.LittleEndian.Uint16(data[p+4:]))
		off := 2 + int(binary.LittleEndian.Uint32(data[p+6:]))
		if off+size > len(data) {
			return nil, fmt.Errorf("COMANI.BMG %q 超出檔尾", name)
		}
		b.entries[name] = bmgEntry{size, off}
		b.names = append(b.names, name)
	}
	return b, nil
}

// Names 回目錄裡的段名（照目錄順序；名稱是原始位元組，不一定是 ASCII）。
func (b *BMG) Names() []string { return b.names }

// StaticPicture是原前兩byte的D0首格，用於共用HD艦型；無資料就回錯誤。
func (b *BMG) StaticPicture(prefix string) (*Image, error) {
	if len(prefix) != 2 {
		return nil, fmt.Errorf("動畫前綴長度不是2")
	}
	if im := b.static[prefix]; im != nil {
		return im, nil
	}
	frames, err := b.Frames(prefix + "D0")
	if err != nil {
		return nil, err
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("靜態圖段是空的")
	}
	planes := PlanesFrom(make([]byte, 640*480))
	if err := frames[0].Draw(planes, 0); err != nil {
		return nil, err
	}
	pix := planes.Pixels()
	im := &Image{W: 256, H: 102, Pix: make([]byte, 256*102), ArtworkKey: fmt.Sprintf("%X", []byte(prefix))}
	for y := 0; y < 102; y++ {
		copy(im.Pix[y*256:], pix[y*640:y*640+256])
	}
	if b.static == nil {
		b.static = make(map[string]*Image)
	}
	b.static[prefix] = im
	return im, nil
}

// Frames 回一段的所有格。
func (b *BMG) Frames(name string) ([]BMGFrame, error) {
	e, ok := b.entries[name]
	if !ok {
		return nil, fmt.Errorf("COMANI.BMG 沒有 %q", name)
	}
	seg := b.data[e.off : e.off+e.size]
	if len(seg) < 2 {
		return nil, fmt.Errorf("COMANI.BMG %q 太短", name)
	}
	n := int(binary.LittleEndian.Uint16(seg))
	var out []BMGFrame
	p := 2
	for i := 0; i < n; i++ {
		if p+10 > len(seg) {
			return nil, fmt.Errorf("COMANI.BMG %q 第 %d 格超出段尾", name, i)
		}
		next := p + 4 + int(binary.LittleEndian.Uint16(seg[p+2:]))
		if next > len(seg) || next < p+10 {
			return nil, fmt.Errorf("COMANI.BMG %q 第 %d 格長度不合理", name, i)
		}
		out = append(out, BMGFrame{
			Kind:   int(seg[p]),
			Offset: int(binary.LittleEndian.Uint16(seg[p+4:])),
			Cols:   int(seg[p+6]), // 高位元組程式碼不讀（`348A:0281`）
			Height: int(seg[p+8]),
			data:   seg[p+10 : next],
		})
		p = next
	}
	return out, nil
}

// Planes 是 640×480 mode 12h 的四個 plane，一列 80 bytes（`[plane][列*80+欄]`）。
type Planes [4][]byte

// PlanesFrom 把 640×480 的色號畫面拆成四個 plane（色號第 k 位 ＝ plane k）。
func PlanesFrom(pix []byte) Planes {
	var p Planes
	for k := range p {
		p[k] = make([]byte, 80*480)
	}
	for i, v := range pix {
		x, y := i%640, i/640
		for k := 0; k < 4; k++ {
			if v>>k&1 != 0 {
				p[k][y*80+x/8] |= 0x80 >> (x % 8)
			}
		}
	}
	return p
}

// Pixels 把四個 plane 合回 640×480 的色號。
func (p Planes) Pixels() []byte {
	out := make([]byte, 640*480)
	for i := range out {
		x, y := i%640, i/640
		var v byte
		for k := 0; k < 4; k++ {
			if p[k][y*80+x/8]&(0x80>>(x%8)) != 0 {
				v |= 1 << k
			}
		}
		out[i] = v
	}
	return out
}

// BMGWindowBase 回演出視窗 (x, y) 格的 VRAM 位移（`3A8D:003A`：y×320＋x＋0xC80）；
// 左視窗 x＝1 → 像素 (8,56)、右視窗 x＝47 → (376,56)。
func BMGWindowBase(x, y int) int { return y*320 + x + 0x0C80 }

// Draw 把一格畫到 base（`BMGWindowBase`）上：格種 0 照 `.AMG` 的狀態機覆寫
// （`348A:02F9`），格種 1 是 XOR 差分（`348A:03A3`：非 0 的 byte 與那一列 XOR、
// 0 後面接「跳過幾列」，可以跨直行、跨 plane）。
func (f BMGFrame) Draw(p Planes, base int) error {
	_, err := f.draw(p, base)
	return err
}

// draw 回吃掉的資料位元組數（每一格都應該剛好吃完自己那一段）。
func (f BMGFrame) draw(p Planes, base int) (int, error) {
	if f.Cols <= 0 || f.Height <= 0 {
		return 0, fmt.Errorf("格的直行數／高是 0")
	}
	origin := (base + f.Offset) & 0xFFFF
	plane, col, cnt, di := 0, 0, 0, origin
	src := f.data
	q := 0
	need := func(n int) error {
		if q+n > len(src) {
			return fmt.Errorf("格資料不夠（%d／%d）", q+n, len(src))
		}
		return nil
	}
	put := func(di int, v byte, xor bool) error {
		if di < 0 || di >= len(p[plane]) {
			return fmt.Errorf("VRAM 位移 %d 超出畫面", di)
		}
		if xor {
			p[plane][di] ^= v
		} else {
			p[plane][di] = v
		}
		return nil
	}
	// nextCol 換直行（走完 w 個直行換 plane）；回 true ＝ 四個 plane 都走完。
	nextCol := func() bool {
		cnt = 0
		col++
		di = origin + col
		if col == f.Cols {
			col, di = 0, origin
			plane++
		}
		return plane == 4
	}
	if f.Kind == 0 {
		for plane < 4 {
			if err := need(2); err != nil {
				return q, err
			}
			a, c := src[q], src[q+1]
			q += 2
			n := 1
			if a == c {
				if err := need(2); err != nil {
					return q, err
				}
				c, n = src[q], int(src[q+1])
				if n == 0 {
					n = 65536
				}
				q += 2
			}
			for ; n > 0; n-- {
				if err := put(di, a, false); err != nil {
					return q, err
				}
				if err := put(di+80, c, false); err != nil {
					return q, err
				}
				di += 160
				cnt++
				if cnt == f.Height && nextCol() {
					break
				}
			}
		}
		return q, nil
	}
	for plane < 4 {
		if err := need(1); err != nil {
			return q, err
		}
		a := src[q]
		q++
		if a != 0 {
			if err := put(di, a, true); err != nil {
				return q, err
			}
			di += 80
			cnt++
			if cnt == f.Height {
				nextCol()
			}
			continue
		}
		if err := need(1); err != nil {
			return q, err
		}
		n := int(src[q])
		q++
		for {
			if left := f.Height - cnt; n >= left {
				n -= left
				if nextCol() || n == 0 {
					break
				}
				continue
			}
			cnt += n
			di += 80 * n
			break
		}
	}
	return q, nil
}
