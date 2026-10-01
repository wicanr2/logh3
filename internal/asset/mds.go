package asset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
)

// MDSMIDI 移植既有匯出器；時間語意仍為 legacy approximation（規格 069）。
// 未解 E4／FB／FC 不猜補，22 首保存 MIDI 是移植的逐位元組基準。
func MDSMIDI(data []byte) ([]byte, error) {
	if len(data) < 256 || len(data) > 1<<20 || !bytes.Equal(data[:2], []byte("M1")) || (len(data)-256)%4 != 0 {
		return nil, fmt.Errorf("MDS 簽章、長度或事件邊界不符")
	}
	type track struct {
		channel byte
		events  [][4]byte
	}
	var tracks []track
	segment := 0
	var events [][4]byte
	flush := func() {
		hasNote := false
		for _, e := range events {
			hasNote = hasNote || e[0] < 128
		}
		if hasNote {
			ch := byte(1)
			if events[0][0] == 0xFC {
				ch = events[0][1]
			} else if segment < 16 {
				for i, c := range data[0x40:0x50] {
					if c == 0xFF {
						break
					}
					if i == segment {
						ch = c
					}
				}
			}
			tracks = append(tracks, track{ch, events})
		}
		segment++
		events = nil
	}
	for p := 256; p < len(data); p += 4 {
		e := [4]byte(data[p : p+4])
		if e[0] == 0xFE {
			if e != [4]byte{0xFE, 0xFE, 0xFE, 0xFE} {
				return nil, fmt.Errorf("MDS 軌界損壞：檔案偏移 0x%X", p)
			}
			flush()
		} else {
			events = append(events, e)
		}
	}
	flush()
	if len(tracks) == 0 || len(tracks) > 65535 {
		return nil, fmt.Errorf("MDS 沒有可匯出的音符軌或軌數超限")
	}
	counts := map[int]int{}
	var lengths []int // 同票採首次出現，與 Python Counter 相同。
	bars, saneBars := 0, 0
	for _, t := range tracks {
		acc := 0
		for _, e := range t.events {
			switch e[0] {
			case 0xFD:
				bars++
				if e[1] <= 2 {
					saneBars++
				}
				if acc != 0 {
					if counts[acc] == 0 {
						lengths = append(lengths, acc)
					}
					counts[acc]++
				}
				acc = 0
			case 0xE0, 0xE2, 0xFA:
				acc += int(e[1])
			default:
				if e[0] < 128 {
					acc += int(e[1])
				} else {
					acc += int(e[0]) // 舊匯出器 raw 的近似，非已證實步進。
				}
			}
		}
	}
	barlen, best := 192, 0
	for _, n := range lengths {
		if counts[n] > best {
			barlen, best = n, counts[n]
		}
	}
	sane := bars > 0 && saneBars*10 > bars*9
	out := bytes.NewBufferString("MThd")
	put32 := func(b *bytes.Buffer, n int) { _ = binary.Write(b, binary.BigEndian, uint32(n)) }
	put32(out, 6)
	_ = binary.Write(out, binary.BigEndian, uint16(1))
	_ = binary.Write(out, binary.BigEndian, uint16(len(tracks)))
	_ = binary.Write(out, binary.BigEndian, uint16(48))
	type message struct {
		at, priority int
		data         []byte
	}
	for _, t := range tracks {
		ch := max(0, min(15, int(t.channel)-1))
		var messages []message
		bar, tick := 0, 0
		for _, e := range t.events {
			at := bar*barlen + tick
			switch {
			case e[0] < 128:
				if e[2] != 0 && e[3] != 0 {
					messages = append(messages, message{at, 0, []byte{byte(0x90 | ch), e[0] & 127, e[3] & 127}},
						message{at + int(e[2]), -1, []byte{byte(0x80 | ch), e[0] & 127, 0}})
				}
				tick += int(e[1])
			case e[0] == 0xFD:
				if sane {
					bar += int(e[1])
				} else {
					bar++
				}
				tick = 0
			case e[0] == 0xFB || e[0] == 0xFC:
				// 舊匯出器不展開反覆、不推進時間。
			case e[0] == 0xE0:
				messages = append(messages, message{at, 0, []byte{byte(0xC0 | ch), e[2] & 127}})
				tick += int(e[1])
			case e[0] == 0xE2:
				messages = append(messages, message{at, 0, []byte{byte(0xB0 | ch), e[2] & 127, e[3] & 127}})
				tick += int(e[1])
			case e[0] == 0xFA:
				us := 60000000 / max(1, int(e[2]))
				messages = append(messages, message{at, -2, []byte{0xFF, 0x51, 3, byte(us >> 16), byte(us >> 8), byte(us)}})
				tick += int(e[1])
			default:
				tick += int(e[0]) // 保留未解 raw 的既有時間近似。
			}
		}
		sort.SliceStable(messages, func(i, j int) bool {
			if messages[i].at != messages[j].at {
				return messages[i].at < messages[j].at
			}
			return messages[i].priority < messages[j].priority
		})
		var payload bytes.Buffer
		last := 0
		for _, m := range messages {
			payload.Write(midiVLQ(max(0, m.at-last)))
			payload.Write(m.data)
			last = max(last, m.at)
		}
		payload.Write([]byte{0, 0xFF, 0x2F, 0})
		out.WriteString("MTrk")
		put32(out, payload.Len())
		out.Write(payload.Bytes())
	}
	return out.Bytes(), nil
}

func midiVLQ(n int) []byte {
	b := []byte{byte(n & 127)}
	for n >>= 7; n != 0; n >>= 7 {
		b = append(b, byte(n&127)|128)
	}
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}
