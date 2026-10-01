package asset

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// FixedOggSerial 固定單一 Ogg 串流的封裝序號，原始音樂封包完全不改（規格 070）。
// 先核對每頁的原 CRC；錯誤不改輸入，不返回半份輸出。
func FixedOggSerial(data []byte, serial uint32) ([]byte, error) {
	if len(data) == 0 || len(data) > 64<<20 {
		return nil, fmt.Errorf("OGG 為空或超過 64 MiB")
	}
	out := bytes.Clone(data)
	var original uint32
	var sequence uint32
	for p := 0; p < len(out); {
		if len(out)-p < 27 || string(out[p:p+4]) != "OggS" || out[p+4] != 0 {
			return nil, fmt.Errorf("OGG 頁頭／版本／邊界錯誤：0x%X", p)
		}
		segments := int(out[p+26])
		end := p + 27 + segments
		if end > len(out) {
			return nil, fmt.Errorf("OGG 段表截斷：0x%X", p)
		}
		for _, n := range out[p+27 : end] {
			end += int(n)
		}
		if end > len(out) {
			return nil, fmt.Errorf("OGG 封包截斷：0x%X", p)
		}
		page := out[p:end]
		flags := page[5]
		current := binary.LittleEndian.Uint32(page[14:18])
		if p == 0 {
			original = current
		}
		if current != original || binary.LittleEndian.Uint32(page[18:22]) != sequence || flags&^byte(7) != 0 ||
			(flags&2 != 0) != (p == 0) || (flags&4 != 0) != (end == len(out)) {
			return nil, fmt.Errorf("OGG 不是完整單串流或頁序／BOS／EOS 錯誤：0x%X", p)
		}
		crc := binary.LittleEndian.Uint32(page[22:26])
		clear(page[22:26])
		if oggCRC(page) != crc {
			return nil, fmt.Errorf("OGG 原始 CRC 不符：0x%X", p)
		}
		binary.LittleEndian.PutUint32(page[14:18], serial)
		binary.LittleEndian.PutUint32(page[22:26], oggCRC(page))
		sequence++
		p = end
	}
	return out, nil
}

func oggCRC(page []byte) uint32 {
	var value uint32
	for _, b := range page {
		value ^= uint32(b) << 24
		for i := 0; i < 8; i++ {
			if value&0x80000000 != 0 {
				value = value<<1 ^ 0x04C11DB7
			} else {
				value <<= 1
			}
		}
	}
	return value
}
